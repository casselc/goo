import {test} from 'node:test';
import assert from 'node:assert/strict';
import {fixture,observed,refused,tick,walk} from './support/dashboard-owner-354.mjs';

test('incoming HTTP response disclosure retains its label and tracks pending/refused reads',async()=>{
 for(const fail of [false,true]){
  const f=fixture(1),widget=f.plan.widgets[0];widget.key='http-server-responses';widget.family='http-server';widget.query.signal='spans';widget.query.groupBy=[{kind:'attribute',key:'http.response.status_code'}];
  await tick();const context=walk(f.host).find(element=>element.className==='dashboard-http-server-context'),summary=context.children.find(element=>element.localName==='summary');
  assert(context);assert.match(summary.textContent,/^Response context · 1 independent read · 1 pending$/);assert.equal(context.dataset.state,'loading');assert(f.box('http-server-responses').parentElement===context.children.find(element=>element.className==='dashboard-grid'));
  f.requests[0].resolve(fail?refused():observed(f.requests[0]));await f.loaded;
  assert.equal(summary.textContent,fail?'Response context · 1 independent read · 1 unavailable':'Response context · 1 independent read');assert.equal(context.dataset.state,fail?'partial':'loaded');
 }
});

test('incoming HTTP response handoff refuses pending and retired owners before native navigation or publication',async()=>{
 let parentReady=false;const f=fixture(2,{canLeave:()=>parentReady}),widget=f.plan.widgets[0];widget.key='http-server-responses';widget.family='http-server';widget.query.signal='spans';widget.query.groupBy=[{kind:'attribute',key:'http.response.status_code'}];
 await tick();f.requests[0].resolve(observed(f.requests[0]));await tick();const box=f.box('http-server-responses'),open=walk(box).find(element=>element.localName==='a'&&element.textContent==='Open widget question'),add=walk(box).find(element=>element.textContent==='Add widget question'),capture=walk(box).find(element=>element.textContent==='Capture widget evidence'),events=[];
 globalThis.dispatchEvent=event=>events.push(event);const attempt=()=>{const event={prevented:false,preventDefault(){this.prevented=true},stopImmediatePropagation(){}};const result=open.onclick(event);return{result,prevented:event.prevented}};
 assert.deepEqual(attempt(),{result:false,prevented:true});add.onclick();capture.onclick();assert.equal(events.length,0);assert.equal(open.getAttribute('aria-disabled'),'true');
 f.requests[1].resolve(observed(f.requests[1]));await f.loaded;assert.deepEqual(attempt(),{result:false,prevented:true},'outer service open must finish too');
 parentReady=true;assert.deepEqual(attempt(),{result:true,prevented:false});add.onclick();capture.onclick();assert.equal(events.length,2);
 for(const control of [open,add,capture]){control.inert=true;if(control===open)assert.deepEqual(attempt(),{result:false,prevented:true});else control.onclick();control.inert=false;}assert.equal(events.length,2);
 f.retire();assert.deepEqual(attempt(),{result:false,prevented:true});add.onclick();capture.onclick();assert.equal(events.length,2);
});
