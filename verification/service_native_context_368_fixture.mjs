import fs from 'node:fs';
import assert from 'node:assert/strict';
import {admitHeld} from '../ui/native-held-record-context.mjs';
import {plan} from '../ui/native-record-context.mjs';
import {mount} from '../tests/support/service-native-context-368.mjs';

const input=JSON.parse(fs.readFileSync(0,'utf8')),original=JSON.stringify(input);
let checks=0;
for(const item of input.cases){
 const {measurement,detail}=item;
 // Wire DTOs may include Go's inert blank scalar beside a typed selector.
 const expected=plan(measurement.question,detail.stream,detail.question.signal,detail.fromNS,detail.toNS);
 const wire=structuredClone(detail.question);
 if(wire.environmentSelector&&wire.environment==='')delete wire.environment;
 assert.deepEqual(wire,expected.question);
 const normalized={...structuredClone(detail),question:expected.question};
 const held={measurement};
 const accepted=admitHeld(held,normalized);
 assert.deepEqual(accepted,normalized);
 assert(Object.isFrozen(accepted));assert(Object.isFrozen(accepted.stream.resource));
 for(const mutate of [
  x=>x.question.toMs++,x=>x.question.signal='metrics',
  x=>x.stream.attributes['metric.only']='foreign',x=>x.stream.resource['service.instance.id']='foreign',
  x=>x.fromNS=String(BigInt(x.fromNS)+1n),x=>x.toNS=String(BigInt(x.toNS)+1n),
  x=>x.readAt+='x',x=>x.basis+='x',x=>x.grain='bucket',
 ]){const bad=structuredClone(normalized);mutate(bad);assert.throws(()=>admitHeld(held,bad));checks++}
 const bucket=structuredClone(held);delete bucket.measurement.histograms;bucket.measurement.histogramTime={buckets:[{fromNS:detail.fromNS,toNS:detail.toNS}]};
 assert.throws(()=>admitHeld(bucket,normalized));
 const gauge=structuredClone(held);gauge.measurement.question.definition.kind='gauge';assert.throws(()=>admitHeld(gauge,normalized));
 checks+=3;
 const f=mount({data:measurement}),anchor=f.view.anchor(normalized),retired=[];
 let paused=false,valid=true;
 const click=()=>{const before=f.opened.length,allowed=valid&&!paused&&!anchor.hidden&&!anchor.inert&&anchor.isConnected;anchor.onclick();assert.equal(f.opened.length,before+(allowed?1:0));if(allowed){paused=true;assert(f.opened.at(-1)[1].current())}checks++};
 const resume=()=>{retired.push(anchor.onclick);const children=[...f.box.children],key=f.box.selectedKey;f.owner.renew();paused=false;assert.deepEqual(f.box.children,children);assert.equal(f.box.selectedKey,key);checks++};
 const stale=()=>{const before=f.opened.length;for(const fn of retired)fn();assert.equal(f.opened.length,before);checks++};
 click();retired.push(anchor.onclick);f.box.hidden=true;assert(f.opened.at(-1)[1].current(),'hidden held source must remain current');f.box.hidden=false;resume();stale();click();resume();
 for(const action of input.actions){if(action===0)click();if(action===1)resume();if(action===2)stale();if(action===3){anchor.hidden=true;click();anchor.hidden=false}if(action===4){anchor.inert=true;click();anchor.inert=false}}
 retired.push(anchor.onclick);f.invalidate();valid=false;stale();click();for(const [,source]of f.opened)assert(!source.current());
}
assert.equal(JSON.stringify(input),original);
process.stdout.write(JSON.stringify({complete:true,checks}));
