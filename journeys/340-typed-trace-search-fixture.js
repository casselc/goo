const {randomBytes}=require('node:crypto');
const attrs=value=>Object.entries(value).map(([key,stringValue])=>({key,value:{stringValue}}));
const CURRENT='deployment.environment.name',LEGACY='deployment.environment';
function fixture(){
 const token=randomBytes(5).toString('hex'),environment='trace340-'+token,clock=Date.now(),records=[],traces=[];
 const service='db340-'+token,operation='SELECT 340 '+token;
 const keys={current:{[CURRENT]:environment},legacy:{[LEGACY]:environment},mixed:{[CURRENT]:environment,[LEGACY]:environment+'-shadow'},wrong:{[CURRENT]:environment+'-other',[LEGACY]:environment},empty:{[CURRENT]:'',[LEGACY]:environment},legacyEmpty:{[LEGACY]:''},unreported:{}};
 function add(role,dbKeys,workerKeys=dbKeys,sameSpan=false){
  const traceId=(token+'0'.repeat(22)).slice(0,24)+(traces.length+1).toString(16).padStart(8,'0');
  const definitions=[['api340-'+token,'GET /340',1,2,dbKeys], [service,operation,sameSpan?2:1,3,dbKeys],...(!sameSpan?[['worker340-'+token,'process 340 '+token,2,4,workerKeys]]:[])];
  const rows=definitions.map(([name,op,status,kind,resource],index)=>({traceId,spanId:(index+1).toString(16).padStart(16,'0'),parentSpanId:index?'0000000000000001':'',service:name,operation:op,startNs:String(BigInt(clock-2000+index*10)*1000000n),durationNs:'80000000',status,kind,resource:{...resource,'service.name':name,'service.namespace':'native340-'+token},attributes:{'test.run':token,'http.request.method':'GET',...(name===service?{'db.system.name':'postgresql','db.operation.name':'SELECT'}:{})}}));
  if(role==='mixed-positive')rows.at(-1).parentSpanId='';
  records.push(...rows);traces.push({role,traceId,records:rows});return traces.at(-1);
 }
 add('legacy-positive',keys.legacy);add('mixed-positive',keys.current,keys.legacy);add('same-span',keys.current,keys.current,true);
 for(let i=0;i<12;i++)add('current-positive-'+i,keys.mixed);
 add('split-policy-negative',keys.legacy,keys.wrong);add('shadow-negative',keys.wrong);add('current-empty',keys.empty);add('legacy-empty',keys.legacyEmpty);add('unreported',keys.unreported);
 const selector=(policy,mode,value)=>({version:'goo/environment-selector/v1',policy,scope:{mode,...(mode==='named'?{value}: {})}});
 const profiles=[{name:'effective',selector:selector('effective','named',environment),expected:15,origin:'activity'}, {name:'current',selector:selector('current-only','named',environment),expected:13,origin:'fixed'}, {name:'empty',selector:selector('effective','empty'),expected:2,origin:'activity'}, {name:'unreported',selector:selector('effective','unreported'),expected:1,origin:'activity'}];
 return {version:'goo.typedTraceFixture340',token,environment,clock,fromMs:clock-30000,toMs:clock+1000,service,operation,records,traces,profiles};
}
function payload(f){return {resourceSpans:f.records.map(r=>({resource:{attributes:attrs(r.resource)},scopeSpans:[{scope:{name:'native.typed.trace340'},spans:[{traceId:r.traceId,spanId:r.spanId,...(r.parentSpanId?{parentSpanId:r.parentSpanId}:{}),name:r.operation,kind:r.kind,startTimeUnixNano:r.startNs,endTimeUnixNano:String(BigInt(r.startNs)+BigInt(r.durationNs)),status:{code:r.status},attributes:attrs(r.attributes)}]}]}))}}
function logs(f){return {resourceLogs:f.traces.map(t=>{const r=t.records[0];return {resource:{attributes:attrs(r.resource)},scopeLogs:[{scope:{name:'native.typed.trace340'},logRecords:[{timeUnixNano:r.startNs,traceId:r.traceId,spanId:r.spanId,severityText:'INFO',body:{stringValue:'trace340 '+t.role},attributes:attrs({'test.run':f.token})}]}]}})}}
function selected(resource,s){const current=Object.hasOwn(resource,CURRENT),legacy=Object.hasOwn(resource,LEGACY);if(s.policy==='current-only')return current&&resource[CURRENT]===s.scope.value; if(!current&&!legacy)return s.scope.mode==='unreported';const v=current?resource[CURRENT]:resource[LEGACY];return s.scope.mode==='empty'?v==='':s.scope.mode==='named'&&v===s.scope.value}
function field(r,f){if(f.kind==='resource')return Object.hasOwn(r.resource,f.key)?r.resource[f.key]:null;if(f.kind==='attribute')return Object.hasOwn(r.attributes,f.key)?r.attributes[f.key]:null;return ({service:r.service,name:r.operation,trace_id:r.traceId,span_id:r.spanId,status_class:r.status===2?'error':r.status===1?'ok':'unset',kind:['Unspecified','Internal','Server','Client','Producer','Consumer'][r.kind],timestamp_ns:r.startNs,duration_ns:r.durationNs})[f.key]??null}
function matches(r,p){if(p.all)return p.all.every(x=>matches(r,x));if(p.any)return p.any.some(x=>matches(r,x));const v=field(r,p.field),x=p.value??'';return p.op==='eq'?v===x:p.op==='ne'?v!==null&&v!==x:p.op==='exists'?v!==null:p.op==='missing'?v===null:p.op==='gte'?v!==null&&BigInt(v)>=BigInt(x):p.op==='lte'?v!==null&&BigInt(v)<=BigInt(x):false}
function oracle(f,q){const groups=new Map();for(const r of f.records){if(!selected(r.resource,q.environmentSelector)||BigInt(r.startNs)<BigInt(q.fromMs)*1000000n||BigInt(r.startNs)>=BigInt(q.toMs)*1000000n)continue;if(!groups.has(r.traceId))groups.set(r.traceId,[]);groups.get(r.traceId).push(r)}return [...groups].filter(([,rs])=>q.conditions.every(p=>rs.some(r=>matches(r,p)))).sort(([a],[b])=>a<b?-1:a>b?1:0)}
module.exports={fixture,payload,logs,selected,field,matches,oracle,CURRENT,LEGACY};
