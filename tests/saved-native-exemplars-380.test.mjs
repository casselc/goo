import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Saved,fixture} from './support/saved-native-context-364.mjs';
import {admitHeld} from '../ui/native-held-record-context.mjs';
import {exemplarPage,exemplarDetail} from './support/histogram-exemplars-380.mjs';

test('Saved admits only the complete retained reported export occurrence and freezes its detached result',()=>{
 const p=exemplarPage(),d=exemplarDetail(p),before=structuredClone(p),out=Saved.admit(p,d);
 assert.deepEqual(out,d);assert(Object.isFrozen(out.question.where));assert(Object.isFrozen(out.stream.resource));assert.deepEqual(p,before);assert(!Object.isFrozen(p.measurement));
 assert.equal(out.fromNS,'1000000001');assert.equal(out.exportStartNS,'1000000000');assert.equal(out.timeNS,'0');assert.equal(out.question.where.all[1].op,'eq');assert.equal(out.question.where.all[1].value,undefined);
 assert.throws(()=>admitHeld({measurement:p.measurement},d),'other held owners must explicitly opt in');
 for(const change of [x=>x.relation='other',x=>delete x.relation,x=>x.sourceRow++,x=>x.exemplarIndex++,x=>x.exportStartNS='1',x=>x.exportEndNS='2',x=>x.timeNS='1',x=>x.traceId='d'.repeat(32),x=>x.spanId='d'.repeat(16),x=>delete x.spanId,x=>x.fromNS='1000000000',x=>x.question.where.all.pop(),x=>x.stream.resource['service.namespace']='other',x=>x.basis='d'.repeat(64),x=>x.readAt='other',x=>x.measurement.fromMs++,x=>x.foreign=true]){
  const bad=structuredClone(d);change(bad);assert.throws(()=>Saved.admit(p,bad));
 }
});
test('Saved missing, changed or ambiguous retained exemplar metadata cannot authorize navigation',()=>{
 for(const change of [p=>delete p.measurement.histogramExemplars,p=>p.measurement.rows.push(structuredClone(p.measurement.rows[0])),p=>p.measurement.histograms.series.push(structuredClone(p.measurement.histograms.series[0])),p=>p.measurement.histogramExemplars.exports[0].items[0].traceId='d'.repeat(32),p=>p.measurement.rows[0].exemplar_total='2',p=>p.measurement.histogramExemplars.exports[0].sourceRow=1,p=>p.receipt.question.measurement.toMs++]){
  const p=exemplarPage(),d=exemplarDetail(p);change(p);assert.throws(()=>Saved.admit(p,d));
 }
 const p=exemplarPage(),d=exemplarDetail(p);p.measurement.histograms.series[0].intervals[1].state='unknown';delete p.measurement.histograms.series[0].intervals[1].distribution;assert.deepEqual(Saved.admit(p,d),d,'a reported association is independent from distribution validity');
});
function event(){return{prevented:0,stopped:0,button:1,preventDefault(){this.prevented++},stopImmediatePropagation(){this.stopped++},stopPropagation(){}}}
test('Saved exemplar departure retires all old gestures and Back rebinds both handlers without changing held evidence',()=>{
 const f=fixture(),p=exemplarPage();f.owner.show(p);const m=f.render();assert.equal(m.options.heldExemplarContext,true);m.detail=exemplarDetail(p);const a=m.anchor();a.dataset.nativeExemplar='0';a.onauxclick=a.onclick;const old=a.onclick;assert.equal(old(),true);const source=f.opened[0][1],dom=f.host.children;
 f.owner.suspend();for(const callback of [old,a.onauxclick]){const e=event();assert.equal(callback(e),false);assert.equal(e.prevented,1);assert.equal(e.stopped,1)}assert.equal(source.current(),true);assert.equal(f.owner.resume(),true);assert.equal(f.host.children,dom);assert.equal(f.render(),m);assert.notEqual(a.onclick,old);assert.equal(a.onauxclick,a.onclick);assert.equal(a.onclick(),true);assert.equal(f.opened.length,2);
 for(const mode of ['hidden','inert','detached','inactive']){if(mode==='inactive')f.active(false);else if(mode==='detached')f.host.children=[];else a[mode]=true;const e=event();assert.equal(a.onauxclick(e),false);assert.equal(e.prevented,1);if(mode==='inactive')f.active(true);else if(mode==='detached')f.host.children=dom;else a[mode]=false}
 f.owner.show(exemplarPage());const e=event();assert.equal(a.onclick(e),false);assert.equal(e.prevented,1);assert.equal(source.current(),false);
});
