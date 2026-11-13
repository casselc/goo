'use strict';
const {test,expect}=require('@playwright/test'),fs=require('node:fs/promises'),path=require('node:path');
const F=require('./support/activity-facets'),P=require('./support/activity-snapshot-profile');
const {startRuntime}=require('./191-generic-runtime'),{startSourceFaultProxy}=require('./support/source-fault-proxy'),{source}=require('./support/cross-app'),{revealControl}=require('./support/explore');
const telemetry=new Set(['/api/tail','/api/explore','/api/explore/page','/api/explore/counted-values','/api/record-aggregates','/api/measurements']);
const checkbox=(v,value)=>v.counted().getByRole('checkbox',{name:'Select counted attribute:cohort '+(value===null?'Missing key':value===''?'Present empty string':JSON.stringify(value)),exact:true});
test('423 profile actual Activity inspection and counted selection row-copy work',async({page,request})=>{
 test.setTimeout(300000);page.setDefaultTimeout(15000);const label=process.env.GOO_SNAPSHOT_PROFILE_LABEL||'candidate';if(!/^[a-z0-9-]+$/.test(label))throw Error('Use a simple profile label');
 const dir=path.resolve('artifacts/activity-snapshot-profile423',label);await fs.mkdir(dir,{recursive:true});const report={version:'goo.activitySnapshotProfile423/1',label,variants:[],errors:[],calls:[]};let runtime,proxy;
 page.on('pageerror',error=>report.errors.push(error.message));page.on('request',r=>{const p=new URL(r.url()).pathname;if(r.method()==='POST'&&telemetry.has(p))report.calls.push({path:p})});
 const counts=()=>({source:report.calls.length,sql:proxy.calls.length}),idle=()=>expect.poll(()=>proxy.calls.every(c=>c.closed),{timeout:60000}).toBe(true);
 await P.install(page);
 try{
  proxy=await startSourceFaultProxy();runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,dir,{clickHouseURL:proxy.url});
  for(const extraAttributes of [0,48]){
   const f=F.fixture('logs',100),payload=F.payload(f);for(const resource of payload.resourceLogs)for(const scope of resource.scopeLogs)for(const record of scope.logRecords)for(let i=0;i<extraAttributes;i++)record.attributes.push({key:'profile.extra.'+String(i).padStart(2,'0'),value:{stringValue:'attribute-'+i+'-'+('x'.repeat(32))}});
   expect((await request.post('http://127.0.0.1:4318/v1/logs',{data:payload})).ok()).toBe(true);
   const raw=()=>source(request,"SELECT ServiceName,ResourceAttributes,LogAttributes AS Attributes,toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns FROM otel_logs WHERE LogAttributes['facet.bank']={bank:String} FORMAT JSONEachRow",{param_bank:f.token,max_result_rows:1000});
   await expect.poll(async()=>(await raw()).filter(r=>r.ResourceAttributes['deployment.environment.name']===f.environment).length,{timeout:45000}).toBe(100);
   const view=await F.saveView(request,runtime.base,f);await page.goto('about:blank');const v=await F.started(page,runtime.base,view,{older:false});await idle();
   const held=await v.tail().locator('goo-record-summary').evaluateAll(es=>es.map(e=>e.record));expect(held).toHaveLength(100);expect(Object.keys(held[0].Attributes).filter(k=>k.startsWith('profile.extra.'))).toHaveLength(extraAttributes);
   const result={extraAttributes,retainedRows:held.length,retainedRecordJSONBytes:Buffer.byteLength(JSON.stringify(held)),measurements:[],reads:[]};report.variants.push(result);
   const local=async(name,run)=>{await idle();const before=counts(),metric=await run();await idle();expect(counts(),name+' must not read telemetry').toEqual(before);result.measurements.push({...metric,before,after:counts()});return metric;};
   const inspectionIndexes=await v.tail().locator('tbody>tr').evaluateAll(rows=>rows.flatMap((row,index)=>row.querySelector('goo-record-summary')?.record?.Attributes?.cohort==='common'?[index]:[]).slice(0,2));expect(inspectionIndexes).toHaveLength(2);result.inspectionIndexes=inspectionIndexes;
   for(let i=0;i<5;i++){const row=v.tail().locator('tbody>tr').nth(inspectionIndexes[i%2]);await local(i===0?'inspect-first':i===1?'inspect-second':'inspect-repeat-'+i,()=>P.click(page,row.getByRole('button',{name:'Inspect',exact:true}),i===0?'inspect-first':i===1?'inspect-second':'inspect-repeat-'+i));await expect(v.inspector()).toBeVisible();}
   const facet=F.facet(v,'attribute');await revealControl(facet);await expect(facet).toBeEnabled();await local('facet-entry',()=>P.click(page,facet,'facet-entry'));await expect(F.load(v,'attribute')).toBeFocused();
   await idle();const before=counts();await page.evaluate(()=>window.__snapshotProfile423.begin('counted-read'));const pending=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/explore/counted-values',{timeout:60000});await F.load(v,'attribute').click();const response=await pending;expect(response.status()).toBe(200);const data=await response.json();expect(data.count).toBe('100');await expect(checkbox(v,null)).toBeVisible();await idle();const metric=await page.evaluate(()=>window.__snapshotProfile423.end());expect(counts()).toEqual({source:before.source+1,sql:before.sql+2});result.reads.push({question:response.request().postDataJSON(),count:data.count,...metric,before,after:counts()});
   await local('select-missing',()=>P.input(page,checkbox(v,null),'select-missing',true));await local('select-empty',()=>P.input(page,checkbox(v,''),'select-empty',true));
   const search=v.counted().getByRole('searchbox',{name:'Search counted values',exact:true});await local('search-loaded',()=>P.input(page,search,'search-loaded','common'));await expect(v.counted()).toContainText('2 selected · 2 hidden by search');
   await local('clear-search',()=>P.input(page,search,'clear-search',''));await local('clear-selection',()=>P.click(page,v.counted().getByRole('button',{name:'Clear selected values',exact:true}),'clear-selection'));
   await local('reselect-missing',()=>P.input(page,checkbox(v,null),'reselect-missing',true));await local('reselect-empty',()=>P.input(page,checkbox(v,''),'reselect-empty',true));
   await local('stage-union',()=>P.click(page,v.counted().getByRole('button',{name:'Stage selected values',exact:true}),'stage-union'));await expect(v.activity()).toContainText('Unapplied scope');expect(await v.tail().locator('goo-record-summary').evaluateAll(es=>es.map(e=>e.record))).toEqual(held);
   result.inspectionMedianMs=result.measurements.filter(m=>m.name.startsWith('inspect')).map(m=>m.durationMs).sort((a,b)=>a-b)[2];
   if(label==='candidate'||process.env.GOO_SNAPSHOT_PROFILE_EXPECT_NO_LOCAL_ROWS==='1')for(const m of result.measurements){if(m.name.startsWith('inspect'))expect(m.retainedRowsCopied,m.name).toBeLessThanOrEqual(200);else expect(m.retainedRowsCopied,m.name).toBe(0);}
  }
  expect(report.errors).toEqual([]);report.status='passed';
 }finally{report.sql=proxy?.calls;await fs.writeFile(path.join(dir,'report.json'),JSON.stringify(report,null,2));await runtime?.stop();await proxy?.stop()}
});
