import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Service from '../../ui/service-count-comparison.mjs';
const questions=JSON.parse(fs.readFileSync(0,'utf8'));
globalThis.fetch=()=>{throw Error('Count projection must not read telemetry')};
const views=[];
for(const q of questions){
 const before=structuredClone(q),count=structuredClone(q);delete count.summary;delete count.spanStatus;
 assert.deepEqual(Service.project(q),count);
 const widget={query:q},data={question:q,state:q.signal==='logs'?'no_data':'observed'};
 const view=Service.create(widget,data);
 assert.deepEqual(view.queries,[count]);
 const pair=view.composition.children.find(child=>child.comparison).comparison;
 assert.deepEqual(pair.baselineQuery,{...count,fromMs:q.fromMs-(q.toMs-q.fromMs),toMs:q.fromMs});
 assert.equal(pair.display.value,'count');
 assert.deepEqual(q,before,'projection must not mutate accepted widget question');
 assert.deepEqual(data,{question:q,state:q.signal==='logs'?'no_data':'observed'});
 for(const mutate of [x=>x.offset=1,x=>x.limit=101,x=>x.groupBy=[],x=>x.groupBy.push({kind:'column',key:'service'}),x=>x.numeric={field:{kind:'attribute',key:'payload.size'},type:'uint64',unit:'By',parse:'canonical-decimal'},x=>{x.fromMs=1;x.toMs=10001}]){
  const bad=structuredClone(q);mutate(bad);assert.throws(()=>Service.project(bad));
 }
 assert.throws(()=>Service.create(widget,{question:{...q,limit:q.limit===100?99:100},state:'observed'}),'foreign widget read refused');
 assert.throws(()=>Service.create(widget,{question:q,state:'changed_between_reads'}),'incoherent source refused');
 assert.throws(()=>Service.create({...widget,measurement:{}},data),'native widgets refused');
 views.push(view);
}
process.stdout.write(JSON.stringify(views));
