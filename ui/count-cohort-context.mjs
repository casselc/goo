/* Query compatibility only; no data, source authority or partition attestation. */
import {eligible as aggregateEligible} from './view-declaration.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {sameQuestionEnvironment} from './record-environment.mjs';
import {identity} from './question-session.mjs';
export function eligibilityReason(q){
 if(q?.numeric||q?.numericHistogram)return 'Count comparisons do not support numeric measures or histograms. Choose an ordinary categorical Count breakdown.';
 if(q?.reportedEdge)return 'Count comparisons do not support reported-edge measures. Choose an ordinary categorical Count breakdown.';
 if(q?.groupTimeSeries)return 'Count comparisons do not support grouped time series. Choose an ordinary categorical Count breakdown.';
 if(q?.spanStatus)return 'Count comparisons do not support span status. Turn off Span status and errors in Measures, then Run.';
 if(q?.logSeverity)return 'Count comparisons do not support numeric log severity. Choose an ordinary categorical Count breakdown.';
 if(!aggregateEligible(q))return 'Count comparisons require one ordinary categorical grouping. Choose a single-field breakdown, then Run.';
 return '';
}
export const eligible=q=>eligibilityReason(q)==='';
export function compatibleCohorts(a,b){
 for(const q of [a,b]){validateRecordQuestion(q);if(!eligible(q))throw Error('Group comparison requires ordinary single-group count cohorts.');}
 if(!sameQuestionEnvironment(a,b))throw Error('Group comparison environment policy differs.');
 const context=q=>{const copy=structuredClone(q);for(const key of ['where','fromMs','toMs','environment','environmentSelector'])delete copy[key];copy.offset=0;copy.limit=100;return identity(copy)};
 if(context(a)!==context(b))throw Error('Group comparison signal, grouping or measure profile differs.');return true;
}
