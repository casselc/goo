import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
const source=readFileSync(new URL('../ui/tail.mjs',import.meta.url),'utf8');
const between=(start,end)=>source.slice(source.indexOf(start),source.indexOf(end,source.indexOf(start)));
function element(){return {hidden:false,disabled:false,dataset:{},attrs:{},writes:0,stateWrites:0,attrWrites:0,_text:'',_state:'unknown',get textContent(){return this._text},set textContent(value){this.writes++;this._text=value},get state(){return this._state},set state(value){this.stateWrites++;this._state=value},getAttribute(key){return this.attrs[key]??null},setAttribute(key,value){this.attrWrites++;this.attrs[key]=value}}}
function controls({compact=true,onResume}={}){
 const document={activeElement:null},pause=element(),resume=element(),other=element(),calls={pause:0,connect:0,align:0};
 for(const control of [pause,resume,other])control.focus=()=>{document.activeElement=control};
 const context={document,pause,resume,compact,selector:{environment:'prod'},staged:false,onResume,viewport:{},header:{},body:{children:[]},pauseStream(){calls.pause++;pause.hidden=compact;resume.hidden=false},alignPausedRecords(){calls.align++},connect(){calls.connect++;resume.hidden=compact;pause.hidden=false}};
 vm.runInNewContext(between('pause.onclick=','stop.onclick='),context);
 return {document,pause,resume,other,calls};
}
test('compact explicit Pause and Resume retain keyboard focus on the available counterpart',()=>{
 const c=controls();c.pause.focus();c.pause.onclick();assert.equal(c.document.activeElement,c.resume);assert.equal(c.calls.pause,1);assert.equal(c.calls.align,1);
 c.resume.onclick();assert.equal(c.document.activeElement,c.pause);assert.equal(c.calls.connect,1);
});
test('programmatic and noncompact controls do not steal focus; refused Resume does not transfer',()=>{
 const c=controls();c.other.focus();c.pause.onclick();assert.equal(c.document.activeElement,c.other);c.resume.onclick();assert.equal(c.document.activeElement,c.other);
 const full=controls({compact:false});full.pause.focus();full.pause.onclick();assert.equal(full.document.activeElement,full.pause);full.resume.focus();full.resume.onclick();assert.equal(full.document.activeElement,full.resume);
 const refused=controls({onResume:()=>false});refused.resume.focus();refused.resume.onclick();assert.equal(refused.document.activeElement,refused.resume);assert.equal(refused.calls.connect,0);
});
test('repeated source polls publish status only when the semantic state or message changes',()=>{
 const status=element(),context={phase:'stopped',section:element(),buffer:{rows:[],evicted:0,observed:0},gap:false,selector:{environment:'prod'},status,compact:false,counts:element(),controller:null,pause:element(),resume:element(),saving:false,generation:1,refreshBatch(){},global:{dispatchEvent(){}},CustomEvent:class{constructor(type,options){this.type=type;this.detail=options.detail}}};
 vm.runInNewContext(between('  function publish(','  function halt(')+';this.publish=publish',context);
 context.publish('unknown','Listening');context.publish('unknown','Listening');assert.equal(status.writes,1);assert.equal(status.stateWrites,0);
 context.publish('partial','Limited');assert.equal(status.writes,2);assert.equal(status.stateWrites,1);
 context.publish('partial','Limited');assert.equal(status.writes,2);assert.equal(status.stateWrites,1);
 context.publish('paused','Paused');assert.equal(status.writes,3);assert.equal(status.stateWrites,2);
 context.gap=true;context.publish('paused','Paused');assert.equal(status.writes,4);assert.match(status.textContent,/Continuity gap/);
});
test('loaded Find count is quiet without a query and avoids rewriting unchanged search results',()=>{
 const findStatus=element(),context={document:{activeElement:null},columnControl:{layout:()=>({fieldColumns:[]})},buffer:{rows:[{key:'one',record:{Body:'alpha'}}]},nodes:new Map([['one',{}]]),RecordPresentation:{matchesLoaded:()=>true},signalKind:'logs',findText:'',findStatus,clearFind:{},find:{value:''},section:element(),viewport:{scrollTop:0,scrollLeft:0},refreshSelection(){},body:{contains:()=>false},UI:{actionVisible:()=>true},findScroll:null};
 vm.runInNewContext(between('  function applyFind(','  function batchState(')+';this.applyFind=applyFind',context);
 context.applyFind();context.applyFind();assert.equal(findStatus.writes,1);assert.equal(findStatus.getAttribute('aria-live'),'off');assert.equal(findStatus.attrWrites,1);
 context.findText='alpha';context.find.value='alpha';context.applyFind();context.applyFind();assert.equal(findStatus.getAttribute('aria-live'),'polite');assert.equal(findStatus.attrWrites,2);assert.equal(findStatus.writes,1);
 context.RecordPresentation.matchesLoaded=()=>false;context.applyFind();assert.equal(findStatus.writes,2);assert.equal(findStatus.textContent,'0 of 1 loaded');
});
