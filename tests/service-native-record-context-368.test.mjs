import {test} from 'node:test';
import assert from 'node:assert/strict';
import {admitHeld} from '../ui/native-held-record-context.mjs';
import * as Environment from '../ui/record-environment.mjs';
import {measurement,detail,mount} from './support/service-native-context-368.mjs';

test('held service histogram admits only actual stream export and complete planned question',()=>{
 const r=measurement(),d=detail(r),before=structuredClone(r),out=admitHeld({measurement:r},d);
 assert.deepEqual(out,d);assert(Object.isFrozen(out.question.where.all));assert.deepEqual(r,before);assert(!Object.isFrozen(r));
 for(const mutate of [x=>x.fromNS=String(BigInt(x.fromNS)+1n),x=>x.toNS=String(BigInt(x.toNS)-1n),x=>x.fromNS=r.question.fromNS,x=>x.stream.attributes['metric.only']='foreign',x=>x.stream.resource['service.instance.id']='replacement',x=>x.question.where.all.pop(),x=>x.question.where.all.push({field:{kind:'attribute',key:'metric.only'},op:'eq',value:'not-copied'}),x=>x.question.limit++,x=>x.measurement.toMs++,x=>x.basis.coverage='complete',x=>x.readAt+='x',x=>x.grain='time-bucket']){const bad=structuredClone(d);mutate(bad);assert.throws(()=>admitHeld({measurement:r},bad))}
 const aggregated=structuredClone(r);delete aggregated.histograms;aggregated.histogramTime={buckets:[{fromNS:d.fromNS,toNS:d.toNS}]};assert.throws(()=>admitHeld({measurement:aggregated},d));
 for(const kind of ['gauge','stock']){const wrong=structuredClone(r);wrong.question.definition.kind=kind;assert.throws(()=>admitHeld({measurement:wrong},{...d,measurement:wrong.question}))}
});
test('service native exact context retains typed environment and namespace, never metric dimensions',()=>{
 for(const [selector,resource]of [[Environment.current('prod'),{'deployment.environment.name':'prod'}],[Environment.effective({mode:'named',value:'prod'}),{'deployment.environment':'prod'}],[Environment.effective({mode:'empty'}),{'deployment.environment.name':'','deployment.environment':'ignored'}],[Environment.effective({mode:'unreported'}),{}]])for(const ns of [undefined,'','team-雪'])for(const signal of ['logs','spans']){
  const r=measurement();delete r.question.definition.environment;r.question.definition.environmentSelector=selector;const stream=r.histograms.series[0].stream;delete stream.resource['deployment.environment.name'];delete stream.resource['deployment.environment'];Object.assign(stream.resource,resource);if(ns!==undefined)stream.resource['service.namespace']=ns;
  const admitted=admitHeld({measurement:r},detail(r,signal));assert.deepEqual(admitted.question.environmentSelector,selector);assert.equal(admitted.question.signal,signal);assert.deepEqual(admitted.question.where.all.map(x=>x.field.key),['service','service.namespace','service.instance.id','timestamp_ns','timestamp_ns']);assert.deepEqual(admitted.question.where.all[1].field,{kind:'resource',key:'service.namespace'});assert.equal(admitted.question.where.all[1].op,ns===undefined?'missing':'eq');if(ns!==undefined)assert.equal(admitted.question.where.all[1].value??'',ns);assert.equal(admitted.question.where.all.at(-1).value,String(BigInt(admitted.toNS)-1n));
 }
});
test('service held origin survives hidden departure and resumes existing DOM with retired action leases',()=>{
 const f=mount(),d=structuredClone(detail(f.data)),a=f.view.anchor(d),old=a.onclick,lazy=f.view.anchor(d),oldLazy=lazy.onclick;
 assert.equal(old(),true);const [admitted,source]=f.opened[0];assert.deepEqual(admitted,d);assert(source.current());assert(Object.isFrozen(source.measurement.histograms.series[0].stream));f.data.histograms.series[0].stream.attributes['metric.only']='caller-mutated';assert.equal(source.measurement.histograms.series[0].stream.attributes['metric.only'],'not-copied');const children=f.box.children,bodyChildren=f.view.host.children;
 f.box.hidden=true;assert(source.current(),'hidden parent is still the held origin');assert.equal(old(),false);assert.equal(oldLazy(),false);assert.equal(f.opened.length,1);
 f.box.hidden=false;assert(source.resume());assert.equal(f.box.children,children);assert.equal(f.view.host.children,bodyChildren);assert.equal(f.disclosure.open,true);assert.equal(f.box.selectedKey,'held-interval');assert.equal(old(),false);assert.equal(oldLazy(),false);assert.equal(a.onclick(),true);assert.equal(f.opened.length,2);
 const latest=f.opened[1][1];assert(latest.resume());a.hidden=true;assert.equal(a.onclick(),false);a.hidden=false;a.inert=true;assert.equal(a.onclick(),false);a.inert=false;
 f.invalidate();assert.equal(source.current(),false);assert.equal(latest.current(),false);assert.equal(latest.resume(),false);assert.equal(a.onclick(),false);
});
test('refused service departure restores live controls but cannot revive the saved callback',()=>{
 const f=mount({accept:false}),a=f.view.anchor(detail(f.data)),old=a.onclick;assert.equal(old(),false);assert.equal(f.opened.length,1);assert.equal(old(),false);assert.equal(f.opened.length,1);assert.equal(a.onclick(),false);assert.equal(f.opened.length,2);
 const foreign=detail(f.data);foreign.toNS=String(BigInt(foreign.toNS)-1n);assert.equal(f.view.anchor(foreign).onclick(),false);assert.equal(f.opened.length,2);
});

test('ordinary service departure renews existing and later-created controls without reviving old functions',()=>{
 const f=mount(),d=detail(f.data),a=f.view.anchor(d),old=a.onclick;const children=f.view.host.children;
 f.owner.renew();assert.equal(old(),false);assert.equal(f.opened.length,0);const b=f.view.anchor(d),oldB=b.onclick;
 f.owner.renew();assert.equal(old(),false);assert.equal(oldB(),false);assert.equal(f.view.host.children,children);assert.equal(b.onclick(),true);assert.equal(f.opened.length,1);
 const source=f.opened[0][1];f.box.isConnected=false;assert.equal(source.current(),false);assert.equal(source.resume(),false);
});
test('service context remains restricted to exact histogram exports',()=>{
 for(const kind of ['counter','gauge','stock']){const r=measurement();r.question.definition.kind=kind;r[kind+'s']=r.histograms;delete r.histograms;const f=mount({data:r});assert.equal(f.view.options.onRecordContext,undefined);assert.equal(f.view.options.bindRecordContextAction,undefined)}
});
