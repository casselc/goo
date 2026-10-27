// Generated ordinary OTLP; not a running k8sclusterreceiver or Kubernetes agent.
const previous=require('./202-pod-memory-fixture'),{randomBytes}=require('node:crypto');
const receiverScope='github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver';
function fixture(){
 const facts=previous.fixture(),memoryTemplates=facts.points.filter(p=>p.case==='primary'&&p.endNS===facts.points[1].endNS);facts.environment='container-ready-'+randomBytes(5).toString('hex');
 facts.primary={...facts.primary,'deployment.environment.name':facts.environment,'container.id':'reported-container-a','k8s.container.name':'worker'};
 for(const record of facts.records)record.resource['deployment.environment.name']=facts.environment;
 facts.points=[];facts.receiverScope=receiverScope;facts.receiverVersion='0.137.0';
 function point(value,options={}){
  const resource={...facts.primary,...(options.resource||{})};for(const key of options.missingResource||[])delete resource[key];
  facts.points.push({kind:options.kind||'gauge',family:options.kind||'gauge',metric:'k8s.container.ready',unit:options.unit??'',value:String(value),valueType:options.double?'double':'int',resource,resourceSchemaURL:options.resourceSchemaURL??'',schemaURL:options.schemaURL??'',scope:options.scope||receiverScope,version:options.version||'0.137.0',attributes:options.attributes||{},flags:options.flags||0,startNS:'0',endNS:String(BigInt(facts.toMs-(options.age||3000))*1000000n),temporality:2,monotonic:false,case:options.case||'primary'});
 }
 point(0,{age:20000});point(1);
 point(1,{resource:{'container.id':'mixed-memory'},case:'mixed-memory'});
 for(const metric of ['k8s.pod.memory.usage','k8s.pod.memory.working_set']){
  const template=memoryTemplates.find(p=>p.metric===metric);facts.points.push({...template,resource:{...facts.primary,'container.id':'mixed-memory'},case:'mixed-memory'});
 }
 point(1,{resource:{'container.id':'cluster-pair'},case:'cluster-pair'});
 const phaseTemplate={...facts.points.at(-1),metric:'k8s.pod.phase',value:'2'};facts.points.push(phaseTemplate);

 point(0,{resource:{'container.id':'reported-container-not-ready'},case:'not-ready'});
 point(0,{resource:{'container.id':'reported-container-replacement'},case:'replacement-id-same-name'});
 point(0,{resourceSchemaURL:previous.semconvSchema,case:'same-id-different-resource-schema'});
 point(0,{resource:{'k8s.cluster.uid':'cluster-2'},case:'same-id-different-resource-map'});
 point(1,{resource:{'container.id':'point-attributes'},case:'point-attributes'});point(0,{resource:{'container.id':'point-attributes'},attributes:{'reported.variant':'other'},case:'same-resource-different-point-attributes'});
 point(1,{resource:{'container.id':'stale'},age:90000,case:'stale'});
 point(1,{resource:{'container.id':'conflict'},case:'conflict'});point(0,{resource:{'container.id':'conflict'},case:'conflict'});
 point(1,{resource:{'container.id':'missing-data'},age:20000,case:'missing-data-earlier-ready'});point(0,{resource:{'container.id':'missing-data'},flags:1,case:'missing-data'});
 point(2,{resource:{'container.id':'unsupported'},case:'unsupported'});point(.5,{resource:{'container.id':'fractional'},double:true,case:'fractional'});
 point(2,{resource:{'container.id':'stale-unsupported'},age:90000,case:'stale-unsupported'});
 point(1,{resource:{'container.id':'wrong-version'},version:'v0.137.0',case:'wrong-version'});
 point(1,{resource:{'container.id':'wrong-scope'},scope:'k8sclusterreceiver',case:'wrong-scope'});
 point(1,{resource:{'container.id':'wrong-schema'},schemaURL:previous.semconvSchema,case:'wrong-schema'});
 point(1,{resource:{'container.id':'wrong-unit'},unit:'1',case:'wrong-unit'});
 point(1,{resource:{'container.id':'wrong-kind'},kind:'sum',case:'wrong-kind'});
 point(1,{resource:{'container.id':''},case:'empty-container-id'});
 point(1,{missingResource:['container.id'],case:'missing-container-id'});
 point(1,{resource:{'container.id':'missing-name'},missingResource:['k8s.container.name'],case:'missing-container-name'});
 point(1,{resource:{'container.id':'missing-pod'},missingResource:['k8s.pod.uid'],case:'missing-pod-uid'});
 return facts;
}
module.exports={...previous,fixture,receiverScope};
