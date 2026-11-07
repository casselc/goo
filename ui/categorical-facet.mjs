import {renderRefusal} from './source-refusal-view.mjs';
/* Passive supplied counted-values surface. The application owns reads and scope. */
import {node,actionVisible} from './components.mjs';
import {declaration,selectionReason} from './facet-declaration.mjs';
import {countedState,projection,valueKey} from './categorical-pair.mjs';
import {cell} from './comparison-visual.mjs';
const global=globalThis,label=value=>value===null?'Missing key':value===''?'Present empty string':JSON.stringify(value);
export function model(value){
 if(!value||typeof value!=='object'||Array.isArray(value)||Object.keys(value).some(key=>!['declaration','values','baseline','baselineLabel','disabled','hasStaged','basisLabel'].includes(key))||value.disabled!==undefined&&typeof value.disabled!=='boolean'||value.hasStaged!==undefined&&typeof value.hasStaged!=='boolean')throw Error('Supply a categorical facet model.');
 for(const key of ['basisLabel','baselineLabel'])if(value[key]!==undefined&&(typeof value[key]!=='string'||!value[key].isWellFormed()||!value[key].trim()||new TextEncoder().encode(value[key]).length>120||/\p{Cc}/u.test(value[key])))throw Error('Supply a bounded literal facet basis label.');
 if(value.baselineLabel!==undefined&&!Object.hasOwn(value,'baseline'))throw Error('A baseline label requires a supplied baseline state.');
 const d=declaration(value.declaration,value.declaration?.scope?.signal),data=countedState(value.values,d.field);
 return {...structuredClone(value),declaration:d,values:data,...(Object.hasOwn(value,'baseline')?{baseline:countedState(value.baseline,d.field)}:{})};
}
// Keep both the field prefix and terminal identifier visible; exact text stays supplied.
export function loadedValuePreview(text){const points=Array.from(text);return points.length>160?points.slice(0,104).join('')+'…'+points.slice(-55).join(''):text;}
const adopt=value=>{try{return model(value)}catch{return null}};
export function loadedMatchesText({visible,total},maximum=1000){if(!Number.isInteger(visible)||!Number.isInteger(total)||visible<0||visible>total||total>maximum||![1000,2000].includes(maximum))throw Error('Supply bounded local facet match counts.');return visible+' of '+total+' loaded values';}
class CategoricalFacet extends (global.HTMLElement||class{}){
 static get observedAttributes(){return ['data']}
 set data(value){this._data=adopt(value);if(this.isConnected)this.render()}
 get data(){if(this._data!==undefined)return structuredClone(this._data);try{const raw=this.getAttribute('data');return raw&&raw.length<=8*1024*1024?model(JSON.parse(raw)):null}catch{return null}}
 set interaction(value){this._interaction=value;if(this.isConnected)this.render()}
 get interaction(){return this._interaction}
 publish(value){
  let data=null,interaction=null;
  try{if(!value||typeof value!=='object'||Array.isArray(value)||Object.keys(value).some(key=>!['data','interaction'].includes(key))||!Object.hasOwn(value,'data')||!Object.hasOwn(value,'interaction'))throw Error('Supply data and interaction together.');
   const supplied=value.interaction;if(supplied!=null&&(typeof supplied!=='object'||Array.isArray(supplied)||supplied.current!==undefined&&typeof supplied.current!=='function'))throw Error('Supply a current-owner callback.');data=model(value.data);interaction=supplied;
  }catch{}
  this._data=data;this._interaction=interaction;if(this.isConnected)this.render();
 }
 connectedCallback(){for(const key of ['data','interaction'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];if(key==='data')this._data=adopt(value);else this._interaction=value}this.render()}
 attributeChangedCallback(){this._data=undefined;if(this.isConnected)this.render()}
 render(){
  const active=global.document?.activeElement,focus=this.contains(active)?active.getAttribute('aria-label'):null;
  if(focus)this._pendingFocus=focus;else if(active&&active!==global.document?.body&&!this.contains(active))this._pendingFocus=null;
  const previousSearch=this.querySelector('[aria-label="Search loaded facet values"]');if(previousSearch)this._caret={start:previousSearch.selectionStart,end:previousSearch.selectionEnd,direction:previousSearch.selectionDirection};
  const previousDetails=this.querySelector('.categorical-facet-receipts')?.open,epoch=this._epoch=(this._epoch||0)+1,data=this.data;
  this.replaceChildren();this.classList.add('filter-counted-values');
  if(!data){this._find='';this._searchKey=null;this._pendingFocus=null;this.dataset.state='unavailable';this.append(node('p','Facet unavailable.'));return}
  const d=data.declaration,v=data.values,paired=Object.hasOwn(data,'baseline'),baseline=data.baseline,pairedValues=paired?projection(v,baseline,d.field):null,loading=v.state==='loading'||baseline?.state==='loading',ready=paired?[v,baseline].some(side=>side.state==='ready'):v.state==='ready';
  const searchKey=JSON.stringify([d.scope.signal,d.field.kind,d.field.key,v.basis||null,v.readAt||null,...(paired?[baseline.basis||null,baseline.readAt||null]:[])]),sameRead=this._searchKey===searchKey;
  if(!sameRead){this._find='';this._page=0;this._caret=null;}this._searchKey=searchKey;const name=d.field.kind+':'+d.field.key,baselineLabel=data.baselineLabel||'Fixed baseline';this.dataset.state=paired?(pairedValues.complete?'ready':[v,baseline].some(side=>side.state==='ready')?'partial':loading?'loading':[v,baseline].some(side=>side.state==='error')?'error':'unloaded'):v.state;
  this.classList.toggle('categorical-facet-paired',paired);
  const current=()=>epoch===this._epoch&&this.isConnected&&!data.disabled&&this._interaction?.current?.()!==false,emit=(type,detail={})=>{if(current())this.dispatchEvent(new CustomEvent(type,{bubbles:true,detail}))};
  const heading=node('header',undefined,'categorical-facet-header'),actions=node('div',undefined,'categorical-facet-actions'),read=node('button','Read values'),clear=node('button','Clear staged');
  heading.append(node('h4',d.display.title||name),actions);read.type=clear.type='button';read.setAttribute('aria-label','Read values');clear.setAttribute('aria-label','Clear staged facet');read.disabled=!!data.disabled||loading;clear.disabled=!!data.disabled||!data.hasStaged;
  read.onclick=()=>{if(!read.disabled)emit('goo:facet-load')};clear.onclick=()=>{if(!clear.disabled)emit('goo:facet-clear')};actions.append(read,clear);this.append(heading,node('small',(d.display.title?name+' · ':'')+(data.basisLabel||(paired?'Both window values':'Applied fixed scope')),'categorical-facet-scope'));
  const restoreFocus=()=>{if(!this._pendingFocus)return;const target=[...this.querySelectorAll('[aria-label]')].find(el=>el.getAttribute('aria-label')===this._pendingFocus&&!el.disabled)||(!read.disabled?read:null);if(target){target.focus({preventScroll:true});this._pendingFocus=null}};
  if(paired)for(const [side,sideLabel,state]of [['selected','Selected',v],['baseline',baselineLabel,baseline]])if(state.state==='error'){
   const status=node('div',undefined,'categorical-facet-side-status'),retry=node('button','Retry '+(side==='selected'?'selected':baselineLabel.toLowerCase())+' values');status.setAttribute('role','status');if(state.refusal)renderRefusal(status,state,{prefix:sideLabel+' values unavailable · '});else status.append(node('span',sideLabel+' values unavailable · '+(state.error||'Read failed.')));retry.type='button';retry.setAttribute('aria-label','Retry '+side+' values');retry.disabled=!!data.disabled||loading;retry.onclick=()=>{if(!retry.disabled)emit('goo:facet-retry',{side})};status.append(retry);this.append(status);
  }
  if(!ready){const status=node('p',loading?'Reading counted values…':paired&&[v,baseline].some(side=>side.state==='error')?'Read available values again, or retry the failed window.':v.state==='error'?v.error:'Read values explicitly. Inspection rows are not facet totals.');status.setAttribute('role','status');if(v.state==='error'&&v.refusal)renderRefusal(status,v);this.append(status);if(this._pendingFocus&&loading){status.tabIndex=-1;status.focus({preventScroll:true})}else restoreFocus();return}
  this.append(node('small',paired?'Selected '+(v.state==='ready'?v.total:'unavailable')+' · '+baselineLabel+' '+(baseline.state==='ready'?baseline.total:'unavailable')+' · independent value reads; upstream coverage unknown':v.total+' matching records · upstream coverage unknown','categorical-facet-count'));
  if(paired&&!pairedValues.complete&&v.state!=='error'&&baseline.state!=='error'){const pending=node('p','The other window is '+(loading?'loading':'not loaded')+'; its counts are unknown.');pending.setAttribute('role','status');this.append(pending);}
  if(paired&&pairedValues.complete&&!pairedValues.rows.length)this.append(node('p','No values observed in either window.'));
  const tools=node('div',undefined,'categorical-facet-tools'),search=node('input'),matches=node('span',undefined,'categorical-facet-matches'),details=node('details',undefined,'categorical-facet-receipts'),summary=node('summary','Read details'),list=node('goo-facet-list',undefined,'categorical-facet-values');
  search.type='search';search.placeholder='Find loaded values';search.setAttribute('aria-label','Search loaded facet values');search.value=this._find||'';matches.setAttribute('aria-label','Loaded facet matches');matches.setAttribute('role','status');matches.setAttribute('aria-live','polite');matches.setAttribute('aria-atomic','true');summary.setAttribute('aria-label','Read details');details.append(summary);details.open=!!sameRead&&!!previousDetails;
  for(const [sideLabel,state]of paired?[['Selected',v],[baselineLabel,baseline]]:[['',v]])if(state.state==='ready'){
   if(paired)details.append(node('p',sideLabel+' · '+(state.label||'Supplied grouping partition')));
   for(const kind of ['count','groups'])if(state.reads?.[kind])details.append(node('p',(sideLabel?sideLabel+' · ':'')+kind+' · read '+state.reads[kind].readAt+' · '+state.reads[kind].basis));
  }
  tools.append(search,matches,details);this.append(tools);
  if(paired){const columns=node('div',undefined,'categorical-facet-pair-columns'),counts=node('div',undefined,'categorical-facet-pair-counts');columns.setAttribute('aria-hidden','true');columns.append(node('span'));counts.append(node('span','Selected'),node('span',baselineLabel));columns.append(counts,node('span'));this.append(columns);}
  const pager=node('nav',undefined,'categorical-facet-pages'),previous=node('button','Previous values'),next=node('button','Next values'),range=node('span');pager.setAttribute('aria-label','Loaded facet pages');previous.type=next.type='button';previous.setAttribute('aria-label','Previous facet page');next.setAttribute('aria-label','Next facet page');range.setAttribute('aria-label','Shown facet values');range.setAttribute('aria-live','polite');pager.append(previous,range,next);this.append(pager,list);
  const items=paired?pairedValues.rows:v.items,searchable=items.map(item=>({item,text:label(item.value).toLocaleLowerCase()}));let rowGeneration=0;
  const drawRows=()=>{const generation=++rowGeneration,query=search.value.toLocaleLowerCase(),matched=searchable.filter(entry=>entry.text.includes(query));this._page=Math.max(0,Math.min(this._page||0,Math.max(0,Math.ceil(matched.length/100)-1)));const start=this._page*100;list.replaceChildren();list.scrollTop=0;matches.textContent=loadedMatchesText({visible:matched.length,total:items.length},paired?2000:1000);range.textContent=matched.length?'Showing '+(start+1)+'–'+Math.min(start+100,matched.length)+' of '+matched.length+' matches':'Showing 0 of 0 matches';previous.hidden=next.hidden=matched.length<=100;previous.disabled=start===0;next.disabled=start+100>=matched.length;
  for(const {item} of matched.slice(start,start+100)){const row=node('div',undefined,'categorical-facet-row'),text=label(item.value),value=node('span',loadedValuePreview(text),'categorical-facet-value'),controls=node('div',undefined,'categorical-facet-row-actions');value.title=text;row.dataset.facetItem='';row.dataset.facetText=text;row.dataset.facetValueKey=valueKey(item.value);
   if(paired){const counts=node('div',undefined,'categorical-facet-pair-counts');for(const [side,key,sideLabel]of [['selection','selected','Selected'],['baseline','baseline',baselineLabel]]){const supplied=item[key],count=cell(counts,supplied.count,side,{max:pairedValues.max,text:supplied.count??'—',geometry:pairedValues.complete});count.dataset.count=supplied.count??'unknown';count.dataset.presence=supplied.state;count.setAttribute('role','group');count.setAttribute('aria-label',sideLabel+' count for '+text);count.title=supplied.state==='absent'?'Absent from this returned grouping; upstream coverage unknown.':supplied.state==='unknown'?sideLabel+' values unavailable.':'Count in this returned grouping; upstream coverage unknown.';count.prepend(node('small',sideLabel,'categorical-facet-side-label'));}row.append(value,counts,controls);}
   else {controls.append(node('span',item.count,'categorical-facet-value-count'));row.append(value,controls);}
   for(const mode of ['include','exclude']){const button=node('button',mode==='include'?'Include':'Exclude');button.type='button';button.setAttribute('aria-label',(mode==='include'?'Include ':'Exclude ')+name+' '+text);const reason=selectionReason(item.value,d.scope.signal,d.field);button.disabled=!!data.disabled||!!reason;if(reason)button.title=reason;button.onclick=()=>{if(generation===rowGeneration&&button.isConnected&&!button.disabled)emit('goo:facet-stage',{value:item.value,mode})};controls.append(button)}
   list.append(row);
  }
  for(const [button,change]of [[previous,-1],[next,1]])button.onclick=()=>{if(!localCurrent()||generation!==rowGeneration||!button.isConnected||button.disabled||button.hidden||!actionVisible(button)||button.closest?.('[inert]'))return;this._page+=change;drawRows();(button.disabled?(change>0?previous:next):button).focus({preventScroll:true})};
  };
  const localCurrent=()=>epoch===this._epoch&&this.isConnected&&search.isConnected&&this._interaction?.current?.()!==false;
  search.oninput=()=>{if(!localCurrent())return;this._find=search.value;this._page=0;drawRows()};
  search.onblur=()=>{if(localCurrent())this._caret={start:search.selectionStart,end:search.selectionEnd,direction:search.selectionDirection}};
  drawRows();restoreFocus();if(this._caret&&sameRead)search.setSelectionRange?.(this._caret.start,this._caret.end,this._caret.direction);

 }
}
if(global.customElements&&!global.customElements.get('goo-categorical-facet'))global.customElements.define('goo-categorical-facet',CategoricalFacet);
export {CategoricalFacet};
