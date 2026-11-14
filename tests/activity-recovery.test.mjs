import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import {revealFocus} from '../ui/dom-visibility.mjs';
import {plan,lease,mount} from '../ui/activity-recovery.mjs';
import {effective} from '../ui/record-environment.mjs';
const copy=value=>structuredClone(value);
function context(){
 const toMs=1800000,read={fromMs:toMs-60000,toMs,rowCount:0,limited:false,observedAt:new Date(toMs+500).toISOString(),revision:3};
 return {active:true,clean:true,phase:'unknown',staged:false,retained:0,evicted:0,gap:false,epoch:7,connection:'a'.repeat(64),selector:{signal:'logs',environment:'',environmentSelector:effective({mode:'unreported'}),where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'worker'},{any:[{field:{kind:'resource',key:'service.namespace'},op:'missing'},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''}]}]}},readWindow:{fromMs:read.fromMs,toMs:read.toMs},completedRead:read};
}
test('wider questions anchor to the completed read and preserve the full detached typed scope',()=>{
 for(const signal of ['logs','spans'])for(const scope of [{mode:'named',value:'prod'},{mode:'empty'},{mode:'unreported'}]){
  const c=context();c.selector.signal=signal;c.selector.environmentSelector=effective(scope);const before=copy(c),q=plan(c);assert.equal(q.signal,signal);assert.equal(q.fromMs,c.completedRead.toMs-15*60000);assert.equal(q.toMs,c.completedRead.toMs);assert.equal(q.limit,100);assert.equal(q.offset,0);assert.deepEqual(q.where,c.selector.where);assert.deepEqual(q.environmentSelector,c.selector.environmentSelector);q.where.all[0].value='changed';q.environmentSelector.scope.mode='broken';assert.deepEqual(c,before);
 }
 const c=context();delete c.selector.signal;assert.equal(plan(c).signal,'logs');
});
test('empty delivered batches are not absence: retained records, limits, failures and incomplete reads refuse',()=>{
 const edits=[c=>c.active=false,c=>c.clean=false,c=>c.staged=true,c=>c.retained=1,c=>c.completedRead=null,c=>c.completedRead.rowCount=1,c=>c.completedRead.limited=true,c=>c.completedRead.observedAt=new Date(c.completedRead.toMs-1).toISOString(),c=>c.completedRead.observedAt='not a read clock',c=>c.completedRead.revision=0,c=>c.epoch=-1,c=>c.connection='changed',c=>c.readWindow.toMs++,c=>{c.completedRead.fromMs=60000;c.completedRead.toMs=120000;c.completedRead.observedAt=new Date(120001).toISOString();c.readWindow={fromMs:60000,toMs:120000}}];
 for(const phase of ['loading','stale','unavailable','partial','stopped'])edits.push(c=>c.phase=phase);
 for(const edit of edits){const c=context();edit(c);assert.throws(()=>plan(c));assert.equal(lease(c),null)}
 // A continuity gap is not repaired by this action. It may still stage a question
 // anchored to a later successful read without asserting upstream completeness.
 const gap=context();gap.gap=true;assert.doesNotThrow(()=>plan(gap));
});
test('ownership survives an intentional pause but retires a new batch, scope epoch or connection',()=>{
 const c=context(),key=lease(c);c.phase='paused';assert.equal(lease(c),key);for(const edit of [c=>c.epoch++,c=>c.completedRead.revision++,c=>c.connection='b'.repeat(64),c=>c.selector.where.all[0].value='other']){const next=copy(c);edit(next);assert.notEqual(lease(next),key)}
});
class Element{
 constructor(tag){this.localName=tag;this.parentElement=null;this.children=[];this.dataset={};this.attrs=new Map();this.hidden=false;this.disabled=false;this.root=false;this.open=true;this._text=''}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}get textContent(){return this._text+this.children.map(c=>c.textContent).join('')}set textContent(value){this._text=String(value);this.children=[]}
 append(...nodes){for(const node of nodes){node.parentElement=this;this.children.push(node)}}setAttribute(k,v){this.attrs.set(k,String(v))}getAttribute(k){return this.attrs.get(k)??null}
 getClientRects(){return this.isConnected?[{}]:[]}matches(selector){return selector===':disabled'?this.disabled:false}
 querySelector(selector){return selector===':scope > summary'?this.children.find(c=>c.localName==='summary')||null:null}
 closest(selector){for(let p=this;p;p=p.parentElement){if(selector==='details:not([open])'&&p.localName==='details'&&!p.open)return p;if(selector==='fieldset:disabled'&&p.localName==='fieldset'&&p.disabled)return p;if(selector.includes('[hidden]')&&p.hidden)return p;if(selector.includes('[inert]')&&p.attrs.has('inert'))return p;if(selector.includes('[aria-hidden="true"]')&&p.attrs.get('aria-hidden')==='true')return p}return null}
 contains(node){return node===this||this.children.some(c=>c.contains(node))}remove(){if(this.parentElement){this.parentElement.children=this.parentElement.children.filter(c=>c!==this);this.parentElement=null}}
}
function fixture({pause,onStage,layout}={}){
 const c=context(),host=new Element('section'),stages=[],errors=[];host.root=true;globalThis.document={hidden:false,createElement:tag=>new Element(tag)};
 const api=mount(host,{current:()=>c,pause:()=>{c.phase='paused';pause?.(c,api)},layout:layout||(()=>({columns:['utc','service','summary'],fieldColumns:[]})),onStage:(...args)=>{stages.push(args);onStage?.(c,...args)},onError:()=>errors.push('unavailable')});const section=host.children[0],button=section.children[0];return {c,host,api,section,button,stages,errors};
}
test('actual recovery control stages locally and preserves button ownership through pause and held return',()=>{
 const f=fixture({pause:(_c,api)=>api.sync()});f.button.onclick();assert.equal(f.stages.length,1);assert.equal(f.c.phase,'paused');assert.deepEqual(f.stages[0][1],{columns:['utc','service','summary'],fieldColumns:[]});assert.equal(f.stages[0][2],f.button);assert.equal(f.section.hidden,false);assert.equal(f.section.dataset.pollRows,'0');assert.equal(f.section.dataset.observedAt,f.c.completedRead.observedAt);
 f.c.active=false;f.api.sync();assert.equal(f.section.hidden,true);f.c.active=true;f.api.sync();assert.equal(f.section.hidden,false);f.button.onclick();assert.equal(f.stages.length,2);
});
test('captured hidden, dirty and retired controls never stage, including pause reentrancy and ABA',()=>{
 for(const edit of [f=>f.button.hidden=true,f=>f.host.attrs.set('inert',''),f=>f.c.clean=false,f=>f.c.phase='stale',f=>f.c.retained=2,f=>document.hidden=true,f=>f.api.dispose()]){const f=fixture(),old=f.button.onclick;edit(f);old();assert.equal(f.stages.length,0)}
 for(const change of [(c,api)=>{c.clean=false;api.sync()},c=>{c.selector.where.all[0].value='other';c.epoch++;c.selector.where.all[0].value='worker';c.epoch++},c=>c.completedRead.revision++]){const f=fixture({pause:change});f.button.onclick();assert.equal(f.stages.length,0)}
 const f=fixture(),old=f.button.onclick;f.c.completedRead.revision++;f.api.sync();old();assert.equal(f.stages.length,0);
});
test('layout, pause and handoff failures are caught without resuming or publishing replacement data',()=>{
 for(const options of [{layout:()=>{throw Error('bad layout')}},{pause:()=>{throw Error('pause failed')}},{onStage:()=>{throw Error('handoff failed')}}]){const f=fixture(options);assert.doesNotThrow(()=>f.button.onclick());assert.equal(f.errors.length,1);assert.ok(['unknown','paused'].includes(f.c.phase));assert.equal(f.c.retained,0)}
});

test('actual Activity handoff refusal and throw restore the paused parent, anchor, route and opener',()=>{
 const source=readFileSync(new URL('../ui/stream-entry.mjs',import.meta.url),'utf8'),begin=source.indexOf('recovery=ActivityRecovery.mount('),statement=source.slice(begin,source.indexOf('\n',begin));assert(begin>=0);
 for(const outcome of ['refused','throw']){
  const host=new Element('section'),opener=new Element('button');host.root=true;host.append(opener);opener.scrollIntoView=()=>{};opener.focus=()=>{document.activeElement=opener};globalThis.document={hidden:false,activeElement:opener};
  const previous={opener:'original anchor'},calls={present:[],leave:0,start:0,resume:0,open:[]};let options;
  const state={enabled:true,advanced:{open:false},returnAnchor:previous,records:true,route:'explore',plotEpoch:7,tailHost:host,tail:{phase:'paused',scope:()=>context(),layout:()=>({columns:['utc']})},draft:{textContent:''},stop(){calls.start++},dirtyScope:()=>false,
   ActivityRecovery:{mount(_host,value){options=value;return {}}},rememberParent(button){state.returnAnchor={opener:button}},present(){host.hidden=!state.enabled;if(host.hidden)document.activeElement=null;calls.present.push({enabled:state.enabled,anchor:state.returnAnchor})},
   onOpenFixed(question,layout){calls.open.push({question,layout});state.location.hash='#question=child';if(outcome==='throw')throw Error('private failure');return false},onLeaveFixed(){calls.leave++},revealFocus,
   location:{hash:'#explore'},history:{replaceState(_state,_title,hash){state.location.hash=hash}},Error};
  vm.runInNewContext(statement,state);const q=plan(context()),layout={columns:['utc','service']};assert.throws(()=>options.onStage(q,layout,opener));options.onError();
  assert.equal(state.enabled,true);assert.equal(state.advanced.open,false);assert.equal(state.returnAnchor,previous);assert.equal(state.location.hash,'#explore');assert.equal(document.activeElement,opener);assert.equal(host.hidden,false);assert.equal(state.tail.phase,'paused');assert.equal(calls.leave,1);assert.equal(calls.start,0);assert.equal(calls.resume,0);assert.deepEqual(calls.open,[{question:q,layout}]);assert.deepEqual(calls.present.map(x=>x.enabled),[false,true]);assert.equal(state.draft.textContent,'Wider search unavailable.');
 }
});

test('actual Activity draftScope hides reverted changes while keeping the observation paused',()=>{
 const source=readFileSync(new URL('../ui/stream-entry.mjs',import.meta.url),'utf8'),begin=source.indexOf('  function draftScope('),statement=source.slice(begin,source.indexOf('\n',begin));assert(begin>=0);
 for(const scenario of [{applied:true,dirty:false,error:null,staged:false,hidden:true,text:''},{applied:true,dirty:true,error:null,staged:false,hidden:false,text:'Unapplied scope'},{applied:true,dirty:true,error:'Invalid condition.',staged:false,hidden:false,text:'Invalid condition.'},{applied:false,dirty:false,error:null,staged:false,hidden:false,text:''},{applied:true,dirty:true,error:null,staged:true,hidden:true,text:'Unapplied scope'}]){
  const calls={pause:0,recovery:0},noop=()=>{},state={numericPlot:null,plotOrigin:null,plotEpoch:1,triage:null,explanation:null,explain:{},triageButton:{},latency:null,logContext:null,valueLookup:null,editor:{setSuggestions:noop},stop(){calls.pause++},applied:{textContent:'Applied'},toolbar:{append:noop},apply:{hidden:false,disabled:!!scenario.error},draft:{textContent:'Unapplied scope'},appliedSelector:scenario.applied?{signal:'logs'}:null,filterError:scenario.error,scopeStaged:scenario.staged,dirtyScope:()=>scenario.dirty,updatePatternAction:noop,renderFilters:noop,updateViewAction:noop,recovery:{sync(){calls.recovery++}}};
  vm.runInNewContext(statement+'\ndraftScope(false);',state);
  assert.equal(state.apply.hidden,scenario.hidden);assert.equal(state.apply.textContent,scenario.applied?'Apply scope':'Start stream');assert.equal(state.draft.textContent,scenario.text);assert.equal(state.apply.disabled,!!scenario.error);assert.equal(calls.pause,1);assert.equal(calls.recovery,1);
 }
});
