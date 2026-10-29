/* Saved observation adapter. Supplied record components stay passive. */
import * as UI from './components.mjs';
import * as Records from './record-presentation.mjs';
import * as Environment from './record-environment.mjs';
import * as InspectionFocus from './inspection-focus.mjs';
import * as Protocol from './kept-protocol.mjs';
import {client} from './kept-client.mjs';
const n=UI.node;
function mount(host,{active=()=>true}={}){let generation=0,focus;
 return {clear(){generation++;focus?.dispose();host.replaceChildren()},show(value){Protocol.validate(value);generation++;focus?.dispose();const own=generation;host.replaceChildren();const summary=n('goo-evidence-summary');summary.data={mode:'frozen',state:'observed',facts:[{label:'Selected observations',value:String(value.members.length)},{label:'Signal',value:value.selector.signal||'logs'},{label:'Scope',value:Environment.questionLabel(value.selector)},{label:'Kept at',value:value.keptAt,format:'utc'}],details:[{label:'Kept ID',value:value.id},{label:'Content digest',value:value.digest}],notice:'Kept from server-observed records without a telemetry reread. Selection size is not a population count.'};host.append(summary);
 const signal=value.selector.signal||'logs',table=Records.table({signal,labels:['UTC','Service','Record','Inspect'],caption:'Selected original server observations',mobileCompact:true}),details=n('section'),inspector=UI.inspector(details,{onClose:()=>focus.close()});focus=InspectionFocus.mount(details,{active,fallback:()=>host.querySelector('button')});for(const [i,m]of value.members.entries()){const {row,cells}=table.row(m.record,signal),button=n('button','Inspect kept record');button.onclick=()=>{if(!active()||own!==generation)return;const content=Records.detail(m.record,signal,{},value.members.map(x=>x.record)),provenance=n('goo-evidence-summary');provenance.data={mode:'frozen',state:'observed',facts:[{label:'Observed at',value:m.observedAt,format:'utc'},{label:'Read window',value:'['+BigInt(m.fromMs)*1000000n+', '+BigInt(m.toMs)*1000000n+')',format:'unix-ns-range'}],details:[{label:'Occurrence key',value:m.key},{label:'Read basis',value:m.basis},{label:'Coverage',value:m.limited?'Read limited; additional records may exist':'Completeness remains unknown'},{label:'Arrival',value:m.late?'Reported late':'No late flag reported'}],notice:'An observed occurrence, not proof of a unique physical event, complete population or reviewed label.'};content.append(provenance);inspector.show({key:String(i),title:'Kept '+signal+' observation',description:Records.description(m.record),content,opener:button});focus.open(button)};cells[3].append(button);table.body.append(row)}host.append(table.viewport,details)}};
}
function library(host){
 const box=n('details'),body=n('div'),statuses=new Map();
 box.append(n('summary','Kept observations and pending keeps'),n('p','Recent keeps in this browser; exact ID links reopen server storage. Labels and XTDB are not required.'),body);host.append(box);
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
export {mount,library};
