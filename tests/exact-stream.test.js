const {test}=require('node:test'),assert=require('node:assert/strict');
const {streamQuestion}=require('../ui/measurements'),w=require('../ui/workspace');
test('stream focus preserves exact bounds and intent through sharing and capture',()=>{
 const stream={resource:{'service.name':'api','service.instance.id':'one','deployment.environment.name':'prod','service.namespace':''},resourceSchemaURL:'resource-v1',scope:'otel.http',version:'1',schemaURL:'otel-v1',name:'requests',unit:'1',attributes:{route:'',method:'GET'}};
 const q={definition:{kind:'counter',environment:'prod',source:{scope:'otel.http',metric:'requests',unit:'1'}},fromMs:1000,toMs:3001,fromNS:'1000000123',toNS:'3000000123'},r={question:q};
 const focused=streamQuestion(r,stream);assert.equal(q.definition.source.exactStream,undefined);
 const step=w.readStepLink(w.stepLink(w.measurementStep(focused)));assert.deepEqual(w.captureQuestion(step).measurement,focused);
 assert.ok(w.evidenceDifference(w.measurementStep(q),{kind:'measurement',measurement:focused}).includes('Question definition differs'));
 stream.attributes.method='POST';assert.equal(focused.definition.source.exactStream.attributes.method,'GET');
 assert.deepEqual(streamQuestion({question:focused},null),q);assert.equal(focused.definition.source.exactStream.resource['service.namespace'],'');
 assert.throws(()=>streamQuestion({question:{definition:{kind:'lifecycle'}}},stream));
});
const {streamDiagnostics,inventoryView,reasonInfo}=require('../ui/measurements');
test('local stream navigation scans complete identities and keeps input immutable',()=>{
 const series=Array.from({length:61},(_,i)=>({stream:{resource:{'service.instance.id':'worker-'+i,'service.namespace':i===60?'hidden-namespace':'shared'},scope:'otel.http',name:'requests',attributes:{route:i===60?'':'/search'},resourceSchemaURL:i===60?'resource-v2':'v1'},state:i===60?'unknown':'ok',intervals:[{state:'ok',delta:'0'}]}));
 const before=JSON.stringify(series);assert.equal(inventoryView(series,'counter').items.length,25);assert.equal(inventoryView(series,'counter',{offset:25}).items[0],series[25]);assert.equal(inventoryView(series,'counter',{offset:500}).items.length,11);
 for(const query of ['hidden-namespace','resource-v2','worker-60']){const view=inventoryView(series,'counter',{query,offset:25});assert.equal(view.items[0],series[60]);assert.equal(view.offset,0);assert.equal(view.total,61)}
 assert.equal(inventoryView(series,'counter',{query:'absent'}).matched,0);assert.deepEqual(inventoryView(series,'counter',{state:'review'}).items,[series[60]]);assert.equal(inventoryView(series,'counter',{size:0}).items.length,25);assert.equal(inventoryView(series,'counter',{size:100}).items.length,25);assert.equal(JSON.stringify(series),before);
});
test('native diagnostics preserve unknowns, zero intervals, gauge freshness and exact reasons',()=>{
 assert.equal(streamDiagnostics({state:'ok',intervals:[{state:'ok',delta:'0'}]},'counter').label,'Window covered');
 const gap=streamDiagnostics({state:'unknown',intervals:[{state:'ok'}]},'counter');assert.equal(gap.attention,true);assert.deepEqual(gap.reasons,['incomplete-window']);
 const mixed=streamDiagnostics({state:'unknown',intervals:[{state:'unknown',reason:'counter-decrease'},{state:'unknown',reason:'counter-decrease'}]},'counter');assert.equal(mixed.review,2);assert.deepEqual(mixed.reasons,['counter-decrease']);assert.match(reasonInfo(mixed.reasons[0]).title,/decreased/);
 const gauge=streamDiagnostics({state:'observed',samples:[{state:'stale',reason:'older than declared freshness'},{state:'observed',value:0}]},'gauge');assert.equal(gauge.label,'Latest observation recorded');assert.equal(gauge.attention,true);assert.equal(gauge.recorded,1);assert.equal(gauge.review,1);
 for(const code of ['new-code','constructor','__proto__']){assert.equal(reasonInfo(code).code,code);assert.match(reasonInfo(code).title,/additional diagnostic/)}
 assert.equal(streamDiagnostics({state:'no_data',intervals:[]},'histogram').recorded,0);
});
test('generated local filter/page sequences never rewrite or fabricate native streams',()=>{
 for(let n=1;n<=100;n++){const series=Array.from({length:n},(_,i)=>({stream:{resource:{'service.instance.id':String(i)},attributes:{route:'/route-'+i}},state:i%3?'ok':'unknown',intervals:[]}));const before=JSON.stringify(series);
  for(let step=0;step<12;step++){const v=inventoryView(series,'counter',{query:step%2?'route-'+step:'',state:step%3?'all':'review',offset:step*25});assert.ok(v.items.length<=25);assert.ok(v.matched<=n);assert.equal(v.total,n);assert.ok(v.items.every(s=>series.includes(s)));assert.equal(JSON.stringify(series),before)}
 }
});

test('identity search matches literal quotes and backslashes rather than encoded JSON',()=>{
 const stream={resource:{'service.instance.id':'one',literal:'quote"\\backslash'},attributes:{empty:''},resourceSchemaURL:'schema-v1'},s={stream,state:'ok',intervals:[]};
 assert.equal(inventoryView([s],'counter',{query:'quote"\\backslash'}).matched,1);
 assert.equal(inventoryView([s],'counter',{query:'empty'}).matched,1);
});
