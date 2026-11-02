import {test,mock} from 'node:test';
import assert from 'node:assert/strict';

// Exercise the application owner with supplied renderers. These fixtures test
// request lifetime/publication, not source response validation or chart painting.
const mounts=[];let waterfallHook=null;
class Element {
 constructor(tag,text){this.tag=tag;this.textContent=text??'';this.children=[];this.parent=null;this.hidden=false;this.open=false;this.disabled=false;this.value='';this.dataset={};this.attributes=new Map();this.classList={add(){},remove(){}};}
 get isConnected(){return this.root===true||!!this.parent?.isConnected;}
 get ownerDocument(){return globalThis.document;}
 append(...nodes){for(const node of nodes){node.parent=this;this.children.push(node)}}
 replaceChildren(...nodes){for(const node of this.children)node.parent=null;this.children=[];this.append(...nodes)}
 before(node){const i=this.parent.children.indexOf(this);node.parent=this.parent;this.parent.children.splice(i,0,node)}
 replaceWith(node){const parent=this.parent;if(!parent)return;const i=parent.children.indexOf(this);parent.children.splice(i,1,node);node.parent=parent;this.parent=null;}
 remove(){if(this.parent){this.parent.children=this.parent.children.filter(x=>x!==this);this.parent=null}}
 set innerHTML(value){this.replaceChildren();for(const match of value.matchAll(/<(\w+)[^>]*\bid="([^"]+)"[^>]*>/g)){const el=new Element(match[1]);el.id=match[2];el.hidden=/\bhidden\b/.test(match[0]);el.disabled=/\bdisabled\b/.test(match[0]);el.value=el.id==='exploreLogScope'?'trace':el.id==='surroundingRadius'?'30000':'';this.append(el)}}
 querySelector(selector){const id=selector.startsWith('#')?selector.slice(1):null;for(const child of this.children){if(child.id===id)return child;const found=child.querySelector(selector);if(found)return found}return null}
 querySelectorAll(){return []}
 setAttribute(k,v){this.attributes.set(k,v)}
 addEventListener(){}
 closest(){return null}
 contains(element){return element===this||this.children.some(child=>child.contains(element))}
 getClientRects(){return this.isConnected?[{}]:[]}
 focus(){document.activeElement=this}
 showModal(){this.open=true}
 close(){this.open=false}
 scrollIntoView(){}
}
mock.module('../ui/components.mjs',{exports:{node:(tag,text)=>new Element(tag,text)}});
mock.module('../ui/namespace.mjs',{exports:{namespace(){}}});
mock.module('../ui/record-interactions.mjs',{exports:{mount(...args){mounts.push(args);return {}}}});
mock.module('../ui/scoped-breakdown.mjs',{exports:{mount(){return {suspendActions(){}}}}});
mock.module('../ui/trace-comparison-view.mjs',{exports:{}});
mock.module('../ui/distributions.mjs',{exports:{durationMS:()=>''}});
mock.module('../ui/trace-journey.mjs',{exports:{finding(){}}});
mock.module('../ui/visualizations.mjs',{exports:{actionVisible:()=>true,breakdown(){},waterfall(){waterfallHook?.()}}});
const {mount}=await import('../ui/record-results.mjs');
const origin={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0};
const trace='0123456789abcdef0123456789abcdef';
const flush=()=>new Promise(resolve=>setImmediate(resolve));
function setup(t){
 const keys=['document','CSS','addEventListener','removeEventListener'],previous=new Map(keys.map(key=>[key,globalThis[key]])),events=new Map();
 globalThis.document={activeElement:null,createElement:tag=>new Element(tag),getElementById:()=>null};globalThis.CSS={escape:x=>x};
 globalThis.addEventListener=(type,fn)=>{if(!events.has(type))events.set(type,new Set());events.get(type).add(fn)};
 globalThis.removeEventListener=(type,fn)=>events.get(type)?.delete(fn);
 const container=new Element('section');container.root=true;
 const requests=[];let active=true,returns=0;
 const owner=mount(container,{active:()=>active,returnFocus:()=>returns++,fetchQuery:(q,signal)=>new Promise((resolve,reject)=>requests.push({q,signal,resolve,reject}))});
 const el=id=>container.querySelector('#'+id);
 const complete=(index,{more=false}={})=>{const r=requests[index];r.resolve({question:structuredClone(r.q),rows:[],count:'0',state:'no_data',basis:'held-'+index,capturedAt:'2026-11-02T00:00:00Z',hasMore:more,warnings:[]})};
 t.after(()=>{owner.dispose();waterfallHook=null;mounts.length=0;for(const [key,value]of previous){if(value===undefined)delete globalThis[key];else globalThis[key]=value}});
 return {owner,requests,el,complete,container,get returns(){return returns},withdraw:async()=>{active=false;for(const fn of events.get('goo:workspace')||[])fn();await flush()},openSurrounding:()=>{owner.table(new Element('div'),[],'logs',origin,'origin');mounts.at(-1).at(-1).openSurrounding(origin,{ServiceName:'api',timestamp_ns:'1500000000'},'origin',null)}};
}

test('trace close aborts its read and a late successful response cannot start linked logs or publish',async t=>{
 const x=setup(t),pending=x.owner.openTrace(origin,trace);assert.equal(x.requests.length,1);assert.ok(x.requests[0].signal instanceof AbortSignal);
 x.el('exploreContextClose').onclick();assert.equal(x.requests[0].signal.aborted,true);assert.equal(x.returns,1);
 x.complete(0);await pending;assert.equal(x.requests.length,1);assert.equal(x.el('exploreContext').hidden,true);assert.equal(mounts.length,0);assert.doesNotMatch(x.el('exploreContextStatus').textContent,/unavailable/);
});

test('a replacement trace aborts its predecessor and keeps the new trace/log request lifetimes separate',async t=>{
 const x=setup(t),first=x.owner.openTrace(origin,trace),second=x.owner.openTrace(origin,'f'.repeat(32));
 assert.equal(x.requests[0].signal.aborted,true);assert.equal(x.requests[1].signal.aborted,false);
 x.complete(0);await first;assert.equal(x.requests.length,2);
 x.complete(1);await flush();assert.equal(x.requests.length,3,x.el('exploreContextStatus').textContent);assert.equal(x.requests[1].signal.aborted,false);assert.notEqual(x.requests[1].signal,x.requests[2].signal);
 x.el('exploreContextClose').onclick();assert.equal(x.requests[2].signal.aborted,true);x.complete(2);await second;assert.equal(mounts.length,0);assert.equal(x.requests.length,3);
});

test('log scope replacement and trace paging abort only the retired dependent work',async t=>{
 const x=setup(t),opening=x.owner.openTrace(origin,trace);x.complete(0,{more:true});await flush();
 const originalLog=x.requests[1];const replacement=x.el('exploreLogRun').onclick();assert.equal(originalLog.signal.aborted,true);assert.equal(x.requests[2].signal.aborted,false);
 originalLog.resolve({});await flush();assert.equal(mounts.length,0);
 const page=x.el('exploreTraceMore').onclick();assert.equal(x.requests[2].signal.aborted,true);assert.equal(x.requests[3].q.offset,100);
 x.complete(2);await replacement;x.owner.invalidate();assert.equal(x.requests[3].signal.aborted,true);x.complete(3);await page;await opening;
 assert.equal(x.requests.length,4);assert.equal(mounts.length,0);
});

test('surrounding context replacement, close and navigation abort without new reads',async t=>{
 const x=setup(t);x.openSurrounding();assert.equal(x.requests.length,1);assert.equal(x.requests[0].q.signal,'logs');
 x.el('surroundingRadius').value='60000';const replacement=x.el('surroundingRadius').onchange();assert.equal(x.requests[0].signal.aborted,true);
 x.complete(0);await flush();assert.equal(mounts.length,1); // only the explicit host used to open surrounding context
 x.el('surroundingClose').onclick();assert.equal(x.requests[1].signal.aborted,true);x.complete(1);await replacement;assert.equal(mounts.length,1);
 x.openSurrounding();await x.withdraw();assert.equal(x.requests[2].signal.aborted,true);x.complete(2);await flush();assert.equal(x.requests.length,3);assert.equal(x.el('surroundingContext').hidden,true);
});

test('dispose aborts pending work; retired controls cannot reopen a request',async t=>{
 const x=setup(t),pending=x.owner.openTrace(origin,trace),run=x.el('exploreLogRun').onclick;
 x.owner.dispose();assert.equal(x.requests[0].signal.aborted,true);x.complete(0);await pending;await run();await x.owner.openTrace(origin,trace);assert.equal(x.requests.length,1);
});

test('retirement during trace rendering does not launch the following linked-log request',async t=>{
 const x=setup(t),pending=x.owner.openTrace(origin,trace);waterfallHook=()=>x.owner.hide();x.complete(0);await pending;assert.equal(x.requests.length,1);assert.equal(x.el('exploreContext').hidden,true);
});

test('ordinary completion preserves explicit close focus/return handling and makes no retries',async t=>{
 const x=setup(t);let returns=0;const pending=x.owner.openTrace(origin,trace,'',null,null,{onClose:()=>{returns++;return true}});x.complete(0);await flush();x.complete(1);await pending;
 assert.equal(x.requests.length,2);assert.equal(mounts.length,1);assert.equal(x.requests.every(r=>!r.signal.aborted),true);
 x.el('exploreContextClose').onclick();assert.equal(returns,1);assert.equal(x.returns,0);assert.equal(x.requests.length,2);
});
