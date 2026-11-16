/* One supplied explicit-histogram occurrence. No acquisition or aggregation. */
import {facts} from './raw-histogram-values.mjs';
const keys=['temporality','recordingState','count','countState','bounds','boundsState','bucketCounts','bucketCountsState','bucketsState'];
const equal=(a,b)=>JSON.stringify(a)===JSON.stringify(b);
const unavailable=()=>({state:'unavailable',eligible:false,reason:'Bucket data unavailable for this point.',rows:[],count:null,maximum:'0',unit:''});
export function model(row){
 try{
  if(!row||typeof row!=='object'||!row.rawRow||!row.stream||typeof row.stream.unit!=='string'||typeof row.occurrenceKey!=='string')return unavailable();
  const projected=facts(row.rawRow,{flags:row.flags,flagsState:row.flagsState}).result;
  if(keys.some(key=>!equal(row[key],projected[key])))return unavailable();
  if(row.flagsState==='valid'&&(!Number.isInteger(row.flags)||row.flags<0||row.flags>4294967295||row.flags!==row.rawRow.flags))return unavailable();
  const base={...unavailable(),unit:row.stream.unit,occurrenceKey:row.occurrenceKey,endNS:row.endNS,temporality:row.temporality,recordingState:row.recordingState,bucketsState:row.bucketsState,count:row.count};
  if(row.recordingState==='no-recorded-value')return {...base,reason:'No recorded value. Inspect raw source fields for this point.'};
  if(row.bucketsState!=='valid')return {...base,reason:row.bucketsState==='absent'?'Bucket data unavailable for this point.':'Invalid bucket vector. Inspect diagnostics and raw source fields.'};
  let maximum=0n;for(const count of row.bucketCounts){const amount=BigInt(count);if(amount>maximum)maximum=amount}
  const bound=index=>String(row.bounds[index]),rawBound=index=>String(row.rawRow.bounds[index]);
  const interval=(index,label)=>row.bounds.length===0?'(−∞, +∞)':index===0?'(−∞, '+label(0)+']':index===row.bounds.length?'('+label(index-1)+', +∞)':'('+label(index-1)+', '+label(index)+']';
  const rows=row.bucketCounts.map((count,index)=>({index,interval:interval(index,bound),rawInterval:interval(index,rawBound),count,fraction:maximum===0n?0:Number(BigInt(count))/Number(maximum)}));
  const eligible=row.recordingState==='recorded';
  return {...base,state:eligible?(row.count==='0'?'no_data':'observed'):'unknown',eligible,reason:eligible?'':'Recording validity unknown. Source bucket counts are not an evaluated distribution.',maximum:String(maximum),rows};
 }catch{return unavailable()}
}
