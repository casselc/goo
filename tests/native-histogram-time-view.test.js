const {test}=require('node:test'),assert=require('node:assert/strict');
const {result,fixture}=require('./fixtures/native-histogram-time.cjs');
const {buildNativeHistogramTimeView,buildNativeHistogramView}=require('../ui/native-histogram-view.mjs');
const {validateMeasurementIntervalSelector,validateMeasurementTimeBucketSelector}=require('../ui/measurement-scope.mjs');
const {validateSnapshot,validateDisplay,histogramIntent}=require('../ui/view-model.mjs');
const {createSourceAdapter,importSnapshot,hasGuarantee,selectRows}=require('../ui/snapshots.mjs');
const source=raw=>createSourceAdapter({id:'time-view'}).adoptSnapshot(raw,{question:raw.scope.selector});

test('fixed-time projection preserves exact scope, derived basis and observation totals without re-evaluating exports',()=>{
 const r=result(),before=JSON.stringify(r),raw=buildNativeHistogramTimeView(r,0,0),snapshot=source(raw);
 assert.equal(snapshot.histogram.total,'7');assert.notEqual(snapshot.histogram.total,r.count);assert.equal(snapshot.sourceBasis,r.histogramTime.basis);
 assert.equal(snapshot.scope.id,r.basis+':native-time:0:0');assert.equal(snapshot.scope.selector.kind,'measurement-time-bucket');assert.equal(snapshot.scope.selector.widthNS,'4000000000');assert.equal(snapshot.scope.selector.bucketIndex,0);
 assert.equal(snapshot.scope.selector.fromNS,'1000000000');assert.equal(snapshot.scope.selector.toNS,'5000000000');
 assert.equal(snapshot.rows[0].values.from,'-inf');assert.equal(snapshot.rows[0].values.to,'3602879701896397/36028797018963968');assert.equal(snapshot.rows.at(-1).values.to,'+inf');
 assert.equal(validateSnapshot(snapshot).ok,true);assert.equal(validateDisplay(snapshot,{kind:'histogram',requiresComplete:true}).ok,true);assert.equal(hasGuarantee(snapshot,'bins'),true);assert.equal(JSON.stringify(r),before);
});
test('a whole-export time bucket can exceed the per-export max-gap width',()=>{
 const r=result(),raw=buildNativeHistogramTimeView(r,0,0);assert(raw);assert(BigInt(raw.scope.selector.toNS)-BigInt(raw.scope.selector.fromNS)>BigInt(r.question.definition.maxGapMs)*1000000n);
 assert.equal(validateMeasurementIntervalSelector(raw.scope.selector),raw.scope.selector);assert.equal(buildNativeHistogramView(r,0,0).scope.selector.kind,'measurement-interval');
});
test('the final selected time bucket is anchored and clipped to the exact requested end',()=>{
 const raw=buildNativeHistogramTimeView(result(),0,2);assert(raw);assert.equal(raw.histogram.total,'2');assert.equal(raw.scope.selector.fromNS,'9000000000');assert.equal(raw.scope.selector.toNS,'11000000000');assert.equal(raw.scope.selector.widthNS,'4000000000');assert.equal(raw.scope.selector.bucketIndex,2);
 for(const change of [s=>s.fromNS='8000000000',s=>s.toNS='13000000000',s=>s.bucketIndex=1,s=>s.widthNS='2000000000',s=>s.bucketIndex=2.5,s=>s.bucketIndex=200,s=>delete s.question.definition.histogramBucketMs,s=>s.extra=true]){const selector=structuredClone(raw.scope.selector);change(selector);assert.throws(()=>validateMeasurementTimeBucketSelector(selector))}
});
test('known-zero time buckets retain the declared empty-distribution diagnostic and exact zero counts',()=>{
 const r=result(),bucket=r.histogramTime.series[0].buckets[1];r.histograms.series[0].intervals[2]={fromNS:'5000000000',toNS:'7000000000',state:'ok',distribution:structuredClone(r.histograms.series[0].intervals[3].distribution)};r.histograms.series[0].state='ok';bucket.state='no_data';bucket.reasons=['empty-distribution'];bucket.distribution={unit:'s',bounds:[.1,.5],bucketCounts:['0','0','0'],count:'0',p95:{state:'no_data'}};
 const raw=buildNativeHistogramTimeView(r,0,1);assert(raw);assert.equal(raw.histogram.total,'0');assert(raw.rows.every(row=>row.state==='no_data'));assert.equal(hasGuarantee(source(raw),'bins'),true);
 for(const reasons of [[],['empty-distribution','empty-distribution'],['incomplete-window'],null]){bucket.reasons=reasons;assert.equal(buildNativeHistogramTimeView(r,0,1),null)}
});
test('unavailable and diagnostic buckets never borrow a neighboring known distribution',()=>{
 const r=result();assert.equal(buildNativeHistogramTimeView(r,0,1),null);assert.equal(buildNativeHistogramTimeView(r,0,3),null);assert.equal(buildNativeHistogramTimeView(r,1,0),null);
 for(const change of [b=>b.state='unavailable',b=>b.reasons=['boundary-crossing'],b=>b.reasons=['bucket-layout-change'],b=>b.state='no_data',b=>delete b.distribution,b=>delete b.reasons]){const v=result();change(v.histogramTime.series[0].buckets[0]);assert.equal(buildNativeHistogramTimeView(v,0,0),null)}
});
test('projection checks retained parent, derived lineage, policy, membership and exact bucket geometry',()=>{
 const mutations=[r=>r.histogramTime.apiVersion='other',r=>r.histogramTime.basis='bad',r=>r.histogramTime.sourceBasis='d'.repeat(64),r=>r.histogramTime.fromNS='0',r=>r.histogramTime.toNS='13000000000',r=>r.histogramTime.widthNS='4000000001',r=>r.histogramTime.series[0].buckets.pop(),r=>r.histogramTime.series[0].buckets[0].fromNS='1000000001',r=>r.histogramTime.series[0].buckets[0].toNS='4999999999',r=>r.histogramTime.series[0].temporality='cumulative',r=>r.histogramTime.series[0].stream.attributes.team='other',r=>r.histograms.series.push(structuredClone(r.histograms.series[0])),r=>r.histogramTime.series.push(structuredClone(r.histogramTime.series[0])),r=>r.histogramTime.series[0].buckets[0].distribution.count='8',r=>r.histogramTime.series[0].buckets[0].distribution.unit='ms',r=>delete r.question.definition.histogramBucketMs,r=>r.readAt='2026-02-31T00:00:00Z'];
 for(const change of mutations){const r=result();change(r);assert.equal(buildNativeHistogramTimeView(r,0,0),null)}
});
test('time series match full original stream identity independently of their retained order',()=>{
 const r=result(),otherRaw=structuredClone(r.histograms.series[0]),otherTime=structuredClone(r.histogramTime.series[0]);otherRaw.stream.resource['service.instance.id']='other';otherTime.stream.resource['service.instance.id']='other';r.histograms.series.push(otherRaw);r.histogramTime.series.unshift(otherTime);
 const raw=buildNativeHistogramTimeView(r,1,0);assert(raw);assert.equal(raw.scope.selector.stream.resource['service.instance.id'],'one');assert.equal(raw.scope.id,r.basis+':native-time:1:0');
});
test('fixed-time imports retain values but no private guarantee or record action intent',()=>{
 const admitted=source(fixture()),held=importSnapshot(admitted);assert.equal(hasGuarantee(held,'bins'),false);assert.equal(validateDisplay(held,{kind:'histogram'}).ok,true);assert.equal(validateDisplay(held,{kind:'histogram',requiresComplete:true}).ok,false);assert.throws(()=>histogramIntent(held,'bin:0'));assert.throws(()=>selectRows(held,['bin:0']));
 const wrong=structuredClone(admitted.scope.selector);wrong.widthNS='8000000000';assert.throws(()=>createSourceAdapter({id:'wrong-time'}).adoptSnapshot(admitted,{question:wrong}));
});
test('short single-bucket scopes retain width policy even when their displayed bounds match',()=>{
 const a=fixture().scope.selector;a.question.toMs=3000;a.toNS='3000000000';a.widthNS='4000000000';a.bucketIndex=0;assert.equal(validateMeasurementTimeBucketSelector(a),a);
 const b=structuredClone(a);b.question.definition.histogramBucketMs=8000;b.widthNS='8000000000';assert.equal(validateMeasurementTimeBucketSelector(b),b);assert.notDeepEqual(a,b);
});
