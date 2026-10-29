const {test}=require('node:test'),assert=require('node:assert/strict'),fs=require('node:fs');
const bank=JSON.parse(fs.readFileSync('testdata/protocol/environment-selectors.json','utf8'));
test('Go/browser environment selector bank preserves policy, canonical scopes and membership',async()=>{
 const E=await import('../ui/record-environment.mjs');
 for(const fixture of bank.cases){const normalized=E.normalize(fixture.selector);assert.deepEqual(normalized,fixture.canonical,fixture.name);assert.ok(E.equal(normalized,fixture.selector));assert.deepEqual(bank.resources.map(r=>E.matches(normalized,r)),fixture.matches,fixture.name);const clone=E.normalize(normalized);clone.scope.mode='broken';assert.deepEqual(E.normalize(fixture.selector),fixture.canonical,'normalization must detach')}
 for(const value of bank.invalid){assert.throws(()=>E.normalize(value));assert.equal(E.equal(value,value),false)}
 for(const raw of bank.invalidJSON)assert.throws(()=>E.normalize(JSON.parse(raw)));
 for(const raw of bank.validJSON)assert.doesNotThrow(()=>E.normalize(JSON.parse(raw)));
 assert.equal(E.equal(E.current('prod'),E.effective({mode:'named',value:'prod'})),false);
 assert.deepEqual(E.assertPreview(E.effective({mode:'unreported'}),{mode:'unreported'}),E.effective({mode:'unreported'}));
 assert.throws(()=>E.assertPreview(E.current('prod'),{mode:'named',value:'prod'}));
 assert.throws(()=>E.assertPreview(E.effective({mode:'empty'}),{mode:'unreported'}));
 assert.throws(()=>E.effective({mode:'named',value:'\ud800'}));
 assert.equal(E.matches(E.effective({mode:'unreported'}),null),false);
});
