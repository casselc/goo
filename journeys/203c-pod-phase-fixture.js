// Generated ordinary OTLP; not a running k8sclusterreceiver or Kubernetes agent.
const previous = require('./202-pod-memory-fixture');
const {randomBytes} = require('node:crypto');
const phaseScope = 'github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver';
function fixture() {
 const facts=previous.fixture(), old=facts.environment; facts.environment='pod-phase-'+randomBytes(5).toString('hex');
 for(const record of facts.records) record.resource['deployment.environment.name']=facts.environment;
 facts.primary['deployment.environment.name']=facts.environment;
 for(const point of facts.points) point.resource['deployment.environment.name']=facts.environment;
 // Keep only the primary memory pair, so phase admission cannot replace either default.
 facts.points=facts.points.filter(p=>p.case==='primary' && p.metric.startsWith('k8s.pod.memory.'));
 facts.phaseScope=phaseScope;
 function phase(uid,value,options={}) { facts.points.push({kind:options.kind||'gauge',family:options.kind||'gauge',metric:'k8s.pod.phase',unit:options.unit??'',value:String(value),valueType:options.double?'double':'int',resource:{...facts.primary,'k8s.pod.uid':uid,...(options.resource||{})},resourceSchemaURL:options.resourceSchemaURL??'',schemaURL:options.schemaURL??'',scope:options.scope||phaseScope,version:options.version||'0.137.0',attributes:options.attributes||{},flags:options.flags||0,startNS:'0',endNS:String(BigInt(facts.toMs-(options.age||3000))*1000000n),temporality:2,monotonic:false,case:options.case||uid}); }
 phase('uid-a',1,{age:20000});phase('uid-a',2);
 phase('phase-only',1,{age:20000});phase('phase-only',2);
 phase('reported-unknown',5);phase('unsupported',7);phase('fractional',1.5,{double:true});
 phase('stale-phase',2,{age:90000});phase('stale-unsupported',7,{age:90000});
 phase('phase-conflict',1);phase('phase-conflict',2);phase('phase-missing',0,{flags:1});
 phase('wrong-version',2,{version:'v0.137.0'});phase('wrong-scope',2,{scope:'k8sclusterreceiver'});
 phase('wrong-schema',2,{schemaURL:previous.semconvSchema});phase('wrong-unit',2,{unit:'1'});phase('wrong-kind',2,{kind:'sum'});
 phase('phase-only',4,{resourceSchemaURL:previous.semconvSchema,case:'same-uid-other-resource-schema'});
 phase('phase-only',3,{resource:{'k8s.cluster.uid':'cluster-2'},case:'same-uid-other-resource-map'});
 return facts;
}
module.exports={...previous,fixture,phaseScope};
