import * as RecordEnvironment from './record-environment.mjs';
/* Application-owned resource return presentation. No reads or persisted state. */
import * as UI from './components.mjs';
import {actionVisible,revealFocus} from './dom-visibility.mjs';
import * as EvidenceSummary from './evidence-summary.mjs';
function originEnvironment(question){const q=question.inventory||question.explore||question,definition=q.definition||question.measurement?.definition;return definition?RecordEnvironment.questionLabel(definition):(q.environmentSelector!=null||q.environment!==undefined?RecordEnvironment.questionLabel(q):'Environment unavailable')}
const active=new WeakMap();
function mount(parent,{className='',id}={}){
 const node=UI.node,bar=node('div');bar.className='resource-return controls '+className;if(id)bar.id=id;bar.hidden=true;parent.prepend(bar);let revision=0,holdRevision=0,origin=null;
 function clear(){revision++;origin=null;bar.hidden=true;bar.replaceChildren();if(active.get(parent)===owner)active.delete(parent)}
 function show({title,question,backLabel,onBack,citation,notice,citationLabel='Retained native origin'}){
  active.get(parent)?.clear();clear();const own=revision,back=node('button',backLabel),label=node('strong',title),scope=node('span','Origin environment: '+originEnvironment(question)),window=node('span'),details=node('details'),summary=node('summary','Origin scope');back.type='button';scope.className='resource-return-environment';window.className='resource-return-window';details.className='resource-return-scope';
  const from=question.fromNS??String(BigInt(question.fromMs)*1000000n),to=question.toNS??String(BigInt(question.toMs)*1000000n);
  window.append('Fixed window: ',EvidenceSummary.timestamp(from),' – ',EvidenceSummary.timestamp(to),' (end excluded)');
  back.onclick=()=>{if(active.get(parent)!==owner||own!==revision||bar.hidden||!actionVisible(back)||back.closest('[inert]'))return;clear();onBack()};details.append(summary,scope,window);
  if(citation){const retained=node('details'),receipt=node('pre',JSON.stringify(citation,null,2));retained.append(node('summary',citationLabel),receipt);details.append(retained)}if(notice)details.append(node('p',notice));
  bar.append(back,label,details);bar.hidden=false;origin={title,question:structuredClone(question)};active.set(parent,owner);
 }
 function hold(question){
  // A retained resource return is ancestry only for the exact question it opened.
  const stable=value=>Array.isArray(value)?value.map(stable):value&&typeof value==='object'?Object.fromEntries(Object.keys(value).sort().map(key=>[key,stable(value[key])])):value;
  if(question&&(!origin?.question.definition||JSON.stringify(stable(origin.question.definition))!==JSON.stringify(stable(question.definition))||['fromMs','toMs','fromNS','toNS'].some(key=>origin.question[key]!==question[key]))){clear();return null}

  if(bar.hidden||active.get(parent)!==owner||!origin)return null;
  const own=revision,token=++holdRevision,snapshot=structuredClone(origin);bar.hidden=true;
  return {...snapshot,restore(){if(active.get(parent)!==owner||own!==revision||token!==holdRevision)return;holdRevision++;bar.hidden=false}};
 }
 const owner={show,clear,hold,dispose(){clear();bar.remove()}};return owner;
}
function hold(parent,question){return active.get(parent)?.hold(question)??null}
function restoreFocus(origin,fallback){revealFocus(origin,fallback)}
export {originEnvironment,mount,hold,restoreFocus};
