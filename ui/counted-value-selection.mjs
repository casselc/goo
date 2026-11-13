/* Atomic counted-value intent only. No source reads or selection persistence. */
import {array} from './intent-shape.mjs';
import {validate,validateField} from './predicate-validation.mjs';

function leaf(field,value,signal){
 validateField(field,signal);
 if(field.kind==='column'&&['duration_ns','timestamp_ns','severity_number'].includes(field.key))throw Error('Counted values require a categorical string field.');
 if(value!==null&&typeof value!=='string')throw Error('Choose a reported string value or the missing key.');
 const condition={field:structuredClone(field),op:value===null?'missing':'eq',...(value===null?{}:{value})};
 return validate(condition,signal);
}
export function valueReason(field,value,signal='logs'){
 try{leaf(field,value,signal);return '';}catch(error){return error.message;}
}
export function predicate(where,field,values,signal='logs'){
 const retained=validate(where,signal);
 if(!Array.isArray(values)||values.length===0||values.length>32)throw Error('Choose at least one counted value within the predicate limit.');
 array(values);
 const seen=new Set(),leaves=values.map(value=>{
  const condition=leaf(field,value,signal);
  if(seen.has(value))throw Error('Choose each counted value only once.');
  seen.add(value);return condition;
 });
 const selected=leaves.length===1?leaves[0]:{any:leaves};
 const combined=retained===undefined?selected:{all:[...(retained.all||[retained]),selected]};
 // The retained question consumes the same node/depth budget as the new OR.
 // Refusal happens before any caller replaces its staged predicate.
 return validate(combined,signal);
}
