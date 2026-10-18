const {test}=require('node:test');const assert=require('node:assert/strict');const {compare}=require('../ui/investigations.js');
test('revision comparison preserves separate evidence bases and exact span identity',()=>{
 const a={question:{fromMs:1,toMs:2},evidence:{operations:{spans:[{traceID:'a',spanID:'b'}],conformance:{basis:'one'}},rates:{basis:'rate-one'},errors:{histograms:'unavailable'}}};
 const b=structuredClone(a);b.evidence.operations.spans.push({traceID:'c',spanID:'d'});b.evidence.operations.conformance.basis='two';
 const result=compare(a,b);assert.equal(result.sameQuestion,true);assert.equal(result.addedSpanIdentities,1);assert.equal(result.removedSpanIdentities,0);assert.equal(result.bases.find(x=>x.component==='rates.basis').changed,false);assert.equal(result.bases.find(x=>x.component==='operations.conformance.basis').changed,true);assert.deepEqual(result.afterErrors,{histograms:'unavailable'});
});
