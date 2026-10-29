const {test,expect}=require('@playwright/test'),fs=require('node:fs/promises'),path=require('node:path'),os=require('node:os'),{execFile}=require('node:child_process'),{promisify}=require('node:util'),{startRuntime}=require('./191-generic-runtime');
test('254 population owner restores selector baselines and refuses changed-policy plans before source reads',async({page})=>{
 test.setTimeout(180000);page.setDefaultTimeout(15000);const owned=await fs.mkdtemp(path.join(os.tmpdir(),'goo-population-254-')),dir=path.resolve('artifacts/population-environment-254');let runtime;
 try{
  const binary=path.join(owned,'goo');await promisify(execFile)('go',['build','-mod=readonly','-o',binary,'./cmd/goo'],{timeout:120000});runtime=await startRuntime(binary,dir);
  const requests=[],errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.startsWith('/api/'))requests.push({path:new URL(r.url()).pathname,body:r.postDataJSON()})});
  await page.goto(runtime.base+'/#questions');await expect(page.locator('#exploreRun')).toBeVisible();
  await page.evaluate(async()=>{const {mount}=await import('/population-workspace.mjs'),host=document.createElement('section');host.id='population-owner-254';host.style.marginLeft='260px';host.style.width='1000px';document.body.append(host);window.population254=mount(host,{current:()=>null,table:()=>{},prefix:'p254',active:()=>true})});
  const host=page.locator('#population-owner-254'),toMs=Date.now(),q={signal:'spans',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'unreported'}},fromMs:toMs-60000,toMs,where:{field:{kind:'column',key:'service'},op:'eq',value:'owned-no-match-254'},groupBy:[{kind:'column',key:'service'}],limit:100,offset:0};
  const before=requests.length;await page.evaluate(q=>window.population254.stage([q],{recent:false,breakdowns:{spans:{baseline:{...q,fromMs:q.fromMs-60000,toMs:q.toMs-60000}}}}),q);expect(requests).toHaveLength(before);
  await expect(host.locator('.population-save button')).toBeDisabled();
  await host.getByRole('button',{name:'Run population view',exact:true}).click();await expect.poll(()=>requests.filter(r=>r.path==='/api/explore').length).toBe(2);await expect(host.locator('[id$=baselineStatus]')).toContainText('Baseline: 0 records');
  expect(requests.filter(r=>r.path==='/api/explore').every(r=>JSON.stringify(r.body.environmentSelector)===JSON.stringify(q.environmentSelector))).toBe(true);await expect.poll(()=>requests.some(r=>r.path==='/api/compare')).toBe(true);
  const publications=requests.filter(r=>r.path==='/api/exploration-views').length;await host.locator('.population-save button').evaluate(b=>b.onclick());expect(requests.filter(r=>r.path==='/api/exploration-views')).toHaveLength(publications);
  // A real planner reply is corrupted in transport: no returned question may
  // trigger an automatic source read after policy validation fails.
  await page.route('**/api/population/plan',async route=>{const response=await route.fetch(),data=await response.json();data.queries[0].environmentSelector.scope={mode:'empty'};await route.fulfill({response,json:data})});
  const reads=requests.filter(r=>r.path==='/api/explore').length;await page.evaluate(q=>window.population254.openRecent(q),q);await expect(host.locator('.population-workspace')).toContainText('Population plan changed environment policy');expect(requests.filter(r=>r.path==='/api/explore')).toHaveLength(reads);
  await page.unroute('**/api/population/plan');await page.evaluate(q=>window.population254.openRecent(q),q);await expect.poll(()=>requests.filter(r=>r.path==='/api/explore').length).toBe(reads+2);await expect(host.locator('.population-status')).toContainText('Population applied');
  expect(errors).toEqual([]);await page.evaluate(()=>window.population254.dispose());
 }finally{await runtime?.stop();await fs.rm(owned,{recursive:true,force:true})}
});
