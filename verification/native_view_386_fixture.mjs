import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Native from '../ui/native-view-declaration.mjs';
const fixtures=JSON.parse(fs.readFileSync(0,'utf8'));let refusals=0;
globalThis.fetch=()=>{throw Error('Display admission/reuse must not read telemetry')};
for(const f of fixtures){const before=JSON.stringify(f);assert.deepEqual(Native.page(f.Page),f.Page);for(const mutate of [p=>delete p.items[0].family,p=>delete p.items[0].selection,p=>p.items[0].selection="time-bucket",p=>p.items[0].kind="pie"]){const p=structuredClone(f.Page);mutate(p);assert.throws(()=>Native.page(p));refusals++}const reused=Native.reuse(f.Source,f.Destination);assert.deepEqual(reused,f.Reused);assert.deepEqual(reused.question,f.Destination);assert.deepEqual(reused.declaration,f.Source.declaration);assert.equal(Object.hasOwn(reused,'id'),false);
 for(const bad of f.Invalid){assert.throws(()=>Native.view(bad));refusals++}
 for(const mutate of [q=>q.definition.kind=q.definition.kind==='gauge'?'histogram':'gauge',q=>q.definition.grain='consumer-group',q=>{q.definition.source.unit+='different';q.definition.source.exactStream.unit=q.definition.source.unit},q=>q.definition.source.resource={kind:'host',id:'host-a'}]){const q=structuredClone(f.Destination);mutate(q);assert.throws(()=>Native.reuse(f.Source,q));refusals++}
 reused.question.definition.source.attributes['only.destination']='mutated';reused.question.definition.source.exactStream.resource['service.instance.id']='mutated';reused.question.definition.source.services[0]='mutated';reused.question.definition.environmentSelector.scope.value='mutated';reused.declaration.kind.name='mutated';assert.equal(JSON.stringify(f),before);
}
process.stdout.write(JSON.stringify({examples:fixtures.length,refusals}));
