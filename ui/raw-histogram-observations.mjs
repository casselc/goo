/* Raw explicit histogram facts and supplied-data detail. No acquisition or evaluation. */
import * as UI from './components.mjs';
import {compactScalar} from './numeric-values.mjs';
import {displayChoice} from './display-choice.mjs';
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
export function detail(host,row,{active=()=>true,bank,snapshot}={}){
 const section=n('section',undefined,'raw-histogram-bucket-detail'),chart=n('goo-chart');
 chart.setAttribute('heading-level','5');chart.snapshot=snapshot??dataset(bank,row);chart.display={kind:'histogram'};
 const picker=displayChoice({legend:'Bucket display',options:[{value:'histogram',label:'Bars'},{value:'table',label:'Table'}],value:'histogram',active,onChange:value=>{chart.display={kind:value}}});
 section.append(picker,chart);host.append(section);
 const stats=n('goo-table'),table=n('table'),head=n('thead'),header=n('tr'),body=n('tbody');stats.className='raw-histogram-source-statistics';stats.setAttribute('density','compact');stats.setAttribute('aria-label','Source histogram statistics');table.setAttribute('aria-label','Source histogram statistics');
 for(const label of ['Field','Source value']){const th=n('th',label);th.scope='col';header.append(th)}head.append(header);
 for(const field of ['sum','min','max']){
  const value=row[field],tr=n('tr'),label=n('th',field),cell=n('td'),raw=value.rawValue,text=raw===null||raw===undefined?'Not reported':typeof raw==='object'?JSON.stringify(raw):String(raw);label.scope='row';cell.dataset.valueState=value.valueState;cell.append(n('span',text===''?'Empty source text':text));
  if(value.valueState!=='finite')cell.append(n('span',' · '+(row.recordingState==='no-recorded-value'?'Not recorded':value.valueState),'evidence-note'));
  tr.append(label,cell);body.append(tr);
 }
 table.append(n('caption','Source columns · producer presence unavailable, including source zero. Not preserved by exporter.'),head,body);stats.append(table);host.append(n('h4','Source statistics'),stats);
}
