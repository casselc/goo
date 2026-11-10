const {test}=require('node:test'),assert=require('node:assert/strict');
const Dashboard=require('../ui/service-dashboard.mjs'),Capabilities=require('../ui/instrument-capabilities.mjs');
const definition={kind:'histogram',title:'Outgoing request duration',source:{scope:'otel/http',metric:'http.client.request.duration',unit:'s'}};
test('native outgoing duration presentation remains independent from the outgoing span lens',()=>{
 const widget={key:'native-1',title:'Outgoing request duration',family:'http-client',measurement:{definition}},plan={profile:'all-spans'},before=structuredClone(widget);
 assert.equal(Dashboard.widgetTitle(widget,plan),'Outgoing request duration [s]');
 assert.match(Dashboard.widgetQualifier(widget),/Native HTTP client request duration.*seconds.*independent per-stream/);
 assert.match(Dashboard.widgetQualifier(widget),/not HTTP client-span records, unique requests, network attempts or remote-server health/);
 assert.equal(Dashboard.recordFamily(widget),null);assert.equal(Dashboard.incomingFamily(widget),null);assert.equal(Dashboard.isHTTP(widget,plan),false);
 assert.equal(Dashboard.profileSelectable({key:'http-client',state:'not_observed'}),false);assert.equal(Dashboard.profileSelectable({key:'http-client',state:'candidate'}),true);
 assert.equal(Dashboard.profileState(plan,[{widget,result:{state:'observed'}}]),null);assert.deepEqual(widget,before);
});
test('browser retains supplied client-duration diagnostics without promoting instrument shape into convention or span proof',()=>{
 const capabilityPlan={version:Capabilities.version,profile:'http-client-request-duration',title:'HTTP client request duration candidate',eligible:true,verification:'not-run',ruleSet:'goo/http-client-request-duration/v1',reasons:[],hints:['Required method/address/port dimensions are diagnostic; streams are not filtered.'],coverage:{state:'partial',streams:[],dimensions:[]}},instrument={metric:'http.client.request.duration',kind:'histogram',unit:'s',schemaURL:'https://opentelemetry.io/schemas/1.45.0',supported:true,definition,capabilityPlan},before=structuredClone(instrument);
 assert.deepEqual(Capabilities.describe(instrument),capabilityPlan);const modified=Capabilities.describe(instrument);modified.hints.push('local');assert.deepEqual(instrument,before);
 // Metadata is displayed exactly as planned by Go, never reclassified in JavaScript.
 assert.deepEqual(Capabilities.describe({...instrument,unit:'ms'}),capabilityPlan);assert.equal(Capabilities.describe({...instrument,capabilityPlan:null}).eligible,false);
 assert.equal(Dashboard.httpClientObserved({environment:'prod',fromMs:1000,toMs:2000},'api',{state:'observed',question:{},rows:[]},{}),false);
});
