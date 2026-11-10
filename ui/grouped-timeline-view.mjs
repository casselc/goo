/* Passive grouped geometry. Source validity and continuation authority are host-owned. */
import './time-bars.mjs';
import {categoryLabel,formatCell,exactText} from './view-values.mjs';
const n=(tag,text)=>{const e=document.createElement(tag);if(text!==undefined)e.textContent=String(text);return e};
const utc=value=>new Date(Number(value)).toISOString();
const known=row=>row.state!=='unknown'&&row.values.count!==null;
const meanDisplay=display=>display?.value==='duration.mean';
const meanKnown=row=>known(row)&&row.cells?.['duration.mean']?.state==='derived'&&row.values['duration.mean']!==null;
const parts=value=>{const [n,d='1']=value.split('/');return {n:BigInt(n),d:BigInt(d)}};
const greater=(a,b)=>{const x=parts(a),y=parts(b);return x.n*y.d>y.n*x.d};
const ms=value=>{let {n,d}=parts(value);d*=1000000n;let a=n,b=d;while(b){const next=a%b;a=b;b=next}n/=a;d/=a;return d===1n?String(n):n+'/'+d};
const label=(snapshot,value)=>categoryLabel(snapshot.fields.find(f=>f.key==='category'),value);
const meanField=snapshot=>snapshot.fields.find(f=>f.key==='duration.mean');
const valueLabel=(snapshot,row,display)=>meanDisplay(display)?meanKnown(row)?formatCell(meanField(snapshot),row.values['duration.mean'],row.cells['duration.mean']):known(row)?'No captured spans':'Unknown mean duration':known(row)?row.values.count:'?';
const rowLabel=(snapshot,row,display)=>label(snapshot,row.values.category)+' · ['+utc(row.values.fromMs)+', '+utc(row.values.toMs)+') · '+(meanDisplay(display)?valueLabel(snapshot,row,display)+' · '+(meanKnown(row)?exactText(meanField(snapshot),row.values['duration.mean'],row.cells['duration.mean'])+' · ':''):'')+(known(row)?row.values.count+' captured records'+(row.values.count==='0'?' · recorded zero':''):'unknown count');
export function groupRows(snapshot,display={}){
 const map=new Map();for(const row of snapshot.rows){const key=JSON.stringify(row.values.category);if(!map.has(key))map.set(key,{key,value:row.values.category,label:label(snapshot,row.values.category),rows:[]});map.get(key).rows.push(row)}
 const result=[...map.values()];for(const group of result)group.rows.sort((a,b)=>Number(a.values.fromMs)-Number(b.values.fromMs));
 if(display.order==='label')result.sort((a,b)=>a.label.localeCompare(b.label)||a.key.localeCompare(b.key));
 const query=(display.query||'').toLocaleLowerCase();return result.filter(group=>group.label.toLocaleLowerCase().includes(query));
}
export const groups=groupRows;
export function ceiling(snapshot,display={}){return snapshot.rows.reduce((max,row)=>meanDisplay(display)?meanKnown(row)&&greater(row.values['duration.mean'],max)?row.values['duration.mean']:max:known(row)&&BigInt(row.values.count)>BigInt(max)?row.values.count:max,'0')}
export function render(snapshot,display,{onSelect}={}){
 const figure=n('figure');figure.className='grouped-timeline-view';const visible=groupRows(snapshot,display),all=groupRows(snapshot),max=ceiling(snapshot,display),mean=meanDisplay(display),maxLabel=mean?formatCell(meanField(snapshot),max,{state:'derived'}):max+' captured records',anyKnown=snapshot.rows.some(mean?meanKnown:known);
 figure.append(n('figcaption',visible.length+' of '+all.length+' supplied groups · '+(anyKnown?'common scale 0–'+maxLabel:mean?'no known means; scale unavailable':'no known counts; scale unavailable')));
 const select=row=>{if(known(row))onSelect?.(row.key)};
 if(!visible.length){figure.append(n('p',all.length?'No groups match Find. The supplied population is unchanged.':'No captured groups in this read.'));return figure}
 const legend=mean?'0 ms = observed zero duration · — = no captured spans · ? = unknown mean':'0 = recorded zero · ? = unknown count';
 if(display.kind==='small-multiples'){
  for(const group of visible){const bars=n('goo-time-bars');bars.setAttribute('density','compact');bars.data={title:group.label,unit:mean?'ms':'captured records',ceiling:mean?ms(max):max,...(mean?{valueFormat:'rational',ceilingLabel:maxLabel}:{}),interactive:!!onSelect,fromLabel:utc(group.rows[0].values.fromMs),toLabel:utc(group.rows.at(-1).values.toMs),points:group.rows.map(row=>({key:row.key,label:rowLabel(snapshot,row,display),state:mean?meanKnown(row)?'observed':known(row)?'unavailable':'unknown':known(row)?row.values.count==='0'?'no_data':'observed':'unknown',...(mean?meanKnown(row)?{value:ms(row.values['duration.mean'])}:{}:known(row)?{value:row.values.count}:{})}))};bars.addEventListener('goo-time-select',event=>{const row=group.rows.find(r=>r.key===event.detail?.key);if(row)select(row)});figure.append(bars)}
  figure.append(n('p',mean?'0 ms = observed zero duration · hatched = no mean (empty or unknown); inspect interval details': '0 = recorded zero · hatched interval = unknown count'));return figure;
 }
 const scroll=n('goo-table'),table=n('table'),head=n('thead'),body=n('tbody'),header=n('tr');scroll.setAttribute('density','compact');scroll.setAttribute('sticky-first-column','');scroll.setAttribute('sticky-header','');scroll.setAttribute('role','region');scroll.tabIndex=0;scroll.setAttribute('aria-label',mean?'Grouped mean duration intervals':'Grouped count intervals');table.append(n('caption',mean?'Supplied mean span duration · UTC intervals are start-inclusive, end-exclusive':'Supplied captured-record counts · UTC intervals are start-inclusive, end-exclusive'));
 const cellButton=row=>{const button=n('button',mean?(meanKnown(row)?valueLabel(snapshot,row,display):known(row)?'—':'?'):known(row)?row.values.count:'?');button.type='button';button.setAttribute('aria-label',rowLabel(snapshot,row,display));button.title=rowLabel(snapshot,row,display);button.dataset.rowKey=row.key;button.dataset.state=row.state;button.disabled=!known(row)||!onSelect;button.addEventListener('click',()=>select(row));return button};
 if(display.kind==='heatmap'){
  const intervals=all[0]?.rows||[];if(intervals.length>32){figure.append(n('p','Heatmap unavailable above 32 intervals. Choose small multiples or Table.'));return figure}
  table.children[0].textContent=(mean?'Mean duration':'Captured records')+' · '+[...new Set(intervals.map(row=>utc(row.values.fromMs).slice(0,10)))].join(' / ')+' UTC · [start, end)';const corner=n('th','Group');corner.scope='col';header.append(corner);for(const row of intervals){const th=n('th',utc(row.values.fromMs).slice(11));th.scope='col';th.title='['+utc(row.values.fromMs)+', '+utc(row.values.toMs)+')';th.setAttribute('aria-label',th.title);header.append(th)}
  for(const group of visible){const tr=n('tr'),th=n('th',group.label);th.scope='row';tr.append(th);for(const row of group.rows){const td=n('td'),button=cellButton(row);if(mean?meanKnown(row):known(row)&&BigInt(row.values.count)>0n){const value=parts(mean?row.values['duration.mean']:row.values.count),top=parts(max),intensity=top.n===0n?0:Number(value.n*top.d*10000n/(value.d*top.n))/10000;button.style.background='rgba(57,201,173,'+(.12+.33*intensity)+')'}td.append(button);tr.append(td)}body.append(tr)}
 }else{
  for(const text of mean?['Group','UTC interval [start, end)','Mean duration','Captured spans']:['Group','UTC interval [start, end)','Captured records']){const th=n('th',text);th.scope='col';header.append(th)}
  const sameDate=new Set(snapshot.rows.map(row=>utc(row.values.fromMs).slice(0,10))).size===1;
  for(const group of visible)for(const row of group.rows){const tr=n('tr'),th=n('th',group.label),bounds='['+utc(row.values.fromMs)+', '+utc(row.values.toMs)+')',interval=n('td',bounds),value=n('td'),context=n('span',sameDate?utc(row.values.fromMs).slice(11):utc(row.values.fromMs));context.className='grouped-interval-context';context.title=bounds;context.setAttribute('aria-label',bounds);th.scope='row';th.append(context);value.title=rowLabel(snapshot,row,display);value.append(cellButton(row));tr.append(th,interval,value);if(mean){const count=n('td',known(row)?row.values.count:'?');count.title=rowLabel(snapshot,row,display);tr.append(count)}body.append(tr)}
 }
 head.append(header);table.append(head,body);scroll.append(table);figure.append(scroll,n('p',legend));return figure;
}
