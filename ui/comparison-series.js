/* Passive, exact count intervals. No resampling, transport, or question ownership. */
(function(global){
 const integer=v=>Number.isSafeInteger(v)&&v>=0&&v<=4102444800000,uint=v=>typeof v==='string'&&/^(0|[1-9][0-9]{0,127})$/.test(v);
 function model(data,alignment='elapsed'){
  const bad={state:'unavailable',sides:[]};if(!['elapsed','utc'].includes(alignment)||!Array.isArray(data?.sides)||data.sides.length!==2)return bad;
  let ceiling=0n;const sides=[];
  for(const s of data.sides){if(!s||typeof s.label!=='string'||s.label.length>512||!integer(s.fromMs)||!integer(s.toMs)||s.toMs<=s.fromMs||!Array.isArray(s.points)||s.points.length>240)return bad;
   let end=s.fromMs;const points=[];
   for(const p of s.points){if(!p||!integer(p.fromMs)||!integer(p.toMs)||p.fromMs<end||p.toMs<=p.fromMs||p.toMs>s.toMs||!['observed','no_data','unknown','unavailable'].includes(p.state))return bad;
    const known=['observed','no_data'].includes(p.state);if(known&&(!uint(p.count)||p.state==='no_data'&&p.count!=='0'))return bad;
    if(known&&BigInt(p.count)>ceiling)ceiling=BigInt(p.count);points.push({...p,count:known?p.count:null});end=p.toMs;
   }sides.push({label:s.label,fromMs:s.fromMs,toMs:s.toMs,points});
  }
  const origin=alignment==='utc'?Math.min(...sides.map(s=>s.fromMs)):0,limit=alignment==='utc'?Math.max(...sides.map(s=>s.toMs)):Math.max(...sides.map(s=>s.toMs-s.fromMs)),width=BigInt(limit-origin);
  for(const s of sides){s.gaps=[];let cursor=s.fromMs;for(const p of s.points){if(p.fromMs>cursor)s.gaps.push({fromMs:cursor,toMs:p.fromMs});cursor=p.toMs}if(cursor<s.toMs)s.gaps.push({fromMs:cursor,toMs:s.toMs});const offset=alignment==='utc'?0:s.fromMs;s.left=Number(BigInt(s.fromMs-offset-origin)*100000n/width)/1000;s.width=Number(BigInt(s.toMs-s.fromMs)*100000n/width)/1000;
   for(const p of [...s.points,...s.gaps]){p.left=Number(BigInt(p.fromMs-offset-origin)*100000n/width)/1000;p.width=Number(BigInt(p.toMs-p.fromMs)*100000n/width)/1000;p.height=p.count!=null&&ceiling?Number(BigInt(p.count)*100000n/ceiling)/1000:0;}
  }return {state:'observed',sides,ceiling:String(ceiling),alignment,origin,limit};
 }
 function brush(data,alignment,side,a,b){const m=model(data,alignment);if(m.state!=='observed'||![0,1].includes(side)||![a,b].every(v=>Number.isFinite(v)&&v>=0&&v<=1))return null;const s=m.sides[side],width=BigInt(m.limit-m.origin),offset=alignment==='utc'?m.origin:s.fromMs;const point=v=>offset+Number(width*BigInt(Math.round(v*1000000))/1000000n);const fromMs=Math.max(s.fromMs,point(Math.min(a,b))),toMs=Math.min(s.toMs,point(Math.max(a,b)));return toMs>fromMs?{side,fromMs,toMs}:null}
 const label=(s,p)=>s.label+' · ['+new Date(p.fromMs).toISOString()+', '+new Date(p.toMs).toISOString()+') · '+(p.count===null?p.state:p.count+' records'+(p.state==='no_data'?' · no captured records; health unknown':''));
 if(global.customElements&&global.HTMLElement){class Series extends global.HTMLElement{
  static get observedAttributes(){return ['data','alignment']}
  set data(v){try{this._data=structuredClone(v)}catch{this._data=null}if(this.isConnected)this.render()}
  get data(){if(this._data!==undefined)return structuredClone(this._data);try{const v=this.getAttribute('data');return v&&v.length<=262144?JSON.parse(v):null}catch{return null}}
  connectedCallback(){if(Object.hasOwn(this,'data')){const v=this.data;delete this.data;this.data=v}this.render()}
  attributeChangedCallback(k){if(k==='data')this._data=undefined;if(this.isConnected)this.render()}
  disconnectedCallback(){this._cancel?.()}
  render(){this._cancel?.();const node=(tag,text,cls)=>{const e=document.createElement(tag);if(text!==undefined)e.textContent=text;if(cls)e.className=cls;return e},m=model(this.data,this.getAttribute('alignment')||'elapsed');this.replaceChildren();this.dataset.state=m.state;if(m.state!=='observed'){this.append(node('p','Count series unavailable; no values inferred.'));return}
   const fig=node('figure'),caption=node('figcaption','Records per interval · shared 0–'+m.ceiling),plot=node('div',undefined,'comparison-series-plot'),status=node('p','Focus an interval for exact UTC and count.','comparison-series-detail');status.setAttribute('role','status');
   m.sides.forEach((s,i)=>{for(const p of s.gaps){const gap=node('div',undefined,'comparison-series-gap');gap.style.left=p.left+'%';gap.style.width=p.width+'%';gap.dataset.side=String(i);gap.title=s.label+' · no supplied interval';plot.append(gap)}const window=node('div',undefined,'comparison-series-window');window.style.left=s.left+'%';window.style.width=s.width+'%';window.dataset.side=String(i);plot.append(window);s.points.forEach((p,index)=>{const b=node('button',undefined,'comparison-series-mark');b.type='button';b.dataset.side=String(i);b.dataset.state=p.state;b.dataset.count=p.count??'';b.style.left=p.left+'%';b.style.width=p.width+'%';b.style.setProperty('--height',p.height+'%');const text=label(s,p);b.title=text;b.setAttribute('aria-label',text);b.onfocus=b.onmouseenter=()=>status.textContent=text;b.onclick=()=>{status.textContent=text;this.dispatchEvent(new CustomEvent('goo-series-select',{bubbles:true,detail:{side:i,index,fromMs:p.fromMs,toMs:p.toMs}}))};plot.append(b)})});
   const axis=node('p',m.alignment==='utc'?new Date(m.origin).toISOString()+' → '+new Date(m.limit).toISOString():'Elapsed from each window start · 0–'+(m.limit/1000)+' seconds','comparison-series-axis');this.bindBrush(plot);fig.append(caption,plot,axis,status);const exact=node('details'),summary=node('summary','Exact intervals and counts'),wrap=node('div',undefined,'comparison-series-table'),table=node('table'),thead=node('thead'),head=node('tr'),body=node('tbody');for(const text of ['Side','From (UTC)','To (UTC, exclusive)','Records','State']){const th=node('th',text);th.scope='col';head.append(th)}thead.append(head);for(const s of m.sides)for(const p of s.points){const row=node('tr');for(const text of [s.label,new Date(p.fromMs).toISOString(),new Date(p.toMs).toISOString(),p.count??'Unavailable',p.state])row.append(node('td',text));body.append(row)}table.append(thead,body);wrap.append(table);exact.append(summary,wrap);exact.append(node('p','Drag from an interval to select a range; Escape cancels. Dashed gaps have no supplied interval; stripes are unknown. No interpolation.'));this.append(fig,node('p','Blue: '+m.sides[0].label+' · Orange: '+m.sides[1].label+' · Counts, not rates.'),exact);
  }
  bindBrush(plot){let gesture=null,suppress=false;const cancel=()=>{const g=gesture;gesture=null;if(g&&plot.hasPointerCapture(g.id))plot.releasePointerCapture(g.id);plot.style.removeProperty('--selection-left');plot.style.removeProperty('--selection-width')};this._cancel=cancel;
   plot.addEventListener('click',e=>{if(suppress&&e.isTrusted&&e.detail){e.preventDefault();e.stopImmediatePropagation();suppress=false}},true);
   plot.onpointerdown=e=>{const mark=e.target.closest('.comparison-series-mark');if(!mark||e.button!==0||!e.isPrimary||e.pointerType==='touch')return;const rect=plot.getBoundingClientRect();if(!rect.width)return;suppress=false;gesture={id:e.pointerId,start:e.clientX,rect,side:Number(mark.dataset.side),moved:false};mark.focus({preventScroll:true});plot.setPointerCapture(e.pointerId)};
   plot.onpointermove=e=>{const g=gesture;if(!g||g.id!==e.pointerId||Math.abs(g.start-e.clientX)<4&&!g.moved)return;g.moved=true;const fraction=x=>Math.max(0,Math.min(1,(x-g.rect.left)/g.rect.width));g.bounds=brush(this.data,this.getAttribute('alignment')||'elapsed',g.side,fraction(g.start),fraction(e.clientX));plot.style.setProperty('--selection-left',Math.min(fraction(g.start),fraction(e.clientX))*100+'%');plot.style.setProperty('--selection-width',Math.abs(fraction(g.start)-fraction(e.clientX))*100+'%')};
   plot.onpointerup=e=>{const g=gesture;if(!g||g.id!==e.pointerId)return;cancel();if(g.moved){suppress=true;if(g.bounds)this.dispatchEvent(new CustomEvent('goo-series-range',{bubbles:true,detail:g.bounds}))}};
   plot.onpointercancel=plot.onlostpointercapture=()=>cancel();plot.onkeydown=e=>{if(e.key==='Escape'){suppress=true;cancel()}};
  }

 }if(!global.customElements.get('goo-comparison-series'))global.customElements.define('goo-comparison-series',Series)}
 const api={model,label,brush};if(typeof module!=='undefined')module.exports=api;else global.GooComparisonSeries=api;
})(typeof window==='undefined'?globalThis:window);
