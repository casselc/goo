const { test, expect } = require('@playwright/test');
const { readStepLink } = require('../ui/workspace');
const { base, json, observe, loaded, snapshot, seed, find, assertAggregate, write, captureBytes } = require('./194-service-projection-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
async function discover(page, facts) {
  await page.goto(base + '/#services'); await page.locator('#inventoryEnvironment').fill(facts.environment); await page.locator('#minutes').fill('2');
  const pending = page.waitForResponse(response => response.url().endsWith('/api/discovery')); await page.locator('#inventoryRun').click(); return json(await pending);
}
async function open(page, observation, facts) {
  const catalog = await discover(page, facts), pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / edge-api', exact: true }).click(); const plan = await json(await pending);
  const dashboard = await loaded(page, 'http-server'); await observation.settle();
  const results = {};
  for (const widget of plan.widgets.filter(widget => widget.query)) { const result = find(observation, widget.query); expect(result).toBeTruthy(); assertAggregate(facts, result); results[widget.key] = result; }
  const supporting = observation.responses.find(response => response.endpoint === '/api/explore' && response.request.signal === 'spans' && !response.request.groupBy)?.body;
  expect(supporting.rows).toHaveLength(100); expect(supporting.rows.some(row => row.SpanKind === 'Server' && row.Attributes['http.request.method'])).toBe(true);
  return { catalog, plan, dashboard, results, supporting };
}
async function returnService(page, observation, expected) {
  const before = observation.source, hashes = [];
  for (let index = 0; index < 8 && !await page.locator('.service-dashboard').isVisible(); index++) { await page.goBack(); hashes.push(page.url()); }
  await expect(page.locator('.service-dashboard')).toBeVisible(); expect(observation.source).toBe(before);
  expect(await snapshot(page, observation)).toEqual(expected); return hashes;
}
async function geometry(page) {
  const widths = await page.locator('[data-widget=dimension] tbody tr').evaluateAll(rows => rows.map(row => [...row.querySelectorAll('.comparison-bar-track')].map(bar => bar.getBoundingClientRect().width)));
  for (const pair of widths) { expect(pair).toHaveLength(2); expect(pair[0]).toBeGreaterThan(50); expect(Math.abs(pair[0] - pair[1])).toBeLessThan(1); }
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true); return widths;
}
test('194 service widgets use rich aggregate reads while records and captures remain full evidence', async ({ page, request }) => {
  test.setTimeout(150000);
  const facts = await seed(request), observation = observe(page), initial = await open(page, observation, facts), events = [];
  const event = action => events.push({ action, source: observation.source, plans: observation.plans });
  expect(initial.results.spans.count).toBe('137'); expect(initial.results.logs.count).toBe('155');
  expect(initial.results.dimension.groups.map(group => [group.g0, group.count])).toEqual(expect.arrayContaining([[null, '3'], ['', '2'], ['/items', '120'], ['/custom', '12']]));
  expect(observation.requests.filter(item => item.endpoint === '/api/record-aggregates')).toHaveLength(3);
  expect(observation.requests.filter(item => item.endpoint === '/api/measurements')).toHaveLength(initial.plan.widgets.filter(widget => widget.measurement && widget.state === 'supported').length);
  expect(observation.requests.filter(item => item.endpoint === '/api/explore')).toHaveLength(1); // Admission witness stays a full record read.
  const heldView = await snapshot(page, observation), desktop = await geometry(page); event('overview');
  await page.evaluate(() => scrollTo({ top: Math.max(0, document.querySelector('#serviceDetail').getBoundingClientRect().top + scrollY - 85), behavior: 'instant' }));
  await page.screenshot({ path: 'docs/journeys/screenshots/194c-service-overview-in-context.png' });
  await initial.dashboard.screenshot({ path: 'docs/journeys/screenshots/194a-rich-service-aggregates.png' });
  await page.setViewportSize({ width: 390, height: 1100 }); const mobile = await geometry(page);
  await initial.dashboard.screenshot({ path: 'docs/journeys/screenshots/194b-mobile-service-aggregates.png' }); await page.setViewportSize({ width: 1440, height: 1100 });
  const logLink = page.locator('[data-widget=logs]').getByRole('link', { name: 'Open widget question', exact: true });
  const staged = readStepLink(await logLink.getAttribute('href')), beforeStage = observation.source; await logLink.click(); expect(observation.source).toBe(beforeStage); event('logs-staged');
  let pending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST');
  await page.locator('#exploreRun').click(); const records = await json(await pending);
  expect(records.count).toBe('155'); expect(records.rows).toHaveLength(100); expect(records.question).toEqual(initial.plan.widgets.find(widget => widget.key === 'logs').query); expect(records.fields.length).toBeGreaterThan(0); event('logs-full-read');
  const logReturn = await returnService(page, observation, heldView); event('logs-return');
  const beforeCapture = observation.source, widget = page.locator('[data-widget=spans]');
  await widget.getByText('Widget actions', { exact: true }).click(); await widget.getByRole('button', { name: 'Capture widget evidence', exact: true }).click(); expect(observation.source).toBe(beforeCapture);
  await page.locator('#captureDraftTitle').fill('Rich service widget, fresh full evidence'); pending = page.waitForResponse(response => response.url().endsWith('/api/captures') && response.request().method() === 'POST');
  await page.locator('#captureDraftConfirm').click(); const saved = await json(await pending);
  expect(saved.explore.count).toBe('137'); expect(saved.explore.rows).toHaveLength(100); expect(saved.explore.fields.length).toBeGreaterThan(0); expect(saved.explore.question).toEqual(initial.plan.widgets.find(widget => widget.key === 'spans').query);
  const storedBefore = await captureBytes(facts, saved.receipt.id, 'capture-before.json');
  expect(storedBefore.data.explore.rows).toHaveLength(137);
  event('fresh-full-capture'); const beforeReopen = observation.source;
  await expect(page.locator('#savedexploreCount')).toContainText('137'); expect(observation.source).toBe(beforeReopen);
  const reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id)); expect(reopened).toEqual(saved); event('frozen-api-reopen');
  const beforeServices = observation.source; await page.locator('[data-workspace=services]').click();
  await expect(page.locator('.service-dashboard')).toBeVisible(); expect(observation.source).toBe(beforeServices);
  const captureReturn = { kind: 'Services navigation', view: await snapshot(page, observation) }; expect(captureReturn.view).toEqual(heldView); event('capture-services-navigation');
  const route = page.locator('[data-widget=dimension] tbody tr').filter({ has: page.getByText(JSON.stringify('/items'), { exact: true }) });
  const beforeRoute = observation.source; await route.getByRole('link', { name: 'Open operation', exact: true }).click(); expect(observation.source).toBe(beforeRoute);
  pending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST');
  await page.locator('#exploreRun').click(); const operation = await json(await pending); expect(operation.count).toBe('120'); expect(operation.rows).toHaveLength(100); event('operation-full-read');
  const row = page.locator('#exploreRows .record-row').first(), traceId = operation.rows[0].TraceId;
  const tracePending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().postDataJSON().signal === 'spans');
  const logsPending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().postDataJSON().signal === 'logs');
  await row.getByRole('button', { name: 'Expand trace ' + traceId, exact: true }).click(); const trace = await json(await tracePending), traceLogs = await json(await logsPending);
  expect(trace.count).toBe('1'); expect(traceLogs.count).toBe('0'); event('trace'); await page.locator('#exploreContextClose').click();
  const operationReturn = await returnService(page, observation, heldView); event('operation-return');
  const last = facts.records.filter(record => record.attributes['test.case'] === 'incoming-current').at(-1), range = { fromMs: last.timeMs, toMs: last.timeMs + 1 };
  expect(initial.supporting.rows.some(row => BigInt(row.timestamp_ns) >= BigInt(range.fromMs) * 1000000n && BigInt(row.timestamp_ns) < BigInt(range.toMs) * 1000000n)).toBe(false);
  await page.locator('.service-dashboard .service-timeline').getByText('Select a time range', { exact: true }).click();
  const beforeRange = observation.source;
  await page.locator('.service-dashboard goo-time-range').evaluate((control, range) => { const start = control.querySelectorAll('input[type=range]')[0]; start.value = String(range.fromMs); start.dispatchEvent(new Event('input', { bubbles: true })); const end = control.querySelectorAll('input[type=range]')[1]; end.value = String(range.toMs); end.dispatchEvent(new Event('input', { bubbles: true })); }, range);
  expect(observation.source).toBe(beforeRange); pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Apply range to overview', exact: true }).click(); const rangeResponse = await pending, rangeRequest = rangeResponse.request().postDataJSON(), rangePlan = await json(rangeResponse);
  await loaded(page, 'http-server'); await observation.settle(); expect(rangeRequest.profile).toBe('http-server'); expect(rangeRequest.httpServerObserved || false).toBe(false);
  const rangeResult = find(observation, rangePlan.widgets.find(widget => widget.key === 'spans').query); assertAggregate(facts, rangeResult); expect(rangeResult.count).toBe('1'); event('range-read');
  const beforeFrozenUI = observation.source; await page.goto(base + '/#capture=' + saved.receipt.id);
  await expect(page.locator('#savedexploreCount')).toContainText('137'); expect(observation.source).toBe(beforeFrozenUI); event('frozen-ui-reopen');
  const storedAfter = await captureBytes(facts, saved.receipt.id, 'capture-after.json'); expect(storedAfter.sha256).toBe(storedBefore.sha256);
  expect(observation.errors).toEqual([]);
  await write(facts, 'ledger.json', { version: 1, facts, initial: { ...initial, dashboard: undefined }, heldView, desktop, mobile, staged, records, logReturn, saved, reopened, storedBefore, storedAfter, captureReturn, operation, trace, traceLogs, operationReturn, range, rangeRequest, rangePlan, rangeResult, events, requests: observation.requests, responses: observation.responses, errors: observation.errors });
});

test('194 late rich aggregate completion cannot reopen a departed service', async ({ page, request }) => {
  test.setTimeout(120000);
  const facts = await seed(request), observation = observe(page), initial = await open(page, observation, facts);
  let release, entered, finished, held;
  const gate = new Promise(resolve => { release = resolve; }), started = new Promise(resolve => { entered = resolve; }), done = new Promise(resolve => { finished = resolve; });
  await page.route('**/api/record-aggregates', async route => {
    const q = route.request().postDataJSON();
    if (q.signal !== 'spans' || q.groupBy[0].key !== 'name') return route.continue();
    const response = await route.fetch(); held = { question: q, result: await response.json() }; assertAggregate(facts, held.result); entered();
    await gate; try { await route.fulfill({ response }); } catch {} finally { finished(); }
  });
  await page.locator('[data-dashboard-profile=all-spans]').click(); await started;
  await page.locator('#serviceDetail').getByRole('button', { name: 'Back to service list', exact: true }).click();
  await expect(page.locator('#serviceDetail')).toBeHidden(); const beforeRelease = observation.source;
  release(); await done; await page.unroute('**/api/record-aggregates'); await observation.settle();
  await expect(page.locator('#serviceDetail')).toBeHidden(); expect(observation.source).toBe(beforeRelease);
  const afterRelease = await page.locator('#serviceDetail').evaluate(element => ({ hidden: element.hidden, visible: !!element.getClientRects().length }));
  const pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / edge-api', exact: true }).click(); const freshPlan = await json(await pending);
  await loaded(page, 'http-server'); await observation.settle(); const fresh = find(observation, freshPlan.widgets.find(widget => widget.key === 'spans').query);
  assertAggregate(facts, fresh); expect(fresh.count).toBe('137'); expect(observation.errors).toEqual([]);
  await write(facts, 'ownership.json', { version: 1, facts, originalPlan: initial.plan, held, beforeRelease, afterRelease, freshPlan, fresh, requests: observation.requests, responses: observation.responses, errors: observation.errors });
});
