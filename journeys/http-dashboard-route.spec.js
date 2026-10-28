const { test, expect } = require('@playwright/test');
const { readStepLink } = require('../ui/workspace.mjs');
const { base, json, seed, observe, loaded, snapshot, write } = require('./187-http-dashboard-support');

test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
test('187 route diagnosis stages exact HTTP operation, expands real trace and returns to held service', async ({ page, request }) => {
  test.setTimeout(120000);
  const facts = await seed(request), observation = observe(page);
  await page.goto(base + '/#services');
  await page.locator('#inventoryEnvironment').fill(facts.environment);
  await page.locator('#minutes').fill('2');
  await page.locator('#inventoryRun').click();
  await expect(page.getByRole('button', { name: 'Open service commerce / edge-api', exact: true })).toBeVisible();
  const planning = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / edge-api', exact: true }).click();
  const plan = await json(await planning);
  const dash = await loaded(page, 'http-server');
  const before = await snapshot(page, observation);
  const route = dash.locator('[data-widget=dimension] tr[data-diagnosis-value]').filter({ has: page.getByRole('cell', { name: '"/items"', exact: true }) });
  await expect(route).toHaveCount(1);
  const link = route.getByRole('link', { name: 'Open operation', exact: true });
  const staged = readStepLink(await link.getAttribute('href'));
  const parent = plan.widgets.find(widget => widget.key === 'dimension').query;
  expect(staged.record).toEqual({ ...parent, where: { all: [parent.where, { field: { kind: 'attribute', key: 'http.route' }, op: 'eq', value: '/items' }] }, offset: 0 });
  const sourceBeforeStage = observation.source;
  await link.click();
  await expect(page.locator('#exploreRun')).toBeVisible();
  expect(observation.source).toBe(sourceBeforeStage);
  observation.event('stage-route');
  let pending = page.waitForResponse(response => response.url().endsWith('/api/explore'));
  await page.locator('#exploreRun').click();
  const operation = await json(await pending);
  expect(operation.question).toEqual(staged.record);
  expect(operation.count).toBe('120');
  expect(operation.rows).toHaveLength(100);
  const expected = facts.records.filter(record => record.namespace === 'commerce' && record.service === 'edge-api' && record.kind === 2 && record.attributes['http.request.method'] && record.attributes['http.route'] === '/items');
  const expectedIDs = new Set(expected.map(record => record.id));
  expect(expected.length).toBe(120);
  expect(new Set(operation.rows.map(row => row.SpanId)).size).toBe(100);
  expect(operation.rows.every(row => expectedIDs.has(row.SpanId))).toBe(true);
  expect(operation.duration.sumNs).toBe('2400000000');
  expect(operation.spanStatus.error).toBe('24');
  expect(observation.source).toBe(sourceBeforeStage + 1);
  observation.event('run-route');

  const traceID = operation.rows[0].TraceId;
  const traceReading = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().postDataJSON().signal === 'spans');
  const logsReading = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().postDataJSON().signal === 'logs');
  await page.getByRole('button', { name: 'Expand trace ' + traceID, exact: true }).first().click();
  const trace = await json(await traceReading), logs = await json(await logsReading);
  await expect(page.locator('#exploreContextSummary')).toContainText('1 captured spans');
  await expect(page.locator('#exploreTrace .waterfall-row')).toHaveCount(1);
  await expect(page.locator('#exploreLogStatus')).toContainText('captured logs');
  for (const result of [trace, logs]) {
    expect(result.question.environment).toBe(parent.environment);
    expect(result.question.fromMs).toBe(parent.fromMs);
    expect(result.question.toMs).toBe(parent.toMs);
    expect(result.question.where).toEqual({ all: [{ field: { kind: 'column', key: 'trace_id' }, op: 'eq', value: traceID }] });
  }
  expect(trace.count).toBe('1');
  expect(trace.rows[0].SpanId).toBe(operation.rows[0].SpanId);
  expect(logs.count).toBe('0');
  expect(observation.source).toBe(sourceBeforeStage + 3);
  observation.event('expand-trace');
  const context = await page.locator('#exploreContext').evaluate(element => ({ text: element.textContent, visibleRows: element.querySelectorAll('.waterfall-row').length }));
  await page.setViewportSize({ width: 1440, height: 1300 });
  await page.locator('#exploreContext').screenshot({ path: 'docs/journeys/screenshots/187e-route-to-original-window-trace.png' });
  await page.locator('#exploreContextClose').click();
  const returnHashes = [];
  // Native links retain both the workspace transition and fixed-question URL.
  // Follow browser Back through those real history entries without another read.
  for (let i = 0; i < 3 && new URL(page.url()).hash !== '#services'; i++) {
    await page.goBack(); returnHashes.push(new URL(page.url()).hash);
  }
  await expect(page).toHaveURL(base + '/#services');
  await loaded(page, 'http-server');
  const after = await snapshot(page, observation);
  expect(after).toEqual(before);
  expect(observation.source).toBe(sourceBeforeStage + 3);
  observation.event('return-held-service');
  expect(observation.errors).toEqual([]);
  await write(facts, 'route.json', { ...facts, plan, before, staged, operation, traceID, trace, logs, context, returnHashes, after,
    events: observation.events, requests: observation.requests, responses: observation.responses });
});
