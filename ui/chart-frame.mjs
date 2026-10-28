/* Passive chart chrome. Callers own plots, actions, source reads and subscriptions. */
const global = globalThis;
 const states={observed:'Observed',no_data:'No data',unknown:'Coverage unknown',loading:'Loading',partial:'Partial result',stale:'Stale result',unavailable:'Unavailable',incompatible:'Incompatible comparison'};
 const text=(value,fallback='')=>typeof value==='string'?value.slice(0,4096):fallback;
 function model(value){const d=value&&typeof value==='object'?value:{};return {title:text(d.title,'Chart'),unit:text(d.unit),window:text(d.window),state:Object.hasOwn(states,d.state)?d.state:'unknown',note:text(d.note),headingLevel:/^[2-6]$/.test(String(d.headingLevel))?String(d.headingLevel):'3'}}
 const fields={title:'heading',unit:'unit',window:'window-label',state:'state',note:'note',headingLevel:'heading-level'};
 if(global.customElements&&global.HTMLElement){
  class GooChartFrame extends global.HTMLElement{
   static get observedAttributes(){return Object.values(fields)}
   set data(value){const d=model(value);this._batch=true;for(const [key,attr] of Object.entries(fields))this.setAttribute(attr,d[key]);this._batch=false;if(this.isConnected)this.render()}
   get data(){return model(Object.fromEntries(Object.entries(fields).map(([key,attr])=>[key,this.getAttribute(attr)])))}
   connectedCallback(){if(Object.hasOwn(this,'data')){const value=this.data;delete this.data;this.data=value}this.render()}
   attributeChangedCallback(){if(this.isConnected&&!this._batch)this.render()}
   render(){
    const d=this.data;
    if(!this.hasAttribute('role'))this.setAttribute('role','figure');
    if(!this.hasAttribute('aria-label')||this.getAttribute('aria-label')===this._label){this.setAttribute('aria-label',d.title);this._label=d.title}
    if(!this._header){const header=document.createElement('div');header.className='chart-frame-header';header.dataset.chartFrameChrome='';this._header=header;this._title=document.createElement('h'+d.headingLevel);this._title.className='chart-frame-title';this._meta=document.createElement('p');this._meta.className='chart-frame-meta';this._state=document.createElement('span');this._state.className='chart-frame-state';this._window=document.createElement('span');this._window.className='chart-frame-window';this._unit=document.createElement('span');this._unit.className='chart-frame-unit';this._note=document.createElement('p');this._note.className='chart-frame-note';this._meta.append(this._state,this._unit,this._window);header.append(this._title,this._meta,this._note)}
    if(this._header.parentNode!==this)this.prepend(this._header);
    if(this._title.tagName!=='H'+d.headingLevel){const heading=document.createElement('h'+d.headingLevel);heading.className='chart-frame-title';this._title.replaceWith(heading);this._title=heading}
    this._title.textContent=d.title;this._state.textContent=states[d.state];this._state.dataset.state=d.state;this._unit.textContent=d.unit;this._unit.hidden=!d.unit;this._window.textContent=d.window;this._window.hidden=!d.window;this._note.textContent=d.note;this._note.hidden=!d.note;
   }
  }
  if(!global.customElements.get('goo-chart-frame'))global.customElements.define('goo-chart-frame',GooChartFrame);
 }
 export {model};
