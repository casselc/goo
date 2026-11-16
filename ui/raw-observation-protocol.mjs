/* Pure raw-point protocol admission. No source acquisition or navigation. */
import * as Environment from './record-environment.mjs';
import {scalarValue} from './raw-observation-values.mjs';
import * as HistogramObservations from './raw-histogram-values.mjs';
import {profile} from './raw-observation-profiles.mjs';
import {object,text} from './intent-shape.mjs';
import {isSumKind as isSum} from './instrument-reporting.mjs';
const has=(o,k)=>Object.hasOwn(o,k);
const fail=()=>{throw Error('Raw observation response does not match this read.');};
const stable=v=>JSON.stringify(v,(_,x)=>x&&typeof x==='object'&&!Array.isArray(x)?Object.fromEntries(Object.entries(x).sort(([a],[b])=>a.localeCompare(b))):x);
export function checkedRequest(input){
 object(input,['question','instrument']);const q=input.question,i=input.instrument;object(q,['environment','environmentSelector','namespace','service','fromMs','toMs']);Environment.requireNarrow(q);
 if(!text(q.service,128,{empty:false})||!Number.isSafeInteger(q.fromMs)||!Number.isSafeInteger(q.toMs)||q.fromMs<0||q.toMs<=q.fromMs||q.toMs-q.fromMs>7200000||q.toMs>=4102444800000)fail();
 if(q.namespace!=null){object(q.namespace,['mode','value']);if(!['missing','value'].includes(q.namespace.mode)||!text(q.namespace.value??'',128)||q.namespace.mode==='missing'&&(q.namespace.value??'')!=='')fail();}
 const sum=isSum(i.kind),histogram=i.kind==='histogram';object(i,['kind','scope','version','schemaURL','metric','unit',...(sum?['temporality','monotonic']:histogram?['temporality']:[])]);if(i.kind!=='gauge'&&!sum&&!histogram||!text(i.metric,256,{empty:false})||!text(i.scope,256)||!text(i.version,128)||!text(i.schemaURL,1024)||!text(i.unit,64))fail();
 if(histogram&&(!has(i,'temporality')||!Number.isInteger(i.temporality)||i.temporality< -2147483648||i.temporality>2147483647))fail();
 if(sum&&(!has(i,'temporality')||!Number.isInteger(i.temporality)||i.temporality< -2147483648||i.temporality>2147483647||typeof i.monotonic!=='boolean'||(i.kind==='stock')!==(i.temporality===2&&!i.monotonic)))fail();
 return structuredClone({question:{...Environment.questionFields(q),...(q.namespace?{namespace:q.namespace}:{}),service:q.service,fromMs:q.fromMs,toMs:q.toMs},instrument:i});
}
function questionKey(q){return stable({...q,environment:q.environment||undefined,namespace:q.namespace?{mode:q.namespace.mode,...(q.namespace.value?{value:q.namespace.value}:{})}:undefined});}
function map(value){
 if(!value||typeof value!=='object'||![Object.prototype,null].includes(Object.getPrototypeOf(value)))fail();
 for(const key of Reflect.ownKeys(value)){const d=Object.getOwnPropertyDescriptor(value,key);if(typeof key!=='string'||!d.enumerable||!has(d,'value')||typeof d.value!=='string'||!key.isWellFormed()||!d.value.isWellFormed())fail();}
}
function timestamp(value){
 if(value===null||value===undefined)return {value:null,state:'absent'};
 if(typeof value!=='string')return {value:null,state:'invalid'};
 if(!/^(0|[1-9]\d*)$/.test(value)||BigInt(value)>9223372036854775807n)return {value,state:'invalid'};
 return {value,state:'valid'};
}

export function checkedResult(value,request){
 const expected=checkedRequest(request);object(value,['version','question','instrument','readAt','rows','limit']);
 const encoded=JSON.stringify(value).replace(/[<>&\u2028\u2029]/g,c=>'\\u'+c.charCodeAt(0).toString(16).padStart(4,'0'));
 if(new TextEncoder().encode(encoded).length>8*1024*1024||value.version!==profile(expected.instrument.kind).version||value.limit!==10000||!Array.isArray(value.rows)||value.rows.length>value.limit||typeof value.readAt!=='string'||!/^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(?:\.\d{1,9})?Z$/.test(value.readAt)||!Number.isFinite(Date.parse(value.readAt))||new Date(value.readAt).toISOString().slice(0,19)!==value.readAt.slice(0,19))fail();
 const echo=checkedRequest({question:value.question,instrument:value.instrument});if(questionKey(echo.question)!==questionKey(expected.question)||stable(echo.instrument)!==stable(expected.instrument))fail();const keys=new Set();
 for(const row of value.rows){
  const histogram=expected.instrument.kind==='histogram';
  object(row,['occurrenceKey','stream','startNS','endNS','startState','flags','flagsState','rawRow','diagnostics',...(histogram?['temporality','recordingState','count','countState','bounds','boundsState','bucketCounts','bucketCountsState','bucketsState','sum','min','max']:['rawValue','value','valueState',...(isSum(expected.instrument.kind)?['temporality','monotonic']:[])])]);if(!text(row.occurrenceKey,128,{empty:false})||keys.has(row.occurrenceKey))fail();keys.add(row.occurrenceKey);
  const s=row.stream,i=expected.instrument,q=expected.question;object(s,['resourceSchemaURL','resource','scope','version','schemaURL','name','unit','attributes']);map(s.resource);map(s.attributes);
  if(typeof s.resourceSchemaURL!=='string'||s.name!==i.metric||s.unit!==i.unit||s.scope!==i.scope||s.version!==i.version||s.schemaURL!==i.schemaURL||s.resource['service.name']!==q.service||!Environment.matches(Environment.requireNarrow(q),s.resource))fail();
  if(q.namespace&&(q.namespace.mode==='missing'?has(s.resource,'service.namespace'):!has(s.resource,'service.namespace')||s.resource['service.namespace']!==(q.namespace.value||'')))fail();
  if(!row.rawRow||typeof row.rawRow!=='object'||Array.isArray(row.rawRow)||!Array.isArray(row.diagnostics))fail();const raw=row.rawRow;
  for(const [wire,key]of [['resource','resource'],['attributes','attributes'],['resourceSchemaURL','resource_schema_url'],['scope','scope'],['version','version'],['schemaURL','schema_url'],['name','name'],['unit','unit']])if(!has(raw,key)||stable(s[wire])!==stable(raw[key]))fail();
  const end=timestamp(raw.end_ns);if(end.state!=='valid'||row.endNS!==end.value||BigInt(row.endNS)<BigInt(q.fromMs)*1000000n||BigInt(row.endNS)>=BigInt(q.toMs)*1000000n)fail();
  const diagnostics=[],add=(field,code)=>diagnostics.push({field,code}),start=timestamp(raw.start_ns);if(start.state!=='valid')add('startNS',start.state);else if(BigInt(start.value)>BigInt(end.value)){start.state='invalid';add('startNS','after-observation');}
  if(row.startState!==start.state||row.startNS!==start.value)fail();
  if(histogram){
   let flags=null,flagsState='absent';if(raw.flags!==null&&raw.flags!==undefined){if(!Number.isInteger(raw.flags)||raw.flags<0||raw.flags>4294967295||row.flagsState==='invalid'&&typeof raw.flags==='number')flagsState='invalid';else{flags=raw.flags;flagsState='valid';if((BigInt(flags)&~1n)!==0n)add('flags','unknown-bits')}}
   if(flagsState!=='valid')add('flags',flagsState);if(row.flagsState!==flagsState||row.flags!==flags||row.temporality!==i.temporality||raw.temporality!==i.temporality)fail();
   if(i.temporality===0)add('temporality','unspecified');else if(i.temporality!==1&&i.temporality!==2)add('temporality','unrecognized');
   const projected=HistogramObservations.facts(raw,{flags,flagsState});for(const field of ['sum','min','max'])object(row[field],['rawValue','value','valueState','presenceState']);
   for(const [key,value]of Object.entries(projected.result))if(stable(row[key])!==stable(value))fail();
   if(stable(row.diagnostics)!==stable([...diagnostics,...projected.diagnostics]))fail();continue;
  }
  if(stable(row.rawValue??null)!==stable(raw.value??null))fail();const scalar=scalarValue(raw.value);if(scalar.state!=='finite')add('value',scalar.state);
  // JSON.parse may round a fractional flags token to an integer. Retain the
  // source adapter's conservative invalid state; rounded bits cannot prove flags.
  let flags=null,flagsState='absent';if(raw.flags!==null&&raw.flags!==undefined){if(!Number.isInteger(raw.flags)||raw.flags<0||raw.flags>4294967295||row.flagsState==='invalid'&&typeof raw.flags==='number')flagsState='invalid';else {flags=raw.flags;flagsState='valid';if((BigInt(flags)&~1n)!==0n)add('flags','unknown-bits');if((flags&1)!==0){scalar.value=null;scalar.state='absent';add('value','no-recorded-value');}}}
  if(flagsState!=='valid')add('flags',flagsState);
  if(isSum(i.kind)){if(row.temporality!==i.temporality||row.monotonic!==i.monotonic||raw.temporality!==i.temporality||raw.monotonic!==i.monotonic)fail();if(i.temporality===0)add('temporality','unspecified');else if(i.temporality!==1&&i.temporality!==2)add('temporality','unrecognized');}
  if(row.valueState!==scalar.state||row.value!==scalar.value||row.flagsState!==flagsState||row.flags!==flags||stable(row.diagnostics)!==stable(diagnostics))fail();
 }
 return structuredClone(value);
}
