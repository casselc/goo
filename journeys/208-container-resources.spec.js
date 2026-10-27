const { test, expect } = require('@playwright/test');
const s = require('./208-container-resources-support');
async function frame(panel, file) {
  await panel.evaluate(async e => { e.scrollIntoView({block:'center'}); await document.fonts.ready; await new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))); });
  const geometry = await panel.evaluate(e => {
    const b=e.getBoundingClientRect(), issues=[];
    for(const n of e.querySelectorAll('th,td,time,button,select,strong')) {
      if(!n.getClientRects().length || n.closest('details:not([open])')) continue;
      const r=n.getBoundingClientRect();
      if(n.matches('tbody th')){const range=document.createRange();range.selectNodeContents(n);const lines=[...range.getClientRects()].filter(r=>r.width>0).length;if(lines>2||r.width<60)issues.push({text:n.textContent,lines,width:r.width});}
      if(r.left<b.left-1 || r.right>b.right+1 || r.top<b.top-1 || r.bottom>b.bottom+1 || r.left<0 || r.right>innerWidth || r.top<60 || r.bottom>innerHeight) issues.push({text:n.textContent.slice(0,100),x:r.x,y:r.y,width:r.width,height:r.height});
    }
    return {width:b.width,height:b.height,scrollWidth:e.scrollWidth,clientWidth:e.clientWidth,issues,clocks:[...e.querySelectorAll('time')].map(n=>({text:n.textContent,title:n.title}))};
  });
  expect(geometry.issues).toEqual([]); expect(geometry.scrollWidth).toBeLessThanOrEqual(geometry.clientWidth+1);
  for(const c of geometry.clocks){expect(c.text).toContain('UTC');expect(c.title).toMatch(/ns since Unix epoch/);}
  await panel.screenshot({path:'docs/journeys/screenshots/'+file}); return geometry;
}
test('208 shared container Resources metadata with independent family values and immutable native custody', async ({page,request}) => {
  test.setTimeout(180000);
  const facts=await s.seed(request), o=s.observe(page), events=[];
  const event=action=>events.push({action,source:o.source(),plans:o.plans(),compositions:o.compositions()});
  const initial=await s.discover(page,facts), primary=s.primaryRow(facts,initial); expect(primary).toBeTruthy(); event('inventory');
  const opened=await s.openResources(page,primary), panel=page.locator('[data-container-resources="'+primary.key+'"]'); await o.settle();
  expect(o.source()).toBe(2); expect(o.plans()).toBe(1); expect(o.compositions()).toBe(0);
  await expect(panel.getByLabel('Container resource source variant',{exact:true})).toHaveCount(1);
  expect(await panel.locator('.container-resources-sources').evaluate(e=>e.open)).toBe(false); event('shared-first-metadata');
  const cluster=initial.response.observations.find(r=>r.resourceSchemaURL===facts.schemaURL&&s.canonical(r.resource)===s.canonical(facts.primary));
  const second=await s.discoverVariant(page,panel,cluster.key); await o.settle();
  expect(o.source()).toBe(3); expect(o.plans()).toBe(2); event('shared-second-metadata');
  const memory=await s.loadFamily(page,panel,'memory'); await o.settle();
  expect(memory.response.signals.map(r=>r.value)).toEqual(['32768','65536','131072']);
  expect(memory.response.percentages.request.state).toBe('unavailable'); expect(o.source()).toBe(6); event('load-memory-only');
  const heldMemory=await s.familyPanel(panel,'memory').innerHTML();
  const cpu=await s.loadFamily(page,panel,'cpu'); await o.settle();
  expect(cpu.response.signals.map(r=>r.value)).toEqual(['0.125','0.25','0.5']);
  expect(cpu.response.percentages.request.state).toBe('unavailable'); expect(o.source()).toBe(9);
  expect(await s.familyPanel(panel,'memory').innerHTML()).toBe(heldMemory); event('load-cpu-retains-memory');
  const heldCPU=await s.familyPanel(panel,'cpu').innerHTML();
  const memoryReload=await s.loadFamily(page,panel,'memory'); await o.settle();
  expect(o.source()).toBe(12); expect(await s.familyPanel(panel,'cpu').innerHTML()).toBe(heldCPU); event('reload-memory-retains-cpu');
  const combined=await s.snapshot(panel);
  await page.setViewportSize({width:1440,height:2200}); const desktop=await frame(panel,'208a-container-resources.png');
  await page.screenshot({path:'docs/journeys/screenshots/208c-container-resources-in-context.png'});
  await page.setViewportSize({width:390,height:2800}); const mobile=await frame(panel,'208b-mobile-container-resources.png');
  await page.setViewportSize({width:1440,height:2000});
  const target=s.familyPanel(panel,'cpu'); await s.openConfiguration(target);
  const retained=await panel.innerHTML(), inspect=target.getByRole('button',{name:'Inspect usage',exact:true});
  await inspect.focus(); await inspect.evaluate(e=>window.__resourcesOpener=e);
  await inspect.click(); await expect(page.locator('#measureRun')).toBeEnabled(); expect(o.source()).toBe(12); event('native-stage');
  await page.getByRole('button',{name:'Back to Kubernetes',exact:true}).click();
  expect(await panel.innerHTML()).toBe(retained); expect(await page.evaluate(()=>document.activeElement===window.__resourcesOpener)).toBe(true); event('retained-native-back');
  await inspect.click(); let wait=page.waitForResponse(r=>r.url().endsWith('/api/measurements')&&r.request().method()==='POST');
  await page.locator('#measureRun').click(); const individual=await s.json(await wait); s.assertRaw(facts,individual); event('native-run');
  await page.locator('#measureCapture').click(); await page.locator('#captureDraftTitle').fill('Container Resources individual CPU observation');
  wait=page.waitForResponse(r=>r.url().endsWith('/api/captures')&&r.request().method()==='POST'); await page.locator('#captureDraftConfirm').click();
  const saved=await s.json(await wait); await expect(page.locator('#captureMeasurementEvidence')).toBeVisible();
  const storedBefore=await s.captureBytes(facts,saved.receipt.id,'capture-before.json'); event('individual-capture');
  await page.locator('[data-workspace=services]').click(); expect(await panel.innerHTML()).toBe(retained);
  const alignedRow=initial.response.observations.find(r=>r.resource['container.id']==='aligned'&&r.resourceSchemaURL===facts.schemaURL);
  const alignedOpened=await s.openResources(page,alignedRow), alignedPanel=page.locator('[data-container-resources="'+alignedRow.key+'"]');
  const alignedMemory=await s.loadFamily(page,alignedPanel,'memory'), alignedCPU=await s.loadFamily(page,alignedPanel,'cpu'); await o.settle();
  for(const result of [alignedMemory,alignedCPU]){expect(result.response.percentages.request.value).toBe('50');expect(result.response.percentages.limit.value).toBe('25');}
  const alignedSnapshot=await s.snapshot(alignedPanel), alignedGeometry=await frame(alignedPanel,'208d-aligned-container-resources.png'); event('aligned-both-families');
  const blocked=[], deny=async route=>{if(route.request().method()==='POST'){blocked.push(new URL(route.request().url()).pathname);await route.abort('internetdisconnected');}else await route.continue();};
  const sourceBefore=o.source(); await page.route('**/api/**',deny); await page.goto(s.base+'/#capture='+saved.receipt.id); await page.reload();
  await expect(page.locator('#captureMeasurementEvidence')).toBeVisible(); expect(o.source()).toBe(sourceBefore); expect(blocked).toEqual([]); await page.unroute('**/api/**',deny);
  const reopened=await s.json(await request.get(s.base+'/api/captures/'+saved.receipt.id)), storedAfter=await s.captureBytes(facts,saved.receipt.id,'capture-after.json');
  expect(reopened).toEqual(saved); expect(storedAfter.sha256).toBe(storedBefore.sha256); event('frozen-reopen'); expect(o.errors).toEqual([]);
  await s.write(facts,'ledger.json',{facts,initial,primary,opened,cluster,second,memory,cpu,memoryReload,heldMemory,heldCPU,combined,desktop,mobile,retained,individual,saved,reopened,storedBefore,storedAfter,aligned:{row:alignedRow,opened:alignedOpened,memory:alignedMemory,cpu:alignedCPU,snapshot:alignedSnapshot,geometry:alignedGeometry},blocked,events,loadIntents:s.loadIntents,requests:o.requests,responses:o.responses});
});
