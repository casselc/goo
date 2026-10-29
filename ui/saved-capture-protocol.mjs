/* Checked stored response boundary. No source reads, DOM or persistence. */
import * as Environment from './record-environment.mjs';
import * as Questions from './question-protocol.mjs';
import * as Session from './question-session.mjs';
import * as Explanation from './explanation.mjs';
function recordIdentity(q){const value=structuredClone(q);delete value.offset;delete value.limit;return Session.identity(value)}
function checkedRecord(expected,actual){Questions.validateRecordQuestion(expected);Questions.validateRecordQuestion(actual);if(recordIdentity(expected)!==recordIdentity(actual))throw Error('Stored record scope differs from its immutable receipt.');}
function validate(page,expectedID){
  const receipt=page?.receipt;if(!/^[a-f0-9]{64}$/.test(receipt?.id)||expectedID!==undefined&&receipt.id!==expectedID)throw Error('Returned capture differs from the requested ID.');
  const q=receipt.question;if(!Environment.hasQuestionSelector(q)&&!Environment.hasQuestionSelector(page.explore?.question)&&!['baseline','selection'].some(side=>Environment.hasQuestionSelector(page.comparison?.request?.[side]?.question)||Environment.hasQuestionSelector(page.explanation?.request?.[side])))return page;
  if(q.kind==='explore'&&page.explore){checkedRecord(q.explore,page.explore.question);const offset=page.explore.question.offset;if(!Number.isInteger(offset)||offset<0||offset>1000||offset%100!==0||page.explore.question.limit!==100)throw Error('Invalid stored record page.');}
  else if(q.kind==='comparison'&&page.comparison){for(const side of ['baseline','selection'])checkedRecord(q.comparison[side],page.comparison.request?.[side]?.question);if(!Environment.sameQuestionEnvironment(q.comparison.baseline,q.comparison.selection))throw Error('Stored comparison environment policies differ.');}
  else if(q.kind==='explanation'&&page.explanation){Explanation.validate(q.explanation);Explanation.validate(page.explanation.request);for(const side of ['baseline','selection'])checkedRecord(q.explanation[side],page.explanation.request[side]);if(Session.identity({fields:q.explanation.fields})!==Session.identity({fields:page.explanation.request.fields}))throw Error('Stored explanation fields differ from its receipt.');}
  else throw Error('This stored capture family does not support environment selectors.');
  return page;
}
function validateRefresh(page,previous){validate(page);if(page.receipt.parent!==previous.receipt.id||Session.identity(page.receipt.question)!==Session.identity(previous.receipt.question))throw Error('Refreshed capture changed its predecessor or immutable question.');return page}
export {validate,validateRefresh};
