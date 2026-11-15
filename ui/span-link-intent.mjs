import {validateRecordQuestion} from './question-protocol.mjs';
import * as Environment from './record-environment.mjs';
import {validateTarget} from './span-link-membership.mjs';

/* Pure admission and planning for one reported OTel link. Held-result ownership,
 * clocks, source reads and revocable navigation capabilities belong to callers. */
const record=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
const canonicalID=(value,length)=>typeof value==='string'&&value.length===length&&/^[0-9a-f]+$/.test(value)&&!/^0+$/.test(value);
function tuple(traceId,spanId){
 if(!canonicalID(traceId,32)||!canonicalID(spanId,16))throw Error('Reported link requires exact nonzero lowercase trace and span IDs.');
 return {traceId,spanId};
}
export function inspect({rows,rowIndex,linkIndex}){
 if(!Array.isArray(rows)||!Number.isInteger(rowIndex)||rowIndex<0||rowIndex>=rows.length||!Object.hasOwn(rows,rowIndex)||!record(rows[rowIndex]))throw Error('Select one held span row.');
 const row=rows[rowIndex],source=tuple(row.TraceId,row.SpanId),arrays=[row.linked_traces,row.linked_spans,row.linked_attributes];
 if(arrays.some(a=>!Array.isArray(a))||arrays.some(a=>a.length!==arrays[0].length))throw Error('Reported link arrays must have matching trace, span and attribute positions.');
 if(!Number.isInteger(linkIndex)||linkIndex<0||linkIndex>=arrays[0].length||arrays.some(a=>!Object.hasOwn(a,linkIndex)))throw Error('Select one reported link position.');
 const target=tuple(arrays[0][linkIndex],arrays[1][linkIndex]),attributes=arrays[2][linkIndex];
 if(!record(attributes)||Object.values(attributes).some(value=>typeof value!=='string'))throw Error('Reported link attributes require a string map.');
 const rowIndexes=[];
 for(let index=0;index<rows.length;index++)if(record(rows[index])&&rows[index].TraceId===target.traceId&&rows[index].SpanId===target.spanId)rowIndexes.push(index);
 return {target,source,attributes:structuredClone(attributes),local:{state:rowIndexes.length===1?'unique':rowIndexes.length?'ambiguous':'absent',rowIndexes}};
}
export function plan({origin,rows,rowIndex,linkIndex}){
 validateRecordQuestion(origin);
 if(origin.signal!=='spans')throw Error('Reported span links require an ordinary spans question.');
 const {target,source,attributes,local}=inspect({rows,rowIndex,linkIndex});
 const question={signal:'spans',...Environment.questionFields(origin),fromMs:origin.fromMs,toMs:origin.toMs,where:{all:[{field:{kind:'column',key:'trace_id'},op:'eq',value:target.traceId},{field:{kind:'column',key:'span_id'},op:'eq',value:target.spanId}]},limit:100,offset:0};
 validateRecordQuestion(question);
 return {relation:'reported-span-link',question,target,local,origin:{question:structuredClone(origin),rowIndex,linkIndex,source},attributes:structuredClone(attributes)};
}

export function planIncoming({origin,rows,rowIndex}){
 validateRecordQuestion(origin);
 if(origin.signal!=='spans')throw Error('Reported span-link discovery requires an ordinary spans question.');
 if(!Array.isArray(rows)||!Number.isInteger(rowIndex)||rowIndex<0||rowIndex>=rows.length||!Object.hasOwn(rows,rowIndex)||!record(rows[rowIndex]))throw Error('Select one held span row.');
 const source=validateTarget({traceId:rows[rowIndex].TraceId,spanId:rows[rowIndex].SpanId}),target=structuredClone(source);
 const question={signal:'spans',...Environment.questionFields(origin),fromMs:origin.fromMs,toMs:origin.toMs,linksTo:structuredClone(target),limit:100,offset:0};
 validateRecordQuestion(question);
 return {relation:'reported-span-link-incoming',question,target,origin:{question:structuredClone(origin),rowIndex,source}};
}
