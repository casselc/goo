/* Raw explicit histogram facts and supplied-data detail. No acquisition or evaluation. */
import * as UI from './components.mjs';
import {compactScalar} from './numeric-values.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
export {facts} from './raw-histogram-values.mjs';
const n=UI.node;
export function summary(row){
 if(row.recordingState==='no-recorded-value')return 'Not recorded';
 const count=row.countState==='valid'?compactScalar(row.count)?.text+' count':'Count '+row.countState;
 const buckets=row.bucketsState==='valid'?row.bucketCounts.length+' buckets':row.bucketsState==='absent'?'Buckets absent':'Buckets invalid';
 return count+' · '+buckets+(row.recordingState==='unknown'?' · recording unknown':'');
}
export function intervals(row){if(row.bucketsState!=='valid')return [];return row.bucketCounts.map((count,index)=>({interval:row.bounds.length===0?'(−∞, +∞)':index===0?'(−∞, '+row.bounds[index]+']':index===row.bounds.length?'('+row.bounds[index-1]+', +∞)':'('+row.bounds[index-1]+', '+row.bounds[index]+']',count}));}
export function detail(host,row,{active=()=>true}={}){
 if(row.recordingState!=='recorded')host.append(n('p',row.recordingState==='no-recorded-value'?'No recorded value. Raw fields remain available below.':'Recording validity unknown. Source values are not an evaluated distribution.','evidence-note'));
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
