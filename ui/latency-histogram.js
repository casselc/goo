/* App adapter: supplied full-scope raw-span buckets, existing staged query owner. */
(function(global){
 const dependency=name=>global[name]||(typeof module!=='undefined'?require(name==='GooDiscovery'?'./discovery':'./population-selection'):null);
 function model(data){
  const order=dependency('GooDiscovery').durationOrder,q=data?.question;
  if(q?.signal!=='spans'||q.numeric||data.numeric||q.groupBy?.length!==1||q.groupBy[0].kind!=='column'||q.groupBy[0].key!=='duration_bucket'||data.state!=='observed'||!Array.isArray(data.groups)||typeof data.count!=='string'||!/^(0|[1-9]\d*)$/.test(data.count)||data.count.length>20||BigInt(data.count)>18446744073709551615n||data.count==='0')return null;
  const counts=new Map();for(const g of data.groups){if(!order.includes(g.g0)||counts.has(g.g0)||typeof g.count!=='string'||!/^(0|[1-9]\d*)$/.test(g.count)||g.count.length>20||BigInt(g.count)>18446744073709551615n)return null;counts.set(g.g0,g.count)}
  if([...counts.values()].reduce((a,b)=>a+BigInt(b),0n)!==BigInt(data.count))return null;
  const points=order.map((label,i)=>({key:String(i),label:label+' · '+(counts.get(label)||'0')+' captured spans',value:counts.get(label)||'0',state:'observed'}));
  return {title:'Span latency distribution',unit:'spans per duration bucket',ceiling:points.reduce((a,p)=>BigInt(p.value)>BigInt(a)?p.value:a,'0'),fromLabel:'0 ms',toLabel:'1,000+ ms · unequal-width buckets',interactive:true,points};
 }
 function selection(data,fromKey,toKey){
  if(!model(data))throw Error('A consistent full-scope raw-span duration distribution is required.');
  if(!/^[0-7]$/.test(fromKey)||!/^[0-7]$/.test(toKey))throw Error('Invalid latency bucket range.');
  const a=Math.min(Number(fromKey),Number(toKey)),b=Math.max(Number(fromKey),Number(toKey)),edges=['0','1','5','10','50','100','500','1000',''];
  return {fromMs:data.question.fromMs,toMs:data.question.toMs,minMS:edges[a],maxMS:edges[b+1]};
 }
 function mount(host,data,{open,readOnly=false,onSelect,onBegin,compact=false}={}){
  const supplied=model(data);if(!supplied)return false;
  const node=global.GooUI.node,section=node('section',undefined,'latency-histogram'),bars=node('goo-time-bars');section.setAttribute('aria-label','Span latency histogram');bars.data={...supplied,interactive:!readOnly&&!!(open||onSelect)};section.append(bars);const labels=node('div',undefined,'latency-bucket-labels');for(const label of ['0–1','1–5','5–10','10–50','50–100','100–500','500–1k','1k+'])labels.append(node('span',label));section.append(labels);host.append(section);
  const hint=node('p',readOnly||!(open||onSelect)?'Supplied bucket counts, not density.': 'Bucket counts, not density. Drag to select; Space to anchor, Shift + arrows to extend, Enter to stage. Touch: tap a bucket or edit exact bounds.');if(compact){bars.setAttribute('aria-description',hint.textContent);hint.title=hint.textContent;hint.textContent='Counts per bucket · select a range'}section.append(hint);
  if(!readOnly&&onSelect){bars.setAttribute('range-select','');bars.addEventListener('pointerdown',()=>onBegin?.(),true);bars.addEventListener('keydown',e=>{if([' ','Enter','ArrowLeft','ArrowRight'].includes(e.key))onBegin?.()},true);const choose=(a,b)=>{const from=Math.min(Number(a),Number(b)),to=Math.max(Number(a),Number(b)),bounds=selection(data,a,b),count=supplied.points.slice(from,to+1).reduce((n,p)=>n+BigInt(p.value),0n);onSelect(bounds,{count:String(count),total:data.count})};bars.addEventListener('goo-time-range',e=>choose(e.detail.fromKey,e.detail.toKey));bars.addEventListener('goo-time-select',e=>choose(e.detail.key,e.detail.key));}
  else if(!readOnly&&open){bars.setAttribute('range-select','');const owner=global.GooPopulationSelection.mount(section,data,{open,onBaseline:true,compact:true});
   const preview=node('p');preview.setAttribute('role','status');section.insertBefore(preview,owner.element);owner.element.addEventListener('input',()=>{preview.textContent='Exact bounds edited. Run to measure this population.';bars.clearRange()});const stage=(a,b)=>{const lo=Math.min(Number(a),Number(b)),hi=Math.max(Number(a),Number(b));selection(data,a,b);const count=supplied.points.slice(lo,hi+1).reduce((n,p)=>n+BigInt(p.value),0n);preview.textContent=count+' of '+data.count+' captured spans selected · '+(BigInt(data.count)-count)+' outside the selected buckets';owner.select(selection(data,a,b))};bars.addEventListener('goo-time-range',e=>stage(e.detail.fromKey,e.detail.toKey));bars.addEventListener('goo-time-select',e=>{bars.selectedKey=e.detail.key;stage(e.detail.key,e.detail.key)});
  }
  return true;
 }
 const api={model,selection,mount};if(typeof module!=='undefined')module.exports=api;else global.GooLatencyHistogram=api;
})(typeof window==='undefined'?globalThis:window);
