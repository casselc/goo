const {test}=require('node:test'),assert=require('node:assert/strict');
const E=require('../ui/record-environment.mjs'),P=require('../ui/question-protocol.mjs'),S=require('../ui/question-session.mjs'),T=require('../ui/tail.mjs'),L=require('../ui/activity-latency.mjs'),A=require('../ui/activity-explanation.mjs');
const bank=require('../testdata/protocol/environment-selectors.json');
test('Activity builders retain typed selectors through shared fixed-question links',()=>{
 for(const fixture of bank.cases){const selector={signal:'spans',environmentSelector:structuredClone(fixture.selector),where:{field:{kind:'attribute',key:'release'},op:'eq',value:'canary'}},context={selector,readWindow:{fromMs:120000,toMs:180000}},before=structuredClone(context),q=L.query(context),window=A.windowQuestion(context);
  assert.deepEqual(E.questionFields(q),E.questionFields(selector));assert.deepEqual(E.questionFields(window),E.questionFields(selector));assert.doesNotThrow(()=>P.validateRecordQuestion(q));assert.doesNotThrow(()=>P.validateLink({version:1,question:q,view:'Records'}));const shared=P.readQuestionLink(P.questionLink(q));assert.deepEqual(shared,{version:1,question:q,view:'Records'});assert.deepEqual(shared.question.environmentSelector,fixture.selector);assert.deepEqual(shared.question.where,selector.where);assert.equal(shared.question.fromMs,120000);assert.equal(shared.question.toMs,180000);const malformed=structuredClone(q);malformed.environmentSelector.policy='future';assert.throws(()=>P.questionLink(malformed));
  const previous=A.previousQuestion(window),narrow=A.selectWindow(window,130000,150000);for(const child of [previous,narrow]){assert.ok(E.sameQuestionEnvironment(child,q));child.environmentSelector.scope.mode='broken';assert.deepEqual(context,before)}
  assert.equal(L.scopeKey(context),L.scopeKey({...context,readWindow:{fromMs:180000,toMs:240000}}));
  const session=new S.Session(),token=session.start(q),returned=structuredClone(q);returned.environment='';returned.environmentSelector=E.normalize(returned.environmentSelector);assert.equal(session.publish(token,{question:returned,basis:'observed',count:'1',state:'observed'}),true);assert.deepEqual(context,before);
 }
});
test('scope policy changes invalidate reads and saved baseline compatibility',()=>{
 const scopes=[E.current('prod'),E.effective({mode:'named',value:'prod'}),E.effective({mode:'empty'}),E.effective({mode:'unreported'})];
 for(let i=0;i<scopes.length;i++)for(let j=0;j<scopes.length;j++){
  const make=selector=>({selector:{signal:'spans',environmentSelector:selector},readWindow:{fromMs:120000,toMs:180000}}),q=L.query(make(scopes[i])),other=L.query(make(scopes[j]));assert.equal(L.scopeKey(make(scopes[i]))===L.scopeKey(make(scopes[j])),i===j);
  const s=new S.Session(),token=s.start(q);if(i===j)assert.equal(s.publish(token,{question:other}),true);else assert.throws(()=>s.publish(token,{question:other}),/does not match/);
  const id='a'.repeat(64),page={receipt:{id,capturedAt:'2026-10-29T00:00:00Z',question:{kind:'explore',explore:q}},explore:{question:q}};if(i===j)assert.deepEqual(A.savedScope(page,id,other),q);else assert.throws(()=>A.savedScope(page,id,other),/same signal and environment/);
 }
});
test('tail response identity handles Go omission defaults without erasing selector policy',()=>{
 for(const fixture of bank.cases){const input={signal:'logs',environmentSelector:fixture.selector,where:null},output={environment:'',environmentSelector:fixture.canonical};assert.equal(T.transportIdentity(input),T.transportIdentity(output));}
 assert.equal(T.transportIdentity({environment:'prod',environmentSelector:null}),T.transportIdentity({signal:'logs',environment:'prod'}));
 assert.notEqual(T.transportIdentity({environment:'prod'}),T.transportIdentity({environmentSelector:E.current('prod')}));
 assert.notEqual(T.transportIdentity({environmentSelector:E.effective({mode:'empty'})}),T.transportIdentity({environmentSelector:E.effective({mode:'unreported'})}));
 assert.throws(()=>T.transportIdentity({environment:'prod',environmentSelector:E.current('prod')}));
 const q={signal:'spans',environmentSelector:E.current('prod'),fromMs:1,toMs:2,reportedEdge:{parent:{},child:{}}};assert.throws(()=>P.validateRecordQuestion(q),/parent edges/);
});
test('invalid selector starts leave the retained session unchanged and never reinterpret empty legacy names',()=>{
 const s=new S.Session(),q={signal:'logs',environment:'prod',fromMs:1,toMs:2},token=s.start(q);s.publish(token,{question:q});const before=s.snapshot();
 for(const bad of [{environment:''},{environment:'prod',environmentSelector:E.effective({mode:'empty'})},{environmentSelector:{version:'future',policy:'effective',scope:{mode:'empty'}}}]){const next={signal:'logs',fromMs:1,toMs:2,...bad};assert.throws(()=>s.start(next));assert.deepEqual(s.snapshot(),before);assert.throws(()=>P.validateRecordQuestion(next));}
 const old={environment:'prod'},explicit={environment:'',environmentSelector:E.current('prod')};assert.ok(E.sameQuestionEnvironment(old,explicit));assert.notEqual(S.identity(old),S.identity(explicit));
});
