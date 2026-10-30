// Independent fixed-bucket fixture for boundary and owner tests.
const labels=['[0,1) ms','[1,5) ms','[5,10) ms','[10,50) ms','[50,100) ms','[100,500) ms','[500,1000) ms','[1000,+inf) ms'];
const edges=['0','1000000','5000000','10000000','50000000','100000000','500000000','1000000000','+inf'];
function attach(data){
 const counts=labels.map(label=>data.groups.find(group=>group.g0===label)?.count||'0'),total=String(counts.reduce((n,count)=>n+BigInt(count),0n)),revision=data.basis||'fixture';
 return {...data,histogramDataset:{version:'goo.dataset/1',id:'duration-'+revision,revision,origin:'source-read',...(data.capturedAt?{capturedAt:data.capturedAt}:{}),shape:'histogram',scope:{id:'duration-scope',generation:revision,signal:'spans',selector:data.question},quality:{extent:'complete',basis:'captured-query',value:'recorded',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision,scopeId:'duration-scope',aspects:['bins']},fields:[{key:'from',label:'From',type:'histogram-bound',unit:'ns'},{key:'to',label:'To',type:'histogram-bound',unit:'ns'},{key:'count',label:'Captured records',type:'number',unit:'records'}],histogram:{kind:'span-duration',binning:'goo-duration-v1',unit:'ns',total,closure:'left-closed-right-open'},rows:counts.map((count,index)=>({key:'bin:'+index,values:{from:edges[index],to:edges[index+1],count},state:count==='0'?'no_data':'observed'}))}};
}
module.exports={attach,labels,edges};
