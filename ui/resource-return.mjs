/* Application-owned resource return presentation. No reads or persisted state. */
import * as UI from './components.mjs';
import * as EvidenceSummary from './evidence-summary.mjs';
const active=new WeakMap();
function mount(parent,{className='',id}={}){
 const node=UI.node,bar=node('div');bar.className='resource-return controls '+className;if(id)bar.id=id;bar.hidden=true;parent.prepend(bar);let revision=0;
 function clear(){revision++;bar.hidden=true;bar.replaceChildren();if(active.get(parent)===owner)active.delete(parent)}
 function show({title,question,backLabel,onBack}){
  active.get(parent)?.clear();clear();const own=revision,back=node('button',backLabel),label=node('strong',title),scope=node('span',question.environment??question.definition?.environment??'Environment unavailable'),window=node('span');window.className='resource-return-window';
  const from=question.fromNS??String(BigInt(question.fromMs)*1000000n),to=question.toNS??String(BigInt(question.toMs)*1000000n);
  window.append('Fixed window: ',EvidenceSummary.timestamp(from),' – ',EvidenceSummary.timestamp(to),' (end excluded)');
  back.onclick=()=>{if(active.get(parent)!==owner||own!==revision||bar.hidden)return;clear();onBack()};bar.append(back,label,scope,window);bar.hidden=false;active.set(parent,owner);
 }
 const owner={show,clear,dispose(){clear();bar.remove()}};return owner;
}
export {mount};
