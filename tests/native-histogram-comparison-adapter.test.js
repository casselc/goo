const {test}=require('node:test'),assert=require('node:assert/strict');
const {comparisonForWindow}=require('../ui/native-histogram-comparison.mjs');
const {fixture}=require('./fixtures/native-histogram-comparison-app.cjs');

test('comparable native windows use held independent operands without reads or input mutation',()=>{
 const r=fixture(),before=JSON.stringify(r),old=global.fetch;global.fetch=()=>{throw Error('Unexpected fetch')};
 try{const pair=comparisonForWindow(r,r.nativeWindows[0]);assert(pair);assert.equal(pair.baseline.histogram.total,'4');assert.equal(pair.selection.histogram.total,'8');assert.equal(pair.baseline.scope.selector.kind,'measurement-window');assert.notEqual(pair.baseline.capturedAt,pair.selection.capturedAt);assert.equal(pair.baseline.origin,'browser-held');assert.equal(pair.baseline.quality.extent,'unknown');assert.equal(pair.selection.quality.extent,'unknown');assert.equal(pair.baseline.completeness,undefined);assert.equal(JSON.stringify(r),before)}finally{global.fetch=old}
});
test('a locally complete member never bypasses the overall withheld decision',()=>{
 const r=fixture(),window=r.nativeWindows[0];for(const state of ['unavailable','incompatible','unknown','unsupported']){r.state=state;assert.equal(comparisonForWindow(r,window),null)}r.state='comparable';window.state='unavailable';window.reasons=['another-member-withholds-comparison'];assert.equal(comparisonForWindow(r,window),null);
});
test('foreign rows and unknown or empty windows do not gain aligned arithmetic',()=>{
 const r=fixture(),window=r.nativeWindows[0];assert.equal(comparisonForWindow(r,structuredClone(window)),null);
 window.selection.state='unavailable';delete window.selection.distribution;assert.equal(comparisonForWindow(r,window),null);
 const empty=fixture(),w=empty.nativeWindows[0];w.selection.state='no_data';w.selection.reasons=['empty-distribution'];w.selection.distribution.count='0';w.selection.distribution.bucketCounts.fill('0');assert.equal(comparisonForWindow(empty,w),null);
});
test('missing transport identity and incompatible bucket layouts are refused',()=>{
 const r=fixture(),window=r.nativeWindows[0];delete r.baseline.nativeHistogramSource;assert.equal(comparisonForWindow(r,window),null);
 const different=fixture(),row=different.nativeWindows[0];row.selection.distribution.bounds=[.25,.5];assert.equal(comparisonForWindow(different,row),null);
});
