import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Versions from '../../ui/service-version-comparison.mjs';
globalThis.fetch=()=>{throw Error('Version projection acquired a source')};
const input=JSON.parse(fs.readFileSync(0,'utf8')),before=structuredClone(input);
const output=input.map(c=>{
 const question=Versions.discoveryQuestion({query:c.parent},c.parentData,c.context);
 const accepted=Versions.admit(question,c.discovery);
 const count=Versions.comparison(c.parent,accepted,c.baseline,c.selection,'count');
 const mean=Versions.comparison(c.parent,accepted,c.baseline,c.selection,'duration.mean');
 const empty={...structuredClone(c.discovery),state:'no_data',count:'0',groups:[]};
 assert.equal(Versions.admit(question,empty).state,'no_data');
 assert.throws(()=>Versions.comparison(c.parent,empty,c.baseline,c.selection));
 for(const mutate of [d=>d.question.toMs++,d=>d.state='changed_between_reads',d=>d.count=String(BigInt(d.count)+1n),d=>d.groups.push(structuredClone(d.groups[0])),d=>d.groups[0].count='01',d=>d.groups[0].g0=1]){
  const bad=structuredClone(c.discovery);mutate(bad);assert.throws(()=>Versions.admit(question,bad));
 }
 assert.throws(()=>Versions.comparison(c.parent,accepted,c.selection,c.selection));
 assert.throws(()=>Versions.comparison(c.parent,accepted,'not-in-observed-values',c.selection));
 assert.throws(()=>Versions.comparison(c.parent,accepted,c.baseline,c.selection,'p95'));
 for(const change of [x=>x.service+='-foreign',x=>x.scope.fromMs++,x=>x.scope.namespace=null]){
  const context=structuredClone(c.context);change(context);assert.throws(()=>Versions.discoveryQuestion({query:c.parent},c.parentData,context));
 }
 assert.throws(()=>Versions.discoveryQuestion({query:c.parent},{...c.parentData,state:'changed_between_reads'},c.context));
 return {question,count,mean};
});
assert.deepEqual(input,before,'version helpers mutated caller configuration or source data');
process.stdout.write(JSON.stringify(output));
