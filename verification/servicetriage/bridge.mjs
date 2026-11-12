import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Triage from '../../ui/service-triage-model.mjs';
const input=JSON.parse(fs.readFileSync(0,'utf8'));
globalThis.fetch=()=>{throw Error('Pure triage planning/admission must not read sources')};
const copy=x=>structuredClone(x);
const output=[];
for(const c of input){
 const p=Triage.plan(c.scope),before=copy(c),admitted=Triage.admit(p,{...c.current,question:p.current},{...c.baseline,question:p.baseline});
 const rows=Triage.rows(admitted,c.catalog),wire=rows.map(r=>({key:r.key,service:r.service,namespace:r.namespace,current:r.current,baseline:r.baseline,countDelta:r.countDelta,errorDelta:r.errorDelta,durationDelta:r.durationDelta,question:Triage.question(p,r)}));
 const orders=Object.fromEntries(['count','error-rate','mean','p95','count-change'].map(metric=>[metric,Triage.sortRows(rows,metric).map(r=>r.key)]));
 for(const mutate of [x=>x.question.fromMs++,x=>x.state='changed_between_reads',x=>x.groups.push(copy(x.groups[0])),x=>x.groups[0].count='0',x=>x.count=String(BigInt(x.count)+1n),x=>x.groups[0].spanStatus.error=String(BigInt(x.groups[0].count)+1n),x=>x.groups[0].duration.sumNs='0']){
  const bad={...copy(c.current),question:copy(p.current)};mutate(bad);assert.throws(()=>Triage.admit(p,bad,{...copy(c.baseline),question:p.baseline}));
 }
 assert.deepEqual(c,before,'pure model must not rewrite source values or catalog');
 assert.throws(()=>Triage.plan({...c.scope,fromMs:1,toMs:10000}),'negative previous window refused before read');
 output.push({plan:p,rows:wire,orders});
}
process.stdout.write(JSON.stringify(output));
