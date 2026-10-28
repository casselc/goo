const { test, expect } = require('@playwright/test');
const { addFilter } = require('./support/explore');
const { readStepLink } = require('../ui/workspace.mjs');
const { base, json, observe, loaded, snapshot, seed, findResult, assertPopulation, write } = require('./189-db-dashboard-support');
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
  expect(data.routes).toHaveLength(4);
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

test('189 current-key database client defaults preserve operation populations and continue to real trace evidence', async ({ page, request }) => {
  test.setTimeout(150000);
  const facts = await seed(request), observation = observe(page), catalog = await discover(page, facts);
  let pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / db-api', exact: true }).click();
  const plan = await json(await pending), dash = await loaded(page, 'db-client');
  await observation.settle();
  expect(plan.declarations).toEqual([]);
  const spanWidget = plan.widgets.find(widget => widget.key === 'spans'), methodWidget = plan.widgets.find(widget => widget.key === 'dimension');
  expect(spanWidget.grain).toBe('client-span-record');
  expect(spanWidget.query.where.all).toContainEqual({ field: { kind: 'column', key: 'kind' }, op: 'eq', value: 'Client' });
  for (const key of ['db.system.name']) expect(spanWidget.query.where.all).toContainEqual({ field: { kind: 'attribute', key }, op: 'ne' });
  expect(methodWidget.query.groupBy).toEqual([{ kind: 'attribute', key: 'db.operation.name' }]);
  expect(methodWidget.query.where).toEqual(spanWidget.query.where);
  const initial = { plan, spans: findResult(observation, spanWidget.query), methods: findResult(observation, methodWidget.query), view: await snapshot(page, observation) };
  assertPopulation(facts, initial.spans, 'db-client'); assertPopulation(facts, initial.methods, 'db-client');
  expect(initial.spans.count).toBe('137'); expect(initial.spans.rows).toBeUndefined();
  expect(initial.spans.spanStatus.error).toBe('26');
  expect(initial.methods.groups.map(group => [group.g0, group.count]).sort()).toEqual([[null, '3'], ['', '2'], ['FETCH', '12'], ['SELECT', '120']].sort());
  const native = plan.widgets.filter(widget => widget.measurement);
  expect(native).toHaveLength(2);
  expect(native[0].measurement.definition.source.metric).toBe('db.client.operation.duration');
  expect(native[0].measurement.definition.source.unit).toBe('s');
  expect(native[0].grain).not.toBe('client-span-record');
  await expect(dash.locator('[data-widget=spans]')).toContainText('Observed client-span records');
  expect(initial.view.rateTitle).toContain('137000/' + (plan.scope.toMs - plan.scope.fromMs));
  for (const widget of native) await expect(dash.locator('[data-widget=' + widget.key + '] > h5')).toHaveText(widget.title + ' [' + widget.measurement.definition.source.unit + ']');
  expect(await dash.innerText()).not.toContain('RAW_QUERY_NOT_A_DEFAULT_189');
  expect(await dash.innerText()).not.toContain('SUPPLIED_SUMMARY_NOT_A_DEFAULT_189');
  observation.event('automatic-db');
  const plannerInput = observation.requests.find(value => value.endpoint === '/api/dashboards/plan').payload;
  const fallbackPlans = [];
  for (const removed of [['db.operation.name'], ['db.operation.name', 'db.namespace']]) {
    const input = { ...structuredClone(plannerInput), profile: 'db-client', fields: plannerInput.fields.filter(field => field.kind !== 'attribute' || !removed.includes(field.key)) };
    const result = await json(await request.post(base + '/api/dashboards/plan', { data: input }));
    const widget = result.widgets.find(widget => widget.key === 'dimension');
    expect(widget.query.groupBy).toEqual([{ kind: 'attribute', key: removed.length === 1 ? 'db.namespace' : 'db.system.name' }]);
    expect(widget.query.where).toEqual(spanWidget.query.where);
    expect(result.scope).toEqual(plan.scope);
    expect(JSON.stringify(result.widgets.map(widget => widget.query?.groupBy))).not.toMatch(/db\.query\.(text|summary)/);
    fallbackPlans.push({ effect: 'pure-metadata-plan', request: input, response: result });
  }


  const layout = {};
  await page.setViewportSize({ width: 1440, height: 1800 }); layout.desktop = await geometry(page);
  await page.getByRole('heading', { name: 'Services', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/189c-db-service-in-context.png' });
  await dash.screenshot({ path: 'docs/journeys/screenshots/189a-db-client-dashboard.png' });
  await page.setViewportSize({ width: 390, height: 1600 }); layout.mobile = await geometry(page);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.getByRole('heading', { name: 'Services', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/189d-mobile-db-in-context.png' });
  await dash.locator('[data-widget=spans]').screenshot({ path: 'docs/journeys/screenshots/189b-mobile-db-summary.png' });
  await page.setViewportSize({ width: 1440, height: 1100 });

  const generic = await profile(page, observation, 'all-spans');
  expect(generic.plan.scope).toEqual(plan.scope);
  generic.result = findResult(observation, generic.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, generic.result, 'all-spans'); expect(generic.result.count).toBe('189');
  observation.event('explicit-all-spans');
  const returned = await profile(page, observation, 'db-client');
  expect(returned.plan.scope).toEqual(plan.scope);
  returned.result = findResult(observation, returned.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, returned.result, 'db-client'); observation.event('explicit-db-return');

  const lastClient = facts.records.filter(record => record.attributes['test.case'] === 'client-select').at(-1);
  const range = { fromMs: lastClient.timeMs, toMs: lastClient.timeMs + 1 };
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
  await loaded(page, 'db-client'); await observation.settle();
  expect(rangeRequest.profile).toBe('db-client'); expect(rangeRequest.dbClientObserved || false).toBe(false);
  expect(rangePlan.scope).toEqual({ ...plan.scope, ...range });
  const rangeResult = findResult(observation, rangePlan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, rangeResult, 'db-client'); expect(rangeResult.count).toBe('1'); observation.event('apply-db-range');
  await dash.locator('.service-timeline').getByText('Select a time range', { exact: true }).click();
  pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await dash.getByRole('button', { name: 'Reset overview window', exact: true }).click();
  const resetPlan = await json(await pending); await loaded(page, 'db-client'); await observation.settle();
  expect(resetPlan.scope).toEqual(plan.scope); observation.event('reset-db-window');

  // The actual dashboard → operation → records → waterfall journey is part of the primary acceptance.
  const beforeMethod = await snapshot(page, observation);
  const method = dash.locator('[data-widget=dimension] tr[data-diagnosis-value]').filter({ has: page.getByRole('cell', { name: '"SELECT"', exact: true }) });
  const link = method.getByRole('link', { name: 'Open client population', exact: true });
  const staged = readStepLink(await link.getAttribute('href'));
  expect(staged.record).toEqual({ ...methodWidget.query, where: { all: [methodWidget.query.where, { field: { kind: 'attribute', key: 'db.operation.name' }, op: 'eq', value: 'SELECT' }] }, offset: 0 });
  const beforeStage = observation.source;
  await dash.locator('[data-dashboard-profile=all-spans]').evaluate(button => { window.__dbStaleProfile = button; });
  await link.click(); await expect(page.locator('#exploreRun')).toBeVisible();
  const plansBeforeStale = observation.plans;
  await page.evaluate(() => window.__dbStaleProfile.click());
  expect(observation.source).toBe(beforeStage); expect(observation.plans).toBe(plansBeforeStale); observation.event('stage-method');
  pending = page.waitForResponse(response => response.url().endsWith('/api/explore'));
  await page.locator('#exploreRun').click(); const operation = await json(await pending);
  expect(operation.question).toEqual(staged.record);
  assertPopulation(facts, operation, 'db-client', 'db-api', { 'db.operation.name': 'SELECT' });
  expect(operation.count).toBe('120');
  expect([...new Set(operation.rows.map(row => row.Attributes['db.system.name']))].sort()).toEqual(['mysql', 'postgresql']);
  expect([...new Set(operation.rows.map(row => row.Attributes['db.namespace']))].sort()).toEqual(['archive', 'catalog']); expect(operation.duration.sumNs).toBe('3600000000');
  expect(observation.source).toBe(beforeStage + 1); observation.event('run-method');
  const beforeFilters = observation.source;
  await addFilter(page, { kind: 'attribute', key: 'db.system.name' }, 'eq', 'postgresql');
  await addFilter(page, { kind: 'attribute', key: 'db.namespace' }, 'eq', 'catalog');
  expect(observation.source).toBe(beforeFilters); observation.event('stage-database-filters');
  pending = page.waitForResponse(value => value.url().endsWith('/api/explore'));
  await page.locator('#exploreRun').click(); const isolated = await json(await pending);
  assertPopulation(facts, isolated, 'db-client', 'db-api', { 'db.operation.name': 'SELECT', 'db.system.name': 'postgresql', 'db.namespace': 'catalog' });
  expect(isolated.count).toBe('108');
  expect(isolated.question).toMatchObject({ environment: plan.scope.environment, fromMs: plan.scope.fromMs, toMs: plan.scope.toMs });
  const leaves = predicate => predicate.all ? predicate.all.flatMap(leaves) : predicate.any ? predicate.any.flatMap(leaves) : [predicate];
  for (const leaf of leaves(staged.record.where)) expect(leaves(isolated.question.where)).toContainEqual(leaf);
  for (const [key, value] of [['db.system.name', 'postgresql'], ['db.namespace', 'catalog']]) expect(leaves(isolated.question.where)).toContainEqual({ field: { kind: 'attribute', key }, op: 'eq', value });
  expect(observation.source).toBe(beforeStage + 2); observation.event('run-isolated-database');
  if (await page.locator('#filterComposer').evaluate(element => element.open)) await page.locator('#filterComposer > summary').click();
  await page.getByRole('tab', { name: 'Records', exact: true }).click();
  expect(observation.source).toBe(beforeStage + 2);
  observation.event('show-isolated-records');
  const traceID = isolated.rows[0].TraceId;
  const traceReading = page.waitForResponse(value => value.url().endsWith('/api/explore') && value.request().postDataJSON().signal === 'spans');
  const logsReading = page.waitForResponse(value => value.url().endsWith('/api/explore') && value.request().postDataJSON().signal === 'logs');
  await page.getByRole('button', { name: 'Expand trace ' + traceID, exact: true }).first().click();
  const trace = await json(await traceReading), logs = await json(await logsReading);
  await expect(page.locator('#exploreTrace .waterfall-row')).toHaveCount(1);
  await expect(page.locator('#exploreLogStatus')).toContainText('captured logs');
  for (const result of [trace, logs]) {
    expect(result.question).toMatchObject({ environment: plan.scope.environment, fromMs: plan.scope.fromMs, toMs: plan.scope.toMs, where: { all: [{ field: { kind: 'column', key: 'trace_id' }, op: 'eq', value: traceID }] } });
  }
  expect(trace.count).toBe('1'); expect(logs.count).toBe('0'); expect(observation.source).toBe(beforeStage + 4); observation.event('expand-method-trace');
  await page.setViewportSize({ width: 1440, height: 1300 });
  await page.locator('#exploreContext').screenshot({ path: 'docs/journeys/screenshots/189e-db-population-to-trace.png' });
  await page.locator('#exploreContextClose').click();
  const returnHashes = [];
  for (let i = 0; i < 6 && new URL(page.url()).hash !== '#services'; i++) { await page.goBack(); returnHashes.push(new URL(page.url()).hash); }
  await expect(page).toHaveURL(base + '/#services'); await loaded(page, 'db-client');
  const afterMethod = await snapshot(page, observation); expect(afterMethod).toEqual(beforeMethod);
  expect(observation.source).toBe(beforeStage + 4); observation.event('return-held-service');

  await page.locator('#serviceDetail').getByRole('button', { name: 'Back to service list', exact: true }).click();
  pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / db-other-role', exact: true }).click();
  const outboundPlan = await json(await pending); await loaded(page, 'all-spans');
  const outbound = { plan: outboundPlan, view: await snapshot(page, observation) }; observation.event('outbound-default-generic');
  const emptyDB = await profile(page, observation, 'db-client');
  expect(emptyDB.view.profileState).toBe('no_data');
  expect(findResult(observation, emptyDB.plan.widgets.find(widget => widget.key === 'spans').query).count).toBe('0'); observation.event('other-role-explicit-empty-db');
  const recovered = await profile(page, observation, 'all-spans');
  const outboundResult = findResult(observation, recovered.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, outboundResult, 'all-spans', 'db-other-role'); expect(outboundResult.count).toBe('5'); observation.event('outbound-generic-recovery');
  await page.locator('#serviceDetail').getByRole('button', { name: 'Back to service list', exact: true }).click();
  pending = page.waitForResponse(value => value.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / hybrid', exact: true }).click();
  const hybridPlan = await json(await pending); await loaded(page, 'http-server'); await observation.settle();
  const hybridHTTP = findResult(observation, hybridPlan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, hybridHTTP, 'http-server', 'hybrid'); expect(hybridHTTP.count).toBe('5');
  observation.event('hybrid-auto-http');
  const hybridRPC = await profile(page, observation, 'rpc-server');
  const hybridRPCResult = findResult(observation, hybridRPC.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, hybridRPCResult, 'rpc-server', 'hybrid'); expect(hybridRPCResult.count).toBe('5');
  expect(hybridRPC.plan.scope).toEqual(hybridPlan.scope); observation.event('hybrid-explicit-rpc');
  const hybridDB = await profile(page, observation, 'db-client');
  const hybridDBResult = findResult(observation, hybridDB.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, hybridDBResult, 'db-client', 'hybrid'); expect(hybridDBResult.count).toBe('5');
  expect(hybridDB.plan.scope).toEqual(hybridPlan.scope); observation.event('hybrid-explicit-db');
  await observation.settle(); expect(observation.errors).toEqual([]);
  await write(facts, 'ledger.json', { ...facts, catalog, initial, fallbackPlans, layout, generic, returned, supporting, rangeRequest, rangePlan, rangeResult, resetPlan,
    beforeMethod, staged, operation, isolated, traceID, trace, logs, returnHashes, afterMethod, outbound, emptyDB, recovered, hybridPlan, hybridHTTP, hybridRPC, hybridRPCResult, hybridDB, hybridDBResult,
    events: observation.events, requests: observation.requests, responses: observation.responses });
});

test('189 optional bounded database client dashboard campaign', async ({ request }) => {
  test.skip(process.env.GOO_DB_DASHBOARD_BOMBADIL !== '1', 'Explicit ordinary OTel campaign');
  test.setTimeout(120000);
  const fs = require('node:fs/promises'), path = require('node:path'), { startBombadil } = require('./support/cross-app');
  const facts = await seed(request), dir = await write(facts, 'fixture.json', facts);
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'), url: base + '/#' + new URLSearchParams({ dbDashboard: facts.environment }), specification: 'verification/bombadil/db-client-dashboard.ts', seconds: 30 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const entries = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
  const states = entries.flatMap(entry => entry.snapshots.filter(value => value.name === 'state').map(value => value.value));
  const phases = [...new Set(states.map(value => value.phase))].sort((a, b) => a - b);
  expect(phases).toEqual(Array.from({ length: 9 }, (_, i) => i));
  expect(states.at(-1)).toMatchObject({ phase: 8, terminal: true, source: 17, plans: 3, proof: 'db:137|all:189|db:137' });
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2));
});
