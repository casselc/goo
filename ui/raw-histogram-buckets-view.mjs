/* Passive categorical count bars for one raw histogram occurrence. */
import './chart-frame.mjs';
import './components.mjs';
import {model} from './raw-histogram-buckets.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
export {model};
const n=(tag,text,cls)=>{const node=document.createElement(tag);if(text!==undefined)node.textContent=text;if(cls)node.className=cls;return node};
if(globalThis.customElements&&globalThis.HTMLElement){
 class GooRawHistogramBuckets extends globalThis.HTMLElement{
  static get observedAttributes(){return ['data','mode','heading','heading-level']}
  set data(value){try{this._data=structuredClone(value)}catch{this._data=null}this._model=undefined;this._offset=0;if(this.isConnected)this.render()}
  get data(){if(this._data!==undefined)return structuredClone(this._data);try{const text=this.getAttribute('data');return text&&text.length<=2097152?JSON.parse(text):null}catch{return null}}
  connectedCallback(){if(Object.hasOwn(this,'data')){const value=this.data;delete this.data;this.data=value}if(!this.hasAttribute('role'))this.setAttribute('role','group');if(!this.hasAttribute('aria-label'))this.setAttribute('aria-label','Raw histogram buckets');this.render()}
  attributeChangedCallback(name){if(name==='data'){this._data=undefined;this._model=undefined;this._offset=0}if(this.isConnected)this.render()}
  render(){
   const revision=this._revision=(this._revision||0)+1,m=this._model??=(model(this.data)),mode=this.getAttribute('mode')||'bars';this.replaceChildren();this.dataset.state=m.state;this.dataset.buckets=String(m.rows.length);this.dataset.eligible=String(m.eligible);
   if(!['bars','table'].includes(mode)){this.append(n('p','Histogram display unavailable.'));return}
   const frame=n('goo-chart-frame');frame.setAttribute('density','compact');frame.data={headingLevel:this.getAttribute('heading-level')||'3',title:this.getAttribute('heading')||'Histogram buckets',unit:m.unit||'Unit not reported',state:m.state==='observed'?'unknown':m.state==='no_data'?'no_data':m.state,note:mode==='bars'&&m.eligible?'Count bars · interval widths are not to scale':''};this.append(frame);
   if(m.reason)frame.append(n('p',m.reason,'raw-histogram-buckets-note'));
   if(!m.rows.length||mode==='bars'&&!m.eligible)return;
   const summary=n('p',m.rows.length+' buckets · '+m.count+' source count'+(m.count==='0'?' · recorded empty point':''),'raw-histogram-buckets-summary');frame.append(summary);
   const viewport=n('goo-table'),table=n('table'),head=n('thead'),header=n('tr'),body=n('tbody'),pages=n('div',undefined,'raw-histogram-buckets-pages'),previous=n('button','Previous buckets'),next=n('button','More buckets'),range=n('span');previous.type=next.type='button';range.setAttribute('role','status');viewport.setAttribute('density','compact');table.setAttribute('aria-label','Source histogram buckets');
   const labels=['Interval','Source count'];for(const label of labels){const th=n('th',label);th.scope='col';header.append(th)}head.append(header);table.append(n('caption','One source point · exact bucket counts','visually-hidden'),head,body);viewport.append(table);pages.append(previous,next,range);frame.append(pages,viewport);let offset=Math.min(this._offset||0,Math.max(0,Math.floor((m.rows.length-1)/100)*100));
   const allowed=button=>revision===this._revision&&this.isConnected&&this.contains(button)&&nativeControlReady(button);
   const draw=()=>{
    this._offset=offset;body.replaceChildren();for(const value of m.rows.slice(offset,offset+100)){
     const entry=n('tr'),interval=n('th',value.interval),amount=n('td'),count=n('span',value.count,'raw-histogram-buckets-count');entry.dataset.bucketIndex=String(value.index);interval.scope='row';interval.dataset.label=labels[0];interval.title=value.rawInterval===value.interval?value.interval:'Source boundaries: '+value.rawInterval;amount.dataset.label=labels[1];
     if(mode==='bars'){const track=n('span',undefined,'raw-histogram-buckets-track'),bar=n('span',undefined,'raw-histogram-buckets-bar');track.setAttribute('aria-hidden','true');bar.style.width=(value.fraction*100)+'%';bar.dataset.nonzero=String(value.count!=='0');track.append(bar);amount.className='raw-histogram-buckets-amount';amount.append(track)}amount.append(count);entry.append(interval,amount);body.append(entry);
    }
    previous.disabled=offset===0;next.disabled=offset+100>=m.rows.length;pages.hidden=m.rows.length<=100;range.textContent=(offset+1)+'–'+Math.min(offset+100,m.rows.length)+' of '+m.rows.length+' buckets';
   };
   previous.onclick=()=>{if(!allowed(previous))return;offset=Math.max(0,offset-100);draw();(previous.disabled?next:previous).focus()};next.onclick=()=>{if(!allowed(next))return;offset+=100;draw();(next.disabled?previous:next).focus()};draw();
  }
 }
 if(!globalThis.customElements.get('goo-raw-histogram-buckets'))globalThis.customElements.define('goo-raw-histogram-buckets',GooRawHistogramBuckets);
}
