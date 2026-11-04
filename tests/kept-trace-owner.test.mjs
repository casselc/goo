import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class El{
 constructor(tag,text){this.tag=tag;this.textContent=text||'';this.children=[];this.parent=null;this.hidden=false;this.dataset={};this.disabled=false}
 get isConnected(){return this.root||!!this.parent?.isConnected}
 append(...nodes){for(const n of nodes){n.parent=this;this.children.push(n)}}
 prepend(...nodes){for(const n of nodes)n.parent=this;this.children.unshift(...nodes)}
 replaceChildren(...nodes){for(const n of this.children)n.parent=null;this.children=[];this.append(...nodes)}
 focus(){document.activeElement=this}
 querySelector(){return this.children.find(n=>n.tag==='button')}
}
const inspectors=[];
mock.module('../ui/components.mjs',{exports:{node:(tag,text)=>new El(tag,text),actionVisible:el=>el.isConnected&&!el.hidden,inspector:(host,{onClose})=>{const cache=new Map();const api={close:onClose,show({key,content}){if(!cache.has(key))cache.set(key,content);host.replaceChildren(cache.get(key))}};inspectors.push(api);return api}}});
mock.module('../ui/record-presentation.mjs',{exports:{table:()=>{const viewport=new El('table'),body=new El('tbody');viewport.append(body);return {viewport,body,row:()=>{const row=new El('tr'),cells=Array.from({length:4},()=>new El('td'));row.append(...cells);return {row,cells}}}},detail:()=>new El('div'),description:()=>''}});
mock.module('../ui/inspection-focus.mjs',{exports:{mount:()=>({open(){},close(){return false},dispose(){}})}});
mock.module('../ui/record-results.mjs',{exports:{traceView:(q,id)=>{if(!id)throw Error('A valid trace ID is required.');return {queries:[{...q,where:{trace:id}}]}}}});
const {mount}=await import('../ui/kept-saved.mjs');
const hash='a'.repeat(64),receipt=()=>({version:'goo/kept-observations/v1',id:hash,digest:hash,intentDigest:hash,connection:hash,keptAt:new Date(200000).toISOString(),selector:{signal:'logs',environment:'prod'},members:[1,2].map(i=>({key:hash+':'+i,fromMs:100000,toMs:160000,observedAt:new Date(200000).toISOString(),basis:hash,record:{timestamp_ns:'150000000000',Body:'original',TraceId:'c'.repeat(32),ResourceAttributes:{'deployment.environment.name':'prod'}}}))});
const find=(root,text)=>{for(const n of root.children){if(n.textContent===text)return n;const found=find(n,text);if(found)return found}};
test('kept inspector continuation detaches receipt, refuses retired member controls and restores same inspector without read',()=>{
 globalThis.document={hidden:false,activeElement:null};globalThis.scrollY=123;globalThis.scrollTo=()=>{};
 try{const host=new El('section');host.root=true;let active=true,calls=[];const owner=mount(host,{active:()=>active,onOpenTraceView:(view,source)=>{calls.push({view,source});active=false;return true}}),v=receipt();owner.show(v);v.members[0].record.TraceId='mutated';
 const inspect=find(host,'Inspect kept record');inspect.onclick();const action=find(host,'Open trace as view');action.onclick();assert.equal(calls.length,1);assert.equal(calls[0].view.queries[0].where.trace,'c'.repeat(32));assert(calls[0].source.current());assert.equal(calls[0].source.onBack(),false);active=true;assert(calls[0].source.onBack());assert.equal(document.activeElement,action);
 assert.equal(inspectors.at(-1).close(),false);inspect.onclick();const repeated=find(host,'Open trace as view');assert.notEqual(repeated,action);repeated.onclick();assert.equal(calls.length,2);active=true;owner.show(receipt());assert.equal(calls[0].source.current(),false);action.onclick();assert.equal(calls.length,2);find(host,'Inspect kept record').onclick();const fresh=find(host,'Open trace as view');document.hidden=true;fresh.onclick();assert.equal(calls.length,2);document.hidden=false;owner.clear();fresh.onclick();assert.equal(calls.length,2);
 }finally{delete globalThis.document;delete globalThis.scrollY;delete globalThis.scrollTo}
});
