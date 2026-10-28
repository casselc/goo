const { test, expect } = require('@playwright/test');
const { readStepLink } = require('../ui/workspace.mjs');
const { base, json, observe, loaded, snapshot, seed, findResult, assertPopulation, write } = require('./188-rpc-dashboard-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });

async function profile(page, observation, key) {
  const start = observation.requests.length;
  const pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.locator('[data-dashboard-profile="' + key + '"]').click();
  const plan = await json(await pending);
  await loaded(page, key); await observation.settle();
  const requests = observation.requests.slice(start);
  expect(requests.filter(value => value.endpoint === '/api/dashboards/plan')).toHaveLength(1);
  expect(requests.filter(value => ['/api/discovery', '/api/inventory', '/api/metric-discovery'].includes(value.endpoint))).toEqual([]);
  expect(requests.filter(value => value.effect === 'source')).toHaveLength(plan.widgets.filter(widget => widget.state === 'supported').length);
  return { plan, view: await snapshot(page, observation), requests };
}
async function geometry(page) {
  const data = await page.locator('.service-dashboard').evaluate(host => ({
    routes: [...host.querySelectorAll('[data-widget=dimension] tbody tr')].map(row => ({
      text: row.textContent, bars: [...row.querySelectorAll('.comparison-bar-cell')].map(cell => ({ value: cell.dataset.value, maximum: cell.dataset.scaleMax,
        track: cell.querySelector('.comparison-bar-track').getBoundingClientRect().width, fill: cell.querySelector('.comparison-bar-fill').getBoundingClientRect().width })),
    })),
    lanes: [...host.querySelectorAll('.service-timeline-lane')].map(row => {
      const label = row.querySelector(':scope>div'), plot = row.querySelector('goo-value-timeline');
      const rect = element => { const box = element.getBoundingClientRect(); return { left: box.left, right: box.right }; };
      return { label: rect(label), plot: rect(plot), children: [...label.children].map(child => ({ ...rect(child), scroll: child.scrollWidth, width: child.clientWidth })) };
    }),
  }));
  expect(data.routes).toHaveLength(2);
  for (const row of data.routes) {
    expect(row.bars).toHaveLength(2);
    expect(row.bars[0].value).toBe(row.bars[1].value);
    expect(row.bars[0].maximum).toBe(row.bars[1].maximum);
    expect(row.bars[0].track).toBeGreaterThan(50);
    expect(Math.abs(row.bars[0].track - row.bars[1].track)).toBeLessThan(1);
    expect(Math.abs(row.bars[0].fill - row.bars[1].fill)).toBeLessThan(1);
  }
  expect(data.lanes.length).toBeGreaterThanOrEqual(5);
  for (const lane of data.lanes) {
    expect(lane.label.right).toBeLessThanOrEqual(lane.plot.left + 1);
    for (const child of lane.children) {
      expect(child.left).toBeGreaterThanOrEqual(lane.label.left - 1);
      expect(child.right).toBeLessThanOrEqual(lane.label.right + 1);
      expect(child.scroll).toBeLessThanOrEqual(child.width + 1);
    }
  }
  return data;
}
async function discover(page, facts) {
  await page.goto(base + '/#services');
  await page.locator('#inventoryEnvironment').fill(facts.environment);
  await page.locator('#minutes').fill('2');
  const pending = page.waitForResponse(response => response.url().endsWith('/api/discovery'));
  await page.locator('#inventoryRun').click(); return json(await pending);
}

test('188 incoming legacy RPC defaults preserve method populations and continue to real trace evidence', async ({ page, request }) => {
  test.setTimeout(150000);
  const facts = await seed(request), observation = observe(page), catalog = await discover(page, facts);
  let pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / rpc-api', exact: true }).click();
  const plan = await json(await pending), dash = await loaded(page, 'rpc-server');
  await observation.settle();
  expect(plan.declarations).toEqual([]);
  const spanWidget = plan.widgets.find(widget => widget.key === 'spans'), methodWidget = plan.widgets.find(widget => widget.key === 'dimension');
  expect(spanWidget.grain).toBe('server-span-record');
  expect(spanWidget.query.where.all).toContainEqual({ field: { kind: 'column', key: 'kind' }, op: 'eq', value: 'Server' });
  for (const key of ['rpc.system', 'rpc.method']) expect(spanWidget.query.where.all).toContainEqual({ field: { kind: 'attribute', key }, op: 'ne' });
  expect(methodWidget.query.groupBy).toEqual([{ kind: 'attribute', key: 'rpc.method' }]);
  expect(methodWidget.query.where).toEqual(spanWidget.query.where);
  const initial = { plan, spans: findResult(observation, spanWidget.query), methods: findResult(observation, methodWidget.query), view: await snapshot(page, observation) };
  assertPopulation(facts, initial.spans, 'rpc-server'); assertPopulation(facts, initial.methods, 'rpc-server');
  expect(initial.spans.count).toBe('132'); expect(initial.spans.rows).toBeUndefined();
  expect(initial.spans.spanStatus.error).toBe('24');
  expect(initial.methods.groups.map(group => [group.g0, group.count]).sort()).toEqual([['CustomCall', '12'], ['catalog.Inventory/GetItem', '120']].sort());
  const native = plan.widgets.filter(widget => widget.measurement);
  expect(native).toHaveLength(2);
  expect(native[0].measurement.definition.source.metric).toBe('rpc.server.duration');
  expect(native[0].measurement.definition.source.unit).toBe('ms');
  expect(native[0].grain).not.toBe('server-span-record');
  await expect(dash.locator('[data-widget=spans]')).toContainText('Observed server-span records');
  expect(initial.view.rateTitle).toContain('132000/' + (plan.scope.toMs - plan.scope.fromMs));
  observation.event('automatic-rpc');

  const layout = {};
  await page.setViewportSize({ width: 1440, height: 1800 }); layout.desktop = await geometry(page);
  await page.getByRole('heading', { name: 'Services', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/188c-rpc-service-in-context.png' });
  await dash.screenshot({ path: 'docs/journeys/screenshots/188a-incoming-rpc-dashboard.png' });
  await page.setViewportSize({ width: 390, height: 1600 }); layout.mobile = await geometry(page);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.getByRole('heading', { name: 'Services', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/188d-mobile-rpc-in-context.png' });
  await dash.locator('[data-widget=spans]').screenshot({ path: 'docs/journeys/screenshots/188b-mobile-rpc-summary.png' });
  await page.setViewportSize({ width: 1440, height: 1100 });

  const generic = await profile(page, observation, 'all-spans');
  expect(generic.plan.scope).toEqual(plan.scope);
  generic.result = findResult(observation, generic.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, generic.result, 'all-spans'); expect(generic.result.count).toBe('203');
  observation.event('explicit-all-spans');
  const returned = await profile(page, observation, 'rpc-server');
  expect(returned.plan.scope).toEqual(plan.scope);
  returned.result = findResult(observation, returned.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, returned.result, 'rpc-server'); observation.event('explicit-rpc-return');

  const lastServer = facts.records.filter(record => record.attributes['test.case'] === 'incoming-method').at(-1);
  const range = { fromMs: lastServer.timeMs, toMs: lastServer.timeMs + 1 };
  const supporting = observation.responses.find(value => value.endpoint === '/api/explore' && value.request.signal === 'spans' && !value.request.groupBy)?.body;
  expect(supporting.rows).toHaveLength(100);
  expect(supporting.rows.some(row => BigInt(row.timestamp_ns) >= BigInt(range.fromMs) * 1000000n && BigInt(row.timestamp_ns) < BigInt(range.toMs) * 1000000n)).toBe(false);
  const beforeRange = observation.source;
  await dash.locator('.service-timeline').getByText('Select a time range', { exact: true }).click();
  await dash.locator('goo-time-range').evaluate((control, range) => {
    const start = control.querySelectorAll('input[type=range]')[0]; start.value = String(range.fromMs); start.dispatchEvent(new Event('input', { bubbles: true }));
    const end = control.querySelectorAll('input[type=range]')[1]; end.value = String(range.toMs); end.dispatchEvent(new Event('input', { bubbles: true }));
  }, range);
  expect(observation.source).toBe(beforeRange);
  expect(await dash.locator('.service-timeline').evaluate(element => ({ fromMs: Number(element.dataset.fromMs), toMs: Number(element.dataset.toMs) }))).toEqual(range);
  pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await dash.getByRole('button', { name: 'Apply range to overview', exact: true }).click();
  const response = await pending, rangeRequest = response.request().postDataJSON(), rangePlan = await json(response);
  await loaded(page, 'rpc-server'); await observation.settle();
  expect(rangeRequest.profile).toBe('rpc-server'); expect(rangeRequest.rpcServerObserved || false).toBe(false);
  expect(rangePlan.scope).toEqual({ ...plan.scope, ...range });
  const rangeResult = findResult(observation, rangePlan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, rangeResult, 'rpc-server'); expect(rangeResult.count).toBe('1'); observation.event('apply-rpc-range');
  await dash.locator('.service-timeline').getByText('Select a time range', { exact: true }).click();
  pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await dash.getByRole('button', { name: 'Reset overview window', exact: true }).click();
  const resetPlan = await json(await pending); await loaded(page, 'rpc-server'); await observation.settle();
  expect(resetPlan.scope).toEqual(plan.scope); observation.event('reset-rpc-window');

  // The actual dashboard → method → records → waterfall journey is part of the primary acceptance.
  const beforeMethod = await snapshot(page, observation);
  const method = dash.locator('[data-widget=dimension] tr[data-diagnosis-value]').filter({ has: page.getByRole('cell', { name: '"catalog.Inventory/GetItem"', exact: true }) });
  const link = method.getByRole('link', { name: 'Open method population', exact: true });
  const staged = readStepLink(await link.getAttribute('href'));
  expect(staged.record).toEqual({ ...methodWidget.query, where: { all: [methodWidget.query.where, { field: { kind: 'attribute', key: 'rpc.method' }, op: 'eq', value: 'catalog.Inventory/GetItem' }] }, offset: 0 });
  const beforeStage = observation.source;
  await dash.locator('[data-dashboard-profile=all-spans]').evaluate(button => { window.__rpcStaleProfile = button; });
  await link.click(); await expect(page.locator('#exploreRun')).toBeVisible();
  const plansBeforeStale = observation.plans;
  await page.evaluate(() => window.__rpcStaleProfile.click());
  expect(observation.source).toBe(beforeStage); expect(observation.plans).toBe(plansBeforeStale); observation.event('stage-method');
  pending = page.waitForResponse(response => response.url().endsWith('/api/explore'));
  await page.locator('#exploreRun').click(); const operation = await json(await pending);
  expect(operation.question).toEqual(staged.record);
  assertPopulation(facts, operation, 'rpc-server', 'rpc-api', 'catalog.Inventory/GetItem');
  expect(operation.count).toBe('120');
  expect([...new Set(operation.rows.map(row => row.Attributes['rpc.system']))].sort()).toEqual(['custom.protocol', 'grpc']);
  expect([...new Set(operation.rows.map(row => row.Attributes['rpc.service']))].sort()).toEqual(['catalog.AlternateInventory', 'catalog.Inventory']); expect(operation.duration.sumNs).toBe('3000000000');
  expect(observation.source).toBe(beforeStage + 1); observation.event('run-method');
  const traceID = operation.rows[0].TraceId;
  const traceReading = page.waitForResponse(value => value.url().endsWith('/api/explore') && value.request().postDataJSON().signal === 'spans');
  const logsReading = page.waitForResponse(value => value.url().endsWith('/api/explore') && value.request().postDataJSON().signal === 'logs');
  await page.getByRole('button', { name: 'Expand trace ' + traceID, exact: true }).first().click();
  const trace = await json(await traceReading), logs = await json(await logsReading);
  await expect(page.locator('#exploreTrace .waterfall-row')).toHaveCount(1);
  await expect(page.locator('#exploreLogStatus')).toContainText('captured logs');
  for (const result of [trace, logs]) {
    expect(result.question).toMatchObject({ environment: plan.scope.environment, fromMs: plan.scope.fromMs, toMs: plan.scope.toMs, where: { all: [{ field: { kind: 'column', key: 'trace_id' }, op: 'eq', value: traceID }] } });
  }
  expect(trace.count).toBe('1'); expect(logs.count).toBe('0'); expect(observation.source).toBe(beforeStage + 3); observation.event('expand-method-trace');
  await page.setViewportSize({ width: 1440, height: 1300 });
  await page.locator('#exploreContext').screenshot({ path: 'docs/journeys/screenshots/188e-rpc-method-to-trace.png' });
  await page.locator('#exploreContextClose').click();
  const returnHashes = [];
  for (let i = 0; i < 3 && new URL(page.url()).hash !== '#services'; i++) { await page.goBack(); returnHashes.push(new URL(page.url()).hash); }
  await expect(page).toHaveURL(base + '/#services'); await loaded(page, 'rpc-server');
  const afterMethod = await snapshot(page, observation); expect(afterMethod).toEqual(beforeMethod);
  expect(observation.source).toBe(beforeStage + 3); observation.event('return-held-service');

  await page.locator('#serviceDetail').getByRole('button', { name: 'Back to service list', exact: true }).click();
  pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / rpc-outbound', exact: true }).click();
  const outboundPlan = await json(await pending); await loaded(page, 'all-spans');
  const outbound = { plan: outboundPlan, view: await snapshot(page, observation) }; observation.event('outbound-default-generic');
  const emptyRPC = await profile(page, observation, 'rpc-server');
  expect(emptyRPC.view.profileState).toBe('no_data');
  expect(findResult(observation, emptyRPC.plan.widgets.find(widget => widget.key === 'spans').query).count).toBe('0'); observation.event('outbound-explicit-empty-rpc');
  const recovered = await profile(page, observation, 'all-spans');
  const outboundResult = findResult(observation, recovered.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, outboundResult, 'all-spans', 'rpc-outbound'); expect(outboundResult.count).toBe('7'); observation.event('outbound-generic-recovery');
  await observation.settle(); expect(observation.errors).toEqual([]);
  await write(facts, 'ledger.json', { ...facts, catalog, initial, layout, generic, returned, supporting, rangeRequest, rangePlan, rangeResult, resetPlan,
    beforeMethod, staged, operation, traceID, trace, logs, returnHashes, afterMethod, outbound, emptyRPC, recovered,
    events: observation.events, requests: observation.requests, responses: observation.responses });
});

test('188 optional bounded RPC dashboard campaign', async ({ request }) => {
  test.skip(process.env.GOO_RPC_DASHBOARD_BOMBADIL !== '1', 'Explicit ordinary OTel campaign');
  test.setTimeout(120000);
  const fs = require('node:fs/promises'), path = require('node:path'), { startBombadil } = require('./support/cross-app');
  const facts = await seed(request), dir = await write(facts, 'fixture.json', facts);
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'), url: base + '/#' + new URLSearchParams({ rpcDashboard: facts.environment }), specification: 'verification/bombadil/rpc-server-dashboard.ts', seconds: 30 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const entries = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
  const states = entries.flatMap(entry => entry.snapshots.filter(value => value.name === 'state').map(value => value.value));
  const phases = [...new Set(states.map(value => value.phase))].sort((a, b) => a - b);
  expect(phases).toEqual(Array.from({ length: 9 }, (_, i) => i));
  expect(states.at(-1)).toMatchObject({ phase: 8, terminal: true, source: 17, plans: 3, proof: 'rpc:132|all:203|rpc:132' });
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2));
});
