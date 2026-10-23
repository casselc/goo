/* Supplied retained-content presentation; no fetch, navigation or publication. */
(function(global){
 // Keep JSON numeric tokens that JavaScript cannot round-trip as explicit text.
 // Matching whole string literals first prevents touching digits inside strings.
 function displayJSON(raw){let exact=0;const converted=raw.replace(/"(?:\\.|[^"\\])*"|-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?/g,token=>{if(token[0]==='"')return token;const n=Number(token);if(!Number.isFinite(n)||String(n)!==token||Number.isInteger(n)&&!Number.isSafeInteger(n)){exact++;return JSON.stringify('Exact JSON number: '+token)}return token});return {value:JSON.parse(converted),exact}}
 function model(content){
  if(!['saved-capture','result-page','tail-selection'].includes(content.origin)||typeof content.manifestJSON!=='string')throw Error('Unsupported retained evidence origin');
  const parsed=displayJSON(content.manifestJSON),m=parsed.value,tail=content.origin==='tail-selection';
  if(m.id!==(tail?content.tail?.id:content.capture?.id)||(tail?(!m.selection||content.capture!=null):(!m.capture||content.tail!=null)))throw Error('Retained content differs from its origin');
  if(tail){const s=m.selection;return {id:m.id,origin:content.origin,signal:s.selector.signal,count:s.members.length,clock:'Retained at',at:content.retainedAt,exact:parsed.exact,manifestJSON:content.manifestJSON,
   details:[{label:'Environment',value:s.selector.environment},{label:'Signal',value:s.selector.signal},{label:'Connection',value:s.connection}],notice:'Observed Activity occurrences. No claim of complete traces, replay or population totals.',
   members:s.members.map((v,i)=>({label:'Stored observation '+(i+1),row:v.record,provenance:{facts:[{label:'Read at',value:v.observedAt,format:'utc'},{label:'Window from',value:new Date(v.fromMs).toISOString(),format:'utc'},{label:'Window before',value:new Date(v.toMs).toISOString(),format:'utc'}],details:[{label:'Read basis',value:v.basis},{label:'Occurrence',value:v.ref.key},{label:'Late observation',value:v.late?'Yes':'No'},{label:'Limited read',value:v.limited?'Yes':'No'}],notice:'Original half-open poll window; event time remains on the observation.'}}))};
  }
  return {id:m.id,origin:content.origin,signal:m.capture.question?.explore?.signal||'logs',count:m.members.length,clock:'Captured at',at:m.capture.capturedAt,exact:parsed.exact,manifestJSON:content.manifestJSON,original:content.origin==='saved-capture'?m.capture.id:null,
   details:[{label:content.origin==='result-page'?'Original result token':'Original capture ID',value:m.capture.id},{label:'Original evidence basis',value:m.basis},{label:'Source count at capture',value:m.sourceCount},{label:'Stored inspection records',value:String(m.storedRecords)}],notice:m.coverage,members:m.members.map(v=>({label:'Stored observation '+(v.index+1),row:v.row}))};
 }
 function summary(value,extra={}){const el=global.GooUI.node('goo-evidence-summary');el.data={mode:'frozen',state:'unknown',facts:[{label:'Origin',value:({'tail-selection':'Activity selection','saved-capture':'Saved capture','result-page':'Result page'})[value.origin]},{label:'Observations',value:String(value.count)},{label:value.clock,value:value.at,format:'utc'}],details:value.details,notice:value.notice,...extra};return el}
 function observations(host,value){const node=global.GooUI.node;if(value.exact)host.append(node('p','Numeric tokens beyond browser precision are shown as explicit text. Original manifest JSON preserves their numeric types.','evidence-note'));
  for(const member of value.members){const detail=node('details'),content=node('div');detail.className='retained-observation';detail.append(node('summary',member.label),content);let built=false;detail.addEventListener('toggle',()=>{if(!detail.open||built)return;built=true;
   const description=global.GooRecordPresentation.description(member.row);description.className='record-description';content.append(description,global.GooRecordPresentation.summary(member.row,value.signal));
   if(member.provenance){const provenance=node('goo-evidence-summary');provenance.data={mode:'frozen',state:'unknown',...member.provenance};content.append(provenance)}
   const attributes=node('details'),fields=node('div');attributes.className='retained-attributes';attributes.append(node('summary','Observation attributes'),fields);let inspected=false;attributes.addEventListener('toggle',()=>{if(attributes.open&&!inspected){inspected=true;global.GooAttributes.mount(fields,member.row,{raw:false,renderField:entry=>entry.key==='timestamp_ns'&&!entry.field?global.GooRecordPresentation.timestamp(member.row):null})}});content.append(attributes);
  });host.append(detail)}
  const raw=node('details');raw.className='retained-original-json';raw.append(node('summary','Original manifest JSON'),node('pre',value.manifestJSON));host.append(raw);
 }

 const api={displayJSON,model,summary,observations};if(typeof module!=='undefined')module.exports=api;else global.GooRetentionPresentation=api;
})(typeof window==='undefined'?globalThis:window);
