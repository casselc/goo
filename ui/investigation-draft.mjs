/* Bounded same-tab authored draft recovery. No DOM, requests or publication. */
import * as LogPatternIdentity from './log-pattern-identity.mjs';
import * as QuestionProtocol from './question-protocol.mjs';
const global=globalThis;
'use strict';
const MAX_BYTES=256*1024,TTL_MS=7*86400000,ID=/^[a-f0-9]{64}$/;
const clone=x=>JSON.parse(JSON.stringify(x)),bytes=x=>new TextEncoder().encode(x).length;
function requireThat(ok,message='Invalid investigation draft'){if(!ok)throw Error(message)}
function object(v,keys){requireThat(v!==null&&typeof v==='object'&&!Array.isArray(v));requireThat(Object.keys(v).every(k=>keys.includes(k)),'Unsupported investigation draft field')}
function text(v,max,multiline=false){requireThat(typeof v==='string'&&bytes(v)<=max&&!/[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?<![\uD800-\uDBFF])[\uDC00-\uDFFF]/u.test(v));requireThat(!(multiline?/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x9F]/:/[\x00-\x1F\x7F-\x9F]/).test(v))}
function integer(v,min,max){requireThat(Number.isSafeInteger(v)&&v>=min&&v<=max)}
function optional(v,k,fn){if(Object.hasOwn(v,k)&&v[k]!==null&&v[k]!==undefined)fn(v[k])}
function strings(v,count,max){requireThat(Array.isArray(v)&&v.length<=count);for(const x of v)text(x,max);requireThat(new Set(v).size===v.length)}
function map(v,count,keyMax,valueMax){requireThat(v&&typeof v==='object'&&!Array.isArray(v)&&Object.keys(v).length<=count);for(const [k,x]of Object.entries(v)){text(k,keyMax);requireThat(k!=='');text(x,valueMax)}}
function namespace(v){object(v,['mode','value']);requireThat(['missing','value'].includes(v.mode));optional(v,'value',x=>text(x,128));requireThat(v.mode!=='missing'||!v.value)}
function field(v){object(v,['kind','key']);requireThat(['column','attribute','resource'].includes(v.kind));text(v.key,128);requireThat(v.key!=='')}
function record(q){
  object(q,['reportedEdge','numeric','numericHistogram','groupTimeSeries','spanStatus','bucketMs','summary','signal','environment','fromMs','toMs','where','groupBy','limit','offset']);
  const walk=(p,depth=0,state={n:0})=>{object(p,['all','any','field','numeric','op','value']);requireThat(depth<=4&&++state.n<=32);if(p.all||p.any){requireThat(!(p.all&&p.any)&&!p.field&&!p.numeric&&!p.op&&!p.value);const list=p.all||p.any;requireThat(Array.isArray(list)&&list.length>0);for(const c of list)walk(c,depth+1,state)}else if(p.numeric){const protocol=QuestionProtocol;protocol.validateNumericPredicate(p)}else{field(p.field);requireThat(['eq','ne','contains','exists','missing','gte','lte'].includes(p.op));optional(p,'value',x=>text(x,1024))}};
  optional(q,'where',x=>walk(x));optional(q,'groupBy',a=>{requireThat(Array.isArray(a)&&a.length<=2);a.forEach(field)});
  optional(q,'numericHistogram',v=>{object(v,['binning']);requireThat(!!q.numeric&&v.binning==='powers-of-two')});
  optional(q,'numeric',v=>{object(v,['field','type','unit','parse','scale']);field(v.field);requireThat(['attribute','resource'].includes(v.field.kind)&&['int64','uint64','decimal64'].includes(v.type)&&v.parse==='canonical-decimal');text(v.unit,64);const protocol=QuestionProtocol;protocol.validateNumericPredicate({numeric:v,op:'valid'})});
  optional(q,'reportedEdge',v=>{object(v,['parent','child','discoveryNamespace']);for(const key of ['parent','child']){object(v[key],['service','namespace']);text(v[key].service,128);namespace(v[key].namespace)}optional(v,'discoveryNamespace',namespace)});
  for(const key of ['groupTimeSeries','spanStatus'])optional(q,key,v=>requireThat(typeof v==='boolean'));
  optional(q,'summary',v=>requireThat(['','span-duration'].includes(v)));optional(q,'bucketMs',v=>requireThat([0,1000,10000,60000,300000].includes(v)));
  const checkPattern=p=>{if(p.all||p.any){(p.all||p.any).forEach(checkPattern);return}if(p.field?.kind==='column'&&p.field.key==='body_pattern_v1'){requireThat(q.signal==='logs'&&['eq','ne','exists','missing'].includes(p.op));if(['eq','ne'].includes(p.op))(LogPatternIdentity).decode(p.value);else requireThat(p.value===undefined||p.value==='')}};if(q.where)checkPattern(q.where);requireThat(q.signal==='logs'||!q.groupBy?.some(f=>f.kind==='column'&&f.key==='body_pattern_v1'));
  optional(q,'limit',v=>integer(v,1,100));optional(q,'offset',v=>integer(v,0,1000));
 }
function native(d){
  object(d,['stateProfile','stockDomain','histogramBucketMs','apiVersion','name','title','kind','environment','lookbackMinutes','maxGapMs','grain','source','lifecycle']);requireThat(bytes(JSON.stringify(d))<=32768);
  requireThat(d.apiVersion==='goo/measurement-package/v1alpha1'&&/^[A-Za-z][A-Za-z0-9_.-]{0,127}$/.test(d.name));text(d.title,160);text(d.environment,128);requireThat(['gauge','stock','counter','histogram','lifecycle','state'].includes(d.kind)&&['per-stream','consumer-group','subject'].includes(d.grain));integer(d.lookbackMinutes,1,120);integer(d.maxGapMs,1,600000);optional(d,'histogramBucketMs',v=>integer(v,1,7200000));optional(d,'stockDomain',v=>requireThat(v===''||d.kind==='stock'&&v==='nonnegative-integer'));if(d.kind==='state')requireThat(['k8scluster-pod-phase/0.137.0','k8scluster-container-ready/0.137.0'].includes(d.stateProfile)&&!d.stockDomain&&!d.histogramBucketMs);else optional(d,'stateProfile',v=>requireThat(v===''));
  const b=d.source;object(b,['resource','exactStream','namespace','services','scope','version','schemaURL','metric','unit','attributes']);const services=['host','kubernetes-pod','kubernetes-container'].includes(b.resource?.kind)&&b.services==null?[]:b.services;strings(services,8,128);requireThat(services.length>0||['host','kubernetes-pod','kubernetes-container'].includes(b.resource?.kind));optional(b,'resource',v=>{object(v,['kind','id']);requireThat((v.kind==='host'&&['gauge','stock'].includes(d.kind)||v.kind==='kubernetes-pod'&&['gauge','state'].includes(d.kind)||v.kind==='kubernetes-container'&&(d.kind==='state'&&d.stateProfile==='k8scluster-container-ready/0.137.0'||d.kind==='gauge'))&&services.length===0&&!b.namespace&&!!b.exactStream);if(v.kind==='kubernetes-container')for(const key of ['k8s.pod.uid','k8s.container.name']){text(b.exactStream.resource?.[key],512);requireThat(b.exactStream.resource[key]!=='')}text(v.id,512);requireThat(v.id!==''&&b.exactStream.resource?.[v.kind==='host'?'host.id':v.kind==='kubernetes-container'?'container.id':'k8s.pod.uid']===v.id)});for(const [key,max]of [['scope',256],['version',128],['schemaURL',1024],['metric',256],['unit',64]])optional(b,key,v=>text(v,max));optional(b,'namespace',namespace);optional(b,'attributes',v=>map(v,16,128,512));
  optional(b,'exactStream',s=>{object(s,['resourceSchemaURL','resource','scope','version','schemaURL','name','unit','attributes']);requireThat(bytes(JSON.stringify(s))<=16384);map(s.resource,64,256,2048);map(['host','kubernetes-pod','kubernetes-container'].includes(b.resource?.kind)&&s.attributes==null?{}:s.attributes,64,256,2048);for(const [key,max]of [['resourceSchemaURL',1024],['scope',256],['version',128],['schemaURL',1024],['name',256],['unit',64]])optional(s,key,v=>text(v,max))});
  optional(d,'lifecycle',l=>{object(l,['logicalService','tenants','accepted','terminal','eventID','eventTime','subject','tenant','release','outcome','success','outcomes','deadlineMs','allowanceMs']);for(const key of ['logicalService','accepted','terminal','eventID','eventTime','subject','tenant','release','outcome','success'])text(l[key],128);strings(l.tenants,16,128);strings(l.outcomes,16,128);integer(l.deadlineMs,1,3600000);integer(l.allowanceMs,0,600000)});
 }
function step(s){
  object(s,['version','kind','scope','view','record','measurement']);object(s.scope,['environment','fromMs','toMs','fromNS','toNS']);text(s.scope.environment,128);requireThat(s.scope.environment!=='');integer(s.scope.fromMs,0,4102444799999);integer(s.scope.toMs,s.scope.fromMs+1,Math.min(s.scope.fromMs+7200000,4102444799999));
  if(s.kind==='record'){record(s.record);requireThat(!s.measurement);text(s.record.environment,128)}else if(s.kind==='measurement'){native(s.measurement);requireThat(!s.record)}else throw Error('Unsupported question step');
  const protocol=QuestionProtocol;protocol.validateStep(s);
 }
function card(c){
  object(c,['kind','review','readBasis','key','evidence','step','role','text','reference','caption']);optional(c,'caption',v=>text(v,160));optional(c,'text',v=>text(v,4000,true));
  if(c.kind==='question'){requireThat(!c.reference&&!c.role&&!c.text&&!c.review&&!c.readBasis);step(c.step);optional(c,'key',v=>requireThat(v===''||ID.test(v)));optional(c,'evidence',v=>{strings(v,16,64);requireThat(v.every(x=>ID.test(x))&&(!v.length||ID.test(c.key)))})}
  else {requireThat(!c.step&&!c.key&&(!c.evidence||Array.isArray(c.evidence)&&c.evidence.length===0));if(c.kind==='note'){requireThat(['hypothesis','finding','conclusion','next-step'].includes(c.role)&&typeof c.text==='string'&&c.text.trim()&&!c.reference&&!c.caption&&!c.review&&!c.readBasis)}else{requireThat(['capture','comparison','deployment','retained'].includes(c.kind)&&ID.test(c.reference)&&!c.role);if(c.kind==='retained'){requireThat(ID.test(c.review)&&!c.text);optional(c,'readBasis',v=>text(v,16384))}else requireThat(!c.review&&!c.readBasis&&(c.kind==='comparison'||!c.text))}}
 }
function validateDraft(d){
  object(d,['parent','document','composer','openedQuestion']);if(d.parent!==null){object(d.parent,['id','recordedAt']);requireThat(ID.test(d.parent.id));text(d.parent.recordedAt,64);requireThat(Number.isFinite(Date.parse(d.parent.recordedAt)))}
  const doc=d.document;object(doc,['title','question','finding','openQuestions','cards']);text(doc.title,120);for(const k of ['question','finding','openQuestions'])text(doc[k],4000,true);requireThat(Array.isArray(doc.cards)&&doc.cards.length<=32);doc.cards.forEach(card);const keys=doc.cards.filter(c=>c.key).map(c=>c.key);requireThat(new Set(keys).size===keys.length);
  const c=d.composer;object(c,['role','text','editing','captureID','caption','destination','comparisonID','comparisonFinding']);requireThat(['hypothesis','finding','conclusion','next-step'].includes(c.role));text(c.text,4000,true);text(c.captureID,64);text(c.caption,160);text(c.comparisonID,64);text(c.comparisonFinding,4000,true);text(c.destination,80);text(d.openedQuestion,80);
  if(c.editing!==null){integer(c.editing,0,doc.cards.length-1);requireThat(doc.cards[c.editing].kind==='note')}
  const destination=x=>{if(!x)return;const target=x.startsWith('legacy:')&&/^legacy:(0|[1-9][0-9]*)$/.test(x)?doc.cards[Number(x.slice(7))]:doc.cards.find(c=>c.key===x);requireThat(target?.kind==='question','Draft question destination is unavailable')};destination(c.destination);destination(d.openedQuestion);return d;
 }
const storageKey=path=>'goo.investigation-draft.v1:'+path;
function randomID(){const crypto=global.crypto||(null);return Array.from(crypto.getRandomValues(new Uint8Array(32)),v=>v.toString(16).padStart(2,'0')).join('')}
class Journal{
  constructor({storage,path,now=()=>Date.now(),id=randomID}){requireThat(typeof path==='string'&&path.startsWith('/')&&bytes(path)<=2048);this.storage=storage;this.path=path;this.now=now;this.id=id;this.observed=undefined}
  inspect(){let raw;try{raw=this.storage.getItem(storageKey(this.path))}catch{return {state:'unavailable',message:'Browser draft storage is unavailable; recovery is not confirmed.'}}this.observed=raw;
   if(raw===null)return {state:'empty',token:raw};try{requireThat(bytes(raw)<=MAX_BYTES,'Local draft exceeds 256 KiB');const e=JSON.parse(raw);object(e,['version','path','id','savedAt','publication','draft']);requireThat(e.version===1&&e.path===this.path&&ID.test(e.id)&&['draft','unknown'].includes(e.publication));const t=Date.parse(e.savedAt);requireThat(typeof e.savedAt==='string'&&Number.isFinite(t)&&t<=this.now()&&this.now()-t<TTL_MS,'Local draft is expired or has an invalid checkpoint time');validateDraft(e.draft);return {state:'ready',entry:clone(e),token:raw}}catch(e){return {state:'invalid',token:raw,message:'Local investigation draft cannot be restored: '+e.message}}}
  checkpoint(draft,{expected=this.observed,publication='draft'}={}){
   validateDraft(draft);let raw;try{raw=this.storage.getItem(storageKey(this.path))}catch{throw Error('Browser draft storage is unavailable; your editor remains in this page.')}
   requireThat(expected!==undefined&&raw===expected,'Local recovery changed; restore or discard it before replacing it');let previous=null;if(raw!==null){const read=this.inspect();requireThat(read.state==='ready','Discard the invalid local draft before replacing it');previous=read.entry}
   const e={version:1,path:this.path,id:this.id(),savedAt:new Date(this.now()).toISOString(),publication:publication==='unknown'||previous?.publication==='unknown'?'unknown':'draft',draft:clone(draft)};requireThat(ID.test(e.id));const next=JSON.stringify(e);requireThat(bytes(next)<=MAX_BYTES,'Local investigation draft exceeds 256 KiB; your editor remains in this page.');try{this.storage.setItem(storageKey(this.path),next)}catch{throw Error('Local draft could not be stored; reload recovery is unavailable.')}this.observed=next;return clone(e);
  }
  beforePublish(draft,options={}){return this.checkpoint(draft,{...options,publication:'unknown'})}
  acknowledge(id){const read=this.inspect();if(read.state!=='ready'||read.entry.id!==id)return false;return this.discard(read.token)}
  discard(token){let raw;try{raw=this.storage.getItem(storageKey(this.path));if(raw!==token)return false;this.storage.removeItem(storageKey(this.path))}catch{throw Error('Local draft could not be discarded from browser storage.')}this.observed=null;return true}
 }

export {Journal,validateDraft,storageKey,MAX_BYTES,TTL_MS};
