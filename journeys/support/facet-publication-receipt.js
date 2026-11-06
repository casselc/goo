// Bounded 346 regression admission over observed render records, never CPU limits.
function admitFacetPublicationReceipt(report){
 const requireFact=(ok,message)=>{if(!ok)throw Error('346 receipt: '+message)};
 requireFact(report.phase==='final','expected final receipt');
 requireFact(report.totals?.sourceRequests===18&&report.totals?.comparisonRequests===2,'source/comparison budget');
 for(const width of [1280,390])for(const paired of [false,true]){
  const key=width+'-'+(paired?'pair':'ordinary'),rows=paired?2000:1000;
  for(const kind of ['back','stage','source-publication']){
   const matches=report.operations.filter(o=>o.operation===key+'-'+kind);requireFact(matches.length===1,key+' '+kind+' unique operation');const op=matches[0],actual=op.renders.filter(r=>r.connected),built=actual.map(r=>r.rowsBuilt).filter(n=>n>0);
   requireFact(op.connectedRenderCount===actual.length,key+' '+kind+' reported count differs');requireFact(op.rowsBuilt===op.renders.reduce((sum,r)=>sum+r.rowsBuilt,0),key+' '+kind+' reported rows differ');
   if(kind==='source-publication'){
    requireFact(JSON.stringify(built.sort((a,b)=>a-b))===JSON.stringify(paired?[1000,2000]:[1000]),key+' independent ready publications');requireFact(op.sourceAfter-op.sourceBefore===(paired?2:1),key+' value read boundary');
   }else{
    requireFact(actual.length===1&&op.rowsBuilt===rows,key+' '+kind+' must build once');requireFact(op.sourceBefore===op.sourceAfter&&op.compareBefore===op.compareAfter,key+' '+kind+' performed IO');
   }
  }
 }
 return {accepted:true,backOperations:4,stageOperations:4,sourcePublicationOperations:4,cpuThresholds:false};
}
module.exports={admitFacetPublicationReceipt};
