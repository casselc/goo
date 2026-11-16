/* Raw explicit histogram facts and supplied-data detail. No acquisition or evaluation. */
import * as UI from './components.mjs';
import {compactScalar} from './numeric-values.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
import {dataset} from './raw-histogram-dataset.mjs';
import './scope-view.mjs';
export {facts} from './raw-histogram-values.mjs';
const n=UI.node;
export function summary(row){
 if(row.recordingState==='no-recorded-value')return 'Not recorded';
 const count=row.countState==='valid'?compactScalar(row.count)?.text+' count':'Count '+row.countState;
 const buckets=row.bucketsState==='valid'?row.bucketCounts.length+' buckets':row.bucketsState==='absent'?'Buckets absent':'Buckets invalid';
 return count+' · '+buckets+(row.recordingState==='unknown'?' · recording unknown':'');
}
export function intervals(row){if(row.bucketsState!=='valid')return [];return row.bucketCounts.map((count,index)=>({interval:row.bounds.length===0?'(−∞, +∞)':index===0?'(−∞, '+row.bounds[index]+']':index===row.bounds.length?'('+row.bounds[index-1]+', +∞)':'('+row.bounds[index-1]+', '+row.bounds[index]+']',count}));}
let displayInstance=0;
export function detail(host,row,{active=()=>true,bank}={}){
 const section=n('section',undefined,'raw-histogram-bucket-detail'),picker=n('fieldset',undefined,'raw-observation-display'),chart=n('goo-chart');
 picker.append(n('legend','Bucket display'));const name='raw-histogram-display-'+(++displayInstance);
 chart.setAttribute('heading-level','5');chart.snapshot=dataset(bank,row);chart.display={kind:'histogram'};
 for(const [value,text]of [['histogram','Bars'],['table','Table']]){
  const label=n('label'),input=n('input');input.type='radio';input.name=name;input.value=value;input.checked=value==='histogram';input.setAttribute('aria-label',text);label.append(input,text);picker.append(label);
  input.onchange=()=>{if(!active()||!nativeControlReady(input)||!input.checked)return;chart.display={kind:value}};
 }
 section.append(picker,chart);host.append(section);
 const stats=n('goo-table'),table=n('table'),head=n('thead'),header=n('tr'),body=n('tbody');stats.setAttribute('density','compact');stats.setAttribute('aria-label','Source histogram statistics');table.setAttribute('aria-label','Source histogram statistics');stats.setAttribute('mobile-layout','stacked');for(const label of ['Statistic','Source value','OTLP presence']){const th=n('th',label);th.scope='col';header.append(th)}head.append(header);
 for(const field of ['sum','min','max']){const value=row[field],tr=n('tr');for(const [index,text]of [field,value.valueState==='finite'?String(value.rawValue):value.valueState,'Not preserved by exporter'].entries()){const cell=n(index===0?'th':'td',text);if(index===0)cell.scope='row';cell.dataset.label=header.children[index].textContent;tr.append(cell)}body.append(tr)}
 table.append(n('caption','Source statistics · presence unavailable, including source zero'),head,body);stats.append(table);host.append(n('h4','Source statistics'),stats);
}
