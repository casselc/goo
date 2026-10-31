const {result}=require('./native-histogram.cjs');
function fixture(){
 const first=result(),stream=structuredClone(first.histograms.series[0].stream),qA=first.question,qB=structuredClone(qA);qB.fromMs=5000;qB.toMs=11000;
 const side=(question,basis,histogramBasis,readAt)=>({state:'observed',count:'2',readAt,receipt:{id:basis,title:'Native captured window',capturedAt:readAt,question:{kind:'measurement',measurement:question}},nativeHistogramSource:{measurementBasis:basis,histogramBasis,streams:[{stream:structuredClone(stream),temporality:'delta',state:'ok'}]}});
 const summary=(question,count,bucketCounts)=>({fromNS:String(BigInt(question.fromMs)*1000000n),toNS:String(BigInt(question.toMs)*1000000n),temporality:'delta',state:'observed',reasons:[],distribution:{unit:'s',bounds:[.125,.5],count,bucketCounts,p95:{state:'unbounded-upper',lower:.5}}});
 const window={state:'comparable',reasons:[],stream,baseline:summary(qA,'4',['1','2','1']),selection:summary(qB,'8',['2','1','5']),p95Relation:'overlap'};
 return {state:'comparable',baseline:side(qA,'a'.repeat(64),'b'.repeat(64),'2026-10-31T00:00:00Z'),selection:side(qB,'c'.repeat(64),'d'.repeat(64),'2026-10-31T00:01:00Z'),nativeWindows:[window],differences:[],warnings:[]};
}
module.exports={fixture};
