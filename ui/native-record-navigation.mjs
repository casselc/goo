/* App-owned retained native context. No source reads, result mutation or custody. */
import * as EvidenceSummary from './evidence-summary.mjs';
import * as UI from './components.mjs';
import * as Workspace from './workspace.mjs';
const global=globalThis;
function visibleWithin(opener,host,active){
  if(global.document?.hidden||!active||!opener?.isConnected||!host?.contains(opener)||!opener.getClientRects().length)return false;
  for(let e=opener;e;e=e.parentElement){if(e.hidden||e.inert||e.getAttribute?.('aria-hidden')==='true')return false;if(e.localName==='details'&&!e.open&&!e.querySelector(':scope>summary')?.contains(opener))return false}
  return true;
 }
function mount(host,{nativeHost,active,stageQuestion,showNative,suspendRecords,holdResourceOrigin=()=>null}={}){
  const node=UI.node,button=node('button','Back to measurement'),summary=node('span','Interval context · trace expansion is broader'),scope=node('details'),body=node('div');
  host.className='native-record-return';host.setAttribute('aria-label','Native interval context');const controls=node('div');controls.className='controls';controls.append(button,summary);scope.append(node('summary','Context scope · trace expansion is broader'),body);host.append(controls,scope);host.hidden=true;
  let parent=null,resourceOrigin=null;
  function releaseOrigin(){resourceOrigin?.restore();resourceOrigin=null}
  const valid=()=>!!parent&&parent.opener.isConnected&&nativeHost.contains(parent.opener);
  function clear(){releaseOrigin();parent=null;host.hidden=true;body.replaceChildren()}
  function restore(hash,{replace=false}={}){
   if(!valid()){clear();return false}if(hash!==parent.hash)return false;
   if(!parent.returned){parent.returned=true;suspendRecords();showNative();host.hidden=true;releaseOrigin();parent.opener.focus({preventScroll:true});global.scrollTo({top:parent.scroll,behavior:'instant'})}
   if(replace)history.replaceState(null,'',parent.hash);return true;
  }
  button.onclick=()=>{if(parent&&!active()&&visibleWithin(button,host,true))restore(parent.hash,{replace:true})};
  function stage(detail,opener){
   if(!visibleWithin(opener,nativeHost,active()))return false;
   const href=Workspace.questionLink(detail.question),exemplar=detail.relation==='reported-exemplar';host.setAttribute('aria-label',exemplar?'Reported exemplar context':'Native interval context');scope.querySelector('summary').textContent=exemplar?'Reported exemplar scope and limits':'Context scope · trace expansion is broader';
   parent={hash:location.hash,opener,scroll:global.scrollY,returned:false};resourceOrigin=resourceOrigin||holdResourceOrigin(detail.measurement);body.replaceChildren();
   summary.className='native-context-path';summary.textContent=[resourceOrigin?.title,detail.measurement?.definition?.title||'Measurement',exemplar?'Reported trace exemplar':detail.question.signal==='spans'?'Spans in interval':'Logs in interval'].filter(Boolean).join(' → ');
   if(resourceOrigin){const origin=node('p');const q=resourceOrigin.question;origin.append(resourceOrigin.title+' · original window: ',EvidenceSummary.timestamp(q.fromNS??String(BigInt(q.fromMs)*1000000n)),' – ',EvidenceSummary.timestamp(q.toNS??String(BigInt(q.toMs)*1000000n)),' (end excluded).');body.append(origin)}
   const bounds=node('p');bounds.append(exemplar?'Original measurement window: ':'Native interval: ',EvidenceSummary.timestamp(detail.fromNS),' – ',EvidenceSummary.timestamp(detail.toNS),' (end excluded).');
   if(exemplar){const exported=node('p');exported.append('Reported source export: ',EvidenceSummary.timestamp(detail.exportStartNS),' – ',EvidenceSummary.timestamp(detail.exportEndNS),' · sample time: ',EvidenceSummary.timestamp(detail.timeNS));body.append(bounds,exported,node('p','Reported trace '+detail.traceId+(detail.spanId?' · reported span hint '+detail.spanId:' · no valid span hint')+'. Run explicitly to read spans by reported trace ID and exact namespace in the original measurement window and environment. Service, instance, metric and metric-attribute filters are omitted. The span hint is not a predicate.'),node('p','This is a reported association, not proof of causality, complete trace capture or bucket contribution. A cumulative sample may predate a derived delta interval.'))}else{
   body.append(bounds,node('p',detail.measurement?.definition?.source?.resource?.kind==='host'?'Records match reported host.id and environment within the exact interval. Other SDK/resource maps may match that ID; this does not prove one physical host. Metric attributes and instrumentation scope are not record filters.':'Records match the reported service, resource namespace and instance. Metric name, metric attributes and instrumentation scope are not record filters. Shared time does not establish cause.'),node('p','Trace expansion is broader: a fresh read by reported trace ID in the enclosing millisecond window may include other instances, namespaces and records just outside the exact native interval. The exact record question remains its origin.'));
   }
   const provenance=node('details'),receipt=node('pre');provenance.append(node('summary','Retained native origin'),receipt);receipt.textContent=JSON.stringify({question:detail.measurement,stream:detail.stream,basis:detail.basis,readAt:detail.readAt,...(exemplar?{reference:{sourceRow:detail.sourceRow,exportStartNS:detail.exportStartNS,exportEndNS:detail.exportEndNS,exemplarIndex:detail.exemplarIndex,timeNS:detail.timeNS,traceId:detail.traceId,...(detail.spanId?{spanId:detail.spanId}:{})}}:{})},null,2);body.append(provenance,node('small','This native origin is retained in this tab. Sharing or saving the record question preserves its filters, not a durable citation to this native result.'));
   host.dataset.fromNs=detail.fromNS;host.dataset.toNs=detail.toNS;host.dataset.signal=detail.question.signal;host.hidden=false;
   stageQuestion(structuredClone(detail.question));history.pushState(null,'',href);return true;
  }
  global.addEventListener('goo:workspace',event=>{if(event.detail.route!=='explore')clear()});
  global.addEventListener('goo:explore-mode',event=>{if(parent&&!valid()){clear();return}if(event.detail.measurements){if(valid()){parent.returned=true;suspendRecords()}releaseOrigin();host.hidden=true}else host.hidden=!valid()||parent.returned});
  return {stage,restore,clear};
 }

export {visibleWithin,mount};
