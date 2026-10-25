const { test, expect } = require('@playwright/test');
const { readStepLink } = require('../ui/workspace');
const { base, json, seed, observe, loaded, snapshot, assertPopulation, write } = require('./187-http-dashboard-support');

test.use({ actionTimeout: 15000, navigationTimeout: 45000 });

async function switchProfile(page, observation, profile) {
  const begin = observation.requests.length;
  const planning = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.locator('[data-dashboard-profile="' + profile + '"]').click();
  const plan = await json(await planning);
  await loaded(page, profile);
  await observation.settle();
  const requests = observation.requests.slice(begin);
  expect(requests.filter(request => request.endpoint === '/api/dashboards/plan')).toHaveLength(1);
  // Switching consumes held discovery metadata; no metadata or fallback reread is needed.
  expect(requests.filter(request => ['/api/discovery', '/api/inventory', '/api/metric-discovery'].includes(request.endpoint))).toEqual([]);
  expect(requests.filter(request => request.effect === 'source')).toHaveLength(plan.widgets.filter(widget => widget.state === 'supported').length);
  return { plan, view: await snapshot(page, observation), requests };
}
function findResult(observation, question) {
  const canonical = value => Array.isArray(value) ? value.map(canonical) : value && typeof value === 'object' ? Object.fromEntries(Object.keys(value).sort().map(key => [key, canonical(value[key])])) : value;
  const stable = value => JSON.stringify(canonical(value));
  // Question properties are canonicalized by Go; the fixed plan query is also echoed by the endpoint.
  return [...observation.responses].reverse().find(response => response.endpoint === '/api/explore' && response.body.question && stable(response.body.question) === stable(question))?.body;
}
async function routeGeometry(page) {
  const rows = await page.locator('[data-widget=dimension] tbody tr').evaluateAll(rows => rows.map(row => ({
    text: row.textContent,
    bars: [...row.querySelectorAll('.comparison-bar-cell')].map(cell => ({
      value: cell.dataset.value, maximum: cell.dataset.scaleMax,
      trackWidth: cell.querySelector('.comparison-bar-track').getBoundingClientRect().width,
      fillWidth: cell.querySelector('.comparison-bar-fill').getBoundingClientRect().width,
    })),
  })));
  expect(rows).toHaveLength(4);
  for (const row of rows) {
    expect(row.bars).toHaveLength(2);
    expect(row.bars[0].value).toBe(row.bars[1].value);
    expect(row.bars[0].maximum).toBe(row.bars[1].maximum);
    expect(row.bars[0].trackWidth).toBeGreaterThan(50);
    expect(Math.abs(row.bars[0].trackWidth - row.bars[1].trackWidth)).toBeLessThan(1);
    expect(Math.abs(row.bars[0].fillWidth - row.bars[1].fillWidth)).toBeLessThan(1);
  }
  return rows;
}
async function laneGeometry(page) {
  const lanes = await page.locator('.service-dashboard .service-timeline-lane').evaluateAll(rows => rows.map(row => {
    const label = row.querySelector(':scope > div'), plot = row.querySelector('goo-value-timeline');
    const bounds = node => { const box = node.getBoundingClientRect(); return { left: box.left, right: box.right, width: box.width }; };
    return { key: row.dataset.lane, label: bounds(label), plot: bounds(plot), children: [...label.children].map(child => ({ ...bounds(child), text: child.textContent, scrollWidth: child.scrollWidth, clientWidth: child.clientWidth })) };
  }));
  expect(lanes.length).toBeGreaterThanOrEqual(5);
  for (const lane of lanes) {
    expect(lane.label.right).toBeLessThanOrEqual(lane.plot.left + 1);
    for (const child of lane.children) {
      expect(child.left).toBeGreaterThanOrEqual(lane.label.left - 1);
      expect(child.right).toBeLessThanOrEqual(lane.label.right + 1);
      expect(child.scrollWidth).toBeLessThanOrEqual(child.clientWidth + 1);
    }
  }
  return lanes;
}
function sameScope(plan, other) {
  expect(other.scope).toEqual(plan.scope);
  expect(other.service).toBe(plan.service);
}
async function inspectPlan(facts, observation, plan) {
  const spans = plan.widgets.find(widget => widget.key === 'spans');
  expect(spans.query.environment).toBe(facts.environment);
  expect(plan.scope.namespace).toEqual({ mode: 'value', value: 'commerce' });
  const result = findResult(observation, spans.query);
  expect(result).toBeTruthy();
  assertPopulation(facts, result, plan.profile);
  const routes = plan.widgets.find(widget => widget.key === 'dimension');
  const routeResult = findResult(observation, routes.query);
  assertPopulation(facts, routeResult, plan.profile);
  const http = plan.profile === 'http-server';
  if (http) {
    expect(spans.grain).toBe('server-span-record');
    expect(spans.query.where.all).toContainEqual({ field: { kind: 'column', key: 'kind' }, op: 'eq', value: 'Server' });
    expect(spans.query.where.all).toContainEqual({ field: { kind: 'attribute', key: 'http.request.method' }, op: 'ne' });
    expect(routes.query.where).toEqual(spans.query.where);
  }
  return { spans: result, routes: routeResult };
}

test('187 Incoming HTTP defaults retain exact server population, native grain and generic fallback', async ({ page, request }) => {
  test.setTimeout(180000);
  const facts = await seed(request), observation = observe(page);
  await page.goto(base + '/#services');
  await expect(page.locator('#inventoryRun')).toBeVisible();
  await page.locator('#inventoryEnvironment').fill(facts.environment);
  await page.locator('#minutes').fill('2');
  const discovering = page.waitForResponse(response => response.url().endsWith('/api/discovery'));
  await page.locator('#inventoryRun').click();
  const catalog = await json(await discovering);
  const planning = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / edge-api', exact: true }).click();
  const plan = await json(await planning);
  const dash = await loaded(page, 'http-server');
  await observation.settle();
  expect(plan.declarations).toEqual([]);
  const initial = { plan, results: await inspectPlan(facts, observation, plan), view: await snapshot(page, observation) };
  expect(initial.results.spans.count).toBe('137');
  expect(initial.results.spans.rows).toHaveLength(100);
  expect(initial.view.rateTitle).toContain('not unique request throughput');
  expect(initial.view.rateTitle).toContain('137000/' + (plan.scope.toMs - plan.scope.fromMs));
  const native = plan.widgets.filter(widget => widget.measurement);
  expect(native).toHaveLength(2);
  expect(native[0].measurement.definition.source.metric).toBe('http.server.request.duration');
  expect(native[0].grain).not.toBe('server-span-record');
  await expect(dash.locator('[data-widget=spans]')).toContainText('Observed server-span records');
  await expect(dash.locator('[data-widget=spans]')).not.toContainText('900.000 ms');
  observation.event('automatic-http');

  const beforeLocal = observation.source;
  await dash.locator('[data-widget=spans] > .dashboard-widget-evidence > summary').click();
  await dash.getByText('Available views and prerequisites', { exact: true }).click();
  expect(observation.source).toBe(beforeLocal);
  await dash.getByText('Available views and prerequisites', { exact: true }).click();
  await dash.locator('[data-widget=spans] > .dashboard-widget-evidence > summary').click();
  observation.event('inspect-held-evidence');
  await page.setViewportSize({ width: 1440, height: 1800 });
  const layout = { desktop: await routeGeometry(page), desktopLanes: await laneGeometry(page) };
  await page.getByRole('heading', { name: 'Services', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/187c-service-overview-in-context.png' });
  await dash.screenshot({ path: 'docs/journeys/screenshots/187a-incoming-http-service-dashboard.png' });
  await page.setViewportSize({ width: 390, height: 1600 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  layout.mobile = await routeGeometry(page);
  layout.mobileLanes = await laneGeometry(page);
  await page.getByRole('heading', { name: 'Services', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/187d-mobile-service-overview-in-context.png' });
  await dash.locator('[data-widget=spans]').screenshot({ path: 'docs/journeys/screenshots/187b-mobile-http-server-records.png' });
  await page.setViewportSize({ width: 1440, height: 1100 });

  const generic = await switchProfile(page, observation, 'all-spans');
  sameScope(plan, generic.plan);
  generic.results = await inspectPlan(facts, observation, generic.plan);
  expect(generic.results.spans.count).toBe('198');
  observation.event('explicit-all-spans');
  const returned = await switchProfile(page, observation, 'http-server');
  sameScope(plan, returned.plan);
  returned.results = await inspectPlan(facts, observation, returned.plan);
  expect(returned.results.spans.count).toBe('137');
  observation.event('explicit-http-return');

  // Narrow to a late one-millisecond interval absent from the held supporting page.
  // The resolved profile must stay explicit instead of being guessed again on Apply.
  const lastServer = facts.records.filter(record => record.attributes['test.case'] === 'incoming-current').at(-1);
  const selected = { fromMs: lastServer.timeMs, toMs: lastServer.timeMs + 1 };
  const supporting = observation.responses.find(response => response.endpoint === '/api/explore' && response.request.signal === 'spans' && !response.request.groupBy)?.body;
  expect(supporting.rows).toHaveLength(100);
  expect(supporting.rows.some(row => BigInt(row.timestamp_ns) >= BigInt(selected.fromMs) * 1000000n && BigInt(row.timestamp_ns) < BigInt(selected.toMs) * 1000000n)).toBe(false);
  const beforeRange = observation.source;
  await dash.locator('.service-timeline').getByText('Select a time range', { exact: true }).click();
  await dash.locator('.service-timeline goo-time-range').evaluate((control, selected) => {
    const start = control.querySelectorAll('input[type=range]')[0];
    start.value = String(selected.fromMs); start.dispatchEvent(new Event('input', { bubbles: true }));
    // Controlled selection redraws the native inputs; resolve the second live input anew.
    const end = control.querySelectorAll('input[type=range]')[1];
    end.value = String(selected.toMs); end.dispatchEvent(new Event('input', { bubbles: true }));
  }, selected);
  expect(observation.source).toBe(beforeRange);
  expect(await dash.locator('.service-timeline').evaluate(section => ({ fromMs: Number(section.dataset.fromMs), toMs: Number(section.dataset.toMs) }))).toEqual(selected);
  const rangePlanning = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await dash.getByRole('button', { name: 'Apply range to overview', exact: true }).click();
  const rangeResponse = await rangePlanning;
  const rangeRequest = rangeResponse.request().postDataJSON(), rangePlan = await json(rangeResponse);
  await loaded(page, 'http-server'); await observation.settle();
  expect(rangeRequest.profile).toBe('http-server');
  expect(rangeRequest.httpServerObserved || false).toBe(false);
  expect(rangePlan.scope).toEqual({ ...plan.scope, ...selected });
  const rangeResult = findResult(observation, rangePlan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, rangeResult, 'http-server');
  expect(BigInt(rangeResult.count)).toBeGreaterThan(0n);
  expect(BigInt(rangeResult.count)).toBeLessThan(137n);
  observation.event('apply-http-range', { selected });
  await dash.locator('.service-timeline').getByText('Select a time range', { exact: true }).click();
  const resetPlanning = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await dash.getByRole('button', { name: 'Reset overview window', exact: true }).click();
  const resetPlan = await json(await resetPlanning);
  await loaded(page, 'http-server'); await observation.settle();
  sameScope(plan, resetPlan);
  observation.event('reset-http-window');

  // Exact scoped widget links stage an intent; only the named Run reads again.
  const widget = dash.locator('[data-widget=spans]');
  await widget.getByText('Widget actions', { exact: true }).click();
  const link = widget.getByRole('link', { name: 'Open widget question', exact: true });
  const focused = readStepLink(await link.getAttribute('href'));
  expect(focused.record).toEqual(plan.widgets.find(widget => widget.key === 'spans').query);
  const beforeStage = observation.source;
  await dash.locator('[data-dashboard-profile=all-spans]').evaluate(button => { window.__httpDashboardRetainedProfile = button; });
  await link.click();
  await expect(page.locator('#exploreRun')).toBeVisible();
  expect(observation.source).toBe(beforeStage);
  const beforeStalePlan = observation.plans;
  await page.evaluate(() => window.__httpDashboardRetainedProfile.click());
  expect(observation.plans).toBe(beforeStalePlan);
  expect(observation.source).toBe(beforeStage);
  observation.event('stage-http-question', { retainedProfileRejected: true });
  const running = page.waitForResponse(response => response.url().endsWith('/api/explore'));
  await page.locator('#exploreRun').click();
  const drilldown = await json(await running);
  assertPopulation(facts, drilldown, 'http-server');
  expect(drilldown.question).toEqual(initial.results.spans.question);
  observation.event('run-http-question');

  // Each new service starts auto; an outgoing-only metadata candidate never implies incoming traffic.
  await page.goto(base + '/#services');
  await page.locator('#inventoryEnvironment').fill(facts.environment);
  await page.locator('#inventoryRun').click();
  await expect(page.getByRole('button', { name: 'Open service commerce / worker', exact: true })).toBeVisible();
  let nextPlan = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / worker', exact: true }).click();
  const workerPlan = await json(await nextPlan);
  await loaded(page, 'all-spans');
  const worker = { plan: workerPlan, view: await snapshot(page, observation) };
  expect(worker.view.families.find(family => family.key === 'http-server').disabled).toBe(true);
  observation.event('generic-worker');
  await page.locator('#serviceDetail').getByRole('button', { name: 'Back to service list', exact: true }).click();
  nextPlan = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / outbound', exact: true }).click();
  const outboundPlan = await json(await nextPlan);
  await loaded(page, 'all-spans');
  const outbound = { plan: outboundPlan, view: await snapshot(page, observation) };
  observation.event('outbound-default-generic');
  const emptyHTTP = await switchProfile(page, observation, 'http-server');
  expect(emptyHTTP.view.profileState).toBe('no_data');
  expect(emptyHTTP.view.text).toContain('All spans');
  const emptyResult = findResult(observation, emptyHTTP.plan.widgets.find(widget => widget.key === 'spans').query);
  expect(emptyResult.count).toBe('0');
  observation.event('outbound-explicit-empty-http');
  const outboundGeneric = await switchProfile(page, observation, 'all-spans');
  const outboundResult = findResult(observation, outboundGeneric.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, outboundResult, 'all-spans', 'outbound');
  observation.event('outbound-generic-recovery');
  await observation.settle();
  expect(observation.errors).toEqual([]);
  await write(facts, 'ledger.json', { ...facts, catalog, layout, initial, generic, returned, supporting, rangeRequest, rangePlan, rangeResult, resetPlan, focused, drilldown, worker, outbound, emptyHTTP, outboundGeneric, events: observation.events, requests: observation.requests, responses: observation.responses });
});

test('187 optional HTTP server dashboard Bombadil campaign', async ({ request }) => {
  test.skip(process.env.GOO_HTTP_DASHBOARD_BOMBADIL !== '1', 'Explicit ordinary OTel campaign');
  test.setTimeout(150000);
  const fs = require('node:fs/promises'), path = require('node:path');
  const { startBombadil } = require('./support/cross-app');
  const facts = await seed(request), dir = await write(facts, 'fixture.json', facts);
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'),
    url: base + '/#' + new URLSearchParams({ httpDashboard: facts.environment }),
    specification: 'verification/bombadil/http-server-dashboard.ts', seconds: 45 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const entries = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
  const states = entries.flatMap(entry => entry.snapshots.filter(snapshot => snapshot.name === 'state').map(snapshot => snapshot.value));
  const phases = [...new Set(states.map(state => state.phase))].sort((a, b) => a - b);
  expect(phases).toEqual(Array.from({ length: 9 }, (_, index) => index));
  expect(states.at(-1)).toMatchObject({ phase: 8, terminal: true, source: 17, plans: 3, proof: 'http:137|all:198|http:137' });
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2));
});
