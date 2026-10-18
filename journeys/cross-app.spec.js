const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { spawn } = require('node:child_process');
const { randomUUID } = require('node:crypto');
const { stripVTControlCharacters } = require('node:util');

const goo = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
const app = process.env.GOO_REPORT_URL || 'http://127.0.0.1:8091';
const frames = path.resolve('docs/journeys/screenshots');

// This test adapter owns physical ClickHouse columns. It deliberately does not
// use goo's generated SQL or API to establish what the source captured.
async function source(request, sql, parameters) {
  const url = new URL(process.env.GOO_TEST_CLICKHOUSE_URL || 'http://127.0.0.1:8123');
  for (const [key, value] of Object.entries({ readonly: 1, max_execution_time: 5,
    max_result_rows: 101, result_overflow_mode: 'throw', max_bytes_to_read: 268435456,
    read_overflow_mode: 'throw', ...parameters })) url.searchParams.set(key, String(value));
  const user = process.env.GOO_TEST_CLICKHOUSE_USER || 'lab';
  const password = process.env.GOO_TEST_CLICKHOUSE_PASSWORD || 'lab-local';
  const response = await request.post(url.href, { data: sql,
    headers: { Authorization: 'Basic ' + Buffer.from(user + ':' + password).toString('base64') } });
  expect(response.ok(), 'independent source query failed: HTTP ' + response.status()).toBe(true);
  const text = (await response.text()).trim();
  return text ? text.split('\n').map(line => JSON.parse(line)) : [];
}

test('26 correlate a real report with goo while Bombadil explores, optionally checking Quint', async ({ page, request }, testInfo) => {
  const run = randomUUID(), dir = testInfo.outputPath('cross-app');
  await fs.mkdir(dir, { recursive: true });
  await fs.mkdir(frames, { recursive: true });
  const ledger = { version: 1, run, events: [], experiment: 'real-report-with-concurrent-goo-navigation' };
  const record = (action, data) => ledger.events.push({ action, run, observedAt: new Date().toISOString(), ...data });
  const persist = () => fs.writeFile(path.join(dir, 'ledger.json'), JSON.stringify(ledger, null, 2));
  const logs = [];
  // Launch the pinned native binary so cancellation reaches Bombadil's SIGINT
  // handler, rather than terminating npm's synchronous wrapper and orphaning it.
  const nativeName = { 'linux-x64': 'bombadil-linux-x64', 'linux-arm64': 'bombadil-linux-arm64',
    'darwin-arm64': 'bombadil-darwin-arm64' }[process.platform + '-' + process.arch];
  expect(nativeName, 'platform supported by the pinned Bombadil package').toBeTruthy();
  const explorer = spawn(path.resolve('node_modules/@antithesishq/bombadil/binaries', nativeName), [
    'browser', 'test', '--headless', '--no-sandbox', '--time-limit', '30s',
    '--exit-on-violation', '--output-path', path.join(dir, 'bombadil'), goo,
    'verification/bombadil/workspace.ts',
  ], { cwd: path.resolve('.'), stdio: ['ignore', 'pipe', 'pipe'], timeout: 65000, killSignal: 'SIGINT' });
  const explorationStartedAt = Date.now();
  explorer.stdout.on('data', data => logs.push(data.toString()));
  explorer.stderr.on('data', data => logs.push(data.toString()));
  let exited = false;
  const finished = new Promise(resolve => {
    explorer.on('error', error => { exited = true; resolve({ error: error.message }); });
    explorer.on('close', (code, signal) => { exited = true; resolve({ code, signal, endedAt: Date.now() }); });
  });
  try {
    await expect.poll(() => stripVTControlCharacters(logs.join('')).includes('goo.navigate('), { timeout: 20000 }).toBe(true);
    ledger.explorationStartedAt = explorationStartedAt;
    ledger.workloadStartedAt = Date.now();
    expect(exited, 'explorer must still be active when the workload starts').toBe(false);
    await page.goto(app);
    await page.getByRole('button', { name: 'Generate report', exact: true }).click();
    await expect(page.locator('#result')).toContainText('ready');
    const report = JSON.parse(await page.locator('#result').innerText());
    expect(report.outcome).toBe('ready');
    expect(report.id).toMatch(/^[A-Za-z0-9]+$/);
    record('accepted', { actor: 'application', subject: report.id });
    await page.screenshot({ path: path.join(frames, '26a-correlated-real-report.png'), fullPage: true });

    const discoverySQL = `SELECT TraceId, SpanId, arrayFirst(a -> a['goo.subject.id'] = {subject:String}, \`Events.Attributes\`)['goo.subject.id'] AS subject
      FROM default.otel_traces WHERE ServiceName = {service:String}
      AND Timestamp >= fromUnixTimestamp64Milli({from:Int64})
      AND ResourceAttributes['deployment.environment.name'] = {environment:String}
      AND arrayExists(a -> a['goo.subject.id'] = {subject:String}
        AND a['goo.subject.namespace'] = 'report' AND a['goo.subject.tenant'] = 'demo', \`Events.Attributes\`)
      ORDER BY SpanId LIMIT 100 FORMAT JSONEachRow`;
    const discoveryParameters = { param_subject: report.id, param_service: 'lab-go-reports',
      param_environment: 'local', param_from: ledger.workloadStartedAt - 60000 };
    let discovery;
    await expect.poll(async () => {
      discovery = await source(request, discoverySQL, discoveryParameters);
      return discovery.length;
    }, { timeout: 60000, intervals: [1000] }).toBeGreaterThan(0);
    const trace = discovery[0].TraceId;
    expect(trace).toMatch(/^[a-f0-9]{32}$/);
    expect(discovery.every(row => row.TraceId === trace && row.subject === report.id)).toBe(true);
    ledger.discovery = { sql: discoverySQL, parameters: discoveryParameters, rows: discovery };

    // Use ordinary UI controls; observe the actual frozen question sent by goo.
    await page.goto(goo + '/#explore');
    await page.locator('#minutes').fill('5');
    await page.locator('#exploreField').selectOption(JSON.stringify({ kind: 'column', key: 'trace_id' }));
    await page.locator('#exploreValue').fill(trace);
    await page.locator('#exploreAdd').click();
    const responsePromise = page.waitForResponse(r => r.url().endsWith('/api/explore') && r.request().method() === 'POST');
    await page.locator('#exploreRun').click();
    const response = await responsePromise;
    expect(response.ok()).toBe(true);
    const question = response.request().postDataJSON(), result = await response.json();
    expect(question.signal).toBe('spans');
    expect(question.environment).toBe('local');
    expect(question.where).toEqual({ all: [{ field: { kind: 'column', key: 'trace_id' }, op: 'eq', value: trace }] });
    const sql = `SELECT TraceId, SpanId, ResourceAttributes['deployment.environment.name'] AS environment
      FROM default.otel_traces WHERE Timestamp >= fromUnixTimestamp64Milli({from:Int64})
      AND Timestamp < fromUnixTimestamp64Milli({to:Int64})
      AND mapContains(ResourceAttributes, 'deployment.environment.name')
      AND ResourceAttributes['deployment.environment.name'] = {environment:String}
      AND TraceId = {trace:String} ORDER BY SpanId LIMIT 101 FORMAT JSONEachRow`;
    const parameters = { param_trace: trace, param_environment: question.environment,
      param_from: question.fromMs, param_to: question.toMs };
    const rows = await source(request, sql, parameters);
    expect(rows.length).toBeGreaterThan(0);
    expect(rows.length).toBeLessThanOrEqual(100);
    const members = rows.map(row => row.SpanId).sort().join(',');
    const evidence = { subject: discovery[0].subject, trace, environment: question.environment,
      fromMs: question.fromMs, toMs: question.toMs, count: rows.length, members };
    record('source_ready', { actor: 'source-oracle', ...evidence });
    ledger.oracle = { sql, parameters, rows };

    expect(result.state).toBe('observed');
    expect(result.count).toBe(String(rows.length));
    expect({ ...result.question, groupBy: result.question.groupBy || [] }).toEqual(question);
    expect(result.rows.map(row => row.SpanId).sort().join(',')).toBe(members);
    await expect(page.locator('#exploreStatus')).toContainText(`${rows.length} matching captured spans`);
    // Check rendered records, not just the response used to populate them.
    const rendered = await page.locator('#exploreRows details pre').allTextContents();
    expect(rendered.map(text => JSON.parse(text).SpanId).sort().join(',')).toBe(members);
    expect(rendered.map(text => JSON.parse(text).TraceId)).toEqual(rows.map(() => trace));
    const displayedQuestion = JSON.parse(await page.locator('#exploreExplain').textContent()).question;
    expect(displayedQuestion).toEqual(question);
    record('displayed', { actor: 'goo-observer', ...evidence, count: rendered.length,
      members: rendered.map(text => JSON.parse(text).SpanId).sort().join(','), basis: result.basis });
    ledger.question = question;
    await page.locator('#exploreRows').getByRole('button', { name: 'Expand trace ' + trace, exact: true }).first().click();
    await expect(page.locator('#exploreContextStatus')).toContainText('across all services');
    await expect(page.locator('#exploreTrace .waterfall-row')).toHaveCount(rows.length);
    await page.screenshot({ path: path.join(frames, '26b-correlated-goo-trace.png'), fullPage: true });

    const outcome = await finished;
    const navigationActions = stripVTControlCharacters(logs.join('')).match(/goo\.navigate\(/g)?.length || 0;
    ledger.exploration = { ...outcome, navigationActions };
    expect(outcome.code, logs.join('').slice(-3000)).toBe(0);
    expect(navigationActions).toBeGreaterThan(0);
    expect(outcome.endedAt).toBeGreaterThan(ledger.workloadStartedAt);
    record('explored', { actor: 'bombadil-summary', steps: navigationActions, passed: outcome.code === 0 });
    await persist();

    if (process.env.GOO_CROSS_APP_QUINT === '1') {
      // Optional expensive model validation over this actual run, not a fixture.
      const verifierEnv = { ...process.env, GOO_CROSS_APP_LEDGER: path.join(dir, 'ledger.json'), NO_COLOR: '1' };
      // quint-go parses the CLI's plain verdict lines. Playwright forces color
      // in its worker environment, which must not leak to this subprocess.
      delete verifierEnv.FORCE_COLOR;
      const verifier = spawn('go', ['test', '-count=1', '-v', '-run', '^TestQuintCrossAppLiveLedger$', '.'], {
        cwd: path.resolve('verification'), env: verifierEnv, timeout: 60000,
        stdio: ['ignore', 'pipe', 'pipe'],
      });
      const output = [];
      verifier.stdout.on('data', data => output.push(data.toString()));
      verifier.stderr.on('data', data => output.push(data.toString()));
      const code = await new Promise((resolve, reject) => { verifier.on('error', reject); verifier.on('exit', resolve); });
      await fs.writeFile(path.join(dir, 'quint-validation.log'), output.join(''));
      expect(code, output.join('')).toBe(0);
    }
  } finally {
    if (!exited) { explorer.kill('SIGINT'); await finished; }
    await persist();
    await fs.writeFile(path.join(dir, 'bombadil.log'), logs.join(''));
    await testInfo.attach('cross-app-ledger', { path: path.join(dir, 'ledger.json'), contentType: 'application/json' });
  }
});
