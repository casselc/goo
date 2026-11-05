/* Bounded kept discovery. Exact resource opening retains the separate verifier. */
import * as Environment from './record-environment.mjs';
const id=v=>typeof v==='string'&&/^[a-f0-9]{64}$/.test(v);
const utf8=value=>typeof value==='string'&&!/[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?<![\uD800-\uDBFF])[\uDC00-\uDFFF]/u.test(value);
function query(value={}){if(!value||Object.keys(value).some(k=>!['origin','signal','text'].includes(k))||!['','tail-selection','result-page','saved-capture'].includes(value.origin||'')||!['','logs','spans'].includes(value.signal||'')||value.text!=null&&(!utf8(value.text)||new TextEncoder().encode(value.text).length>256))throw Error('Invalid kept discovery filter.');return {origin:value.origin||'',signal:value.signal||'',text:value.text||''}}
function summary(v){
 if(!v||Object.keys(v).some(k=>!['id','version','origin','signal','environment','environmentSelector','observations','keptAt','digest','match'].includes(k))||!id(v.id)||!id(v.digest)||!Number.isInteger(v.observations)||v.observations<1||v.observations>20||!Number.isFinite(Date.parse(v.keptAt))||!['logs','spans'].includes(v.signal)||(v.environment!=null)===(v.environmentSelector!=null))throw Error('Invalid kept discovery metadata.');
 if(v.version==='goo/kept-observations/v1'?v.origin!=='tail-selection':v.version!=='goo/kept-records/v1'||!['result-page','saved-capture'].includes(v.origin))throw Error('Invalid kept discovery origin.');Environment.selection(v);return v;
}
function page(v,expected={}){
 const q=query(expected),integer=(x,max)=>Number.isInteger(x)&&x>=0&&x<=max;
 if(v?.version!=='goo/kept-search/v1'||JSON.stringify(query(v.query))!==JSON.stringify(q)||!integer(v.total,1000)||!integer(v.eligible,v.total)||!integer(v.matched,v.total)||!integer(v.checked,q.text?64:1000)||!integer(v.bytes,32*1024*1024)||!integer(v.matchingRecords,400)||!Array.isArray(v.items)||v.items.length>20||v.items.length>v.matched||typeof v.hasMore!=='boolean'||(v.hasMore?typeof v.nextCursor!=='string'||!v.nextCursor||v.nextCursor.length>1024:!!v.nextCursor))throw Error('Invalid kept discovery page.');
 const ids=new Set();let records=0;
 for(const item of v.items){summary(item);if(ids.has(item.id)||q.origin&&item.origin!==q.origin||q.signal&&item.signal!==q.signal)throw Error('Kept discovery differs from its filter.');ids.add(item.id);
  if(q.text){const m=item.match,w=m?.witness;if(!m||Object.keys(m).some(k=>!['records','witness'].includes(k))||!integer(m.records,item.observations)||m.records<1||!w||Object.keys(w).some(k=>!['memberIndex','field','text'].includes(k))||!integer(w.memberIndex,item.observations-1)||![item.signal==='logs'?'Body':'SpanName','ServiceName','TraceId','SpanId'].includes(w.field)||!utf8(w.text)||Array.from(w.text).length>320||!w.text.includes(q.text))throw Error('Invalid kept content witness.');records+=m.records;
  }else if(item.match!=null)throw Error('Unexpected kept content witness.');
 }
 if(!q.text&&(v.eligible!==v.matched||v.checked!==v.eligible||v.bytes!==0||v.matchingRecords!==0))throw Error('Invalid kept metadata counts.');
 if(q.text&&(v.items.length!==v.matched||v.matched>v.checked||v.checked>v.eligible||v.hasMore&&v.checked===0||v.checked>0&&v.bytes===0||records!==v.matchingRecords))throw Error('Invalid kept content counts.');
 return v;
}
function repair(v){if(!v||!Number.isInteger(v.indexed)||v.indexed<0||v.indexed>64||!Number.isInteger(v.bytes)||v.bytes<0||v.bytes>32*1024*1024||!Number.isInteger(v.completed)||v.completed<v.indexed||!Number.isInteger(v.total)||v.total<v.completed||v.total>1000||typeof v.hasMore!=='boolean'||(v.hasMore?!v.indexed||v.completed>=v.total||typeof v.nextCursor!=='string'||!v.nextCursor||v.nextCursor.length>1024:v.completed!==v.total||!!v.nextCursor))throw Error('Invalid kept index repair progress.');return v}
export {query,summary,page,repair};
