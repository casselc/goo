import {test} from 'node:test';
import assert from 'node:assert/strict';
import {populationScopePresentation} from '../ui/population-workspace.mjs';
const selector=(mode='named',value='prod',policy='effective')=>({version:'goo/environment-selector/v1',policy,scope:{mode,...(mode==='named'?{value}:{})}});
const question=()=>({signal:'logs',environmentSelector:selector(),fromMs:61000,toMs:121000,limit:100,offset:0,groupBy:[{kind:'column',key:'service'}],where:{field:{kind:'attribute',key:'tenant'},op:'eq',value:'selected'}});
test('compact scope facts retain exact UTC bounds, full policies and distinct baseline definition identity',()=>{
 const q=question(),baseline={...structuredClone(q),fromMs:1000,toMs:61000,where:{field:{kind:'attribute',key:'tenant'},op:'eq',value:'baseline'}},before=structuredClone({q,baseline}),view=populationScopePresentation([q],baseline);
 assert.equal(view.environment,'"prod"');assert.equal(view.policy,'effective');assert.match(view.environmentDetail,/policy effective/);assert.match(view.baselineEnvironmentDetail,/policy effective/);assert.equal(view.baselineLabel,'Previous window');assert.equal(view.selectedBounds,'[1970-01-01T00:01:01.000Z, 1970-01-01T00:02:01.000Z)');assert.equal(view.baselineBounds,'[1970-01-01T00:00:01.000Z, 1970-01-01T00:01:01.000Z)');assert.match(view.selectedWindow,/UTC \(end excluded\)/);assert.deepEqual({q,baseline},before);
 for(const change of [x=>x.where.value='other',x=>x.limit++,x=>x.environmentSelector.policy='current-only',x=>x.groupBy=[{kind:'column',key:'scope'}]]){const next=structuredClone(baseline);change(next);assert.notEqual(populationScopePresentation([q],next).key,view.key)}
 const fixed={...baseline,fromMs:q.fromMs,toMs:q.toMs};assert.equal(populationScopePresentation([q],fixed).baselineLabel,'Fixed baseline window');
});
test('empty and unreported names cannot collapse into named strings or hide effective policy',()=>{
 const facts=[];for(const [mode,value]of [['empty',''],['unreported',''],['named','Empty environment label'],['named','Unreported environment']]){const q={...question(),environmentSelector:selector(mode,value)},view=populationScopePresentation([q]);facts.push(view);assert.match(view.environmentDetail,/policy effective/);assert.equal(view.baselineLabel,null);assert.equal(view.baselineBounds,null)}
 assert.equal(new Set(facts.map(view=>view.environment)).size,4);assert.equal(new Set(facts.map(view=>view.key)).size,4);
 const q=question();delete q.environmentSelector;q.environment='prod';assert.match(populationScopePresentation([q]).environmentDetail,/policy current-only/);
});
test('scope presentation performs no IO, preserves source objects and shows cross-day exact UTC',()=>{
 const q={...question(),fromMs:86399000,toMs:86401000},before=structuredClone(q),fetch=globalThis.fetch;globalThis.fetch=()=>{throw Error('presentation performed IO')};
 try{const view=populationScopePresentation([q]);assert.match(view.selectedWindow,/1970-01-01T23:59:59.000Z–1970-01-02T00:00:01.000Z/);assert.deepEqual(q,before);assert.throws(()=>populationScopePresentation([{...q,toMs:86402000},q]));}finally{globalThis.fetch=fetch}
});
