/* Passive bounded record display. The application owns acquisition and actions. */
import * as UI from './components.mjs';
import * as Pages from './record-page.mjs';
import * as Records from './record-presentation.mjs';
import * as Columns from './record-columns.mjs';
import {recordDisplay} from './view-composition.mjs';

const defaults=Object.freeze({columns:Object.freeze(['utc','service','summary','context']),fieldColumns:Object.freeze([])});
const object=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
function ownData(value,keys){
 if(!object(value)||![Object.prototype,null].includes(Object.getPrototypeOf(value))||Object.getOwnPropertySymbols(value).length)throw Error('Use plain record display options.');
 const descriptors=Object.getOwnPropertyDescriptors(value);
 for(const [key,d]of Object.entries(descriptors))if(!keys.includes(key)||!d.enumerable||!Object.hasOwn(d,'value'))throw Error('Use plain record display options.');
 return Object.fromEntries(Object.entries(descriptors).map(([key,d])=>[key,d.value]));
}
function arrayData(value,max){
 if(!Array.isArray(value)||value.length>max||Object.getOwnPropertySymbols(value).length)throw Error('Use bounded record display lists.');
 const ds=Object.getOwnPropertyDescriptors(value);if(Object.keys(ds).length!==value.length+1)throw Error('Use dense record display lists.');
 return Array.from({length:value.length},(_,i)=>{const d=ds[i];if(!d||!d.enumerable||!Object.hasOwn(d,'value'))throw Error('Use plain record display lists.');return d.value});
}
export function displayOptions(value=defaults){
 const raw=ownData(value,['columns','fieldColumns']),columns=arrayData(raw.columns,4),fields=arrayData(raw.fieldColumns===undefined?[]:raw.fieldColumns,8).map(field=>ownData(field,['kind','key']));
 const checked=Columns.validate({columns,fieldColumns:fields});
 return Object.freeze({columns:Object.freeze(checked.columns),fieldColumns:Object.freeze(checked.fieldColumns.map(field=>Object.freeze(field)))});
}
export function fieldValue(record,field){return Columns.presentation(record,field);}
// Context consumes reported identifiers only. Bound malformed public values too:
// compact metadata must not clone rich page payloads once per displayed record.
export function contextRecord(record){
 const out={};for(const [key,max]of [['TraceId',32],['SpanId',16],['ParentSpanId',16]])if(Object.hasOwn(record,key)){const value=record[key];out[key]=value==null?null:typeof value==='string'&&value.length<=max?value:'invalid'}return out;
}
export const contextRecords=rows=>rows.map(contextRecord);
export function listModel(page,display=defaults){
 page=Pages.adoptPage(page);display=displayOptions(display);
 recordDisplay({version:'goo.view/1',scope:{kind:'record-query',signal:page.signal},kind:{name:'record-list',version:1},display},page.question);
 return {page,display,rows:page.rows.map((record,index)=>({index,record,description:Records.describe(record,page.signal),fields:display.fieldColumns.map(field=>({field,...fieldValue(record,field)}))}))};
}
const Base=globalThis.HTMLElement||class{};
export class RecordList extends Base{
 static get observedAttributes(){return ['page','display']}
 set page(value){try{this._page=value==null?null:Pages.adoptPage(value);this._pageError=''}catch(error){this._page=null;this._pageError=error.message}if(this.isConnected)this.render()}
 get page(){if(this._page!==undefined)return this._page;try{const text=this.getAttribute('page');if(!text)return this._page=null;if(text.length>Pages.RECORD_PAGE_LIMITS.maxBytes)throw Error('Record page exceeds the supplied-data budget.');return this._page=Pages.adoptPage(JSON.parse(text))}catch(error){this._pageError=error.message;return this._page=null}}
 set display(value){try{this._display=displayOptions(value);this._displayError=''}catch(error){this._display=null;this._displayError=error.message}if(this.isConnected)this.render()}
 get display(){if(this._display!==undefined)return this._display;try{const text=this.getAttribute('display');if(text&&text.length>8192)throw Error('Record display exceeds its options budget.');return this._display=displayOptions(text?JSON.parse(text):defaults)}catch(error){this._displayError=error.message;return this._display=null}}
 connectedCallback(){for(const key of ['page','display'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
 attributeChangedCallback(name){if(name==='page'){this._page=undefined;this._pageError=''}else{this._display=undefined;this._displayError=''}if(this.isConnected)this.render()}
 render(){
  const n=UI.node,page=this.page,display=this.display;this.replaceChildren();delete this.dataset.signal;this.dataset.state='unavailable';
  if(!page||!display){const status=n('p',this._pageError||this._displayError||'Supply a bounded record page.','record-list-unavailable');status.setAttribute('role','status');this.append(status);return}
  let model;try{model=listModel(page,display)}catch(error){const status=n('p','Record list unavailable · '+error.message,'record-list-unavailable');status.setAttribute('role','status');this.append(status);return}
  this.dataset.signal=page.signal;this.dataset.state=page.quality.state;
  const continuation=page.page.hasMore===null?'Continuation unknown':page.page.hasMore?'More records reported':'No further page reported',facts=n('p',page.page.returned+' supplied records · offset '+page.page.offset+' · '+continuation,'record-list-facts');facts.title='Bounded supplied page; not a population total. Independent source reads; upstream coverage unknown.';
  this.append(facts);if(page.quality.state==='changed_between_reads')this.append(n('p','Source changed between independent reads; these supplied rows do not establish one coherent population.','record-list-warning'));
  if(!model.rows.length){this.append(n('p','No supplied records on this page.','record-list-empty'));return}
  const ancestryRows=contextRecords(page.rows);
  const list=n('ol');list.setAttribute('aria-label','Supplied '+page.signal+' records');list.start=page.page.offset+1;this.append(list);
  for(const {index,record,fields}of model.rows){
   const item=n('li',undefined,'record-list-item');item.dataset.recordIndex=String(index);const line=n('div',undefined,'record-list-meta');
   if(display.columns.includes('utc'))line.append(Records.timestamp(record));if(display.columns.includes('service'))line.append(Records.service(record));if(line.childNodes.length)item.append(line);
   if(display.columns.includes('summary'))item.append(Records.summary(record,page.signal,{compact:true}));
   if(fields.length){const values=n('dl',undefined,'record-list-fields');for(const value of fields){const pair=n('div'),label=value.field.kind+':'+value.field.key,text=n('dd',value.text);text.dataset.valueState=value.state;text.dataset.valueType=value.type;text.setAttribute('aria-label',label+' · '+value.type+' value · '+value.text);if(value.title)text.title=value.title;pair.append(n('dt',label),text);values.append(pair)}item.append(values)}
   if(display.columns.includes('context')){const context=n('goo-record-context');context.setAttribute('signal',page.signal);context.record=ancestryRows[index];context.records=ancestryRows;item.append(context)}
   const details=n('details',undefined,'record-list-details');details.append(n('summary','Details'));let populated=false;
   details.ontoggle=()=>{if(!details.open||populated||this.page!==page||!this.isConnected||!details.isConnected)return;populated=true;details.append(Records.detail(record,page.signal,{},ancestryRows))};
   item.append(details);list.append(item);
  }
 }
}
if(globalThis.customElements&&!globalThis.customElements.get('goo-record-list'))globalThis.customElements.define('goo-record-list',RecordList);
