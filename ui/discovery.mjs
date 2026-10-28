/* Pure context and search contracts. Time proximity never implies identity. */

const global=globalThis;
function millisecondsToNS(value){
  if(!/^(?:0|[1-9]\d*)(?:\.\d{1,6})?$/.test(value))throw Error('Duration requires non-negative milliseconds with at most six decimal places.');
  const [whole,fraction='']=value.split('.'),ns=BigInt(whole)*1000000n+BigInt(fraction.padEnd(6,'0'));
  if(ns>18446744073709551615n)throw Error('Duration exceeds uint64 nanoseconds.');return String(ns);
 }
function surroundingQuestion(origin,row,radiusMS=30000,offset=0){
  if(!Number.isSafeInteger(radiusMS)||radiusMS<1000||radiusMS>300000)throw Error('Context radius must be 1–300 seconds.');
  if(typeof row.timestamp_ns!=='string'||!/^\d+$/.test(row.timestamp_ns)||typeof row.ServiceName!=='string')throw Error('Captured log timestamp and service required.');
  const center=BigInt(row.timestamp_ns)/1000000n;if(center<BigInt(origin.fromMs)||center>=BigInt(origin.toMs))throw Error('Anchor is outside the captured question window.');
  const fromMS=Math.max(origin.fromMs,Number(center)-radiusMS),toMS=Math.min(origin.toMs,Number(center)+radiusMS+1);
  if(!Number.isSafeInteger(fromMS)||!Number.isSafeInteger(toMS)||fromMS>=toMS)throw Error('Anchor is outside the captured question window.');
  return {signal:'logs',environment:origin.environment,fromMs:fromMS,toMs:toMS,where:{field:{kind:'column',key:'service'},op:'eq',value:row.ServiceName},limit:100,offset};
 }
function severity(value){const s=String(value||'').toUpperCase();return /^(FATAL|CRITICAL|EMERG|ALERT)/.test(s)?'fatal':/^(ERROR|ERR)/.test(s)?'error':/^WARN/.test(s)?'warn':/^INFO/.test(s)?'info':/^(DEBUG|TRACE)/.test(s)?'debug':'unset'}
function searchPredicates(signal,{service='',message='',level='',minimum='',errors=false}={}){
  const result=[],add=(key,op,value)=>result.push({field:{kind:'column',key},op,value});
  if(service!=='')add('service','eq',service);
  if(signal==='logs'){if(message!=='')add('body','contains',message);if(level!=='')add('severity','eq',level)}
  else if(signal==='spans'){if(minimum!=='')add('duration_ns','gte',millisecondsToNS(minimum));if(errors)add('status','eq','Error')}
  else throw Error('Choose spans or logs.');return result;
 }
function sameLog(a,b){if(!a||!b)return false;const stable=o=>Object.entries(o||{}).sort(([a],[b])=>a.localeCompare(b));const identity=r=>JSON.stringify([r.timestamp_ns,r.ServiceName,r.TraceId,r.SpanId,r.Body,r.SeverityText,stable(r.ResourceAttributes),stable(r.Attributes)]);return identity(a)===identity(b)}
const durationOrder=['[0,1) ms','[1,5) ms','[5,10) ms','[10,50) ms','[50,100) ms','[100,500) ms','[500,1000) ms','[1000,+inf) ms'];

export {durationOrder,sameLog,millisecondsToNS,surroundingQuestion,severity,searchPredicates};
