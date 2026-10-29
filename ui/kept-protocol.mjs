/* Reference-only exact keep protocol. No DOM or network. */
import * as Environment from './record-environment.mjs';
import * as Questions from './question-protocol.mjs';
import * as Session from './question-session.mjs';
const id=v=>typeof v==='string'&&/^[a-f0-9]{64}$/.test(v),key=v=>typeof v==='string'&&/^[a-f0-9]{64}:[1-9][0-9]{0,3}$/.test(v),copy=v=>structuredClone(v);
function intent(body){
 if(body?.version==='goo/keep-records/v1')return recordIntent(body);
 if(!body||Object.keys(body).some(k=>!['version','requestID','connection','selector','refs'].includes(k))||body.version!=='goo/keep-tail/v1'||!id(body.requestID)||!id(body.connection)||!Array.isArray(body.refs)||body.refs.length<1||body.refs.length>20)throw Error('Choose 1–20 observed records from one connection.');
 const s=body.selector;if(!s||Object.keys(s).some(k=>!['signal','environment','environmentSelector','where'].includes(k))||!['logs','spans'].includes(s.signal||'logs'))throw Error('Invalid kept selector.');Environment.selection(s);
 const refs=copy(body.refs);if(refs.some(r=>!r||Object.keys(r).some(k=>!['token','key'].includes(k))||!id(r.token)||!key(r.key))||new Set(refs.map(r=>r.key)).size!==refs.length)throw Error('Invalid observed occurrence references.');refs.sort((a,b)=>a.token<b.token?-1:a.token>b.token?1:a.key<b.key?-1:a.key>b.key?1:0);
 return {version:body.version,requestID:body.requestID,connection:body.connection,selector:{...copy(s),signal:s.signal||'logs',...(s.environmentSelector!=null?{environment:''}:{})},refs};
}
function entry(path,body,createdAt=new Date().toISOString()){if(path!==(body?.version==='goo/keep-records/v1'?'/records':'/tail')||!Number.isFinite(Date.parse(createdAt))||new TextEncoder().encode(JSON.stringify(body)).length>16384)throw Error('Invalid pending keep');return {version:1,path,body:intent(body),createdAt}}
function validate(value,expected){
 if(value?.version==='goo/kept-records/v1')return validateRecords(value,expected);
 if(value?.version!=='goo/kept-observations/v1'||!id(value.id)||!id(value.digest)||!id(value.intentDigest)||!id(value.connection)||!Number.isFinite(Date.parse(value.keptAt))||!Array.isArray(value.members)||value.members.length<1||value.members.length>20)throw Error('Invalid kept observation receipt.');
 const selection=Environment.selection(value.selector),signal=value.selector.signal||'logs',seen=new Set();if(!['logs','spans'].includes(signal))throw Error('Invalid kept signal.');
 for(const m of value.members){const r=m.record;if(!key(m.key)||seen.has(m.key)||!Number.isSafeInteger(m.fromMs)||!Number.isSafeInteger(m.toMs)||m.fromMs<0||m.toMs-m.fromMs!==60000||(!Number.isFinite(Date.parse(m.observedAt))||Date.parse(m.observedAt)<m.toMs)||!id(m.basis)||!r||typeof r.timestamp_ns!=='string'||!/^(0|[1-9][0-9]*)$/.test(r.timestamp_ns)||typeof r[signal==='logs'?'Body':'SpanName']!=='string'||!Environment.matches(selection,r.ResourceAttributes))throw Error('Kept record differs from its scope or provenance.');const ns=BigInt(r.timestamp_ns);if(ns<BigInt(m.fromMs)*1000000n||ns>=BigInt(m.toMs)*1000000n)throw Error('Kept record is outside its observed window.');seen.add(m.key)}
 if(expected){const q=intent(expected),selector=s=>Session.identity({...s,signal:s.signal||'logs'});if(value.id!==q.requestID||value.connection!==q.connection||selector(value.selector)!==selector(q.selector)||JSON.stringify([...seen].sort())!==JSON.stringify(q.refs.map(r=>r.key).sort()))throw Error('Keep acknowledgment differs from the submitted selection.');}
 return value;
}
function recordIntent(body){
 if(!body||Object.keys(body).some(k=>!['version','requestID','resultToken','captureID','indexes','expected'].includes(k))||body.version!=='goo/keep-records/v1'||!id(body.requestID)||!!body.resultToken===!!body.captureID||!id(body.resultToken||body.captureID)||!Array.isArray(body.indexes)||body.indexes.length<1||body.indexes.length>20||body.indexes.some(i=>!Number.isInteger(i)||i<0||i>=10000)||new Set(body.indexes).size!==body.indexes.length)throw Error('Choose 1–20 indexes from one held result or saved capture.');
 if(body.expected){if(Object.keys(body.expected).some(k=>!['question','basis'].includes(k))||!id(body.expected.basis))throw Error('Invalid expected record scope.');Questions.validateRecordQuestion(body.expected.question)}
 return {version:body.version,requestID:body.requestID,...(body.expected?{expected:copy(body.expected)}:{}),...(body.resultToken?{resultToken:body.resultToken}:{captureID:body.captureID}),indexes:[...body.indexes].sort((a,b)=>a-b)};
}
function validateRecords(v,expected){
 if(!id(v.id)||!id(v.intentDigest)||!id(v.digest)||!Number.isFinite(Date.parse(v.keptAt))||!Number.isFinite(Date.parse(v.observedAt))||typeof v.basis!=='string'||!v.basis||v.basis.length>4096||!['observed','no_data','changed_between_reads'].includes(v.state)||!Number.isInteger(v.returnedRecords)||v.returnedRecords<1||v.returnedRecords>10000||typeof v.recordsLimited!=='boolean'||!Array.isArray(v.members)||v.members.length<1||v.members.length>20)throw Error('Invalid kept records receipt.');
 if(v.origin==='saved-capture'){if(!id(v.captureID)||!Number.isFinite(Date.parse(v.captureAt)))throw Error('Invalid original capture.')}else if(v.origin!=='result-page'||v.captureID||v.captureAt)throw Error('Invalid kept records origin.');
 Questions.validateRecordQuestion(v.question);const selection=Environment.selection(v.question);let previous=-1;
 for(const m of v.members){const r=m.record;if(!Number.isInteger(m.index)||m.index<=previous||m.index>=v.returnedRecords||!id(m.digest)||!r||typeof r.timestamp_ns!=='string'||!/^(0|[1-9][0-9]*)$/.test(r.timestamp_ns)||typeof r[v.question.signal==='logs'?'Body':'SpanName']!=='string'||!Environment.matches(selection,r.ResourceAttributes))throw Error('Kept record differs from its scope or index.');const ns=BigInt(r.timestamp_ns);if(ns<BigInt(v.question.fromMs)*1000000n||ns>=BigInt(v.question.toMs)*1000000n)throw Error('Kept record is outside its window.');previous=m.index}
 if(expected){const q=recordIntent(expected);if(q.expected&&(Session.identity(q.expected.question)!==Session.identity(v.question)||q.expected.basis!==v.basis))throw Error('Kept record scope differs from the submitted expectation.');if(q.requestID!==v.id||(q.captureID?'saved-capture':'result-page')!==v.origin||q.captureID&&q.captureID!==v.captureID||JSON.stringify(q.indexes)!==JSON.stringify(v.members.map(m=>m.index)))throw Error('Keep acknowledgment differs from the submitted records.');}
 return v;
}
export {intent,entry,validate};
