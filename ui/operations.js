/* Operation evidence selections stay on one validated response; no new queries. */
(function(global){
 const minute=60000000000n;
 const edges=[0n,1000000n,10000000n,50000000n,100000000n,500000000n,1000000000n,10000000000n];
 const labels=['<1 ms','1–10 ms','10–50 ms','50–100 ms','100–500 ms','0.5–1 s','1–10 s','≥10 s'];
 function select(spans,filter={}){
  return spans.filter(s=>{
   if(filter.operation!==undefined&&s.operation!==filter.operation||filter.role!==undefined&&s.role!==filter.role||filter.outcome!==undefined&&s.outcome!==filter.outcome||filter.traceID!==undefined&&s.traceID!==filter.traceID)return false;
   const start=BigInt(s.startNS),duration=BigInt(s.durationNS);
   return (filter.fromNS===undefined||start>=BigInt(filter.fromNS))&&(filter.toNS===undefined||start<BigInt(filter.toNS))&&(filter.minNS===undefined||duration>=BigInt(filter.minNS))&&(filter.maxNS===undefined||duration<BigInt(filter.maxNS));
  });
 }
 function distribution(spans,fromMs,toMs){
  const from=BigInt(fromMs)*1000000n,to=BigInt(toMs)*1000000n;
  const cells=new Map();
  for(const s of select(spans,{fromNS:from,toNS:to})){
   const start=BigInt(s.startNS),duration=BigInt(s.durationNS);let band=edges.length-1;
   while(band>0&&duration<edges[band])band--;
   const bucket=start/minute*minute,key=bucket+':'+band;
   if(!cells.has(key))cells.set(key,{bucketNS:String(bucket),band,count:0,fromNS:String(bucket<from?from:bucket),toNS:String(bucket+minute>to?to:bucket+minute),minNS:String(edges[band]),...(band+1<edges.length?{maxNS:String(edges[band+1])}:{})});
   cells.get(key).count++;
  }
  return [...cells.values()].sort((a,b)=>BigInt(a.bucketNS)<BigInt(b.bucketNS)?-1:BigInt(a.bucketNS)>BigInt(b.bucketNS)?1:a.band-b.band);
 }
 function waterfallRows(spans,service){
  return spans.map(s=>({TraceId:s.traceID,SpanId:s.spanID,ParentSpanId:s.parentSpanID,timestamp_ns:s.startNS,duration_ns:s.durationNS,duration_ms:Number(BigInt(s.durationNS))/1e6,SpanName:s.operation,ServiceName:service,StatusCode:s.outcome==='ok'?'Unset':'Error',SpanAttributes:s.attributes,operationOutcome:s.outcome}));
 }
 // A structured tuple prevents delimiter collisions and keeps tenant/namespace scope.
 function subjectKey(subject){
  if(!subject)return null;
  const parts=['environment','service','namespace','tenant','id'].map(k=>subject[k]);
  return parts.every(v=>typeof v==='string'&&v.length>0)?JSON.stringify(parts):null;
 }
 function subjectEvents(spans,subject){
  const key=subjectKey(subject);if(key===null)return [];
  const events=[];
  for(const span of spans)for(const [index,event] of (span.events||[]).entries())if(!event.diagnostic&&subjectKey(event.subject)===key)events.push({event,span,index});
  return events.sort((a,b)=>BigInt(a.event.timeNS)<BigInt(b.event.timeNS)?-1:BigInt(a.event.timeNS)>BigInt(b.event.timeNS)?1:a.span.traceID.localeCompare(b.span.traceID)||a.span.spanID.localeCompare(b.span.spanID)||a.index-b.index);
 }
 function subjectSpans(spans,subject){const members=new Set(subjectEvents(spans,subject).map(x=>x.span));return spans.filter(s=>members.has(s));}
 function node(tag,text){const n=document.createElement(tag);if(text!==undefined)n.textContent=text;return n;}
 function observationTable(id,labels,label){const viewport=node('goo-table'),table=node('table'),head=node('thead'),header=node('tr'),body=node('tbody');viewport.setAttribute('density','compact');viewport.setAttribute('aria-label',label);table.id=id;for(const text of labels){const th=node('th',text);th.scope='col';header.append(th)}head.append(header);table.append(head,body);viewport.append(table);return {viewport,body}}
 function mount(container,data,{prefix=''}={}){
  const snapshot=data.spans, basis=data.conformance.basis;
  const enabled=new Set(data.package?.definition.views||['operation-overview','duration-heatmap','trace-waterfall','subject-events']);
  const heat=node('div'),inspection=node('div'),waterfall=node('div');waterfall.id=prefix+'operationWaterfall';
 const subjects=node('section');subjects.id=prefix+'operationSubjects';
  const heading=node('h3','Operation duration heatmap'),selection=node('p');selection.id=prefix+'operationSelection';
  const clear=node('button','Show all operations');
  const selectedNote=node('p');selectedNote.id=prefix+'operationEvidenceStatus';
  heading.hidden=!enabled.has('duration-heatmap');heat.hidden=heading.hidden;subjects.hidden=!enabled.has('subject-events');
  container.append(clear,heading,selection,heat,selectedNote,inspection,waterfall,subjects);
  let current={},selectedFilter={},navigation={};
  function subjectButton(subject){
   const b=node('button',`Follow subject ${subject.tenant}/${subject.namespace}/${subject.id}`);
   b.onclick=()=>showSubject(subject);return b;
  }
  function openTrace(traceID){
   navigation={traceID};
   waterfall.replaceChildren(node('h3','Trace '+traceID));
   const expand=node('button','Expand trace across services and find linked logs');expand.onclick=()=>global.dispatchEvent(new CustomEvent('goo:expand-trace',{detail:{traceID,basis:data.conformance.basis,question:{environment:data.source.environment,fromMs:data.fromMs,toMs:data.toMs}}}));waterfall.append(expand);
   const trace=select(snapshot,{traceID});
   global.GooVisuals.waterfall(waterfall,waterfallRows(trace,data.source.service),false,undefined,'operationSpanDetails');
   subjects.replaceChildren(node('h3','Observed subject events in this trace'));
   let count=0;
   for(const s of trace)for(const event of s.events||[]){count++;const row=node('p');row.append(global.GooRecordPresentation.timestamp({timestamp_ns:event.timeNS}),` · ${event.name}${event.outcome?' · '+event.outcome:''} `);if(event.subject&&!event.diagnostic)row.append(subjectButton(event.subject));else row.append(node('span',event.diagnostic||'No scoped subject identity'));subjects.append(row);}
   if(!count)subjects.append(node('p','No observed subject events. Absence does not prove failure or nonexistence.'));
   waterfall.scrollIntoView({block:'nearest'});
  }
  function showSubject(subject){
   navigation={subject};
   const events=subjectEvents(snapshot,subject),spans=subjectSpans(snapshot,subject),traceIDs=new Set(spans.map(s=>s.traceID));
   subjects.replaceChildren(node('h3',`Subject ${subject.tenant}/${subject.namespace}/${subject.id}`));
   const note=node('p',`${events.length} observed event${events.length===1?'':'s'} · ${spans.length} event-bearing invocation${spans.length===1?'':'s'} · ${traceIDs.size} trace${traceIDs.size===1?'':'s'}. Scope: ${subject.environment}/${subject.service}/${subject.namespace}/${subject.tenant}. Same captured evidence and span-start window. Matching subject identity is a declared relation, not a parent edge or proof of retry, commit or durable completion.`);note.id=prefix+'operationSubjectStatus';subjects.append(note);
   const {viewport,body}=observationTable(prefix+'operationSubjectEvents',['Event time (UTC)','Event','Reported outcome','Invocation / trace context'],'Declared subject events in captured evidence');
   for(const {event,span} of events){const row=node('tr');const time=node('td');time.append(global.GooRecordPresentation.timestamp({timestamp_ns:event.timeNS}));row.append(time,node('td',event.name),node('td',event.outcome||'not reported'));const cell=node('td'),b=node('button','Open subject trace');b.setAttribute('aria-label','Open subject trace '+span.traceID);b.dataset.traceId=span.traceID;b.title=span.traceID;b.onclick=()=>openTrace(span.traceID);cell.append(node('span',span.operation+' '),b);row.append(cell);body.append(row);}
   subjects.append(viewport);subjects.scrollIntoView({block:'nearest'});
  }
  function showMatches(filter){
   selectedFilter={...filter};navigation={};
   const matches=select(snapshot,filter);inspection.replaceChildren();waterfall.replaceChildren();subjects.replaceChildren();
   selectedNote.textContent=`${matches.length} matching completed spans · same captured evidence · ${new Date(data.fromMs).toISOString()} to ${new Date(data.toMs).toISOString()} · ${data.source.service}/${data.source.environment}/${data.source.instance}. Trace completeness unknown; spans outside this operation scope/window are not fetched.`;
   if(!matches.length){inspection.append(node('p','No observed spans in this selection; absence is not proof of health.'));return;}
   const {viewport,body}=observationTable(prefix+'operationMatches',['Operation','Outcome','Duration','Trace'],'Operation invocations in captured selection');
   for(const s of matches){const row=node('tr');const duration=node('td',global.GooUI.durationMS(s.durationNS));duration.title=s.durationNS+' ns';row.append(node('td',s.operation),node('td',s.outcome),duration);const cell=node('td'),button=node('button','Open trace');button.setAttribute('aria-label','Open trace '+s.traceID);button.dataset.traceId=s.traceID;button.title=s.traceID;button.dataset.spanId=s.spanID;button.onclick=()=>openTrace(s.traceID);
   if(enabled.has('trace-waterfall'))cell.append(button);else cell.append(node('span',s.traceID));row.append(cell);body.append(row);}
   inspection.append(viewport);
  }
  function choose(filter){
   current={...filter};selection.textContent=filter.operation?`${filter.operation} · ${filter.role}${filter.outcome?' · '+filter.outcome:''}`:'All observed operations';
   heat.replaceChildren();if(!enabled.has('duration-heatmap')){showMatches(current);return;}const spans=select(snapshot,current),cells=distribution(spans,data.fromMs,data.toMs);
   heat.append(node('p',`${spans.length} observed durations · raw spans · UTC · select a populated cell. Empty cells mean no observations, not healthy behavior.`));
   const grid=node('table');grid.className='heatmap';grid.id=prefix+'operationHeatmap';const header=node('tr');header.append(node('th','Duration / UTC'));
   const times=[];for(let t=BigInt(data.fromMs)*1000000n/minute*minute;t<BigInt(data.toMs)*1000000n;t+=minute)times.push(String(t));
   times.forEach(t=>header.append(node('th',new Date(Number(BigInt(t)/1000000n)).toISOString().slice(11,19))));grid.append(header);
   const byKey=new Map(cells.map(c=>[c.bucketNS+':'+c.band,c])),max=Math.max(1,...cells.map(c=>c.count));
   for(let band=labels.length-1;band>=0;band--){const row=node('tr');row.append(node('th',labels[band]));for(const t of times){const cell=byKey.get(t+':'+band),td=node('td'),button=node('button',cell?String(cell.count):'·');button.className='heat-cell';button.disabled=!cell;button.dataset.count=cell?cell.count:0;button.setAttribute('aria-label',`${new Date(Number(BigInt(t)/1000000n)).toISOString()} ${labels[band]}: ${cell?cell.count:0} operation spans`);
    if(cell){for(const key of ['fromNS','toNS','minNS','maxNS'])if(cell[key]!==undefined)button.dataset[key]=cell[key];button.style.background=`rgba(57,201,173,${0.18+0.82*Math.log1p(cell.count)/Math.log1p(max)})`;button.onclick=()=>{const filter={...current,fromNS:cell.fromNS,toNS:cell.toNS,minNS:cell.minNS,...(cell.maxNS?{maxNS:cell.maxNS}:{})};selection.textContent=`${current.operation||'All operations'}${current.outcome?' · '+current.outcome:''} · ${labels[band]} · ${new Date(Number(BigInt(cell.fromNS)/1000000n)).toISOString()} (half-open cell)`;showMatches(filter);};}td.append(button);row.append(td);}grid.append(row);}
   const scroll=node('div');scroll.className='heatmap-scroll';scroll.append(grid);heat.append(scroll);showMatches(current);
  }
  clear.onclick=()=>choose({});container.dataset.basis=basis;choose({});return{choose,getState:()=>({filter:{...selectedFilter},...navigation}),restore:state=>{choose(state?.filter||{});if(state?.subject)showSubject(state.subject);if(state?.traceID)openTrace(state.traceID)}};
 }
 function overview(el,data,{prefix=''}={}){
 const textNode=node;
 const table=document.createElement('table'),header=document.createElement('tr');
 ['Operation','Role','Evidence','Calls','Errors','Panics','Duration samples'].forEach(c=>header.append(textNode('th',c)));table.append(header);
 const explorer=mount(el('operationExplorer'),data,{prefix});
 for(const row of data.conformance.operations){const tr=document.createElement('tr');const observed=row.status==='observed-compatible';
 for(const [column,value] of [["operation",row.name],["role",row.role],["status",observed?'Observed · compatible':'Unobserved · unknown'],["calls",observed?row.calls:'unknown'],["errors",observed?row.errors:'unknown'],["panics",observed?row.panics:'unknown'],["durations",observed?row.durationsNS.length:'unknown']]){
 const td=document.createElement('td');if(observed&&['operation','calls','errors','panics'].includes(column)&&value!==0){const button=textNode('button',value);button.dataset.operation=row.name;button.dataset.selection=column;button.setAttribute('aria-label',`Inspect ${column} for ${row.name}`);button.onclick=()=>explorer.choose({operation:row.name,role:row.role,...(column==='errors'?{outcome:'error'}:column==='panics'?{outcome:'panic'}:{})});td.append(button)}else td.textContent=value;tr.append(td)}table.append(tr)}
 el('operationView').append(table);
 el('operationExplain').textContent=JSON.stringify({package:data.package,manifestRevision:data.conformance.manifestRevision,basis:data.conformance.basis,source:data.source,fromMs:data.fromMs,toMs:data.toMs,query:data.query,parameters:data.parameters,duplicates:data.duplicateRecords},null,2);
 return explorer;
 }
 function renderRates(container,data,choose){
  container.hidden=false;container.replaceChildren(node('summary','Native counter rates'));
  container.append(node('p',data.warnings.join(' ')));
  for(const entry of data.rates){
   const box=node('section');box.append(node('h3',entry.contract.name+' · '+entry.contract.metric));
   if(!entry.result.series.length)box.append(node('p','No observed counter intervals — no data, not zero.'));
   for(const series of entry.result.series){
    const a=series.stream.attributes,known=series.intervals.filter(i=>i.state==='ok');
    box.append(node('p',`${a['goo.operation.name']} / ${a['goo.operation.role']} / ${a['goo.operation.outcome']} · window coverage: ${series.state} · ${known.length} supported intervals`));
    const ns='http://www.w3.org/2000/svg',svg=document.createElementNS(ns,'svg');svg.setAttribute('viewBox','0 0 900 110');svg.setAttribute('role','img');svg.setAttribute('aria-label','Observed '+entry.contract.name+' rate for '+a['goo.operation.name']+' '+a['goo.operation.outcome']);svg.style.width='100%';
    const max=Math.max(1,...known.map(i=>i.perSecond)),from=BigInt(entry.result.fromNS),extent=Number(BigInt(entry.result.toNS)-from);
    for(const interval of known){const line=document.createElementNS(ns,'line');line.setAttribute('x1',20+860*Number(BigInt(interval.fromNS)-from)/extent);line.setAttribute('x2',20+860*Number(BigInt(interval.toNS)-from)/extent);line.setAttribute('y1',90-70*interval.perSecond/max);line.setAttribute('y2',90-70*interval.perSecond/max);line.setAttribute('stroke','#67ddbb');line.setAttribute('stroke-width','3');svg.append(line);}
    box.append(svg);const table=node('table'),header=node('tr');['Observed interval · UTC','Count delta','Invocations/s','Evidence / investigation'].forEach(t=>header.append(node('th',t)));table.append(header);
    for(const interval of series.intervals){const tr=node('tr');tr.append(node('td',new Date(Number(BigInt(interval.fromNS)/1000000n)).toISOString()+' → '+new Date(Number(BigInt(interval.toNS)/1000000n)).toISOString()),node('td',interval.delta??'unknown'),node('td',interval.state==='ok'?interval.perSecond.toPrecision(4):'unknown'));const cell=node('td');
     if(interval.state==='ok'){const b=node('button','Inspect contextual spans');b.onclick=()=>choose({operation:a['goo.operation.name'],role:a['goo.operation.role'],outcome:a['goo.operation.outcome'],fromNS:interval.fromNS,toNS:interval.toNS});cell.append(b);}else cell.append(node('span',interval.reason));tr.append(cell);table.append(tr);}
    const scroll=node('div');scroll.style.maxHeight='220px';scroll.style.overflow='auto';scroll.append(table);box.append(scroll);
   }
   box.append(node('details'));const explain=box.lastChild;explain.append(node('summary','Counter contract and evidence basis'),node('pre',JSON.stringify({contract:entry.contract,basis:entry.result.basis,fromNS:entry.result.fromNS,toNS:entry.result.toNS},null,2)));container.append(box);
  }
 }
 const api={select,distribution,waterfallRows,subjectKey,subjectEvents,subjectSpans,mount,overview,renderRates};if(typeof module!=='undefined')module.exports=api;else global.GooOperations=api;
})(typeof window==='undefined'?globalThis:window);
