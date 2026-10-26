const { test, expect } = require('@playwright/test');
const { observe, seed, discover, write } = require('./200-runtime-support');
test('200 held actual runtime stock response cannot publish after service departure', async ({ page, request }) => {
  test.setTimeout(120000); const facts = await seed(request), observation = observe(page), catalog = await discover(page, facts);
  let release, entered, finished, held;
  const gate = new Promise(r => { release = r; }), started = new Promise(r => { entered = r; }), done = new Promise(r => { finished = r; });
  await page.route('**/api/measurements', async route => {
    if (route.request().postDataJSON().definition.source.metric !== 'go.memory.used') return route.continue();
    const response = await route.fetch(); held = { request: route.request().postDataJSON(), status: response.status(), response: await response.json() };
    entered(); await gate; try { await route.fulfill({ response }); } catch {} finally { finished(); }
  });
  await page.getByRole('button', { name: 'Open service ' + facts.namespace + ' / ' + facts.service, exact: true }).click();
  await started; expect(held.status).toBe(200);
  await page.locator('#serviceDetail').getByRole('button', { name: 'Back to service list', exact: true }).click();
  const snapshot = async () => ({ url: page.url(), source: observation.source(), rendered: await page.locator('.runtime-observations').count(), detailHidden: await page.locator('#serviceDetail').isHidden() });
  const before = await snapshot(); release(); await done; await page.unroute('**/api/measurements'); await observation.settle();
  const after = await snapshot(); expect(after).toEqual(before); expect(after.detailHidden).toBe(true); expect(observation.errors).toEqual([]);
  await write(facts, 'ownership.json', { facts, catalog, held, before, after, requests: observation.requests, responses: observation.responses });
});
