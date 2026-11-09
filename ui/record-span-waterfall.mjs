/* Passive record-page timing. Data and mounting handles grant no source authority. */
import * as UI from './components.mjs';
import * as Pages from './record-page.mjs';
import * as Records from './record-presentation.mjs';
import * as Columns from './record-columns.mjs';
import {displayOptions,contextRecords} from './record-list.mjs';
import {recordDisplay} from './view-composition.mjs';
import {layout,rowElement,durationLabel} from './waterfall.mjs';

const defaults=Object.freeze({columns:Object.freeze(['utc','service','summary','context']),fieldColumns:Object.freeze([])});
const bounded=(value,max)=>typeof value==='string'?value.length<=max?value:value.slice(0,max)+'…':'';
// Geometry never copies a rich page per observation. Invalid IDs remain invalid,
// while display/Details retain the original immutable supplied record.
export function timingRecord(record){
 const out={};for(const [key,max]of [['timestamp_ns',20],['duration_ns',20],['TraceId',32],['SpanId',16],['ParentSpanId',16],['ServiceName',512],['SpanName',1024],['StatusCode',16]]){
  const value=record[key];if(value===undefined||value===null){if(key==='ParentSpanId')out[key]=value;continue}
  if(typeof value==='string')out[key]=bounded(value,max);else out[key]='invalid';
 }return out;
}
export function waterfallModel(page,display=defaults){
 page=Pages.adoptPage(page);display=displayOptions(display);
 recordDisplay({version:'goo.view/1',scope:{kind:'record-query',signal:page.signal},kind:{name:'record-span-waterfall',version:1},display},page.question);
 const geometry=layout(page.rows.map(timingRecord),{origin:'trace-relative'});
 return {page,display,geometry,rows:page.rows.map((record,index)=>({index,record,description:Records.describe(record,'spans'),fields:display.fieldColumns.map(field=>({field,...Columns.presentation(record,field)}))}))};
}
const Base=globalThis.HTMLElement||class{};
export class RecordSpanWaterfall extends Base{
 static get observedAttributes(){return ['page','display','summary-provided']}
 set summaryProvided(value){if(value===true)this.setAttribute('summary-provided','');else this.removeAttribute('summary-provided')}
 get summaryProvided(){return this.getAttribute('summary-provided')!==null}
 set page(value){try{this._page=value==null?null:Pages.adoptPage(value);this._pageError=''}catch(error){this._page=null;this._pageError=error.message}if(this.isConnected)this.render()}
 get page(){if(this._page!==undefined)return this._page;try{const text=this.getAttribute('page');if(!text)return this._page=null;if(text.length>Pages.RECORD_PAGE_LIMITS.maxBytes)throw Error('Record page exceeds the supplied-data budget.');return this._page=Pages.adoptPage(JSON.parse(text))}catch(error){this._pageError=error.message;return this._page=null}}
 set display(value){try{this._display=displayOptions(value);this._displayError=''}catch(error){this._display=null;this._displayError=error.message}if(this.isConnected)this.render()}
 get display(){if(this._display!==undefined)return this._display;try{const text=this.getAttribute('display');if(text&&text.length>8192)throw Error('Record display exceeds its options budget.');return this._display=displayOptions(text?JSON.parse(text):defaults)}catch(error){this._displayError=error.message;return this._display=null}}
 connectedCallback(){for(const key of ['page','display','summaryProvided'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}if(this.page!==this._renderedPage||this.display!==this._renderedDisplay||this.summaryProvided!==this._renderedSummaryProvided)this.render()}
 attributeChangedCallback(name){if(name==='page'){this._page=undefined;this._pageError=''}else if(name==='display'){this._display=undefined;this._displayError=''}if(this.isConnected)this.render()}
 recordItems(){if(this.page!==this._renderedPage||this.display!==this._renderedDisplay||this.summaryProvided!==this._renderedSummaryProvided)this.render();return this._recordItems||Object.freeze([])}
 render(){
  this._recordItems=Object.freeze([]);this._renderedPage=null;this._renderedDisplay=null;this._renderedSummaryProvided=null;this.replaceChildren();delete this.dataset.signal;this.dataset.state='unavailable';
  const n=UI.node,page=this.page,display=this.display;
  if(!page||!display){this.append(n('p',this._pageError||this._displayError||'Supply a bounded span page.','record-span-waterfall-unavailable'));return}
  let model;try{model=waterfallModel(page,display)}catch(error){this.append(n('p','Span timing unavailable · '+error.message,'record-span-waterfall-unavailable'));return}
  this._renderedPage=page;this._renderedDisplay=display;this._renderedSummaryProvided=this.summaryProvided;this.dataset.signal='spans';this.dataset.state=page.quality.state;
  const continuation=page.page.hasMore===null?'Continuation unknown':page.page.hasMore?'More records reported':'No further page reported';
  if(!this.summaryProvided)this.append(n('p',page.page.returned+' supplied spans · offset '+page.page.offset+' · '+continuation,'record-span-waterfall-facts'));
  const scale=model.geometry.scaleExtent,interpretation='Each trace starts at its earliest timed supplied span. Relative origins remove gaps between traces; one common duration scale preserves comparison. Supplied extent is not complete trace latency. Async links are not parent edges.',notice=n('p',(scale===null?'Timing unavailable':((this.summaryProvided?'Relative trace time: 0–':'Shared scale: 0–')+durationLabel(scale)))+(this.summaryProvided?' · coverage unknown':' · each trace starts at its earliest timed supplied span · coverage unknown'),'record-span-waterfall-scale');notice.title='Relative origins remove gaps between traces; one common duration scale preserves comparison. Supplied extent is not complete trace latency. Async links are not parent edges.';this.append(notice);
  if(page.quality.state==='changed_between_reads')this.append(n('p','Source changed between independent reads; supplied rows do not establish one coherent population.','record-span-waterfall-warning'));
  if(!page.rows.length){this.append(n('p','No supplied spans on this page.','record-span-waterfall-empty'));return}
  const ancestryRows=contextRecords(page.rows),byIndex=new Map(model.geometry.spans.map(span=>[span.index,span])),hosts=new Array(page.rows.length);
  for(const group of model.geometry.groups){
   const section=n('section',undefined,'record-span-waterfall-group'),heading=n('div',undefined,'record-span-waterfall-heading'),identity=n('strong',group.traceId?'Trace '+group.traceId.slice(0,8)+'…'+group.traceId.slice(-8):'Unreported trace · observation '+(page.page.offset+group.rowIndexes[0]+1));
   if(group.traceId){identity.title=group.traceId;identity.setAttribute('aria-label','Trace '+group.traceId)}
   heading.append(identity);if(group.origin!==null&&group.rowIndexes.length>1){const time=Records.timestamp({timestamp_ns:String(group.origin)},{compact:this.summaryProvided});time.title='Earliest timed supplied span · '+time.title;heading.append(time)}else if(group.origin===null)heading.append(n('span','Supplied origin unavailable'));
   section.append(heading);section.setAttribute('aria-label',group.traceId?'Supplied trace fragment '+group.traceId:'Unreported trace observation '+(page.page.offset+group.rowIndexes[0]+1));
   const ordered=model.geometry.spans.filter(span=>span.groupKey===group.key);
   for(const span of ordered){
    const {index}=span,{record,description,fields}=model.rows[index],item=n('div',undefined,'record-span-waterfall-item'),timeHost=n('div',undefined,'record-span-waterfall-meta');item.dataset.recordIndex=String(index);timeHost.append(Records.timestamp(record));item.append(timeHost);
    const label=bounded((display.columns.includes('service')?description.qualifiedService+' · ':'')+description.message,1536),timingButton=rowElement(byIndex.get(index),{label});item.append(timingButton);
    if(fields.length){const values=n('dl',undefined,'record-span-waterfall-fields');for(const value of fields){const pair=n('div'),label=value.field.kind+':'+value.field.key,text=n('dd',bounded(value.text,512));text.dataset.valueState=value.state;text.dataset.valueType=value.type;text.setAttribute('aria-label',label+' · '+value.type+' value · '+bounded(value.text,512));text.title=value.title|| (value.text.length>512?'Inspect for the exact supplied value.':value.text);pair.append(n('dt',label),text);values.append(pair)}item.append(values)}
    if(display.columns.includes('context')){const context=n('goo-record-context');context.setAttribute('signal','spans');context.record=ancestryRows[index];context.records=ancestryRows;item.append(context)}
    const details=n('details',undefined,'record-span-waterfall-details');details.append(n('summary','Details'));let populated=false;
    details.ontoggle=()=>{if(!details.open||populated||this.page!==page||this.display!==display||!this.isConnected||!details.isConnected)return;populated=true;details.append(Records.detail(record,'spans',{},ancestryRows))};
    timingButton.onclick=()=>{if(this.page===page&&this.display===display&&this.isConnected&&timingButton.isConnected&&this.recordItems()[index]?.timingButton===timingButton)details.open=true};
    const actionsHost=n('div',undefined,'record-span-waterfall-actions'),detailHost=n('div',undefined,'record-span-waterfall-inspector');actionsHost.hidden=true;detailHost.hidden=true;item.append(actionsHost,details,detailHost);section.append(item);
    hosts[index]=Object.freeze({index,element:item,actionsHost,detailHost,details,timingButton,timeHost});
   }this.append(section);
  }
  if(this.summaryProvided){const details=n('details',undefined,'record-span-waterfall-details');details.append(n('summary','Timing scope'),n('p',interpretation));this.append(details)}
  this._recordItems=Object.freeze(hosts);
 }
}
if(globalThis.customElements&&!globalThis.customElements.get('goo-record-span-waterfall'))globalThis.customElements.define('goo-record-span-waterfall',RecordSpanWaterfall);
