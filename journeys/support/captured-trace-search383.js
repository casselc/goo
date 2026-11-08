// Independent physical-source acceptance adapter; never product presentation code.
const Base=require('./saved-trace-search382');
async function fixture(request){
 const f=await Base.fixture(request),first=f.first,other=f.other;
 const namespace=Object.hasOwn(first.resource,'service.namespace')?Base.leaf('resource','service.namespace','eq',first.resource['service.namespace']):Base.leaf('resource','service.namespace','missing');
 const absent='goo.acceptance.absent383';
 if(f.rows.some(r=>Object.hasOwn(r.attributes,absent)))throw Error('Acceptance missing-field key is actually reported');
 const requestValue={version:'goo/trace-search/v1',environmentSelector:f.parentQuestion.environmentSelector,fromMs:f.parentQuestion.fromMs,toMs:f.parentQuestion.toMs,conditions:[
  {all:[Base.leaf('column','trace_id','eq',first.traceId),Base.leaf('column','service','eq',first.service),namespace]},
  {all:[Base.leaf('column','service','eq',first.service),Base.leaf('column','name','eq',first.operation)]},
  Base.leaf('column','service','eq',other.service),
  {all:[Base.leaf('attribute',absent,'missing'),Base.leaf('column','duration_ns','lte',f.highThreshold)]}
 ],limit:f.limit};
 return {...f,searchRequest:requestValue,absentField:{kind:'attribute',key:absent},emptyQualifiedValues:f.rows.flatMap(r=>['resource','attributes'].flatMap(container=>Object.entries(r[container]).filter(([,v])=>v==='').map(([key])=>({spanId:r.spanId,kind:container==='resource'?'resource':'attribute',key,value:''}))))};
}
function checkResult(expect,f,result){
 const oracle=Base.oracle(f.rows,result.request);expect(result.matchedTraces).toBe(String(oracle.length));expect(result.items.map(i=>i.traceId)).toEqual(oracle.slice(0,result.request.limit).map(i=>i.traceId));expect(result.hasMore).toBe(oracle.length>result.request.limit);
 for(const item of result.items){const group=oracle.find(o=>o.traceId===item.traceId);for(const witness of item.witnesses){const matching=group.matching[witness.conditionIndex];expect(witness.matchingRecords).toBe(String(matching.length));const row=matching.find(r=>r.spanId===witness.spanId&&r.timestampNS===witness.timestampNs&&r.service===witness.service&&r.operation===witness.operation);expect(row,'Witness must be an independently read actual matching record').toBeTruthy();for(const value of witness.fieldValues){const source=value.field.kind==='resource'?row.resource:value.field.kind==='attribute'?row.attributes:null;const actual=source?(Object.hasOwn(source,value.field.key)?source[value.field.key]:null):({trace_id:row.traceId,span_id:row.spanId,service:row.service,name:row.operation,duration_ns:row.durationNS,timestamp_ns:row.timestampNS,kind:row.kind}[value.field.key]??null);expect(value.value).toBe(actual)}}}
 expect(oracle).toHaveLength(1);expect(oracle[0].matching[0].some(a=>oracle[0].matching[1].some(b=>a.spanId===b.spanId&&a.timestampNS===b.timestampNS))).toBe(true);expect(result.items[0].witnesses[0].service).not.toBe(result.items[0].witnesses[2].service);expect(result.items[0].witnesses[3].fieldValues).toContainEqual({field:f.absentField,value:null});return oracle;
}
module.exports={...Base,fixture,checkResult};
