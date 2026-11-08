import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
let reads=[],scopes=[],opens=[],suspends=0;
mock.module('../ui/service-browser.mjs',{exports:{mount:()=>({setScope:q=>scopes.push(structuredClone(q)),suspend:()=>suspends++,open:(...args)=>opens.push(args)})}});
mock.module('../ui/visualizations.mjs',{exports:{actionVisible:el=>{for(let n=el;n;n=n.parentElement)if(n.hidden||n.inert)return false;return true}}});
mock.module('../ui/distributed-services.mjs',{exports:{mount(){}}});
mock.module('../ui/queries.mjs',{exports:{postJSON:(url,q,signal)=>new Promise(resolve=>reads.push({url,q:structuredClone(q),signal,resolve}))}});
const I=await import('../ui/inventory.mjs');
const fixed={environmentFields:{environment:'prod'},windowMode:'fixed',fromUTC:'2026-11-08T12:00:00.001Z',toUTC:'2026-11-08T12:10:00Z'};
test('catalog scopes preserve exact UTC, empty/missing namespaces and typed environment policy',()=>{
 for(const [mode,namespace]of [['all',undefined],['missing',{mode:'missing'}],['value',{mode:'value',value:''}]]){const input={...fixed,namespaceMode:mode},q=I.catalogQuestion(input);assert.equal(q.fromMs,Date.parse(fixed.fromUTC));assert.equal(q.toMs,Date.parse(fixed.toUTC));assert.deepEqual(q.namespace,namespace);assert.equal(q.environment,'prod');assert.equal(input.namespace,undefined)}
 const selector={version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'unreported'}};assert.deepEqual(I.catalogQuestion({...fixed,environmentFields:{environmentSelector:selector}}).environmentSelector,selector);
 const q=I.catalogQuestion({environmentFields:{environment:'prod'},minutes:5,nowMs:600000});assert.equal(q.fromMs,300000);assert.equal(q.toMs,600000);assert.equal(q.namespace,undefined);
});
test('catalog rejects non-UTC, rolled dates, invalid bounds and invalid namespaces before source acquisition',()=>{
 for(const patch of [{fromUTC:'2026-11-08T12:00:00'},{fromUTC:'2026-11-08T12:00:00+00:00'},{fromUTC:'2026-02-30T12:00:00Z'},{fromUTC:'2026-11-08T12:00:60Z'},{toUTC:fixed.fromUTC},{toUTC:'2026-11-08T15:00:00Z'},{namespaceMode:'unknown'},{namespaceMode:'value',namespaceValue:'x'.repeat(129)},{windowMode:'unknown'}])assert.throws(()=>I.catalogQuestion({...fixed,...patch}));
 for(const minutes of [0,121,NaN,Infinity])assert.throws(()=>I.catalogQuestion({environmentFields:{environment:'prod'},minutes,nowMs:600000}));
});
test('catalog optional-source note never equates missing table with empty or executable data',()=>{
 assert.equal(I.catalogSourceNote({diagnostics:[]}), '');assert.equal(I.catalogSourceNote({}), '');assert.equal(I.catalogSourceNote({diagnostics:[{kind:'exponential-histogram',reason:'different'}]}),'');assert.match(I.catalogSourceNote({diagnostics:[{kind:'exponential-histogram',reason:'optional-table-missing'}]}),/unavailable.*incomplete/);
});
class Element{
 get isConnected(){return this.root||!!this.parentElement?.isConnected}
 constructor(tag){this.tag=tag;this.children=[];this.parentElement=null;this.hidden=false;this.dataset={};this.value='';this.textContent='';this.listeners=new Map()}
 append(...nodes){for(const n of nodes){n.parentElement=this;this.children.push(n)}}replaceChildren(...nodes){for(const n of this.children)n.parentElement=null;this.children=[];this.append(...nodes)}
 contains(node){return this===node||this.children.some(child=>child.contains(node))}
 setAttribute(key,value){this[key]=String(value)}addEventListener(key,value){this.listeners.set(key,value)}
 set innerHTML(html){this.replaceChildren();const stack=[this];for(const token of html.matchAll(/<\/?([\w-]+)([^>]*)>|([^<]+)/g)){if(token[3]){const n=new Element('#text');n.textContent=token[3];stack.at(-1).append(n);continue}if(token[0].startsWith('</')){stack.pop();continue}const n=new Element(token[1]);for(const a of token[2].matchAll(/([\w-]+)(?:="([^"]*)")?/g))n[a[1]]=a[2]??true;stack.at(-1).append(n);if(n.tag!=='input')stack.push(n)}}
 closest(selector){for(let n=this;n;n=n.parentElement)if(selector==='[inert]'?n.inert:selector==='.controls'&&n.class=== 'controls')return n;return null}
}
const walk=root=>[root,...root.children.flatMap(walk)];
function setup(){reads=[];scopes=[];opens=[];suspends=0;const host=new Element('main'),minutes=new Element('input');minutes.id='minutes';minutes.value='10';const root=new Element('body');root.root=true;root.append(host,minutes);globalThis.document={createElement:t=>new Element(t),getElementById:id=>walk(root).find(n=>n.id===id)};globalThis.fetch=async()=>({ok:true,json:async()=>[]});const owner=I.mount(host),el=id=>document.getElementById(id);el('inventoryWindow').value='fixed';el('inventoryNamespace').value='value';el('inventoryFromUTC').value=fixed.fromUTC;el('inventoryToUTC').value=fixed.toUTC;el('inventoryEnvironment').value='prod';return{host,el,owner}}
function finish(read,diagnostics=[],signals=[]){read.resolve({ok:true,data:{question:read.q,signals,dependencies:[],diagnostics}})}
test('mounted catalog scope edits abort reads and retire late results; fixed UTC ignores unrelated global lookback',async()=>{
 const f=setup(),pending=f.el('inventoryRun').onclick(),old=reads[0];assert.deepEqual(old.q.namespace,{mode:'value',value:''});assert.equal(old.signal.aborted,false);f.el('inventoryNamespace').value='missing';f.el('inventoryNamespace').onchange();assert(old.signal.aborted);finish(old);await pending;assert.equal(f.el('serviceCatalog').children.length,0);assert.match(f.el('inventoryStatus').textContent,/Scope changed/);assert.equal(scopes.at(-1),null);
 const next=f.el('inventoryRun').onclick(),current=reads[1];assert.deepEqual(current.q.namespace,{mode:'missing'});document.getElementById('minutes').listeners.get('input')();assert.equal(current.signal.aborted,false);finish(current,[{kind:'exponential-histogram',reason:'optional-table-missing'}]);await next;assert.deepEqual(scopes.at(-1),current.q);assert(walk(f.el('serviceCatalog')).some(n=>n.dataset.sourceDiagnostic==='optional-table-missing'));f.el('inventoryFromUTC').oninput();assert.equal(f.el('serviceCatalog').children.length,0);assert.equal(scopes.at(-1),null);
});
test('mounted catalog invalid fixed scope acquires nothing and external fixed scope populates exact controls',async()=>{
 const f=setup();f.el('inventoryFromUTC').value='invalid';await f.el('inventoryRun').onclick();assert.equal(reads.length,0);assert.match(f.el('inventoryStatus').textContent,/UTC/);
 const selector={version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'unreported'}},q={environmentSelector:selector,namespace:{mode:'value',value:''},fromMs:Date.parse(fixed.fromUTC),toMs:Date.parse(fixed.toUTC)},pending=f.owner.openQuestion(q);assert.deepEqual(reads[0].q,q);assert.equal(f.el('inventoryWindow').value,'fixed');assert.equal(f.el('inventoryFromUTC').value,fixed.fromUTC);assert.equal(f.el('inventoryNamespace').value,'value');assert.equal(f.el('inventoryNamespaceValue').value,'');assert.equal(f.el('inventoryNamespaceField').hidden,false);finish(reads[0]);await pending;
});
test('mounted catalog hidden, inert and disabled Discover callbacks cannot acquire sources',async()=>{
 const f=setup(),run=f.el('inventoryRun'),controls=run.closest('.controls');for(const key of ['hidden','inert']){controls[key]=true;await run.onclick();controls[key]=false;assert.equal(reads.length,0)}run.disabled=true;await run.onclick();assert.equal(reads.length,0);
});

test('catalog row callbacks retain their rendered evidence lease across search, scope changes and reattached old DOM',async()=>{
 const f=setup(),signal={service:'api',signal:'metrics',count:'13',firstMs:Date.parse(fixed.fromUTC),lastMs:Date.parse(fixed.fromUTC)};
 let pending=f.el('inventoryRun').onclick();finish(reads[0],[],[signal]);await pending;
 const rowButton=()=>walk(f.el('serviceCatalog')).find(n=>n.tag==='button'&&n['aria-label']?.startsWith('Open service ')),old=rowButton();assert(old);assert.equal(old.textContent,'Open service');assert.match(old['aria-label'],/^Open service api$/);const table=walk(f.el('serviceCatalog')).find(n=>n.tag==='goo-table');assert.equal(table.density,'compact');assert.equal(table['mobile-layout'],'stacked');assert(walk(table).some(n=>n.tag==='caption'));assert.equal(walk(table).filter(n=>n.tag==='th'&&n.scope==='col').length,6);old.onclick();assert.equal(opens.length,1);
 f.el('inventorySearch').oninput();old.onclick();assert.equal(opens.length,1);const current=rowButton();current.onclick();assert.equal(opens.length,2);
 current.inert=true;current.onclick();current.inert=false;assert.equal(opens.length,2);
 f.el('inventoryNamespace').value='missing';f.el('inventoryNamespace').onchange();pending=f.el('inventoryRun').onclick();finish(reads[1],[],[{...signal,service:'worker'}]);await pending;f.el('serviceCatalog').append(old);old.onclick();current.onclick();assert.equal(opens.length,2);rowButton().onclick();assert.equal(opens.length,3);assert.equal(opens.at(-1)[0],'worker');
});
