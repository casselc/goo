/* Independent observation check: never calls Goo's adapter or query builder. */
import fs from 'node:fs';
import assert from 'node:assert/strict';
const path=process.argv[2];if(!path)throw Error('Pass the original readiness report');
const original=JSON.parse(fs.readFileSync(path,'utf8'));
const attrs=values=>Object.fromEntries((values||[]).map(a=>[a.key,String(a.value.stringValue??a.value.intValue)]));
function bits(value){const bytes=new ArrayBuffer(8),view=new DataView(bytes);view.setFloat64(0,value,false);return String(view.getBigUint64(0,false))}
function check(report){
 assert.equal(report.status,'observed-unjudged');assert.equal(report.rows.length,3);assert.equal(report.traceRows.length,4);
 let points=0,examples=0,losses=0;
 for(const resource of report.plan.input.resourceMetrics)for(const scope of resource.scopeMetrics)for(const metric of scope.metrics)for(const point of metric.histogram.dataPoints){
  const resourceMap=attrs(resource.resource.attributes),pointMap=attrs(point.attributes),matches=report.rows.filter(row=>row.ScopeVersion===scope.scope.version&&row.ResourceAttributes['service.instance.id']===resourceMap['service.instance.id']&&row.Attributes['probe.case']===pointMap['probe.case']);
  assert.equal(matches.length,1);const row=matches[0];points++;
  assert.deepEqual(row.ResourceAttributes,resourceMap);assert.deepEqual(row.Attributes,pointMap);assert.equal(row.ResourceSchemaUrl,resource.schemaUrl);assert.equal(row.ScopeSchemaUrl,scope.schemaUrl);assert.equal(row.ScopeName,scope.scope.name);assert.equal(row.MetricName,metric.name);assert.equal(row.MetricUnit,metric.unit);assert.equal(row.startNS,point.startTimeUnixNano);assert.equal(row.timeNS,point.timeUnixNano);
  const exemplars=point.exemplars;
  for(const key of ['Exemplars.FilteredAttributes','Exemplars.TimeUnix','Exemplars.Value','Exemplars.TraceId','Exemplars.SpanId','exemplarTimes','exemplarValues','exemplarValueBits'])assert.equal(row[key].length,exemplars.length,key);
  exemplars.forEach((ex,index)=>{
   examples++;assert.equal(row['Exemplars.TraceId'][index],ex.traceId||'0'.repeat(32));assert.equal(row['Exemplars.SpanId'][index],ex.spanId||'0'.repeat(16));assert.equal(row.exemplarTimes[index],ex.timeUnixNano);assert.deepEqual(row['Exemplars.FilteredAttributes'][index],attrs(ex.filteredAttributes));
   const stored=Number(ex.asInt??ex.asDouble);assert.equal(row.exemplarValueBits[index],bits(stored));assert.equal(Number(row.exemplarValues[index]),stored);
   if(ex.asInt!==undefined&&BigInt(ex.asInt)!==BigInt(stored))losses++;
  });
 }
 assert.equal(points,3);assert.equal(examples,7);assert.equal(losses,1);
 const scoped=report.traceRows.filter(r=>r.ResourceAttributes['deployment.environment.name']===report.plan.environment&&r.ResourceAttributes['service.namespace']===report.plan.namespace);
 assert.equal(scoped.length,2);assert.equal(new Set(scoped.map(r=>r.ServiceName)).size,2);assert.deepEqual(new Set(scoped.map(r=>r.SpanId)),new Set([report.plan.spanA,report.plan.spanB]));assert.ok(scoped.every(r=>r.TraceId===report.plan.traceID));
 for(const item of report.requests)assert.equal(item.status,['known-traces','histogram'].includes(item.case)?200:400);
 for(const rows of Object.values(report.variantRows))assert.equal(rows.length,0);
 return {points,examples,scopedTraceRows:scoped.length,lossyIntegers:losses};
}
const result=check(original);
for(const change of [r=>r.rows[0]['Exemplars.TraceId'][0]='f'.repeat(32),r=>r.rows[0].exemplarTimes.reverse(),r=>r.rows[0]['Exemplars.SpanId'].pop()]){const mutant=structuredClone(original);change(mutant);assert.throws(()=>check(mutant),'Corrupted source association was accepted')}
process.stdout.write(JSON.stringify({...result,rejectedMutants:3,decision:'reported-reference subset retained; original scalar type and precision not preserved'})+'\n');
