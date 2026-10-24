'use strict';

const {test, expect} = require('@playwright/test');
const {randomBytes} = require('node:crypto');
const fs = require('node:fs/promises');
const path = require('node:path');
const {questionLink} = require('../ui/workspace');
const {startBombadil} = require('./support/cross-app');
const base = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
const patternField = {kind: 'column', key: 'body_pattern_v1'};
const isResponse = (endpoint, response) => response.request().method() === 'POST' && new URL(response.url()).pathname === '/api/' + endpoint;
async function json(response) {
  expect(response.ok(), await response.text()).toBe(true);
  return response.json();
}

// Independent byte scanner: no production normalizer/SQL/display helper is used.
// OTLP JSON strings are valid Unicode; malformed physical UTF-8 is tested by the
// independent native/backend corpus, not invented through a JSON source fixture.
function pattern(body) {
  const bytes = Buffer.from(body, 'utf8');
  if (bytes.length === 0) return 'lp1:e';
  if (bytes.length > 256) return 'lp1:o';
  const chunks = [];
  let literal = [], digits = false;
  for (const byte of bytes) {
    if (byte >= 48 && byte <= 57) {
      if (!digits) {chunks.push(Buffer.from(literal).toString('hex')); literal = []}
      digits = true;
    } else {literal.push(byte); digits = false}
  }
  chunks.push(Buffer.from(literal).toString('hex'));
  return 'lp1:p:' + chunks.join('n');
}
function value(record, field) {
  if (field.kind === 'resource') return Object.hasOwn(record.resource, field.key) ? record.resource[field.key] : null;
  if (field.kind === 'attribute') return Object.hasOwn(record.attributes, field.key) ? record.attributes[field.key] : null;
  return {service: record.service, body: record.body, severity: record.severity, trace_id: record.traceId, span_id: record.id, body_pattern_v1: pattern(record.body)}[field.key];
}
function matches(record, predicate) {
  if (!predicate) return true;
  if (predicate.all) return predicate.all.every(item => matches(record, item));
  if (predicate.any) return predicate.any.some(item => matches(record, item));
  if (predicate.not) return !matches(record, predicate.not);
  const actual = value(record, predicate.field), literal = predicate.value ?? '';
  if (predicate.op === 'missing') return actual === null;
  if (predicate.op === 'exists') return actual !== null;
  if (predicate.op === 'eq') return actual !== null && actual === literal;
  if (predicate.op === 'ne') return actual !== null && actual !== literal;
  throw Error('Unknown independent predicate ' + JSON.stringify(predicate));
}
function members(records, question) {
  return records.filter(record => question.signal === 'logs' && record.environment === question.environment && record.timeMs >= question.fromMs && record.timeMs < question.toMs && matches(record, question.where));
}
function assertResult(result, records) {
  const question = result.question, rows = members(records, question);
  expect(result.count).toBe(String(rows.length));
  expect(result.rows).toHaveLength(Math.min(question.limit, Math.max(0, rows.length - question.offset)));
  const expectedRows = new Map(rows.map(record => [record.id, record]));
  const pageIDs = new Set();
  for (const row of result.rows) {
    const expected = expectedRows.get(row.SpanId);
    expect(expected).toBeTruthy();
    expect(row.Body).toBe(expected.body);
    expect(row.timestamp_ns).toBe(String(BigInt(expected.timeMs) * 1000000n));
    expect(pageIDs.has(row.SpanId)).toBe(false);
    pageIDs.add(row.SpanId);
  }
  if (question.groupBy?.length) {
    const groups = new Map();
    for (const record of rows) {
      const key = JSON.stringify(question.groupBy.map(field => value(record, field)));
      groups.set(key, (groups.get(key) || 0n) + 1n);
    }
    expect(result.groups).toHaveLength(groups.size);
    for (const group of result.groups) expect(group.count).toBe(String(groups.get(JSON.stringify(question.groupBy.map((_, index) => group['g' + index])))));
    expect(result.groups.reduce((count, group) => count + BigInt(group.count), 0n)).toBe(BigInt(rows.length));
  }
  if (question.bucketMs) {
    expect(result.timeSeries).toHaveLength(Math.ceil((question.toMs - question.fromMs) / question.bucketMs));
    for (const point of result.timeSeries) expect(point.count).toBe(String(rows.filter(record => record.timeMs >= point.fromMs && record.timeMs < point.toMs).length));
    expect(result.timeSeries.reduce((count, point) => count + BigInt(point.count), 0n)).toBe(BigInt(rows.length));
  }
  return rows;
}

const bodies = {
  common: index => 'handled request id=' + (1000 + index),
  rising: index => 'retry delay=' + (1000 + index) + 'ms',
  currentOnly: index => 'checkout event=' + (1000 + index),
  baselineOnly: index => 'cache refresh generation=' + (1000 + index),
  lateOnly: index => 'baseline late record id=' + (1000 + index),
  empty: () => '',
  oversize: () => 'x'.repeat(257)
};
function fraction(numerator, denominator) {
  let a = numerator < 0n ? -numerator : numerator, b = denominator;
  while (b) [a, b] = [b, a % b];
  numerator /= a; denominator /= a;
  return denominator === 1n ? String(numerator) : numerator + '/' + denominator;
}
function groupOracle(records, question) {
  const counts = new Map();
  for (const record of members(records, question)) {
    const id = pattern(record.body);
    counts.set(id, (counts.get(id) || 0n) + 1n);
  }
  return counts;
}
function comparisonOracle(records, baseline, selection) {
  const a = groupOracle(records, baseline), z = groupOracle(records, selection);
  const ac = BigInt(members(records, baseline).length), zc = BigInt(members(records, selection).length);
  const ad = BigInt(baseline.toMs - baseline.fromMs), zd = BigInt(selection.toMs - selection.fromMs);
  return [...new Set([...a.keys(), ...z.keys()])].sort().map(id => {
    const an = a.get(id) || 0n, zn = z.get(id) || 0n;
    return {
      id, baselineCount: String(an), selectionCount: String(zn), countDifference: String(zn - an),
      baselineRate: fraction(an * 1000n, ad), selectionRate: fraction(zn * 1000n, zd),
      rateDifference: fraction(1000n * (zn * ad - an * zd), ad * zd),
      baselineShare: ac ? fraction(an * 100n, ac) : null,
      selectionShare: zc ? fraction(zn * 100n, zc) : null,
      shareDifferencePP: ac && zc ? fraction(100n * (zn * ac - an * zc), ac * zc) : null,
      observation: an === 0n ? 'newly-observed' : zn === 0n ? 'absent' : 'observed'
    };
  });
}
const attributes = values => Object.entries(values).map(([key, stringValue]) => ({key, value: {stringValue}}));
async function emit(request, records) {
  const resourceLogs = records.map(record => ({
    resource: {attributes: attributes({'service.name': record.service, 'deployment.environment.name': record.environment, ...record.resource})},
    scopeLogs: [{scope: {name: 'ordinary.log.pattern.time.worker'}, logRecords: [{
      timeUnixNano: String(BigInt(record.timeMs) * 1000000n), traceId: record.traceId, spanId: record.id,
      severityText: 'INFO', body: {stringValue: record.body}, attributes: attributes(record.attributes)
    }]}]
  }));
  expect((await request.post('http://127.0.0.1:4318/v1/logs', {data: {resourceLogs}})).ok()).toBe(true);
}
async function fixture(request) {
  const environment = 'log-pattern-time-' + randomBytes(5).toString('hex'), now = Date.now();
  const windows = {
    current: {fromMs: now - 60000, toMs: now},
    previous: {fromMs: now - 120000, toMs: now - 60000},
    saved: {fromMs: now - 300000, toMs: now - 180000}
  };
  const selector = {signal: 'logs', environment, where: {all: [
    {field: {kind: 'resource', key: 'service.namespace'}, op: 'eq', value: 'commerce'},
    {any: [
      {field: {kind: 'column', key: 'service'}, op: 'eq', value: 'checkout-api'},
      {field: {kind: 'column', key: 'service'}, op: 'eq', value: 'queue-worker'}
    ]},
    {field: {kind: 'attribute', key: 'team'}, op: 'eq', value: 'core'}
  ]}};
  function recordsFor(phase, counts) {
    const window = windows[phase], duration = window.toMs - window.fromMs;
    const kinds = Object.entries(counts).flatMap(([kind, count]) => Array.from({length: count}, () => kind));
    return kinds.map((kind, index) => ({
      phase, kind, environment, signal: 'logs', id: randomBytes(8).toString('hex'), traceId: randomBytes(16).toString('hex'),
      timeMs: window.fromMs + 500 + Math.floor(index * (duration - 1000) / kinds.length),
      service: index % 3 === 0 ? 'checkout-api' : 'queue-worker', severity: 'INFO',
      body: bodies[kind](index), attributes: {team: 'core', 'work.sequence': String(index)}, resource: {'service.namespace': 'commerce'}
    }));
  }
  const records = [
    ...recordsFor('current', {common: 160, rising: 160, currentOnly: 40, empty: 20, oversize: 20}),
    ...recordsFor('previous', {common: 160, rising: 40, baselineOnly: 160, empty: 20, oversize: 20}),
    ...recordsFor('saved', {common: 320, rising: 40, baselineOnly: 20, empty: 10, oversize: 10})
  ];
  const excluded = records.filter((_, index) => index % 40 === 0).map((record, index) => ({
    ...record, id: randomBytes(8).toString('hex'), phase: 'excluded',
    service: index % 3 === 0 ? 'unrelated-worker' : record.service,
    attributes: {team: index % 3 === 1 ? 'other' : 'core'},
    resource: {'service.namespace': index % 3 === 2 ? 'other' : 'commerce'}
  }));
  records.push(...excluded);
  await emit(request, records);
  const current = {...selector, ...windows.current, groupBy: [patternField], bucketMs: 1000, limit: 100, offset: 0};
  const previous = {...current, ...windows.previous};
  const savedQuestion = {...selector, ...windows.saved, groupBy: [{kind: 'column', key: 'service'}], bucketMs: 10000, limit: 100, offset: 0};
  const readyProbes = [];
  for (const question of [current, previous, savedQuestion]) {
    let ready;
    await expect.poll(async () => {ready = await json(await request.post(base + '/api/explore', {data: question})); return ready.count}, {timeout: 45000, intervals: [500]}).toBe('400');
    assertResult(ready, records); readyProbes.push(ready);
  }
  const scopeCapture = await json(await request.post(base + '/api/captures', {data: {title: 'Original120-second service scope for pattern comparison', question: {kind: 'explore', explore: savedQuestion}}}));
  expect(scopeCapture.explore.count).toBe('400');
  return {version: 1, environment, records, windows, selector, current, previous, savedQuestion, readyProbes, scopeCapture};
}
function lateRecords(fixture) {
  const window = fixture.windows.saved;
  return Array.from({length: 12}, (_, index) => ({
    phase: 'late-saved-baseline', kind: 'lateOnly', environment: fixture.environment, signal: 'logs',
    id: randomBytes(8).toString('hex'), traceId: randomBytes(16).toString('hex'),
    timeMs: window.fromMs + 1000 + index * 100,
    service: index % 2 ? 'queue-worker' : 'checkout-api', severity: 'INFO', body: bodies.lateOnly(index),
    attributes: {team: 'core', 'work.sequence': 'late-' + index}, resource: {'service.namespace': 'commerce'}
  }));
}

function watch(page) {
  const state = {sourceOperations: 0, protocols: 0, writes: 0, captureGets: 0, events: [], requests: [], sourceResponses: [], failedRequests: [], errors: []};
  page.on('pageerror', error => state.errors.push(error.message));
  page.on('request', request => {
    const endpoint = new URL(request.url()).pathname;
    if (request.method() === 'GET' && /^\/api\/captures\/[a-f0-9]{64}(?:\/receipt)?$/.test(endpoint)) {state.captureGets++; state.requests.push({effect: 'stored-scope', endpoint}); return}
    if (request.method() !== 'POST') return;
    const payload = request.postDataJSON();
    let effect;
    if (['/api/explore', '/api/captures'].includes(endpoint)) {state.sourceOperations++; effect = 'source'}
    else if (endpoint === '/api/compare') {state.protocols++; effect = 'comparison'}
    else if (endpoint === '/api/exploration-views') {state.writes++; effect = 'configuration'}
    if (effect) state.requests.push({effect, endpoint, payload});
  });
  page.on('response', async response => {
    if (!isResponse('explore', response)) return;
    try {state.sourceResponses.push({request: response.request().postDataJSON(), status: response.status(), response: await response.json()})}
    catch (error) {state.errors.push('Could not retain source response: ' + error.message)}
  });
  page.on('requestfailed', request => {
    if (request.method() === 'POST' && new URL(request.url()).pathname === '/api/explore') state.failedRequests.push({endpoint: '/api/explore', request: request.postDataJSON(), failure: request.failure()});
  });
  state.event = action => state.events.push({action, ...Object.fromEntries(['sourceOperations', 'protocols', 'writes', 'captureGets'].map(key => [key, state[key]]))});
  return state;
}
async function storedBytes(id) {
  expect(id).toMatch(/^[a-f0-9]{64}$/);
  const {promisify} = require('node:util'), {execFile} = require('node:child_process');
  const environment = {...process.env}; for (const key of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete environment[key];
  const {stdout} = await promisify(execFile)('docker', ['--host=unix:///var/run/docker.sock', 'exec', process.env.GOO_QUERY_CONTAINER || 'grammar-observability-packages-1', 'cat', '/app/artifacts/investigations/exploration/' + id + '.json'], {env: environment, encoding: 'buffer', maxBuffer: 4 * 1024 * 1024});
  return stdout;
}
function hash(bytes) {return require('node:crypto').createHash('sha256').update(bytes).digest('hex')}
async function preserveOriginal(directory, capture) {
  await fs.mkdir(directory, {recursive: true});
  const bytes = await storedBytes(capture.receipt.id);
  const custody = {file: 'scope-capture.json', bytes: bytes.length, sha256: hash(bytes), receiptID: capture.receipt.id, copiedAt: new Date().toISOString(), source: 'read-only actual query-container capture store'};
  await fs.writeFile(path.join(directory, custody.file), bytes);
  await fs.writeFile(path.join(directory, 'capture-custody.json'), JSON.stringify(custody, null, 2) + '\n');
  return custody;
}
async function observeStages(page) {
  await page.addInitScript(() => {
    window.__patternTimeStages = [];
    window.addEventListener('goo:question-state', event => {
      if (event.detail.action === 'stage') window.__patternTimeStages.push(structuredClone(event.detail.requested));
    });
  });
}
async function runCurrent(page, records) {
  const pending = page.waitForResponse(response => isResponse('explore', response));
  await page.locator('#exploreRun').click();
  const result = await json(await pending);
  assertResult(result, records);
  return result;
}
function assertPair(pair, records) {
  const baseline = pair.request.baseline, selection = pair.request.selection;
  expect(baseline.question.groupBy).toEqual([patternField]);
  expect(selection.question.groupBy).toEqual([patternField]);
  for (const side of [baseline, selection]) {
    const expected = groupOracle(records, side.question);
    expect(side.count).toBe(String(members(records, side.question).length));
    expect(side.groups).toHaveLength(expected.size);
    for (const group of side.groups) expect(group.count).toBe(String(expected.get(group.g0)));
  }
  expect(pair.response.baselineCount).toBe(baseline.count);
  expect(pair.response.selectionCount).toBe(selection.count);
  if (baseline.state !== 'changed_between_reads' && selection.state !== 'changed_between_reads' && baseline.count !== '0' && selection.count !== '0') {
    const expected = comparisonOracle(records, baseline.question, selection.question);
    expect(pair.response.rows).toHaveLength(expected.length);
    for (const row of pair.response.rows) {
      const own = expected.find(item => item.id === row.value);
      expect(own).toBeTruthy();
      expect(row.baselineCount).toBe(own.baselineCount);
      expect(row.selectionCount).toBe(own.selectionCount);
    }
  }
  return comparisonOracle(records, baseline.question, selection.question);
}

const section = page => page.locator('.pattern-time-comparison:visible');
async function openChoice(page) {
  const choice = page.locator('.pattern-time-choice:visible');
  if (!await choice.evaluate(element => element.open)) await choice.locator(':scope>summary').click();
  return choice;
}
async function runPair(page, state, records) {
  const before = state.sourceResponses.length;
  const pending = page.waitForResponse(response => isResponse('compare', response));
  await page.locator('.population-run:visible').click();
  const response = await pending;
  const pair = {request: response.request().postDataJSON(), response: await json(response)};
  assertPair(pair, records);
  await expect(section(page)).toHaveAttribute('data-state', 'known');
  await expect.poll(() => state.sourceResponses.length).toBe(before + 2);
  const reads = state.sourceResponses.slice(before);
  pair.baseline = reads.find(read => read.response.question.fromMs === pair.request.baseline.question.fromMs && read.response.question.toMs === pair.request.baseline.question.toMs).response;
  pair.selection = reads.find(read => read.response.question.fromMs === pair.request.selection.question.fromMs && read.response.question.toMs === pair.request.selection.question.toMs).response;
  for (const result of [pair.baseline, pair.selection]) assertResult(result, records);
  return pair;
}
async function exactView(page, pair, records, metric) {
  const host = section(page);
  await host.getByRole('combobox', {name: 'Pattern comparison measure', exact: true}).selectOption(metric);
  const model = await page.evaluate(({request, response, metric}) => window.GooLogPatternComparison.model(request, response, {metric}), {...pair, metric});
  expect(model).toBeTruthy();
  const expected = comparisonOracle(records, pair.request.baseline.question, pair.request.selection.question);
  expect(model.rows).toHaveLength(expected.length);
  const rows = await host.locator('tbody>tr[data-pattern]').evaluateAll(elements => elements.map(element => ({...element.dataset, label: element.querySelector('th').title, cells: [...element.querySelectorAll('.comparison-bar-cell')].map(cell => {
    const fill = cell.querySelector('.comparison-bar-fill'), track = cell.querySelector('.comparison-bar-track'), text = cell.querySelector('.comparison-value');
    return {value: cell.dataset.value, max: cell.dataset.scaleMax, side: cell.dataset.side, exact: text.title, fillPercent: fill?.style.width ?? null, fillWidth: fill?.getBoundingClientRect().width ?? null, trackWidth: track.getBoundingClientRect().width};
  })})));
  expect(rows).toHaveLength(expected.length);
  for (const own of expected) {
    const mapped = {baselineCount: own.baselineCount, currentCount: own.selectionCount, baselineRate: own.baselineRate, currentRate: own.selectionRate, baselineShare: own.baselineShare, currentShare: own.selectionShare, countDifference: own.countDifference, rateDifference: own.rateDifference, shareDifference: own.shareDifferencePP, presence: own.observation === 'newly-observed' ? 'not-observed-in-baseline' : own.observation === 'absent' ? 'not-observed-in-current' : 'observed-in-both'};
    expect(model.rows.find(row => row.id === own.id)).toMatchObject(mapped);
    const row = rows.find(row => row.pattern === own.id);
    for (const [key, value] of Object.entries(mapped)) expect(row[key]).toBe(value === null ? 'unavailable' : value);
    expect(row.cells).toHaveLength(2);
    expect(row.cells[0].max).toBe(row.cells[1].max);
    const unit = metric === 'count' ? 'logs' : metric === 'rate' ? 'logs/s' : '%';
    const baseline = mapped['baseline' + (metric === 'count' ? 'Count' : metric === 'rate' ? 'Rate' : 'Share')];
    const current = mapped['current' + (metric === 'count' ? 'Count' : metric === 'rate' ? 'Rate' : 'Share')];
    expect(row.cells[0].exact).toBe(baseline === null ? 'Unavailable' : baseline + ' ' + unit + ' (exact)');
    expect(row.cells[1].exact).toBe(current === null ? 'Unavailable' : current + ' ' + unit + ' (exact)');
  }
  return {metric, model, rows};
}
function commonGeometry(view) {
  const row = view.rows.find(item => item.pattern === pattern(bodies.common(0)));
  return {baseline: row.cells[0], current: row.cells[1]};
}
async function stagedQuestions(population) {
  const scope = population.locator('details').filter({has: population.page().getByText('Signal selectors and independent read scopes', {exact: true})});
  return JSON.parse(await scope.locator(':scope>pre').textContent()).queries;
}

test('183 real log patterns compare previous and reread saved windows with exact rates and shares', async ({page, request}) => {
  test.setTimeout(300000);
  page.setDefaultTimeout(15000);
  page.setDefaultNavigationTimeout(45000);
  const f = await fixture(request), state = watch(page), directory = path.resolve('artifacts/log-pattern-time', f.environment);
  const custody = await preserveOriginal(directory, f.scopeCapture);
  await observeStages(page);
  await page.goto(base + '/' + questionLink(f.current));
  expect(state.sourceOperations).toBe(0); state.event('current-stage');
  const current = await runCurrent(page, f.records);
  expect(current.count).toBe('400'); state.event('current-run');
  let choice = await openChoice(page);
  await expect(choice.getByRole('combobox', {name: 'Pattern baseline scope', exact: true})).toHaveValue('previous');
  await choice.getByRole('button', {name: 'Stage pattern comparison', exact: true}).click();
  expect(state.sourceOperations).toBe(1); state.event('previous-stage');
  const previous = await runPair(page, state, f.records);
  expect(previous.request.baseline.question).toEqual(f.previous);
  expect(previous.request.selection.question).toEqual(f.current);
  state.event('previous-run');
  const previousViews = [];
  for (const metric of ['rate', 'count', 'share']) previousViews.push(await exactView(page, previous, f.records, metric));
  expect(state.sourceOperations).toBe(3); expect(state.protocols).toBe(1); state.event('previous-local-metrics');
  const absentID = pattern(bodies.baselineOnly(0));
  const absent = section(page).locator('tr[data-pattern="' + absentID + '"]');
  await expect(absent).toHaveAttribute('data-presence', 'not-observed-in-current');
  await absent.getByText('Inspect a side', {exact: true}).click();
  await absent.getByRole('link', {name: 'Open baseline pattern logs', exact: true}).click();
  await expect(page.locator('#exploreStatus')).toContainText('No evidence loaded');
  const sideQuestion = await page.evaluate(() => window.__patternTimeStages.at(-1));
  expect([sideQuestion.fromMs, sideQuestion.toMs]).toEqual([f.previous.fromMs, f.previous.toMs]);
  expect(members(f.records, sideQuestion).map(record => record.id).sort()).toEqual(members(f.records, f.previous).filter(record => pattern(record.body) === absentID).map(record => record.id).sort());
  expect(state.sourceOperations).toBe(3); state.event('side-stage');
  const sideResult = await runCurrent(page, f.records);
  expect(sideResult.count).toBe('160'); state.event('side-run');
  await page.getByRole('button', {name: 'Back to question', exact: true}).click();
  await expect(section(page)).toBeVisible();
  expect(state.sourceOperations).toBe(4); state.event('side-back');
  await page.getByRole('button', {name: 'Back to question', exact: true}).click();
  await expect(page.locator('.log-pattern-results:visible')).toBeVisible();
  state.event('parent-return');

  const late = lateRecords(f);
  await emit(request, late);
  let lateReadyProbe;
  await expect.poll(async () => {lateReadyProbe = await json(await request.post(base + '/api/explore', {data: f.savedQuestion})); return lateReadyProbe.count}, {timeout: 45000, intervals: [500]}).toBe('412');
  const allRecords = [...f.records, ...late];
  assertResult(lateReadyProbe, allRecords);
  choice = await openChoice(page);
  await choice.getByRole('combobox', {name: 'Pattern baseline scope', exact: true}).selectOption('saved');
  await choice.getByRole('textbox', {name: 'Saved pattern baseline capture', exact: true}).fill(f.scopeCapture.receipt.id);
  expect(state.sourceOperations).toBe(4); state.event('saved-scope-edit');
  const stored = page.waitForResponse(response => response.request().method() === 'GET' && new URL(response.url()).pathname === '/api/captures/' + f.scopeCapture.receipt.id);
  await choice.getByRole('button', {name: 'Load saved baseline scope', exact: true}).click();
  const loadedScope = await json(await stored);
  expect(loadedScope).toEqual(f.scopeCapture);
  await expect(choice).toContainText('saved counts are not reused');
  expect(state.sourceOperations).toBe(4); state.event('saved-scope-load');
  await choice.getByRole('button', {name: 'Stage pattern comparison', exact: true}).click();
  expect(state.sourceOperations).toBe(4); state.event('saved-stage');
  const savedPair = await runPair(page, state, allRecords);
  expect(savedPair.request.baseline.count).toBe('412');
  expect(savedPair.request.selection.count).toBe('400');
  expect(savedPair.request.baseline.question.where).toEqual(f.savedQuestion.where);
  expect([savedPair.request.baseline.question.fromMs, savedPair.request.baseline.question.toMs]).toEqual([f.savedQuestion.fromMs, f.savedQuestion.toMs]);
  for (const result of [savedPair.baseline, savedPair.selection]) expect(Date.parse(result.capturedAt)).toBeGreaterThanOrEqual(Date.parse(lateReadyProbe.capturedAt));
  state.event('saved-run');
  const savedViews = {};
  for (const metric of ['rate', 'count', 'share']) savedViews[metric] = await exactView(page, savedPair, allRecords, metric);
  const geometry = {rate: commonGeometry(savedViews.rate), count: commonGeometry(savedViews.count)};
  expect(geometry.rate.baseline.value).toBe('8/3'); expect(geometry.rate.current.value).toBe('8/3');
  expect(geometry.rate.baseline.fillPercent).toBe(geometry.rate.current.fillPercent);
  expect(Math.abs(geometry.rate.baseline.fillWidth - geometry.rate.current.fillWidth)).toBeLessThan(1);
  expect(geometry.count.baseline.value).toBe('320'); expect(geometry.count.current.value).toBe('160');
  expect(parseFloat(geometry.count.baseline.fillPercent)).toBe(2 * parseFloat(geometry.count.current.fillPercent));
  expect(state.sourceOperations).toBe(6); expect(state.protocols).toBe(2); state.event('saved-local-metrics');
  await exactView(page, savedPair, allRecords, 'rate');
  await page.setViewportSize({width: 1440, height: 1900});
  await section(page).evaluate(element => {element.scrollIntoView({block: 'start'}); window.scrollBy(0, -120)});
  await section(page).screenshot({path: 'docs/journeys/screenshots/183a-unequal-window-pattern-rates.png'});
  await page.setViewportSize({width: 390, height: 2400});
  const mobile = await section(page).locator('tbody>tr[data-pattern]').evaluateAll(rows => rows.map(row => ({id: row.dataset.pattern, cells: [...row.querySelectorAll('.comparison-bar-cell')].map(cell => {
    const box = cell.getBoundingClientRect(), text = cell.querySelector('.comparison-value');
    return {left: box.left, right: box.right, viewport: innerWidth, exact: text.title, noTextClip: text.scrollWidth <= text.clientWidth};
  })})));
  for (const row of mobile) for (const cell of row.cells) {expect(cell.left).toBeGreaterThanOrEqual(0); expect(cell.right).toBeLessThanOrEqual(cell.viewport); expect(cell.noTextClip).toBe(true)}
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await section(page).evaluate(element => {element.scrollIntoView({block: 'start'}); window.scrollBy(0, -120)});
  await section(page).screenshot({path: 'docs/journeys/screenshots/183b-mobile-pattern-time-comparison.png'});
  await page.setViewportSize({width: 1440, height: 1100});

  const population = page.locator('.population-workspace:visible').last();
  const saveView = population.locator('.population-save');
  if (!await saveView.evaluate(element => element.open)) await saveView.locator(':scope>summary').click();
  await saveView.getByRole('textbox', {name: 'Exploration view title', exact: true}).fill('Fixed pattern time comparison');
  const configured = page.waitForResponse(response => isResponse('exploration-views', response));
  await saveView.getByRole('button', {name: 'Save exploration view', exact: true}).click();
  const reusable = await json(await configured);
  expect(state.sourceOperations).toBe(6); state.event('view-save');
  await population.getByRole('button', {name: 'Capture comparison (fresh reads)', exact: true}).click();
  await page.locator('#captureDraftTitle').fill('Freshly captured pattern comparison');
  const captured = page.waitForResponse(response => isResponse('captures', response));
  await page.locator('#captureDraftConfirm').click();
  const saved = await json(await captured);
  await expect(page.locator('#captureGeneralID')).toHaveValue(saved.receipt.id);
  const capturePair = {request: saved.comparison.request, response: saved.comparison.result};
  assertPair(capturePair, allRecords);
  expect(capturePair.request.baseline.count).toBe('412'); expect(capturePair.request.selection.count).toBe('400');
  state.event('capture-fresh');
  const captureSources = {};
  for (const [side, id] of [['baseline', saved.comparison.baselineId], ['selection', saved.comparison.selectionId]]) {
    captureSources[side] = await json(await request.get(base + '/api/captures/' + id));
    assertResult(captureSources[side].explore, allRecords);
    expect(Date.parse(captureSources[side].explore.capturedAt)).toBeGreaterThanOrEqual(Date.parse(savedPair[side].capturedAt));
  }
  const beforeReopen = state.sourceOperations;
  await page.reload();
  await expect(section(page)).toHaveAttribute('data-state', 'known');
  const reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id));
  expect(reopened).toEqual(saved);
  const frozenView = await exactView(page, capturePair, allRecords, 'rate');
  expect(state.sourceOperations).toBe(beforeReopen); state.event('capture-reopen');
  await page.goto(base + '/#exploration-view=' + reusable.id);
  const reopenedPopulation = page.locator('.population-workspace:visible').last();
  await expect(reopenedPopulation).toHaveAttribute('data-phase', 'staged');
  expect(state.sourceOperations).toBe(beforeReopen); state.event('view-reopen');
  const reused = await runPair(page, state, allRecords);
  expect(reused.request.baseline.question).toEqual(savedPair.request.baseline.question);
  expect(reused.request.selection.question).toEqual(savedPair.request.selection.question);
  state.event('view-run');
  const scopeCaptureAfter = await json(await request.get(base + '/api/captures/' + f.scopeCapture.receipt.id));
  expect(scopeCaptureAfter).toEqual(f.scopeCapture);
  const after = await storedBytes(f.scopeCapture.receipt.id);
  expect(hash(after)).toBe(custody.sha256);
  expect(after.equals(await fs.readFile(path.join(directory, custody.file)))).toBe(true);
  await fs.writeFile(path.join(directory, 'scope-capture.after.json'), after);
  expect(state.errors).toEqual([]); state.event('original-unchanged');
  await fs.writeFile(path.join(directory, 'ledger.json'), JSON.stringify({...f, custody, current, previous, previousViews, sideOpen: {side: 'baseline', id: absentID, parent: previous.request.baseline.question, question: sideQuestion, result: sideResult}, lateRecords: late, lateReadyProbe, loadedScope, savedPair, savedViews, geometry, mobile, reusable, saved, captureSources, reopened, frozenView, reused, scopeCaptureAfter, events: state.events, requests: state.requests, sourceResponses: state.sourceResponses}, null, 2));
});

test('183 successful empty baseline has zero rate and undefined share', async ({page, request}) => {
  test.setTimeout(180000);
  page.setDefaultTimeout(15000); page.setDefaultNavigationTimeout(45000);
  const f = await fixture(request), state = watch(page);
  await page.goto(base + '/' + questionLink(f.previous));
  const current = await runCurrent(page, f.records); state.event('current-run');
  const choice = await openChoice(page);
  await choice.getByRole('button', {name: 'Stage pattern comparison', exact: true}).click();
  expect(state.sourceOperations).toBe(1); state.event('empty-stage');
  const pair = await runPair(page, state, f.records);
  expect(pair.request.baseline.count).toBe('0'); expect(pair.request.baseline.state).toBe('no_data');
  expect(pair.request.selection.count).toBe('400');
  expect(pair.baseline.rows).toEqual([]); expect(pair.baseline.groups).toEqual([]);
  const views = {};
  for (const metric of ['rate', 'count', 'share']) views[metric] = await exactView(page, pair, f.records, metric);
  for (const row of views.rate.model.rows) {expect(row.baselineRate).toBe('0'); expect(row.baselineShare).toBeNull(); expect(row.shareDifference).toBeNull(); expect(row.presence).toBe('not-observed-in-baseline')}
  expect(state.sourceOperations).toBe(3); state.event('empty-run');
  await section(page).getByRole('combobox', {name: 'Pattern comparison measure', exact: true}).selectOption('share');
  await expect(section(page)).toContainText('Share unavailable');
  await page.setViewportSize({width: 1440, height: 1800});
  await section(page).evaluate(element => {element.scrollIntoView({block: 'start'}); window.scrollBy(0, -120)});
  await section(page).screenshot({path: 'docs/journeys/screenshots/183c-empty-baseline-share-unavailable.png'});
  expect(state.errors).toEqual([]);
  const directory = path.resolve('artifacts/log-pattern-time', f.environment);
  await fs.mkdir(directory, {recursive: true});
  await fs.writeFile(path.join(directory, 'empty.json'), JSON.stringify({...f, current, pair, views, events: state.events, requests: state.requests, sourceResponses: state.sourceResponses}, null, 2));
});

test('183 real baseline delivery failure remains unavailable and held responses lose ownership', async ({page, request}) => {
  test.setTimeout(240000);
  page.setDefaultTimeout(15000); page.setDefaultNavigationTimeout(45000);
  const f = await fixture(request), state = watch(page), observations = [];
  await observeStages(page);
  await page.goto(base + '/' + questionLink(f.current));
  const current = await runCurrent(page, f.records); state.event('current-run');
  let choice = await openChoice(page);
  await choice.getByRole('button', {name: 'Stage pattern comparison', exact: true}).click();
  state.event('failure-stage');
  let arrived, release, heldResponse;
  const ready = new Promise(resolve => {arrived = resolve}), gate = new Promise(resolve => {release = resolve});
  await page.route('**/api/explore', async route => {
    const question = route.request().postDataJSON();
    if (question.fromMs !== f.previous.fromMs || question.toMs !== f.previous.toMs) {await route.continue(); return}
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    heldResponse = {request: question, status: response.status(), response: await response.json(), transportFault: 'Abort browser delivery after actual upstream success'};
    arrived(); await gate; await route.abort('failed');
  });
  await page.locator('.population-run:visible').click();
  await ready;
  assertResult(heldResponse.response, f.records);
  release(); await page.unrouteAll({behavior: 'wait'});
  await expect(page.locator('.population-workspace:visible').last()).not.toHaveAttribute('data-phase', 'running');
  expect(state.sourceOperations).toBe(3); expect(state.protocols).toBe(0);
  await expect(page.locator('.population-workspace:visible').last()).toContainText(/unavailable|failed/i);
  expect(await section(page).count()).toBe(0);
  expect(await page.getByText(/Not observed in (baseline|current) read/).count()).toBe(0);
  observations.push({action: 'failed-baseline', comparisonSections: await section(page).count(), absentLabels: await page.getByText(/Not observed in (baseline|current) read/).count(), sourceOperations: state.sourceOperations, protocols: state.protocols});
  state.event('failure-delivered');

  // A subsequent real response is held across navigation. Delivery must not
  // publish a hidden result or make a pure comparison request for that old owner.
  await page.getByRole('button', {name: 'Back to question', exact: true}).click();
  choice = await openChoice(page);
  await choice.getByRole('button', {name: 'Stage pattern comparison', exact: true}).click();
  let heldArrive, heldRelease, late;
  const lateReady = new Promise(resolve => {heldArrive = resolve}), lateGate = new Promise(resolve => {heldRelease = resolve});
  await page.route('**/api/explore', async route => {
    const question = route.request().postDataJSON();
    if (question.fromMs !== f.previous.fromMs || question.toMs !== f.previous.toMs) {await route.continue(); return}
    const response = await route.fetch();
    expect(response.ok()).toBe(true); late = {request: question, response: await response.json()};
    heldArrive(); await lateGate;
    try {await route.fulfill({response})} catch { /* Actual owner canceled its request. */ }
  });
  await page.locator('.population-run:visible').click(); await lateReady;
  state.event('held-run');
  await page.goto(base + '/#definitions');
  heldRelease(); await page.unrouteAll({behavior: 'wait'});
  await page.goto(base + '/#explore');
  expect(state.sourceOperations).toBe(5); expect(state.protocols).toBe(0);
  expect(await section(page).count()).toBe(0);
  observations.push({action: 'held-return', comparisonSections: await section(page).count(), sourceOperations: state.sourceOperations, protocols: state.protocols});
  state.event('held-return');
  const recovered = await runPair(page, state, f.records);
  expect(state.sourceOperations).toBe(7); state.event('recovered-run');
  const row = section(page).locator('tr[data-pattern="' + pattern(bodies.currentOnly(0)) + '"]');
  await row.getByText('Inspect a side', {exact: true}).click();
  await row.getByRole('link', {name: 'Open current pattern logs', exact: true}).evaluate(element => {window.__oldPatternTimeSide = element});
  await page.goto(base + '/#definitions');
  const before = await page.evaluate(() => ({url: location.href, stages: window.__patternTimeStages.length, connected: window.__oldPatternTimeSide.isConnected}));
  await page.evaluate(() => window.__oldPatternTimeSide.dispatchEvent(new MouseEvent('click', {bubbles: true, cancelable: true, view: window})));
  const after = await page.evaluate(() => ({url: location.href, stages: window.__patternTimeStages.length}));
  expect(before.connected).toBe(true); expect(after).toEqual({url: before.url, stages: before.stages});
  expect(state.sourceOperations).toBe(7);
  observations.push({action: 'stale-side', before, after, sourceOperations: state.sourceOperations}); state.event('stale-side');
  expect(state.errors).toEqual([]);
  const directory = path.resolve('artifacts/log-pattern-time', f.environment);
  await fs.mkdir(directory, {recursive: true});
  await fs.writeFile(path.join(directory, 'ownership.json'), JSON.stringify({...f, current, heldResponse, late, recovered, observations, events: state.events, requests: state.requests, sourceResponses: state.sourceResponses, failedRequests: state.failedRequests}, null, 2));
});

test('183 Bombadil compares real log-pattern windows and returns from exact side inspection', async ({request}) => {
  test.skip(process.env.GOO_PATTERN_TIME_BOMBADIL !== '1', 'Optional serialized ordinary OTel campaign.');
  test.setTimeout(180000);
  const f = await fixture(request), directory = path.resolve('artifacts/log-pattern-time', f.environment, 'bombadil');
  await fs.mkdir(directory, {recursive: true});
  await fs.writeFile(path.join(directory, 'fixture.json'), JSON.stringify(f, null, 2));
  const actor = await startBombadil({dir: directory, url: base + '/' + questionLink(f.current), specification: 'verification/bombadil/log-pattern-time-comparisons.ts', seconds: 55});
  try {await actor.finish()} finally {await actor.stop()}
});
