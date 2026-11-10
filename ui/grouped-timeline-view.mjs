/* Passive grouped count geometry. Source validity and continuation authority are host-owned. */
import './time-bars.mjs';
import {categoryLabel} from './view-values.mjs';
const n=(tag,text)=>{const e=document.createElement(tag);if(text!==undefined)e.textContent=String(text);return e};
const utc=value=>new Date(Number(value)).toISOString();
const known=row=>row.state!=='unknown'&&row.values.count!==null;
const label=(snapshot,value)=>categoryLabel(snapshot.fields.find(f=>f.key==='category'),value);
const rowLabel=(snapshot,row)=>label(snapshot,row.values.category)+' · ['+utc(row.values.fromMs)+', '+utc(row.values.toMs)+') · '+(known(row)?row.values.count+' captured records'+(row.values.count==='0'?' · recorded zero':''):'unknown count');
export function groupRows(snapshot,display={}){
 const map=new Map();for(const row of snapshot.rows){const key=JSON.stringify(row.values.category);if(!map.has(key))map.set(key,{key,value:row.values.category,label:label(snapshot,row.values.category),rows:[]});map.get(key).rows.push(row)}
 const result=[...map.values()];for(const group of result)group.rows.sort((a,b)=>Number(a.values.fromMs)-Number(b.values.fromMs));
 if(display.order==='label')result.sort((a,b)=>a.label.localeCompare(b.label)||a.key.localeCompare(b.key));
 const query=(display.query||'').toLocaleLowerCase();return result.filter(group=>group.label.toLocaleLowerCase().includes(query));
}
export const groups=groupRows;
export function ceiling(snapshot){return snapshot.rows.reduce((max,row)=>known(row)&&BigInt(row.values.count)>BigInt(max)?row.values.count:max,'0')}
export function render(snapshot,display,{onSelect}={}){
 const figure=n('figure');figure.className='grouped-timeline-view';const visible=groupRows(snapshot,display),all=groupRows(snapshot),max=ceiling(snapshot);
 figure.append(n('figcaption',visible.length+' of '+all.length+' supplied groups · '+(snapshot.rows.some(known)?'common scale 0–'+max+' captured records':'no known counts; scale unavailable')));
 const select=row=>{if(known(row))onSelect?.(row.key)};
 if(!visible.length){figure.append(n('p',all.length?'No groups match Find. The supplied population is unchanged.':'No captured groups in this read.'));return figure}
 if(display.kind==='small-multiples'){
  for(const group of visible){const bars=n('goo-time-bars');bars.setAttribute('density','compact');bars.data={title:group.label,unit:'captured records',ceiling:max,interactive:!!onSelect,fromLabel:utc(group.rows[0].values.fromMs),toLabel:utc(group.rows.at(-1).values.toMs),points:group.rows.map(row=>({key:row.key,label:rowLabel(snapshot,row),state:known(row)?row.values.count==='0'?'no_data':'observed':'unknown',...(known(row)?{value:row.values.count}:{})}))};bars.addEventListener('goo-time-select',event=>{const row=group.rows.find(r=>r.key===event.detail?.key);if(row)select(row)});figure.append(bars)}
  figure.append(n('p','0 = recorded zero · hatched interval = unknown count'));return figure;
 }
 const scroll=n('goo-table'),table=n('table'),head=n('thead'),body=n('tbody'),header=n('tr');scroll.setAttribute('density','compact');scroll.setAttribute('role','region');scroll.tabIndex=0;scroll.setAttribute('aria-label','Grouped count intervals');table.append(n('caption','Supplied captured-record counts · UTC intervals are start-inclusive, end-exclusive'));
 const cellButton=row=>{const button=n('button',known(row)?row.values.count:'?');button.type='button';button.setAttribute('aria-label',rowLabel(snapshot,row));button.title=rowLabel(snapshot,row);button.dataset.rowKey=row.key;button.dataset.state=row.state;button.disabled=!known(row)||!onSelect;button.addEventListener('click',()=>select(row));return button};
 if(display.kind==='heatmap'){
  const intervals=all[0]?.rows||[];if(intervals.length>32){figure.append(n('p','Heatmap unavailable above 32 intervals. Choose small multiples or Table.'));return figure}
  table.children[0].textContent='Captured records · '+[...new Set(intervals.map(row=>utc(row.values.fromMs).slice(0,10)))].join(' / ')+' UTC · [start, end)';const corner=n('th','Group');corner.scope='col';header.append(corner);for(const row of intervals){const th=n('th',utc(row.values.fromMs).slice(11));th.scope='col';th.title='['+utc(row.values.fromMs)+', '+utc(row.values.toMs)+')';th.setAttribute('aria-label',th.title);header.append(th)}
  for(const group of visible){const tr=n('tr'),th=n('th',group.label);th.scope='row';tr.append(th);for(const row of group.rows){const td=n('td'),button=cellButton(row);if(known(row)&&BigInt(row.values.count)>0n){const intensity=Number(BigInt(row.values.count)*10000n/(BigInt(max)||1n))/10000;button.style.background='rgba(57,201,173,'+(.12+.33*intensity)+')'}td.append(button);tr.append(td)}body.append(tr)}
 }else{
  for(const text of ['Group','UTC interval [start, end)','Captured records']){const th=n('th',text);th.scope='col';header.append(th)}
  for(const group of visible)for(const row of group.rows){const tr=n('tr'),th=n('th',group.label),interval=n('td','['+utc(row.values.fromMs)+', '+utc(row.values.toMs)+')'),count=n('td');th.scope='row';count.append(cellButton(row));tr.append(th,interval,count);body.append(tr)}
 }
 head.append(header);table.append(head,body);scroll.append(table);figure.append(scroll,n('p','0 = recorded zero · ? = unknown count'));return figure;
}
