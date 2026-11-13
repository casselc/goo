import assert from 'node:assert/strict';
import fs from 'node:fs';
import {plan} from '../../ui/counted-selection-comparison.mjs';
import * as Composition from '../../ui/view-composition.mjs';
const samples=JSON.parse(fs.readFileSync(0,'utf8')),clone=structuredClone,key=f=>f.kind+':'+f.key;
let reads=0;globalThis.fetch=()=>{reads++;throw Error('planning acquired source data')};
// This oracle uses finite records and explicit absent/present values; it does
// not reuse the production predicate/complement helper.
function matches(p,row){
 if(!p)return true;if(p.all)return p.all.every(c=>matches(c,row));if(p.any)return p.any.some(c=>matches(c,row));
 const v=row.get(key(p.field));
 if(p.op==='missing')return v===undefined;if(p.op==='exists')return v!==undefined;
 if(p.op==='eq')return v!==undefined&&v===(p.value??'');
 if(p.op==='ne')return v!==undefined&&v!==(p.value??'');
 throw Error('unsupported independent oracle operator '+p.op);
}
for(const [i,sample]of samples.entries()){
 const field={kind:sample.Kind,key:'cohort'},other={kind:sample.Kind==='attribute'?'resource':'attribute',key:'cohort'},tenant={kind:'attribute',key:'tenant'};
 const choices=[null,'',sample.Value,'Missing key'],values=[...new Set(choices.filter((_,index)=>(i+index)%3!==0))];
 const where={all:[{field:tenant,op:'eq',value:sample.Tenant},{any:[{field:other,op:'eq',value:'blue'},{field:other,op:'missing'}]}]};
 const parent={signal:sample.Signal,environmentSelector:{version:'goo/environment-selector/v1',policy:sample.Policy,scope:{mode:'named',value:sample.Environment}},fromMs:sample.From,toMs:sample.From+60000,limit:1,offset:0,groupBy:[clone(field)],where};
 const before=clone(parent),selectedValues=clone(values),view=plan(parent,field,values),selected=view.queries[0],rest=Composition.comparison(view.composition).baselineQuery;
 Composition.view(view);assert.equal(Composition.comparison(view.composition).display.value,'count');
 for(const q of [selected,rest]){
  assert.equal(q.signal,parent.signal);assert.deepEqual(q.environmentSelector,parent.environmentSelector);assert.equal(q.fromMs,parent.fromMs);assert.equal(q.toMs,parent.toMs);assert.deepEqual(q.groupBy,parent.groupBy);
  assert.equal(q.offset,0);assert.equal(q.limit,100);
 }
 for(const v of [undefined,'','Missing key',sample.Value,'other','雪'])for(const peer of [undefined,'blue','green'])for(const t of [sample.Tenant,'foreign']){
  const row=new Map([[key(tenant),t]]);if(v!==undefined)row.set(key(field),v);if(peer!==undefined)row.set(key(other),peer);
  const inside=matches(parent.where,row),wanted=values.some(value=>value===null?v===undefined:v===value),a=matches(selected.where,row),b=matches(rest.where,row);
  assert.equal(a,inside&&wanted,'selection must use qualified exact values');assert.equal(b,inside&&!wanted,'rest is relative to the unchanged parent');assert(!(a&&b),'selected/rest overlap');assert.equal(a||b,inside,'selected/rest must cover parent semantics');
 }
 assert.deepEqual(parent,before);assert.deepEqual(values,selectedValues);assert.deepEqual(plan(parent,field,values),view,'planning is deterministic');
 // The namespace, ordinary profile and whole AST budget are admission rules.
 assert.throws(()=>plan(parent,other,values));assert.throws(()=>plan(parent,field,[]));assert.throws(()=>plan(parent,field,[...values,values[0]]));
 for(const mutate of [q=>q.offset=1,q=>q.summary='span-duration',q=>q.spanStatus=true,q=>q.groupBy.push(other),q=>q.bucketMs=1000]){const unsupported=clone(parent);mutate(unsupported);assert.throws(()=>plan(unsupported,field,values));}
 const crowded=clone(parent);crowded.where={all:Array.from({length:30},(_,index)=>({field:{kind:'attribute',key:'tenant'+index},op:'eq',value:'x'}))};assert.throws(()=>plan(crowded,field,[null,'',sample.Value]));
 // A new parent window/policy produces new fixed operands, never a cached plan.
 const changed=clone(parent);changed.fromMs+=60000;changed.toMs+=60000;changed.environmentSelector.policy=changed.environmentSelector.policy==='effective'?'current-only':'effective';const next=plan(changed,field,values);assert.equal(next.queries[0].fromMs,changed.fromMs);assert.deepEqual(Composition.comparison(next.composition).baselineQuery.environmentSelector,changed.environmentSelector);assert.deepEqual(view.queries[0].environmentSelector,parent.environmentSelector);
 assert.equal(reads,0);
}
process.stdout.write(JSON.stringify(samples.length));
