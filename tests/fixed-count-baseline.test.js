const {test}=require('node:test'),assert=require('node:assert/strict');
const {readFixedCountBaseline}=require('../ui/population-workspace.mjs'),C=require('../ui/captured-count-comparison.mjs');
const {aggregateFixture}=require('./fixtures/record-aggregate-fixture.cjs');
function fixture(){const {result}=aggregateFixture();result.question.summary='span-duration';delete result.question.spanStatus;delete result.spanStatus;for(const item of [...result.groups,...result.timeSeries])delete item.spanStatus;return result}
test('fixed string baseline acquires aggregates once with exact intent and receipts, without full snapshot authority',async()=>{
 const data=fixture(),q=structuredClone(data.question),before=structuredClone(q),controller=new AbortController(),calls=[];
 const result=await readFixedCountBaseline(q,controller.signal,{post:async(path,body,signal)=>{calls.push({path,body,signal});return {ok:true,data}}});
 assert.equal(calls.length,1);assert.equal(calls[0].path,'/api/record-aggregates');assert.equal(calls[0].signal,controller.signal);assert.deepEqual(calls[0].body,q);assert.notEqual(calls[0].body,q);
 assert.deepEqual(result.reads,data.reads);assert.equal(result.count,'2');assert.deepEqual(C.cohort(result),C.cohort(data));for(const key of ['viewDataset','summaryDataset','rows','fields','recordPage'])assert.equal(result[key],undefined);
 data.question.environment='changed';data.reads.count.basis='changed';assert.deepEqual(q,before);assert.deepEqual(result.question,before);assert.equal(result.reads.count.basis,'a'.repeat(64));
});
test('numeric column grouping retains the existing full acquisition without an error-driven fallback',async()=>{
 for(const [signal,key]of [['spans','duration_ns'],['spans','timestamp_ns'],['logs','severity_number']]){
  const q={signal,environment:'prod',fromMs:1000,toMs:2000,limit:100,offset:0,groupBy:[{kind:'column',key}]},calls=[];
  const result=await readFixedCountBaseline(q,undefined,{post:async(path,body)=>{calls.push(path);return {ok:true,data:{question:body,count:'0',groups:[],rows:[],fields:[],hasMore:false,state:'no_data',basis:'b'.repeat(64),capturedAt:'2026-10-25T00:00:01Z'}}}});
  assert.deepEqual(calls,['/api/explore']);assert.equal(result.version,undefined);assert.equal(result.count,'0');
 }
});
test('failed or malformed aggregate reads fail once, preserving diagnostic and empty states',async()=>{
 for(const patch of [{viewDataset:{}},{question:{...fixture().question,environment:'other'}},{reads:{}},{count:'3'}]){
  const data={...fixture(),...patch},calls=[];await assert.rejects(()=>readFixedCountBaseline(fixture().question,undefined,{post:async path=>{calls.push(path);return {ok:true,data}}}));assert.deepEqual(calls,['/api/record-aggregates']);
 }
 let calls=0;await assert.rejects(()=>readFixedCountBaseline(fixture().question,undefined,{post:async()=>{calls++;return {ok:false,data:{error:'source unavailable'}}}}),/source unavailable/);assert.equal(calls,1);
 const changed={...fixture(),state:'changed_between_reads',groups:[]};assert.equal((await readFixedCountBaseline(changed.question,undefined,{post:async()=>({ok:true,data:changed})})).state,'changed_between_reads');
 const empty={...fixture(),state:'no_data',count:'0',groups:[],duration:{state:'no_data',count:'0',sumNs:'0'}};empty.timeSeries=[{fromMs:1000,toMs:2000,count:'0',state:'no_data',duration:{...empty.duration}}];assert.equal((await readFixedCountBaseline(empty.question,undefined,{post:async()=>({ok:true,data:empty})})).state,'no_data');
});
test('fixed baseline owns its request and refuses aborted or late fulfillment without retry',async()=>{
 const data=fixture(),q=structuredClone(data.question),controller=new AbortController();let finish,body,calls=0;
 const pending=readFixedCountBaseline(q,controller.signal,{post:(path,request,signal)=>{calls++;body=request;assert.equal(signal,controller.signal);return new Promise(resolve=>finish=resolve)}});q.environment='changed';assert.equal(body.environment,'prod');controller.abort();finish({ok:true,data});await assert.rejects(pending,{name:'AbortError'});assert.equal(calls,1);
 await assert.rejects(()=>readFixedCountBaseline(data.question,controller.signal,{post:async()=>{calls++}}),{name:'AbortError'});assert.equal(calls,1);
 await assert.rejects(()=>readFixedCountBaseline({...data.question,spanStatus:true},undefined,{post:async()=>{calls++}}));assert.equal(calls,1);
});
