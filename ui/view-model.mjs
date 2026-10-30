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
 if(display?.scale==='population'){const unavailable=populationUnavailable(snapshot);if(unavailable)return {ok:false,reasons:[...new Set([...result.reasons,unavailable])]}}
 return result;
}
export function capabilities(snapshot){
 const described=describeCapabilities(snapshot);
 return Object.freeze({...described,completeAspects:Object.freeze(described.completeAspects.filter(a=>hasGuarantee(snapshot,a)))});
}
export const supportsSelection=snapshot=>capabilities(snapshot).selectRows;

// Exact denominator facts and rounded presentation remain separate. The fraction
// keeps the captured numerator/denominator, without a whole-query claim on copies.
function populationUnavailable(snapshot){
 if(!hasGuarantee(snapshot,'groups'))return 'Population shares require verified coverage of the full query.';
 const p=snapshot.partition;
 if(snapshot.shape!=='ranked'||!p||p.kind!=='disjoint-record-groups'||p.measure!=='count')return 'Captured-population shares are unavailable for this result';
 if(!['recorded','derived'].includes(snapshot.quality.value)||snapshot.rows.some(row=>row.state==='unknown'))return 'Population shares require known captured counts';
 if(p.total==='0')return 'No captured population is available for shares';
 return null;
}
export function populationShare(snapshot,rowKey){
 const unavailable=populationUnavailable(snapshot);if(unavailable)return {ok:false,reason:unavailable};
 const p=snapshot.partition;
 const row=snapshot.rows.find(row=>row.key===rowKey);if(!row)return {ok:false,reason:'Choose a supplied group'};
 const n=BigInt(row.values.count),total=BigInt(p.total),rounded=(n*1000n+total/2n)/total;
 const percent=n===0n?'0%':n===total?'100%':rounded===0n?'<0.1%':rounded>=1000n?'>99.9%':`${n*1000n%total?'≈ ':''}${rounded/10n}${rounded%10n?'.'+rounded%10n:''}%`;
 return {ok:true,count:row.values.count,total:p.total,fraction:row.values.count+'/'+p.total,percent,coordinate:Number(n)/Number(total)*100};
}
