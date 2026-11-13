import {object} from './intent-shape.mjs';
import {validateRecordQuestion as admitRecord} from './record-intent.mjs';
/* Pure question/step/link protocol. No navigation, DOM, source reads or publication. */
import * as RecordEnvironment from './record-environment.mjs';
import * as QuestionSession from './question-session.mjs';
import {validateNumericPredicate} from './numeric-predicate.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
const global=globalThis;
// Versioned links contain questions, never evidence or source credentials.
function exactBounds(q){return q.fromNS!==undefined||q.toNS!==undefined?{fromNS:q.fromNS,toNS:q.toNS}:{}}
function questionFromEvidence(receipt){
  if(!/^[a-f0-9]{64}$/.test(receipt?.id))throw Error('An exact capture ID is required.');
  const q=receipt.question;let step;if(q?.kind==='measurement'&&q.measurement)step=measurementStep(q.measurement);else if(q?.kind==='explore'&&q.explore)step=recordStep(q.explore);else if(q?.kind==='trace-search'&&q.traceSearch)step=traceSearchStep(q.traceSearch);else throw Error('This evidence cannot become a reusable question.');
  validateStep(step);return {kind:'question',step,evidence:[receipt.id]};
 }
function captureQuestion(step){validateStep(step);if(step.kind==='trace-search')return {kind:'trace-search',traceSearch:structuredClone(step.traceSearch)};return step.kind==='record'?{kind:'explore',explore:structuredClone(step.record)}:{kind:'measurement',measurement:{definition:structuredClone(step.measurement),fromMs:step.scope.fromMs,toMs:step.scope.toMs,...exactBounds(step.scope)}}}
// Comparison-only Go zero/omitempty defaults; stored definitions and digests are untouched.
function nativeDefinitionForComparison(definition){
  const d=structuredClone(definition),source=d.source;delete d.environment;delete d.environmentSelector;
  for(const k of ['scope','version','schemaURL'])if(source[k]===undefined||source[k]===null)source[k]='';
  for(const k of ['metric','unit'])if(source[k]==='')delete source[k];
  if(source.attributes===null||source.attributes&&typeof source.attributes==='object'&&!Array.isArray(source.attributes)&&Object.keys(source.attributes).length===0)delete source.attributes;
  for(const k of ['namespace','exactStream','resource'])if(source[k]===null)delete source[k];
  if(source.namespace?.value==='')delete source.namespace.value;
  if(['host','kubernetes-pod','kubernetes-container'].includes(source.resource?.kind)&&source.services==null)source.services=[];
  if(['host','kubernetes-pod','kubernetes-container'].includes(source.resource?.kind)&&source.exactStream&&source.exactStream.attributes==null)source.exactStream.attributes={};
  if(d.stateProfile===null||d.stateProfile==='')delete d.stateProfile;
  if(d.stockDomain===null||d.stockDomain==='')delete d.stockDomain;
  for(const k of ['histogramBucketMs','lifecycle'])if(d[k]===null)delete d[k];
  return d;
 }
function evidenceDifference(step,capture){
  const identify=QuestionSession?.identity||(null);const expected=captureQuestion(step),actual=capture,changes=[];if(expected.kind!==actual.kind)return ['Signal family differs'];
  const a=expected.explore||expected.measurement||expected.traceSearch,b=actual.explore||actual.measurement||actual.traceSearch;if(!b)return ['Question unavailable'];
  if(a.definition? !RecordEnvironment.sameQuestionEnvironment(a.definition,b.definition) : !RecordEnvironment.sameQuestionEnvironment(a,b))changes.push('Environment differs');
  if(a.fromMs!==b.fromMs||a.toMs!==b.toMs||a.fromNS!==b.fromNS||a.toNS!==b.toNS)changes.push('Fixed window differs');
  function definition(q){if(q.definition)return stable(nativeDefinitionForComparison(q.definition));const d=structuredClone(q);for(const k of ['environment','environmentSelector','fromMs','toMs','offset',...(expected.kind==='trace-search'?[]:['limit'])])delete d[k];return identify?identify(d):stable(d)}
  function stable(x){if(Array.isArray(x))return '['+x.map(stable).join(',')+']';if(x&&typeof x==='object')return '{'+Object.keys(x).sort().map(k=>JSON.stringify(k)+':'+stable(x[k])).join(',')+'}';return JSON.stringify(x)}
  if(definition(a)!==definition(b))changes.push('Question definition differs');return changes;
 }
function recordStep(q,view='Records'){validateRecordQuestion(q,view);return {version:1,kind:'record',scope:{...RecordEnvironment.questionFields(q),fromMs:q.fromMs,toMs:q.toMs},view,record:structuredClone(q)}}
function measurementStep(q){return {version:1,kind:'measurement',scope:{...RecordEnvironment.questionFields(q.definition),fromMs:q.fromMs,toMs:q.toMs,...exactBounds(q)},view:'Overview',measurement:structuredClone(q.definition)}}
function traceSearchStep(request){validateTraceSearchRequest(request);return {version:1,kind:'trace-search',scope:{environment:request.environment||'',...RecordEnvironment.questionFields(request),fromMs:request.fromMs,toMs:request.toMs},view:'Traces',traceSearch:structuredClone(request)}}
function traceSearchRequestBytes(request){const normalized=value=>Array.isArray(value)?value.map(normalized):value&&typeof value==='object'?Object.fromEntries(Object.entries(value).filter(([key,item])=>!(key==='environment'&&item===''||key==='value'&&item===''||key==='scale'&&item===0)).map(([key,item])=>[key,normalized(item)])):value;const wire=JSON.stringify(normalized(request)).replace(/[<>&\u2028\u2029]/g,c=>'\\u'+c.charCodeAt(0).toString(16).padStart(4,'0'));return new TextEncoder().encode(wire).length;}
function validateTraceSearchRequest(request){
 const object=(value,keys)=>value&&typeof value==='object'&&!Array.isArray(value)&&Object.keys(value).every(key=>keys.includes(key));
 const validJSON=value=>{if(typeof value==='string')return value.isWellFormed();if(value===null)return false;if(typeof value==='boolean')return true;if(typeof value==='number')return Number.isFinite(value);if(Array.isArray(value))return value.every(validJSON);return value&&typeof value==='object'&&(Object.getPrototypeOf(value)===Object.prototype||Object.getPrototypeOf(value)===null)&&Object.entries(value).every(([key,item])=>key.isWellFormed()&&validJSON(item))};
 if(!object(request,['version','environment','environmentSelector','fromMs','toMs','conditions','limit'])||!validJSON(request)||request.version!=='goo/trace-search/v1'||!Number.isSafeInteger(request.fromMs)||!Number.isSafeInteger(request.toMs)||request.fromMs<0||request.toMs<=request.fromMs||request.toMs-request.fromMs>7200000||request.toMs>4102444800000||!Number.isInteger(request.limit)||request.limit<1||request.limit>100||!Array.isArray(request.conditions)||request.conditions.length<1||request.conditions.length>4)throw Error('Invalid fixed trace-search request.');
 RecordEnvironment.selection(request);let nodes=0;const visit=p=>{nodes++;if(p.all||p.any){if(!object(p,[p.all?'all':'any']))throw Error('Invalid span-condition group fields.');for(const child of p.all||p.any)visit(child)}else if(!object(p,p.numeric?['numeric','op','value']:['field','op','value'])||p.field&&!object(p.field,['kind','key'])||Object.hasOwn(p,'value')&&typeof p.value!=='string')throw Error('Unsupported span-condition fields.');};for(const condition of request.conditions){if(!condition)throw Error('Each span condition requires a predicate.');validatePredicate(condition,'spans');visit(condition)}if(nodes>64)throw Error('Trace search exceeds 64 predicate nodes.');if(traceSearchRequestBytes(request)>65536)throw Error('Trace-search intent exceeds the 64 KiB request budget.');return request;
}
function validateStep(s){
  const closed=(value,keys)=>{try{object(value,keys);return true}catch{return false}};
  if(!closed(s,['version','kind','scope','view','record','measurement','traceSearch'])||!closed(s.scope,['environment','environmentSelector','fromMs','toMs','fromNS','toNS']))throw Error('Invalid question step fields.');
  const b=s.scope;if(s?.version!==1||!b||!Number.isSafeInteger(b.fromMs)||!Number.isSafeInteger(b.toMs)||b.fromMs<0||b.toMs<=b.fromMs||b.toMs-b.fromMs>7200000||b.toMs>4102444800000)throw Error('Invalid question step scope.');
  if(b.fromNS!==undefined||b.toNS!==undefined){if(s.kind!=='measurement'||s.measurement?.kind==='lifecycle'||typeof b.fromNS!=='string'||typeof b.toNS!=='string'||!/^(0|[1-9][0-9]*)$/.test(b.fromNS)||!/^(0|[1-9][0-9]*)$/.test(b.toNS))throw Error('Invalid exact native scope.');const f=BigInt(b.fromNS),t=BigInt(b.toNS);if(t<=f||t-f>7200000000000n||t>=4102444800000000000n||BigInt(b.fromMs)!==f/1000000n||BigInt(b.toMs)!==(t+999999n)/1000000n)throw Error('Exact scope differs from enclosing bounds.');}
  if(s.kind==='record'){validateRecordQuestion(s.record,s.view);RecordEnvironment.selection(b);if(Object.hasOwn(s,'traceSearch')||Object.hasOwn(s,'measurement')||!s.record||!RecordEnvironment.sameQuestionEnvironment(s.record,b)||s.record.fromMs!==b.fromMs||s.record.toMs!==b.toMs)throw Error('Record step scope differs.');validateLink({version:1,question:s.record,view:s.view})}
  else if(s.kind==='measurement'){if(Object.hasOwn(s,'traceSearch')||Object.hasOwn(s,'record')||s.view!=='Overview'||!s.measurement||!RecordEnvironment.sameQuestionEnvironment(s.measurement,b)||!s.measurement.source||typeof s.measurement.source!=='object'||Array.isArray(s.measurement.source)||s.measurement.environmentSelector!=null&&(!['gauge','counter','histogram'].includes(s.measurement.kind)||s.measurement.grain!=='per-stream'||s.measurement.source.resource!=null||s.measurement.kind==='histogram'&&!['s','ms'].includes(s.measurement.source.unit)))throw Error('Invalid measurement step.');}
  else if(s.kind==='trace-search'){if(Object.hasOwn(s,'record')||Object.hasOwn(s,'measurement')||s.view!=='Traces'||!s.traceSearch||b.fromNS!==undefined||b.toNS!==undefined)throw Error('Trace search requires one typed request.');validateTraceSearchRequest(s.traceSearch);if(!RecordEnvironment.sameQuestionEnvironment(s.traceSearch,b)||s.traceSearch.fromMs!==b.fromMs||s.traceSearch.toMs!==b.toMs)throw Error('Trace-search request differs from step scope.');}
  else throw Error('Unsupported question step.');return s;
 }
function stepLink(s){validateStep(s);const hash='#step='+encodeURIComponent(JSON.stringify(s));if(hash.length>(s.kind==='trace-search'?262144:48000))throw Error('Question step is too large to share.');return hash}
function readStepLink(hash){if(!hash.startsWith('#step='))return null;if(hash.length>262144)throw Error('Question step is too large.');const step=validateStep(JSON.parse(decodeURIComponent(hash.slice(6))));if(hash.length>(step.kind==='trace-search'?262144:48000))throw Error('Question step is too large.');return step}
function sameWindow(a,b){return RecordEnvironment.sameQuestionEnvironment(a,b)&&a.fromMs===b.fromMs&&a.toMs===b.toMs&&a.fromNS===b.fromNS&&a.toNS===b.toNS}
function questionLink(question,view='Records'){
  validateLink({version:1,question,view});
  const hash=stepLink(recordStep(question,view));if(hash.length>16000)throw Error('Question is too large to share in a link.');return hash;
 }
// Browser intent guards only; Go remains authoritative for execution semantics.

function validateLink(value){return validateRecord(value)}
function validateRecordQuestion(question,view='Records'){validateRecord({version:1,question,view});return question}
function validateRecord(value){
  object(value,['version','question','view'],'Unsupported shared question fields.');
  if(value?.version!==1||!['Records','Distribution'].includes(value.view))throw Error('Invalid shared question view.');
  admitRecord(value.question);return value;
 }

function readQuestionLink(hash){if(hash.startsWith('#step=')){const s=readStepLink(hash);return s.kind==='record'?{version:1,question:s.record,view:s.view}:null}if(!hash.startsWith('#question='))return null;if(hash.length>16000)throw Error('Question link is too large.');return validateLink(JSON.parse(decodeURIComponent(hash.slice(10))));}

export {questionFromEvidence,captureQuestion,evidenceDifference,recordStep,measurementStep,traceSearchStep,traceSearchRequestBytes,validateTraceSearchRequest,validateStep,stepLink,readStepLink,sameWindow,questionLink,readQuestionLink,validateLink,validateNumericPredicate,validateRecordQuestion};
