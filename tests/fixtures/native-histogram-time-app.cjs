const {fixture:intervalFixture}=require('./native-histogram-app.cjs');
function fixture(){
 const r=intervalFixture(),s=r.histograms.series[0];
 r.question.toMs=14000;r.question.definition.maxGapMs=3000;r.question.definition.histogramBucketMs=4000;
 r.histograms.toNS='14000000000';r.histograms.maxGapNS='3000000000';r.count='6';
 const distribution=(count,bucketCounts)=>({unit:'s',bounds:[.125,.5],bucketCounts,count,p95:count==='0'?{state:'no_data'}:{state:'unbounded-upper',lower:.5}});
 s.temporality='delta';s.intervals=[
  {fromNS:'1000000000',toNS:'3000000000',state:'ok',distribution:distribution('3',['1','1','1'])},
  {fromNS:'3000000000',toNS:'5000000000',state:'ok',distribution:distribution('4',['1','2','1'])},
  {fromNS:'5000000000',toNS:'7000000000',state:'ok',distribution:distribution('0',['0','0','0'])},
  {fromNS:'7000000000',toNS:'9000000000',state:'ok',distribution:distribution('0',['0','0','0'])},
  {fromNS:'9000000000',toNS:'13000000000',state:'unknown',reason:'histogram-no-recorded-value'},
  {fromNS:'13000000000',toNS:'14000000000',state:'ok',distribution:distribution('2',['1','0','1'])}
 ];
 r.histogramTime={apiVersion:'goo/histogram-time-result/v1alpha1',basis:'c'.repeat(64),sourceBasis:r.histograms.basis,fromNS:r.histograms.fromNS,toNS:r.histograms.toNS,widthNS:'4000000000',series:[{stream:structuredClone(s.stream),temporality:s.temporality,buckets:[
  {fromNS:'1000000000',toNS:'5000000000',state:'observed',reasons:[],distribution:distribution('7',['2','3','2'])},
  {fromNS:'5000000000',toNS:'9000000000',state:'no_data',reasons:['empty-distribution'],distribution:distribution('0',['0','0','0'])},
  {fromNS:'9000000000',toNS:'13000000000',state:'unavailable',reasons:['histogram-no-recorded-value']},
  {fromNS:'13000000000',toNS:'14000000000',state:'observed',reasons:[],distribution:distribution('2',['1','0','1'])}
 ]}]};
 return r;
}
module.exports={fixture};
