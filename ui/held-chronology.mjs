/* Passive held observations from independent Activity lanes, never an aggregate. */
import {object,text,array} from './intent-shape.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
import * as Environment from './record-environment.mjs';
import {describe,findQuery,matchesLoaded,validID,contextRecords} from './record-presentation.mjs';
export const limits=Object.freeze({maxRowsPerLane:100,maxBytes:8*1024*1024,maxNodes:250000,maxDepth:32});
const signals=['logs','spans'],laneFields=['selector','phase','staged','find','readWindow','completedRead','connection','evicted','gap','rows'],itemFields=['key','record','readWindow','observedAt','limited','late'];
const admitted=new WeakSet(),encoder=new TextEncoder(),own=(v,k)=>Object.hasOwn(v,k),fail=message=>{throw Error(message||'Invalid held Activity observations.');};
function selected(input,keys){
 if(!input||typeof input!=='object'||Array.isArray(input)||![Object.prototype,null].includes(Object.getPrototypeOf(input)))fail();
 const out={};for(const key of keys){const d=Object.getOwnPropertyDescriptor(input,key);if(!d||!d.enumerable||!own(d,'value'))fail('Held observation fields must be plain data.');out[key]=d.value}return out;
}
function immutable(input){
 let nodes=0,bytes=0;const seen=new Set();
 function copy(value,depth){
  if(++nodes>limits.maxNodes||depth>limits.maxDepth)fail('Held observations exceed the supplied-data budget.');
  if(value===null||typeof value==='boolean')return value;
  if(typeof value==='number'){if(!Number.isFinite(value))fail();return value}
  if(typeof value==='string'){if(!value.isWellFormed()||(bytes+=encoder.encode(value).length)>limits.maxBytes)fail('Held observations exceed the supplied-data budget.');return value}
  if(!value||typeof value!=='object'||seen.has(value)||Object.getOwnPropertySymbols(value).length)fail('Held observations require acyclic plain JSON data.');
  const list=Array.isArray(value);if(list)array(value);else if(![Object.prototype,null].includes(Object.getPrototypeOf(value)))fail();
  const descriptors=Object.getOwnPropertyDescriptors(value),out=list?[]:{};seen.add(value);
  for(const [key,d]of Object.entries(descriptors)){if(list&&key==='length')continue;if(!d.enumerable||!own(d,'value')||!key.isWellFormed())fail('Held observations require data properties.');bytes+=encoder.encode(key).length;if(bytes>limits.maxBytes)fail('Held observations exceed the supplied-data budget.');Object.defineProperty(out,key,{value:copy(d.value,depth+1),enumerable:true,writable:false,configurable:false})}
  seen.delete(value);return Object.freeze(out);
 }
 return copy(input,0);
}
function closed(value,keys){object(value,keys);if(Object.keys(value).length!==keys.length)fail('Held observation fields are incomplete.');}
const natural=value=>Number.isSafeInteger(value)&&value>=0;
function window(value){if(value===null)return;closed(value,['fromMs','toMs']);if(!natural(value.fromMs)||!natural(value.toMs)||value.toMs-value.fromMs!==60000||value.toMs>=4102444800000)fail('Original Activity read window unavailable.');}
function clock(value){if(!text(value,40,{empty:false})||!/^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(?:\.\d{1,9})?Z$/.test(value)||!Number.isFinite(Date.parse(value))||new Date(value).toISOString().slice(0,19)!==value.slice(0,19))fail('Original Activity read clock unavailable.');}
function validate(data){
 closed(data,['version','lanes']);if(data.version!=='goo.held-activity/1')fail('Unsupported held Activity profile.');closed(data.lanes,signals);
 for(const signal of signals){
  const lane=data.lanes[signal];closed(lane,laneFields);closed(lane.find,['text','revision']);findQuery(lane.find.text);if(!natural(lane.find.revision))fail();
  if(!['stopped','loading','paused','unknown','partial','stale','unavailable'].includes(lane.phase)||typeof lane.staged!=='boolean'||typeof lane.gap!=='boolean'||!natural(lane.evicted)||lane.connection!==null&&!/^[a-f0-9]{64}$/.test(lane.connection))fail();
  array(lane.rows);if(lane.rows.length>limits.maxRowsPerLane)fail('Held lane exceeds its row budget.');
  if(lane.selector===null){
   if(!['stopped','paused'].includes(lane.phase)||lane.rows.length||lane.readWindow!==null||lane.completedRead!==null||lane.connection!==null||lane.evicted!==0||lane.gap||lane.find.text!=='')fail('Unconfigured lanes cannot carry observation history.');
   continue;
  }
  object(lane.selector,['signal','environment','environmentSelector','where']);if((own(lane.selector,'signal')?lane.selector.signal:'logs')!==signal)fail('Held lane signal differs from its source selector.');Environment.requireNarrow(lane.selector);validatePredicate(lane.selector.where,signal);
  window(lane.readWindow);
  if(lane.completedRead!==null){const r=lane.completedRead;closed(r,['fromMs','toMs','rowCount','limited','observedAt','revision']);window({fromMs:r.fromMs,toMs:r.toMs});clock(r.observedAt);if(Date.parse(r.observedAt)<r.toMs||!natural(r.rowCount)||typeof r.limited!=='boolean'||!natural(r.revision)||r.revision===0)fail();if(!lane.readWindow||r.fromMs!==lane.readWindow.fromMs||r.toMs!==lane.readWindow.toMs)fail('Latest read facts disagree.');}
  const keys=new Set();
  for(const item of lane.rows){closed(item,itemFields);if(typeof item.key!=='string'||!/^[a-f0-9]{64}:\d+$/.test(item.key)||keys.has(item.key))fail('Held occurrence identity is missing or duplicated.');keys.add(item.key);objectRecord(item.record);if(typeof item.record[signal==='logs'?'Body':'SpanName']!=='string')fail('Held record differs from its lane signal.');if(item.readWindow===null)fail();window(item.readWindow);clock(item.observedAt);if(Date.parse(item.observedAt)<item.readWindow.toMs||typeof item.limited!=='boolean'||typeof item.late!=='boolean')fail();}
 }
 return data;
}
function objectRecord(value){if(!value||typeof value!=='object'||Array.isArray(value)||![Object.prototype,null].includes(Object.getPrototypeOf(value)))fail();}
export function prepare(lanes){
 object(lanes,signals);if(!signals.every(signal=>own(lanes,signal)))fail('Choose both held lane descriptors.');
 const projected={};for(const signal of signals){const lane=selected(lanes[signal],laneFields),rows=lane.rows;array(rows);if(rows.length>limits.maxRowsPerLane)fail('Held lane exceeds its row budget.');projected[signal]={...lane,rows:rows.map(item=>selected(item,itemFields))};}
 return adopt({version:'goo.held-activity/1',lanes:projected});
}
export function adopt(value){if(value&&typeof value==='object'&&admitted.has(value))return value;const data=validate(immutable(value));admitted.add(data);return data;}
export function model(value,{query=''}={}){
 const data=adopt(value);findQuery(query);const rows=[],counts={},matched={};
 for(const signal of signals){const lane=data.lanes[signal];counts[signal]=lane.rows.length;matched[signal]=0;for(const [index,item]of lane.rows.entries()){
  if(!matchesLoaded(item.record,signal,[],query))continue;matched[signal]++;const time=describe(item.record,signal).time,placed=!!time;
  rows.push({key:signal+':'+item.key,signal,item,index,placed,timeNS:placed?item.record.timestamp_ns:null});
 }}
 rows.sort((a,b)=>{if(a.placed!==b.placed)return a.placed?-1:1;if(a.placed){const x=BigInt(a.timeNS),y=BigInt(b.timeNS);if(x!==y)return x>y?-1:1}return signals.indexOf(a.signal)-signals.indexOf(b.signal)||a.index-b.index});
 return {data,rows,placed:rows.filter(row=>row.placed),unplaced:rows.filter(row=>!row.placed),counts,matched,query};
}

// Exact reported correlation over held occurrences; never a trace-completeness claim.
export function related(value,{signal,key,spanOnly=false}={}){
 const data=adopt(value);
 if(!signals.includes(signal)||typeof key!=='string'||typeof spanOnly!=='boolean')fail('Choose a held occurrence.');
 const item=data.lanes[signal].rows.find(item=>item.key===key);
 if(!item||!validID(item.record.TraceId,32))fail('A valid reported trace ID is required.');
 const traceId=item.record.TraceId,spanId=validID(item.record.SpanId,16)?item.record.SpanId:null;
 if(spanOnly&&!spanId)fail('A valid reported span ID is required.');
 const current=model(data),rows=current.rows.filter(entry=>entry.item.record.TraceId===traceId&&(!spanOnly||entry.item.record.SpanId===spanId));
 const matched=Object.fromEntries(signals.map(signal=>[signal,rows.filter(entry=>entry.signal===signal).length]));
 return {...current,rows,placed:rows.filter(row=>row.placed),unplaced:rows.filter(row=>!row.placed),matched,traceId,spanId,spanOnly};
}

const parentProjections=new WeakMap();
export function parentRecords(value){const data=adopt(value);let projection=parentProjections.get(data);if(!projection){projection=contextRecords(data.lanes.spans.rows.map(item=>item.record));parentProjections.set(data,projection)}return projection}
