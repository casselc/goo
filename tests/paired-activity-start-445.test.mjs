import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as PairPlan from '../ui/paired-activity-plan.mjs';
import {canonical} from '../ui/tail.mjs';
const source=readFileSync(new URL('../ui/paired-activity.mjs',import.meta.url),'utf8');
const actions=source.slice(source.indexOf(' function adopt('),source.indexOf(' function pause('));
function fixture(){
 const calls={clears:0,stages:[],starts:[],pauses:[]},state={allowed:true,blocked:{logs:false,spans:false}},old={signal:'logs',environment:'old'},lanes=new Map();
 const held=()=>Object.fromEntries([...lanes].map(([signal,lane])=>[signal,structuredClone(lane.held)]));
 const context={PairPlan:{create:(value,conditions)=>PairPlan.create(structuredClone(value),structuredClone(conditions))},Tail:{canonical},structuredClone,epoch:4,disposed:false,selector:structuredClone(old),plan:PairPlan.create(old),lanes,available:()=>state.allowed,clearInspection(){calls.clears++}};
 for(const signal of ['logs','spans']){
  const lane={epoch:3,held:{selector:{signal,environment:'old'},keys:[signal+'-held'],phase:'paused'}};
  lane.tail={ready:()=>state.allowed&&!state.blocked[signal],stage(selector){calls.stages.push(signal);lane.held={selector:structuredClone(selector),keys:[],phase:'paused'};state.onStage?.(signal)},start(selector){calls.starts.push(signal);if(state.refuse?.[signal])return false;lane.held.selector=structuredClone(selector);lane.held.phase='loading';state.onStart?.(signal);return true},pause(){calls.pauses.push(signal);lane.held.phase='paused'}};lanes.set(signal,lane);
 }
 vm.runInNewContext(actions+'\nthis.api={stage,start,ready};',context);
 return {context,state,calls,held};
}

test('paired initial refusal preserves both held banks and applied pair identity',()=>{
 for(const withdraw of [f=>f.state.allowed=false,f=>f.state.blocked.logs=true,f=>f.state.blocked.spans=true]){
  const f=fixture(),before=f.held(),selector=structuredClone(f.context.selector);withdraw(f);
  assert.equal(f.context.api.start({signal:'logs',environment:'new'}),false);
  assert.deepEqual(f.held(),before);assert.deepEqual(f.context.selector,selector);assert.equal(f.context.epoch,4);assert.deepEqual(f.calls,{clears:0,stages:[],starts:[],pauses:[]});
 }
});

test('whole-pair predicate refusal happens before either bank is cleared or source is attempted',()=>{
 const f=fixture(),before=f.held();assert.throws(()=>f.context.api.start({signal:'logs',environment:'new',where:{field:{kind:'column',key:'body'},op:'eq',value:'message'}}),/spans/);
 assert.deepEqual(f.held(),before);assert.equal(f.calls.clears,0);assert.deepEqual(f.calls.starts,[]);
});

test('both local scopes are staged before the first accepted source start',()=>{
 const f=fixture();f.state.onStart=()=>{assert.deepEqual(f.calls.stages,['logs','spans']);for(const lane of Object.values(f.held()))assert.equal(lane.selector.environment,'new')};
 assert.equal(f.context.api.start({signal:'spans',environment:'new'}),true);assert.deepEqual(f.calls.starts,['logs','spans']);assert.equal(f.context.epoch,5);assert.deepEqual(f.calls.pauses,[]);
});

test('synchronous withdrawal between lane starts cancels accepted new work without false success',()=>{
 const f=fixture();f.state.onStart=signal=>{if(signal==='logs')f.state.allowed=false};
 assert.equal(f.context.api.start({signal:'logs',environment:'new'}),false);assert.deepEqual(f.calls.starts,['logs']);assert.deepEqual(f.calls.pauses,['logs']);
 for(const lane of Object.values(f.held())){assert.equal(lane.selector.environment,'new');assert.deepEqual(lane.keys,[]);assert.equal(lane.phase,'paused')}
});

test('withdrawal while staging performs no source work and keeps both new local scopes coherent',()=>{
 const f=fixture();f.state.onStage=signal=>{if(signal==='logs')f.state.blocked.spans=true};
 assert.equal(f.context.api.start({environment:'new'}),false);assert.deepEqual(f.calls.stages,['logs','spans']);assert.deepEqual(f.calls.starts,[]);assert.deepEqual(f.calls.pauses,[]);
 for(const lane of Object.values(f.held()))assert.equal(lane.selector.environment,'new');
});

test('an old start cannot pause work belonging to a replacement owner epoch',()=>{
 const f=fixture();f.state.onStart=()=>{f.context.epoch++};
 assert.equal(f.context.api.start({environment:'new'}),false);assert.deepEqual(f.calls.starts,['logs']);assert.deepEqual(f.calls.pauses,[]);
});


test('a child local refusal is not reported as successful paired acquisition',()=>{
 const f=fixture();f.state.refuse={spans:true};assert.equal(f.context.api.start({environment:'new'}),false);
 assert.deepEqual(f.calls.starts,['logs','spans']);assert.deepEqual(f.calls.pauses,['logs']);
 for(const lane of Object.values(f.held()))assert.equal(lane.phase,'paused');
});


test('a replacement owner during first staging cannot be overwritten by the old sibling stage',()=>{
 const f=fixture();f.state.onStage=signal=>{if(signal==='logs'){f.state.onStage=null;assert.equal(f.context.api.stage({environment:'replacement'}),true)}};
 assert.equal(f.context.api.start({environment:'withdrawn'}),false);
 assert.deepEqual(f.calls.stages,['logs','logs','spans']);assert.deepEqual(f.calls.starts,[]);assert.deepEqual(f.calls.pauses,[]);
 assert.equal(f.context.selector.environment,'replacement');assert.equal(f.context.epoch,6);
 for(const lane of Object.values(f.held()))assert.equal(lane.selector.environment,'replacement');
});
