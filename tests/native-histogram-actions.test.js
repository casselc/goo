'use strict';
const {test}=require('node:test');
const assert=require('node:assert/strict');
const {fixture}=require('./fixtures/native-histogram.cjs');
const {createSourceAdapter,importSnapshot,hasGuarantee}=require('../ui/snapshots.mjs');
const {createActionContext}=require('../ui/scope-actions.mjs');
const {histogramIntent}=require('../ui/view-model.mjs');

test('instrument observation snapshots cannot acquire record-population action bindings',()=>{
 const raw=fixture();
 assert(raw,'the known native interval must construct');
 const source=createSourceAdapter({id:'native.actions'});
 const snapshot=source.adoptSnapshot(raw,{question:raw.scope.selector});
 assert.equal(hasGuarantee(snapshot,'bins'),true);
 const context=createActionContext({epoch:()=>1,active:()=>true,canStage:()=>true});
 for(const value of [snapshot,importSnapshot(snapshot)]){
  for(const role of ['current','baseline']){
   assert.throws(()=>context.bind({role,slot:'histogram',epoch:1,snapshot:value,question:value.scope.selector,coherent:true}),error=>error.code==='scope-family');
  }
  assert.throws(()=>histogramIntent(value,'bin:0'),/native|instrument|record/i);
 }
 assert.equal(context.canAct({}),false);
 for(const method of ['histogram','duration','resolve','inspect']){
  const refused=context[method]({},{type:'scope-histogram'});
  assert.equal(refused.ok,false);
  assert.equal(refused.code,'foreign-binding');
 }
});
