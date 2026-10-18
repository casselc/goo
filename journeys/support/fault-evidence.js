const { expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { source } = require('./cross-app');
const goo = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';

// Test contracts over captured records. These labels are not product SLO results
// or a claim of business durability/capture completeness.
function summarizeCaptured(spans, logs) {
  const retrySeen = logs.some(row => row.Body === 'reservation.retry');
  const stalled = logs.some(row => row.Body === 'reservation.stalled');
  const terminalRows = logs.filter(row => row.Body === 'reservation.terminal');
  const terminalSeen = terminalRows.length > 0;
  const terminals = [...new Set(terminalRows.map(row => row.Attributes['app.outcome'] || ''))];
  const terminal = terminals.length === 1 ? terminals[0] : '';
  const attempts = spans.filter(row => row.SpanName === 'inventory process').length;
  const errors = spans.filter(row => String(row.StatusCode).toLowerCase() === 'error').length;
  let meaning = 'unknown-no-data';
  if (terminals.length > 1) meaning = 'unknown-conflicting-terminals';
  else if (stalled && terminals.length === 0) meaning = 'unknown-no-terminal';
  else if (terminal === 'rejected' && !retrySeen && !stalled) meaning = 'rejected';
  else if (terminal === 'confirmed' && retrySeen && !stalled && attempts === 2 && errors > 0) meaning = 'confirmed-after-retry';
  return { retrySeen, stalled, terminalSeen, terminal, attempts, errors, meaning };
}

async function capturedScenario(dir, scenario) {
  const records = (await fs.readFile(path.join(dir, 'trace.jsonl'), 'utf8')).trim().split('\n').map(JSON.parse);
  expect(records.filter(record => record.action?.Custom?.name === 'sample.submit-scenario')).toHaveLength(1);
  const states = records.flatMap(record => record.snapshots || []).filter(snapshot => snapshot.name === 'scenarioResult')
    .map(snapshot => snapshot.value).filter(value => value.id);
  expect(states.length).toBeGreaterThan(0);
  expect(new Set(states.map(value => value.id)).size).toBe(1);
  expect(states.every(value => value.returnedScenario === scenario && value.scenario === scenario)).toBe(true);
  expect(states[0].id).toMatch(/^[a-f0-9-]{36}$/);
  return { id: states[0].id, scenario };
}

async function readSubject(request, signal, id, fromMs, toMs) {
  // Physical exporter columns stay in this independent test source adapter.
  const spans = signal === 'spans', attributes = spans ? 'SpanAttributes' : 'LogAttributes';
  const sql = `SELECT TraceId, SpanId, ServiceName, toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns,
    ${attributes} AS Attributes, ${spans ? 'SpanName, StatusCode' : 'Body'}
    FROM default.${spans ? 'otel_traces' : 'otel_logs'}
    WHERE Timestamp >= fromUnixTimestamp64Milli({from:Int64}) AND Timestamp < fromUnixTimestamp64Milli({to:Int64})
    AND ResourceAttributes['deployment.environment.name'] = {environment:String}
    AND ${attributes}['app.reservation.id'] = {subject:String} ORDER BY Timestamp, TraceId, SpanId LIMIT 101 FORMAT JSONEachRow`;
  const parameters = { param_from: fromMs, param_to: toMs, param_environment: 'local', param_subject: id };
  const rows = await source(request, sql, parameters);
  return { sql, parameters, rows };
}

const expectedMeaning = { retry: 'confirmed-after-retry', reject: 'rejected', missing: 'unknown-no-terminal' };
const canonical = value => value && typeof value === 'object'
  ? Array.isArray(value) ? value.map(canonical) : Object.fromEntries(Object.keys(value).sort().map(key => [key, canonical(value[key])])) : value;
function members(rows, signal) {
  return rows.map(row => JSON.stringify(canonical({ trace: row.TraceId, span: row.SpanId, service: row.ServiceName,
    time: row.timestamp_ns, name: signal === 'spans' ? row.SpanName : row.Body, attributes: row.Attributes,
    ...(signal === 'spans' ? { status: row.StatusCode } : {}) }))).sort();
}

async function inspectScenario({ page, request, result, startedAt, frameName }) {
  // Positive evidence gates readiness. Missing capture alone never satisfies it.
  const fromMs = startedAt - 60000, toMs = Date.now();
  let readiness;
  await expect.poll(async () => {
    const spans = await readSubject(request, 'spans', result.id, fromMs, toMs);
    const logs = await readSubject(request, 'logs', result.id, fromMs, toMs);
    readiness = { spans, logs, facts: summarizeCaptured(spans.rows, logs.rows), fromMs, toMs };
    return readiness.facts.meaning;
  }, { timeout: 45000, intervals: [1000] }).toBe(expectedMeaning[result.scenario]);
  const evidence = { result, readiness, reads: {} };
  const rendered = {};
  for (const signal of ['spans', 'logs']) {
    await page.goto(goo + '/#explore');
    if (await page.locator('#exploreContextClose').isVisible()) await page.locator('#exploreContextClose').click();
    await page.locator('#exploreClear').click();
    await page.locator('#exploreSignal').selectOption(signal);
    await page.locator('#minutes').fill('5');
    const run = async () => {
      const response = page.waitForResponse(r => r.url().endsWith('/api/explore') && r.request().method() === 'POST');
      await page.locator('#exploreRun').click();
      const r = await response; expect(r.ok()).toBe(true);
      return { question: r.request().postDataJSON(), data: await r.json() };
    };
    await run(); // Discover ordinary OTel attributes from the actual source.
    await page.locator('#exploreField').selectOption(JSON.stringify({ kind: 'attribute', key: 'app.reservation.id' }));
    await page.locator('#exploreValue').fill(result.id);
    await page.locator('#exploreAdd').click();
    const shown = await run(), q = shown.question;
    expect(q.environment).toBe('local'); expect(q.signal).toBe(signal);
    expect(q.where).toEqual({ all: [{ field: { kind: 'attribute', key: 'app.reservation.id' }, op: 'eq', value: result.id }] });
    const oracle = await readSubject(request, signal, result.id, q.fromMs, q.toMs);
    expect(oracle.rows.length).toBeGreaterThan(0); expect(oracle.rows.length).toBeLessThanOrEqual(100);
    expect(shown.data.state).toBe('observed'); expect(shown.data.count).toBe(String(oracle.rows.length));
    await expect(page.locator('#exploreStatus')).toContainText(`${oracle.rows.length} matching captured ${signal}`);
    rendered[signal] = (await page.locator('#exploreRows details pre').allTextContents()).map(JSON.parse);
    expect(members(rendered[signal], signal)).toEqual(members(oracle.rows, signal));
    const displayedQuestion = JSON.parse(await page.locator('#exploreExplain').textContent()).question;
    expect(displayedQuestion).toEqual(q);
    evidence.reads[signal] = { question: q, displayedQuestion, basis: shown.data.basis, oracle, rendered: rendered[signal] };
  }
  const observed = summarizeCaptured(rendered.spans, rendered.logs);
  expect(observed.meaning).toBe(expectedMeaning[result.scenario]);
  evidence.observed = observed;
  // Exercise the actual disclosure, then leave the full log sequence visible
  // for documentation instead of allowing one expanded JSON row to obscure it.
  const body = result.scenario === 'missing' ? 'reservation.stalled' : 'reservation.terminal';
  const row = page.locator('#exploreRows .record-detail-row').filter({ hasText: body }).first();
  const disclosure = row.getByText('Captured attributes', { exact: true });
  await disclosure.click();
  const key = result.scenario === 'missing' ? 'app.scenario' : 'app.outcome';
  const attribute = row.locator('.attribute-row').filter({ has: page.locator('.attribute-key').filter({ hasText: new RegExp('^' + key.replaceAll('.', '\\.') + '$') }) });
  await expect(attribute).toBeVisible();
  await expect(attribute.locator('.attribute-value-text')).toHaveText(result.scenario === 'missing' ? 'missing' : result.scenario === 'retry' ? 'confirmed' : 'rejected');
  await disclosure.click();
  await page.screenshot({ path: path.join('docs/journeys/screenshots', frameName), fullPage: true });
  return evidence;
}

module.exports = { summarizeCaptured, capturedScenario, readSubject, inspectScenario, expectedMeaning };
