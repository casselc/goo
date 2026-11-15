/* Pure reported-link tuple admission. A match describes a span record, not a
 * delivery, unique invocation, or count of its possibly repeated link entries. */
import {object,array} from './intent-shape.mjs';
const canonical=(value,length)=>typeof value==='string'&&value.length===length&&/^[0-9a-f]+$/.test(value)&&!/^0+$/.test(value);
export function validateTarget(target){
 object(target,['traceId','spanId'],'Choose one exact reported trace/span pair.');
 if(!canonical(target.traceId,32)||!canonical(target.spanId,16))throw Error('Reported links require exact nonzero lowercase trace and span IDs.');
 return {traceId:target.traceId,spanId:target.spanId};
}
function attributes(value){
 try{object(value,Object.keys(value));return Object.values(value).every(item=>typeof item==='string')}catch{return false}
}
export function matches(row,target){
 const wanted=validateTarget(target);
 try{
  object(row,Object.keys(row));const arrays=[row.linked_traces,row.linked_spans,row.linked_attributes];
  for(const values of arrays)array(values);
  if(arrays.some(values=>values.length!==arrays[0].length))return false;
  return arrays[0].some((traceId,index)=>traceId===wanted.traceId&&arrays[1][index]===wanted.spanId&&attributes(arrays[2][index]));
 }catch{return false}
}
