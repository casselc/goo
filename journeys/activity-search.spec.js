const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises');
const path=require('node:path');
const {randomBytes}=require('node:crypto');
const {startRuntime}=require('./191-generic-runtime');
const {source,startBombadil}=require('./support/cross-app');

test('434 Find stages literal source messages in the applied fixed window',async({page,request})=>{
 test.setTimeout(180000);
 if(!process.env.GOO_GENERIC_BINARY)throw Error('Set GOO_GENERIC_BINARY to the compiled Go binary');
 const token=randomBytes(6).toString('hex'),dir=path.resolve('artifacts/activity-search434',token);
 await fs.mkdir(dir,{recursive:true});
 const report={version:'goo.activitySearch434/1',token,cases:[],calls:[],errors:[]};
 const runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,dir);
 const sources=new Set(['/api/tail','/api/explore','/api/explore/page','/api/record-aggregates','/api/measurements','/api/inventory','/api/discovery','/api/environment-scopes','/api/environment-scopes/records','/api/source-capabilities','/api/metric-scopes','/api/metric-discovery','/api/trace-search','/api/trace-search/page','/api/explore/counted-values','/api/population/explain']);
 const reads=()=>report.calls.filter(c=>c.method==='POST'&&sources.has(c.path)).length;
 page.on('pageerror',e=>report.errors.push(e.message));
 page.on('request',req=>{const pathname=new URL(req.url()).pathname;if(pathname.startsWith('/api/')){let body;try{body=req.postDataJSON()}catch{}report.calls.push({path:pathname,method:req.method(),body});}});
 await page.addInitScript(()=>{window.__search434Questions=[];window.addEventListener('goo:question-state',e=>window.__search434Questions.push(structuredClone(e.detail)));});
 const attrs=object=>Object.entries(object).map(([key,stringValue])=>({key,value:{stringValue}}));
 async function emit(environment,bank,body,{traceId=randomBytes(16).toString('hex'),age=1000}={}){
  const response=await request.post('http://127.0.0.1:4318/v1/logs',{data:{resourceLogs:[{resource:{attributes:attrs({'service.name':'ordinary-search434','deployment.environment.name':environment,'search434.bank':bank})},scopeLogs:[{scope:{name:'ordinary.search'},logRecords:[{timeUnixNano:String(BigInt(Date.now()-age)*1000000n),severityText:'INFO',traceId,spanId:randomBytes(8).toString('hex'),body:{stringValue:body}}]}]}]}});
  expect(response.ok()).toBe(true);
 }
 async function held(tail){return tail.evaluate(root=>({keys:[...root.querySelectorAll('[data-tail-key]')].map(row=>row.dataset.tailKey),find:root.querySelector('input[type=search]').value,readWindow:{fromMs:Number(root.dataset.lastTo)-60000,toMs:Number(root.dataset.lastTo)},selected:root.dataset.selected}));}
 try{
  const id=randomBytes(16).toString('hex');
  for(const scenario of [{name:'case',term:'Needle434-'+token,body:'Needle434-'+token,expected:'1'},{name:'literal',term:"  雪%_'[.*] "+token+'  ',body:"  雪%_'[.*] "+token+'  ',expected:'1'},{name:'id-only',term:id,body:'message without requested identifier '+token,traceId:id,expected:'0'}]){
   const bank=token+'-'+scenario.name,environment='search434-'+bank;
   await emit(environment,bank,scenario.body,{traceId:scenario.traceId});
   if(scenario.name==='case')await emit(environment,bank,scenario.body.toLowerCase());
   await emit(environment,'outside-'+bank,scenario.term+' outside selector');
   await emit(environment,bank,scenario.term+' outside window',{age:5*60000});
   await expect.poll(async()=>Number((await source(request,'SELECT count() AS n FROM default.otel_logs WHERE ResourceAttributes[\'search434.bank\']={bank:String} FORMAT JSONEachRow',{param_bank:bank}))[0].n),{timeout:45000}).toBe(scenario.name==='case'?3:2);
   const selector={signal:'logs',environment,where:{any:[{field:{kind:'resource',key:'search434.bank'},op:'eq',value:bank},{field:{kind:'column',key:'service'},op:'eq',value:'never-'+bank}]}};
   const response=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title:'Message search · '+scenario.name,selector,columns:['utc','service','summary']}});expect(response.status()).toBe(201);const view=await response.json();
   const entry={scenario:scenario.name,selector,term:scenario.term,steps:[]};report.cases.push(entry);const step=(name,before,query)=>entry.steps.push({name,sourceReads:reads()-before,...(query?{query}:{} )});
   await page.goto('about:blank');await page.goto(runtime.base+'/#activity-view='+view.id);
   const shell=page.locator('.stream-entry'),tail=shell.locator('.live-tail'),find=tail.getByRole('searchbox',{name:'Find loaded records',exact:true}),search=tail.getByRole('button',{name:'Search log messages',exact:true});
   let before=reads();await shell.getByRole('button',{name:'Start stream',exact:true}).click();await expect(tail.locator('[data-tail-key]')).toHaveCount(scenario.name==='case'?2:1);step('loaded',before);
   // Pause once a completed poll exists: later operations have an observable zero-read budget.
   await shell.getByRole('button',{name:'Pause stream',exact:true}).click();before=reads();await find.fill(scenario.term);await expect(search).toBeVisible();expect(reads()).toBe(before);step('find',before);
   entry.heldBefore=await held(tail);entry.localMatched=Number(await tail.getAttribute('data-matched'));expect(entry.localMatched).toBe(scenario.name==='case'?2:1);
   await search.evaluate(element=>window.__search434Old={element,handler:element.onclick});
   before=reads();await shell.getByLabel('Live environment',{exact:true}).fill(environment+'-dirty');await expect(search).toBeHidden();await page.evaluate(()=>window.__search434Old.handler.call(window.__search434Old.element));expect(reads()).toBe(before);step('dirty_refused',before);
   await shell.getByLabel('Live environment',{exact:true}).fill(environment);await expect(search).toBeVisible();
   if(scenario.name==='case')for(const width of [1440,390,320]){await page.setViewportSize({width,height:900});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:`docs/journeys/screenshots/activity-search-${width}.png`});}
   await page.setViewportSize({width:1440,height:1000});
   before=reads();await search.focus();await page.keyboard.press('Enter');
   const fixed=page.locator('.activity-analysis').filter({visible:true}),run=fixed.locator('#exploreRun');await expect(fixed).toHaveAttribute('data-question-phase','draft');await expect(run).toBeFocused();expect(reads()).toBe(before);
   const question=await page.evaluate(()=>window.__search434Questions.filter(e=>e.action==='stage').at(-1).requested);entry.question=question;entry.poll={...entry.heldBefore.readWindow};
   expect(question.where).toEqual({all:[selector.where,{field:{kind:'column',key:'body'},op:'contains',value:scenario.term}]});expect(question.fromMs).toBe(entry.poll.fromMs);expect(question.toMs).toBe(entry.poll.toMs);expect(question.environment).toBe(environment);step('stage',before,question);
   if(scenario.name==='case')await page.screenshot({path:'docs/journeys/screenshots/activity-search-staged-1440.png'});
   before=reads();const running=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/explore');await run.click();const resultResponse=await running;expect(resultResponse.ok(),await resultResponse.text()).toBe(true);const result=await resultResponse.json();expect(result.count).toBe(scenario.expected);expect(result.rows.map(r=>r.Body)).toEqual(scenario.expected==='0'?[]:[scenario.body]);
   const requested=report.calls.filter(c=>c.path==='/api/explore'&&c.method==='POST').at(-1).body;expect(requested.where).toEqual(question.where);expect(requested.fromMs).toBe(question.fromMs);expect(requested.toMs).toBe(question.toMs);entry.firstRun={question:requested,count:result.count,result};step('run',before,requested);
   before=reads();await page.getByRole('button',{name:'Return to paused activity',exact:true}).click();await expect(tail).toHaveAttribute('data-state','paused');await expect(search).toBeFocused();entry.heldAfter=await held(tail);expect(entry.heldAfter).toEqual(entry.heldBefore);expect(reads()).toBe(before);step('return',before);
   before=reads();await search.evaluate(element=>window.__search434Retire={element,handler:element.onclick});await find.fill('');await expect(search).toBeHidden();await page.evaluate(()=>window.__search434Retire.handler.call(window.__search434Retire.element));expect(reads()).toBe(before);step('retire',before);
  }
  if(process.env.GOO_SEARCH434_BOMBADIL==='1'){
   const environment='search434actor-'+token,bank=token+'-actor',term='ActorNeedle434-'+token;
   await emit(environment,bank,term);
   await expect.poll(async()=>Number((await source(request,'SELECT count() AS n FROM default.otel_logs WHERE ResourceAttributes[\'search434.bank\']={bank:String} FORMAT JSONEachRow',{param_bank:bank}))[0].n),{timeout:45000}).toBe(1);
   const made=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title:'Actor message search',selector:{signal:'logs',environment,where:{all:[{field:{kind:'resource',key:'search434.bank'},op:'eq',value:bank}]}},columns:['utc','service','summary']}});expect(made.status()).toBe(201);const view=await made.json();
   await page.goto('about:blank');
   const params=new URLSearchParams({search434Env:environment,search434Bank:bank,search434Term:term});
   const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:runtime.base+'/?'+params+'#activity-view='+view.id,specification:'verification/bombadil/activity-search434.ts',seconds:30,width:390,height:900});
   try{const campaign=await actor.finish({snapshot:{name:'state',version:'goo.activitySearch434Actor/1'}}),complete=campaign.states.find(s=>s.phase==='complete');expect(complete).toBeTruthy();expect(complete.localSafe).toBe(true);expect(complete.scopeSafe).toBe(true);expect(complete.heldSafe).toBe(true);expect(complete.forbidden).toEqual([]);expect([complete.tailReads,complete.exploreReads]).toEqual([1,1]);report.bombadil={complete,snapshots:campaign.states.length};}finally{await actor.stop();}
  }
  expect(report.errors).toEqual([]);report.status='passed';
 }catch(error){report.status='failed';report.failure=error.message;await page.screenshot({path:path.join(dir,'failure.png')}).catch(()=>{});throw error;}
 finally{await page.goto('about:blank').catch(()=>{});await runtime.stop();report.cleanup=runtime.receipt.cleanup;await fs.writeFile(path.join(dir,'report.json'),JSON.stringify(report,null,2));}
});
