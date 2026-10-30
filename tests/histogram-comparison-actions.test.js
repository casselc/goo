'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict');
const {fixture}=require('./fixtures/histogram-comparison.cjs');
const {asHistogramComparison}=require('../ui/snapshots.mjs');
const {histogramIntent}=require('../ui/view-model.mjs');
const {createHistogramComparisonActions}=require('../ui/histogram-comparison-actions.mjs');
function pair(){const raw=fixture();raw.baseline.scope.selector.fromMs=1000;raw.baseline.scope.selector.toMs=2000;raw.selection.scope.selector.fromMs=3000;raw.selection.scope.selector.toMs=4000;raw.baseline.capturedAt='2026-10-30T10:00:00Z';raw.selection.capturedAt='2026-10-30T11:00:00Z';return asHistogramComparison(raw)}
const action=(comparison,side,key)=>({side,intent:histogramIntent(comparison[side],key,key)});
const denied=(result,code)=>{assert.equal(result.ok,false);assert.equal(result.code,code)};
test('explicit side planning retains independent captured questions and coefficient bounds',()=>{
 const comparison=pair(),owner=createHistogramComparisonActions({comparison,active:()=>true});
 const a=owner.resolve(action(comparison,'baseline','bin:2')),b=owner.resolve(action(comparison,'selection','bin:5'));
 assert.equal(a.ok,true);assert.equal(b.ok,true);assert.equal(a.count,'3');assert.equal(b.count,'2');
 assert.equal(a.question,comparison.baseline.scope.selector);assert.equal(b.question,comparison.selection.scope.selector);assert.equal(a.question.fromMs,1000);assert.equal(b.question.fromMs,3000);assert(Object.isFrozen(a));assert(Object.isFrozen(a.question));
 assert.equal(a.from,'2');assert.equal(a.to,'4');assert.equal(b.from,'16');assert.equal(b.to,'32');
});
test('range intents cannot switch sides, forge identities or select empty bins',()=>{
 const comparison=pair(),owner=createHistogramComparisonActions({comparison,active:()=>true}),selected=action(comparison,'selection','bin:5');
 denied(owner.resolve({...selected,side:'baseline'}),'identity-mismatch');
 denied(owner.resolve({side:'current',intent:selected.intent}),'invalid-side');
 denied(owner.resolve({...selected,intent:{...selected.intent,toKey:'bin:999'}}),'invalid-membership');
 denied(owner.resolve({...selected,intent:{...selected.intent,fromKey:'bin:6'}}),'invalid-membership');
 denied(owner.resolve(action(comparison,'selection','bin:1')),'empty-selection');
 denied(owner.resolve({...selected,intent:{...selected.intent,extra:true}}),'invalid-intent');
});
test('an empty side does not disable explicit planning of the nonempty side',()=>{
 const raw=fixture();raw.baseline.histogram.support={total:'0',valid:'0',missing:'0',empty:'0',invalid:'0'};for(const row of raw.baseline.rows){row.values.count='0';row.state='no_data'}raw.state='no_data';
 const comparison=asHistogramComparison(raw),owner=createHistogramComparisonActions({comparison,active:()=>true});
 denied(owner.resolve(action(comparison,'baseline','bin:2')),'empty-selection');assert.equal(owner.resolve(action(comparison,'selection','bin:5')).ok,true);
});
test('diagnostic or withdrawn pairs cannot plan even when their counts remain known',async()=>{
 const raw=fixture();raw.state='diagnostic';raw.reason='source-reads-disagree';const comparison=asHistogramComparison(raw),diagnostic=createHistogramComparisonActions({comparison,active:()=>true});
 denied(diagnostic.resolve(action(comparison,'baseline','bin:2')),'diagnostic');
 const observed=pair();let active=true;const owner=createHistogramComparisonActions({comparison:observed,active:()=>active}),intent=action(observed,'baseline','bin:2');assert.equal(owner.resolve(intent).ok,true);
 await Promise.resolve();owner.revoke();const fresh=createHistogramComparisonActions({comparison:observed,active:()=>active});assert.equal(fresh.resolve(intent).ok,true);denied(owner.resolve(intent),'retired');active=false;denied(fresh.resolve(intent),'inactive');
});
test('held maximum geometry remains a sentinel rather than a numeric threshold',()=>{
 const raw=fixture();raw.selection.rows[5].values.count='0';raw.selection.rows[5].state='no_data';raw.selection.rows.at(-1).values.count='2';raw.selection.rows.at(-1).state='observed';
 const comparison=asHistogramComparison(raw),owner=createHistogramComparisonActions({comparison,active:()=>true}),result=owner.resolve(action(comparison,'selection','bin:64'));
 assert.equal(result.ok,true);assert.equal(result.to,'18446744073709551616');assert.equal(result.toBin,65);assert.equal(result.question.where,undefined);
});
test('unbranded pairs, asynchronous owner state and hostile accessors fail closed',()=>{
 assert.throws(()=>createHistogramComparisonActions({comparison:fixture(),active:()=>true}),TypeError);
 const comparison=pair(),owner=createHistogramComparisonActions({comparison,active:()=>Promise.resolve(true)});denied(owner.resolve(action(comparison,'baseline','bin:2')),'owner-state');
 const current=createHistogramComparisonActions({comparison,active:()=>true}),hostile={side:'baseline'};let invoked=0;Object.defineProperty(hostile,'intent',{enumerable:true,get(){invoked++;return {}}});denied(current.resolve(hostile),'invalid-side');assert.equal(invoked,0);
});
