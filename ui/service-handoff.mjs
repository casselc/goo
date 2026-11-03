/* Checked navigation context only. No evidence, reads, profile selection or return callbacks. */
import * as Environment from './record-environment.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
const version='goo.service-handoff/v1',own=(o,k)=>Object.prototype.hasOwnProperty.call(o,k),plain=x=>x&&typeof x==='object'&&!Array.isArray(x)&&[Object.prototype,null].includes(Object.getPrototypeOf(x));
function closed(x,keys){if(!plain(x)||Object.keys(x).some(k=>!keys.includes(k)))throw Error('Invalid service context fields.');}
function text(x,empty=false){return typeof x==='string'&&(empty||x.length>0)&&x.isWellFormed()&&!x.includes('\0')&&new TextEncoder().encode(x).length<=128;}
function namespace(x){closed(x,['mode','value']);if(!['missing','value'].includes(x.mode)||x.mode==='missing'&&own(x,'value')||x.mode==='value'&&!text(x.value,true))throw Error('Service namespace must have an exact reported value or a known missing key.');return structuredClone(x);}
function origin(q){validateRecordQuestion(q);if(q.where)validatePredicate(q.where,q.signal);if(q.toMs>=4102444800000)throw Error('This observation window is outside service discovery bounds.');return Environment.selection(q);}
function freeze(x){if(x&&typeof x==='object'){Object.values(x).forEach(freeze);Object.freeze(x)}return x;}
export function validate(value){
 closed(value,['version','service','namespace','environmentSelector','fromMs','toMs','origin']);if(value.version!==version||!text(value.service))throw Error('Open service requires a nonempty reported service name (up to 128 bytes).');
 const selected=origin(value.origin);if(!Environment.equal(selected,value.environmentSelector)||value.fromMs!==value.origin.fromMs||value.toMs!==value.origin.toMs)throw Error('Service context must retain the observation environment policy and completed window.');namespace(value.namespace);
 return freeze(structuredClone(value));
}
export function create(question,row){
 const selected=origin(question);if(!plain(row)||!text(row.ServiceName))throw Error('Open service requires a nonempty reported service name (up to 128 bytes).');
 if(!own(row,'ResourceAttributes')||!plain(row.ResourceAttributes))throw Error('Service namespace reporting is unknown for this observation. Use manual service discovery.');
 const resource=row.ResourceAttributes,ns=own(resource,'service.namespace')?{mode:'value',value:resource['service.namespace']}:{mode:'missing'};namespace(ns);
 if(!Environment.matches(selected,resource))throw Error('Observed resource environment does not match this Activity scope.');
 if(typeof row.timestamp_ns!=='string'||row.timestamp_ns.length>20||!/^(0|[1-9][0-9]*)$/.test(row.timestamp_ns)||BigInt(row.timestamp_ns)<BigInt(question.fromMs)*1000000n||BigInt(row.timestamp_ns)>=BigInt(question.toMs)*1000000n)throw Error('This observation is outside its completed read window.');
 return validate({version,service:row.ServiceName,namespace:ns,environmentSelector:selected,fromMs:question.fromMs,toMs:question.toMs,origin:structuredClone(question)});
}
export function reason(question,row){try{create(question,row);return ''}catch(error){return error.message}}
export function scope(value){const checked=validate(value);return {...Environment.questionFields(checked.origin),namespace:structuredClone(checked.namespace),fromMs:checked.fromMs,toMs:checked.toMs};}
