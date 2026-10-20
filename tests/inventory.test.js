const {test}=require('node:test'),assert=require('node:assert/strict');const {question}=require('../ui/inventory');
test('inventory navigation retains fixed scope and exact literal identities without claiming a shared evidence read',()=>{const source={question:{environment:'production',fromMs:1000,toMs:2000},basis:'capture'};const row={service:"api'\\\n",name:'<work>'};const q=question(source,row);assert.deepEqual(q,{signal:'spans',environment:'production',fromMs:1000,toMs:2000,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:row.service},{field:{kind:'column',key:'name'},op:'eq',value:row.name}]},groupBy:[],limit:100,offset:0});assert.equal(q.basis,undefined);assert.equal(source.question.signal,undefined);});

test('metric discovery carries the same fixed inventory window without claiming a span population',()=>{const {metricContext}=require('../ui/inventory');const context=metricContext({question:{environment:'prod',fromMs:1234,toMs:5678}}, {service:'worker',name:'job'});assert.deepEqual(context,{service:'worker',environment:'prod',fromMs:1234,toMs:5678});assert.equal(context.where,undefined)});

const {scopeQuestion,suggestions}=require('../ui/service-browser');
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
 const {catalogRows}=require('../ui/inventory');const source=[{service:'worker',signal:'metrics',count:'18446744073709551615',firstMs:10,lastMs:20},{service:'api',signal:'spans',count:'3',firstMs:11,lastMs:19},{service:'worker',signal:'logs',count:'2',firstMs:5,lastMs:22}];const rows=catalogRows(source);
 assert.equal(rows.length,2);assert.equal(rows[1].signals.metrics.count,'18446744073709551615');assert.equal(rows[1].signals.spans,undefined);assert.equal(rows[1].firstMs,5);assert.equal(rows[1].lastMs,22);rows[1].signals.metrics.count='0';assert.equal(source[0].count,'18446744073709551615');assert.throws(()=>catalogRows([source[0],source[0]]),/Duplicate/);
});
test('distributed relationships keep unresolved diagnostics out of edges and stage explicitly wider scoped questions',()=>{
 const {model,widerQuestion}=require('../ui/distributed-services'),evidence={dependencies:[{kind:'unresolved',child:'worker',parent:'',count:'9007199254740993'},{kind:'edge',child:'worker',parent:'api',count:'4'},{kind:'ambiguous',parent:'',child:'other',count:'2'}]};const before=structuredClone(evidence),m=model(evidence);assert.equal(m.edges.length,1);assert.equal(m.diagnostics.unresolved,'9007199254740993');assert.deepEqual(widerQuestion({environment:'prod',fromMs:10,toMs:20},m.edges[0]),{signal:'spans',environment:'prod',fromMs:10,toMs:20,limit:100,offset:0,where:{field:{kind:'column',key:'service'},op:'eq',value:'worker'}});assert.deepEqual(evidence,before);
});
test('dashboard result matching rejects same-scope wrong populations and detached capture intent',()=>{
 const {matches,question}=require('../ui/service-dashboard'),q={signal:'logs',environment:'prod',fromMs:10,toMs:20,limit:100,offset:0,where:{field:{kind:'column',key:'service'},op:'eq',value:'api'}},w={query:q};assert.equal(matches(w,{question:structuredClone(q)}),true);assert.equal(matches(w,{question:{...q,where:{...q.where,value:'other'}}}),false);const capture=question(w);capture.explore.where.value='changed';assert.equal(q.where.value,'api');
});
test('native dashboard summaries preserve latest unknown endpoints, exact stream identity and histogram ranges',()=>{
 const {nativeSummary}=require('../ui/service-dashboard'),stream={resource:{'service.instance.id':'one'},attributes:{region:'a'},unit:'1'};const data={gauges:{series:[{stream,state:'unknown',samples:[{atNS:'1000',state:'observed',value:.2},{atNS:'2000',state:'unknown',reason:'no recorded value'}]}]}};assert.equal(nativeSummary(data)[0].value,'Unavailable');assert.equal(nativeSummary(data)[0].extent,'Endpoint 2000 ns');data.gauges.series[0].state='observed';data.gauges.series[0].samples[1]={atNS:'2000',state:'observed',value:.25};assert.equal(nativeSummary(data)[0].value,'0.25 1');assert.deepEqual(nativeSummary(data)[0].stream,stream);const hist={histograms:{series:[{stream:{...stream,unit:'ms'},state:'observed',intervals:[{fromNS:'1',toNS:'10',state:'observed',distribution:{count:'4',unit:'ms',p95:{state:'bucket-bound',lower:10,upper:null}}}]}]}};assert.equal(nativeSummary(hist)[0].value,'4 observations · p95 10 … +∞ ms');
});
test('service namespace selectors partition missing, present-empty and literal names throughout questions',()=>{
 const id=require('../ui/service-identity'),{scopeQuestion}=require('../ui/service-browser'),{sameScope}=require('../ui/service-dashboard');const scope={environment:'prod',fromMs:1000,toMs:2000};const missing=id.namespace(null),empty=id.namespace(''),named=id.namespace('payments');assert.notEqual(id.key('api',missing),id.key('api',empty));assert.notEqual(id.key('api',named),id.key('api',null));for(const namespace of [missing,empty,named]){const q=scopeQuestion({...scope,namespace},'api');assert.equal(q.namespace,undefined);assert.deepEqual(q.where.all[1],id.predicate(namespace));assert.equal(sameScope({...scope,namespace},{...scope,namespace:structuredClone(namespace)}),true)}assert.equal(sameScope({...scope,namespace:empty},{...scope,namespace:missing}),false);
 const rows=require('../ui/inventory').catalogRows([{service:'api',signal:'logs',count:'1',firstMs:1,lastMs:2},{service:'api',namespace:'',signal:'logs',count:'2',firstMs:1,lastMs:2},{service:'api',namespace:'payments',signal:'logs',count:'3',firstMs:1,lastMs:2}]);assert.equal(rows.length,3);
});
test('declared services retain absent telemetry, exact namespaces, multiple bindings and independent provenance',()=>{
 const {catalogModel}=require('../ui/inventory');const evidence={question:{environment:'prod',fromMs:1,toMs:2},signals:[{service:'api',namespace:'payments',signal:'spans',count:'7',firstMs:1,lastMs:1}]};
 const pkg=(revision,namespace,instance='one')=>({revision,definition:{name:revision,title:revision,resource:{service:'api',environment:'prod',instance,...(namespace?{namespace}:{})},instrumentation:{name:'sdk',version:'1'}}});
 const before=JSON.stringify(evidence),packages=[pkg('unbound'),pkg('missing',{mode:'missing'}),pkg('orders',{mode:'value',value:'orders'}),pkg('other',{mode:'value',value:'orders'},'two')];const rows=catalogModel(evidence,packages);
 assert.equal(rows.length,3);const observed=rows.find(r=>r.observed);assert.equal(observed.signals.spans.count,'7');assert.equal(observed.declarations[0].namespaceMatch,'unspecified');const orders=rows.find(r=>r.namespace?.value==='orders');assert.equal(orders.observed,false);assert.equal(orders.lastMs,null);assert.deepEqual(Object.keys(orders.signals),[]);assert.equal(orders.declarations.length,2);assert.equal(JSON.stringify(evidence),before);
 assert.deepEqual(catalogModel(evidence,[...packages].reverse()),rows);assert.deepEqual(catalogModel({...evidence,question:{...evidence.question,environment:'other'}},packages).flatMap(r=>r.declarations),[]);
 const empty=catalogModel({...evidence,signals:[]},packages);assert.equal(empty.length,3);assert.equal(empty.filter(r=>r.namespace===null).length,1);assert.equal(empty.every(r=>!r.observed),true);
 const scoped=catalogModel({...evidence,question:{...evidence.question,namespace:{mode:'value',value:'payments'}}},packages);assert.equal(scoped.length,1);assert.equal(scoped[0].declarations.length,1);
});
