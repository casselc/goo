const { test, expect } = require('@playwright/test');
const { base, json, observe, seed, write } = require('./199-host-support');
test('199 held actual host responses cannot publish after resource-owner departure', async ({ page, request }) => {
  test.setTimeout(120000); const facts = await seed(request), observation = observe(page);
  await page.goto(base + '/#services'); await page.getByRole('tab', { name: 'Hosts', exact: true }).click();
  await page.getByLabel('Host environment', { exact: true }).fill(facts.environment); await page.locator('#minutes').fill('2');
  let pending = page.waitForResponse(r => r.url().endsWith('/api/host-discovery'));
  await page.getByRole('button', { name: 'Find hosts', exact: true }).click(); const inventory = await json(await pending);
  const primary = inventory.hosts.find(h => h.resource['host.id'] === 'host-a' && h.resource['cloud.account.id'] === 'account-a' && !Object.hasOwn(h.resource, 'service.name'));
  let release, entered, finished, held;
  const gate = new Promise(r => { release = r; }), started = new Promise(r => { entered = r; }), done = new Promise(r => { finished = r; });
  await page.route('**/api/measurements', async route => {
    if (route.request().postDataJSON().definition.source.metric !== 'system.cpu.utilization') return route.continue();
    const response = await route.fetch(); held = { request: route.request().postDataJSON(), status: response.status(), response: await response.json() };
    entered(); await gate; try { await route.fulfill({ response }); } catch {} finally { finished(); }
  });
  await page.locator('[data-host-key="' + primary.key + '"]').click(); await started; expect(held.status).toBe(200);
  await page.getByRole('tab', { name: 'Services', exact: true }).click();
  const before = { url: page.url(), source: observation.source(), widgets: await page.locator('.host-widget').count() };
  release(); await done; await page.unroute('**/api/measurements'); await observation.settle();
  const after = { url: page.url(), source: observation.source(), widgets: await page.locator('.host-widget').count() };
  expect(after).toEqual(before); expect(observation.errors).toEqual([]);
  await write(facts, 'ownership.json', { facts, inventory, primary, held, before, after, requests: observation.requests, responses: observation.responses });
});
