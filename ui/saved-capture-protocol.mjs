import {validateResult as validateSearchResult} from './trace-search-result.mjs';
import {validatePosition} from './trace-search-position.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
/* Checked stored response boundary. No source reads, DOM or persistence. */
import * as Environment from './record-environment.mjs';
import * as Questions from './question-protocol.mjs';
import * as Session from './question-session.mjs';
import * as Explanation from './explanation.mjs';
function recordIdentity(q){const value=structuredClone(q);delete value.offset;delete value.limit;return Session.identity(value)}
function checkedRecord(expected,actual){Questions.validateRecordQuestion(expected);Questions.validateRecordQuestion(actual);validatePredicate(expected.where,expected.signal);validatePredicate(actual.where,actual.signal);if(recordIdentity(expected)!==recordIdentity(actual))throw Error('Stored record scope differs from its immutable receipt.');}
const arms={explore:'explore',inventory:'inventory',comparison:'comparison',measurement:'measurement',explanation:'explanation','trace-search':'traceSearch'};
const object=value=>value!==null&&typeof value==='object'&&!Array.isArray(value),own=(value,key)=>Object.hasOwn(value,key),id=value=>typeof value==='string'&&/^[a-f0-9]{64}$/.test(value);
function clock(value){
 if(typeof value!=='string'||value.length>35)return false;
 const m=/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d{1,9})?(?:Z|[+-]\d{2}:\d{2})$/.exec(value);if(!m)return false;
 const [,y,mo,d,h,mi,se]=m.map(Number),days=[31,y%4===0&&(y%100!==0||y%400===0)?29:28,31,30,31,30,31,31,30,31,30,31];return mo>=1&&mo<=12&&d>=1&&d<=days[mo-1]&&h<24&&mi<60&&se<60&&Number.isFinite(Date.parse(value));
}
function fixedWindow(q){if(!object(q)||!Number.isSafeInteger(q.fromMs)||!Number.isSafeInteger(q.toMs)||q.fromMs<0||q.toMs<=q.fromMs||q.toMs-q.fromMs>7200000||q.toMs>4102444800000)throw Error('Stored capture requires a bounded fixed question.');}
function inventoryQuestion(q){fixedWindow(q);Environment.selection(q);if(Object.keys(q).some(k=>!['environment','environmentSelector','fromMs','toMs','namespace','service'].includes(k)))throw Error('Invalid stored inventory question.');if(own(q,'service')&&(typeof q.service!=='string'||!q.service.isWellFormed()||q.service.includes('\0')||new TextEncoder().encode(q.service).length>128))throw Error('Invalid stored inventory service.');if(q.namespace!=null){const n=q.namespace;if(!object(n)||Object.keys(n).some(k=>!['mode','value'].includes(k))||!['missing','value'].includes(n.mode)||n.value!==undefined&&(typeof n.value!=='string'||!n.value.isWellFormed()||n.value.includes('\0')||new TextEncoder().encode(n.value).length>128)||n.mode==='missing'&&n.value!==undefined&&n.value!=='')throw Error('Invalid stored inventory namespace.');}}
function captureFamily(question){
 if(!object(question)||!own(arms,question.kind))throw Error('Unsupported stored capture family.');const selected=arms[question.kind];
 if(!object(question[selected])||Object.keys(question).some(k=>k!=='kind'&&k!==selected&&!(selected==='traceSearch'&&k==='traceSearchPosition')))throw Error('Stored capture question requires one exclusive non-null family.');
 return selected;
}
function receipt(page,expectedID){
 const r=page?.receipt;if(!object(r)||Object.keys(r).some(k=>!['version','id','parent','title','capturedAt','question'].includes(k))||r.version!==1||!id(r.id)||expectedID!==undefined&&r.id!==expectedID||typeof r.title!=='string'||!r.title.trim()||!r.title.isWellFormed()||new TextEncoder().encode(r.title).length>120||/\p{Cc}/u.test(r.title)||!clock(r.capturedAt)||own(r,'parent')&&r.parent!==''&&!id(r.parent))throw Error('Invalid immutable capture receipt.');
 return captureFamily(r.question);
}
function checkedQuestion(q,key){
 if(key==='explore'){Questions.validateRecordQuestion(q.explore);validatePredicate(q.explore.where,q.explore.signal)}
 else if(key==='inventory')inventoryQuestion(q.inventory);
 else if(key==='comparison'){if(!object(q.comparison)||q.comparison.distribution!=null&&typeof q.comparison.distribution!=='boolean')throw Error('Invalid stored comparison question.');for(const side of ['baseline','selection']){Questions.validateRecordQuestion(q.comparison[side]);validatePredicate(q.comparison[side].where,q.comparison[side].signal)}if(!Environment.sameQuestionEnvironment(q.comparison.baseline,q.comparison.selection))throw Error('Stored comparison environment policies differ.');}
 else if(key==='explanation')Explanation.validate(q.explanation);
 else if(key==='traceSearch'){Questions.validateTraceSearchRequest(q.traceSearch);if(own(q,'traceSearchPosition'))validatePosition(q.traceSearchPosition,q.traceSearch);}
 else Questions.validateStep(Questions.measurementStep(q.measurement));
}
function exclusivePayload(page,key,{reference=false}={}){if(!object(page))throw Error('Invalid stored capture envelope.');for(const arm of Object.values(arms))if(own(page,arm)&&(reference||arm!==key))throw Error('Stored capture has mixed payload families.');if(!reference&&!object(page[key]))throw Error('Stored capture requires its selected result.');}
function validateReceiptReference(page,expectedID){if(!object(page)||Object.keys(page).some(k=>!['receipt','definitionRevision','readAt','summary'].includes(k)))throw Error('Invalid receipt-only capture envelope.');const key=receipt(page,expectedID);exclusivePayload(page,key,{reference:true});checkedQuestion(page.receipt.question,key);if(['retainedRecords','recordsLimited','nextCursor','previousCursor'].some(k=>own(page,k))||(page.definitionRevision!==''&&!id(page.definitionRevision))||typeof page.readAt!=='string'||page.readAt!==''&&!clock(page.readAt)||!object(page.summary)||Object.values(page.summary).some(value=>typeof value!=='string'))throw Error('Invalid receipt-only capture reference.');return page;}
function cursor(page,key,offset){const present=own(page,key);if(offset===null){if(present)throw Error('Unexpected stored page cursor.');return}const value=page[key];if(!present||typeof value!=='string'||value.length>128||!/^[A-Za-z0-9_-]+$/.test(value))throw Error('Stored page cursor unavailable.');let decoded;try{decoded=atob(value.replace(/-/g,'+').replace(/_/g,'/'))}catch{throw Error('Invalid stored page cursor.')}if(decoded!==page.receipt.id+':'+offset||btoa(decoded).replace(/\+/g,'-').replace(/\//g,'_').replace(/=+$/,'')!==value)throw Error('Stored page cursor differs from its capture and offset.');}
function validate(page,expectedID){
 if(!object(page)||Object.keys(page).some(k=>!['receipt',...Object.values(arms),'retainedRecords','recordsLimited','nextCursor','previousCursor'].includes(k)))throw Error('Invalid stored capture envelope.');const key=receipt(page,expectedID);exclusivePayload(page,key);const q=page.receipt.question,result=page[key];checkedQuestion(q,key);
 if(!Number.isInteger(page.retainedRecords)||page.retainedRecords<0||page.retainedRecords>10000||typeof page.recordsLimited!=='boolean')throw Error('Invalid stored capture delivery metadata.');
 if(key==='explore'){
  checkedRecord(q.explore,result.question);const offset=result.question.offset,n=page.retainedRecords,rows=result.rows===null?[]:result.rows;if(q.explore.offset!==0||q.explore.limit!==100||!Number.isInteger(offset)||offset<0||offset>=1000||offset%100!==0||result.question.limit!==100||n>1000||offset>0&&offset>=n||!Array.isArray(rows)||rows.length!==Math.min(100,n-offset)||typeof result.hasMore!=='boolean'||result.hasMore!==(offset+rows.length<n))throw Error('Invalid stored record page.');
  cursor(page,'nextCursor',result.hasMore?offset+rows.length:null);cursor(page,'previousCursor',offset>0?offset-100:null);
 }else{
  cursor(page,'nextCursor',null);cursor(page,'previousCursor',null);
  if(key==='measurement'){if(!object(result.question)||Session.identity(q.measurement)!==Session.identity(result.question)||result.rows!==null&&!Array.isArray(result.rows)||page.retainedRecords!==(result.rows?.length??0)||page.recordsLimited)throw Error('Stored measurement differs from its immutable receipt or bounded input.');}
  else if(key==='traceSearch'){validateSearchResult(result,q.traceSearch,q.traceSearchPosition);if(page.retainedRecords!==result.items.length||page.recordsLimited!==result.hasMore)throw Error('Stored trace-search capture is not a record page.');}
  else {if(page.retainedRecords!==0||page.recordsLimited)throw Error('Stored aggregate capture has no record pages.');
   if(key==='inventory'){inventoryQuestion(result.question);if(Session.identity(q.inventory)!==Session.identity(result.question)||result.rows!==null&&!Array.isArray(result.rows)||(result.rows?.length??0)>1000||!id(result.basis)||!clock(result.capturedAt)||typeof result.count!=='string'||!/^(0|[1-9]\d*)$/.test(result.count)||result.bucketLabels!==null&&!Array.isArray(result.bucketLabels)||result.warnings!==null&&!Array.isArray(result.warnings)||!object(result.plan))throw Error('Stored inventory differs from its immutable receipt.');}
   else if(key==='comparison'){const expected=q.comparison.distribution,actual=result.request?.distribution;if(expected!=null&&typeof expected!=='boolean'||actual!=null&&typeof actual!=='boolean'||(expected===true)!==(actual===true))throw Error('Stored comparison distribution intent differs from its immutable receipt.');for(const side of ['baseline','selection'])checkedRecord(q.comparison[side],result.request?.[side]?.question);}
   else if(key==='explanation'){Explanation.validate(result.request);for(const side of ['baseline','selection'])checkedRecord(q.explanation[side],result.request[side]);if(Session.identity({fields:q.explanation.fields})!==Session.identity({fields:result.request.fields}))throw Error('Stored explanation fields differ from its receipt.');}
  }
 }
 return page;
}
function validateVisitLocation(page,requestedCursor=''){
 if(typeof requestedCursor!=='string')throw Error('Invalid requested capture cursor.');
 if(!page.explore){if(requestedCursor!=='')throw Error('This capture family has no record pages.');return page;}
 const offset=page.explore.question.offset;if(requestedCursor===''){if(offset!==0)throw Error('Stored response differs from the requested first page.');}
 else cursor({receipt:page.receipt,requestedCursor},'requestedCursor',offset);
 return page;
}
function validateVisit(page,expectedID,requestedCursor=''){validate(page,expectedID);return validateVisitLocation(page,requestedCursor)}
function validateRefresh(page,previous){validate(page);if(page.receipt.parent!==previous.receipt.id||Session.identity(page.receipt.question)!==Session.identity(previous.receipt.question))throw Error('Refreshed capture changed its predecessor or immutable question.');return page}
export {captureFamily,checkedRecord,validate,validateRefresh,validateReceiptReference,validateVisit,validateVisitLocation};
