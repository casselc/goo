/* Rendering consumes validated metric distributions, never raw exporter fields. */

const global=globalThis;
function node(tag,text){const n=document.createElement(tag);if(text!==undefined)n.textContent=text;return n;}
function bucketLabel(bounds,index,unit){return `${index? '('+bounds[index-1]: '(-∞'}, ${index<bounds.length?bounds[index]+']':'+∞)'} ${unit}`;}
function p95Label(range,unit){if(range.state==='no_data')return 'P95: no observations';return `P95 bucket range: ${range.lower===undefined?'(-∞':'('+range.lower}, ${range.upper===undefined?'+∞)':range.upper+']'} ${unit} · no interpolation`;}
function render(container,data,choose){
  container.hidden=false;container.replaceChildren(node('summary','Native duration histogram heatmap'),node('p',data.warnings.join(' ')));
  for(const entry of data.histograms){
   if(!entry.result.series.length){container.append(node('p','No observed histogram intervals — no data, not zero.'));continue;}
   for(const series of entry.result.series){
    const attrs=series.stream.attributes,box=node('section'),selection=node('div');selection.className='histogramSelection';
    box.append(node('h3',`${attrs['goo.operation.name']} / ${attrs['goo.operation.outcome']} · ${series.stream.unit}`),node('p',`Window coverage: ${series.state}. Columns are returned intervals; uncovered edges are not filled. Grey ? cells are unknown, 0 cells are observed empty buckets.`));
    const intervals=series.intervals,bounds=entry.contract.bounds,table=node('table');table.className='heatmap native-histogram';const header=node('tr');header.append(node('th','Duration / interval end · UTC'));
    for(const interval of intervals){const cell=node('th',new Date(Number(BigInt(interval.toNS)/1000000n)).toISOString().slice(11,19));cell.title=interval.fromNS+' → '+interval.toNS+' ns';header.append(cell);}table.append(header);
    if(!intervals.length)box.append(node('p','A baseline alone cannot establish a histogram delta.'));
    let max=1n;for(const interval of intervals)for(const value of interval.distribution?.bucketCounts||[])if(BigInt(value)>max)max=BigInt(value);
    for(let bucket=bounds.length;bucket>=0;bucket--){const row=node('tr');row.append(node('th',bucketLabel(bounds,bucket,series.stream.unit)));
     for(const interval of intervals){const known=interval.state==='ok',count=known?interval.distribution.bucketCounts[bucket]:null,cell=node('td'),button=node('button',known?count:'?');button.className='heat-cell';button.disabled=!known||BigInt(count)===0n;button.dataset.state=interval.state;if(known)button.dataset.count=count;
      button.setAttribute('aria-label',`${attrs['goo.operation.name']} ${bucketLabel(bounds,bucket,series.stream.unit)}: ${known?count+' observations':'unknown '+interval.reason}`);
      if(known&&BigInt(count)>0n)button.style.background=`rgba(57,201,173,${.18+.82*Math.log1p(Number(BigInt(count)))/Math.log1p(Number(max))})`;
      button.onclick=()=>{selection.replaceChildren(node('h3',bucketLabel(bounds,bucket,series.stream.unit)),node('p',`${count} observed bucket population · ${interval.fromNS} → ${interval.toNS} ns`),node('p',p95Label(interval.distribution.p95,series.stream.unit)),node('p','Span navigation is time/operation context only; it does not assert bucket membership or exact contributors.'));const inspect=node('button','Inspect histogram context');inspect.onclick=()=>choose({operation:attrs['goo.operation.name'],role:attrs['goo.operation.role'],outcome:attrs['goo.operation.outcome'],fromNS:interval.fromNS,toNS:interval.toNS});selection.append(inspect);};cell.append(button);row.append(cell);
     }table.append(row);
    }
    const scroll=node('div');scroll.className='heatmap-scroll';scroll.append(table);box.append(scroll,selection);container.append(box);
   }
   const explain=node('details');explain.append(node('summary','Histogram contract and evidence basis'),node('pre',JSON.stringify({contract:entry.contract,basis:entry.result.basis,fromNS:entry.result.fromNS,toNS:entry.result.toNS},null,2)));container.append(explain);
  }
 }

export {bucketLabel,p95Label,render};
