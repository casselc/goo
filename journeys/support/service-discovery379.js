// Retained real Collector exports; this journey never adds telemetry.
const Fixture376=require('./unsupported-instruments376');
const {source}=require('./cross-app');
async function fixtures(){
 const f=await Fixture376.fixtures();
 const catalogQuestion={environment:f.unsupported.environmentSelector.scope.value,namespace:structuredClone(f.unsupported.namespace),fromMs:f.unsupported.fromMs,toMs:f.unsupported.toMs};
 return {...f,catalogQuestion,instrumentQuestion:{...catalogQuestion,service:f.unsupported.service}};
}
async function exponentialRows(request,q){
 return source(request,"SELECT ServiceName AS service,ScopeName AS scope,ScopeVersion AS version,ScopeSchemaUrl AS schemaURL,MetricName AS metric,MetricUnit AS unit,AggregationTemporality AS temporality,ResourceAttributes AS resource,Attributes AS attributes,toString(toUnixTimestamp64Nano(TimeUnix)) AS timeNS,toString(Count) AS count FROM otel_metrics_exp_histogram WHERE ServiceName={service:String} AND ResourceAttributes['deployment.environment.name']={environment:String} AND ResourceAttributes['service.namespace']={namespace:String} AND TimeUnix>=fromUnixTimestamp64Milli({from:Int64}) AND TimeUnix<fromUnixTimestamp64Milli({to:Int64}) ORDER BY MetricName,TimeUnix FORMAT JSONEachRow",{param_service:q.service,param_environment:q.environmentSelector.scope.value,param_namespace:q.namespace.value,param_from:q.fromMs,param_to:q.toMs});
}
async function ordinaryRows(request,q){
 const selection=q.environmentSelector,environment=selection.scope.value,current="ResourceAttributes['deployment.environment.name']={environment:String}",predicate=selection.policy==='current-only'?current:"if(mapContains(ResourceAttributes,'deployment.environment.name'),ResourceAttributes['deployment.environment.name'],ResourceAttributes['deployment.environment'])={environment:String}";
 const families=[['spans','otel_traces','Timestamp'],['logs','otel_logs','Timestamp'],['metrics','otel_metrics_gauge','TimeUnix'],['metrics','otel_metrics_sum','TimeUnix'],['metrics','otel_metrics_histogram','TimeUnix']];
 return source(request,"SELECT signal,toString(sum(n)) AS count FROM ("+families.map(([signal,table,time])=>"SELECT '"+signal+"' AS signal,count() AS n FROM "+table+" WHERE ServiceName={service:String} AND "+predicate+" AND ResourceAttributes['service.namespace']={namespace:String} AND "+time+">=fromUnixTimestamp64Milli({from:Int64}) AND "+time+"<fromUnixTimestamp64Milli({to:Int64})").join(' UNION ALL ')+") GROUP BY signal HAVING sum(n)>0 ORDER BY signal FORMAT JSONEachRow",{param_service:q.service,param_environment:environment,param_namespace:q.namespace.value,param_from:q.fromMs,param_to:q.toMs});
}
module.exports={fixtures,exponentialRows,ordinaryRows,groupRows:Fixture376.groupRows};
