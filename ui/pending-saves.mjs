/* Explicit native-save retry journal. This module performs no network requests. */

const global=globalThis;
const storageKey='goo.pending-native-saves.v1',id=/^[a-f0-9]{64}$/;
function entry(path,body,createdAt=new Date().toISOString()){
  if(!/^\/(?:tail|[a-f0-9]{64}\/annotations)?$/.test(path||'/')||!body||!id.test(body.requestID)||typeof createdAt!=='string'||!Number.isFinite(Date.parse(createdAt)))throw Error('Invalid pending save');
  const raw=JSON.stringify(body);if(new TextEncoder().encode(raw).length>16384)throw Error('Pending save exceeds request bound');
  const value=JSON.parse(raw);if(path==='/tail'){if(value.version!=='goo/tail-evidence/v1alpha1'||!id.test(value.connection)||!value.selector||!Array.isArray(value.refs)||value.refs.length<1||value.refs.length>20||value.refs.some(r=>!id.test(r.token)||!/^([a-f0-9]{64}):([1-9][0-9]{0,3})$/.test(r.key))||new Set(value.refs.map(r=>r.key)).size!==value.refs.length||Object.keys(value).some(k=>!['version','requestID','connection','selector','refs','tags','note'].includes(k)))throw Error('Invalid pending Activity selection')}else if(path&&path!=='/'){if(!value.target||value.target.kind!=='evidence'||!id.test(value.target.id)||path!=='/'+value.target.id+'/annotations')throw Error('Pending review target differs')}else if(!Array.isArray(value.indexes)||value.indexes.length<1||value.indexes.length>20||(!!value.captureID===!!value.resultToken)||!id.test(value.captureID||value.resultToken))throw Error('Invalid pending selection');
  return {version:1,path:path||'/',body:value,createdAt};
 }
function intent(e){const b={...e.body};delete b.awaitToken;return JSON.stringify([e.path,b])}
class Journal{
  constructor(storage,{key=storageKey,validate=entry}={}){this.storage=storage;this.key=key;this.validate=validate}
  list(){let raw;try{raw=this.storage.getItem(this.key)}catch{throw Error('Pending saves cannot be read from browser storage')};if(!raw)return [];if(raw.length>524288)throw Error('Pending save journal exceeds bound');let items;try{items=JSON.parse(raw)}catch{throw Error('Pending save journal is unreadable; no save was sent')};if(!Array.isArray(items)||items.length>16)throw Error('Invalid pending save journal');const keys=new Set();return items.map(e=>{if(e.version!==1)throw Error('Unsupported pending save version');const v=this.validate(e.path,e.body,e.createdAt);if(keys.has(v.body.requestID))throw Error('Duplicate pending request key');keys.add(v.body.requestID);return v})}
  write(items){const raw=JSON.stringify(items);if(raw.length>524288)throw Error('Pending save journal exceeds bound');try{this.storage.setItem(this.key,raw)}catch{throw Error('Pending save could not be preserved; no request was sent')}}
  begin(path,body){const next=this.validate(path,body),items=this.list(),previous=items.find(e=>e.body.requestID===body.requestID);if(previous){if(intent(previous)!==intent(next))throw Error('Pending request key belongs to another intent');return previous}if(items.length>=16)throw Error('Resolve or forget a pending save before adding another');items.push(next);this.write(items);return next}
  complete(key){this.write(this.list().filter(e=>e.body.requestID!==key))}
 }

export {Journal,entry,storageKey};
