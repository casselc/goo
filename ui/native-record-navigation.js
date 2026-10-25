/* App-owned retained native context. No source reads, result mutation or custody. */
(function(global){
 function visibleWithin(opener,host,active){
  if(global.document?.hidden||!active||!opener?.isConnected||!host?.contains(opener)||!opener.getClientRects().length)return false;
  for(let e=opener;e;e=e.parentElement){if(e.hidden||e.getAttribute?.('aria-hidden')==='true')return false;if(e.localName==='details'&&!e.open&&!e.querySelector(':scope>summary')?.contains(opener))return false}
  return true;
 }
 function mount(host,{nativeHost,active,stageQuestion,showNative,suspendRecords}={}){
  const node=global.GooUI.node,button=node('button','Back to measurement'),summary=node('span','Interval context · trace expansion is broader'),scope=node('details'),body=node('div');
  host.className='native-record-return';host.setAttribute('aria-label','Native interval context');const controls=node('div');controls.className='controls';controls.append(button,summary);scope.append(node('summary','Context scope'),body);host.append(controls,scope);host.hidden=true;
  let parent=null;
  const valid=()=>!!parent&&parent.opener.isConnected&&nativeHost.contains(parent.opener);
  function clear(){parent=null;host.hidden=true;body.replaceChildren()}
  function restore(hash,{replace=false}={}){
   if(!valid()){clear();return false}if(hash!==parent.hash)return false;
   if(!parent.returned){parent.returned=true;suspendRecords();showNative();host.hidden=true;parent.opener.focus({preventScroll:true});global.scrollTo({top:parent.scroll,behavior:'instant'})}
   if(replace)history.replaceState(null,'',parent.hash);return true;
  }
  button.onclick=()=>{if(parent&&!active()&&visibleWithin(button,host,true))restore(parent.hash,{replace:true})};
  function stage(detail,opener){
   if(!visibleWithin(opener,nativeHost,active()))return false;
   const href=global.GooWorkspace.questionLink(detail.question);
   parent={hash:location.hash,opener,scroll:global.scrollY,returned:false};body.replaceChildren();
   const bounds=node('p');bounds.append('Native interval: ',global.GooEvidenceSummary.timestamp(detail.fromNS),' – ',global.GooEvidenceSummary.timestamp(detail.toNS),' (end excluded).');
   body.append(bounds,node('p','Records match the reported service, resource namespace and instance. Metric name, metric attributes and instrumentation scope are not record filters. Shared time does not establish cause.'),node('p','Trace expansion is broader: a fresh read by reported trace ID in the enclosing millisecond window may include other instances, namespaces and records just outside the exact native interval. The exact record question remains its origin.'));
   const provenance=node('details'),receipt=node('pre');provenance.append(node('summary','Retained native origin'),receipt);receipt.textContent=JSON.stringify({question:detail.measurement,stream:detail.stream,basis:detail.basis,readAt:detail.readAt},null,2);body.append(provenance,node('small','This native origin is retained in this tab. Sharing or saving the record question preserves its filters, not a durable citation to this native result.'));
   host.dataset.fromNs=detail.fromNS;host.dataset.toNs=detail.toNS;host.dataset.signal=detail.question.signal;host.hidden=false;
   stageQuestion(structuredClone(detail.question));history.pushState(null,'',href);return true;
  }
  global.addEventListener('goo:workspace',event=>{if(event.detail.route!=='explore')clear()});
  global.addEventListener('goo:explore-mode',event=>{if(event.detail.measurements){if(valid())suspendRecords();host.hidden=true}else host.hidden=!valid()||parent.returned});
  return {stage,restore,clear};
 }
 const api={visibleWithin,mount};if(typeof module!=='undefined')module.exports=api;else global.GooNativeRecordNavigation=api;
})(typeof window==='undefined'?globalThis:window);
