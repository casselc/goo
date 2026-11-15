import {test} from 'node:test';
import assert from 'node:assert/strict';
import {plan} from '../ui/span-link-intent.mjs';
import * as Environment from '../ui/record-environment.mjs';
const source={TraceId:'a'.repeat(32),SpanId:'b'.repeat(16)},target={TraceId:'c'.repeat(32),SpanId:'d'.repeat(16)};
const fixture=()=>({origin:{signal:'spans',environment:'prod',fromMs:1700000000000,toMs:1700000060000,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'producer'},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:'source'}]},groupBy:[{kind:'column',key:'service'}],limit:20,offset:100},rows:[{...source,linked_traces:[target.TraceId],linked_spans:[target.SpanId],linked_attributes:[{'messaging.message.id':'order-42'}]}],rowIndex:0,linkIndex:0});
test('target plan preserves exact environment/window but intentionally drops population filters',()=>{
 for(const scope of [{environment:'prod'},{environmentSelector:Environment.current('prod')},{environmentSelector:Environment.effective({mode:'named',value:'prod'})},{environmentSelector:Environment.effective({mode:'empty'})},{environmentSelector:Environment.effective({mode:'unreported'})}]){
  const args=fixture();delete args.origin.environment;Object.assign(args.origin,scope);const out=plan(args);
  assert.deepEqual(Environment.questionFields(out.question),Environment.questionFields(args.origin));assert.equal(out.question.fromMs,args.origin.fromMs);assert.equal(out.question.toMs,args.origin.toMs);
  assert.deepEqual(out.question.where,{all:[{field:{kind:'column',key:'trace_id'},op:'eq',value:target.TraceId},{field:{kind:'column',key:'span_id'},op:'eq',value:target.SpanId}]});assert.equal(out.question.limit,100);assert.equal(out.question.offset,0);assert.equal(out.question.groupBy,undefined);assert.equal(out.local.state,'absent');
 }
});
test('local exact tuple selects only unique target and reports duplicates as ambiguous',()=>{
 const args=fixture();args.rows.push({...target},{...target,TraceId:'e'.repeat(32)});assert.deepEqual(plan(args).local,{state:'unique',rowIndexes:[1]});args.rows.push({...target});assert.deepEqual(plan(args).local,{state:'ambiguous',rowIndexes:[1,3]});
 args.rows[0].linked_traces.push(source.TraceId);args.rows[0].linked_spans.push(source.SpanId);args.rows[0].linked_attributes.push({});args.linkIndex=1;assert.deepEqual(plan(args).local,{state:'unique',rowIndexes:[0]});
});
test('link arrays are correlated by position, not independent membership or invented span fallback',()=>{
 const args=fixture();args.rows[0].linked_traces.push('e'.repeat(32));args.rows[0].linked_spans.push('f'.repeat(16));args.rows[0].linked_attributes.push({});args.linkIndex=1;assert.deepEqual(plan(args).target,{traceId:'e'.repeat(32),spanId:'f'.repeat(16)});
 for(const change of [a=>delete a.rows[0].linked_attributes,a=>a.rows[0].linked_spans.pop(),a=>a.rows[0].linked_traces.pop(),a=>a.rows[0].linked_attributes[1]=null,a=>a.rows[0].linked_attributes[1]={count:1},a=>delete a.rows[0].linked_spans[1]]){const next=structuredClone(args);change(next);assert.throws(()=>plan(next))}
});
test('invalid IDs, indexes, scope and origin refuse before a question can escape',()=>{
 for(const value of ['',undefined,'0'.repeat(32),'A'.repeat(32),'a'.repeat(31),'a'.repeat(33)]){const args=fixture();args.rows[0].linked_traces[0]=value;assert.throws(()=>plan(args))}
 for(const value of ['',undefined,'0'.repeat(16),'D'.repeat(16),'d'.repeat(15)]){const args=fixture();args.rows[0].linked_spans[0]=value;assert.throws(()=>plan(args))}
 for(const change of [a=>a.rowIndex=-1,a=>a.rowIndex=0.5,a=>a.linkIndex=1,a=>a.linkIndex=-1,a=>a.origin.signal='logs',a=>a.origin.toMs=a.origin.fromMs,a=>a.rows[0].SpanId='invalid',a=>a.origin.environmentSelector=Environment.effective({mode:'all'})]){const args=fixture();change(args);assert.throws(()=>plan(args))}
});
test('plan is detached and does not fabricate a read or custody clock',()=>{
 const args=fixture(),before=structuredClone(args),out=plan(args);out.origin.question.where.all[0].value='changed';out.question.where.all[0].value='changed';out.attributes['messaging.message.id']='changed';out.local.rowIndexes.push(99);assert.deepEqual(args,before);assert.equal(out.readAt,undefined);assert.equal(out.custody,undefined);assert.equal(out.origin.readAt,undefined);
});
