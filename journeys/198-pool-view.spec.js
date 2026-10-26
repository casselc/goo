const { test, expect } = require('@playwright/test');
const { base, json, observe, seed, write, captureBytes, open } = require('./198-pool-view-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
const row = (page, pool = 'orders', instance = 'worker-a') => page.locator('.pool-connection-row[data-pool="' + pool + '"][data-instance="' + instance + '"]');
async function inspect(page, target, keyboard = false) {
 const panel = target.locator('.pool-signals'); if (await panel.evaluate(e => e.open)) await panel.locator(':scope > summary').click();
 const pending = page.waitForResponse(r => r.url().endsWith('/api/pool-signals/plan') && r.request().method() === 'POST');
 if (keyboard) { await panel.locator(':scope > summary').focus(); await page.keyboard.press('Enter'); } else await panel.locator(':scope > summary').click();
 const response = await pending; return { request: response.request().postDataJSON(), response: await json(response) };
}
async function load(page, target) {
 const pending = page.waitForResponse(r => r.url().endsWith('/api/pool-signals/view') && r.request().method() === 'POST');
 await target.locator('[data-pool-load]').click(); const response = await pending; const result = await json(response);
 await expect(target.locator('.pool-aligned-view')).toBeVisible(); return { request: response.request().postDataJSON(), response: result };
}
async function snap(target) { return target.locator('.pool-aligned-view').evaluate(e => ({ text: e.textContent, html: e.innerHTML, state: e.dataset.state })); }
async function geometry(page, target) {
 expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
 const tracks = await target.locator('.pool-aligned-view .comparison-bar-track').evaluateAll(xs => xs.map(e => e.getBoundingClientRect().width));
 for (const width of tracks) { expect(width).toBeGreaterThan(50); expect(Math.abs(width - tracks[0])).toBeLessThan(1); }
 const maxima = await target.locator('.pool-aligned-view .comparison-bar-cell').evaluateAll(xs => xs.map(e => e.dataset.scaleMax)); expect(new Set(maxima).size).toBe(1); expect(maxima[0]).toBe('12');
 const value = await target.evaluate(e => { const r = e.getBoundingClientRect(); return { left: r.left, right: r.right, width: r.width, scrollWidth: e.scrollWidth, clientWidth: e.clientWidth }; });
 expect(value.left).toBeGreaterThanOrEqual(0); expect(value.right).toBeLessThanOrEqual(await page.evaluate(() => innerWidth)); expect(value.scrollWidth).toBeLessThanOrEqual(value.clientWidth + 1); return { ...value, tracks, maxima };
}
test('198 load compact aligned pool evidence with exact ratio and retained independent clocks', async ({ page, request }) => {
 test.setTimeout(180000); const facts = await seed(request), observation = observe(page), initial = await open(page, facts, observation), events = [];
 const event = action => events.push({ action, source: observation.source(), poolPlans: observation.poolPlans(), views: observation.views() });
 expect(observation.source()).toBe(9); expect(initial.plan.widgets.filter(w => w.measurement && w.state === 'supported')).toHaveLength(2); expect(Buffer.byteLength(JSON.stringify(initial.request))).toBeLessThanOrEqual(16384); event('overview');
 const primaryPlan = await inspect(page, row(page), true); expect(observation.source()).toBe(9); event('keyboard-open');
 const waitCheck = row(page).getByLabel('Load acquisition wait', { exact: true }); await waitCheck.focus(); await page.keyboard.press('Space');
 await expect(row(page).locator('[data-pool-load]')).toContainText('1 source read'); await page.keyboard.press('Space'); await expect(row(page).locator('[data-pool-load]')).toContainText('2 source reads');
 expect(observation.source()).toBe(9); event('keyboard-selection');
 const primary = await load(page, row(page)); expect(observation.source()).toBe(11);
 expect(primary.response.utilization).toEqual({ state: 'observed', reason: 'matching-reported-observation-and-epoch', numerator: '3', denominator: '10' });
 await expect(row(page).locator('.pool-utilization')).toContainText('30%'); await expect(row(page).locator('.pool-utilization')).toHaveAttribute('title', '3/10 of reported maximum'); event('load-pair');
 let cached = await snap(row(page)); const panel = row(page).locator('.pool-signals'), beforeLocal = observation.source();
 await panel.locator(':scope > summary').focus(); await page.keyboard.press('Enter'); await expect(panel).toHaveJSProperty('open', false);
 await page.keyboard.press('Enter'); await expect(panel).toHaveJSProperty('open', true); expect(await snap(row(page))).toEqual(cached); expect(observation.source()).toBe(beforeLocal); event('cached-keyboard-reopen');
 await row(page).getByText('Independent reads and limitations', { exact: true }).click(); expect(observation.source()).toBe(beforeLocal); cached = await snap(row(page));
 await page.setViewportSize({ width: 1440, height: 1800 }); const desktop = await geometry(page, row(page)); await row(page).screenshot({ path: 'docs/journeys/screenshots/198a-aligned-pool-evidence.png' });
 await row(page).scrollIntoViewIfNeeded(); await page.evaluate(() => { const e = document.querySelector('.pool-connection-row[data-pool=orders][data-instance=worker-a]'); scrollBy(0, e.getBoundingClientRect().top - 90); }); await page.screenshot({ path: 'docs/journeys/screenshots/198c-pool-view-in-context.png' });
 await page.setViewportSize({ width: 390, height: 2200 }); const mobile = await geometry(page, row(page)); await row(page).screenshot({ path: 'docs/journeys/screenshots/198b-mobile-aligned-pool.png' }); expect(observation.source()).toBe(beforeLocal); await page.setViewportSize({ width: 1440, height: 1400 });
 const partialPlan = await inspect(page, row(page, 'wait-only')); expect(partialPlan.response.capacity).toHaveLength(0); expect(partialPlan.response.wait).toHaveLength(1);
 const partial = await load(page, row(page, 'wait-only')); expect(observation.source()).toBe(12); expect(partial.request.capacity).toBeUndefined(); expect(partial.response.utilization.reason).toBe('capacity-not-supplied'); event('partial-one-read');
 await row(page).locator('.pool-signals > summary').click(); expect(await snap(row(page))).toEqual(cached); expect(observation.source()).toBe(12); event('cached-row-return');
 const beforePreview = observation.source(); await row(page).locator('[data-pool-signal=capacity]').click(); await expect(page.locator('#measureRun')).toBeEnabled(); expect(observation.source()).toBe(beforePreview); event('individual-staged');
 let pending = page.waitForResponse(r => r.url().endsWith('/api/measurements') && r.request().method() === 'POST'); await page.locator('#measureRun').click(); const individual = await json(await pending); expect(individual.question).toEqual(primary.request.capacity.question); event('individual-run');
 await page.locator('#measureCapture').click(); await page.locator('#captureDraftTitle').fill('Individual aligned pool capacity'); pending = page.waitForResponse(r => r.url().endsWith('/api/captures') && r.request().method() === 'POST'); await page.locator('#captureDraftConfirm').click(); const saved = await json(await pending); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible(); expect(saved.receipt.question.kind).toBe('measurement');
 const storedBefore = await captureBytes(facts, saved.receipt.id, 'capture-before.json'), reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id)); expect(reopened).toEqual(saved); event('individual-capture');
 const frozen = observation.source(); await page.locator('[data-workspace=services]').click(); await expect(row(page).locator('.pool-aligned-view')).toBeVisible(); expect(await snap(row(page))).toEqual(cached); expect(observation.source()).toBe(frozen); event('retained-service-return');
 const boundary = [];
 for (const item of [{ pool: 'orders', instance: 'worker-b', reason: 'zero-capacity' }, { pool: 'time-mismatch', reason: 'observation-time-mismatch' }, { pool: 'epoch-mismatch', reason: 'epoch-mismatch' }, { pool: 'archive', reason: 'used-unavailable' }, { pool: 'conflicting', reason: 'used-unavailable' }, { pool: 'used-exceeds', reason: 'used-exceeds-capacity' }]) {
  const stream = initial.countResult.stocks.series.find(s => s.stream.attributes['db.client.connection.pool.name'] === item.pool && s.stream.resource['service.instance.id'] === (item.instance || 'worker-a')).stream;
  const planRequest = await page.evaluate(({ parent, stream, instruments }) => window.GooPoolSignals.request(parent, stream, instruments), { parent: initial.countResult.question, stream, instruments: facts.discovery.instruments });
  const planResponse = await json(await request.post(base + '/api/pool-signals/plan', { data: planRequest })); expect(planResponse.capacity).toHaveLength(1);
  const capacityRequest = planResponse.capacity[0].question, capacityResult = await json(await request.post(base + '/api/measurements', { data: capacityRequest }));
  const composeRequest = { count: initial.countResult, stream, capacity: capacityResult }, composeResponse = await json(await request.post(base + '/api/pool-signals/view', { data: composeRequest }));
  expect(composeResponse.utilization.reason).toBe(item.reason); expect(composeResponse.utilization.state).toBe(item.reason === 'used-exceeds-capacity' ? 'inconsistent' : 'unavailable');
  boundary.push({ ...item, planRequest, planResponse, capacityRequest, capacityResult, composeRequest, composeResponse });
 }
 const beforeReload = observation.source(); await page.goto(base + '/#capture=' + saved.receipt.id); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible(); await page.reload(); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible(); expect(observation.source()).toBe(beforeReload);
 const storedAfter = await captureBytes(facts, saved.receipt.id, 'capture-after.json'); expect(storedAfter.sha256).toBe(storedBefore.sha256); event('frozen-individual-reload'); expect(observation.errors).toEqual([]);
 await write(facts, 'ledger.json', { version: 1, facts, initial, primaryPlan, primary, cached, desktop, mobile, partialPlan, partial, individual, saved, reopened, storedBefore, storedAfter, boundary, events, requests: observation.requests, responses: observation.responses });
});
