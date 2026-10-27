// Generated ordinary OTLP resource observations; no Kubernetes API or verified owner hierarchy.
const previous=require('./201-kubernetes-fixture'),{randomBytes}=require('node:crypto');
function fixture(){
 const toMs=Date.now(),environment='workload-context-'+randomBytes(5).toString('hex');
 const primary={'deployment.environment.name':environment,'k8s.cluster.uid':'cluster-a','k8s.cluster.name':'shared-cluster','k8s.namespace.name':'team-a','k8s.pod.uid':'pod-a','k8s.pod.name':'same-pod','container.id':'container-a','k8s.container.name':'worker','k8s.deployment.name':'same-work','k8s.deployment.uid':'deployment-a','k8s.replicaset.name':'same-work-rs','k8s.replicaset.uid':'replicaset-a','telemetry.sdk.language':'python'};
 const facts={environment,fromMs:toMs-120000,toMs,schemaURL:previous.schemaURL,primary,points:[],records:[]};
 function resource(changes={},missing=[]){const r={...primary,...changes};for(const k of missing)delete r[k];return r;}
 function variant(label,r){facts.points.push({kind:'gauge',family:'gauge',metric:'fixture.workload.observed',unit:'{item}',value:'1',attributes:{},resource:{...r},resourceSchemaURL:previous.schemaURL,schemaURL:'',scope:'ordinary.workload',version:'1',startNS:'0',endNS:String(BigInt(toMs-3000)*1000000n),case:label});return r;}
 function record(family,r,label,age=3000){const id=randomBytes(8).toString('hex');facts.records.push({family,signal:family,id,spanId:id,traceId:randomBytes(16).toString('hex'),resource:{...r,'service.name':'ordinary-workload-app','service.instance.id':'app-worker'},resourceSchemaURL:previous.schemaURL,timeNS:String(BigInt(toMs-age-facts.records.length)*1000000n),name:'reported-workload-context',body:'Reported workload '+id,attributes:{'work.id':id},case:label});}
 variant('primary',primary);for(const family of ['logs','spans'])record(family,primary,'primary');
 const replacement=variant('replacement-container',resource({'container.id':'container-replacement'}));for(const family of ['logs','spans'])record(family,replacement,'replacement-container');
 const sibling=variant('same-workload-other-pod',resource({'k8s.pod.uid':'pod-b','container.id':'container-b','k8s.replicaset.name':'other-rs','k8s.replicaset.uid':'other-rs-id'}));for(const family of ['logs','spans'])record(family,sibling,'same-workload-other-pod');
 const controls=[['other-workload-uid',{'k8s.deployment.uid':'deployment-b'}],['other-namespace',{'k8s.namespace.name':'team-b'}],['other-cluster',{'k8s.cluster.uid':'cluster-b'}],['empty-namespace',{'k8s.namespace.name':''}],['empty-cluster',{'k8s.cluster.uid':''}],['empty-workload-uid',{'k8s.deployment.uid':''}]];
 for(const [label,changes] of controls){const r=variant(label,resource(changes));for(const family of ['logs','spans'])record(family,r,label);}
 const nameOnly=variant('name-only',resource({'k8s.pod.uid':'name-only','container.id':'name-only'},['k8s.deployment.uid']));for(const family of ['logs','spans'])record(family,nameOnly,'name-only');
 variant('missing-namespace',resource({'k8s.pod.uid':'missing-namespace'},['k8s.namespace.name']));
 variant('missing-cluster',resource({'k8s.pod.uid':'missing-cluster'},['k8s.cluster.uid']));
 variant('empty-name',resource({'k8s.pod.uid':'empty-name','k8s.deployment.name':''},['k8s.deployment.uid']));
 variant('no-workload',resource({'k8s.pod.uid':'no-workload'},['k8s.deployment.uid','k8s.deployment.name','k8s.replicaset.name','k8s.replicaset.uid']));
 for(const kind of ['statefulset','daemonset','job','cronjob'])variant('reported-'+kind,resource({'k8s.pod.uid':'reported-'+kind,['k8s.'+kind+'.name']:'same-work',['k8s.'+kind+'.uid']:kind+'-id'},['k8s.deployment.uid','k8s.deployment.name','k8s.replicaset.name','k8s.replicaset.uid']));
 for(const [r,value,age] of [[primary,'0',20000],[primary,'1',3000],[replacement,'0',3000]])facts.points.push({kind:'gauge',family:'gauge',metric:'k8s.container.ready',unit:'',value,attributes:{},resource:{...r},resourceSchemaURL:previous.schemaURL,schemaURL:'',scope:'github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver',version:'0.137.0',startNS:'0',endNS:String(BigInt(toMs-age)*1000000n),case:'reported-readiness'});
 const wrongEnv=resource({'deployment.environment.name':environment+'-excluded'});for(const family of ['logs','spans'])record(family,wrongEnv,'wrong-environment');
 for(const family of ['logs','spans'])record(family,primary,'outside-window',150000);
 return facts;
}
module.exports={...previous,fixture};
