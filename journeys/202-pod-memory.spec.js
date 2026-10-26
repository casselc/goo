const { test, expect } = require('@playwright/test');
const { base, json, observe, seed, discover, assertInventory, canonical, openMetrics, assertRaw, captureBytes, write } = require('./202-pod-memory-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
async function panelSnapshot(panel) {
  return panel.evaluate(e => ({ text: e.textContent, html: e.innerHTML, widgets: [...e.querySelectorAll('.pod-metric-widget')].map(w => ({ key: w.dataset.podMetric,
    text: w.textContent, value: w.querySelector(':scope > p strong')?.textContent, title: w.querySelector(':scope > p strong')?.title,
    timeline: w.querySelector('goo-value-timeline')?.data })) }));
}
async function geometry(panel) {
  const result = await panel.evaluate(e => { const box = e.getBoundingClientRect(), issues = [];
    for (const node of e.querySelectorAll('h4,p,button,goo-value-timeline')) { if (!node.getClientRects().length || node.closest('details:not([open])')) continue; const r = node.getBoundingClientRect(); if (r.left < box.left - 1 || r.right > box.right + 1 || r.top < box.top - 1 || r.bottom > box.bottom + 1) issues.push(node.tagName); }
    return { width: box.width, height: box.height, scrollWidth: e.scrollWidth, clientWidth: e.clientWidth, issues,
      plotHeights: [...e.querySelectorAll('.value-timeline-plot')].map(p => p.getBoundingClientRect().height),
      clocks: [...e.querySelectorAll('.pod-metric-widget > p time')].map(t => ({ text: t.textContent, title: t.title, datetime: t.dateTime })) };
  }); expect(result.issues).toEqual([]); expect(result.scrollWidth).toBeLessThanOrEqual(result.clientWidth + 1);
  for (const height of result.plotHeights) expect(height).toBe(90);
  for (const clock of result.clocks) { expect(clock.text).toContain('UTC'); expect(clock.title).toMatch(/^\d+ ns since Unix epoch$/); expect(clock.datetime).toMatch(/\.\d{9}Z$/); }
  return result;
}
test('202 exact receiver pod memory, strict byte presentation and frozen individual capture', async ({ page, request }) => {
  test.setTimeout(210000); const facts = await seed(request), observation = observe(page), events = [];
  const event = action => events.push({ action, source: observation.source(), plans: observation.plans() });
  const initial = await discover(page, facts); assertInventory(facts, initial.response); event('inventory'); expect(observation.source()).toBe(1);
  const primary = initial.response.observations.find(r => canonical(r.resource) === canonical(facts.primary) && r.resourceSchemaURL === ''); expect(primary).toBeTruthy();
  for (const row of initial.response.observations.filter(r => r.actionable && r.resourceSchemaURL === null)) await expect(page.locator('[data-pod-observe="' + row.key + '"]')).toBeDisabled();
  for (const row of initial.response.observations.filter(r => !r.actionable)) expect(await page.locator('[data-pod-observe="' + row.key + '"]').count()).toBe(0);
  const clusterKey = JSON.stringify([['missing'], ['present', 'shared-cluster']]), namespaceKey = JSON.stringify(['missing']);
  await page.getByLabel('Reported Kubernetes cluster', { exact: true }).selectOption(clusterKey); await page.getByLabel('Reported Kubernetes namespace', { exact: true }).selectOption(namespaceKey);
  await page.getByLabel('Find reported pod resource', { exact: true }).fill('uid-a'); expect(observation.source()).toBe(1);
  const primaryRead = await openMetrics(page, primary); await observation.settle(); event('primary-pair'); expect(observation.source()).toBe(4);
  expect(primaryRead.plan.widgets).toHaveLength(2); expect(primaryRead.plan.widgets.map(w => w.question.definition.source.metric).sort()).toEqual(['k8s.pod.memory.usage','k8s.pod.memory.working_set']);
  for (const widget of primaryRead.plan.widgets) { expect(widget.profile).toBe('collector-kubeletstats-v0.137.0'); expect(widget.question.definition.source.scope).toBe(facts.receiverScope); expect(widget.question.definition.source.version).toBe('0.137.0'); expect(widget.question.definition.source.schemaURL || '').toBe(''); }
  const primaryResults = observation.responses.filter(r => r.endpoint === '/api/measurements' && r.body?.question).map(r => r.body); expect(primaryResults).toHaveLength(2); primaryResults.forEach(r => assertRaw(facts,r));
  const panel = page.locator('.pod-metrics-panel[data-pod-metrics="' + primary.key + '"]');
  const usageWidget = primaryRead.plan.widgets.find(w => w.question.definition.source.metric === 'k8s.pod.memory.usage');
  await expect(panel.locator('[data-pod-metric="' + usageWidget.key + '"] > p strong')).toHaveAttribute('title','32768 By');
  await page.setViewportSize({ width:1440,height:2000 }); const desktop = await geometry(panel); await panel.screenshot({ path:'docs/journeys/screenshots/202a-pod-memory.png' });
  await page.screenshot({ path:'docs/journeys/screenshots/202c-pod-memory-in-context.png' });
  await page.setViewportSize({ width:390,height:2200 }); const mobile = await geometry(panel); await panel.screenshot({ path:'docs/journeys/screenshots/202b-mobile-pod-memory.png' }); expect(observation.source()).toBe(4); await page.setViewportSize({ width:1440,height:1800 });
  const retained = await panelSnapshot(panel), inspect = panel.locator('[data-pod-metric="' + usageWidget.key + '"]').getByRole('button',{ name:'Inspect measurement',exact:true });
  await inspect.focus(); await inspect.evaluate(e => { window.__podMemoryOpener=e; }); await inspect.click(); await expect(page.locator('#measureRun')).toBeEnabled(); expect(observation.source()).toBe(4); event('native-staged');
  expect(await page.locator('[data-native-record-context]').count()).toBe(0);
  await page.getByRole('button',{ name:'Back to Kubernetes',exact:true }).click(); expect(await panelSnapshot(panel)).toEqual(retained); expect(observation.source()).toBe(4); expect(await page.evaluate(() => document.activeElement===window.__podMemoryOpener)).toBe(true); event('retained-back');
  await inspect.click(); await expect(page.locator('#measureRun')).toBeEnabled(); let pending=page.waitForResponse(r=>r.url().endsWith('/api/measurements')&&r.request().method()==='POST');
  await page.locator('#measureRun').click(); const individual=await json(await pending); assertRaw(facts,individual); event('native-run');
  await page.locator('#measureCapture').click(); await page.locator('#captureDraftTitle').fill('Exact reported pod memory usage'); pending=page.waitForResponse(r=>r.url().endsWith('/api/captures')&&r.request().method()==='POST'); await page.locator('#captureDraftConfirm').click(); const saved=await json(await pending); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible();
  const storedBefore=await captureBytes(facts,saved.receipt.id,'capture-before.json'), reopened=await json(await request.get(base+'/api/captures/'+saved.receipt.id)); expect(reopened).toEqual(saved); event('individual-capture');
  const beforeReturn=observation.source(); await page.locator('[data-workspace=services]').click(); expect(await panelSnapshot(panel)).toEqual(retained); expect(observation.source()).toBe(beforeReturn); event('capture-retained-return');
  await page.getByLabel('Find reported pod resource',{exact:true}).fill(''); await page.getByLabel('Reported Kubernetes cluster',{exact:true}).selectOption('*'); await page.getByLabel('Reported Kubernetes namespace',{exact:true}).selectOption('*');
  const boundary=[];
  for (const item of [{uid:'uid-b',state:'observed',value:8192},{uid:'semconv',state:'observed',value:12345,profile:'otel-k8s-v1.37.0'},{uid:'stale',state:'stale'},{uid:'conflict',state:'unknown'},{uid:'missing-data',state:'unknown'},{uid:'negative',state:'invalid',value:-1},{uid:'fractional',state:'invalid',value:1.5},{uid:'unsafe',state:'invalid',value:9007199254740992}]) {
    const row=initial.response.observations.find(r=>r.resource['k8s.pod.uid']===item.uid&&typeof r.resourceSchemaURL==='string'); expect(row).toBeTruthy(); const before=observation.source(), responseIndex=observation.responses.length, read=await openMetrics(page,row); await observation.settle(); expect(observation.source()).toBe(before+2); expect(read.plan.widgets).toHaveLength(1);
    const results=observation.responses.slice(responseIndex).filter(r=>r.endpoint==='/api/measurements'&&r.body?.question).map(r=>r.body); expect(results).toHaveLength(1); assertRaw(facts,results[0]);
    if(item.profile) { expect(read.plan.widgets[0].profile).toBe(item.profile); expect(read.plan.widgets[0].question.definition.source.metric).toBe('k8s.pod.memory.usage'); }
    const target=page.locator('.pod-metrics-panel[data-pod-metrics="'+row.key+'"]'), snapshot=await panelSnapshot(target); expect(snapshot.widgets).toHaveLength(1);
    if(item.state==='invalid') { expect(snapshot.widgets[0].value).toBe('Invalid reported byte count'); expect(snapshot.widgets[0].timeline.points.every(p=>p.state==='unknown'&&p.value===undefined)).toBe(true); expect(results[0].gauges.series[0].samples.at(-1).state).toBe('observed'); }
    else if(item.state==='unknown') { expect(snapshot.widgets[0].value).toBe('Unavailable'); expect(results[0].gauges.series[0].samples.at(-1).state).toBe('unknown'); }
    else { expect(results[0].gauges.series[0].samples.at(-1).state).toBe(item.state); }
    if(item.value!==undefined) expect(results[0].gauges.series[0].samples.at(-1).value).toBe(item.value);
    boundary.push({...item,row,read,results,snapshot}); event('boundary-'+item.uid);
  }
  const beforeFrozen=observation.source(), blocked=[];
  const denyPosts=async route=>{if(route.request().method()==='POST'){blocked.push(new URL(route.request().url()).pathname);await route.abort('internetdisconnected')}else await route.continue()};
  await page.route('**/api/**',denyPosts); await page.goto(base+'/#capture='+saved.receipt.id); await page.reload(); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible(); expect(observation.source()).toBe(beforeFrozen); expect(blocked).toEqual([]); await page.unroute('**/api/**',denyPosts);
  const storedAfter=await captureBytes(facts,saved.receipt.id,'capture-after.json'); expect(storedAfter.sha256).toBe(storedBefore.sha256); event('frozen-reopen-no-posts'); expect(observation.errors).toEqual([]);
  await write(facts,'ledger.json',{facts,initial,primary,primaryRead,primaryResults,desktop,mobile,retained,individual,saved,reopened,storedBefore,storedAfter,boundary,blocked,events,requests:observation.requests,responses:observation.responses});
});
