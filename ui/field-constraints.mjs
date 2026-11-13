/* Exact authored-clause intent. No query execution, DOM access or retained data. */
import {validateRecordQuestion} from './record-intent.mjs';
import {validate as validatePredicate,validateField} from './predicate-validation.mjs';
import {identity} from './question-session.mjs';
import {object,array} from './intent-shape.mjs';

function canonical(value){
 if(Array.isArray(value))return value.map(canonical);
 if(value&&typeof value==='object')return Object.fromEntries(Object.keys(value).sort().map(key=>[key,canonical(value[key])]));
 return value;
}
const key=value=>JSON.stringify(canonical(value));
const sameField=(a,b)=>a?.kind===b.kind&&a?.key===b.key;

export function plan(question,field){
 validateRecordQuestion(question);validateField(field,question.signal);
 const clauses=[];
 function visit(clause,path,conjunctive){
  const group=clause.all?'all':clause.any?'any':null;
  if(group){clause[group].forEach((child,index)=>visit(child,[...path,group,index],conjunctive&&group==='all'));return;}
  if(!sameField(clause.numeric?.field||clause.field,field))return;
  clauses.push({path,clause:structuredClone(clause),clauseKey:key(clause),removable:conjunctive,reason:conjunctive?'':'Inside an OR group. Edit the condition or group instead.'});
 }
 if(question.where)visit(question.where,[],true);
 return {questionKey:identity(question),contentKey:key(question),field:structuredClone(field),clauses};
}

function selected(question,intent,choice){
 validateRecordQuestion(question);
 object(intent,['questionKey','contentKey','field','clauses'],'Choose the original applied constraint.');
 object(choice,['path','clause','clauseKey','removable','reason'],'Choose one exact applied condition.');
 array(choice.path);
 if(choice.path.length%2||choice.path.some((part,index)=>index%2?!Number.isInteger(part)||part<0:!['all','any'].includes(part))||typeof choice.clauseKey!=='string'||typeof choice.removable!=='boolean'||typeof choice.reason!=='string')throw Error('Choose one exact applied condition.');
 validatePredicate(choice.clause,question.signal);
 const current=plan(question,intent.field);
 if(intent.questionKey!==current.questionKey||intent.contentKey!==current.contentKey)throw Error('Scope changed. Choose the field again.');
 const actual=current.clauses.find(entry=>key(entry.path)===key(choice.path));
 if(!actual||key(actual)!==key(choice))throw Error('Condition changed. Choose the field again.');
 return actual;
}

export function resolve(question,intent,choice){
 return structuredClone(selected(question,intent,choice).path);
}

export function remove(question,intent,choice){
 const actual=selected(question,intent,choice);
 if(!actual.removable)throw Error(actual.reason);
 const next=structuredClone(question),path=actual.path;
 function without(clause,depth){
  if(depth===path.length)return undefined;
  const group=path[depth],index=path[depth+1],child=without(clause[group][index],depth+2);
  if(child===undefined)clause[group].splice(index,1);else clause[group][index]=child;
  // Only remove emptied groups on the chosen conjunction ancestry. Preserve
  // every sibling, its order and authored single-child groups exactly.
  return clause[group].length?clause:undefined;
 }
 const where=without(next.where,0);
 if(where===undefined)delete next.where;else next.where=where;
 validateRecordQuestion(next);
 return next;
}
