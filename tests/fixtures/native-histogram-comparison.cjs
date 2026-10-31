const {buildNativeHistogramComparison}=require('../../ui/native-histogram-view.mjs');
function inputs(){
 const stream={resource:{'service.name':'worker','service.instance.id':'one','deployment.environment.name':'prod'},scope:'otel.work',version:'1',schemaURL:'',name:'work.duration',unit:'s',attributes:{route:'/work'}};
 const definition={apiVersion:'goo/measurement-package/v1alpha1',name:'native-pair',title:'Worker duration',kind:'histogram',environment:'prod',lookbackMinutes:2,maxGapMs:3000,grain:'per-stream',source:{services:['worker'],scope:stream.scope,version:stream.version,schemaURL:stream.schemaURL,metric:stream.name,unit:'s'}};
 const distribution=(count,bucketCounts)=>({unit:'s',bounds:[.1,.5],bucketCounts,count,p95:{state:'unbounded-upper',lower:.5}});
 function result(side,fromMs,toMs,counts){const fromNS=String(BigInt(fromMs)*1000000n),toNS=String(BigInt(toMs)*1000000n);return {question:{definition:structuredClone(definition),fromMs,toMs},basis:(side?'c':'a').repeat(64),readAt:side?'2026-10-31T01:00:01.123456789Z':'2026-10-31T01:00:00.123456789Z',state:'observed',count:side?'3':'2',rows:[],histograms:{apiVersion:'goo/histogram-result/v1alpha1',basis:(side?'d':'b').repeat(64),fromNS,toNS,maxGapNS:'3000000000',warnings:[],series:[{stream:structuredClone(stream),temporality:'delta',state:'ok',intervals:[]}]},distribution:distribution(String(counts.reduce((a,b)=>a+b,0)),counts.map(String))}}
 const baseline=result(0,1000,5000,[1,2,1]),selection=result(1,20000,26000,[1,2,3]);
 function summary(r){return {fromNS:r.histograms.fromNS,toNS:r.histograms.toNS,temporality:'delta',state:'observed',reasons:[],distribution:r.distribution}}
 const window={stream:structuredClone(stream),state:'comparable',reasons:[],baseline:summary(baseline),selection:summary(selection),buckets:[],p95Relation:'overlap-or-unbounded'};
 const side=r=>({receipt:{question:{kind:'measurement',measurement:structuredClone(r.question)}},readAt:r.readAt,state:r.state,count:r.count,nativeHistogramSource:{measurementBasis:r.basis,histogramBasis:r.histograms.basis,streams:r.histograms.series.map(({stream,temporality,state})=>({stream:structuredClone(stream),temporality,state}))}});
 return {baseline,selection,window,evidence:{state:'comparable',baseline:side(baseline),selection:side(selection),nativeWindows:[window]}};
}
function fixture(){const {baseline,selection,window}=inputs(),pair=buildNativeHistogramComparison(baseline,selection,window);if(!pair)throw Error('Invalid native comparison fixture');return pair}
module.exports={inputs,fixture};
