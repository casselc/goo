/* Passive supplied-span geometry and native-button display. App owners retain all
 * source reads and actions. Subtract exact ns before projecting coordinates. */
import {durationMS} from './components.mjs';
import {unixNs} from './evidence-summary.mjs';
const global=globalThis;
const validID=(value,size)=>typeof value==='string'&&new RegExp('^[a-f0-9]{'+size+'}$').test(value)&&!/^0+$/.test(value);
const ns=value=>typeof value==='string'&&/^(0|[1-9][0-9]{0,19})$/.test(value)&&BigInt(value)<=18446744073709551615n;
const compare=(a,b)=>a.start===null?(b.start===null?a.index-b.index:1):b.start===null?-1:a.start<b.start?-1:a.start>b.start?1:a.index-b.index;
const percent=(value,range)=>Number(value*100000000n/range)/1000000;

export function layout(rows,{origin='global'}={}){
 if(!Array.isArray(rows)||rows.length>1000||!['global','trace-relative'].includes(origin))throw Error('Use bounded spans and a supported supplied origin.');
 const spans=rows.map((row,index)=>{
  const start=ns(row?.timestamp_ns)?BigInt(row.timestamp_ns):null,duration=ns(row?.duration_ns)?BigInt(row.duration_ns):null,traceValid=validID(row?.TraceId,32),spanValid=validID(row?.SpanId,16);
  return {row,index,start,duration,timingValid:start!==null&&duration!==null,traceValid,spanValid,groupKey:traceValid?'trace:'+row.TraceId:'observation:'+index,parentIndex:null,parentState:'parent-unreported',depth:0,left:null,width:null};
 });
 const byID=new Map(),groupMap=new Map();
 for(const span of spans){
  if(!groupMap.has(span.groupKey))groupMap.set(span.groupKey,{key:span.groupKey,traceId:span.traceValid?span.row.TraceId:null,rowIndexes:[],origin:null,extent:null});
  const group=groupMap.get(span.groupKey);group.rowIndexes.push(span.index);
  if(span.traceValid&&span.spanValid){const key=span.row.TraceId+':'+span.row.SpanId;if(!byID.has(key))byID.set(key,[]);byID.get(key).push(span)}
 }
 for(const span of spans){
  if(!span.traceValid){span.parentState='trace-unreported';continue}
  if(!span.spanValid){span.parentState='span-unreported';continue}
  span.duplicateId=byID.get(span.row.TraceId+':'+span.row.SpanId).length>1;
  const parent=span.row.ParentSpanId;
  if(parent===undefined||parent===null)continue;
  if(parent===''||typeof parent==='string'&&/^0{16}$/.test(parent)){span.parentState='no-reported-parent';continue}
  if(!validID(parent,16)){span.parentState='invalid-parent';continue}
  const candidates=byID.get(span.row.TraceId+':'+parent)||[];
  if(candidates.length>1){span.parentState='ambiguous-parent';span.ambiguousParent=true;continue}
  if(!candidates.length){span.parentState='parent-not-held';span.missingParent=true;continue}
  span.parentIndex=candidates[0].index;span.parentState='held-parent';
 }
 for(const span of spans){
  const seen=new Set([span.index]);let current=span;
  while(current.parentIndex!==null){
   if(seen.has(current.parentIndex)){span.cycle=true;break}
   seen.add(current.parentIndex);span.depth++;current=spans[current.parentIndex];
  }
  span.ancestryState=span.cycle?'cycle':current.parentState;
  if(span.cycle)span.depth=0;
 }
 let start=null,end=null;
 for(const group of groupMap.values()){
  let last=null;
  for(const index of group.rowIndexes){const span=spans[index];if(!span.timingValid)continue;const finish=span.start+span.duration;if(group.origin===null||span.start<group.origin)group.origin=span.start;if(last===null||finish>last)last=finish;if(start===null||span.start<start)start=span.start;if(end===null||finish>end)end=finish}
  if(group.origin!==null)group.extent=last-group.origin;
 }
 const groups=[...groupMap.values()],extent=start===null?null:end-start,scaleExtent=origin==='global'?extent:groups.reduce((max,group)=>group.extent!==null&&(max===null||group.extent>max)?group.extent:max,null),range=scaleExtent||1n;
 for(const span of spans)if(span.timingValid){const base=origin==='global'?start:groupMap.get(span.groupKey).origin;span.left=percent(span.start-base,range);span.width=percent(span.duration,range)}
 const ordered=[];
 for(const group of groups){
  const local=group.rowIndexes.map(index=>spans[index]).sort(compare),children=new Map(),visited=new Set();
  for(const span of local)if(span.parentIndex!==null){if(!children.has(span.parentIndex))children.set(span.parentIndex,[]);children.get(span.parentIndex).push(span)}
  for(const list of children.values())list.sort(compare);
  function visit(root){const pending=[root];while(pending.length){const span=pending.pop();if(visited.has(span.index))continue;visited.add(span.index);ordered.push(span);const nested=children.get(span.index)||[];for(let i=nested.length-1;i>=0;i--)pending.push(nested[i])}}
  local.filter(span=>span.parentIndex===null).forEach(visit);local.forEach(visit);
 }
 return {spans:ordered,groups,start:start??0n,end:end??0n,extent,scaleExtent,originMode:origin};
}
export function validData(data){
 return !!data&&Array.isArray(data.rows)&&data.rows.length<=1000&&(data.facts===undefined||Array.isArray(data.facts)&&data.facts.length===data.rows.length&&data.facts.every(f=>Array.isArray(f)&&f.length<=8&&f.every(v=>v&&typeof v.label==='string'&&typeof v.value==='string')))&&(!Object.hasOwn(data,'truncated')||typeof data.truncated==='boolean')&&data.rows.every(r=>r&&ns(r.timestamp_ns)&&ns(r.duration_ns)&&['TraceId','SpanId'].every(k=>typeof r[k]==='string')&&['ParentSpanId','ServiceName','SpanName','StatusCode'].every(k=>r[k]===undefined||typeof r[k]==='string'));
}
export function durationLabel(value){const n=typeof value==='bigint'?value:ns(value)?BigInt(value):null;return n===null||n<0n?'Duration unavailable':n>0n&&n<1000n?'<0.001 ms':durationMS(String(n));}
const node=(tag,text,cls)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n};
export function parentLabel(span){
 const labels={'trace-unreported':'Valid trace ID not reported','span-unreported':'Valid span ID not reported','parent-unreported':'Parent not reported','invalid-parent':'Invalid reported parent','ambiguous-parent':'Ambiguous reported parent','parent-not-held':'Parent outside supplied spans'};
 return [...(span.duplicateId?['Duplicate reported span ID']:[]),...(span.cycle?['Reported parent cycle']:labels[span.parentState]?[labels[span.parentState]]:[])].join(' · ');
}
// One row renderer shared by the inspector and record-page presentation. Host
// adapters may add siblings; interactive controls are never nested in this button.
export function rowElement(span,{facts=[],label}={}){
 const r=span.row,line=node('button',undefined,'waterfall-row');line.type='button';line.dataset.rowIndex=String(span.index);line.dataset.spanId=typeof r.SpanId==='string'?r.SpanId:'';line.dataset.traceId=span.traceValid?r.TraceId:'';line.dataset.parentState=span.parentState;line.dataset.timingState=span.timingValid?'observed':'unavailable';
 const issue=parentLabel(span),title=label||`${typeof r.ServiceName==='string'&&r.ServiceName||'(service unspecified)'} · ${typeof r.SpanName==='string'&&r.SpanName||'(span unnamed)'}`;
 const duration=durationLabel(span.duration),time=span.start===null?'Timestamp unavailable':unixNs(String(span.start))?.exact||'Timestamp unavailable';
 line.setAttribute('aria-label','Inspect supplied span · '+[title,time,duration,issue,...facts.map(f=>f.label+': '+f.value)].filter(Boolean).join(' · '));line.title=time+' · '+duration+(span.duration===null?'':' · '+span.duration+' ns')+(issue?' · '+issue:'');
 const text=node('span',title,'span-label');text.style.paddingLeft=Math.min(span.depth,12)*12+'px';if(issue)text.append(node('span',issue,'span-parent-state'));for(const fact of facts)text.append(node('span',fact.label+': '+fact.value,'span-fact'));
 const track=node('span',undefined,'span-track');track.setAttribute('aria-hidden','true');
 if(span.timingValid){const bar=node('span',undefined,'span-bar');bar.style.left=span.left+'%';bar.style.width=span.width+'%';bar.dataset.error=r.StatusCode==='Error'?'true':'false';bar.dataset.zero=String(span.duration===0n);track.append(bar)}else{track.dataset.state='unavailable';track.append(node('span','Timing unavailable','span-timing-unavailable'))}
 line.append(text,track,node('span',duration+(r.StatusCode==='Error'?' · error':''),'span-duration'));return line;
}
if(global.customElements&&global.HTMLElement){
 class GooTraceWaterfall extends global.HTMLElement{
  static get observedAttributes(){return ['data','selected-index']}
  set data(value){try{this._data=structuredClone(value)}catch{this._data=null}if(this.isConnected)this.render()}
  get data(){if(this._data!==undefined)return structuredClone(this._data);try{const raw=this.getAttribute('data');return raw&&raw.length<=65536?JSON.parse(raw):null}catch{return null}}
  set selectedIndex(value){this.setAttribute('selected-index',value)}get selectedIndex(){const raw=this.getAttribute('selected-index');return raw!==null&&/^(0|[1-9][0-9]*)$/.test(raw)&&Number(raw)<1000?Number(raw):-1}
  connectedCallback(){for(const key of ['data','selectedIndex'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
  attributeChangedCallback(name){if(name==='data')this._data=undefined;if(this.isConnected){if(name==='selected-index')this.updateSelection();else this.render()}}
  updateSelection(){for(const row of this.querySelectorAll('.waterfall-row'))row.setAttribute('aria-pressed',String(Number(row.dataset.rowIndex)===this.selectedIndex))}
  render(){const input=this.data;this.replaceChildren();if(!validData(input)){this.dataset.state='unavailable';this.append(node('p','Waterfall unavailable: unsupported span data.','waterfall-status'));return}
   const data=layout(input.rows);this.dataset.state=input.rows.length?'observed':'empty';
   this.append(node('p',`Displayed span extent: ${durationLabel(data.extent??0n)}. Relative to earliest returned span. Evidence completeness unknown${input.truncated?'; span list truncated':''}. Async links are references, not parent edges.`,'waterfall-status'));
   if(!input.rows.length){this.append(node('p','No spans in this selection.'));return}
   this.append(node('p','Parent layout follows unambiguous reported IDs in supplied rows. These rows are not unique invocation counts.','waterfall-status waterfall-identity-note'));
   let group;
   for(const span of data.spans){if(group!==span.groupKey){group=span.groupKey;this.append(node('h3',span.traceValid?'Trace '+span.row.TraceId:'Unreported trace context · observation '+(span.index+1)))}this.append(rowElement(span,{facts:input.facts?.[span.index]||[]}))}this.updateSelection();
  }
 }
 if(!global.customElements.get('goo-trace-waterfall'))global.customElements.define('goo-trace-waterfall',GooTraceWaterfall);
}
