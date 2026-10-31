/* Contextual preservation choices. This module never reads or publishes data. */
import {node} from './components.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
const meanings=Object.freeze({
 question:Object.freeze({title:'Question',note:'Fixed query and scope · no telemetry read.'}),
 view:Object.freeze({title:'View definition',note:'Queries and presentation · no telemetry read.'}),
 evidence:Object.freeze({title:'Evidence',note:'Opens a capture form. Confirmation makes a new independent read.'})
});
export function meaning(kind){if(!Object.hasOwn(meanings,kind))throw Error('Choose a supported preservation kind.');return {...meanings[kind]};}
export function menu(label){const root=node('details',undefined,'preserve-menu');root.setAttribute('aria-label',label);root.append(node('summary','Preserve'));return {root,group(kind,...children){const copy=meaning(kind),section=node('div',undefined,'preserve-section');section.dataset.preserveKind=kind;section.append(node('strong',copy.title),node('small',copy.note));section.append(...children);root.append(section);return section;}};}
export function selectedQuestion(queries,signal=''){
 if(!Array.isArray(queries)||!queries.length||queries.length>2)throw Error('Choose a configured record question.');
 const chosen=queries.length===1?queries[0].signal:signal;
 if(!chosen)throw Error('Choose a signal before preserving a question or evidence.');
 const matches=queries.filter(query=>query.signal===chosen);
 if(matches.length!==1)throw Error('Choose one configured record signal.');
 const question={...structuredClone(matches[0]),offset:0};validateRecordQuestion(question);return question;
}
export function availability({active=true,connected=true,busy=false,pending=false,applied=true}={}){
 if(!active||!connected)return 'Return to this active context before preserving it.';
 if(busy)return 'Wait for the current read to finish.';
 if(pending)return 'Apply or clear pending selections and drafts first.';
 if(!applied)return 'Run this question before capturing evidence.';
 return '';
}
