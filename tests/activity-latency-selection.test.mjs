import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class Element{
 constructor(tag,text=''){this.tag=tag;this.raw=text;this.children=[];this.parentElement=null;this.attributes={};this.dataset={};this.listeners=new Map();this.hidden=false;this.disabled=false;this.value=''}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}
 get textContent(){return this.raw+this.children.map(n=>n.textContent).join('')}
 set textContent(value){this.raw=value;this.replaceChildren()}
 append(...nodes){for(const n of nodes){n.remove();n.parentElement=this;this.children.push(n)}}
 replaceChildren(...nodes){for(const n of this.children)n.parentElement=null;this.children=[];this.append(...nodes)}
 remove(){if(this.parentElement)this.parentElement.children=this.parentElement.children.filter(n=>n!==this);this.parentElement=null}
 setAttribute(key,value){this.attributes[key]=String(value)}getAttribute(key){return this.attributes[key]??null}
 addEventListener(type,callback){this.listeners.set(type,[...(this.listeners.get(type)||[]),callback])}
 dispatch(type,detail={}){for(const callback of this.listeners.get(type)||[])callback({type,detail,...detail})}
 closest(selector){for(let p=this;p;p=p.parentElement)if(selector==='[inert]'&&p.inert)return p;return null}
 focus(){document.activeElement=this}
}
const walk=root=>root.children.flatMap(n=>[n,...walk(n)]),node=(tag,text='',cls='')=>{const e=new Element(tag,text);e.className=cls;return e},visible=e=>{if(!e?.isConnected||document.hidden)return false;for(let p=e;p;p=p.parentElement)if(p.hidden||p.inert)return false;return true};
const actualUI=await import('../ui/components.mjs'),actualVisibility=await import('../ui/dom-visibility.mjs');
let reads=[],views=[];
mock.module('../ui/workspace.mjs',{exports:{validateRecordQuestion:(await import('../ui/question-protocol.mjs')).validateRecordQuestion}});
mock.module('../ui/components.mjs',{exports:{...actualUI,node,disclosure:(label,children)=>{const e=node('details');e.append(...children);return e}}});
mock.module('../ui/dom-visibility.mjs',{exports:{...actualVisibility,actionVisible:visible,nativeControlReady:e=>visible(e)&&!e.disabled}});
mock.module('../ui/queries.mjs',{exports:{postJSON:(url,q,signal)=>new Promise(resolve=>reads.push({url,q:structuredClone(q),signal,resolve}))}});
mock.module('../ui/activity-baseline.mjs',{exports:{mount:()=>({pause(){},invalidate(){},pending(){},setCurrent(){},dispose(){}})}});
// The passive histogram has no read or staging owner; retain its callbacks so
// adapter leases are exercised even when a retired component calls directly.
mock.module('../ui/latency-histogram.mjs',{exports:{durationSnapshot:d=>d.histogramDataset,mount:(host,data,options)=>{views.push(options);return{invalidate(){},suspend(){}}}}});
const {mount}=await import('../ui/activity-latency.mjs');
const {attach}=await import('./support/duration-dataset.cjs');
function setup(t){
 reads=[];views=[];globalThis.document={hidden:false,activeElement:null};const root=node('main');root.root=true;const toolbar=node('div'),host=node('div');root.append(toolbar,host);
 const c={selector:{signal:'spans',environment:'prod',where:{any:[{field:{kind:'attribute',key:'tenant'},op:'missing'},{field:{kind:'attribute',key:'tenant'},op:'eq',value:''}]}},readWindow:{fromMs:120000,toMs:180000},active:true,clean:true,phase:'partial'},staged=[],comparisons=[];let afterPause=()=>{},pauses=0;
 const owner=mount(host,{controlsHost:toolbar,current:()=>c,pauseStream:()=>{pauses++;c.phase='paused';owner.sync();afterPause()},onStage:where=>staged.push(where),onCompare:(...args)=>comparisons.push(args)});t.after(()=>owner.dispose());
 const button=name=>[...walk(toolbar),...walk(host)].find(e=>e.tag==='button'&&(e.textContent===name||e.attributes['aria-label']===name));
 return {c,root,host,owner,staged,comparisons,button,section:()=>host.children[0],pauseWith:fn=>afterPause=fn,pauses:()=>pauses,select:()=>views.at(-1).onSelect({fromMs:120000,toMs:180000,minMS:'5',maxMS:'10'},{count:'2',total:'3'}),async open(){button('Show latency').onclick();const r=reads.at(-1);r.resolve({ok:true,data:attach({question:r.q,basis:'basis',capturedAt:'2026-11-12T12:00:00Z',count:'3',state:'observed',groups:[{g0:'[5,10) ms',count:'2'},{g0:'[1000,+inf) ms',count:'1'}]})});await new Promise(resolve=>setImmediate(resolve))}};
}
test('latency selection holds its exact parent then stages once; Compare remains a local held handoff',async t=>{
 const f=setup(t);await f.open();assert.equal(views.length,1);f.select();assert.equal(f.staged.length,1);assert.deepEqual(f.staged[0].all[0],f.c.selector.where);f.button('Compare selected latency with rest').onclick();assert.equal(f.comparisons.length,1);assert.equal(f.comparisons[0][0].fromMs,120000);assert.equal(f.comparisons[0][1].toMs,180000);assert.equal(reads.length,1);
});
for(const state of ['dirty','hidden','disabled','disposed','scope','renderABA'])test('latency selection withdraws when hold changes '+state,async t=>{
 const f=setup(t);await f.open();f.pauseWith(()=>{if(state==='dirty')f.c.clean=false;if(state==='hidden')f.section().hidden=true;if(state==='disabled')f.host.children[0].children[1].disabled=true;if(state==='disposed')f.owner.dispose();if(state==='scope'){f.c.selector.environment='other';f.owner.sync()}if(state==='renderABA'){f.c.active=false;f.owner.sync();f.c.active=true;f.owner.sync()}});f.select();assert.equal(f.staged.length,0);assert.equal(reads.length,1);
});
for(const state of ['dirty','hidden','disabled','disposed','scope','renderABA'])test('latency Compare withdraws when hold changes '+state,async t=>{
 const f=setup(t);await f.open();f.select();const compare=f.button('Compare selected latency with rest');f.pauseWith(()=>{if(state==='dirty')f.c.clean=false;if(state==='hidden')f.section().hidden=true;if(state==='disabled')compare.disabled=true;if(state==='disposed')f.owner.dispose();if(state==='scope'){f.c.selector.environment='other';f.owner.sync()}if(state==='renderABA'){f.c.active=false;f.owner.sync();f.c.active=true;f.owner.sync()}});compare.onclick();assert.equal(f.comparisons.length,0);assert.equal(reads.length,1);
});
test('inert latency callbacks do not even pause or read',async t=>{const f=setup(t);await f.open();const old=views[0].onSelect,pauses=f.pauses();f.root.inert=true;old({},{});assert.equal(f.pauses(),pauses);assert.equal(f.staged.length,0);assert.equal(reads.length,1)});
test('latency Compare retains the already staged dirty filter without applying or rereading',async t=>{const f=setup(t);await f.open();f.select();f.c.clean=false;f.owner.sync();f.button('Compare selected latency with rest').onclick();assert.equal(f.comparisons.length,1);assert.equal(reads.length,1)});

test('return to held latency disables the retired Compare without pausing or reacquiring',async t=>{
 const f=setup(t);await f.open();f.select();const compare=f.button('Compare selected latency with rest'),retired=compare.onclick;
 assert.equal(compare.disabled,false);const pauses=f.pauses();f.c.active=false;f.owner.sync();assert.equal(compare.disabled,true);f.c.active=true;f.owner.sync();assert.equal(compare.disabled,true);retired();assert.equal(f.comparisons.length,0);assert.equal(f.pauses(),pauses);assert.equal(reads.length,1);assert.equal(f.section().dataset.fromMs,'120000');assert.equal(f.section().dataset.toMs,'180000');
});

test('Compare rejects a foreign provider scope before sync without pausing or reading',async t=>{
 const f=setup(t);await f.open();f.select();const compare=f.button('Compare selected latency with rest'),retired=compare.onclick,pauses=f.pauses();f.c.selector.environment='foreign';retired();assert.equal(f.comparisons.length,0);assert.equal(f.pauses(),pauses);assert.equal(reads.length,1);f.owner.sync();assert.equal(compare.disabled,true);
});
