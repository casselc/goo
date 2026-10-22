/* Ratios consume aligned Go results; span navigation remains contextual. */
(function(global){
 function node(tag,text){const n=document.createElement(tag);if(text!==undefined)n.textContent=text;return n;}
 function label(i){return i.state==='ok'?`${i.percent.toFixed(2)}% (${i.errors}/${i.calls})`:i.state==='no_data'?'No data · zero traffic':`Unknown · ${i.reason}`;}
 function render(container,data,choose,{prefix=''}={}){
  let selected;const choices=new Map();
  container.hidden=false;container.replaceChildren(node('summary','Operation error ratios'),node('p','Calls, errors and error rate share one counter evidence basis. Every declared outcome must have a valid aligned interval. Missing telemetry stays unknown; zero traffic is no data. Span links are contextual, not exact counter contributors.'));
  for(const entry of data.ratios){
   if(!entry.result.series.length)container.append(node('p','No observed outcome streams · no data.'));
   for(const series of entry.result.series){
    const attrs=series.stream.attributes,box=node('section');box.append(node('h3',`${attrs['goo.operation.name']} / ${attrs['goo.operation.role']}`),node('p','Window coverage: '+series.state));
    const svg=document.createElementNS('http://www.w3.org/2000/svg','svg');svg.setAttribute('viewBox','0 0 900 140');svg.setAttribute('role','img');svg.setAttribute('aria-label','Observed error percentage for '+attrs['goo.operation.name']);svg.style.width='100%';
    function mark(tag,attributes,text){const n=document.createElementNS(svg.namespaceURI,tag);for(const [k,v]of Object.entries(attributes))n.setAttribute(k,String(v));if(text!==undefined)n.textContent=text;svg.append(n);return n;}
    const from=BigInt(entry.result.fromNS),extent=Number(BigInt(entry.result.toNS)-from),x=t=>40+840*Number(BigInt(t)-from)/extent;
    mark('text',{x:0,y:20,fill:'#a9bfce','font-size':12},'100%');mark('text',{x:10,y:120,fill:'#a9bfce','font-size':12},'0%');
    mark('line',{x1:40,x2:880,y1:120,y2:120,stroke:'#536779'});
    for(const i of series.intervals){if(i.state==='ok')mark('line',{x1:x(i.fromNS),x2:x(i.toNS),y1:120-i.percent,y2:120-i.percent,stroke:'#67ddbb','stroke-width':4});else mark('rect',{x:x(i.fromNS),y:20,width:x(i.toNS)-x(i.fromNS),height:100,fill:i.state==='unknown'?'#536779':'#183044',opacity:.45});}
    box.append(svg,node('p','Error percentage by observed interval · UTC · teal: supported ratio; grey: unknown; dark: no data / zero traffic. Intervals are not joined across gaps.'));
    const table=node('table');table.className='error-ratio-table';const header=node('tr');['Interval · UTC','Calls','Errors','Error rate','Investigate'].forEach(t=>header.append(node('th',t)));table.append(header);
    for(const interval of series.intervals){const row=node('tr');row.dataset.state=interval.state;row.append(node('td',new Date(Number(BigInt(interval.fromNS)/1000000n)).toISOString().slice(11,23)+' → '+new Date(Number(BigInt(interval.toNS)/1000000n)).toISOString().slice(11,23)),node('td',interval.calls??'unknown'),node('td',interval.errors??'unknown'),node('td',label(interval)));const cell=node('td');
     if(interval.state==='ok'&&BigInt(interval.errors)>0n){const key=JSON.stringify({contract:entry.contract.name,stream:series.stream,fromNS:interval.fromNS,toNS:interval.toNS});const b=node('button','Why did errors rise?');b.onclick=()=>{
      selected=key;selection.replaceChildren(node('h3','Investigate observed errors'),node('p',`${attrs['goo.operation.name']} · ${interval.errors} errors / ${interval.calls} calls · ${label(interval)}. This is one observed interval, not an automated increase or cause detector.`),node('p','Choose an error outcome, inspect time-matched spans, then open a trace waterfall. Counts may differ because metric and span queries have separate evidence bases.'));
      for(const outcome of entry.contract.errorOutcomes){const next=node('button','Inspect '+outcome+' context');next.onclick=()=>{choose({operation:attrs['goo.operation.name'],role:attrs['goo.operation.role'],outcome,fromNS:interval.fromNS,toNS:interval.toNS});selection.append(node('p','Context selected. Open an available trace below; an empty selection is not proof that the metric is incorrect.'));};selection.append(next)}
     };choices.set(key,b);cell.append(b)}else cell.append(node('span',interval.reason||'No observed errors'));row.append(cell);table.append(row);}
    const scroll=node('div');scroll.style.maxHeight='300px';scroll.style.overflow='auto';scroll.append(table);box.append(scroll);container.append(box);
   }
   const explain=node('details');explain.append(node('summary','Ratio policy and evidence basis'),node('pre',JSON.stringify({contract:entry.contract,policy:entry.result.policy,basis:entry.result.basis,counterBasis:entry.result.counterBasis},null,2)));container.append(explain);
  }
  const selection=node('section');selection.id=prefix+'errorRatioSelection';container.append(selection);
  return {getState:()=>selected,restore:key=>{if(choices.has(key)){container.open=true;choices.get(key).click()}}};
 }
 const api={label,render};if(typeof module!=='undefined')module.exports=api;else global.GooRatios=api;
})(typeof window==='undefined'?globalThis:window);
