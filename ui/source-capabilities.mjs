/* Explicit fixed-adapter checks. Display elements never acquire source data. */
import * as UI from './components.mjs';
import * as Queries from './queries.mjs';
import * as SourceRefusal from './source-refusal.mjs';
import {nativeControlReady,completionControlReady} from './dom-visibility.mjs';

const families=Object.freeze([
  {kind:'logs',table:'otel_logs',label:'Logs',unit:'records'},
  {kind:'spans',table:'otel_traces',label:'Spans',unit:'records'},
  {kind:'gauge',table:'otel_metrics_gauge',label:'Gauges',unit:'exported-points'},
  {kind:'counter',table:'otel_metrics_sum',label:'Counters',unit:'exported-points'},
  {kind:'histogram',table:'otel_metrics_histogram',label:'Histograms',unit:'exported-points'},
  {kind:'exponential-histogram',table:'otel_metrics_exp_histogram',label:'Exponential histograms',unit:'exported-points'},
]);
const states=new Set(['compatible','absent','incompatible','unreadable']);
const observations=new Set(['observed','empty','unreadable','unknown']);
const object=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
function keys(value,allowed){return object(value)&&Object.keys(value).every(key=>allowed.includes(key))}
function boundedText(value){return typeof value==='string'&&value.isWellFormed()&&new TextEncoder().encode(value).length<=256&&!/\p{Cc}/u.test(value)}
function validCapturedAt(value){return typeof value==='string'&&value.length<=64&&/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?Z$/.test(value)&&Number.isFinite(Date.parse(value))&&new Date(value).toISOString().slice(0,19)===value.slice(0,19)}
function validateResult(value,question){
  const invalid=()=>{throw Error('Source capability response rejected.');};
  if(!keys(value,['version','capturedAt','question','families'])||!validCapturedAt(value.capturedAt)||value.version!=='goo/source-capabilities/v1'||!keys(value.question,['fromMs','toMs'])||value.question.fromMs!==question.fromMs||value.question.toMs!==question.toMs||!Array.isArray(value.families)||value.families.length!==families.length)invalid();
  const checked=new Map();
  for(const row of value.families){
    if(!keys(row,['kind','table','state','detail','observation']))invalid();
    const profile=families.find(family=>family.kind===row.kind);
    if(!profile||checked.has(row.kind)||row.table!==profile.table||!states.has(row.state)||!boundedText(row.detail)||!keys(row.observation,['state','unit'])||!observations.has(row.observation.state)||row.observation.unit!==profile.unit)invalid();
    if(row.state!=='compatible'&&row.observation.state!=='unknown')invalid();
    checked.set(row.kind,structuredClone(row));
  }
  return {version:value.version,capturedAt:value.capturedAt,question:{...question},families:families.map(profile=>checked.get(profile.kind))};
}
function statusOf(family){
  if(family.state==='absent')return 'Table absent';
  if(family.state==='incompatible')return 'Unsupported schema';
  if(family.state==='unreadable')return 'Source unavailable';
  const metric=['gauge','counter','histogram','exponential-histogram'].includes(family.kind);
  if(family.observation.state==='observed')return metric?'Points observed':'Records observed';
  if(family.observation.state==='empty')return metric?'No exported points in window':'No records in window';
  if(family.observation.state==='unreadable')return 'Read unavailable';
  return 'Not checked for records';
}
// Catalogs disclose excluded families without interpreting them as empty data.
function sourceNote(sources){
  if(!Array.isArray(sources))return '';
  const reasons={absent:'table absent',incompatible:'unsupported schema',unreadable:'source unavailable'};
  const excluded=sources.flatMap(source=>{const family=families.find(f=>f.kind===source.kind),reason=reasons[source.state];return family&&reason?[family.label+' ('+reason+')']:[]});
  return excluded.length?'Sources not included: '+excluded.join(', ')+'. Available sources only.':'';
}
function mount(host,{minutes,onRead,active=()=>true,onSelect,onOpenMetrics}={}){
  const n=UI.node,section=n('section',undefined,'source-capabilities'),controls=n('div',undefined,'controls'),read=n('button','Read available signals'),status=n('p'),output=n('div');
  section.setAttribute('aria-label','Available telemetry signals');
  read.type='button';status.setAttribute('role','status');status.className='source-capability-status';
  controls.append(read);section.append(controls,status,output);host.append(section);
  let disposed=false,generation=0,controller=null,adopted=null;
  const current=control=>!disposed&&active()&&!document.hidden&&nativeControlReady(control);
  function clear(){generation++;controller?.abort();controller=null;adopted=null;read.disabled=false;output.replaceChildren();status.textContent='';section.removeAttribute('aria-busy')}
  function render(result,g){
    const viewport=n('goo-table'),table=n('table'),caption=n('caption','Across all services and environments · '+new Date(result.question.fromMs).toISOString()+' → '+new Date(result.question.toMs).toISOString()),head=n('thead'),row=n('tr'),body=n('tbody');
    viewport.className='source-capability-table';viewport.setAttribute('density','compact');viewport.setAttribute('aria-label','Supported telemetry sources');
    for(const title of ['Signal','Source state','Use']){const th=n('th',title);th.scope='col';row.append(th)}head.append(row);
    for(const family of result.families){
      const profile=families.find(item=>item.kind===family.kind),tr=n('tr'),name=n('th',profile.label),state=n('td'),action=n('td'),label=n('span',statusOf(family));
      name.scope='row';label.dataset.state=family.state;state.append(label);
      const unit=profile.unit==='records'?'Records':'Exported points';
      if(family.state==='compatible')state.append(n('small',unit+' · presence check'));
      if(family.state==='compatible'&&['observed','empty'].includes(family.observation.state)&&['logs','spans'].includes(family.kind)&&onSelect){
        const choose=n('button','Choose '+family.kind);choose.type='button';
        choose.onclick=()=>{if(!current(choose)||g!==generation||adopted!==result)return;if(onSelect(family.kind)!==false)clear()};action.append(choose);
      }
      tr.append(name,state,action);body.append(tr);
    }
    table.append(caption,head,body);viewport.append(table);const details=n('details'),summary=n('summary','Source details'),list=n('ul');for(const family of result.families)list.append(n('li',family.table+' · '+family.detail));details.append(summary,n('p','Checked '+result.capturedAt),list);const metricAction=n('div',undefined,'controls');if(onOpenMetrics&&result.families.some(f=>['gauge','counter','histogram','exponential-histogram'].includes(f.kind)&&f.state==='compatible'&&['observed','empty'].includes(f.observation.state))){const open=n('button','Find metric scopes');open.type='button';open.onclick=()=>{if(!current(open)||g!==generation||adopted!==result)return;if(onOpenMetrics({...result.question})!==false)clear()};metricAction.append(open)}output.replaceChildren(viewport,metricAction,details);
  }
  read.onclick=async()=>{
    if(!current(read))return;
    const count=Number(minutes.value);if(!Number.isInteger(count)||count<1||count>120){status.textContent='Choose 1–120 minutes.';return}
    onRead?.();if(!current(read))return;
    clear();const g=generation,toMs=Date.now(),question={fromMs:toMs-count*60000,toMs},own=new AbortController();controller=own;read.disabled=true;section.setAttribute('aria-busy','true');status.textContent='Checking supported signals…';
    try{
      const reply=await Queries.postJSON('/api/source-capabilities',question,own.signal);
      if(g!==generation||disposed||!active()||document.hidden||!completionControlReady(read))return;
      if(!reply.ok)throw SourceRefusal.responseError(reply,'Source checks unavailable.');
      adopted=validateResult(reply.data,question);render(adopted,g);status.textContent=adopted.families.some(family=>family.state==='compatible'&&['observed','empty'].includes(family.observation.state)&&['logs','spans'].includes(family.kind))?'Choose Logs or Spans to find environments. Coverage remains unknown.':'Signal checks complete. Coverage remains unknown.';
    }catch(error){
      if(g!==generation||disposed||error.name==='AbortError'||!active()||document.hidden||!completionControlReady(read))return;
      status.textContent=error instanceof SourceRefusal.SourceReadError?error.message:'Source checks unavailable.';
    }finally{
      // Release this request's private busy state even when publication was withdrawn.
      if(controller===own){controller=null;read.disabled=false;section.removeAttribute('aria-busy')}
    }
  };
  const changed=()=>clear();minutes.addEventListener('input',changed);
  return {clear,dispose(){clear();disposed=true;minutes.removeEventListener('input',changed);section.remove()}};
}
export {families,validateResult,statusOf,sourceNote,mount};
