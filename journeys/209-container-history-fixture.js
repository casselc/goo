// Generated ordinary OTLP; not an actual receiver, Kubernetes agent or configuration event.
const base=require('./208-container-resources-fixture'),{randomBytes}=require('node:crypto');
function fixture(){
 const original=base.fixture(),toMs=Date.now(),environment='container-history-'+randomBytes(5).toString('hex');
 const points=original.points.filter(p=>p.case.startsWith('primary-')).map(p=>({...p,resource:{...p.resource,'deployment.environment.name':environment},endNS:String(BigInt(toMs-3000)*1000000n)}));
 const primary={...original.primary,'deployment.environment.name':environment};
 function add(family,role,age,value,options={}){const template=points.find(p=>p.familyName===family&&p.role===role);points.push({...template,value:String(value),endNS:String(BigInt(toMs-age)*1000000n),valueType:options.double?'double':template.valueType,flags:options.flags||0,case:options.case||'history'});}
 for(const [age,value,options] of [[110000,8192,{}],[100000,16384,{}],[15000,1.5,{double:true,case:'invalid-bytes'}],[8000,32768,{flags:1,case:'missing-byte-observation'}]])add('memory','workingSet',age,value,options);
 for(const role of ['request','limit'])add('memory',role,50000,role==='request'?32768:65536,{case:'different-reported-memory-value'});
 for(let i=0;i<265;i++)add('cpu','usage',100000-i*350,(i%7+1)/100,{case:'paged-cpu-observation'});
 add('cpu','usage',7000,.2,{flags:1,case:'missing-cpu-observation'});add('cpu','usage',6000,.3,{case:'conflicting-cpu-observation'});add('cpu','usage',6000,.4,{case:'conflicting-cpu-observation'});
 for(const role of ['request','limit'])add('cpu',role,50000,role==='request'?.125:.25,{case:'different-reported-cpu-value'});
 const fromNS=String(BigInt(toMs-20000)*1000000n),toNS=String(BigInt(toMs-10000)*1000000n),records=[];
 for(const family of ['logs','spans'])for(const [label,at,resource] of [
  ['before-lower',BigInt(fromNS)-1n,{}],['at-lower',BigInt(fromNS),{}],['inside',BigInt(fromNS)+123456n,{}],['before-upper',BigInt(toNS)-1n,{}],['at-upper',BigInt(toNS),{}],['after-upper',BigInt(toNS)+1n,{}],['replacement-container',BigInt(fromNS)+1000000n,{'container.id':'replacement'}],['other-namespace',BigInt(fromNS)+1000000n,{'k8s.namespace.name':'other'}]]){
  const id=randomBytes(8).toString('hex');records.push({family,signal:family,id,spanId:id,traceId:randomBytes(16).toString('hex'),resource:{...primary,...resource,'service.name':'history-context-app','service.instance.id':'record-worker','record.extra':'allowed'},resourceSchemaURL:original.schemaURL,timeNS:String(at),name:label,body:'History context '+label,attributes:{'work.id':id,'fixture.case':label},case:label});
 }
 return {...original,environment,fromMs:toMs-120000,toMs,primary,points,records,contextRange:{fromNS,toNS},rawDuplicateCaveat:'OTLP double encoding cannot preserve distinct original decimal spellings which round to one Float64; independent kernel properties cover that refusal.'};
}
module.exports={...base,fixture};
