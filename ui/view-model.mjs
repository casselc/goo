/* Public view contracts: structural data validation plus runtime guarantees. */
export * from './view-schema.mjs';
import {validateDisplay as validateStructure,capabilities as describeCapabilities} from './view-schema.mjs';
import {hasGuarantee} from './snapshots.mjs';
const aspect=snapshot=>snapshot?.shape==='time-series'?'intervals':snapshot?.shape==='summary'?'summary':'groups';
export function validateDisplay(snapshot,display){
 const result=validateStructure(snapshot,display);
 if(display?.requiresComplete&&!hasGuarantee(snapshot,aspect(snapshot))){
  const completeReason=reason=>/^Complete captured-query (groups|intervals|summary) are required for this scope and revision$/.test(reason)||/^Complete extent requires matching captured-query (groups|intervals|summary) attestation$/.test(reason);
  return {ok:false,reasons:[...result.reasons.filter(reason=>!completeReason(reason)),'Complete-only display requires verified coverage of the full query.']};
 }
 return result;
}
export function capabilities(snapshot){
 const described=describeCapabilities(snapshot);
 return Object.freeze({...described,completeAspects:Object.freeze(described.completeAspects.filter(a=>hasGuarantee(snapshot,a)))});
}
export const supportsSelection=snapshot=>capabilities(snapshot).selectRows;
