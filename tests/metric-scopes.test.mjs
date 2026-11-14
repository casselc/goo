import {test} from 'node:test';
import assert from 'node:assert/strict';
import {validateResult,selection,mount} from '../ui/metric-scopes.mjs';
import * as Environment from '../ui/record-environment.mjs';
import {families} from '../ui/source-capabilities.mjs';
const question={fromMs:100000,toMs:160000};
const choice=(mode='unreported',ns={mode:'missing'})=>({service:'same-service',namespace:ns,environment:{mode,...(mode==='named'?{value:'prod'}:{})},environmentSelector:Environment.effective({mode,...(mode==='named'?{value:'prod'}:{})}),points:'9007199254740993',firstMs:120000,lastMs:150000,selectable:true});
function result(q=question){return {version:'goo/metric-scopes/v1',question:{...q},capturedAt:'2026-11-14T03:00:00.123456789Z',sources:families.filter(f=>!['logs','spans'].includes(f.kind)).map(f=>({kind:f.kind,table:f.table,state:'compatible',detail:'Discovery profile ready.'})),diagnostics:[],choices:[{...choice(),firstMs:q.fromMs+1,lastMs:q.toMs-1}],unit:'exported-points',warning:'Independent reads; coverage unknown.'}}
test('metric scopes retain exact typed environment, namespace, point strings and the discovery window',()=>{
 for(const mode of ['named','empty','unreported'])for(const ns of [{mode:'missing'},{mode:'value'},{mode:'value',value:'team'}]){
  const raw=result();raw.choices=[choice(mode,ns)];const checked=validateResult(raw,question),selected=selection(checked.choices[0],checked.question);
  assert.deepEqual(selected,{service:'same-service',namespace:ns,environment:'',environmentSelector:Environment.effective(raw.choices[0].environment),...question});assert.equal(checked.choices[0].points,'9007199254740993');raw.choices[0].service='changed';assert.equal(checked.choices[0].service,'same-service');
 }
});
test('nonselectable identities remain visible but never become discoverable intent',()=>{
 const raw=result(),row=raw.choices[0];row.service='';row.selectable=false;row.reason='Service name required.';assert.doesNotThrow(()=>validateResult(raw,question));assert.throws(()=>selection(row,question));
 row.environment={mode:'named',value:'x'.repeat(129)};delete row.environmentSelector;assert.doesNotThrow(()=>validateResult(raw,question));assert.throws(()=>selection(row,question));
});
test('metric scope validation rejects ambiguous identities, invented support and mismatched windows',()=>{
 const edits=[r=>r.question.toMs++,r=>r.unit='requests',r=>r.choices.push(structuredClone(r.choices[0])),r=>r.choices[0].namespace=null,r=>r.choices[0].environmentSelector.policy='current-only',r=>r.choices[0].environment={mode:'empty'},r=>r.choices[0].points=2,r=>r.choices[0].points='0',r=>r.choices[0].points=String(1n<<128n),r=>r.choices[0].firstMs=0,r=>r.choices[0].lastMs=160000,r=>r.choices[0].selectable=false,r=>r.sources.pop(),r=>r.sources[0].table='arbitrary',r=>r.sources[0].state='healthy',r=>{for(const source of r.sources)source.state='absent';r.diagnostics=r.sources.map(source=>({kind:source.kind,reason:'absent'}))},r=>{r.sources[0].state='absent'},r=>{r.sources[0].state='absent';r.diagnostics=[{kind:'gauge',reason:'unreadable'}]},r=>{r.sources[0].state='absent';r.diagnostics=[{kind:'gauge',reason:'absent'},{kind:'gauge',reason:'absent'}]},r=>r.capturedAt='2026-02-30T00:00:00Z'];
 for(const edit of edits){const raw=result();edit(raw);assert.throws(()=>validateResult(raw,question),/response rejected/)}
 assert.throws(()=>selection(choice(),{fromMs:0,toMs:7200001}));
});
class Element{
 constructor(tag){this.localName=tag;this.children=[];this.parentElement=null;this.attrs=new Map();this.dataset={};this.disabled=false;this.hidden=false;this.open=false;this.root=false;this.listeners=new Map();this._text=''}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}get textContent(){return this._text+this.children.map(c=>c.textContent).join('')}set textContent(value){this._text=String(value);this.replaceChildren()}
 append(...nodes){for(const n of nodes){n.parentElement=this;this.children.push(n)}}replaceChildren(...nodes){for(const n of this.children)n.parentElement=null;this.children=[];this.append(...nodes)}
 setAttribute(k,v){this.attrs.set(k,String(v))}getAttribute(k){return this.attrs.get(k)??null}removeAttribute(k){this.attrs.delete(k)}
 addEventListener(k,f){this.listeners.set(k,f)}getClientRects(){return this.isConnected?[{}]:[]}scrollIntoView(){}focus(){globalThis.document.activeElement=this}
 contains(n){return this===n||this.children.some(c=>c.contains(n))}matches(s){return s===':disabled'?this.disabled:false}querySelector(s){return [':scope > summary',':scope>summary'].includes(s)?this.children.find(c=>c.localName==='summary')||null:null}
 closest(s){for(let p=this;p;p=p.parentElement){if(s==='details:not([open])'&&p.localName==='details'&&!p.open)return p;if(s==='fieldset:disabled'&&p.localName==='fieldset'&&p.disabled)return p;if(s.includes('[inert]')&&p.attrs.has('inert'))return p;if(s.includes('[hidden]')&&p.hidden)return p;if(s.includes('[aria-hidden="true"]')&&p.getAttribute('aria-hidden')==='true')return p}return null}
 remove(){if(this.parentElement){this.parentElement.children=this.parentElement.children.filter(c=>c!==this);this.parentElement=null}}
}
function fixture(){
 const visibilityEvents=new EventTarget(),listeners=new Set(),document={hidden:false,activeElement:null,addEventListener(type,fn){visibilityEvents.addEventListener(type,fn);listeners.add(fn)},removeEventListener(type,fn){visibilityEvents.removeEventListener(type,fn);listeners.delete(fn)},dispatchEvent:event=>visibilityEvents.dispatchEvent(event)},host=new Element('section'),calls=[],pending=[],selected=[];host.root=true;globalThis.document=document;
 globalThis.document.createElement=tag=>new Element(tag);
 globalThis.fetch=(url,options={})=>{if(String(url).endsWith('/cancel'))return Promise.resolve(Response.json({}));calls.push({path:url,q:JSON.parse(options.body),signal:options.signal});return new Promise(resolve=>pending.push(resolve))};
 const api=mount(host,{onSelect:next=>{selected.push(next);return true}}),box=host.children[0],read=box.children[1].children[1],minutes=box.children[1].children[0].children[0],output=box.children[4];
 const find=(e,label)=>e.localName==='button'&&e.textContent===label?e:e.children.map(c=>find(c,label)).find(Boolean);
 return {api,box,read,minutes,output,calls,pending,selected,find,document,listeners,reply:()=>Response.json(result(calls.at(-1).q))};
}
test('actual controls read only explicitly, retain a handed-off window and stage locally',async()=>{
 const f=fixture();assert.equal(f.calls.length,0);f.api.open(question);assert.equal(f.calls.length,0);const reading=f.read.onclick();assert.deepEqual({...f.calls[0].q},question);f.pending.shift()(f.reply());await reading;const choose=f.find(f.output,'Choose metric scope');assert(choose);choose.onclick();assert.equal(f.selected.length,1);assert.equal(f.calls.length,1);choose.onclick();assert.equal(f.selected.length,1);
 f.api.open(question);f.minutes.value='1';f.minutes.oninput();const fresh=f.read.onclick();assert.equal(f.calls[1].q.toMs-f.calls[1].q.fromMs,60000);assert.notEqual(f.calls[1].q.toMs,question.toMs);f.pending.shift()(f.reply());await fresh;assert.ok(f.find(f.output,'Choose metric scope'));
});
test('closed/inert controls and retired pending requests cannot publish or stage',async()=>{
 const f=fixture();await f.read.onclick();assert.equal(f.calls.length,0);f.api.open();const reading=f.read.onclick();f.box.setAttribute('inert','');f.pending.shift()(f.reply());await reading;assert.equal(f.output.children.length,0);assert.equal(f.read.disabled,false);f.box.removeAttribute('inert');const retry=f.read.onclick();f.api.clear();assert.equal(f.calls[1].signal.aborted,true);f.pending.shift()(f.reply());await retry;assert.equal(f.output.children.length,0);
});

test('hidden-to-visible transitions retire pending discovery before a late reply and permit a fresh retry',async()=>{
 const f=fixture();f.api.open();const reading=f.read.onclick(),signal=f.calls[0].signal;f.document.hidden=true;f.document.dispatchEvent(new Event('visibilitychange'));assert.equal(signal.aborted,true);assert.equal(f.read.disabled,false);f.document.hidden=false;f.document.dispatchEvent(new Event('visibilitychange'));f.pending.shift()(f.reply());await reading;assert.equal(f.output.children.length,0);assert.equal(f.calls.length,1);
 const retry=f.read.onclick();f.pending.shift()(f.reply());await retry;assert.ok(f.find(f.output,'Choose metric scope'));assert.equal(f.calls.length,2);assert.equal(f.listeners.size,1);f.api.dispose();assert.equal(f.listeners.size,0);
});

test('supplied choices retain source order while the local view leads with usable scopes and folded exclusions',async()=>{
 const f=fixture();f.api.open(question);const reading=f.read.onclick(),raw=result(question);raw.choices.unshift({...choice(),service:'',selectable:false,reason:'Service name is empty or unreported.'});for(const source of raw.sources.slice(1))source.state='absent';raw.diagnostics=raw.sources.slice(1).map(source=>({kind:source.kind,reason:'absent'}));
 f.pending.shift()(Response.json(raw));await reading;const viewport=f.output.children[0],table=viewport.children[0],body=table.children[2],excluded=f.output.children[1];assert.equal(viewport.localName,'goo-table');assert.equal(body.children[0].children[0].textContent,'"same-service"');assert.equal(body.children.at(-1).children[0].textContent,'Empty or unreported service');assert.equal(raw.choices[0].service,'');assert.equal(f.find(f.output,'Choose metric scope').disabled,false);assert.equal(excluded.localName,'details');assert.equal(excluded.children[0].textContent,'3 sources excluded');assert.equal(excluded.open,false);assert.equal(f.box.children[2].hidden,true);
});
