/* Observed membership and supported interval coverage are separate facts. */
(function(global){
 function node(tag,text){const n=document.createElement(tag);if(text!==undefined)n.textContent=text;return n;}
 function coverageLabel(i){return `${i.supported}/${i.expected} members support this exact interval`;}
 function comparisonLabel(i){return i.state==='comparable'?'Comparable windows · no combined percentage':'Unknown · no supported match';}
 function renderComparison(container,comparison,prefix=''){
  let selected;const choices=new Map();
  const panel=node('section');panel.id=prefix+'fleetComparison';panel.append(node('h3','Compare actual member windows'),node('p','Policy: '+comparison.policy.mode+' · maximum endpoint skew '+comparison.policy.maxSkewNS+' ns. This permits side-by-side comparison, not exact aggregation or clock-accuracy claims. Exact fleet totals below never use this tolerance.'));
  for(const series of comparison.series){const box=node('section');box.className='fleet-comparison-series';box.append(node('h3',series.stream.attributes['goo.operation.name']+' / '+series.stream.attributes['goo.operation.role']));const table=node('table'),head=node('tr');['Reference window · UTC','Comparison','Start / end spread (ns)',...series.intervals[0]?.members.map(m=>m.instance)||[],'Explain'].forEach(t=>{const th=node('th',t);th.style.position='sticky';th.style.top='0';th.style.background='#122235';head.append(th)});table.append(head);
   for(const i of series.intervals){const row=node('tr');row.className='fleet-comparison-row';row.dataset.state=i.state;row.dataset.fromNs=i.fromNS;row.append(node('td',new Date(Number(BigInt(i.fromNS)/1000000n)).toISOString().slice(11,23)+' → '+new Date(Number(BigInt(i.toNS)/1000000n)).toISOString().slice(11,23)),node('td',comparisonLabel(i)),node('td',(i.startSpreadNS??'unknown')+' / '+(i.endSpreadNS??'unknown')));
    for(const member of i.members){const cell=node('td');cell.style.whiteSpace='pre-line';cell.textContent=member.fromNS?`${new Date(Number(BigInt(member.fromNS)/1000000n)).toISOString().slice(11,23)} → ${new Date(Number(BigInt(member.toNS)/1000000n)).toISOString().slice(11,23)}\n${member.calls??'unknown'} calls · ${member.errors??'unknown'} errors\n${member.state==='ok'?member.percent.toFixed(2)+'%':member.state==='no_data'?'No data · zero completed calls':member.state}`:member.reason;row.append(cell)}
    const cell=node('td'),button=node('button','Explain member comparison');const key=JSON.stringify({stream:series.stream,fromNS:i.fromNS,toNS:i.toNS});button.onclick=()=>{selected=key;selection.replaceChildren(node('h3',comparisonLabel(i)),node('p','Actual timestamps and counts are retained. No common-window fleet percentage is computed. Host clock accuracy and ingestion completeness remain unverified.'),node('pre',JSON.stringify({policy:comparison.policy,referenceInstance:series.referenceInstance,comparison:i,basis:comparison.basis},null,2)));};choices.set(key,button);cell.append(button);row.append(cell);table.append(row)}
   const scroll=node('div');scroll.style.maxHeight='360px';scroll.style.overflow='auto';scroll.append(table);box.append(scroll);panel.append(box)
  }
  const selection=node('section');selection.id=prefix+'fleetComparisonSelection';panel.append(selection);container.append(panel);
  return {getState:()=>selected,restore:key=>{if(choices.has(key)){container.open=true;choices.get(key).click()}}};
 }
 function render(container,entry,{prefix=''}={}){
  container.hidden=false;container.replaceChildren(node('summary','Fleet membership and coverage'),node('p','Membership is declared for the whole window. Observed does not mean complete. Fleet counts require every expected member and exactly aligned, valid outcome intervals.'));
  const result=entry.result,observed=result.members.filter(m=>m.state==='observed').length;
  container.append(node('h3',`${observed}/${result.members.length} declared members observed`));
  const members=node('table');members.id=prefix+'fleetMembers';const header=node('tr');['Expected instance','Source observation'].forEach(t=>header.append(node('th',t)));members.append(header);
  for(const member of result.members){const row=node('tr');row.dataset.instance=member.instance;row.append(node('td',member.instance),node('td',member.state));members.append(row)}container.append(members);
  if(entry.contract.description)container.append(node('p',entry.contract.description));container.append(node('p','Ordinary operation views remain scoped to the selected instance.'));
  const comparison=result.comparison?renderComparison(container,result.comparison,prefix):null;
  container.append(node('h3','Exact fleet totals · identical boundaries required'));
  if(!result.series.length)container.append(node('p','No observed operation streams. No fleet total can be established.'));
  for(const series of result.series){const box=node('section'),attrs=series.stream.attributes;box.append(node('h3',`${attrs['goo.operation.name']} / ${attrs['goo.operation.role']}`),node('p','Window coverage: '+series.state));
   const table=node('table');table.className='fleet-intervals';const h=node('tr');['Interval · UTC','Coverage','Fleet calls','Fleet errors','Fleet error rate','Explain'].forEach(t=>h.append(node('th',t)));table.append(h);
   for(const i of series.intervals){const row=node('tr');row.dataset.state=i.state;row.dataset.supported=i.supported;row.append(node('td',new Date(Number(BigInt(i.fromNS)/1000000n)).toISOString().slice(11,23)+' → '+new Date(Number(BigInt(i.toNS)/1000000n)).toISOString().slice(11,23)),node('td',coverageLabel(i)),node('td',i.calls??'unknown'),node('td',i.errors??'unknown'),node('td',i.state==='ok'?`${i.percent.toFixed(2)}% (${i.errors}/${i.calls})`:i.state==='no_data'?'No data · zero completed calls':'Unknown'));
    const cell=node('td'),b=node('button','Explain fleet coverage');b.onclick=()=>{selection.replaceChildren(node('h3','Fleet interval coverage'),node('p',coverageLabel(i)+'. '+(i.reason||i.state)),node('p','Unknown members do not become zero and supported members do not become a fleet total. Independently timed exports may be unaligned even when all members are observed.'));
     for(const member of i.members)selection.append(node('p',member.instance+' · '+member.state));};cell.append(b);row.append(cell);table.append(row)}
   const scroll=node('div');scroll.style.maxHeight='280px';scroll.style.overflow='auto';scroll.append(table);box.append(scroll);container.append(box);
  }
  const selection=node('section');selection.id=prefix+'fleetCoverageSelection';container.append(selection);
  const explain=node('details');explain.append(node('summary','Fleet membership policy and evidence basis'),node('pre',JSON.stringify({contract:entry.contract,policy:result.policy,basis:result.basis,members:result.members,fromNS:result.fromNS,toNS:result.toNS},null,2)));container.append(explain);
  return comparison;
 }
 const api={coverageLabel,comparisonLabel,render};if(typeof module!=='undefined')module.exports=api;else global.GooFleet=api;
})(typeof window==='undefined'?globalThis:window);
