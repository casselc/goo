import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
let reads=[];
class Element{
 constructor(tag,text=''){this.tag=tag;this.raw=text;this.children=[];this.parentElement=null;this.attributes={};this.dataset={};this.hidden=false;this.disabled=false;this.value=''}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}
 contains(n){return n===this||walk(this).includes(n)}
 get textContent(){return this.raw+this.children.map(n=>n.textContent).join('')}
 set textContent(text){this.raw=text;this.children=[]}
 append(...nodes){for(const n of nodes){if(n.parentElement)n.parentElement.children=n.parentElement.children.filter(x=>x!==n);n.parentElement=this;this.children.push(n)}}
 replaceChildren(...nodes){for(const n of this.children)n.parentElement=null;this.children=[];this.raw='';this.append(...nodes)}
 setAttribute(k,v){this.attributes[k]=String(v)}
 closest(selector){for(let n=this;n;n=n.parentElement)if(selector==='[inert]'&&n.inert)return n;return null}
 querySelector(selector){return walk(this).find(n=>selector==='[role="status"]'?n.attributes.role==='status':selector==='[aria-pressed="true"]'?n.attributes['aria-pressed']==='true':n.tag===selector)}
 focus(){document.activeElement=this}
}
const walk=root=>root.children.flatMap(n=>[n,...walk(n)]);
const node=(tag,text='',cls='')=>{const n=new Element(tag,text);n.className=cls;return n};
const visible=n=>{if(!n?.isConnected||document.hidden)return false;for(let p=n;p;p=p.parentElement)if(p.hidden||p.inert)return false;return true};
mock.module('../ui/components.mjs',{exports:{node}});
mock.module('../ui/dom-visibility.mjs',{exports:{actionVisible:visible,nativeControlReady:n=>visible(n)&&!n.disabled}});
mock.module('../ui/queries.mjs',{exports:{postJSON:(url,q,signal)=>new Promise((resolve,reject)=>reads.push({url,q,signal,resolve,reject}))}});
mock.module('../ui/service-triage-model.mjs',{exports:{
 plan:scope=>({scope,current:{...scope},baseline:{...scope,fromMs:scope.fromMs-(scope.toMs-scope.fromMs),toMs:scope.fromMs}}),
 admit:(plan,current,baseline)=>{if(!current.admitted||!baseline.admitted)throw Error('Incomplete server summary');return{plan,current,baseline}},
 rows:snapshot=>snapshot.current.rows||[],sortRows:rows=>rows,
}});
const {mount}=await import('../ui/service-triage.mjs');
function setup(t){
 reads=[];const listeners=new Map();globalThis.document={hidden:false,activeElement:null,addEventListener:(k,f)=>listeners.set(k,f),removeEventListener:k=>listeners.delete(k)};
 const root=node('main');root.root=true;const host=node('section'),signalView=node('div');root.append(host,signalView);const opened=[],owner=mount();let current=true;
 owner.setScope({environment:'prod',fromMs:1000,toMs:2000});const attach=(options={})=>owner.attach(host,{signalView,rows:options.rows||[],search:options.search||'',openService:(...args)=>opened.push(args),current:()=>current});attach();
 t.after(()=>owner.dispose());return{root,host,signalView,owner,opened,listeners,attach,retire:()=>current=false,button:text=>walk(host).find(n=>n.tag==='button'&&n.textContent===text),table:()=>walk(host).find(n=>n.tag==='table')};
}
const complete=(read,rows=[])=>read.resolve({ok:true,data:{admitted:true,rows}});
const sample={key:'["api","missing",""]',service:'api',namespace:{mode:'missing'},current:{count:'12',spanStatus:{error:'3'},duration:{state:'observed',sumNs:'24000000',count:'12'}},baseline:{count:'10'},countDelta:'2'};

test('service triage requires an explicit current control and two admitted windows; cached modes and ranking never read',async t=>{
 const f=setup(t),measure=f.button('Measure server spans');assert.equal(reads.length,0);measure.focus();const work=measure.onclick();assert.equal(reads.length,2);assert.equal(f.signalView.hidden,true);assert.equal(f.table(),undefined);assert.equal(document.activeElement.attributes.role,'status');
 complete(reads[0],[sample]);await Promise.resolve();assert.equal(f.table(),undefined);complete(reads[1]);await work;
 assert(f.table());assert.match(f.table().textContent,/12\+2 vs previous25\.00%3\/122\.000 ms/);assert.equal(document.activeElement,f.button('Measure server spans'));
 f.button('Signal counts').onclick();assert.equal(f.signalView.hidden,false);f.button('Server summary').onclick();assert.equal(f.signalView.hidden,true);const rank=walk(f.host).find(n=>n.tag==='select');rank.value='mean';rank.onchange();assert.equal(reads.length,2);assert(f.table());
 f.attach();assert.equal(f.signalView.hidden,true);assert.equal(walk(f.host).find(n=>n.tag==='select').value,'mean');assert.equal(reads.length,2);f.root.append(measure);await measure.onclick();assert.equal(reads.length,2,'reattached old controls cannot acquire sources');
});

test('failed or incomplete second window never publishes a ranked table or synthetic zero result',async t=>{
 const f=setup(t),work=f.button('Measure server spans').onclick();complete(reads[0],[sample]);reads[1].resolve({ok:true,data:{admitted:false}});await work;
 assert.equal(f.table(),undefined);assert.match(f.host.textContent,/Summary unavailable.*Incomplete server summary/);assert.equal(reads.length,2);
});

test('scope replacement and cancellation retire late responses and retained callbacks',async t=>{
 const f=setup(t),old=f.button('Measure server spans'),work=old.onclick();f.button('Cancel summary').onclick();assert(reads.every(r=>r.signal.aborted));complete(reads[0],[sample]);complete(reads[1]);await work;assert.equal(f.table(),undefined);assert.match(f.host.textContent,/canceled/);
 f.owner.setScope({environment:'other',fromMs:2000,toMs:3000});f.attach();old.onclick();assert.equal(reads.length,2);const next=f.button('Measure server spans').onclick();f.owner.setScope(null);complete(reads[2],[sample]);complete(reads[3]);await next;assert.equal(f.table(),undefined);
});

test('hidden, inert, departed and disabled source callbacks refuse reads; hiding during a read aborts it',async t=>{
 const f=setup(t),button=f.button('Measure server spans');for(const flag of ['hidden','inert']){f.root[flag]=true;await button.onclick();f.root[flag]=false}button.disabled=true;await button.onclick();button.disabled=false;assert.equal(reads.length,0);
 const work=button.onclick();document.hidden=true;f.listeners.get('visibilitychange')();assert(reads.every(r=>r.signal.aborted));complete(reads[0],[sample]);complete(reads[1]);await work;document.hidden=false;assert.equal(f.table(),undefined);
 f.attach();f.retire();await f.button('Measure server spans').onclick();assert.equal(reads.length,2);
});

test('service opening preserves its live opener; stopped complete views do not replace the retained table',async t=>{
 const f=setup(t),work=f.button('Measure server spans').onclick();complete(reads[0],[sample]);complete(reads[1]);await work;const button=f.button('Open service');button.onclick();assert.equal(f.opened.length,1);assert.equal(f.opened[0][1],button);f.owner.suspend();assert.equal(button.isConnected,true);assert.equal(f.button('Open service'),button);f.root.hidden=true;button.onclick();assert.equal(f.opened.length,1);assert.equal(reads.length,2);
});


test('mode switches and cancel restore focused replacements while asynchronous completion never steals moved focus',async t=>{
 const f=setup(t),work=f.button('Measure server spans').onclick();const outside=node('button','Other control');f.root.append(outside);outside.focus();complete(reads[0],[sample]);complete(reads[1]);await work;assert.equal(document.activeElement,outside);
 const signals=f.button('Signal counts');signals.focus();signals.onclick();assert.equal(document.activeElement,f.button('Measure server spans'));assert.equal(f.button('Signal counts'),undefined);const summary=f.button('Server summary');summary.focus();summary.onclick();assert.equal(document.activeElement,f.button('Measure server spans'));assert.equal(reads.length,2);
 const reread=f.button('Measure server spans').onclick(),cancel=f.button('Cancel summary');cancel.focus();cancel.onclick();assert.equal(document.activeElement,f.button('Measure server spans'));complete(reads[2],[sample]);complete(reads[3]);await reread;
});

test('unqualified declarations remain unavailable rather than a synthetic zero or missing-namespace group',async t=>{
 const f=setup(t),work=f.button('Measure server spans').onclick(),wildcard={key:'wildcard',service:'declared',namespace:null,qualified:false,current:null,baseline:null,countDelta:null};complete(reads[0],[wildcard]);complete(reads[1]);await work;assert.match(f.table().textContent,/All namespacesNamespace unspecified in declarationUnavailableUnavailableUnavailable/);const opener=f.button('Open service');opener.onclick();assert.equal(f.opened[0][2],null);
});


test('signal and declaration search matches persist through cached triage mode changes without source reads',async t=>{
 const f=setup(t),work=f.button('Measure server spans').onclick();complete(reads[0],[sample]);complete(reads[1]);await work;
 const catalog=[{service:'api',namespace:{mode:'missing'},signals:{logs:{count:'4'}},declarations:[{title:'Billing integration',name:'billing'}]}];
 for(const search of ['logs','Billing']){f.attach({rows:catalog,search});assert.equal(walk(f.table()).filter(n=>n.tag==='tbody')[0].children.length,1);f.button('Signal counts').onclick();f.button('Server summary').onclick();assert.equal(walk(f.table()).filter(n=>n.tag==='tbody')[0].children.length,1)}
 f.attach({rows:catalog,search:'absent'});assert.equal(walk(f.table()).filter(n=>n.tag==='tbody')[0].children.length,0);assert.equal(reads.length,2);
});


test('callbacks bind their creating draw even when old controls are reattached inside a replacement scope',async t=>{
 const f=setup(t),initial=f.button('Measure server spans').onclick();complete(reads[0],[sample]);complete(reads[1]);await initial;
 const oldMeasure=f.button('Measure server spans'),oldSignals=f.button('Signal counts'),oldRank=walk(f.host).find(n=>n.tag==='select');
 f.owner.setScope({environment:'replacement',fromMs:2000,toMs:3000});f.attach();f.host.append(oldMeasure,oldSignals,oldRank);
 assert.equal(f.host.contains(oldMeasure),true);await oldMeasure.onclick();oldSignals.onclick();oldRank.value='mean';oldRank.onchange();assert.equal(reads.length,2);assert.equal(f.signalView.hidden,false);
 const measure=f.button('Measure server spans'),work=measure.onclick(),oldCancel=f.button('Cancel summary');assert.equal(reads.length,4);oldCancel.onclick();complete(reads[2],[sample]);complete(reads[3]);await work;
 const next=f.button('Measure server spans').onclick();assert.equal(reads.length,6);f.host.append(oldCancel);oldCancel.onclick();assert.equal(reads[4].signal.aborted,false);assert.equal(reads[5].signal.aborted,false);complete(reads[4],[sample]);complete(reads[5]);await next;assert(f.table());
});
