const { test, expect } = require('@playwright/test');
const { observe, seed, write, discover, assertInventory, canonical, json } = require('./201-kubernetes-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
async function rowGeometry(row) {
  const geometry = await row.evaluate(e => {
    const box = e.getBoundingClientRect(), issues = [];
    for (const node of e.querySelectorAll('td,p,small,strong,a,time,input,select,button')) {
      if (!node.getClientRects().length || node.closest('details:not([open])')) continue;
      const r = node.getBoundingClientRect();
      if (r.left < box.left - 1 || r.right > box.right + 1 || r.top < box.top - 1 || r.bottom > box.bottom + 1) issues.push({ tag: node.tagName, text: node.textContent?.slice(0, 80) });
    }
    return { width: box.width, height: box.height, scrollWidth: e.scrollWidth, clientWidth: e.clientWidth, issues,
      clocks: [...e.querySelectorAll('time')].map(t => ({ text: t.textContent, title: t.title, datetime: t.dateTime })) };
  });
  expect(geometry.issues).toEqual([]); expect(geometry.scrollWidth).toBeLessThanOrEqual(geometry.clientWidth + 1);
  expect(geometry.clocks.length).toBeGreaterThan(0);
  for (const clock of geometry.clocks) { expect(clock.text).toContain('UTC'); expect(clock.title).toMatch(/^\d+ ns since Unix epoch$/); expect(clock.datetime).toMatch(/\.\d{9}Z$/); }
  return geometry;
}
test('201 full raw Kubernetes resource variants and UID-qualified record context', async ({ page, request }) => {
  test.setTimeout(150000); const facts = await seed(request), observation = observe(page), events = [];
  const event = action => events.push({ action, source: observation.source(), plans: observation.plans() });
  const initial = await discover(page, facts); assertInventory(facts, initial.response); event('discovery'); expect(observation.source()).toBe(1);
  const rows = initial.response.observations;
  expect(rows.some(r => r.podUIDState === 'missing' && !r.actionable)).toBe(true); expect(rows.some(r => r.podUIDState === 'empty' && !r.actionable)).toBe(true);
  for (const state of ['missing', 'empty']) expect(await page.locator('tr[data-pod-uid-state="' + state + '"] [data-pod-signal]').count()).toBe(0);
  expect(rows.some(r => r.resourceSchemaURL === null && r.families.some(f => f.family === 'spans'))).toBe(true);
  expect(rows.some(r => r.resourceSchemaURL === '' && r.families.some(f => f.family === 'gauge'))).toBe(true);
  const clusterKey = JSON.stringify([['missing'], ['present', 'shared-cluster']]), namespaceKey = JSON.stringify(['missing']);
  const cluster = page.getByLabel('Reported Kubernetes cluster', { exact: true }), namespace = page.getByLabel('Reported Kubernetes namespace', { exact: true }), search = page.getByLabel('Find reported pod resource', { exact: true });
  await cluster.selectOption(clusterKey); await namespace.selectOption(namespaceKey); await search.fill('uid-a');
  const filtered = rows.filter(r => !Object.hasOwn(r.resource, 'k8s.cluster.uid') && !Object.hasOwn(r.resource, 'k8s.namespace.name') && Object.values(r.resource).some(v => v.toLowerCase().includes('uid-a')));
  await expect(page.locator('tr[data-pod-variant]')).toHaveCount(filtered.length); expect(observation.source()).toBe(1); event('local-filters');
  const primary = rows.find(r => canonical(r.resource) === canonical(facts.primary) && r.resourceSchemaURL === facts.schemaURL); expect(primary).toBeTruthy();
  const target = page.locator('tr[data-pod-variant="' + primary.key + '"]');
  const evidence = page.locator('tr[data-pod-evidence="' + primary.key + '"]');
  await evidence.locator('details > summary').click(); await expect(evidence.locator('time').first()).toBeVisible(); expect(observation.source()).toBe(1);
  await page.setViewportSize({ width: 1440, height: 2600 }); const desktop = await rowGeometry(evidence);
  await evidence.screenshot({ path: 'docs/journeys/screenshots/201a-pod-resource-evidence.png' });
  await page.screenshot({ path: 'docs/journeys/screenshots/201c-kubernetes-in-context.png' });
  await page.setViewportSize({ width: 390, height: 2800 }); const mobile = await rowGeometry(evidence);
  await evidence.screenshot({ path: 'docs/journeys/screenshots/201b-mobile-pod-resource.png' });
  expect(observation.source()).toBe(1); await page.setViewportSize({ width: 1440, height: 1800 });
  const retained = await page.locator('.kubernetes-browser').innerHTML(), context = [];
  for (const signal of ['logs', 'spans']) {
    const link = target.locator('[data-pod-signal="' + signal + '"]'); await link.focus(); await link.evaluate(e => { window.__podContextOpener = e; });
    const before = observation.source(); await link.click(); await expect(page.locator('#exploreRun')).toBeEnabled(); expect(observation.source()).toBe(before); event(signal + '-staged');
    const pending = page.waitForResponse(r => r.url().endsWith('/api/explore') && r.request().method() === 'POST');
    await page.locator('#exploreRun').click(); const response = await pending, result = await json(response), question = response.request().postDataJSON();
    const predicates = question.where.all; expect(predicates.filter(p => p.op === 'missing').map(p => p.field.key).sort()).toEqual(['k8s.cluster.uid', 'k8s.namespace.name']);
    expect(predicates.some(p => p.field.key.startsWith('service.'))).toBe(false);
    for (const [key, value] of Object.entries(facts.primary)) expect(predicates).toContainEqual({ field: { kind: 'resource', key }, op: 'eq', value });
    const members = facts.records.filter(r => r.family === signal && Object.entries(facts.primary).every(([k, v]) => r.resource[k] === v) && !Object.hasOwn(r.resource, 'k8s.cluster.uid') && !Object.hasOwn(r.resource, 'k8s.namespace.name'));
    expect(result.count).toBe(String(members.length)); expect(result.rows).toHaveLength(members.length);
    const actualMembers = result.rows.map(r => r.Attributes?.['work.id'] || r.LogAttributes?.['work.id'] || r.SpanAttributes?.['work.id']);
    expect(actualMembers.sort()).toEqual(members.map(r => r.id).sort());
    expect(result.rows.some(r => r.ResourceAttributes['service.name'] === 'pod-context-app')).toBe(true);
    expect(question.fromMs).toBe(initial.request.fromMs); expect(question.toMs).toBe(initial.request.toMs);
    context.push({ signal, question, result, expectedMembers: members.map(r => r.id).sort() }); event(signal + '-run');
    const afterRead = observation.source(); await page.getByRole('button', { name: 'Back to Kubernetes', exact: true }).click();
    expect(await page.locator('.kubernetes-browser').innerHTML()).toBe(retained); expect(observation.source()).toBe(afterRead);
    await expect(cluster).toHaveValue(clusterKey); await expect(namespace).toHaveValue(namespaceKey); await expect(search).toHaveValue('uid-a');
    expect(await page.evaluate(() => document.activeElement === window.__podContextOpener)).toBe(true); event(signal + '-retained-back');
  }
  expect(observation.source()).toBe(3); expect(observation.errors).toEqual([]);
  await write(facts, 'ledger.json', { facts, initial, primary, clusterKey, namespaceKey, filteredKeys: filtered.map(r => r.key), desktop, mobile, retained, context, events, requests: observation.requests, responses: observation.responses });
});
