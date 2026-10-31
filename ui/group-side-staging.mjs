/* Pure group-side question intent. Acquisition and action authority stay with the host. */
import {asGroupComparison} from './group-count-comparison.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
const keys=['type','side','key','datasetId','revision','scopeId','generation'];
function closed(intent){
 if(!intent||typeof intent!=='object'||Array.isArray(intent)||![Object.prototype,null].includes(Object.getPrototypeOf(intent)))throw TypeError('Use a plain group-side intent.');
 const names=Reflect.ownKeys(intent);if(names.length!==keys.length||keys.some(key=>!Object.hasOwn(intent,key)))throw TypeError('Use the exact group-side intent fields.');
 for(const key of names){const d=Object.getOwnPropertyDescriptor(intent,key);if(typeof key!=='string'||!keys.includes(key)||!d?.enumerable||!Object.hasOwn(d,'value')||typeof d.value!=='string')throw TypeError('Group-side intents require string data fields.');}
}
function selected(comparison,side,key){
 const pair=asGroupComparison(comparison);
 if(!['baseline','selection'].includes(side)||typeof key!=='string')throw TypeError('Choose baseline or current and a supplied group key.');
 if(pair.state==='diagnostic'||pair.baselineAccounting!=='coherent'||pair.selectionAccounting!=='coherent')throw TypeError('Diagnostic group comparisons cannot stage questions.');
 const row=pair.rows.find(row=>row.key===key),cell=row?.[side];
 if(!row||cell.presence!=='present'||cell.state!=='derived'||BigInt(cell.value)<=0n)throw TypeError('Choose a positive group reported in this side.');
 return {pair,row,snapshot:pair[side]};
}
export function groupSideIntent(comparison,side,key){
 const {snapshot}=selected(comparison,side,key);
 return {type:'scope-group-side',side,key,datasetId:snapshot.id,revision:snapshot.revision,scopeId:snapshot.scope.id,generation:snapshot.scope.generation};
}
export function stagedGroupQuestion(comparison,intent){
 closed(intent);const {pair,row}=selected(comparison,intent.side,intent.key),expected=groupSideIntent(pair,intent.side,intent.key);
 if(keys.some(key=>intent[key]!==expected[key]))throw TypeError('Group-side intent differs from its selected held snapshot.');
 const question=structuredClone(pair[intent.side].scope.selector),leaf={field:structuredClone(pair.field),op:row.value===null?'missing':'eq',...(row.value===null?{}:{value:row.value})};
 question.offset=0;question.where=question.where?{all:[question.where,leaf]}:leaf;
 validateRecordQuestion(question);validatePredicate(question.where,question.signal);return question;
}
export const stageGroupSide=stagedGroupQuestion;
