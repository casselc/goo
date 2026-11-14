import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import {families,validateResult,statusOf} from '../ui/source-capabilities.mjs';
import {nativeControlReady,completionControlReady} from '../ui/dom-visibility.mjs';
import * as SourceRefusal from '../ui/source-refusal.mjs';
const question={fromMs:100000,toMs:160000};
const result=q=>({version:'goo/source-capabilities/v1',capturedAt:'2026-11-14T01:00:00.123456789Z',question:{...q},families:families.map(f=>({kind:f.kind,table:f.table,state:'compatible',detail:'Supported schema.',observation:{state:'observed',unit:f.unit}}))});
test('fixed source profiles validate independently of response order and detach caller data',()=>{
 const raw=result(question);raw.families.reverse();const checked=validateResult(raw,question);assert.deepEqual(checked.families.map(f=>f.kind),families.map(f=>f.kind));raw.families[0].observation.state='empty';assert.equal(checked.families.at(-1).observation.state,'observed');assert.equal(statusOf(checked.families[0]),'Records observed');assert.equal(statusOf(checked.families[2]),'Points observed');assert.equal(statusOf({...checked.families[2],observation:{state:'empty'}}),'No exported points in window');
 for(const [state,observation,label] of [['absent','unknown','Table absent'],['incompatible','unknown','Unsupported schema'],['unreadable','unknown','Source unavailable'],['compatible','empty','No records in window'],['compatible','unreadable','Read unavailable'],['compatible','unknown','Not checked for records']])assert.equal(statusOf({state,observation:{state:observation}}),label);
});
test('malformed, duplicate, unknown and mismatched source profiles are refused',()=>{
 const edits=[r=>delete r.capturedAt,r=>r.capturedAt='2026-02-31T00:00:00Z',r=>r.capturedAt='2026-11-14T01:00:00+00:00',r=>r.capturedAt='not a clock',r=>r.question.toMs++,r=>r.question.extra=true,r=>r.families.pop(),r=>r.families[1]=structuredClone(r.families[0]),r=>r.families[0].table='user_supplied',r=>r.families[0].state='healthy',r=>r.families[0].observation.unit='requests',r=>r.families[0].observation.count=1,r=>r.families[0].detail='x'.repeat(257),r=>r.families[0].detail='secret\nerror',r=>{r.families[0].state='absent';r.families[0].observation.state='observed'},r=>r.families[0].observation=null];
 for(const edit of edits){const raw=result(question);edit(raw);assert.throws(()=>validateResult(raw,question),/response rejected/)}
});
class Element{
 constructor(tag){this.localName=tag;this.children=[];this.parentElement=null;this.attrs=new Map();this.dataset={};this.disabled=false;this.hidden=false;this.open=true;this.listeners=new Map();this._text='';this.root=false}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}get textContent(){return this._text+this.children.map(c=>c.textContent||'').join('')}set textContent(v){this._text=String(v);this.replaceChildren()}
 append(...nodes){for(const node of nodes){node.parentElement=this;this.children.push(node)}}replaceChildren(...nodes){for(const c of this.children)c.parentElement=null;this.children=[];this.append(...nodes)}
 setAttribute(k,v){this.attrs.set(k,String(v))}getAttribute(k){return this.attrs.get(k)??null}removeAttribute(k){this.attrs.delete(k)}
 addEventListener(k,f){this.listeners.set(k,f)}removeEventListener(k){this.listeners.delete(k)}
 getClientRects(){return this.isConnected?[{}]:[]}
 matches(selector){return selector===':disabled'?this.disabled:false}
 contains(node){return node===this||this.children.some(c=>c.contains(node))}
 querySelector(selector){return selector===':scope > summary'?this.children.find(c=>c.localName==='summary')||null:null}
 closest(selector){for(let p=this;p;p=p.parentElement){if(selector==='details:not([open])'&&p.localName==='details'&&!p.open)return p;if(selector==='fieldset:disabled'&&p.localName==='fieldset'&&p.disabled)return p;if(selector.includes('[inert]')&&p.attrs.has('inert'))return p;if(selector.includes('[hidden]')&&p.hidden)return p;if(selector.includes('[aria-hidden="true"]')&&p.getAttribute('aria-hidden')==='true')return p}return null}
 remove(){if(this.parentElement){const i=this.parentElement.children.indexOf(this);this.parentElement.children.splice(i,1);this.parentElement=null}}
}
function fixture(){
 const document={hidden:false},host=new Element('details'),minutes=new Element('input'),calls=[],pending=[],selections=[];host.root=true;minutes.value='2';host.append(minutes);globalThis.document=document;
 const text=readFileSync(new URL('../ui/source-capabilities.mjs',import.meta.url),'utf8').replace(/^import .*;\n/gm,'').replace(/^export .*;$/gm,'');
 const context={document,UI:{node:(tag,text,cls)=>{const e=new Element(tag);if(text!==undefined)e.textContent=text;if(cls)e.className=cls;return e}},Queries:{postJSON:async(url,q,signal)=>{calls.push({url,q,signal});return new Promise(resolve=>pending.push(resolve))}},SourceRefusal,nativeControlReady,completionControlReady,TextEncoder,AbortController,Date,Set,Map,structuredClone};
 vm.runInNewContext(text+';this.mount=mount',context);const api=context.mount(host,{minutes,active:()=>true,onSelect:kind=>{selections.push(kind);return true}}),section=host.children[1],read=section.children[0].children[0],status=section.children[1],output=section.children[2];
 const find=(root,label)=>root.localName==='button'&&root.textContent===label?root:root.children.map(c=>find(c,label)).find(Boolean);
 return {host,minutes,api,read,status,output,calls,pending,selections,find,reply:()=>({ok:true,data:result(calls.at(-1).q)})};
}
test('mount and local signal choice do not read; actual controls refuse closed carriers and retired results',async()=>{
 const f=fixture();assert.equal(f.calls.length,0);f.host.open=false;await f.read.onclick();assert.equal(f.calls.length,0);f.host.open=true;
 const reading=f.read.onclick();assert.equal(f.calls.length,1);f.pending.shift()(f.reply());await reading;const choose=f.find(f.output,'Choose logs');assert.ok(choose);choose.onclick();assert.deepEqual(f.selections,['logs']);assert.equal(f.calls.length,1);choose.onclick();assert.deepEqual(f.selections,['logs']);assert.equal(f.output.children.length,0);
});
test('withdrawn pending reply releases busy state without publication and fresh retry succeeds',async()=>{
 const f=fixture(),reading=f.read.onclick();assert.equal(f.read.disabled,true);f.host.setAttribute('inert','');f.pending.shift()(f.reply());await reading;assert.equal(f.output.children.length,0);assert.equal(f.read.disabled,false);f.host.removeAttribute('inert');
 const retry=f.read.onclick();f.pending.shift()(f.reply());await retry;assert.ok(f.find(f.output,'Choose spans'));assert.equal(f.calls.length,2);
});
test('minute edits and dispose retire pending replies and abort their source leases',async()=>{
 for(const dispose of [false,true]){const f=fixture(),reading=f.read.onclick(),signal=f.calls[0].signal;if(dispose)f.api.dispose();else f.minutes.listeners.get('input')();assert.equal(signal.aborted,true);f.pending.shift()(f.reply());await reading;assert.equal(f.output.children.length,0);assert.equal(f.read.disabled,false)}
});
