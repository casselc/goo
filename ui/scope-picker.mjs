/* Application adapter; shared native disclosure and buttons, no source reads. */
import * as UI from './components.mjs';
import {selectorText} from './question-session.mjs';
import * as Environment from './record-environment.mjs';
import {identity,normalizeIntent} from './scope-history.mjs';
import {nativeControlReady,completionControlReady} from './dom-visibility.mjs';

export function scopeSummary(intent){
 if(intent.kind==='saved-view'){const saved=normalizeIntent(intent);return {primary:saved.title,meta:'Saved view',policy:'Exact revision',window:saved.id.slice(0,12)+'…',exactWindow:'Saved view ID '+saved.id,exact:'Saved view · '+saved.title+' · '+saved.id};}
 const q=intent.question||intent.selector,selection=Environment.selection(q),scope=selection.scope;
 const environment=scope.mode==='named'?JSON.stringify(scope.value):scope.mode==='empty'?'Empty environment':'Unreported environment';
 const kind=intent.kind==='activity'?'Activity':'Fixed question',signal=q.signal==='logs'?'Logs':'Spans';
 const primary=q.where?selectorText(q.where):'All '+signal.toLowerCase();
 let window='Rolling 60s',exactWindow='Rolling last 60 seconds; Start stream explicitly reads a new window.';
 if(intent.kind==='question'){
  const from=new Date(q.fromMs).toISOString(),to=new Date(q.toMs).toISOString();
  window=from.slice(0,10)===to.slice(0,10)?from.slice(0,10)+' · '+from.slice(11,19)+'–'+to.slice(11,19)+' UTC':from.slice(0,19).replace('T',' ')+' → '+to.slice(0,19).replace('T',' ')+' UTC';
  exactWindow='['+from+', '+to+') · end excluded';
 }
 return {primary,meta:kind+' · '+signal+' · '+environment+(intent.kind==='question'&&q.offset>0?' · from record '+(q.offset+1):''),policy:selection.policy==='effective'?'Current / legacy':'Current key',window,exactWindow,exact:primary+' · '+Environment.questionLabel(q)+' · '+exactWindow};
}

export function mount(host,{journal,current,onStage,visits=()=>[],onVisit=()=>{}}){
 const n=UI.node,root=n('details',undefined,'scope-history'),summary=n('summary','Recent / Pinned'),status=n('p'),content=n('div'),tools=n('div',undefined,'controls'),pin=n('button','Pin current scope'),clear=n('button','Clear recent'),reset=n('button','Reset scope history'),label=n('label'),remember=n('input');const search=n('input');search.type='search';search.maxLength=128;search.setAttribute('aria-label','Find recent or pinned scopes');search.placeholder='Find a scope or saved view';remember.type='checkbox';remember.checked=journal.durable;label.append(remember,document.createTextNode('Remember scopes on this browser'));status.setAttribute('role','status');tools.append(pin,clear,reset);root.append(summary,search,content,tools,label,n('small','20 recent · 30 days. 20 pins until removed. Stage loads configuration, never telemetry.'),status);host.prepend(root);
 function action(fn){const focused=globalThis.document.activeElement;try{fn();status.textContent='';}catch(e){status.textContent=e.message}render();if(focused&&!focused.isConnected)summary.focus({preventScroll:true});}
 let revision=0,pending=null,disposed=false;
 function retire(){const old=pending;pending=null;old?.controller.abort();if(old?.button.isConnected){old.button.disabled=false;old.button.textContent='Stage';}root.setAttribute('aria-busy','false');}
 function available(control){return !disposed&&!host.hidden&&nativeControlReady(control);}
 function complete(control){return !disposed&&!host.hidden&&completionControlReady(control);}
 function stageIntent(intent,button,own,row){
  if(own!==revision||!root.open||!row.isConnected||!available(button))return;
  retire();const invocation=pending={controller:new AbortController(),button};
  const current=()=>pending===invocation&&own===revision&&root.open&&row.isConnected&&complete(button);
  let result;try{result=onStage(structuredClone(intent),{opener:button,signal:invocation.controller.signal,current});}catch(error){pending=null;status.textContent=error.message;return;}
  if(!result||typeof result.then!=='function'){if(pending===invocation){pending=null;if(result===true){root.open=false;render();}else status.textContent=result===false?'Current investigation kept.':'Scope could not be staged.';}return;}
  button.disabled=true;button.textContent='Loading…';root.setAttribute('aria-busy','true');status.textContent='Loading saved view…';
  Promise.resolve(result).then(adopted=>{if(pending!==invocation)return;const eligible=current();pending=null;root.setAttribute('aria-busy','false');button.disabled=false;button.textContent='Stage';if(!eligible)return;if(adopted===true){status.textContent='';root.open=false;render();}else status.textContent=adopted===false?'Current investigation kept.':'Scope could not be staged.';},error=>{if(pending!==invocation)return;const eligible=current();pending=null;root.setAttribute('aria-busy','false');button.disabled=false;button.textContent='Stage';if(eligible)status.textContent=error?.message||'Saved view could not be loaded.';});
 }
 function description(intent,{id,visit}={}){
  const text=scopeSummary(intent),body=n('div',undefined,'scope-history-summary'),title=n('strong',text.primary,'scope-history-title'),meta=n('span',text.meta,'scope-history-meta'),window=n('span',text.window,'scope-history-window'),policy=n('span',text.policy,'scope-history-policy');
  body.id=id;body.title=text.exact;title.title=text.primary;meta.append(policy);window.title=text.exactWindow;
  if(visit)body.append(n('span',visit,'scope-history-visit'));
  body.append(title,meta,window);return body;
 }
 function render(){retire();root.setAttribute('aria-busy','false');const own=++revision;content.replaceChildren();const query=search.value?.trim().toLowerCase()||'',items=journal.list().filter(entry=>!query||Object.values(scopeSummary(entry.intent)).join(' ').toLowerCase().includes(query));for(const pinned of [true,false]){const section=n('section'),heading=n('h3',pinned?'Pinned':'Recent');section.append(heading);for(const entry of items.filter(x=>x.pinned===pinned)){
  const row=n('div',undefined,'scope-history-row'),actions=n('div',undefined,'scope-history-actions'),stage=n('button','Stage'),toggle=n('button',pinned?'Unpin':'Pin'),forget=n('button','Forget');row.dataset.kind=entry.intent.kind;row.dataset.pinned=String(pinned);row.dataset.scopeId=identity(entry.intent);
  const title=description(entry.intent,{id:'scope-entry-'+own+'-'+content.children.length+'-'+section.children.length});for(const button of [stage,toggle,forget])button.setAttribute('aria-describedby',title.id);for(const button of [stage,toggle,forget])button.type='button';actions.append(stage,toggle,forget);row.append(title,actions);
  const owns=control=>own===revision&&root.open&&row.isConnected&&available(control);stage.onclick=()=>stageIntent(entry.intent,stage,own,row);toggle.onclick=()=>{if(owns(toggle))action(()=>pinned?journal.unpin(entry.intent):journal.pin(entry.intent));};forget.onclick=()=>{if(owns(forget))action(()=>journal.forget(entry.intent));};section.append(row);
 }if(!section.querySelector('button'))section.append(n('small','None'));content.append(section);}
 const held=n('section');held.append(n('h3','Held in this tab'));for(const visit of visits()){
  const row=n('div',undefined,'scope-history-row scope-history-held'),button=n('button',visit.current?'Current':'Return'),actions=n('div',undefined,'scope-history-actions'),id='scope-visit-'+own+'-'+visit.id;
  button.type='button';
  const state=visit.phase==='applied'?'Held results':visit.phase==='draft'?'Staged view':'Held view';
  let body;if(visit.configuration){body=description(visit.configuration,{id,visit:'Visit '+visit.id+' · '+state});}else{body=n('div',undefined,'scope-history-summary');body.id=id;body.title=visit.title;body.append(n('strong',visit.title,'scope-history-title'));}
  button.dataset.visitId=String(visit.id);button.disabled=visit.current;button.setAttribute('aria-label',(visit.current?'Current visit: ':'Return to ')+visit.title);button.setAttribute('aria-describedby',id);button.onclick=()=>{if(own!==revision||!root.open||!available(button))return;onVisit(visit.id);root.open=false;};actions.append(button);row.append(body,actions);held.append(row);
 }if(!held.querySelector('button'))held.append(n('small','None'));content.append(held);remember.checked=journal.durable;pin.disabled=!current();if(journal.error&&!status.textContent)status.textContent=journal.error;
 }
 for(const button of [pin,clear,reset])button.type='button';
 pin.onclick=()=>{if(!available(pin))return;action(()=>{const intent=current();if(!intent)throw Error('Apply or stage a scope first.');journal.pin(intent);});};clear.onclick=()=>{if(available(clear))action(()=>journal.clearRecent());};reset.onclick=()=>{if(available(reset))action(()=>journal.reset());};remember.onchange=()=>{if(available(remember))action(()=>journal.persistence(remember.checked));};search.oninput=()=>{if(available(search))render();};root.addEventListener('toggle',()=>{if(root.open)render();else retire()});render();return {root,refresh:render,dispose(){disposed=true;retire();revision++;root.setAttribute('aria-busy','false');}};
}
