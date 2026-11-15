import {test} from 'node:test';
import assert from 'node:assert/strict';
import {profile,supports} from '../ui/raw-observation-profiles.mjs';
import {checkedRequest} from '../ui/raw-observations.mjs';

test('raw-read routing is closed and keeps source families explicit',()=>{
 for(const [kind,label,path,version] of [
  ['gauge','gauge','metric-observations','gauge'],
  ['counter','sum','sum-observations','sum'],
  ['stock','sum','sum-observations','sum'],
  ['histogram','histogram','histogram-observations','histogram'],
 ]){
  assert.equal(supports(kind),true);
  const selected=profile(kind);
  assert.equal(selected.label,label);
  assert.equal(selected.path,'/api/'+path);
  assert.equal(selected.version,'goo/raw-'+version+'-observations/v1');
 }
 for(const kind of [undefined,null,false,0,'','logs','metrics','state','exponential-histogram','future','toString','__proto__',new String('gauge')]){
  assert.equal(supports(kind),false);
  assert.throws(()=>profile(kind),/do not support/);
 }
});

test('routing metadata and exact instrument field lists cannot be modified by a caller',()=>{
 for(const kind of ['gauge','counter','stock','histogram']){
  const selected=profile(kind),before=structuredClone(selected);
  assert.equal(Object.isFrozen(selected),true);
  assert.equal(Object.isFrozen(selected.instrumentFields),true);
  assert.throws(()=>selected.path='/api/measurements',TypeError);
  assert.throws(()=>selected.instrumentFields.push('cursor'),TypeError);
  assert.throws(()=>selected.instrumentFields[0]='fromMs',TypeError);
  const copy=structuredClone(selected);copy.instrumentFields.push('readAt');copy.path='/api/captures';
  assert.deepEqual(profile(kind),before);
 }
});

test('discovery field projection retains the exact family tuple and excludes runtime metadata',()=>{
 const common=['kind','scope','version','schemaURL','metric','unit'];
 for(const kind of ['gauge','counter','stock','histogram']){
  const fields=[...common,...(kind==='gauge'?[]:kind==='histogram'?['temporality']:['temporality','monotonic'])];
  assert.deepEqual(profile(kind).instrumentFields,fields);
  const catalog={kind,scope:'ordinary',version:'',schemaURL:'',metric:'work',unit:'',temporality:2,monotonic:kind!=='stock',records:42,capabilityPlan:{eligible:false},cursor:'not-an-instrument',readAt:'not-source-authority'};
  const instrument=Object.fromEntries(profile(kind).instrumentFields.map(key=>[key,catalog[key]]));
  const request={question:{environment:'prod',service:'worker',fromMs:0,toMs:60000},instrument};
  assert.deepEqual(checkedRequest(request),request);
  assert.equal(Object.hasOwn(instrument,'records'),false);
  assert.equal(Object.hasOwn(instrument,'cursor'),false);
  assert.equal(Object.hasOwn(instrument,'readAt'),false);
  if(kind==='histogram')assert.equal(Object.hasOwn(instrument,'monotonic'),false);
  if(kind==='gauge')assert.equal(Object.hasOwn(instrument,'temporality'),false);
 }
});
