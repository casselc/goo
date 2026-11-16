import {test} from 'node:test';
import assert from 'node:assert/strict';
import {dataset,model} from '../ui/raw-gauge-points.mjs';
import {validateSnapshot,validateDisplay,eligibleKinds,selectionIntent,scopeIntent,capabilities,measureValue} from '../ui/view-schema.mjs';
import {bank} from './support/raw-gauge-points455-fixture.mjs';
test('raw gauges use the shared dataset/kind grammar, exact scope and occurrence membership',()=>{
 const input=bank([{rawValue:'18446744073709551615'},{rawValue:'-0'}]),before=structuredClone(input),snapshot=dataset(input),m=model(snapshot);
 assert.equal(validateSnapshot(snapshot).ok,true);assert.deepEqual(eligibleKinds(snapshot),['table','points']);assert.equal(snapshot.version,'goo.dataset/1');assert.equal(snapshot.shape,'observations');assert.equal(snapshot.scope.signal,'metrics');assert.equal(snapshot.sourceBasis,undefined);assert.equal(snapshot.quality.extent,'unknown');assert.equal(m.readAt,input.readAt);assert.equal(m.atLimit,false);assert.equal(m.points[0].rawValues[0],'18446744073709551615');assert.equal(m.points[1].rawValues[0],'-0');assert.equal(capabilities(snapshot).selectRows,true);assert.equal(selectionIntent(snapshot,'1').key,'1');assert.throws(()=>scopeIntent(snapshot,'value','1'));assert.equal(validateDisplay(snapshot,{kind:'points',requiresComplete:true}).ok,false);for(const kind of ['line','bar','histogram','time-bar'])assert.equal(validateDisplay(snapshot,{kind}).ok,false);m.points[0].rawValues[0]='changed';assert.deepEqual(input,before);
});
test('only known recorded finite observations plot; all withheld occurrences remain accounted',()=>{
 const m=model(dataset(bank([{rawValue:'0'},{rawValue:'-3'},{flags:1},{flags:2},{flags:3},{flags:null},{flags:'invalid'},{rawValue:'NaN'},{rawValue:{invalid:true}},{rawValue:null}])));
 assert.equal(m.counts.loaded,10);assert.equal(m.counts.eligible,2);assert.equal(m.counts.plotted,2);assert.equal(m.counts.unplotted,8);assert.equal(m.counts.noRecordedValue,2);assert.equal(m.counts.unknownRecording,3);assert.equal(m.counts.invalidValue,3);assert.equal(m.excluded.length,8);const states=dataset(bank([{flags:0},{flags:2},{flags:1},{rawValue:'NaN'}]));assert.deepEqual(states.rows.map(row=>row.state),['recorded','unknown','not-recorded','invalid']);assert.deepEqual(states.rows.map(row=>measureValue(states,row,'value').state),['recorded','unknown','no_data','unknown']);assert.equal(new Set([...m.points.flatMap(point=>point.occurrenceKeys),...m.excluded.map(row=>row.key)]).size,10);
});
test('finite geometry survives extreme, constant, negative and subnormal values with exact ns subtraction',()=>{
 for(const values of [['-1.7976931348623157e308','0','1.7976931348623157e308'],['5e-324','1e-323'],['-3','-3'],['0']]){
  const source=bank(values.map(rawValue=>({rawValue}))),m=model(dataset(source));for(const point of m.points){assert.ok(Number.isFinite(point.x)&&point.x>=0&&point.x<=1);assert.ok(Number.isFinite(point.y)&&point.y>=0&&point.y<=1)}assert.equal(m.points[0].x,1/60000000000);assert.equal(m.fromNS,'1700000000000000000');
 }
 const exact=bank([{rawValue:'1.7976931348623157e308'},{rawValue:'1.7976931348623155e308'}]);assert.deepEqual(model(dataset(exact)).points.map(p=>p.y),[1,0]);
});
test('coincident coordinates retain every member including distinct raw spellings and rounding collisions',()=>{
 const endNS='1700000000000000001',source=bank([{rawValue:'9007199254740992',endNS},{rawValue:'9007199254740993',endNS},{rawValue:'9007199254740992',endNS}]),m=model(dataset(source));assert.equal(m.points.length,1);assert.equal(m.points[0].multiplicity,3);assert.deepEqual(m.points[0].occurrenceKeys,['1','2','3']);assert.deepEqual(m.points[0].rawValues,['9007199254740992','9007199254740993','9007199254740992']);assert.equal(m.counts.plotted,3);
});
test('render budget is explicit, full-bank domains remain and at-limit coverage is not complete',()=>{
 const source=bank(Array.from({length:1001},(_,i)=>({rawValue:String(i)}))),snapshot=dataset(source),m=model(snapshot);assert.equal(snapshot.rows.length,1001);assert.equal(m.points.length,1000);assert.equal(m.counts.eligible,1001);assert.equal(m.counts.plotted,1000);assert.equal(m.counts.notShownEligible,1);assert.equal(m.excluded.at(-1).reason,'visual-budget');assert.equal(m.domain.max,1000);assert.equal(m.points.at(-1).y,.999);
 const capped=dataset(bank(Array.from({length:10000},()=>({rawValue:'1',endNS:'1700000000000000001'}))));assert.equal(capped.quality.extent,'unknown');assert.equal(model(capped).atLimit,true);assert.equal(validateDisplay(capped,{kind:'points',requiresComplete:true}).ok,false);
});
test('forged metadata, values, recording certainty and raw selector cannot bypass admission',()=>{
 const snapshot=dataset(bank());
 for(const change of [s=>s.scope.selector.instrument.kind='counter',s=>s.quality.extent='complete',s=>s.quality.extent='partial',s=>s.quality.value='recorded',s=>s.rawGauge.limit=1,s=>s.rows[0].values.value='4',s=>s.rows[0].state='unknown',s=>s.rows[0].observation.flags=1,s=>s.rows[0].observation.rawRow.flags=1,s=>s.rows[0].key='other',s=>s.rows[0].observation.endNS='1700000060000000000',s=>s.fields[2].unit='s',s=>s.sourceBasis='invented',s=>s.rawGauge.extra=true]){const next=structuredClone(snapshot);change(next);assert.equal(validateSnapshot(next).ok,false);assert.equal(model(next).state,'unavailable')}
 assert.equal(model(dataset(bank([]))).state,'no_data');
});
test('projection and geometry are detached source-free pure operations',()=>{const previous=globalThis.fetch;let calls=0;globalThis.fetch=()=>{calls++;throw Error('Unexpected IO')};try{const source=bank(),snapshot=dataset(source);model(snapshot);snapshot.rows[0].observation.rawRow.resource.changed='local';assert.equal(source.rows[0].rawRow.resource.changed,undefined);assert.equal(calls,0)}finally{globalThis.fetch=previous}});
