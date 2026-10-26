const { test, expect } = require('@playwright/test');
const { base, json, observe, seed, open, write, captureBytes, assertRaw } = require('./200-runtime-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
async function geometry(page) {
  const snapshot = await page.locator('.runtime-observations').evaluateAll(elements => elements.map(e => {
    const box = e.getBoundingClientRect(), plot = e.querySelector('.value-timeline-plot').getBoundingClientRect(), issues = [];
    for (const node of e.querySelectorAll('p,button,a,goo-value-timeline')) {
      if (node.closest('details:not([open])') || !node.getClientRects().length) continue;
      const r = node.getBoundingClientRect();
      if (r.left < box.left - 1 || r.right > box.right + 1 || r.top < box.top - 1 || r.bottom > box.bottom + 1) issues.push(node.tagName);
    }
    return { metric: e.dataset.runtimeMetric, instance: e.dataset.instance, state: e.dataset.state,
      text: e.textContent, valueTitle: e.querySelector(':scope > p strong').title, issues,
      plotHeight: plot.height, scrollWidth: e.scrollWidth, clientWidth: e.clientWidth,
      clocks: [...e.querySelectorAll(':scope > p time')].map(t => ({ text: t.textContent, title: t.title, datetime: t.dateTime })) };
  }));
  for (const row of snapshot) { expect(row.issues).toEqual([]); expect(row.scrollWidth).toBeLessThanOrEqual(row.clientWidth + 1); expect(row.plotHeight).toBe(90); for (const clock of row.clocks) { expect(clock.text).toContain('UTC'); expect(clock.title).toMatch(/^\d+ ns since Unix epoch$/); } }
  return snapshot;
}
test('200 exact Go runtime stocks, heap target and immutable native evidence', async ({ page, request }) => {
  test.setTimeout(150000); const facts = await seed(request), observation = observe(page), events = [];
  const event = action => events.push({ action, source: observation.source(), plans: observation.plans() });
  const initial = await open(page, facts, observation); event('overview');
  const widgets = initial.plan.widgets.filter(w => w.family === 'go-runtime');
  expect(widgets).toHaveLength(2); expect(widgets.map(w => w.measurement.definition.source.metric).sort()).toEqual(['go.memory.gc.goal', 'go.memory.used']);
  const results = observation.responses.filter(r => r.endpoint === '/api/measurements' && r.body?.question?.definition.source.metric.startsWith('go.')).map(r => r.body);
  expect(results).toHaveLength(2); results.forEach(r => assertRaw(facts, r));
  expect(results[0].question.definition.source.exactStream.resource).toEqual(results[1].question.definition.source.exactStream.resource);
  await expect(page.locator('.runtime-observations')).toHaveCount(2);
  const beforeLocal = observation.source();
  await page.setViewportSize({ width: 1440, height: 1600 }); const desktop = await geometry(page);
  await page.locator('.service-dashboard').screenshot({ path: 'docs/journeys/screenshots/200a-runtime-observations.png' });
  await page.screenshot({ path: 'docs/journeys/screenshots/200c-runtime-in-context.png' });
  await page.setViewportSize({ width: 390, height: 2200 }); const mobile = await geometry(page);
  await page.locator('.service-dashboard').screenshot({ path: 'docs/journeys/screenshots/200b-mobile-runtime-observations.png' });
  expect(observation.source()).toBe(beforeLocal); await page.setViewportSize({ width: 1440, height: 1600 });
  const card = page.locator('.runtime-observations[data-runtime-metric="go.memory.used"]');
  await card.getByText('Exact stream and meaning', { exact: true }).click();
  const retained = await page.locator('.service-dashboard').innerHTML();
  await card.getByRole('link', { name: 'Inspect exact runtime measurement', exact: true }).click();
  await expect(page.locator('#measureRun')).toBeEnabled(); expect(observation.source()).toBe(beforeLocal); event('native-staged');
  await page.locator('#serviceReturn').getByRole('button', { name: 'Back to service', exact: true }).click();
  expect(await page.locator('.service-dashboard').innerHTML()).toBe(retained); expect(observation.source()).toBe(beforeLocal); event('retained-back');
  await card.getByRole('link', { name: 'Inspect exact runtime measurement', exact: true }).click();
  await expect(page.locator('#measureRun')).toBeEnabled();
  let pending = page.waitForResponse(r => r.url().endsWith('/api/measurements') && r.request().method() === 'POST');
  await page.locator('#measureRun').click(); const individual = await json(await pending); assertRaw(facts, individual); event('native-run');
  await page.locator('#measureCapture').click(); await page.locator('#captureDraftTitle').fill('Exact reported Go runtime observation');
  pending = page.waitForResponse(r => r.url().endsWith('/api/captures') && r.request().method() === 'POST');
  await page.locator('#captureDraftConfirm').click(); const saved = await json(await pending); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible();
  const storedBefore = await captureBytes(facts, saved.receipt.id, 'capture-before.json'); event('individual-capture');
  const beforeReopen = observation.source(), reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id)); expect(reopened).toEqual(saved);
  await page.locator('[data-workspace=services]').click();
  expect(await page.locator('.service-dashboard').innerHTML()).toBe(retained); expect(observation.source()).toBe(beforeReopen); event('capture-retained-service-return');
  const goroutineOption = initial.plan.nativeOptions.find(option => option.instrument.metric === 'go.goroutine.count' && option.instrument.definition?.source.exactStream);
  expect(goroutineOption).toBeTruthy(); const picker = page.locator('.dashboard-native-picker');
  await picker.locator(':scope > summary').click(); await picker.getByRole('button', { name: 'Clear selection', exact: true }).click();
  await picker.locator('input[type=checkbox][value="' + goroutineOption.key + '"]').check();
  expect(observation.source()).toBe(beforeReopen); event('goroutine-selection-staged');
  const beforeApplyResponses = observation.responses.length;
  pending = page.waitForResponse(r => r.url().endsWith('/api/dashboards/plan') && r.request().method() === 'POST');
  await picker.getByRole('button', { name: 'Apply and reread overview', exact: true }).click();
  const goroutinePlanResponse = await pending, goroutinePlan = await json(goroutinePlanResponse);
  await expect(page.locator('.service-dashboard')).toHaveAttribute('data-state', 'loaded'); await observation.settle();
  const selectedWidgets = goroutinePlan.widgets.filter(w => w.measurement && w.state === 'supported');
  expect(selectedWidgets).toHaveLength(1); expect(selectedWidgets[0].measurement.definition.source.metric).toBe('go.goroutine.count');
  expect(goroutinePlanResponse.request().postDataJSON().nativeSelection).toEqual([goroutineOption.key]);
  const goroutineReads = observation.responses.slice(beforeApplyResponses).filter(r => r.endpoint === '/api/measurements' && r.body?.question);
  expect(goroutineReads).toHaveLength(1); const goroutine = goroutineReads[0].body; assertRaw(facts, goroutine);
  const goroutineValue = goroutine.stocks.series[0].samples.at(-1).value; expect([7, 11]).toContain(goroutineValue);
  await expect(page.locator('.runtime-observations')).toHaveCount(1);
  const goroutineCard = page.locator('.runtime-observations[data-runtime-metric="go.goroutine.count"]');
  await expect(goroutineCard.locator(':scope > p strong')).toHaveText(String(goroutineValue) + ' goroutines');
  await expect(goroutineCard.locator(':scope > p strong')).toHaveAttribute('title', String(goroutineValue) + ' {goroutine}'); event('goroutine-alone-read');
  const beforeFrozenUI = observation.source(); await page.goto(base + '/#capture=' + saved.receipt.id); await page.reload();
  await expect(page.locator('#captureMeasurementEvidence')).toBeVisible(); expect(observation.source()).toBe(beforeFrozenUI);
  const storedAfter = await captureBytes(facts, saved.receipt.id, 'capture-after.json'); expect(storedAfter.sha256).toBe(storedBefore.sha256); event('immutable-reopen');
  expect(observation.errors).toEqual([]);
  await write(facts, 'ledger.json', { facts, initial, results, desktop, mobile, retained, individual, saved, reopened, goroutineOption, goroutinePlan, goroutine, storedBefore, storedAfter, events, requests: observation.requests, responses: observation.responses });
});
