/* Application-owned rolling view configuration. Opening never executes telemetry. */
import * as DefinitionLibrary from './definition-library.mjs';
import * as QuestionSession from './question-session.mjs';
import * as RecordColumns from './record-columns.mjs';
import * as UI from './components.mjs';
import * as RecordEnvironment from './record-environment.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import * as PairedPlan from './paired-activity-plan.mjs';
import {object,text,array} from './intent-shape.mjs';
import {validateField,validate as validatePredicate} from './predicate-validation.mjs';
const global=globalThis;
const id=/^[a-f0-9]{64}$/;
const own=(value,key)=>Object.hasOwn(value,key);
function layout(value,signal){
 object(value,['columns','fieldColumns','where'],'Save columns, qualified fields and lane conditions only.');
 array(value.columns);if(Object.hasOwn(value,'fieldColumns'))array(value.fieldColumns);
 const checked=RecordColumns.validate(value,['utc','service','summary']);
 for(const field of checked.fieldColumns)validateField(field,signal);
 if(own(value,'where')){
  if(value.where===undefined)throw Error('Omit an unset lane condition.');
  const where=validatePredicate(value.where,signal);
  checked.where=where;
 }
 return checked;
}
function definition(v,identified){
 object(v,['version',...(identified?['id']:[]),'title','selector','columns','fieldColumns','baselineCapture','lanes'],'Invalid Activity view fields.');
 if(v.version!==1||identified&&(!id.test(v.id)||typeof v.id!=='string')||!text(v.title,120,{empty:false})||!v.title.trim())throw Error('Invalid Activity view');
 object(v.selector,['signal','environment','environmentSelector','where'],'Invalid Activity selector fields');
 if(own(v.selector,'signal')&&typeof v.selector.signal!=='string')throw Error('Choose a string Activity signal.');
 const signal=v.selector.signal||'logs';
 if(signal==='both'){
  for(const key of ['columns','fieldColumns','baselineCapture'])if(own(v,key))throw Error('Paired views save independent lane columns without a baseline.');
  object(v.lanes,['logs','spans'],'Choose exactly logs and spans presentation.');
  if(!own(v.lanes,'logs')||!own(v.lanes,'spans'))throw Error('Choose both lane presentations.');
  const lanes={logs:layout(v.lanes.logs,'logs'),spans:layout(v.lanes.spans,'spans')};
  PairedPlan.create({...v.selector,signal:'logs'},{logs:lanes.logs.where,spans:lanes.spans.where});
  return {...structuredClone(v),selector:{...structuredClone(v.selector),signal},lanes};
 }
 if(!['logs','spans'].includes(signal)||own(v,'lanes'))throw Error('Choose one signal or paired lane presentations.');
 validateRecordQuestion({...v.selector,signal,fromMs:0,toMs:60000,limit:100,offset:0});
 if(v.baselineCapture!==undefined&&v.baselineCapture!==''&&(typeof v.baselineCapture!=='string'||!id.test(v.baselineCapture)||signal!=='spans'))throw Error('Saved baseline requires an exact capture ID and spans.');
 return {...structuredClone(v),selector:{...structuredClone(v.selector),signal},...layout({columns:v.columns,...(own(v,'fieldColumns')?{fieldColumns:v.fieldColumns}:{})},signal)};
}
function model(v){return definition(v,true)}
export function input(v){return definition(v,false)}
export function presentation(v){return v.selector.signal==='both'?{lanes:structuredClone(v.lanes)}:{columns:[...v.columns],fieldColumns:structuredClone(v.fieldColumns||[])}}
const signalLabel=signal=>signal==='both'?'Logs + spans':signal==='spans'?'Spans':'Logs';
const laneConditions=v=>v.selector.signal==='both'?['logs','spans'].filter(signal=>v.lanes[signal].where!==undefined).map(signal=>signalLabel(signal)+': '+QuestionSession.selectorText(v.lanes[signal].where)).join(' · '):'';
async function api(path='',body,signal){const r=await fetch('/api/activity-views'+path,{...(body?{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(input(body))}:{}),signal}),v=await r.json();if(!r.ok)throw Error(v.error||'Activity views unavailable');return v}
function share(viewID){if(!id.test(viewID))throw Error('Invalid Activity view ID');const button=UI.node('button','Copy view link');button.onclick=async()=>{try{await navigator.clipboard.writeText(location.origin+location.pathname+'#activity-view='+viewID);button.textContent='Link copied'}catch{button.textContent='Copy unavailable; use the view link'}};return button}
export function scopeFields(value){if(value===undefined||value===null||value==='')return null;return RecordEnvironment.questionFields(typeof value==='string'?{environment:value}:value)}
export function scopeQuery(value){const fields=scopeFields(value);return !fields?{}:fields.environmentSelector!=null?{environmentSelector:JSON.stringify(fields.environmentSelector)}:{environment:fields.environment}}
const scopeLabel=value=>value?RecordEnvironment.questionLabel(value):'';
const scopeIdentity=value=>value?JSON.stringify(RecordEnvironment.selection(value)):'';
function library(host,{environment,embedded=false}={}){
 let appliedScope=scopeFields(environment);
  const n=UI.node,box=n(embedded?'div':'details'),label=appliedScope&&!embedded?'Saved Activity views in this environment policy':embedded?'Saved Activity views in this scope':'Saved Activity views',status=n('p'),search=n('input'),refresh=n('button','Refresh Activity views'),more=n('button','Next Activity views'),first=n('button','First Activity views'),body=n('div');box.className='activity-view-library';if(embedded){box.setAttribute('role','region');box.setAttribute('aria-label','Saved Activity views in applied scope')}box.append(n(embedded?'h3':'summary',label),n('p',embedded?'Exact applied environment policy; other filters may differ. Opening stages a view; Start reads telemetry.':appliedScope?'Exact environment policy; saved filters may target other services. Opening stages; Start reads telemetry.':'Reusable rolling selectors and columns. Opening stages a view; Start reads telemetry.'),refresh,search,status,body,first,more);host.append(box);search.type='search';search.setAttribute('aria-label','Find loaded Activity views');status.setAttribute('role','status');more.hidden=first.hidden=true;let generation=0,cursor='',next='',controller=null,active=false,disposed=false;
  const catalog=DefinitionLibrary.mount(body,{catalogLabel:'Activity views',describe:v=>({family:'activity',revision:v.id,title:v.title,name:RecordEnvironment.questionLabel(v.selector),summary:signalLabel(v.selector.signal)+' · '+RecordEnvironment.questionLabel(v.selector)+(laneConditions(v)?' · '+laneConditions(v):''),search:v.title+' '+RecordEnvironment.questionLabel(v.selector)+' '+QuestionSession.selectorText(v.selector.where)+' '+laneConditions(v)}),detail:raw=>{const v=model(raw),content=n('div'),link=n('a','Open Activity view');link.href='#activity-view='+v.id;content.append(n('p',RecordEnvironment.questionLabel(v.selector)+' · '+signalLabel(v.selector.signal)+' · '+QuestionSession.selectorText(v.selector.where)),...(v.selector.signal==='both'?['logs','spans'].map(signal=>n('p',signalLabel(signal)+': '+(v.lanes[signal].where===undefined?'Shared filter':QuestionSession.selectorText(v.lanes[signal].where))+' · '+v.lanes[signal].columns.join(', '))):[n('p','Columns: '+v.columns.join(', '))]),...(v.baselineCapture?[n('p','Captured latency baseline · '+v.baselineCapture.slice(0,12)+' · loads after explicit Start and Show latency.')]:[]),link,share(v.id));return content}});
  async function load(after=''){if(disposed||embedded&&(!active||!appliedScope))return;controller?.abort();controller=new AbortController();const own=++generation;status.textContent='Loading Activity view configurations…';const q=new URLSearchParams();if(after)q.set('cursor',after);for(const [key,value]of Object.entries(scopeQuery(appliedScope)))q.set(key,value);try{const v=await api(q.size?'?'+q:'',undefined,controller.signal);if(own!==generation)return;const items=v.items.map(model);if(appliedScope&&items.some(item=>!RecordEnvironment.sameQuestionEnvironment(item.selector,appliedScope)))throw Error('Returned Activity view differs from the applied environment policy.');catalog.set(items);cursor=after;next=v.nextCursor||'';first.hidden=!cursor;more.hidden=!next;status.textContent=v.items.length+' views on this page'+(appliedScope?' in '+scopeLabel(appliedScope):'')+'. Search covers this loaded page.'}catch(e){if(own===generation&&e.name!=='AbortError')status.textContent=e.message}}
  if(!embedded)box.addEventListener('toggle',()=>{if(box.open)load();else{generation++;controller?.abort()}});refresh.onclick=()=>load(cursor);more.onclick=()=>load(next);first.onclick=()=>load();search.oninput=()=>catalog.search(search.value);
  function cancel(){generation++;controller?.abort()}
  function setScope(value,visible){if(!embedded||disposed)return;const nextScope=scopeFields(value),changed=scopeIdentity(appliedScope)!==scopeIdentity(nextScope),opened=!active&&visible;appliedScope=nextScope;active=visible;refresh.disabled=!appliedScope;search.disabled=!appliedScope;if(changed){cancel();cursor=next='';search.value='';catalog.clear();catalog.search('');more.hidden=first.hidden=true;status.textContent=appliedScope?'Applied scope: '+scopeLabel(appliedScope)+'. Open Views to load configurations.':'Apply an environment scope to find saved views.'}if(!active){cancel();return}if(changed||opened)load()}
  const changed=()=>{cancel();if(!embedded&&!disposed&&!appliedScope&&location.hash==='#activity-views'){if(box.open)load();else box.open=true}};if(!embedded){queueMicrotask(changed);global.addEventListener('goo:workspace',changed)}return {root:box,setScope,refresh:()=>load(),dispose(){disposed=true;cancel();global.removeEventListener('goo:workspace',changed);catalog.clear();box.remove()}};
 }
const value={model,api,share,library};
export {model,api,share,library};
