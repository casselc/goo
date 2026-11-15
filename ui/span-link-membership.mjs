/* Pure reported-link tuple admission. A match describes a span record, not a
 * delivery, unique invocation, or count of its possibly repeated link entries. */
import {object,array} from './intent-shape.mjs';
const canonical=(value,length)=>typeof value==='string'&&value.length===length&&/^[0-9a-f]+$/.test(value)&&!/^0+$/.test(value);
export function validateTraceId(value){if(!canonical(value,32))throw Error('A canonical nonzero lowercase trace ID is required.');return value}
export function validateSpanId(value){if(!canonical(value,16))throw Error('A canonical nonzero lowercase span ID is required.');return value}
export function validateTarget(target){
 object(target,['traceId','spanId'],'Choose one exact reported trace/span pair.');
 return {traceId:validateTraceId(target.traceId),spanId:validateSpanId(target.spanId)};
}
function attributes(value){
 try{object(value,Object.keys(value));return Object.values(value).every(item=>typeof item==='string')}catch{return false}
}
function linkArrays(row){
 object(row,Object.keys(row));const arrays=[row.linked_traces,row.linked_spans,row.linked_attributes];
 for(const values of arrays)array(values);
 if(arrays.some(values=>values.length!==arrays[0].length))throw Error('Reported link arrays must have matching trace, span and attribute positions.');
 return arrays;
}
export function linkAt(row,index){
 const arrays=linkArrays(row);
 if(!Number.isInteger(index)||index<0||index>=arrays[0].length||arrays.some(values=>!Object.hasOwn(values,index)))throw Error('Select one reported link position.');
 const target=validateTarget({traceId:arrays[0][index],spanId:arrays[1][index]});
 if(!attributes(arrays[2][index]))throw Error('Reported link attributes require a string map.');
 return {target,attributes:structuredClone(arrays[2][index])};
}
export function matches(row,target){
 const wanted=validateTarget(target);
 try{
  const arrays=linkArrays(row);
  return arrays[0].some((traceId,index)=>traceId===wanted.traceId&&arrays[1][index]===wanted.spanId&&attributes(arrays[2][index]));
 }catch{return false}
}
