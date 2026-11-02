/* App-level continuation over supplied full counts. No IO or persisted view grammar. */
import * as Composition from './view-composition.mjs';
import {cohort} from './captured-count-comparison.mjs';
import {eligibilityReason} from './count-cohort-context.mjs';
import {identity} from './question-session.mjs';
import {hasGuarantee,isSnapshot} from './snapshots.mjs';
const frozen=x=>{if(x&&typeof x==='object'){for(const child of Object.values(x))frozen(child);Object.freeze(x)}return x};
const key=q=>identity({...q,offset:0,limit:100});
function definition(composition,q){Composition.validate(composition,q);const unsupported=eligibilityReason(q);if(unsupported)throw Error(unsupported);if(Composition.aggregate(composition)?.display.value!=='count')throw Error('Select Count in the breakdown before comparing a selection.');if(Composition.comparison(composition))throw Error('Remove the existing count comparison before comparing a selection.');}
export function full(composition,q,data){definition(composition,q);const snapshot=data?.viewDataset;if(!isSnapshot(snapshot)||!hasGuarantee(snapshot,'groups')||snapshot.shape!=='ranked'||key(data.question)!==key(q)||key(snapshot.scope.selector)!==key(q)||snapshot.revision!==data.basis||snapshot.capturedAt!==data.capturedAt||!Array.isArray(data.groups))throw Error('Run the count breakdown to acquire full groups before comparing a selection.');return data;}
export function reason(composition,q,data){try{full(composition,q,data);return ''}catch(error){return error.message}}
export function plan(composition,parent,selected,mode,rest,data){
 full(composition,parent,data);definition(composition,selected);if(!['parent','rest'].includes(mode))throw Error('Choose held parent or rest.');
 const context=q=>{const next=structuredClone(q);delete next.where;return key(next)};if(context(parent)!==context(selected)||mode==='rest'&&(!rest||context(parent)!==context(rest)))throw Error('Selected and rest questions must retain the exact parent context.');
 return Object.freeze({mode,question:frozen(structuredClone(selected)),parent:frozen(structuredClone(parent)),...(mode==='parent'?{held:data}:{rest:frozen(structuredClone(rest))})});
}
export function matches(intent,composition,q){try{definition(composition,q);return !!intent&&key(intent.question)===key(q)}catch{return false}}
export function request(intent,composition,selected,baseline){
 if(!matches(intent,composition,selected.question))throw Error('Selection comparison scope changed.');full(composition,selected.question,selected);full(composition,intent.mode==='parent'?intent.parent:intent.rest,baseline);
 return {baseline:cohort(baseline),selection:cohort(selected)};
}
