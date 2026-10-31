import {bindIntervalRange} from './interval-range.mjs';
/* Supplied count-bar snapshots. Question scope and source actions belong to adapters. */
const global = globalThis;
 const uint=value=>typeof value==='string'&&/^(0|[1-9][0-9]*)$/.test(value)&&value.length<=128;
 function rational(value){if(typeof value!=='string'||value.length>256||!/^(0|[1-9]\d*)(?:\/([1-9]\d*))?$/.test(value))return null;const [n,d='1']=value.split('/');return {n:BigInt(n),d:BigInt(d)}}
 function model(data){
  if(!data||!(data.valueFormat==='rational'?rational(data.ceiling):uint(data.ceiling))||!Array.isArray(data.points)||data.points.length>257)return {state:'unavailable',points:[]};
  const parse=data.valueFormat==='rational'?rational:v=>uint(v)?{n:BigInt(v),d:1n}:null,ceiling=parse(data.ceiling),keys=new Set(),points=[];
  for(const p of data.points){
   if(!p||typeof p.key!=='string'||!p.key||p.key.length>512||keys.has(p.key)||typeof p.label!=='string')return {state:'unavailable',points:[]};keys.add(p.key);
   const state=p.state||'observed';if(!['observed','no_data','unknown','unavailable'].includes(state))return {state:'unavailable',points:[]};
   const known=state==='observed'||state==='no_data',parsed=known?parse(p.value):null;if(known&&(!parsed||parsed.n*ceiling.d>ceiling.n*parsed.d||state==='no_data'&&parsed.n!==0n))return {state:'unavailable',points:[]};
   const value=known?p.value:null,positive=known&&parsed.n>0n;
   points.push({key:p.key,label:p.label,state,value,positive,height:positive&&ceiling.n>0n?Number(parsed.n*ceiling.d*10000n/(parsed.d*ceiling.n))/100:0,...(typeof p.href==='string'?{href:p.href}:{})});
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
    this._rangeGesture?.clear();
    this.onpointerdown=this.onpointermove=this.onpointerup=this.onpointercancel=this.onkeydown=null;this._ignoreClick=false;this._rangeGesture=null;this._clearRange=null;const m=model(this.data);this.replaceChildren();this.dataset.state=m.state;
    if(m.state!=='observed'){this.append(node('p','Time chart unavailable.','time-bars-unavailable'));return}
    const figure=node('figure',undefined,'time-bars-figure'),caption=node('figcaption',m.title),scale=node('p','0–'+m.ceiling+' '+m.unit+' · supplied scale','time-bars-scale'),bars=node('div',undefined,'time-bars');
    for(const p of m.points){const href=safeHref(p.href),tag=href?'a':m.interactive?'button':'span',bar=node(tag,undefined,'time-bar');if(tag==='button')bar.type='button';if(href)bar.href=href;
     bar.dataset.key=p.key;bar.dataset.state=p.state;bar.dataset.positive=String(p.positive);if(p.value!==null)bar.dataset.count=p.value;bar.style.setProperty('--bar-height',p.height+'%');bar.setAttribute('aria-label',p.label);bar.title=p.label;
     if(tag==='span'){bar.setAttribute('role','img')}
     if(tag==='button'){bar.setAttribute('aria-pressed',String(this.selectedKey===p.key));bar.addEventListener('click',e=>{if(!this.isConnected||!this.contains(bar))return;if(this._rangeGesture?.ignoreClick(e)){return}this.dispatchEvent(new CustomEvent('goo-time-select',{bubbles:true,detail:{key:p.key}}))})}else if(href&&this.selectedKey===p.key)bar.setAttribute('aria-current','true');
     if(p.state==='no_data')bar.classList.add('no-data');bars.append(bar)
    }
    const axis=node('div',undefined,'time-axis');axis.append(node('span',m.fromLabel),node('span',m.toLabel));figure.append(caption,scale,bars,axis);this.append(figure);if(m.interactive&&this.hasAttribute('range-select')&&!m.points.some(p=>p.href))this.bindRange(m);
   }

   bindRange(m){
    this._rangeGesture=bindIntervalRange(this,{plot:()=>this.querySelector(':scope > .time-bars-figure > .time-bars'),items:()=>[...(this.querySelector(':scope > .time-bars-figure > .time-bars')?.children||[])],keys:m.points.map(point=>point.key),onRange:(fromKey,toKey)=>this.dispatchEvent(new CustomEvent('goo-time-range',{bubbles:true,detail:{fromKey,toKey}}))});
    this._clearRange=()=>this._rangeGesture.clear();
   }

  }
  if(!global.customElements.get('goo-time-bars'))global.customElements.define('goo-time-bars',GooTimeBars);
 }
 export {model,safeHref,rational};
