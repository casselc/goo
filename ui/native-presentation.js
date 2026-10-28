/* Supplied native values and coverage. No source reads, query ownership or evidence clocks. */
(function(global){
 const node=(...args)=>(typeof module!=='undefined'?require('./components.mjs'):global.GooUI).node(...args);
 const reasonDescriptions={
  'missing-coverage':['Exports do not cover this time range.','Excluded time is not a measured zero or evidence of ingestion completeness.'],
  'unknown-interval':['This interval cannot be evaluated.','Inspect retained exports before selecting a supported scope.'],
  'empty-distribution':['The complete bucket records zero observations.','No percentile is defined for an empty population.'],
  'missing-exact-stream':['This exact stream is missing from one window.','Inspect the full identity; matching an instrument name or replacement instance is insufficient.'],
  'unknown-observation':['No evaluated observation is available.','Inspect the fixed window and exported values; absence is not a zero.'],
  'unsupported-temporality':['The exported temporality is unsupported.','Inspect delta or cumulative semantics before comparing windows.'],
  'bucket-layout-differs':['The windows use different histogram buckets.','Compare compatible native bucket layouts.'],
  'boundary-crossing':['An observation crosses a window boundary.','Choose bounds aligned with complete exports; Goo does not split an increment.'],
  'incomplete-window':['The whole window cannot be evaluated.','Inspect interval boundaries and export coverage; missing time is not zero.'],
  'incomplete-or-overlapping-coverage':['Export coverage is incomplete or overlapping.','Check missing boundaries, gaps and repeated exports before comparing rates.'],
  'incomplete-or-empty-window':['A complete window is required.','Use a window with observed start and end boundaries.'],
  'counter-decrease':['The counter decreased.','A reset cannot be treated as a negative increment. Inspect the counter epoch.'],
  'counter-epoch-change':['The counter epoch changed.','Compare within one epoch rather than subtracting across a restart.'],
  'invalid-counter-epoch':['The counter epoch is invalid.','Inspect exported start times and temporality.'],
  'histogram-epoch-change':['The histogram epoch changed.','Do not subtract bucket counts across an epoch change.'],
  'histogram-decrease':['Histogram counts decreased.','Inspect resets and bucket observations before comparing distributions.'],
  'invalid-histogram-epoch':['The histogram epoch is invalid.','Inspect exported start times and temporality.'],
  'bucket-layout-change':['Histogram bucket boundaries changed.','Compare compatible native bucket layouts.'],
  'excessive-gap':['Observations exceed the declared gap limit.','Inspect export cadence or explicitly revise the package policy.'],
  'no-recorded-value':['The exporter marked a value as absent.','Absence is unknown, not a measured zero.'],
  'no recorded value':['The exporter marked a value as absent.','Absence is unknown, not a measured zero.'],
  'older than declared freshness':['An observation exceeds the freshness limit.','Inspect export cadence. Stored evidence keeps its fixed evaluation time.'],
  'conflicting observations at one timestamp':['Observations conflict at one timestamp.','Inspect duplicate exports; Goo does not choose a value arbitrarily.'],
  'stale-export-endpoint':['The latest export is too old for a live rate.','Check the selected stream and export cadence. Frozen windows remain historical evidence.'],
  'no-observed-endpoint':['No export endpoint was observed in the bounded search.','Check the exact identity and source window; a replacement instance is not substituted.'],
  'missing-baseline-stream':['This exact stream is missing from the baseline.','Capture a compatible baseline; matching only the instrument name is insufficient.'],
  'missing-selection-stream':['This exact stream is missing from the current window.','Inspect the selected identity; another instance is not a replacement.'],
  'temporality-differs':['Metric temporality changed.','Do not compare cumulative and delta windows as if they were equivalent.']
 };
 function reasonInfo(code){const [title,guidance]=(Object.hasOwn(reasonDescriptions,code)?reasonDescriptions[code]:null)||['The source reported an additional diagnostic.','Inspect its exact code and retained observations before interpreting this window.'];return {code,title,guidance}}
 function streamDiagnostics(series,kind){
  const observations=['gauge','stock'].includes(kind)?series.samples||[]:series.intervals||[],good=['gauge','stock'].includes(kind)?'observed':'ok',recorded=observations.filter(x=>x.state===good).length,review=observations.length-recorded;
  const reasons=[...new Set(observations.map(x=>x.reason).filter(Boolean))];
  const attention=series.state!==good||review>0;
  if(attention&&!reasons.length)reasons.push(['gauge','stock'].includes(kind)?'unknown-observation':'incomplete-window');
  const label=['gauge','stock'].includes(kind)?(series.state==='observed'?'Latest observation recorded':series.state==='stale'?'Latest observation stale':'Observation unavailable'):(series.state==='ok'?'Window covered':series.state==='no_data'?'No evaluated intervals':'Window needs review');
  return {state:series.state,attention,label,recorded,review,reasons};
 }
 function latest(s,kind){
   const gauge=['gauge','stock'].includes(kind),points=(gauge?s.samples:s.intervals)||[],last=points.reduce((a,p)=>!a||BigInt(p.atNS||p.toNS)>BigInt(a.atNS||a.toNS)?p:a,null);
   let value='Unavailable';const validSeries=kind==='stock'?['observed','unknown','stale'].includes(s.state):gauge?s.state==='observed':['ok','unknown'].includes(s.state);
   if(validSeries&&(kind==='stock'?['observed','stale'].includes(last?.state):last?.state===(gauge?'observed':'ok'))){
    if(gauge&&typeof last.value==='number'&&Number.isFinite(last.value))value=String(last.value)+' '+s.stream.unit;
    else if(kind==='counter'&&typeof last.perSecond==='number'&&Number.isFinite(last.perSecond))value=String(last.perSecond)+' '+s.stream.unit+'/s';
    else if(kind==='histogram'&&last.distribution?.unit===s.stream.unit){const d=last.distribution,formatter=typeof module!=='undefined'?require('./histograms'):global.GooHistograms;value=d.count+' observations · '+(['bucket-bound','unbounded-upper','unbounded-lower','unbounded','no_data'].includes(d.p95.state)?formatter.p95Label(d.p95,d.unit):'P95 unavailable')}
   }
   return {stream:structuredClone(s.stream),kind,seriesState:s.state,state:last?.state||s.state,value,extent:last?(gauge?'Endpoint '+last.atNS+' ns':'Latest interval ['+last.fromNS+', '+last.toNS+') ns'):'No retained endpoint or interval',reason:last?.reason||'',fromNS:last?.fromNS,toNS:last?.toNS,atNS:last?.atNS,...(kind==='stock'?{startNS:last?.startNS}:{})};
 }
 function nativeSummary(data){const family=data.gauges||data.stocks||data.counters||data.histograms;if(!family)return [];const kind=data.gauges?'gauge':data.stocks?'stock':data.counters?'counter':'histogram';return family.series.slice(0,4).map(s=>latest(s,kind))}
 function rateText(value){if(typeof value!=='string'||! /^-?\d+(\/[1-9]\d*)?$/.test(value)||value.length>256)return 'Unavailable';const [a,b='1']=value.split('/'),n=BigInt(a),d=BigInt(b),v=((n<0n?-n:n)*1000n+d/2n)/d;return (n<0n&&v!==0n?'-':'')+String(v/1000n)+'.'+String(v%1000n).padStart(3,'0')}
 function windowValue(stream,summary){
  if(!windowCoverage(summary).validated)return 'Unavailable';
  if(summary.rate!==undefined)return summary.delta+' increments · '+rateText(summary.rate)+' '+stream.unit+'/s';
  const d=summary.distribution;if(!d||d.unit!==stream.unit)return 'Unavailable';
  const formatter=typeof module!=='undefined'?require('./histograms'):global.GooHistograms;
  return d.count+' histogram observations · '+(['bucket-bound','unbounded-upper','unbounded-lower','unbounded','no_data'].includes(d.p95?.state)?formatter.p95Label(d.p95,d.unit):'P95 unavailable');
 }
 function renderWindowMember(host,member){const section=node('section');section.className='native-window-member';section.append(node('p',(member.stream.resource?.['service.instance.id']||'Unknown instance')+' · '+member.stream.name+' ['+member.stream.unit+']'),node('strong',windowValue(member.stream,member.summary)));renderWindowCoverage(section,'Export coverage',member.summary);section.append(global.GooAttributes.streamInspection(member.stream));host.append(section);return section}
 function renderReasons(host,codes){
  const unique=[...new Set(codes.filter(Boolean))];if(!unique.length)return;
  const list=node('ul');list.className='native-reason-list';for(const code of unique){const info=reasonInfo(code),item=node('li');item.append(node('strong',info.title),node('span',' '+info.guidance));list.append(item)}
  const exact=node('details');exact.append(node('summary','Diagnostic codes'),node('code',unique.join(' · ')));host.append(list,exact);
 }
 function extent(host,value){const p=node('p');if(value.atNS!==undefined)p.append('Endpoint ',global.GooEvidenceSummary.timestamp(value.atNS));else if(value.fromNS!==undefined&&value.toNS!==undefined)p.append('Latest interval [',global.GooEvidenceSummary.timestamp(value.fromNS),' → ',global.GooEvidenceSummary.timestamp(value.toNS),')');else p.textContent='No retained endpoint or interval';if(value.kind==='stock')p.append(' · accumulation epoch ',value.startNS?global.GooEvidenceSummary.timestamp(value.startNS):node('span','Unavailable'));host.append(p)}
 function renderStreamDiagnostics(host,s,kind,{exportCoverage=false,compact=false,showLatest=true}={}){const m=streamDiagnostics(s,kind),section=node(compact?'details':'section');section.className='native-diagnostics'+(compact?' native-coverage-details':'');section.dataset.state=m.state;section.append(node(compact?'summary':'h4',(exportCoverage?'Export intervals · ':'')+m.label));
  if(showLatest){const value=latest(s,kind);const row=node('p');row.append(node('strong',value.value));section.append(row);extent(section,value)}
  section.append(global.GooUI.stats([{label:['gauge','stock'].includes(kind)?'Fresh observations':'Recorded intervals',value:String(m.recorded)},{label:['gauge','stock'].includes(kind)?'Observations needing review':'Intervals needing review',value:String(m.review)}]));renderReasons(section,m.reasons);section.append(node('small',['gauge','stock'].includes(kind)?'Independent samples; freshness is evaluated at the fixed window end.':'Recorded intervals do not by themselves establish whole-window coverage.'));host.append(section);return section;
 }
 function windowCoverage(side){return side?{state:side.state,fromNS:side.fromNS,toNS:side.toNS,temporality:side.temporality,reasons:[...new Set(side.reasons||[])],validated:side.state==='observed'&&!(side.reasons||[]).length}:{state:'missing',reasons:[],validated:false}}
 function renderWindowCoverage(host,label,side){const m=windowCoverage(side),section=node('div');section.className='native-window-coverage';section.dataset.state=m.state;const p=node('p');if(side)p.append(label+': [',global.GooEvidenceSummary.timestamp(m.fromNS),' → ',global.GooEvidenceSummary.timestamp(m.toNS),') · '+m.temporality+' · '+m.state+' · '+(m.validated?'Validated export coverage':'Coverage not validated'));else p.textContent=label+': missing exact stream';section.append(p);renderReasons(section,m.reasons);host.append(section);return section}
 function instrumentLabel(name,unit){return (typeof name==='string'&&name?name:'Instrument not reported')+' ['+(typeof unit==='string'&&unit?unit:'unit not reported')+']'}
 const api={instrumentLabel,windowValue,renderWindowMember,reasonInfo,streamDiagnostics,latest,nativeSummary,rateText,windowCoverage,renderReasons,renderStreamDiagnostics,renderWindowCoverage};if(typeof module!=='undefined')module.exports=api;else global.GooNativePresentation=api;
})(typeof window==='undefined'?globalThis:window);
