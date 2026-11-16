import {test} from 'node:test';
import assert from 'node:assert/strict';
import {initial,adopt,transition,withContext,contexts,pages,topics} from '../ui/guidance-content.mjs';
test('closed immutable reading state preserves page across every context',()=>{
 let state=initial();assert.equal(Object.isFrozen(state),true);assert.equal(state.guidePage,null);state=transition(state,'start-guide');
 for(let index=0;index<20;index++)state=transition(state,'next');assert.equal(state.guidePage,4);
 for(const context of contexts){state=withContext(state,context);assert.equal(state.guidePage,4);assert.equal(state.context,context);assert.equal(Object.isFrozen(state),true)}
 assert.equal(transition(state,'replay').guidePage,0);assert.equal(transition(state,'about').guidePage,null);
 for(let index=0;index<20;index++)state=transition(state,'previous');assert.equal(state.guidePage,0);
 assert.throws(()=>transition(state,'run'));assert.equal(Object.isFrozen(pages[0]),true);assert.equal(Object.isFrozen(topics.activity.topics[0]),true);
});
test('admission rejects getters without invocation, exotic objects, cycles, unknown fields and out of bounds state',()=>{
 let invoked=0;const getter={context:'activity',guidePage:null,get version(){invoked++;return 'goo.guidance/1'}};
 for(const input of [null,[],new Date(),getter,{...initial(),extra:true},{...initial(),guidePage:-1},{...initial(),guidePage:5},{...initial(),guidePage:1.5},{...initial(),context:'constructor'}])assert.throws(()=>adopt(input));
 const cycle={...initial()};cycle.guidePage=cycle;assert.throws(()=>adopt(cycle));assert.equal(invoked,0);const symbol={...initial(),[Symbol()]:1};assert.throws(()=>adopt(symbol));
});
