/* Raw explicit histogram facts and supplied-data detail. No acquisition or evaluation. */
import * as UI from './components.mjs';
import {compactScalar} from './numeric-values.mjs';
import {scalarValue} from './raw-observation-values.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
const n=UI.node,maximum=18446744073709551615n;
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
export function summary(row){
 if(row.recordingState==='no-recorded-value')return 'Not recorded';
 const count=row.countState==='valid'?compactScalar(row.count)?.text+' count':'Count '+row.countState;
 const buckets=row.bucketsState==='valid'?row.bucketCounts.length+' buckets':row.bucketsState==='absent'?'Buckets absent':'Buckets invalid';
 return count+' · '+buckets+(row.recordingState==='unknown'?' · recording unknown':'');
}
export function intervals(row){if(row.bucketsState!=='valid')return [];return row.bucketCounts.map((count,index)=>({interval:row.bounds.length===0?'(−∞, +∞)':index===0?'(−∞, '+row.bounds[index]+']':index===row.bounds.length?'('+row.bounds[index-1]+', +∞)':'('+row.bounds[index-1]+', '+row.bounds[index]+']',count}));}
export function detail(host,row,{active=()=>true}={}){
 const status=n('p',row.recordingState==='no-recorded-value'?'No recorded value. Raw source fields remain available below.':row.recordingState==='unknown'?'Recording validity is unknown. Source values are not an evaluated distribution.':'One exported point; not an evaluated or merged distribution.','evidence-note');host.append(status);
 const bucketRows=intervals(row),section=n('section'),title=n('h4','Buckets');section.append(title);host.append(section);
 if(!bucketRows.length)section.append(n('p',row.bucketsState==='absent'?'Bucket data unavailable for this point.':'Invalid bucket vector; inspect diagnostics and raw source fields.'));
 else {
  const controls=n('div',undefined,'controls'),previous=n('button','Previous buckets'),next=n('button','More buckets'),count=n('span'),viewport=n('goo-table'),table=n('table'),head=n('thead'),header=n('tr'),body=n('tbody');previous.type=next.type='button';count.setAttribute('role','status');viewport.setAttribute('density','compact');viewport.setAttribute('aria-label','Source histogram buckets');table.setAttribute('aria-label','Source histogram buckets');viewport.setAttribute('mobile-layout','stacked');
  for(const label of ['Interval'+(row.stream.unit?' ('+row.stream.unit+')':''),'Source count']){const th=n('th',label);th.scope='col';header.append(th)}head.append(header);table.append(n('caption','One source point · exact bucket counts'),head,body);viewport.append(table);controls.append(previous,next,count);section.append(controls,viewport);let offset=0;
  function draw(){body.replaceChildren();for(const value of bucketRows.slice(offset,offset+100)){const tr=n('tr'),interval=n('td',value.interval),amount=n('td',value.count);interval.dataset.label=header.children[0].textContent;amount.dataset.label='Source count';tr.append(interval,amount);body.append(tr)}previous.hidden=next.hidden=bucketRows.length<=100;previous.disabled=offset===0;next.disabled=offset+100>=bucketRows.length;count.textContent=offset+1+'–'+Math.min(offset+100,bucketRows.length)+' of '+bucketRows.length+' buckets';}
  previous.onclick=()=>{if(!active()||!nativeControlReady(previous))return;offset-=100;draw()};next.onclick=()=>{if(!active()||!nativeControlReady(next))return;offset+=100;draw()};draw();
 }
 const stats=n('goo-table'),table=n('table'),head=n('thead'),header=n('tr'),body=n('tbody');stats.setAttribute('density','compact');stats.setAttribute('aria-label','Source histogram statistics');table.setAttribute('aria-label','Source histogram statistics');stats.setAttribute('mobile-layout','stacked');for(const label of ['Statistic','Source value','OTLP presence']){const th=n('th',label);th.scope='col';header.append(th)}head.append(header);
 for(const field of ['sum','min','max']){const value=row[field],tr=n('tr');for(const [index,text]of [field,value.valueState==='finite'?String(value.rawValue):value.valueState,'Not preserved by exporter'].entries()){const cell=n(index===0?'th':'td',text);if(index===0)cell.scope='row';cell.dataset.label=header.children[index].textContent;tr.append(cell)}body.append(tr)}
 table.append(n('caption','Source statistics · presence unavailable, including source zero'),head,body);stats.append(table);host.append(n('h4','Source statistics'),stats);
}
