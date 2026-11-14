const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises');
const path=require('node:path');
const {randomBytes}=require('node:crypto');
const {startRuntime}=require('./191-generic-runtime');
const {source,startBombadil}=require('./support/cross-app');
const {identity}=require('../ui/question-session.mjs');

const attrs=object=>Object.entries(object).map(([key,stringValue])=>({key,value:{stringValue}}));
const telemetry=new Set(['/api/tail','/api/explore','/api/explore/page','/api/explore/counted-values','/api/record-aggregates','/api/inventory','/api/discovery','/api/environment-scopes','/api/environment-scopes/records','/api/source-capabilities','/api/metric-scopes','/api/metric-discovery','/api/measurements','/api/trace-search','/api/trace-search/page','/api/population/explain']);

test('432 sparse Activity stages a wider fixed event-time read and returns without restarting',async({page,request})=>{
 test.setTimeout(240000);
 if(!process.env.GOO_GENERIC_BINARY)throw Error('Set GOO_GENERIC_BINARY to the compiled Go binary');
 const token=randomBytes(6).toString('hex'),dir=path.resolve('artifacts/sparse-data432',token);
 await fs.mkdir(dir,{recursive:true});
 const report={version:'goo.sparseData432/1',token,cases:[],calls:[],errors:[]};
 const runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,dir);
 const reads=()=>report.calls.filter(call=>call.method==='POST'&&telemetry.has(call.path)).length;
 const lastTail=()=>structuredClone(report.calls.filter(call=>call.path==='/api/tail'&&call.method==='POST').at(-1).body);
 const post=pathname=>page.waitForResponse(response=>response.request().method()==='POST'&&new URL(response.url()).pathname===pathname);
 page.on('pageerror',error=>report.errors.push(error.message));
 page.on('request',req=>{const pathname=new URL(req.url()).pathname;if(pathname.startsWith('/api/')){let body;try{body=req.postDataJSON()}catch{}report.calls.push({method:req.method(),path:pathname,body});}});
 await page.addInitScript(()=>{window.__sparse432Questions=[];window.addEventListener('goo:question-state',event=>window.__sparse432Questions.push(structuredClone(event.detail)));});
 async function emit(signal,resource,body,at){
  const resourceObject={attributes:attrs({'service.name':'ordinary-sparse432','telemetry.sdk.language':'python',...resource})};
  const value=signal==='logs'?{resourceLogs:[{resource:resourceObject,scopeLogs:[{scope:{name:'ordinary.sparse'},logRecords:[{timeUnixNano:String(BigInt(at)*1000000n),observedTimeUnixNano:String(BigInt(Date.now())*1000000n),severityText:'INFO',body:{stringValue:body}}]}]}]}:
   {resourceSpans:[{resource:resourceObject,scopeSpans:[{scope:{name:'ordinary.sparse'},spans:[{traceId:randomBytes(16).toString('hex'),spanId:randomBytes(8).toString('hex'),name:body,kind:2,startTimeUnixNano:String(BigInt(at)*1000000n),endTimeUnixNano:String(BigInt(at)*1000000n+1000000n),status:{code:1}}]}]}]};
  const response=await request.post('http://127.0.0.1:4318/v1/'+(signal==='logs'?'logs':'traces'),{data:value});expect(response.ok()).toBe(true);
 }
 async function waitRecord(signal,body){
  const table=signal==='logs'?'otel_logs':'otel_traces',column=signal==='logs'?'Body':'SpanName';
  await expect.poll(async()=>Number((await source(request,`SELECT count() AS n FROM default.${table} WHERE ${column}={body:String} FORMAT JSONEachRow`,{param_body:body}))[0].n),{timeout:45000}).toBe(1);
 }
 async function held(shell){
  const tail=shell.locator('.live-tail'),recovery=shell.locator('.activity-recovery');
  const metadata=await recovery.evaluate(element=>({fromMs:Number(element.dataset.pollFromMs),toMs:Number(element.dataset.pollToMs),observedAt:element.dataset.observedAt,rowCount:Number(element.dataset.pollRows),limited:element.dataset.limited==='true'}));
  return {selector:lastTail(),readWindow:{fromMs:metadata.fromMs,toMs:metadata.toMs},observedAt:metadata.observedAt,retained:Number(await tail.getAttribute('data-retained')),phase:await tail.getAttribute('data-state'),metadata};
 }
 try{
  for(const signal of ['logs','spans']){
   const bank=token+'-'+signal,environment='sparse432-'+token;
   const environmentFields=signal==='logs'?{environment}:{environment:'',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'empty'}}};
   const resourceEnvironment=signal==='logs'?{'deployment.environment.name':environment}:{'deployment.environment.name':'','deployment.environment':'ignored-'+token};
   const selector={signal,...environmentFields,where:{all:[{field:{kind:'resource',key:'sparse432.bank'},op:'eq',value:bank}]}};
   const at=Date.now()-5*60000,original=bank+' old event',late=bank+' late arrival';
   await emit(signal,{...resourceEnvironment,'sparse432.bank':bank},original,at);
   await emit(signal,{...resourceEnvironment,'sparse432.bank':'other-'+bank},bank+' wrong filter',Date.now()-1000);
   await emit(signal,{'deployment.environment.name':'other-'+environment,'sparse432.bank':bank},bank+' wrong environment',Date.now()-1000);
   await waitRecord(signal,original);
   const title='Delayed '+signal+' · bank '+bank;
   const saved=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title,selector,columns:['utc','service','summary']}});
   expect(saved.status(),await saved.text()).toBe(201);const view=await saved.json();
   const entry={signal,selector:null,poll:null,steps:[],bootstrap:{viewId:view.id,title},oldEventMs:at};report.cases.push(entry);
   const step=(name,before,query)=>entry.steps.push({name,sourceReads:reads()-before,...(query?{query}:{} )});
   await page.goto('about:blank');let before=reads();await page.goto(runtime.base+'/#activity-view='+view.id);
   const shell=page.locator('.stream-entry'),tail=shell.locator('.live-tail'),recovery=shell.locator('.activity-recovery'),wider=shell.getByRole('button',{name:'Search wider window',exact:true});
   await expect(shell.getByRole('button',{name:'Start stream',exact:true})).toBeVisible();expect(reads()).toBe(before);
   await shell.getByRole('button',{name:'Start stream',exact:true}).click();
   await expect(wider).toBeVisible();await expect(tail).toHaveAttribute('data-retained','0');
   entry.selector=lastTail();expect(entry.selector.where).toEqual(selector.where);
   if(signal==='logs')expect(entry.selector.environment).toBe(environment);else expect(entry.selector.environmentSelector).toEqual(environmentFields.environmentSelector);
   step('completed',before,entry.selector);
   await wider.evaluate(element=>window.__sparse432Old={element,handler:element.onclick});
   before=reads();const filters=shell.locator('.entry-filter-options');await filters.locator(':scope>summary').click();
   const editor=filters.locator('.filter-editor');await editor.getByRole('button',{name:'Add condition',exact:true}).click();
   await expect(wider).toBeHidden();await page.evaluate(()=>window.__sparse432Old.handler.call(window.__sparse432Old.element));expect(reads()).toBe(before);
   const addedLabel=await editor.locator('input[aria-label^="Condition "][aria-label$=" field"]').last().getAttribute('aria-label');const addedID=addedLabel.match(/^Condition (\d+) field$/)[1];await editor.getByRole('button',{name:'Remove condition '+addedID,exact:true}).click();await filters.locator(':scope>summary').click();await expect(wider).toBeVisible();step('dirty_refused',before);
   entry.heldBefore=await held(shell);entry.poll={...entry.heldBefore.metadata,retained:entry.heldBefore.retained};delete entry.heldBefore.metadata;
   expect(entry.poll.retained).toBe(0);expect(entry.poll.rowCount).toBe(0);expect(entry.poll.limited).toBe(false);expect(Number.isFinite(Date.parse(entry.poll.observedAt))).toBe(true);
   if(signal==='logs'){
    const facts=recovery.locator('details');await facts.locator(':scope>summary').click();
    for(const width of [1440,390,320]){await page.setViewportSize({width,height:900});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:`docs/journeys/screenshots/sparse-recovery-entry-${width}.png`});}
    await facts.locator(':scope>summary').click();await page.setViewportSize({width:1440,height:1000});
   }
   before=reads();await wider.focus();await page.keyboard.press('Enter');
   const fixed=page.locator('.activity-analysis').filter({visible:true}),run=fixed.locator('#exploreRun');await expect(fixed).toHaveAttribute('data-question-phase','draft');await expect(run).toBeFocused();expect(reads()).toBe(before);
   const staged=await page.evaluate(()=>window.__sparse432Questions.filter(event=>event.action==='stage').at(-1).requested);
   entry.question=staged;expect(staged.toMs).toBe(entry.poll.toMs);expect(staged.fromMs).toBe(entry.poll.toMs-15*60000);expect(staged.where).toEqual(selector.where);expect(staged.signal).toBe(signal);
   expect(identity(staged)).toBe(identity({...entry.selector,signal,fromMs:entry.poll.toMs-15*60000,toMs:entry.poll.toMs,limit:100,offset:0}));step('stage',before,staged);
   if(signal==='logs')for(const width of [1440,390,320]){await page.setViewportSize({width,height:900});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:`docs/journeys/screenshots/sparse-recovery-staged-${width}.png`});}
   await page.setViewportSize({width:1440,height:1000});
   before=reads();let running=post('/api/explore');await run.click();let response=await running;expect(response.ok(),await response.text()).toBe(true);const first=await response.json();
   const firstRequest=report.calls.filter(call=>call.path==='/api/explore'&&call.method==='POST').at(-1).body;
   expect(identity(firstRequest)).toBe(identity(staged));expect(first.count).toBe('1');expect(first.rows).toHaveLength(1);expect(first.rows[0][signal==='logs'?'Body':'SpanName']).toBe(original);expect(first.state).toBe('observed');
   entry.firstRun={question:firstRequest,count:first.count,capturedAt:first.capturedAt,result:first};step('run',before,firstRequest);
   await emit(signal,{...resourceEnvironment,'sparse432.bank':bank},late,at+1);await waitRecord(signal,late);
   before=reads();running=post('/api/explore');await run.click();response=await running;expect(response.ok(),await response.text()).toBe(true);const second=await response.json();
   const secondRequest=report.calls.filter(call=>call.path==='/api/explore'&&call.method==='POST').at(-1).body;
   expect(identity(secondRequest)).toBe(identity(staged));expect(second.count).toBe('2');expect(second.rows).toHaveLength(2);expect(second.capturedAt).not.toBe(first.capturedAt);
   entry.secondRun={question:secondRequest,count:second.count,capturedAt:second.capturedAt,result:second};step('rerun',before,secondRequest);
   if(signal==='logs')await page.screenshot({path:'docs/journeys/screenshots/sparse-recovery-results-1440.png'});
   before=reads();await page.getByRole('button',{name:'Return to paused activity',exact:true}).click();await expect(tail).toHaveAttribute('data-state','paused');await expect(wider).toBeFocused();await expect(tail).toHaveAttribute('data-retained','0');
   entry.heldAfter=await held(shell);delete entry.heldAfter.metadata;expect(entry.heldAfter.selector).toEqual(entry.heldBefore.selector);expect(entry.heldAfter.readWindow).toEqual(entry.heldBefore.readWindow);expect(entry.heldAfter.observedAt).toBe(entry.heldBefore.observedAt);step('return',before);
   if(signal==='logs'){await page.setViewportSize({width:390,height:900});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:'docs/journeys/screenshots/sparse-recovery-held-390.png'});}
   await page.setViewportSize({width:1440,height:1000});
   before=reads();const resume=shell.getByRole('button',{name:'Resume with recent '+signal,exact:true});await resume.click();await expect(wider).toBeVisible();step('resume',before,lastTail());
   before=reads();await wider.evaluate(element=>window.__sparse432Retire={element,handler:element.onclick});await shell.getByLabel('Live environment',{exact:true}).fill('retired-'+bank);await expect(wider).toBeHidden();await page.evaluate(()=>window.__sparse432Retire.handler.call(window.__sparse432Retire.element));expect(reads()).toBe(before);step('retire',before);
  }
  if(process.env.GOO_SPARSE432_BOMBADIL==='1'){
   const bank=token+'-actor',environment='sparse432-actor-'+token,at=Date.now()-5*60000,body=bank+' old event';
   await emit('logs',{'deployment.environment.name':environment,'sparse432.bank':bank},body,at);await waitRecord('logs',body);
   const made=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title:'Delayed actor logs · bank '+bank,selector:{signal:'logs',environment,where:{all:[{field:{kind:'resource',key:'sparse432.bank'},op:'eq',value:bank}]}},columns:['utc','service','summary']}});expect(made.status()).toBe(201);const view=await made.json();
   const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:runtime.base+'/?sparse432Env='+encodeURIComponent(environment)+'&sparse432Bank='+encodeURIComponent(bank)+'#activity-view='+view.id,specification:'verification/bombadil/sparse-data432.ts',seconds:30,width:390,height:900});
   try{const campaign=await actor.finish({snapshot:{name:'state',version:'goo.sparseRecovery432Actor/1'}}),complete=campaign.states.find(state=>state.phase==='complete');expect(complete).toBeTruthy();expect(complete.localSafe).toBe(true);expect(complete.scopeSafe).toBe(true);expect(complete.forbidden).toEqual([]);expect([complete.tailReads,complete.exploreReads]).toEqual([2,1]);report.bombadil={complete,snapshots:campaign.states.length};}finally{await actor.stop();}
  }
  expect(report.errors).toEqual([]);report.status='passed';
 }catch(error){report.status='failed';report.failure=error.message;report.uiText=await page.locator('.stream-entry').innerText().catch(()=>null);await page.screenshot({path:path.join(dir,'failure.png')}).catch(()=>{});throw error;}
 finally{try{await page.goto('about:blank',{timeout:5000}).catch(()=>{});}finally{await runtime.stop();await fs.writeFile(path.join(dir,'report.json'),JSON.stringify(report,null,2));}}
});
