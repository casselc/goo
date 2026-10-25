const { test, expect } = require('@playwright/test');
const { addFilter } = require('./support/explore');
const { profileLayout } = require('./190-profile-layout');
const { readStepLink } = require('../ui/workspace');
const { base, json, observe, loaded, snapshot, seed, findResult, assertPopulation, write } = require('./190-messaging-dashboard-support');
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

test('190 messaging roles preserve destination populations and continue to real trace evidence', async ({ page, request }) => {
  test.setTimeout(150000);
  const facts = await seed(request), observation = observe(page), catalog = await discover(page, facts);
  let pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / messaging-app', exact: true }).click();
  const plan = await json(await pending), dash = await loaded(page, 'messaging-consumer');
  await observation.settle();
  expect(plan.declarations).toEqual([]);
  const spanWidget = plan.widgets.find(widget => widget.key === 'spans'), methodWidget = plan.widgets.find(widget => widget.key === 'dimension');
  expect(spanWidget.grain).toBe('consumer-span-record');
  expect(spanWidget.query.where.all).toContainEqual({ field: { kind: 'column', key: 'kind' }, op: 'eq', value: 'Consumer' });
  for (const key of ['messaging.system']) expect(spanWidget.query.where.all).toContainEqual({ field: { kind: 'attribute', key }, op: 'ne' });
  expect(methodWidget.query.groupBy).toEqual([{ kind: 'attribute', key: 'messaging.destination.name' }]);
  expect(methodWidget.query.where).toEqual(spanWidget.query.where);
  const initial = { plan, spans: findResult(observation, spanWidget.query), methods: findResult(observation, methodWidget.query), view: await snapshot(page, observation) };
  assertPopulation(facts, initial.spans, 'messaging-consumer'); assertPopulation(facts, initial.methods, 'messaging-consumer');
  expect(initial.spans.count).toBe('137'); expect(initial.spans.rows).toHaveLength(100);
  expect(initial.spans.spanStatus.error).toBe('26');
  expect(initial.methods.groups.map(group => [group.g0, group.count]).sort()).toEqual([[null, '3'], ['', '2'], ['returns', '12'], ['orders', '120']].sort());
  const native = plan.widgets.filter(widget => widget.measurement);
  expect(native).toHaveLength(2);
  expect(native[0].measurement.definition.source.metric).toBe('messaging.process.duration');
  expect(native[0].measurement.definition.source.unit).toBe('s');
  expect(native[0].grain).not.toBe('consumer-span-record');
  await expect(dash.locator('[data-widget=spans]')).toContainText('Observed consumer-span records');
  expect(initial.view.rateTitle).toContain('137000/' + (plan.scope.toMs - plan.scope.fromMs));
  for (const widget of native) await expect(dash.locator('[data-widget=' + widget.key + '] > h5')).toHaveText(widget.title + ' [' + widget.measurement.definition.source.unit + ']');
  expect(initial.view.rateTitle).toContain('not unique messages or processing completions');
  observation.event('automatic-consumer');
  const plannerInput = observation.requests.find(value => value.endpoint === '/api/dashboards/plan').payload;
  const fallbackPlans = [];
  for (const removed of [['messaging.destination.name'], ['messaging.destination.name', 'messaging.operation.name']]) {
    const input = { ...structuredClone(plannerInput), profile: 'messaging-consumer', fields: plannerInput.fields.filter(field => field.kind !== 'attribute' || !removed.includes(field.key)) };
    const result = await json(await request.post(base + '/api/dashboards/plan', { data: input }));
    const widget = result.widgets.find(widget => widget.key === 'dimension');
    expect(widget.query.groupBy).toEqual([{ kind: 'attribute', key: removed.length === 1 ? 'messaging.operation.name' : 'messaging.system' }]);
    expect(widget.query.where).toEqual(spanWidget.query.where);
    expect(result.scope).toEqual(plan.scope);
    fallbackPlans.push({ effect: 'pure-metadata-plan', request: input, response: result });
  }


  const layout = {};
  await page.setViewportSize({ width: 1440, height: 1800 }); layout.desktop = await geometry(page); layout.desktopProfiles = await profileLayout(page);
  await page.getByRole('heading', { name: 'Services', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/190c-messaging-service-in-context.png' });
  await dash.screenshot({ path: 'docs/journeys/screenshots/190a-messaging-consumer-dashboard.png' });
  await page.setViewportSize({ width: 390, height: 1600 }); layout.mobile = await geometry(page); layout.mobileProfiles = await profileLayout(page, true);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.getByRole('heading', { name: 'Services', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/190d-mobile-messaging-in-context.png' });
  await dash.locator('[data-widget=spans]').screenshot({ path: 'docs/journeys/screenshots/190b-mobile-consumer-summary.png' });
  await page.setViewportSize({ width: 1440, height: 1100 });

  const producer = await profile(page, observation, 'messaging-producer');
  producer.result = findResult(observation, producer.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, producer.result, 'messaging-producer'); expect(producer.result.count).toBe('120');
  expect(producer.plan.scope).toEqual(plan.scope);
  expect(producer.plan.widgets.filter(widget => widget.measurement).every(widget => widget.family !== 'messaging-consumer')).toBe(true);
  observation.event('explicit-producer');
  const generic = await profile(page, observation, 'all-spans');
  expect(generic.plan.scope).toEqual(plan.scope);
  generic.result = findResult(observation, generic.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, generic.result, 'all-spans'); expect(generic.result.count).toBe('309');
  observation.event('explicit-all-spans');
  const returned = await profile(page, observation, 'messaging-consumer');
  expect(returned.plan.scope).toEqual(plan.scope);
  returned.result = findResult(observation, returned.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, returned.result, 'messaging-consumer'); observation.event('explicit-consumer-return');

  const lastConsumer = facts.records.filter(record => record.attributes['test.case'] === 'consumer-orders').at(-1);
  const range = { fromMs: lastConsumer.timeMs, toMs: lastConsumer.timeMs + 1 };
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
  await loaded(page, 'messaging-consumer'); await observation.settle();
  expect(rangeRequest.profile).toBe('messaging-consumer'); expect(rangeRequest.messagingConsumerObserved || false).toBe(false);
  expect(rangePlan.scope).toEqual({ ...plan.scope, ...range });
  const rangeResult = findResult(observation, rangePlan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, rangeResult, 'messaging-consumer'); expect(rangeResult.count).toBe('1'); observation.event('apply-consumer-range');
  await dash.locator('.service-timeline').getByText('Select a time range', { exact: true }).click();
  pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await dash.getByRole('button', { name: 'Reset overview window', exact: true }).click();
  const resetPlan = await json(await pending); await loaded(page, 'messaging-consumer'); await observation.settle();
  expect(resetPlan.scope).toEqual(plan.scope); observation.event('reset-consumer-window');

  // The actual dashboard → operation → records → waterfall journey is part of the primary acceptance.
  const beforeDestination = await snapshot(page, observation);
  const method = dash.locator('[data-widget=dimension] tr[data-diagnosis-value]').filter({ has: page.getByRole('cell', { name: '"orders"', exact: true }) });
  const link = method.getByRole('link', { name: 'Open consumer population', exact: true });
  const staged = readStepLink(await link.getAttribute('href'));
  expect(staged.record).toEqual({ ...methodWidget.query, where: { all: [methodWidget.query.where, { field: { kind: 'attribute', key: 'messaging.destination.name' }, op: 'eq', value: 'orders' }] }, offset: 0 });
  const beforeStage = observation.source;
  await dash.locator('[data-dashboard-profile=all-spans]').evaluate(button => { window.__messagingStaleProfile = button; });
  await link.click(); await expect(page.locator('#exploreRun')).toBeVisible();
  const plansBeforeStale = observation.plans;
  await page.evaluate(() => window.__messagingStaleProfile.click());
  expect(observation.source).toBe(beforeStage); expect(observation.plans).toBe(plansBeforeStale); observation.event('stage-method');
  pending = page.waitForResponse(response => response.url().endsWith('/api/explore'));
  await page.locator('#exploreRun').click(); const operation = await json(await pending);
  expect(operation.question).toEqual(staged.record);
  assertPopulation(facts, operation, 'messaging-consumer', 'messaging-app', { 'messaging.destination.name': 'orders' });
  expect(operation.count).toBe('120');
  expect([...new Set(operation.rows.map(row => row.Attributes['messaging.system']))].sort()).toEqual(['kafka', 'rabbitmq']);
  expect([...new Set(operation.rows.map(row => row.Attributes['messaging.consumer.group.name']))].sort()).toEqual(['workers-a', 'workers-b']); expect(operation.duration.sumNs).toBe('3600000000');
  expect(observation.source).toBe(beforeStage + 1); observation.event('run-method');
  const beforeFilters = observation.source;
  await addFilter(page, { kind: 'attribute', key: 'messaging.system' }, 'eq', 'kafka');
  await addFilter(page, { kind: 'attribute', key: 'messaging.consumer.group.name' }, 'eq', 'workers-a');
  expect(observation.source).toBe(beforeFilters); observation.event('stage-system-group-filters');
  pending = page.waitForResponse(value => value.url().endsWith('/api/explore'));
  await page.locator('#exploreRun').click(); const isolated = await json(await pending);
  assertPopulation(facts, isolated, 'messaging-consumer', 'messaging-app', { 'messaging.destination.name': 'orders', 'messaging.system': 'kafka', 'messaging.consumer.group.name': 'workers-a' });
  expect(isolated.count).toBe('108');
  expect(isolated.question).toMatchObject({ environment: plan.scope.environment, fromMs: plan.scope.fromMs, toMs: plan.scope.toMs });
  const leaves = predicate => predicate.all ? predicate.all.flatMap(leaves) : predicate.any ? predicate.any.flatMap(leaves) : [predicate];
  for (const leaf of leaves(staged.record.where)) expect(leaves(isolated.question.where)).toContainEqual(leaf);
  for (const [key, value] of [['messaging.system', 'kafka'], ['messaging.consumer.group.name', 'workers-a']]) expect(leaves(isolated.question.where)).toContainEqual({ field: { kind: 'attribute', key }, op: 'eq', value });
  expect(observation.source).toBe(beforeStage + 2); observation.event('run-isolated-consumer');
  if (await page.locator('#filterComposer').evaluate(element => element.open)) await page.locator('#filterComposer > summary').click();
  await page.getByRole('tab', { name: 'Records', exact: true }).click();
  expect(observation.source).toBe(beforeStage + 2);
  observation.event('show-isolated-records');
  const linkedFact = facts.records.find(record => record.id === facts.linkedConsumer);
  expect(isolated.rows[0].SpanId).toBe(linkedFact.id);
  const linkBefore = observation.source;
  await page.locator('#exploreRows .record-row').first().getByRole('button', { name: 'Inspect record', exact: true }).click();
  const inspector = page.locator('#exploreRows .ui-inspector');
  await inspector.locator('.raw-record > summary').click();
  const inspected = JSON.parse(await inspector.locator('.raw-record > pre').textContent());
  expect(inspected.linked_traces).toEqual(linkedFact.links.map(link => link.traceId));
  expect(inspected.linked_spans).toEqual(linkedFact.links.map(link => link.spanId));
  expect(inspected.linked_attributes).toEqual(linkedFact.links.map(link => link.attributes));
  await expect(inspector.locator('.raw-record > pre')).toContainText(linkedFact.links[0].traceId);
  await inspector.getByRole('button', { name: 'Close selected item details', exact: true }).click();
  expect(observation.source).toBe(linkBefore); observation.event('inspect-reported-link');
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
  const traceOrigin = JSON.parse(await page.locator('#exploreContextExplain').textContent());
  expect(traceOrigin.origin).toEqual(isolated.question);
  expect(trace.rows[0].linked_traces).toEqual(inspected.linked_traces);
  await expect(page.locator('#exploreTrace .waterfall-status:not(.waterfall-identity-note)')).toContainText('Async links are references, not parent edges.');
  expect(trace.count).toBe('1'); expect(logs.count).toBe('0'); expect(observation.source).toBe(beforeStage + 4); observation.event('expand-method-trace');
  await page.setViewportSize({ width: 1440, height: 1300 });
  await page.locator('#exploreContext').screenshot({ path: 'docs/journeys/screenshots/190e-consumer-population-to-trace.png' });
  await page.locator('#exploreContextClose').click();
  const returnHashes = [];
  for (let i = 0; i < 6 && new URL(page.url()).hash !== '#services'; i++) { await page.goBack(); returnHashes.push(new URL(page.url()).hash); }
  await expect(page).toHaveURL(base + '/#services'); await loaded(page, 'messaging-consumer');
  const afterDestination = await snapshot(page, observation); expect(afterDestination).toEqual(beforeDestination);
  expect(observation.source).toBe(beforeStage + 4); observation.event('return-held-service');

  await page.locator('#serviceDetail').getByRole('button', { name: 'Back to service list', exact: true }).click();
  pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / messaging-other-role', exact: true }).click();
  const outboundPlan = await json(await pending); await loaded(page, 'all-spans');
  const outbound = { plan: outboundPlan, view: await snapshot(page, observation) }; observation.event('outbound-default-generic');
  const emptyConsumer = await profile(page, observation, 'messaging-consumer');
  expect(emptyConsumer.view.profileState).toBe('no_data');
  expect(findResult(observation, emptyConsumer.plan.widgets.find(widget => widget.key === 'spans').query).count).toBe('0'); observation.event('other-role-explicit-empty-consumer');
  const recovered = await profile(page, observation, 'all-spans');
  const outboundResult = findResult(observation, recovered.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, outboundResult, 'all-spans', 'messaging-other-role'); expect(outboundResult.count).toBe('5'); observation.event('outbound-generic-recovery');
  await page.locator('#serviceDetail').getByRole('button', { name: 'Back to service list', exact: true }).click();
  pending = page.waitForResponse(value => value.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / hybrid', exact: true }).click();
  const hybridPlan = await json(await pending); await loaded(page, 'http-server'); await observation.settle();
  const hybridHTTP = findResult(observation, hybridPlan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, hybridHTTP, 'http-server', 'hybrid'); expect(hybridHTTP.count).toBe('3');
  observation.event('hybrid-auto-http');
  const hybridRPC = await profile(page, observation, 'rpc-server');
  const hybridRPCResult = findResult(observation, hybridRPC.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, hybridRPCResult, 'rpc-server', 'hybrid'); expect(hybridRPCResult.count).toBe('3');
  expect(hybridRPC.plan.scope).toEqual(hybridPlan.scope); observation.event('hybrid-explicit-rpc');
  const hybridDB = await profile(page, observation, 'db-client');
  const hybridDBResult = findResult(observation, hybridDB.plan.widgets.find(widget => widget.key === 'spans').query);
  assertPopulation(facts, hybridDBResult, 'db-client', 'hybrid'); expect(hybridDBResult.count).toBe('3');
  expect(hybridDB.plan.scope).toEqual(hybridPlan.scope); observation.event('hybrid-explicit-db');
  await observation.settle(); expect(observation.errors).toEqual([]);
  await write(facts, 'ledger.json', { ...facts, catalog, initial, fallbackPlans, layout, producer, generic, returned, supporting, rangeRequest, rangePlan, rangeResult, resetPlan,
    beforeDestination, staged, operation, isolated, inspected, traceOrigin, traceID, trace, logs, returnHashes, afterDestination, outbound, emptyConsumer, recovered, hybridPlan, hybridHTTP, hybridRPC, hybridRPCResult, hybridDB, hybridDBResult,
    events: observation.events, requests: observation.requests, responses: observation.responses });
});

test('190 optional bounded messaging dashboard campaign', async ({ request }) => {
  test.skip(process.env.GOO_MESSAGING_DASHBOARD_BOMBADIL !== '1', 'Explicit ordinary OTel campaign');
  test.setTimeout(120000);
  const fs = require('node:fs/promises'), path = require('node:path'), { startBombadil } = require('./support/cross-app');
  const facts = await seed(request), dir = await write(facts, 'fixture.json', facts);
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'), url: base + '/#' + new URLSearchParams({ messagingDashboard: facts.environment }), specification: 'verification/bombadil/messaging-dashboard.ts', seconds: 30 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const entries = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
  const states = entries.flatMap(entry => entry.snapshots.filter(value => value.name === 'state').map(value => value.value));
  const phases = [...new Set(states.map(value => value.phase))].sort((a, b) => a - b);
  expect(phases).toEqual(Array.from({ length: 11 }, (_, i) => i));
  expect(states.at(-1)).toMatchObject({ phase: 10, terminal: true, source: 25, plans: 5, proof: 'consumer:137|all:309|consumer:137|producer:120|consumer:137' });
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2));
});
