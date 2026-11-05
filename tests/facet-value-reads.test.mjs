import assert from 'node:assert/strict';
import {test,mock} from 'node:test';
let requests=[];
mock.module('../ui/queries.mjs',{exports:{postJSON:(path,question,signal)=>new Promise(resolve=>requests.push({path,question:structuredClone(question),signal,resolve}))}});
const Reads=await import('../ui/facet-value-reads.mjs');
const field={kind:'attribute',key:'http.request.method'},clock='2026-11-05T12:00:00.000Z';
const question=(baseline=false)=>({signal:'logs',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},fromMs:baseline?1000:61000,toMs:baseline?61000:121000,limit:baseline?50:100,offset:0,groupBy:[{kind:'column',key:'service'}],where:{field:{kind:'resource',key:'service.name'},op:'eq',value:baseline?'worker':'api'}});
const scope=()=>({selected:question(),baseline:question(true)});
const response=(request,value='POST',count='2')=>({ok:true,data:{version:'goo/counted-values/v1',question:structuredClone(request.question),state:'observed',count,groups:[{g0:value,count}],capturedAt:clock,basis:'a'.repeat(64),reads:{count:{readAt:clock,basis:'b'.repeat(64)},groups:{readAt:'2026-11-05T12:00:00.001Z',basis:'c'.repeat(64)}}}});
function fixture({now=scope(),initial=null,onPause}={}){
 requests=[];const states=[];let loader,paused=0;
 const editor={setValueLoader(value){loader=value},setFacetValues(value){states.push(value)}};
 const f={now,states,get loader(){return loader},get paused(){return paused}};
 f.owner=Reads.mount(editor,{current:()=>f.now,pause:()=>{paused++;onPause?.(f)},initial});return f;
}
const tick=()=>new Promise(resolve=>setImmediate(resolve));

test('one explicit facet read independently acquires both complete original scopes',async()=>{
 const f=fixture(),original=structuredClone(f.now),pending=f.loader(field);assert.equal(requests.length,2);assert.equal(f.paused,1);
 for(const [i,side]of ['selected','baseline'].entries()){
  const request=requests[i];assert.equal(request.path,'/api/explore/counted-values');assert.deepEqual(request.question.where,original[side].where);assert.deepEqual(request.question.environmentSelector,original[side].environmentSelector);assert.equal(request.question.fromMs,original[side].fromMs);assert.equal(request.question.toMs,original[side].toMs);assert.equal(request.question.limit,1);assert.equal(request.question.offset,0);assert.deepEqual(request.question.groupBy,[field]);
 }
 requests[1].resolve(response(requests[1],'DELETE','3'));await tick();assert.equal(f.states.at(-1).baseline.state,'ready');assert.equal(f.states.at(-1).values.state,'loading');
 requests[0].resolve(response(requests[0]));await pending;assert.equal(f.states.at(-1).values.total,'2');assert.equal(f.states.at(-1).baseline.total,'3');assert.deepEqual(f.now,original);f.owner.dispose();
});

test('failed side retries alone and retains the independent healthy witness and clock',async()=>{
 const f=fixture(),pending=f.loader(field);requests[0].resolve(response(requests[0]));requests[1].resolve({ok:false,data:{error:'Baseline unavailable'}});await pending;
 const healthy=structuredClone(f.states.at(-1).values);assert.equal(f.states.at(-1).baseline.state,'error');assert.equal(f.states.at(-1).baseline.error,'Baseline unavailable');
 await f.owner.retry('selected');await f.owner.retry('other');assert.equal(requests.length,2);
 const retry=f.owner.retry('baseline');assert.equal(requests.length,3);assert.equal(f.states.at(-1).baseline.state,'loading');assert.deepEqual(f.states.at(-1).values,healthy);await f.owner.retry('baseline');assert.equal(requests.length,3);
 requests[2].resolve(response(requests[2],'DELETE'));await retry;assert.deepEqual(f.states.at(-1).values,healthy);assert.equal(f.states.at(-1).baseline.state,'ready');assert.equal(f.paused,2);f.owner.dispose();
});

test('explicit reread replaces both sides rather than combining a new side with an old partition',async()=>{
 const f=fixture(),first=f.loader(field);requests[0].resolve(response(requests[0]));requests[1].resolve(response(requests[1]));await first;
 const second=f.loader(field);assert.equal(requests.length,4);assert.equal(f.states.at(-1).values.state,'loading');assert.equal(f.states.at(-1).baseline.state,'loading');requests[2].resolve(response(requests[2]));requests[3].resolve(response(requests[3]));await second;f.owner.dispose();
});

test('changing only the baseline retires both pending reads; A-B-A cannot revive their callbacks',async()=>{
 const f=fixture(),a=structuredClone(f.now),oldLoader=f.loader,pending=oldLoader(field),old=requests.slice();f.now={...f.now,baseline:{...f.now.baseline,where:{field:{kind:'attribute',key:'tenant'},op:'eq',value:'other'}}};f.owner.sync();assert(old.every(request=>request.signal.aborted));assert.equal(f.states.at(-1),undefined);assert.equal(requests.length,2);
 f.now=a;old[0].resolve(response(old[0]));old[1].resolve(response(old[1]));await pending;assert.equal(f.states.at(-1),undefined);assert.equal(requests.length,2);
 f.owner.dispose();await oldLoader(field);await f.owner.retry('baseline');assert.equal(requests.length,2);
});

test('a whole-pair change at response admission refuses even without a preceding sync',async()=>{
 const f=fixture(),pending=f.loader(field),old=requests.slice();f.now={...f.now,baseline:{...f.now.baseline,limit:51}};old[0].resolve(response(old[0]));await tick();assert(old.every(request=>request.signal.aborted));old[1].resolve(response(old[1]));await pending;assert.equal(f.states.at(-1),undefined);assert.equal(f.states.some(value=>value?.values.state==='ready'||value?.baseline?.state==='ready'),false);f.owner.dispose();
});

test('restoring a retained healthy/error pair enables one failed-side retry without rereading the healthy side',async()=>{
 const f=fixture(),pending=f.loader(field);requests[0].resolve(response(requests[0]));requests[1].resolve({ok:false,data:{error:'Retry this operand'}});await pending;
 const retained={scope:structuredClone(f.now),field:structuredClone(field),...structuredClone(f.states.at(-1))},oldLoader=f.loader;f.owner.dispose();const restored=fixture({now:structuredClone(retained.scope),initial:retained});assert.equal(requests.length,0);assert.equal(restored.states.length,0);
 retained.values.items[0].count='999';await oldLoader(field);assert.equal(requests.length,0);const retry=restored.owner.retry('baseline');assert.equal(requests.length,1);assert.equal(restored.states.at(-1).values.items[0].count,'2');assert.equal(restored.states.at(-1).baseline.state,'loading');requests[0].resolve(response(requests[0],'DELETE'));await retry;assert.equal(restored.states.at(-1).values.readAt,clock);assert.equal(restored.states.at(-1).baseline.state,'ready');restored.owner.dispose();
});

test('retained errors cannot retry against a different current pair or a departed owner',async()=>{
 const original=scope(),initial={scope:original,field,values:{state:'error',error:'old'},baseline:{state:'error',error:'old'}};
 const f=fixture({initial,now:{...scope(),baseline:{...question(true),limit:51}}});await f.owner.retry('baseline');assert.equal(requests.length,0);f.now=original;f.owner.dispose();await f.owner.retry('baseline');assert.equal(requests.length,0);
});

test('ordinary facets still make one explicit counted-values request and dispose rejects late responses',async()=>{
 const f=fixture({now:{selected:question()}}),pending=f.loader(field);assert.equal(requests.length,1);assert.equal(f.states.at(-1).baseline,undefined);const request=requests[0];f.owner.dispose();assert(request.signal.aborted);const before=f.states.length;request.resolve(response(request));await pending;assert.equal(f.states.length,before);assert.equal(f.loader,null);
});

test('reentrant pause retirement cannot restart a departed read or retry a cleared bound scope',async()=>{
 const initial={scope:scope(),field,values:{state:'error',error:'old'},baseline:{state:'error',error:'old'}};
 for(const action of ['load','retry']){const f=fixture({initial,onPause:f=>f.owner.dispose()});await (action==='load'?f.loader(field):f.owner.retry('baseline'));assert.equal(requests.length,0);assert.equal(f.loader,null)}
 const f=fixture({initial,onPause:f=>f.owner.invalidate()});await f.owner.retry('baseline');assert.equal(requests.length,0);assert.equal(f.states.at(-1),undefined);f.owner.dispose();
});

test('read identity binds original policies, full predicates, grouping, limits and qualified field',()=>{
 const original=scope(),key=Reads.identity(original,field);
 for(const change of [s=>s.baseline.limit++,s=>s.baseline.where.value='different',s=>s.baseline.groupBy=[{kind:'column',key:'scope'}],s=>s.baseline.environmentSelector.policy='current-only',s=>s.selected.toMs++]){const next=structuredClone(original);change(next);assert.notEqual(Reads.identity(next,field),key)}
 assert.notEqual(Reads.identity(original,{kind:'resource',key:field.key}),key);assert.throws(()=>Reads.identity({selected:question(),baseline:null},{kind:'column',key:'duration_ns'}));
});
