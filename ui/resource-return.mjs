/* Application-owned resource return presentation. No reads or persisted state. */
import * as UI from './components.mjs';
import * as EvidenceSummary from './evidence-summary.mjs';
const active=new WeakMap();
function mount(parent,{className='',id}={}){
 const node=UI.node,bar=node('div');bar.className='resource-return controls '+className;if(id)bar.id=id;bar.hidden=true;parent.prepend(bar);let revision=0,holdRevision=0,origin=null;
 function clear(){revision++;origin=null;bar.hidden=true;bar.replaceChildren();if(active.get(parent)===owner)active.delete(parent)}
 function show({title,question,backLabel,onBack}){
  active.get(parent)?.clear();clear();const own=revision,back=node('button',backLabel),label=node('strong',title),scope=node('span',question.environment??question.definition?.environment??'Environment unavailable'),window=node('span');window.className='resource-return-window';
  const from=question.fromNS??String(BigInt(question.fromMs)*1000000n),to=question.toNS??String(BigInt(question.toMs)*1000000n);
  window.append('Fixed window: ',EvidenceSummary.timestamp(from),' – ',EvidenceSummary.timestamp(to),' (end excluded)');
  back.onclick=()=>{if(active.get(parent)!==owner||own!==revision||bar.hidden)return;clear();onBack()};bar.append(back,label,scope,window);bar.hidden=false;origin={title,question:structuredClone(question)};active.set(parent,owner);
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
function restoreFocus(origin,fallback){
 const target=origin?.isConnected&&UI.actionVisible(origin)?origin:fallback;
 target?.scrollIntoView({block:'center',behavior:'instant'});target?.focus({preventScroll:true});
}
export {mount,hold,restoreFocus};
