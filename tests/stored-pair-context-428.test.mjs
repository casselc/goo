import {test} from 'node:test';
import assert from 'node:assert/strict';
import {storedPairContext} from '../ui/distributions.mjs';
import {contextModel} from '../ui/comparison-visual.mjs';
const query={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,groupBy:[{kind:'column',key:'service'}]};
test('428 stored pair context uses immutable child references and original per-side source clocks without inventing custody',()=>{
 const pair={baselineId:'a'.repeat(64),selectionId:'b'.repeat(64),request:{baseline:{question:query,count:'2',state:'observed',basis:'baseline-read',capturedAt:'2026-11-14T01:02:03Z'},selection:{question:{...query,fromMs:61000,toMs:121000},count:'0',state:'no_data',basis:'selected-read',capturedAt:'2026-11-14T01:02:04Z'}},result:{projection:{origin:'browser-held'}}},before=structuredClone(pair),context=storedPairContext(pair);
 assert.equal(contextModel(context),true);assert.deepEqual(context.sides.map(s=>s.label),['Baseline','Selected']);
 for(const [i,side]of context.sides.entries()){assert.equal(side.data.mode,'frozen');assert.equal(side.data.details.find(f=>f.label==='Capture ID').value,i?pair.selectionId:pair.baselineId);assert.equal(side.data.details.find(f=>f.label==='Evidence storage').value,'Stored child capture');assert.equal(side.data.details.find(f=>f.label==='Source read').value,i?pair.request.selection.capturedAt:pair.request.baseline.capturedAt);assert.equal(side.data.details.some(f=>/custody/i.test(f.label)),false);}
 assert.equal(context.sides[1].data.state,'no_data');assert.deepEqual(pair,before);assert.throws(()=>storedPairContext({...pair,selectionId:'invalid'}));
});
