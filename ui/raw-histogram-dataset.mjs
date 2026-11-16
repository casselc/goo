/* One held source occurrence, never an interval evaluation or merged distribution. */
import {checkedRequest,checkedResult} from './raw-observation-protocol.mjs';
import {state,values,rawHistogramFields} from './raw-histogram-schema.mjs';
import {validateSnapshot} from './view-schema.mjs';
const canonical=value=>JSON.stringify(value,(_,entry)=>entry&&typeof entry==='object'&&!Array.isArray(entry)?Object.fromEntries(Object.entries(entry).sort(([a],[b])=>a.localeCompare(b))):entry);
function identity(value){let a=2166136261,b=2246822519;for(const byte of new TextEncoder().encode(canonical(value))){a=Math.imul(a^byte,16777619);b=Math.imul(b^byte,3266489917)}return (a>>>0).toString(16).padStart(8,'0')+(b>>>0).toString(16).padStart(8,'0')}
export function dataset(data,occurrence){
 const request=checkedRequest({question:data?.question,instrument:data?.instrument}),bank=checkedResult(data,request);
 if(bank.instrument.kind!=='histogram')throw Error('Bucket displays require raw histogram observations.');
 const key=typeof occurrence==='string'?occurrence:occurrence?.occurrenceKey,row=bank.rows.find(row=>row.occurrenceKey===key);
 if(!row||typeof occurrence==='object'&&canonical(row)!==canonical(occurrence))throw Error('Choose one occurrence from this held histogram bank.');
 const id='raw-histogram:'+identity({request,key}),revision='held:'+identity({request,readAt:bank.readAt,row});
 const snapshot={version:'goo.dataset/1',id,revision,origin:'source-read',capturedAt:bank.readAt,shape:'observations',scope:{id,generation:revision,signal:'metrics',selector:request},rawHistogram:{request,readAt:bank.readAt,limit:bank.limit},fields:rawHistogramFields(),rows:[{key,state:state(row),values:values(row),observation:row}],quality:{extent:'unknown',basis:'captured-query',value:'unknown',sampling:'unknown',freshness:'fixed'}};
 const admitted=validateSnapshot(snapshot);if(!admitted.ok)throw Error(admitted.reasons.join('; '));return structuredClone(snapshot);
}
