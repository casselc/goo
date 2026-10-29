const {test}=require('node:test'),assert=require('node:assert/strict');
const {RecordScopeDraft}=require('../ui/record-scope-draft.mjs'),E=require('../ui/record-environment.mjs'),bank=require('../testdata/protocol/environment-selectors.json');
test('retained record scope survives independent filter, signal and fixed/fresh bound changes',()=>{
 for(const fixture of bank.cases){const origin={signal:'logs',environmentSelector:structuredClone(fixture.selector),fromMs:1000,toMs:61000},before=structuredClone(origin),draft=new RecordScopeDraft();draft.restore(origin);
  for(const signal of ['logs','spans'])for(const bounds of [{fromMs:1000,toMs:61000},{fromMs:61000,toMs:121000}])for(const where of [undefined,{field:{kind:'column',key:'service'},op:'eq',value:'worker'}]){
   const next={signal,...draft.snapshot(),...bounds,where};assert.ok(draft.matches(next));assert.ok(E.sameQuestionEnvironment(next,origin));next.environmentSelector.scope.mode='changed';assert.deepEqual(origin,before);assert.ok(draft.matches(origin));
  }
  const fields=draft.snapshot();fields.environmentSelector.scope.mode='changed';assert.ok(draft.matches(origin));
  for(const bad of [{environment:''},{environment:'prod',environmentSelector:E.effective({mode:'empty'})}]){assert.throws(()=>draft.restore(bad));assert.ok(draft.matches(origin));}
  draft.setName('prod');assert.deepEqual(draft.snapshot(),{environment:'prod'});assert.equal(draft.matches(origin),false);assert.ok(draft.matches({environment:'prod'}));
  draft.setName('');assert.equal(draft.matches({environment:''}),false);assert.throws(()=>E.questionFields(draft.snapshot()));
 }
});
test('scope ownership distinguishes representation and policy changes without conflating empty or missing names',()=>{
 const draft=new RecordScopeDraft('prod');assert.ok(draft.matches({environment:'prod',environmentSelector:null}));assert.equal(draft.matches({environmentSelector:E.current('prod')}),false);
 const scopes=[E.current('prod'),E.effective({mode:'named',value:'prod'}),E.effective({mode:'empty'}),E.effective({mode:'unreported'})];
 for(let i=0;i<scopes.length;i++){draft.restore({environmentSelector:scopes[i]});for(let j=0;j<scopes.length;j++)assert.equal(draft.matches({environment:'',environmentSelector:scopes[j]}),i===j)}
});
