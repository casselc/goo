import './evidence-summary.mjs';
/* Application presenters for typed evidence; no reads or persisted schema. */
import * as ServiceIdentity from './service-identity.mjs';
import * as Environment from './record-environment.mjs';
const recordEnvironment=q=>q?.environmentSelector?Environment.label(q.environmentSelector):q?.environment;
const global=globalThis;
const clock=(label,value)=>({label,value,...(/^\d{4}-\d{2}-\d{2}T.*Z$/.test(value)?{format:'utc'}:{})});
function measurement(page,{changes,linked=false}={}){
  const receipt=page.receipt,q=receipt?.question?.measurement;if(!q)return null;const d=q.definition,r=page.measurement,meta=page.summary,identity=ServiceIdentity;
  const from=q.fromNS??String(BigInt(q.fromMs)*1000000n),to=q.toNS??String(BigInt(q.toMs)*1000000n),revision=page.definitionRevision||r?.plan?.revision;
  const facts=[{label:'Population',value:d.kind+' · '+(d.source.metric||'Bound events')},{label:'Scope',value:d.environment+' · '+(['host','kubernetes-pod'].includes(d.source.resource?.kind)?(d.source.resource.kind==='host'?'Host ':'Pod ')+d.source.resource.id:(d.source.services||[]).join(', ')+' · '+identity.label(d.source.namespace))},{label:'Fixed window · UTC',value:'['+from+', '+to+')',format:'unix-ns-range'},{label:'Source records',value:String(meta?.count??r?.count??'Unavailable')},{label:'Binding',value:d.kind==='lifecycle'?'Declared subject and tenant binding':d.source.exactStream?'Pinned exact OTel stream':'All matching independent streams'},{label:'Declared grain',value:d.grain||'Unavailable'}];
  const details=[{label:'Capture ID',value:receipt.id},clock('Source read',page.readAt||r?.readAt||'Unavailable in retained metadata'),clock('Stored custody time',receipt.capturedAt),{label:'Exact question bounds · UTC',value:'['+from+', '+to+')',format:'unix-ns-range'},{label:'Definition',value:d.name},{label:'Definition revision',value:revision||'Unavailable in retained metadata'},{label:'Evidence basis',value:meta?.basis||r?.basis||'Unavailable in retained metadata'},{label:'Scope library',value:d.source.scope||'Omitted'},{label:'Library version',value:d.source.version||'Omitted'},{label:'Schema URL',value:d.source.schemaURL||'Omitted'},{label:'Unit',value:d.source.unit||'No scalar unit'}];
  if(d.histogramBucketMs!==undefined)details.push({label:'Histogram time width · ms',value:String(d.histogramBucketMs)});
  const p=r?.lifecycle?.receipt?.policy;if(p)for(const [label,key]of[['Business evaluation · UTC','evaluationTimeNs'],['Knowledge cutoff · UTC','evidenceCutoffNs']])if(p[key]!==undefined)details.push({label,value:String(p[key]),format:'unix-ns'});
  let notice='Independent capture read; matching windows do not imply a shared snapshot. Source completeness remains unverified.';if(changes)notice=(changes.length?changes.join(' · '):'Matches this question and fixed scope')+'. '+notice;
  if(linked&&!r)notice+=' Result metadata only; full stream identities and diagnostics remain in the capture.';
  return {mode:'frozen',state:meta?.state||r?.state||'unknown',facts,notice,details};
 }
// Full capture pages and metadata-only document references share receipt facts.
function record(page,{changes,linked=false}={}){const receipt=page.receipt,q=receipt?.question?.explore;if(!q)return null;const r=page.explore,meta=page.summary,count=String(r?.count??meta?.count??'Unavailable'),state=r?.state||meta?.state||'unknown',predicate=JSON.stringify(q.where||{});const facts=[{label:'Population',value:q.signal+' · captured records'},{label:'Scope',value:recordEnvironment(q)},{label:'Fixed window · UTC',value:'['+BigInt(q.fromMs)*1000000n+', '+BigInt(q.toMs)*1000000n+')',format:'unix-ns-range'},{label:'Source records',value:count}];const details=[{label:'Question definition',value:'Ad hoc fixed question; no reusable definition identity is inferred.'},{label:'Capture ID',value:receipt.id},clock('Source read',(r?r.capturedAt:page.readAt)||'Unavailable in retained metadata'),clock('Stored custody time',receipt.capturedAt),{label:'Evidence basis',value:r?.basis||meta?.basis||'Unavailable in retained metadata'},{label:'Record predicate',value:predicate.length<=8192?predicate:'Selector exceeds summary display budget; inspect the exact receipt.'}];if(r)details.push({label:'Displayed page',value:String(r.rows.length)+' records · offset '+r.question.offset},{label:'Retained inspection records',value:String(page.retainedRecords)},{label:'Retention limit',value:page.recordsLimited?'More source records were not retained':'All records returned by the capture read were retained; source coverage unverified'});let notice='Independent capture read; matching windows do not imply a shared snapshot. Coverage remains unknown.';if(state==='changed_between_reads')notice='Source count and page reads disagreed. '+notice;if(changes)notice=(changes.length?changes.join(' · '):'Matches this question and fixed scope')+'. '+notice;if(linked&&!r)notice+=' Metadata only; inspect stored pages in the capture.';return {mode:'frozen',state:state==='changed_between_reads'?'unknown':state,facts,details,notice}}
// Plan scope is validated by the dashboard adapter before publishing any result.
function serviceWidget(widget,result,plan){
  const q=result.question,d=q.definition,native=!!widget.measurement,identity=ServiceIdentity;
  const from=q.fromNS??String(BigInt(q.fromMs)*1000000n),to=q.toNS??String(BigInt(q.toMs)*1000000n);
  const facts=[{label:'Population',value:native?d.kind+' · '+d.source.metric+' ['+d.source.unit+']':q.signal+' · captured records'},{label:'Scope',value:plan.scope.environment+' · '+plan.service+' · '+identity.label(plan.scope.namespace)},{label:'Fixed window · UTC',value:'['+from+', '+to+')',format:'unix-ns-range'}];
  const details=[clock('Source read',(native?result.readAt:result.capturedAt)||'Unavailable in result'),{label:'Evidence basis',value:result.basis||'Unavailable in result'},{label:'Source records',value:String(result.count??'Unavailable')},{label:'Exact question bounds · UTC',value:'['+from+', '+to+')',format:'unix-ns-range'},{label:'Declared grain',value:widget.grain||d?.grain||'Unavailable'}];
  if(native)details.push({label:'Binding',value:d.source.exactStream?'Pinned exact OTel stream':'All matching independent streams'},{label:'Scope library',value:d.source.scope||'Omitted'},{label:'Library version',value:d.source.version||'Omitted'},{label:'Schema URL',value:d.source.schemaURL||'Omitted'});
  return {mode:'fixed',state:result.state||'unknown',facts:facts.slice(0,1),details:[facts[1],facts[2],...details],notice:'Independent read · completeness unverified. Capturing evidence performs a new read.'};
 }
function comparisonSide(side){
  const receipt=side.receipt,q=receipt.question.measurement||receipt.question.explore,d=q.definition,identity=ServiceIdentity;
  const from=q.fromNS??String(BigInt(q.fromMs)*1000000n),to=q.toNS??String(BigInt(q.toMs)*1000000n);
  const facts=[{label:'Population',value:d?d.kind+' · '+(d.source.metric||'Bound events')+' ['+(d.source.unit||'No scalar unit')+']':q.signal+' · captured records'},{label:'Source records',value:String(side.count??'Unavailable')},{label:'Fixed window · UTC',value:'['+from+', '+to+')',format:'unix-ns-range'}];
  const details=[{label:'Capture ID',value:receipt.id},clock('Source read',side.readAt||'Unavailable in retained metadata'),clock('Stored custody time',receipt.capturedAt),{label:'Environment',value:recordEnvironment(q)||d.environment}];
  if(d)details.push({label:'Scope',value:(['host','kubernetes-pod'].includes(d.source.resource?.kind)?(d.source.resource.kind==='host'?'Host ':'Pod ')+d.source.resource.id:(d.source.services||[]).join(', ')+' · '+identity.label(d.source.namespace))},{label:'Declared grain',value:d.grain||'Unavailable'},{label:'Binding',value:d.source.exactStream?'Pinned exact OTel stream':'All matching independent streams'},{label:'Scope library',value:d.source.scope||'Omitted'},{label:'Library version',value:d.source.version||'Omitted'},{label:'Schema URL',value:d.source.schemaURL||'Omitted'});
  else details.push({label:'Record predicate',value:JSON.stringify(q.where||{})});
  return {mode:'frozen',state:side.state||'unknown',facts,details,notice:'Independent stored read · source completeness unverified.'};
 }
// Supplied definitions and read metadata only; never infer completeness or custody.
function comparisonContextSide(q,{mode='fixed',state='unknown',count,readAt,basis,capturedAt}={}){
  const from=q.fromNS??String(BigInt(q.fromMs)*1000000n),to=q.toNS??String(BigInt(q.toMs)*1000000n);
  const facts=[{label:'Source records',value:String(count??'Unavailable')},{label:'Window · UTC',value:'['+from+', '+to+')',format:'unix-ns-range'}];
  const details=[{label:'Environment',value:recordEnvironment(q)||q.definition?.environment||'Unavailable'},clock('Source read',readAt||'Unavailable in supplied metadata'),{label:'Evidence basis',value:basis||'Unavailable in supplied metadata'}];
  if(capturedAt)details.push(clock('Stored custody time',capturedAt));
  return {mode,state:state==='changed_between_reads'?'unknown':state,facts,details,notice:''};
 }
function element(data){const e=document.createElement('goo-evidence-summary');e.data=data;return e}

export {measurement,record,serviceWidget,comparisonSide,comparisonContextSide,element};
