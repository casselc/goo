const {test}=require('node:test'),assert=require('node:assert/strict');
const E=require('../ui/record-environment.mjs'),{traceQuestion}=require('../ui/record-results.mjs'),{surroundingQuestion}=require('../ui/discovery.mjs'),{inspectionQuestion}=require('../ui/tail.mjs');
const bank=require('../testdata/protocol/environment-selectors.json');
const trace='0123456789abcdef0123456789abcdef';
test('record context constructors preserve every checked selector representation without aliasing',()=>{
 for(const fixture of bank.cases)for(const legacy of [{},{environment:''}]){
  const q={signal:'logs',...legacy,environmentSelector:structuredClone(fixture.selector),fromMs:1000,toMs:61000,where:{field:{kind:'column',key:'body'},op:'contains',value:'private'}},before=structuredClone(q);
  const contexts=[traceQuestion(q,trace),traceQuestion(q,trace,'0123456789abcdef','logs'),surroundingQuestion(q,{timestamp_ns:'2000000000',ServiceName:'worker'},1000),inspectionQuestion(q,'spans',{readWindow:{fromMs:1000,toMs:61000}})];
  for(const context of contexts){assert.deepEqual(E.questionFields(context),E.questionFields(q),fixture.name);assert.ok(E.sameQuestionEnvironment(context,q));assert.notEqual(context.environmentSelector,q.environmentSelector);assert.notEqual(context.environmentSelector.scope,q.environmentSelector.scope);context.environmentSelector.scope.mode='broken';assert.deepEqual(q,before);assert.equal(JSON.stringify(context.where||{}).includes('private'),false)}
 }
});
test('semantic equality preserves current policy and distinguishes empty, absent and effective names',()=>{
 const legacy={environment:'prod'},explicit={environmentSelector:E.current('prod')};assert.ok(E.sameQuestionEnvironment(legacy,explicit));
 const values=[legacy,{environmentSelector:E.effective({mode:'named',value:'prod'})},{environmentSelector:E.effective({mode:'empty'})},{environmentSelector:E.effective({mode:'unreported'})}];
 for(let i=0;i<values.length;i++)for(let j=0;j<values.length;j++)assert.equal(E.sameQuestionEnvironment(values[i],values[j]),i===j);
 assert.deepEqual(E.questionFields(legacy),legacy);assert.deepEqual(E.questionFields({environment:'prod',environmentSelector:null}),{environment:'prod',environmentSelector:null});
 for(const q of [{},{environment:''},{environment:'prod',environmentSelector:E.current('prod')},{environment:null,environmentSelector:E.effective({mode:'empty'})},...bank.invalid.map(environmentSelector=>({environmentSelector}))]){assert.throws(()=>E.questionFields(q));assert.equal(E.sameQuestionEnvironment(q,q),false)}
});
test('context builders reject ambiguous environment input before constructing a replacement question',()=>{
 const q={environment:'prod',environmentSelector:E.effective({mode:'empty'}),fromMs:1000,toMs:61000};
 assert.throws(()=>traceQuestion(q,trace));assert.throws(()=>surroundingQuestion(q,{timestamp_ns:'2000000000',ServiceName:'worker'}));assert.throws(()=>inspectionQuestion(q,'logs',{readWindow:{fromMs:1000,toMs:61000}}));
});
