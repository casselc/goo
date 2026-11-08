import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
const {result:histogramResult}=createRequire(import.meta.url)('./fixtures/native-histogram.cjs');

// Minimal DOM implements the actual mount's form and action lifecycles. No source
// math, failure presenter or Measurements callback is reimplemented here.
class Element{
 constructor(tag){this.localName=tag;this.children=[];this.parentElement=null;this.attributes=new Map();this.dataset={};this.className='';this._text='';this.value='';this.hidden=false;this.disabled=false;this.open=false;this.style={};this.classList={add:n=>this.className+=' '+n,remove:()=>{},toggle:()=>{}}}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}get firstChild(){return this.children[0]??null}get childNodes(){return this.children}
 get textContent(){return this._text+this.children.map(c=>c.textContent).join('')}set textContent(value){this.replaceChildren();this._text=String(value??'')}
 set innerHTML(html){this.replaceChildren();const stack=[this],voids=new Set(['input','br','hr','img']);for(const token of html.matchAll(/<\/?([\w-]+)([^>]*)>|([^<]+)/g)){if(token[3]){const text=new Element('#text');text._text=token[3];stack.at(-1).append(text);continue}if(token[0].startsWith('</')){stack.pop();continue}const node=new Element(token[1]);for(const a of token[2].matchAll(/([\w-]+)(?:="([^"]*)")?/g))node.setAttribute(a[1],a[2]??'');stack.at(-1).append(node);if(!voids.has(node.localName))stack.push(node)}}
 append(...nodes){for(let node of nodes){if(typeof node==='string'){const text=new Element('#text');text._text=node;node=text}node.remove();node.parentElement=this;this.children.push(node)}}
 replaceChildren(...nodes){for(const c of this.children)c.parentElement=null;this.children=[];this._text='';this.append(...nodes)}remove(){if(this.parentElement)this.parentElement.children=this.parentElement.children.filter(c=>c!==this);this.parentElement=null}
 after(node){const parent=this.parentElement;if(!parent)return;node.remove();node.parentElement=parent;parent.children.splice(parent.children.indexOf(this)+1,0,node)}
 insertBefore(node,before){node.remove();node.parentElement=this;const index=this.children.indexOf(before);this.children.splice(index<0?this.children.length:index,0,node)}
 setAttribute(key,value){this.attributes.set(key,String(value));if(key==='id'||key==='value')this[key]=String(value);if(key==='class')this.className=String(value);if(key==='hidden'||key==='inert'||key==='open')this[key]=true;if(key.startsWith('data-'))this.dataset[key.slice(5).replace(/-([a-z])/g,(_,c)=>c.toUpperCase())]=String(value)}getAttribute(key){return this.attributes.get(key)??null}hasAttribute(key){return this.attributes.has(key)}
 matches(selector){if(selector==='details:not([open])')return this.localName==='details'&&!this.open;if(selector.startsWith('#'))return this.id===selector.slice(1);if(selector.startsWith('.'))return this.className.split(' ').includes(selector.slice(1));const a=selector.match(/^\[([\w-]+)(?:="([^"]*)")?\]$/);if(a)return a[1]==='inert'?!!this.inert:a[1]==='hidden'?!!this.hidden:a[2]===undefined?this.hasAttribute(a[1]):this.getAttribute(a[1])===a[2];return this.localName===selector}
 querySelectorAll(selector){if(selector===':scope > summary')return this.children.filter(c=>c.localName==='summary');return this.children.flatMap(c=>[...(selector.split(',').some(s=>c.matches(s.trim()))?[c]:[]),...c.querySelectorAll(selector)])}querySelector(selector){return this.querySelectorAll(selector)[0]??null}
 closest(selector){for(let n=this;n;n=n.parentElement)if(selector.split(',').some(s=>n.matches(s.trim())))return n;return null}contains(node){return this===node||this.children.some(c=>c.contains(node))}
 getClientRects(){for(let n=this;n;n=n.parentElement)if(n.hidden)return[];return this.isConnected?[{}]:[]}focus(){document.activeElement=this}select(){}scrollIntoView(){}addEventListener(){}removeEventListener(){}
}
let current;
mock.module('../ui/native-histogram-reuse.mjs',{exports:{mount:(_host,options)=>{current.nativeViews=options;return{invalidate(){},refresh(){}}}}});
mock.module('../ui/queries.mjs',{exports:{postJSON:async(path,body,signal)=>current.post(path,body,signal),jsonResponse:async pending=>{const r=await pending;return{ok:r.ok,status:r.status,data:await r.json()}}}});
const Measurements=await import('../ui/measurements.mjs');
const NativeDisplay=await import('../ui/native-histogram-declaration.mjs');
const Refusals=await import('../ui/source-refusal.mjs');
function setup(replies=[]){
 const body=new Element('body');body.root=true;const host=new Element('main');body.append(host);const events=new EventTarget();
 globalThis.document={body,hidden:false,activeElement:null,createElement:t=>new Element(t),createTextNode:t=>{const n=new Element('#text');n._text=t;return n},getElementById:id=>body.querySelector('#'+id)};
 globalThis.addEventListener=(...args)=>events.addEventListener(...args);globalThis.dispatchEvent=event=>events.dispatchEvent(event);globalThis.CustomEvent=class extends Event{constructor(type,options={}){super(type);this.detail=options.detail}};globalThis.location={hash:'#explore',origin:'http://goo.test',pathname:'/'};globalThis.history={pushState(){}};
 const q=structuredClone(histogramResult().question);q.fromNS=String(BigInt(q.fromMs)*1000000n+1n);q.toNS=String(BigInt(q.toMs)*1000000n-1n);q.definition.environment='';q.definition.environmentSelector={version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'empty'}};q.definition.source.namespace={mode:'value',value:''};
 const plan={question:q,definition:q.definition,yaml:'fixed native definition',revision:'a'.repeat(64)},reads=[],pending=[...replies];
 current={post:async(path,input,signal)=>{assert.equal(path,'/api/measurements');reads.push({question:structuredClone(input),signal});const reply=pending.shift();if(reply instanceof Error)throw reply;if(typeof reply==='function')return reply(input,signal);return reply??success(input)},nativeViews:null};
 globalThis.fetch=async(_url,options={})=>({ok:true,status:200,json:async()=>options.method==='POST'?structuredClone(plan):[]});
 const owner=Measurements.mount(host),el=id=>host.querySelector('#'+id),open=()=>owner.open({yaml:'authored input',fromMs:q.fromMs,toMs:q.toMs,fromNS:q.fromNS,toNS:q.toNS});
 return{host,q,plan,reads,pending,el,open,emit:(type,detail)=>events.dispatchEvent(new CustomEvent(type,{detail})),native:()=>current.nativeViews};
}
function success(question){const value=histogramResult();value.question=structuredClone(question);value.state='no_data';value.warnings=[];value.count='0';value.rows=[];value.histograms.series=[];return{ok:true,status:200,data:value}}
function refused(kind='timeout'){return{ok:false,status:502,data:{error:Refusals.reasons[kind],refusal:{kind,detail:'A bounded source explanation.'}}}}
async function failed(kind='timeout'){const h=setup([refused(kind)]);await h.open();await h.el('measureRun').onclick();return h}

test('mounted native Retry reuses exact typed question, nanoseconds and display intent with explicit focus recovery',async()=>{
 const h=setup([refused('timeout')]);await h.open();const declaration=NativeDisplay.create(h.q,'table');assert(h.native().onApply({question:structuredClone(h.q),declaration}));h.el('measureRun').focus();await h.el('measureRun').onclick();assert.equal(h.reads.length,1);const retry=h.el('measureRetry'),old=retry.onclick;assert.equal(retry.textContent,'Retry measurement');assert.equal(document.activeElement,retry);assert.equal(h.el('measureRun').hidden,true);assert.equal(h.el('measureCapture').disabled,true);assert.equal(h.el('measureStatus').querySelector('.source-refusal').dataset.sourceRefusalKind,'timeout');assert.equal(h.el('measureStatus').querySelector('details').open,false);assert.deepEqual(h.reads[0].question,h.q);
 retry.focus();await retry.onclick();assert.equal(h.reads.length,2);assert.deepEqual(h.reads[1].question,h.reads[0].question);assert.equal(h.el('measureCapture').disabled,false);assert.match(h.el('measureStatus').textContent,/table export interval display/);assert.equal(document.activeElement.localName,'h3');assert.equal(retry.isConnected,false);old();assert.equal(h.reads.length,2);
});
test('mounted native errors keep closed refusal kinds and raw unknown source text out of presentation',async()=>{
 for(const kind of ['timeout','budget','malformed','unavailable','cancelled']){const h=await failed(kind);assert.equal(h.reads.length,1);assert.equal(h.el('measureStatus').querySelector('.source-refusal').dataset.sourceRefusalKind,kind);assert.equal(!!h.el('measureRetry'),kind!=='cancelled')}
 for(const reply of [{ok:false,status:503,data:{error:'secret SQL/password'}},{ok:false,status:502,data:{error:'secret SQL/password',refusal:{kind:'invented',detail:'secret SQL/password'}}},new TypeError('secret SQL/password'),new SyntaxError('secret SQL/password')]){const h=setup([reply]);await h.open();await h.el('measureRun').onclick();assert(h.el('measureRetry'));assert.doesNotMatch(h.el('measureStatus').textContent,/secret|password|SQL/)}
 const h=setup([{ok:false,status:400,data:{error:'Invalid fixed question.'}}]);await h.open();await h.el('measureRun').onclick();assert.equal(h.el('measureRetry'),null);assert.match(h.el('measureStatus').textContent,/Invalid fixed question/);
});
test('retry callback leases retire on edit, preview, cancel, navigation and display replacement',async()=>{
 for(const change of [h=>h.el('measureYAML').oninput(),h=>h.el('measureFrom').oninput(),h=>h.el('measureTo').oninput(),h=>h.el('measureFixed').onclick(),h=>h.el('measureCancel').onclick(),h=>h.emit('goo:workspace',{route:'services'}),h=>h.emit('goo:explore-mode',{measurements:false}),h=>h.native().onApply({question:structuredClone(h.q),declaration:NativeDisplay.create(h.q,'histogram')})]){const h=await failed(),retry=h.el('measureRetry'),old=retry.onclick;await change(h);await old();assert.equal(h.reads.length,1);assert.equal(retry.isConnected,false)}
});
test('hidden, detached, inert, disabled and changed-question retry gestures perform no read',async()=>{
 for(const change of [(h,b)=>b.hidden=true,(h,b)=>b.remove(),(h,b)=>b.inert=true,(h,b)=>h.host.inert=true,(h,b)=>b.disabled=true,(h,b)=>document.hidden=true,(h,b)=>h.native().current().question.toNS='5000000000']){const h=await failed(),retry=h.el('measureRetry');change(h,retry);await retry.onclick();assert.equal(h.reads.length,1)}
});
test('focused repeated failure replaces retry without losing focus, and late recovery is retired by navigation',async()=>{
 const h=setup([refused('budget'),refused('unavailable')]);await h.open();await h.el('measureRun').onclick();const first=h.el('measureRetry');first.focus();await first.onclick();const next=h.el('measureRetry');assert.notEqual(next,first);assert.equal(document.activeElement,next);await first.onclick();assert.equal(h.reads.length,2);
 let finish;h.pending.push(()=>new Promise(resolve=>finish=resolve));const running=next.onclick();await Promise.resolve();assert.equal(h.reads.length,3);const last=h.reads.at(-1);h.emit('goo:workspace',{route:'services'});assert.equal(last.signal.aborted,true);finish(success(last.question));await running;assert.equal(h.el('measureRetry'),null);assert.equal(h.el('measureCapture').disabled,true);assert.equal(h.el('measureResult').children.length,0);
});


test('late catalog failure cannot overwrite a current source refusal, read, result, preview or cancellation',async()=>{
 for(const state of ['refusal','pending','result','preview','cancel']){
  const h=setup([refused('timeout')]);await h.open();let rejectMetadata,finishRead,running;
  if(state==='refusal')await h.el('measureRun').onclick();
  else if(state==='pending'){h.pending.splice(0,h.pending.length,input=>new Promise(resolve=>finishRead=()=>resolve(success(input))));running=h.el('measureRun').onclick();await Promise.resolve()}
  else if(state==='result'){h.pending.length=0;await h.el('measureRun').onclick()}
  globalThis.fetch=()=>new Promise((_resolve,reject)=>rejectMetadata=reject);const loading=h.el('measureRefresh').onclick();
  if(state==='preview'){globalThis.fetch=async()=>({ok:true,status:200,json:async()=>structuredClone(h.plan)});await h.el('measureFixed').onclick()}
  else if(state==='cancel')h.el('measureCancel').onclick();
  const before=h.el('measureStatus').textContent;rejectMetadata(new Error('catalog unavailable'));await loading;assert.equal(h.el('measureStatus').textContent,before);
  if(state==='pending'){h.el('measureCancel').onclick();finishRead();await running}
 }
 const h=setup();await h.open();globalThis.fetch=async()=>{throw Error('catalog unavailable')};await h.el('measureRefresh').onclick();assert.match(h.el('measureStatus').textContent,/Packages unavailable/);
});
