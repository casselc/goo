const {randomBytes}=require('node:crypto');
const attrs=value=>Object.entries(value).map(([key,stringValue])=>({key,value:{stringValue}}));
function fixture(){
 const token=randomBytes(5).toString('hex'),base=BigInt(Date.now()-30000)*1000000n+123457n;
 function make(role,offset,counts,bounds){
  const fromNS=String(base+BigInt(offset)),toNS=String(base+BigInt(offset)+2000321000n),unknownToNS=String(BigInt(toNS)+2000321000n),environment='native-view371-'+token+'-'+role,service='service-'+role+'-'+token,namespace='team-'+role,instance=role+'-instance',metric=role+'.request.duration',dimensions={'cohort':role,'http.route':'/'+role},resource={'service.name':service,'service.namespace':namespace,'service.instance.id':instance,...(role==='target'?{'deployment.environment':environment}:{'deployment.environment.name':environment})};
  const stream={resource,scope:'ordinary.native371',version:'1',schemaURL:'https://opentelemetry.io/schemas/1.30.0',name:metric,unit:'s',attributes:dimensions};
  const question={definition:{apiVersion:'goo/measurement-package/v1alpha1',name:role+'-duration',title:role==='source'?'Source duration':'Target duration',kind:'histogram',environmentSelector:{version:'goo/environment-selector/v1',policy:role==='target'?'effective':'current-only',scope:{mode:'named',value:environment}},lookbackMinutes:1,maxGapMs:10000,grain:'per-stream',source:{services:[service],namespace:{mode:'value',value:namespace},scope:stream.scope,version:stream.version,schemaURL:stream.schemaURL,metric,unit:'s',attributes:{cohort:role},...(role==='target'?{exactStream:stream}:{})}},fromMs:Number(BigInt(fromNS)/1000000n),toMs:Number((BigInt(unknownToNS)+999999n)/1000000n),fromNS,toNS:unknownToNS};
  const count=counts.reduce((n,x)=>n+BigInt(x),0n).toString(),points=[{startTimeUnixNano:fromNS,timeUnixNano:toNS,count,bucketCounts:counts,explicitBounds:bounds,attributes:attrs(dimensions)},...(role==='target'?[{startTimeUnixNano:toNS,timeUnixNano:unknownToNS,flags:1,count:'0',bucketCounts:['0','0','0'],explicitBounds:bounds,attributes:attrs(dimensions)}]:[])];
  return{role,environment,service,namespace,instance,metric,stream,question,fromNS,toNS,unknownToNS,count,bounds,counts,resource,points};
 }
 return{token,source:make('source',0,['1','2','1'],[0.01,0.1]),target:make('target',7000321000,['2','3','2'],[0.02,0.2])};
}
function payload(f){return{resourceMetrics:[f.source,f.target].map(c=>({resource:{attributes:attrs(c.resource)},scopeMetrics:[{scope:{name:c.stream.scope,version:c.stream.version},schemaUrl:c.stream.schemaURL,metrics:[{name:c.metric,unit:'s',histogram:{aggregationTemporality:1,dataPoints:c.points}}]}]}))}}
const declaration=kind=>({version:'goo.view/1',scope:{kind:'measurement-question',signal:'metrics'},selection:{kind:'export-interval'},kind:{name:kind,version:1},display:{category:'bucket',value:'observation-count',requiresComplete:true}});
module.exports={fixture,payload,declaration};
