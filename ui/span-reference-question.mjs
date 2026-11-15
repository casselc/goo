/* Deliberate reference expansion, separate from population-preserving projection.
 * Only the supplied environment policy and fixed window carry into this query. */
import {validateRecordQuestion} from './record-intent.mjs';
import * as Environment from './record-environment.mjs';
import {validateTarget,validateTraceId,validateSpanId} from './span-link-membership.mjs';
function fixedReference(scope,signal,offset){
 Environment.requireNarrow(scope);
 return validateRecordQuestion({signal,...Environment.questionFields(scope),fromMs:scope.fromMs,toMs:scope.toMs,limit:100,offset});
}
export function traceQuestion(scope,traceId,spanId='',signal='spans',offset=0){
 validateTraceId(traceId);if(spanId!=='')validateSpanId(spanId);
 const q=fixedReference(scope,signal,offset),all=[{field:{kind:'column',key:'trace_id'},op:'eq',value:traceId}];
 if(spanId!=='')all.push({field:{kind:'column',key:'span_id'},op:'eq',value:spanId});
 return validateRecordQuestion({...q,where:{all}});
}
export function incomingQuestion(scope,target,offset=0){
 return validateRecordQuestion({...fixedReference(scope,'spans',offset),linksTo:validateTarget(target)});
}
