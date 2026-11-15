import {test} from 'node:test';
import assert from 'node:assert/strict';
import {present} from '../ui/condition-summary.mjs';
import {identity} from '../ui/question-session.mjs';
test('single and lane summaries share exact human units and unchanged raw predicates',()=>{
 const where={all:[{field:{kind:'column',key:'duration_ns'},op:'gte',value:'100000001'},{any:[{field:{kind:'attribute',key:'tenant'},op:'eq',value:''},{field:{kind:'attribute',key:'tenant'},op:'missing'}]}]},before=structuredClone(where),key=identity({where}),single={},lane={};
 present(single,where);present(lane,where,{empty:'All in shared scope'});
 assert.deepEqual(single,lane);assert.match(single.textContent,/100\.000001 ms/);assert.equal(single.title,'Exact predicate · '+JSON.stringify(where));assert.deepEqual(where,before);assert.equal(identity({where}),key);
 present(single,undefined);assert.equal(single.textContent,'Applied · All records in scope');assert.equal(single.title,'All records in scope');
 present(lane,where,{prefix:''});assert(!lane.textContent.startsWith('Applied'));assert.match(lane.title,/100000001/);
});
