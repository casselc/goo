// Generated ordinary OTLP; not a running kubeletstats/k8scluster receiver or Kubernetes agent.
const previous=require('./202-pod-memory-fixture'),{randomBytes}=require('node:crypto');
const scopes={usage:'github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver',request:'github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver',limit:'github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver'};
const names={usage:'container.cpu.usage',request:'k8s.container.cpu_request',limit:'k8s.container.cpu_limit'},clusterSchema='https://opentelemetry.io/schemas/1.18.0';
function fixture(){
 const toMs=Date.now(),environment='container-cpu-'+randomBytes(5).toString('hex'),primary={'deployment.environment.name':environment,'k8s.cluster.uid':'cluster-a','k8s.cluster.name':'shared-cluster','k8s.namespace.name':'team-a','k8s.pod.uid':'pod-a','k8s.pod.name':'same-pod','container.id':'container-a','k8s.container.name':'worker','telemetry.sdk.language':'python'};
 const facts={environment,fromMs:toMs-120000,toMs,primary,schemaURL:clusterSchema,scopes,names,points:[],records:[]};
 function point(role,value,options={}){const resource={...primary,...options.resource};for(const key of options.missingResource||[])delete resource[key];facts.points.push({kind:options.kind||'gauge',family:options.kind||'gauge',role,metric:options.metric||names[role],unit:options.unit??'{cpu}',value:String(value),valueType:options.integer?'int':'double',resource,resourceSchemaURL:options.resourceSchemaURL??clusterSchema,schemaURL:options.schemaURL??'',scope:options.scope||scopes[role],version:options.version||'0.137.0',attributes:options.attributes||{},flags:options.flags||0,startNS:'0',endNS:String(BigInt(toMs-(options.age||3000))*1000000n),temporality:2,monotonic:options.monotonic??false,case:options.case||'primary'});}
 // Realistic receiver variants have different raw schema/map identities: independent values only.
 point('usage',0.125,{resourceSchemaURL:'',resource:{'k8s.node.name':'node-a'},case:'primary-usage'});point('request',0.25,{case:'primary-request'});point('limit',0.5,{case:'primary-limit'});
 const cases=[['aligned',0.125,0.25,0.5],['over-100',0.75,0.25,0.5],['zero-denominator',0.125,0,0],['partial',0.125,null,null],['negative',-0.125,0.25,0.5],['fractional',0.1,0.2,0.4],['large','1000000',0.25,0.5],['stale',0.125,0.25,0.5],['conflict',0.125,0.25,0.5],['missing',0.125,0.25,0.5],['time-mismatch',0.125,0.25,0.5],['attributes-mismatch',0.125,0.25,0.5]];
 for(const [label,working,request,limit] of cases){const resource={'container.id':label,'k8s.pod.uid':'pod-'+label},values={usage:working,request,limit};for(const role of ['usage','request','limit']){const value=values[role];if(value===null)continue;const options={resource,case:label,double:label==='fractional'&&role==='usage',age:label==='stale'?90000:label==='time-mismatch'&&role!=='usage'?4000:3000,flags:label==='missing'&&role==='usage'?1:0,attributes:label==='attributes-mismatch'&&role!=='usage'?{'reported.variant':'other'}:{}};point(role,value,options);if(label==='conflict'&&role==='usage')point(role,0.0625,options);}}
 // Same pod/name with a replacement ID must not join the primary authority.
 for(const role of ['usage','request','limit'])point(role,0.875,{resource:{'container.id':'container-replacement'},case:'replacement-id'});
 point('usage',111,{resourceSchemaURL:'',resource:{'container.id':'container-a','k8s.cluster.uid':'cluster-b'},case:'wrong-cluster'});
 point('usage',112,{resourceSchemaURL:'',resource:{'container.id':'container-a','k8s.namespace.name':'team-b'},case:'wrong-namespace'});
 point('usage',113,{resource:{'container.id':''},case:'empty-container-id'});point('usage',114,{missingResource:['container.id'],case:'missing-container-id'});
 for(const [label,options] of [['wrong-scope',{scope:'kubeletstatsreceiver'}],['wrong-version',{version:'v0.137.0'}],['wrong-unit',{unit:'1'}],['wrong-schema',{schemaURL:clusterSchema}],['wrong-kind',{kind:'sum'}]])point('usage',117,{...options,resource:{'container.id':label},case:label});
 point('usage',1.5,{metric:'container.cpu.time',unit:'s',kind:'sum',monotonic:true,case:'separate-cumulative-cpu-time'});
 return facts;
}
module.exports={...previous,fixture,scopes,names,clusterSchema};
