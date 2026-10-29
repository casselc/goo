const {test,expect}=require('@playwright/test'),fs=require('node:fs/promises'),path=require('node:path'),os=require('node:os'),{execFile}=require('node:child_process'),{promisify}=require('node:util'),{startRuntime}=require('./191-generic-runtime');
test('253 fixed owner preserves checked environment scope across edits without activating selector entry',async({page})=>{
 test.setTimeout(180000);const owned=await fs.mkdtemp(path.join(os.tmpdir(),'goo-scope-253-')),dir=path.resolve('artifacts/record-scope-253');let runtime;
 try{
  const binary=path.join(owned,'goo');await promisify(execFile)('go',['build','-mod=readonly','-o',binary,'./cmd/goo'],{timeout:120000});runtime=await startRuntime(binary,dir);
  const requests=[],errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{const p=new URL(r.url()).pathname;if(r.method()==='POST'&&p==='/api/explore')requests.push(r.postDataJSON())});
  await page.goto(runtime.base+'/#questions');await expect(page.locator('#exploreRun')).toBeVisible();
  // Mount the production fixed owner through its internal API. No new public
  // selector entry is asserted; this is a bounded owner integration test.
  await page.evaluate(async()=>{const {mount}=await import('/exploration.mjs');const host=document.createElement('section');host.id='scope-owner-253';document.body.append(host);window.scope253=mount(host,{prefix:'s253',managed:true})});
  const host=page.locator('#scope-owner-253'),selector=scope=>({version:'goo/environment-selector/v1',policy:'effective',scope});
  for(const scope of [{mode:'unreported'},{mode:'empty'},{mode:'named',value:'local'}]){
   const environmentSelector=selector(scope),toMs=Date.now(),q={signal:'logs',environmentSelector,fromMs:toMs-60000,toMs,limit:100,offset:0};const before=requests.length;
   await page.evaluate(q=>window.scope253.openQuestion(q,[],{stage:true}),q);expect(requests).toHaveLength(before);
   await expect(host.locator('#s253exploreQuestionState')).toContainText(scope.mode==='unreported'?'Unreported environment':scope.mode==='empty'?'Empty environment label':'current/legacy precedence');
   await host.locator('#s253discoveryMessage').fill('owned-no-match-253');expect(requests).toHaveLength(before);
   let response=page.waitForResponse(r=>r.url().endsWith('/api/explore')&&r.request().method()==='POST');await host.locator('#s253exploreRun').click();expect((await response).status()).toBe(200);await expect(host).toHaveAttribute('data-question-phase','applied');expect(requests.at(-1).environmentSelector).toEqual(environmentSelector);expect(requests.at(-1).fromMs).toBe(q.fromMs);expect(requests.at(-1).toMs).toBe(q.toMs);
   await expect(host.locator('#s253exploreCreatePackage')).toBeDisabled();await expect(host.locator('#s253deploymentLoad')).toBeDisabled();
   // Changing signal/fresh bounds must not turn an unreported scope into "".
   await host.locator('#s253exploreSignal').selectOption('spans');await host.locator('#s253exploreFreshWindow').click();response=page.waitForResponse(r=>r.url().endsWith('/api/explore')&&r.request().method()==='POST');await host.locator('#s253exploreRun').click();expect((await response).status()).toBe(200);await expect(host).toHaveAttribute('data-question-phase','applied');expect(requests.at(-1).environmentSelector).toEqual(environmentSelector);expect(requests.at(-1).signal).toBe('spans');
   const phase=await host.getAttribute('data-question-phase');expect(await page.evaluate(q=>{try{window.scope253.openQuestion({...q,environment:'ambiguous'},[],{stage:true});return false}catch{return true}},q)).toBe(true);await expect(host).toHaveAttribute('data-question-phase',phase);
  }
  // A pinned baseline must not override an explicit environment edit.
  const toMs=Date.now();await page.evaluate(q=>window.scope253.openQuestion(q,[],{stage:true}),{signal:'spans',environmentSelector:selector({mode:'unreported'}),fromMs:toMs-60000,toMs,limit:100,offset:0,groupBy:[{kind:'column',key:'service'}]});
  let pinnedRead=page.waitForResponse(r=>r.url().endsWith('/api/explore')&&r.request().method()==='POST');await host.locator('#s253exploreRun').click();expect((await pinnedRead).status()).toBe(200);await expect(host).toHaveAttribute('data-question-phase','applied');await host.locator('#s253baselinePin').click();await expect(host.locator('#s253baselineStatus')).toContainText('Baseline:');
  await host.locator('#s253exploreEnvironment').fill('local');const response=page.waitForResponse(r=>r.url().endsWith('/api/explore')&&r.request().method()==='POST');await host.locator('#s253exploreRun').click();expect((await response).status()).toBe(200);await expect(host).toHaveAttribute('data-question-phase','applied');expect(requests.at(-1).environment).toBe('local');expect(requests.at(-1).environmentSelector).toBeUndefined();
  expect(errors).toEqual([]);await page.evaluate(()=>window.scope253.dispose());
 }finally{await runtime?.stop();await fs.rm(owned,{recursive:true,force:true})}
});
