/* Application-owned bounded fleet reads. Display components receive supplied data. */
import * as UI from './components.mjs';
import * as Queries from './queries.mjs';
import * as Model from './service-triage-model.mjs';
import * as CatalogSearch from './service-catalog-search.mjs';
import * as Identity from './service-identity.mjs';
import * as Values from './distribution-values.mjs';
import * as EvidenceSummary from './evidence-summary.mjs';
import {nativeControlReady,actionVisible} from './dom-visibility.mjs';

const node=UI.node;
const rankings=[['count-change','Increase in captured spans'],['count','Captured server spans'],['error-rate','Recorded Error fraction'],['mean','Mean duration'],['name','Service name']];
function mount(){
 let scope=null,revision=0,controller=null,pending=false,snapshot=null,error='',mode='signals',rank='count-change',view=null,drawRevision=0;
 function live(control){return !!scope&&view?.current()&&(!control||view.host.contains(control))&&actionVisible(view.host)&&!view.host.closest('[inert]')}
 function stop(message=''){
  revision++;controller?.abort();controller=null;
  if(pending){pending=false;error=message||'Read stopped. Measure again.'}
 }
 function setScope(value){stop();scope=value?structuredClone(value):null;snapshot=null;error='';mode='signals';view=null}
 async function measure(button){
  if(pending||!live(button)||!nativeControlReady(button))return;
  const restoreFocus=document.activeElement===button;let focusTarget=null;
  let plan;try{plan=Model.plan(scope)}catch(e){mode='triage';error=e.message;draw();return}
  stop();const own=revision,request=new AbortController();controller=request;pending=true;snapshot=null;error='';mode='triage';draw();
  if(restoreFocus&&live()){focusTarget=view.host.querySelector('[role="status"]');focusTarget.tabIndex=-1;focusTarget.focus({preventScroll:true})}
  try{
   const read=async q=>{const response=await Queries.postJSON('/api/record-aggregates',q,request.signal);if(!response.ok)throw Error(response.data?.error||'Server-span read failed');return response.data};
   const [current,baseline]=await Promise.all([read(plan.current),read(plan.baseline)]);
   if(own!==revision||request.signal.aborted||!live())return;
   snapshot=Model.admit(plan,current,baseline);
  }catch(e){if(own===revision&&!request.signal.aborted){error=e.message||'Server-span summary unavailable';request.abort()}}
  finally{if(own===revision){const restore=restoreFocus&&document.activeElement===focusTarget;pending=false;controller=null;draw();if(restore&&live())view.host.querySelector('button')?.focus({preventScroll:true})}}
 }
 function draw(){
  if(!view)return;const ownDraw=++drawRevision,ready=control=>ownDraw===drawRevision&&live(control);const {host,signalView,rows:catalogRows,search,openService}=view;
  host.replaceChildren();host.className='service-triage';host.dataset.mode=mode;signalView.hidden=mode==='triage';
  const controls=node('div',undefined,'controls'),measureButton=node('button',pending?'Reading server spans…':'Measure server spans'),signals=node('button','Signal counts');
  measureButton.disabled=pending;measureButton.onclick=()=>{if(ready(measureButton))return measure(measureButton)};
  signals.setAttribute('aria-pressed',String(mode==='signals'));signals.onclick=()=>{if(!nativeControlReady(signals)||!ready(signals))return;const restore=document.activeElement===signals;mode='signals';draw();if(restore&&live())view.host.querySelector('button')?.focus({preventScroll:true})};controls.append(measureButton);if(mode==='triage')controls.append(signals);
  if(snapshot&&mode==='signals'){const summary=node('button','Server summary');summary.onclick=()=>{if(!nativeControlReady(summary)||!ready(summary))return;const restore=document.activeElement===summary;mode='triage';draw();if(restore&&live())view.host.querySelector('button')?.focus({preventScroll:true})};controls.append(summary)}
  if(pending){const cancel=node('button','Cancel summary');cancel.onclick=()=>{if(!nativeControlReady(cancel)||!ready(cancel))return;const restore=document.activeElement===cancel;stop('Summary canceled. Measure again.');draw();if(restore&&live())view.host.querySelector('button')?.focus({preventScroll:true})};controls.append(cancel)}
  host.append(controls);if(mode!=='triage')return;
  const status=node('p',pending?'Reading current and previous windows…':error?'Summary unavailable · '+error:snapshot?'Captured Server spans · current and preceding window · coverage unknown':'Measure server spans to compare this fixed scope.','service-triage-status');status.setAttribute('role','status');host.append(status);
  if(!snapshot)return;
  const windows=node('p',undefined,'service-triage-windows');for(const [label,q]of [['Current',snapshot.plan.current],['Previous',snapshot.plan.baseline]]){const fact=EvidenceSummary.formatFact({format:'unix-ns-range',value:'['+BigInt(q.fromMs)*1000000n+', '+BigInt(q.toMs)*1000000n+')'}),part=node('span',label+' '+fact.text);part.title='['+new Date(q.fromMs).toISOString()+', '+new Date(q.toMs).toISOString()+')';windows.append(part)}host.append(windows);
  const ranking=node('select'),label=node('label','Rank services');ranking.setAttribute('aria-label','Rank services');for(const [value,text]of rankings){const option=node('option',text);option.value=value;ranking.append(option)}ranking.value=rank;label.append(ranking);controls.append(label);
  ranking.onchange=()=>{if(!nativeControlReady(ranking)||!ready(ranking))return;const restore=document.activeElement===ranking;rank=ranking.value;draw();if(restore&&live())view.host.querySelector('select')?.focus({preventScroll:true})};
  const catalogByKey=new Map(catalogRows.map(row=>[Identity.key(row.service,row.namespace),row])),all=Model.rows(snapshot,catalogRows),rows=Model.sortRows(all,rank).filter(row=>CatalogSearch.matches(catalogByKey.get(row.key)||{...row,signals:row.current||row.baseline?{spans:{}}:{}},search));
  const tableWrap=node('goo-table'),table=node('table'),head=node('thead'),headers=node('tr'),body=node('tbody'),titles=['Service','Server spans','Recorded Error','Mean','Investigate'];
  tableWrap.setAttribute('density','compact');tableWrap.setAttribute('mobile-layout','stacked');tableWrap.setAttribute('aria-label','Server-span service triage');table.className='service-triage-table';
  const caption=node('caption','Captured Server-span summaries · '+rows.length+' of '+all.length+' services shown');table.append(caption);for(const title of titles){const th=node('th',title);th.scope='col';headers.append(th)}head.append(headers);table.append(head,body);
  const catalogKeys=new Set(catalogRows.map(row=>Identity.key(row.service,row.namespace)));
  for(const row of rows){
   const tr=node('tr');tr.dataset.service=row.service;tr.dataset.namespace=Identity.key('',row.namespace);const name=node('td'),label=node('strong',row.service||'(empty service name)');name.append(label,node('small',Identity.label(row.namespace),'service-catalog-namespace'));if(row.qualified===false)name.append(node('small','Namespace unspecified in declaration','service-triage-origin'));
   if((row.current||row.baseline)&&!catalogKeys.has(row.key))name.append(node('small',row.current?'Observed in summary read':'Observed in previous window','service-triage-origin'));
   const count=node('td',row.qualified===false?'Unavailable':row.current?.count||'0');if(row.countDelta!==null){const delta=node('small',(BigInt(row.countDelta)>0n?'+':'')+row.countDelta+' vs previous','service-triage-delta');delta.title='Change in captured Server-span record count between adjacent equal-duration windows';count.append(delta)}
   const error=node('td'),current=row.current;
   if(current?.spanStatus&&BigInt(current.count)>0n){error.append(node('span',Values.percent(current.spanStatus.error,current.count)+'%'),node('small',current.spanStatus.error+'/'+current.count,'service-triage-detail'))}else error.textContent='Unavailable';
   const duration=node('td',current?.duration?.state==='observed'?Values.durationMS(current.duration.sumNs,current.duration.count):'Unavailable'),action=node('td'),open=node('button','Open service');
   open.setAttribute('aria-label','Open service '+Identity.name(row.service||'(empty service name)',row.namespace));open.onclick=()=>{if(!nativeControlReady(open)||!ready(open))return;stop();openService(row.service,open,row.namespace)};action.append(open);tr.append(name,count,error,duration,action);[...tr.children].forEach((cell,index)=>cell.dataset.label=titles[index]);body.append(tr);
  }
  tableWrap.append(table);host.append(tableWrap);if(!rows.length)host.append(node('p','No services match this local search.','empty-state'));
  const details=node('details'),summary=node('summary','Summary scope and source evidence'),meaning=node('p','Counts describe captured Server-kind span records. Recorded Error is reported span status; neither it nor missing observations establish request success or service health. Each window is an independent read; no complete upstream population is claimed.');
  const pre=node('pre');pre.textContent=JSON.stringify(snapshot,null,2);details.append(summary,meaning,pre);host.append(details);
 }
 const hidden=()=>{if(document.hidden){const wasPending=pending;stop();if(wasPending)draw()}};document.addEventListener('visibilitychange',hidden);
 return {setScope,attach(host,{signalView,rows,search='',openService,current=()=>true}){view={host,signalView,rows,search,openService,current};draw()},suspend(){const wasPending=pending;stop();if(wasPending)draw()},dispose(){stop();view=null;document.removeEventListener('visibilitychange',hidden)}};
}
export {mount};
