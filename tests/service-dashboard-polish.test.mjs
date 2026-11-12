import {test} from 'node:test';
import assert from 'node:assert/strict';
import {fixture,observed,tick,walk} from './support/dashboard-owner-354.mjs';

test('service widget keeps one heading and one compact footer with existing actions and disclosed evidence',async()=>{
 const f=fixture(1);await tick();f.requests[0].resolve(observed(f.requests[0]));const owner=await f.loaded,box=f.box('logs0'),elements=walk(box);
 assert.equal(box.children.filter(child=>child.localName==='h5').length,1);
 const footers=elements.filter(element=>element.className==='dashboard-widget-footer');assert.equal(footers.length,1);
 const footer=footers[0],evidence=walk(footer).find(element=>element.className==='dashboard-widget-evidence');
 assert(evidence);assert.notEqual(evidence.open,true);assert.equal(evidence.children[0].textContent,'Scope and evidence · observed');
 for(const label of ['Open widget question','Add widget question','Capture widget evidence'])assert(walk(footer).some(element=>element.textContent===label));
 assert.equal(walk(footer).find(element=>element.textContent==='Open widget question').href,'#step=checked');
 assert.equal(owner.results[0].result.count,'2');assert.equal(owner.results[0].result.capturedAt,'2026-11-06T00:00:00Z');
 const reads=f.requests.length,metadata=f.metadata.length;evidence.open=true;evidence.open=false;assert.equal(f.requests.length,reads);assert.equal(f.metadata.length,metadata);
 owner.dispose();
});
