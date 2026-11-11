const {test}=require('node:test'),assert=require('node:assert/strict'),T=require('../ui/view-template-handoff.mjs'),C=require('../ui/view-composition.mjs'),V=require('../ui/view-declaration.mjs');
const field={kind:'attribute',key:'db.operation.name'},other={kind:'resource',key:'db.namespace'};
function query(service='source',group=field){return {signal:'spans',environment:'prod',fromMs:1000,toMs:61000,limit:5,offset:0,summary:'span-duration',spanStatus:true,groupBy:[group],where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:service},{field:{kind:'column',key:'kind'},op:'eq',value:'Client'},{field:{kind:'attribute',key:'db.system.name'},op:'ne',value:''},{any:[{field:{kind:'attribute',key:'zone'},op:'missing'},{field:{kind:'attribute',key:'zone'},op:'eq',value:''}]}]}};}
function source(){const q=query();return {version:1,id:'a'.repeat(64),title:'Team database view',queries:[q],composition:C.create(q,V.create(q)),template:T.templateFor('otel.db-client')};}
function envelope(s,q){return {template:{id:s.id,version:1,title:s.title},view:{version:1,title:s.title,queries:[{...structuredClone(q),offset:0}],composition:structuredClone(s.composition),template:structuredClone(s.template)}};}
test('authored template explicitly binds destination category without copying source query or data',()=>{const s=source(),q={...query('destination',other),fromMs:60000,toMs:120000,offset:0,environment:'other'},before=structuredClone({s,q}),input=T.authoredRequest(s,q),view=T.authoredResponse(envelope(s,q),s,q);assert.deepEqual(input,{source:s,query:q});assert.deepEqual(view.queries,[{...q,offset:0}]);assert.deepEqual(view.template,s.template);assert.deepEqual(view.composition,s.composition);view.queries[0].where.all[0].value='mutated';input.source.composition.children.reverse();assert.deepEqual({s,q},before);assert.throws(()=>C.reuse(s.composition,s.queries[0],q),'ordinary composition still requires exact grouping');});
test('template profiles require mandatory current semantic selectors rather than labels or weakening OR',()=>{const s=source();for(const change of [q=>q.where={any:q.where.all.slice(0,3)},q=>q.where.all[1].value='Server',q=>q.where.all[2].field.key='db.system',q=>q.where.all[0].field.kind='resource']){const bad=structuredClone(s);change(bad.queries[0]);assert.throws(()=>T.templateView(bad));}const ordinary={...s,template:T.templateFor('ordinary')};assert.doesNotThrow(()=>T.templateView(ordinary));assert.throws(()=>T.binding({...ordinary.template,rows:[]}));});
test('template enabling refuses nonportable children, local Find and runtime payloads without stripping them',()=>{for(const change of [s=>s.composition.children[0].aggregate.display.query='SELECT',s=>s.composition.children.push({timeline:{}}),s=>s.rows=[],s=>s.template.profile='anything',s=>s.template.group='source',s=>s.version=2]){const s=source(),before=structuredClone(s);change(s);const input=structuredClone(s);assert.throws(()=>T.templateView(s));assert.deepEqual(s,input);assert.equal(before.queries[0].where.all[0].value,'source');}});
test('response rejects source identity, query, binding and presentation substitution; missing duration refuses',()=>{const s=source(),q=query('target',other);for(const change of [r=>r.template.id='b'.repeat(64),r=>r.view.queries[0].environment='wrong',r=>r.view.template.profile='ordinary',r=>r.view.composition.children.reverse(),r=>r.view.rows=[]]){const r=envelope(s,q);change(r);assert.throws(()=>T.authoredResponse(r,s,q));}s.composition.children[0].aggregate.display.value='duration.p95';s.composition.children[0].aggregate.display.scale='max';const absent=query('target',other);delete absent.summary;assert.throws(()=>T.authoredRequest(s,absent));});
test('authored expansion declines late replies and performs no source or publication callbacks',async()=>{const s=source(),q=query('target',other);let active=true,release,calls=0;const gate=new Promise(done=>release=done),pending=T.expandAuthored(s,q,{current:()=>active,post:async input=>{calls++;assert.deepEqual(input.query,q);await gate;return envelope(s,q)}});active=false;release();assert.equal(await pending,null);assert.equal(calls,1);assert.equal(await T.expandAuthored(s,q,{current:()=>false,post:()=>{throw Error('must not call')}}),null);});
test('database service handoff chooses the shared pinned DB descriptor and retains exact widget scope',async()=>{const q=query(),widget={family:'db-client',key:'dimension',state:'supported',query:q},request=T.request(widget);assert.equal(request.template,'otel.db-client-operations');assert.equal(T.customizeLabel(widget),'database client view');const s=source(),result={template:{id:'otel.db-client-operations',version:1,title:'Database client activity'},view:{...envelope(s,q).view,title:'Database client activity'}};assert.deepEqual(await T.expand(widget,{post:async()=>result}),result.view);assert.deepEqual(widget.query,q);});
test('question codec accepts Go zero/empty omissions while retaining exact predicates and qualified grouping',()=>{const s=source(),q={...query('target',other),spanStatus:false,groupTimeSeries:false};const wire=envelope(s,q),encode=p=>{if(p.value==='')delete p.value;for(const child of [...(p.all||[]),...(p.any||[])])encode(child)};delete wire.view.queries[0].spanStatus;delete wire.view.queries[0].groupTimeSeries;encode(wire.view.queries[0].where);const before=structuredClone({s,q,wire});assert.deepEqual(T.authoredResponse(wire,s,q).queries,wire.view.queries);assert.deepEqual({s,q,wire},before);for(const change of [p=>p.where.all[3].any[1].op='missing',p=>p.where.all[3].any[1].op='ne',p=>p.groupBy[0].kind='attribute',p=>p.where.all[3].any[1].field.kind='resource',p=>p.where.all[0].value='different-service',p=>p.where.all[3].any.pop()]){const bad=structuredClone(wire);change(bad.view.queries[0]);assert.throws(()=>T.authoredResponse(bad,s,q));}});

test('time-aware templates retain supported timeline presentation and exact destination query',()=>{
 const Timeline=require('../ui/timeline-declaration.mjs');
 for(const value of ['count','duration.mean','duration.p95'])for(const kind of ['time-bar','table'])for(const hidden of [false,true]){
  const s=source();s.queries[0].bucketMs=10000;
  const timeline={...Timeline.create(s.queries[0]),kind:{name:kind,version:1},display:{value,title:'UTC workload timing',requiresComplete:true}};
  s.composition.children=[s.composition.children[1],{timeline,hidden},s.composition.children[0]];
  const q={...query('target',other),environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'unreported'}},fromMs:65000,toMs:125001,bucketMs:10000,offset:0};delete q.environment;
  const before=structuredClone({s,q});assert.deepEqual(T.templateView(s),s);assert.deepEqual(T.authoredRequest(s,q),{source:s,query:q});
  const result=T.authoredResponse(envelope(s,q),s,q);assert.deepEqual(result.composition,s.composition);assert.deepEqual(result.queries,[{...q,offset:0}]);assert.deepEqual({s,q},before);
 }
});

test('timeline reuse refuses missing or different intervals before any metadata request, including hidden timelines',async()=>{
 const Timeline=require('../ui/timeline-declaration.mjs'),s=source();s.queries[0].bucketMs=10000;s.composition.children.unshift({timeline:Timeline.create(s.queries[0]),hidden:true});
 for(const bucketMs of [undefined,1000,60000]){
  const q={...query('target',other),bucketMs},before=structuredClone({s,q});let calls=0;
  assert.throws(()=>T.authoredRequest(s,q),/requires 10 seconds intervals.*configure.*Run explicitly/);
  assert.throws(()=>T.authoredResponse(envelope(s,q),s,q),/requires 10 seconds intervals/);
  await assert.rejects(T.expandAuthored(s,q,{post:async()=>{calls++;throw Error('metadata request must not run')}}),/requires 10 seconds intervals/);
  assert.equal(calls,0);assert.deepEqual({s,q},before);
 }
 const tooWide={...query('target',other),bucketMs:10000,fromMs:0,toMs:2400001};assert.throws(()=>T.authoredRequest(s,tooWide),/at most 240/);
 const duration=structuredClone(s);C.timeline(duration.composition).display.value='duration.p95';const unsupported={...query('target',other),bucketMs:10000};delete unsupported.summary;assert.throws(()=>T.authoredRequest(duration,unsupported),/query-supported timeline/);
});

test('time-aware replies cannot change interval or timeline presentation and remain epoch guarded',async()=>{
 const Timeline=require('../ui/timeline-declaration.mjs'),s=source();s.queries[0].bucketMs=10000;s.composition.children.unshift({timeline:Timeline.create(s.queries[0]),hidden:true});const q={...query('target',other),bucketMs:10000};
 for(const change of [r=>r.view.queries[0].bucketMs=1000,r=>C.timeline(r.view.composition).display.value='duration.p95',r=>C.timeline(r.view.composition).display.title='Changed',r=>C.timeline(r.view.composition).display.requiresComplete=true,r=>r.view.composition.children[0].hidden=false,r=>r.view.composition.children.reverse()]){
  const response=envelope(s,q);change(response);assert.throws(()=>T.authoredResponse(response,s,q));
 }
 let active=true,release,calls=0;const wait=new Promise(resolve=>release=resolve),pending=T.expandAuthored(s,q,{current:()=>active,post:async()=>{calls++;await wait;return envelope(s,q)}});active=false;release();assert.equal(await pending,null);assert.equal(calls,1);
 assert.equal(T.intervalLabel(1000),'1 second');assert.equal(T.intervalLabel(10000),'10 seconds');assert.equal(T.intervalLabel(60000),'1 minute');assert.equal(T.intervalLabel(300000),'5 minutes');assert.equal(T.intervalLabel(undefined),'Not configured');
});

test('ordinary log templates retain count timelines without acquiring span duration semantics',()=>{
 const Timeline=require('../ui/timeline-declaration.mjs'),q={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:5,offset:0,bucketMs:10000,groupBy:[{kind:'column',key:'service'}]},s={version:1,id:'c'.repeat(64),title:'Log workload',queries:[q],composition:C.create(q,V.create(q)),template:T.templateFor('ordinary')};
 s.composition=C.setTimeline(s.composition,Timeline.create(q),q);const target={...q,environment:'other',groupBy:[{kind:'attribute',key:'queue'}],fromMs:2000,toMs:62001};assert.deepEqual(T.authoredRequest(s,target),{source:s,query:target});assert.deepEqual(T.authoredResponse(envelope(s,target),s,target).composition,s.composition);
 C.timeline(s.composition).display.value='duration.mean';assert.throws(()=>T.templateView(s),/query-supported timeline/);
});

function comparisonSource(){const s=source();delete s.queries[0].spanStatus;s.composition=C.setComparison(s.composition,require('../ui/captured-count-comparison.mjs').create(s.queries[0],'b'.repeat(64)),s.queries[0]);s.template=T.templateFor('otel.db-client',{baseline:true});return s;}
function comparisonTarget(){const q=query('destination',other);delete q.spanStatus;return {...q,environment:'target-environment',fromMs:65000,toMs:125001};}
function comparisonEnvelope(s,q,id){const value=envelope(s,q);C.comparison(value.view.composition).baselineCapture=id;return value;}
test('comparison templates bind only an explicit destination baseline and preserve every other declaration',()=>{
 const Timeline=require('../ui/timeline-declaration.mjs'),id='d'.repeat(64);
 for(const view of ['summary','groups'])for(const order of ['change','label'])for(const hidden of [false,true]){
  const s=comparisonSource(),q=comparisonTarget();s.queries[0].bucketMs=q.bucketMs=10000;
  const comparison=s.composition.children.find(child=>child.comparison);comparison.hidden=hidden;comparison.comparison.display={value:'count',title:'Population change',view,order};
  s.composition=C.setTimeline(s.composition,{...Timeline.create(s.queries[0]),display:{value:'duration.p95',title:'Timing',requiresComplete:true}},s.queries[0]);s.composition.children.reverse();
  const before=structuredClone({s,q}),input=T.authoredRequest(s,q,id),result=T.authoredResponse(comparisonEnvelope(s,q,id),s,q,id),expected=structuredClone(s.composition);C.comparison(expected).baselineCapture=id;
  assert.deepEqual(input,{source:s,query:q,baselineCapture:id});assert.deepEqual(result.composition,expected);assert.deepEqual(result.queries,[q]);assert.deepEqual(result.template,s.template);assert.deepEqual({s,q},before);
  C.comparison(result.composition).display.title='Mutation';assert.deepEqual({s,q},before);
 }
 // An identical reference is permitted only when chosen explicitly again.
 const s=comparisonSource(),q=comparisonTarget(),explicit=C.comparison(s.composition).baselineCapture;
 assert.equal(T.authoredRequest(s,q,explicit).baselineCapture,explicit);assert.throws(()=>T.authoredRequest(s,q),/explicit destination baseline/);
});
test('baseline parameters and bindings are conditional, closed and never fall back to the specimen',async()=>{
 const s=comparisonSource(),q=comparisonTarget(),before=structuredClone({s,q});
 for(const id of [undefined,null,'','z'.repeat(64),'A'.repeat(64),'b'.repeat(63),{}]){
  let calls=0;assert.throws(()=>T.authoredRequest(s,q,id),/explicit destination baseline/);
  await assert.rejects(T.expandAuthored(s,q,{baselineCapture:id,post:async()=>{calls++;throw Error('must not call')}}),/explicit destination baseline/);assert.equal(calls,0);
 }
 assert.deepEqual({s,q},before);
 for(const bad of [{...s.template,baseline:'source'},{...s.template,baseline:null},{...s.template,extra:true}])assert.throws(()=>T.binding(bad));
 const missing=structuredClone(s);delete missing.template.baseline;assert.throws(()=>T.templateView(missing),/baseline binding/);
 const ordinary=source();assert.throws(()=>T.authoredRequest(ordinary,query('destination',other),'d'.repeat(64)),/only allowed for comparison/);ordinary.template.baseline='destination';assert.throws(()=>T.templateView(ordinary),/only when a comparison/);
 const unsupported={...q,spanStatus:true};assert.throws(()=>T.authoredRequest(s,unsupported,'d'.repeat(64)),/Captured count comparison/);
});
test('comparison replies permit exactly the chosen reference substitution and fence late baseline requests',async()=>{
 const s=comparisonSource(),q=comparisonTarget(),id='d'.repeat(64);
 for(const change of [r=>C.comparison(r.view.composition).baselineCapture='b'.repeat(64),r=>C.comparison(r.view.composition).display.view='groups',r=>C.comparison(r.view.composition).display.title='Changed',r=>r.view.composition.children.at(-1).hidden=true,r=>r.view.composition.children.reverse(),r=>r.view.template.baseline='source',r=>r.view.queries[0].environment='source']){
  const response=comparisonEnvelope(s,q,id);change(response);assert.throws(()=>T.authoredResponse(response,s,q,id));
 }
 let baseline=id,release;const wait=new Promise(resolve=>release=resolve),pending=T.expandAuthored(s,q,{baselineCapture:id,current:()=>baseline===id,post:async input=>{assert.equal(input.baselineCapture,id);await wait;return comparisonEnvelope(s,q,id)}});baseline='e'.repeat(64);release();assert.equal(await pending,null);
 assert.throws(()=>T.authoredResponse(comparisonEnvelope(s,q,id),s,q,'e'.repeat(64)),/changed the selected presentation/);
});
test('explicit comparison edits derive only the destination binding and preserve an enabled timeline profile',()=>{
 const Timeline=require('../ui/timeline-declaration.mjs'),s=source();delete s.queries[0].spanStatus;s.queries[0].bucketMs=10000;s.composition=C.setTimeline(s.composition,Timeline.create(s.queries[0]),s.queries[0]);
 const original=structuredClone(s),comparison=require('../ui/captured-count-comparison.mjs').create(s.queries[0],'b'.repeat(64));
 const withComparison=C.setComparison(s.composition,comparison,s.queries[0]),bound=T.bindingForComposition(s.template,withComparison);assert.deepEqual(bound,{...s.template,baseline:'destination'});assert.doesNotThrow(()=>T.templateView({...s,composition:withComparison,template:bound}));
 const removed=C.setComparison(withComparison,null,s.queries[0]),ordinary=T.bindingForComposition(bound,removed);assert.deepEqual(ordinary,s.template);assert.doesNotThrow(()=>T.templateView({...s,composition:removed,template:ordinary}));assert.deepEqual(s,original);
});

test('category templates retain literal facet identity across existing ordinary and OTel bindings',()=>{
 const Facet=require('../ui/facet-declaration.mjs');
 for(const profile of ['ordinary','otel.db-client','otel.http-server'])for(const hidden of [false,true]){
  const s=source();s.template=T.templateFor(profile);if(profile==='otel.http-server'){s.queries[0].groupBy=[{kind:'attribute',key:'http.route'}];s.queries[0].where.all[1].value='Server';s.queries[0].where.all[2].field.key='http.request.method';}
  const literal={kind:'attribute',key:profile==='otel.http-server'?'http.request.method':'team.routing-policy'},facet=Facet.create(s.queries[0],literal,'Team facet');s.composition=C.setFacet(s.composition,facet,s.queries[0]);s.composition=C.setHidden(s.composition,'facet',hidden,s.queries[0]);
  const q={...structuredClone(s.queries[0]),environment:'',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'other'}},fromMs:61000,toMs:121000,groupBy:[{kind:'resource',key:'service.namespace'}]};q.where.all[0].value='target';const before=structuredClone({s,q});
  assert.deepEqual(T.templateView(s),s);assert.deepEqual(T.authoredRequest(s,q),{source:s,query:q});const result=T.authoredResponse(envelope(s,q),s,q);assert.deepEqual(C.facet(result.composition),facet);assert.equal(result.composition.children.find(c=>c.facet).hidden,hidden);assert.deepEqual(result.template,s.template);assert.deepEqual(result.queries,[q]);assert.deepEqual({s,q},before);
  const binding=T.bindingForComposition(s.template,s.composition);assert.deepEqual(binding,s.template);
 }
});
test('facet template expansion refuses literal presentation substitution and unsupported destination profiles',()=>{
 const Facet=require('../ui/facet-declaration.mjs'),s=source();s.composition=C.setFacet(s.composition,Facet.create(s.queries[0],{kind:'attribute',key:'http.request.method'},'Literal method'),s.queries[0]);const q=query('target',other);
 for(const change of [f=>f.field.key='http.method',f=>f.field.kind='resource',f=>f.display.title='Other title',f=>f.display.basis='page']){const response=envelope(s,q);change(C.facet(response.view.composition));assert.throws(()=>T.authoredResponse(response,s,q));}
 const numeric={field:{kind:'attribute',key:'weight'},type:'int64',unit:'1',parse:'canonical-decimal'};assert.throws(()=>T.authoredRequest(s,{...q,numeric}));assert.throws(()=>T.authoredRequest(s,{...q,reportedEdge:{}}));assert.throws(()=>T.authoredRequest(s,{...q,groupTimeSeries:true}));assert.throws(()=>T.authoredRequest(s,{...q,limit:101}));
 const captured=require('../ui/captured-count-comparison.mjs');const noStatus={...s.queries[0]};delete noStatus.spanStatus;assert.throws(()=>C.setComparison(s.composition,captured.create(noStatus,'b'.repeat(64)),noStatus));
 const bad=structuredClone(s);bad.queries[0].where={any:bad.queries[0].where.all};assert.throws(()=>T.templateView(bad));
});
test('HTTP service handoff admits the shipped literal method facet without changing the widget question',()=>{
 const Facet=require('../ui/facet-declaration.mjs'),q=query('http-service',{kind:'attribute',key:'http.route'});q.where.all[1].value='Server';q.where.all[2].field.key='http.request.method';
 const view={version:1,title:'HTTP server routes',queries:[q],composition:C.setFacet(C.create(q,V.create(q)),Facet.create(q,{kind:'attribute',key:'http.request.method'},'HTTP methods'),q),template:T.templateFor('otel.http-server')},widget={family:'http-server',key:'dimension',state:'supported',query:q},response={template:{id:'otel.http-server-routes',version:1,title:view.title},view},before=structuredClone({q,widget,response});
 assert.deepEqual(T.response(response,q),view);assert.deepEqual(T.request(widget).widget.query,q);assert.deepEqual({q,widget,response},before);
 const selected=require('../ui/population-workspace.mjs').selectionQueries([q],[{signal:'spans',field:{kind:'attribute',key:'http.request.method'},value:'GET',mode:'include'}])[0];assert.doesNotThrow(()=>T.templateView({...view,queries:[selected]}));assert.deepEqual(C.facet(T.templateView({...view,queries:[selected]}).composition),C.facet(view.composition));
 const numeric={...q,numeric:{field:{kind:'attribute',key:'payload'},type:'int64',unit:'1',parse:'canonical-decimal'}},numericView={version:1,title:view.title,queries:[numeric],composition:C.create(numeric,V.create(numeric))};assert.deepEqual(T.response({template:response.template,view:numericView},numeric).queries,[numeric]);assert.equal(C.facet(numericView.composition),undefined);
});
