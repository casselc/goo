import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Groups from '../../ui/exception-groups.mjs';
globalThis.fetch=()=>{throw Error('Pure exception projection acquired a source')};
const input=JSON.parse(fs.readFileSync(0,'utf8')),before=structuredClone(input);
const result=input.map(c=>{
 const question=Groups.question(c.parent),dto=Groups.admit(question,c.dto),staged=Groups.stageQuestion(c.parent,c.fingerprint);
 for(const mutate of [d=>d.question.toMs++,d=>d.count=String(BigInt(d.count)+1n),d=>d.groups.push(structuredClone(d.groups[0])),d=>d.groups[0].count='01',d=>d.groups[0].firstNS=String(BigInt(c.parent.fromMs)*1000000n-1n),d=>d.groups[0].lastNS=String(BigInt(c.parent.toMs)*1000000n),d=>d.groups[0].affected[0].namespace=7,d=>d.groups[0].affected.push(structuredClone(d.groups[0].affected[0])),d=>d.rows=[]]){
  const bad=structuredClone(c.dto);mutate(bad);assert.throws(()=>Groups.admit(question,bad),'unsafe metadata accepted');
 }
 assert.throws(()=>Groups.stageQuestion(c.parent,'ef1:c:garbage'));
 assert.throws(()=>Groups.question({...c.parent,offset:1}));
 assert.deepEqual(dto,c.dto);
 return {question,staged};
});
assert.deepEqual(input,before,'projection rewrote its caller state');
process.stdout.write(JSON.stringify(result));
