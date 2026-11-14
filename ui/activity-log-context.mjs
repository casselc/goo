/* Activity owns acquisition and selection. Charts only receive supplied counts. */
import './time-bars.mjs';
import {ReadGate} from './aggregate-read-gate.mjs';
import {holdSelection} from './aggregate-selection-lease.mjs';
import * as Model from './activity-log-model.mjs';
import * as UI from './components.mjs';
import * as Queries from './queries.mjs';
import * as Values from './distribution-values.mjs';
import {actionVisible,nativeControlReady} from './dom-visibility.mjs';

function mount(host,{controlsHost,current,pauseStream,onExplain,onStage}={}){
 const n=UI.node,toggle=n('button','Show volume'),section=n('section',undefined,'activity-log-context'),controls=n('div',undefined,'activity-log-controls'),refresh=n('button','Refresh'),cadence=n('select'),pause=n('button','Pause'),status=n('p'),chart=n('goo-time-bars'),readInfo=n('p'),severity=n('details'),severityTitle=n('summary','Severity text'),breakdown=n('div'),explain=n('button','Explain'),bandLabel=n('label','Numeric severity '),bandSelect=n('select'),filterBand=n('button','Filter band'),bandHost=n('div',undefined,'activity-log-bands');
 section.setAttribute('aria-label','Activity log volume');section.hidden=true;toggle.setAttribute('aria-expanded','false');toggle.title='Read full-scope log counts separately from the retained stream buffer.';
 explain.setAttribute('aria-label','Explain window');explain.title='Explain window';refresh.setAttribute('aria-label','Refresh log volume');pause.setAttribute('aria-label','Pause log volume');cadence.setAttribute('aria-label','Log volume update cadence');
 for(const [value,label]of [['0','Manual updates'],['10000','Every 10 seconds'],['30000','Every 30 seconds']]){const option=n('option',label);option.value=value;cadence.append(option)}
 chart.setAttribute('range-select','');chart.setAttribute('density','compact');status.setAttribute('role','status');status.className=readInfo.className='activity-log-status';
 bandSelect.setAttribute('aria-label','Numeric severity');filterBand.type='button';filterBand.hidden=true;bandLabel.append(bandSelect);bandHost.append(bandLabel,filterBand);
 controls.append(refresh,cadence,pause,explain);severity.append(severityTitle,breakdown);section.append(controls,bandHost,chart,readInfo,severity,status);controlsHost.append(toggle);host.append(section);
 const gate=new ReadGate();let opened=false,automatic=false,disposed=false,controller=null,timer=null,data=null,selected=null,renderVersion=0,controlVersion=0,bound=null,wasActive=false,readWhilePaused=false,timerRevision=0,band='all',bandReason='';
 const active=c=>!!c?.active&&c.clean&&!!Model.scopeKey(c),streaming=c=>['unknown','partial'].includes(c?.phase);
 const scopeOf=c=>Model.scopeKey(c)||(data&&c?.selector?Model.scopeKey({...c,readWindow:{fromMs:data.question.fromMs,toMs:data.question.toMs}}):null),scope=()=>scopeOf(current());
 const visible=()=>!disposed&&opened&&active(current())&&!section.hidden&&actionVisible(section);
 function explainLabel(label){explain.setAttribute('aria-label',label);explain.title=label}
 function emit(phase){section.dataset.phase=phase;bandSelect.disabled=!data||!visible();filterBand.disabled=!data||!visible()||!!bandReason;section.dataset.auto=String(automatic);section.dataset.generation=String(gate.generation);refresh.disabled=gate.loading||!active(current());cadence.disabled=!active(current());const focused=document.activeElement===pause;pause.disabled=!automatic&&!gate.loading;pause.hidden=pause.disabled;explain.disabled=!data||!visible()||!!bandReason;if(pause.hidden&&focused&&visible()&&nativeControlReady(refresh))refresh.focus({preventScroll:true})}
 function cancel(){timerRevision++;clearTimeout(timer);timer=null;controller?.abort();controller=null;gate.pause()}
 function freeze(message='Held'){automatic=false;cancel();status.textContent=message;emit('paused')}
 function invalidate(message='Refresh for this scope.'){
  cancel();automatic=false;data=selected=null;band='all';bandReason='';bandSelect.replaceChildren();filterBand.hidden=true;delete section.dataset.severityBand;explainLabel('Explain window');renderVersion++;chart.selectedKey=null;chart.data=null;readInfo.replaceChildren();breakdown.replaceChildren();severity.hidden=true;delete section.dataset.fromMs;delete section.dataset.toMs;status.textContent=message;emit('paused');wire();
 }
 function schedule(){const scheduled=++timerRevision,key=bound;clearTimeout(timer);timer=null;if(!disposed&&opened&&automatic&&!gate.loading&&active(current())&&streaming(current()))timer=setTimeout(()=>{if(scheduled!==timerRevision||key!==bound||scope()!==key||!automatic)return;timer=null;load(false)},Number(cadence.value))}
 function allowed(captured,version,key){return visible()&&nativeControlReady(chart)&&data===captured&&renderVersion===version&&scope()===key}
 function render(){
  const captured=data;if(!captured)return;selected=null;band='all';explainLabel('Explain window');const version=++renderVersion,key=scope(),bandCounts=Model.severityBands(captured);let viewRevision=0;
  const all=n('option','All logs · '+captured.count);all.value='all';bandSelect.replaceChildren(all);for(const item of bandCounts){const bounds=item.key==='unspecified'?'0':item.key==='invalid'?'>24':item.min+'–'+item.max,option=n('option',item.label+' '+bounds+' · '+item.count);option.value=item.key;bandSelect.append(option)}bandSelect.value='all';bandSelect.disabled=false;
  const draw=()=>{const own=++viewRevision,selectedBand=band;filterBand.onclick=()=>{if(!nativeControlReady(filterBand)||!allowed(captured,version,key)||own!==viewRevision||band!==selectedBand||band==='all'||!onStage)return;try{const where=Model.bandWhere(captured,selectedBand);holdSelection(()=>nativeControlReady(filterBand)&&allowed(captured,version,key)&&own===viewRevision&&band===selectedBand,()=>{freeze('Held');pauseStream?.()},()=>onStage(where,filterBand))}catch(error){status.textContent=error.message}};bandReason='';try{Model.select(captured,'0','5',band)}catch(error){bandReason=error.message}selected=null;chart.clearRange?.();chart.selectedKey=null;chart.data=Model.timeline(captured,band);const count=band==='all'?captured.count:captured.logSeverity[band],rate=Model.rate(captured,band);readInfo.textContent=(band==='all'?count+' logs':count+' of '+captured.count+' logs')+' · '+rate.text;readInfo.title='Window ['+new Date(captured.question.fromMs).toISOString()+', '+new Date(captured.question.toMs).toISOString()+') · source read '+captured.capturedAt+' · exact rate '+rate.exact;filterBand.hidden=band==='all'||!onStage;filterBand.textContent='Filter '+(bandCounts.find(item=>item.key===band)?.label||'band');filterBand.disabled=!!bandReason;filterBand.title=bandReason||'Stage this exact numeric severity band within the applied scope. Apply reads a new recent window.';explain.disabled=!!bandReason;section.dataset.severityBand=band;explainLabel('Explain window')};draw();
  bandSelect.onchange=()=>{if(!nativeControlReady(bandSelect)||!allowed(captured,version,key))return;const next=bandSelect.value;try{Model.timeline(captured,next);holdSelection(()=>nativeControlReady(bandSelect)&&allowed(captured,version,key),()=>{freeze('Held');pauseStream?.()},()=>{band=next;draw();status.textContent='Held · '+(band==='all'?'all logs':'numeric severity; text is independent.')})}catch(error){bandSelect.value=band;status.textContent=error.message}};
  const begin=()=>holdSelection(()=>allowed(captured,version,key),()=>{freeze('Held');pauseStream?.()});
  const choose=(from,to)=>{if(!allowed(captured,version,key)||!begin())return;try{selected=Model.select(captured,from,to,band);if(from===to)chart.clearRange?.();chart.selectedKey=from===to?from:null;explainLabel('Explain interval');status.textContent='Selected '+new Date(selected.fromMs).toISOString().slice(11,23)+' → '+new Date(selected.toMs).toISOString().slice(11,23)+' UTC';explain.disabled=false}catch(error){status.textContent=error.message}};
  // Capture handlers pause before a shared pointer/keyboard range gesture starts.
  chartBegin=begin;chartChoose=choose;
  const groups=Model.severity(captured),table=n('table'),head=n('thead'),tr=n('tr'),body=n('tbody');for(const title of ['Reported severity text','Captured logs','Share']){const th=n('th',title);th.scope='col';tr.append(th)}head.append(tr);table.append(n('caption','Reported severity text · full window'),head,body);
  for(const group of groups){const row=n('tr'),label=n('td');row.dataset.severity=group.value;label.append(n('code',group.value===''?'Empty severity text':JSON.stringify(group.value)));row.append(label,n('td',group.count),n('td',Values.percent(group.count,captured.count)+'%'));body.append(row)}
  const source=n('details'),pre=n('pre');source.append(n('summary','Volume read details'),n('p','Severity text is exact producer text; severity numbers remain independent. Counts measure the full queried window, not the retained buffer. Independent count/group/time reads do not prove upstream completeness.'));pre.textContent=JSON.stringify(captured,null,2);source.append(pre);breakdown.replaceChildren(table,source);severityTitle.textContent='Severity text · '+groups.length+' values';severity.hidden=false;
  section.dataset.fromMs=String(captured.question.fromMs);section.dataset.toMs=String(captured.question.toMs);
 }
 let chartBegin=()=>{},chartChoose=()=>{};
 chart.addEventListener('pointerdown',()=>chartBegin(),true);chart.addEventListener('focusin',()=>chartBegin());chart.addEventListener('keydown',event=>{if([' ','Enter','ArrowLeft','ArrowRight'].includes(event.key))chartBegin()},true);
 chart.addEventListener('goo-time-range',event=>chartChoose(event.detail.fromKey,event.detail.toKey));chart.addEventListener('goo-time-select',event=>chartChoose(event.detail.key,event.detail.key));
 async function load(explicit=true){
  const c=current();if(!visible()||!active(c)||!explicit&&(!automatic||!streaming(c))||gate.loading)return;
  gate.bind(Model.scopeKey(c));gate.resume();const token=gate.start(),key=gate.scope,q=Model.query(c),request=new AbortController();readWhilePaused=!streaming(c);controller=request;status.textContent='Reading…';emit('loading');
  try{const response=await Queries.postJSON('/api/record-aggregates',q,request.signal);if(!visible()||!active(current())||scope()!==key||!gate.accept(token,key))return;if(!response.ok)throw Error(response.data?.error||'Source unavailable');data=Model.admit(q,response.data);render();status.textContent=data.state==='no_data'?'No captured logs.':automatic?'Updating · '+Number(cadence.value)/1000+'s':'Held';emit('ready')}
  catch(error){if(token===gate.generation&&!disposed&&error.name!=='AbortError'){automatic=false;data=selected=null;band='all';bandReason='';bandSelect.replaceChildren();filterBand.hidden=true;delete section.dataset.severityBand;explainLabel('Explain window');renderVersion++;chart.selectedKey=null;chart.data=null;readInfo.replaceChildren();breakdown.replaceChildren();severity.hidden=true;status.textContent='Volume unavailable · '+error.message;emit('error')}}
  finally{if(token===gate.generation){controller=null;readWhilePaused=false;gate.loading=false;refresh.disabled=!active(current());schedule()}}
 }
 function wire(){
  const version=++controlVersion,key=bound,ready=(control,clean=true)=>!disposed&&version===controlVersion&&key===bound&&scope()===key&&nativeControlReady(control)&&!!current()?.active&&(!clean||current().clean);
  toggle.onclick=()=>{if(!ready(toggle,!opened))return;opened=!opened;section.hidden=!opened;toggle.textContent=opened?'Hide volume':'Show volume';toggle.setAttribute('aria-expanded',String(opened));if(opened){gate.bind(scope());return load()}else{invalidate('Volume closed.');toggle.disabled=!active(current())}};
  refresh.onclick=()=>{if(!ready(refresh)||!opened)return;automatic=Number(cadence.value)>0&&streaming(current());return load()};
  cadence.onchange=()=>{if(!ready(cadence)||!opened)return;automatic=Number(cadence.value)>0&&streaming(current());if(automatic){schedule();status.textContent='Updating · '+Number(cadence.value)/1000+'s';emit('paused')}else freeze()};
  pause.onclick=()=>{if(ready(pause)&&opened)freeze()};
  explain.onclick=()=>{if(!ready(explain)||!visible()||!data||bandReason)return;const captured=data,version=renderVersion,key=scope(),selectedBand=band;try{const parent=Model.select(captured,'0','5',band),selection=selected?structuredClone(selected):structuredClone(parent);holdSelection(()=>ready(explain)&&allowed(captured,version,key)&&band===selectedBand,()=>{freeze('Held');pauseStream?.()},()=>onExplain?.({parent,selection,...(band!=='all'?{logSeverityBand:band}:{})},explain))}catch(error){status.textContent=error.message}};
 }
 function sync(){
  const c=current(),key=scopeOf(c);section.hidden=!opened||(c?.selector?.signal||'logs')!=='logs'||!c?.active;toggle.hidden=(c?.selector?.signal||'logs')!=='logs';
  if(key!==bound){invalidate();bound=key;gate.bind(key);wire()}
  toggle.disabled=!opened&&!active(c);bandSelect.disabled=!data||!active(c)||section.hidden;filterBand.disabled=!data||!active(c)||section.hidden||!!bandReason;explain.disabled=!data||!active(c)||section.hidden||!!bandReason;refresh.disabled=!active(c)||gate.loading;cadence.disabled=!active(c);
  if(!c?.active){if(wasActive)freeze('Held');wasActive=false;return}wasActive=true;
  if(!c.clean){if(automatic||gate.loading)freeze('Held · unapplied scope');return}
  if(!streaming(c)&&(automatic||gate.loading&&!readWhilePaused))freeze('Held');
  explain.disabled=!data||!visible()||!!bandReason;
 }
 wire();sync();
 return {sync,invalidate,dispose(){disposed=true;controlVersion++;renderVersion++;cancel();toggle.remove();section.remove()}};
}
export {mount};
