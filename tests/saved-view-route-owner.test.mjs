import assert from 'node:assert/strict';
import {test} from 'node:test';
import {fixture,tick} from './support/saved-view-owner.mjs';
const a='a'.repeat(64),b='b'.repeat(64);
test('saved-view paired history notifications acquire one definition per visit',async()=>{
 const f=fixture();await tick();f.route(a);assert.equal(f.reads.length,1);f.reads[0].resolve();await tick();assert.equal(f.adoptions.length,1);const status=f.panel().textContent;f.emit('hashchange');f.emit('popstate');assert.equal(f.reads.length,1);assert.equal(f.panel().textContent,status);f.owner.dispose();
});
test('saved definition departure inside Explore retires before abort and refuses late adoption',async()=>{
 const f=fixture();await tick();f.route(a,['hashchange']);let observedAbort=false;f.reads[0].signal.addEventListener('abort',()=>{observedAbort=true;f.reads[0].resolve()});location.hash='#question=other';f.emit();await tick();assert(observedAbort);assert.equal(f.adoptions.length,0);f.route(b);f.reads.at(-1).resolve();await tick();assert.equal(f.adoptions.length,1);f.owner.dispose();
});
test('queued departures and A-B-A visits cannot renew an old definition lease',async()=>{
 const f=fixture();await tick();f.route(a,['hashchange']);const old=f.reads[0];f.emit('hashchange','#question=departed');assert(old.signal.aborted);assert.equal(f.reads.length,2);old.resolve();await tick();assert.equal(f.adoptions.length,0);f.reads[1].resolve();await tick();assert.equal(f.adoptions.length,1);
 f.route(b);const pendingB=f.reads.at(-1);f.route(a);assert(pendingB.signal.aborted);pendingB.resolve();await tick();assert.equal(f.adoptions.length,1);f.reads.at(-1).resolve();await tick();assert.equal(f.adoptions.length,2);f.owner.dispose();
});
test('metadata failures retry explicitly once, with old controls retired and focus retained only by the retry task',async()=>{
 const f=fixture();await tick();f.route(a);f.reads[0].reject();await tick();const retry=f.panel().children.at(-1);assert.equal(retry.textContent,'Retry opening view');f.emit();assert.equal(f.reads.length,1);retry.focus();retry.onclick();retry.onclick();assert.equal(f.reads.length,2);assert.equal(document.activeElement,f.panel());f.reads[1].reject();await tick();const next=f.panel().children.at(-1);assert.notEqual(next,retry);assert.equal(document.activeElement,next);
 next.onclick();const other={};document.activeElement=other;f.reads[2].reject();await tick();assert.equal(document.activeElement,other);f.route(b);retry.onclick();next.onclick();assert.equal(f.reads.length,4);f.owner.dispose();f.reads.at(-1).resolve();await tick();assert.equal(f.adoptions.length,0);
});
test('wrong definition identity refuses admission and passive history is not a retry',async()=>{
 const f=fixture();await tick();f.route(a);f.reads[0].resolve(b);await tick();assert.equal(f.adoptions.length,0);assert.match(f.panel().textContent,/identity differs/);f.emit('popstate');f.emit('hashchange');assert.equal(f.reads.length,1);f.owner.dispose();
});
test('pending departure to plain Explore releases obsolete loading UI without telemetry reads',async()=>{
 const f=fixture();await tick();f.route(a);assert.equal(f.panel().hidden,false);f.route(null);assert.equal(f.panel().hidden,true);assert(f.reads[0].signal.aborted);f.reads[0].resolve();await tick();assert.equal(f.adoptions.length,0);assert(f.reads.every(r=>r.method==='GET'&&r.url.startsWith('/api/exploration-views/')));f.owner.dispose();
});
test('suspend and dispose retire requests including the queued initial restore',async()=>{
 const f=fixture();await tick();f.route(a);f.owner.suspend();assert(f.reads[0].signal.aborted);assert.equal(f.panel().hidden,true);assert.equal(f.panel().children.length,0);f.reads[0].resolve();await tick();assert.equal(f.adoptions.length,0);f.emit();assert.equal(f.reads.length,1);f.owner.dispose();f.emit();assert.equal(f.reads.length,1);
 const next=fixture();location.hash='#exploration-view='+b;next.owner.dispose();await tick();assert.equal(next.reads.length,0);
});
test('admitted parent callback cannot adopt after reentrant invalidation or disposal',async()=>{
 for(const method of ['invalidate','dispose']){
  let f;f=fixture({onAdopt:()=>f.owner[method]()});await tick();f.route(a);f.reads[0].resolve();await tick();assert.equal(f.adoptions.length,1);assert.equal(f.panel().hidden,true);assert(f.reads[0].signal.aborted);assert.notEqual(f.panel().dataset.phase,'applied');f.emit();assert.equal(f.reads.length,1);f.owner.dispose();
 }
});
test('workspace departure clears even a completed visit marker for intentional return',async()=>{
 const f=fixture();await tick();f.route(a);f.reads[0].resolve();await tick();assert.equal(f.adoptions.length,1);location.hash='#services';f.workspace('services');location.hash='#exploration-view='+a;f.workspace('explore');f.emit();assert.equal(f.reads.length,2);f.owner.dispose();
});

test('secondary entry controls retain identity while active population owns them, then return and dispose without orphans',async()=>{const f=fixture({externalControls:true});await tick();const controls=f.controlsHost.children[0],library=controls.children.find(el=>el.tagName==='details');assert(controls);assert(library);assert.equal(controls.parentElement,f.controlsHost);assert.equal(f.host.contains(controls),false);f.route(a);assert.equal(f.reads.length,1);assert.equal(controls.parentElement,f.host);assert.equal(f.controlsHost.contains(controls),false);assert.equal(f.host.children.indexOf(controls)<f.host.children.indexOf(f.panel()),true);assert.equal(controls.children.includes(library),true);f.owner.invalidate();assert.equal(controls.parentElement,f.controlsHost);assert.equal(f.host.contains(controls),false);assert.equal(controls.children.includes(library),true);assert(f.reads[0].signal.aborted);f.reads[0].resolve();await tick();assert.equal(controls.parentElement,f.controlsHost);assert.equal(f.adoptions.length,0);assert(f.reads.every(read=>read.method==='GET'&&read.url.startsWith('/api/exploration-views/')));f.owner.dispose();assert.equal(controls.parentElement,null);assert.equal(f.controlsHost.children.length,0);assert.deepEqual(f.host.children,[f.panel()]);assert.equal(f.panel().hidden,true);assert.equal(f.panel().children.length,0)});
