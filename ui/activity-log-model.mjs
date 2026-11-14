/* Pure log-window projection. No stream, source, DOM or publication authority. */
import * as Environment from './record-environment.mjs';
import * as Aggregate from './record-aggregate-protocol.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {identity} from './question-session.mjs';
import {bands,bandPredicate} from './log-severity.mjs';
const copy=value=>structuredClone(value),own=(value,key)=>Object.hasOwn(value,key);
const validCount=value=>typeof value==='string'&&/^(0|[1-9][0-9]*)$/.test(value)&&BigInt(value)<(1n<<64n);
export function query(context){
 const selector=context?.selector,window=context?.readWindow;
 if(!selector||typeof selector!=='object'||Array.isArray(selector)||Object.keys(selector).some(key=>!['signal','environment','environmentSelector','where'].includes(key))||(selector.signal||'logs')!=='logs'||!window||!Number.isSafeInteger(window.fromMs)||!Number.isSafeInteger(window.toMs)||window.fromMs<0||window.toMs-window.fromMs!==60000||window.toMs>4102444800000)throw Error('Wait for an applied log scope and completed 60-second window.');
 const q={signal:'logs',...Environment.questionFields(selector),fromMs:window.fromMs,toMs:window.toMs,limit:1,offset:0,groupBy:[{kind:'column',key:'severity'}],bucketMs:10000,logSeverity:true};
 if(own(selector,'where'))q.where=copy(selector.where);
 validateRecordQuestion(q);if(!Aggregate.canRead(q))throw Error('Unsupported log context question.');return q;
}
export function scopeKey(context){try{const q=query(context);delete q.fromMs;delete q.toMs;return identity(q)}catch{return null}}
function checkedQuestion(q){
 const selector={...Environment.questionFields(q),...(q?.signal!==undefined?{signal:q.signal}:{}),...(q?.where!==undefined?{where:copy(q.where)}:{})};
 const expected=query({selector,readWindow:q});if(identity(expected)!==identity(q))throw Error('Log context question differs from the supported window projection.');return expected;
}
export function admit(q,value){
 checkedQuestion(q);Aggregate.recordAggregates(q,value);
 if(!['observed','no_data'].includes(value.state))throw Error('Log context reads disagree. Refresh explicitly.');
 // SeverityText is an exact nonnullable physical string; empty is not a numeric severity.
 if(value.groups.some(group=>typeof group.g0!=='string'))throw Error('Invalid reported severity text.');return copy(value);
}
function accepted(value){return admit(value?.question,value)}
function countRate(count,widthMs){
 if(!validCount(count)||!Number.isSafeInteger(widthMs)||widthMs<=0)throw Error('Invalid captured-log rate.');
 const numerator=BigInt(count)*1000n,denominator=BigInt(widthMs);let a=numerator,b=denominator;while(b){const remainder=a%b;a=b;b=remainder}const n=numerator/a,d=denominator/a,rounded=(numerator*1000n+denominator/2n)/denominator;
 return {numerator:String(n),denominator:String(d),exact:d===1n?String(n):n+'/'+d,text:rounded/1000n+'.'+String(rounded%1000n).padStart(3,'0')+' logs/s'};
}
function bandInfo(key){if(key==='all')return {key,label:'All logs'};const band=bands.find(item=>item.key===key);if(!band)throw Error('Choose a numeric severity band.');return band}
export function severityBands(value){const snapshot=accepted(value);return bands.map(band=>({...band,count:snapshot.logSeverity[band.key]}))}
export function rate(value,band='all'){const snapshot=accepted(value);bandInfo(band);return countRate(band==='all'?snapshot.count:snapshot.logSeverity[band],snapshot.question.toMs-snapshot.question.fromMs)}
export function timeline(value,band='all'){
 const snapshot=accepted(value),info=bandInfo(band),q=snapshot.question,ceiling=snapshot.timeSeries.reduce((max,point)=>BigInt(point.count)>max?BigInt(point.count):max,1n);
 return {title:(band==='all'?'Captured logs':info.label+' numeric severity')+' · 10-second buckets',unit:'log records',interactive:true,ceiling:String(ceiling),fromLabel:new Date(q.fromMs).toISOString().slice(11,23)+' UTC',toLabel:new Date(q.toMs).toISOString().slice(11,23)+' UTC',points:snapshot.timeSeries.map((point,index)=>{const count=band==='all'?point.count:point.logSeverity[band];return {key:String(index),state:count==='0'?'no_data':'observed',value:count,label:'['+new Date(point.fromMs).toISOString()+', '+new Date(point.toMs).toISOString()+') · '+count+' captured '+(band==='all'?'logs':info.label+' logs')+' · '+countRate(count,point.toMs-point.fromMs).text}})};
}
export function severity(value){
 const snapshot=accepted(value);return snapshot.groups.map(group=>({value:group.g0,label:group.g0===''?'Empty severity text':group.g0,count:group.count,share:{numerator:group.count,denominator:snapshot.count}})).sort((a,b)=>{const n=BigInt(a.count)-BigInt(b.count);return n>0n?-1:n<0n?1:a.value<b.value?-1:a.value>b.value?1:0});
}
export function select(value,fromKey,toKey=fromKey,band='all'){
 const snapshot=accepted(value);if(typeof fromKey!=='string'||typeof toKey!=='string'||!/^[0-5]$/.test(fromKey)||!/^[0-5]$/.test(toKey))throw Error('Choose a time bucket inside the held log window.');
 const a=Number(fromKey),b=Number(toKey),q=copy(snapshot.question);bandInfo(band);delete q.groupBy;delete q.bucketMs;delete q.logSeverity;if(band!=='all'){const predicate=bandPredicate(band);q.where=q.where?{all:[q.where,predicate]}:predicate}q.limit=100;q.offset=0;q.fromMs=snapshot.timeSeries[Math.min(a,b)].fromMs;q.toMs=snapshot.timeSeries[Math.max(a,b)].toMs;validateRecordQuestion(q);return q;
}

export function bandWhere(value,band){if(band==='all')throw Error('Choose a numeric severity band.');return select(value,'0','5',band).where}
