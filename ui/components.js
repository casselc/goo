/* Shared presentation primitives. Query semantics and source reads belong to callers. */
(function(global){
 function node(tag,text,cls){const e=document.createElement(tag);if(text!==undefined)e.textContent=text;if(cls)e.className=cls;return e}
 function disclosure(label,children=[],{open=false,className=''}={}){const e=node('details',undefined,'ui-disclosure '+className);e.open=open;e.append(node('summary',label),...children);return e}
 function renderStats(items,{className=''}={}){const list=node('dl',undefined,'ui-stats '+className);for(const item of items){const e=node('div');e.append(node('dt',item.label),node('dd',item.value??'Unavailable'));if(item.note)e.append(node('small',item.note));list.append(e)}return list}
 function actions(primary,secondary,{label='More actions'}={}){const group=node('div',undefined,'ui-actions');group.append(...primary);if(secondary.length){const body=node('div',undefined,'ui-actions-body');body.append(...secondary);const more=disclosure(label,[body],{className:'ui-action-menu'});group.append(more);more.addEventListener('keydown',event=>{if(event.key==='Escape'){more.open=false;more.querySelector('summary').focus()}});body.addEventListener('click',event=>{if(event.target.closest('button,a')&&!event.target.disabled)more.open=false})}return group}
 function durationMS(ns,count='1'){if(ns===undefined||ns===null||count===undefined||count===null)return 'Unavailable';const n=BigInt(ns),d=BigInt(count)*1000000n;if(d===0n)return '—';const v=(n*1000n+d/2n)/d;return String(v/1000n)+'.'+String(v%1000n).padStart(3,'0')+' ms'}
 function latencyModel(d,ceilingNs){const valid=v=>typeof v==='string'&&/^\d+$/.test(v)&&v.length<=128;if(!d||d.state!=='observed'||!valid(d.count)||BigInt(d.count)===0n)return {state:d?.state==='observed'?'unavailable':d?.state||'unavailable',quantiles:[]};const quantiles=[['p50','p50Ns'],['p95','p95Ns'],['p99','p99Ns']].filter(([,k])=>valid(d[k])).map(([label,key])=>({label,ns:d[key],value:durationMS(d[key])}));let maximum=quantiles.reduce((m,p)=>BigInt(p.ns)>m?BigInt(p.ns):m,0n);if(ceilingNs!==undefined&&ceilingNs!==null){if(!valid(ceilingNs)||BigInt(ceilingNs)<maximum)return {state:'unavailable',quantiles:[]};maximum=BigInt(ceilingNs)}return {state:d.state,count:d.count,ceilingNs:maximum.toString(),ceiling:durationMS(maximum.toString()),mean:valid(d.sumNs)?durationMS(d.sumNs,d.count):'Unavailable',quantiles:quantiles.map(p=>({...p,position:maximum===0n?0:Number(BigInt(p.ns)*10000n/maximum)/100})),min:valid(d.minNs)?durationMS(d.minNs):'Unavailable',max:valid(d.maxNs)?durationMS(d.maxNs):'Unavailable'}}
 function renderLatency(host,d,{title='Span latency',headingLevel=3,ceilingNs}={}){const model=latencyModel(d,ceilingNs),root=node('section',undefined,'ui-latency');root.dataset.component='latency';root.dataset.state=model.state;if(model.ceilingNs!==undefined)root.dataset.ceilingNs=model.ceilingNs;root.append(node('h'+(Number.isInteger(headingLevel)&&headingLevel>=2&&headingLevel<=6?headingLevel:3),title,'ui-latency-title'));if(model.state!=='observed'||!model.count||BigInt(model.count)===0n){root.append(node('p',model.state==='no_data'?'No captured spans. Duration measures are unavailable.':'Span duration measures unavailable.'));host.append(root);return root}const body=node('div',undefined,'ui-latency-body'),mean=renderStats([{label:'Mean',value:model.mean}],{className:'ui-latency-mean'}),quantiles=node('div',undefined,'ui-quantiles');quantiles.append(renderStats(model.quantiles));const rail=node('div',undefined,'ui-quantile-rail');rail.setAttribute('aria-hidden','true');for(const p of model.quantiles){const marker=node('span',undefined,'ui-quantile-marker');marker.style.left=p.position+'%';marker.title=p.label+' · '+p.value;marker.dataset.quantile=p.label;marker.dataset.ns=p.ns;rail.append(marker)}const scale=node('small','0–'+model.ceiling+' · percentile positions');body.append(mean,quantiles);root.append(body);const exact=renderStats([{label:'Min',value:model.min},{label:'Max',value:model.max},{label:'Captured spans',value:model.count},{label:'Duration sum (ns)',value:d.sumNs}]);for(const p of model.quantiles)exact.append(renderStats([{label:p.label+' exact (ns)',value:p.ns}]).firstChild);root.append(disclosure('Latency details',[rail,scale,exact,node('p','Full captured span population. Exact upper order-statistic percentiles, displayed to 0.001 ms. Percentile markers are not a histogram or end-to-end request latency. Sampling and source completeness remain unknown.')]));host.append(root);return root}
 // App-owned selection keys and content; no query, navigation or subscriptions.
 function inspector(host,{headingLevel=4,onClose=()=>{}}={}){
  const heading=node('h'+(Number.isInteger(headingLevel)&&headingLevel>=2&&headingLevel<=6?headingLevel:4)),header=node('div',undefined,'ui-inspector-heading'),close=node('button','Close details'),summary=node('div',undefined,'ui-inspector-summary'),body=node('div',undefined,'ui-inspector-body'),controls=node('div',undefined,'ui-inspector-actions'),cache=new Map();let active=null,opener=null;
  host.classList.add('ui-inspector');host.setAttribute('role','region');host.setAttribute('aria-label','Selected item details');host.hidden=true;close.type='button';close.setAttribute('aria-label','Close selected item details');header.append(heading,close);host.append(header,summary,body,controls);
  function remember(){if(!host.hidden&&active!==null&&cache.has(active))cache.get(active).scroll=body.scrollTop}
  close.onclick=()=>{remember();host.hidden=true;onClose(active,opener);if(opener?.isConnected)opener.focus({preventScroll:true})};
  return {has:key=>cache.has(key),show({key,title,description,content,actions=[],opener:origin}){if(typeof key!=='string'||!key)throw Error('Inspector requires an opaque selection key');remember();if(!cache.has(key)){if(!(content instanceof Node))throw Error('Inspector requires caller-owned content');cache.set(key,{content,actions,description,scroll:0});if(cache.size>32)cache.delete(cache.keys().next().value)}const entry=cache.get(key);cache.delete(key);cache.set(key,entry);heading.textContent=title;const text=description??entry.description??'';summary.replaceChildren();if(text instanceof Node)summary.append(text);else summary.textContent=text;if(active!==key||body.firstChild!==entry.content){body.replaceChildren(entry.content);controls.replaceChildren(...entry.actions)}active=key;opener=origin;host.dataset.key=key;host.hidden=false;body.scrollTop=entry.scroll;},clear(){active=null;opener=null;cache.clear();body.replaceChildren();controls.replaceChildren();summary.replaceChildren();heading.textContent='';delete host.dataset.key;host.hidden=true}};
 }
 function parseAttribute(element,name){try{const value=element.getAttribute(name);return value&&value.length<=65536?JSON.parse(value):null}catch{return null}}
 function latency(host,summary,{title='Span latency',headingLevel=3}={}){const e=node('goo-latency');e.heading=title;e.headingLevel=headingLevel;e.summary=summary;host.append(e);return e}
 function stats(items,{className=''}={}){const e=node('goo-stat-list',undefined,className);e.items=items;return e}
 if(global.customElements&&global.HTMLElement){
  // These shells never replace children: adapters retain row/action identity and focus.
  class GooFacetList extends global.HTMLElement{
   static get observedAttributes(){return ['query']}
   get query(){return this.getAttribute('query')||''}set query(value){this.setAttribute('query',value??'')}
   connectedCallback(){if(!this.hasAttribute('role'))this.setAttribute('role','group');if(Object.hasOwn(this,'query')){const value=this.query;delete this.query;this.query=value}this.refresh()}
   attributeChangedCallback(){if(this.isConnected)this.refresh()}
   refresh(){const q=this.query.toLocaleLowerCase(),rows=[...this.querySelectorAll(':scope > [data-facet-item]')];for(const row of rows)row.hidden=!(row.getAttribute('data-facet-text')||'').toLocaleLowerCase().includes(q);return {visible:rows.filter(row=>!row.hidden).length,total:rows.length}}
  }
  if(!global.customElements.get('goo-facet-list'))global.customElements.define('goo-facet-list',GooFacetList);
  class GooTable extends global.HTMLElement{
   connectedCallback(){if(!this.hasAttribute('tabindex'))this.tabIndex=0;if(!this.hasAttribute('role'))this.setAttribute('role','region');if(!this.hasAttribute('aria-label')&&!this.hasAttribute('aria-labelledby'))this.setAttribute('aria-label','Data table');}
  }
  class GooResultState extends global.HTMLElement{
   static get observedAttributes(){return ['state','announce']}
   get state(){return this.getAttribute('state')||'unknown'}set state(value){this.setAttribute('state',value)}
   connectedCallback(){if(Object.hasOwn(this,'state')){const value=this.state;delete this.state;this.state=value}if(!this.hasAttribute('role'))this.setAttribute('role','status');if(!this.hasAttribute('aria-atomic'))this.setAttribute('aria-atomic','true');this.syncAnnouncement()}
   attributeChangedCallback(){if(this.isConnected)this.syncAnnouncement()}
   syncAnnouncement(){this.setAttribute('aria-live',this.getAttribute('announce')==='off'?'off':'polite')}
  }
  if(!global.customElements.get('goo-table'))global.customElements.define('goo-table',GooTable);
  if(!global.customElements.get('goo-result-state'))global.customElements.define('goo-result-state',GooResultState);
  class GooLatency extends global.HTMLElement{
   static get observedAttributes(){return ['summary','heading','heading-level','ceiling-ns']}
   set summary(value){this._summary=structuredClone(value);if(this.isConnected)this.render()}
   get summary(){return structuredClone(this._summary===undefined?parseAttribute(this,'summary'):this._summary)}
   set ceilingNs(value){if(value===undefined||value===null)this.removeAttribute('ceiling-ns');else this.setAttribute('ceiling-ns',value)}get ceilingNs(){return this.getAttribute('ceiling-ns')??undefined}
   set heading(value){this.setAttribute('heading',value)}get heading(){return this.getAttribute('heading')||'Span latency'}
   set headingLevel(value){this.setAttribute('heading-level',value)}get headingLevel(){const n=Number(this.getAttribute('heading-level'));return Number.isInteger(n)&&n>=2&&n<=6?n:3}
   connectedCallback(){for(const key of ['summary','heading','headingLevel','ceilingNs'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}attributeChangedCallback(name){if(name==='summary')this._summary=undefined;if(this.isConnected)this.render()}
   render(){this.replaceChildren();renderLatency(this,this.summary,{title:this.heading,headingLevel:this.headingLevel,ceilingNs:this.ceilingNs})}
  }
  class GooStatList extends global.HTMLElement{
   static get observedAttributes(){return ['items']}
   set items(value){this._items=structuredClone(value);if(this.isConnected)this.render()}
   get items(){return structuredClone(this._items===undefined?parseAttribute(this,'items'):this._items)}
   connectedCallback(){if(Object.hasOwn(this,'items')){const value=this.items;delete this.items;this.items=value}this.render()}attributeChangedCallback(){this._items=undefined;if(this.isConnected)this.render()}
   render(){const items=this.items;this.replaceChildren(renderStats(Array.isArray(items)?items.slice(0,32).filter(x=>x&&typeof x.label==='string'&&['string','number'].includes(typeof x.value)):[]))}
  }
  if(!global.customElements.get('goo-latency'))global.customElements.define('goo-latency',GooLatency);
  if(!global.customElements.get('goo-stat-list'))global.customElements.define('goo-stat-list',GooStatList);
 }
 const api={node,disclosure,stats,actions,inspector,durationMS,latencyModel,latency,parseAttribute};if(typeof module!=='undefined')module.exports=api;else global.GooUI=api;
})(typeof window==='undefined'?globalThis:window);
