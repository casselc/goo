import * as UI from './components.mjs';
import * as Tail from './tail.mjs';
import * as Chronology from './held-chronology.mjs';
import './held-chronology-view.mjs';
import * as RecordColumns from './record-columns.mjs';
import {object} from './intent-shape.mjs';
import {validateField} from './predicate-validation.mjs';
import * as PairPlan from './paired-activity-plan.mjs';
import * as LaneFilter from './paired-lane-filter.mjs';
import * as RecordEnvironment from './record-environment.mjs';
import * as RecordPresentation from './record-presentation.mjs';
import {nativeControlReady,revealFocus} from './dom-visibility.mjs';

/* Two source owners; an explicit hold can order their immutable observations. */
let sequence=0;
export function mount(host,{current=()=>true,onOpenFixed,onChange=()=>{}}={}){
 const id='paired-activity-'+(++sequence),root=UI.node('div',undefined,'paired-activity'),grid=UI.node('div',undefined,'paired-activity-lanes'),detail=UI.node('section');
 const toolbar=UI.node('div',undefined,'paired-activity-view'),hold=UI.node('button','Hold chronology'),back=UI.node('button','Back to lanes'),notice=UI.node('p',undefined,'paired-activity-view-status'),chronology=UI.node('goo-held-chronology');
 hold.type=back.type='button';hold.title='Pause both streams and order all held records by event time.';hold.disabled=true;back.hidden=true;chronology.hidden=true;notice.hidden=true;notice.setAttribute('role','status');toolbar.append(hold,back,notice);root.dataset.view='lanes';root.setAttribute('role','group');root.setAttribute('aria-label','Paired live activity');root.append(toolbar,grid,chronology,detail);host.append(root);
 let disposed=false,epoch=0,plan=null,selector=null,inspection=null,lastFocus=null,view='lanes',viewRevision=0,frozen=null;
 const originals=new Map(),detailGroups=new Map();
 const lanes=new Map(),available=()=>!disposed&&current()&&root.isConnected&&!document.hidden&&!root.closest('[hidden], [aria-hidden="true"], [inert]');
 const inspector=UI.inspector(detail,{onClose:(key,opener)=>{inspection=null;chronology.selectedKey=null;revealFocus(opener,view==='chronology'?(chronology.querySelector('[data-related-heading]')||chronology.querySelector('input')):hold);return false}});
 const close=detail.querySelector('[aria-label="Close selected item details"]'),closeInspection=close.onclick;close.onclick=event=>{if(available()&&nativeControlReady(close))closeInspection.call(close,event)};
 const heading=detail.querySelector('.ui-inspector-heading>h4');heading.id=id+'-detail';heading.tabIndex=-1;detail.setAttribute('aria-labelledby',heading.id);
 function clearInspection(){inspection=null;lastFocus=null;detailGroups.clear();inspector.clear();chronology.selectedKey=null}
 function retireLane(signal){retireChronology();const lane=lanes.get(signal);lane.epoch++;if(inspection?.signal===signal)clearInspection()}
 function windowText(lane,state){
  lane.owner=state.owner;lane.section.dataset.phase=state.phase;refreshHold();
  const window=state.readWindow;
  if(!window){lane.window.textContent='Window pending';lane.window.removeAttribute('title');return}
  const from=new Date(window.fromMs).toISOString(),to=new Date(window.toMs).toISOString(),sameDay=from.slice(0,10)===to.slice(0,10);
  lane.window.textContent=(sameDay?'':from.slice(0,10)+' ')+from.slice(11,19)+'–'+(sameDay?'':to.slice(0,10)+' ')+to.slice(11,19)+' UTC';
  lane.window.title='Last observed window · ['+from+', '+to+')';
 }
 function inspect(signal,item,opener){
  const lane=lanes.get(signal);if(!available()||!lane||!plan||!nativeControlReady(opener))return false;
  const own=epoch,laneEpoch=lane.epoch,origin=structuredClone(plan[signal]),captured=structuredClone(item),key=own+':'+signal+':'+item.key+':'+item.readWindow.fromMs+':'+item.readWindow.toMs;
  const lease=()=>available()&&epoch===own&&lane.epoch===laneEpoch&&inspection?.key===key;
  for(const old of detailGroups.keys())if(!inspector.has(old))detailGroups.delete(old);
  let group=detailGroups.get(key);
  if(!group){
   const content=UI.node('div'),context=UI.node('goo-record-context'),analyze=UI.node('button','Analyze observation window');context.setAttribute('signal',signal);context.record=captured.record;content.append(context);analyze.type='button';analyze.setAttribute('aria-label','Analyze observation window · '+signal);
   RecordPresentation.attributes(content,captured.record);
   content.append(UI.disclosure('Original observation window',[UI.node('p',new Date(item.readWindow.fromMs).toISOString()+' → '+new Date(item.readWindow.toMs).toISOString()),...(item.observedAt?[UI.stats([{label:'Source read UTC',value:item.observedAt,format:'utc'}],{density:'compact'})]:[])]));
   const related=UI.node('button','Related held records');related.type='button';group={content,analyze,related,context};detailGroups.set(key,group);
  }
  const {content,analyze,related,context}=group;
  context.records=frozen?Chronology.parentRecords(frozen):[];
  related.hidden=view!=='chronology';related.disabled=!RecordPresentation.validID(captured.record.TraceId,32);related.title=related.disabled?'Valid reported trace ID required':'Explore matching reported IDs in held logs and spans; no source read.';
  related.onclick=()=>{if(!lease()||view!=='chronology'||!nativeControlReady(related))return;chronology.showRelated(signal,item.key)};
  analyze.disabled=typeof onOpenFixed!=='function';
  analyze.onclick=()=>{
   if(!lease()||!nativeControlReady(analyze))return;
   const question=Tail.recordQuestion(origin,signal,captured),layout=lane.tail.layout();
   if(!pause('Paused for fixed investigation.')||!lease()||!nativeControlReady(analyze))return;lastFocus=analyze;
   onOpenFixed(question,{...layout,columns:[...layout.columns,'context']},analyze);
  };
  inspection={key,signal,opener,analyze,lease};if(view==='chronology')chronology.selectedKey=signal+':'+item.key;lastFocus=opener;
  inspector.show({key,title:signal==='logs'?'Observed log':'Observed span',description:RecordPresentation.description(captured.record),content,actions:[analyze,related],opener});
  revealFocus(heading,opener);return true;
 }
 for(const signal of ['logs','spans']){
  const section=UI.node('section',undefined,'paired-activity-lane'),header=UI.node('header',undefined,'paired-activity-heading'),title=UI.node('h3',signal==='logs'?'Logs':'Spans'),window=UI.node('small','Window pending','paired-activity-window'),body=UI.node('div',undefined,'paired-activity-body');
  title.id=id+'-'+signal;section.setAttribute('aria-labelledby',title.id);section.setAttribute('role','region');section.dataset.signal=signal;section.dataset.pairedLane=signal;section.dataset.phase='stopped';header.append(title,window);section.append(header,body);grid.append(section);
  const lane={section,window,tail:null,filter:null,epoch:0};lanes.set(signal,lane);
  lane.tail=Tail.mount(body,{open:true,compact:true,canUse:available,onState:state=>windowText(lane,state),onFind:state=>{section.dataset.findRevision=String(state.find.revision)},onResume:()=>{if(!available()||!plan)return false;if(lane.filter?.dirty()){lane.filter.resolve();return false}retireLane(signal);if(Tail.transportIdentity(lane.tail.scope().selector)!==Tail.transportIdentity(plan[signal])){lane.tail.start(plan[signal]);return false}return true},onInspect:(item,opener)=>inspect(signal,item,opener)});
  lane.filter=LaneFilter.mount(header,{signal,current:()=>available()&&!!plan,base:()=>{const now=Date.now(),window=lane.tail.scope().readWindow||{fromMs:now-60000,toMs:now};return {signal,...RecordEnvironment.questionFields(selector),...window,limit:100,offset:0}},records:()=>lane.tail.snapshot().rows.map(row=>row.record),validate:where=>PairPlan.laneSelector(selector,signal,{[signal]:where}),onChange,
   onApply:(where,opener,capability)=>{
    if(!available()||!plan||!capability.current()||!nativeControlReady(opener)||!lane.tail.ready())return false;
    const next=PairPlan.create(selector,{...conditions(),[signal]:where}),own=epoch;retireLane(signal);const laneEpoch=lane.epoch;
    if(!lane.tail.start(next[signal]))return false;
    if(epoch!==own||lane.epoch!==laneEpoch||!capability.current())return false;
    plan[signal]=next[signal];return true;
   }});

 }
 function refreshHold(){
  const observed=[...lanes.values()].some(lane=>{const state=lane.tail?.scope();return state&&(state.retained>0||state.completedRead||state.readWindow)});
  hold.disabled=!available()||!plan||!observed;hold.hidden=view!=='lanes';back.hidden=view!=='chronology';
 }
 function retireChronology(){
  if(view!=='chronology'&&!frozen)return;
  viewRevision++;view='lanes';root.dataset.view=view;frozen=null;originals.clear();chronology.data=null;chronology.hidden=true;grid.hidden=false;clearInspection();refreshHold();
 }
 hold.onclick=()=>{
  if(!available()||!plan||!nativeControlReady(hold)||view!=='lanes')return false;
  const own=epoch,sourcePlan=plan,revision=++viewRevision,laneEpochs=[...lanes.values()].map(lane=>({lane,epoch:lane.epoch}));
  const retained=()=>available()&&epoch===own&&plan===sourcePlan&&viewRevision===revision&&laneEpochs.every(state=>state.lane.epoch===state.epoch);
  notice.hidden=true;if(!pause('Paused for held chronology.')||!retained())return false;
  try{
   const snapshots=Object.fromEntries([...lanes].map(([signal,lane])=>[signal,lane.tail.snapshot()]));
   const next=Chronology.prepare(snapshots);
   for(const signal of ['logs','spans'])if(next.lanes[signal].selector!==null&&Tail.transportIdentity(next.lanes[signal].selector)!==Tail.transportIdentity(plan[signal]))throw Error('Lane scope changed; return to its current observations.');
   if(!retained())return false;
   originals.clear();for(const signal of ['logs','spans'])for(const item of snapshots[signal].rows)originals.set(signal+':'+item.key,{signal,item});
   clearInspection();frozen=next;chronology.data=next;view='chronology';root.dataset.view=view;grid.hidden=true;chronology.hidden=false;refreshHold();revealFocus(chronology.querySelector('input'),back);return true;
  }catch(error){notice.textContent='Chronology unavailable · '+error.message;notice.hidden=false;return false}
 };
 back.onclick=()=>{if(!available()||view!=='chronology'||!nativeControlReady(back))return false;retireChronology();revealFocus(hold);return true};
 chronology.addEventListener('goo-chronology-inspect',event=>{
  if(!available()||view!=='chronology'||!frozen||!nativeControlReady(chronology))return;
  const {signal,key}=event.detail||{};if(!['logs','spans'].includes(signal)||typeof key!=='string')return;
  const opener=event.composedPath?.().find(element=>element?.localName==='button'&&chronology.contains(element)),original=originals.get(signal+':'+key);
  if(!original||!opener||!nativeControlReady(opener))return;
  inspect(signal,original.item,opener);
 });
 chronology.addEventListener('goo-chronology-return',event=>{
  if(!available()||view!=='chronology'||!frozen||event.target!==chronology||!nativeControlReady(chronology))return;
  const selectedKey=event.detail?.selectedKey,original=originals.get(selectedKey);
  if(!original||chronology.selectedKey!==selectedKey)return;
  const opener=[...chronology.querySelectorAll('[data-chronology-inspect]')].find(button=>button.dataset.chronologyInspect===selectedKey)||chronology.querySelector('input');
  if(opener&&nativeControlReady(opener))inspect(original.signal,original.item,opener);
 });
 function adopt(value,next=PairPlan.create(value,conditions())){retireChronology();const changed=Tail.canonical(selector)!==Tail.canonical(value);epoch++;for(const lane of lanes.values())lane.epoch++;clearInspection();selector=structuredClone(value);plan=next;refreshHold();if(changed)for(const lane of lanes.values())lane.filter?.retire();return next}
 function conditions(){return Object.fromEntries([...lanes].map(([signal,lane])=>[signal,lane.filter?.where()]))}
 function stage(value,presentation){
  if(disposed)return false;const checked={},local={};
  if(presentation!==undefined){object(presentation,['logs','spans'],'Choose both lane presentations.');for(const signal of ['logs','spans']){object(presentation[signal],['columns','fieldColumns','where'],'Save lane conditions and qualified fields only.');local[signal]=presentation[signal].where;const {where,...layout}=presentation[signal];checked[signal]=RecordColumns.validate(layout,['utc','service','summary']);for(const field of checked[signal].fieldColumns)validateField(field,signal)}}
  const next=PairPlan.create(value,local);adopt(value,next);for(const [signal,lane]of lanes)lane.filter?.reset(local[signal]);const own=epoch;
  for(const signal of ['logs','spans']){if(epoch!==own)return false;lanes.get(signal).tail.stage(next[signal],checked[signal],{message:'Ready. Start streams explicitly.'})}
  return true;
 }
 function ready(){return available()&&[...lanes.values()].every(lane=>lane.tail.ready())}
 function start(value=selector){
  if(!ready())return false;const next=PairPlan.create(value,conditions());if(!ready())return false;
  adopt(value,next);const own=epoch,started=[];
  // Adopt both local selectors before any read, so withdrawal cannot leave an
  // old sibling bank labelled with the new pair's scope.
  for(const signal of ['logs','spans']){if(epoch!==own)return false;lanes.get(signal).tail.stage(next[signal],undefined,{message:'Ready. Start streams explicitly.'})}
  const withdraw=()=>{if(epoch===own)for(const signal of started)lanes.get(signal).tail.pause('Start withdrawn. Resume explicitly.');return false};
  for(const signal of ['logs','spans']){
   if(epoch!==own||!ready())return withdraw();
   if(!lanes.get(signal).tail.start(next[signal]))return withdraw();started.push(signal);
   if(epoch!==own||!ready())return withdraw();
  }
  return true;
 }
 function pause(reason='Streams paused. Resume each lane explicitly.'){
  if(disposed||!plan)return false;const own=epoch,sourcePlan=plan,states=[...lanes.values()].map(lane=>({lane,epoch:lane.epoch}));
  const retained=()=>!disposed&&epoch===own&&plan===sourcePlan&&states.every(state=>state.lane.epoch===state.epoch);
  for(const {lane}of states){if(!retained())return false;lane.tail.pause(reason)}
  return retained();
 }
 function reset(){if(disposed)return;retireChronology();epoch++;plan=null;selector=null;clearInspection();for(const lane of lanes.values()){lane.epoch++;lane.tail.reset();lane.filter?.reset(undefined)}}
 function sync(){const ready=available();for(const lane of lanes.values())lane.tail.setReadBlocked(ready?'':'Apply shared scope before resuming.');for(const lane of lanes.values())lane.filter?.sync();if(inspection)inspection.analyze.disabled=typeof onOpenFixed!=='function'||!inspection.lease();refreshHold()}
 function layout(){return Object.fromEntries([...lanes].map(([signal,lane])=>[signal,{...lane.tail.layout(),...(lane.filter?.where()===undefined?{}:{where:lane.filter.where()})}]))}
 function snapshot(){return {view,selector:structuredClone(selector),lanes:Object.fromEntries([...lanes].map(([signal,lane])=>[signal,{...lane.tail.snapshot(),owner:lane.owner,signal,condition:lane.filter?.condition()}]))}}
 function focusStart(){lanes.get('logs').tail.focusStart()}
 return {stage,start,ready,layout,pause,reset,snapshot,focusStart,sync,retireHeld:retireChronology,dirty:()=>[...lanes.values()].some(lane=>lane.filter?.dirty()),restoreFocus(){if(!available())return false;const fallback=[...root.querySelectorAll('button')].find(button=>/^(Resume|Retry|Start)/.test(button.textContent)&&nativeControlReady(button));return !!revealFocus(lastFocus,inspection?.opener||fallback)},dispose(){if(disposed)return;disposed=true;epoch++;frozen=null;originals.clear();clearInspection();for(const lane of lanes.values()){lane.filter?.dispose();lane.tail.dispose()}root.remove()}};
}
