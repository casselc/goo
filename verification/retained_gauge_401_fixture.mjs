import fs from 'node:fs';
import assert from 'node:assert/strict';
import {plan} from '../ui/native-observation-context.mjs';
import {admitHeld} from '../ui/native-held-record-context.mjs';
import {admit} from '../ui/saved-native-record-context.mjs';
import * as SavedProtocol from '../ui/saved-capture-protocol.mjs';
const vectors=JSON.parse(fs.readFileSync(0,'utf8')),norm=value=>{const n=structuredClone(value);if(n.question?.environmentSelector&&n.question.environment==='')delete n.question.environment;return n};
globalThis.fetch=()=>{throw Error('Retained gauge admission must not read telemetry')};
let plans=0,controls=0;
for(const v of vectors){const before=JSON.stringify(v);SavedProtocol.validateVisit(v.page,v.page.receipt.id,'');
 for(const p of v.plans){const computed=plan(v.result,v.stream,p.question.signal,v.atNS);assert.deepEqual(computed,norm(p));const detail={...computed,measurement:v.result.question,stream:v.stream,basis:v.result.basis,readAt:v.result.readAt};
  assert.deepEqual(admitHeld({measurement:v.result},detail),detail);assert.deepEqual(admit(v.page,detail),detail);plans++;
  for(const mutate of [x=>x.atNS=String(BigInt(x.atNS)+1n),x=>x.fromNS=String(BigInt(x.fromNS)+1n),x=>x.toNS=String(BigInt(x.toNS)+1n),x=>x.radiusNS='1',x=>x.relation='point-context',x=>x.grain='point-context',x=>x.question.toMs++,x=>x.measurement.toMs++,x=>x.stream.resource['service.instance.id']='foreign',x=>x.basis='foreign',x=>x.readAt='foreign']){const bad=structuredClone(detail);mutate(bad);assert.throws(()=>admitHeld({measurement:v.result},bad));assert.throws(()=>admit(v.page,bad));controls++;}
  for(const mutate of [x=>x.gauges.series.push(structuredClone(x.gauges.series[0])),x=>x.gauges.series[0].samples.push(structuredClone(x.gauges.series[0].samples[0])),x=>x.gauges.series[0].samples[0].atNS=String(BigInt(v.atNS)+1n)]){const bad=structuredClone(v.result);mutate(bad);assert.throws(()=>admitHeld({measurement:bad},detail));controls++;}
  for(const mutate of [x=>x.receipt.title='',x=>x.receipt.capturedAt='not-a-clock',x=>x.receipt.question.measurement.toMs++,x=>x.receipt.id='bad',x=>x.explore={}]){const bad=structuredClone(v.page);mutate(bad);assert.throws(()=>admit(bad,detail));controls++;}
 }
 const foreign=structuredClone(v.page);foreign.receipt.id='b'.repeat(64);assert.throws(()=>SavedProtocol.validateVisit(foreign,v.page.receipt.id,''));controls++;
 assert.equal(JSON.stringify(v),before);
}
process.stdout.write(JSON.stringify({Cases:vectors.length,Plans:plans,Controls:controls}));
