import {validateRecordQuestion} from './question-protocol.mjs';
import {traceQuestion,incomingQuestion} from './span-reference-question.mjs';
import {validateTarget,linkAt} from './span-link-membership.mjs';

/* Pure admission and planning for one reported OTel link. Held-result ownership,
 * clocks, source reads and revocable navigation capabilities belong to callers. */
const record=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
export function inspect({rows,rowIndex,linkIndex}){
 if(!Array.isArray(rows)||!Number.isInteger(rowIndex)||rowIndex<0||rowIndex>=rows.length||!Object.hasOwn(rows,rowIndex)||!record(rows[rowIndex]))throw Error('Select one held span row.');
 const row=rows[rowIndex],source=validateTarget({traceId:row.TraceId,spanId:row.SpanId}),{target,attributes}=linkAt(row,linkIndex);
 const rowIndexes=[];
 for(let index=0;index<rows.length;index++)if(record(rows[index])&&rows[index].TraceId===target.traceId&&rows[index].SpanId===target.spanId)rowIndexes.push(index);
 return {target,source,attributes:structuredClone(attributes),local:{state:rowIndexes.length===1?'unique':rowIndexes.length?'ambiguous':'absent',rowIndexes}};
}
export function plan({origin,rows,rowIndex,linkIndex}){
 validateRecordQuestion(origin);
 if(origin.signal!=='spans')throw Error('Reported span links require an ordinary spans question.');
 const {target,source,attributes,local}=inspect({rows,rowIndex,linkIndex});
 const question=traceQuestion(origin,target.traceId,target.spanId);
 validateRecordQuestion(question);
 return {relation:'reported-span-link',question,target,local,origin:{question:structuredClone(origin),rowIndex,linkIndex,source},attributes:structuredClone(attributes)};
}

export function planIncoming({origin,rows,rowIndex}){
 validateRecordQuestion(origin);
 if(origin.signal!=='spans')throw Error('Reported span-link discovery requires an ordinary spans question.');
 if(!Array.isArray(rows)||!Number.isInteger(rowIndex)||rowIndex<0||rowIndex>=rows.length||!Object.hasOwn(rows,rowIndex)||!record(rows[rowIndex]))throw Error('Select one held span row.');
 const source=validateTarget({traceId:rows[rowIndex].TraceId,spanId:rows[rowIndex].SpanId}),target=structuredClone(source);
 const question=incomingQuestion(origin,target);
 validateRecordQuestion(question);
 return {relation:'reported-span-link-incoming',question,target,origin:{question:structuredClone(origin),rowIndex,source}};
}
