import {test} from 'node:test';
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
import * as Native from '../ui/native-histogram-declaration.mjs';
const {result}=createRequire(import.meta.url)('./fixtures/native-histogram.cjs');
const id=n=>n.toString(16).padStart(64,'0'),item=n=>({id:id(n),title:'Native '+n+' 雪',kind:n%2?'histogram':'table',unit:'s'}),clone=structuredClone;
test('native metadata page admission keeps candidate loss explicit and the cursor at the last scanned ID',()=>{
 const raw={items:Array.from({length:20},(_,i)=>i+1).filter(n=>n!==2).map(item),partial:true,diagnostics:[{id:id(2),reason:'unavailable-artifact'}],nextCursor:id(20)},before=clone(raw),checked=Native.page(raw,{after:id(0)});assert.deepEqual(checked,raw);checked.items[0].title='caller mutation';assert.deepEqual(raw,before);
 for(const mutate of [p=>p.partial=false,p=>p.diagnostics[0].reason='missing',p=>p.nextCursor=id(2),p=>p.items.push(item(1)),p=>p.items.reverse(),p=>p.items[0].basis='captured',p=>p.total=3,p=>p.items[0].unit='By',p=>p.items[0].kind='line',p=>p.diagnostics.push({id:id(1),reason:'unavailable-artifact'})]){const bad=clone(raw);mutate(bad);assert.throws(()=>Native.page(bad,{after:id(0)}))}
 assert.throws(()=>Native.page(raw,{after:id(1)}));assert.throws(()=>Native.page({items:[item(1)],partial:false,diagnostics:[],nextCursor:id(1)}));assert.throws(()=>Native.page({...raw,items:Array.from({length:21},(_,i)=>item(i+1)),diagnostics:[],partial:false,nextCursor:id(21)}));assert.doesNotThrow(()=>Native.page({items:[],partial:true,diagnostics:[{id:id(1),reason:'unavailable-artifact'}]}));assert.doesNotThrow(()=>Native.page({items:[],partial:false,diagnostics:[]}));
});
test('summary compatibility checks only exact units after validating the complete destination question',()=>{
 const q=result().question,before=clone(q);assert.equal(Native.compatibility(item(1),q),true);assert.equal(Native.compatibility({...item(1),unit:'ms'},q),false);assert.deepEqual(q,before);
 for(const mutate of [q=>q.definition.kind='counter',q=>q.definition.grain='consumer-group',q=>q.definition.histogramBucketMs=1000,q=>q.definition.source.resource=null]){const invalid=clone(q);mutate(invalid);assert.throws(()=>Native.compatibility(item(1),invalid))}
});
