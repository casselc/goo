import {validateRecordQuestion} from './record-intent.mjs';
import {identity} from './question-session.mjs';
export const recordCELContract=Object.freeze({operation:'record.refine',library:'goo.cel.otel/v1',compiler:'goo.record-cel/1',target:'clickhouse',profile:'fixed-record/1'});
const clone=value=>structuredClone(value),hash=value=>typeof value==='string'&&/^[a-f0-9]{64}$/.test(value);
const keys=['version','representation','source','base','binding','parentRevision','loweringDigest'];
const bindingKeys=['operation','signal','baseIdentity','slot','epoch','revision','library','compiler','target','profile'];
function closed(value,allowed){if(!value||typeof value!=='object'||Array.isArray(value)||Object.keys(value).some(key=>!allowed.includes(key)))throw Error('Unexpected CEL authoring fields.');}
function token(value){return typeof value==='string'&&value.isWellFormed()&&value.length>0&&new TextEncoder().encode(value).length<=128&&!/[\p{White_Space}\p{Cc}]/u.test(value);}
// Browser admission checks shape and complete ordinary parent conservation. Go
// alone recompiles source and verifies the content and lowering digests.
export function condition(value,question){
 closed(value,keys);if(value.version!=='goo.cel-authoring/1'||value.representation!=='cel'||typeof value.source!=='string'||!value.source.isWellFormed()||!value.source||new TextEncoder().encode(value.source).length>4096||value.source.includes('\0')||!token(value.parentRevision)||!hash(value.loweringDigest))throw Error('Require checked CEL source provenance.');
 validateRecordQuestion(value.base);validateRecordQuestion(question);closed(value.binding,bindingKeys);
 const b=value.binding;if(bindingKeys.some(key=>!Object.hasOwn(b,key))||b.signal!==question.signal||b.signal!==value.base.signal||!hash(b.baseIdentity)||!['slot','epoch','revision'].every(key=>token(b[key]))||Object.entries(recordCELContract).some(([key,expected])=>b[key]!==expected))throw Error('CEL authoring contract is unavailable.');
 const policy=clone(question);if(value.base.where)policy.where=clone(value.base.where);else delete policy.where;
 if(identity(policy)!==identity(value.base))throw Error('CEL authoring base policy differs from its question.');
 if(value.base.where){if(!Array.isArray(question.where?.all)||question.where.all.length!==2||identity({where:question.where.all[0]})!==identity({where:value.base.where}))throw Error('CEL authoring changed its complete parent.');}else if(!question.where)throw Error('CEL authoring requires an additional condition.');
 return clone(value);
}
export function authoring(value,queries){
 if(value===undefined)return undefined;closed(value,['logs','spans']);if(!Object.keys(value).length)throw Error('Omit empty CEL authoring.');
 const result={};for(const [signal,entry]of Object.entries(value)){const q=queries.find(q=>q.signal===signal);if(!q)throw Error('CEL authoring has no matching record query.');result[signal]=condition(entry,q);}return result;
}
export function capture(snapshot,receipt){
 if(!snapshot||!receipt?.loweringDigest)return null;
 return condition({version:'goo.cel-authoring/1',representation:'cel',source:receipt.expression,base:snapshot.base,binding:receipt.binding,parentRevision:snapshot.parentRevision,loweringDigest:receipt.loweringDigest},receipt.question);
}
