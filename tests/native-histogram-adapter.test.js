const {test}=require('node:test'),assert=require('node:assert/strict');
const {nativeHistogramSnapshot}=require('../ui/measurements.mjs');
const {createSourceAdapter,hasGuarantee}=require('../ui/snapshots.mjs');
const {fixture}=require('./fixtures/native-histogram-app.cjs');

test('native app projection defaults to held data and never mutates retained evidence',()=>{
 const r=fixture(),before=JSON.stringify(r),s=r.histograms.series[0],view=nativeHistogramSnapshot(r,s,0);
 assert.equal(view.histogram.kind,'otel-explicit');assert.equal(view.histogram.total,'4');assert.equal(view.histogram.grain,'observations');
 assert.equal(view.origin,'browser-held');assert.equal(hasGuarantee(view,'bins'),false);assert.equal(JSON.stringify(r),before);
});
test('a live request owner admits only its matching measurement interval',()=>{
 const r=fixture(),s=r.histograms.series[0],sourceAdapter=createSourceAdapter({id:'native.app.test'}),options={sourceAdapter,sourceQuestion:structuredClone(r.question)};
 const view=nativeHistogramSnapshot(r,s,0,options);assert.equal(hasGuarantee(view,'bins'),true);assert.equal(view.scope.signal,'metrics');assert.deepEqual(view.scope.selector.stream,s.stream);
 assert.throws(()=>nativeHistogramSnapshot(r,s,0,{sourceAdapter}),/requested measurement/);
 const changed=structuredClone(r.question);changed.definition.environment='other';assert.throws(()=>nativeHistogramSnapshot(r,s,0,{sourceAdapter,sourceQuestion:changed}),/requested measurement/);
});
test('unknown and foreign intervals have no earlier known distribution fallback',()=>{
 const r=fixture(),s=r.histograms.series[0];assert.equal(nativeHistogramSnapshot(r,s,1),null);assert.equal(nativeHistogramSnapshot(r,s,99),null);
 assert.equal(nativeHistogramSnapshot(r,structuredClone(s),0),null);
 const zero=nativeHistogramSnapshot(r,s,2);assert.equal(zero.histogram.total,'0');assert(zero.rows.every(row=>row.values.count==='0'&&row.state==='no_data'));
});
test('the selected stream remains independent from other streams and later layout changes',()=>{
 const r=fixture(),s=r.histograms.series[0];r.histograms.series.push(structuredClone(s));r.histograms.series[1].stream.resource['service.instance.id']='two';r.histograms.series[1].intervals[0].distribution.count='6';r.histograms.series[1].intervals[0].distribution.bucketCounts=['2','2','2'];
 assert.equal(nativeHistogramSnapshot(r,s,0).histogram.total,'4');assert.equal(nativeHistogramSnapshot(r,r.histograms.series[1],0).histogram.total,'6');
 s.intervals.push({fromNS:'4000000000',toNS:'5000000000',state:'ok',distribution:{unit:'s',bounds:[.25],bucketCounts:['1','2'],count:'3',p95:{state:'unbounded-upper',lower:.25}}});
 assert.equal(nativeHistogramSnapshot(r,s,3).rows.length,2);assert.equal(nativeHistogramSnapshot(r,s,0).rows.length,3);
});
