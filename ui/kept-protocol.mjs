/* Reference-only exact keep protocol. No DOM or network. */
import * as Environment from './record-environment.mjs';
import * as Session from './question-session.mjs';
const id=v=>typeof v==='string'&&/^[a-f0-9]{64}$/.test(v),key=v=>typeof v==='string'&&/^[a-f0-9]{64}:[1-9][0-9]{0,3}$/.test(v),copy=v=>structuredClone(v);
function intent(body){
 if(!body||Object.keys(body).some(k=>!['version','requestID','connection','selector','refs'].includes(k))||body.version!=='goo/keep-tail/v1'||!id(body.requestID)||!id(body.connection)||!Array.isArray(body.refs)||body.refs.length<1||body.refs.length>20)throw Error('Choose 1–20 observed records from one connection.');
 const s=body.selector;if(!s||Object.keys(s).some(k=>!['signal','environment','environmentSelector','where'].includes(k))||!['logs','spans'].includes(s.signal||'logs'))throw Error('Invalid kept selector.');Environment.selection(s);
 const refs=copy(body.refs);if(refs.some(r=>!r||Object.keys(r).some(k=>!['token','key'].includes(k))||!id(r.token)||!key(r.key))||new Set(refs.map(r=>r.key)).size!==refs.length)throw Error('Invalid observed occurrence references.');refs.sort((a,b)=>a.token<b.token?-1:a.token>b.token?1:a.key<b.key?-1:a.key>b.key?1:0);
 return {version:body.version,requestID:body.requestID,connection:body.connection,selector:{...copy(s),signal:s.signal||'logs',...(s.environmentSelector!=null?{environment:''}:{})},refs};
}
function entry(path,body,createdAt=new Date().toISOString()){if(path!=='/tail'||!Number.isFinite(Date.parse(createdAt)))throw Error('Invalid pending keep');return {version:1,path,body:intent(body),createdAt}}
function validate(value,expected){
 if(value?.version!=='goo/kept-observations/v1'||!id(value.id)||!id(value.digest)||!id(value.intentDigest)||!id(value.connection)||!Number.isFinite(Date.parse(value.keptAt))||!Array.isArray(value.members)||value.members.length<1||value.members.length>20)throw Error('Invalid kept observation receipt.');
 const selection=Environment.selection(value.selector),signal=value.selector.signal||'logs',seen=new Set();if(!['logs','spans'].includes(signal))throw Error('Invalid kept signal.');
 for(const m of value.members){const r=m.record;if(!key(m.key)||seen.has(m.key)||!Number.isSafeInteger(m.fromMs)||!Number.isSafeInteger(m.toMs)||m.fromMs<0||m.toMs-m.fromMs!==60000||(!Number.isFinite(Date.parse(m.observedAt))||Date.parse(m.observedAt)<m.toMs)||!id(m.basis)||!r||typeof r.timestamp_ns!=='string'||!/^(0|[1-9][0-9]*)$/.test(r.timestamp_ns)||typeof r[signal==='logs'?'Body':'SpanName']!=='string'||!Environment.matches(selection,r.ResourceAttributes))throw Error('Kept record differs from its scope or provenance.');const ns=BigInt(r.timestamp_ns);if(ns<BigInt(m.fromMs)*1000000n||ns>=BigInt(m.toMs)*1000000n)throw Error('Kept record is outside its observed window.');seen.add(m.key)}
 if(expected){const q=intent(expected),selector=s=>Session.identity({...s,signal:s.signal||'logs'});if(value.id!==q.requestID||value.connection!==q.connection||selector(value.selector)!==selector(q.selector)||JSON.stringify([...seen].sort())!==JSON.stringify(q.refs.map(r=>r.key).sort()))throw Error('Keep acknowledgment differs from the submitted selection.');}
 return value;
}
export {intent,entry,validate};
