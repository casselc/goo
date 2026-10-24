const {openFilters,inspectedRecords}=require('./explore');
const { expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { spawn } = require('node:child_process');
const { stripVTControlCharacters } = require('node:util');
const goo = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
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

async function inspectReport({ page, request, report, ledger, record, frameName }) {
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
    await page.goto(goo + '/#questions');
    // Navigating to the current hash may keep the document and inspector alive.
    // A revisit explicitly closes context and starts a fresh trace selection.
    if (await page.locator('#exploreContextClose').isVisible()) await page.locator('#exploreContextClose').click();
    await require('./explore').clearFilters(page);
    await page.locator('#minutes').fill('5');
    await require('./explore').addFilter(page,{kind:'column',key:'trace_id'},'eq',trace);
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
    const rendered = (await inspectedRecords(page)).map(record=>JSON.stringify(record));
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
    await page.screenshot({ path: path.join(frames, frameName), fullPage: true });
}

async function startBombadil({ dir, url, specification, seconds }) {
  await fs.mkdir(dir, { recursive: true });
  const nativeName = { 'linux-x64': 'bombadil-linux-x64', 'linux-arm64': 'bombadil-linux-arm64',
    'darwin-arm64': 'bombadil-darwin-arm64' }[process.platform + '-' + process.arch];
  expect(nativeName).toBeTruthy();
  const logs = [], startedAt = Date.now();
  const activeFile = path.join(dir, '.active');
  await fs.writeFile(activeFile, JSON.stringify({startedAt, owner:'bombadil'}), {flag:'wx'});
  const processHandle = spawn(path.resolve('node_modules/@antithesishq/bombadil/binaries', nativeName), [
    'browser', 'test', '--headless', '--no-sandbox', '--time-limit', seconds + 's',
    '--exit-on-violation', '--output-path', dir, url, specification,
  ], { stdio: ['ignore', 'pipe', 'pipe'], timeout: (seconds + 30) * 1000, killSignal: 'SIGINT' });
  processHandle.stdout.on('data', data => logs.push(data.toString()));
  processHandle.stderr.on('data', data => logs.push(data.toString()));
  let exited = false;
  const finished = new Promise(resolve => {
    let completing = false;
    const complete = async result => {
      if (completing) return; completing = true; exited = true;
      try {
        const record = {...result, startedAt, endedAt:Date.now(), status:result.code===0?'passed':'failed'};
        const temporary = path.join(dir, 'run-result.json.tmp');
        await fs.writeFile(temporary, JSON.stringify(record, null, 2));
        await fs.rename(temporary, path.join(dir, 'run-result.json'));
        await fs.rm(activeFile, {force:true});
        resolve(record);
      } catch (error) { resolve({error:error.message}); }
    };
    processHandle.on('error', error => complete({error:error.message}));
    processHandle.on('close', (code, signal) => complete({code, signal}));
  });
  const save = () => fs.writeFile(path.join(dir, 'process.log'), logs.join(''));
  return {
    active: () => !exited,
    async ready(action) {
      await expect.poll(() => stripVTControlCharacters(logs.join('')).includes(action), { timeout: 20000 }).toBe(true);
    },
    async finish() {
      const result = await finished; await save();
      expect(result.code, logs.join('').slice(-3000)).toBe(0);
      return result;
    },
    async stop() { if (!exited) { processHandle.kill('SIGINT'); await finished; } await save(); },
  };
}

async function verifyWithGo(dir, testName, variables) {
  const env = { ...process.env, ...variables, NO_COLOR: '1' };
  delete env.FORCE_COLOR;
  const output = [];
  const child = spawn('go', ['test', '-count=1', '-v', '-run', '^' + testName + '$', '.'], {
    cwd: path.resolve('verification'), env, timeout: 60000, stdio: ['ignore', 'pipe', 'pipe'],
  });
  child.stdout.on('data', data => output.push(data.toString()));
  child.stderr.on('data', data => output.push(data.toString()));
  const code = await new Promise((resolve, reject) => { child.on('error', reject); child.on('close', resolve); });
  await fs.writeFile(path.join(dir, testName + '.log'), output.join(''));
  expect(code, output.join('')).toBe(0);
}

async function capturedReport(dir) {
  const records = (await fs.readFile(path.join(dir, 'trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
  const submissions = records.filter(record => record.action?.Custom?.name === 'sample.generate-report');
  expect(submissions, 'one generated submission, independently of the result property').toHaveLength(1);
  const states = records.flatMap(record => record.snapshots || []).filter(snapshot => snapshot.name === 'reportResult')
    .map(snapshot => snapshot.value).filter(value => value.id);
  expect(states.length, 'actual returned identity must be captured').toBeGreaterThan(0);
  expect(new Set(states.map(value => value.id)).size).toBe(1);
  expect(states.every(value => value.outcome === 'ready')).toBe(true);
  expect(states[0].id).toMatch(/^[A-Za-z0-9]+$/);
  return { id: states[0].id, outcome: states[0].outcome };
}

module.exports = { inspectReport, startBombadil, verifyWithGo, capturedReport, source };
