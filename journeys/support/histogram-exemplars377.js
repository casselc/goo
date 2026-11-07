const fs=require('node:fs/promises'),path=require('node:path');
async function fixture(){
 const probePath=path.resolve(process.env.GOO_HISTOGRAM_EXEMPLARS_377_REPORT||'artifacts/histogram-exemplars377/report.json'),raw=JSON.parse(await fs.readFile(probePath,'utf8'));
 if(raw.status!=='observed-unjudged'||raw.rows.length!==3||raw.traceRows.length!==4)throw Error('Retained real377 source probe required; no replacement ingestion');
 const p=raw.plan,primary=raw.rows.find(r=>r.Attributes['probe.case']==='primary'),empty=raw.rows.find(r=>r.Attributes['probe.case']==='empty-exemplars'),fromNS=primary.startNS,toNS=String(BigInt(empty.timeNS)+1n);
 const question={definition:{apiVersion:'goo/measurement-package/v1alpha1',name:'reported-exemplars377',title:'Reported histogram exemplars',kind:'histogram',environmentSelector:{version:'goo/environment-selector/v1',policy:'current-only',scope:{mode:'named',value:p.environment}},lookbackMinutes:1,maxGapMs:10000,grain:'per-stream',source:{services:[p.service],namespace:{mode:'value',value:p.namespace},scope:primary.ScopeName,version:primary.ScopeVersion,schemaURL:primary.ScopeSchemaUrl,metric:primary.MetricName,unit:primary.MetricUnit,attributes:{'http.route':'/work-雪'}}},fromMs:Number(BigInt(fromNS)/1000000n),toMs:Number((BigInt(toNS)+999999n)/1000000n),fromNS,toNS};
 return{probePath,token:raw.token,question,primary,empty,traceID:p.traceID,spanA:p.spanA,spanB:p.spanB,namespace:p.namespace,environment:p.environment,service:p.service,rawRows:raw.rows,rawTraceRows:raw.traceRows};
}
module.exports={fixture};
