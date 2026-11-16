/* Explicit raw acquisition belongs to this application adapter; held display is local. */
import * as UI from './components.mjs';
import * as Attributes from './attributes.mjs';
import * as Evidence from './evidence-summary.mjs';
import * as Refusal from './source-refusal.mjs';
import {renderRefusal} from './source-refusal-view.mjs';
import {compactScalar} from './numeric-values.mjs';
import * as HistogramObservations from './raw-histogram-observations.mjs';
import * as GaugePoints from './raw-gauge-points.mjs';
import './raw-gauge-points-view.mjs';
import {profile} from './raw-observation-profiles.mjs';
import {checkedRequest,checkedResult} from './raw-observation-protocol.mjs';
export {checkedRequest,checkedResult} from './raw-observation-protocol.mjs';
import * as RawRecordContext from './raw-observation-context.mjs';
import {nativeControlReady,completionControlReady,revealFocus} from './dom-visibility.mjs';
import {isSumKind as isSum,temporalityLabel as temporality,sumReportingLabel as reportedSum} from './instrument-reporting.mjs';
const n=UI.node,has=(o,k)=>Object.hasOwn(o,k);

const stable=v=>JSON.stringify(v,(_,x)=>x&&typeof x==='object'&&!Array.isArray(x)?Object.fromEntries(Object.entries(x).sort(([a],[b])=>a.localeCompare(b))):x);
function instance(row){const r=row.stream.resource;return has(r,'service.instance.id')?(r['service.instance.id']||'Empty instance label'):'Instance not reported';}
function recording(row){if(row.recordingState)return row.recordingState;if(row.flagsState!=='valid')return 'unknown';return row.flags&1?'no-recorded-value':row.flags===0?'recorded':'unknown';}
function scalar(row){const amount=row.valueState==='finite'?(compactScalar(row.rawValue)||compactScalar(row.value)):null,state=row.diagnostics.some(d=>d.field==='value'&&d.code==='no-recorded-value')?'not recorded':row.valueState;return amount?amount.text+(row.stream.unit?' '+row.stream.unit:'')+(recording(row)==='unknown'?' · recording unknown':''):'Unavailable · '+state;}
function pointFacts(data,row){
 const histogram=data.instrument.kind==='histogram';
 const start=row.startNS==='0'?{value:'Not reported · raw 0'}:row.startState==='valid'?{value:row.startNS,format:'unix-ns'}:{value:'Unavailable · '+row.startState};
 return [
  {label:'Observation time',value:row.endNS,format:'unix-ns'},
  {label:'Reported start',...start},
  ...(histogram?[{label:'Count',value:row.count??row.countState},{label:'Buckets',value:row.bucketsState},{label:'Recording',value:row.recordingState}]:[{label:'Raw value',value:has(row,'rawValue')?JSON.stringify(row.rawValue):'Not reported'},{label:'Value state',value:row.valueState},{label:'Recording',value:recording(row)}]),
  {label:'Unit',value:row.stream.unit||'Not reported'},
  ...(isSum(data.instrument.kind)?[{label:'Temporality',value:temporality(row.temporality)},{label:'Monotonicity',value:row.monotonic?'Monotonic':'Nonmonotonic'}]:histogram?[{label:'Temporality',value:temporality(row.temporality)}]:[]),
  {label:'Flags',value:row.flagsState==='valid'?String(row.flags):'Unavailable · '+row.flagsState},
 ];
}
function readContext(data,question){
 const range='['+BigInt(question.fromMs)*1000000n+', '+BigInt(question.toMs)*1000000n+')';
 const context=UI.stats([...(isSum(data.instrument.kind)?[{label:'Reported sum',value:reportedSum(data.instrument)}]:[]),{label:'Source read UTC',value:data.readAt,format:'utc'},{label:'Fixed window',value:range,format:'unix-ns-range'}],{density:'compact',className:'raw-read-context'});context.dataset.rawReadContext='';return context;
}
let displayInstance=0;
function renderRows(host,data,{active=()=>true,onRecordContext,contextLease=()=>()=>false}={}){
 const controls=n('div',undefined,'raw-observation-tools'),label=n('label','Find '),find=n('input'),count=n('p'),previous=n('button','Previous observations'),next=n('button','More observations'),viewport=n('goo-table'),table=n('table'),head=n('thead'),header=n('tr'),body=n('tbody'),panel=n('section');find.type='search';find.placeholder='Loaded points';find.title='Find within held source points; no source read.';find.setAttribute('aria-label','Find held observations');label.append(find);count.setAttribute('role','status');previous.type=next.type='button';previous.setAttribute('aria-label','Previous observations');next.setAttribute('aria-label','More observations');previous.textContent='Previous';next.textContent='More';controls.append(label,previous,next,count);viewport.setAttribute('density','compact');viewport.setAttribute('mobile-layout','stacked');viewport.setAttribute('aria-label','Raw '+profile(data.instrument.kind).label+' observations');for(const name of ['Observed UTC',data.instrument.kind==='histogram'?'Count / buckets':'Value','Reported instance','Details']){const th=n('th',name);th.scope='col';header.append(th);}head.append(header);table.append(n('caption','Exported '+profile(data.instrument.kind).label+' observations · repeated rows retained'),head,body);viewport.append(table);host.replaceChildren(controls,viewport,panel);
 const inspector=UI.inspector(panel,{onClose:(key,opener)=>{if(!active())return false;const rowOpener=!pointView?[...body.children].find(row=>row.dataset.occurrenceKey===key)?.querySelector('button'):null;if(revealFocus(opener,rowOpener,{disclosures:true}))return false;const currentMode=displayPicker?.querySelectorAll('input');revealFocus(currentMode?[...currentMode].find(input=>input.checked):find);return false}}),rows=data.rows.map((row,index)=>({row,index,text:stable(row).toLocaleLowerCase()})),contextGroups=new Map();let offset=0,revision=0,selectedKey=null,pointView=false,plot=null,displayPicker=null,pointsReady=false,displayNote=null;
 function inspectOccurrence(index,opener){
  if(!active()||!nativeControlReady(opener))return false;
  const row=rows[index]?.row;if(!row)return false;selectedKey=row.occurrenceKey;if(plot)plot.selectedKey=selectedKey;
  const content=n('div');content.append(UI.stats(pointFacts(data,row),{density:'compact',className:'raw-point-facts'}));if(data.instrument.kind==='histogram')HistogramObservations.detail(content,row,{active,bank:data});Attributes.mountStream(content,row.stream);if(row.diagnostics.length){const list=n('ul');for(const d of row.diagnostics)list.append(n('li',d.field+': '+d.code));content.append(UI.disclosure('Diagnostics',[list]));}content.append(UI.disclosure('Raw source row',[n('pre',JSON.stringify(row.rawRow,null,2))]));
    let actions=[];
    if(onRecordContext){
     for(const key of contextGroups.keys())if(!inspector.has(key))contextGroups.delete(key);
     let group=contextGroups.get(row.occurrenceKey);
     if(!group){const notice=n('p',has(row.stream.resource,'service.instance.id')?(row.stream.resource['service.instance.id']===''?'Reported empty instance label · nearby time context, not causality.':'Reported instance · nearby time context, not causality.'):'Service-wide · instance not reported. Nearby time does not establish cause.','evidence-note'),status=n('p');status.setAttribute('role','status');const buttons=['logs','spans'].map(signal=>{const button=n('button','Nearby '+signal);button.type='button';button.dataset.rawRecordContext=signal;return button});group={buttons,status};contextGroups.set(row.occurrenceKey,group);content.insertBefore(notice,content.firstChild);content.append(status)}
     const current=contextLease(),selected=()=>current()&&!panel.hidden&&panel.dataset.key===row.occurrenceKey;
     for(const button of group.buttons){button.disabled=!current();button.title='Stage nearby '+button.dataset.rawRecordContext+'; Run reads explicitly.';button.onclick=()=>{if(!active()||!nativeControlReady(button)||!selected())return false;try{const detail=RawRecordContext.plan(data,row.occurrenceKey,button.dataset.rawRecordContext);return onRecordContext(detail,button,{current:selected})===true}catch(error){group.status.textContent='Nearby context unavailable · '+error.message;return false}}}
     actions=group.buttons;
    }
    inspector.show({key:row.occurrenceKey,title:(data.instrument.kind==='histogram'?'Histogram':isSum(data.instrument.kind)?'Sum':'Gauge')+' observation '+(index+1),description:row.stream.name,content,actions,opener});
  const heading=panel.querySelector('h4');if(heading){heading.tabIndex=-1;revealFocus(heading)}
  return true;
 }
 function draw(){const epoch=++revision,query=find.value.toLocaleLowerCase(),matched=rows.filter(row=>row.text.includes(query));offset=Math.min(offset,Math.max(0,Math.floor((matched.length-1)/100)*100));body.replaceChildren();
  for(const {row,index}of matched.slice(offset,offset+100)){const tr=n('tr'),time=n('td'),value=n('td'),producer=n('td',instance(row)),action=n('td'),inspect=n('button','Inspect');tr.dataset.occurrenceKey=row.occurrenceKey;if(row.occurrenceKey===selectedKey){tr.dataset.selected='true';tr.setAttribute('aria-current','true')}time.append(Evidence.timestamp(row.endNS));value.textContent=data.instrument.kind==='histogram'?HistogramObservations.summary(row):scalar(row)+(row.stream.unit?'':' · unit not reported');if(data.instrument.kind==='histogram'&&row.countState==='valid')value.title='Exact source count: '+row.count;inspect.type='button';inspect.setAttribute('aria-label','Inspect observation '+(index+1));inspect.onclick=()=>{if(epoch!==revision||pointView)return false;return inspectOccurrence(index,inspect)};action.append(inspect);tr.append(time,value,producer,action);[...tr.children].forEach((cell,i)=>cell.dataset.label=header.children[i].textContent);body.append(tr);}
  previous.hidden=next.hidden=matched.length<=100;previous.disabled=offset===0;next.disabled=offset+100>=matched.length;count.textContent=(matched.length?offset+1:0)+'–'+Math.min(offset+100,matched.length)+' of '+matched.length+(query?' matches · '+rows.length+' loaded':' points')+(!matched.length&&rows.length?' · No matches':'');
 }
 function showDisplay(points){
  if(points&&!pointsReady){plot.data=GaugePoints.dataset(data);pointsReady=true;plot.selectedKey=selectedKey}
  pointView=points;controls.hidden=points;viewport.hidden=points;plot.hidden=!points;
  for(const input of displayPicker.querySelectorAll('input'))input.checked=input.value===(points?'points':'table');
  displayNote.textContent=points?'All '+rows.length+' loaded'+(find.value?' · Table Find not applied':''):'';displayNote.hidden=!points;
  if(!points&&selectedKey){const query=find.value.toLocaleLowerCase(),matched=rows.filter(item=>item.text.includes(query)),at=matched.findIndex(item=>item.row.occurrenceKey===selectedKey);if(at>=0)offset=Math.floor(at/100)*100;else{displayNote.textContent='Selected observation is outside Table Find.';displayNote.hidden=false}draw();}
 }
 if(data.instrument.kind==='gauge'){
  displayPicker=n('fieldset',undefined,'raw-observation-display');displayPicker.append(n('legend','Observation display'));const name='raw-observation-display-'+(++displayInstance);
  for(const [value,text]of [['table','Table'],['points','Points']]){const label=n('label'),input=n('input');input.type='radio';input.name=name;input.value=value;input.checked=value==='table';input.setAttribute('aria-label',text);label.append(input,text);displayPicker.append(label);input.onchange=()=>{if(!active()||!nativeControlReady(input)||!input.checked)return;showDisplay(value==='points')}}
  plot=n('goo-raw-gauge-points');plot.hidden=true;displayNote=n('p',undefined,'raw-observation-display-note');displayNote.hidden=true;controls.after(displayPicker);displayPicker.after(displayNote);displayNote.after(plot);
  plot.addEventListener('goo-raw-gauge-point',event=>{
   if(!pointView||!active()||!nativeControlReady(plot))return;
   const keys=event.detail?.occurrenceKeys;if(!Array.isArray(keys)||keys.length!==1||typeof keys[0]!=='string')return;
   const opener=event.composedPath?.().find(element=>element?.localName==='button'&&plot.contains(element));if(!opener||!nativeControlReady(opener))return;
   const index=rows.findIndex(item=>item.row.occurrenceKey===keys[0]);if(index<0)return;inspectOccurrence(index,opener);
  });
 }
 find.oninput=()=>{if(!active()||!nativeControlReady(find))return;offset=0;draw();};previous.onclick=()=>{if(!active()||!nativeControlReady(previous))return;offset-=100;draw();};next.onclick=()=>{if(!active()||!nativeControlReady(next))return;offset+=100;draw();};draw();if(!rows.length)host.append(n('p','No exported observations returned for this instrument and fixed window.'));return {retireContexts(){for(const group of contextGroups.values())for(const button of group.buttons){button.disabled=true;button.title='Inspect this point again after the read.'}}};
}

export function mount(host,{active=()=>true,post,onRecordContext}={}){
 let generation=0,contextRevision=0,viewer=null,controller=null,lease=null,held=null,readButton=null,lastOpener=null;const section=n('section'),heading=n('h3','Raw observations'),controls=n('div',undefined,'controls'),cancel=n('button','Cancel read'),close=n('button','Close observations'),status=n('p'),receipt=n('div'),display=n('div');section.dataset.rawObservations='';section.hidden=true;status.setAttribute('role','status');cancel.type=close.type='button';cancel.hidden=true;controls.append(cancel,close);section.append(heading,controls,status,receipt,display);host.append(section);
 function stopRead(){generation++;controller?.abort();controller=null;if(readButton)readButton.disabled=false;readButton=null;cancel.hidden=true;}
 function retire(){contextRevision++;viewer?.retireContexts();viewer=null;stopRead();lease=null;held=null;section.hidden=true;display.replaceChildren();receipt.replaceChildren();status.replaceChildren();}
 // Navigation ends acquisition authority, not the fixed observations already held.
 function suspend(){if(!held){retire();return;}const pending=controller!==null;stopRead();if(pending){status.className='';status.textContent='Read canceled by navigation. Previous held observations retained.';}}
 cancel.onclick=()=>{if(!nativeControlReady(cancel))return;status.className='';stopRead();status.textContent=held?'Read canceled. Previous held observations retained.':'Read canceled. No observations adopted.';if(active()&&nativeControlReady(lastOpener))lastOpener.focus({preventScroll:true});};close.onclick=()=>{if(!nativeControlReady(close))return;const origin=lastOpener;retire();if(active()&&nativeControlReady(origin))origin.focus({preventScroll:true});};
 async function read(request,opener){if(!active()||!nativeControlReady(opener))return;let checked;try{checked=checkedRequest(request);}catch(error){return;}const identity=stable(checked);if(lease!==identity){retire();lease=identity;}contextRevision++;viewer?.retireContexts();controller?.abort();const current=++generation,abort=new AbortController(),followFocus=document.activeElement===opener;controller=abort;readButton=opener;lastOpener=opener;opener.disabled=true;section.hidden=false;cancel.hidden=false;status.className='';heading.textContent=(checked.instrument.kind==='histogram'?'Raw histogram observations':isSum(checked.instrument.kind)?'Raw sum observations':'Raw observations')+' · '+checked.instrument.metric;status.textContent='Reading bounded '+profile(checked.instrument.kind).label+' observations…';if(followFocus){heading.tabIndex=-1;heading.scrollIntoView({block:'start',behavior:'instant'});heading.focus({preventScroll:true});}
  try{const reply=await post(profile(checked.instrument.kind).path,checked,abort.signal);if(current!==generation)return;if(!active()||!completionControlReady(opener)||!completionControlReady(section)){retire();return;}if(!reply.ok)throw Refusal.responseError(reply,'Raw observations unavailable.');let data;try{data=checkedResult(reply.data,checked);}catch{throw new Refusal.SourceReadError({kind:'malformed',detail:'The raw observation response did not match its fixed read.'});}held=data;receipt.replaceChildren(readContext(data,checked.question));viewer=renderRows(display,data,{active:()=>active()&&held===data&&lease===identity,onRecordContext,contextLease:()=>{const revision=contextRevision;return ()=>held===data&&lease===identity&&revision===contextRevision&&controller===null}});receipt.append(UI.disclosure('Read scope and limits',[n('p','Source points in this fixed read; duplicates remain separate. No continuous producer identity, evaluation or collection completeness is inferred.')]));status.className='visually-hidden';status.textContent=data.rows.length+' points loaded.';
  }catch(error){if(current!==generation)return;if(!active()||!completionControlReady(opener)||!completionControlReady(section)){retire();return;}const failure=error instanceof Refusal.SourceReadError?error:new Refusal.SourceReadError({kind:error?.name==='AbortError'?'cancelled':error instanceof SyntaxError?'malformed':'unavailable',detail:'The raw observation read could not be completed.'});status.className='';renderRefusal(status,failure,{prefix:held?'Previous held observations retained. ':''});
  }finally{if(current===generation){controller=null;opener.disabled=false;readButton=null;cancel.hidden=true;}}
 }
 return {read,retire,suspend};
}
