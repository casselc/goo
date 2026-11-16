import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as Views from '../ui/activity-views.mjs';
import * as ConditionSummary from '../ui/condition-summary.mjs';
import * as Environment from '../ui/record-environment.mjs';
import * as PairPlan from '../ui/paired-activity-plan.mjs';
import {canonical as pairCanonical} from '../ui/tail.mjs';
import * as RecordColumns from '../ui/record-columns.mjs';
import {object} from '../ui/intent-shape.mjs';
import {validateField} from '../ui/predicate-validation.mjs';
import {canonical} from '../ui/tail.mjs';

const plain=value=>JSON.parse(JSON.stringify(value));
const layout=()=>({columns:['utc','summary'],fieldColumns:[]});
const draft=()=>({version:1,title:'API streams',selector:{signal:'both',environment:'prod',where:{any:[{field:{kind:'resource',key:'team'},op:'missing'},{field:{kind:'resource',key:'team'},op:'eq',value:''}]}},lanes:{logs:{columns:['summary','utc'],fieldColumns:[{kind:'attribute',key:'http.route'}]},spans:{columns:['utc','service','summary'],fieldColumns:[{kind:'column',key:'duration_ns'}]}}});
const stored=value=>({...structuredClone(value),id:'a'.repeat(64)});

test('paired definition retains exact scope, shared predicate and independent detached presentation',()=>{
 for(const scope of [{environment:'prod'},{environmentSelector:Environment.current('prod')},{environmentSelector:Environment.effective({mode:'named',value:'prod'})},{environmentSelector:Environment.effective({mode:'empty'})},{environmentSelector:Environment.effective({mode:'unreported'})}]){
  const v=draft();v.selector={signal:'both',...scope,where:v.selector.where};const before=structuredClone(v),out=Views.input(v);assert.deepEqual(out,before);assert.deepEqual(Views.model(stored(v)),stored(v));
  const presentation=Views.presentation(out);presentation.lanes.logs.columns.reverse();presentation.lanes.spans.fieldColumns[0].key='mutated';out.selector.where.any[0].op='exists';assert.deepEqual(v,before);
 }
 assert.equal(Views.input({version:1,title:'Logs',selector:{environment:'prod'},columns:['utc','summary']}).selector.signal,'logs');
});

test('paired definition rejects execution state, mixed layouts, lane omissions and family-only conditions as a whole',()=>{
 const mutations=[
  v=>v.columns=[],v=>v.fieldColumns=[],v=>v.baselineCapture='',v=>v.readWindow={},v=>v.health='ok',v=>v.selector.fromMs=0,v=>v.selector.cursor='x',v=>v.selector.environmentSelector=Environment.effective({mode:'all'}),
  v=>delete v.lanes.logs,v=>v.lanes.metrics=layout(),v=>v.lanes.logs.window={},v=>v.lanes.logs.fieldColumns=[{kind:'column',key:'duration_ns'}],v=>v.lanes.spans.fieldColumns=[{kind:'column',key:'body'}],v=>v.lanes.spans.columns=['utc','summary','context'],
  v=>v.selector.where={field:{kind:'column',key:'body'},op:'eq',value:'only logs'},v=>v.selector.where={all:[v.selector.where,{field:{kind:'column',key:'name'},op:'eq',value:'only spans'}]},v=>v.selector.signal='logs',v=>v.title='',v=>v.title='😀'.repeat(31),
 ];
 for(const change of mutations){const v=draft();change(v);assert.throws(()=>Views.input(v))}
 for(const signal of [false,0,null])assert.throws(()=>Views.input({version:1,title:'Logs',selector:{signal,environment:'prod'},columns:['utc','summary']}));
 assert.throws(()=>Views.model(draft()));assert.throws(()=>Views.input(stored(draft())));
 const sparse=draft();sparse.lanes.logs.columns.length=5;assert.throws(()=>Views.input(sparse));
 const single={version:1,title:'Logs',selector:{signal:'logs',environment:'prod'},columns:['utc','summary'],lanes:{logs:layout(),spans:layout()}};assert.throws(()=>Views.input(single));
});

const stream=readFileSync(new URL('../ui/stream-entry.mjs',import.meta.url),'utf8');
const saveCode=stream.slice(stream.indexOf('  function viewPresentation(){'),stream.indexOf('  function stageView(v,'));
function saveFixture(){
 const state={pair:true,dirty:false,laneDirty:false,cel:false,ready:true,layout:draft().lanes},calls={requests:[],refresh:0},context={enabled:true,records:true,route:'explore',viewGeneration:1,viewSaving:false,appliedSelector:draft().selector,currentView:null,titleDirty:true,viewTitle:{value:'API streams'},saveView:{disabled:false},viewSummary:{},viewStatus:{replaceChildren(...children){this.children=children}},shell:{addEventListener(){}},syncViews(){},pairMode:()=>state.pair,paired:{layout:()=>structuredClone(state.layout),dirty:()=>state.laneDirty},tail:{layout},latency:{baselineReference:()=> 'b'.repeat(64)},dirtyScope:()=>state.dirty,editor:{hasUnappliedCELDraft:()=>state.cel&&!state.checked},nativeControlReady:()=>state.ready,Tail:{canonical},structuredClone,n:(tag,text)=>({tag,text}),viewLibrary:{refresh(){calls.refresh++}}};
 context.ActivityViews={...Views,input:value=>Views.input(plain(value)),model:Views.model,share:id=>({id}),api:async(_path,payload)=>{calls.requests.push(payload);return new Promise(resolve=>{state.resolve=resolve})}};
 vm.runInNewContext(saveCode+'\nthis.update=updateViewAction;',context);return {state,calls,context};
}
test('paired save writes only applied configuration and refuses dirty, hidden or pending actions',async()=>{
 for(const condition of ['dirty','laneDirty','cel','ready']){const f=saveFixture();f.state[condition]=condition!=='ready';await f.context.saveView.onclick();assert.equal(f.calls.requests.length,0)}
 const invalid=saveFixture();invalid.context.viewTitle.value='😀'.repeat(31);await invalid.context.saveView.onclick();assert.equal(invalid.calls.requests.length,0);assert.equal(invalid.context.viewSaving,false);assert.match(invalid.context.viewStatus.textContent,/Invalid Activity view/);
 const f=saveFixture(),pending=f.context.saveView.onclick();assert.equal(f.calls.requests.length,1);assert.deepEqual(f.calls.requests[0],draft());assert.equal(f.context.saveView.disabled,true);await f.context.saveView.onclick();assert.equal(f.calls.requests.length,1);f.state.resolve(stored(f.calls.requests[0]));await pending;assert.equal(f.context.currentView.id,'a'.repeat(64));assert.equal(f.calls.refresh,1);assert.equal(f.context.viewSaving,false);assert.equal(f.context.titleDirty,false);
});
test('late save never adopts a configuration after lane, title, mode or route ownership changes',async()=>{
 for(const change of [f=>f.state.laneDirty=true,f=>f.state.layout.spans.columns.reverse(),f=>f.context.viewTitle.value='',f=>{f.state.pair=false;f.context.appliedSelector={signal:'logs',environment:'prod'}},f=>f.context.viewGeneration++]){
  const f=saveFixture(),pending=f.context.saveView.onclick();change(f);f.state.resolve(stored(f.calls.requests[0]));await pending;assert.equal(f.context.currentView,null);assert.equal(f.calls.refresh,0);assert.equal(f.context.titleDirty,true);assert.equal(f.context.viewSaving,false);
 }
 const f=saveFixture(),pending=f.context.saveView.onclick(),other=stored(f.calls.requests[0]);other.selector.environment='other';f.state.resolve(other);await pending;assert.equal(f.context.currentView,null);assert.match(f.context.viewStatus.textContent,/differs from the submitted/);
});

const pairSource=readFileSync(new URL('../ui/paired-activity.mjs',import.meta.url),'utf8'),stageCode=pairSource.slice(pairSource.indexOf(' function adopt('),pairSource.indexOf(' function ready('));
function stageFixture(){
 const calls=[],state={disposed:false,epoch:0,selector:{environment:'old'},plan:null,retireChronology(){},refreshHold(){},structuredClone,Tail:{canonical:pairCanonical},PairPlan:{create:(value,conditions)=>PairPlan.create(plain(value),conditions===undefined?undefined:plain(conditions))},RecordColumns:{validate:(value,available)=>RecordColumns.validate(plain(value),available)},object:(value,keys,message)=>object(plain(value),keys,message),validateField:(field,signal)=>validateField(plain(field),signal),clearInspection(){calls.push('clear')},lanes:new Map(['logs','spans'].map(signal=>[signal,{epoch:0,tail:{stage(q,layout){calls.push({signal,q:plain(q),layout:plain(layout)})}}}]))};
 vm.runInNewContext(stageCode+'\nthis.stage=stage;',state);return {state,calls};
}
test('paired recall validates both lane layouts before clearing and stages ordinary selectors without starting',()=>{
 const v=draft(),f=stageFixture();assert.equal(f.state.stage({...v.selector,signal:'logs'},v.lanes),true);assert.equal(f.calls.length,3);assert.equal(f.calls[1].q.signal,'logs');assert.equal(f.calls[2].q.signal,'spans');assert.deepEqual(f.calls[1].q.where,v.selector.where);assert.deepEqual(f.calls[1].layout,v.lanes.logs);assert.deepEqual(f.calls[2].layout,v.lanes.spans);
 for(const change of [v=>delete v.spans,v=>v.spans.columns=['utc','context','summary'],v=>v.logs.fieldColumns=[{kind:'column',key:'name'}]]){const bad=structuredClone(v.lanes),f=stageFixture();change(bad);assert.throws(()=>f.state.stage({...v.selector,signal:'logs'},bad));assert.deepEqual(f.calls,[]);assert.equal(f.state.epoch,0)}
});

const recallCode=stream.slice(stream.indexOf('  function stageView(v,'),stream.indexOf('  async function restoreView()'));
function recallFixture(){
 const calls=[],node=()=>({replaceChildren(){},remove(){}}),context={ConditionSummary,structuredClone,viewConflict:node(),stop(){calls.push('pause')},editor:{retireCEL(){},set(where,signal){calls.push(['editor',signal])}},scopeStaged:true,numericPlot:null,plotOrigin:{},plotEpoch:0,triage:null,explanation:null,explain:{},triageButton:{},latency:{invalidate(){},restoreBaseline(value){calls.push(['baseline',value])}},logContext:null,onLeaveFixed:null,gate:{pause(){}},enabled:false,records:false,returnAnchor:{},selectedOpener:{},environmentDraft:{restore(){}},renderEnvironment(){},signal:{value:'logs'},previousSignal:'logs',filters:{restore(){}},filterError:null,apply:{},renderFilters(){},clearInspection(){},currentView:null,titleDirty:true,viewTitle:{},viewSummary:{},viewStatus:node(),draft:{},applied:Object.assign(node(),{textContent:'Applied · old environment',title:'old selector'}),viewOptions:{open:true},advanced:{open:true},n(){},ActivityViews:{share(){}},global:{dispatchEvent(){calls.push('navigate')}},CustomEvent:class{},present(){calls.push('present')},updateViewAction(){},revealFocus(){calls.push('focus-pair')},tail:{reset(){calls.push('reset-single')},stage(value){calls.push(['stage-single',plain(value)])},focusStart(){calls.push('focus-single')}},paired:{reset(){calls.push('reset-pair')},stage(value,lanes){calls.push(['stage-pair',plain(value),plain(lanes)])}}};
 context.pairMode=()=>context.signal.value==='both';context.editorSignal=()=>context.pairMode()?'logs':context.signal.value;
 vm.runInNewContext(recallCode+'\nthis.recall=stageView;',context);return {context,calls};
}
test('single↔pair recall only stages; interactive focus follows presentation while initial recall stays quiet',()=>{
 for(const pair of [false,true])for(const focus of [false,true]){
  const f=recallFixture(),v=pair?stored(draft()):{version:1,id:'b'.repeat(64),title:'Single spans',selector:{signal:'spans',environment:'prod'},columns:['utc','summary']};
  f.context.recall(v,{focus});assert.equal(f.context.gate.eligible,false);assert.equal(f.context.enabled,true);assert.equal(f.context.viewOptions.open,false);assert.equal(f.context.apply.hidden,!pair);assert.match(f.context.applied.textContent,/prod/);assert.deepEqual(JSON.parse(f.context.applied.title),v.selector);
  const staged=f.calls.find(call=>Array.isArray(call)&&call[0]===(pair?'stage-pair':'stage-single'));assert(staged);assert.equal(staged[1].signal,pair?'logs':'spans');if(pair)assert.deepEqual(staged[2],v.lanes);
  assert.equal(f.calls.includes(pair?'reset-single':'reset-pair'),true);assert.equal(f.calls.includes(pair?'focus-pair':'focus-single'),focus);if(focus)assert(f.calls.indexOf('present')<f.calls.indexOf(pair?'focus-pair':'focus-single'));
 }
});
const clickCode=stream.slice(stream.indexOf('  function viewLinkClick(event){'),stream.indexOf('  function present(){'));
test('native view links carry focus intent and re-open the current hash without waiting for a hashchange',()=>{
 const hash='#activity-view='+'a'.repeat(64),context={viewFocusRequest:null,location:{hash},nativeControlReady:()=>true,restoreView(){context.restores++},restores:0};vm.runInNewContext(clickCode+'\nthis.click=viewLinkClick;',context);
 const link={getAttribute:()=>hash},event={button:0,target:{closest:()=>link},preventDefault(){this.prevented=true}};context.click(event);assert.equal(context.restores,1);assert.equal(event.prevented,true);assert.equal(context.viewFocusRequest.hash,hash);
 context.location.hash='#explore';context.viewFocusRequest=null;context.click({...event,prevented:false});assert.equal(context.restores,1);assert.equal(context.viewFocusRequest.hash,hash);
 for(const patch of [{ctrlKey:true},{button:1},{defaultPrevented:true}]){context.viewFocusRequest=null;context.click({...event,...patch});assert.equal(context.viewFocusRequest,null)}
 context.nativeControlReady=()=>false;context.viewFocusRequest=null;context.click(event);assert.equal(context.viewFocusRequest,null);
});
