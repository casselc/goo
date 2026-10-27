const {test,expect}=require('@playwright/test'),s=require('./209-container-history-support');
test.beforeEach(()=>test.skip(true,'Historical manual Resources entrance; current automatic entry and ownership are exercised in journey 210. Native/API/property/raw contracts remain active.'));
test('209 actual delayed pure history respects family and global owners',async({page,request})=>{
 test.setTimeout(150000);const facts=await s.seed(request),o=s.observe(page),cases=[];
 for(const reason of ['role','disclosure','scope','close','sibling-load-survives']){
  const initial=await s.discover(page,facts),primary=s.primaryRow(facts,initial);await s.openResources(page,primary);const panel=page.locator('[data-container-resources="'+primary.key+'"]'),memory=s.familyPanel(panel,'memory'),cpu=s.familyPanel(panel,'cpu');
  let release,entered,finish,held;const gate=new Promise(r=>release=r),started=new Promise(r=>entered=r),done=new Promise(r=>finish=r);
  await page.route('**/api/container-resources/history',async route=>{
   const payload=route.request().postDataJSON();if(!payload.memory){await route.continue();return;}
   const response=await route.fetch();held={request:payload,status:response.status(),body:await response.json()};entered();await gate;try{await route.fulfill({response});}catch{}finally{finish();}
  });
  await s.openConfiguration(memory);await memory.locator('[data-container-memory-load]').click();await started;expect(held.status).toBe(200);const heldCPU=await cpu.innerHTML(),countsHeld={source:o.source(),plans:o.plans(),compositions:o.compositions(),histories:o.histories()};let sibling;
  if(reason==='role')await memory.getByLabel('Container memory working-set',{exact:true}).selectOption('');
  else if(reason==='disclosure')await memory.locator('.container-memory-configuration > summary').click();
  else if(reason==='close')await panel.getByRole('button',{name:'Close container resources',exact:true}).click();
  else if(reason==='scope')await page.getByLabel('Kubernetes environment',{exact:true}).fill(facts.environment+'-changed');
  else sibling=await s.loadFamily(page,panel,'cpu');
  const snapshotOwner=async()=>await panel.count()?s.snapshot(panel):{present:false};
  const before=await snapshotOwner();release();await done;await page.unroute('**/api/container-resources/history');await o.settle();
  if(reason==='sibling-load-survives'){
   await expect(s.history(panel,'memory')).toHaveCount(1);await expect(s.history(panel,'cpu')).toHaveCount(1);expect(o.source()).toBe(countsHeld.source+1);expect(o.histories()).toBe(countsHeld.histories+1);
  }else{
   await expect(s.history(panel,'memory')).toHaveCount(0);expect(await snapshotOwner()).toEqual(before);expect(o.source()).toBe(countsHeld.source);expect(o.histories()).toBe(countsHeld.histories);
   if(reason!=='scope')expect(await cpu.innerHTML()).toBe(heldCPU);
  }
  cases.push({reason,initial,primary,held,heldCPU,countsHeld,sibling,before,after:await snapshotOwner(),countsAfter:{source:o.source(),plans:o.plans(),compositions:o.compositions(),histories:o.histories()}});
 }
 expect(o.errors).toEqual([]);await s.write(facts,'ownership.json',{facts,cases,requests:o.requests,responses:o.responses});
});
