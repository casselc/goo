'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict');
const {fixture}=require('./fixtures/native-histogram-comparison.cjs');
const {asHistogramComparison}=require('../ui/snapshots.mjs');
const {createHistogramComparisonActions}=require('../ui/histogram-comparison-actions.mjs');
const {createActionContext}=require('../ui/scope-actions.mjs');
const {histogramIntent}=require('../ui/view-model.mjs');

test('native window comparisons remain passive and cannot acquire record range owners',()=>{
 const pair=asHistogramComparison(fixture());
 assert.throws(()=>createHistogramComparisonActions({comparison:pair,active:()=>true}),/numeric histogram operands/);
 const context=createActionContext({epoch:()=>1,active:()=>true,canStage:()=>true});
 for(const snapshot of [pair.baseline,pair.selection]){
  assert.throws(()=>histogramIntent(snapshot,'bin:0'),/native|instrument|record/i);
  assert.throws(()=>context.bind({role:'current',slot:'histogram',epoch:1,snapshot,question:snapshot.scope.selector,coherent:true}),error=>error.code==='scope-family');
 }
 assert.equal(context.canAct(pair),false);
});
