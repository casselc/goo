const {result}=require('./native-histogram.cjs');
function fixture(){
 const r=result(),s=r.histograms.series[0];
 const interval=(from,to,count,bucketCounts)=>({fromNS:from,toNS:to,state:'ok',distribution:{unit:'s',bounds:[.125,.5],bucketCounts,count,p95:count==='0'?{state:'no_data'}:{state:'unbounded-upper',lower:.5}}});
 s.intervals=[interval('1000000000','2000000000','4',['1','2','1']),{fromNS:'2000000000',toNS:'3000000000',state:'unknown',reason:'histogram-no-recorded-value'},interval('3000000000','4000000000','0',['0','0','0'])];
 return {...r,state:'observed',count:'3',warnings:[],rows:[]};
}
module.exports={fixture};
