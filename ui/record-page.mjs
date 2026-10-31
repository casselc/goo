/* Supplied rich-record pages are bounded observations, not aggregate Datasets.
   This profile never carries completeness witnesses or execution tokens. */
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';

export const RECORD_PAGE_LIMITS=Object.freeze({maxBytes:16*1024*1024,maxNodes:250000,maxDepth:32});
const object=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
const encoder=new TextEncoder();

// Walk data descriptors, rather than reading properties, so an extension cannot
// smuggle mutable class instances, accessors or cycles into a passive component.
function immutableJSON(input){
 const path=new Set();let nodes=0,bytes=0;
 function copy(value,depth){
  if(++nodes>RECORD_PAGE_LIMITS.maxNodes||depth>RECORD_PAGE_LIMITS.maxDepth)throw Error('Record page exceeds the supplied-data budget.');
  if(value===null||typeof value==='boolean')return value;
  if(typeof value==='number'){if(!Number.isFinite(value))throw Error('Record page requires finite JSON values.');return value}
  if(typeof value==='string'){if(!value.isWellFormed())throw Error('Record page requires valid Unicode text.');if(value.length>RECORD_PAGE_LIMITS.maxBytes||(bytes+=encoder.encode(value).length)>RECORD_PAGE_LIMITS.maxBytes)throw Error('Record page exceeds the supplied-data budget.');return value}
  if(typeof value!=='object'||path.has(value))throw Error('Record page requires acyclic JSON data.');
  const array=Array.isArray(value),prototype=Object.getPrototypeOf(value);
  if(!array&&prototype!==Object.prototype&&prototype!==null)throw Error('Record page requires plain JSON objects.');
  if(Object.getOwnPropertySymbols(value).length)throw Error('Record page requires JSON keys.');
  const descriptors=Object.getOwnPropertyDescriptors(value),keys=Object.keys(descriptors);
  if(array&&(keys.length!==value.length+1||keys.some(key=>key!=='length'&&(!/^(0|[1-9][0-9]*)$/.test(key)||Number(key)>=value.length))))throw Error('Record page requires dense JSON arrays.');
  path.add(value);const out=array?[]:{};
  for(const key of keys){
   if(array&&key==='length')continue;
   const d=descriptors[key];if(!d.enumerable||!Object.hasOwn(d,'value'))throw Error('Record page requires data properties.');
   if(!key.isWellFormed())throw Error('Record page requires valid Unicode keys.');
   bytes+=encoder.encode(key).length;if(bytes>RECORD_PAGE_LIMITS.maxBytes)throw Error('Record page exceeds the supplied-data budget.');
   Object.defineProperty(out,key,{value:copy(d.value,depth+1),enumerable:true,writable:false,configurable:false});
  }
  path.delete(value);return Object.freeze(out);
 }
 return copy(input,0);
}

// Normalize the known Go response's record aspect. Rich record attributes and
// nanosecond strings remain unchanged; page length never becomes a population total.
export function fromExploration(data){
 if(!object(data)||![Object.prototype,null].includes(Object.getPrototypeOf(data)))throw Error('A plain exploration response is required.');
 const descriptors=Object.getOwnPropertyDescriptors(data),selected={};
 for(const key of ['question','basis','capturedAt','state','rows','hasMore']){
  const d=descriptors[key];if(key==='hasMore'&&!d)continue;
  if(!d||!Object.hasOwn(d,'value'))throw Error('Record-page source fields must be data properties.');
  if(key==='hasMore'&&d.value===undefined)continue;
  selected[key]=d.value;
 }
 const input=immutableJSON(selected);
 validateRecordQuestion(input.question);validatePredicate(input.question.where,input.question.signal);
 if(!/^[0-9a-f]{64}$/.test(input.basis)||typeof input.capturedAt!=='string'||!/^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(?:\.\d{1,9})?Z$/.test(input.capturedAt)||!Number.isFinite(Date.parse(input.capturedAt))||new Date(input.capturedAt).toISOString().slice(0,19)!==input.capturedAt.slice(0,19))throw Error('Record page requires its source basis and UTC read clock.');
 if(!['observed','no_data','changed_between_reads'].includes(input.state))throw Error('Unsupported record-page observation state.');
 const {limit,offset}=input.question;
 if(!Number.isInteger(limit)||limit<1||limit>200||!Number.isInteger(offset)||offset<0||offset>1000||!Array.isArray(input.rows)||input.rows.length>limit||input.rows.some(row=>!object(row)))throw Error('Record page differs from its bounded delivery question.');
 if(input.hasMore!==undefined&&typeof input.hasMore!=='boolean')throw Error('Record-page continuation must be reported or unknown.');
 return Object.freeze({version:'goo.record-page/1',signal:input.question.signal,question:input.question,basis:input.basis,capturedAt:input.capturedAt,rows:input.rows,page:Object.freeze({offset,limit,returned:input.rows.length,hasMore:input.hasMore??null}),quality:Object.freeze({extent:'bounded-page',upstream:'unknown',consistency:'independent-reads',state:input.state})});
}
