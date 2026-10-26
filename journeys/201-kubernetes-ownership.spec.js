const { test, expect } = require('@playwright/test');
const { base, observe, seed, write } = require('./201-kubernetes-support');
test('201 actual pod discovery responses cannot publish after scope edit or tab departure', async ({ page, request }) => {
  test.setTimeout(120000); const facts = await seed(request), observation = observe(page), cases = [];
  await page.goto(base + '/#services'); await page.locator('#entitiesKubernetes').click();
  const environment = page.getByLabel('Kubernetes environment', { exact: true }); await environment.fill(facts.environment); await page.locator('#minutes').fill('2');
  for (const reason of ['environment', 'tab']) {
    let release, entered, finished, held;
    const gate = new Promise(r => { release = r; }), started = new Promise(r => { entered = r; }), done = new Promise(r => { finished = r; });
    await page.route('**/api/pod-discovery', async route => {
      const response = await route.fetch(); held = { request: route.request().postDataJSON(), status: response.status(), response: await response.json() };
      entered(); await gate; try { await route.fulfill({ response }); } catch {} finally { finished(); }
    });
    await page.getByRole('button', { name: 'Find pods', exact: true }).click(); await started; expect(held.status).toBe(200);
    if (reason === 'environment') await environment.fill(facts.environment + '-draft'); else await page.locator('#entitiesServices').click();
    const snapshot = async () => ({ url: page.url(), source: observation.source(), rows: await page.locator('tr[data-pod-variant]').count(), kubernetesHidden: await page.locator('.kubernetes-browser').isHidden() });
    const before = await snapshot(); release(); await done; await page.unroute('**/api/pod-discovery'); await observation.settle(); const after = await snapshot();
    expect(after).toEqual(before); expect(after.rows).toBe(0); cases.push({ reason, held, before, after });
    if (reason === 'environment') await environment.fill(facts.environment);
  }
  expect(observation.source()).toBe(2); expect(observation.errors).toEqual([]);
  await write(facts, 'ownership.json', { facts, cases, requests: observation.requests, responses: observation.responses });
});
