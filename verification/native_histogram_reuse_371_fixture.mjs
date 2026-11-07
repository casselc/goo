import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Native from '../ui/native-histogram-declaration.mjs';
const input=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(input);
globalThis.fetch=()=>{throw Error('Display configuration attempted a source read')};
const reused=Native.reuse(input.source,input.destination);
assert.equal(Object.hasOwn(reused,'id'),false);
assert.deepEqual(reused.question,input.destination);
assert.deepEqual(reused.declaration,input.source.declaration);
assert.equal(reused.title,input.source.title);
let refusals=0;
for(const mutate of [
 q=>q.definition.kind='counter',q=>q.definition.grain='consumer-group',
 q=>q.definition.histogramBucketMs=1000,q=>q.definition.source.resource={kind:'host',id:'host-a'},
 q=>{q.definition.source.unit=q.definition.source.unit==='s'?'ms':'s';if(q.definition.source.exactStream)q.definition.source.exactStream.unit=q.definition.source.unit},
]){const q=structuredClone(input.destination);mutate(q);assert.throws(()=>Native.reuse(input.source,q));refusals++}
for(const mutate of [
 v=>v.declaration.display.value='count',v=>v.declaration.display.requiresComplete=false,
 v=>v.declaration.selection.kind='time-bucket',v=>v.declaration.selection.intervalIndex=0,
 v=>v.declaration.display.bins=[],v=>v.basis='donor-evidence',
]){const source=structuredClone(input.source);mutate(source);assert.throws(()=>Native.reuse(source,input.destination));refusals++}
assert.equal(JSON.stringify(input),before);
process.stdout.write(JSON.stringify({reused,refusals,unchanged:true}));
