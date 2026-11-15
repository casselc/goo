import {test} from 'node:test';
import assert from 'node:assert/strict';
import {create,reason} from '../ui/paired-activity-plan.mjs';
import * as Environment from '../ui/record-environment.mjs';

const field=(kind,key,value='x')=>({field:{kind,key},op:'eq',value});
test('pair admission conserves complete selector structure and detaches both families',()=>{
 const where={all:[field('resource','service.namespace'),{any:[field('attribute','example.region',''),field('column','service')]}]},selector={signal:'spans',environment:'prod',where},before=structuredClone(selector),plan=create(selector);
 assert.deepEqual(plan,{logs:{signal:'logs',environment:'prod',where},spans:{signal:'spans',environment:'prod',where}});
 assert.deepEqual(selector,before);assert.notEqual(plan.logs.where,selector.where);assert.notEqual(plan.logs.where,plan.spans.where);
 plan.logs.where.all[1].any[0].field.key='other';assert.deepEqual(selector,before);assert.deepEqual(plan.spans.where,before.where);
});
test('named empty and unreported environment policies stay exact, never broaden to All',()=>{
 const selectors=[{environment:'prod'},...['named','empty','unreported'].map(mode=>({environment:'',environmentSelector:Environment.effective({mode,...(mode==='named'?{value:'prod'}:{})})}))];
 for(const selector of selectors){const plan=create(selector);for(const signal of ['logs','spans'])assert.deepEqual(plan[signal],{signal,...selector});assert.equal(reason(selector),'')}
 for(const selector of [{environment:''},{environment:'',environmentSelector:Environment.effective({mode:'all'})},{environment:'prod',environmentSelector:Environment.effective({mode:'empty'})},{environmentSelector:{version:'goo/environment-selector/v1',policy:'current-only',scope:{mode:'empty'}}}])assert.throws(()=>create(selector));
});
test('one unsupported leaf refuses whole pair before any effect, including nested and typed distinctions',()=>{
 const oldFetch=globalThis.fetch;let reads=0;globalThis.fetch=()=>{reads++;throw Error('planner attempted source IO')};
 try{
  for(const [signal,key]of [['logs','body'],['spans','name'],['logs','severity_number'],['spans','duration_ns']]){
   const where={all:[field('column','service'),{any:[field('attribute','ok'),field('column',key,key==='severity_number'||key==='duration_ns'?'1':'x')]}]},selector={signal,environment:'prod',where},before=structuredClone(selector);
   assert.match(reason(selector),/Shared filter is unavailable for/);assert.throws(()=>create(selector));assert.deepEqual(selector,before);
  }
  const numeric={numeric:{field:{kind:'attribute',key:'http.request.body.size'},type:'uint64',unit:'By',parse:'canonical-decimal'},op:'gte',value:'9007199254740993'},plan=create({environment:'prod',where:numeric});assert.deepEqual(plan.logs.where,numeric);assert.deepEqual(plan.spans.where,numeric);
  assert.equal(reads,0);
 }finally{globalThis.fetch=oldFetch}
});
test('closed pair input rejects clocks, unknown profiles, malformed shapes and accessors',()=>{
 for(const patch of [{signal:'both'},{signal:null},{fromMs:1},{numeric:{}},{reportedEdge:{}},{where:null},{where:{all:[]}}])assert.notEqual(reason({environment:'prod',...patch}),'');
 let invoked=0;const selector={environment:'prod'};Object.defineProperty(selector,'where',{enumerable:true,get(){invoked++;return undefined}});assert.throws(()=>create(selector));assert.equal(invoked,0);
 const invalid={environment:'prod',where:{field:{kind:'attribute',key:'x'},op:'eq',value:'x'.repeat(1025)}};assert.throws(()=>create(invalid));
});
