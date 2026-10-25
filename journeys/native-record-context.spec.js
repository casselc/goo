const { test, expect } = require('@playwright/test');
const { measurementStep, stepLink, readQuestionLink } = require('../ui/workspace');
const { members } = require('./192-native-context-fixture');
const { base, json, question, seed, observe, write } = require('./192-native-context-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
async function openNative(page, observation, q, streams = 3) {
  await page.goto(base + '/' + stepLink(measurementStep(q)));
  await expect(page.locator('#measureRun')).toBeEnabled();
  const before = observation.source();
  const pending = page.waitForResponse(response => response.url().endsWith('/api/measurements') && response.request().method() === 'POST');
  await page.locator('#measureRun').click(); const result = await json(await pending);
  expect(observation.source()).toBe(before + 1);
  await expect(page.locator('#measureResult .native-stream-list > details')).toHaveCount(streams);
  return result;
}
async function selectInterval(page, namespace = 'inventory', key = '0') {
  const host = page.locator('#measureResult');
  await host.getByLabel('Find a stream', { exact: true }).fill(namespace);
  const stream = host.locator('.native-stream-list > details');
  await expect(stream).toHaveCount(1);
  if (!await stream.evaluate(element => element.open)) await stream.locator(':scope > summary').click();
  const point = stream.locator('goo-value-timeline button[data-key="' + key + '"]');
  await point.click();
  await expect(stream.locator('.native-record-context')).toBeVisible();
  return stream;
}
async function retained(stream) {
  return stream.evaluate(element => ({ text: element.textContent, open: element.open,
    parentEvidence: element.closest('#measureResult').querySelector(':scope > details pre')?.textContent,
    timeline: element.querySelector('goo-value-timeline').data,
    selected: element.querySelector('goo-value-timeline').selectedKey,
    bounds: { ...element.querySelector('.native-record-context').dataset },
  }));
}
function assertRecords(facts, result, namespace = 'inventory', interval) {
  const expected = members(facts, namespace, interval);
  expect(result.count).toBe(String(expected.length));
  expect(result.rows).toHaveLength(Math.min(100, expected.length));
  const byID = new Map(expected.map(record => [record.id, record]));
  for (const row of result.rows) {
    const id = result.question.signal === 'spans' ? row.SpanId : row.SpanId;
    const record = byID.get(id); expect(record, 'Unexpected interval record ' + id).toBeTruthy();
    expect(row.timestamp_ns).toBe(record.timestampNS);
    expect(row.ResourceAttributes['service.instance.id']).toBe('worker-a');
  }
}

test('192 native interval stages exact record context and returns to retained native evidence', async ({ page, request }) => {
  test.setTimeout(150000);
  const facts = await seed(request), observation = observe(page), events = [];
  const event = (action, extra = {}) => events.push({ action, source: observation.source(), ...extra });
  const native = await openNative(page, observation, question(facts)); event('native-run');
  const stream = await selectInterval(page), before = await retained(stream); event('interval-selected');
  expect(before.bounds.fromNs).toBe(facts.fromNS); expect(before.bounds.toNs).toBe(facts.toNS);
  const link = stream.locator('[data-native-record-context=spans]');
  const staged = readQuestionLink(await link.getAttribute('href'));
  await link.evaluate(element => { window.nativeContextOpener = element; });
  const beforeStage = observation.source(); await link.click();
  await expect(page.locator('#exploreRun')).toBeVisible(); expect(observation.source()).toBe(beforeStage);
  await expect(page.locator('#nativeRecordReturn')).toContainText(native.basis);
  await expect(page.locator('#nativeRecordReturn')).toContainText(native.readAt);
  event('records-staged');
  let pending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST');
  await page.locator('#exploreRun').click(); const records = await json(await pending);
  assertRecords(facts, records); expect(records.count).toBe('115'); event('records-run');
  const where = records.question.where.all;
  expect(where).toContainEqual({ field: { kind: 'column', key: 'timestamp_ns' }, op: 'gte', value: facts.fromNS });
  expect(where).toContainEqual({ field: { kind: 'column', key: 'timestamp_ns' }, op: 'lte', value: String(BigInt(facts.toNS) - 1n) });
  expect(where.map(leaf => leaf.field.key).sort()).toEqual(['service', 'service.instance.id', 'service.namespace', 'timestamp_ns', 'timestamp_ns'].sort());
  await page.screenshot({ path: 'docs/journeys/screenshots/192a-native-interval-record-context.png' });
  if (await page.locator('#resultRecords').isVisible()) await page.locator('#resultRecords').click();
  const row = page.locator('#exploreRows tbody tr').filter({ has: page.getByText('inside-0', { exact: true }) });
  await expect(row).toHaveCount(1);
  pending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST' && response.request().postDataJSON().signal === 'spans');
  const logsPending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST' && response.request().postDataJSON().signal === 'logs');
  await row.getByRole('button', { name: 'Expand trace ' + facts.traceId, exact: true }).click();
  const trace = await json(await pending), linkedLogs = await json(await logsPending);
  expect(trace.count).toBe('4'); expect(linkedLogs.count).toBe('4');
  expect(trace.question.fromMs).toBe(records.question.fromMs); expect(trace.question.toMs).toBe(records.question.toMs);
  await expect(page.locator('#exploreContextExplain')).toContainText(facts.fromNS); event('trace-expanded');
  const traceOrigin = JSON.parse(await page.locator('#exploreContextExplain').textContent());
  expect(traceOrigin.origin).toEqual(records.question);
  await expect(page.locator('#nativeRecordReturn')).toContainText('Trace expansion is broader');
  // The root-owned retained-return controller supplies this native parent action.
  await page.getByRole('button', { name: 'Back to measurement', exact: true }).click();
  const after = await retained(stream); expect(after).toEqual(before);
  expect(await page.evaluate(() => document.activeElement === window.nativeContextOpener)).toBe(true); event('native-return');
  expect(observation.source()).toBe(beforeStage + 3);
  const logsLink = stream.locator('[data-native-record-context=logs]'); await logsLink.click(); event('logs-staged');
  pending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST');
  await page.locator('#exploreRun').click(); const logs = await json(await pending); assertRecords(facts, logs); event('logs-run');
  await page.setViewportSize({ width: 390, height: 1000 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.screenshot({ path: 'docs/journeys/screenshots/192b-mobile-exact-interval-logs.png' });
  await page.getByRole('button', { name: 'Back to measurement', exact: true }).click(); event('logs-return');
  expect(await retained(stream)).toEqual(before); expect(observation.errors).toEqual([]);
  await write(facts, 'ledger.json', { version: 1, facts, native, before, staged, records, trace, linkedLogs, traceOrigin, after, logs, events,
    requests: observation.requests, responses: observation.responses, errors: observation.errors });
});


test('192 histogram, missing namespace, unknown export and empty interval retain honest record context', async ({ page, request }) => {
  test.setTimeout(150000);
  const facts = await seed(request), observation = observe(page), cases = [];
  const scenarios = [
    { name: 'histogram-empty-namespace', kind: 'histogram', namespace: '', signal: 'spans', count: '1' },
    { name: 'counter-missing-namespace', kind: 'counter', namespace: null, signal: 'logs', count: '1' },
    { name: 'unknown-export', kind: 'counter', unknown: true, namespace: 'inventory', signal: 'logs', count: '115' },
    { name: 'known-empty-interval', kind: 'counter', namespace: 'inventory', signal: 'spans', key: '1', count: '0' },
  ];
  for (const scenario of scenarios) {
    const q = question(facts, scenario.kind, scenario.unknown);
    if (scenario.namespace === null) q.definition.source.namespace = { mode: 'missing' };
    else q.definition.source.namespace = { mode: 'value', value: scenario.namespace };
    const native = await openNative(page, observation, q, 1);
    const stream = await selectInterval(page, '', scenario.key || '0'), view = await retained(stream);
    const intervals = (native.counters || native.histograms).series[0].intervals;
    const selectedInterval = intervals[Number(scenario.key || '0')];
    if (scenario.unknown) { expect(selectedInterval.state).toBe('unknown'); expect(selectedInterval.perSecond).toBeUndefined(); }
    const link = stream.locator('[data-native-record-context=' + scenario.signal + ']');
    const staged = readQuestionLink(await link.getAttribute('href')), beforeStage = observation.source();
    await link.click(); await expect(page.locator('#exploreRun')).toBeVisible(); expect(observation.source()).toBe(beforeStage);
    const pending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST');
    await page.locator('#exploreRun').click(); const result = await json(await pending);
    expect(result.count).toBe(scenario.count);
    assertRecords(facts, result, scenario.namespace, { fromNS: selectedInterval.fromNS, toNS: selectedInterval.toNS });
    await page.getByRole('button', { name: 'Back to measurement', exact: true }).click();
    expect(await retained(stream)).toEqual(view); expect(observation.source()).toBe(beforeStage + 1);
    cases.push({ ...scenario, native, view, selectedInterval, staged, result, sourceBeforeStage: beforeStage, sourceAfterReturn: observation.source() });
  }
  expect(observation.errors).toEqual([]);
  await write(facts, 'cases.json', { version: 1, facts, cases, requests: observation.requests, responses: observation.responses, errors: observation.errors });
});


test('192 late record response and hidden or closed native actions cannot take ownership', async ({ page, request }) => {
  test.setTimeout(120000);
  const facts = await seed(request), observation = observe(page), events = [];
  const native = await openNative(page, observation, question(facts)), stream = await selectInterval(page), before = await retained(stream);
  const opener = stream.locator('[data-native-record-context=spans]');
  await opener.evaluate(element => { window.retainedNativeAction = element; });
  await opener.click();
  const stagedURL = page.url(), beforeHidden = observation.source();
  await page.evaluate(() => window.retainedNativeAction.click());
  expect(page.url()).toBe(stagedURL); expect(observation.source()).toBe(beforeHidden);
  events.push({ action: 'hidden-action-rejected', source: observation.source(), url: page.url() });
  let release, held, settled;
  const finished = new Promise(resolve => { settled = resolve; });
  const gate = new Promise(resolve => { release = resolve; });
  let observed; const started = new Promise(resolve => { observed = resolve; });
  await page.route('**/api/explore', async route => {
    const response = await route.fetch(); held = await response.json(); observed();
    await gate; try { await route.fulfill({ response }); } catch {} finally { settled(); }
  });
  await page.locator('#exploreRun').click(); await started;
  events.push({ action: 'record-read-held', source: observation.source(), heldQuestion: held.question });
  await page.getByRole('button', { name: 'Back to measurement', exact: true }).click();
  expect(await retained(stream)).toEqual(before);
  events.push({ action: 'return-during-read', source: observation.source(), visible: await page.locator('#measurementWorkbench').isVisible() });
  const cancelledView = await page.locator('#exploration').evaluate(host => ({ count: host.querySelector('#exploreCount')?.textContent, rows: host.querySelectorAll('#exploreRows tbody tr').length, state: host.querySelector('#exploreQuestionState')?.textContent }));
  release(); await finished; await page.unroute('**/api/explore');
  await expect(page.locator('#measurementWorkbench')).toBeVisible();
  expect(await retained(stream)).toEqual(before);
  const afterLate = await page.locator('#exploration').evaluate(host => ({ count: host.querySelector('#exploreCount')?.textContent, rows: host.querySelectorAll('#exploreRows tbody tr').length, state: host.querySelector('#exploreQuestionState')?.textContent }));
  expect(afterLate).toEqual(cancelledView);
  events.push({ action: 'late-response-rejected', source: observation.source(), visible: await page.locator('#measurementWorkbench').isVisible(), cancelledView, afterLate });
  // Close the retained inspector, then dispatch its still-connected old link.
  await stream.getByRole('button', { name: 'Close selected item details', exact: true }).click();
  const closedURL = page.url(), beforeClosed = observation.source();
  await page.evaluate(() => window.retainedNativeAction.click());
  expect(page.url()).toBe(closedURL); expect(observation.source()).toBe(beforeClosed);
  events.push({ action: 'closed-action-rejected', source: observation.source(), url: page.url() });
  await stream.locator('goo-value-timeline button[data-key="0"]').click();
  await stream.locator('[data-native-record-context=spans]').click();
  const beforeBack = observation.source(); await page.goBack();
  await expect(page.locator('#measurementWorkbench')).toBeVisible();
  expect(observation.source()).toBe(beforeBack); expect(await retained(stream)).toEqual(before);
  events.push({ action: 'browser-back-retains-native', source: observation.source(), visible: await page.locator('#measurementWorkbench').isVisible() });
  expect(observation.errors).toEqual([]);
  await write(facts, 'ownership.json', { version: 1, facts, native, before, held, events, requests: observation.requests, responses: observation.responses, errors: observation.errors });
});
