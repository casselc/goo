/* Public passive scoped displays. The host owns scope, requests and selection. */
import {validateSnapshot, validateDisplay, selectionIntent, scopeIntent, measureValue, supportsSelection, capabilities,displayRows,populationShare,fractionFacts,histogramIntent} from './view-model.mjs';
import {formatCell,supportText,exactText,categoryText,isFraction,fractionContextText} from './view-values.mjs';
import {asSnapshot,parseSnapshotJSON} from './snapshots.mjs';
import {rationalDisplay} from './numeric-values.mjs';
import './chart-frame.mjs';
import './time-bars.mjs';

const n=(tag,text,cls)=>{const node=document.createElement(tag);if(text!==undefined)node.textContent=String(text);if(cls)node.className=cls;return node};
const exact=categoryText;
const emit=(host,name,detail)=>host.dispatchEvent(new CustomEvent(name,{detail,bubbles:true,composed:true}));
const svgNode=(tag,attrs)=>{const node=document.createElementNS('http://www.w3.org/2000/svg',tag);for(const [key,value] of Object.entries(attrs))node.setAttribute(key,String(value));return node};
function frozenConfig(value){
  if(value===undefined)return undefined;
  if(!value||typeof value!=='object'||Array.isArray(value)||![Object.prototype,null].includes(Object.getPrototypeOf(value)))throw TypeError('Supply display options.');
  const keys=Reflect.ownKeys(value);if(keys.length>10)throw TypeError('Too many display options.');
  const result={};for(const key of keys){
    if(typeof key!=='string'||!['kind','title','category','value','start','end','requiresComplete','scale','query','order'].includes(key))throw TypeError('Unsupported display option.');
    const descriptor=Object.getOwnPropertyDescriptor(value,key);if(!Object.hasOwn(descriptor,'value')||!descriptor.enumerable)throw TypeError('Display options must contain plain values.');
    const item=descriptor.value;if(key==='requiresComplete'?typeof item!=='boolean':typeof item!=='string'||item.length>1024)throw TypeError('Display options require bounded strings and a boolean completeness flag.');result[key]=item;
  }
  return Object.freeze(result);
}

function quality(snapshot){
  const box=n('div',undefined,'scope-view-quality');
  const q=snapshot.quality;
  for(const text of [snapshot.origin,`${q.extent} ${snapshot.shape==='time-series'?'intervals':snapshot.shape==='histogram'?'bins':snapshot.shape==='summary'?'summary':'groups'}`,`${q.value} values`,`${q.sampling} upstream sampling`,`${q.freshness} snapshot`])if(text)box.append(n('span',text));
  box.title='Extent describes the captured query, not complete upstream telemetry.';
  return box;
}

function capturedClock(value){
  if(typeof value!=='string')return '';
  const time=Date.parse(value);
  return Number.isFinite(time)?'Captured '+new Date(time).toISOString():'Capture time unavailable';
}

function utc(value){
  const time=Number(value);
  return Number.isSafeInteger(time)&&Number.isFinite(new Date(time).getTime())?new Date(time).toISOString():exact(value);
}

function cellValue(field,value,knowledge,facts,exactBoundary=false){
  const cell=n('td'),text=n('span',exactBoundary?rationalDisplay(value)+(field.unit?' '+field.unit:''):formatCell(field,value,knowledge));cell.append(text);cell.title=exactText(field,value,knowledge,facts);
  if(field.type==='timestamp'&&value!==null&&value!==undefined)cell.title=String(value)+' ms since Unix epoch';
  const needsExact=field.type==='timestamp'||field.type==='rational'||field.type==='number'&&value!==null&&value!==undefined&&(typeof value==='number'||/[./eE]/.test(String(value)));
  if(knowledge||needsExact){
    if(knowledge)cell.dataset.cellState=knowledge.state;
    const details=n('details',undefined,'scope-view-cell-support'),summary=n('summary',knowledge?knowledge.state+(knowledge.support?' · '+knowledge.support.valid+'/'+knowledge.support.total+' valid':''):'Exact value');
    details.append(summary,n('p',exactText(field,value,knowledge,facts),'scope-view-exact'));cell.append(details);
  }
  return cell;
}

function table(host,snapshot,display,rows=displayRows(snapshot,display)){
  const frame=n('goo-table'),table=n('table'),head=n('thead'),row=n('tr'),body=n('tbody');
  frame.setAttribute('density','compact');
  table.append(n('caption',display.scale==='population'?'Supplied values · shares of '+snapshot.partition.total+' captured records':'Supplied values'));
  const mapped=[display.category,display.value].map(key=>snapshot.fields.find(field=>field.key===key));
  const fields=display.kind==='table'?snapshot.fields:mapped.every(Boolean)?mapped:snapshot.fields;
  for(const field of fields){const th=n('th',field.label+(isFraction(field)?' · 0–100%':field.unit&&field.unit!=='ns'?' · '+field.unit:''));th.scope='col';row.append(th)}
  if(display.scale==='population'){const th=n('th','Share');th.scope='col';th.title='Share of the full captured query population';row.append(th)}
  if(snapshot.shape==='time-series'){const th=n('th','Interval state');th.scope='col';row.append(th)}
  const selectable=supportsSelection(snapshot);
  if(selectable){const th=n('th','Selection');th.scope='col';row.append(th)}head.append(row);
  for(const entry of rows){
    const tr=n('tr');tr.dataset.key=entry.key;
    for(const field of fields)tr.append(cellValue(field,entry.values[field.key],entry.cells?.[field.key],fractionFacts(snapshot,entry,field.key),snapshot.shape==='histogram'&&['from','to'].includes(field.key)));
    if(display.scale==='population'){const share=populationShare(snapshot,entry.key),cell=n('td',share.ok?share.percent:'Unavailable');if(share.ok){cell.title=share.fraction+' captured records (exact)';cell.dataset.share=share.fraction}tr.append(cell)}
    if(snapshot.shape==='time-series')tr.append(n('td',entry.state==='no_data'?'No data · zero captured records':entry.state==='unknown'?'Unknown coverage':'Observed'));
    if(selectable){const firstField=fields.find(field=>field.key===display.category)||fields[0];
      const cell=n('td'),button=n('button','Select');button.type='button';button.setAttribute('aria-label','Select '+(snapshot.dimensions?.length?snapshot.dimensions.map(key=>exact(entry.values[key])).join(' / '):firstField.type==='timestamp'?utc(entry.values[firstField.key]):exact(entry.values[firstField.key])));button.onclick=()=>emit(host,'goo-select',selectionIntent(snapshot,entry.key));cell.append(button);tr.append(cell)}body.append(tr);
  }
  table.append(head,body);frame.append(table);return frame;
}

function timeBars(host,snapshot,display){
  const figure=n('figure',undefined,'scope-time'),track=n('div',undefined,'scope-time-track');
  const valueKey=display.value||'count',field=snapshot.fields.find(item=>item.key===valueKey);
  const from=BigInt(snapshot.interval.fromMs),to=BigInt(snapshot.interval.toMs),span=to-from;
  const percent=value=>Number(value*100000000n/span)/1000000;
  const rows=[...snapshot.rows].sort((a,b)=>BigInt(a.values.fromMs)<BigInt(b.values.fromMs)?-1:1);
  const values=rows.map(row=>measureValue(snapshot,row,valueKey)),max=isFraction(field)?1:Math.max(0,...values.map(value=>value.coordinate??0));
  const gaps=[];let cursor=from;
  const addGap=(start,end)=>{
    const gap=n('span',undefined,'scope-time-gap');gap.style.left=percent(start-from)+'%';gap.style.width=percent(end-start)+'%';
    gap.title=utc(start.toString())+' → '+utc(end.toString())+' · Unreported interval, not zero';gap.setAttribute('aria-hidden','true');track.append(gap);gaps.push([start,end]);
  };
  for(const [index,row] of rows.entries()){
    const start=BigInt(row.values.fromMs),end=BigInt(row.values.toMs);
    if(start>cursor)addGap(cursor,start);
    cursor=end;
    const value=values[index],cell=row.cells?.[valueKey],state=valueKey==='count'?row.state:value.coordinate===null?'unknown':value.coordinate===0?'no_data':'observed',button=n('button',undefined,'scope-time-interval');button.type='button';button.dataset.key=row.key;button.dataset.state=state;button.dataset.cellState=value.state;
    button.style.left=percent(start-from)+'%';button.style.width=percent(end-start)+'%';
    const range=utc(row.values.fromMs)+' → '+utc(row.values.toMs),meaning=valueKey==='count'&&row.state==='unknown'?'Unknown coverage':formatCell(field,row.values[valueKey],cell);
    button.title=range+' · '+meaning+' · '+exactText(field,row.values[valueKey],cell,fractionFacts(snapshot,row,valueKey))+(value.reason?' · '+value.reason:'')+' · ['+row.values.fromMs+', '+row.values.toMs+') ms';button.setAttribute('aria-label','Select '+range+' · '+field.label+': '+meaning+(isFraction(field)?' · '+fractionContextText(fractionFacts(snapshot,row,valueKey)):'')+(supportText(cell)?' · '+supportText(cell):''));
    const fill=n('span',undefined,'scope-time-fill');fill.setAttribute('aria-hidden','true');fill.style.height=(value.coordinate!==null&&max?value.coordinate/max*100:0)+'%';button.append(fill);
    if(state==='observed'){const count=n('span',formatCell(field,row.values[valueKey],cell),'scope-time-count');count.setAttribute('aria-hidden','true');button.append(count)}
    button.onclick=()=>emit(host,'goo-select',selectionIntent(snapshot,row.key));track.append(button);
  }
  if(cursor<to)addGap(cursor,to);
  const axis=n('figcaption',undefined,'scope-time-axis');axis.append(n('span',utc(snapshot.interval.fromMs)),n('span',utc(snapshot.interval.toMs)));
  const legend=n('p',undefined,'scope-time-legend');legend.append(n('span',valueKey==='count'?'Captured records':field.label+(isFraction(field)?' · 0–100%':''),'scope-time-observed-key'),n('span',valueKey==='count'?'No data · zero':'Known zero','scope-time-empty-key'),n('span',valueKey==='count'?'Unknown / unreported':'Unavailable / unreported','scope-time-unknown-key'));
  figure.append(measureLabel(field),track,axis,legend);
  if(gaps.length){const details=n('details',undefined,'scope-view-values');details.append(n('summary',gaps.length+' unreported interval'+(gaps.length===1?'':'s')));const list=n('ul');for(const [start,end] of gaps)list.append(n('li',utc(start.toString())+' → '+utc(end.toString())+' · Unknown, not zero'));details.append(list);figure.append(details)}
  return figure;
}


function histogram(host,snapshot){
 const figure=n('section',undefined,'scope-histogram'),support=snapshot.histogram.support,unit=snapshot.histogram.binding.unit;
 const context=n('p',support.valid+'/'+support.total+' captured records have valid numeric values · '+support.missing+' missing · '+support.empty+' empty · '+support.invalid+' invalid','scope-time-legend');
 figure.append(context);
 if(support.valid==='0'){figure.append(n('p',support.total==='0'?'No captured records. The fixed bins are known empty.':'No valid numeric values. Empty bins do not imply an empty captured population.','scope-view-empty'));return figure}
 const all=snapshot.rows,first=all.findIndex(row=>row.values.count!=='0');let last=all.length-1;while(last>first&&all[last].values.count==='0')last--;
 const rows=all.slice(first,last+1),ceiling=all.reduce((n,row)=>BigInt(row.values.count)>n?BigInt(row.values.count):n,0n).toString();
 const scroll=n('div',undefined,'scope-histogram-scroll'),bars=n('goo-time-bars');bars.style.minWidth=Math.max(440,rows.length*8)+'px';bars.setAttribute('range-select','');scroll.append(bars);
 bars.data={title:'Captured numeric distribution',unit:'records',ceiling,interactive:true,fromLabel:rationalDisplay(rows[0].values.from)+' '+unit,toLabel:rationalDisplay(rows.at(-1).values.to)+' '+unit,points:rows.map(row=>({key:row.key,label:'['+rationalDisplay(row.values.from)+', '+rationalDisplay(row.values.to)+') '+unit+' · '+row.values.count+' captured records',value:row.values.count,state:row.state}))};
 const range=(fromKey,toKey)=>emit(host,'goo-histogram-range',histogramIntent(snapshot,fromKey,toKey));
 bars.addEventListener('goo-time-range',event=>{event.stopPropagation();range(event.detail.fromKey,event.detail.toKey)});
 bars.addEventListener('goo-time-select',event=>{event.stopPropagation();range(event.detail.key,event.detail.key)});
 figure.append(scroll,n('small','Equal-width bars · unequal numeric ranges · counts, not density · outer empty bins hidden; all bins remain in Exact values.','scope-view-coordinate-note'));
 return figure;
}

function bars(host,snapshot,display,rows=displayRows(snapshot,display)){
  const list=n('div',undefined,'scope-view-bars'),field=snapshot.fields.find(item=>item.key===display.value),values=snapshot.rows.map(row=>measureValue(snapshot,row,display.value));
  const max=isFraction(field)?1:Math.max(0,...values.map(value=>value.coordinate??0)),byKey=new Map(snapshot.rows.map((row,i)=>[row.key,values[i]]));list.append(measureLabel(field));
  if(display.scale==='population')list.append(n('small',snapshot.partition.total+' captured records · full query population','scope-view-coordinate-note'));
  for(const row of rows){
    const value=byKey.get(row.key),share=display.scale==='population'?populationShare(snapshot,row.key):null,cell=row.cells?.[display.value],button=n('button',undefined,'scope-view-bar');button.type='button';button.dataset.key=row.key;button.dataset.cellState=value.state;
    button.append(n('span',exact(row.values[display.category]),'scope-view-label'),n('span',formatCell(field,row.values[display.value],cell)+(share?.ok?' · '+share.percent:''),'scope-view-value'));
    if(value.coordinate!==null){const track=n('span',undefined,'scope-view-track'),fill=n('span',undefined,'scope-view-fill');track.setAttribute('aria-hidden','true');fill.style.width=(share?.ok?share.coordinate:max?value.coordinate/max*100:0)+'%';track.append(fill);button.append(track)}
    button.title=exactText(field,row.values[display.value],cell,fractionFacts(snapshot,row,display.value))+(share?.ok?' · '+share.fraction+' captured records (exact share)':'')+(value.reason?' · '+value.reason:'');
    button.setAttribute('aria-label','Select '+exact(row.values[display.category])+': '+formatCell(field,row.values[display.value],cell)+(share?.ok?' · '+share.percent+' of captured population':'')+(isFraction(field)?' · '+fractionContextText(fractionFacts(snapshot,row,display.value)):'')+(supportText(cell)?' · '+supportText(cell):''));button.onclick=()=>emit(host,'goo-select',selectionIntent(snapshot,row.key));list.append(button);
  }
  return list;
}

function measureLabel(field){
  const label=n('p',field.label+(isFraction(field)?' · 0–100%':field.unit&&field.unit!=='ns'?' · '+field.unit:''),'scope-view-measure');
  label.title=field.aggregation?field.aggregation.method+(field.aggregation.quantile?' · q='+field.aggregation.quantile:''):field.unit||'';return label;
}

function line(snapshot,display){
  const figure=n('figure',undefined,'scope-view-line'),svg=svgNode('svg',{viewBox:'0 0 640 180','aria-hidden':'true',focusable:'false'});
  // Numeric coordinates are disposable projections. Exact supplied values remain in the table.
  const points=snapshot.rows.map(row=>[Number(row.values[display.category]),measureValue(snapshot,row,display.value).coordinate]).sort((a,b)=>a[0]-b[0]);
  // Scale before subtracting so finite extreme coordinates cannot overflow their span.
  const valueField=snapshot.fields.find(field=>field.key===display.value),fractionAxis=isFraction(valueField);
  const scaleX=Math.max(1,...points.map(point=>Math.abs(point[0]))),scaleY=fractionAxis?1:Math.max(1,...points.map(point=>Math.abs(point[1])));
  const xs=points.map(point=>point[0]/scaleX),ys=points.map(point=>point[1]/scaleY),minX=Math.min(...xs),maxX=Math.max(...xs),minY=fractionAxis?0:Math.min(0,...ys),maxY=fractionAxis?1:Math.max(0,...ys);
  const project=([x,y])=>[24+(maxX===minX?.5:(x/scaleX-minX)/(maxX-minX))*592,156-(maxY===minY?.5:(y/scaleY-minY)/(maxY-minY))*132];
  svg.append(svgNode('path',{d:'M24 24 V156 H616',class:'scope-view-axis'}));
  // Unknown/estimated values may be points; a connecting line would invent continuity.
  if((snapshot.quality.value==='recorded'||snapshot.quality.value==='derived')&&!snapshot.rows.some(row=>measureValue(snapshot,row,display.value).state==='estimated'))svg.append(svgNode('polyline',{points:points.map(point=>project(point).join(',')).join(' '),class:'scope-view-path'}));
  for(const point of points){const [cx,cy]=project(point);svg.append(svgNode('circle',{cx,cy,r:3,class:'scope-view-point'}))}
  const x=snapshot.fields.find(field=>field.key===display.category),y=snapshot.fields.find(field=>field.key===display.value);
  figure.append(svg,n('figcaption',`${x.label} → ${y.label}${fractionAxis?' · 0–100%':y.unit&&y.unit!=='ns'?' · '+y.unit:''}`));return figure;
}

class SuppliedElement extends (globalThis.HTMLElement||class {}) {
  static get observedAttributes(){return ['snapshot','display','field']}
  connectedCallback(){for(const key of ['snapshot','display','field'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
  attributeChangedCallback(){this.render()}
  set snapshot(value){this.setInput('snapshot',value)}
  get snapshot(){return this.read('snapshot')}
  normalize(name,value){return name==='snapshot'?asSnapshot(value):frozenConfig(value)}
  setInput(name,value){this._attributeCache?.delete(name);try{this['_'+name]=this.normalize(name,value);if(this._inputErrors)delete this._inputErrors[name]}catch(error){this['_'+name]=null;(this._inputErrors??={})[name]=error.message||'Invalid supplied value.'}this.render()}
  read(name){
    if(Object.hasOwn(this,'_'+name))return this['_'+name];
    const text=this.getAttribute(name);if(text===null)return undefined;
    const cache=(this._attributeCache??=new Map()).get(name);if(cache?.text===text)return cache.value;
    let value;try{if(name==='display'&&text.length>16384)throw TypeError('Display options are too large.');value=name==='snapshot'?parseSnapshotJSON(text):this.normalize(name,JSON.parse(text));if(this._inputErrors)delete this._inputErrors[name]}catch(error){value=null;(this._inputErrors??={})[name]=error instanceof SyntaxError?'Invalid '+name+' JSON.':error.message||'Invalid supplied value.'}
    this._attributeCache.set(name,{text,value});return value;
  }
  refuse(reasons){const status=n('p',reasons.join('; '),'scope-view-unavailable');status.setAttribute('role','status');this.replaceChildren(status)}
}

class Chart extends SuppliedElement {
  set display(value){this.setInput('display',value)}
  get display(){return this.read('display')}
  clearRange(){this.querySelector('goo-time-bars')?.clearRange()}
  set selectedKey(value){const bars=this.querySelector('goo-time-bars');if(bars)bars.selectedKey=value}
  get selectedKey(){return this.querySelector('goo-time-bars')?.selectedKey??null}
  render(){
    try{
      const snapshot=this.read('snapshot'),display=this.read('display');
      if(!snapshot||!display){this.refuse([this._inputErrors?.snapshot||this._inputErrors?.display||'Supply a dataset snapshot and display mapping.']);return}
      const checked=validateDisplay(snapshot,display);if(!checked.ok){
        const rawDisplay={...display,scale:'max',requiresComplete:false};
        if(display.scale==='population'&&!display.requiresComplete&&snapshot.shape==='ranked'&&['bar','table'].includes(display.kind)&&validateDisplay(snapshot,rawDisplay).ok){
          const frame=n('goo-chart-frame',undefined,'scope-view');frame.data={title:display.title||'Scoped breakdown',window:capturedClock(snapshot.capturedAt),state:'unavailable'};
          const status=n('p',checked.reasons.join('; '),'scope-view-unavailable');status.setAttribute('role','status');const values=table(this,snapshot,{...rawDisplay,kind:'table'});values.querySelector('caption').textContent='Supplied values · shares unavailable';frame.append(quality(snapshot),status,values);this.replaceChildren(frame);return;
        }
        this.refuse(checked.reasons);return;
      }
      const rows=snapshot.shape==='ranked'?displayRows(snapshot,display):snapshot.rows;
      const root=n('goo-chart-frame',undefined,'scope-view');root.data={title:display.title||(snapshot.shape==='histogram'?'Captured numeric distribution':snapshot.shape==='time-series'?'Captured records over time':snapshot.shape==='summary'?'Population summary':'Scoped breakdown'),window:capturedClock(snapshot.capturedAt),state:snapshot.quality.extent==='unknown'||snapshot.quality.value==='unknown'?'unknown':snapshot.quality.extent==='partial'?'partial':!snapshot.rows.length||snapshot.shape==='histogram'&&snapshot.histogram.support.total==='0'||snapshot.shape==='summary'&&snapshot.rows[0].values.count==='0'?'no_data':'observed'};root.dataset.valueKnowledge=snapshot.quality.value;root.append(quality(snapshot));
      if(!snapshot.rows.length&&display.kind!=='time-bar')root.append(n('p',snapshot.shape==='time-series'?'No supplied intervals. Coverage is unknown.':'No supplied groups.','scope-view-empty'));
      else if(!rows.length&&snapshot.shape==='ranked')root.append(n('p','No matching groups. '+snapshot.rows.length+' groups remain in this result.','scope-view-empty'));
      else if(display.kind==='table')root.append(table(this,snapshot,display,rows));
      else if(snapshot.quality.value==='unknown'){root.append(n('p','Geometry unavailable: value knowledge is unknown. Supplied values remain inspectable.','scope-view-unavailable'),table(this,snapshot,display))}
      else if(['bar','time-bar'].includes(display.kind)&&snapshot.rows.some(row=>(measureValue(snapshot,row,display.value||'count').coordinate??0)<0)){root.append(n('p','This bar kind requires nonnegative values. Use the exact table for signed values.','scope-view-unavailable'),table(this,snapshot,display))}
      else if(display.kind==='line'&&snapshot.rows.some(row=>row.values[display.category]===null||row.values[display.category]===undefined||measureValue(snapshot,row,display.value).coordinate===null)){root.append(n('p','Line unavailable: a coordinate is missing. Exact values remain available.','scope-view-unavailable'),table(this,snapshot,display))}
      else{
        root.append(display.kind==='histogram'?histogram(this,snapshot):display.kind==='time-bar'?timeBars(this,snapshot,display):display.kind==='bar'?bars(this,snapshot,display,rows):line(snapshot,display));
        const fallback=n('details',undefined,'scope-view-values');fallback.append(n('summary','Exact values'),table(this,snapshot,display,rows));root.append(fallback);
        const note=n('small','Approximate geometry','scope-view-coordinate-note');note.title='Plot coordinates are approximate; exact source values and methods remain in cell tooltips.';root.append(note);
      }
      this.replaceChildren(root);
      const clock=root.querySelector('.chart-frame-window');if(clock&&typeof snapshot.capturedAt==='string')clock.title=snapshot.capturedAt;
    }catch(error){this.refuse([error instanceof SyntaxError?'Invalid snapshot or display JSON.':error.message||'Unable to display supplied data.'])}
  }
}

class Facet extends SuppliedElement {
  set field(value){this._field=value;this.render()}
  get field(){return this._field??this.getAttribute('field')}
  render(){
    try{
      const snapshot=this.read('snapshot');if(!snapshot){this.refuse(['Supply a dataset snapshot.']);return}
      const checked=validateSnapshot(snapshot);if(!checked.ok){this.refuse(checked.reasons);return}
      if(!capabilities(snapshot).filterDimensions){this.refuse(['This dataset has no filterable group dimensions.']);return}
      const field=snapshot.fields.find(item=>item.key===this.field);if(!field){this.refuse(['Choose a supplied facet field.']);return}
      if(field.aggregation||snapshot.dimensions&&!snapshot.dimensions.includes(field.key)){this.refuse(['Choose a group dimension. Aggregate measures are not record filters.']);return}
      const root=n('div',undefined,'scope-view-facet');root.setAttribute('role','group');root.setAttribute('aria-label','Filter '+field.label);root.append(n('span',field.label,'scope-view-facet-label'));
      const values=new Map();for(const row of snapshot.rows){const value=row.values[field.key];if(value!==undefined)values.set(typeof value+':'+String(value),value)}
      for(const value of values.values()){const button=n('button',field.type==='timestamp'&&value!==null?utc(value):exact(value));button.type='button';if(field.type==='timestamp'&&value!==null)button.title=String(value)+' ms since Unix epoch';button.onclick=()=>emit(this,'goo-scope',scopeIntent(snapshot,field.key,value));root.append(button)}
      if(!values.size)root.append(n('span','No supplied values.'));
      root.append(n('small','From supplied groups · stage a filter','scope-view-facet-note'));this.replaceChildren(root);
    }catch(error){this.refuse([error instanceof SyntaxError?'Invalid snapshot JSON.':error.message||'Unable to display supplied facets.'])}
  }
}

if(globalThis.customElements&&globalThis.HTMLElement){
  if(!customElements.get('goo-chart'))customElements.define('goo-chart',Chart);
  if(!customElements.get('goo-scope-facet'))customElements.define('goo-scope-facet',Facet);
}
