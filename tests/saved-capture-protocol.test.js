const {test}=require('node:test'),assert=require('node:assert/strict'),P=require('../ui/saved-capture-protocol.mjs'),E=require('../ui/record-environment.mjs'),bank=require('../testdata/protocol/environment-selectors.json');
const id='a'.repeat(64),query=selector=>({signal:'logs',environmentSelector:structuredClone(selector),fromMs:1000,toMs:2000,limit:100,offset:0}),fields=[{kind:'column',key:'service'}];
test('Saved selectors preserve receipt scope across stored paging and supported comparison/explanation families',()=>{
 for(const fixture of bank.cases){const q=query(fixture.selector),request={version:'goo/explanation/v1',baseline:structuredClone(q),selection:structuredClone(q),fields},pages=[{receipt:{id,question:{kind:'explore',explore:q}},explore:{question:{...structuredClone(q),environment:'',offset:100}}},{receipt:{id,question:{kind:'comparison',comparison:{baseline:q,selection:q}}},comparison:{request:{baseline:{question:structuredClone(q)},selection:{question:structuredClone(q)}}}},{receipt:{id,question:{kind:'explanation',explanation:request}},explanation:{request:structuredClone(request)}}];
  for(const page of pages){const before=structuredClone(page);assert.equal(P.validate(page,id),page);assert.deepEqual(page,before);assert.throws(()=>P.validate(page,'b'.repeat(64)));const bad=structuredClone(page),actual=bad.explore?.question||bad.comparison?.request.selection.question||bad.explanation?.request.selection;actual.environmentSelector=E.effective({mode:fixture.selector.scope.mode==='empty'?'unreported':'empty'});assert.throws(()=>P.validate(bad,id));}
 }
});
test('Saved boundary refuses scope widening, invalid pages and unsupported selector families',()=>{
 const q=query(E.effective({mode:'unreported'})),page={receipt:{id,question:{kind:'explore',explore:q}},explore:{question:structuredClone(q)}};
 for(const mutate of [p=>p.explore.question.offset=1,p=>p.explore.question.limit=200,p=>p.explore.question.toMs++,p=>{delete p.explore.question.environmentSelector;p.explore.question.environment='prod'},p=>p.explore.question.where={field:fields[0],op:'eq',value:'foreign'},p=>p.receipt.question.kind='measurement']){const bad=structuredClone(page);mutate(bad);assert.throws(()=>P.validate(bad,id));}
 const legacy={receipt:{id,question:{kind:'explore',explore:{environment:'prod'}}}};assert.equal(P.validate(legacy,id),legacy);
});
test('refresh requires the requested predecessor and exact immutable selector question',()=>{
 const q=query(E.effective({mode:'unreported'})),before={receipt:{id,question:{kind:'explore',explore:q}},explore:{question:structuredClone(q)}},fresh=structuredClone(before);fresh.receipt.id='b'.repeat(64);fresh.receipt.parent=id;assert.equal(P.validateRefresh(fresh,before),fresh);
 for(const mutate of [p=>p.receipt.parent='c'.repeat(64),p=>{p.receipt.question.explore.environmentSelector.scope={mode:'empty'};p.explore.question.environmentSelector.scope={mode:'empty'}}]){const bad=structuredClone(fresh);mutate(bad);assert.throws(()=>P.validateRefresh(bad,before));}
 const legacy={receipt:{id,question:{kind:'explore',explore:{...q,environment:'prod'}}},explore:{question:structuredClone(q)}};delete legacy.receipt.question.explore.environmentSelector;assert.throws(()=>P.validate(legacy,id));
});
