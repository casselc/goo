/* Immutable data admission. Source adapters are trusted host extension points,
   not authentication or proof that an arbitrary HTTP response is authoritative. */
import {validateSnapshot} from './view-schema.mjs';
import {validateComparison} from './view-comparison.mjs';
import {identity as questionIdentity} from './question-session.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';

export const SNAPSHOT_LIMITS=Object.freeze({maxBytes:8*1024*1024,maxNodes:250000,maxDepth:32});
export const COMPARISON_LIMITS=Object.freeze({maxBytes:16*1024*1024+65536,maxNodes:500000,maxDepth:34});
const witnesses=new WeakMap(),comparisons=new WeakSet(),encoder=new TextEncoder();
let nextSource=0,nextLocal=0;
const record=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
const own=(value,key)=>Object.prototype.hasOwnProperty.call(value,key);
const aspect=snapshot=>snapshot.shape==='time-series'?'intervals':snapshot.shape==='summary'?'summary':'groups';
export class SnapshotError extends TypeError {
 constructor(code,message){super(message);this.name='SnapshotError';this.code=code}
}
const fail=(code,message)=>{throw new SnapshotError(code,message)};

// Inspect descriptors before any clone/property traversal. Count alias appearances
// as serialized data, while the path set rejects cycles without rejecting DAGs.
function preflight(input,limits=SNAPSHOT_LIMITS,{allowUndefined=false}={}){
 const path=new Set();let bytes=0,nodes=0;
 const add=n=>{bytes+=n;if(bytes>limits.maxBytes)fail('budget-bytes','Snapshot exceeds the normalized byte budget')};
 const stringBytes=value=>{
  const remaining=limits.maxBytes-bytes;
  if(value.length>remaining)fail('budget-bytes','Snapshot exceeds the normalized byte budget');
  let n=2;
  for(let i=0;i<value.length;i++){
   const code=value.charCodeAt(i);
   if(code===34||code===92)n+=2;
   else if(code<32)n+=[8,9,10,12,13].includes(code)?2:6;
   else if(code<128)n++;
   else if(code<2048)n+=2;
   else if(code>=55296&&code<=56319&&value.charCodeAt(i+1)>=56320&&value.charCodeAt(i+1)<=57343){n+=4;i++}
   else if(code>=55296&&code<=57343)n+=6;
   else n+=3;
   if(n>remaining)fail('budget-bytes','Snapshot exceeds the normalized byte budget');
  }
  return n;
 };
 function visit(value,depth){
  if(depth>limits.maxDepth)fail('budget-depth','Snapshot exceeds the normalized depth budget');
  if(++nodes>limits.maxNodes)fail('budget-nodes','Snapshot exceeds the normalized node budget');
  if(value===null){add(4);return}
  switch(typeof value){
   case 'string':add(stringBytes(value));return;
   case 'number':if(!Number.isFinite(value))fail('non-json','Snapshot numbers must be finite JSON values');add(String(value).length);return;
   case 'boolean':add(value?4:5);return;
   case 'undefined':if(allowUndefined){add(1);return}fail('non-json','Snapshot values cannot be undefined');break;
   case 'object':break;
   default:fail('non-json','Snapshot values must be plain JSON data');
  }
  const array=Array.isArray(value),prototype=Object.getPrototypeOf(value);
  if(array?prototype!==Array.prototype:prototype!==Object.prototype&&prototype!==null)fail('prototype','Snapshot values must use plain JSON objects and arrays');
  if(path.has(value))fail('cycle','Snapshot values cannot contain cycles');
  path.add(value);add(2);
  const keys=Reflect.ownKeys(value),length=array?Object.getOwnPropertyDescriptor(value,'length')?.value:null;
  if(keys.length-(array?1:0)>limits.maxNodes-nodes)fail('budget-nodes','Snapshot exceeds the normalized node budget');
  if(array){
   if(!Number.isSafeInteger(length)||length>limits.maxNodes)fail('budget-nodes','Snapshot exceeds the normalized node budget');
   if(keys.length!==length+1)fail('non-json','Snapshot arrays must be dense with no named properties');
  }
  for(const key of keys){
   if(array&&key==='length')continue;
   if(typeof key!=='string')fail('non-json','Snapshot values cannot contain symbol properties');
   const d=Object.getOwnPropertyDescriptor(value,key);
   if(!d)fail('non-json','Snapshot properties changed during inspection');
   if(!own(d,'value'))fail('accessor','Snapshot values cannot contain getters or setters');
   if(!d.enumerable)fail('non-json','Snapshot values cannot contain hidden properties');
   if(array&&(!/^(?:0|[1-9]\d*)$/.test(key)||Number(key)>=length))fail('non-json','Snapshot arrays must use ordinary indices');
   add(stringBytes(key)+2);visit(d.value,depth+1);
  }
  path.delete(value);
 }
 try{visit(input,0)}catch(error){if(error instanceof SnapshotError)throw error;fail('non-json','Unable to inspect plain snapshot data')}
 return {bytes,nodes};
}
function freeze(value,seen=new Set()){
 if(value===null||typeof value!=='object'||seen.has(value)||Object.isFrozen(value))return value;
 seen.add(value);for(const child of Object.values(value))freeze(child,seen);return Object.freeze(value);
}
function checked(raw){const result=validateSnapshot(raw);if(!result.ok)fail('shape',result.reasons.join('; '))}
function make(raw,info,token=null){
 freeze(raw);
 const publicInfo=freeze({...info,guarantees:[...info.guarantees]});
 witnesses.set(raw,{info:publicInfo,token,guarantees:new Set(info.guarantees)});
 return raw;
}
export const isSnapshot=value=>value!==null&&typeof value==='object'&&witnesses.has(value);
export const getSnapshotInfo=value=>isSnapshot(value)?witnesses.get(value).info:null;
export const hasGuarantee=(value,guarantee)=>isSnapshot(value)&&witnesses.get(value).guarantees.has(guarantee);
function hold(raw){
 if(!record(raw))fail('shape','A dataset snapshot object is required');
 if(typeof raw.id!=='string'||!raw.id.length||raw.id.length>1024||typeof raw.revision!=='string'||!raw.revision.length||raw.revision.length>1024)fail('shape','Raw snapshots require bounded dataset identity claims');
 if(raw.origin==='source-read'&&raw.sourceBasis===undefined)raw.sourceBasis=raw.revision;
 const id='goo.local/'+ ++nextLocal;raw.id=id;raw.revision=id;
 raw.origin='browser-held';delete raw.completeness;
 if(record(raw.quality))raw.quality.extent='unknown';
 return raw;
}
export function importSnapshot(raw){
 preflight(raw);
 const copy=hold(structuredClone(raw));checked(copy);
 return make(copy,{entry:'held',sourceId:null,sourceInstance:null,parent:null,projection:null,guarantees:[]});
}
export const asSnapshot=input=>isSnapshot(input)?input:importSnapshot(input);
function options(value,allowed){
 preflight(value,SNAPSHOT_LIMITS,{allowUndefined:true});
 if(!record(value)||Object.keys(value).some(key=>!allowed.includes(key)))fail('options','Unsupported snapshot constructor options');
}
function checkedQuestion(question){
 preflight(question,SNAPSHOT_LIMITS,{allowUndefined:true});
 try{validateRecordQuestion(question);validatePredicate(question.where,question.signal)}catch(error){fail('source-question','Invalid source question: '+error.message)}
}
function sourceQuestion(raw,question){
 checkedQuestion(question);checkedQuestion(raw.scope.selector);
 if(questionIdentity(raw.scope.selector)!==questionIdentity(question))fail('source-question','Source snapshot does not match the checked requested question');
 if(raw.scope.signal!==question.signal)fail('source-question','Source snapshot signal does not match its question');
 if(raw.shape==='time-series'&&(raw.interval.fromMs!==String(question.fromMs)||raw.interval.toMs!==String(question.toMs)||raw.interval.bucketMs!==String(question.bucketMs)))fail('source-question','Source interval geometry does not match its question');
 if(raw.shape==='ranked'&&(raw.dimensions?.length??0)!==(question.groupBy?.length??0))fail('source-question','Source group dimensions do not match its question');
 for(const field of raw.fields){
  if(field.aggregation?.family==='numeric'&&(!question.numeric||field.unit!==question.numeric.unit))fail('source-question','Source numeric measure unit does not match its binding');
  if(field.aggregation?.family==='duration'&&(question.signal!=='spans'||question.summary!=='span-duration'))fail('source-question','Source duration measures require the checked duration profile');
 }
}
export function createSourceAdapter(configuration){
 options(configuration,['id']);
 const id=configuration.id;
 if(typeof id!=='string'||!id.length||id.length>128||encoder.encode(id).length>128||id.trim()!==id||/[\p{Cc}]/u.test(id))fail('options','A bounded logical source id is required');
 const token={},sourceInstance='goo.source/'+ ++nextSource;
 return Object.freeze({adoptSnapshot(raw,configuration){
  options(configuration,['question']);
  if(!own(configuration,'question'))fail('options','Source adoption requires the checked question');
  if(isSnapshot(raw)){
   const prior=witnesses.get(raw);
   if(prior.token!==token||prior.info.entry!=='source')fail('source-instance','A snapshot handle cannot be adopted by another source adapter');
   sourceQuestion(raw,configuration.question);return raw;
  }
  preflight(raw);checked(raw);
  if(raw.origin!=='source-read')fail('source-origin','Source adoption requires an explicit source-read snapshot');
  sourceQuestion(raw,configuration.question);
  const copy=structuredClone(raw),guarantees=copy.quality.extent==='complete'?[aspect(copy)]:[];
  return make(copy,{entry:'source',sourceId:id,sourceInstance,parent:null,projection:null,guarantees},token);
 }});
}
export function selectRows(snapshot,keys){
 if(!isSnapshot(snapshot))fail('snapshot-required','Row projection requires an admitted snapshot handle');
 if(!['ranked','time-series'].includes(snapshot.shape))fail('projection-shape','Population summaries cannot be projected into rows');
 preflight(keys);
 if(!Array.isArray(keys)||keys.some(key=>typeof key!=='string')||new Set(keys).size!==keys.length)fail('projection-keys','Projection keys must be distinct supplied row keys');
 const wanted=new Set(keys),available=new Set(snapshot.rows.map(row=>row.key));
 if(keys.some(key=>!available.has(key)))fail('projection-keys','Projection contains a foreign row key');
 if(keys.length===snapshot.rows.length)return snapshot;
 const id='goo.local/'+ ++nextLocal,{completeness,...rest}=snapshot;
 const projected={...rest,id,revision:id,origin:'browser-held',rows:snapshot.rows.filter(row=>wanted.has(row.key)),quality:{...snapshot.quality,extent:snapshot.quality.extent==='unknown'?'unknown':'partial'}};
 checked(projected);
 const parent=getSnapshotInfo(snapshot);
 return make(projected,{entry:'projection',sourceId:parent.sourceId,sourceInstance:parent.sourceInstance,parent:{id:snapshot.id,revision:snapshot.revision},projection:{kind:'rows',keys:projected.rows.map(row=>row.key)},guarantees:[]});
}
export function asComparison(raw){
 if(raw!==null&&typeof raw==='object'&&comparisons.has(raw))return raw;
 preflight(raw,COMPARISON_LIMITS);
 if(!record(raw))fail('shape','A comparison envelope is required');
 const baseline=asSnapshot(raw.baseline),selection=asSnapshot(raw.selection);
 const metadata={...raw};delete metadata.baseline;delete metadata.selection;
 const copy={...structuredClone(metadata),baseline,selection},result=validateComparison(copy);
 if(!result.ok)fail('shape',result.reasons.join('; '));
 freeze(copy);comparisons.add(copy);return copy;
}
function parse(text,limits,admit){
 if(typeof text!=='string')fail('non-json','Snapshot JSON must be a string');
 if(text.length>limits.maxBytes)fail('budget-bytes','Snapshot JSON exceeds the byte budget');
 let bytes=0;
 for(let i=0;i<text.length;i++){const code=text.charCodeAt(i);if(code<128)bytes++;else if(code<2048)bytes+=2;else if(code>=55296&&code<=56319&&text.charCodeAt(i+1)>=56320&&text.charCodeAt(i+1)<=57343){bytes+=4;i++}else bytes+=3;if(bytes>limits.maxBytes)fail('budget-bytes','Snapshot JSON exceeds the byte budget')}
 let raw;try{raw=JSON.parse(text)}catch{fail('non-json','Invalid snapshot JSON')}
 return admit(raw);
}
export const parseSnapshotJSON=text=>parse(text,SNAPSHOT_LIMITS,asSnapshot);
export const parseComparisonJSON=text=>parse(text,COMPARISON_LIMITS,asComparison);
