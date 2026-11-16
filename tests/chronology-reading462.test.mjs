import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Reading from '../ui/chronology-reading.mjs';
import * as Held from '../ui/held-chronology.mjs';
import * as Records from '../ui/record-presentation.mjs';
import {lanes} from './support/held-chronology458-fixture.mjs';
const key='a'.repeat(64)+':1',selectedKey='logs:'+key;
test('reading return is independent of later anchor, span refinement and selected-row visibility',()=>{
 const start=Reading.page(Reading.find(Reading.initial(),'original Find'),1,200),related=Reading.enter(start,{signal:'logs',key},{selectedKey,scroll:53}),changed=Reading.span(Reading.enter(related,{signal:'spans',key},{selectedKey:'spans:'+key,scroll:0}),true);
 assert.deepEqual(Reading.leave(changed).snapshot,{query:'original Find',offset:100,selectedKey,scroll:53});assert.deepEqual(Reading.leave(changed).state,start);assert.throws(()=>Reading.find(related,'hidden Find'));assert.equal(start.query,'original Find');assert.equal(Object.isFrozen(changed.returnPoint),true);assert.throws(()=>Reading.span(start,true));assert.throws(()=>Reading.find(start,'x'.repeat(257)));assert.throws(()=>Reading.fit(start,201));assert.throws(()=>Reading.page(start,100,200));
});
test('generated reading campaigns preserve the first return point and reset barriers',()=>{
 for(let seed=1;seed<=100;seed++){
  let state=Reading.initial(),saved=null,rng=seed;
  for(let i=0;i<80;i++){
   rng=(Math.imul(rng,1664525)+1013904223)>>>0;const action=rng%6;
   if(action===0){state=Reading.initial();saved=null}
   else if(action===1&&!state.related)state=Reading.find(state,'query '+i);
   else if(action===2)state=Reading.page(state,i%2?1:-1,200);
   else if(action===3){if(!state.related)saved={query:state.query,offset:state.offset,selectedKey:i%2?selectedKey:null,scroll:i};state=Reading.enter(state,{signal:i%2?'logs':'spans',key},{selectedKey:i%2?selectedKey:null,scroll:i})}
   else if(action===4&&state.related)state=Reading.span(state,i%2===0);
   else if(action===5&&state.related){const result=Reading.leave(state);assert.deepEqual(result.snapshot,saved);assert.equal(result.state.query,saved.query);assert.equal(result.state.offset,saved.offset);state=result.state;saved=null}
   assert.equal(Object.isFrozen(state),true);assert.ok([0,100].includes(state.offset));assert.deepEqual(state.returnPoint,saved);
  }
 }
});
test('parent context caches immutable banks without collapsing ambiguous observations or leaking rich payloads',()=>{
 const source=lanes(),trace='a'.repeat(32),span='b'.repeat(16);source.spans.rows[0].record={SpanName:'root',TraceId:trace,SpanId:span,ParentSpanId:'',Attributes:{large:'x'.repeat(10000)}};source.spans.rows.push({...source.spans.rows[0],key:'a'.repeat(64)+':2'});const bank=Held.prepare(source),projection=Held.parentRecords(bank);assert.equal(Held.parentRecords(bank),projection);assert.equal(projection.length,2);assert.equal(Object.isFrozen(projection[0]),true);assert.equal(projection[0].Attributes,undefined);assert.equal(Records.ancestry({TraceId:trace,SpanId:'c'.repeat(16),ParentSpanId:span},projection).state,'ambiguous-parent');
 source.spans.rows[0].record.SpanId='d'.repeat(16);assert.notEqual(Held.parentRecords(Held.prepare(source)),projection);assert.equal(projection[0].SpanId,span);
 for(const parent of [undefined,null,'','0'.repeat(16),'bad',{rich:true}]){const record={TraceId:trace,SpanId:span,...(parent!==undefined?{ParentSpanId:parent}:{})};assert.deepEqual(Records.ancestry(record,[record]),Records.ancestry(Records.contextRecord(record),Records.contextRecords([record])))}
});
