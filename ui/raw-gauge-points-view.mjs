/* Passive supplied-data gauge scatter. Acquisition and inspection belong to callers. */
import './chart-frame.mjs';
import './components.mjs';
import {model} from './raw-gauge-points.mjs';
import {compactScalar} from './numeric-values.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
export {model};
const n=(tag,text,cls)=>{const e=document.createElement(tag);if(text!==undefined)e.textContent=text;if(cls)e.className=cls;return e};
const raw=value=>value===null||value===undefined?'Not reported':typeof value==='string'?value:String(value);
function utc(value){try{const ns=BigInt(value),iso=new Date(Number(ns/1000000n)).toISOString();return {short:iso.slice(11,23)+' UTC',exact:iso.replace(/\.\d{3}Z$/,'.'+String(ns%1000000000n).padStart(9,'0')+'Z')}}catch{return {short:'Unavailable time',exact:'Unavailable time'}}}
const coordinate=value=>compactScalar(value)?.text??String(value);
if(globalThis.customElements&&globalThis.HTMLElement){
 class GooRawGaugePoints extends globalThis.HTMLElement{
  static get observedAttributes(){return ['data','selected-key','mode','heading']}
  set data(value){try{this._data=structuredClone(value)}catch{this._data=null}if(this.isConnected)this.render()}
  get data(){if(this._data!==undefined)return structuredClone(this._data);try{const text=this.getAttribute('data');return text&&text.length<=2097152?JSON.parse(text):null}catch{return null}}
  set selectedKey(value){if(value===null||value===undefined)this.removeAttribute('selected-key');else this.setAttribute('selected-key',String(value))}
  get selectedKey(){return this.getAttribute('selected-key')}
  connectedCallback(){for(const key of ['data','selectedKey'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}if(!this.hasAttribute('role'))this.setAttribute('role','group');if(!this.hasAttribute('aria-label'))this.setAttribute('aria-label','Raw gauge points');this.render()}
  attributeChangedCallback(name){if(name==='data')this._data=undefined;if(this.isConnected){if(name==='selected-key')this.syncSelection();else this.render()}}
  syncSelection(){for(const button of this.querySelectorAll('[data-point-key]'))button.setAttribute('aria-pressed',String(this._groups?.get(button.dataset.pointKey)?.occurrenceKeys.includes(this.selectedKey)||false));for(const button of this.querySelectorAll('[data-occurrence-key]'))button.setAttribute('aria-pressed',String(button.dataset.occurrenceKey===this.selectedKey))}
  render(){
   const revision=this._revision=(this._revision||0)+1,data=this.data,m=model(data),mode=this.getAttribute('mode')||'points';this.replaceChildren();this.dataset.state=m.state;this._groups=new Map();
   if(m.state==='unavailable'||!['points','table'].includes(mode)){this.append(n('p','Gauge points unavailable.'));return}
   const frame=n('goo-chart-frame'),summary=n('p',undefined,'raw-gauge-points-summary'),plotArea=n('div',undefined,'raw-gauge-points-area'),scale=n('div',undefined,'raw-gauge-points-scale'),plot=n('div',undefined,'raw-gauge-points-plot'),axis=n('div',undefined,'raw-gauge-points-axis'),chooser=n('section',undefined,'raw-gauge-points-chooser');chooser.hidden=true;
   frame.data={title:this.getAttribute('heading')||'Gauge observations',unit:m.unit||'Unit not reported',state:m.state==='no_data'?'no_data':'unknown',note:mode==='table'?'':'Approximate positions may overlap. Observation values lists every row.'};
   for(const key of ['loaded','plotted','unplotted'])this.dataset[key]=String(m.counts[key]);
   summary.textContent=mode==='table'?m.counts.loaded+' held observations':m.counts.plotted+' plotted · '+m.counts.unplotted+' unplotted · '+m.points.length+' locations'+(m.counts.undisplayed?' · 1,000-location display limit':'')+(m.atLimit?' · read budget '+m.counts.loaded:'');summary.setAttribute('role','status');
   const min=m.domain?.min,max=m.domain?.max;scale.append(n('span',max===null||max===undefined?'':coordinate(max)),n('span',min===null||min===undefined?'':coordinate(min)));scale.title='Approximate value coordinates; inspect an observation for its raw value.';
   plot.setAttribute('role','group');plot.setAttribute('aria-label','Gauge value by observation time; arrow keys move between plotted locations');
   const indexes=new Map((data.rows||[]).map((row,index)=>[row.key,index+1])),allowed=button=>revision===this._revision&&this.isConnected&&this.contains(button)&&nativeControlReady(button);
   const emit=(keys,button)=>{if(allowed(button))button.dispatchEvent(new CustomEvent('goo-raw-gauge-point',{bubbles:true,detail:{occurrenceKeys:[...keys]}}))};
   const controls=[];
   const memberButton=member=>{const number=indexes.get(member.key)||'? ',button=n('button','Observation '+number+' · '+raw(member.rawValue));button.type='button';button.dataset.occurrenceKey=member.key;button.title=utc(member.endNS).exact+' · Raw value: '+raw(member.rawValue)+(m.unit?' '+m.unit:'');button.setAttribute('aria-label','Inspect observation '+number+' · '+raw(member.rawValue));button.onclick=()=>emit([member.key],button);return button};
   function members(point){
    chooser.hidden=false;let offset=0;const heading=n('h4',point.multiplicity+' observations at this position'),hint=n('p','Same plotted coordinates; raw values may differ.'),list=n('div',undefined,'raw-gauge-points-members'),paging=n('div',undefined,'raw-gauge-points-pages'),previous=n('button','Previous observations'),next=n('button','More observations'),range=n('span');previous.type=next.type='button';paging.append(previous,next,range);chooser.replaceChildren(heading,hint,list,paging);
    const draw=()=>{list.replaceChildren(...point.members.slice(offset,offset+100).map(memberButton));previous.disabled=offset===0;next.disabled=offset+100>=point.members.length;paging.hidden=point.members.length<=100;range.textContent=(offset+1)+'–'+Math.min(offset+100,point.members.length)+' of '+point.members.length;};
    previous.onclick=()=>{if(!allowed(previous))return;offset=Math.max(0,offset-100);draw();list.querySelector('button')?.focus()};next.onclick=()=>{if(!allowed(next))return;offset+=100;draw();list.querySelector('button')?.focus()};draw();list.querySelector('button')?.focus();
   }
   for(const [index,point]of (mode==='points'?m.points:[]).entries()){
    this._groups.set(point.key,point);const button=n('button',point.multiplicity>1?String(point.multiplicity):'', 'raw-gauge-points-mark');button.type='button';button.dataset.pointKey=point.key;button.dataset.occurrences=String(point.multiplicity);button.style.left=(point.x*100)+'%';button.style.top=((1-point.y)*100)+'%';button.tabIndex=index===0?0:-1;
    const label=point.multiplicity>1?point.multiplicity+' observations at '+utc(point.endNS).exact+'; choose an observation':'Observation '+indexes.get(point.occurrenceKeys[0])+' · '+utc(point.endNS).exact+' · raw value '+raw(point.rawValues[0])+(m.unit?' '+m.unit:'');button.title=label;button.setAttribute('aria-label',label);button.setAttribute('aria-pressed',String(point.occurrenceKeys.includes(this.selectedKey)));
    button.onclick=()=>{if(!allowed(button))return;for(const [i,control]of controls.entries())control.tabIndex=i===index?0:-1;if(point.multiplicity>1){members(point);emit(point.occurrenceKeys,button)}else{chooser.hidden=true;emit(point.occurrenceKeys,button)}};
    button.onkeydown=event=>{if(!allowed(button))return;let next;if(['ArrowRight','ArrowDown'].includes(event.key))next=Math.min(controls.length-1,index+1);else if(['ArrowLeft','ArrowUp'].includes(event.key))next=Math.max(0,index-1);else if(event.key==='Home')next=0;else if(event.key==='End')next=controls.length-1;else return;event.preventDefault();for(const [i,control]of controls.entries())control.tabIndex=i===next?0:-1;controls[next].focus()};controls.push(button);plot.append(button);
   }
   if(!m.points.length)plot.append(n('p',m.counts.loaded?'No eligible recorded values to plot.':'No held observations.','raw-gauge-points-empty'));
   for(const bound of [m.fromNS,m.toNS]){const time=utc(bound),element=n('time',time.short);element.dateTime=time.exact;element.title=time.exact;axis.append(element)}
   plotArea.append(scale,plot);if(mode==='points')frame.append(summary,plotArea,axis,chooser);else frame.append(summary);
   const more=n('details',undefined,'raw-gauge-points-excluded'),excludedSummary=n('summary',m.counts.unplotted+' unplotted observations');
   const reasons=[];for(const [key,label]of [['unknownRecording','Recording unknown'],['noRecordedValue','Not recorded'],['invalidValue','Value unavailable'],['undisplayed','Plot display limit']])if(m.counts[key])reasons.push(label+': '+m.counts[key]);more.append(excludedSummary,n('p',reasons.join(' · ')));more.hidden=m.counts.unplotted===0||mode==='table';if(mode==='points')frame.append(more);
   const values=n('details',undefined,'raw-gauge-points-values'),valuesSummary=n('summary','Observation values'),valueList=n('div',undefined,'raw-gauge-points-members'),valuePaging=n('div',undefined,'raw-gauge-points-pages'),before=n('button','Previous values'),after=n('button','More values'),position=n('span');before.type=after.type='button';valuePaging.append(before,after,position);values.append(valuesSummary,valueList,valuePaging);let valueOffset=0;
   const reasonLabels={'recording-unknown':'Recording unknown','not-recorded':'Not recorded','invalid-value':'Value unavailable','visual-budget':'Plot display limit'},excludedByKey=new Map((m.excluded||[]).map(item=>[item.key,reasonLabels[item.reason]||'Unavailable']));
   const drawValues=()=>{
    const viewport=n('goo-table'),table=n('table'),head=n('thead'),header=n('tr'),body=n('tbody'),labels=['Observed UTC','Raw value'+(m.unit?' · '+m.unit:''),'Recording / value','Inspect'];viewport.setAttribute('density','compact');viewport.setAttribute('mobile-layout','stacked');table.append(n('caption','Held raw gauge observations · source occurrences retained'));
    for(const label of labels){const th=n('th',label);th.scope='col';header.append(th)}head.append(header);table.append(head,body);
    for(const row of (data.rows||[]).slice(valueOffset,valueOffset+100)){
     const observation=row.observation,entry=n('tr'),bound=utc(observation.endNS),time=n('time',bound.short),cell=n('td'),button=memberButton({key:row.key,endNS:observation.endNS,rawValue:observation.rawValue});time.dateTime=bound.exact;time.title=bound.exact;cell.append(time);button.textContent='Inspect';const action=n('td');action.append(button);entry.append(cell,n('td',raw(observation.rawValue)),n('td',excludedByKey.get(row.key)||'Recorded'),action);[...entry.children].forEach((td,index)=>td.dataset.label=labels[index]);body.append(entry);
    }
    viewport.append(table);valueList.replaceChildren(viewport);before.disabled=valueOffset===0;after.disabled=valueOffset+100>=data.rows.length;valuePaging.hidden=data.rows.length<=100;position.textContent=(data.rows.length?valueOffset+1:0)+'–'+Math.min(valueOffset+100,data.rows.length)+' of '+data.rows.length;this.syncSelection();
   };
   values.ontoggle=()=>{if(values.open&&revision===this._revision&&this.isConnected)drawValues()};before.onclick=()=>{if(!allowed(before))return;valueOffset=Math.max(0,valueOffset-100);drawValues();valueList.querySelector('button')?.focus()};after.onclick=()=>{if(!allowed(after))return;valueOffset+=100;drawValues();valueList.querySelector('button')?.focus()};if(mode==='table'){values.open=true;valuesSummary.hidden=true}frame.append(values);this.append(frame);if(mode==='table')drawValues();this.syncSelection();
  }
 }
 if(!globalThis.customElements.get('goo-raw-gauge-points'))globalThis.customElements.define('goo-raw-gauge-points',GooRawGaugePoints);
}
