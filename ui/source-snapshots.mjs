/* Trusted application admission for known Go exploration response slots.
   Generic components/imports never promote JSON claims into source guarantees. */
import {createSourceAdapter,asComparison,importSnapshot,getSnapshotInfo} from './snapshots.mjs';
import {identity} from './question-session.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
import {validate as validateSaved} from './saved-capture-protocol.mjs';

const source=createSourceAdapter({id:'goo.explore'}),stored=createSourceAdapter({id:'goo.verified-saved-explore'});
const slots=Object.freeze({viewDataset:'ranked',timeDataset:'time-series',summaryDataset:'summary',histogramDataset:'histogram'});
const record=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
function questionKey(question,paging=false){
  validateRecordQuestion(question);validatePredicate(question.where,question.signal);
  const copy=structuredClone(question);if(paging){delete copy.offset;delete copy.limit}return identity(copy);
}
function sameQuestion(expected,actual,paging=false){
  if(questionKey(expected,paging)!==questionKey(actual,paging))throw Error('Returned snapshot scope differs from the requested question.');
}
function admit(data,expected,adapter,{paging=false}={}){
  if(!record(data))throw Error('An exploration response is required.');
  sameQuestion(expected,data.question,paging);
  const result={...data};
  for(const [slot,shape] of Object.entries(slots)){
    const value=data[slot];if(value===undefined||value===null)continue;
    if(!record(value)||value.shape!==shape)throw Error('Unexpected exploration snapshot shape in '+slot+'.');
    sameQuestion(expected,value.scope?.selector,paging);
    if(value.revision!==data.basis||value.scope.generation!==data.basis)throw Error('Snapshot basis differs from its exploration response.');
    if(value.capturedAt!==data.capturedAt)throw Error('Snapshot clock differs from its exploration response.');
    // A stored record page changes only delivery paging. Its aggregate keeps the
    // independently checked original selector and original source read clock.
    result[slot]=adapter.adoptSnapshot(value,{question:paging?value.scope.selector:expected});
  }
  return result;
}

export function adoptExploration(data,requestedQuestion){return admit(data,requestedQuestion,source)}

export function adoptStoredExplore(page){
  validateSaved(page);
  if(page?.receipt?.question?.kind!=='explore'||!record(page.explore))throw Error('A verified stored exploration page is required.');
  const expected=page.receipt.question.explore;
  // SavedProtocol historically checked selector pages only. All stored pages
  // must match their immutable receipt, including named-environment pages.
  sameQuestion(expected,page.explore.question,true);
  return {...page,explore:admit(page.explore,expected,stored,{paging:true})};
}

export function importComparisonResult(data,requested){
  if(!record(data))throw Error('A comparison response is required.');
  if(data.viewComparison===undefined||data.viewComparison===null)return data;
  for(const side of ['baseline','selection'])sameQuestion(requested?.[side]?.question,data.viewComparison?.[side]?.scope?.selector);
  // /compare computes over browser-supplied cohorts. The HTTP response is not a
  // fresh authoritative source read; both operands remain conservative imports.
  const held=value=>getSnapshotInfo(value)?.entry==='held'?value:importSnapshot(value);
  return {...data,viewComparison:asComparison({...data.viewComparison,baseline:held(data.viewComparison.baseline),selection:held(data.viewComparison.selection)})};
}
