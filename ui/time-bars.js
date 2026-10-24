/* Supplied count-bar snapshots. Question scope and source actions belong to adapters. */
(function(global){
 const uint=value=>typeof value==='string'&&/^(0|[1-9][0-9]*)$/.test(value)&&value.length<=128;
 function model(data){
  if(!data||!uint(data.ceiling)||!Array.isArray(data.points)||data.points.length>256)return {state:'unavailable',points:[]};
  const ceiling=BigInt(data.ceiling),keys=new Set(),points=[];
  for(const p of data.points){
   if(!p||typeof p.key!=='string'||!p.key||p.key.length>512||keys.has(p.key)||typeof p.label!=='string')return {state:'unavailable',points:[]};keys.add(p.key);
   const state=p.state||'observed';if(!['observed','no_data','unknown','unavailable'].includes(state))return {state:'unavailable',points:[]};
   const known=state==='observed'||state==='no_data';if(known&&(!uint(p.value)||BigInt(p.value)>ceiling||state==='no_data'&&p.value!=='0'))return {state:'unavailable',points:[]};
   const value=known?p.value:null,positive=known&&BigInt(value)>0n;
   points.push({key:p.key,label:p.label,state,value,positive,height:positive&&ceiling>0n?Number(BigInt(value)*10000n/ceiling)/100:0,...(typeof p.href==='string'?{href:p.href}:{})});
  }
  return {state:'observed',title:typeof data.title==='string'?data.title:'Captured records over time',unit:typeof data.unit==='string'?data.unit:'records',ceiling:data.ceiling,fromLabel:typeof data.fromLabel==='string'?data.fromLabel:'',toLabel:typeof data.toLabel==='string'?data.toLabel:'',interactive:data.interactive===true,points};
 }
 function safeHref(value){if(typeof value!=='string'||!value.trim()||value.trim()!==value)return null;try{const url=new URL(value,'https://component.invalid/');return ['https:','http:'].includes(url.protocol)?value:null}catch{return null}}
 function node(tag,text,cls){const e=document.createElement(tag);if(text!==undefined)e.textContent=text;if(cls)e.className=cls;return e}
 if(global.customElements&&global.HTMLElement){
  class GooTimeBars extends global.HTMLElement{
   static get observedAttributes(){return ['data','selected-key','range-select']}
   set data(value){try{this._data=structuredClone(value)}catch{this._data=null}if(this.isConnected)this.render()}
   get data(){if(this._data!==undefined)return structuredClone(this._data);try{const raw=this.getAttribute('data');return raw&&raw.length<=65536?JSON.parse(raw):null}catch{return null}}
   set selectedKey(value){if(value===null||value===undefined)this.removeAttribute('selected-key');else this.setAttribute('selected-key',value)}get selectedKey(){return this.getAttribute('selected-key')}
   connectedCallback(){for(const key of ['data','selectedKey'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
   attributeChangedCallback(name){if(name==='data')this._data=undefined;if(this.isConnected){if(name==='selected-key')this.syncSelection();else this.render()}}
   clearRange(){this._clearRange?.();this.selectedKey=null}
   syncSelection(){for(const bar of this.querySelectorAll(':scope > .time-bars-figure > .time-bars > .time-bar')){const selected=this.selectedKey===bar.dataset.key;if(bar.tagName==='BUTTON')bar.setAttribute('aria-pressed',String(selected));else if(bar.tagName==='A'){if(selected)bar.setAttribute('aria-current','true');else bar.removeAttribute('aria-current')}}}
   render(){
    this.onpointerdown=this.onpointermove=this.onpointerup=this.onpointercancel=this.onkeydown=null;this._ignoreClick=false;this._clearRange=null;const m=model(this.data);this.replaceChildren();this.dataset.state=m.state;
    if(m.state!=='observed'){this.append(node('p','Time chart unavailable.','time-bars-unavailable'));return}
    const figure=node('figure',undefined,'time-bars-figure'),caption=node('figcaption',m.title),scale=node('p','0–'+m.ceiling+' '+m.unit+' · supplied scale','time-bars-scale'),bars=node('div',undefined,'time-bars');
    for(const p of m.points){const href=safeHref(p.href),tag=href?'a':m.interactive?'button':'span',bar=node(tag,undefined,'time-bar');if(tag==='button')bar.type='button';if(href)bar.href=href;
     bar.dataset.key=p.key;bar.dataset.state=p.state;bar.dataset.positive=String(p.positive);if(p.value!==null)bar.dataset.count=p.value;bar.style.setProperty('--bar-height',p.height+'%');bar.setAttribute('aria-label',p.label);bar.title=p.label;
     if(tag==='span'){bar.setAttribute('role','img')}
     if(tag==='button'){bar.setAttribute('aria-pressed',String(this.selectedKey===p.key));bar.addEventListener('click',e=>{if(this._ignoreClick&&e.detail>0){this._ignoreClick=false;return}this.dispatchEvent(new CustomEvent('goo-time-select',{bubbles:true,detail:{key:p.key}}))})}else if(href&&this.selectedKey===p.key)bar.setAttribute('aria-current','true');
     if(p.state==='no_data')bar.classList.add('no-data');bars.append(bar)
    }
    const axis=node('div',undefined,'time-axis');axis.append(node('span',m.fromLabel),node('span',m.toLabel));figure.append(caption,scale,bars,axis);this.append(figure);if(m.interactive&&this.hasAttribute('range-select')&&!m.points.some(p=>p.href))this.bindRange(m,bars);
   }

   bindRange(m,plot){
    const bars=[...plot.children];let anchor=null,end=null,committed=null,pointer=null;
    const at=e=>{const b=e.target.closest?.('.time-bar');return b&&plot.contains(b)?bars.indexOf(b):-1};
    const paint=()=>{const range=anchor===null?committed:[Math.min(anchor,end),Math.max(anchor,end)];bars.forEach((b,i)=>{b.dataset.rangeSelected=String(!!range&&i>=range[0]&&i<=range[1])})};
    const release=()=>{if(pointer!==null&&plot.hasPointerCapture(pointer))plot.releasePointerCapture(pointer);pointer=null};
    const commit=()=>{if(anchor===null)return;committed=[Math.min(anchor,end),Math.max(anchor,end)];this.dispatchEvent(new CustomEvent('goo-time-range',{bubbles:true,detail:{fromKey:m.points[committed[0]].key,toKey:m.points[committed[1]].key}}));anchor=end=null;paint()};
    const cancel=()=>{release();anchor=end=null;paint()};this._clearRange=()=>{committed=null;cancel()};
    this.onpointerdown=e=>{const i=at(e);if(e.button!==0||e.pointerType==='touch'||i<0)return;this._ignoreClick=false;anchor=end=i;pointer=e.pointerId;plot.setPointerCapture(pointer);paint();e.preventDefault()};
    this.onpointermove=e=>{if(pointer!==e.pointerId)return;const x=e.clientX;let nearest=0,distance=Infinity;bars.forEach((b,i)=>{const r=b.getBoundingClientRect(),d=Math.abs(x-(r.left+r.width/2));if(d<distance){distance=d;nearest=i}});end=nearest;paint()};
    this.onpointerup=e=>{if(pointer!==e.pointerId)return;release();this._ignoreClick=true;commit()};this.onpointercancel=cancel;plot.onlostpointercapture=()=>{if(pointer!==null)cancel()};
    this.onkeydown=e=>{const i=at(e);if(i<0)return;if(e.key==='Escape'){cancel();e.preventDefault();return}if(e.key===' '){e.preventDefault();e.stopPropagation();if(anchor===null){anchor=end=i;paint()}else{end=i;commit()}return}if(e.shiftKey&&['ArrowLeft','ArrowRight'].includes(e.key)){e.preventDefault();anchor=anchor??i;end=Math.max(0,Math.min(bars.length-1,i+(e.key==='ArrowRight'?1:-1)));bars[end].focus();paint()}if(e.key==='Enter'&&anchor!==null){e.preventDefault();e.stopPropagation();end=i;commit()}};
   }

  }
  if(!global.customElements.get('goo-time-bars'))global.customElements.define('goo-time-bars',GooTimeBars);
 }
 const api={model,safeHref};if(typeof module!=='undefined')module.exports=api;else global.GooTimeBars=api;
})(typeof window==='undefined'?globalThis:window);
