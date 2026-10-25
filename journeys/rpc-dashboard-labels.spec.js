const { test, expect } = require('@playwright/test');
const { base, json, seed, observe, loaded, snapshot, write } = require('./188-rpc-dashboard-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });

test('188 same-name native RPC unit labels remain distinct without changing count axes', async ({ page, request }) => {
  test.setTimeout(120000);
  const facts = await seed(request), observation = observe(page);
  await page.goto(base + '/#services');
  await page.locator('#inventoryEnvironment').fill(facts.environment);
  await page.locator('#minutes').fill('2');
  await page.locator('#inventoryRun').click();
  await expect(page.getByRole('button', { name: 'Open service commerce / rpc-api', exact: true })).toBeVisible();
  const planning = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service commerce / rpc-api', exact: true }).click();
  const plan = await json(await planning), dash = await loaded(page, 'rpc-server');
  const before = await snapshot(page, observation);
  const bindings = plan.widgets.filter(widget => widget.measurement).map(widget => ({ key: widget.key, name: widget.measurement.definition.source.metric, unit: widget.measurement.definition.source.unit }));
  expect(bindings.map(binding => [binding.name, binding.unit])).toEqual([['rpc.server.duration', 'ms'], ['rpc.server.duration', 's']]);
  for (const binding of bindings) await expect(dash.locator('[data-widget=' + binding.key + '] > h5')).toHaveText(binding.name + ' [' + binding.unit + ']');
  const readSnapshot = () => dash.evaluate(host => ({
    native: [...host.querySelectorAll('.service-timeline-lane[data-lane^="native-"]')].map(row => ({
      key: row.dataset.lane, label: row.querySelector('strong').textContent, data: row.querySelector('goo-value-timeline').data,
    })),
    lanes: [...host.querySelectorAll('.service-timeline-lane')].map(row => {
      const label = row.querySelector(':scope>div'), plot = row.querySelector('goo-value-timeline');
      const rect = element => { const box = element.getBoundingClientRect(); return { left: box.left, right: box.right }; };
      return { label: rect(label), plot: rect(plot), children: [...label.children].map(child => ({ ...rect(child), scroll: child.scrollWidth, width: child.clientWidth })) };
    }),
    bars: [...host.querySelectorAll('[data-widget=dimension] tbody tr')].map(row => [...row.querySelectorAll('.comparison-bar-cell')].map(cell => ({ value: cell.dataset.value,
      track: cell.querySelector('.comparison-bar-track').getBoundingClientRect().width, fill: cell.querySelector('.comparison-bar-fill').getBoundingClientRect().width }))),
  }));
  const check = value => {
    expect(value.native).toHaveLength(2);
    for (const binding of bindings) {
      const lane = value.native.find(lane => lane.key.startsWith(binding.key + '-'));
      expect(lane.label).toContain(binding.name + ' [' + binding.unit + ']');
      expect(lane.data.unit).toBe('observations per interval');
      expect(lane.data.points.some(point => point.value === '4')).toBe(true);
    }
    for (const lane of value.lanes) {
      expect(lane.label.right).toBeLessThanOrEqual(lane.plot.left + 1);
      for (const child of lane.children) {
        expect(child.left).toBeGreaterThanOrEqual(lane.label.left - 1);
        expect(child.right).toBeLessThanOrEqual(lane.label.right + 1);
        expect(child.scroll).toBeLessThanOrEqual(child.width + 1);
      }
    }
    for (const bars of value.bars) {
      expect(bars).toHaveLength(2); expect(bars[0].value).toBe(bars[1].value);
      expect(bars[0].track).toBeGreaterThan(50);
      expect(Math.abs(bars[0].track - bars[1].track)).toBeLessThan(1);
      expect(Math.abs(bars[0].fill - bars[1].fill)).toBeLessThan(1);
    }
  };
  const beforeLocal = observation.source;
  await page.setViewportSize({ width: 1440, height: 1800 });
  const desktop = await readSnapshot(); check(desktop);
  await page.getByRole('heading', { name: 'Services', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/188c-rpc-service-in-context.png' });
  await dash.screenshot({ path: 'docs/journeys/screenshots/188a-incoming-rpc-dashboard.png' });
  await page.setViewportSize({ width: 390, height: 1600 });
  const mobile = await readSnapshot(); check(mobile);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.getByRole('heading', { name: 'Services', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'docs/journeys/screenshots/188d-mobile-rpc-in-context.png' });
  await dash.locator('[data-widget=spans]').screenshot({ path: 'docs/journeys/screenshots/188b-mobile-rpc-summary.png' });
  expect(observation.source).toBe(beforeLocal); expect(observation.source).toBe(9); expect(observation.plans).toBe(1);
  expect((await snapshot(page, observation)).evidence).toEqual(before.evidence);
  expect(observation.errors).toEqual([]);
  observation.event('unit-label-display-replay', { sourceFreeDisplay: true });
  await write(facts, 'labels.json', { ...facts, plan, bindings, before, desktop, mobile, sourceFreeDisplay: true,
    events: observation.events, requests: observation.requests, responses: observation.responses });
});
