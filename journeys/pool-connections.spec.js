const { test, expect } = require('@playwright/test');
const { readStepLink } = require('../ui/workspace.mjs');
const { base, json, observe, loaded, seed, write, captureBytes, metadataControlRequest } = require('./196-pool-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
async function geometry(page, host) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  const rows = await host.locator('.pool-connection-row').evaluateAll(elements => elements.map(e => { const r = e.getBoundingClientRect(); return { pool: e.dataset.pool, instance: e.dataset.instance, left: r.left, right: r.right, width: r.width, scroll: e.scrollWidth, client: e.clientWidth }; }));
  for (const row of rows) { expect(row.left).toBeGreaterThanOrEqual(0); expect(row.right).toBeLessThanOrEqual(await page.evaluate(() => innerWidth)); expect(row.scroll).toBeLessThanOrEqual(row.client + 1); }
  return rows;
}
test('196 named pool counts auto assemble exact states without totals or health inference', async ({ page, request }) => {
  test.setTimeout(150000); const facts = await seed(request), observation = observe(page), events = [];
  const event = action => events.push({ action, source: observation.source, plans: observation.plans });
  await page.goto(base + '/#services'); await page.locator('#inventoryEnvironment').fill(facts.environment); await page.locator('#minutes').fill('2');
  let pending = page.waitForResponse(r => r.url().endsWith('/api/discovery')); await page.locator('#inventoryRun').click(); const catalog = await json(await pending);
  pending = page.waitForResponse(r => r.url().endsWith('/api/dashboards/plan')); await page.getByRole('button', { name: 'Open service inventory / pool-api', exact: true }).click();
  const planResponse = await pending, planRequest = planResponse.request().postDataJSON(), plan = await json(planResponse);
  const dashboard = await loaded(page, 'http-server'); await observation.settle(); event('overview');
  const nativeWidgets = plan.widgets.filter(w => w.measurement && w.state === 'supported'); expect(nativeWidgets).toHaveLength(2);
  expect(nativeWidgets[0].measurement.definition.source.metric).toBe('http.server.request.duration');
  const widget = nativeWidgets.find(w => w.family === 'database-pool-connections'); expect(widget).toBeTruthy();
  expect(widget.measurement.definition.stockDomain).toBe('nonnegative-integer');
  expect(plan.families.some(f => f.key === 'database-pool-connections')).toBe(false);
  const response = observation.responses.find(r => r.endpoint === '/api/measurements' && r.body?.question.definition.stockDomain === 'nonnegative-integer');
  expect(response).toBeTruthy(); const widgetResult = response.body; expect(widgetResult.stocks.series).toHaveLength(10);
  const box = dashboard.locator('[data-widget="' + widget.key + '"]'), host = box.locator('.pool-connections'); await expect(host).toBeVisible();
  const poolModel = await page.evaluate(async (result)=>{const GooPoolConnections=await import('/pool-connections.mjs');return GooPoolConnections.model(result)}, widgetResult);
  expect(poolModel.ineligible).toHaveLength(2); expect(poolModel.groups).toHaveLength(5);
  const ordersA = host.locator('.pool-connection-row[data-pool="orders"][data-instance="worker-a"]');
  await expect(ordersA.locator('[data-pool-state=used]')).toHaveAttribute('data-value', '3');
  await expect(ordersA.locator('[data-pool-state=idle]')).toHaveAttribute('data-value', '12');
  await expect(ordersA.locator('[data-pool-state=reserved]')).toHaveAttribute('data-value', '2');
  await expect(host.locator('[data-pool="orders"][data-instance="worker-b"] [data-pool-state=used]')).toHaveAttribute('data-value', '21');
  await expect(host.locator('[data-pool="orders"][data-instance="worker-b"] [data-pool-state=idle]')).toHaveAttribute('data-value', '0');
  await expect(host.locator('[data-pool=archive] [data-pool-state=used]')).toHaveAttribute('data-state', 'stale');
  for (const pool of ['missing-report', 'conflicting']) { const cell = host.locator('[data-pool="' + pool + '"] [data-pool-state=used]'); await expect(cell).toHaveAttribute('data-state', 'unknown'); await expect(cell).toHaveAttribute('data-value', ''); }
  const beforeLocal = observation.source; await host.getByLabel('Find named pool', { exact: true }).fill('orders');
  await expect(host.locator('.pool-connection-row')).toHaveCount(2); await host.getByLabel('Pool connection state', { exact: true }).selectOption('state:reserved');
  await expect(ordersA.locator('[data-pool-state=reserved]')).toHaveAttribute('data-value', '2');
  await host.getByLabel('Pool connection state', { exact: true }).selectOption('all'); await host.getByLabel('Find named pool', { exact: true }).fill('');
  await ordersA.getByText('Observation times and exact streams', { exact: true }).click();
  await expect(ordersA).toContainText('epoch'); expect(observation.source).toBe(beforeLocal); event('local-filter-disclosure');
  const domRows = await host.locator('.pool-connection-row').evaluateAll(rows => rows.map(e => ({ pool: e.dataset.pool, instance: e.dataset.instance, text: e.textContent, cells: [...e.querySelectorAll('[data-pool-state]')].map(c => ({ ...c.dataset, text: c.textContent })) })));
  const desktop = await geometry(page, host); await host.screenshot({ path: 'docs/journeys/screenshots/196a-named-pool-counts.png' });
  await page.setViewportSize({ width: 390, height: 1800 }); const mobile = await geometry(page, host); await host.screenshot({ path: 'docs/journeys/screenshots/196b-mobile-pool-states.png' });
  expect(observation.source).toBe(beforeLocal); await page.setViewportSize({ width: 1440, height: 1200 });
  const openLink = box.getByRole('link', { name: 'Open widget question', exact: true }), staged = readStepLink(await openLink.getAttribute('href'));
  await openLink.click(); await expect(page.locator('#measureRun')).toBeEnabled(); expect(observation.source).toBe(beforeLocal); event('native-staged');
  pending = page.waitForResponse(r => r.url().endsWith('/api/measurements') && r.request().method() === 'POST'); await page.locator('#measureRun').click();
  const native = await json(await pending); expect(native.question).toEqual(widget.measurement); expect(native.stocks).toEqual(widgetResult.stocks); event('native-run');
  await page.locator('#measureCapture').click(); await page.locator('#captureDraftTitle').fill('Pinned named pool observations');
  pending = page.waitForResponse(r => r.url().endsWith('/api/captures') && r.request().method() === 'POST'); await page.locator('#captureDraftConfirm').click(); const saved = await json(await pending);
  await expect(page.locator('#captureMeasurementEvidence')).toBeVisible(); expect(saved.measurement.question.definition.stockDomain).toBe('nonnegative-integer');
  const storedBefore = await captureBytes(facts, saved.receipt.id, 'capture-before.json'); event('capture-fresh');
  const beforeReopen = observation.source; await page.reload(); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible();
  const reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id)); expect(reopened).toEqual(saved); expect(observation.source).toBe(beforeReopen);
  const storedAfter = await captureBytes(facts, saved.receipt.id, 'capture-after.json'); expect(storedAfter.sha256).toBe(storedBefore.sha256); event('frozen-reopened');
  const genericInstrument = facts.discovery.instruments.find(i => i.metric === 'work.generic.stock'), genericQuestion = { ...widget.measurement, definition: genericInstrument.definition };
  const generic = await json(await request.post(base + '/api/measurements', { data: genericQuestion })); expect(generic.stocks.series[0].samples.some(s => s.value === -1.25)).toBe(true);
  const domainQuestion = structuredClone(genericQuestion); domainQuestion.definition.stockDomain = 'nonnegative-integer';
  const rejected = await request.post(base + '/api/measurements', { data: domainQuestion }); expect(rejected.status()).toBe(502);
  const invalidDomainProbe = { genericQuestion, generic, domainQuestion, status: rejected.status(), response: await rejected.json() };
  // Negative caller-supplied metadata controls, not invented positive telemetry.
  const observed = facts.discovery.instruments.find(i => i.capabilityPlan?.profile === 'database-pool-connections');
  const metadataControls = [];
  for (const kind of ['missing-keys', 'witness-after-preview-limit']) {
    const controlRequest = metadataControlRequest(planRequest.scope, facts.service, observed, kind);
    const controlResponse = await json(await request.post(base + '/api/dashboards/plan', { data: controlRequest }));
    expect(controlResponse.widgets.some(w => w.family === 'database-pool-connections')).toBe(false);
    expect(controlResponse.widgets.some(w => w.measurement?.definition.kind === 'stock')).toBe(true);
    metadataControls.push({ kind, evidence: 'deliberate caller-supplied negative metadata using actual observed streams', request: controlRequest, response: controlResponse });
  }
  expect(observation.errors).toEqual([]);
  await write(facts, 'ledger.json', { version: 1, facts, catalog, planRequest, plan, widgetResult, poolModel, domRows, desktop, mobile, staged, native, saved, reopened, storedBefore, storedAfter, invalidDomainProbe, metadataControls, events, requests: observation.requests, responses: observation.responses });
});
