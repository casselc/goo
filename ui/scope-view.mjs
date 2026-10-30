/* Public passive scoped displays. The host owns scope, requests and selection. */
import {validateSnapshot, validateDisplay, selectionIntent, scopeIntent} from './view-model.mjs';
import './chart-frame.mjs';

const n=(tag,text,cls)=>{const node=document.createElement(tag);if(text!==undefined)node.textContent=String(text);if(cls)node.className=cls;return node};
const exact=value=>value===undefined?'Not supplied':value===null?'Missing':value===''?'Present empty string':String(value);
const emit=(host,name,detail)=>host.dispatchEvent(new CustomEvent(name,{detail,bubbles:true,composed:true}));
const svgNode=(tag,attrs)=>{const node=document.createElementNS('http://www.w3.org/2000/svg',tag);for(const [key,value] of Object.entries(attrs))node.setAttribute(key,String(value));return node};

function quality(snapshot){
  const box=n('div',undefined,'scope-view-quality');
  const q=snapshot.quality;
  for(const text of [`${q.extent} groups`,`${q.value} values`,`${q.sampling} upstream sampling`,`${q.freshness} snapshot`])box.append(n('span',text));
  box.title='Extent describes the captured query, not complete upstream telemetry.';
  return box;
}

function capturedClock(value){
  if(typeof value!=='string')return '';
  const time=Date.parse(value);
  return Number.isFinite(time)?'Captured '+new Date(time).toISOString():'Capture time unavailable';
}

function table(host,snapshot,display){
  const frame=n('goo-table'),table=n('table'),head=n('thead'),row=n('tr'),body=n('tbody');
  frame.setAttribute('density','compact');
  table.append(n('caption','Exact supplied values'));
  const mapped=[display.category,display.value].map(key=>snapshot.fields.find(field=>field.key===key));
  const fields=mapped.every(Boolean)?mapped:snapshot.fields;
  for(const field of fields){const th=n('th',field.label+(field.unit?' · '+field.unit:''));th.scope='col';row.append(th)}
  const th=n('th','Selection');th.scope='col';row.append(th);head.append(row);
  for(const entry of snapshot.rows){
    const tr=n('tr');tr.dataset.key=entry.key;
    for(const field of fields)tr.append(n('td',exact(entry.values[field.key])));
    const cell=n('td'),button=n('button','Select');button.type='button';button.setAttribute('aria-label','Select '+exact(entry.values[display.category??fields[0].key]));button.onclick=()=>emit(host,'goo-select',selectionIntent(snapshot,entry.key));cell.append(button);tr.append(cell);body.append(tr);
  }
  table.append(head,body);frame.append(table);return frame;
}

function bars(host,snapshot,display){
  const list=n('div',undefined,'scope-view-bars'),numbers=snapshot.rows.map(row=>row.values[display.value]===null||row.values[display.value]===undefined?0:Number(row.values[display.value]));
  const max=Math.max(0,...numbers.map(Math.abs));
  for(const [index,row] of snapshot.rows.entries()){
    const button=n('button',undefined,'scope-view-bar');button.type='button';button.dataset.key=row.key;
    button.append(n('span',exact(row.values[display.category]),'scope-view-label'),n('span',exact(row.values[display.value]),'scope-view-value'));
    if(row.values[display.value]!==null&&row.values[display.value]!==undefined){const track=n('span',undefined,'scope-view-track'),fill=n('span',undefined,'scope-view-fill');track.setAttribute('aria-hidden','true');fill.style.width=(max?Math.abs(numbers[index])/max*100:0)+'%';fill.dataset.sign=numbers[index]<0?'negative':'positive';track.append(fill);button.append(track)}
    button.setAttribute('aria-label','Select '+exact(row.values[display.category])+': '+exact(row.values[display.value]));button.onclick=()=>emit(host,'goo-select',selectionIntent(snapshot,row.key));list.append(button);
  }
  return list;
}

function line(snapshot,display){
  const figure=n('figure',undefined,'scope-view-line'),svg=svgNode('svg',{viewBox:'0 0 640 180','aria-hidden':'true',focusable:'false'});
  // Numeric coordinates are disposable projections. Exact supplied values remain in the table.
  const points=snapshot.rows.map(row=>[Number(row.values[display.category]),Number(row.values[display.value])]).sort((a,b)=>a[0]-b[0]);
  // Scale before subtracting so finite extreme coordinates cannot overflow their span.
  const scaleX=Math.max(1,...points.map(point=>Math.abs(point[0]))),scaleY=Math.max(1,...points.map(point=>Math.abs(point[1])));
  const xs=points.map(point=>point[0]/scaleX),ys=points.map(point=>point[1]/scaleY),minX=Math.min(...xs),maxX=Math.max(...xs),minY=Math.min(0,...ys),maxY=Math.max(0,...ys);
  const project=([x,y])=>[24+(maxX===minX?.5:(x/scaleX-minX)/(maxX-minX))*592,156-(maxY===minY?.5:(y/scaleY-minY)/(maxY-minY))*132];
  svg.append(svgNode('path',{d:'M24 24 V156 H616',class:'scope-view-axis'}));
  // Unknown/estimated values may be points; a connecting line would invent continuity.
  if(snapshot.quality.value==='recorded'||snapshot.quality.value==='derived')svg.append(svgNode('polyline',{points:points.map(point=>project(point).join(',')).join(' '),class:'scope-view-path'}));
  for(const point of points){const [cx,cy]=project(point);svg.append(svgNode('circle',{cx,cy,r:3,class:'scope-view-point'}))}
  const x=snapshot.fields.find(field=>field.key===display.category),y=snapshot.fields.find(field=>field.key===display.value);
  figure.append(svg,n('figcaption',`${x.label} → ${y.label}${y.unit?' · '+y.unit:''}`));return figure;
}

class SuppliedElement extends (globalThis.HTMLElement||class {}) {
  static get observedAttributes(){return ['snapshot','display','field']}
  connectedCallback(){for(const key of ['snapshot','display','field'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
  attributeChangedCallback(){this.render()}
  set snapshot(value){this._snapshot=value;this.render()}
  get snapshot(){return this.read('snapshot')}
  read(name){if(Object.hasOwn(this,'_'+name))return this['_'+name];const text=this.getAttribute(name);return text===null?undefined:JSON.parse(text)}
  refuse(reasons){const status=n('p',reasons.join('; '),'scope-view-unavailable');status.setAttribute('role','status');this.replaceChildren(status)}
}

class Chart extends SuppliedElement {
  set display(value){this._display=value;this.render()}
  get display(){return this.read('display')}
  render(){
    try{
      const snapshot=this.read('snapshot'),display=this.read('display');
      if(!snapshot||!display){this.refuse(['Supply a dataset snapshot and display mapping.']);return}
      const checked=validateDisplay(snapshot,display);if(!checked.ok){this.refuse(checked.reasons);return}
      const root=n('goo-chart-frame',undefined,'scope-view');root.data={title:display.title||'Scoped breakdown',window:capturedClock(snapshot.capturedAt),state:!snapshot.rows.length?'no_data':snapshot.quality.extent==='partial'?'partial':snapshot.quality.extent==='unknown'||snapshot.quality.value==='unknown'?'unknown':'observed'};root.dataset.valueKnowledge=snapshot.quality.value;root.append(quality(snapshot));
      if(!snapshot.rows.length)root.append(n('p','No supplied groups.','scope-view-empty'));
      else if(display.kind==='table')root.append(table(this,snapshot,display));
      else if(snapshot.quality.value==='unknown'){root.append(n('p','Geometry unavailable: value knowledge is unknown. Supplied values remain inspectable.','scope-view-unavailable'),table(this,snapshot,display))}
      else if(display.kind==='bar'&&snapshot.rows.some(row=>row.values[display.value]!==null&&Number(row.values[display.value])<0)){root.append(n('p','This bar kind requires nonnegative values. Use the exact table for signed values.','scope-view-unavailable'),table(this,snapshot,display))}
      else if(display.kind==='line'&&snapshot.rows.some(row=>[display.category,display.value].some(key=>row.values[key]===null||row.values[key]===undefined))){root.append(n('p','Line unavailable: a coordinate is missing. Exact values remain available.','scope-view-unavailable'),table(this,snapshot,display))}
      else{
        root.append(display.kind==='bar'?bars(this,snapshot,display):line(snapshot,display));
        const fallback=n('details',undefined,'scope-view-values');fallback.append(n('summary','Exact values'),table(this,snapshot,display));root.append(fallback);
        root.append(n('small','Geometry uses approximate numeric coordinates; exact values are preserved.','scope-view-coordinate-note'));
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
      const field=snapshot.fields.find(item=>item.key===this.field);if(!field){this.refuse(['Choose a supplied facet field.']);return}
      const root=n('div',undefined,'scope-view-facet');root.setAttribute('role','group');root.setAttribute('aria-label','Filter '+field.label);root.append(n('span',field.label,'scope-view-facet-label'));
      const values=new Map();for(const row of snapshot.rows){const value=row.values[field.key];if(value!==undefined)values.set(typeof value+':'+String(value),value)}
      for(const value of values.values()){const button=n('button',exact(value));button.type='button';button.onclick=()=>emit(this,'goo-scope',scopeIntent(snapshot,field.key,value));root.append(button)}
      if(!values.size)root.append(n('span','No supplied values.'));
      root.append(n('small','From supplied groups · stage a filter','scope-view-facet-note'));this.replaceChildren(root);
    }catch(error){this.refuse([error instanceof SyntaxError?'Invalid snapshot JSON.':error.message||'Unable to display supplied facets.'])}
  }
}

if(globalThis.customElements&&globalThis.HTMLElement){
  if(!customElements.get('goo-chart'))customElements.define('goo-chart',Chart);
  if(!customElements.get('goo-scope-facet'))customElements.define('goo-scope-facet',Facet);
}
