/* Pure population conservation. Projection changes never rebuild selectors. */
import {validateRecordQuestion} from './record-intent.mjs';
import {object} from './intent-shape.mjs';

export function populationQuery(parent,paging={}){
 object(paging,['limit','offset'],'Population projections accept paging only.');
 const {limit=100,offset=0}=paging;
 validateRecordQuestion(parent);
 const question=structuredClone(parent);
 for(const key of ['numeric','numericHistogram','groupBy','summary','spanStatus','logSeverity','bucketMs','groupTimeSeries'])delete question[key];
 question.limit=limit;question.offset=offset;
 validateRecordQuestion(question);
 return question;
}
