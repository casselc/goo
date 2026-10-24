/* Small multiples share a fixed captured scale. Local controls never read telemetry. */
(function(global){
 function mount(host,data,{onSelect,onRectangle}={}){
  const node=(tag,text)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=String(text);return n},api=global.GooDistributions,apiNumeric=global.GooNumericMeasures;
  host.replaceChildren();host.hidden=false;host.classList.add('grouped-time-chart');host.append(node('h3','Compare breakdowns over time'));
  const controls=node('div');controls.className='controls';const label=node('label','Chart measure'),select=node('select');select.setAttribute('aria-label','Grouped chart measure');select.dataset.seriesMeasure='true';
  for(const [key,title]of [['count','Captured record count'],...(data.numeric?[['numeric-mean','Numeric mean (absolute magnitude)'],['numeric-sum','Numeric sum (absolute magnitude)'],['numeric-min','Numeric minimum (absolute magnitude)'],['numeric-max','Numeric maximum (absolute magnitude)']]:[]),...(data.duration?[['mean','Mean span duration'],['p95','p95 span duration'],['max','Maximum span duration']]:[]),...(data.spanStatus?[['error-rate','Recorded error percentage']]:[])]){const o=node('option',title);o.value=key;select.append(o)}label.append(select);
  const searchLabel=node('label','Find a series'),search=node('input');search.type='search';search.setAttribute('aria-label','Find a captured series');searchLabel.append(search);controls.append(label,searchLabel);host.append(controls);
  const scale=node('p'),list=node('div'),status=node('p','Select an interval to inspect exact bounds, count and available duration measures.');status.setAttribute('role','status');status.className='series-interval-detail';list.className='series-grid';host.append(scale,list,status);
  const groups=data.groupedTimeSeries||[];
  if(data.question.signal==='spans'&&data.question.groupBy?.length===1&&data.question.groupBy[0].kind==='column'&&data.question.groupBy[0].key==='duration_bucket'&&data.timeSeries?.length&&data.timeSeries.length<=32&&data.state!=='changed_between_reads'){
   controls.hidden=scale.hidden=list.hidden=status.hidden=true;host.querySelector('h3').textContent='Span duration over time';const order=global.GooDiscovery.durationOrder,grid=node('goo-bucket-heatmap');
   const columns=data.timeSeries.map((p,i)=>({key:String(i),label:new Date(p.fromMs).toISOString(),cells:order.map(key=>{const g=groups.find(g=>g.values[0]===key),point=g?.points.find(x=>x.fromMs===p.fromMs&&x.toMs===p.toMs);return point?{state:'observed',value:point.count}:{state:'unknown'}})}));
   grid.data={title:'Raw span duration by UTC interval · half-open buckets',ceiling:columns.reduce((max,c)=>c.cells.reduce((n,p)=>p.value&&BigInt(p.value)>BigInt(n)?p.value:n,max),'0'),rows:order.map(key=>({key,label:key})),columns};
   if(onRectangle){grid.setAttribute('range-select','');grid.addEventListener('goo-heatmap-range',e=>onRectangle(global.GooPopulationSelection.rectangle(data,e.detail)))}
   host.append(grid);
  }
  function draw(){
   const metric=select.value;let max={n:0n,d:1n};for(const g of groups)for(const p of g.points){const v=api.groupMetric(p,metric);if(v&&v.n*max.d>max.n*v.d)max=v}
   const display=p=>metric.startsWith('numeric-')?apiNumeric.format(p.numeric,metric.slice(8))+' · '+apiNumeric.coverage(p.numeric):metric==='count'?p.count:metric==='error-rate'?(p.count==='0'?'No captured data':api.percent(p.spanStatus.error,p.count)+'% recorded Error'):p.duration?.state==='observed'?api.durationMS(p.duration[metric==='mean'?'sumNs':metric==='p95'?'p95Ns':'maxNs'],metric==='mean'?p.count:'1'):'No captured data';
   scale.textContent=groups.length+' captured series · common scale across all captured series · '+(metric.startsWith('numeric-')?'absolute magnitude; labels retain signs; only valid values enter arithmetic':metric==='count'?'record counts':metric==='error-rate'?'recorded Error / all captured spans; Unset and unknown remain in the denominator':'span duration; exact measures rounded for display');list.replaceChildren();
   for(const g of groups){
    const heading=data.question.groupBy.map((f,i)=>f.kind+':'+f.key+' = '+api.valueLabel(g.values[i],f)).join(' · '),card=node('article');card.className='series-card';card.dataset.series=JSON.stringify(g.values);card.dataset.search=heading.toLocaleLowerCase();card.append(node('h4',heading));const bars=node('div');bars.className='time-bars';
    for(const p of g.points){const b=node('button'),value=api.groupMetric(p,metric);b.className='time-bar'+(p.state==='no_data'?' no-data':'');b.dataset.state=p.state;b.style.setProperty('--bar-height',api.metricWidth(value,max)+'%');const text=heading+' · '+new Date(p.fromMs).toISOString()+' → '+new Date(p.toMs).toISOString()+' (exclusive) · '+display(p)+' · '+p.count+' captured records';b.setAttribute('aria-label',text);b.title=text;b.onclick=()=>{status.textContent=text+(p.state==='no_data'?' · no captured data; health unknown':'');onSelect?.(p,data.question.groupBy.map((field,i)=>({field,value:g.values[i]})))};bars.append(b)}
    const axis=node('div');axis.className='time-axis';axis.append(node('span',new Date(data.question.fromMs).toISOString()),node('span',new Date(data.question.toMs).toISOString()));card.append(bars,axis);list.append(card)
   }
   filter();status.textContent='Select an interval to inspect exact bounds, count and available duration measures.';
  }
  function filter(){for(const card of list.children)card.hidden=!card.dataset.search.includes(search.value.toLocaleLowerCase())}
  select.onchange=draw;search.oninput=filter;draw();
  if(!groups.length)host.append(node('p','No observed breakdown series in this read. No resource health or missing-instrument conclusion is implied.'));
  host.append(node('p','At most 8 observed series and 1,000 observed cells; over-budget queries fail instead of hiding groups. Empty intervals mean no captured records. Missing keys and empty strings are distinct. Percentiles use each cell’s raw span durations; scalar percentiles are never averaged. Recorded span status is not an application outcome policy. The split chart, total chart and breakdowns are independent reads.'+(data.state==='changed_between_reads'?' Source reads disagree; do not treat this as one population snapshot.':'')));
 }
 global.GooGroupedTimeSeries={mount};
})(window);
