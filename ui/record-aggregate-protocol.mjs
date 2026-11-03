/* Raw record-aggregate DTO admission only. No rows, datasets or source witnesses.
 * Returns the caller-owned transport object, as the service adapter did; cohort
 * construction owns any later copy. Validation grants no snapshot capability. */
import {identity} from './question-session.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
// Capability preflight only; server compilation remains authoritative.
export function canRead(q){try{validateRecordQuestion(q);if(q.groupBy?.length!==1||(q.offset??0)!==0||q.numeric||q.numericHistogram||q.groupTimeSeries||q.summary&&!['span-duration'].includes(q.summary)||q.signal!=='spans'&&(q.summary||q.spanStatus))return false;const field=q.groupBy[0];validatePredicate({field,op:'exists'},q.signal);if(q.where)validatePredicate(q.where,q.signal);return !(field.kind==='column'&&['duration_ns','timestamp_ns','severity_number'].includes(field.key));}catch{return false}}
function closed(value,keys){if(!value||typeof value!=='object'||Array.isArray(value)||![Object.prototype,null].includes(Object.getPrototypeOf(value)))throw Error('Invalid record aggregates: expected a plain DTO');for(const key of Reflect.ownKeys(value)){const d=Object.getOwnPropertyDescriptor(value,key);if(typeof key!=='string'||!keys.includes(key)||!d?.enumerable||!Object.hasOwn(d,'value'))throw Error('Invalid record aggregates: unexpected DTO field');}}
export function recordAggregates(q,data){
  const fail=message=>{throw Error('Invalid record aggregates: '+message)},uint=(v,bits=64)=>typeof v==='string'&&!/\s/u.test(v)&&/^(0|[1-9][0-9]*)$/.test(v)&&BigInt(v)<(1n<<BigInt(bits));
  closed(data,['version','question','capturedAt','basis','count','state','groups','duration','spanStatus','timeSeries','warnings','reads']);
  if(data?.version!=='goo/record-aggregates/v1')fail('unsupported response version');try{if(!q||identity(q)!==identity(data.question))fail('question differs')}catch{fail('question differs')}
  if(!['observed','no_data','changed_between_reads'].includes(data.state)||!uint(data.count)||typeof data.basis!=='string'||!/^[0-9a-f]{64}$/.test(data.basis)||data.basis.length!==64||!Number.isFinite(Date.parse(data.capturedAt))||!Array.isArray(data.warnings)||data.warnings.some(w=>typeof w!=='string'))fail('state or provenance');
  for(const key of ['rows','fields','hasMore','resultToken','plan'])if(Object.hasOwn(data,key))fail('inspection payload is not an aggregate projection');
  closed(data.reads,['count','groups',...(q.bucketMs?['timeSeries']:[])]);
  const receiptKeys=['count','groups',...(q.bucketMs?['timeSeries']:[])];for(const key of receiptKeys){const r=data.reads?.[key];closed(r,['readAt','basis']);if(!r||!Number.isFinite(Date.parse(r.readAt))||typeof r.basis!=='string'||r.basis.length!==64||!/^[0-9a-f]{64}$/.test(r.basis))fail('missing '+key+' read receipt')}
  function measures(item){
   if(q.summary){const d=item.duration;closed(d,['state','count','sumNs','minNs','p50Ns','p95Ns','p99Ns','maxNs']);if(!d||!uint(d.count)||d.count!==item.count||!uint(d.sumNs,128))fail('duration summary');const n=BigInt(d.count),sum=BigInt(d.sumNs),keys=['minNs','p50Ns','p95Ns','p99Ns','maxNs'];if(n===0n){if(d.state!=='no_data'||sum!==0n||keys.some(k=>d[k]!==undefined))fail('empty duration summary')}else{if(d.state!=='observed'||keys.some(k=>!uint(d[k])))fail('duration values');const values=keys.map(k=>BigInt(d[k]));if(values.some((v,i)=>i&&v<values[i-1])||sum<values[0]*n||sum>values[4]*n)fail('duration bounds')}}else if(item.duration!==undefined)fail('unrequested duration');
   if(q.spanStatus){const d=item.spanStatus;closed(d,['count','error','ok','unset','unknown']);if(!d||d.count!==item.count||!['count','error','ok','unset','unknown'].every(k=>uint(d[k]))||['error','ok','unset','unknown'].reduce((n,k)=>n+BigInt(d[k]),0n)!==BigInt(d.count))fail('span status')}else if(item.spanStatus!==undefined)fail('unrequested span status');
  }
  measures(data);if(!Array.isArray(data.groups)||data.groups.length>1000)fail('groups');let groupCount=0n;const seen=new Set();for(const g of data.groups){closed(g,['g0','count','duration','spanStatus']);if(!g||!Object.hasOwn(g,'g0')||!(g.g0===null||typeof g.g0==='string')||seen.has(g.g0)||!uint(g.count)||g.count==='0')fail('group');seen.add(g.g0);groupCount+=BigInt(g.count);measures(g)}
  let seriesCount=0n;if(q.bucketMs){if(!Array.isArray(data.timeSeries)||data.timeSeries.length!==Math.ceil((q.toMs-q.fromMs)/q.bucketMs)||data.timeSeries.length>240)fail('time series');for(const [i,t]of data.timeSeries.entries()){closed(t,['fromMs','toMs','state','count','duration','spanStatus']);if(t.fromMs!==q.fromMs+i*q.bucketMs||t.toMs!==Math.min(t.fromMs+q.bucketMs,q.toMs)||!uint(t.count)||t.state!==(t.count==='0'?'no_data':'observed'))fail('time bucket');measures(t);seriesCount+=BigInt(t.count)}}else if(data.timeSeries!==undefined)fail('unrequested time series');
  if(data.state!=='changed_between_reads'){
   if(groupCount!==BigInt(data.count)||q.bucketMs&&seriesCount!==BigInt(data.count)||data.state!==(data.count==='0'?'no_data':'observed'))fail('population conservation');
   for(const items of [data.groups,...(q.bucketMs?[data.timeSeries]:[])]){if(q.summary&&items.reduce((n,p)=>n+BigInt(p.duration.sumNs),0n)!==BigInt(data.duration.sumNs))fail('duration conservation');if(q.spanStatus)for(const key of ['error','ok','unset','unknown'])if(items.reduce((n,p)=>n+BigInt(p.spanStatus[key]),0n)!==BigInt(data.spanStatus[key]))fail('status conservation')}
  }return data;
 }
