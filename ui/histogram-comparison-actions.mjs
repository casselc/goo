/* Pure explicit side planning over a captured pair. This is not current-scope
   H1 staging, source authentication, or permission to relabel a held baseline. */
import {isHistogramComparison} from './snapshots.mjs';
import {edges as coefficientEdges} from './numeric-bins.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';

const own=(value,key)=>Object.prototype.hasOwnProperty.call(value,key);
const refusal=(code,reason)=>({ok:false,code,reason});
function plain(value,keys){
 if(!value||typeof value!=='object'||![Object.prototype,null].includes(Object.getPrototypeOf(value)))return null;
 const names=Reflect.ownKeys(value);if(names.length!==keys.length||names.some(key=>typeof key!=='string'||!keys.includes(key)))return null;
 const result={};for(const key of keys){const d=Object.getOwnPropertyDescriptor(value,key);if(!d||!own(d,'value')||!d.enumerable)return null;result[key]=d.value}return result;
}
const identityKeys=['type','datasetId','revision','scopeId','generation','fromKey','toKey'];
export function createHistogramComparisonActions(configuration){
 const config=plain(configuration,['comparison','active']);
 if(!config||!isHistogramComparison(config.comparison)||typeof config.active!=='function')throw TypeError('Bind an admitted histogram comparison and a synchronous active callback.');
 const comparison=config.comparison;
 if(comparison.baseline.histogram.kind!=='numeric'||comparison.selection.histogram.kind!=='numeric')throw TypeError('Record range planning requires numeric histogram operands.');
 let retired=false;
 function resolve(value){
  try{
   if(retired)return refusal('retired','This comparison has been withdrawn.');
   const active=config.active();if(typeof active!=='boolean')return refusal('owner-state','Comparison owner must return a synchronous boolean.');
   if(!active)return refusal('inactive','This comparison is not active.');
   if(comparison.state==='diagnostic')return refusal('diagnostic','Source reads disagree; a diagnostic comparison cannot plan a range.');
   const selection=plain(value,['side','intent']);if(!selection||!['baseline','selection'].includes(selection.side))return refusal('invalid-side','Choose an explicit captured comparison side.');
   const intent=plain(selection.intent,identityKeys),snapshot=comparison[selection.side];
   if(!intent||intent.type!=='scope-histogram')return refusal('invalid-intent','Choose a supplied histogram range.');
   if(intent.datasetId!==snapshot.id||intent.revision!==snapshot.revision||intent.scopeId!==snapshot.scope.id||intent.generation!==snapshot.scope.generation)return refusal('identity-mismatch','Range does not belong to the captured side.');
   if(typeof intent.fromKey!=='string'||typeof intent.toKey!=='string')return refusal('invalid-membership','Range endpoints must be supplied bin keys.');
   const rows=snapshot.rows,fromBin=rows.findIndex(row=>row.key===intent.fromKey),last=rows.findIndex(row=>row.key===intent.toKey);
   if(fromBin<0||last<fromBin)return refusal('invalid-membership','Range endpoints must be ordered supplied bin keys.');
   const toBin=last+1,count=rows.slice(fromBin,toBin).reduce((sum,row)=>sum+BigInt(row.values.count),0n);
   if(count===0n)return refusal('empty-selection','The chosen side has no captured records in this range.');
   const question=snapshot.scope.selector;
   try{validateRecordQuestion(question);validatePredicate(question.where,question.signal)}catch(error){return refusal('invalid-question','The retained side question cannot be planned: '+error.message)}
   if((question.groupBy??[]).length>1)return refusal('unsupported-scope','Range planning supports zero or one retained breakdown.');
   const edges=coefficientEdges(snapshot.histogram.binding.type);
   return Object.freeze({ok:true,side:selection.side,question,fromBin,toBin,from:edges[fromBin],to:edges[toBin],count:String(count)});
  }catch(error){return refusal('invalid-intent',error.message||'Unable to resolve this comparison range.');}
 }
 return Object.freeze({resolve,revoke(){retired=true}});
}
