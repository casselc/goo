/* Application adapter: one explicit lane-condition draft, no telemetry acquisition. */
import * as UI from './components.mjs';
import * as FilterEditor from './filter-editor.mjs';
import * as Attributes from './attributes.mjs';
import * as QuestionSession from './question-session.mjs';
import * as Queries from './queries.mjs';
import {nativeControlReady,revealFocus} from './dom-visibility.mjs';

export function mount(host,{signal,current,base,validate,onApply,onChange=()=>{},records=()=>[]}={}){
 const n=UI.node,singular=signal==='logs'?'log':'span',label=signal==='logs'?'Log':'Span',details=n('details',undefined,'paired-lane-filters'),summary=n('summary',label+' filters'),applied=n('small',undefined,'paired-lane-condition'),editorHost=n('fieldset'),legend=n('legend',label+' conditions'),controls=n('div',undefined,'controls'),apply=n('button','Apply '+singular+' filter'),discard=n('button','Discard '+singular+' draft'),status=n('p');
 details.dataset.laneFilter=signal;editorHost.append(legend);apply.type=discard.type='button';status.setAttribute('role','status');controls.append(apply,discard);details.append(summary,editorHost,controls,status);host.append(applied,details);
 let where,revision=0,error=null,disposed=false;
 const allowed=()=>!disposed&&current(),identity=value=>JSON.stringify(FilterEditor.normalize(value)),draft=()=>editor.snapshot(),dirty=()=>!!error||editor.hasUnappliedCELDraft()||identity(draft())!==identity(where);
 function notify(){sync();onChange()}
 const editor=FilterEditor.mount(editorHost,{signal,canMutate:()=>allowed()&&details.open,canGroup:()=>false,canPlot:()=>false,onChange:(value,failure)=>{
  revision++;error=failure;
  if(!error)try{validate(value)}catch(e){error=e.message}
  status.textContent=error||'Draft · Apply reads only '+signal+'.';notify();
 }});
 editor.enableCEL({base:()=>{const q=structuredClone(base()),condition=FilterEditor.normalize(draft());if(condition===undefined)delete q.where;else q.where=condition;return q},context:(q,abortSignal)=>Queries.postJSON('/api/cel/context',{base:q},abortSignal),check:(q,abortSignal)=>Queries.postJSON('/api/cel/plan',q,abortSignal),current:allowed});
 function sync(){
  const ready=allowed(),changed=dirty();details.dataset.phase=changed?'draft':'applied';editorHost.disabled=!ready;
  apply.disabled=!ready||!changed||!!error||editor.hasUnappliedCELDraft();discard.disabled=!ready;discard.hidden=!changed&&!editor.hasCELDraft();
  applied.textContent='Applied · '+(where?QuestionSession.selectorText(where):'All in shared scope');applied.title='Applied '+singular+' condition · '+applied.textContent;
  apply.title='Read a new recent '+singular+' window; the other lane is unchanged.';editor.syncCEL();
 }
 apply.onclick=()=>{
  if(!allowed()||!nativeControlReady(apply)||apply.disabled)return false;
  let candidate;try{candidate=FilterEditor.normalize(draft());validate(candidate)}catch(e){error=e.message;status.textContent=error;notify();return false}
  const own=revision,focused=document.activeElement===apply,lease=()=>allowed()&&revision===own;
  if(onApply(structuredClone(candidate),apply,{current:lease})!==true){if(lease()){status.textContent='Not applied. Draft kept.';notify()}return false}
  if(!lease())return false;where=structuredClone(candidate);error=null;status.textContent='Applied '+singular+' condition.';notify();if(lease()&&focused)revealFocus(summary);return true;
 };
 discard.onclick=()=>{if(!allowed()||!nativeControlReady(discard))return;revision++;editor.resetCEL();editor.set(where,signal);error=null;status.textContent='Applied condition kept.';notify();revealFocus(summary)};
 details.addEventListener('toggle',()=>{if(details.open&&allowed())editor.setSuggestions({...Attributes.suggestionData(records()),label:'Held '+signal+' · incomplete suggestions'});sync()});
 sync();
 return {resolve(){if(!allowed())return false;details.open=true;status.textContent=error||'Apply or discard this draft before resuming.';sync();return !!revealFocus(apply.disabled?discard:apply,summary)},where:()=>structuredClone(where),dirty,condition:()=>({...(where===undefined?{}:{where:structuredClone(where)}),draft:structuredClone(draft()),dirty:dirty(),error}),sync,
  reset(value){revision++;where=structuredClone(value);editor.resetCEL();editor.set(value,signal);error=null;status.textContent='';sync()},
  retire(){revision++;editor.retireCEL();sync()},
  dispose(){disposed=true;revision++;editor.dispose();details.remove()}};
}
