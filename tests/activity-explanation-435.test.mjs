import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class Element{
 constructor(tag,text=''){this.tag=tag;this.raw=text;this.children=[];this.parentElement=null;this.attributes={};this.dataset={};this.listeners=new Map();this.hidden=false;this.disabled=false;this.value=''}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}get textContent(){return this.raw+this.children.map(n=>n.textContent).join('')}set textContent(value){this.raw=value;this.replaceChildren()}
 append(...nodes){for(const n of nodes){n.remove();n.parentElement=this;this.children.push(n)}}replaceChildren(...nodes){for(const n of this.children)n.parentElement=null;this.children=[];this.append(...nodes)}remove(){if(this.parentElement)this.parentElement.children=this.parentElement.children.filter(n=>n!==this);this.parentElement=null}
 setAttribute(key,value){this.attributes[key]=String(value)}getAttribute(key){return this.attributes[key]??null}addEventListener(type,callback){this.listeners.set(type,[...(this.listeners.get(type)||[]),callback])}dispatch(type,detail={}){for(const callback of this.listeners.get(type)||[])callback({type,detail,...detail})}
 closest(selector){for(let p=this;p;p=p.parentElement)if(selector==='[inert]'&&p.inert)return p;return null}focus(){document.activeElement=this}scrollIntoView(){}
}
const walk=root=>root.children.flatMap(n=>[n,...walk(n)]),node=(tag,text='',cls='')=>{const e=new Element(tag,text);e.className=cls;return e},visible=e=>{if(!e?.isConnected||document.hidden)return false;for(let p=e;p;p=p.parentElement)if(p.hidden||p.inert)return false;return true};
let reads=[],children=[];
mock.module('../ui/components.mjs',{exports:{node,actionVisible:visible}});
mock.module('../ui/time-bars.mjs',{exports:{}});
mock.module('../ui/dom-visibility.mjs',{exports:{actionVisible:visible,nativeControlReady:e=>visible(e)&&!e.disabled}});
mock.module('../ui/queries.mjs',{exports:{postJSON:(url,q,signal)=>new Promise(resolve=>reads.push({url,q:structuredClone(q),signal,resolve}))}});
mock.module('../ui/attributes.mjs',{exports:{suggestionData:()=>({fields:[]})}});
mock.module('../ui/explanation.mjs',{exports:{mount:(_host,options)=>{const child={options,opened:false,invalidated:0,disposed:0};children.push(child);return {open(){child.opened=true},invalidate(){child.invalidated++},dispose(){child.disposed++}}}}});
const Volume=await import('../ui/activity-log-context.mjs'),Explanation=await import('../ui/activity-explanation.mjs');
const summary=(count,band)=>Object.fromEntries(['count','trace','debug','info','warn','error','fatal','unspecified','invalid'].map(key=>[key,key==='count'||key===band?String(count):'0']));
function response(q){const receipt={basis:'a'.repeat(64),readAt:'2026-11-12T12:00:00Z'};return {version:'goo/record-aggregates/v1',question:structuredClone(q),capturedAt:receipt.readAt,basis:'b'.repeat(64),state:'observed',count:'3',logSeverity:{...summary('3'),info:'1',error:'2'},groups:[{g0:'FATAL',count:'1',logSeverity:summary('1','info')},{g0:'ERROR',count:'2',logSeverity:summary('2','error')}],warnings:[],reads:{count:receipt,groups:receipt,timeSeries:receipt},timeSeries:[2,0,0,0,0,1].map((count,i)=>({fromMs:q.fromMs+i*10000,toMs:q.fromMs+(i+1)*10000,count:String(count),state:count?'observed':'no_data',logSeverity:summary(count,i===5?'info':'error')}))}}
function setup(t,{onPause}={}){
 reads=[];children=[];globalThis.document={hidden:false,activeElement:null};const root=node('main');root.root=true;const toolbar=node('div'),volumeHost=node('div'),explanationHost=node('div');root.append(toolbar,volumeHost,explanationHost);
 const c={selector:{signal:'logs',environment:'prod',where:{any:[{field:{kind:'attribute',key:'tenant'},op:'missing'},{field:{kind:'attribute',key:'tenant'},op:'eq',value:''}]}},readWindow:{fromMs:120000,toMs:180000},rows:[17,9,20,undefined,'17.0'].map(SeverityNumber=>({record:{timestamp_ns:'120000000000',SeverityNumber}})),active:true,clean:true,phase:'partial'};
 let explanation,volume,lastPinned;const pause=()=>{c.phase='paused';onPause?.(c);volume?.sync();explanation?.sync()};
 explanation=Explanation.mount(explanationHost,{current:()=>c,pauseStream:pause});volume=Volume.mount(volumeHost,{controlsHost:toolbar,current:()=>c,pauseStream:pause,onExplain:(value,opener)=>{lastPinned=value;return explanation.open(value,opener)}});t.after(()=>{volume.dispose();explanation.dispose()});
 const all=()=>[...walk(toolbar),...walk(volumeHost)],button=name=>all().find(e=>e.tag==='button'&&(e.textContent===name||e.attributes['aria-label']===name));
 return {c,volume,explanation,button,band:()=>all().find(e=>e.attributes['aria-label']==='Numeric severity'),chart:()=>all().find(e=>e.tag==='goo-time-bars'),explanationPanel:()=>explanationHost.children[0],guide:()=>walk(explanationHost).find(e=>e.tag==='goo-time-bars'),lastPinned:()=>lastPinned,async open(){const work=button('Show volume').onclick();reads[0].resolve({ok:true,data:response(reads[0].q)});await work}};
}
test('actual volume-to-explanation integration carries only the declared numeric refinement and remains active across held sync',async t=>{
 const f=setup(t);await f.open();f.band().value='error';f.band().onchange();f.chart().dispatch('goo-time-select',{key:'0'});f.button('Explain interval').onclick();
 assert.equal(f.lastPinned().logSeverityBand,'error');assert.equal(f.explanationPanel().hidden,false);assert.equal(f.explanationPanel().dataset.severityBand,'error');assert.equal(children.length,1);assert.equal(children[0].opened,true);assert.equal(children[0].options.canAct(),true);assert.deepEqual(children[0].options.request.selection,f.lastPinned().selection);assert.deepEqual(children[0].options.request.baseline.where,f.lastPinned().parent.where);assert.deepEqual(f.guide().data.points.map(p=>p.value),['2','0','0','0','0','0']);assert.equal(reads.length,1);
 f.c.readWindow={fromMs:130000,toMs:190000};f.explanation.sync();assert.equal(f.explanationPanel().hidden,false);assert.equal(children[0].options.canAct(),true);f.c.active=false;f.explanation.sync();assert.equal(f.explanationPanel().hidden,true);assert.equal(children[0].options.canAct(),false);f.c.active=true;f.explanation.sync();assert.equal(f.explanationPanel().hidden,false);assert.equal(children[0].options.canAct(),true);
 f.c.selector.where={field:{kind:'attribute',key:'tenant'},op:'eq',value:'foreign'};f.explanation.sync();assert.equal(f.explanationPanel().hidden,true);assert.equal(children[0].options.canAct(),false);assert.equal(children[0].disposed,1);
});
test('all-band volume keeps the original two-key pinned contract and all held occurrence guide',async t=>{
 const f=setup(t);await f.open();f.button('Explain window').onclick();assert.deepEqual(Object.keys(f.lastPinned()).sort(),['parent','selection']);assert.equal(f.explanationPanel().hidden,false);assert.equal(f.explanationPanel().dataset.severityBand,'all');assert.equal(f.guide().data.points[0].value,'5');assert.equal(children[0].options.canAct(),true);assert.equal(reads.length,1);
});
test('explanation cannot adopt a scope changed reentrantly during its pause',t=>{
 const f=setup(t,{onPause:c=>{c.clean=false}}),parent=Explanation.windowQuestion(f.c),opened=f.explanation.open({parent,selection:parent});assert.equal(opened,false);assert.equal(f.explanationPanel().hidden,true);assert.equal(children.length,0);
});
