/* Application-owned refresh, snapshot custody and cancellation. Display elements stay passive. */
(function(global){
 function displayState(phase,comparison){if(phase==='current'||phase==='frozen')return comparison?.state==='comparable'?'observed':comparison?.state==='incompatible'?'incompatible':'unavailable';return phase==='paused'?(comparison?'stale':'unknown'):['loading','stale','unavailable'].includes(phase)?phase:'unknown'}
 function measureView(comparison){
  if(comparison?.state!=='comparable')return {};
  const m=comparison.measures||{},a=m.baselineDuration,b=m.selectionDuration,x=m.baselineStatus,y=m.selectionStatus,out={};
  if(a?.state==='observed'&&b?.state==='observed'&&a.count===comparison.baseline.count&&b.count===comparison.selection.count){const ns=[a,b].flatMap(d=>['p50Ns','p95Ns','p99Ns'].map(k=>d[k]));if(ns.every(v=>typeof v==='string'&&/^\d+$/.test(v)))out.latency={baseline:a,current:b,maximumNs:ns.reduce((m,v)=>BigInt(v)>BigInt(m)?v:m,'0')}}
  if(x&&y&&x.count===comparison.baseline.count&&y.count===comparison.selection.count&&x.count!=='0'&&y.count!=='0'&&m.errorDifferencePP!==undefined)out.status={baseline:x,current:y,delta:m.errorDifferencePP};
  return out;
 }
 function rateText(value){if(typeof value!=='string'||! /^-?\d+(\/[1-9]\d*)?$/.test(value)||value.length>256)return 'Unavailable';const [a,b='1']=value.split('/'),n=BigInt(a),d=BigInt(b),v=((n<0n?-n:n)*1000n+d/2n)/d;return (n<0n&&v!==0n?'-':'')+String(v/1000n)+'.'+String(v%1000n).padStart(3,'0')}
 function mount(host,{baseline,current=()=>true,onLink,kind='record'}){
  const n=(tag,text)=>{const e=document.createElement(tag);if(text!==undefined)e.textContent=text;return e},section=n('section'),heading=n('h4',kind==='counter'?'Live counter rate':'Live record measures'),controls=n('div'),start=n('button','Start live comparison'),pause=n('button','Pause live comparison'),freeze=n('button','Freeze displayed result'),rescale=n('button','Rescale comparison'),status=n('p','Uses the baseline question and duration. Full-population reads; independent of the log tail.'),scope=n('details'),output=n('goo-chart-frame'),receipt=n('p'),link=n('button','Add frozen result to investigation');
  if(kind==='counter')status.textContent='Requires a baseline with one complete exact OTel stream. Refresh follows the latest observed export endpoint; gaps, resets and stale endpoints withhold rates.';const scopeText=n('p');scope.append(n('summary','Baseline window and source read'),scopeText);scope.hidden=true;output.hidden=true;section.className='live-comparison';section.dataset.baseline=baseline;controls.className='controls';status.setAttribute('role','status');controls.append(start,pause,freeze,rescale);section.append(heading,controls,status,scope,output,receipt,link);link.hidden=true;host.append(section);pause.disabled=freeze.disabled=rescale.disabled=true;
  let generation=0,active=false,timer,controller,last=null,ceiling=null,latencyCeiling=null,rateCeiling=null,busy=false;
  const valid=()=>section.isConnected&&current(),emit=phase=>{section.dataset.phase=phase;output.setAttribute('state',displayState(phase,last?.comparison));global.dispatchEvent(new CustomEvent('goo:live-comparison-state',{detail:{phase,baseline,generation,token:last?.token||null,basis:last?.basis||null,readAt:last?.comparison.selection.readAt||null}}))};
  function stop(){active=false;generation++;clearTimeout(timer);controller?.abort();pause.disabled=true;start.disabled=false;freeze.disabled=!last||busy||Date.now()>=Date.parse(last.expiresAt);emit('paused')}
  function renderCounter(r){
   const a=r.baseline,b=r.selection,q=b.receipt.question.measurement,w=r.nativeWindows?.[0];
   output.setAttribute('heading','Exact-stream counter rate');output.setAttribute('unit',q.definition.source.unit+'/s');
   output.dataset.basis=last.basis||'';output.dataset.readAt=b.readAt;output.dataset.fromNs=q.fromNS||String(BigInt(q.fromMs)*1000000n);output.dataset.toNs=q.toNS||String(BigInt(q.toMs)*1000000n);
   output.setAttribute('state',displayState('current',r));scope.hidden=false;
   const width=BigInt(output.dataset.toNs)-BigInt(output.dataset.fromNs);
   output.setAttribute('window-label',(last.native?.endpointNS?'Export ended ':'No observed endpoint · searched to ')+new Date(Number(BigInt(output.dataset.toNs)/1000000n)).toISOString()+' · '+global.GooUI.durationMS(String(width))+' window');
   scopeText.textContent='Current exact native scope ['+output.dataset.fromNs+', '+output.dataset.toNs+') ns · source read '+b.readAt+' · basis '+last.basis+' · endpoint policy '+last.native?.policy+' · export age '+(last.native?.endpointNS?global.GooUI.durationMS(last.native.lagNS):'unknown');
   const coverage=n('details');coverage.append(n('summary','Stream identity and window coverage'));output.append(coverage);
   for(const item of r.nativeWindows||[]){if(global.GooAttributes)coverage.append(global.GooAttributes.disclosure({ResourceAttributes:item.stream.resource,Attributes:item.stream.attributes,scope:item.stream.scope,version:item.stream.version,schema:item.stream.schemaURL,resourceSchema:item.stream.resourceSchemaURL}));for(const [label,s]of[['Baseline',item.baseline],['Current',item.selection]])coverage.append(n('p',label+': '+(s?'['+s.fromNS+', '+s.toNS+') ns · '+s.temporality+' · '+s.state+' · '+s.reasons.join(' · '):'missing exact stream')));for(const reason of item.reasons)coverage.append(n('p',reason))}
   const reasons=[...(last.native?.reasons||[]),...(r.nativeWindows||[]).flatMap(x=>[...x.reasons,...(x.selection?.reasons||[])])];
   if(r.state!=='comparable'||r.nativeWindows?.length!==1||w?.rateDifference===undefined){output.append(n('p','Counter rate unavailable · '+(reasons.join(' · ')||'missing exact stream or complete boundary observations')+' · no interpolation or older-window fallback.'));rescale.disabled=true;return}
   const visual=global.GooComparisonVisual,values=[w.baseline.rate,w.selection.rate],max=visual.maximum(values);if(rateCeiling===null)rateCeiling=max;
   rescale.disabled=visual.maximum([max,rateCeiling])===rateCeiling;
   if(!rescale.disabled){output.append(n('p','Counter rate exceeds the fixed axis (0–'+rateText(rateCeiling)+'). Rescale explicitly to view both bars.'));return}
   const figure=visual.pair(output,{title:'Validated counter rate',values,display:values.map(v=>rateText(v)+' '+w.stream.unit+'/s'),ceiling:rateCeiling,unit:w.stream.unit+'/s',delta:rateText(w.rateDifference),labels:['Frozen baseline','Latest export window']});figure.dataset.measure='counter-rate';output.dataset.scaleMax=rateCeiling;
   const exact=n('details');exact.append(n('summary','Exact increments and rates'));for(const [label,s]of[['Baseline',w.baseline],['Current',w.selection]])exact.append(n('p',label+' · '+s.delta+' validated increments · exact rate '+s.rate+' '+w.stream.unit+'/s'));exact.append(n('p','Exact current − baseline: '+w.rateDifference+' '+w.stream.unit+'/s'));output.append(exact,n('small','One exact OTel stream. No fleet aggregation or inference of missing increments. Retained observations do not establish ingestion completeness.'));
  }
  function render(){
   output.hidden=false;for(const child of Array.from(output.children))if(!child.hasAttribute('data-chart-frame-chrome'))child.remove();
   output.removeAttribute('data-scale-max');output.setAttribute('heading','Baseline and rolling current');output.setAttribute('unit','record population');
   if(!last)return;
   const r=last.comparison;if(r.selection.receipt.question.measurement){renderCounter(r);return}const a=r.baseline,b=r.selection,q=b.receipt.question.explore,view=measureView(r);
   output.dataset.basis=last.basis||'';output.dataset.readAt=b.readAt;output.dataset.fromMs=q.fromMs;output.dataset.toMs=q.toMs;
   output.setAttribute('state',displayState('current',r));scope.hidden=false;
   scopeText.textContent='Baseline ['+new Date(a.receipt.question.explore.fromMs).toISOString()+', '+new Date(a.receipt.question.explore.toMs).toISOString()+') · source read '+b.readAt+' · basis '+(last.basis||'unavailable')+' · coverage unknown';
   output.setAttribute('window-label','['+new Date(q.fromMs).toISOString()+', '+new Date(q.toMs).toISOString()+')');
   if(r.state!=='comparable'){output.append(n('p',r.state+' · '+r.differences.join(' · ')+' · numeric comparison withheld'));if(b.state==='no_data')output.append(n('p','No matching records in the rolling current window. Latency and recorded error share are unavailable.'));else if(a.state==='no_data')output.append(n('p','No matching records in the baseline window. Numeric comparison is unavailable.'));rescale.disabled=true;return}
   const max=global.GooComparisonVisual.maximum([a.count,b.count]);if(ceiling===null)ceiling=max;
   const countOverflow=BigInt(max)>BigInt(ceiling);
   if(countOverflow)output.append(n('p','Current count exceeds the fixed axis (0–'+ceiling+'). Rescale explicitly to view both bars.'));
   else{const counts=global.GooComparisonVisual.pair(output,{title:'Matching source records',values:[a.count,b.count],labels:['Frozen baseline','Rolling current'],unit:'records',delta:r.countDifference,ceiling});counts.dataset.measure='record-count';output.dataset.scaleMax=ceiling}
   let latencyOverflow=false;
   if(view.latency){
    if(latencyCeiling===null)latencyCeiling=view.latency.maximumNs;
    latencyOverflow=BigInt(view.latency.maximumNs)>BigInt(latencyCeiling);
    if(latencyOverflow){const message=n('p','Span latency exceeds the fixed axis (0–'+global.GooUI.durationMS(latencyCeiling)+'). Rescale explicitly to view percentiles.');message.dataset.latencyOverflow='true';output.append(message)}
    else{const pair=n('div');pair.className='live-latency-pair';pair.dataset.latencyCeilingNs=latencyCeiling;
     for(const [title,summary,side]of[['Frozen baseline',view.latency.baseline,'baseline'],['Rolling current',view.latency.current,'current']]){const element=n('goo-latency');element.dataset.side=side;element.ceilingNs=latencyCeiling;element.heading=title;element.headingLevel=5;element.summary=summary;pair.append(element)}
     output.append(pair,n('small','Span latency · mean and exact p50 / p95 / p99, shared fixed scale. Source records, not end-to-end request latency.'))}
   }else if(q.summary==='span-duration')output.append(n('p','Span latency unavailable for this snapshot.'));
   if(view.status){const {baseline:x,current:y}=view.status,pct=global.GooDistributions.percent;
    const values=[pct(x.error,x.count),pct(y.error,y.count)];
    const figure=global.GooComparisonVisual.pair(output,{title:'Recorded error share',values,display:[values[0]+'% · '+x.error+'/'+x.count+' spans',values[1]+'% · '+y.error+'/'+y.count+' spans'],unit:'%',ceiling:'100',delta:view.status.delta+' pp',labels:['Frozen baseline','Rolling current']});figure.dataset.measure='recorded-error-share';
    const details=n('details');details.append(n('summary','Recorded status coverage'));
    for(const [label,s]of[['Baseline',x],['Current',y]])details.append(n('p',label+' · Error '+s.error+' · Ok '+s.ok+' · Unset '+s.unset+' · Unknown '+s.unknown+' / '+s.count+' spans'));
    details.append(n('p','All matching spans form the denominator. Unset and unknown are not success. Sampling and telemetry completeness remain unknown.'));output.append(details);
   }else if(q.spanStatus)output.append(n('p','Recorded error share unavailable for this snapshot.'));
   if(q.signal==='spans'&&q.summary!=='span-duration'&&!q.spanStatus)output.append(n('small','This saved question requests counts only. Capture a question with span duration and recorded status to compare those measures.'));
   rescale.disabled=!(countOverflow||latencyOverflow);
  }
  async function poll(){if(!active||!valid())return stop();const g=generation;controller=new AbortController();freeze.disabled=true;emit('loading');status.textContent='Refreshing full-population measures…';
   try{const response=await global.GooQueries.post('/api/captures/live',{baseline,...(last?{replace:last.token}:{})},controller.signal),data=await response.json();if(!response.ok)throw Error(data.error||'Current read unavailable');if(g!==generation||!valid())return;last=data;render();freeze.disabled=false;status.textContent='Current as of the displayed source read · refresh every 10 seconds. Freeze expires in two minutes.';emit('current')}
   catch(e){if(g!==generation||!valid())return;status.textContent=(last?'Stale':'Unavailable')+' · '+e.message+(last?' · previous result retained with its original window.':' · no completed current result.');freeze.disabled=!last||Date.now()>=Date.parse(last.expiresAt);emit(last?'stale':'unavailable')}
   finally{if(active&&g===generation){if(valid())timer=setTimeout(poll,10000);else stop()}}
  }
  start.onclick=()=>{if(busy||!valid())return;active=true;generation++;start.disabled=true;pause.disabled=false;poll()};pause.onclick=stop;
  rescale.onclick=()=>{if(!last)return;if(last.comparison.selection.receipt.question.measurement){const w=last.comparison.nativeWindows?.[0];if(w?.baseline?.rate===undefined||w?.selection?.rate===undefined)return;rateCeiling=global.GooComparisonVisual.maximum([w.baseline.rate,w.selection.rate]);render();return}ceiling=global.GooComparisonVisual.maximum([last.comparison.baseline.count,last.comparison.selection.count]);latencyCeiling=measureView(last.comparison).latency?.maximumNs??null;render()};
  freeze.onclick=async()=>{if(!last||busy||!valid())return;stop();const selected=last;busy=true;start.disabled=freeze.disabled=true;status.textContent='Freezing the displayed snapshot; no telemetry reread…';try{const response=await fetch('/api/captures/live',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({freeze:selected.token})}),data=await response.json();if(!response.ok)throw Error(data.error);if(!valid())return;receipt.textContent='Frozen displayed evidence: '+data.receipt.id;receipt.className='reference-id';section.dataset.frozen=data.receipt.id;const open=n('a',' Open frozen capture');open.href='#capture='+data.receipt.id;receipt.append(open);link.hidden=!onLink;link.onclick=()=>{try{onLink(data.receipt.id)}catch(e){status.textContent=e.message}};status.textContent='Exact displayed result saved in the capture library. Add it to an investigation explicitly.';emit('frozen')}catch(e){if(valid()){status.textContent=e.message;emit('stale')}}finally{busy=false;start.disabled=false;freeze.disabled=false}};
  const visibility=()=>{if(document.hidden)stop()};document.addEventListener('visibilitychange',visibility);
  // Scope and document changes stop the application subscription immediately.
  const leave=()=>stop();global.addEventListener('goo:workspace',leave);global.addEventListener('hashchange',leave);
  let disposed=false;const observer=new MutationObserver(()=>{if(!section.isConnected)dispose()});observer.observe(document.body,{childList:true,subtree:true});function dispose(){if(disposed)return;disposed=true;stop();observer.disconnect();document.removeEventListener('visibilitychange',visibility);global.removeEventListener('goo:workspace',leave);global.removeEventListener('hashchange',leave);section.remove()}return {dispose};
 }
 const api={displayState,measureView,rateText,mount};if(typeof module!=='undefined')module.exports=api;else global.GooLiveComparison=api;
})(typeof window==='undefined'?globalThis:window);
