const { test, expect } = require('@playwright/test');
const { json, observe, seed, write, open } = require('./198-pool-view-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
const row = page => page.locator('.pool-connection-row[data-pool=orders][data-instance=worker-a]');
async function metadata(page) {
 const panel = row(page).locator('.pool-signals'); if (await panel.evaluate(e => e.open)) return;
 const pending = page.waitForResponse(r => r.url().endsWith('/api/pool-signals/plan') && r.request().method() === 'POST'); await panel.locator(':scope > summary').click(); await json(await pending);
}
test('198 actual optional source responses cannot publish after close, selection change or service departure', async ({ page, request }) => {
 test.setTimeout(150000); const facts = await seed(request), observation = observe(page), initial = await open(page, facts, observation), cases = [];
 for (const reason of ['close', 'selection', 'leave']) {
  await metadata(page); const panel = row(page).locator('.pool-signals'); await panel.getByLabel('Load acquisition wait', { exact: true }).uncheck();
  let release, entered, finished, held; const gate = new Promise(r => { release = r; }), started = new Promise(r => { entered = r; }), done = new Promise(r => { finished = r; });
  await page.route('**/api/measurements', async route => { if (route.request().postDataJSON().definition.source.metric !== 'db.client.connection.max') return route.continue(); const response = await route.fetch(); held = { request: route.request().postDataJSON(), status: response.status(), response: await response.json() }; entered(); await gate; try { await route.fulfill({ response }); } catch {} finally { finished(); } });
  const beforeLoad = { source: observation.source(), views: observation.views() }; await panel.locator('[data-pool-load]').click(); await started; expect(held.status).toBe(200);
  if (reason === 'close') await panel.locator(':scope > summary').click();
  else if (reason === 'selection') {
   // Current UI leaves selection controls enabled while source reads are pending.
   await panel.getByLabel('Load capacity', { exact: true }).uncheck();
  } else await page.locator('#serviceDetail').getByRole('button', { name: 'Back to service list', exact: true }).click();
  const beforeRelease = { source: observation.source(), views: observation.views(), url: page.url(), viewCount: await row(page).locator('.pool-aligned-view').count() };
  release(); await done; await page.unroute('**/api/measurements'); await observation.settle();
  const afterRelease = { source: observation.source(), views: observation.views(), url: page.url(), viewCount: await row(page).locator('.pool-aligned-view').count() }; expect(afterRelease).toEqual(beforeRelease); expect(afterRelease.viewCount).toBe(0); expect(afterRelease.views).toBe(beforeLoad.views);
  cases.push({ reason, deliberateRetainedEvent: reason === 'selection', held, beforeLoad, beforeRelease, afterRelease });
  if (reason === 'close') await panel.locator(':scope > summary').click();
  if (reason !== 'leave') { await panel.getByLabel('Load capacity', { exact: true }).check(); }
 }
 expect(observation.errors).toEqual([]); await write(facts, 'ownership.json', { version: 1, facts, initial, cases, requests: observation.requests, responses: observation.responses });
});
