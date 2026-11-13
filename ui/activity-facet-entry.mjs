/* Local intent for the existing counted-values owner. No source reads or data cache. */
import {query} from './filter-values.mjs';
import {identity} from './question-session.mjs';
export function create(context,field,observationWindow){
 if(!context?.active||!context.clean||!context.selector||!context.readWindow||['loading','stopped'].includes(context.phase))throw Error('Apply Activity scope and wait for a completed window before exploring values.');
 const question=query({...context.selector,signal:context.selector.signal||'logs',...context.readWindow},field);
 if(!observationWindow||!Number.isSafeInteger(observationWindow.fromMs)||!Number.isSafeInteger(observationWindow.toMs)||observationWindow.fromMs<0||observationWindow.toMs<=observationWindow.fromMs)throw Error('The inspected observation window is unavailable.');
 const differs=question.fromMs!==observationWindow.fromMs||question.toMs!==observationWindow.toMs;
 return {question,observationWindow:structuredClone(observationWindow),label:'Latest completed Activity window · '+new Date(question.fromMs).toISOString()+' → '+new Date(question.toMs).toISOString()+(differs?' · inspected record window differs':'')};
}
export function current(intent,context){try{return identity(intent.question)===identity(create(context,intent.question.groupBy[0],intent.observationWindow).question)}catch{return false}}
