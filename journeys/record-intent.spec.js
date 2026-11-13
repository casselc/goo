'use strict';
const {test,expect}=require('@playwright/test');
const {randomBytes}=require('node:crypto');
const fs=require('node:fs/promises'),path=require('node:path');
const {source,startBombadil}=require('./support/cross-app');
const {identity}=require('../ui/question-session.mjs');
const base=process.env.GOO_QUERY_URL||'http://127.0.0.1:8088';
const settle=page=>page.evaluate(()=>new Promise(done=>requestAnimationFrame(()=>requestAnimationFrame(done))));
const attrs=value=>Object.entries(value).map(([key,stringValue])=>({key,value:{stringValue}}));

// This journey exercises the served modules and the actual route owner. The
// independent Go/JS generated differential lives in verification/recordintent.
test('Record intent survives links and history without weaker staging or implicit reads',async({page,request})=>{
 test.setTimeout(180000);page.setDefaultTimeout(15000);
 const dir=path.resolve('artifacts/record-intent417'),shots=path.resolve('docs/journeys/screenshots');await fs.mkdir(dir,{recursive:true});await fs.mkdir(shots,{recursive:true});
 const token=randomBytes(5).toString('hex'),environment='record-intent-'+token,unicodeKey='雪'.repeat(42)+'ab',unicodeValue='🦉'.repeat(256),now=Date.now(),fromMs=now-10000,toMs=now+1000;
 const selector={version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:environment}};
 const bank={field:{kind:'attribute',key:'intent417.bank'},op:'eq',value:token};
 const q={signal:'logs',environmentSelector:selector,fromMs,toMs,limit:100,where:{all:[bank,{field:{kind:'attribute',key:unicodeKey},op:'eq',value:unicodeValue}]}};
 const report={version:'goo.recordIntent417/1',question:q,token,calls:[],facts:[],errors:[]};
 page.on('pageerror',error=>report.errors.push(error.message));
 page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.startsWith('/api/')){let body;try{body=r.postDataJSON()}catch{}report.calls.push({path:new URL(r.url()).pathname,body});}});
 const local=async(name,action)=>{const before=report.calls.length;await action();await settle(page);expect(report.calls.length,name+' must remain source-free').toBe(before);report.facts.push({name,before,after:report.calls.length,phase:await page.locator('#exploration').getAttribute('data-question-phase'),held:await rows().count()>0});};
 const rows=()=>page.locator('#exploreRows .record-row');
 try{
  await page.goto(base+'/#questions');await expect(page.locator('#exploreRun')).toBeVisible();
  const records=Array.from({length:140},(_,i)=>({timeUnixNano:String(BigInt(now-5000+i)*1000000n),severityText:'INFO',severityNumber:9,body:{stringValue:'ordinary work '+i},attributes:attrs({'intent417.bank':token,[unicodeKey]:unicodeValue,'work.latency':'42','work.id':String(i)})}));
  const ingested=await request.post('http://127.0.0.1:4318/v1/logs',{data:{resourceLogs:[{resource:{attributes:attrs({'service.name':'worker-'+token,'deployment.environment.name':environment})},scopeLogs:[{scope:{name:'ordinary.record-intent'},logRecords:records}]}]}});expect(ingested.ok()).toBe(true);
  const raw=()=>source(request,"SELECT LogAttributes['work.id'] AS id,toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns FROM otel_logs WHERE LogAttributes['intent417.bank']={bank:String} FORMAT JSONEachRow",{param_bank:token,max_result_rows:200});
  await expect.poll(async()=>(await raw()).length,{timeout:45000}).toBe(140);report.raw=await raw();
  await local('shared-admission-and-history',async()=>{
   report.protocol=await page.evaluate(async({q,token,bank})=>{
    const P=await import('/question-protocol.mjs'),H=await import('/scope-history.mjs');
    const clone=structuredClone,leaf=()=>clone(bank),nested=(depth,p)=>depth?{all:[nested(depth-1,p)]}:p;
    const valid=[
     {name:'UTF-8 field/value exact boundaries',query:clone(q)},
     {name:'32 predicate nodes',query:{...clone(q),where:{all:Array.from({length:31},leaf)}}},
     {name:'depth four',query:{...clone(q),where:nested(4,leaf())}},
     {name:'typed numeric threshold',query:{...clone(q),where:{all:[leaf(),{numeric:{field:{kind:'attribute',key:'work.latency'},type:'uint64',unit:'ns',parse:'canonical-decimal'},op:'gte',value:'42'}]}}}
    ];
    const storagePath='/record-intent417-'+token,clock=Date.now(),journal=new H.Journal({storage:()=>localStorage,path:storagePath,now:()=>clock});
    const values=[];let persisted;
    try{
     for(const {name,query}of valid){const before=JSON.stringify(query);P.validateRecordQuestion(query);const link=P.questionLink(query),read=P.readQuestionLink(link);if(JSON.stringify(read.question)!==before||JSON.stringify(query)!==before)throw Error(name+' mutated or changed intent');journal.remember({kind:'question',question:query,view:'Records'});journal.pin({kind:'question',question:query,view:'Records'});values.push({name,query,link});}
     journal.persistence(true);const restored=new H.Journal({storage:()=>localStorage,path:storagePath,now:()=>clock});if(restored.error)throw Error(restored.error);if(JSON.stringify(restored.list())!==JSON.stringify(journal.list()))throw Error('Persisted question configuration changed');persisted=restored.list();
     const invalid=[
      ['unknown field member',{...clone(q),where:{field:{kind:'attribute',key:'x',source:'untrusted'},op:'eq',value:'x'}}],
      ['unknown predicate member',{...clone(q),where:{...leaf(),resultToken:'untrusted'}}],
      ['mixed group and leaf',{...clone(q),where:{all:[leaf()],field:{kind:'attribute',key:'x'},op:'eq',value:'x'}}],
      ['33 predicate nodes',{...clone(q),where:{all:Array.from({length:32},leaf)}}],
      ['depth five',{...clone(q),where:nested(5,leaf())}],
      ['129 byte field',{...clone(q),where:{field:{kind:'attribute',key:'雪'.repeat(43)},op:'exists'}}],
      ['1028 byte value',{...clone(q),where:{field:{kind:'attribute',key:'x'},op:'eq',value:'🦉'.repeat(257)}}],
      ['ill-formed Unicode',{...clone(q),where:{field:{kind:'attribute',key:'x'},op:'eq',value:'\ud800'}}],
      ['page limit 101',{...clone(q),limit:101}],
      ['log duration summary',{...clone(q),summary:'span-duration'}],
      ['duplicate grouping',{...clone(q),groupBy:[{kind:'column',key:'service'},{kind:'column',key:'service'}]}]
     ];
     const rejected=[];
     for(const [name,query]of invalid){const before=JSON.stringify(query),entries=JSON.stringify(journal.list()),step={version:1,kind:'record',view:'Records',scope:{environmentSelector:clone(query.environmentSelector),fromMs:query.fromMs,toMs:query.toMs},record:query},raw='#step='+encodeURIComponent(JSON.stringify(step)),checks=[()=>P.validateRecordQuestion(query),()=>P.questionLink(query),()=>P.readQuestionLink(raw),()=>journal.remember({kind:'question',question:query,view:'Records'})],errors=[];for(const check of checks){let error;try{check()}catch(e){error=e.message}if(!error)throw Error(name+' passed a configuration boundary');errors.push(error);}if(JSON.stringify(query)!==before||JSON.stringify(journal.list())!==entries)throw Error(name+' mutated input or remembered intent');rejected.push({name,errors,raw});}
     return {valid:values,rejected,persisted,hash:values[0].link};
    }finally{localStorage.removeItem(journal.key);}
   },{q,token,bank});
   expect(report.protocol.valid).toHaveLength(4);expect(report.protocol.rejected).toHaveLength(11);expect(report.protocol.persisted).toHaveLength(4);
  });
  await local('stage-exact-question',async()=>{await page.evaluate(hash=>{location.hash=hash},report.protocol.hash);await expect(page.locator('#exploreStatus')).toContainText('Fixed question staged.');await expect(rows()).toHaveCount(0);});
  const pending=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/explore');await page.locator('#exploreRun').click();const response=await pending;expect(response.status()).toBe(200);report.result=await response.json();expect(identity(report.result.question)).toBe(identity(q));expect(report.result.count).toBe('140');expect(report.result.rows).toHaveLength(100);for(const row of report.result.rows)expect(report.raw.some(r=>r.id===row.Attributes['work.id'])).toBe(true);await expect(rows()).toHaveCount(100);report.facts.push({name:'explicit-run',after:report.calls.length,phase:await page.locator('#exploration').getAttribute('data-question-phase'),held:await rows().count()>0});
  const held=await rows().evaluateAll(nodes=>nodes.map(n=>n.textContent));expect(held).toHaveLength(100);
  await local('invalid-stage-keeps-held-results',async()=>{
   for(const invalid of report.protocol.rejected){await page.evaluate(hash=>{location.hash=hash},invalid.raw);await expect(page.locator('#exploreStatus')).toContainText('Cannot restore question');await expect(rows()).toHaveCount(100);expect(await rows().evaluateAll(nodes=>nodes.map(n=>n.textContent))).toEqual(held);}
  });
  await page.setViewportSize({width:390,height:844});await page.locator('#exploreStatus').scrollIntoViewIfNeeded();expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:path.join(shots,'record-intent-refusal-390.png')});
  await local('stage-valid-after-refusal',async()=>{await page.evaluate(hash=>{location.hash=hash},report.protocol.hash);await expect(page.locator('#exploreStatus')).toContainText('Fixed question staged.');await expect(rows()).toHaveCount(0);});
  expect(report.calls).toHaveLength(1);expect(report.calls[0]).toEqual({path:'/api/explore',body:q});expect(report.errors).toEqual([]);
  if(process.env.GOO_RECORD_INTENT_BOMBADIL==='1'){
   const url=new URL(base+'/'+report.protocol.hash);url.searchParams.set('expected',JSON.stringify({question:q,qhash:report.protocol.hash,invalidraw:report.protocol.rejected.at(-1).raw,count:'140'}));
   const actor=await startBombadil({dir:path.join(dir,'bombadil-'+token),url:String(url),specification:'verification/bombadil/record-intent417.ts',seconds:35,width:390,height:844});
   try{
    const campaign=await actor.finish({snapshot:{name:'state',version:'goo.recordIntent417Actor/1'}}),complete=campaign.states.find(s=>s.phase==='complete'),refused=campaign.states.find(s=>s.phase==='refused'),observed=refused?.observed;
    expect(complete).toBeTruthy();expect(refused).toBeTruthy();expect(complete.forbidden).toEqual([]);expect(complete.posts).toBe(1);expect(complete.rows).toBe(0);expect(refused.held).toBe(true);expect(refused.rows).toBe(100);expect(refused.ownerPhase).toBe('applied');expect(observed.count).toBe('140');expect(observed.rowIds).toHaveLength(100);expect(new Set(observed.rowIds).size).toBe(100);expect(observed.rowIds.every(id=>report.raw.some(r=>r.id===id))).toBe(true);expect(identity(observed.request)).toBe(identity(q));expect(identity(observed.question)).toBe(identity(q));
    report.bombadil={snapshots:campaign.states.length,complete,refused,observed,run:campaign.run};
   }finally{await actor.stop();}
  }
  report.status='passed';
 }finally{await fs.writeFile(path.join(dir,'report.json'),JSON.stringify(report,null,2));}
});
