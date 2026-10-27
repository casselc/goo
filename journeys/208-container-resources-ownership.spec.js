const {test,expect}=require('@playwright/test'),s=require('./208-container-resources-support');
test('208 family-local ownership preserves sibling loads and withdraws only affected work',async({page,request})=>{
 test.setTimeout(180000);const facts=await s.seed(request),o=s.observe(page),cases=[];
 for(const reason of ['next-source-close','metadata-auto-close','memory-ambiguity','memory-role','global-close']){
  const initial=await s.discover(page,facts),primary=s.primaryRow(facts,initial);await s.openResources(page,primary);
  const panel=page.locator('[data-container-resources="'+primary.key+'"]'),cpu=s.familyPanel(panel,'cpu'),memory=s.familyPanel(panel,'memory');
  await s.openConfiguration(cpu);
  await s.openConfiguration(memory);
  const expected=reason.startsWith('memory-')||reason==='global-close'?2:1;
  let release,entered,finish;const gate=new Promise(r=>release=r),started=new Promise(r=>entered=r),done=new Promise(r=>finish=r),held=[];let finished=0;
  await page.route('**/api/measurements',async route=>{const response=await route.fetch();held.push({request:route.request().postDataJSON(),status:response.status(),body:await response.json()});if(held.length===expected)entered();await gate;try{await route.fulfill({response});}catch{}finally{if(++finished===expected)finish();}});
  const loadIntents=[];
  for(const [family,target] of [['cpu',cpu],...(expected===2?[['memory',memory]]:[])])loadIntents.push({family,selection:await target.locator('.container-memory-choices select').evaluateAll(ns=>ns.map(n=>({key:n.value,label:n.getAttribute('aria-label')})))});
  await cpu.locator('[data-container-cpu-load]').click();if(expected===2)await memory.locator('[data-container-memory-load]').click();await started;
  for(const h of held)expect(h.status).toBe(200);
  let metadata;
  if(reason==='next-source-close'){
   const sources=panel.locator('.container-resources-sources');await sources.locator('summary').click();
   const cluster=initial.response.observations.find(r=>r.resourceSchemaURL===facts.schemaURL&&s.canonical(r.resource)===s.canonical(facts.primary));
   await panel.getByLabel('Container resource source variant',{exact:true}).selectOption(cluster.key);await sources.locator('summary').click();
  }else if(reason==='metadata-auto-close'){
   const alternative=initial.response.observations.find(r=>r.resource['reported.memory.variant']==='alternative');metadata=await s.discoverVariant(page,panel,alternative.key);
  }else if(reason==='memory-ambiguity'){
   const alternative=initial.response.observations.find(r=>r.resource['reported.memory.variant']==='alternative');metadata=await s.discoverVariant(page,panel,alternative.key);
  }else if(reason==='memory-role')await memory.getByLabel('Container memory working-set',{exact:true}).selectOption('');
  else await panel.getByRole('button',{name:'Close container resources',exact:true}).click();
  const before=await s.snapshot(panel),countsBefore={source:o.source(),plans:o.plans(),compositions:o.compositions()};
  release();await done;await page.unroute('**/api/measurements');await o.settle();
  if(reason==='global-close'){
   await expect(panel).toBeHidden();expect(await s.snapshot(panel)).toEqual(before);expect(o.compositions()).toBe(countsBefore.compositions);
  }else{
   await expect(cpu.locator('.container-cpu-view')).toHaveCount(1);expect(o.compositions()).toBe(countsBefore.compositions+1);
   if(expected===2)await expect(memory.locator('.container-memory-view')).toHaveCount(0);
  }
  cases.push({reason,loadIntents,initial,primary,metadata,held,before,after:await s.snapshot(panel),countsBefore,countsAfter:{source:o.source(),plans:o.plans(),compositions:o.compositions()}});
 }
 expect(o.errors).toEqual([]);await s.write(facts,'ownership.json',{facts,cases,requests:o.requests,responses:o.responses});
});
test('208 shared metadata, pure composition and tab withdrawal reject actual late replies',async({page,request})=>{
 test.setTimeout(120000);const facts=await s.seed(request),o=s.observe(page),cases=[];
 for(const phase of ['metadata','composition','tab']){
  const initial=await s.discover(page,facts),primary=s.primaryRow(facts,initial);
  if(phase!=='metadata')await s.openResources(page,primary);
  const panel=page.locator('[data-container-resources="'+primary.key+'"]');
  let release,entered,finish,held;const gate=new Promise(r=>release=r),started=new Promise(r=>entered=r),done=new Promise(r=>finish=r);
  const pattern=phase==='metadata'?'**/api/pod-metric-discovery':phase==='composition'?'**/api/container-cpu/view':'**/api/measurements';
  await page.route(pattern,async route=>{const response=await route.fetch();held={request:route.request().postDataJSON(),status:response.status(),body:await response.json()};entered();await gate;try{await route.fulfill({response});}catch{}finally{finish();}});
  if(phase==='metadata')await page.locator('[data-container-resources-open="'+primary.key+'"]').click();
  else {const cpu=s.familyPanel(panel,'cpu');await s.openConfiguration(cpu);await cpu.locator('[data-container-cpu-load]').click();}
  await started;expect(held.status).toBe(200);
  if(phase==='metadata')await panel.locator('.container-resources-sources > summary').click();
  else if(phase==='composition')await s.familyPanel(panel,'cpu').getByLabel('Container CPU usage',{exact:true}).selectOption('');
  else await page.locator('#entitiesServices').click();
  const before=await s.snapshot(panel),countsBefore={source:o.source(),plans:o.plans(),compositions:o.compositions()};release();await done;await page.unroute(pattern);await o.settle();
  const after=await s.snapshot(panel);expect(after).toEqual(before);expect({source:o.source(),plans:o.plans(),compositions:o.compositions()}).toEqual(countsBefore);
  cases.push({phase,initial,primary,held,before,after,countsBefore});
 }
 expect(o.errors).toEqual([]);await s.write(facts,'withdrawal.json',{facts,cases,requests:o.requests,responses:o.responses});
});
