const { test, expect } = require('@playwright/test');
const runtime = require('./200-runtime-support');
const { write } = require('./201-kubernetes-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
test('201 leaving Services cancels pending support and widget owners while completed views remain usable', async ({ page, request }) => {
  test.setTimeout(150000); const facts = await runtime.seed(request), observation = runtime.observe(page), catalog = await runtime.discover(page, facts), cases = [];
  const openButton = () => page.getByRole('button', { name: 'Open service ' + facts.namespace + ' / ' + facts.service, exact: true });
  const snapshot = async () => ({ url: page.url(), source: observation.source(), plans: observation.requests.filter(r => r.endpoint === '/api/dashboards/plan').length,
    runtimeCards: await page.locator('.runtime-observations').count(), detailHidden: await page.locator('#serviceDetail').isHidden(),
    dashboardState: await page.locator('.service-dashboard').evaluateAll(nodes => nodes[0]?.getAttribute('data-state') ?? null) });
  for (const phase of ['supporting', 'widget']) {
    let release, entered, finished, held;
    const gate = new Promise(r => { release = r; }), started = new Promise(r => { entered = r; }), done = new Promise(r => { finished = r; });
    const pattern = phase === 'supporting' ? '**/api/explore' : '**/api/measurements';
    await page.route(pattern, async route => {
      const question = route.request().postDataJSON();
      if (phase === 'supporting' ? question.signal !== 'spans' : question.definition.source.metric !== 'go.memory.used') return route.continue();
      const response = await route.fetch(); held = { request: question, status: response.status(), response: await response.json() };
      entered(); await gate; try { await route.fulfill({ response }); } catch {} finally { finished(); }
    });
    const beforeOpen = await snapshot(); await openButton().click(); await started; expect(held.status).toBe(200);
    await page.locator('#entitiesKubernetes').click(); const beforeRelease = await snapshot();
    release(); await done; await page.unroute(pattern); await observation.settle(); const afterRelease = await snapshot();
    expect(afterRelease).toEqual(beforeRelease); expect(afterRelease.detailHidden).toBe(true);
    if (phase === 'supporting') expect(afterRelease.plans).toBe(beforeOpen.plans);
    cases.push({ phase, held, beforeOpen, beforeRelease, afterRelease });
    await page.locator('#entitiesServices').click(); await expect(openButton()).toBeVisible();
  }
  const pending = page.waitForResponse(r => r.url().endsWith('/api/dashboards/plan') && r.request().method() === 'POST');
  await openButton().click(); const response = await pending, plan = await runtime.json(response);
  await expect(page.locator('.service-dashboard')).toHaveAttribute('data-state', 'loaded'); await observation.settle();
  const retainedHTML = await page.locator('.service-dashboard').innerHTML(), beforeTabs = observation.source();
  await page.locator('#entitiesKubernetes').click(); await page.locator('#entitiesServices').click();
  expect(await page.locator('.service-dashboard').innerHTML()).toBe(retainedHTML); expect(observation.source()).toBe(beforeTabs);
  const option = plan.nativeOptions.find(o => o.instrument.metric === 'go.goroutine.count' && o.instrument.definition?.source.exactStream); expect(option).toBeTruthy();
  const picker = page.locator('.dashboard-native-picker'); await picker.locator(':scope > summary').click();
  await picker.getByRole('button', { name: 'Clear selection', exact: true }).click(); await picker.locator('input[value="' + option.key + '"]').check();
  expect(observation.source()).toBe(beforeTabs); const beforeApplyResponses = observation.responses.length;
  await picker.getByRole('button', { name: 'Apply and reread overview', exact: true }).click();
  await expect(page.locator('.service-dashboard')).toHaveAttribute('data-state', 'loaded'); await observation.settle();
  await expect(page.locator('.runtime-observations')).toHaveCount(1); await expect(page.locator('.runtime-observations')).toHaveAttribute('data-runtime-metric', 'go.goroutine.count');
  const applied = observation.responses.slice(beforeApplyResponses).find(r => r.endpoint === '/api/measurements' && r.body?.question.definition.source.metric === 'go.goroutine.count'); expect(applied).toBeTruthy(); runtime.assertRaw(facts, applied.body);
  expect(observation.errors).toEqual([]);
  await write(facts, 'services-ownership.json', { facts, catalog, cases, plan, retainedHTML, beforeTabs, afterApply: await snapshot(), option, applied, requests: observation.requests, responses: observation.responses });
});
