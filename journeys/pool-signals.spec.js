const { test, expect } = require('@playwright/test');
const { base, json, observe, seed, write, captureBytes, open } = require('./197-pool-signals-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
const orders = page => page.locator('.pool-connection-row[data-pool=orders][data-instance=worker-a]');
async function retained(page) { return page.locator('.pool-connections').evaluate(e => ({ text: e.textContent, html: e.innerHTML, search: e.querySelector('[aria-label="Find named pool"]').value, state: e.querySelector('[aria-label="Pool connection state"]').value })); }
async function planSignals(page, row) {
  const panel = row.locator('.pool-signals');
  if (await panel.evaluate(e => e.open)) await panel.locator(':scope > summary').click();
  const pending = page.waitForResponse(r => r.url().endsWith('/api/pool-signals/plan') && r.request().method() === 'POST');
  await panel.locator(':scope > summary').click(); const response = await pending;
  return { request: response.request().postDataJSON(), response: await json(response) };
}
test('197 inspect exact pool capacity and wait with explicit reads and frozen captures', async ({ page, request }) => {
  test.setTimeout(180000); const facts = await seed(request), observation = observe(page), initial = await open(page, facts, observation), events = [];
  const event = action => events.push({ action, source: observation.source(), plans: observation.plans() });
  expect(Buffer.byteLength(JSON.stringify(initial.request))).toBeLessThanOrEqual(16384); expect(initial.plan.widgets.filter(w => w.measurement && w.state === 'supported')).toHaveLength(2); expect(observation.source()).toBe(9); event('overview');
  const beforePlan = observation.source(), paired = await planSignals(page, orders(page)); expect(observation.source()).toBe(beforePlan);
  expect(paired.response.capacity).toHaveLength(1); expect(paired.response.wait).toHaveLength(1); expect(paired.response.version).toBe('goo/pool-signals/v1');
  expect(Buffer.byteLength(JSON.stringify(paired.request))).toBeLessThanOrEqual(65536); event('signals-planned');
  const desktop = await orders(page).boundingBox(); await page.setViewportSize({ width: 1440, height: 1600 });
  await orders(page).screenshot({ path: 'docs/journeys/screenshots/197a-pool-signal-options.png' });
  await page.setViewportSize({ width: 390, height: 1600 }); expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  const mobile = await orders(page).boundingBox(); await orders(page).screenshot({ path: 'docs/journeys/screenshots/197b-mobile-pool-options.png' });
  expect(observation.source()).toBe(beforePlan); await page.setViewportSize({ width: 1440, height: 1400 });
  const results = [];
  for (const kind of ['capacity', 'wait']) {
    const held = await retained(page), before = observation.source(), option = paired.response[kind][0];
    await orders(page).locator('[data-pool-signal=' + kind + ']').click(); await expect(page.locator('#measureRun')).toBeEnabled();
    expect(observation.source()).toBe(before); event(kind + '-staged');
    let pending = page.waitForResponse(r => r.url().endsWith('/api/measurements') && r.request().method() === 'POST');
    await page.locator('#measureRun').click(); const result = await json(await pending); expect(result.question).toEqual(option.question); expect(result.question.definition.source.namespace).toEqual({ mode: 'value', value: 'inventory' });
    if (kind === 'capacity') { expect(result.question.definition.stockDomain).toBe('nonnegative-integer'); expect(result.stocks.series).toHaveLength(1); expect(result.stocks.series[0].samples.map(s => s.value)).toEqual([20, 0]); expect(result.stocks.series[0].samples[1].startNS).not.toBe(result.stocks.series[0].samples[0].startNS); }
    else { expect(result.histograms.series).toHaveLength(1); const interval = result.histograms.series[0].intervals.find(i => i.state === 'ok'); expect(interval.distribution.count).toBe('4'); expect(interval.distribution.bucketCounts).toEqual(['1', '2', '1', '0']); expect(interval.distribution.p95).toEqual({ state: 'bucket-bound', lower: 0.05, upper: 0.1 }); }
    if (kind === 'wait') {
      const selectedIndex = result.histograms.series[0].intervals.findIndex(i => i.state === 'ok');
      await page.locator('#measureResult goo-value-timeline button[data-key="' + selectedIndex + '"]').click();
      await expect(page.locator('#measureResult')).toContainText('Selected interval');
      await page.locator('#measureResult goo-chart-frame.native-timeline').screenshot({ path: 'docs/journeys/screenshots/197c-exact-wait-interval.png' });
    }
    event(kind + '-run');
    await page.locator('#measureCapture').click(); await page.locator('#captureDraftTitle').fill('Independent pool ' + kind);
    pending = page.waitForResponse(r => r.url().endsWith('/api/captures') && r.request().method() === 'POST'); await page.locator('#captureDraftConfirm').click(); const saved = await json(await pending);
    await expect(page.locator('#captureMeasurementEvidence')).toBeVisible(); expect(saved.measurement.question).toEqual(option.question);
    const storedBefore = await captureBytes(facts, saved.receipt.id, kind + '-before.json'); event(kind + '-captured');
    const frozen = observation.source(); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible();
    const reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id)); expect(reopened).toEqual(saved); expect(observation.source()).toBe(frozen);
    const storedAfter = await captureBytes(facts, saved.receipt.id, kind + '-after.json'); expect(storedAfter.sha256).toBe(storedBefore.sha256); event(kind + '-reopened');
    // Preserve the current SPA owner: explicit Services navigation returns without a new source read.
    await page.locator('[data-workspace=services]').click();
    if (!await page.locator('.pool-connections').isVisible()) throw Error('Retained service owner disappeared after capture navigation');
    expect(observation.source()).toBe(frozen); expect(await retained(page)).toEqual(held); event(kind + '-service-return');
    results.push({ kind, option, result, saved, reopened, storedBefore, storedAfter, retained: held });
  }
  const beforeBack = observation.source(); await orders(page).locator('[data-pool-signal=capacity]').click(); await expect(page.locator('#measureRun')).toBeEnabled();
  await page.locator('#serviceReturn').getByRole('button', { name: 'Back to service', exact: true }).click(); await expect(page.locator('.pool-connections')).toBeVisible(); expect(observation.source()).toBe(beforeBack); event('staged-back');
  const partial = await planSignals(page, page.locator('.pool-connection-row[data-pool=archive]')); expect(partial.response.capacity).toHaveLength(0); expect(partial.response.wait).toHaveLength(1);
  const missing = await planSignals(page, page.locator('.pool-connection-row[data-pool=conflicting]')); expect(missing.response.capacity).toHaveLength(0); expect(missing.response.wait).toHaveLength(0); event('partial-and-unconfirmed');
  const beforeFrozenReload = observation.source(), lastCapture = results.at(-1).saved;
  await page.goto(base + '/#capture=' + lastCapture.receipt.id); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible();
  await page.reload(); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible(); expect(observation.source()).toBe(beforeFrozenReload); event('frozen-ui-reload');
  expect(observation.errors).toEqual([]);
  await write(facts, 'ledger.json', { version: 1, facts, initial, paired, results, partial, missing, desktop, mobile, events, requests: observation.requests, responses: observation.responses });
});

test('197 late actual pool metadata plan cannot publish after leaving its service', async ({ page, request }) => {
  test.setTimeout(120000); const facts = await seed(request), observation = observe(page), initial = await open(page, facts, observation);
  let release, entered, finished, held;
  const gate = new Promise(resolve => { release = resolve; }), started = new Promise(resolve => { entered = resolve; }), done = new Promise(resolve => { finished = resolve; });
  await page.route('**/api/pool-signals/plan', async route => { const response = await route.fetch(); held = { request: route.request().postDataJSON(), response: await response.json() }; entered(); await gate; try { await route.fulfill({ response }); } catch {} finally { finished(); } });
  await orders(page).locator('.pool-signals > summary').click(); await started;
  await page.locator('#serviceDetail').getByRole('button', { name: 'Back to service list', exact: true }).click();
  await expect(page.locator('#serviceDetail')).toBeHidden(); const before = { url: page.url(), source: observation.source(), optionCount: await page.locator('[data-pool-signal]').count() };
  release(); await done; await page.unroute('**/api/pool-signals/plan'); await observation.settle();
  const after = { url: page.url(), source: observation.source(), optionCount: await page.locator('[data-pool-signal]').count() }; expect(after).toEqual(before); await expect(page.locator('#serviceDetail')).toBeHidden(); expect(observation.errors).toEqual([]);
  await write(facts, 'ownership.json', { version: 1, facts, initial, held, before, after, requests: observation.requests, responses: observation.responses });
});
