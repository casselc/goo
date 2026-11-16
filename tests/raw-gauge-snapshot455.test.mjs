import {test} from 'node:test';
import assert from 'node:assert/strict';
import {dataset,model} from '../ui/raw-gauge-points.mjs';
import {bank} from './support/raw-gauge-points455-fixture.mjs';
import {importSnapshot,asSnapshot,getSnapshotInfo,hasGuarantee,selectRows,createSourceAdapter,parseSnapshotJSON,RAW_OBSERVATION_LIMITS} from '../ui/snapshots.mjs';
import {validateSnapshot,capabilities,selectionIntent,validateDisplay} from '../ui/view-schema.mjs';
test('generic import downgrades raw gauges to immutable held data without inventing a source basis',()=>{
 const raw=dataset(bank([{rawValue:'7'},{flags:2}])),before=structuredClone(raw),held=importSnapshot(raw);assert.equal(held.origin,'browser-held');assert.notEqual(held.id,raw.id);assert.equal(held.sourceBasis,undefined);assert.equal(held.capturedAt,raw.capturedAt);assert.equal(held.quality.extent,'unknown');assert.equal(validateSnapshot(held).ok,true);assert.deepEqual(getSnapshotInfo(held).guarantees,[]);assert.equal(hasGuarantee(held,'groups'),false);assert.equal(asSnapshot(held),held);assert.equal(capabilities(held).selectRows,true);assert.equal(capabilities(held).filterDimensions,false);assert.equal(selectionIntent(held,'1').datasetId,held.id);assert.equal(validateDisplay(held,{kind:'points'}).ok,true);assert.equal(Object.isFrozen(held.rows[0].observation),true);assert.deepEqual(raw,before);assert.equal(model(held).counts.plotted,1);
});
test('held row projection preserves certainty and read metadata without a completeness or source attestation',()=>{
 const raw=dataset(bank([{rawValue:'7'},{flags:2}])),held=importSnapshot(raw),projected=selectRows(held,['2']);assert.equal(projected.origin,'browser-held');assert.equal(projected.sourceBasis,undefined);assert.equal(projected.quality.extent,'unknown');assert.equal(validateSnapshot(projected).ok,true);assert.deepEqual(projected.rawGauge,held.rawGauge);assert.equal(projected.rows[0].state,'unknown');assert.equal(model(projected).counts.plotted,0);assert.equal(model(projected).counts.loaded,1);assert.equal(getSnapshotInfo(projected).entry,'projection');assert.equal(validateDisplay(projected,{kind:'points',requiresComplete:true}).ok,false);const source=createSourceAdapter({id:'test.raw'});assert.throws(()=>source.adoptSnapshot(raw,{question:raw.scope.selector}));
});

test('raw profile admits the 10000-row bank within its bounded object/SSR envelope without changing ordinary limits',()=>{
 const raw=dataset(bank(Array.from({length:10000},()=>({rawValue:'1',endNS:'1700000000000000001'})))),held=importSnapshot(raw);assert.equal(held.rows.length,10000);assert.equal(held.sourceBasis,undefined);assert.equal(held.quality.extent,'unknown');assert.equal(parseSnapshotJSON(JSON.stringify(raw)).rows.length,10000);
 const forged=structuredClone(raw);forged.rawGauge.request.instrument.kind='counter';assert.throws(()=>importSnapshot(forged));const oversize=structuredClone(raw);oversize.extra='x'.repeat(RAW_OBSERVATION_LIMITS.maxBytes);assert.throws(()=>importSnapshot(oversize),error=>error.code==='budget-bytes');
});
