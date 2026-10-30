/* Lazy projection of one retained evaluated interval. No source reads or merging. */
import {validateHistogramMeasurementQuestion,validateMeasurementIntervalSelector,binary64Rational} from './measurement-scope.mjs';
import {validateSnapshot} from './view-schema.mjs';
const basis=value=>typeof value==='string'&&/^[a-f0-9]{64}$/.test(value);
const count=value=>typeof value==='string'&&/^(0|[1-9]\d{0,19})$/.test(value)&&BigInt(value)<=18446744073709551615n;
function clock(value){
 if(typeof value!=='string')return false;
 const match=/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d{1,9})?(?:Z|[+-]\d{2}:\d{2})$/.exec(value);if(!match)return false;
 const [,year,month,day,hour,minute,second]=match.map(Number),days=[31,year%4===0&&(year%100!==0||year%400===0)?29:28,31,30,31,30,31,31,30,31,30,31];
 return month>=1&&month<=12&&day>=1&&day<=days[month-1]&&hour<24&&minute<60&&second<60&&Number.isFinite(Date.parse(value));
}
export function buildNativeHistogramView(result,seriesIndex,intervalIndex){
 try{
  const question=result.question,bounds=validateHistogramMeasurementQuestion(question),histograms=result.histograms;
  if(!basis(result.basis)||!clock(result.readAt)||!histograms||histograms.apiVersion!=='goo/histogram-result/v1alpha1'||!basis(histograms.basis)||histograms.fromNS!==bounds.fromNS||histograms.toNS!==bounds.toNS||histograms.maxGapNS!==String(BigInt(question.definition.maxGapMs)*1000000n)||!Number.isInteger(seriesIndex)||seriesIndex<0||!Number.isInteger(intervalIndex)||intervalIndex<0)return null;
  const series=histograms.series?.[seriesIndex],interval=series?.intervals?.[intervalIndex],distribution=interval?.distribution;
  if(!series||!['delta','cumulative'].includes(series.temporality)||!interval||interval.state!=='ok'||interval.reason!==undefined&&interval.reason!==''||!distribution||distribution.unit!==series.stream?.unit||!['s','ms'].includes(distribution.unit))return null;
  const selector={kind:'measurement-interval',question:structuredClone(question),stream:structuredClone(series.stream),fromNS:interval.fromNS,toNS:interval.toNS};validateMeasurementIntervalSelector(selector);
  if(!Array.isArray(distribution.bounds)||distribution.bounds.length>256||distribution.bounds.some((value,index)=>typeof value!=='number'||!Number.isFinite(value)||value<0||index>0&&value<=distribution.bounds[index-1])||!Array.isArray(distribution.bucketCounts)||distribution.bucketCounts.length!==distribution.bounds.length+1||!count(distribution.count)||distribution.bucketCounts.some(value=>!count(value)))return null;
  const total=distribution.bucketCounts.reduce((sum,value)=>sum+BigInt(value),0n);if(total!==BigInt(distribution.count))return null;
  const scopeID=result.basis+':native:'+seriesIndex+':'+intervalIndex,unit=distribution.unit,edges=['-inf',...distribution.bounds.map(binary64Rational),'+inf'];
  const snapshot={version:'goo.dataset/1',id:scopeID+':bins',revision:result.basis,scope:{id:scopeID,generation:result.basis,signal:'metrics',selector},shape:'histogram',origin:'source-read',sourceBasis:histograms.basis,capturedAt:result.readAt,fields:[{key:'from',label:'Bin start (exclusive)',type:'histogram-bound',unit},{key:'to',label:'Bin end (inclusive)',type:'histogram-bound',unit},{key:'count',label:'Instrument observations',type:'number',unit:'observations'}],rows:distribution.bucketCounts.map((value,index)=>({key:'bin:'+index,state:value==='0'?'no_data':'observed',values:{from:edges[index],to:edges[index+1],count:value}})),quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:result.basis,scopeId:scopeID,aspects:['bins']},histogram:{kind:'otel-explicit',binning:'explicit',grain:'observations',temporality:series.temporality,unit,total:distribution.count,closure:'left-open-right-closed'}};
  return validateSnapshot(snapshot).ok?snapshot:null;
 }catch{return null}
}
