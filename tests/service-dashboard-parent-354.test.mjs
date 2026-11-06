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
 addEventListener(){}
 querySelector(selector){return this.children.find(n=>selector==='.service-loading'&&n.className==='service-loading')||null}
 remove(){if(this.parent){this.parent.children=this.parent.children.filter(n=>n!==this);this.parent=null}}
}
const {rpcServerCurrentObserved,matches}=await import('../ui/service-dashboard.mjs');
const requests=[],loads=[];
mock.module('../ui/resource-return.mjs',{exports:{mount:()=>({clear(){},show(){}}),restoreFocus(){}}});
mock.module('../ui/deployments.mjs',{exports:{mount:()=>({setQuestion(){}})}});
mock.module('../ui/activity-views.mjs',{exports:{library:()=>({dispose(){}})}});
mock.module('../ui/meaning.mjs',{exports:{mount(){}}});
mock.module('../ui/service-dashboard.mjs',{exports:{matches,httpServerObserved:()=>false,rpcServerObserved:()=>false,rpcServerCurrentObserved,dbClientObserved:()=>false,messagingConsumerObserved:()=>false,messagingProducerObserved:()=>false,nativePicker:()=>null,load:(_host,input,options)=>new Promise(resolve=>loads.push({input,options,resolve}))}});
mock.module('../ui/queries.mjs',{exports:{postJSON:(url,q,signal)=>new Promise((resolve,reject)=>requests.push({url,q:structuredClone(q),signal,resolve,reject}))}});
mock.module('../ui/visualizations.mjs',{exports:{actionVisible:el=>!!el?.isConnected&&!el.closest('[hidden]')}});
const {mount}=await import('../ui/service-browser.mjs');
const Handoff=await import('../ui/service-handoff.mjs');
function setup(t){const events=[],previous={document:globalThis.document,dispatchEvent:globalThis.dispatchEvent};globalThis.document={createElement:tag=>new Element(tag),querySelector:()=>null};globalThis.dispatchEvent=event=>events.push(event);const host=new Element('section');host.root=true;const catalog=new Element('div');host.append(catalog);const browser=mount(host,{catalog});requests.length=0;t.after(()=>{browser.invalidate();globalThis.document=previous.document;globalThis.dispatchEvent=previous.dispatchEvent});return {host,browser,events,detail:host.children.at(-1),buttons:()=>host.children.at(-1).children.filter(n=>n.tag==='button')}}

const tick=()=>new Promise(resolve=>setImmediate(resolve));
test('browser keeps a retry pending when initial dashboard load returns and retires it on suspend',async t=>{const f=setup(t);loads.length=0;const scope={environment:'prod',fromMs:1000,toMs:61000};f.browser.setScope(scope);const opened=f.browser.open('api',new Element('button'));for(const request of requests){if(request.url==='/api/discovery')request.resolve({ok:true,data:{signals:[]}});else request.reject(Error('Independent supporting read unavailable'))}await tick();assert.equal(loads.length,1);const load=loads[0];load.options.onPendingChange(true);load.resolve({plan:{},results:[]});await opened;assert.equal(load.options.signal.aborted,false);assert.equal(f.browser.suspend(),true);assert.equal(load.options.signal.aborted,true);load.options.onPendingChange(false);assert.equal(f.browser.suspend(),false);});
test('old dashboard pending finalizers cannot clear a new service owner',async t=>{const f=setup(t);loads.length=0;f.browser.setScope({environment:'prod',fromMs:1000,toMs:61000});const open=f.browser.open('api',new Element('button'));for(const request of requests){if(request.url==='/api/discovery')request.resolve({ok:true,data:{signals:[]}});else request.reject(Error('Independent supporting read unavailable'))}await tick();const old=loads[0];old.options.onPendingChange(true);old.resolve({});await open;const later=f.browser.open('worker',new Element('button')),next=requests.slice(4);old.options.onPendingChange(false);assert.equal(f.browser.suspend(),true);for(const read of next)read.reject(Error('Retired'));await later;});


test('service template delegation keeps fixed scope, stages only, and retires hidden or replaced openers',async t=>{
 const f=setup(t);loads.length=0;const scope={environment:'prod',fromMs:1000,toMs:61000};f.browser.setScope(scope);const opening=f.browser.open('api',new Element('button'));for(const read of requests){if(read.url==='/api/discovery')read.resolve({ok:true,data:{signals:[]}});else read.reject(Error('Independent context unavailable'))}await tick();const load=loads[0];load.options.onPendingChange(false);load.resolve({});await opening;const origin=new Element('button');f.detail.append(origin);const q={signal:'spans',environment:'prod',fromMs:1000,toMs:61000,summary:'span-duration',spanStatus:true,bucketMs:1000,groupBy:[{kind:'column',key:'name'}],limit:100,offset:0,where:{field:{kind:'column',key:'service'},op:'eq',value:'api'}},before=requests.length;
 load.options.onUseTemplate(q,origin);const event=f.events.find(e=>e.type==='goo:open-population');assert(event);assert.deepEqual(event.detail,{question:q,templates:true,stage:true});assert.equal(requests.length,before);q.where.value='changed';assert.equal(event.detail.question.where.value,'api');const count=f.events.length;origin.hidden=true;load.options.onUseTemplate(q,origin);assert.equal(f.events.length,count);origin.hidden=false;f.browser.invalidate();load.options.onUseTemplate(q,origin);assert.equal(f.events.length,count);
});


test('current RPC auto selection survives a narrowed window without a new positive witness or hidden fallback',async t=>{
 const f=setup(t);loads.length=0;const scope={environment:'prod',fromMs:1000,toMs:61000};f.browser.setScope(scope);const opening=f.browser.open('api',new Element('button'));for(const read of requests){if(read.url==='/api/discovery')read.resolve({ok:true,data:{signals:[{service:'api',signal:'spans',count:'1'}]}});else if(read.url==='/api/explore')read.resolve({ok:true,data:{state:'observed',question:structuredClone(read.q),count:'1',fields:[{kind:'attribute',key:'rpc.system.name'}],rows:[{ServiceName:'api',SpanKind:'Server',timestamp_ns:'1500000000',ResourceAttributes:{'deployment.environment.name':'prod'},Attributes:{'rpc.system.name':'grpc'}}]}});else read.reject(Error('No native source observed'))}await tick();const initial=loads[0];assert.equal(initial.input.rpcServerCurrentObserved,true);initial.options.onPlan({profile:'rpc-server-current'});initial.resolve({});await opening;const requested=requests.length,later=initial.options.onRange({scope:{...scope,fromMs:1600}});await tick();assert.equal(loads.length,2);assert.equal(loads[1].input.profile,'rpc-server-current');assert.equal(loads[1].input.rpcServerCurrentObserved,false,'held row is outside the narrowed window');assert.equal(requests.length,requested,'profile hints must not probe again');loads[1].resolve({});await later;
});
