import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {prepare} from '../ui/raw-observation-dataset.mjs';
import {dataset as histogramDataset} from '../ui/raw-histogram-dataset.mjs';
import {dataset as gaugeDataset} from '../ui/raw-gauge-points.mjs';
import {bank as gaugeBank} from './support/raw-gauge-points455-fixture.mjs';
const histogramBank=()=>JSON.parse(JSON.parse(readFileSync(new URL('../testdata/http/raw-histogram447.json',import.meta.url))).cases.find(c=>c.name==='observed').response);

test('prepare conserves exact projections while isolating caller mutations and frozen readback',()=>{
 const source=histogramBank();source.rows.push({...structuredClone(source.rows[0]),occurrenceKey:'second'});const expected=histogramDataset(source,'second'),bank=prepare(source);
 assert.equal(Object.isFrozen(bank),true);assert.equal(Object.isFrozen(bank.data.rows[1].rawRow.bucket_counts),true);assert.deepEqual(bank.histogram('second'),expected);
 source.rows[1].rawRow.bucket_counts[0]='forged';source.rows[1].stream.unit='forged';source.rows.length=0;
 assert.deepEqual(bank.histogram('second'),expected);assert.throws(()=>bank.data.rows.push({}));assert.throws(()=>bank.data.rows[1].rawRow.bucket_counts[0]='forged');
 const projected=bank.histogram('second');projected.rows[0].observation.rawRow.bucket_counts[0]='forged';projected.rawHistogram.request.question.service='forged';assert.deepEqual(bank.histogram('second'),expected);
 assert.throws(()=>bank.histogram('missing'));assert.throws(()=>bank.histogram(bank.data.rows[0]));assert.throws(()=>bank.gauge());
});
test('read-boundary preparation matches the original expected request and rejects any malformed bank member',()=>{
 const source=histogramBank(),expected={question:structuredClone(source.question),instrument:structuredClone(source.instrument)};assert.doesNotThrow(()=>prepare(source,expected));expected.question.service='other';assert.throws(()=>prepare(source,expected));
 const malformed=histogramBank();malformed.rows.push({...structuredClone(malformed.rows[0]),occurrenceKey:'bad',count:'999'});assert.throws(()=>prepare(malformed));
 const duplicate=histogramBank();duplicate.rows.push(structuredClone(duplicate.rows[0]));assert.throws(()=>prepare(duplicate));
 const forged=histogramBank();forged.rows[0].count='999';assert.throws(()=>histogramDataset(forged,'1'));assert.throws(()=>histogramDataset(histogramBank(),{...histogramBank().rows[0],count:'999'}));
});
test('prepared gauge copies preserve source identity, raw values, certainty and cached projection isolation',()=>{
 const source=gaugeBank([{rawValue:'7'},{flags:2},{rawValue:'NaN'},{flags:1}]),expected=gaugeDataset(source),bank=prepare(source);assert.deepEqual(bank.gauge(),expected);assert.deepEqual(bank.gauge().rows.map(row=>row.state),['recorded','unknown','invalid','not-recorded']);
 source.rows[0].rawValue='changed';const first=bank.gauge();first.rows[0].values.value={numerator:'999',denominator:'1'};first.rows[0].observation.rawRow.value='changed';assert.deepEqual(bank.gauge(),expected);assert.throws(()=>bank.histogram('1'));
});
test('preparation and repeated projections do not fetch or weaken scope/completeness semantics',()=>{
 const previous=globalThis.fetch;globalThis.fetch=()=>{throw Error('Unexpected source read')};try{const bank=prepare(histogramBank());for(let i=0;i<3;i++){const snapshot=bank.histogram('1');assert.deepEqual(snapshot.quality,{extent:'unknown',basis:'captured-query',value:'unknown',sampling:'unknown',freshness:'fixed'});assert.equal(snapshot.sourceBasis,undefined);assert.equal(snapshot.completeness,undefined);assert.equal(snapshot.capturedAt,bank.data.readAt);assert.deepEqual(snapshot.scope.selector.question,bank.data.question)}}finally{globalThis.fetch=previous}
});
