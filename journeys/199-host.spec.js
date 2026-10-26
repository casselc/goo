const { test, expect } = require('@playwright/test');
const { base, json, observe, seed, write, captureBytes } = require('./199-host-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
async function geometry(page) {
  const issues = await page.locator('.host-detail .host-widget').evaluateAll(elements => elements.flatMap(element => {
    const box = element.getBoundingClientRect(), result = [];
    if (box.left < -1 || box.right > innerWidth + 1) result.push('widget outside viewport');
    if (element.scrollWidth > element.clientWidth + 1) result.push('widget horizontal overflow');
    for (const title of element.querySelectorAll('h4,p,button,goo-value-timeline')) {
      if (title.closest('details:not([open])') || !title.getClientRects().length) continue;
      const r = title.getBoundingClientRect();
      if (r.left < box.left - 1 || r.right > box.right + 1) result.push('content outside widget');
      if (r.top < box.top - 1 || r.bottom > box.bottom + 1) result.push('content vertically outside widget');
    }
    return result;
  }));
  expect(issues).toEqual([]); return { issues, viewport: page.viewportSize() };
}
test('199 exact host resources and explicit native observations', async ({ page, request }) => {
  test.setTimeout(150000);
  const facts = await seed(request), observation = observe(page), events = [];
  const event = action => events.push({ action, source: observation.source(), plans: observation.hostPlans() });
  await page.goto(base + '/#services'); await page.getByRole('tab', { name: 'Hosts', exact: true }).click();
  await page.getByLabel('Host environment', { exact: true }).fill(facts.environment); await page.locator('#minutes').fill('2');
  event('draft');
  let pending = page.waitForResponse(r => r.url().endsWith('/api/host-discovery') && r.request().method() === 'POST');
  await page.getByRole('button', { name: 'Find hosts', exact: true }).click(); const inventoryResponse = await pending;
  const inventory = await json(inventoryResponse), discoveryQuestion = inventoryResponse.request().postDataJSON();
  expect(inventory.records).toBe(facts.points.length);
  const primary = inventory.hosts.find(h => h.resource['host.id'] === 'host-a' && h.resource['cloud.account.id'] === 'account-a' && !Object.hasOwn(h.resource, 'service.name'));
  expect(primary).toBeTruthy();
  expect(inventory.hosts.some(h => Object.hasOwn(h.resource, 'service.name') && h.resource['service.name'] === '')).toBe(true);
  event('discover'); const beforeOpen = observation.source();
  pending = page.waitForResponse(r => r.url().endsWith('/api/host-dashboard/plan'));
  await page.locator('[data-host-key="' + primary.key + '"]').click(); const planResponse = await pending, plan = await json(planResponse);
  await expect(page.locator('.host-widget')).toHaveCount(2); await observation.settle();
  expect(observation.source()).toBe(beforeOpen + 2); expect(plan.widgets).toHaveLength(2);
  const results = observation.responses.filter(r => r.endpoint === '/api/measurements' && r.body?.question).map(r => r.body);
  for (const result of results) {
    expect(result.question.definition.source.resource).toEqual({ kind: 'host', id: 'host-a' });
    expect(result.question.definition.source.exactStream.resource).toEqual(primary.resource);
    expect(result.question.definition.source.services || []).toEqual([]);
  }
  event('host-observations');
  await page.setViewportSize({ width: 1440, height: 1600 }); const desktop = await geometry(page);
  await page.locator('.host-detail').screenshot({ path: 'docs/journeys/screenshots/199a-host-observations.png' });
  await page.screenshot({ path: 'docs/journeys/screenshots/199c-hosts-in-context.png' });
  await page.setViewportSize({ width: 390, height: 2200 }); const mobile = await geometry(page);
  await page.locator('.host-detail').screenshot({ path: 'docs/journeys/screenshots/199b-mobile-host-observations.png' });
  expect(observation.source()).toBe(beforeOpen + 2);
  await page.setViewportSize({ width: 1440, height: 1600 });
  const heldHTML = await page.locator('.host-detail').innerHTML(), context = [];
  for (const signal of ['logs', 'spans']) {
    const before = observation.source();
    await page.getByRole('link', { name: 'Matching resource ' + signal, exact: true }).click();
    await expect(page.locator('#exploreRun')).toBeEnabled(); expect(observation.source()).toBe(before); event(signal + '-staged');
    pending = page.waitForResponse(r => r.url().endsWith('/api/explore') && r.request().method() === 'POST');
    await page.locator('#exploreRun').click(); const response = await pending, result = await json(response), question = response.request().postDataJSON();
    expect(result.count).toBe('1');
    expect(question.where).toEqual({ all: Object.entries(primary.resource).map(([key, value]) => ({ field: { kind: 'resource', key }, op: 'eq', value })) });
    expect(question.where.all.some(p => p.field.key === 'service.name')).toBe(false);
    expect(result.rows[0].ResourceAttributes['service.name']).toBe('host-context-app');
    context.push({ signal, question, result }); event(signal + '-run');
    const after = observation.source(); await page.getByRole('button', { name: 'Back to host', exact: true }).click();
    expect(await page.locator('.host-detail').innerHTML()).toBe(heldHTML); expect(observation.source()).toBe(after); event(signal + '-retained-back');
  }
  const beforePreview = observation.source();
  await page.locator('.host-widget').first().getByRole('button', { name: 'Inspect measurement', exact: true }).click();
  await expect(page.locator('#measureRun')).toBeEnabled(); expect(observation.source()).toBe(beforePreview); event('native-preview');
  pending = page.waitForResponse(r => r.url().endsWith('/api/measurements') && r.request().method() === 'POST');
  await page.locator('#measureRun').click(); const individual = await json(await pending); event('native-run');
  await page.locator('#measureCapture').click(); await page.locator('#captureDraftTitle').fill('Exact reported host stream');
  pending = page.waitForResponse(r => r.url().endsWith('/api/captures') && r.request().method() === 'POST');
  await page.locator('#captureDraftConfirm').click(); const saved = await json(await pending);
  await expect(page.locator('#captureMeasurementEvidence')).toBeVisible();
  const storedBefore = await captureBytes(facts, saved.receipt.id, 'capture-before.json'); event('individual-capture');
  const beforeReopen = observation.source(), reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id));
  expect(reopened).toEqual(saved); await page.reload(); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible();
  expect(observation.source()).toBe(beforeReopen);
  const storedAfter = await captureBytes(facts, saved.receipt.id, 'capture-after.json');
  expect(storedAfter.sha256).toBe(storedBefore.sha256); event('immutable-reopen');
  await write(facts, 'ledger.json', { facts, discoveryQuestion, inventory, primary, planRequest: planResponse.request().postDataJSON(), plan, results, desktop, mobile, context, individual, saved, reopened, storedBefore, storedAfter, events, requests: observation.requests, responses: observation.responses });
  expect(observation.errors).toEqual([]);
});
module.exports = { geometry };
