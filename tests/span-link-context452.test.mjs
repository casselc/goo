import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class Element{
 constructor(tag,text=''){this.tag=tag;this.textContent=text;this.children=[];this.hidden=false;this.disabled=false;this.isConnected=true;this.dataset={};this.style={};this.classList={contains:()=>false};this.scrollTop=0;}
 append(...children){this.children.push(...children);for(const child of children)if(typeof child==='object')child.parentElement=this;}
 replaceChildren(...children){this.children=[];this.append(...children)}
 setAttribute(){}
 focus(){globalThis.focused=this}
}
mock.module('../ui/components.mjs',{exports:{node:(...args)=>new Element(...args)}});
mock.module('../ui/dom-visibility.mjs',{exports:{nativeControlReady:element=>element.isConnected&&!element.disabled&&!element.hidden,revealFocus:element=>element.focus()}});
mock.module('../ui/record-presentation.mjs',{exports:{table(){throw Error('not used')},attributes(){}}});
mock.module('../ui/source-refusal-view.mjs',{exports:{renderRefusal:(host,error)=>{host.textContent=error.message}}});
const {mount,admitResult}=await import('../ui/span-link-context.mjs');
const q={signal:'spans',environment:'prod',fromMs:1000,toMs:2000,where:{all:[{field:{kind:'column',key:'trace_id'},op:'eq',value:'a'.repeat(32)},{field:{kind:'column',key:'span_id'},op:'eq',value:'b'.repeat(16)}]},limit:100,offset:0};
const intent={question:q,target:{traceId:'a'.repeat(32),spanId:'b'.repeat(16)},origin:{},attributes:{}};
const row={TraceId:'a'.repeat(32),SpanId:'b'.repeat(16),timestamp_ns:'1500000000',ResourceAttributes:{'deployment.environment.name':'prod'}};
test('lookup result admission checks exact question, target, environment and half-open time',()=>{
 const data={question:q,rows:[row]};assert.equal(admitResult(data,intent),data);
 for(const change of [{TraceId:'c'.repeat(32)},{SpanId:'c'.repeat(16)},{timestamp_ns:'2000000000'},{timestamp_ns:'999999999'},{ResourceAttributes:{'deployment.environment.name':'other'}}])assert.throws(()=>admitResult({...data,rows:[{...row,...change}]},intent));
 assert.throws(()=>admitResult({...data,question:{...q,toMs:2001}},intent));
 assert.equal(admitResult({...data,rows:[row,row]},intent).rows.length,2,'duplicate observations retained');
});
function setup(){const parent=new Element('dialog'),sibling=new Element('div'),hidden=new Element('div'),host=new Element('section'),opener=new Element('button');hidden.hidden=true;parent.append(sibling,hidden,host);const reads=[];let current=true;const owner=mount(host,{current:()=>current,fetchQuery:(q,signal)=>new Promise((resolve,reject)=>reads.push({q,signal,resolve,reject}))});return {owner,reads,host,sibling,hidden,opener,withdraw(){current=false},run:()=>host.children[1].children[1],back:()=>host.children[1].children[0]};}
test('stage and Back are source-free, restore held visibility and focus; cached Run is revoked',()=>{
 const x=setup();assert.equal(x.owner.stage({intent,opener:x.opener}),true);assert.equal(x.reads.length,0);assert.equal(x.sibling.hidden,true);assert.equal(x.host.dataset.phase,'staged');const stale=x.run();x.back().onclick();assert.equal(x.sibling.hidden,false);assert.equal(x.hidden.hidden,true);assert.equal(x.host.hidden,true);assert.equal(globalThis.focused,x.opener);stale.onclick();assert.equal(x.reads.length,0);
});
test('only explicit Run reads; Back aborts and late response cannot publish',async()=>{
 const x=setup();x.owner.stage({intent,opener:x.opener});const run=x.run(),pending=run.onclick();assert.equal(x.reads.length,1);assert.deepEqual(x.reads[0].q,q);x.back().onclick();assert.equal(x.reads[0].signal.aborted,true);x.reads[0].resolve({question:q,rows:[],count:'0'});await pending;assert.equal(x.host.hidden,true);assert.equal(x.host.children.length,0);
});
test('source refusal keeps held origin and retry; no-data result is distinct from delivery failure',async()=>{
 const x=setup();x.owner.stage({intent,opener:x.opener});let pending=x.run().onclick();x.reads[0].reject(Error('source refused'));await pending;assert.equal(x.host.dataset.phase,'unavailable');assert.equal(x.run().disabled,false);pending=x.run().onclick();x.reads[1].resolve({question:q,rows:[],count:'0',basis:'basis',capturedAt:'read'});await pending;assert.equal(x.host.dataset.phase,'ready');assert.match(x.host.children[4].textContent,/absence does not establish missing delivery/);assert.equal(x.reads.length,2);const provenance=JSON.parse(x.host.children[6].children[1].textContent);assert.equal(provenance.lookup.basis,'basis');assert.equal(provenance.lookup.capturedAt,'read');assert.equal(x.host.children[5].children.at(-1).textContent,'Read read');
});
test('new stage retires old response and its controls without mutating caller intent',async()=>{
 const x=setup(),input=structuredClone(intent);x.owner.stage({intent:input,opener:x.opener});const oldRun=x.run(),pending=oldRun.onclick();input.question.toMs=9999;assert.equal(x.reads[0].q.toMs,2000);
 x.owner.back();x.owner.stage({intent,opener:x.opener});assert.equal(x.host.dataset.phase,'staged');x.reads[0].resolve({question:q,rows:[],count:'0'});await pending;assert.equal(x.host.dataset.phase,'staged');oldRun.onclick();assert.equal(x.reads.length,1);x.owner.invalidate();assert.equal(x.sibling.hidden,false);assert.equal(x.hidden.hidden,true);
});

test('an empty page from changed source reads does not claim there are no matching spans',async()=>{
 const x=setup();x.owner.stage({intent,opener:x.opener});const pending=x.run().onclick();x.reads[0].resolve({question:q,rows:[],count:'2',state:'changed_between_reads'});await pending;assert.match(x.host.children[4].textContent,/No span observations returned on this page/);assert.doesNotMatch(x.host.children[4].textContent,/No matching spans in this environment/);
});

test('Back restores held scroll before focusing an already visible opener without recentering',()=>{
 const x=setup(),parent=x.host.parentElement,previous={height:globalThis.innerHeight,width:globalThis.innerWidth};
 globalThis.innerHeight=800;globalThis.innerWidth=500;parent.scrollTop=42;parent.getBoundingClientRect=()=>({top:0,bottom:800,left:0,right:500});x.opener.getBoundingClientRect=()=>({top:100,bottom:130,left:20,right:200});let focused=false;x.opener.focus=options=>{assert.deepEqual(options,{preventScroll:true});assert.equal(parent.scrollTop,42);focused=true};
 try{x.owner.stage({intent,opener:x.opener});parent.scrollTop=0;x.owner.back();assert.equal(focused,true);assert.equal(parent.scrollTop,42)}finally{if(previous.height===undefined)delete globalThis.innerHeight;else globalThis.innerHeight=previous.height;if(previous.width===undefined)delete globalThis.innerWidth;else globalThis.innerWidth=previous.width}
});
