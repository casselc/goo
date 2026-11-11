import fs from 'node:fs';
import assert from 'node:assert/strict';
import {validateSnapshot} from '../ui/view-schema.mjs';
import {adoptExploration} from '../ui/source-snapshots.mjs';
import {isRate,isFraction,formatCell,exactText,defaultSummaryFields} from '../ui/view-values.mjs';
import * as Templates from '../ui/view-template-handoff.mjs';
import * as Timeline from '../ui/timeline-declaration.mjs';
const vectors=JSON.parse(fs.readFileSync(0,'utf8'));
globalThis.fetch=()=>{throw Error('Rate display and reuse must not perform source reads')};
const gcd=(a,b)=>b?gcd(b,a%b):a;
const exact=(count,ms)=>{const n=BigInt(count)*1000n,d=BigInt(ms),g=gcd(n,d);return d/g===1n?String(n/g):`${n/g}/${d/g}`;};
let controls=0;
for(const v of vectors){const before=JSON.stringify(v),q=v.data.question;
 for(const d of [v.data.summaryDataset,v.data.timeDataset]){
  const valid=validateSnapshot(d);assert.equal(valid.ok,true,valid.reasons.join('; '));
  const f=d.fields.find(f=>f.key==='record.rate');assert.ok(isRate(f));assert.equal(isFraction(f),false);assert.equal(f.unit,'records/s');assert.equal(f.type,'rational');assert.equal(f.nullable,true);assert.deepEqual(f.aggregation,{family:'rate',kind:'records-per-second',method:'exact-rational'});
  assert.ok(!defaultSummaryFields(d.fields).some(f=>f.key==='record.rate'),'rate must not silently grow automatic cards');
  for(const row of d.rows){const elapsed=d.shape==='summary'?q.toMs-q.fromMs:Number(row.values.toMs)-Number(row.values.fromMs),cell=row.cells['record.rate'];assert.equal(row.values['record.rate'],exact(row.values.count,elapsed));assert.deepEqual(cell.rate,{count:row.values.count,elapsedMs:String(elapsed)});assert.equal(cell.state,'derived');assert.ok(!formatCell(f,row.values['record.rate'],cell).includes('%'));assert.ok(exactText(f,row.values['record.rate'],cell).includes(String(elapsed)));}
  for(const mutate of [x=>x.rows[0].values['record.rate']='999999999999999999999999',x=>x.rows[0].cells['record.rate'].rate.count='18446744073709551616',x=>x.rows[0].cells['record.rate'].rate.elapsedMs='0',x=>delete x.rows[0].cells['record.rate'].rate,x=>x.fields.find(f=>f.key==='record.rate').aggregation.family='ratio',x=>x.rows[0].cells['record.rate'].rate.elapsedMs='01']){const bad=structuredClone(d);mutate(bad);assert.equal(validateSnapshot(bad).ok,false,'forged rate value/support must fail');controls++;}
 }
 const unknown=structuredClone(v.data.timeDataset);delete unknown.completeness;unknown.quality.extent='partial';const missing=unknown.rows[0];missing.state='unknown';missing.values.count=null;missing.values['record.rate']=null;missing.cells['record.rate']={state:'unknown'};
 assert.equal(validateSnapshot(unknown).ok,true,'explicit unreported interval must remain unknown');
 for(const mutate of [x=>x.rows[0].values['record.rate']='0',x=>x.rows[0].cells['record.rate'].rate={count:'0',elapsedMs:'1000'}]){const bad=structuredClone(unknown);mutate(bad);assert.equal(validateSnapshot(bad).ok,false,'unknown cannot manufacture zero/support');controls++;}
 adoptExploration(v.data,q);
 Templates.templateView(v.summary);assert.deepEqual(Templates.authoredRequest(v.summary,v.destination),{source:v.summary,query:v.destination});assert.deepEqual(Templates.authoredResponse(v.expansion,v.summary,v.destination).queries,[v.destination]);
 Timeline.validate(v.timeline.composition.children.find(c=>c.timeline).timeline,q);
 for(const mutate of [x=>x.groupBy=[{kind:'column',key:'service'}],x=>x.groupTimeSeries=true]){const bad=structuredClone(q);mutate(bad);assert.throws(()=>Timeline.validate(v.timeline.composition.children.find(c=>c.timeline).timeline,bad));controls++;}
 assert.equal(JSON.stringify(v),before);
}
console.log(`${vectors.length} whole Go/JS rate carriers; ${controls} exact-value/support/unsupported-query refusal controls`);
