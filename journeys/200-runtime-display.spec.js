const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const { observe, seed, open, write, assertRaw, captureBytes } = require('./200-runtime-support');
test('200 runtime labels and compact cards with a fresh display fixture', async ({ page, request }) => {
  test.skip(process.env.GOO_RUNTIME_DISPLAY !== '1', 'Explicit display-only replay'); test.setTimeout(90000);
  const facts = await seed(request), observation = observe(page), initial = await open(page, facts, observation);
  const results = observation.responses.filter(r => r.endpoint === '/api/measurements' && r.body?.question?.definition.source.metric.startsWith('go.')).map(r => r.body);
  expect(results).toHaveLength(2); results.forEach(result => assertRaw(facts, result));
  const sourceBefore = observation.source(), plansBefore = observation.plans(); expect(sourceBefore).toBe(8); expect(plansBefore).toBe(1);
  const goal = page.locator('.runtime-observations[data-runtime-metric="go.memory.gc.goal"]');
  await expect(goal.locator(':scope > p').nth(1)).toHaveText(/^Instance worker-[ab]$/);
  const geometry = [];
  for (const [width, height, frame] of [[1440, 1600, '200a-runtime-observations.png'], [390, 2200, '200b-mobile-runtime-observations.png']]) {
    await page.setViewportSize({ width, height });
    const snapshot = await page.locator('.runtime-observations').evaluateAll(elements => elements.map(e => {
      const box = e.getBoundingClientRect(), issues = [], plot = e.querySelector('.value-timeline-plot').getBoundingClientRect();
      for (const node of e.querySelectorAll('p,button,a,goo-value-timeline')) {
        if (node.closest('details:not([open])') || !node.getClientRects().length) continue;
        const r = node.getBoundingClientRect();
        if (r.left < box.left - 1 || r.right > box.right + 1 || r.top < box.top - 1 || r.bottom > box.bottom + 1) issues.push(node.tagName);
      }
      return { metric: e.dataset.runtimeMetric, instance: e.dataset.instance, text: e.textContent,
        metadata: e.querySelectorAll(':scope > p')[1].textContent, valueTitle: e.querySelector(':scope > p strong').title,
        plotHeight: plot.height, issues, scrollWidth: e.scrollWidth, clientWidth: e.clientWidth,
        clocks: [...e.querySelectorAll(':scope > p time')].map(t => ({ text: t.textContent, title: t.title, datetime: t.dateTime })) };
    }));
    for (const entry of snapshot) {
      expect(entry.issues).toEqual([]); expect(entry.scrollWidth).toBeLessThanOrEqual(entry.clientWidth + 1); expect(entry.plotHeight).toBe(90);
      expect(entry.metadata).not.toMatch(/·\s*$/);
      for (const clock of entry.clocks) { expect(clock.text).toContain('UTC'); expect(clock.title).toMatch(/^\d+ ns since Unix epoch$/); expect(clock.datetime).toMatch(/\.\d{9}Z$/); }
    }
    geometry.push({ width, height, snapshot });
    await page.locator('.service-dashboard').screenshot({ path: 'docs/journeys/screenshots/' + frame });
  }
  await page.setViewportSize({ width: 1440, height: 1600 }); await page.locator('.service-dashboard').scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/200c-runtime-in-context.png' });
  expect(observation.source()).toBe(sourceBefore); expect(observation.plans()).toBe(plansBefore); expect(observation.errors).toEqual([]);
  const original = JSON.parse(await fs.readFile(process.env.GOO_RUNTIME_LEDGER || 'artifacts/runtime-observations/runtime-observed-220f90ed2a/ledger.json', 'utf8'));
  const existingCaptureAfterDisplay = await captureBytes(facts, original.saved.receipt.id, 'existing-capture-after-display.json');
  expect(existingCaptureAfterDisplay.sha256).toBe(original.storedBefore.sha256);
  await write(facts, 'display.json', { facts, initial, results, geometry, source: sourceBefore, plans: plansBefore, existingCaptureAfterDisplay, requests: observation.requests, responses: observation.responses });
});
