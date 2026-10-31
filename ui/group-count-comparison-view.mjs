/* Display-only aligned groups. The host owns acquisition, compatibility and scope. */
import {node} from './components.mjs';
import {cell} from './comparison-visual.mjs';
import {asGroupComparison,projection,categoryPreview} from './group-count-comparison.mjs';
import {groupSideIntent} from './group-side-staging.mjs';
const global=globalThis;
function config(value={}){
 if(!value||typeof value!=='object'||Array.isArray(value)||![Object.prototype,null].includes(Object.getPrototypeOf(value)))throw TypeError('Use checked group display options.');
 const result={};for(const key of Reflect.ownKeys(value)){const d=Object.getOwnPropertyDescriptor(value,key);if(typeof key!=='string'||!['order','title'].includes(key)||!d.enumerable||!Object.hasOwn(d,'value'))throw TypeError('Use plain checked group display options.');result[key]=d.value}
 if(result.order!==undefined&&!['','change','label'].includes(result.order)||result.title!==undefined&&(typeof result.title!=='string'||!result.title.isWellFormed()||new TextEncoder().encode(result.title).length>120||/\p{Cc}/u.test(result.title)||result.title!==''&&!result.title.trim()))throw TypeError('Use checked group display options.');
 return {order:result.order||'change',title:result.title||'Count change by group'};
}
export function groupStatus(model,{first=0,last=0}={}){
 const message=model.total===0?(model.diagnostic?'No supplied groups available in this diagnostic comparison.':'No groups observed in either supplied population.'):model.matched===0?'No matching groups. '+model.total+' groups in the held comparison.':'Showing '+first+'–'+last+' of '+model.matched+' matching groups · '+model.total+' total groups.';
 return message+(model.diagnostic?' Source reads disagree; differences and geometry unavailable.':model.total&&model.matched?' Shared count scale: 0–'+model.domainMax+' records.':'');
}
class GroupCountComparison extends (global.HTMLElement||class {}){
 static get observedAttributes(){return ['comparison','display']}
 connectedCallback(){for(const key of ['comparison','display','interaction'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
 attributeChangedCallback(name){const raw=this.getAttribute(name);if(raw===null){if(name==='comparison'){this._comparison=null;this._comparisonError=''}else{this._display=config();this._displayError=''}if(this.isConnected)this.render();return}try{if(raw.length>18*1024*1024)throw Error('Supplied comparison is too large.');this[name]=JSON.parse(raw)}catch(error){if(name==='comparison'){this._comparison=null;this._comparisonError=error.message}else this._displayError=error.message;if(this.isConnected)this.render()}}
 set comparison(value){try{this._comparison=asGroupComparison(value);this._comparisonError='';this._page=0;if(this.isConnected)this.render()}catch(error){this._comparison=null;this._comparisonError=error.message;if(this.isConnected)this.render()}}
 get comparison(){return this._comparison}
 set display(value){try{const next=config(value),hadError=!!this._displayError,changed=next.order!==this._display?.order||next.title!==this._display?.title;this._display=next;this._displayError='';if((changed||hadError)&&this.isConnected)this.render()}catch(error){this._displayError=error.message;if(this.isConnected)this.render()}}
 get display(){return {...(this._display||config())}}
 get presentation(){return {query:this._query||'',page:this._page||0,order:this.display.order}}
 disconnectedCallback(){this._actionEpoch=(this._actionEpoch||0)+1;this._interaction=null;}
 set interaction(value){
  this._actionEpoch=(this._actionEpoch||0)+1;this._interaction=null;
  try{if(value!==null&&value!==undefined){
   if(!value||typeof value!=='object'||Array.isArray(value)||![Object.prototype,null].includes(Object.getPrototypeOf(value))||Reflect.ownKeys(value).length!==2)throw TypeError('Group staging needs an explicit host binding.');
   const out={};for(const key of ['current','stage']){const d=Object.getOwnPropertyDescriptor(value,key);if(!d?.enumerable||!Object.hasOwn(d,'value')||typeof d.value!=='function')throw TypeError('Group staging needs host current and stage callbacks.');out[key]=d.value}this._interaction=Object.freeze(out);
  }}finally{if(this.isConnected)this.render()}
 }
 get interaction(){return this._interaction||null}
 refreshInteraction(){for(const entry of this._stageButtons||[])entry.button.disabled=!this.stageCurrent(entry.binding,entry.epoch,entry.comparison);}
 stageCurrent(binding,epoch,comparison){try{return !!binding&&binding===this._interaction&&epoch===this._actionEpoch&&comparison===this._comparison&&this.isConnected&&this.getClientRects().length>0&&!this.hidden&&binding.current()===true}catch{return false}}
 stageButton(target,row,side,comparison,binding,epoch){
  const cell=row?.[side];if(!binding||comparison.state==='diagnostic'||comparison.baselineAccounting!=='coherent'||comparison.selectionAccounting!=='coherent'||cell?.presence!=='present'||cell.state!=='derived'||cell.value==='0')return;
  const button=node('button',row[side].value+' records','group-count-stage'),preview=categoryPreview(row.value).preview;button.type='button';button.dataset.groupStageSide=side;button.dataset.groupStageKey=row.key;
  button.setAttribute('aria-label','Explore '+(side==='baseline'?'baseline':'current')+' group: '+preview);button.title='Stage this '+(side==='baseline'?'baseline':'current')+' fixed question. Run reads independently.';
  button.disabled=!this.stageCurrent(binding,epoch,comparison);button.onclick=()=>{if(!this.stageCurrent(binding,epoch,comparison)||!button.isConnected||!button.getClientRects().length)return;let intent;try{intent=groupSideIntent(comparison,side,row.key)}catch{return}binding.stage(intent,button)};
  target.replaceChildren(button);this._stageButtons.push({button,binding,epoch,comparison});
 }

 refuse(message){this.replaceChildren(node('p',message||'Group comparison unavailable.'));this.firstChild.setAttribute('role','status')}
 render(){
  this._actionEpoch=(this._actionEpoch||0)+1;
  if(this._comparisonError||this._displayError||!this._comparison){this.refuse(this._comparisonError||this._displayError||'Group comparison unavailable.');return}
  const display=this._display||config();this._display=display;this._query??='';this._page??=0;
  const root=node('section',undefined,'group-count-comparison'),title=node('h5',display.title),controls=node('div',undefined,'group-count-controls'),find=node('input'),order=node('select'),findLabel=node('label','Find groups '),orderLabel=node('label','Order '),body=node('div'),status=node('p'),nav=node('div',undefined,'group-count-controls'),previous=node('button','Previous compared groups'),next=node('button','Next compared groups');
  find.type='search';find.value=this._query;find.setAttribute('aria-label','Find compared groups');find.maxLength=4096;findLabel.append(find);order.setAttribute('aria-label','Group comparison order');for(const [value,text]of [['change','Count change'],['label','Group name']]){const option=node('option',text);option.value=value;order.append(option)}order.value=display.order;orderLabel.append(order);controls.append(findLabel,orderLabel);status.setAttribute('role','status');nav.append(previous,next);root.append(title,controls,body,status,nav);this.replaceChildren(root);
  const draw=()=>{
   const epoch=this._actionEpoch=(this._actionEpoch||0)+1,binding=this._interaction,comparison=this._comparison;this._stageButtons=[];
   const model=projection(this._comparison,{query:this._query,order:this._display.order,page:this._page});this._page=model.page;this.dataset.domainMax=model.domainMax??'unavailable';this.dataset.groupCount=String(model.total);this.dataset.visibleCount=String(model.rows.length);this.dataset.page=String(model.page);this.dataset.state=this._comparison.state;body.replaceChildren();
   const viewport=node('goo-table'),table=node('table'),head=node('thead'),headers=node('tr'),rows=node('tbody');viewport.setAttribute('aria-label','Compared group counts');table.className='group-count-table';table.setAttribute('aria-label','Compared group counts');for(const label of [this._comparison.field.kind+':'+this._comparison.field.key,'Baseline','Current held','Change']){const th=node('th',label);th.scope='col';headers.append(th)}head.append(headers);table.append(head,rows);viewport.append(table);body.append(viewport);
   for(const row of model.rows){const tr=node('tr'),label=node('th'),category=categoryPreview(row.value);label.scope='row';label.className='group-count-category';tr.dataset.groupKey=row.key;label.title=category.label;if(category.long){const exact=node('details'),summary=node('summary',category.preview),value=node('code',category.label);summary.setAttribute('aria-label','Show exact group value: '+category.preview);exact.append(summary,value);label.append(exact)}else label.textContent=category.label;tr.append(label);
    for(const side of ['baseline','selection']){const td=node('td');td.dataset.side=side;const value=row[side].value;td.dataset.presence=row[side].presence;if(model.diagnostic){td.append(node('span',value===null?'Unavailable':value+' records','group-count-value'));td.title=row[side].reason||'Source reads disagree; geometry is unavailable.'}else{const bar=cell(td,value,side,{max:model.domainMax,unit:'records',text:value+' records'});const primary=bar.querySelector('.comparison-value');primary.classList.add('group-count-value');this.stageButton(primary,row,side,comparison,binding,epoch);if(row[side].presence==='absent'){td.title='Not observed in '+(side==='baseline'?'baseline':'current')+' supplied captured groups; aligned count is zero.';td.setAttribute('aria-label','0 records. '+td.title)}}tr.append(td)}
    const delta=node('td');delta.dataset.side='change';const value=row.change.value;delta.append(node('span',value===null?'Unavailable':(BigInt(value)>0n?'+':'')+value+' records','group-count-value'));delta.title=row.change.reason||'Exact current held minus baseline count';tr.append(delta);rows.append(tr);
   }
   const shown=model.rows.length,first=shown?model.page*model.pageSize+1:0,last=shown?first+shown-1:0;
   status.textContent=groupStatus(model,{first,last});previous.disabled=model.page===0;next.disabled=(model.page+1)*model.pageSize>=model.matched;
  };
  find.oninput=()=>{this._query=find.value;this._page=0;draw()};order.onchange=()=>{this._display={...this._display,order:order.value};this._page=0;draw();this.dispatchEvent(new CustomEvent('goo:groupdisplaychange',{detail:{order:order.value},bubbles:true}))};previous.onclick=()=>{this._page=Math.max(0,this._page-1);draw()};next.onclick=()=>{this._page++;draw()};draw();
 }
}
if(global.customElements&&global.HTMLElement&&!global.customElements.get('goo-group-count-comparison'))global.customElements.define('goo-group-count-comparison',GroupCountComparison);
export {GroupCountComparison};
