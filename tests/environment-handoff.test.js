const {test}=require('node:test'),assert=require('node:assert/strict'),{activitySelector}=require('../ui/environment-adoption.mjs'),E=require('../ui/record-environment.mjs'),{RecordScopeDraft}=require('../ui/record-scope-draft.mjs'),bank=require('../testdata/protocol/environment-selectors.json');
test('preview handoff accepts only its checked effective scope and preserves detached policy',()=>{
 for(const fixture of bank.cases){for(const signal of ['logs','spans']){
  if(fixture.canonical.policy!=='effective'){assert.throws(()=>activitySelector(signal,fixture.selector,fixture.canonical.scope));continue}
  const original=structuredClone(fixture.selector),handoff=activitySelector(signal,original,fixture.canonical.scope),owner=new RecordScopeDraft('prod');owner.restore(handoff);assert.equal(handoff.signal,signal);assert.equal(handoff.environment,'');assert.ok(E.equal(handoff.environmentSelector,fixture.canonical));assert.ok(owner.matches(handoff));handoff.environmentSelector.scope.mode='broken';assert.deepEqual(original,fixture.selector);assert.ok(owner.matches({environmentSelector:fixture.canonical}));owner.setName('replacement');assert.equal(owner.snapshot().environmentSelector,undefined);
  assert.throws(()=>activitySelector(signal,fixture.selector,{mode:fixture.canonical.scope.mode==='empty'?'unreported':'empty'}));
 }}
 assert.throws(()=>activitySelector('metrics',E.effective({mode:'empty'}),{mode:'empty'}));
});
