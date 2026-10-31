const {test}=require('node:test'),assert=require('node:assert/strict');
const {nativeHistogramSnapshot}=require('../ui/measurements.mjs');
const {createSourceAdapter,hasGuarantee}=require('../ui/snapshots.mjs');
const {createActionContext}=require('../ui/scope-actions.mjs');
const {fixture}=require('./fixtures/native-histogram-time-app.cjs');
test('time bucket projection retains observation totals, selected coverage and immutable parent data',()=>{
 const r=fixture(),before=JSON.stringify(r),s=r.histograms.series[0],view=nativeHistogramSnapshot(r,s,0,{bucketed:true});
 assert(view);assert.equal(view.histogram.total,'7');assert.notEqual(view.histogram.total,r.count);
 assert.equal(view.scope.selector.kind,'measurement-time-bucket');assert.equal(view.scope.selector.widthNS,'4000000000');
 assert.equal(view.scope.selector.fromNS,'1000000000');assert.equal(view.scope.selector.toNS,'5000000000');
 assert.equal(view.sourceBasis,r.histogramTime.basis);assert.equal(hasGuarantee(view,'bins'),false);assert.equal(JSON.stringify(r),before);
});
test('live bucket admission checks requested question and cannot stage record-population actions',()=>{
 const r=fixture(),s=r.histograms.series[0],sourceAdapter=createSourceAdapter({id:'native.time.app'}),options={bucketed:true,sourceAdapter,sourceQuestion:structuredClone(r.question)};
 const view=nativeHistogramSnapshot(r,s,0,options);assert.equal(hasGuarantee(view,'bins'),true);
 const context=createActionContext({epoch:()=>1,active:()=>true,canStage:()=>true});
 assert.throws(()=>context.bind({role:'current',slot:'histogram',epoch:1,snapshot:view,question:view.scope.selector,coherent:true}),e=>e.code==='scope-family');
 assert.throws(()=>nativeHistogramSnapshot(r,s,0,{...options,sourceQuestion:{...r.question,toMs:13000}}),/requested measurement/);
});
test('known empty, unavailable and clipped final buckets remain distinct without an earlier-value fallback',()=>{
 const r=fixture(),s=r.histograms.series[0];
 const empty=nativeHistogramSnapshot(r,s,1,{bucketed:true});assert.equal(empty.histogram.total,'0');assert(empty.rows.every(row=>row.state==='no_data'));
 assert.equal(nativeHistogramSnapshot(r,s,2,{bucketed:true}),null);assert.equal(nativeHistogramSnapshot(r,s,4,{bucketed:true}),null);
 const final=nativeHistogramSnapshot(r,s,3,{bucketed:true});assert.equal(final.histogram.total,'2');assert.equal(final.scope.selector.toNS,'14000000000');
 assert.equal(final.scope.selector.widthNS,'4000000000');assert.equal(final.scope.selector.bucketIndex,3);
});
test('derived stream lookup uses full identity while preserving original source membership',()=>{
 const r=fixture(),s=r.histograms.series[0],other=structuredClone(r.histogramTime.series[0]);other.stream.attributes={other:'stream'};const rawOther=structuredClone(s);rawOther.stream=structuredClone(other.stream);r.histograms.series.push(rawOther);
 r.histogramTime.series.unshift(other);
 const view=nativeHistogramSnapshot(r,s,0,{bucketed:true});assert.equal(view.histogram.total,'7');assert.deepEqual(view.scope.selector.stream,s.stream);assert.match(view.scope.id,/:native-time:1:0$/);
 assert.equal(nativeHistogramSnapshot(r,structuredClone(s),0,{bucketed:true}),null);
 r.histogramTime.series.push(structuredClone(r.histogramTime.series[1]));assert.equal(nativeHistogramSnapshot(r,s,0,{bucketed:true}),null);
});
