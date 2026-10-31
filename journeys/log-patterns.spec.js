'use strict';

const {test, expect} = require('@playwright/test');
const {randomBytes} = require('node:crypto');
const fs = require('node:fs/promises');
const path = require('node:path');
const {questionLink} = require('../ui/workspace.mjs');
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
const attributes = values => Object.entries(values).map(([key, stringValue]) => ({key, value: {stringValue}}));
async function emit(request, records) {
  const resourceLogs = records.map(record => ({
    resource: {attributes: attributes({'service.name': record.service, 'deployment.environment.name': record.environment, ...record.resource})},
    scopeLogs: [{scope: {name: 'ordinary.log.pattern.worker'}, logRecords: [{
      timeUnixNano: String(BigInt(record.timeMs) * 1000000n), traceId: record.traceId, spanId: record.id,
      severityText: record.severity, body: {stringValue: record.body}, attributes: attributes(record.attributes)
    }]}]
  }));
  expect((await request.post('http://127.0.0.1:4318/v1/logs', {data: {resourceLogs}})).ok()).toBe(true);
}
function body(index) {
  if (index % 3 === 0 && index < 360) return 'request id=' + (1000 + index) + ' completed in ' + (index * 3) + 'ms';
  const corpus = [
    'request id=00042 completed in 007ms', 'request id=42 completed in 7ms',
    'temperature=-12.50C', 'temperature=+003.004C', 'temperature=12.50C',
    'peer=10.12.0.9 port=443', 'peer=192.168.12.34 port=8443',
    'task=550e8400-e29b-41d4-a716-446655440000',
    'task=550e8400-e29b-41d4-a716-446655440001',
    '１２３ are not ASCII digits', 'échec 123 ✓', 'échec 999 ✓',
    'café123', 'cafe\u0301123', 'replacement \ufffd 123',
    'n', '123', '000007', 'literal 0xCAFE', 'literal 0xBEEF',
    '<img src=x onerror="window.__patternInjected=true">12',
    'line one\nline two\tvalue=42\r\n', 'nul\u0000value=123',
    '', 'x'.repeat(256), 'x'.repeat(257), 'é'.repeat(128), 'é'.repeat(129),
    'new connection observed', 'queue draining', 'unexpected punctuation +.-/:',
    'value=1 value=23', 'value=12 value=3'
  ];
  if (index === 399) return 'Cache was refreshed ✓';
  return corpus[index % corpus.length];
}
async function fixture(request, {highCardinality = false} = {}) {
  const environment = 'log-patterns-' + randomBytes(5).toString('hex');
  const toMs = Date.now(), fromMs = toMs - 60000;
  const alphabet = index => {
    let text = '';
    do {text = String.fromCharCode(97 + index % 26) + text; index = Math.floor(index / 26) - 1} while (index >= 0);
    return text;
  };
  const count = highCardinality ? 1100 : 400;
  const records = Array.from({length: count}, (_, index) => ({
    environment, signal: 'logs', id: randomBytes(8).toString('hex'), traceId: randomBytes(16).toString('hex'),
    timeMs: fromMs + 1000 + index * (highCardinality ? 35 : 100), service: index % 3 === 0 ? 'checkout-api' : 'queue-worker', severity: 'INFO',
    body: highCardinality ? 'independent alphabetic shape ' + alphabet(index) : body(index),
    attributes: {team: 'core', 'work.sequence': String(index)}, resource: {'service.namespace': 'commerce'}
  }));
  if (!highCardinality) for (let index = 0; index < 30; index++) records.push({
    ...records[index], id: randomBytes(8).toString('hex'),
    service: index % 3 === 0 ? 'unrelated-worker' : 'checkout-api',
    attributes: {team: index % 3 === 1 ? 'other' : 'core'},
    resource: {'service.namespace': index % 3 === 2 ? 'other' : 'commerce'}
  });
  await emit(request, records);
  const selector = {signal: 'logs', environment, where: {all: [
    {field: {kind: 'resource', key: 'service.namespace'}, op: 'eq', value: 'commerce'},
    {any: [
      {field: {kind: 'column', key: 'service'}, op: 'eq', value: 'checkout-api'},
      {field: {kind: 'column', key: 'service'}, op: 'eq', value: 'queue-worker'}
    ]},
    {field: {kind: 'attribute', key: 'team'}, op: 'eq', value: 'core'}
  ]}};
  const question = {...selector, fromMs, toMs, limit: 100, offset: 0};
  let readyProbe;
  await expect.poll(async () => {readyProbe = await json(await request.post(base + '/api/explore', {data: question})); return readyProbe.count}, {timeout: 45000, intervals: [500]}).toBe(String(count));
  const view = await json(await request.post(base + '/api/activity-views', {data: {version: 1, title: 'Log pattern investigation', selector, columns: ['utc', 'service', 'summary']}}));
  return {version: 1, environment, records, selector, question, view, highCardinality, readyProbe};
}
function watch(page) {
  const state = {reads: 0, streams: 0, plans: 0, protocols: 0, writes: 0, events: [], requests: [], errors: []};
  page.on('pageerror', error => state.errors.push(error.message));
  page.on('request', request => {
    if (request.method() !== 'POST') return;
    const endpoint = new URL(request.url()).pathname, payload = request.postDataJSON();
    let effect;
    if (endpoint === '/api/tail') {state.streams++; effect = 'stream'}
    else if (['/api/explore', '/api/captures', '/api/measurements', '/api/discovery', '/api/inventory'].includes(endpoint)) {state.reads++; effect = 'source'}
    else if (endpoint === '/api/compare') {state.protocols++; effect = 'comparison'}
    else if (['/api/population/plan', '/api/population/explain/plan'].includes(endpoint)) {state.plans++; effect = 'plan'}
    else if (['/api/telemetry-packages', '/api/exploration-views'].includes(endpoint)) {state.writes++; effect = 'publication'}
    if (effect) state.requests.push({endpoint, payload, effect});
  });
  state.event = action => state.events.push({action, ...Object.fromEntries(['reads', 'streams', 'plans', 'protocols', 'writes'].map(key => [key, state[key]]))});
  return state;
}

async function observeStages(page) {
  await page.addInitScript(() => {
    window.__patternStages = [];
    window.addEventListener('goo:question-state', event => {
      if (event.detail.action === 'stage') window.__patternStages.push(structuredClone(event.detail.requested));
    });
  });
}
async function staged(page) {
  await expect(page.locator('#exploreStatus')).toContainText(/No evidence loaded|Analysis staged\. Run analysis reads this fixed window/);
  const question = await page.evaluate(() => window.__patternStages.at(-1));
  expect(question).toBeTruthy();
  return question;
}
async function run(page, records) {
  const pending = page.waitForResponse(response => isResponse('explore', response));
  await page.locator('#exploreRun').click();
  const result = await json(await pending);
  assertResult(result, records);
  return result;
}
function assertPatternRows(rows, result) {
  expect(rows).toHaveLength(result.groups.length);
  const expected = new Map(result.groups.map(group => [group.g0, group.count]));
  for (const row of rows) {
    expect(row.count).toBe(expected.get(row.id));
    expect(row.share).toBe(row.count + '/' + result.count + ' of all matching records (exact)');
    if (row.id === 'lp1:e') expect(row.label).toBe('Empty body · unclassified');
    if (row.id === 'lp1:o') expect(row.label).toBe('Body exceeds 256-byte bound · unclassified');
  }
}
async function patternRows(panel) {
  return panel.locator('tbody>tr[data-pattern]').evaluateAll(rows => rows.map(row => ({
    id: row.dataset.pattern, state: row.dataset.classification,
    label: row.querySelector('th').textContent,
    count: row.querySelector('td').textContent.split(' · ')[0],
    share: row.querySelector('td').title
  })));
}

// Final source/owner selectors are frozen with the deployed integration before execution.
test('182 real log patterns retain full-population counts, exact trends and immutable evidence', async ({page, request}) => {
  test.setTimeout(300000);
  page.setDefaultTimeout(15000);
  page.setDefaultNavigationTimeout(45000);
  const f = await fixture(request), state = watch(page);
  await observeStages(page);
  await page.goto(base + '/#activity-view=' + f.view.id);
  const activity = page.locator('.stream-entry'), tail = activity.locator('.live-tail');
  await activity.getByRole('button', {name: 'Start stream', exact: true}).click();
  await expect(tail).toHaveAttribute('data-last-to', /^\d+$/);
  await expect(tail.locator('tbody>tr')).not.toHaveCount(0);
  state.event('activity-start');
  await activity.locator('.entry-investigate-options>summary').click();
  await activity.getByRole('button', {name: 'Log patterns', exact: true}).click();
  await expect(tail).toHaveAttribute('data-state', 'paused');
  const toMs = Number(await tail.getAttribute('data-last-to'));
  const applied = {selector: f.selector, readWindow: {fromMs: toMs - 60000, toMs}};
  const original = await staged(page);
  expect(original).toEqual({...f.selector, ...applied.readWindow, groupBy: [patternField], bucketMs: 1000, limit: 100, offset: 0});
  expect(state.reads).toBe(0);
  state.event('patterns-stage');
  const result = await run(page, f.records);
  expect(BigInt(result.count)).toBeGreaterThan(100n);
  expect(result.rows).toHaveLength(100);
  expect(result.groups.some(group => group.g0 === 'lp1:e')).toBe(true);
  expect(result.groups.some(group => group.g0 === 'lp1:o')).toBe(true);
  state.event('patterns-run');
  const panel = page.locator('.log-pattern-results:visible');
  await expect(panel).toContainText('inferred text shapes, not issues or causes');
  const initialRows = await patternRows(panel);
  assertPatternRows(initialRows, result);
  expect(initialRows.map(row => BigInt(row.count))).toEqual(initialRows.map(row => BigInt(row.count)).sort((a, b) => a > b ? -1 : a < b ? 1 : 0));
  const beforeSort = state.reads;
  await panel.getByRole('combobox', {name: 'Sort log patterns', exact: true}).selectOption('least');
  const rareRows = await patternRows(panel);
  assertPatternRows(rareRows, result);
  expect(rareRows[0].count).toBe('1');
  await panel.getByRole('combobox', {name: 'Sort log patterns', exact: true}).selectOption('most');
  expect(state.reads).toBe(beforeSort);
  state.event('sort-local');
  const frequentID = pattern('request id=1000 completed in 0ms');
  const frequent = panel.locator('tr[data-pattern="' + frequentID + '"]');
  const frequentCount = BigInt(result.groups.find(group => group.g0 === frequentID).count);
  expect(frequentCount).toBeGreaterThan(1n);
  expect(initialRows[0].id).toBe(frequentID);
  await expect(frequent.locator('th')).toHaveText('request id=⟪digits⟫ completed in ⟪digits⟫ms');
  expect(await page.evaluate(() => window.__patternInjected === true)).toBe(false);

  // A decoded label is escaped text; the independently inspected source body is unchanged.
  const htmlIndex = result.rows.findIndex(row => row.Body.includes('<img src=x'));
  expect(htmlIndex).toBeGreaterThanOrEqual(0);
  await page.locator('#resultRecords').click();
  await page.locator('#exploreRows .record-row').nth(htmlIndex).getByRole('button', {name: 'Inspect record', exact: true}).click();
  const inspector = page.locator('#exploreRows .record-detail:visible');
  const rawDetail = JSON.parse(await inspector.locator('.raw-record>pre').textContent());
  expect(rawDetail.Body).toBe(result.rows[htmlIndex].Body);
  expect(await inspector.locator('img').count()).toBe(0);
  expect(state.reads).toBe(1);
  await page.locator('#exploreRows').getByRole('button', {name: 'Close selected item details', exact: true}).click();
  await page.locator('#resultDistribution').click();
  state.event('raw-inspection');
  await page.setViewportSize({width: 1440, height: 1800});
  await panel.evaluate(element => {element.scrollIntoView({block: 'start'}); window.scrollBy(0, -120)});
  await panel.screenshot({path: 'docs/journeys/screenshots/182a-full-population-log-patterns.png'});

  await frequent.getByRole('link', {name: 'Open pattern logs', exact: true}).click();
  const selectedStage = await staged(page);
  expect(selectedStage.signal).toBe('logs');
  expect(selectedStage.environment).toBe(original.environment);
  expect([selectedStage.fromMs, selectedStage.toMs]).toEqual([original.fromMs, original.toMs]);
  expect(selectedStage.groupBy).toBeUndefined();
  expect(members(f.records, selectedStage).map(record => record.id).sort()).toEqual(members(f.records, original).filter(record => pattern(record.body) === frequentID).map(record => record.id).sort());
  expect(state.reads).toBe(1);
  state.event('pattern-trend-stage');
  const selected = await run(page, f.records);
  state.event('pattern-trend-run');
  await page.getByRole('button', {name: 'Back to question', exact: true}).click();
  expect(state.reads).toBe(2);
  state.event('pattern-back');
  await frequent.getByRole('button', {name: 'Compare services with rest', exact: true}).click();
  expect(state.reads).toBe(2);
  state.event('comparison-stage');
  const comparisonResponse = page.waitForResponse(response => isResponse('compare', response));
  await page.locator('.population-run:visible').click();
  const cr = await comparisonResponse, comparison = {request: cr.request().postDataJSON(), response: await json(cr)};
  const selectionQuery = comparison.request.selection.question, restQuery = comparison.request.baseline.question;
  const selectedFacts = members(f.records, selectionQuery), restFacts = members(f.records, restQuery), parentFacts = members(f.records, original);
  expect(selectedFacts.map(record => record.id).sort()).toEqual(parentFacts.filter(record => pattern(record.body) === frequentID).map(record => record.id).sort());
  expect(new Set([...selectedFacts, ...restFacts].map(record => record.id)).size).toBe(parentFacts.length);
  expect(selectedFacts.length + restFacts.length).toBe(parentFacts.length);
  expect(restFacts.some(record => pattern(record.body) === 'lp1:e')).toBe(true);
  expect(restFacts.some(record => pattern(record.body) === 'lp1:o')).toBe(true);
  expect(comparison.request.selection.count).toBe(String(selectedFacts.length));
  expect(comparison.request.baseline.count).toBe(String(restFacts.length));
  const serviceGroups = facts => {
    const counts = new Map();
    for (const record of facts) counts.set(record.service, (counts.get(record.service) || 0) + 1);
    return [...counts].map(([g0, count]) => ({g0, count: String(count)})).sort((a, b) => a.g0.localeCompare(b.g0));
  };
  for (const [cohort, facts] of [[comparison.request.selection, selectedFacts], [comparison.request.baseline, restFacts]]) {
    expect(cohort.question.groupBy).toEqual([{kind: 'column', key: 'service'}]);
    expect(cohort.groups.map(group => ({g0: group.g0, count: group.count})).sort((a, b) => a.g0.localeCompare(b.g0))).toEqual(serviceGroups(facts));
  }
  expect(new Set(parentFacts.map(record => record.service)).size).toBe(2);

  expect(state.reads).toBe(4);
  state.event('comparison-run');

  await page.goto(base + '/' + questionLink(original));
  const parentAgain = await run(page, f.records);
  expect(parentAgain.groups).toEqual(result.groups);
  await require('./support/explore').openPreserve(page);await page.locator('#exploreSaveCapture').click();
  await page.locator('#captureDraftTitle').fill('Observed log patterns in fixed service scope');
  const captured = page.waitForResponse(response => isResponse('captures', response));
  await page.locator('#captureDraftConfirm').click();
  const saved = await json(await captured);
  await expect(page.locator('#captureGeneralID')).toHaveValue(saved.receipt.id);
  expect(saved.explore.groups).toEqual(result.groups);
  expect(saved.receipt.question.explore.groupBy).toEqual([patternField]);
  state.event('capture');
  const reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id));
  expect(reopened).toEqual(saved);
  expect(state.reads).toBe(6);
  state.event('reopen');
  const savedPatterns = page.locator('.log-pattern-results:visible');
  assertPatternRows(await patternRows(savedPatterns), saved.explore);
  await page.setViewportSize({width: 390, height: 1800});
  await savedPatterns.evaluate(element => {element.scrollIntoView({block: 'start'}); window.scrollBy(0, -120)});
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await savedPatterns.screenshot({path: 'docs/journeys/screenshots/182b-saved-log-patterns-mobile.png'});
  await page.setViewportSize({width: 1440, height: 1100});
  await page.locator('#savedexploreCreatePackage').click();
  await page.locator('#packageName').fill(f.environment);
  await page.locator('#packageTitle').fill('Reusable service log patterns');
  await page.locator('#packageLookback').fill('3');
  let packageResponse = page.waitForResponse(response => isResponse('telemetry-packages/draft', response));
  await page.locator('#packageBuild').click();
  const draft = await json(await packageResponse);
  expect(draft.question.groupBy).toEqual([patternField]);
  const yaml = await page.locator('#packageYAML').inputValue();
  expect(yaml).toContain('body_pattern_v1');
  expect(state.reads).toBe(6);
  state.event('reuse-stage');
  packageResponse = page.waitForResponse(response => isResponse('telemetry-packages', response));
  await page.locator('#packageSave').click();
  const reusable = await json(await packageResponse);
  expect(state.reads).toBe(6);
  state.event('reuse-save');
  const freshResponse = page.waitForResponse(response => isResponse('explore', response), {timeout: 45000});
  await page.locator('#packageUse').click();
  const reused = await json(await freshResponse);
  assertResult(reused, f.records);
  expect(reused.question.groupBy).toEqual([patternField]);
  expect(reused.question.where).toEqual(original.where);
  expect(reused.question.toMs - reused.question.fromMs).toBe(180000);
  expect(await json(await request.get(base + '/api/captures/' + saved.receipt.id))).toEqual(saved);
  state.event('reuse-run');
  await page.getByRole('button', {name: 'Return to paused activity', exact: true}).click();
  await expect(tail).toHaveAttribute('data-state', 'paused');
  await expect(activity.locator('.entry-investigate-options>summary')).toBeFocused();
  expect(state.streams).toBe(1);
  expect(state.errors).toEqual([]);
  state.event('return-paused');
  const directory = path.resolve('artifacts/log-patterns', f.environment);
  await fs.mkdir(directory, {recursive: true});
  await fs.writeFile(path.join(directory, 'ledger.json'), JSON.stringify({...f, applied, original, result, initialRows, rareRows, rawDetail, frequentID, selectedStage, selected, comparison, parentAgain, saved, reopened, draft, reusable, reused, yaml, events: state.events, requests: state.requests}, null, 2));
});

test('182 stale pattern links cannot stage hidden or changed fixed scopes', async ({page, request}) => {
  test.setTimeout(180000);
  page.setDefaultTimeout(15000);
  page.setDefaultNavigationTimeout(45000);
  const f = await fixture(request), state = watch(page), observations = [];
  await observeStages(page);
  const question = {...f.question, groupBy: [patternField], bucketMs: 1000};
  await page.goto(base + '/' + questionLink(question));
  const result = await run(page, f.records);
  const frequentID = pattern('request id=1000 completed in 0ms');
  const link = page.locator('.log-pattern-results:visible').locator('tr[data-pattern="' + frequentID + '"]').getByRole('link', {name: 'Open pattern logs', exact: true});
  await link.evaluate(element => {window.__oldPatternLink = element});
  state.event('original-run');
  async function obsolete(action, mutate) {
    await mutate();
    const before = await page.evaluate(() => ({url: location.href, stages: window.__patternStages.length, connected: window.__oldPatternLink.isConnected}));
    await page.evaluate(() => window.__oldPatternLink.dispatchEvent(new MouseEvent('click', {bubbles: true, cancelable: true, view: window})));
    const after = await page.evaluate(() => ({url: location.href, stages: window.__patternStages.length}));
    expect(after.url).toBe(before.url);
    expect(after.stages).toBe(before.stages);
    expect(state.reads).toBe(1);
    observations.push({action, before, after, reads: state.reads});
    state.event(action);
  }
  await obsolete('hidden-owner', async () => page.goto(base + '/#definitions'));
  expect(observations[0].before.connected).toBe(true);
  await page.goto(base + '/#explore');
  await obsolete('dirty-owner', async () => page.locator('#exploreEnvironment').fill(f.environment + '-draft'));
  expect(state.errors).toEqual([]);
  const directory = path.resolve('artifacts/log-patterns', f.environment);
  await fs.mkdir(directory, {recursive: true});
  await fs.writeFile(path.join(directory, 'ownership.json'), JSON.stringify({...f, question, result, frequentID, observations, events: state.events, requests: state.requests}, null, 2));
});

test('182 Bombadil investigates real full-population log patterns', async ({request}) => {
  test.skip(process.env.GOO_LOG_PATTERNS_BOMBADIL !== '1', 'Optional serialized ordinary OTel campaign.');
  test.setTimeout(180000);
  const f = await fixture(request), question = {...f.question, groupBy: [patternField], bucketMs: 1000};
  const directory = path.resolve('artifacts/log-patterns', f.environment, 'bombadil');
  await fs.mkdir(directory, {recursive: true});
  await fs.writeFile(path.join(directory, 'fixture.json'), JSON.stringify({...f, question}, null, 2));
  const actor = await startBombadil({dir: directory, url: base + '/' + questionLink(question), specification: 'verification/bombadil/log-patterns.ts', seconds: 55});
  try {await actor.finish()} finally {await actor.stop()}
});

test('182 more than one thousand real log patterns fail closed without a top list', async ({page, request}) => {
  test.setTimeout(180000);
  page.setDefaultTimeout(15000);
  page.setDefaultNavigationTimeout(45000);
  const f = await fixture(request, {highCardinality: true}), state = watch(page);
  const identities = new Set(members(f.records, f.question).map(record => pattern(record.body)));
  expect(identities.size).toBe(1100);
  expect(f.readyProbe.count).toBe('1100');
  const question = {...f.question, groupBy: [patternField], bucketMs: 1000};
  await page.goto(base + '/' + questionLink(question));
  await expect(page.locator('#exploreStatus')).toContainText('No evidence loaded');
  expect(state.reads).toBe(0);
  state.event('high-cardinality-stage');
  const pending = page.waitForResponse(response => isResponse('explore', response));
  await page.locator('#exploreRun').click();
  const response = await pending, refusal = {status: response.status(), request: response.request().postDataJSON(), response: await response.json()};
  expect(response.ok()).toBe(false);
  expect(refusal.response.error).toContain('more than 1000 groups; narrow query; no partial grouping totals');
  expect(refusal.response.groups).toBeUndefined();
  await expect(page.locator('#exploreStatus')).toContainText('more than 1000 groups');
  expect(await page.locator('.log-pattern-results:visible tr[data-pattern]').count()).toBe(0);
  expect(state.reads).toBe(1);
  expect(state.errors).toEqual([]);
  state.event('high-cardinality-refused');
  await page.setViewportSize({width: 1440, height: 1000});
  await page.locator('#exploreStatus').screenshot({path: 'docs/journeys/screenshots/182c-pattern-cardinality-refusal.png'});
  const directory = path.resolve('artifacts/log-patterns', f.environment);
  await fs.mkdir(directory, {recursive: true});
  await fs.writeFile(path.join(directory, 'cardinality.json'), JSON.stringify({...f, question, distinctPatterns: identities.size, refusal, events: state.events, requests: state.requests}, null, 2));
});
