const {test}=require('node:test'),assert=require('node:assert/strict');
const C=require('../ui/view-composition.mjs'),D=require('../ui/view-declaration.mjs'),H=require('../ui/view-template-handoff.mjs');
const query=()=>({signal:'spans',environment:'prod',fromMs:1000,toMs:2000,limit:1,offset:0,groupBy:[{kind:'column',key:'service'}],summary:'span-duration'});
const make=()=>{const q=query();return {q,c:C.create(q,D.create(q),{columns:['utc','summary'],fieldColumns:[{kind:'attribute',key:'http.route'}]})}};
test('a checked composition orders two supplied aspects without duplicating scope or query settings',()=>{const {q,c}=make(),before=structuredClone({q,c}),next=C.reverse(c,q);assert.deepEqual(next.children.map(C.role),['records','aggregate']);assert.equal(C.aggregate(c).scope.signal,'spans');assert.equal(C.records(c).scope.signal,'spans');assert.deepEqual({q,c},before);assert.deepEqual(C.validate(JSON.parse(JSON.stringify(next)),q),next);const aggregate=C.setAggregate(next,{...C.aggregate(next),display:{...C.aggregate(next).display,value:'duration.p95',scale:'max'}},q);assert.equal(C.aggregate(aggregate).display.value,'duration.p95');assert.equal(C.aggregate(c).display.value,'count');assert.deepEqual(C.records(aggregate),C.records(c));});
test('local visibility retains hidden intent but refuses an invisible composition',()=>{const {q,c}=make(),hidden=C.setHidden(c,'records',true,q);assert.equal(C.visible(hidden,'records'),false);assert.equal(C.visible(hidden,'aggregate'),true);assert.deepEqual(C.records(hidden),C.records(c));assert.throws(()=>C.setHidden(hidden,'aggregate',true,q));assert.throws(()=>C.setHidden(c,'missing',true,q));assert.equal(C.visible(c,'missing'),false);});
test('role, scope, version and option ownership are closed before saving',()=>{const {q,c}=make();for(const mutate of [c=>c.version='other',c=>c.kind.name='grid',c=>c.children.push(c.children[0]),c=>c.children[1]={aggregate:C.aggregate(c)},c=>c.children[0].records=null,c=>c.children[0].query=q,c=>c.children[1].records.scope.signal='logs',c=>c.children[1].records.kind.version=2,c=>c.children[1].records.display.requiresComplete=true,c=>c.children[1].records.display.columns=['summary'],c=>c.children[1].records.display.fieldColumns=[{kind:'column',key:'body'}],c=>c.children[1].records.display.fieldColumns=[{kind:'attribute',key:'\ud800'}],c=>c.children[0].hidden=null]){const bad=structuredClone(c);mutate(bad);assert.throws(()=>C.validate(bad,q));}});
test('composition metadata does not admit flat or captured options as an additional authority',()=>{const {q,c}=make(),base={version:1,title:'Services',queries:[q],composition:c};assert.deepEqual(C.view(base),base);for(const [key,value]of [['layout',''],['columns',[]],['fieldColumns',null],['displays',{}],['breakdowns',{}],['tail',null]])assert.throws(()=>C.view({...base,[key]:value}));assert.throws(()=>C.view({...base,composition:null}));assert.throws(()=>C.view({...base,queries:[q,{...q,signal:'logs'}]}));assert.deepEqual(C.declarations(base),{spans:C.aggregate(c)});assert.equal(C.declaration(base,'spans'),C.aggregate(c));assert.equal(C.declaration(base,'logs'),undefined);});
test('record options are checked against the parent signal and edited without touching aggregate intent',()=>{const {q,c}=make(),next=C.setRecords(c,{columns:['utc','service','summary'],fieldColumns:[{kind:'resource',key:'service.namespace'}]},q);assert.deepEqual(C.aggregate(next),C.aggregate(c));assert.deepEqual(C.records(c).display.columns,['utc','summary']);assert.deepEqual(C.records(next).display.columns,['utc','service','summary']);assert.throws(()=>C.setRecords(c,{columns:['utc','summary'],fieldColumns:Array.from({length:9},(_,i)=>({kind:'attribute',key:'field'+i}))},q));});
test('HTTP template handoff admits composition before navigation and keeps flat options absent',()=>{const {q,c}=make(),view={version:1,title:'HTTP server routes',queries:[q],composition:c},response={template:{id:'otel.http-server-routes',version:1,title:'HTTP server routes'},view};const staged=H.response(response,q);assert.deepEqual(staged,view);staged.composition.children[0].aggregate.display.title='Changed';assert.equal(C.aggregate(view.composition).display.title,undefined);assert.throws(()=>H.response({...response,view:{...view,layout:'records'}},q));assert.throws(()=>H.response({...response,view:{...view,rows:[]}},q));});

test('record list is a pinned alternate kind with the same page options and template role',()=>{const {q,c}=make(),before=structuredClone(c),list=C.setRecordKind(c,'record-list',q);assert.equal(C.records(list).kind.name,'record-list');assert.deepEqual(C.records(list).display,C.records(c).display);assert.deepEqual(C.aggregate(list),C.aggregate(c));assert.deepEqual(c,before);assert.equal(C.records(C.setRecords(list,{columns:['utc','summary'],fieldColumns:[]},q)).kind.name,'record-list');assert.equal(C.records(C.reuse(list,q,{...q,environment:'other',fromMs:3000,toMs:4000})).kind.name,'record-list');assert.deepEqual(C.setRecordKind(list,'record-table',q),c);assert.throws(()=>C.setRecordKind(list,'record-detail',q));const source={version:1,title:'Record list',queries:[q],composition:list,template:H.templateFor()};assert.doesNotThrow(()=>H.templateView(source));const dest={...q,groupBy:[{kind:'attribute',key:'other'}]};assert.doesNotThrow(()=>H.authoredRequest(source,dest));assert.equal(C.records(source.composition).kind.name,'record-list');});


test('span waterfall is a spans-only record kind with unchanged options and template ownership',()=>{
 const {q,c}=make(),before=structuredClone({q,c}),fields=Array.from({length:8},(_,i)=>({kind:i%2?'attribute':'resource',key:'field'+i}));
 const next=C.setRecords(C.setRecordKind(c,'record-span-waterfall',q),{columns:['utc','service','summary','context'],fieldColumns:fields},q);
 assert.equal(C.records(next).kind.name,'record-span-waterfall');assert.deepEqual(C.records(next).display.fieldColumns,fields);assert.deepEqual(C.aggregate(next),C.aggregate(c));assert.deepEqual({q,c},before);
 const source={version:1,title:'Supplied span timing',queries:[q],composition:next,template:H.templateFor()};
 const destination={...q,environment:'other',fromMs:3000,toMs:4000,groupBy:[{kind:'attribute',key:'db.namespace'}]};
 assert.doesNotThrow(()=>H.templateView(source));assert.deepEqual(H.authoredRequest(source,destination),{source,query:destination});
 assert.equal(C.records(C.reuse(next,q,{...q,fromMs:3000,toMs:4000})).kind.name,'record-span-waterfall');
 assert.equal(C.records(C.setHidden(next,'records',true,q)).kind.name,'record-span-waterfall');
 const logs={...q,signal:'logs'};delete logs.summary;const logComposition=C.create(logs,D.create(logs),{columns:['utc','summary']});
 assert.throws(()=>C.setRecordKind(logComposition,'record-span-waterfall',logs),/requires a spans query/);
 const hidden=structuredClone(logComposition);hidden.children[1].records.kind.name='record-span-waterfall';hidden.children[1].hidden=true;assert.throws(()=>C.validate(hidden,logs),/requires a spans query/);
});

test('records-only saved views retain raw scope and reuse presentation without grouping',()=>{
 for(const signal of ['logs','spans'])for(const groupBy of [undefined,[],[{kind:'column',key:'service'}],[{kind:'column',key:'service'},{kind:'attribute',key:'team'}]]){
  const q={signal,environment:'prod',fromMs:1000,toMs:2000,limit:10,offset:0,...(groupBy===undefined?{}:{groupBy}),where:{field:{kind:'column',key:'trace_id'},op:'eq',value:'trace'}};
  const before=structuredClone(q),c=C.createRecords(q,{columns:['utc','summary'],fieldColumns:[{kind:'resource',key:'service.namespace'}]});
  assert.equal(C.recordOnly(c),true);assert.equal(C.aggregate(c),undefined);assert.deepEqual(C.declarations({composition:c}),{});
  const kind=signal==='spans'?'record-span-waterfall':'record-list',next=C.setRecordKind(c,kind,q),destination={...q,environment:'staging',fromMs:3000,toMs:4000,offset:10};
  const reused=C.reuse(next,q,destination);assert.equal(C.records(reused).kind.name,kind);assert.deepEqual(C.records(reused).display,C.records(c).display);assert.deepEqual(q,before);assert.deepEqual(destination,{...before,environment:'staging',fromMs:3000,toMs:4000,offset:10});
  const view={version:1,title:'Raw records',queries:[q],composition:next};assert.deepEqual(C.view(JSON.parse(JSON.stringify(view))),view);assert.throws(()=>H.templateView({...view,template:H.templateFor()}));
 }
});
test('records-only admission refuses absent records, hidden sole records and aggregate-less siblings',()=>{
 const q={signal:'logs',environment:'prod',fromMs:1000,toMs:2000,limit:10,offset:0},c=C.createRecords(q);
 assert.throws(()=>C.setHidden(c,'records',true,q));assert.throws(()=>C.validate({...c,children:[]},q));assert.throws(()=>C.setRecordKind(c,'record-span-waterfall',q));
 for(const directive of ['numeric','numericHistogram','reportedEdge','groupTimeSeries'])assert.throws(()=>C.createRecords({...q,[directive]:true}));
 assert.throws(()=>C.createRecords({...q,where:{field:{kind:'column',key:'not_a_column'},op:'eq',value:'x'}}));
 assert.throws(()=>C.createRecords({...q,signal:'metrics'}));
 const grouped=query(),full=C.create(grouped,D.create(grouped));const record=C.createRecords(grouped);
 const timeline={version:'goo.view/1',scope:{kind:'record-query',signal:'spans'},kind:{name:'time-series',version:1},display:{value:'count'}};
 assert.throws(()=>C.validate({...record,children:[...record.children,{timeline}]},grouped));assert.equal(C.recordOnly(full),false);
});
test('records-only preserves unused ordinary directives without permitting specialized profiles',()=>{
 const q={...query(),groupBy:[{kind:'column',key:'service'},{kind:'attribute',key:'team'}],bucketMs:1000,spanStatus:true,numeric:null,numericHistogram:null,reportedEdge:null,groupTimeSeries:false};
 const before=structuredClone(q),c=C.createRecords(q);assert.deepEqual(C.view({version:1,title:'Exact records',queries:[q],composition:c}).queries,[before]);assert.deepEqual(q,before);
 const timeline=require('../ui/timeline-declaration.mjs').create({...q,groupBy:[{kind:'column',key:'service'}]});
 assert.throws(()=>C.validate({...c,children:[...c.children,{timeline}]},{...q,groupBy:[{kind:'column',key:'service'}]}),/Records-only/);
});
