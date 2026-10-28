const {test}=require('node:test'),assert=require('node:assert/strict');const {question}=require('../ui/inventory.mjs');
test('inventory navigation retains fixed scope and exact literal identities without claiming a shared evidence read',()=>{const source={question:{environment:'production',fromMs:1000,toMs:2000},basis:'capture'};const row={service:"api'\\\n",name:'<work>'};const q=question(source,row);assert.deepEqual(q,{signal:'spans',environment:'production',fromMs:1000,toMs:2000,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:row.service},{field:{kind:'column',key:'name'},op:'eq',value:row.name}]},groupBy:[],limit:100,offset:0});assert.equal(q.basis,undefined);assert.equal(source.question.signal,undefined);});

test('metric discovery carries the same fixed inventory window without claiming a span population',()=>{const {metricContext}=require('../ui/inventory.mjs');const context=metricContext({question:{environment:'prod',fromMs:1234,toMs:5678}}, {service:'worker',name:'job'});assert.deepEqual(context,{service:'worker',environment:'prod',fromMs:1234,toMs:5678});assert.equal(context.where,undefined)});

const {scopeQuestion,suggestions}=require('../ui/service-browser.mjs');
test('service suggestions choose observed exact convention aliases and narrow only matching spans',()=>{
 const scope={environment:'prod',fromMs:1000,toMs:2000}, service="worker'\n";
 const fields=[{kind:'attribute',key:'http.route'},{kind:'attribute',key:'db.system'},{kind:'resource',key:'messaging.system'}];
 const choices=suggestions(scope,service,fields);
 assert.deepEqual(choices.map(s=>s.available),[true,true,false]);
 for(const s of choices.filter(s=>s.available)){
  assert.deepEqual(s.question.where.all[0],scopeQuestion(scope,service).where);
  assert.equal(s.question.where.all[1].op,'exists');
  assert.deepEqual(s.question.groupBy,[s.field]);assert.equal(s.question.fromMs,1000);assert.equal(s.question.toMs,2000);
 }
 assert.equal(choices[1].field.key,'db.system');assert.equal(choices[2].question,null);
 assert.equal(fields.length,3);assert.equal(scope.signal,undefined);
 const newer=suggestions(scope,service,[{kind:'attribute',key:'db.system'},{kind:'attribute',key:'db.system.name'}]);
 assert.equal(newer[1].field.key,'db.system.name');
});

test('catalog retains separate exact signal counts, groups services and leaves source evidence detached',()=>{
 const {catalogRows}=require('../ui/inventory.mjs');const source=[{service:'worker',signal:'metrics',count:'18446744073709551615',firstMs:10,lastMs:20},{service:'api',signal:'spans',count:'3',firstMs:11,lastMs:19},{service:'worker',signal:'logs',count:'2',firstMs:5,lastMs:22}];const rows=catalogRows(source);
 assert.equal(rows.length,2);assert.equal(rows[1].signals.metrics.count,'18446744073709551615');assert.equal(rows[1].signals.spans,undefined);assert.equal(rows[1].firstMs,5);assert.equal(rows[1].lastMs,22);rows[1].signals.metrics.count='0';assert.equal(source[0].count,'18446744073709551615');assert.throws(()=>catalogRows([source[0],source[0]]),/Duplicate/);
});
test('distributed relationships keep unresolved diagnostics out of edges and stage explicitly wider scoped questions',()=>{
 const {model,widerQuestion}=require('../ui/distributed-services.mjs'),evidence={dependencies:[{kind:'unresolved',child:'worker',parent:'',count:'9007199254740993'},{kind:'edge',child:'worker',parent:'api',count:'4'},{kind:'ambiguous',parent:'',child:'other',count:'2'}]};const before=structuredClone(evidence),m=model(evidence);assert.equal(m.edges.length,1);assert.equal(m.diagnostics.unresolved,'9007199254740993');assert.deepEqual(widerQuestion({environment:'prod',fromMs:10,toMs:20},m.edges[0]),{signal:'spans',environment:'prod',fromMs:10,toMs:20,limit:100,offset:0,where:{field:{kind:'column',key:'service'},op:'eq',value:'worker'}});assert.deepEqual(evidence,before);
});
test('dashboard result matching rejects same-scope wrong populations and detached capture intent',()=>{
 const {matches,question}=require('../ui/service-dashboard.mjs'),q={signal:'logs',environment:'prod',fromMs:10,toMs:20,limit:100,offset:0,where:{field:{kind:'column',key:'service'},op:'eq',value:'api'}},w={query:q};assert.equal(matches(w,{question:structuredClone(q)}),true);assert.equal(matches(w,{question:{...q,where:{...q.where,value:'other'}}}),false);const capture=question(w);capture.explore.where.value='changed';assert.equal(q.where.value,'api');
});
test('native dashboard summaries preserve latest unknown endpoints, exact stream identity and histogram ranges',()=>{
 const {nativeSummary}=require('../ui/service-dashboard.mjs'),stream={resource:{'service.instance.id':'one'},attributes:{region:'a'},unit:'1'};const data={gauges:{series:[{stream,state:'unknown',samples:[{atNS:'1000',state:'observed',value:.2},{atNS:'2000',state:'unknown',reason:'no recorded value'}]}]}};assert.equal(nativeSummary(data)[0].value,'Unavailable');assert.equal(nativeSummary(data)[0].extent,'Endpoint 2000 ns');data.gauges.series[0].state='observed';data.gauges.series[0].samples[1]={atNS:'2000',state:'observed',value:.25};assert.equal(nativeSummary(data)[0].value,'0.25 1');assert.deepEqual(nativeSummary(data)[0].stream,stream);const hist={histograms:{series:[{stream:{...stream,unit:'ms'},state:'ok',intervals:[{fromNS:'1',toNS:'10',state:'ok',distribution:{count:'4',unit:'ms',p95:{state:'unbounded-upper',lower:10}}}]}]}};assert.equal(nativeSummary(hist)[0].value,'4 observations · P95 bucket range: (10, +∞) ms · no interpolation');
});
test('service namespace selectors partition missing, present-empty and literal names throughout questions',()=>{
 const id=require('../ui/service-identity.mjs'),{scopeQuestion}=require('../ui/service-browser.mjs'),{sameScope}=require('../ui/service-dashboard.mjs');const scope={environment:'prod',fromMs:1000,toMs:2000};const missing=id.namespace(null),empty=id.namespace(''),named=id.namespace('payments');assert.notEqual(id.key('api',missing),id.key('api',empty));assert.notEqual(id.key('api',named),id.key('api',null));for(const namespace of [missing,empty,named]){const q=scopeQuestion({...scope,namespace},'api');assert.equal(q.namespace,undefined);assert.deepEqual(q.where.all[1],id.predicate(namespace));assert.equal(sameScope({...scope,namespace},{...scope,namespace:structuredClone(namespace)}),true)}assert.equal(sameScope({...scope,namespace:empty},{...scope,namespace:missing}),false);
 const rows=require('../ui/inventory.mjs').catalogRows([{service:'api',signal:'logs',count:'1',firstMs:1,lastMs:2},{service:'api',namespace:'',signal:'logs',count:'2',firstMs:1,lastMs:2},{service:'api',namespace:'payments',signal:'logs',count:'3',firstMs:1,lastMs:2}]);assert.equal(rows.length,3);
});
test('declared services retain absent telemetry, exact namespaces, multiple bindings and independent provenance',()=>{
 const {catalogModel}=require('../ui/inventory.mjs');const evidence={question:{environment:'prod',fromMs:1,toMs:2},signals:[{service:'api',namespace:'payments',signal:'spans',count:'7',firstMs:1,lastMs:1}]};
 const pkg=(revision,namespace,instance='one')=>({revision,definition:{name:revision,title:revision,resource:{service:'api',environment:'prod',instance,...(namespace?{namespace}:{})},instrumentation:{name:'sdk',version:'1'}}});
 const before=JSON.stringify(evidence),packages=[pkg('unbound'),pkg('missing',{mode:'missing'}),pkg('orders',{mode:'value',value:'orders'}),pkg('other',{mode:'value',value:'orders'},'two')];const rows=catalogModel(evidence,packages);
 assert.equal(rows.length,3);const observed=rows.find(r=>r.observed);assert.equal(observed.signals.spans.count,'7');assert.equal(observed.declarations[0].namespaceMatch,'unspecified');const orders=rows.find(r=>r.namespace?.value==='orders');assert.equal(orders.observed,false);assert.equal(orders.lastMs,null);assert.deepEqual(Object.keys(orders.signals),[]);assert.equal(orders.declarations.length,2);assert.equal(JSON.stringify(evidence),before);
 assert.deepEqual(catalogModel(evidence,[...packages].reverse()),rows);assert.deepEqual(catalogModel({...evidence,question:{...evidence.question,environment:'other'}},packages).flatMap(r=>r.declarations),[]);
 const empty=catalogModel({...evidence,signals:[]},packages);assert.equal(empty.length,3);assert.equal(empty.filter(r=>r.namespace===null).length,1);assert.equal(empty.every(r=>!r.observed),true);
 const scoped=catalogModel({...evidence,question:{...evidence.question,namespace:{mode:'value',value:'payments'}}},packages);assert.equal(scoped.length,1);assert.equal(scoped[0].declarations.length,1);
});
test('reported edge handoffs retain namespaces and captured parent-search scope without source reads',()=>{
 const {edgeQuestion}=require('../ui/distributed-services.mjs'),{recordStep,stepLink,readStepLink}=require('../ui/workspace.mjs');const scope={environment:'prod',fromMs:1000,toMs:61000},edge={parent:'api',parentNamespace:'payments',child:'api',childNamespace:''};const q=edgeQuestion(scope,edge);assert.equal(q.summary,'span-duration');assert.deepEqual(q.reportedEdge.child.namespace,{mode:'value'});assert.equal(q.reportedEdge.parent.namespace.value,'payments');assert.deepEqual(readStepLink(stepLink(recordStep(q))).record,q);
 const visual=require('../ui/comparison-visual.mjs');assert.deepEqual(visual.scale(['900719925474099300','1801439850948198600']).widths,[50,100]);assert.deepEqual(visual.scale([null,'0'],'100').widths,[null,0]);assert.deepEqual(visual.scale(['0.25','0.5'],'1').widths,[25,50]);assert.equal(visual.maximum(['0.01','0.1']),'0.1');
});
test('native cards show evaluated latest intervals without claiming whole-window coverage or hiding a newer unknown',()=>{
 const {nativeSummary}=require('../ui/service-dashboard.mjs'),stream={resource:{'service.instance.id':'one','service.namespace':''},resourceSchemaURL:'resource-v1',attributes:{route:'/search'},unit:'{request}'};
 const data={counters:{series:[{stream,state:'unknown',intervals:[{fromNS:'100',toNS:'200',state:'ok',delta:'10',perSecond:5}]}]}};
 let s=nativeSummary(data)[0];assert.equal(s.value,'5 {request}/s');assert.equal(s.seriesState,'unknown');assert.equal(s.extent,'Latest interval [100, 200) ns');assert.deepEqual(s.stream,stream);s.stream.resource['service.namespace']='changed';assert.equal(stream.resource['service.namespace'],'');
 data.counters.series[0].intervals.push({fromNS:'200',toNS:'300',state:'unknown',reason:'counter-decrease'});s=nativeSummary(data)[0];assert.equal(s.value,'Unavailable');assert.equal(s.reason,'counter-decrease');assert.equal(s.toNS,'300');
 data.counters.series[0].intervals=[{fromNS:'100',toNS:'200',state:'ok',delta:'0',perSecond:0}];assert.equal(nativeSummary(data)[0].value,'0 {request}/s');
 const histogram={histograms:{series:[{stream:{...stream,unit:'s'},state:'unknown',intervals:[{fromNS:'100',toNS:'200',state:'ok',distribution:{unit:'s',count:'0',p95:{state:'no_data'}}}]}]}};assert.equal(nativeSummary(histogram)[0].value,'0 observations · P95: no observations');histogram.histograms.series[0].intervals.push({fromNS:'200',toNS:'300',state:'unknown',reason:'bucket-layout-change'});assert.equal(nativeSummary(histogram)[0].value,'Unavailable');
 const all=Array.from({length:7},(_,i)=>({...data.counters.series[0],stream:{...stream,resource:{'service.instance.id':String(i)}}}));assert.equal(nativeSummary({counters:{series:all}}).length,4);assert.equal(all.length,7);
});
