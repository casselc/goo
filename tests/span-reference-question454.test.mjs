import {test} from 'node:test';
import assert from 'node:assert/strict';
import {traceQuestion,incomingQuestion} from '../ui/span-reference-question.mjs';
import {linkAt,matches} from '../ui/span-link-membership.mjs';
const trace='a'.repeat(32),span='b'.repeat(16),origin={signal:'spans',environment:'prod',fromMs:1000,toMs:2000,where:{field:{kind:'column',key:'service'},op:'eq',value:'worker'},linksTo:{traceId:'c'.repeat(32),spanId:'d'.repeat(16)},groupBy:[{kind:'column',key:'service'}],limit:1,offset:100};
test('deliberate trace and reference expansions retain only exact environment/window and declared target',()=>{
 const before=structuredClone(origin),whole=traceQuestion(origin,trace),target=traceQuestion(origin,trace,span),logs=traceQuestion(origin,trace,span,'logs',100),inverse=incomingQuestion(origin,{traceId:trace,spanId:span});
 for(const q of [whole,target,logs,inverse]){assert.equal(q.environment,'prod');assert.equal(q.fromMs,1000);assert.equal(q.toMs,2000);assert.equal(q.groupBy,undefined);assert.equal(q.reportedEdge,undefined)}
 assert.equal(whole.where.all.length,1);assert.equal(whole.linksTo,undefined);assert.equal(target.where.all.length,2);assert.equal(logs.signal,'logs');assert.equal(logs.offset,100);assert.equal(inverse.where,undefined);assert.deepEqual(inverse.linksTo,{traceId:trace,spanId:span});inverse.linksTo.spanId='changed';assert.deepEqual(origin,before);
});
test('invalid optional IDs and fixed-scope aliases refuse instead of broadening',()=>{
 for(const invalid of [null,false,0,'0'.repeat(16),'B'.repeat(16)])assert.throws(()=>traceQuestion(origin,trace,invalid));
 for(const patch of [{toMs:1000},{fromMs:null},{environmentSelector:null},{offset:1001}]){
  if(patch.offset)assert.throws(()=>traceQuestion(origin,trace,'','spans',patch.offset));else assert.throws(()=>traceQuestion({...origin,...patch},trace));
 }
});
test('outgoing reference admission and inverse membership share array/attribute rules without rejecting unrelated entries',()=>{
 const row={linked_traces:['invalid',trace,trace],linked_spans:['invalid',span,span],linked_attributes:[null,{},null]},out=linkAt(row,1);assert.deepEqual(out.target,{traceId:trace,spanId:span});assert.equal(matches(row,out.target),true);out.attributes.new='detached';assert.deepEqual(row.linked_attributes[1],{});assert.throws(()=>linkAt(row,0));assert.throws(()=>linkAt(row,2));assert.throws(()=>linkAt({...row,linked_spans:[span]},0));assert.equal(matches({...row,linked_spans:[span]},out.target),false);
});
