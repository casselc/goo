import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as TailRecovery from '../ui/tail-recovery.mjs';
import {SSEParser,transportIdentity,canonical} from '../ui/tail.mjs';
import {nativeControlReady} from '../ui/dom-visibility.mjs';
const missing={reason:'source_missing',retry:'explicit',refusal:{kind:'unavailable',detail:'The requested record source is missing.'}},timeout={reason:'source_timeout',retry:'automatic',refusal:{kind:'timeout',detail:'The source read exceeded its time limit.'}};
test('closed tail outcomes distinguish retry policy without trusting prose, arbitrary statuses or mismatched source kinds',()=>{
 for(const data of [missing,timeout,{reason:'connection_ended',retry:'automatic'},{...missing,reason:'source_unavailable',refusal:{kind:'unavailable',detail:'The source request could not be completed.'}}])assert.deepEqual(TailRecovery.gap(data),data);
 for(const data of [{...missing,retry:'automatic'},{...timeout,retry:'explicit'},{...missing,extra:1},{...missing,reason:'guessed'},{...missing,refusal:{kind:'unavailable',detail:'password SQL privatehost'}},{reason:'connection_ended',retry:'automatic',refusal:missing.refusal},null])assert.deepEqual(TailRecovery.gap(data),TailRecovery.malformed());
 for(const [status,error,reason,retry] of [[400,'Invalid tail selector.','invalid_selector','explicit'],[429,'Live stream capacity reached.','capacity','automatic'],[500,'Live streaming unavailable.','server_unavailable','automatic']]){const value={error,reason,retry};assert.equal(TailRecovery.http(status,value).retry,retry);assert.deepEqual(TailRecovery.http(status,{...value,error:'raw secret'}),TailRecovery.malformed());assert.deepEqual(TailRecovery.http(502,value),TailRecovery.malformed())}
});
const source=readFileSync(new URL('../ui/tail.mjs',import.meta.url),'utf8'),between=(a,b)=>source.slice(source.indexOf(a),source.indexOf(b,source.indexOf(a)));
class Element{
 constructor(){this.hidden=false;this.disabled=false;this.isConnected=true;this.dataset={};this.textContent='';this.state='unknown'}
 closest(selector){if(selector==='[inert]'&&this.inert)return this;if(selector==='[hidden], [aria-hidden="true"]'&&this.hidden)return this;return null}getClientRects(){return this.hidden?[]:[{}]}getAttribute(){return null}matches(selector){return selector===':disabled'&&this.disabled}focus(){document.activeElement=this}
}
function fixture(response){
 globalThis.document={hidden:false,activeElement:null};const rows=[{key:'held',record:{Body:'held'}}],timers=[],calls={fetch:0,clear:0,render:[],publish:[],resume:0};
 const state={TailRecovery,SSEParser,transportIdentity,canonical,nativeControlReady,renderRefusal:(_host,error)=>calls.render.push(error),document,AbortController,TextDecoder,Date,Number,JSON,Error,structuredClone,RegExp,
  staged:false,signalKind:'logs',resume:new Element(),retry:new Element(),pause:new Element(),status:new Element(),failureHost:new Element(),section:new Element(),terminalFailure:null,renderedFailure:null,completedRead:{successful:true},readRevision:1,generation:0,controller:null,timer:null,watchdog:null,selector:{environment:'prod'},sequence:0,lastBatch:0,lastTo:null,gap:false,phase:'unknown',compact:true,connection:'a'.repeat(64),localInspector:null,buffer:{rows,append(items){rows.push(...items)}},
  clear(){calls.clear++;rows.length=0;state.connection=null},reconcile(){},setInterval:()=>1,clearInterval(){},clearTimeout(){},setTimeout(fn){timers.push(fn);return timers.length},
  halt(){state.generation++;state.controller?.abort();state.controller=null;state.timer=null},publish(phase,text){state.phase=phase;state.status.textContent=text;state.pause.hidden=!state.controller;state.pause.disabled=!state.controller;state.resume.hidden=!!state.controller;calls.publish.push({phase,text})},
  onResume(){calls.resume++;return true},fetch:async()=>{calls.fetch++;if(response instanceof Error)throw response;return typeof response==='function'?response():response}};
 vm.runInNewContext(between('  function clearFailure(','  function pauseStream(')+between('  function pauseStream(','\n  pause.onclick=')+'\nthis.connect=connect;this.pauseStream=pauseStream;',state);return {state,calls,timers,rows};
}
const http=(status,value)=>({ok:false,status,json:async()=>value});
const stream=frames=>({ok:true,status:200,headers:{get:()=> 'text/event-stream'},body:{getReader(){let sent=false;return {read:async()=>sent?{done:true}:(sent=true,{done:false,value:new TextEncoder().encode(frames.map((f,i)=>'id: '+(i+1)+'\nevent: '+f.event+'\ndata: '+JSON.stringify(f.data)+'\n\n').join(''))}),releaseLock(){}}}}});
const reset={event:'reset',data:{selector:{environment:'prod'},replay:false,windowMs:60000,connection:'b'.repeat(64)}};
test('actual tail connect stops terminal HTTP/source/protocol outcomes, retains rows and clears successful read metadata',async()=>{
 for(const response of [http(400,{error:'Invalid tail selector.',reason:'invalid_selector',retry:'explicit'}),http(502,{error:'secret'}),stream([reset,{event:'gap',data:missing}]),stream([reset,{event:'unexpected',data:{secret:'raw SQL'}}])]){
  const f=fixture(response);await f.state.connect();assert.equal(f.calls.fetch,1);assert.equal(f.timers.length,0);assert.equal(f.state.phase,'unavailable');assert.equal(f.state.completedRead,null);assert.equal(f.rows.length,1);assert.equal(f.calls.clear,0);assert.equal(f.state.retry.hidden,false);assert.equal(f.state.resume.hidden,true);assert.equal(f.calls.render.length,1);assert.doesNotMatch(f.calls.publish.map(x=>x.text).join(' '),/secret|SQL/);
  const previous=f.state.terminalFailure;f.state.pauseStream('Paused for inspection.');assert.equal(f.state.phase,'unavailable');assert.equal(f.state.terminalFailure,previous);assert.equal(f.state.retry.hidden,false);assert.equal(f.rows.length,1);assert.equal(f.calls.render.length,1,'inspection renews Retry without replacing held diagnostic DOM');
 }
});
test('actual transient transport, timeout, rollover and recognized capacity keep the two-second reconnect path',async()=>{
 for(const response of [new TypeError('secret network host'),http(429,{error:'Live stream capacity reached.',reason:'capacity',retry:'automatic'}),stream([reset,{event:'gap',data:timeout}]),stream([reset,{event:'gap',data:{reason:'connection_ended',retry:'automatic'}}]),stream([])]){const f=fixture(response);await f.state.connect();assert.equal(f.state.phase,'stale');assert.equal(f.state.completedRead,null);assert.equal(f.timers.length,1);assert.equal(f.state.terminalFailure,null);assert.equal(f.rows.length,1);assert.doesNotMatch(f.state.status.textContent,/secret/)}
});
test('actual Retry rechecks native visibility, dirty scope and generation after the owner guard without clearing held rows',async()=>{
 for(const edit of [f=>f.state.retry.hidden=true,f=>f.state.retry.disabled=true,f=>f.state.retry.inert=true,f=>document.hidden=true,f=>f.state.generation++,f=>f.state.onResume=()=>false,f=>f.state.onResume=()=>{f.state.generation++;return true},f=>f.state.onResume=()=>{throw Error('refused')}]){const f=fixture(stream([reset,{event:'gap',data:missing}]));await f.state.connect();const handler=f.state.retry.onclick;edit(f);assert.doesNotThrow(handler);assert.equal(f.calls.fetch,1);assert.equal(f.rows.length,1);assert.equal(f.calls.clear,0)}
 const f=fixture(stream([reset,{event:'gap',data:missing}]));await f.state.connect();const old=f.state.retry.onclick;f.state.pauseStream();old();assert.equal(f.calls.fetch,1);document.activeElement=f.state.retry;f.state.retry.onclick();assert.equal(f.calls.fetch,2);assert.equal(f.rows.length,1);assert.equal(f.calls.clear,0);assert.equal(document.activeElement,f.state.pause);
});

test('malformed batches never replace held observations or count as a successful empty read; explicit retry replaces only on accepted batch',async()=>{
 const batch={event:'batch',data:{fromMs:60000,toMs:120000,observedAt:'1970-01-01T00:02:00.001Z',limited:false,rows:[{key:'c'.repeat(64)+':1',record:{timestamp_ns:'90000000000',Body:'new'}}]}};
 for(const edit of [b=>delete b.data.limited,b=>b.data.observedAt='invalid',b=>b.data.observedAt='1970-01-01T00:01:59Z',b=>b.data.rows[0].key='bad',b=>delete b.data.rows[0].record.timestamp_ns]){const bad=structuredClone(batch);edit(bad);const f=fixture(stream([reset,bad]));await f.state.connect();assert.equal(f.state.terminalFailure.reason,'source_malformed');assert.equal(f.calls.clear,0);assert.equal(f.rows[0].key,'held');assert.equal(f.state.completedRead,null);assert.equal(f.timers.length,0)}
 const f=fixture(stream([reset,{event:'gap',data:missing}]));await f.state.connect();let release;f.state.fetch=()=>new Promise(resolve=>release=resolve);f.state.retry.onclick();assert.equal(f.rows[0].key,'held');assert.equal(f.calls.clear,0);release(stream([reset,batch,{event:'gap',data:missing}]));await new Promise(resolve=>setImmediate(resolve));assert.equal(f.calls.clear,1);assert.equal(f.rows.length,1);assert.equal(f.rows[0].record.Body,'new');assert.equal(f.state.terminalFailure.reason,'source_missing');assert.equal(f.state.completedRead,null);
});

test('failed keyboard Retry returns only its disappearing control focus and inspection preserves held diagnostic DOM',async()=>{
 const f=fixture(stream([reset,{event:'gap',data:missing}]));await f.state.connect();f.state.failureHost.openTechnicalDetails=true;const old=f.state.retry.onclick;document.activeElement=f.state.failureHost;f.state.pauseStream('Paused for inspection.');assert.equal(f.calls.render.length,1);assert.equal(f.state.failureHost.openTechnicalDetails,true);assert.equal(document.activeElement,f.state.failureHost);old();assert.equal(f.calls.fetch,1);
 document.activeElement=f.state.retry;f.state.retry.onclick();assert.equal(document.activeElement,f.state.pause);await new Promise(resolve=>setImmediate(resolve));assert.equal(f.state.phase,'unavailable');assert.equal(document.activeElement,f.state.retry);assert.equal(f.calls.render.length,2);
 const unrelated=new Element();document.activeElement=unrelated;f.state.retry.onclick();await new Promise(resolve=>setImmediate(resolve));assert.equal(document.activeElement,unrelated,'new refusal never steals row or inspector focus');
});

test('a gap before the required reset is malformed even when its supplied retry policy is automatic',async()=>{
 for(const data of [{reason:'connection_ended',retry:'automatic'},timeout,missing]){const f=fixture(stream([{event:'gap',data}]));await f.state.connect();assert.equal(f.state.phase,'unavailable');assert.equal(f.state.terminalFailure.reason,'source_malformed');assert.equal(f.timers.length,0);assert.equal(f.rows[0].key,'held');assert.equal(f.state.completedRead,null)}
});

test('actual loaded-match selection remains local for a stopped terminal failure and rejects ordinary unavailable or live ownership',()=>{
 const sourceBody=between('  function batchState(','  function resetFind('),row={key:'held',record:{Body:'match'}},items=new Set(),calls={writes:0,reads:0,add:0};
 const node=(tag,text)=>{const e=new Element();e.tagName=tag;e.textContent=text;e.children=[];e.setAttribute=()=>{};e.append=(...children)=>e.children.push(...children);e.replaceChildren=(...children)=>e.children=children;return e};
 for(const [phase,failure,controller,enabled]of [['unavailable',missing,null,true],['paused',null,null,true],['unavailable',null,null,false],['unavailable',missing,{},false],['loading',null,{},false]]){
  items.clear();calls.add=0;const batchHost=node('div'),state={phase,terminalFailure:failure,controller,RecordPresentation:{findQuery:()=>{},matchesLoaded:()=>true},find:{value:'match'},findText:'match',columnControl:{layout:()=>({fieldColumns:[]})},buffer:{rows:[row]},signalKind:'logs',keepBusy:false,keepIntent:null,kept:null,saving:false,pendingIntent:null,outcome:null,selector:{environment:'prod'},selected:{planMany:()=>({added:items.has(row.key)?0:1}),addMany(rows){calls.add++;rows.forEach(r=>items.add(r.key));return{added:1}}},selectionGeneration:0,generation:7,connection:'a'.repeat(64),canonical,batchHost,batchSignature:'',batchRevision:0,batchButton:null,batchCommitting:false,n:node,document:{activeElement:null,hidden:false},UI:{actionVisible:e=>!e.hidden},save:{hidden:true},keep:{disabled:false},keepStatus:node('span'),refreshSelection(){},fetch(){calls.reads++}};
  vm.runInNewContext(sourceBody+'\nthis.refreshBatch=refreshBatch;this.batchState=batchState;',state);state.refreshBatch();const button=state.batchButton;assert.equal(button.disabled,!enabled);button.onclick();assert.equal(calls.add,enabled?1:0);assert.equal(calls.reads,0);assert.equal(calls.writes,0);
  if(enabled){state.refreshBatch();assert.equal(state.batchButton.disabled,true,'existing proof/membership plan prevents repeated selection');const old=button.onclick;state.generation++;state.refreshBatch();old();assert.equal(calls.add,1,'retired matching control cannot add again')}
 }
});

test('authoring blocks native tail start/retry and retained connect without reading or dropping held rows',async()=>{
 const {state,calls,rows}=fixture(http(500,{error:'Live streaming unavailable.',reason:'server_unavailable',retry:'automatic'}));
 const body=source.match(/setReadBlocked\(reason=''\)\{([^}]+)\}/)[1];
 state.saving=false;vm.runInNewContext('this.block=function(reason=""){'+body+'}',state);
 state.block('Check the expression before applying.');
 assert.equal(state.resume.disabled,true);assert.equal(state.retry.disabled,true);
 assert.equal(state.resume.title,'Check the expression before applying.');
 await state.connect();assert.equal(calls.fetch,0);assert.equal(rows.length,1);assert.equal(calls.clear,0);
 state.block('');assert.equal(state.resume.disabled,false);assert.equal(state.retry.disabled,false);
 await state.connect();assert.equal(calls.fetch,1);
});
