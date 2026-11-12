'use strict';
const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path');
const {questionLink}=require('../ui/workspace.mjs');
const {identity}=require('../ui/question-session.mjs');
const F=require('./support/activity-log-context');
const {startRuntime}=require('./191-generic-runtime');
const {startSourceFaultProxy}=require('./support/source-fault-proxy');
const {source,startBombadil}=require('./support/cross-app');
const settle=page=>page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));

test('Recent and pinned scopes reuse intent separately from retained results',async({page,request})=>{
 test.setTimeout(180000);page.setDefaultTimeout(15000);
 const dir=path.resolve('artifacts/recent-pinned-scopes'),shots=path.resolve('docs/journeys/screenshots');await fs.mkdir(dir,{recursive:true});await fs.mkdir(shots,{recursive:true});
 const report={version:'goo.recentPinnedScopes/1',calls:[],facts:[],errors:[]};let runtime,proxy;
 page.on('pageerror',error=>report.errors.push(error.message));
 page.on('request',r=>{const name=new URL(r.url()).pathname;if(!name.startsWith('/api/'))return;let body;try{body=r.postDataJSON()}catch{}report.calls.push({path:name,method:r.method(),request:body})});
 const counters=()=>({posts:report.calls.filter(c=>c.method==='POST').length,sql:proxy.calls.length});
 const local=async(name,action)=>{await expect.poll(()=>proxy.calls.every(c=>c.closed)).toBe(true);const before=counters();await action();await settle(page);expect(counters(),name+' must not read telemetry').toEqual(before);report.facts.push({name,before,after:counters()})};
 const picker=()=>page.locator('.scope-history');
 const openPicker=async()=>{if(!await picker().evaluate(n=>n.open))await picker().locator(':scope>summary').click()};
 const stage=async q=>{await page.evaluate(hash=>{const link=document.createElement('a');link.href=hash;link.textContent='Open scoped question';document.body.append(link);link.click();link.remove()},questionLink(q));await expect(page.locator('#exploreStatus')).toContainText('Fixed question staged.')};
 const rows=()=>page.locator('#exploreRows .record-row');
 const read=async q=>{const next=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/explore');await page.locator('#exploreRun').click();const response=await next,value=await response.json();expect(response.ok()).toBe(true);expect(identity(value.question)).toBe(identity(q));expect(value.count).toBe(String(F.members(report.raw,q).length));await expect(rows()).toHaveCount(Math.min(q.limit,F.members(report.raw,q).length));await expect.poll(()=>proxy.calls.every(c=>c.closed)).toBe(true);report.reads??=[];report.reads.push(value);return value};
 try{
  proxy=await startSourceFaultProxy();runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,dir,{clickHouseURL:proxy.url});await page.goto(runtime.base+'/#questions');await expect(page.locator('#exploreRun')).toBeVisible();
  const f=F.fixture();report.fixture=f;expect((await request.post('http://127.0.0.1:4318/v1/logs',{data:F.logs(f)})).ok()).toBe(true);
  const raw=()=>source(request,"SELECT ServiceName,SeverityText,SeverityNumber,Body,ResourceAttributes,LogAttributes,toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns FROM otel_logs WHERE LogAttributes['activity.bank']={bank:String} FORMAT JSONEachRow",{param_bank:f.token,max_result_rows:2000});await expect.poll(async()=>(await raw()).length,{timeout:45000}).toBe(f.rows.length);report.raw=await raw();
  const q={...structuredClone(f.selector),fromMs:Math.min(...f.rows.map(r=>r.timeMs))-1,toMs:Math.max(...f.rows.map(r=>r.timeMs))+1,limit:100,offset:0},q2={...structuredClone(q),where:{all:[q.where,{field:{kind:'column',key:'severity'},op:'eq',value:'INFO'}]}};report.questions=[q,q2];
  await local('stage-first',()=>stage(q));await read(q);expect(report.reads[0].count).toBe('624');
  await openPicker();
  // Scope rows and their controls are supplied by the shared application picker.
  const recent=()=>picker().locator('[data-pinned="false"][data-kind="question"]');
  const pinned=()=>picker().locator('[data-pinned="true"][data-kind="question"]');
  await expect(recent()).toHaveCount(1);
  await local('pin',async()=>{await recent().getByRole('button',{name:'Pin',exact:true}).click();await expect(pinned()).toHaveCount(1)});
  await local('stage-other',()=>stage(q2));await read(q2);expect(report.reads[1].count).toBe('156');
  await openPicker();
  const held=picker().locator('[data-visit-id]').first();
  await local('held-return',async()=>{await held.click();await expect(rows()).toHaveCount(100);expect(await rows().locator('goo-record-summary').evaluateAll(nodes=>nodes.map(n=>n.record))).toEqual(report.reads[0].rows)});
  const viewResponse=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title:'Reusable log scope',selector:f.selector,columns:['utc','service','summary']}});expect(viewResponse.status()).toBe(201);const savedActivity=await viewResponse.json();report.activityView=savedActivity;
  await local('stage-saved-activity',async()=>{await page.evaluate(id=>{location.hash='activity-view='+id},savedActivity.id);await expect(page.getByRole('region',{name:'Live activity',exact:true}).locator('[aria-label="Activity view title"]')).toHaveValue('Reusable log scope');await expect(page.getByRole('button',{name:'Start stream',exact:true})).toBeVisible()});
  await openPicker();const pinnedActivity=()=>picker().locator('[data-pinned="true"][data-kind="activity"]');
  await local('pin-activity',async()=>{await picker().getByRole('button',{name:'Pin current scope',exact:true}).click();await expect(pinnedActivity()).toHaveCount(1)});
  const activity=()=>page.getByRole('region',{name:'Live activity',exact:true}),liveSignal=()=>activity().getByRole('combobox',{name:'Live signal',exact:true});
  await local('dirty-activity-refusal',async()=>{await liveSignal().selectOption('spans');await pinnedActivity().getByRole('button',{name:'Stage',exact:true}).click();await expect(picker().getByRole('status')).toContainText('Apply or clear Activity changes');await expect(liveSignal()).toHaveValue('spans');await liveSignal().selectOption('logs')});
  await local('stage-clean-activity',async()=>{await pinnedActivity().getByRole('button',{name:'Stage',exact:true}).click();await expect(activity().getByRole('button',{name:'Start stream',exact:true})).toBeVisible()});
  const beforeStream=counters();await activity().getByRole('button',{name:'Start stream',exact:true}).click();await expect(activity().locator('.live-tail tbody>tr')).toHaveCount(100);await activity().getByRole('button',{name:'Pause stream',exact:true}).click();await expect(activity().locator('.live-tail')).toHaveAttribute('data-state','paused');await expect.poll(()=>proxy.calls.every(c=>c.closed)).toBe(true);expect(counters().posts).toBeGreaterThan(beforeStream.posts);expect(counters().sql).toBeGreaterThan(beforeStream.sql);report.facts.push({name:'explicit-stream',before:beforeStream,after:counters()});
  await openPicker();await local('held-return-from-activity',async()=>{await picker().locator('[data-visit-id]').first().click();await expect(rows()).toHaveCount(100);expect(await rows().locator('goo-record-summary').evaluateAll(nodes=>nodes.map(n=>n.record))).toEqual(report.reads[0].rows)});
  await openPicker();
  await local('stage-pinned',async()=>{await pinned().getByRole('button',{name:'Stage',exact:true}).click();await expect(page.locator('#exploreStatus')).toContainText('Fixed question staged.');await expect(rows()).toHaveCount(0)});
  await openPicker();await local('recall-rolling-activity',async()=>{await pinnedActivity().getByRole('button',{name:'Stage',exact:true}).click();await expect(page.getByRole('button',{name:'Start stream',exact:true})).toBeVisible();await expect(page.getByRole('region',{name:'Live activity',exact:true}).locator('.live-tail')).toHaveAttribute('data-state','paused')});
  await openPicker();await local('return-to-fixed-intent',async()=>{await pinned().getByRole('button',{name:'Stage',exact:true}).click();await expect(page.locator('#exploreStatus')).toContainText('Fixed question staged.')});
  await openPicker();
  for(const width of [1440,390,320]){await page.setViewportSize({width,height:width===1440?1000:844});await picker().scrollIntoViewIfNeeded();expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:path.join(shots,'recent-pinned-scopes-'+width+'.png')})}
  await page.setViewportSize({width:1440,height:1000});
  const remember=picker().getByRole('checkbox',{name:'Remember scopes on this browser',exact:true});await expect(remember).not.toBeChecked();
  report.defaultStorage=await page.evaluate(()=>Object.keys(localStorage).filter(k=>/scope/i.test(k)));expect(report.defaultStorage).toEqual([]);
  await local('opt-in',async()=>{await remember.check();await expect(remember).toBeChecked()});
  await local('reload-configuration',async()=>{await page.reload();await expect(picker()).toBeVisible();await openPicker();await expect(pinned()).toHaveCount(1);const freshVisit=picker().locator('[data-visit-id]');await expect(freshVisit).toHaveCount(1);await expect(freshVisit).toBeDisabled();await expect(rows()).toHaveCount(0);await expect(page.locator('#exploreStatus')).toContainText('Fixed question staged.');await expect(picker().getByRole('checkbox',{name:'Remember scopes on this browser',exact:true})).toBeChecked()});
  report.persisted=await page.evaluate(()=>Object.entries(localStorage).filter(([k])=>/scope/i.test(k)));expect(report.persisted.length).toBeGreaterThan(0);for(const [,value]of report.persisted){expect(value).not.toContain('ordinary captured work');expect(value).not.toContain('timestamp_ns');expect(value).not.toContain('ResourceAttributes')}
  await local('clear-history',async()=>{await picker().getByRole('button',{name:'Clear recent',exact:true}).click();await expect(recent()).toHaveCount(0);await expect(pinned()).toHaveCount(1)});
  await local('unpin',async()=>{await pinned().getByRole('button',{name:'Unpin',exact:true}).click();await expect(pinned()).toHaveCount(0)});
  await local('disable-persistence',async()=>{await picker().getByRole('checkbox',{name:'Remember scopes on this browser',exact:true}).uncheck()});
  if(process.env.GOO_SCOPE_HISTORY_BOMBADIL==='1'){
   const actorDir=path.join(dir,'bombadil-'+f.token),before=counters(),actor=await startBombadil({dir:actorDir,url:runtime.base+'/'+questionLink(q),specification:'verification/bombadil/recent-pinned-scopes.ts',seconds:30,width:390,height:844});
   try{await actor.finish();const trace=(await fs.readFile(path.join(actorDir,'trace.jsonl'),'utf8')).trim().split('\n').map(JSON.parse),states=trace.flatMap(line=>line.snapshots||[]).map(s=>s.value).filter(s=>s?.version==='goo.scopeHistoryActor/1');expect(states.some(s=>s.phase==='complete')).toBe(true);expect(counters()).toEqual(before);report.bombadil={snapshots:states.length,complete:true,before,after:counters()}}finally{await actor.stop()}
  }
  expect(report.errors).toEqual([]);
 }finally{report.sql=proxy?.calls;await fs.writeFile(path.join(dir,'report.json'),JSON.stringify(report,null,2));await runtime?.stop();await proxy?.stop()}
});
