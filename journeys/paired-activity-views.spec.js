'use strict';
const {test,expect}=require('@playwright/test');
const {randomBytes}=require('node:crypto');
const fs=require('node:fs/promises'),path=require('node:path');
const {sourceBoundary}=require('./support/source-boundary');
const {startRuntime}=require('./191-generic-runtime');
const {source,verifyWithGo,startBombadil}=require('./support/cross-app');
const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));

test('446 save and recall paired Activity without acquiring telemetry',async({page,request})=>{
 test.setTimeout(300000);page.setDefaultTimeout(15000);
 if(!process.env.GOO_GENERIC_BINARY)throw Error('GOO_GENERIC_BINARY must name the final embedded runtime');
 const token=randomBytes(6).toString('hex'),environment='paired446-'+token;
 const dir=path.resolve('artifacts/paired-activity-views446',token);
 const where={all:[{field:{kind:'resource',key:'paired446.bank'},op:'eq',value:token},{any:[{field:{kind:'attribute',key:'cohort'},op:'eq',value:'alpha'},{field:{kind:'attribute',key:'cohort'},op:'missing'}]}]};
 const report={version:'goo.activity-views446/1',status:'running',cleanup:false,errors:[],steps:[],api:[],source:[]};
 await fs.mkdir(dir,{recursive:true});let runtime,meter;
 const shell=()=>page.locator('.stream-entry'),pair=()=>page.locator('.paired-activity');
 const lane=s=>pair().locator('[data-paired-lane="'+s+'"]'),tail=s=>lane(s).locator('.live-tail');
 const rows=s=>tail(s).locator('tbody>tr[data-tail-key]'),options=()=>shell().locator('.activity-view-options');
 const tails=()=>report.api.filter(c=>c.path==='/api/tail');
 const sourcePosts=()=>report.api.filter(c=>['/api/tail','/api/explore','/api/explore/page','/api/discovery','/api/inventory','/api/record-aggregates','/api/explore/counted-values'].includes(c.path));
 page.on('pageerror',e=>report.errors.push(e.message));
 page.on('request',r=>{const u=new URL(r.url());if(r.method()==='POST'&&u.pathname.startsWith('/api/')){let body;try{body=r.postDataJSON()}catch{}report.api.push({path:u.pathname,body})}});
 await page.addInitScript(()=>{window.__views446States={};window.__views446Questions=[];for(const name of ['goo:tail-state','goo:tail-find'])addEventListener(name,e=>{if(e.detail?.owner)window.__views446States[e.detail.owner]=structuredClone(e.detail)});addEventListener('goo:question-state',e=>window.__views446Questions.push(structuredClone(e.detail)))});
 async function snapshot(){const lanes=await page.evaluate(()=>Object.fromEntries(['logs','spans'].map(signal=>{const region=document.querySelector('[data-paired-lane="'+signal+'"]'),root=region?.querySelector('.live-tail'),state=window.__views446States[root?.dataset.owner]||{};return [signal,{phase:root?.dataset.state||'stopped',keys:[...root?.querySelectorAll('tbody>tr[data-tail-key]')||[]].map(r=>r.dataset.tailKey),connection:state.connection||null,find:{text:region?.querySelector('[aria-label="Find loaded records"]')?.value||''},readWindow:state.readWindow||null,selector:state.selector||null,generation:state.generation||0}]})));return {reads:Object.fromEntries(['logs','spans'].map(s=>[s,tails().filter(c=>(c.body.signal||'logs')===s).length])),lanes}}
 async function phase(name,action,run,{signal,local=false}={}){const before=await snapshot(),posts=sourcePosts().length,sql=meter.calls.length;await run();await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));const after=await snapshot();report.steps.push({name,action,...(signal?{lane:signal}:{}),before,after});if(local){expect(sourcePosts().length,name+' source posts').toBe(posts);expect(meter.calls.length,name+' source SQL').toBe(sql)}return {before,after}}
 async function emit(signal,count,label){const at=Date.now()-300,records=Array.from({length:count},(_,i)=>({traceId:randomBytes(16).toString('hex'),spanId:randomBytes(8).toString('hex'),name:label+' '+signal+' '+i,time:String(BigInt(at+i)*1000000n)}));const resource={attributes:attrs({'service.name':'ordinary-worker','deployment.environment.name':environment,'paired446.bank':token})};const data=signal==='logs'?{resourceLogs:[{resource,scopeLogs:[{logRecords:records.map(r=>({timeUnixNano:r.time,traceId:r.traceId,spanId:r.spanId,severityNumber:9,body:{stringValue:r.name},attributes:attrs({cohort:'alpha'})}))}]}]}:{resourceSpans:[{resource,scopeSpans:[{spans:records.map(r=>({traceId:r.traceId,spanId:r.spanId,name:r.name,kind:2,startTimeUnixNano:r.time,endTimeUnixNano:String(BigInt(r.time)+2500000n),attributes:attrs({cohort:'alpha'}),status:{code:1}}))}]}]};const res=await request.post('http://127.0.0.1:4318/v1/'+(signal==='logs'?'logs':'traces'),{data});expect(res.ok(),await res.text()).toBe(true);const table=signal==='logs'?'otel_logs':'otel_traces';await expect.poll(async()=>(await source(request,'SELECT SpanId FROM '+table+' WHERE SpanId={span:String} FORMAT JSONEachRow',{param_span:records[0].spanId})).length,{timeout:45000}).toBe(1);return records}
 async function pauseBoth(){for(const signal of ['logs','spans']){const b=lane(signal).getByRole('button',{name:'Pause stream',exact:true});if(await b.isVisible())await b.click();await expect(tail(signal)).toHaveAttribute('data-state','paused')}}
 async function assertLayouts(saved){for(const signal of ['logs','spans'])expect(await tail(signal).locator('thead th').evaluateAll(xs=>xs.filter(x=>!x.hidden).map(x=>x.dataset.recordCore||x.dataset.recordField).filter(Boolean))).toEqual([...saved.lanes[signal].columns,...(saved.lanes[signal].fieldColumns||[]).map(f=>JSON.stringify([f.kind,f.key]))])}
 async function openViews(){if(!await options().evaluate(e=>e.open))await options().locator(':scope>summary').click()}
 try{
  const initial={logs:await emit('logs',4,'initial'),spans:await emit('spans',3,'initial')};
  meter=await sourceBoundary(process.env.GOO_TEST_CLICKHOUSE_URL||'http://127.0.0.1:8123');
  runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,dir,{clickHouseURL:meter.url});
  const seedResponse=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title:'Single starting scope',selector:{signal:'logs',environment,where},columns:['utc','service','summary']}});
  expect(seedResponse.status(),await seedResponse.text()).toBe(201);const seed=await seedResponse.json();
  await page.goto(runtime.base+'/#activity-view='+seed.id);await expect(shell().getByRole('button',{name:'Start stream',exact:true})).toBeVisible();
  await page.getByRole('combobox',{name:'Live signal',exact:true}).selectOption('both');
  await shell().getByRole('button',{name:/^(Start streams|Apply scope)$/}).click();
  for(const signal of ['logs','spans']){await expect(rows(signal)).toHaveCount(initial[signal].length);expect(await tail(signal).locator('goo-record-summary').evaluateAll(xs=>xs.map(x=>x.record.SpanId).sort())).toEqual(initial[signal].map(r=>r.spanId).sort())}
  await pauseBoth();
  const logsPicker=tail('logs').locator('goo-record-columns');await logsPicker.getByText('Record fields',{exact:true}).click();await logsPicker.getByRole('button',{name:'Show attribute:cohort',exact:true}).click();await logsPicker.getByRole('button',{name:'Move Summary earlier',exact:true}).click();
  const spansPicker=tail('spans').locator('goo-record-columns');await spansPicker.getByText('Record fields',{exact:true}).click();await spansPicker.getByRole('button',{name:'Move Summary earlier',exact:true}).click();await spansPicker.getByRole('button',{name:'Move Summary earlier',exact:true}).click();
  await openViews();await options().getByRole('textbox',{name:'Activity view title',exact:true}).fill('Worker logs and spans');
  const beforeSave={posts:sourcePosts().length,sql:meter.calls.length};const publication=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/activity-views'&&r.request().method()==='POST');await options().getByRole('button',{name:'Save Activity view',exact:true}).click();const published=await publication;expect(published.status(),await published.text()).toBe(201);const saved=report.saved=await published.json();expect(sourcePosts().length).toBe(beforeSave.posts);expect(meter.calls.length).toBe(beforeSave.sql);
  expect(saved.selector).toEqual({signal:'both',environment,where});expect(saved.lanes.logs.columns).toEqual(['utc','summary','service']);expect(saved.lanes.logs.fieldColumns).toEqual([{kind:'attribute',key:'cohort'}]);expect(saved.lanes.spans.columns).toEqual(['summary','utc','service']);
  expect(Object.keys(saved).sort()).toEqual(['id','lanes','selector','title','version']);await options().evaluate(e=>e.open=false);await page.setViewportSize({width:1440,height:1100});await page.screenshot({path:'docs/journeys/screenshots/446-saved-pair-1440.png'});
  // A dirty draft cannot be silently saved as another scope. Recall makes the
  // same explicit keep/discard decision across paired and single views.
  await page.getByRole('textbox',{name:'Live environment',exact:true}).fill(environment+'-draft');await openViews();await expect(options().getByRole('button',{name:'Save Activity view',exact:true})).toBeDisabled();
  const beforeRecall={posts:sourcePosts().length,sql:meter.calls.length,tails:tails().length};
  await options().getByRole('link',{name:'Open saved Activity view',exact:true}).click();await expect(shell().getByRole('button',{name:'Keep current scope',exact:true})).toBeVisible();await shell().getByRole('button',{name:'Keep current scope',exact:true}).click();await expect(page.getByRole('textbox',{name:'Live environment',exact:true})).toHaveValue(environment+'-draft');
  await page.evaluate(id=>{history.replaceState(null,'','#explore');location.hash='#activity-view='+id},saved.id);const recallDiscard=shell().getByRole('button',{name:'Open view and discard draft',exact:true});await recallDiscard.focus();await page.keyboard.press('Enter');await expect(pair()).toBeVisible();await expect(shell().getByRole('button',{name:'Start streams',exact:true})).toBeFocused();
  await page.reload();await expect(shell().getByRole('button',{name:'Start streams',exact:true})).toBeVisible();
  const recalledResponse=await request.get(runtime.base+'/api/activity-views/'+saved.id);expect(recalledResponse.status()).toBe(200);report.recalled=await recalledResponse.json();expect(report.recalled).toEqual(saved);
  for(const signal of ['logs','spans'])await expect(rows(signal)).toHaveCount(0);await assertLayouts(saved);
  await expect(tail('logs').locator('th[data-record-field]')).toContainText('attribute:cohort');
  report.recall={tailPosts:tails().length-beforeRecall.tails,sourceSQL:meter.calls.length-beforeRecall.sql,startVisible:true,rows:{logs:await rows('logs').count(),spans:await rows('spans').count()}};
  expect(sourcePosts().length).toBe(beforeRecall.posts);expect(report.recall).toEqual({tailPosts:0,sourceSQL:0,startVisible:true,rows:{logs:0,spans:0}});
  for(const width of [1440,390,320]){await page.setViewportSize({width,height:1100});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:'docs/journeys/screenshots/446-recalled-pair-'+width+'.png'})}await page.setViewportSize({width:1440,height:1100});
  await emit('logs',2,'recalled');await emit('spans',2,'recalled');report.steps=[];
  await phase('start-recalled-pair','start',async()=>{const b=shell().getByRole('button',{name:'Start streams',exact:true});await b.focus();await page.keyboard.press('Enter');for(const s of ['logs','spans'])await expect(rows(s)).not.toHaveCount(0);await assertLayouts(saved)});
  await phase('pause-recalled-pair','pause',pauseBoth,{local:true});
  const held=await snapshot(),denied=meter.arm('refuse','spans');
  await phase('independent-span-refusal','refuse',async()=>{await lane('spans').getByRole('button',{name:'Resume with recent spans',exact:true}).click();await denied.started;await expect(tail('spans')).toHaveAttribute('data-state','unavailable')},{signal:'spans'});
  expect((await snapshot()).lanes.logs).toEqual(held.lanes.logs);
  await phase('retry-spans','resume',async()=>{await lane('spans').getByRole('button',{name:'Retry',exact:true}).click();await expect(rows('spans')).not.toHaveCount(0);await expect(tail('spans')).not.toHaveAttribute('data-state','unavailable')},{signal:'spans'});
  await phase('pause-spans','pause',pauseBoth,{local:true});
  await phase('inspect-recalled-log','inspect',()=>rows('logs').first().getByRole('button',{name:'Inspect',exact:true}).click(),{signal:'logs',local:true});
  const analyze=pair().getByRole('button',{name:'Analyze observation window · logs',exact:true});
  await phase('fixed-drill-in','navigate',async()=>{await analyze.focus();await page.keyboard.press('Enter');await expect(page.getByRole('button',{name:'Return to paused activity',exact:true})).toBeVisible()},{local:true});
  await phase('retained-return','return',async()=>{await page.getByRole('button',{name:'Return to paused activity',exact:true}).click();await expect(analyze).toBeFocused();await expect(pair()).toBeVisible()},{local:true});
  const beforeSingle={posts:sourcePosts().length,sql:meter.calls.length};await page.evaluate(id=>{location.hash='#activity-view='+id},seed.id);const discard=shell().getByRole('button',{name:'Open view and discard draft',exact:true});await expect.poll(async()=>await page.getByRole('combobox',{name:'Live signal',exact:true}).inputValue()==='logs'||await discard.isVisible()).toBe(true);if(await discard.isVisible())await discard.click();await expect(page.getByRole('combobox',{name:'Live signal',exact:true})).toHaveValue('logs');await expect(shell().getByRole('button',{name:'Start stream',exact:true})).toBeVisible();expect(sourcePosts().length).toBe(beforeSingle.posts);expect(meter.calls.length).toBe(beforeSingle.sql);
  await emit('logs',3,'actor');await emit('spans',3,'actor');
  const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:runtime.base+'/#activity-view='+saved.id,specification:'verification/bombadil/paired-views446.ts',seconds:25,width:390,height:1100});
  try{const campaign=await actor.finish({snapshot:{name:'state',version:'goo.paired-views446/1'}}),complete=campaign.states.find(s=>s.phase===9);expect(complete,'actor saved and recalled actual paired configuration').toBeTruthy();expect(complete).toMatchObject({tailRequests:4,saveRequests:1,forbiddenPosts:0,saved:true,recalled:true,readFreeRecall:true,layoutRestored:true,heldRestored:true,focusRestored:true});report.bombadil={complete,snapshots:campaign.states.length}}finally{await actor.stop()}
  expect(report.errors).toEqual([]);report.status='passed';
 }catch(error){report.status='failed';report.failure=error.message;await page.screenshot({path:path.join(dir,'failure.png')}).catch(()=>{});throw error}
 finally{await page.goto('about:blank').catch(()=>{});await runtime?.stop();await meter?.stop();report.cleanup=runtime?.receipt.cleanup||false;report.source=meter?.calls||[];await fs.writeFile(path.join(dir,'report.json'),JSON.stringify(report,null,2))}
 if(process.env.GOO_ACTIVITY_VIEWS446_VERIFY==='1')await verifyWithGo(dir,'TestQuintActivityViews446Observed',{GOO_ACTIVITY_VIEWS446_REPORT:path.join(dir,'report.json')},{timeoutSeconds:120});
});
