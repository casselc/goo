/* Pure raw explicit histogram admission facts; no display or source owner. */
import {scalarValue} from './raw-observation-values.mjs';
const maximum=18446744073709551615n;
const absent=value=>value===null||value===undefined;
function uint64(value){if(absent(value))return {value:null,state:'absent'};if(typeof value!=='string'||value.length>20||!/^(0|[1-9]\d*)$/.test(value)||BigInt(value)>maximum)return {value:null,state:'invalid'};return {value,state:'valid'}}
function bounds(value){if(absent(value))return {value:null,state:'absent'};if(!Array.isArray(value))return {value:null,state:'invalid'};const out=[];for(const entry of value){const parsed=scalarValue(entry);if(parsed.state!=='finite'||out.length&&parsed.value<=out.at(-1))return {value:null,state:'invalid'};out.push(parsed.value)}return {value:out,state:'valid'}}
function counts(value){if(absent(value))return {value:null,state:'absent'};if(!Array.isArray(value))return {value:null,state:'invalid'};const out=[];for(const entry of value){const parsed=uint64(entry);if(parsed.state!=='valid')return {value:null,state:'invalid'};out.push(parsed.value)}return {value:out,state:'valid'}}
function statistic(raw){const value=scalarValue(raw);return {rawValue:raw??null,value:value.value,valueState:value.state,presenceState:'unavailable'}}
export function facts(raw,{flags,flagsState}){
 const diagnostics=[],add=(field,code)=>diagnostics.push({field,code});
 if(Array.isArray(raw.bounds)&&raw.bounds.length>10000||Array.isArray(raw.bucket_counts)&&raw.bucket_counts.length>10001)throw Error('Histogram point exceeds vector budget.');
 const count=uint64(raw.count),bound=bounds(raw.bounds),bucket=counts(raw.bucket_counts);
 let bucketsState='valid',bucketDiagnostic='';
 if(count.state==='invalid'||bound.state==='invalid'||bucket.state==='invalid'){bucketsState='invalid';bucketDiagnostic='invalid'}
 else if(count.state==='absent'||bound.state==='absent'||bucket.state==='absent'){bucketsState='absent';bucketDiagnostic='absent'}
 else if(!bound.value.length&&!bucket.value.length){bucketsState='absent';bucketDiagnostic='absent'}
 else if(bucket.value.length!==bound.value.length+1){bucketsState='invalid';bucketDiagnostic='shape'}
 else if(bucket.value.reduce((sum,value)=>sum+BigInt(value),0n)!==BigInt(count.value)){bucketsState='invalid';bucketDiagnostic='count-mismatch'}
 const recordingState=flagsState==='valid'&&(flags&1)!==0?'no-recorded-value':flagsState==='valid'&&flags===0?'recorded':'unknown';
 const result={temporality:raw.temporality,recordingState,count:count.value,countState:count.state,bounds:bound.value,boundsState:bound.state,bucketCounts:bucket.value,bucketCountsState:bucket.state,bucketsState,sum:statistic(raw.sum),min:statistic(raw.min),max:statistic(raw.max)};
 for(const [field,state]of [['count',count.state],['bounds',bound.state],['bucketCounts',bucket.state]])if(state!=='valid')add(field,state);
 if(bucketDiagnostic)add('buckets',bucketDiagnostic);
 if(recordingState==='no-recorded-value'){
  result.count=null;result.countState='absent';result.bounds=null;result.boundsState='absent';result.bucketCounts=null;result.bucketCountsState='absent';result.bucketsState='absent';add('count','no-recorded-value');add('buckets','no-recorded-value');
 }
 for(const field of ['sum','min','max']){if(result[field].valueState!=='finite')add(field,result[field].valueState);if(recordingState==='no-recorded-value'){result[field].value=null;result[field].valueState='absent';add(field,'no-recorded-value')}add(field,'presence-unavailable')}
 return {result,diagnostics};
}
