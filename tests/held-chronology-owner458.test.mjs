import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
import {Element as MinimalElement} from './support/raw-observation-dom.mjs';
import * as RealUI from '../ui/components.mjs';
import * as RealRecords from '../ui/record-presentation.mjs';
import {lane as heldLane,item} from './support/held-chronology458-fixture.mjs';
const canonical=value=>JSON.stringify(value,(_,v)=>v&&typeof v==='object'&&!Array.isArray(v)?Object.fromEntries(Object.entries(v).sort(([a],[b])=>a.localeCompare(b))):v);
const tails=[],filters=[];let onPause;
class Element extends MinimalElement{
 constructor(tag){super(tag);this.handlers=new Map()}
 addEventListener(name,handler){this.handlers.set(name,handler)}
 querySelector(selector){const parts=selector.split('>');return parts.length===2?super.querySelector(parts[0])?.querySelector(parts[1])??null:super.querySelector(selector)}
 set data(value){this._data=value;this.replaceChildren();if(value){const find=new Element('input');find.type='search';this.append(find)}}
 get data(){return this._data}
}
mock.module('../ui/components.mjs',{exports:{...RealUI,node:(tag,text,className)=>{const node=new Element(tag);if(text!==undefined)node.textContent=text;if(className)node.className=className;return node}}});
mock.module('../ui/record-presentation.mjs',{exports:{...RealRecords,attributes:(host,record)=>host.append(new Element('record-fields')),description:record=>record.Body??record.SpanName}});
mock.module('../ui/held-chronology-view.mjs',{exports:{}});
mock.module('../ui/tail.mjs',{exports:{canonical,transportIdentity:selector=>{const value=structuredClone(selector);if(value?.signal==='logs')delete value.signal;return canonical(value)},recordQuestion:(selector,signal,entry)=>({signal,...structuredClone(selector),...structuredClone(entry.readWindow),limit:100,offset:0}),mount:(host,options)=>{
 const signal=host.parentElement.dataset.signal,state={...heldLane(signal),selector:null,completedRead:null,readWindow:null,connection:null},calls={starts:[],stages:[],pauses:[]};
 const publish=()=>options.onState({...structuredClone(state),owner:signal+'-owner',retained:state.rows.length});
 const tail={options,state,calls,publish,ready:()=>true,scope:()=>({...structuredClone(state),retained:state.rows.length}),snapshot:()=>structuredClone(state),layout:()=>({columns:['utc','service','summary'],fieldColumns:[]}),stage(selector){state.selector=structuredClone(selector);if(signal==='logs')delete state.selector.signal;state.rows=[];state.completedRead=null;state.readWindow=null;state.phase='paused';state.staged=true;calls.stages.push(structuredClone(selector));publish()},start(selector){state.selector=structuredClone(selector);if(signal==='logs')delete state.selector.signal;state.phase='loading';state.staged=false;calls.starts.push(structuredClone(selector));publish();return true},pause(reason){calls.pauses.push(reason);state.phase='paused';publish();onPause?.(signal,tail)},reset(){state.selector=null;state.rows=[];state.completedRead=null;state.readWindow=null;state.connection=null;state.gap=false;state.evicted=0;state.find={text:'',revision:state.find.revision+1};state.phase='stopped';publish()},setReadBlocked(){},focusStart(){},dispose(){}};tails.push(tail);return tail;
}}});
mock.module('../ui/paired-lane-filter.mjs',{exports:{mount:(host,options)=>{
 let where,dirty=false;const filter={options,where:()=>structuredClone(where),dirty:()=>dirty,condition:()=>({where,dirty}),resolve(){dirty=false;return true},reset(value){where=structuredClone(value);dirty=false},retire(){},sync(){},dispose(){},apply(value,opener){if(options.onApply(value,opener,{current:()=>true})){where=structuredClone(value);return true}return false}};filters.push(filter);return filter;
}}});
const {mount}=await import('../ui/paired-activity.mjs');
const button=(host,text)=>host.querySelectorAll('button').find(node=>node.textContent===text||node.getAttribute('aria-label')===text);
function fixture(){
 const body=new Element('body');body.root=true;globalThis.Node=Element;globalThis.document={body,hidden:false,activeElement:null,createElement:tag=>new Element(tag)};const host=new Element('main'),opener=new Element('button');body.append(host,opener);let active=true;onPause=undefined;const handoffs=[];
 const owner=mount(host,{current:()=>active,onOpenFixed:(...args)=>handoffs.push(args)}),pair=tails.slice(-2),laneFilters=filters.slice(-2);owner.stage({environment:'prod'});
 for(const [index,tail]of pair.entries()){
  const signal=index?'spans':'logs',snapshot=heldLane(signal,[item(signal,0),item(signal,1,'1700000000000000002')]);Object.assign(tail.state,snapshot);if(signal==='logs')delete tail.state.selector.signal;tail.state.phase='unknown';tail.state.find={text:'lane-only Find',revision:3};tail.publish();
 }
 owner.sync();const chronology=host.querySelector('goo-held-chronology'),hold=button(host,'Hold chronology'),back=button(host,'Back to lanes');
 const inspect=(signal='logs',key=pair[signal==='logs'?0:1].state.rows[0].key)=>{const control=new Element('button');control.textContent='Inspect';chronology.append(control);chronology.handlers.get('goo-chronology-inspect')({detail:{signal,key},composedPath:()=>[control,chronology]});return control};
 return {owner,host,opener,chronology,hold,back,handoffs,logs:pair[0],spans:pair[1],logFilter:laneFilters[0],spanFilter:laneFilters[1],inspect,setActive:value=>active=value};
}
const stats=f=>({starts:f.logs.calls.starts.length+f.spans.calls.starts.length,rows:f.logs.state.rows.map(row=>row.key),otherRows:f.spans.state.rows.map(row=>row.key),find:f.logs.state.find.text,otherFind:f.spans.state.find.text});

test('Hold pauses both lanes, freezes stripped observations, and Back preserves buffers/Find without a source start',()=>{
 const f=fixture(),before=stats(f);assert.equal(f.hold.disabled,false);assert.equal(f.hold.onclick(),true);assert.equal(f.logs.calls.pauses.length,1);assert.equal(f.spans.calls.pauses.length,1);assert.equal(f.owner.snapshot().view,'chronology');assert.equal(f.chronology.hidden,false);assert.equal(Object.isFrozen(f.chronology.data.lanes.logs.rows[0].record),true);assert.equal(JSON.stringify(f.chronology.data).includes('SECRET-TOKEN'),false);const record=f.chronology.data.lanes.logs.rows[0].record;f.logs.state.rows[0].record.Body='later mutation';assert.notEqual(record.Body,'later mutation');assert.equal(f.back.onclick(),true);assert.equal(f.owner.snapshot().view,'lanes');assert.equal(f.chronology.hidden,true);assert.deepEqual(stats(f),before);assert.equal(document.activeElement,f.hold);
});
test('reentrant reset or replacement during the first pause cannot freeze or pause replacement work',()=>{
 for(const operation of ['reset','stage','start']){
  const f=fixture();let once=true;onPause=signal=>{if(signal!=='logs'||!once)return;once=false;if(operation==='reset')f.owner.reset();else if(operation==='stage')f.owner.stage({environment:'next'});else f.owner.start({environment:'next'})};
  assert.equal(f.hold.onclick(),false);assert.equal(f.owner.snapshot().view,'lanes');assert.equal(f.chronology.data,undefined);assert.equal(f.spans.calls.pauses.length,0,'replacement span owner was not paused');if(operation==='reset')assert.equal(f.owner.snapshot().selector,null);else assert.equal(f.owner.snapshot().selector.environment,'next');if(operation==='start')assert.equal(f.spans.calls.starts.length,1);
 }
});
test('held inspection routes only a known visible opener and retires after Back, stage, start or lane Apply',()=>{
 for(const operation of ['back','stage','start','apply']){
  const f=fixture();assert.equal(f.hold.onclick(),true);const original=f.inspect();assert.equal(button(f.host,'Analyze observation window').disabled,false);const handler=f.chronology.handlers.get('goo-chronology-inspect');
  if(operation==='back')f.back.onclick();else if(operation==='stage')f.owner.stage({environment:'next'});else if(operation==='start')f.owner.start({environment:'next'});else assert.equal(f.logFilter.apply({field:{kind:'column',key:'severity_number'},op:'gte',value:'17'},f.opener),true);
  handler({detail:{signal:'logs',key:f.logs.state.rows[0]?.key??'a'.repeat(64)+':0'},composedPath:()=>[original,f.chronology]});assert.equal(f.owner.snapshot().view,'lanes');assert.equal(f.chronology.selectedKey,null);assert.equal(button(f.host,'Analyze observation window'),undefined);assert.equal(f.handoffs.length,0);
 }
 const f=fixture();f.hold.onclick();const handler=f.chronology.handlers.get('goo-chronology-inspect'),foreign=new Element('button');f.host.append(foreign);handler({detail:{signal:'logs',key:f.logs.state.rows[0].key},composedPath:()=>[foreign,f.chronology]});assert.equal(button(f.host,'Analyze observation window'),undefined);f.inspect('logs','foreign');assert.equal(button(f.host,'Analyze observation window'),undefined);
});
test('cached Analyze is reused and refreshed while retaining its occurrence’s original window and source clock',()=>{
 const f=fixture(),row=f.logs.state.rows[0];row.readWindow={fromMs:1699999940000,toMs:1700000000000};f.hold.onclick();const opener=f.inspect(),analyze=button(f.host,'Analyze observation window'),detail=f.host.querySelector('.ui-inspector'),content=detail.querySelector('.ui-inspector-body').firstChild;assert.ok(detail.textContent.includes('2023-11-14T22:12:20.000Z'));assert.deepEqual(content.querySelector('goo-stat-list').items,[{label:'Source read UTC',value:row.observedAt,format:'utc'}]);f.logs.state.rows[0].readWindow={fromMs:1700000000000,toMs:1700000060000};button(f.host,'Close details').onclick();assert.equal(document.activeElement,opener);f.inspect();assert.equal(button(f.host,'Analyze observation window'),analyze);assert.equal(detail.querySelector('.ui-inspector-body').firstChild,content);analyze.onclick();assert.equal(f.handoffs.length,1);assert.deepEqual(f.handoffs[0][0],{signal:'logs',environment:'prod',fromMs:1699999940000,toMs:1700000000000,limit:100,offset:0});assert.equal(f.handoffs[0][2],analyze);assert.equal(f.handoffs[0][1].columns.at(-1),'context');assert.equal(f.handoffs[0][0].membership,undefined);assert.equal(stats(f).starts,0);
});
test('withdrawal during Analyze pause never stages a fixed question from retired observations',()=>{
 const f=fixture();f.hold.onclick();f.inspect();const analyze=button(f.host,'Analyze observation window');let once=true;onPause=signal=>{if(signal==='logs'&&once){once=false;f.owner.start({environment:'next'})}};const previous=f.spans.calls.pauses.length;analyze.onclick();assert.equal(f.handoffs.length,0);assert.equal(f.spans.calls.pauses.length,previous);assert.equal(f.spans.calls.starts.length,1);assert.equal(f.owner.snapshot().view,'lanes');
});
test('visibility and withdrawal guards reject stale control use, and Close falls back to visible Find after rerender',()=>{
 const f=fixture();f.setActive(false);assert.equal(f.hold.onclick(),false);f.setActive(true);f.hold.hidden=true;assert.equal(f.hold.onclick(),false);f.hold.hidden=false;f.hold.onclick();const opener=f.inspect();opener.remove();const find=f.chronology.querySelector('input');button(f.host,'Close details').onclick();assert.equal(document.activeElement,find);assert.equal(f.chronology.selectedKey,null);assert.equal(f.handoffs.length,0);f.chronology.hidden=true;f.inspect();assert.equal(button(f.host,'Analyze observation window').parentElement.parentElement.hidden,true);assert.equal(stats(f).starts,0);
});

test('a reentrant single-lane Apply changes the ownership vector and prevents stale hold adoption',()=>{
 const f=fixture();let once=true;onPause=signal=>{if(signal==='logs'&&once){once=false;assert.equal(f.spanFilter.apply({field:{kind:'column',key:'duration_ns'},op:'gte',value:'100'},f.opener),true)}};assert.equal(f.hold.onclick(),false);assert.equal(f.owner.snapshot().view,'lanes');assert.equal(f.chronology.data,undefined);assert.equal(f.logs.calls.pauses.length,1);assert.equal(f.spans.calls.pauses.length,0,'new span stream remains owned by its Apply');assert.equal(f.spans.calls.starts.length,1);assert.equal(f.owner.layout().spans.where.value,'100');assert.equal(f.owner.snapshot().selector.environment,'prod');
});

test('a closed unconfigured sibling does not block useful held data or acquire an invented selector',()=>{
 const f=fixture();f.spans.reset();f.owner.sync();assert.equal(f.hold.disabled,false);assert.equal(f.hold.onclick(),true);assert.equal(f.owner.snapshot().view,'chronology');assert.equal(f.chronology.data.lanes.logs.rows.length,2);assert.equal(f.chronology.data.lanes.spans.selector,null);assert.equal(f.chronology.data.lanes.spans.completedRead,null);assert.deepEqual(f.chronology.data.lanes.spans.rows,[]);assert.equal(f.spans.calls.starts.length,0);f.inspect();assert.equal(button(f.host,'Analyze observation window').disabled,false);f.back.onclick();assert.equal(f.owner.snapshot().view,'lanes');assert.equal(f.spans.state.selector,null);assert.deepEqual(f.spans.state.rows,[]);assert.equal(f.spans.calls.starts.length,0);
});

test('related held intent requires a current chronology inspector and valid reported trace identity',()=>{
 const f=fixture();f.logs.state.rows[0].record.TraceId='a'.repeat(32);f.hold.onclick();f.inspect();const related=button(f.host,'Related held records'),calls=[];f.chronology.showRelated=(...args)=>calls.push(args);const before=stats(f);assert.equal(related.disabled,false);related.onclick();assert.deepEqual(calls,[['logs',f.logs.state.rows[0].key]]);assert.deepEqual(stats(f),before);related.hidden=true;related.onclick();assert.equal(calls.length,1);related.hidden=false;f.back.onclick();related.onclick();assert.equal(calls.length,1);assert.equal(f.handoffs.length,0);
 const invalid=fixture();invalid.hold.onclick();invalid.inspect();assert.equal(button(invalid.host,'Related held records').disabled,true);
});

test('Return restores an inspector through a visible Find fallback when its original row is absent',()=>{
 const f=fixture();f.hold.onclick();f.inspect();const selected=f.chronology.selectedKey;f.inspect('spans');assert.match(f.host.textContent,/Observed span/);const restore=f.chronology.handlers.get('goo-chronology-return');f.chronology.selectedKey=selected;restore({target:f.chronology,detail:{selectedKey:selected}});assert.match(f.host.textContent,/Observed log/);assert.equal(f.chronology.selectedKey,selected);assert.equal(f.handoffs.length,0);const old=button(f.host,'Analyze observation window');f.back.onclick();restore({target:f.chronology,detail:{selectedKey:selected}});assert.equal(f.owner.snapshot().view,'lanes');old.onclick();assert.equal(f.handoffs.length,0);
});
