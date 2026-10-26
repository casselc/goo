const { test, expect } = require('@playwright/test');
const { base, json, observe, seed, write } = require('./199-host-support');
test('199 compact host plot display replay with fresh ordinary telemetry', async ({ page, request }) => {
  test.skip(process.env.GOO_HOST_DISPLAY !== '1', 'Explicit display replay'); test.setTimeout(90000);
  const facts = await seed(request), observation = observe(page);
  await page.goto(base + '/#services'); await page.getByRole('tab', { name: 'Hosts', exact: true }).click();
  await page.getByLabel('Host environment', { exact: true }).fill(facts.environment); await page.locator('#minutes').fill('2');
  let pending = page.waitForResponse(r => r.url().endsWith('/api/host-discovery'));
  await page.getByRole('button', { name: 'Find hosts', exact: true }).click(); const inventory = await json(await pending);
  const primary = inventory.hosts.find(h => h.resource['host.id'] === 'host-a' && h.resource['cloud.account.id'] === 'account-a' && !Object.hasOwn(h.resource, 'service.name'));
  pending = page.waitForResponse(r => r.url().endsWith('/api/host-dashboard/plan'));
  await page.locator('button[data-host-key="' + primary.key + '"]').click(); const plan = await json(await pending);
  await expect(page.locator('.host-widget')).toHaveCount(2); await observation.settle(); expect(observation.source()).toBe(3);
  const geometries = [];
  for (const [width, height, frame] of [[1440, 1600, '199a-host-observations.png'], [390, 1800, '199b-mobile-host-observations.png']]) {
    await page.setViewportSize({ width, height });
    const geometry = await page.locator('.host-widget').evaluateAll(elements => elements.map(e => {
      const box = e.getBoundingClientRect(), plot = e.querySelector('.value-timeline-plot').getBoundingClientRect();
      const issues = [];
      for (const node of e.querySelectorAll('h4,p,button,goo-value-timeline')) {
        if (node.closest('details:not([open])') || !node.getClientRects().length) continue;
        const r = node.getBoundingClientRect(); if (r.left < box.left - 1 || r.right > box.right + 1 || r.top < box.top - 1 || r.bottom > box.bottom + 1) issues.push(node.tagName);
      }
      const clocks = [...e.querySelectorAll(':scope > p time')].map(t => ({ text: t.textContent, title: t.title, datetime: t.dateTime }));
      return { clocks, width: box.width, scrollWidth: e.scrollWidth, clientWidth: e.clientWidth, plotHeight: plot.height, issues };
    }));
    for (const entry of geometry) { expect(entry.issues).toEqual([]); expect(entry.scrollWidth).toBeLessThanOrEqual(entry.clientWidth + 1); expect(entry.plotHeight).toBe(90); expect(entry.clocks).toHaveLength(1); for (const clock of entry.clocks) { expect(clock.text).toContain('UTC'); expect(clock.title).toMatch(/^\d+ ns since Unix epoch$/); expect(clock.datetime).toMatch(/\.\d{9}Z$/); } }
    geometries.push({ width, height, geometry });
    await page.locator('.host-detail').screenshot({ path: 'docs/journeys/screenshots/' + frame });
  }
  await page.setViewportSize({ width: 1440, height: 1600 }); await page.locator('.host-detail').scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/199c-hosts-in-context.png' });
  expect(observation.source()).toBe(3); expect(observation.errors).toEqual([]);
  await write(facts, 'display.json', { facts, inventory, primary, plan, geometries, source: observation.source(), requests: observation.requests, responses: observation.responses });
});
