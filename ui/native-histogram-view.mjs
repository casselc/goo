/* Lazy projection of one retained evaluated distribution. No reads or merging. */
import {validateHistogramMeasurementQuestion,validateMeasurementIntervalSelector,binary64Rational,sameNativeStream} from './measurement-scope.mjs';
import {validateSnapshot} from './view-schema.mjs';
const basis=value=>typeof value==='string'&&/^[a-f0-9]{64}$/.test(value);
const count=value=>typeof value==='string'&&/^(0|[1-9]\d{0,19})$/.test(value)&&BigInt(value)<=18446744073709551615n;
function clock(value){
 if(typeof value!=='string')return false;
 const match=/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d{1,9})?(?:Z|[+-]\d{2}:\d{2})$/.exec(value);if(!match)return false;
 const [,year,month,day,hour,minute,second]=match.map(Number),days=[31,year%4===0&&(year%100!==0||year%400===0)?29:28,31,30,31,30,31,31,30,31,30,31];
 return month>=1&&month<=12&&day>=1&&day<=days[month-1]&&hour<24&&minute<60&&second<60&&Number.isFinite(Date.parse(value));
}
function retainedHistogram(result){
 const question=result.question,bounds=validateHistogramMeasurementQuestion(question),histograms=result.histograms;
 if(!basis(result.basis)||!clock(result.readAt)||!histograms||!Array.isArray(histograms.series)||histograms.apiVersion!=='goo/histogram-result/v1alpha1'||!basis(histograms.basis)||histograms.fromNS!==bounds.fromNS||histograms.toNS!==bounds.toNS||histograms.maxGapNS!==String(BigInt(question.definition.maxGapMs)*1000000n))return null;
 return {question,bounds,histograms};
}
function nativeView(result,series,distribution,selector,scopeID,sourceBasis){
 if(!distribution||distribution.unit!==series.stream?.unit||!['s','ms'].includes(distribution.unit)||!['delta','cumulative'].includes(series.temporality))return null;
 validateMeasurementIntervalSelector(selector);
 if(!Array.isArray(distribution.bounds)||distribution.bounds.length>256||distribution.bounds.some((value,index)=>typeof value!=='number'||!Number.isFinite(value)||value<0||index>0&&value<=distribution.bounds[index-1])||!Array.isArray(distribution.bucketCounts)||distribution.bucketCounts.length!==distribution.bounds.length+1||!count(distribution.count)||distribution.bucketCounts.some(value=>!count(value)))return null;
 const total=distribution.bucketCounts.reduce((sum,value)=>sum+BigInt(value),0n);if(total!==BigInt(distribution.count))return null;
 const unit=distribution.unit,edges=['-inf',...distribution.bounds.map(binary64Rational),'+inf'];
 const snapshot={version:'goo.dataset/1',id:scopeID+':bins',revision:result.basis,scope:{id:scopeID,generation:result.basis,signal:'metrics',selector},shape:'histogram',origin:'source-read',sourceBasis,capturedAt:result.readAt,fields:[{key:'from',label:'Bin start (exclusive)',type:'histogram-bound',unit},{key:'to',label:'Bin end (inclusive)',type:'histogram-bound',unit},{key:'count',label:'Instrument observations',type:'number',unit:'observations'}],rows:distribution.bucketCounts.map((value,index)=>({key:'bin:'+index,state:value==='0'?'no_data':'observed',values:{from:edges[index],to:edges[index+1],count:value}})),quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:result.basis,scopeId:scopeID,aspects:['bins']},histogram:{kind:'otel-explicit',binning:'explicit',grain:'observations',temporality:series.temporality,unit,total:distribution.count,closure:'left-open-right-closed'}};
 return validateSnapshot(snapshot).ok?snapshot:null;
}
export function buildNativeHistogramView(result,seriesIndex,intervalIndex){
 try{
  const parent=retainedHistogram(result);if(!parent||!Number.isInteger(seriesIndex)||seriesIndex<0||!Number.isInteger(intervalIndex)||intervalIndex<0)return null;
  const series=parent.histograms.series?.[seriesIndex],interval=series?.intervals?.[intervalIndex];
  if(!series||!interval||interval.state!=='ok'||interval.reason!==undefined&&interval.reason!=='')return null;
  const selector={kind:'measurement-interval',question:structuredClone(parent.question),stream:structuredClone(series.stream),fromNS:interval.fromNS,toNS:interval.toNS};
  return nativeView(result,series,interval.distribution,selector,result.basis+':native:'+seriesIndex+':'+intervalIndex,parent.histograms.basis);
 }catch{return null}
}
export function buildNativeHistogramTimeView(result,seriesIndex,bucketIndex){
 try{
  const parent=retainedHistogram(result),time=result.histogramTime;
  if(!parent||parent.question.definition.histogramBucketMs===undefined||!Number.isInteger(seriesIndex)||seriesIndex<0||!Number.isSafeInteger(bucketIndex)||bucketIndex<0||!time||!Array.isArray(time.series)||time.apiVersion!=='goo/histogram-time-result/v1alpha1'||!basis(time.basis)||time.sourceBasis!==parent.histograms.basis||time.fromNS!==parent.bounds.fromNS||time.toNS!==parent.bounds.toNS)return null;
  const width=BigInt(parent.question.definition.histogramBucketMs)*1000000n;
  if(time.widthNS!==String(width))return null;
  const series=time.series?.[seriesIndex],bucket=series?.buckets?.[bucketIndex],expectedCount=Number((BigInt(time.toNS)-BigInt(time.fromNS)-1n)/width+1n);
  if(!series||!Array.isArray(series.buckets)||series.buckets.length!==expectedCount||!bucket||!Array.isArray(bucket.reasons)||!bucket.distribution||!count(bucket.distribution.count))return null;
  const originals=parent.histograms.series?.filter(source=>sameNativeStream(source.stream,series.stream));
  if(originals?.length!==1||originals[0].temporality!==series.temporality||time.series.filter(candidate=>sameNativeStream(candidate.stream,series.stream)).length!==1)return null;
  const empty=bucket.distribution.count==='0';
  if(empty?(bucket.state!=='no_data'||bucket.reasons.length!==1||bucket.reasons[0]!=='empty-distribution'):(bucket.state!=='observed'||bucket.reasons.length!==0))return null;
  const selector={kind:'measurement-time-bucket',question:structuredClone(parent.question),stream:structuredClone(series.stream),fromNS:bucket.fromNS,toNS:bucket.toNS,widthNS:time.widthNS,bucketIndex};
  return nativeView(result,series,bucket.distribution,selector,result.basis+':native-time:'+seriesIndex+':'+bucketIndex,time.basis);
 }catch{return null}
}
