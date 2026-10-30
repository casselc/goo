/* App adapter over captured raw-span duration bins; hosts own range effects. */
import './scope-view.mjs';
import {asSnapshot} from './snapshots.mjs';
import {validateSnapshot} from './view-model.mjs';
import * as PopulationSelection from './population-selection.mjs';
import {identity} from './question-session.mjs';
import {node} from './components.mjs';

function durationSnapshot(data){
 try{const snapshot=asSnapshot(data?.histogramDataset);return snapshot.shape==='histogram'&&snapshot.histogram?.kind==='span-duration'&&snapshot.histogram.total===data.count&&identity({...snapshot.scope.selector,limit:undefined,offset:undefined})===identity({...data.question,limit:undefined,offset:undefined})&&validateSnapshot(snapshot).ok?snapshot:null}catch{return null}
}
function mount(host,data,{open,readOnly=false,onSelect,onBegin,compact=false,canAct=()=>false,resolveRange}={}){
 const snapshot=durationSnapshot(data);if(!snapshot)return false;
 const section=node('section',undefined,'latency-histogram'),chart=node('goo-chart');section.setAttribute('aria-label','Span latency histogram');chart.snapshot=snapshot;chart.display={kind:'histogram',title:'Span latency distribution'};section.append(chart);host.append(section);
 let paused=false,owner=null,preview=null;const allowed=()=>!paused&&!readOnly&&section.isConnected&&canAct();
 if(!readOnly&&(onSelect||open)&&resolveRange){
  const captured={...data,question:snapshot.scope.selector,count:snapshot.histogram.total};
  if(open&&!onSelect){preview=node('p');preview.setAttribute('role','status');section.append(preview);owner=PopulationSelection.mount(section,captured,{open,onBaseline:true,compact:true,canAct:allowed});owner.element.addEventListener('input',()=>{preview.textContent='Exact bounds edited. Run to measure this population.';chart.clearRange()})}
  const begin=()=>{if(allowed())onBegin?.()};chart.addEventListener('pointerdown',begin,true);chart.addEventListener('keydown',event=>{if([' ','Enter','ArrowLeft','ArrowRight'].includes(event.key))begin()},true);
  chart.addEventListener('goo-histogram-range',event=>{
   if(!allowed())return;const range=resolveRange(event.detail);if(!range?.ok||!allowed())return;
   const bounds={fromMs:range.question.fromMs,toMs:range.question.toMs,minMS:range.minMS,maxMS:range.maxMS};
   if(onSelect)onSelect(bounds,{count:range.count,total:range.total});else{preview.textContent=range.count+' of '+range.total+' captured spans selected · '+String(BigInt(range.total)-BigInt(range.count))+' outside the selected buckets';owner.select(bounds);}
  });
 }
 const pause=()=>{paused=true;owner?.suspend?.()};
 return {chart,suspend:pause,suspendActions:pause,invalidate(){paused=true;owner?.invalidate?.()}};
}
export {durationSnapshot,mount};
