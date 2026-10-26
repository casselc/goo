const { test, expect } = require('@playwright/test');
const { observe, seed, discover, assertInventory, canonical, write } = require('./201-kubernetes-support');
test('201 full-width resource evidence and family labels with fresh ordinary telemetry', async ({ page, request }) => {
  test.skip(process.env.GOO_KUBERNETES_DISPLAY !== '1', 'Explicit display replay'); test.setTimeout(90000);
  const facts = await seed(request), observation = observe(page), initial = await discover(page, facts); assertInventory(facts, initial.response);
  const primary = initial.response.observations.find(r => canonical(r.resource) === canonical(facts.primary) && r.resourceSchemaURL === facts.schemaURL); expect(primary).toBeTruthy();
  const cluster = page.getByLabel('Reported Kubernetes cluster', { exact: true }), namespace = page.getByLabel('Reported Kubernetes namespace', { exact: true }), search = page.getByLabel('Find reported pod resource', { exact: true });
  const clusterKey = JSON.stringify([['missing'], ['present', 'shared-cluster']]), namespaceKey = JSON.stringify(['missing']);
  await cluster.selectOption(clusterKey); await namespace.selectOption(namespaceKey); await search.fill('uid-a');
  const target = page.locator('tr[data-pod-variant="' + primary.key + '"]'), evidence = page.locator('tr[data-pod-evidence="' + primary.key + '"]'), summary = evidence.locator('details > summary');
  await summary.click(); await expect(evidence.locator('time').first()).toBeVisible();
  for (const label of ['Logs: 2', 'Gauge points: 1', 'Sum points: 1', 'Histogram points: 1']) await expect(evidence).toContainText(label);
  expect(observation.source()).toBe(1); const snapshots = [];
  for (const [width, height, frame] of [[1440, 2600, '201a-pod-resource-evidence.png'], [390, 2800, '201b-mobile-pod-resource.png']]) {
    await page.setViewportSize({ width, height });
    const snapshot = await evidence.evaluate(e => {
      const rect = e.getBoundingClientRect(), table = e.closest('table').getBoundingClientRect(), issues = [];
      for (const node of e.querySelectorAll('td,p,small,strong,a,time,input,select,button')) {
        if (!node.getClientRects().length || node.closest('details:not([open])')) continue;
        const r = node.getBoundingClientRect(); if (r.left < rect.left - 1 || r.right > rect.right + 1 || r.top < rect.top - 1 || r.bottom > rect.bottom + 1) issues.push({ tag: node.tagName, text: node.textContent?.slice(0, 60) });
      }
      const values = [...e.querySelectorAll('goo-field-value')].map(node => ({ text: node.textContent, height: node.getBoundingClientRect().height, width: node.getBoundingClientRect().width }));
      return { width: rect.width, height: rect.height, tableWidth: table.width, issues, values, scrollWidth: e.scrollWidth, clientWidth: e.clientWidth,
        clocks: [...e.querySelectorAll('time')].map(t => ({ text: t.textContent, title: t.title, datetime: t.dateTime })), text: e.textContent };
    });
    expect(Math.abs(snapshot.width - snapshot.tableWidth)).toBeLessThanOrEqual(2); expect(snapshot.issues).toEqual([]); expect(snapshot.scrollWidth).toBeLessThanOrEqual(snapshot.clientWidth + 1);
    expect(snapshot.values).toHaveLength(Object.keys(facts.primary).length);
    for (const value of snapshot.values) expect(value.height).toBeLessThan(90);
    expect(snapshot.clocks).toHaveLength(8); for (const clock of snapshot.clocks) { expect(clock.text).toContain('UTC'); expect(clock.title).toMatch(/^\d+ ns since Unix epoch$/); expect(clock.datetime).toMatch(/\.\d{9}Z$/); }
    snapshots.push({ width, height, snapshot }); await evidence.screenshot({ path: 'docs/journeys/screenshots/' + frame });
  }
  await summary.click(); await summary.click(); expect(observation.source()).toBe(1);
  const retained = await page.locator('.kubernetes-browser').innerHTML(), link = target.locator('[data-pod-signal=logs]');
  await link.focus(); await link.evaluate(e => { window.__podDisplayOpener = e; }); await link.click(); await expect(page.locator('#exploreRun')).toBeEnabled(); expect(observation.source()).toBe(1);
  await page.getByRole('button', { name: 'Back to Kubernetes', exact: true }).click();
  expect(await page.locator('.kubernetes-browser').innerHTML()).toBe(retained); expect(observation.source()).toBe(1); expect(await page.evaluate(() => document.activeElement === window.__podDisplayOpener)).toBe(true);
  await expect(cluster).toHaveValue(clusterKey); await expect(namespace).toHaveValue(namespaceKey); await expect(search).toHaveValue('uid-a');
  await page.setViewportSize({ width: 1440, height: 2600 }); await target.scrollIntoViewIfNeeded(); await page.screenshot({ path: 'docs/journeys/screenshots/201c-kubernetes-in-context.png' });
  expect(observation.source()).toBe(1); expect(observation.errors).toEqual([]);
  await write(facts, 'display.json', { facts, initial, primary, snapshots, source: observation.source(), requests: observation.requests, responses: observation.responses });
});
