const {test}=require('node:test'),assert=require('node:assert/strict');
const A=require('../ui/view-artifact.mjs'),C=require('../ui/view-composition.mjs'),D=require('../ui/view-declaration.mjs'),H=require('../ui/view-template-handoff.mjs');
const labels={namespace:'payments',name:'trace-view',revision:'v1'};
function view(){const q={signal:'spans',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0,groupBy:[{kind:'column',key:'service'}]};return {version:1,title:'Trace work',queries:[q],composition:C.create(q,D.create(q),{columns:['utc','summary'],fieldColumns:[]})}}
const result=artifact=>({artifact,yaml:'version: goo.view-artifact/1\n'});
test('portable transport preserves a checked definition and accepts only known Go wire omissions',()=>{
 const v=view(),before=structuredClone(v),artifact=A.artifact(v,labels),wire=structuredClone(artifact);delete wire.view.composition.children[0].aggregate.display.query;delete wire.view.composition.children[0].aggregate.display.requiresComplete;delete wire.view.composition.children[1].records.display.fieldColumns;
 const missingLimit=structuredClone(artifact);delete missingLimit.view.queries[0].limit;assert.throws(()=>A.artifact(missingLimit.view,labels));assert.throws(()=>A.response(result(missingLimit),artifact));
 assert.deepEqual(A.response(result(wire),artifact).artifact,wire);assert.deepEqual(v,before);artifact.view.title='Changed';assert.deepEqual(v,before);
 const typed={...v.queries[0],environment:'',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},groupTimeSeries:false};assert.throws(()=>A.artifact({...v,queries:[{...typed,numeric:null}]},labels));const source=A.artifact({...v,queries:[typed]},labels),omitted=structuredClone(source);delete omitted.view.queries[0].groupTimeSeries;delete omitted.view.queries[0].environment;assert.doesNotThrow(()=>A.response(result(omitted),source));
});
test('codec response cannot substitute fixed scope or authored display intent',()=>{
 const expected=A.artifact(view(),labels);
 for(const mutate of [a=>a.name='other',a=>a.view.title='other',a=>a.view.queries[0].fromMs=0,a=>a.view.composition.children[0].aggregate.kind.name='table',a=>a.view.composition.children[0].aggregate.display.order='label',a=>a.view.composition.children[1].records.kind.name='record-list',a=>a.view.composition.children[1].records.display.columns.push('service'),a=>a.view.composition.children[1].records.display.fieldColumns=[{kind:'attribute',key:'other'}],a=>a.view.template=H.templateFor()]){const changed=structuredClone(expected);mutate(changed);assert.throws(()=>A.response(result(changed),expected));}
});
test('portable envelope refuses local evidence identities and runtime payloads without stripping them',()=>{
 const v=view();assert.deepEqual(A.artifact({...v,id:''},labels).view,v);for(const extra of [{id:'a'.repeat(64)},{rows:[]},{resultToken:'token'},{basis:'basis'},{unknown:true}])assert.throws(()=>A.artifact({...v,...extra},labels));
 const baseline=C.setComparison(v.composition,require('../ui/captured-count-comparison.mjs').create(v.queries[0],'a'.repeat(64)),v.queries[0]);assert.throws(()=>A.artifact({...v,composition:baseline},labels),/local to this store/);
 assert.throws(()=>A.artifact({version:1,title:'Tail',queries:v.queries,layout:'records',tail:{signal:'spans',baseline:'a'.repeat(64)}},labels),/local to this store/);
 for(const bad of [{...labels,namespace:'Bad'},{...labels,name:'bad-'},{...labels,revision:'../v1'},{...labels,id:'x'}])assert.throws(()=>A.artifact(v,bad));
 assert.throws(()=>A.yamlInput(''));assert.throws(()=>A.yamlInput('é'.repeat(32769)));assert.deepEqual(A.yamlInput('a: b'),{yaml:'a: b'});
});
test('fixed query baselines remain portable configuration without importing their observations',()=>{
 const q=view().queries[0],v={version:1,title:'Selection vs rest',queries:[q],layout:'charts-records',columns:['utc','summary'],breakdowns:{spans:{distribution:true,metric:'count',comparisonMetric:'count',baseline:{...q,where:{field:{kind:'column',key:'service'},op:'ne',value:'api'}}}}};assert.deepEqual(A.artifact(v,labels).view,v);
});
test('transport leases reject input ABA, owner changes and late hidden replies',()=>{
 const expected={active:true,generation:2,revision:3,serial:4,inputRevision:5};assert(A.current(expected,{...expected}));for(const change of [{active:false},{generation:3},{revision:4},{serial:5},{inputRevision:7}])assert.equal(A.current(expected,{...expected,...change}),false);assert.equal(A.current({...expected,active:false},expected),false);
});
test('fixed query comparisons travel as configuration while capture references remain local',()=>{
 const v=view(),q=v.queries[0],baseline={...q,where:{field:{kind:'column',key:'service'},op:'ne',value:'api'}};v.composition=C.setComparison(v.composition,require('../ui/captured-count-comparison.mjs').createQuery(q,baseline),q);const expected=A.artifact(v,labels),wire=structuredClone(expected);wire.view.composition.children.find(x=>x.comparison).comparison.display.order='';assert.doesNotThrow(()=>A.response(result(wire),expected));const changed=structuredClone(wire);changed.view.composition.children.find(x=>x.comparison).comparison.baselineQuery.where.value='other';assert.throws(()=>A.response(result(changed),expected));assert.deepEqual(A.artifact(v,labels).view,v);
});
test('fixed comparison transport accepts only known display and baseline-query wire omissions',()=>{
 const P=require('../ui/captured-count-comparison.mjs'),q={...view().queries[0],environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}}};delete q.environment;
 const baseline={...structuredClone(q),environment:'',bucketMs:0,summary:'',spanStatus:false,groupTimeSeries:false,where:{field:{kind:'column',key:'service'},op:'ne',value:'api'}};
 for(const key of ['numeric','numericHistogram','reportedEdge'])assert.throws(()=>P.createQuery(q,{...baseline,[key]:null}));
 const comparison={...P.createQuery(q,baseline),display:{value:'count',title:'',view:'',order:''}},v={version:1,title:'Two fixed queries',queries:[q],composition:C.setComparison(C.create(q,D.create(q)),comparison,q)},expected=A.artifact(v,labels),wire=structuredClone(expected),child=wire.view.composition.children.find(x=>x.comparison).comparison;
 for(const key of ['title','view','order'])delete child.display[key];for(const key of ['bucketMs','summary','spanStatus','groupTimeSeries'])delete child.baselineQuery[key];wire.view.queries[0].environment='';
 assert.doesNotThrow(()=>A.response(result(wire),expected));
 for(const mutate of [x=>x.baselineQuery.where.value='changed',x=>x.baselineQuery.fromMs=2000,x=>x.baselineQuery.groupBy=[{kind:'column',key:'name'}],x=>x.display.view='groups',x=>x.display.title='Changed']){const changed=structuredClone(wire);mutate(changed.view.composition.children.find(x=>x.comparison).comparison);assert.throws(()=>A.response(result(changed),expected));}
});
