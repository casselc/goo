import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Versions from '../ui/service-version-comparison.mjs';
import * as Composition from '../ui/view-composition.mjs';
const scope={environment:'prod',namespace:{mode:'value',value:''},fromMs:0,toMs:60000};
const q={signal:'spans',environment:'prod',fromMs:0,toMs:60000,limit:100,offset:0,summary:'span-duration',spanStatus:true,bucketMs:10000,groupBy:[{kind:'column',key:'name'}],where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''},{field:{kind:'column',key:'name'},op:'eq',value:'GET /items'}]}};
const context={service:'api',scope};
const data=question=>({version:'goo/record-aggregates/v1',question:structuredClone(question),capturedAt:'2026-11-12T00:00:00Z',basis:'a'.repeat(64),count:'10',state:'observed',groups:[{g0:null,count:'1'},{g0:'',count:'2'},{g0:'a',count:'3'},{g0:'b',count:'4'}],warnings:[],reads:{count:{readAt:'2026-11-12T00:00:00Z',basis:'b'.repeat(64)},groups:{readAt:'2026-11-12T00:00:00Z',basis:'c'.repeat(64)}}});

test('415 projection preserves qualified parent and same-window versions do not require a previous window',()=>{
 const before=structuredClone(q),query=Versions.discoveryQuestion({query:q},{question:q,state:'observed'},context);
 assert.deepEqual(query.where,q.where);assert.deepEqual(query.groupBy,[{kind:'resource',key:'service.version'}]);assert.equal(query.limit,1);for(const k of ['summary','spanStatus','bucketMs'])assert(!Object.hasOwn(query,k));
 for(const metric of ['count','duration.mean'])for(const [a,b]of [[null,''],['a','b']]){const view=Versions.comparison(q,data(query),a,b,metric),selected=view.queries[0],declaration=Composition.comparison(view.composition),baseline=declaration.baselineQuery;assert.deepEqual(selected.groupBy,q.groupBy);assert.equal(selected.fromMs,0);assert.deepEqual(selected.where.all[0],q.where);assert.deepEqual(baseline.where.all[0],q.where);assert.equal(selected.where.all[1].value,b);assert.equal(baseline.where.all[1].op,a===null?'missing':'eq');assert.equal(selected.summary,metric==='duration.mean'?'span-duration':undefined);assert.equal(declaration.display.value,metric);assert(view.title.includes('Comparison'));assert(view.title.includes('baseline'));assert.equal(selected.spanStatus,undefined);}
 assert.deepEqual(q,before);
});

test('415 refuses wrong namespace, widget mismatch, incomplete/disagreeing discovery and unknown pairs',()=>{
 const query=Versions.discoveryQuestion({query:q},{question:q,state:'observed'},context);
 for(const c of [{...context,service:'other'},{...context,scope:{...scope,namespace:null}},{...context,scope:{...scope,namespace:{mode:'missing'}}},{...context,scope:{...scope,toMs:60001}}])assert.throws(()=>Versions.discoveryQuestion({query:q},{question:q,state:'observed'},c));
 assert.throws(()=>Versions.discoveryQuestion({query:q},{question:{...q,limit:1},state:'observed'},context));
 for(const mutate of [d=>d.state='changed_between_reads',d=>d.groups.pop(),d=>d.groups.push(d.groups[0]),d=>d.count='11',d=>d.question.where={field:{kind:'column',key:'service'},op:'eq',value:'other'},d=>d.rows=[]]){const d=data(query);mutate(d);assert.throws(()=>Versions.admit(query,d));}
 for(const [a,b,m]of [['a','a','count'],['no','b','count'],['a','b','error.rate']])assert.throws(()=>Versions.comparison(q,data(query),a,b,m));
 assert.equal(Versions.valueLabel(null),'Not reported');assert.equal(Versions.valueLabel(''),'Present empty');assert.equal(Versions.valueLabel('Not reported'),'"Not reported"');assert.equal(Versions.choiceReason('normal'),'');assert(Versions.choiceReason('x'.repeat(4097)));
});

test('419 recorded Error choices preserve cohort intent and request explicit status support',()=>{
 const before=structuredClone(q),query=Versions.discoveryQuestion({query:q},{question:q,state:'observed'},context);
 for(const metric of ['status.error','status.error_fraction']){
  const view=Versions.comparison(q,data(query),'a','b',metric),selected=view.queries[0],declaration=Composition.comparison(view.composition),baseline=declaration.baselineQuery;
  assert.equal(selected.spanStatus,true);assert.equal(baseline.spanStatus,true);assert.equal(selected.summary,undefined);
  assert.deepEqual(selected.where.all[0],q.where);assert.deepEqual(baseline.where.all[0],q.where);assert.deepEqual(selected.groupBy,q.groupBy);
  assert.equal(selected.fromMs,q.fromMs);assert.equal(selected.toMs,q.toMs);assert.equal(selected.environment,q.environment);
  assert.equal(declaration.display.value,metric);assert.equal(declaration.display.view,'summary');
  assert.throws(()=>Composition.setComparison(view.composition,{...declaration,display:{...declaration.display,view:'groups'}},selected));
  assert.throws(()=>Composition.setComparison(view.composition,{...declaration,baselineQuery:{...baseline,spanStatus:false}},selected));
  assert.throws(()=>Composition.setComparison(view.composition,{...declaration,kind:{name:'captured-count-comparison',version:1},baselineCapture:'a'.repeat(64),baselineQuery:undefined},selected));
 }
 assert.deepEqual(q,before);
});
