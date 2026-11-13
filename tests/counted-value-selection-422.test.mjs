import {test} from 'node:test';
import assert from 'node:assert/strict';
import {predicate,valueReason} from '../ui/counted-value-selection.mjs';
const field={kind:'attribute',key:'country'},retained={all:[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'resource',key:'country'},op:'eq',value:'deployment'}]};
function matches(p,row){if(p.all)return p.all.every(x=>matches(x,row));if(p.any)return p.any.some(x=>matches(x,row));const value=row[p.field.kind]?.[p.field.key];return p.op==='missing'?value===undefined:value===p.value;}

test('422 one field uses OR while preserving every retained constraint and input',()=>{
 const before=structuredClone({retained,field}),values=['US','DE'],combined=predicate(retained,field,values);
 assert.deepEqual(combined,{all:[...retained.all,{any:[{field,op:'eq',value:'US'},{field,op:'eq',value:'DE'}]}]});
 for(const country of ['US','DE','FR'])for(const service of ['api','other'])for(const deployment of ['deployment','other']){
  assert.equal(matches(combined,{attribute:{country},column:{service},resource:{country:deployment}}),service==='api'&&deployment==='deployment'&&['US','DE'].includes(country));
 }
 assert.deepEqual({retained,field},before);assert.deepEqual(values,['US','DE']);combined.all[0].value='changed';assert.deepEqual({retained,field},before);
});

test('422 missing, present empty and literal lookalikes remain distinct',()=>{
 const values=[null,'','Missing key','Present empty string'],combined=predicate(undefined,field,values);
 assert.deepEqual(combined.any.map(p=>[p.op,p.value]),[['missing',undefined],['eq',''],['eq','Missing key'],['eq','Present empty string']]);
 for(const value of values){const row={attribute:value===null?{}:{country:value}};assert(matches(combined,row));}
 assert(!matches(combined,{attribute:{country:'US'}}));assert.deepEqual(predicate(undefined,field,['']),{field,op:'eq',value:''});
 assert.deepEqual(predicate({field,op:'eq',value:'US'},field,['US','DE']).all[0],{field,op:'eq',value:'US'});
});

test('422 complete combined node/depth bounds refuse atomically rather than truncate',()=>{
 const leaf={field:{kind:'resource',key:'x'},op:'eq',value:'x'},nearLimit={all:Array.from({length:29},()=>structuredClone(leaf))},before=structuredClone(nearLimit);
 assert.doesNotThrow(()=>predicate(nearLimit,field,['US']));assert.throws(()=>predicate(nearLimit,field,['US','DE']),/32/);assert.deepEqual(nearLimit,before);
 const deepest={any:[{any:[{any:[{any:[leaf]}]}]}]};assert.throws(()=>predicate(deepest,field,['US']),/nested/);
 assert.throws(()=>predicate(undefined,field,Array.from({length:32},(_,i)=>String(i))),/32/);
});

test('422 unsupported, duplicate, oversized and malformed choices never form a predicate',()=>{
 for(const values of [[],['US','US'],[null,null],[undefined],[false],[1],['\ud800'],['a'.repeat(1025)],['\0']])assert.throws(()=>predicate(retained,field,values));
 const sparse=[];sparse.length=1;assert.throws(()=>predicate(retained,field,sparse));
 let reads=0;const getter=[];Object.defineProperty(getter,0,{get(){reads++;return 'US'},enumerable:true});assert.throws(()=>predicate(retained,field,getter));assert.equal(reads,0);
 assert.equal(valueReason(field,'é'.repeat(512)),'');assert(valueReason(field,'é'.repeat(513)));assert.equal(valueReason(field,null),'');
 for(const [signal,key]of [['spans','duration_ns'],['spans','timestamp_ns'],['logs','severity_number']])assert(valueReason({kind:'column',key},'1',signal));
 assert(valueReason({kind:'column',key:'not_known'},'value'));
});
