import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class Element{
 constructor(tag){this.tag=tag;this.children=[];this.parent=null;this.hidden=false;this.inert=false;this.dataset={};this.classList={add(){}}}
 get isConnected(){return this.root||!!this.parent?.isConnected}
 append(...nodes){for(const n of nodes){n.parent=this;this.children.push(n)}}
 replaceChildren(...nodes){for(const n of this.children)n.parent=null;this.children=[];this.append(...nodes)}
 setAttribute(k,v){this[k]=v}
 closest(selector){for(let n=this;n;n=n.parent){if(selector==='[inert]'?n.inert:n.hidden)return n}return null}
 getClientRects(){return this.isConnected&&!this.closest('[hidden]')?[{}]:[]}
 focus(){document.activeElement=this}
 scrollIntoView(){}
 querySelector(selector){return this.children.find(n=>selector==='.service-loading'&&n.className==='service-loading')||null}
 remove(){if(this.parent){this.parent.children=this.parent.children.filter(n=>n!==this);this.parent=null}}
}
const requests=[];
mock.module('../ui/resource-return.mjs',{exports:{mount:()=>({clear(){},show(){}}),restoreFocus(){}}});
mock.module('../ui/deployments.mjs',{exports:{mount:()=>({setQuestion(){}})}});
mock.module('../ui/activity-views.mjs',{exports:{library:()=>({dispose(){}})}});
mock.module('../ui/meaning.mjs',{exports:{mount(){}}});
mock.module('../ui/queries.mjs',{exports:{postJSON:(url,q,signal)=>new Promise((resolve,reject)=>requests.push({url,q:structuredClone(q),signal,resolve,reject}))}});
mock.module('../ui/visualizations.mjs',{exports:{actionVisible:el=>!!el?.isConnected&&!el.closest('[hidden]')}});
const {mount}=await import('../ui/service-browser.mjs');
const Handoff=await import('../ui/service-handoff.mjs');
function setup(t){const previous={document:globalThis.document,dispatchEvent:globalThis.dispatchEvent};globalThis.document={createElement:tag=>new Element(tag),querySelector:()=>null};globalThis.dispatchEvent=()=>{};const host=new Element('section');host.root=true;const catalog=new Element('div');host.append(catalog);const browser=mount(host,{catalog});requests.length=0;t.after(()=>{browser.invalidate();globalThis.document=previous.document;globalThis.dispatchEvent=previous.dispatchEvent});return {host,browser,detail:host.children.at(-1),buttons:()=>host.children.at(-1).children.filter(n=>n.tag==='button')}}
// Uses the same pure constructor as the source row adapter; source data remains
// supplied here because this test isolates destination ownership and no-I/O.
function intent(){return Handoff.create({signal:'logs',environment:'prod',fromMs:100000,toMs:160000,limit:100,offset:0,where:{field:{kind:'column',key:'body'},op:'eq',value:'needle'}},{ServiceName:'api',timestamp_ns:'120000000000',ResourceAttributes:{'deployment.environment.name':'prod','service.namespace':''}})}

test('staging is detached and source-free; named Back returns once without a read',t=>{const f=setup(t);let returns=0;const value=intent();f.browser.stageService(value,{onBack:()=>returns++});assert.equal(requests.length,0);assert.equal(f.detail['aria-label'],'Service handoff');assert.match(f.detail.children[2].textContent,/current key only/);const filters=f.detail.children.find(n=>n.tag==='details');assert.equal(filters.children[0].textContent,'Activity filters are not carried');assert.equal(filters.children[1].textContent,'Origin signal: logs');assert.equal(filters.children[2].textContent,'Log message = \"needle\"');const [back,read]=f.buttons();assert.equal(back.textContent,'Back to paused Activity');assert.equal(read.textContent,'Read service overview');back.onclick();back.onclick();read.onclick();assert.equal(returns,1);assert.equal(requests.length,0)});
test('only explicit Read starts exact scoped overview; Back cancels all pending reads',t=>{const f=setup(t);let returns=0;f.browser.stageService(intent(),{onBack:()=>returns++});const [oldBack,read]=f.buttons();read.onclick();assert.equal(requests.length,4);for(const r of requests){assert.equal(r.q.environment,'prod');if(r.url==='/api/explore'){assert.deepEqual(r.q.where,{all:[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''}]})}else assert.deepEqual(r.q.namespace,{mode:'value',value:''});assert.equal(r.q.fromMs,100000);assert.equal(r.q.toMs,160000);assert.equal(r.signal.aborted,false)}oldBack.onclick();assert.equal(returns,0);f.buttons()[0].onclick();assert.equal(returns,1);assert(requests.every(r=>r.signal.aborted));read.onclick();assert.equal(requests.length,4)});
test('retired, hidden, inert, detached and withdrawn stage controls cannot acquire or return',t=>{const f=setup(t);let returns=0,current=true;f.browser.stageService(intent(),{onBack:()=>returns++,current:()=>current});const [back,read]=f.buttons();for(const k of ['hidden','inert']){f.host[k]=true;read.onclick();back.onclick();f.host[k]=false}current=false;read.onclick();back.onclick();current=true;f.browser.stageService(intent(),{onBack:()=>returns++});read.onclick();back.onclick();assert.equal(requests.length,0);assert.equal(returns,0);f.host.root=false;f.buttons()[0].onclick();f.buttons()[1].onclick();assert.equal(requests.length,0);assert.equal(returns,0)});
test('replacement and navigation suspend cancel the old destination read',t=>{const f=setup(t);f.browser.stageService(intent(),{onBack(){}});f.buttons()[1].onclick();const old=requests.slice();f.browser.stageService(intent(),{onBack(){}});assert(old.every(r=>r.signal.aborted));f.buttons()[1].onclick();f.browser.suspend();assert(requests.every(r=>r.signal.aborted))});

test('Cancel service reads preserves a source-free staged return owner',t=>{const f=setup(t);let returns=0;f.browser.stageService(intent(),{onBack:()=>returns++});f.buttons()[1].onclick();const cancel=f.buttons().find(n=>n.textContent==='Cancel service reads');cancel.onclick();assert(requests.every(r=>r.signal.aborted));assert.equal(f.detail['aria-label'],'Service handoff');assert.equal(requests.length,4);f.buttons()[0].onclick();assert.equal(returns,1);assert.equal(requests.length,4)});

test('completed service reads retain named Activity Back without another acquisition',async t=>{const f=setup(t);let returns=0;const value=structuredClone(intent());f.browser.stageService(value,{onBack:()=>returns++});value.service='foreign';f.buttons()[1].onclick();assert.equal(requests[1].q.where.all[0].value,'api');for(const r of requests)r.reject(Error('Unavailable supplied source'));await new Promise(resolve=>setImmediate(resolve));assert.equal(f.buttons()[0].textContent,'Back to paused Activity');assert.equal(requests.length,4);f.buttons()[0].onclick();assert.equal(returns,1);assert.equal(requests.length,4)});
