import * as NativeRecordContext from './native-record-context.mjs';
import * as NativeExemplarContext from './native-exemplar-context.mjs';
const clone=value=>structuredClone(value);
function freeze(value){if(value&&typeof value==='object'){for(const item of Object.values(value))freeze(item);Object.freeze(value)}return value}
const canonical=value=>JSON.stringify(value,(_key,item)=>item&&typeof item==='object'&&!Array.isArray(item)?Object.fromEntries(Object.keys(item).sort().map(key=>[key,item[key]])):item);
const equal=(a,b)=>canonical(a)===canonical(b);
// Admission is against retained exports, never a synthetic time bucket or a caller's stream.
function admitHeld(held,detail,{reportedExemplars=false}={}){
 const r=held.measurement,kind=r.question.definition?.kind;
 if(!['counter','histogram'].includes(kind)||!detail||!equal(detail.measurement,r.question)||!equal(detail.basis,r.basis)||detail.readAt!==r.readAt)throw Error('Record context requires the held native export interval.');
 if(detail.relation!==undefined){
  if(detail.relation!=='reported-exemplar'||!reportedExemplars||kind!=='histogram')throw Error('Reported exemplar context is unavailable in this held owner.');
  const matches=r.histograms?.series?.filter(item=>equal(item.stream,detail.stream));
  if(matches?.length!==1)throw Error('Reported exemplar context requires one held exact stream.');
  const stream=matches[0].stream,plan=NativeExemplarContext.plan(r,stream,detail.exportEndNS,detail.exemplarIndex);
  const expected={...plan,measurement:r.question,stream,basis:r.basis,readAt:r.readAt};
  if(!equal(detail,expected))throw Error('Reported exemplar context changed the retained export occurrence.');
  return freeze(clone(expected));
 }
 const series=(kind==='counter'?r.counters:r.histograms)?.series;
 const s=series?.find(item=>equal(item.stream,detail.stream));
 if(!s||!Array.isArray(s.intervals)||!s.intervals.some(item=>item.fromNS===detail.fromNS&&item.toNS===detail.toNS))throw Error('Record context requires the held exact stream and export bounds.');
 const plan=NativeRecordContext.plan(r.question,s.stream,detail.question?.signal,detail.fromNS,detail.toNS);
 if(!equal(detail.question,plan.question)||detail.grain!==plan.grain)throw Error('Record context changed the planned record question.');
 return freeze(clone({...plan,measurement:r.question,stream:s.stream,basis:r.basis,readAt:r.readAt}));
}
export {admitHeld};
