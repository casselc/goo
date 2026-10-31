const {test}=require('node:test'),assert=require('node:assert/strict');
const {fixture,inputs}=require('./fixtures/native-histogram-comparison.cjs');
const {buildNativeHistogramWindowView,buildNativeHistogramComparison,buildNativeHistogramComparisonFromEvidence}=require('../ui/native-histogram-view.mjs');
const {validateHistogramComparison}=require('../ui/view-histogram-comparison.mjs');
const {model,sideData,displayOptions}=require('../ui/aligned-histogram.mjs');
const {asHistogramComparison,parseHistogramComparisonJSON,hasGuarantee}=require('../ui/snapshots.mjs');
const {fixture:numericFixture}=require('./fixtures/histogram-comparison.cjs');

test('native window comparisons retain independent complete scopes and clocks without source proof',()=>{
 const input=inputs(),before=JSON.stringify(input),pair=buildNativeHistogramComparison(input.baseline,input.selection,input.window);
 assert(pair);assert.equal(validateHistogramComparison(pair).ok,true);
 for(const [key,total,basis] of [['baseline','4','b'],['selection','6','d']]){const side=pair[key];assert.equal(side.scope.selector.kind,'measurement-window');assert.equal(side.histogram.total,total);assert.equal(side.sourceBasis,basis.repeat(64));assert.equal(side.origin,'browser-held');assert.equal(side.quality.extent,'unknown');assert.equal(side.completeness,undefined);assert.equal(side.histogram.support,undefined)}
 assert.notEqual(pair.baseline.capturedAt,pair.selection.capturedAt);assert.notEqual(pair.baseline.scope.selector.fromNS,pair.selection.scope.selector.fromNS);assert.equal(JSON.stringify(input),before);
});
test('count and observation-share align exact bins with independent denominators and fixed percentage scale',()=>{
 const raw=fixture(),counts=model(raw),share=model(raw,{scale:'observation-share'});
 assert.equal(counts.ceiling,'3');assert.equal(counts.native,true);assert.equal(counts.all.length,3);assert.equal(counts.points.length,3);assert.equal(share.ceiling,'100');
 assert.deepEqual(share.denominators,['4','6']);assert.deepEqual(share.all[0].shares,['25','50/3']);assert.equal(share.all[0].shareDifference,'-25/3');assert.equal(share.all.at(-1).shareDifference,'25');assert.equal(counts.all.at(-1).difference,'2');
 assert.equal(sideData(share,0).interactive,false);assert.equal(sideData(share,1).interactive,false);assert.equal(sideData(counts,0).fromLabel,'−∞');assert.equal(sideData(counts,0).toLabel,'+∞');assert.match(counts.all[1].label,/\]$/);assert.match(counts.all[0].label,/≈ 0.1 s/);
});
test('native scales never borrow numeric parsing or record-population semantics',()=>{
 const raw=fixture();for(const scale of ['population-share','valid-share'])assert.equal(model(raw,{scale}),null);
 assert.equal(model(numericFixture(),{scale:'observation-share'}),null);assert.deepEqual(displayOptions({scale:'observation-share'}),{scale:'observation-share'});
});
test('known zero native operands remain displayable with unavailable shares but the app constructor preserves its existing empty-window gate',()=>{
 const pair=fixture();pair.state='no_data';pair.baseline.histogram.total='0';for(const row of pair.baseline.rows){row.values.count='0';row.state='no_data'}
 assert.equal(validateHistogramComparison(pair).ok,true);const share=model(pair,{scale:'observation-share'});assert(share);assert(share.all.every(row=>row.shares[0]===null&&row.shareDifference===null));assert.equal(share.all.at(-1).shares[1],'50');assert.equal(model(pair).all.at(-1).difference,'3');
 pair.selection.histogram.total='0';for(const row of pair.selection.rows){row.values.count='0';row.state='no_data'}assert.equal(model(pair).empty,true);
 const {baseline,selection,window}=inputs();window.state='unavailable';window.reasons=['incomplete-or-empty-window'];window.baseline.state='no_data';window.baseline.reasons=['empty-distribution'];window.baseline.distribution.count='0';window.baseline.distribution.bucketCounts.fill('0');
 assert(buildNativeHistogramWindowView(baseline,window.stream,window.baseline));assert.equal(buildNativeHistogramComparison(baseline,selection,window),null);
});
test('mixed families, incompatible streams, temporality, layouts, scopes and native diagnostics refuse',()=>{
 for(const mutate of [p=>p.selection=numericFixture().selection,p=>p.selection.scope.selector.stream.resource['service.instance.id']='other',p=>p.selection.histogram.temporality='cumulative',p=>p.selection.rows[1].values.to='1/4',p=>p.selection.scope.selector.kind='measurement-interval',p=>{p.state='diagnostic';p.reason='source-reads-disagree'},p=>p.state='no_data']){const pair=fixture();mutate(pair);assert.equal(validateHistogramComparison(pair).ok,false);assert.equal(model(pair),null)}
});
test('window constructors refuse partial coverage, ambiguity, drift and incompatible definitions without re-evaluation',()=>{
 for(const mutate of [x=>x.window.baseline.fromNS='2000000000',x=>x.window.baseline.reasons=['gap'],x=>x.window.baseline.state='unavailable',x=>x.baseline.histograms.series[0].state='unknown',x=>x.baseline.histograms.series.push(structuredClone(x.baseline.histograms.series[0])),x=>x.selection.question.definition.title='Different definition',x=>x.window.selection.temporality='cumulative',x=>x.window.selection.distribution.bounds=[.2,.8]]){const input=inputs();mutate(input);assert.equal(buildNativeHistogramComparison(input.baseline,input.selection,input.window),null)}
});
test('compact comparison metadata builds the same held pair without rows or interval arrays',()=>{
 const input=inputs(),before=JSON.stringify(input.evidence),pair=buildNativeHistogramComparisonFromEvidence(input.evidence,input.window);
 assert.deepEqual(pair,buildNativeHistogramComparison(input.baseline,input.selection,input.window));assert.equal(JSON.stringify(input.evidence),before);
 assert.equal(input.evidence.baseline.nativeHistogramSource.streams[0].intervals,undefined);assert.equal(input.evidence.baseline.rows,undefined);
});
test('compact metadata admission preserves the overall and member comparison gates',()=>{
 for(const mutate of [x=>x.evidence.state='unavailable',x=>x.window.state='unavailable',x=>x.evidence.baseline.state='unknown',x=>x.evidence.baseline.receipt.question.kind='explore',x=>delete x.evidence.baseline.nativeHistogramSource,x=>x.evidence.selection.nativeHistogramSource.measurementBasis='',x=>x.evidence.baseline.nativeHistogramSource.streams.push(structuredClone(x.evidence.baseline.nativeHistogramSource.streams[0])),x=>x.evidence.baseline.nativeHistogramSource.streams[0].state='unknown']){const input=inputs();mutate(input);assert.equal(buildNativeHistogramComparisonFromEvidence(input.evidence,input.window),null)}
});
test('normalization and SSR stay held, immutable and ineligible for complete-only display',()=>{
 const pair=asHistogramComparison(fixture()),held=parseHistogramComparisonJSON(JSON.stringify(pair));assert(Object.isFrozen(pair));assert(Object.isFrozen(pair.baseline));assert.equal(pair.baseline.origin,'browser-held');assert.equal(hasGuarantee(pair.baseline,'bins'),false);assert.equal(hasGuarantee(held.selection,'bins'),false);assert(model(held,{scale:'observation-share'}));assert.equal(model(held,{requiresComplete:true}),null);
});

test('native aligned models retain all257 buckets and exact uint64 counts without numeric coercion',()=>{
 const input=inputs();for(const side of ['baseline','selection']){const d=input.window[side].distribution;d.bounds=Array.from({length:256},(_,i)=>i);d.bucketCounts=Array.from({length:257},(_,i)=>i===256?'18446744073709551615':'0');d.count='18446744073709551615'}
 const pair=buildNativeHistogramComparison(input.baseline,input.selection,input.window),counts=model(pair),shares=model(pair,{scale:'observation-share'});assert(pair);assert.equal(counts.all.length,257);assert.equal(counts.points.length,257);assert.equal(counts.ceiling,'18446744073709551615');assert.equal(counts.all.at(-1).difference,'0');assert.deepEqual(shares.all.at(-1).shares,['100','100']);assert.equal(sideData(counts,0).points.at(-1).value,'18446744073709551615');
});
