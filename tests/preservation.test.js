const {test}=require('node:test'),assert=require('node:assert/strict'),P=require('../ui/preservation.mjs');
const query=signal=>({signal,environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},fromMs:1000,toMs:2000,limit:10,offset:20,where:{any:[{field:{kind:'attribute',key:'region'},op:'missing'},{field:{kind:'attribute',key:'region'},op:'eq',value:''}]}});
test('Preserve distinguishes configuration from independent evidence reads',()=>{
 assert.match(P.meaning('question').note,/no telemetry read/);assert.match(P.meaning('view').note,/no telemetry read/);assert.match(P.meaning('evidence').note,/Confirmation makes a new independent read/);
 const copy=P.meaning('question');copy.note='changed';assert.notEqual(P.meaning('question').note,copy.note);assert.throws(()=>P.meaning('capture'));
});
test('Preserve requires explicit multi-signal choice and retains independent typed root intent',()=>{
 const queries=['spans','logs'].map(query),before=structuredClone(queries);assert.throws(()=>P.selectedQuestion(queries),/Choose a signal/);
 for(const signal of ['spans','logs']){const selected=P.selectedQuestion(queries,signal);assert.deepEqual(selected,{...before.find(q=>q.signal===signal),offset:0});selected.where.any[1].value='changed';selected.environmentSelector.scope.value='changed';assert.deepEqual(queries,before)}
 assert.equal(P.selectedQuestion([queries[1]]).signal,'logs');assert.throws(()=>P.selectedQuestion(queries,'metrics'));assert.throws(()=>P.selectedQuestion([queries[0],queries[0]],'spans'));assert.throws(()=>P.selectedQuestion([]));assert.throws(()=>P.selectedQuestion([{...queries[0],fromMs:NaN}]));assert.throws(()=>P.selectedQuestion([{...queries[0],unexpected:'not durable'}]));
});
test('preservation refuses stale owners, in-flight reads, pending edits and unapplied evidence',()=>{
 assert.equal(P.availability(),'');assert.match(P.availability({active:false}),/active context/);assert.match(P.availability({connected:false}),/active context/);assert.match(P.availability({busy:true}),/current read/);assert.match(P.availability({pending:true}),/pending selections/);assert.match(P.availability({applied:false}),/Run this question/);
});
