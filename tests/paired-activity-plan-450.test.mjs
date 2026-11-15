import {test} from 'node:test';
import assert from 'node:assert/strict';
import {create,laneSelector,reason} from '../ui/paired-activity-plan.mjs';
const leaf=(kind,key,op='eq',value='')=>({field:{kind,key},op,...(value===undefined?{}:{value})});
const shared={environment:'prod',where:leaf('column','service','eq','worker')};
test('450 signal-specific lanes conserve shared scope and admit severity versus duration independently',()=>{
 const conditions={logs:leaf('column','severity_number','gte','17'),spans:leaf('column','duration_ns','gte','9007199254740993')},before=structuredClone({shared,conditions}),plan=create(shared,conditions);
 for(const signal of ['logs','spans'])assert.deepEqual(plan[signal],{signal,environment:'prod',where:{all:[shared.where,conditions[signal]]}});
 assert.deepEqual(laneSelector(shared,'logs',conditions),plan.logs);
 plan.logs.where.all[0].value='changed';plan.spans.where.all[1].value='changed';
 assert.deepEqual({shared,conditions},before);assert.equal(plan.spans.where.all[0].value,'worker');
 assert.equal(reason(shared,conditions),'');
});
test('450 omitted lane conditions preserve shared predicates and exact missing/empty semantics',()=>{
 for(const conditions of [undefined,{}, {logs:undefined,spans:undefined}])assert.deepEqual(create(shared,conditions),create(shared));
 const conditions={logs:leaf('resource','service.instance.id','missing'),spans:leaf('resource','service.instance.id','eq','')};
 const plan=create({environment:'prod'},conditions);
 assert.deepEqual(plan.logs.where,conditions.logs);assert.deepEqual(plan.spans.where,conditions.spans);
 assert.notEqual(plan.logs.where,conditions.logs);assert.notEqual(plan.logs.where,plan.spans.where);
});
test('450 invalid lane or shared predicates refuse whole pair rather than omit restrictions',()=>{
 for(const conditions of [null,[],{other:undefined},{logs:null},{spans:{}},{logs:{all:[]}},{logs:leaf('column','duration_ns','gte','1')},{spans:leaf('column','severity_number','gte','1')},{logs:leaf('column','severity_number','gte','18446744073709551616')}]){
  assert.throws(()=>create(shared,conditions));assert.notEqual(reason(shared,conditions),'');
 }
 assert.throws(()=>create({...shared,where:leaf('column','severity_number','gte','17')},{logs:leaf('column','body','contains','error')}));
 assert.throws(()=>laneSelector(shared,'both'));
 let accessed=0;const conditions={};Object.defineProperty(conditions,'logs',{enumerable:true,get(){accessed++;return undefined}});assert.throws(()=>create(shared,conditions));assert.equal(accessed,0);
});
test('450 combined predicate budgets apply after composition and preserve accepted boundary',()=>{
 const group=n=>({all:Array.from({length:n},()=>leaf('attribute','region','eq','west'))});
 // shared group + 28 leaves + lane leaf + combined group = 31 nodes.
 assert.doesNotThrow(()=>create({environment:'prod',where:group(28)},{logs:leaf('column','severity_number','gte','1')}));
 assert.doesNotThrow(()=>create({environment:'prod',where:group(29)},{logs:leaf('column','severity_number','gte','1')}));
 assert.throws(()=>create({environment:'prod',where:group(30)},{logs:leaf('column','severity_number','gte','1')}),/32 conditions/);
 const nested=depth=>depth?{all:[nested(depth-1)]}:leaf('attribute','region','eq','west');
 assert.doesNotThrow(()=>create({environment:'prod',where:nested(3)},{spans:leaf('column','duration_ns','gte','1')}));
 assert.throws(()=>create({environment:'prod',where:nested(4)},{spans:leaf('column','duration_ns','gte','1')}),/nested levels/);
});
