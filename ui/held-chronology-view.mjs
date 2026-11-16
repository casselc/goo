/* Passive local ordering of held observations. The application owns all actions. */
import * as UI from './components.mjs';
import * as Records from './record-presentation.mjs';
import * as Conditions from './condition-summary.mjs';
import * as Environment from './record-environment.mjs';
import {adopt,model,limits} from './held-chronology.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
const n=UI.node,signals=['logs','spans'],label=signal=>signal==='logs'?'Logs':'Spans';
function clock(value){const time=n('time',value.replace('T',' ').replace('Z',' UTC'));time.dateTime=value;time.title=value;time.tabIndex=0;return time}
function sourceFacts(signal,lane){
 const details=n('details',undefined,'held-chronology-source'),summary=n('summary'),title=n('strong',label(signal));
 const phase=({paused:'Paused',unavailable:'Unavailable',stopped:'Stopped',loading:'Pending',unknown:'Held',partial:'Partial',stale:'Stale'})[lane.phase],history=lane.rows.length>0||lane.readWindow!==null;
 const state=lane.selector===null?'Not configured':lane.completedRead?phase:history?phase+' · latest read unavailable':lane.phase==='unavailable'?'Unavailable · not read':lane.staged?'Not started':lane.phase==='loading'?'Pending':'Not read';
 details.dataset.signal=signal;summary.append(title,' · '+lane.rows.length+' held · '+state+(lane.gap?' · continuity gap':'')+(lane.evicted?' · '+lane.evicted+' evicted':''));details.append(summary);
 const facts=n('dl'),fact=(name,value)=>{const pair=n('div'),term=n('dt',name),definition=n('dd');typeof value==='string'?definition.textContent=value:definition.append(value);pair.append(term,definition);facts.append(pair)};
 if(lane.completedRead){fact('Latest read UTC',clock(lane.completedRead.observedAt));fact('Latest read window','['+new Date(lane.completedRead.fromMs).toISOString()+', '+new Date(lane.completedRead.toMs).toISOString()+')');fact('Latest delivery',lane.completedRead.rowCount+' rows'+(lane.completedRead.limited?' · limited':''));}
 else fact('Latest read','Unavailable');
 if(lane.selector===null)fact('Scope','Not configured');
 else {fact('Scope',Environment.questionLabel(lane.selector));const condition=n('span');Conditions.present(condition,lane.selector.where,{prefix:'',empty:'All records in scope'});fact('Source condition',condition);}
 if(lane.find.text)fact('Lane Find','Not applied here · '+lane.find.text);
 details.append(facts);return details;
}
if(globalThis.customElements&&globalThis.HTMLElement){
 class HeldChronology extends globalThis.HTMLElement{
  static get observedAttributes(){return ['data','selected-key']}
  set data(value){try{this._data=adopt(value);this._error=''}catch(error){this._data=null;this._error=error.message}this._query='';this._offset=0;if(this.isConnected)this.render()}
  get data(){if(this._data!==undefined)return this._data?structuredClone(this._data):null;try{const text=this.getAttribute('data');if(!text)return null;if(text.length>limits.maxBytes)throw Error('Held observations exceed the supplied-data budget.');this._data=adopt(JSON.parse(text));return structuredClone(this._data)}catch(error){this._error=error.message;return null}}
  set selectedKey(value){if(value===undefined||value===null)this.removeAttribute('selected-key');else this.setAttribute('selected-key',String(value))}
  get selectedKey(){return this.getAttribute('selected-key')}
  connectedCallback(){for(const key of ['data','selectedKey'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}if(!this.hasAttribute('role'))this.setAttribute('role','group');if(!this.hasAttribute('aria-label'))this.setAttribute('aria-label','Held log and span chronology');this.render()}
  attributeChangedCallback(name){if(name==='selected-key'){if(this.isConnected)this.syncSelection();return}this._data=undefined;this._query='';this._offset=0;if(this.isConnected)this.render()}
  syncSelection(){for(const row of this.querySelectorAll('[data-chronology-key]')){if(row.dataset.chronologyKey===this.selectedKey){row.setAttribute('aria-current','true');row.dataset.selected='true'}else{row.removeAttribute('aria-current');delete row.dataset.selected}}}
  render(){
   const revision=this._revision=(this._revision||0)+1;this.replaceChildren();const supplied=this.data;
   let current;try{current=model(supplied,{query:this._query||''})}catch(error){this.dataset.state='unavailable';const message=n('p','Held chronology unavailable · '+(this._error||error.message),'held-chronology-status');message.setAttribute('role','status');this.append(message);return}
   this.dataset.state='held';const data=current.data,sources=n('div',undefined,'held-chronology-sources');for(const signal of signals)sources.append(sourceFacts(signal,data.lanes[signal]));
   const controls=n('div',undefined,'held-chronology-controls'),findLabel=n('label','Find '),find=n('input'),counts=n('span',undefined,'held-chronology-counts'),previous=n('button','Previous records'),next=n('button','More records'),range=n('span');find.type='search';find.maxLength=256;find.value=this._query||'';find.placeholder='Held records';find.setAttribute('aria-label','Find held records');findLabel.append(find);previous.type=next.type='button';counts.setAttribute('role','status');range.setAttribute('role','status');controls.append(findLabel,previous,next,range,counts);
   const records=Records.table({compact:true,context:false,labels:['Observed UTC','Service','Record'],caption:'Held records · newest event time first',label:'Held log and span chronology'}),{viewport,body}=records;
   records.element.querySelector('caption').title='Event-time presentation only; not causality or a shared source order. Ties use lane then held occurrence order. Unavailable timestamps follow.';
   const note=n('p',undefined,'held-chronology-note');note.hidden=!signals.some(signal=>data.lanes[signal].find.text);note.textContent='All held rows · lane Find not applied';
   this.append(sources,controls,note,viewport);
   const allowed=button=>revision===this._revision&&this.isConnected&&this.contains(button)&&nativeControlReady(button);
   const draw=()=>{
    const length=current.rows.length;this._offset=Math.min(this._offset||0,Math.max(0,Math.floor((length-1)/100)*100));body.replaceChildren();let unplaced=false;
    for(const entry of current.rows.slice(this._offset,this._offset+100)){
     if(!entry.placed&&!unplaced){const row=n('tr',undefined,'held-chronology-unplaced'),title=n('td','Timestamp unavailable');title.colSpan=3;row.append(title);body.append(row);unplaced=true}
     const {row,cells}=records.row(entry.item.record,entry.signal),kind=n('span',entry.signal==='logs'?'Log':'Span','held-chronology-kind'),inspect=n('button','Inspect');kind.dataset.signal=entry.signal;row.dataset.chronologyKey=entry.key;row.dataset.signal=entry.signal;row.dataset.sourceKey=entry.item.key;row.dataset.placed=String(entry.placed);cells[2].insertBefore(kind,cells[2].firstChild);inspect.type='button';inspect.dataset.chronologyInspect=entry.key;inspect.title='Observed '+entry.item.observedAt+' · Original window ['+new Date(entry.item.readWindow.fromMs).toISOString()+', '+new Date(entry.item.readWindow.toMs).toISOString()+')'+(entry.item.limited?' · limited delivery':'')+(entry.item.late?' · late event time':'');inspect.setAttribute('aria-label','Inspect '+(entry.signal==='logs'?'log':'span')+' '+(entry.index+1));inspect.onclick=()=>{if(!allowed(inspect))return;inspect.dispatchEvent(new CustomEvent('goo-chronology-inspect',{bubbles:true,composed:true,detail:{signal:entry.signal,key:entry.item.key}}))};cells[2].append(inspect);body.append(row);
    }
    this.dataset.logs=String(current.matched.logs);this.dataset.spans=String(current.matched.spans);this.dataset.unplaced=String(current.unplaced.length);
    counts.textContent='Logs '+current.matched.logs+'/'+current.counts.logs+' · Spans '+current.matched.spans+'/'+current.counts.spans+' held';
    previous.hidden=next.hidden=length<=100;previous.disabled=this._offset===0;next.disabled=this._offset+100>=length;range.hidden=length<=100;range.textContent=length?'Rows '+(this._offset+1)+'–'+Math.min(this._offset+100,length):'';
    note.hidden=!signals.some(signal=>data.lanes[signal].find.text)&&length!==0;note.textContent=length?(signals.some(signal=>data.lanes[signal].find.text)?'All held rows · lane Find not applied':''):(current.query?'No matching held records.':'No held records. Source coverage remains unknown.');this.syncSelection();
   };
   find.oninput=()=>{if(!allowed(find))return;try{current=model(data,{query:find.value});this._query=find.value;this._offset=0;find.removeAttribute('aria-invalid');draw()}catch(error){find.setAttribute('aria-invalid','true');counts.textContent=error.message}};
   previous.onclick=()=>{if(!allowed(previous))return;this._offset=Math.max(0,this._offset-100);draw();(previous.disabled?next:previous).focus()};next.onclick=()=>{if(!allowed(next))return;this._offset+=100;draw();(next.disabled?previous:next).focus()};draw();
  }
 }
 if(!globalThis.customElements.get('goo-held-chronology'))globalThis.customElements.define('goo-held-chronology',HeldChronology);
}
