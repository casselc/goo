/* Query compatibility only; no data, source authority or partition attestation. */
import {eligible as aggregateEligible} from './view-declaration.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {sameQuestionEnvironment} from './record-environment.mjs';
import {identity} from './question-session.mjs';
export const eligible=q=>aggregateEligible(q)&&!q.numeric&&!q.numericHistogram&&!q.reportedEdge&&!q.spanStatus&&!q.groupTimeSeries;
export function compatibleCohorts(a,b){
 for(const q of [a,b]){validateRecordQuestion(q);if(!eligible(q))throw Error('Group comparison requires ordinary single-group count cohorts.');}
 if(!sameQuestionEnvironment(a,b))throw Error('Group comparison environment policy differs.');
 const context=q=>{const copy=structuredClone(q);for(const key of ['where','fromMs','toMs','environment','environmentSelector'])delete copy[key];copy.offset=0;copy.limit=100;return identity(copy)};
 if(context(a)!==context(b))throw Error('Group comparison signal, grouping or measure profile differs.');return true;
}
