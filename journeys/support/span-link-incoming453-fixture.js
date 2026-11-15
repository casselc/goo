'use strict';
const {randomBytes}=require('node:crypto');
const {expect}=require('@playwright/test');
const {source}=require('./cross-app');
const id=n=>randomBytes(n).toString('hex');
const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
function fixture(){
 const token=id(6),environment='incoming453-'+token,clock=Date.now(),ids={targetTrace:id(16),targetSpan:id(8),emptySpan:id(8),otherTrace:id(16),otherSpan:id(8)},base={'fixture.bank':token};
 const resource=(service,extra={})=>({'service.name':service,'deployment.environment.name':environment,...base,...extra});
 const target={traceId:ids.targetTrace,spanId:ids.targetSpan},link=(traceId=target.traceId,spanId=target.spanId)=>({traceId,spanId,attributes:{'messaging.message.id':'job-'+token}});
 const refs=Array.from({length:104},(_,i)=>({traceId:id(16),spanId:id(8),at:clock-20000+i,name:'publish '+i,resource:i%2?{'service.name':'publisher-b-'+token,'deployment.environment':environment,'service.namespace':'workers',...base}:resource('publisher-a-'+token,{'service.namespace':'background'}),attributes:{population:'other',occurrence:String(i)},links:i===0?[link(),link()]:[link()]}));
 refs.push({...refs[7],at:clock-20000+104,name:'duplicate span observation',attributes:{...refs[7].attributes,occurrence:'duplicate'}});
 const spans=[{...target,at:clock-10000,name:'selected consumer',resource:resource('consumer-'+token,{'service.namespace':'selected'}),attributes:{population:'selected'},links:[]},{traceId:ids.targetTrace,spanId:ids.emptySpan,parentSpanId:ids.targetSpan,at:clock-9000,name:'unreferenced child',resource:resource('consumer-'+token,{'service.namespace':'selected'}),attributes:{population:'selected'},links:[]},...refs,
 {traceId:id(16),spanId:id(8),at:clock-19000,name:'crossed tuple decoy',resource:resource('publisher-a-'+token),links:[link(ids.targetTrace,ids.otherSpan),link(ids.otherTrace,ids.targetSpan)]},
 {traceId:id(16),spanId:id(8),at:clock-19000,name:'wrong effective environment',resource:resource('publisher-a-'+token,{'deployment.environment.name':environment+'-wrong','deployment.environment':environment}),links:[link()]},
 {traceId:id(16),spanId:id(8),at:clock-120000,name:'outside original window',resource:resource('publisher-a-'+token),links:[link()]}];
 const origin={signal:'spans',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:environment}},fromMs:clock-60000,toMs:clock+1000,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'consumer-'+token},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:'selected'},{field:{kind:'attribute',key:'population'},op:'eq',value:'selected'},{field:{kind:'column',key:'span_id'},op:'eq',value:ids.targetSpan}]},limit:100,offset:0};
 return {token,environment,clock,ids,target,origin,spans};
}
async function emit(request,f){
 const payload={resourceSpans:f.spans.map(r=>({resource:{attributes:attrs(r.resource)},scopeSpans:[{scope:{name:'ordinary.incoming'},spans:[{traceId:r.traceId,spanId:r.spanId,parentSpanId:r.parentSpanId||'',name:r.name,kind:r.traceId===f.ids.targetTrace?5:4,startTimeUnixNano:String(BigInt(r.at)*1000000n+123456n),endTimeUnixNano:String(BigInt(r.at)*1000000n+1123456n),attributes:attrs(r.attributes||{}),links:r.links.map(l=>({...l,attributes:attrs(l.attributes)})),status:{code:1}}]}]}))};
 const response=await request.post('http://127.0.0.1:4318/v1/traces',{data:payload});expect(response.ok(),await response.text()).toBe(true);
 const query="SELECT TraceId AS traceId,SpanId AS spanId,toString(toUnixTimestamp64Nano(Timestamp)) AS atNS,ServiceName AS service,ResourceAttributes AS resource,SpanAttributes AS attributes,`Links.TraceId` AS linkedTraces,`Links.SpanId` AS linkedSpans,`Links.Attributes` AS linkedAttributes FROM otel_traces WHERE ResourceAttributes['fixture.bank']={bank:String} FORMAT JSONEachRow";
 await expect.poll(async()=>(await source(request,query,{param_bank:f.token,max_result_rows:1000})).length,{timeout:45000}).toBe(f.spans.length);
 return source(request,query,{param_bank:f.token,max_result_rows:1000});
}
const key=r=>r.traceId+'/'+r.spanId+'@'+r.atNS;
function matches(row,target,origin){
 const at=BigInt(row.atNS),resource=row.resource,environment=Object.hasOwn(resource,'deployment.environment.name')?resource['deployment.environment.name']:resource['deployment.environment'];
 if(at<BigInt(origin.fromMs)*1000000n||at>=BigInt(origin.toMs)*1000000n||environment!==origin.environmentSelector.scope.value)return false;
 const traces=row.linkedTraces,spans=row.linkedSpans,attributes=row.linkedAttributes;if(!Array.isArray(traces)||!Array.isArray(spans)||!Array.isArray(attributes)||traces.length!==spans.length||traces.length!==attributes.length)return false;
 return traces.some((trace,i)=>trace===target.traceId&&spans[i]===target.spanId);
}
function members(oracle,target,origin){return oracle.filter(r=>matches(r,target,origin)).sort((a,b)=>BigInt(a.atNS)<BigInt(b.atNS)?-1:BigInt(a.atNS)>BigInt(b.atNS)?1:0)}
module.exports={fixture,emit,key,matches,members};
