/* Application adapter: one explicit lane-condition draft, no telemetry acquisition. */
import * as UI from './components.mjs';
import * as FilterEditor from './filter-editor.mjs';
import * as Attributes from './attributes.mjs';
import {PredicateDraft} from './predicate-draft.mjs';
import * as ConditionSummary from './condition-summary.mjs';
import * as Queries from './queries.mjs';
import {nativeControlReady,revealFocus} from './dom-visibility.mjs';

export function mount(host,{signal,current,base,validate,onApply,onChange=()=>{},records=()=>[]}={}){
 const n=UI.node,singular=signal==='logs'?'log':'span',label=signal==='logs'?'Log':'Span',details=n('details',undefined,'paired-lane-filters'),summary=n('summary',label+' filters'),applied=n('small',undefined,'paired-lane-condition'),editorHost=n('fieldset'),legend=n('legend',label+' conditions'),controls=n('div',undefined,'controls'),apply=n('button','Apply '+singular+' filter'),discard=n('button','Discard '+singular+' draft'),status=n('p');
 details.dataset.laneFilter=signal;editorHost.append(legend);apply.type=discard.type='button';status.setAttribute('role','status');controls.append(apply,discard);details.append(summary,editorHost,controls,status);host.append(applied,details);
 let disposed=false;const state=new PredicateDraft({validate:where=>{if(where!==undefined||current())validate(where)}});
 const allowed=()=>!disposed&&current(),draft=()=>editor.snapshot(),dirty=()=>state.dirty()||editor.hasUnappliedCELDraft();
 function notify(){sync();onChange()}
 const editor=FilterEditor.mount(editorHost,{signal,canMutate:()=>allowed()&&details.open,canGroup:()=>false,canPlot:()=>false,onChange:(value,failure)=>{
  state.stage(FilterEditor.normalize(draft()),{error:failure,pending:editor.hasUnappliedCELDraft()});const {error}=state.snapshot();status.textContent=error||'Draft · Apply reads only '+signal+'.';notify();
 }});
 editor.enableCEL({base:()=>{const q=structuredClone(base()),condition=FilterEditor.normalize(draft());if(condition===undefined)delete q.where;else q.where=condition;return q},context:(q,abortSignal)=>Queries.postJSON('/api/cel/context',{base:q},abortSignal),check:(q,abortSignal)=>Queries.postJSON('/api/cel/plan',q,abortSignal),current:allowed});
 function sync(){
  const ready=allowed(),changed=dirty(),{error}=state.snapshot();details.dataset.phase=changed?'draft':'applied';editorHost.disabled=!ready;
  apply.disabled=!ready||!changed||!!error||editor.hasUnappliedCELDraft();discard.disabled=!ready;discard.hidden=!changed&&!editor.hasCELDraft();
  ConditionSummary.present(applied,state.applied(),{empty:'All in shared scope'});
  apply.title='Read a new recent '+singular+' window; the other lane is unchanged.';editor.syncCEL();
 }
 apply.onclick=()=>{
  if(!allowed()||!nativeControlReady(apply)||apply.disabled)return false;
  let candidate;try{candidate=state.candidate()}catch(e){status.textContent=e.message;notify();return false}
  const own=state.token(),focused=document.activeElement===apply,lease=()=>allowed()&&state.token()===own;
  if(onApply(structuredClone(candidate),apply,{current:lease})!==true){if(lease()){status.textContent='Not applied. Draft kept.';notify()}return false}
  if(!lease()||!state.accept(own))return false;const committed=state.token();status.textContent='Applied '+singular+' condition.';notify();if(allowed()&&state.token()===committed&&focused)revealFocus(summary);return true;
 };
 discard.onclick=()=>{if(!allowed()||!nativeControlReady(discard))return;state.discard();editor.resetCEL();editor.set(state.draft(),signal);status.textContent='Applied condition kept.';notify();revealFocus(summary)};
 details.addEventListener('toggle',()=>{if(details.open&&allowed())editor.setSuggestions({...Attributes.suggestionData(records()),label:'Held '+signal+' · incomplete suggestions'});sync()});
 sync();
 return {resolve(){if(!allowed())return false;details.open=true;status.textContent=state.snapshot().error||'Apply or discard this draft before resuming.';sync();return !!revealFocus(apply.disabled?discard:apply,summary)},where:()=>state.applied(),dirty,condition:()=>{const value=state.snapshot();return {...(value.applied===undefined?{}:{where:value.applied}),draft:value.draft,dirty:dirty(),error:value.error}},sync,
  reset(value){state.reset(FilterEditor.normalize(value));editor.resetCEL();editor.set(state.draft(),signal);status.textContent='';sync()},
  retire(){state.retire();editor.retireCEL();sync()},
  dispose(){disposed=true;state.retire();editor.dispose();details.remove()}};
}
