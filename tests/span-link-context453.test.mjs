import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class Element{
 constructor(tag,text=''){this.tag=tag;this.textContent=text;this.children=[];this.hidden=false;this.disabled=false;this.isConnected=true;this.dataset={};this.style={};this.classList={contains:()=>false};this.attributes={};this.scrollTop=0;}
 append(...children){for(const child of children){if(typeof child==='string')this.textContent+=child;else{this.children.push(child);child.parentElement=this}}}
 replaceChildren(...children){this.children=[];this.append(...children)}
 setAttribute(key,value){this.attributes[key]=value}
 focus(){globalThis.focused=this;if(globalThis.document)document.activeElement=this}
}
mock.module('../ui/components.mjs',{exports:{node:(...args)=>new Element(...args)}});
mock.module('../ui/dom-visibility.mjs',{exports:{nativeControlReady:element=>element.isConnected&&!element.disabled&&!element.hidden,revealFocus:element=>element.focus()}});
mock.module('../ui/record-presentation.mjs',{exports:{table(){return {body:new Element('tbody'),viewport:new Element('goo-table'),row(){return {row:new Element('tr'),cells:[new Element('td'),new Element('td'),new Element('td')]}}}},attributes(){}}});
mock.module('../ui/source-refusal-view.mjs',{exports:{renderRefusal:(host,error)=>{host.textContent=error.message}}});
const {mount,admitResult}=await import('../ui/span-link-context.mjs');

const target={traceId:'a'.repeat(32),spanId:'b'.repeat(16)},q={signal:'spans',environment:'prod',fromMs:1000,toMs:2000,linksTo:target,limit:2,offset:0};
const intent={relation:'reported-span-link-incoming',question:q,target,origin:{source:target}};
const row={TraceId:'c'.repeat(32),SpanId:'d'.repeat(16),timestamp_ns:'1500000000',ResourceAttributes:{'deployment.environment.name':'prod'},linked_traces:[target.traceId],linked_spans:[target.spanId],linked_attributes:[{}]};
function setup(){const parent=new Element('dialog'),sibling=new Element('div'),host=new Element('section'),opener=new Element('button');parent.append(sibling,host);const reads=[];let current=true;const owner=mount(host,{current:()=>current,fetchQuery:(q,signal)=>new Promise((resolve,reject)=>reads.push({q,signal,resolve,reject}))});const controls=()=>host.children[1];return {owner,reads,host,sibling,opener,run:()=>controls().children[1],previous:()=>controls().children[2].children[0],next:()=>controls().children[2].children[1],back:()=>controls().children[0],withdraw(){current=false}};}
const complete=(x,i,{rows=[row],count='3',more=false,state='observed'}={})=>x.reads[i].resolve({question:structuredClone(x.reads[i].q),rows,count,hasMore:more,state,basis:'basis-'+i,capturedAt:'2026-11-15T00:00:00Z'});
test('incoming result admission checks same link position, environment, time and exact page question',()=>{
 assert.equal(admitResult({question:q,rows:[row]},intent).rows.length,1);
 for(const change of [{linked_traces:[target.traceId,'e'.repeat(32)],linked_spans:['f'.repeat(16),target.spanId],linked_attributes:[{},{}]},{linked_attributes:[]},{ResourceAttributes:{'deployment.environment.name':'other'}},{timestamp_ns:'2000000000'}])assert.throws(()=>admitResult({question:q,rows:[{...row,...change}]},intent));
 assert.throws(()=>admitResult({question:{...q,offset:2},rows:[row]},intent));
});
test('incoming stage and explicit page reads retain selector; failed page retries same offset',async()=>{
 const x=setup();x.owner.stage({intent,opener:x.opener});assert.equal(x.reads.length,0);assert.equal(x.host.children[0].textContent,'Referencing spans');assert.equal(x.host.children[6].children[0].textContent,'Question and reference target');assert.equal(x.run().textContent,'Run reference search');
 let pending=x.run().onclick();complete(x,0,{rows:[row,row],more:true});await pending;assert.equal(x.next().disabled,false);assert.equal(x.previous().disabled,true);assert.match(x.host.children[4].textContent,/1–2 shown/);
 pending=x.next().onclick();assert.deepEqual(x.reads[1].q,{...q,offset:2});x.reads[1].reject(Error('source refused'));await pending;assert.equal(x.host.dataset.phase,'unavailable');assert.equal(JSON.parse(x.host.children[6].children[1].textContent).lookup.question.offset,2);
 pending=x.run().onclick();assert.deepEqual(x.reads[2].q,{...q,offset:2});complete(x,2);await pending;assert.equal(x.next().disabled,true);assert.equal(x.previous().disabled,false);assert.match(x.host.children[4].textContent,/3–3 shown/);assert.match(x.host.children[4].textContent,/Fresh read; pages may change/);
 pending=x.previous().onclick();assert.equal(x.reads[3].q.offset,0);complete(x,3,{rows:[row,row],more:true});await pending;const stale=x.next();x.back().onclick();stale.onclick();assert.equal(x.reads.length,4);assert.equal(x.host.hidden,true);assert.equal(x.sibling.hidden,false);
});
test('paging at the budget boundary refuses another read and Back aborts a pending page',async()=>{
 const x=setup(),bounded={...intent,question:{...q,offset:1000,limit:100}};x.owner.stage({intent:bounded,opener:x.opener});let pending=x.run().onclick();complete(x,0,{more:true,count:'2000'});await pending;assert.equal(x.next().disabled,true);x.next().onclick();assert.equal(x.reads.length,1);assert.match(x.host.children[4].textContent,/Paging limit reached/);
 pending=x.previous().onclick();assert.equal(x.reads[1].q.offset,900);x.back().onclick();assert.equal(x.reads[1].signal.aborted,true);complete(x,1);await pending;assert.equal(x.host.hidden,true);assert.equal(x.host.children.length,0);
});
