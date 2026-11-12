/* Configuration only: no telemetry, source authority or retained results. */
import * as Protocol from './question-protocol.mjs';
import * as Session from './question-session.mjs';
import * as Environment from './record-environment.mjs';
import * as Columns from './record-columns.mjs';
const day=86400000,maximum=20;
const closed=(v,keys)=>v&&typeof v==='object'&&!Array.isArray(v)&&Object.keys(v).every(k=>keys.includes(k));
function fields(q){
 const check=(v,keys)=>{if(!closed(v,keys))throw Error('Unsupported scope fields.');};
 const field=v=>check(v,['kind','key']);
 const numeric=v=>{check(v,['field','type','unit','parse','scale']);field(v.field);};
 const pred=v=>{if(v.all||v.any){check(v,[v.all?'all':'any']);(v.all||v.any).forEach(pred);}else{check(v,v.numeric?['numeric','op','value']:['field','op','value']);if(v.numeric)numeric(v.numeric);else field(v.field);}};
 if(q.environmentSelector)Environment.normalize(q.environmentSelector);
 if(q.where)pred(q.where);if(q.groupBy)q.groupBy.forEach(field);if(q.numeric)numeric(q.numeric);if(q.numericHistogram)check(q.numericHistogram,['binning']);
 if(q.reportedEdge){check(q.reportedEdge,['parent','child','discoveryNamespace']);for(const side of [q.reportedEdge.parent,q.reportedEdge.child]){check(side,['service','namespace']);check(side.namespace,['mode','value']);}if(q.reportedEdge.discoveryNamespace)check(q.reportedEdge.discoveryNamespace,['mode','value']);}
}
export function normalizeIntent(raw){
 if(new TextEncoder().encode(JSON.stringify(raw)).length>16384)throw Error('Scope configuration exceeds 16 KiB.');
 if(!closed(raw,raw?.kind==='question'?['kind','question','view']:['kind','selector','columns','fieldColumns','baselineCapture']))throw Error('Invalid scope configuration.');
 if(raw.kind==='question'){Protocol.validateRecordQuestion(raw.question,raw.view||'Records');fields(raw.question);const question=structuredClone(raw.question);question.offset=0;return {kind:'question',question,view:raw.view||'Records'};}
 if(raw.kind!=='activity'||!closed(raw.selector,['signal','environment','environmentSelector','where'])||!['logs','spans'].includes(raw.selector.signal))throw Error('Invalid Activity scope.');
 Protocol.validateRecordQuestion({...raw.selector,fromMs:0,toMs:60000});fields(raw.selector);for(const field of raw.fieldColumns||[])if(!closed(field,['kind','key']))throw Error('Invalid field columns.');
 const layout=Columns.validate({columns:raw.columns,fieldColumns:raw.fieldColumns||[]},['utc','service','summary']);
 if(raw.baselineCapture!==undefined&&(!/^[a-f0-9]{64}$/.test(raw.baselineCapture)||raw.selector.signal!=='spans'))throw Error('Invalid baseline reference.');
 return {kind:'activity',selector:structuredClone(raw.selector),...layout,...(raw.baselineCapture?{baselineCapture:raw.baselineCapture}:{})};
}
export function identity(raw){const v=normalizeIntent(raw);return v.kind==='question'?v.kind+Session.identity(v.question)+v.view:v.kind+Session.identity({...v.selector,fromMs:0,toMs:60000})+JSON.stringify([v.columns,v.fieldColumns,v.baselineCapture||'']);}
export function describe(raw){const v=normalizeIntent(raw),q=v.question||v.selector;return (v.kind==='activity'?'Activity':'Fixed question')+' · '+q.signal+' · '+Environment.questionLabel(q);}
export class Journal{
 constructor({storage=()=>globalThis.localStorage,path=globalThis.location?.pathname||'/',now=()=>Date.now()}={}){this.storage=storage;this.now=now;this.key='goo.scope-history.v1:'+path;this.items=[];this.durable=false;this.error='';try{const raw=storage().getItem(this.key);if(raw){if(new TextEncoder().encode(raw).length>262144)throw Error('Scope history exceeds bound.');const v=JSON.parse(raw);if(!closed(v,['version','items'])||v.version!==1||!Array.isArray(v.items)||v.items.length>40)throw Error('Invalid scope history.');this.items=v.items.map(x=>this.entry(x)).filter(x=>x.pinned||this.fresh(x));this.durable=true;this.bound();}}catch(e){this.error='Browser history unavailable · '+e.message;}}
 entry(x){if(!closed(x,['intent','usedAt','pinned'])||!Number.isSafeInteger(x.usedAt)||x.usedAt<0||x.usedAt>this.now()||typeof x.pinned!=='boolean')throw Error('Invalid history entry.');return {intent:normalizeIntent(x.intent),usedAt:x.usedAt,pinned:x.pinned};}
 fresh(x){return this.now()-x.usedAt<=30*day;}
 bound(){const seen=new Set();this.items=this.items.filter(x=>x.pinned||this.fresh(x)).sort((a,b)=>b.usedAt-a.usedAt).filter(x=>{const key=identity(x.intent);if(seen.has(key))return false;seen.add(key);return true;});this.items=[...this.items.filter(x=>x.pinned).slice(0,maximum),...this.items.filter(x=>!x.pinned).slice(0,maximum)];}
 save(){this.bound();if(this.durable){try{const raw=JSON.stringify({version:1,items:this.items});if(new TextEncoder().encode(raw).length>262144)throw Error('Scope history exceeds bound.');this.storage().setItem(this.key,raw);this.error='';}catch(e){this.error='Browser history not saved · '+e.message;throw Error(this.error);}}}
 list(){this.bound();return structuredClone(this.items);}
 remember(raw){const intent=normalizeIntent(raw),key=identity(intent),old=this.items.find(x=>identity(x.intent)===key);this.items=this.items.filter(x=>identity(x.intent)!==key);this.items.unshift({intent,usedAt:this.now(),pinned:old?.pinned||false});this.save();}
 pin(raw){const intent=normalizeIntent(raw),key=identity(intent);this.bound();const old=this.items.find(x=>identity(x.intent)===key);if(!old?.pinned&&this.items.filter(x=>x.pinned).length>=maximum)throw Error('Unpin a scope before pinning another (20 maximum).');this.items=this.items.filter(x=>identity(x.intent)!==key);this.items.unshift({intent,usedAt:this.now(),pinned:true});this.save();}
 unpin(raw){const x=this.items.find(x=>identity(x.intent)===identity(raw));if(x)x.pinned=false;this.save();}
 forget(raw){this.items=this.items.filter(x=>identity(x.intent)!==identity(raw));this.save();}
 clearRecent(){this.items=this.items.filter(x=>x.pinned);this.save();}
 persistence(enabled){if(enabled){this.durable=true;try{this.save();}catch(e){this.durable=false;throw e;}}else{this.storage().removeItem(this.key);this.durable=false;this.error='';}}
 reset(){this.storage().removeItem(this.key);this.items=[];this.durable=false;this.error='';}
}
