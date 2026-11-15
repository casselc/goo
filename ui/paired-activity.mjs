import * as UI from './components.mjs';
import * as Tail from './tail.mjs';
import * as RecordColumns from './record-columns.mjs';
import {object} from './intent-shape.mjs';
import {validateField} from './predicate-validation.mjs';
import * as PairPlan from './paired-activity-plan.mjs';
import * as LaneFilter from './paired-lane-filter.mjs';
import * as RecordEnvironment from './record-environment.mjs';
import * as RecordPresentation from './record-presentation.mjs';
import {nativeControlReady,revealFocus} from './dom-visibility.mjs';

/* One session owner over two ordinary tails. No merged chronology or population. */
let sequence=0;
export function mount(host,{current=()=>true,onOpenFixed,onChange=()=>{}}={}){
 const id='paired-activity-'+(++sequence),root=UI.node('div',undefined,'paired-activity'),grid=UI.node('div',undefined,'paired-activity-lanes'),detail=UI.node('section');
 root.setAttribute('role','group');root.setAttribute('aria-label','Paired live activity');root.append(grid,detail);host.append(root);
 let disposed=false,epoch=0,plan=null,selector=null,inspection=null,lastFocus=null;
 const lanes=new Map(),available=()=>!disposed&&current()&&root.isConnected&&!document.hidden&&!root.closest('[hidden], [aria-hidden="true"], [inert]');
 const inspector=UI.inspector(detail,{onClose:()=>{inspection=null;return true}});
 const close=detail.querySelector('[aria-label="Close selected item details"]'),closeInspection=close.onclick;close.onclick=event=>{if(available()&&nativeControlReady(close))closeInspection.call(close,event)};
 const heading=detail.querySelector('.ui-inspector-heading>h4');heading.id=id+'-detail';heading.tabIndex=-1;detail.setAttribute('aria-labelledby',heading.id);
 function clearInspection(){inspection=null;lastFocus=null;inspector.clear()}
 function retireLane(signal){const lane=lanes.get(signal);lane.epoch++;if(inspection?.signal===signal)clearInspection()}
 function windowText(lane,state){
  lane.owner=state.owner;lane.section.dataset.phase=state.phase;
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
  const content=UI.node('div'),analyze=UI.node('button','Analyze observation window');analyze.type='button';analyze.setAttribute('aria-label','Analyze observation window · '+signal);
  RecordPresentation.attributes(content,captured.record);
  content.append(UI.disclosure('Original observation window',[UI.node('p',new Date(item.readWindow.fromMs).toISOString()+' → '+new Date(item.readWindow.toMs).toISOString())]));
  analyze.disabled=typeof onOpenFixed!=='function';
  analyze.onclick=()=>{
   if(!lease()||!nativeControlReady(analyze))return;
   const question=Tail.recordQuestion(origin,signal,captured),layout=lane.tail.layout();
   pause('Paused for fixed investigation.');lastFocus=analyze;
   onOpenFixed(question,{...layout,columns:[...layout.columns,'context']},analyze);
  };
  inspection={key,signal,opener,analyze,lease};lastFocus=opener;
  inspector.show({key,title:signal==='logs'?'Observed log':'Observed span',description:RecordPresentation.description(captured.record),content,actions:[analyze],opener});
  revealFocus(heading,opener);return true;
 }
 for(const signal of ['logs','spans']){
  const section=UI.node('section',undefined,'paired-activity-lane'),header=UI.node('header',undefined,'paired-activity-heading'),title=UI.node('h3',signal==='logs'?'Logs':'Spans'),window=UI.node('small','Window pending','paired-activity-window'),body=UI.node('div',undefined,'paired-activity-body');
  title.id=id+'-'+signal;section.setAttribute('aria-labelledby',title.id);section.setAttribute('role','region');section.dataset.signal=signal;section.dataset.pairedLane=signal;section.dataset.phase='stopped';header.append(title,window);section.append(header,body);grid.append(section);
  const lane={section,window,tail:null,filter:null,epoch:0};lanes.set(signal,lane);
  lane.tail=Tail.mount(body,{open:true,compact:true,canUse:available,onState:state=>windowText(lane,state),onFind:state=>{section.dataset.findRevision=String(state.find.revision)},onResume:()=>{if(!available()||!plan)return false;if(lane.filter?.dirty()){lane.filter.resolve();return false}retireLane(signal);if(Tail.transportIdentity(lane.tail.scope().selector)!==Tail.transportIdentity(plan[signal])){lane.tail.start(plan[signal]);return false}return true},onInspect:(item,opener)=>inspect(signal,item,opener)});
  lane.filter=LaneFilter.mount(header,{signal,current:()=>available()&&!!plan,base:()=>{const now=Date.now(),window=lane.tail.scope().readWindow||{fromMs:now-60000,toMs:now};return {signal,...RecordEnvironment.questionFields(selector),...window,limit:100,offset:0}},records:()=>lane.tail.snapshot().rows.map(row=>row.record),validate:where=>PairPlan.create(selector,{...conditions(),[signal]:where}),onChange,
   onApply:(where,opener,capability)=>{
    if(!available()||!plan||!capability.current()||!nativeControlReady(opener)||!lane.tail.ready())return false;
    const next=PairPlan.create(selector,{...conditions(),[signal]:where}),own=epoch;retireLane(signal);const laneEpoch=lane.epoch;
    if(!lane.tail.start(next[signal]))return false;
    if(epoch!==own||lane.epoch!==laneEpoch||!capability.current())return false;
    plan[signal]=next[signal];return true;
   }});

 }
 function adopt(value,next=PairPlan.create(value,conditions())){const changed=Tail.canonical(selector)!==Tail.canonical(value);epoch++;for(const lane of lanes.values())lane.epoch++;clearInspection();selector=structuredClone(value);plan=next;if(changed)for(const lane of lanes.values())lane.filter?.retire();return next}
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
 function pause(reason='Streams paused. Resume each lane explicitly.'){if(disposed||!plan)return;for(const lane of lanes.values())lane.tail.pause(reason)}
 function reset(){if(disposed)return;epoch++;plan=null;selector=null;clearInspection();for(const lane of lanes.values()){lane.epoch++;lane.tail.reset();lane.filter?.reset(undefined)}}
 function sync(){const ready=available();for(const lane of lanes.values())lane.tail.setReadBlocked(ready?'':'Apply shared scope before resuming.');for(const lane of lanes.values())lane.filter?.sync();if(inspection)inspection.analyze.disabled=typeof onOpenFixed!=='function'||!inspection.lease()}
 function layout(){return Object.fromEntries([...lanes].map(([signal,lane])=>[signal,{...lane.tail.layout(),...(lane.filter?.where()===undefined?{}:{where:lane.filter.where()})}]))}
 function snapshot(){return {selector:structuredClone(selector),lanes:Object.fromEntries([...lanes].map(([signal,lane])=>[signal,{...lane.tail.snapshot(),owner:lane.owner,signal,condition:lane.filter?.condition()}]))}}
 function focusStart(){lanes.get('logs').tail.focusStart()}
 return {stage,start,ready,layout,pause,reset,snapshot,focusStart,sync,dirty:()=>[...lanes.values()].some(lane=>lane.filter?.dirty()),restoreFocus(){if(!available())return false;const fallback=[...root.querySelectorAll('button')].find(button=>/^(Resume|Retry|Start)/.test(button.textContent)&&nativeControlReady(button));return !!revealFocus(lastFocus,inspection?.opener||fallback)},dispose(){if(disposed)return;disposed=true;epoch++;clearInspection();for(const lane of lanes.values()){lane.filter?.dispose();lane.tail.dispose()}root.remove()}};
}
