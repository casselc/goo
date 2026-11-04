/* Saved observation adapter. Supplied record components stay passive. */
import * as UI from './components.mjs';
import * as Records from './record-presentation.mjs';
import * as RecordResults from './record-results.mjs';
import * as Environment from './record-environment.mjs';
import * as InspectionFocus from './inspection-focus.mjs';
import * as Catalog from './kept-catalog.mjs';
import * as Protocol from './kept-protocol.mjs';
import {client} from './kept-client.mjs';
const n=UI.node;
function traceView(value,memberIndex){
 const kept=structuredClone(value);Protocol.validate(kept);
 if(!Number.isInteger(memberIndex)||memberIndex<0||memberIndex>=kept.members.length)throw Error('Choose a kept observation.');
 const page=kept.version==='goo/kept-records/v1',member=kept.members[memberIndex],q=page?kept.question:{...kept.selector,fromMs:member.fromMs,toMs:member.toMs};
 return RecordResults.traceView(q,member.record.TraceId);
}
function mount(host,{active=()=>true,onOpenTraceView}={}){let generation=0,focus,inspection=0;
 return {clear(){generation++;focus?.dispose();host.replaceChildren()},show(value){value=structuredClone(value);Protocol.validate(value);const page=value.version==='goo/kept-records/v1',question=page?value.question:value.selector,members=page?value.members.map(m=>({...m,key:(value.origin==='result-page'?'Page index '+m.index+' · original source offset '+question.offset:'Stored row '+m.index),fromMs:question.fromMs,toMs:question.toMs,observedAt:value.observedAt,basis:value.basis,limited:value.recordsLimited})):value.members;generation++;focus?.dispose();const own=generation;host.replaceChildren();const summary=n('goo-evidence-summary');summary.data={mode:'frozen',state:'observed',facts:[{label:'Selected observations',value:String(value.members.length)},{label:'Signal',value:question.signal||'logs'},{label:'Scope',value:Environment.questionLabel(question)},{label:'Kept at',value:value.keptAt,format:'utc'}],details:[{label:'Kept ID',value:value.id},{label:'Content digest',value:value.digest}],notice:(page?(value.origin==='result-page'?'Kept from the original fixed result page':'Kept from the original saved capture'):'Kept from server-observed records')+' without a telemetry reread. Selection size is not a population count.'};host.append(summary);
 const signal=question.signal||'logs',table=Records.table({signal,labels:['UTC','Service','Record','Inspect'],caption:'Selected original server observations',mobileCompact:true}),details=n('section'),inspector=UI.inspector(details,{onClose:()=>{inspection++;return focus.close()}});focus=InspectionFocus.mount(details,{active,fallback:()=>host.querySelector('button')});for(const [i,m]of members.entries()){const {row,cells}=table.row(m.record,signal),button=n('button','Inspect kept record');button.onclick=()=>{if(!active()||own!==generation||document.hidden||!UI.actionVisible(button))return;const memberVisit=++inspection,content=Records.detail(m.record,signal,{},members.map(x=>x.record)),provenance=n('goo-evidence-summary');provenance.data={mode:'frozen',state:'observed',facts:[{label:'Observed at',value:m.observedAt,format:'utc'},{label:'Read window',value:'['+BigInt(m.fromMs)*1000000n+', '+BigInt(m.toMs)*1000000n+')',format:'unix-ns-range'}],details:[{label:page?'Occurrence position':'Occurrence key',value:m.key},{label:'Read basis',value:m.basis},{label:'Coverage',value:m.limited?'Read limited; additional records may exist':'Completeness remains unknown'},...(page?[{label:'Origin',value:value.origin},{label:'Original result state',value:value.state},...(value.captureID?[{label:'Original capture',value:value.captureID},{label:'Capture custody time',value:value.captureAt,format:'utc'}]:[])]:[{label:'Arrival',value:m.late?'Reported late':'No late flag reported'}])],notice:'An observed occurrence, not proof of a unique physical event, complete population or reviewed label.'};content.prepend(provenance);
 if(onOpenTraceView){
  const action=n('button','Open trace as view'),status=n('p',undefined,'evidence-note');let view;
  try{view=traceView(value,i);status.textContent='New trace question · original observation window; kept evidence unchanged.'}catch(e){action.disabled=true;status.textContent='Trace view unavailable · '+e.message}
  const current=()=>own===generation&&memberVisit===inspection&&host.isConnected&&content.isConnected;
  action.onclick=()=>{if(!current()||!active()||document.hidden||!UI.actionVisible(action))return;const scroll=globalThis.scrollY;const onBack=()=>{if(!current()||!active())return false;action.focus({preventScroll:true});globalThis.scrollTo({top:scroll,behavior:'instant'});return true};try{if(onOpenTraceView(structuredClone(view),{current,onBack})===false)status.textContent='Trace view unavailable. Kept observation remains unchanged.'}catch(e){status.textContent='Trace view unavailable · '+e.message}};
  const scope=n('details');scope.append(n('summary','Trace question scope'),n('p','Original record filters are not carried.'),n('pre',JSON.stringify(question.where||{},null,2)));content.prepend(action,status,scope);
 }
 inspector.show({key:i+':'+memberVisit,title:'Kept '+signal+' observation',description:Records.description(m.record),content,opener:button});focus.open(button)};cells[3].append(button);table.body.append(row)}host.append(table.viewport,details)}};
}
function library(host){
 const box=n('details'),body=n('div'),statuses=new Map();
 box.append(n('summary','Kept observations and pending keeps'));Catalog.mount(box);box.append(n('p','Local browser recent links and pending recovery below. Find kept observations above discovers server storage.'),body);host.append(box);
 function render(){
  const focused=body.contains(document.activeElement)?document.activeElement:null,focusID=focused?.dataset.keepID,focusAction=focused?.dataset.keepAction;
  body.replaceChildren();
  try{
   for(const v of client.recent()){const a=n('a',v.count+' kept observations · '+v.keptAt);a.href='#kept='+v.id;a.dataset.keepID=v.id;a.dataset.keepAction='open';const p=n('p');p.append(a);body.append(p)}
   for(const item of client.pending()){
    const id=item.body.requestID,busy=client.busy(id),line=n('p'),status=n('span',(busy?'Keeping observations…':statuses.get(id)||'Keep acknowledgment unresolved')+' · '+id.slice(0,12)+' '),retry=n('button','Retry unchanged keep'),open=n('a','Open by request ID'),forget=n('button','Forget pending request');
    status.tabIndex=-1;status.dataset.keepID=id;status.dataset.keepAction='status';open.href='#kept='+id;
    for(const [control,action]of [[retry,'retry'],[open,'open'],[forget,'forget']]){control.dataset.keepID=id;control.dataset.keepAction=action}
    retry.disabled=forget.disabled=busy;forget.title='Remove only this browser retry record. A committed server keep is preserved; save its exact ID link first.';
    forget.onclick=()=>{client.forget(id);statuses.delete(id);render()};
    retry.onclick=async()=>{try{const v=await client.send(item.body);statuses.delete(id)}catch(e){statuses.set(id,e.message+' · No telemetry reread was made.')}render()};
    line.append(status,retry,open,forget);body.append(line);
   }
   if(!body.childNodes.length)body.append(n('p','No recent or pending keeps in this browser.'));
  }catch(e){body.append(n('p',e.message))}
  if(focusID){const matches=[...body.querySelectorAll('[data-keep-id]')].filter(e=>e.dataset.keepID===focusID),target=matches.find(e=>e.dataset.keepAction===focusAction&&!e.disabled)||matches.find(e=>e.dataset.keepAction==='open')||matches[0];target?.focus({preventScroll:true})}
 }
 box.addEventListener('toggle',()=>{if(box.open)render()});globalThis.addEventListener('goo:keeps-changed',()=>{if(box.open)render()});return {render};
}
export {traceView,mount,library};
