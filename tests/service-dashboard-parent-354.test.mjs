import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class Element{
 constructor(tag){this.tag=tag;this.children=[];this.parent=null;this.hidden=false;this.inert=false;this.dataset={};this.value='';this.classList={add(){}}}
 get isConnected(){return this.root||!!this.parent?.isConnected}
 append(...nodes){for(const n of nodes){n.parent=this;this.children.push(n)}}
 insertBefore(node,before){const index=this.children.indexOf(before);node.parent=this;if(index<0)this.children.push(node);else this.children.splice(index,0,node)}
 replaceChildren(...nodes){for(const n of this.children)n.parent=null;this.children=[];this.append(...nodes)}
 setAttribute(k,v){this[k]=v}
 closest(selector){for(let n=this;n;n=n.parent){if(selector==='[inert]'?n.inert:n.hidden)return n}return null}
 getClientRects(){return this.isConnected&&!this.closest('[hidden]')?[{}]:[]}
 focus(){document.activeElement=this}
 scrollIntoView(){}
 querySelector(selector){return this.children.find(n=>selector==='.service-loading'&&n.className==='service-loading')||null}
 remove(){if(this.parent){this.parent.children=this.parent.children.filter(n=>n!==this);this.parent=null}}
}
const requests=[],loads=[];
mock.module('../ui/resource-return.mjs',{exports:{mount:()=>({clear(){},show(){}}),restoreFocus(){}}});
mock.module('../ui/deployments.mjs',{exports:{mount:()=>({setQuestion(){}})}});
mock.module('../ui/activity-views.mjs',{exports:{library:()=>({dispose(){}})}});
mock.module('../ui/meaning.mjs',{exports:{mount(){}}});
mock.module('../ui/service-dashboard.mjs',{exports:{httpServerObserved:()=>false,rpcServerObserved:()=>false,dbClientObserved:()=>false,messagingConsumerObserved:()=>false,messagingProducerObserved:()=>false,nativePicker:()=>null,load:(_host,input,options)=>new Promise(resolve=>loads.push({input,options,resolve}))}});
mock.module('../ui/queries.mjs',{exports:{postJSON:(url,q,signal)=>new Promise((resolve,reject)=>requests.push({url,q:structuredClone(q),signal,resolve,reject}))}});
mock.module('../ui/visualizations.mjs',{exports:{actionVisible:el=>!!el?.isConnected&&!el.closest('[hidden]')}});
const {mount}=await import('../ui/service-browser.mjs');
const Handoff=await import('../ui/service-handoff.mjs');
function setup(t){const events=[],previous={document:globalThis.document,dispatchEvent:globalThis.dispatchEvent};globalThis.document={createElement:tag=>new Element(tag),querySelector:()=>null};globalThis.dispatchEvent=event=>events.push(event);const host=new Element('section');host.root=true;const catalog=new Element('div');host.append(catalog);const browser=mount(host,{catalog});requests.length=0;t.after(()=>{browser.invalidate();globalThis.document=previous.document;globalThis.dispatchEvent=previous.dispatchEvent});return {host,browser,events,detail:host.children.at(-1),buttons:()=>host.children.at(-1).children.filter(n=>n.tag==='button')}}

const tick=()=>new Promise(resolve=>setImmediate(resolve));
test('browser keeps a retry pending when initial dashboard load returns and retires it on suspend',async t=>{const f=setup(t);loads.length=0;const scope={environment:'prod',fromMs:1000,toMs:61000};f.browser.setScope(scope);const opened=f.browser.open('api',new Element('button'));for(const request of requests){if(request.url==='/api/discovery')request.resolve({ok:true,data:{signals:[]}});else request.reject(Error('Independent supporting read unavailable'))}await tick();assert.equal(loads.length,1);const load=loads[0];load.options.onPendingChange(true);load.resolve({plan:{},results:[]});await opened;assert.equal(load.options.signal.aborted,false);assert.equal(f.browser.suspend(),true);assert.equal(load.options.signal.aborted,true);load.options.onPendingChange(false);assert.equal(f.browser.suspend(),false);});
test('old dashboard pending finalizers cannot clear a new service owner',async t=>{const f=setup(t);loads.length=0;f.browser.setScope({environment:'prod',fromMs:1000,toMs:61000});const open=f.browser.open('api',new Element('button'));for(const request of requests){if(request.url==='/api/discovery')request.resolve({ok:true,data:{signals:[]}});else request.reject(Error('Independent supporting read unavailable'))}await tick();const old=loads[0];old.options.onPendingChange(true);old.resolve({});await open;const later=f.browser.open('worker',new Element('button')),next=requests.slice(4);old.options.onPendingChange(false);assert.equal(f.browser.suspend(),true);for(const read of next)read.reject(Error('Retired'));await later;});
