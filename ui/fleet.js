/* Observed membership and supported interval coverage are separate facts. */
(function(global){
 function node(tag,text){const n=document.createElement(tag);if(text!==undefined)n.textContent=text;return n;}
 function coverageLabel(i){return `${i.supported}/${i.expected} members support this exact interval`;}
 function render(container,entry){
  container.hidden=false;container.replaceChildren(node('summary','Fleet membership and coverage'),node('p','Membership is declared for the whole window. Observed does not mean complete. Fleet counts require every expected member and exactly aligned, valid outcome intervals.'));
  const result=entry.result,observed=result.members.filter(m=>m.state==='observed').length;
  container.append(node('h3',`${observed}/${result.members.length} declared members observed`));
  const members=node('table');members.id='fleetMembers';const header=node('tr');['Expected instance','Source observation'].forEach(t=>header.append(node('th',t)));members.append(header);
  for(const member of result.members){const row=node('tr');row.dataset.instance=member.instance;row.append(node('td',member.instance),node('td',member.state));members.append(row)}container.append(members);
  if(entry.contract.description)container.append(node('p',entry.contract.description));container.append(node('p','Ordinary operation views remain scoped to the selected instance.'));
  if(!result.series.length)container.append(node('p','No observed operation streams. No fleet total can be established.'));
  for(const series of result.series){const box=node('section'),attrs=series.stream.attributes;box.append(node('h3',`${attrs['goo.operation.name']} / ${attrs['goo.operation.role']}`),node('p','Window coverage: '+series.state));
   const table=node('table');table.className='fleet-intervals';const h=node('tr');['Interval · UTC','Coverage','Fleet calls','Fleet errors','Fleet error rate','Explain'].forEach(t=>h.append(node('th',t)));table.append(h);
   for(const i of series.intervals){const row=node('tr');row.dataset.state=i.state;row.dataset.supported=i.supported;row.append(node('td',new Date(Number(BigInt(i.fromNS)/1000000n)).toISOString().slice(11,23)+' → '+new Date(Number(BigInt(i.toNS)/1000000n)).toISOString().slice(11,23)),node('td',coverageLabel(i)),node('td',i.calls??'unknown'),node('td',i.errors??'unknown'),node('td',i.state==='ok'?`${i.percent.toFixed(2)}% (${i.errors}/${i.calls})`:i.state==='no_data'?'No data · zero completed calls':'Unknown'));
    const cell=node('td'),b=node('button','Explain fleet coverage');b.onclick=()=>{selection.replaceChildren(node('h3','Fleet interval coverage'),node('p',coverageLabel(i)+'. '+(i.reason||i.state)),node('p','Unknown members do not become zero and supported members do not become a fleet total. Independently timed exports may be unaligned even when all members are observed.'));
     for(const member of i.members)selection.append(node('p',member.instance+' · '+member.state));};cell.append(b);row.append(cell);table.append(row)}
   const scroll=node('div');scroll.style.maxHeight='280px';scroll.style.overflow='auto';scroll.append(table);box.append(scroll);container.append(box);
  }
  const selection=node('section');selection.id='fleetCoverageSelection';container.append(selection);
  const explain=node('details');explain.append(node('summary','Fleet membership policy and evidence basis'),node('pre',JSON.stringify({contract:entry.contract,policy:result.policy,basis:result.basis,members:result.members,fromNS:result.fromNS,toNS:result.toNS},null,2)));container.append(explain);
 }
 const api={coverageLabel,render};if(typeof module!=='undefined')module.exports=api;else global.GooFleet=api;
})(typeof window==='undefined'?globalThis:window);
