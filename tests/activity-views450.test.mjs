import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Views from '../ui/activity-views.mjs';
const condition=key=>({field:{kind:'column',key},op:'eq',value:'error'});
const draft=()=>({version:1,title:'Independent lanes',selector:{signal:'both',environment:'prod',where:condition('service')},lanes:{logs:{columns:['utc','summary'],where:condition('body')},spans:{columns:['utc','summary'],where:condition('name')}}});
test('saved paired lane conditions remain distinct and detached across input/model/presentation',()=>{
 const v=draft(),accepted=Views.input(v),held=Views.model({...accepted,id:'a'.repeat(64)}),shown=Views.presentation(held);
 assert.deepEqual(held.lanes.logs.where,v.lanes.logs.where);assert.deepEqual(held.lanes.spans.where,v.lanes.spans.where);
 shown.lanes.logs.where.value='mutated';accepted.lanes.spans.where.value='other';assert.equal(v.lanes.logs.where.value,'error');assert.equal(held.lanes.spans.where.value,'error');
 delete v.selector.where;assert.doesNotThrow(()=>Views.input(v));
});
test('whole saved definition rejects wrong-family, malformed and combined-budget lane predicates',()=>{
 for(const mutate of [v=>v.lanes.logs.where=condition('name'),v=>v.lanes.spans.where=condition('body'),v=>v.lanes.logs.where=null,v=>v.lanes.logs.where=undefined,v=>v.lanes.logs.where={},v=>{v.selector.where={all:Array.from({length:30},()=>condition('service'))}}]){const v=draft();mutate(v);assert.throws(()=>Views.input(v))}
});
