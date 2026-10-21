/* Passive bounded time selection. No query, navigation or source lifecycle. */
(function(global){
 function model(d){if(!d||![d.fromMs,d.toMs].every(Number.isSafeInteger)||d.fromMs<0||d.toMs<=d.fromMs||d.toMs>4102444800000)return null;const fromMs=d.selection?.fromMs??d.fromMs,toMs=d.selection?.toMs??d.toMs;if(![fromMs,toMs].every(Number.isSafeInteger)||fromMs<d.fromMs||toMs>d.toMs||fromMs>=toMs)return null;return {fromMs:d.fromMs,toMs:d.toMs,selection:{fromMs,toMs}}}
 if(global.customElements&&global.HTMLElement){class GooTimeRange extends global.HTMLElement{
  static get observedAttributes(){return ['data']}
  set data(v){try{this._data=structuredClone(v)}catch{this._data=null}if(this.isConnected)this.render()}
  get data(){if(this._data!==undefined)return structuredClone(this._data);try{return JSON.parse(this.getAttribute('data'))}catch{return null}}
  connectedCallback(){if(Object.hasOwn(this,'data')){const d=this.data;delete this.data;this.data=d}this.render()}
  attributeChangedCallback(){this._data=undefined;if(this.isConnected)this.render()}
  render(){const d=model(this.data);this.replaceChildren();this.dataset.state=d?'observed':'unavailable';if(!d){this.textContent='Time range unavailable.';return}for(const [key,label]of [['fromMs','Range start (UTC)'],['toMs','Range end (UTC, exclusive)']]){const wrap=document.createElement('label'),input=document.createElement('input'),output=document.createElement('output');wrap.textContent=label;input.type='range';input.step='1';input.min=String(d.fromMs);input.max=String(d.toMs);input.value=String(d.selection[key]);input.setAttribute('aria-label',label);output.textContent=new Date(d.selection[key]).toISOString();input.oninput=()=>{const value=Number(input.value),other=key==='fromMs'?d.selection.toMs:d.selection.fromMs;if(key==='fromMs'?value>=other:value<=other){input.value=String(d.selection[key]);return}d.selection[key]=value;this._data=structuredClone(d);output.textContent=new Date(value).toISOString();input.setAttribute('aria-valuetext',output.textContent);this.dispatchEvent(new CustomEvent('goo-range-select',{bubbles:true,detail:structuredClone(d.selection)}))};input.setAttribute('aria-valuetext',output.textContent);wrap.append(input,output);this.append(wrap)}}
 }if(!global.customElements.get('goo-time-range'))global.customElements.define('goo-time-range',GooTimeRange)}
 const api={model};if(typeof module!=='undefined')module.exports=api;else global.GooTimeRange=api;
})(typeof window==='undefined'?globalThis:window);
