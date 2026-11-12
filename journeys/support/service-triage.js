'use strict';
const {randomBytes}=require('node:crypto');
const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
const key=(service,namespace)=>JSON.stringify([service,namespace??null]);
function fixture(){
 const token=randomBytes(5).toString('hex'),environment='triage-'+token,toMs=Date.now()-1000,fromMs=toMs-60000,rows=[];
 const add=(service,namespace,window,durations,errorIndices=[],extra={})=>durations.forEach((durationMs,i)=>rows.push({service,namespace,window,durationNs:String(BigInt(durationMs)*1000000n),id:randomBytes(8).toString('hex'),traceId:randomBytes(16).toString('hex'),timeMs:(window==='current'?fromMs:fromMs-60000)+1000+i*100,status:errorIndices.includes(i)?2:1,kind:2,environment,...extra}));
 add('checkout','shop','current',[1,2,3,4,5,6],[0,4]);add('checkout','shop','previous',[2,2,2],[1]);
 add('checkout',undefined,'current',[10,20]);add('checkout',undefined,'previous',[5]);
 add('checkout','','current',[8,9,10],[1]);add('checkout','','previous',[4,4,4,4]);
 add('retired-api','shop','previous',[7,11],[0]);add('new-api','shop','current',[30],[0]);
 add('client-only','shop','current',[99,99],[0,1],{kind:3});
 add('checkout','foreign','current',[77]);
 add('environment-control','shop','current',[900],[0],{environment:'foreign-'+token,legacyEnvironment:environment});
 const scope={environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:environment}},fromMs,toMs};
 return{version:'goo.serviceTriageBank/1',token,environment,fromMs,toMs,scope,rows,logService:'log-only',metricService:'metric-only'};
}
function resource(r){return{attributes:attrs({'service.name':r.service,...(r.namespace===undefined?{}:{'service.namespace':r.namespace}),'deployment.environment.name':r.environment,...(r.legacyEnvironment?{'deployment.environment':r.legacyEnvironment}:{})})};}
function traces(f){return{resourceSpans:f.rows.map(r=>({resource:resource(r),scopeSpans:[{scope:{name:'ordinary.service-triage'},spans:[{traceId:r.traceId,spanId:r.id,name:'GET /items',kind:r.kind,status:{code:r.status},startTimeUnixNano:String(BigInt(r.timeMs)*1000000n),endTimeUnixNano:String(BigInt(r.timeMs)*1000000n+BigInt(r.durationNs)),attributes:attrs({'triage.bank':f.token,'http.request.method':'GET','http.route':'/items'})}]}]}))};}
function logs(f){return{resourceLogs:[{resource:resource({service:f.logService,namespace:'shop',environment:f.environment}),scopeLogs:[{scope:{name:'ordinary.service-triage'},logRecords:[{timeUnixNano:String(BigInt(f.fromMs+1000)*1000000n),severityNumber:9,severityText:'INFO',body:{stringValue:'ordinary log-only service'}}]}]}]};}
function metrics(f){return{resourceMetrics:[{resource:resource({service:f.metricService,namespace:'shop',environment:f.environment}),scopeMetrics:[{scope:{name:'ordinary.service-triage'},metrics:[{name:'process.memory.utilization',unit:'1',gauge:{dataPoints:[{timeUnixNano:String(BigInt(f.fromMs+1000)*1000000n),asDouble:.25}]}}]}]}]};}
function environment(row,q){const r=row.ResourceAttributes;return Object.hasOwn(r,'deployment.environment.name')?r['deployment.environment.name']:q.environmentSelector?.policy==='effective'?r['deployment.environment']:undefined;}
function matches(row,p){if(!p)return true;if(p.all)return p.all.every(x=>matches(row,x));if(p.any)return p.any.some(x=>matches(row,x));const value=p.field.kind==='resource'?row.ResourceAttributes[p.field.key]:p.field.kind==='attribute'?row.SpanAttributes[p.field.key]:({service:row.ServiceName,kind:row.SpanKind,name:row.SpanName,status:row.StatusCode})[p.field.key];if(p.op==='eq')return value===(p.value??'');if(p.op==='missing')return value===undefined;if(p.op==='exists')return value!==undefined;throw Error('Unsupported independent oracle predicate '+p.op);}
function members(raw,q){const expected=q.environmentSelector?.scope.value??q.environment;return raw.filter(r=>environment(r,q)===expected&&BigInt(r.timestamp_ns)>=BigInt(q.fromMs)*1000000n&&BigInt(r.timestamp_ns)<BigInt(q.toMs)*1000000n&&matches(r,q.where));}
function expected(raw,q){const selected=members(raw,q),groups=new Map();for(const r of selected){const namespace=r.ResourceAttributes['service.namespace']??null,k=key(r.ServiceName,namespace),g=groups.get(k)||{service:r.ServiceName,namespace,count:0n,error:0n,sumNs:0n};g.count++;g.error+=r.StatusCode==='Error'?1n:0n;g.sumNs+=BigInt(r.duration_ns);groups.set(k,g)}return{count:String(selected.length),groups:[...groups.values()].map(g=>({...g,count:String(g.count),error:String(g.error),sumNs:String(g.sumNs)})).sort((a,b)=>key(a.service,a.namespace).localeCompare(key(b.service,b.namespace)))};}
function assertResult(expect,raw,q,result){const oracle=expected(raw,q);expect(result.question).toEqual(q);expect(result.state).toBe(oracle.count==='0'?'no_data':'observed');expect(result.count).toBe(oracle.count);const actual=result.groups.map(g=>({service:g.g0,namespace:g.g1,count:g.count,error:g.spanStatus.error,sumNs:g.duration.sumNs})).sort((a,b)=>key(a.service,a.namespace).localeCompare(key(b.service,b.namespace)));expect(actual).toEqual(oracle.groups);for(const forbidden of ['rows','fields','plan','hasMore','resultToken'])expect(Object.hasOwn(result,forbidden)).toBe(false);return oracle;}
module.exports={fixture,traces,logs,metrics,key,members,expected,assertResult};
