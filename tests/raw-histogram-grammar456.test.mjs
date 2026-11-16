import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {dataset} from '../ui/raw-histogram-dataset.mjs';
import {validateSnapshot,validateDisplay,eligibleKinds} from '../ui/view-schema.mjs';
import {importSnapshot,parseSnapshotJSON,getSnapshotInfo,createSourceAdapter} from '../ui/snapshots.mjs';
import {facts} from '../ui/raw-histogram-values.mjs';
const original=JSON.parse(JSON.parse(readFileSync(new URL('../testdata/http/raw-histogram447.json',import.meta.url))).cases.find(value=>value.name==='observed').response);
function bank(change=()=>{}){const value=structuredClone(original),row=value.rows[0];change(row.rawRow);const flags=row.rawRow.flags;row.flags=flags??null;row.flagsState=flags===undefined||flags===null?'absent':'valid';const prefix=row.flagsState==='absent'?[{field:'flags',code:'absent'}]:(BigInt(flags)&~1n)!==0n?[{field:'flags',code:'unknown-bits'}]:[];const projected=facts(row.rawRow,{flags:row.flags,flagsState:row.flagsState});Object.assign(row,projected.result);row.diagnostics=[...prefix,...projected.diagnostics];return value}

test('one raw histogram occurrence uses existing histogram/table grammar without claiming interval coverage',()=>{
 const source=bank(),snapshot=dataset(source,source.rows[0]);assert.equal(validateSnapshot(snapshot).ok,true);assert.deepEqual(eligibleKinds(snapshot),['table','histogram']);assert.equal(validateDisplay(snapshot,{kind:'histogram'}).ok,true);assert.equal(validateDisplay(snapshot,{kind:'points'}).ok,false);assert.equal(validateDisplay(snapshot,{kind:'histogram',requiresComplete:true}).ok,false);assert.equal(snapshot.quality.extent,'unknown');assert.equal(snapshot.sourceBasis,undefined);assert.equal(snapshot.rows[0].state,'recorded');assert.equal(snapshot.rows[0].values.count,'3');assert.deepEqual(snapshot.rows[0].observation,source.rows[0]);source.rows[0].rawRow.count='other';assert.equal(snapshot.rows[0].observation.rawRow.count,'3');assert.throws(()=>dataset(bank(),'missing'));assert.throws(()=>dataset(bank(),{...bank().rows[0],count:'4'}));
});
test('raw histogram grammar preserves all recording and bucket certainty states and arbitrary unit/negative edges',()=>{
 for(const [change,state] of [[raw=>raw.flags=2,'unknown'],[raw=>delete raw.flags,'unknown'],[raw=>raw.flags=1,'not-recorded'],[raw=>raw.count='4','invalid'],[raw=>raw.bounds=['-1'],'recorded']]){const source=bank(change),snapshot=dataset(source,source.rows[0]);assert.equal(snapshot.rows[0].state,state);assert.equal(validateSnapshot(snapshot).ok,true);assert.equal(validateDisplay(snapshot,{kind:'table'}).ok,true)}
 const source=bank();source.instrument.unit='widgets';source.rows[0].stream.unit=source.rows[0].rawRow.unit='widgets';assert.equal(validateDisplay(dataset(source,'1'),{kind:'histogram'}).ok,true);
});
test('held and SSR histogram snapshots retain original clock and certainty without source authority',()=>{
 const source=bank(),raw=dataset(source,'1'),held=importSnapshot(raw);assert.equal(held.origin,'browser-held');assert.equal(held.sourceBasis,undefined);assert.equal(held.capturedAt,source.readAt);assert.deepEqual(held.rawHistogram,raw.rawHistogram);assert.deepEqual(getSnapshotInfo(held).guarantees,[]);assert.equal(Object.isFrozen(held.rows[0].observation.rawRow),true);assert.equal(parseSnapshotJSON(JSON.stringify(raw)).rows[0].key,'1');assert.throws(()=>createSourceAdapter({id:'test.raw'}).adoptSnapshot(raw,{question:raw.scope.selector}));
 for(const change of [s=>s.quality.extent='partial',s=>s.quality.value='recorded',s=>s.sourceBasis='pretend',s=>s.rows.push(structuredClone(s.rows[0])),s=>s.rows[0].state='unknown',s=>s.rows[0].values.count='4',s=>s.rawGauge={},s=>s.rawHistogram.request.instrument.kind='gauge']){const forged=structuredClone(raw);change(forged);assert.equal(validateSnapshot(forged).ok,false)}
});
test('raw histogram profile preserves 10001 exact buckets, including counts above safe integers',()=>{
 const source=bank(raw=>{raw.bounds=Array.from({length:10000},(_,i)=>String(i-5000));raw.count='18446744073709551615';raw.bucket_counts=['18446744073709551615',...Array(10000).fill('0')]});const raw=dataset(source,'1');assert.equal(raw.rows[0].values.count,'18446744073709551615');assert.equal(parseSnapshotJSON(JSON.stringify(raw)).rows[0].observation.bucketCounts.length,10001);assert.equal(importSnapshot(raw).quality.extent,'unknown');
});
