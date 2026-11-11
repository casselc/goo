import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Histogram from '../ui/histogram-declaration.mjs';
import * as Composition from '../ui/view-composition.mjs';
import * as Templates from '../ui/view-template-handoff.mjs';
import {createSourceAdapter,importSnapshot} from '../ui/snapshots.mjs';
import {validateSnapshot,validateDisplay} from '../ui/view-model.mjs';

const vectors=JSON.parse(fs.readFileSync(0,'utf8'));let controls=0;
globalThis.fetch=()=>{throw Error('Histogram configuration and passive display cannot read sources')};
for(const [i,v]of vectors.entries()){
 const before=JSON.stringify(v),q=v.source.queries[0],leaf=v.source.composition.children.find(c=>c.histogram).histogram;
 assert(validateSnapshot(v.dataset).ok);assert.deepEqual(v.dataset.histogram.binding,q.numeric);
 assert.deepEqual(v.dataset.histogram.support,v.support);assert.deepEqual(v.dataset.rows.map(r=>r.values),v.rows);
 assert.equal(v.dataset.quality.extent,v.expectedComplete?'complete':'unknown');
 Histogram.validate(leaf,q);Composition.validate(v.source.composition,q);Templates.templateView(v.source);
 assert.deepEqual(Templates.authoredRequest(v.source,v.destination),{source:v.source,query:v.destination});
 assert.deepEqual(Templates.authoredResponse(v.expansion,v.source,v.destination).queries,[v.destination]);
 assert.deepEqual(Composition.reuse(v.source.composition,q,v.destination),v.expansion.view.composition);
 const source=createSourceAdapter({id:'histogram407-'+i}).adoptSnapshot(v.dataset,{question:q});
 for(const kind of ['histogram','table']){
  const display={kind};
  assert(validateDisplay(source,display).ok);
  const complete=validateDisplay(source,{...display,requiresComplete:true});assert.equal(complete.ok,v.expectedComplete,complete.reasons.join('; '));
  assert(!validateDisplay(importSnapshot(source),{...display,requiresComplete:true}).ok);
 }
 for(const mutate of [x=>x.numeric.unit+=' ',x=>x.numeric.field.kind='resource',x=>x.numeric.field.key+='.other',x=>{x.numeric.type='decimal64';x.numeric.scale=7},x=>x.numeric.parse='coerce',x=>x.numericHistogram.binning='linear',x=>x.numericHistogram=null,x=>x.offset=1,x=>x.groupBy=[{kind:'column',key:'service'}]]){
  const bad=structuredClone(v.destination);mutate(bad);assert.throws(()=>Templates.authoredRequest(v.source,bad));controls++;
 }
 for(const mutate of [x=>x.display.value='numeric.mean',x=>x.display.category='to',x=>x.kind.name='bar',x=>x.display.normalization='density',x=>x.snapshot=v.dataset]){
  const bad=structuredClone(leaf);mutate(bad);assert.throws(()=>Histogram.validate(bad,q));controls++;
 }
 assert.equal(JSON.stringify(v),before);
}
console.log(`${vectors.length} complete histogram source/view/store/template carriers; ${controls} incompatible-binding and declaration refusal controls`);
