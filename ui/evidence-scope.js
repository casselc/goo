/* Application presenter for typed measurement captures; no reads or persisted schema. */
(function(global){
 function measurement(page,{changes,linked=false}={}){
  const receipt=page.receipt,q=receipt?.question?.measurement;if(!q)return null;const d=q.definition,r=page.measurement,meta=page.summary,identity=typeof module!=='undefined'?require('./service-identity'):global.GooServiceIdentity;
  const from=q.fromNS??String(BigInt(q.fromMs)*1000000n),to=q.toNS??String(BigInt(q.toMs)*1000000n),revision=page.definitionRevision||r?.plan?.revision;
  const facts=[{label:'Population',value:d.kind+' · '+(d.source.metric||'Bound events')},{label:'Scope',value:d.environment+' · '+d.source.services.join(', ')+' · '+identity.label(d.source.namespace)},{label:'Fixed window · UTC',value:'['+new Date(q.fromMs).toISOString()+', '+new Date(q.toMs).toISOString()+')'},{label:'Source records',value:String(meta?.count??r?.count??'Unavailable')},{label:'Binding',value:d.kind==='lifecycle'?'Declared subject and tenant binding':d.source.exactStream?'Pinned exact OTel stream':'All matching independent streams'},{label:'Declared grain',value:d.grain||'Unavailable'}];
  const details=[{label:'Capture ID',value:receipt.id},{label:'Source read',value:page.readAt||r?.readAt||'Unavailable in retained metadata'},{label:'Stored custody time',value:receipt.capturedAt},{label:'Exact question bounds · ns',value:'['+from+', '+to+')'},{label:'Definition',value:d.name},{label:'Definition revision',value:revision||'Unavailable in retained metadata'},{label:'Evidence basis',value:meta?.basis||r?.basis||'Unavailable in retained metadata'},{label:'Scope library',value:d.source.scope||'Omitted'},{label:'Library version',value:d.source.version||'Omitted'},{label:'Schema URL',value:d.source.schemaURL||'Omitted'},{label:'Unit',value:d.source.unit||'No scalar unit'}];
  if(d.histogramBucketMs!==undefined)details.push({label:'Histogram time width · ms',value:String(d.histogramBucketMs)});
  const p=r?.lifecycle?.receipt?.policy;if(p)for(const [label,key]of[['Business evaluation · ns','evaluationTimeNs'],['Knowledge cutoff · ns','evidenceCutoffNs']])if(p[key]!==undefined)details.push({label,value:String(p[key])});
  let notice='Independent capture read; matching windows do not imply a shared snapshot. Source completeness remains unverified.';if(changes)notice=(changes.length?changes.join(' · '):'Matches this question and fixed scope')+'. '+notice;
  if(linked&&!r)notice+=' Result metadata only; full stream identities and diagnostics remain in the capture.';
  return {mode:'frozen',state:meta?.state||r?.state||'unknown',facts,notice,details};
 }
 function element(data){const e=document.createElement('goo-evidence-summary');e.data=data;return e}
 const api={measurement,element};if(typeof module!=='undefined')module.exports=api;else global.GooEvidenceScope=api;
})(typeof window==='undefined'?globalThis:window);
