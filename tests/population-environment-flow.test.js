const {test}=require('node:test'),assert=require('node:assert/strict'),E=require('../ui/record-environment.mjs'),P=require('../ui/population-workspace.mjs'),S=require('../ui/population-selection.mjs'),X=require('../ui/explanation.mjs'),L=require('../ui/log-patterns.mjs'),C=require('../ui/log-pattern-comparison.mjs');
const bank=require('../testdata/protocol/environment-selectors.json'),field={kind:'column',key:'service'},base=environmentSelector=>({signal:'spans',environmentSelector:structuredClone(environmentSelector),fromMs:120000,toMs:180000,limit:100,offset:0,groupBy:[field]});
test('population transforms, region/rest and pattern planning retain detached checked scopes',()=>{
 for(const fixture of bank.cases){const q=base(fixture.selector),before=structuredClone(q),queries=[q,{...structuredClone(q),signal:'logs'}],intent=[{signal:'spans',field,value:'api',mode:'include'}];
  const rectangle=S.partition(q,{fromMs:130000,toMs:150000,minMS:'1',maxMS:'5'}),parts=P.partitionQueries(queries,intent),transforms=[...P.rangeQueries(queries,{fromMs:130000,toMs:150000}),...P.widenQueries(queries,{fromMs:60000,toMs:180000}),...P.selectionQueries(queries,intent),...parts.selected,...parts.rest,rectangle.selected,rectangle.rest,P.recentQuestion(q,240000,1)];
  for(const out of transforms){assert.ok(E.sameQuestionEnvironment(out,q));assert.notEqual(out.environmentSelector,q.environmentSelector);out.environmentSelector.scope.mode='broken';assert.deepEqual(q,before)}
  const patterns=L.question({...q,signal:'logs'}),previous=C.previousQuestion(patterns);assert.ok(C.isPair({baseline:{question:previous},selection:{question:patterns}}));assert.ok(E.sameQuestionEnvironment(C.plannedPair(patterns,previous).baseline,q));
  assert.throws(()=>P.definition('Unsupported',queries,'records',['utc']));
 }
});
test('population plans reject changed policy, dropped scope and altered source before execution',()=>{
 for(const fixture of bank.cases){const q=base(fixture.selector),other={...structuredClone(q),signal:'logs'},plan={version:'goo/population/v1',queries:[q,other],notices:[]};assert.deepEqual(P.plannedQueries(q,undefined,plan),plan.queries);
  for(const mutation of [p=>p.queries[0].environmentSelector=E.effective({mode:q.environmentSelector.scope.mode==='empty'?'unreported':'empty'}),p=>{delete p.queries[0].environmentSelector;p.queries[0].environment='local'},p=>p.queries[1].fromMs++,p=>p.queries[1].signal='spans',p=>p.queries[0].where={field,op:'eq',value:'foreign'},p=>p.queries.push(other)]){const bad=structuredClone(plan);mutation(bad);assert.throws(()=>P.plannedQueries(q,undefined,bad));}
  const out=P.plannedQueries(q,other,plan);out[0].environmentSelector.scope.mode='broken';assert.deepEqual(q,plan.queries[0]);
 }
});
test('comparison/explanation and aligned population scope refuse mismatched environment policies',()=>{
 const scopes=[E.current('prod'),E.effective({mode:'named',value:'prod'}),E.effective({mode:'empty'}),E.effective({mode:'unreported'})];
 for(let i=0;i<scopes.length;i++)for(let j=0;j<scopes.length;j++){
  const a=base(scopes[i]),b=base(scopes[j]),request={version:'goo/explanation/v1',baseline:a,selection:b,fields:[field]},patterns={baseline:{question:L.question({...a,signal:'logs'})},selection:{question:L.question({...b,signal:'logs'})}};
  assert.equal(C.isPair(patterns),i===j);if(i===j){assert.doesNotThrow(()=>X.validate(request));assert.doesNotThrow(()=>P.widenQueries([a,b],{fromMs:60000,toMs:180000}));}else{assert.throws(()=>X.validate(request));assert.throws(()=>P.rangeQueries([a,b],{fromMs:130000,toMs:150000}));assert.throws(()=>P.widenQueries([a,b],{fromMs:60000,toMs:180000}));}
 }
 const legacy={...base(E.current('prod')),environment:'prod'};delete legacy.environmentSelector;assert.doesNotThrow(()=>X.validate({version:'goo/explanation/v1',baseline:legacy,selection:base(E.current('prod')),fields:[field]}));
});
test('numeric side continuations cannot change scope, policy, binding or parent filters',()=>{
 const {checkedSelected}=require('../ui/numeric-distribution-comparison.mjs');
 for(const fixture of bank.cases){const parent={...base(fixture.selector),where:{field,op:'eq',value:'api'},numeric:{field:{kind:'attribute',key:'size'},type:'uint64',unit:'bytes',parse:'canonical-decimal'},numericHistogram:{binning:'powers-of-two'}},selected={...structuredClone(parent),where:{all:[structuredClone(parent.where),{numeric:structuredClone(parent.numeric),op:'gte',value:'2'}]}};assert.equal(checkedSelected(parent,selected),selected);
  for(const change of [q=>q.environmentSelector=E.effective({mode:parent.environmentSelector.scope.mode==='empty'?'unreported':'empty'}),q=>q.signal='logs',q=>q.toMs++,q=>q.numeric.unit='items',q=>q.where.all[0].value='foreign']){const bad=structuredClone(selected);change(bad);assert.throws(()=>checkedSelected(parent,bad));}
 }
});
