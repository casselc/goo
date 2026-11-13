/* Generated categorical scope only; the host owns leases, reads and navigation. */
import {validateRecordQuestion} from './record-intent.mjs';
import {validate,validateField} from './predicate-validation.mjs';
import {predicate as selectedPredicate} from './counted-value-selection.mjs';
import * as Comparison from './captured-count-comparison.mjs';
import * as Declaration from './view-declaration.mjs';
import * as Composition from './view-composition.mjs';

function parentQuestion(question,field){
 validateRecordQuestion(question);validateField(field,question.signal);
 if((question.offset??0)!==0||question.groupBy?.length!==1||question.groupBy[0].kind!==field.kind||question.groupBy[0].key!==field.key||question.numeric||question.numericHistogram||question.summary||question.spanStatus||question.groupTimeSeries||question.bucketMs||question.reportedEdge||!Comparison.eligible(question))throw Error('Compare values from one ordinary page-zero categorical Count scope.');
 const parent=structuredClone(question);parent.limit=100;parent.offset=0;return parent;
}
function combine(where,clause,signal){return validate(where?{all:[...(where.all||[where]),clause]}:clause,signal);}

export function plan(question,field,values){
 const parent=parentQuestion(question,field),selected=structuredClone(parent),rest=structuredClone(parent);
 // This also validates dense bounded values, uniqueness and every typed literal.
 selected.where=selectedPredicate(parent.where,field,values,parent.signal);
 const leaf=(op,value)=>({field:structuredClone(field),op,...(value===undefined?{}:{value})});
 const present=[leaf('exists'),...values.filter(value=>value!==null).map(value=>leaf('ne',value))],presentRest=present.length===1?present[0]:{all:present};
 const complement=values.includes(null)?presentRest:{any:[leaf('missing'),presentRest]};
 rest.where=combine(parent.where,complement,parent.signal);
 validateRecordQuestion(selected);validateRecordQuestion(rest);
 const declaration=Comparison.createQuery(selected,rest,'count');declaration.display.title='Rest of scope vs selected values';declaration.display.view='summary';
 const composition=Composition.setComparison(Composition.create(selected,Declaration.create(selected)),declaration,selected);
 return Composition.view({version:1,title:'Selected vs rest of scope',queries:[selected],composition});
}
export function reason(question,field,values){try{plan(question,field,values);return '';}catch(error){return error.message;}}
