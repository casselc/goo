/* Passive population summaries. Go owns aggregates, comparison changes and their basis. */
import {validateSnapshot,measureValue} from './view-model.mjs';
import {validateComparison} from './view-comparison.mjs';
import {signedScale} from './comparison-visual.mjs';
import {formatCell,exactText} from './view-values.mjs';
import './chart-frame.mjs';

const n=(tag,text,cls)=>{const e=document.createElement(tag);if(text!==undefined)e.textContent=String(text);if(cls)e.className=cls;return e};
const copy=value=>structuredClone(value);
const numeric=field=>field?.type==='number'||field?.type==='rational';
let tableSequence=0;
const clock=value=>{if(value===null||value===undefined)return null;const d=new Date(typeof value==='string'&&/^-?\d+$/.test(value)?Number(value):value);return Number.isFinite(d.getTime())?d.toISOString().replace('T',' ').replace('Z',' UTC'):null};

function options(display,fields,{comparison=false}={}){
  if(display===undefined)display={};
  if(!display||typeof display!=='object'||Array.isArray(display))throw TypeError('Supply display options.');
  const allowed=comparison?['fields','title']:['fields','title','compact'];
  if(Object.keys(display).some(key=>!allowed.includes(key)))throw TypeError('Unsupported display option.');
  if(display.title!==undefined&&(typeof display.title!=='string'||display.title.length>512))throw TypeError('Use a short display title.');
  if(display.compact!==undefined&&typeof display.compact!=='boolean')throw TypeError('compact must be boolean.');
  const keys=display.fields??fields.filter(numeric).map(field=>field.key);
  if(!Array.isArray(keys)||!keys.length||keys.length>64||new Set(keys).size!==keys.length||keys.some(key=>!fields.some(field=>field.key===key&&numeric(field))))throw TypeError('Choose distinct supplied numeric measures.');
  return {...display,fields:keys,compact:display.compact!==false,configured:display.fields!==undefined};
}

function provenance(snapshot,label){
  const root=n('div',undefined,'measure-scope');if(label)root.append(n('strong',label));
  const q=snapshot.quality,facts=n('div',undefined,'measure-scope-facts');
  for(const text of [snapshot.scope.signal,snapshot.origin,`${q.extent} captured summary`,`${q.sampling} upstream sampling`])if(text)facts.append(n('span',text));
  root.append(facts);
  if(snapshot.capturedAt){const time=n('time',clock(snapshot.capturedAt)||'Capture time unavailable');time.dateTime=snapshot.capturedAt;time.title=snapshot.capturedAt;root.append(time)}
  else root.append(n('span','Capture time not supplied','measure-clock-missing'));
  const selector=snapshot.scope.selector,from=clock(selector.fromMs),to=clock(selector.toMs);
  if(selector.fromMs!==undefined&&selector.toMs!==undefined&&from&&to)root.append(n('p',from+' → '+to,'measure-window'));
  const details=n('details',undefined,'measure-provenance'),dl=n('dl');details.append(n('summary','Scope & quality'));
  for(const [name,value] of [['Scope',snapshot.scope.id],['Generation',snapshot.scope.generation],['Dataset',snapshot.id],['Revision',snapshot.revision],['Basis',q.basis],['Source basis',snapshot.sourceBasis],['Captured at',snapshot.capturedAt],['Values',q.value],['Delivery mode',q.freshness]]){
    if(value===undefined)continue;const group=n('div');group.append(n('dt',name),n('dd',typeof value==='object'?JSON.stringify(value):value));dl.append(group);
  }
  details.append(dl,n('pre',JSON.stringify(selector,null,2)));
  if(snapshot.completeness)details.append(n('p','Attested aspects: '+snapshot.completeness.aspects.join(', ')));
  root.append(details);return root;
}

function valueCell(field,row,{delta}={}){
  const cell=delta?undefined:row?.cells?.[field?.key],raw=delta?delta.value:row?.values?.[field?.key];
  const body=n('div',undefined,'measure-cell');body.dataset.state=delta?.state||cell?.state||'recorded';
  const text=delta&&delta.state!=='derived'?'Unavailable':field?formatCell(field,raw,cell):'Unavailable';
  body.append(n('span',text,'measure-value'));
  const details=n('details',undefined,'measure-exact'),status=delta?delta.state:cell?.state;
  details.append(n('summary',status?(cell?.support?`${status} · ${cell.support.valid}/${cell.support.total} valid`:status):'Exact value'));
  details.append(n('p',delta?(delta.state==='derived'?exactText({...field,aggregation:undefined},raw):delta.reason||'Change is unavailable.'):field?exactText(field,raw,cell):'Measure not supplied.'));
  if(delta?.state==='derived')details.append(n('p','Selection − baseline · exact supplied change'));
  body.append(details);return body;
}

function measureList(snapshot,fields){
  const dl=n('dl',undefined,'measure-stats'),row=snapshot.rows[0];
  for(const field of fields){const item=n('div'),dd=n('dd');item.dataset.field=field.key;dd.append(valueCell(field,row));item.append(n('dt',field.label),dd);dl.append(item)}
  return dl;
}

function summaryBody(snapshot,display){
  const root=n('div',undefined,'measure-summary'),fields=display.fields.map(key=>snapshot.fields.find(field=>field.key===key));
  const groups=[['Captured records',fields.filter(field=>!field.aggregation)],['Numeric values',fields.filter(field=>field.aggregation?.family==='numeric')],['Duration',fields.filter(field=>field.aggregation?.family==='duration')]];
  for(const [title,items] of groups){if(!items.length)continue;const section=n('section',undefined,'measure-group');section.append(n('h4',title));
    const primary=display.compact&&!display.configured?items.filter(field=>!field.aggregation||['mean','quantile'].includes(field.aggregation.kind)):items;
    const more=items.filter(field=>!primary.includes(field));if(primary.length)section.append(measureList(snapshot,primary));
    if(more.length){const details=n('details',undefined,'measure-more');details.append(n('summary','More measures · '+more.length),measureList(snapshot,more));section.append(details)}
    root.append(section);
  }
  return root;
}

function compatibleFields(a,b){return a&&b&&a.type===b.type&&a.unit===b.unit&&['family','kind','method','quantile'].every(key=>a.aggregation?.[key]===b.aggregation?.[key])}

function comparisonBody(comparison,display){
  const viewport=n('goo-table'),table=n('table',undefined,'measure-comparison'),head=n('thead'),header=n('tr'),body=n('tbody');viewport.setAttribute('density','compact');
  const prefix='measure-comparison-'+(++tableSequence);table.setAttribute('role','table');head.setAttribute('role','rowgroup');body.setAttribute('role','rowgroup');header.setAttribute('role','row');
  table.append(n('caption','Independent captured values · supplied changes'));
  for(const [index,text] of ['Measure','Baseline','Selection','Change'].entries()){const th=n('th',text);th.scope='col';th.id=prefix+'-column-'+index;th.setAttribute('role','columnheader');header.append(th)}head.append(header);
  const snapshots=[comparison.baseline,comparison.selection],rows=snapshots.map(snapshot=>snapshot.rows[0]);
  for(const [index,key] of display.fields.entries()){
    const fields=snapshots.map(snapshot=>snapshot.fields.find(field=>field.key===key)),field=fields.find(Boolean),tr=n('tr');tr.dataset.field=key;
    tr.setAttribute('role','row');const name=n('th',field.label);name.scope='row';name.id=prefix+'-row-'+index;name.setAttribute('role','rowheader');tr.append(name);
    const values=snapshots.map((snapshot,index)=>fields[index]?measureValue(snapshot,rows[index],key):null);
    const comparable=compatibleFields(...fields)&&comparison.changes[key]?.state==='derived'&&!['diagnostic','incompatible'].includes(comparison.state);
    const raw=values.map(value=>value?.coordinate!==null&&value?.coordinate!==undefined?value.exact:null),domain=comparable?signedScale(raw):null;
    const plot=domain?.known&&raw.filter(value=>value!==null).every(value=>typeof value==='string'&&value.length<=256);
    if(plot){const axis=n('small','Shared scale','measure-axis-label');axis.title=domain.min+' → '+domain.max+(field.unit?' '+field.unit:'')+' · signed zero axis';name.append(axis)}
    for(let side=0;side<2;side++){
      const cell=n('td');cell.dataset.side=side?'selection':'baseline';cell.setAttribute('role','cell');cell.setAttribute('headers',name.id+' '+prefix+'-column-'+(side+1));
      const label=n('span',side?'Selection':'Baseline','measure-side-label');label.setAttribute('aria-hidden','true');cell.append(label,valueCell(fields[side],rows[side]));
      if(plot){
        const track=n('div',undefined,'measure-comparison-track');track.dataset.min=domain.min;track.dataset.max=domain.max;track.style.setProperty('--measure-zero',domain.zero+'%');track.setAttribute('aria-hidden','true');
        const bar=domain.bars[side];if(bar){const fill=n('span',undefined,'measure-comparison-fill');fill.dataset.side=side?'selection':'baseline';fill.style.left=bar.start+'%';fill.style.width=bar.width+'%';track.append(fill)}else track.dataset.unavailable='true';cell.querySelector('.measure-value').after(track);
      }
      tr.append(cell);
    }
    const delta=n('td');delta.dataset.side='change';delta.setAttribute('role','cell');delta.setAttribute('headers',name.id+' '+prefix+'-column-3');const label=n('span','Change','measure-side-label');label.setAttribute('aria-hidden','true');delta.append(label,valueCell(field,null,{delta:comparison.changes[key]||{state:'unavailable',value:null,reason:'No supplied change.'}}));tr.append(delta);body.append(tr);
  }
  table.append(head,body);viewport.append(table);return viewport;
}

class Supplied extends (globalThis.HTMLElement||class {}){
  static get observedAttributes(){return ['snapshot','comparison','display']}
  connectedCallback(){for(const key of ['snapshot','comparison','display'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
  attributeChangedCallback(){this.render()}
  read(key){return Object.hasOwn(this,'_'+key)?copy(this['_'+key]):this.hasAttribute(key)?JSON.parse(this.getAttribute(key)):undefined}
  setValue(key,value){try{this['_'+key]=copy(value)}catch{this['_'+key]=null}this.render()}
  set snapshot(value){this.setValue('snapshot',value)}get snapshot(){return this.read('snapshot')}
  set comparison(value){this.setValue('comparison',value)}get comparison(){return this.read('comparison')}
  set display(value){this.setValue('display',value)}get display(){return this.read('display')}
  refuse(reasons){const status=n('p',reasons.join('; '),'measure-unavailable');status.setAttribute('role','status');this.replaceChildren(status)}
}

class Summary extends Supplied{
  render(){try{
    const snapshot=this.snapshot,valid=validateSnapshot(snapshot);if(!valid.ok){this.refuse(valid.reasons);return}if(snapshot.shape!=='summary'){this.refuse(['A population summary snapshot is required.']);return}
    const display=options(this.display,snapshot.fields),root=n('goo-chart-frame',undefined,'measure-frame');root.data={title:display.title||'Population summary',state:snapshot.quality.extent==='unknown'||snapshot.quality.value==='unknown'?'unknown':snapshot.quality.extent==='partial'?'partial':snapshot.rows[0].values.count==='0'?'no_data':'observed'};
    root.append(provenance(snapshot),summaryBody(snapshot,display));this.replaceChildren(root);
  }catch(error){this.refuse([error instanceof SyntaxError?'Invalid summary JSON.':error.message||'Summary unavailable.'])}}
}

class Comparison extends Supplied{
  render(){try{
    const comparison=this.comparison,valid=validateComparison(comparison);if(!valid.ok){this.refuse(valid.reasons);return}
    const fields=[...new Map([...comparison.baseline.fields,...comparison.selection.fields].map(field=>[field.key,field])).values()],display=options(this.display,fields,{comparison:true});
    const root=n('goo-chart-frame',undefined,'measure-frame'),provenances=n('div',undefined,'measure-comparison-provenance');root.data={title:display.title||'Population comparison',state:comparison.state==='diagnostic'?'unknown':comparison.state};
    provenances.append(provenance(comparison.baseline,'Baseline'),provenance(comparison.selection,'Selection'));root.append(provenances,comparisonBody(comparison,display));this.replaceChildren(root);
  }catch(error){this.refuse([error instanceof SyntaxError?'Invalid comparison JSON.':error.message||'Comparison unavailable.'])}}
}

if(globalThis.customElements&&globalThis.HTMLElement){
  if(!customElements.get('goo-measure-summary'))customElements.define('goo-measure-summary',Summary);
  if(!customElements.get('goo-measure-comparison'))customElements.define('goo-measure-comparison',Comparison);
}
