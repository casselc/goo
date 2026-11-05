/* Passive supplied counted-values surface. The application owns reads and scope. */
import {node} from './components.mjs';
import {declaration,selectionReason} from './facet-declaration.mjs';
const global=globalThis,label=value=>value===null?'Missing key':value===''?'Present empty string':JSON.stringify(value);
export function model(value){
 if(!value||typeof value!=='object'||Array.isArray(value)||Object.keys(value).some(key=>!['declaration','values','disabled','hasStaged','basisLabel'].includes(key))||value.disabled!==undefined&&typeof value.disabled!=='boolean'||value.hasStaged!==undefined&&typeof value.hasStaged!=='boolean')throw Error('Supply a categorical facet model.');if(value.basisLabel!==undefined&&(typeof value.basisLabel!=='string'||!value.basisLabel.isWellFormed()||!value.basisLabel.trim()||new TextEncoder().encode(value.basisLabel).length>120||/\p{Cc}/u.test(value.basisLabel)))throw Error('Supply a bounded literal facet basis label.');const d=declaration(value.declaration,value.declaration?.scope?.signal),data=value.values||{state:'unloaded'};
 if(Object.keys(data).some(key=>!['state','field','items','total','readAt','basis','reads','label','error'].includes(key))||!['unloaded','loading','error','ready'].includes(data.state))throw Error('Supply a facet state.');
 if(data.state==='ready'){const uint=v=>typeof v==='string'&&/^(0|[1-9][0-9]*)$/.test(v)&&v.length<=20&&BigInt(v)<=18446744073709551615n;if(data.field?.kind!==d.field.kind||data.field?.key!==d.field.key||!Array.isArray(data.items)||data.items.length>1000||!uint(data.total)||!Number.isFinite(Date.parse(data.readAt))||typeof data.basis!=='string'||!data.basis)throw Error('Supply matching bounded counted facet values.');let sum=0n;const seen=new Set();for(const item of data.items){if(!(item.value===null||typeof item.value==='string'&&item.value.isWellFormed())||!uint(item.count)||BigInt(item.count)===0n||seen.has(item.value))throw Error('Supply unique positive categorical counts.');seen.add(item.value);sum+=BigInt(item.count)}if(sum!==BigInt(data.total))throw Error('Supplied facet counts must partition their supplied total.');for(const name of ['count','groups']){const receipt=data.reads?.[name];if(!receipt||!Number.isFinite(Date.parse(receipt.readAt))||typeof receipt.basis!=='string'||!/^[a-f0-9]{64}$/.test(receipt.basis))throw Error('Supply both independent read receipts.');}}
 return {...structuredClone(value),declaration:d,values:structuredClone(data)};
}
export function loadedMatchesText({visible,total}){if(!Number.isInteger(visible)||!Number.isInteger(total)||visible<0||visible>total||total>1000)throw Error('Supply bounded local facet match counts.');return visible+' of '+total+' loaded values';}
class CategoricalFacet extends (global.HTMLElement||class{}){
 static get observedAttributes(){return ['data']}
 set data(value){try{this._data=model(value)}catch{this._data=null}if(this.isConnected)this.render()}
 get data(){if(this._data!==undefined)return structuredClone(this._data);try{const raw=this.getAttribute('data');return raw&&raw.length<=8*1024*1024?model(JSON.parse(raw)):null}catch{return null}}
 set interaction(value){this._interaction=value;if(this.isConnected)this.render()}
 get interaction(){return this._interaction}
 connectedCallback(){for(const key of ['data','interaction'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
 attributeChangedCallback(){this._data=undefined;if(this.isConnected)this.render()}
 render(){
  const active=global.document?.activeElement,focus=this.contains(active)?active.getAttribute('aria-label'):null;
  if(focus)this._pendingFocus=focus;else if(active&&active!==global.document?.body&&!this.contains(active))this._pendingFocus=null;
  const previousDetails=this.querySelector('.categorical-facet-receipts')?.open,epoch=this._epoch=(this._epoch||0)+1,data=this.data;
  this.replaceChildren();this.classList.add('filter-counted-values');
  if(!data){this._find='';this._searchKey=null;this._pendingFocus=null;this.dataset.state='unavailable';this.append(node('p','Facet unavailable.'));return}
  const d=data.declaration,v=data.values,searchKey=JSON.stringify([d.scope.signal,d.field.kind,d.field.key,v.basis||null,v.readAt||null]),sameRead=this._searchKey===searchKey;
  if(!sameRead)this._find='';this._searchKey=searchKey;const name=d.field.kind+':'+d.field.key;this.dataset.state=v.state;
  const current=()=>epoch===this._epoch&&this.isConnected&&!data.disabled&&this._interaction?.current?.()!==false,emit=(type,detail={})=>{if(current())this.dispatchEvent(new CustomEvent(type,{bubbles:true,detail}))};
  const heading=node('header',undefined,'categorical-facet-header'),actions=node('div',undefined,'categorical-facet-actions'),read=node('button','Read values'),clear=node('button','Clear staged');
  heading.append(node('h4',d.display.title||name),actions);read.type=clear.type='button';read.setAttribute('aria-label','Read values');clear.setAttribute('aria-label','Clear staged facet');read.disabled=!!data.disabled||v.state==='loading';clear.disabled=!!data.disabled||!data.hasStaged;
  read.onclick=()=>{if(!read.disabled)emit('goo:facet-load')};clear.onclick=()=>{if(!clear.disabled)emit('goo:facet-clear')};actions.append(read,clear);this.append(heading,node('small',(d.display.title?name+' · ':'')+(data.basisLabel||'Applied fixed scope'),'categorical-facet-scope'));
  const restoreFocus=()=>{if(!this._pendingFocus)return;const target=[...this.querySelectorAll('[aria-label]')].find(el=>el.getAttribute('aria-label')===this._pendingFocus&&!el.disabled)||(!read.disabled?read:null);if(target){target.focus({preventScroll:true});this._pendingFocus=null}};
  if(v.state!=='ready'){const status=node('p',v.state==='loading'?'Reading counted values…':v.state==='error'?v.error:'Read values explicitly. Inspection rows are not facet totals.');status.setAttribute('role','status');this.append(status);if(this._pendingFocus&&v.state==='loading'){status.tabIndex=-1;status.focus({preventScroll:true})}else restoreFocus();return}
  this.append(node('small',v.total+' matching records · upstream coverage unknown','categorical-facet-count'));
  const tools=node('div',undefined,'categorical-facet-tools'),search=node('input'),matches=node('span',undefined,'categorical-facet-matches'),details=node('details',undefined,'categorical-facet-receipts'),summary=node('summary','Read details'),list=node('goo-facet-list',undefined,'categorical-facet-values');
  search.type='search';search.placeholder='Find loaded values';search.setAttribute('aria-label','Search loaded facet values');search.value=this._find||'';matches.setAttribute('aria-label','Loaded facet matches');matches.setAttribute('role','status');matches.setAttribute('aria-live','polite');matches.setAttribute('aria-atomic','true');summary.setAttribute('aria-label','Read details');details.append(summary);details.open=!!sameRead&&!!previousDetails;
  for(const kind of ['count','groups'])if(v.reads?.[kind])details.append(node('p',kind+' · read '+v.reads[kind].readAt+' · '+v.reads[kind].basis));
  tools.append(search,matches,details);this.append(tools,list);
  for(const item of v.items){const row=node('div',undefined,'categorical-facet-row'),text=label(item.value),value=node('span',text.length>160?text.slice(0,157)+'…':text,'categorical-facet-value'),controls=node('div',undefined,'categorical-facet-row-actions'),count=node('span',item.count,'categorical-facet-value-count');value.title=text;row.dataset.facetItem='';row.dataset.facetText=text;controls.append(count);
   for(const mode of ['include','exclude']){const button=node('button',mode==='include'?'Include':'Exclude');button.type='button';button.setAttribute('aria-label',(mode==='include'?'Include ':'Exclude ')+name+' '+text);const reason=selectionReason(item.value,d.scope.signal,d.field);button.disabled=!!data.disabled||!!reason;if(reason)button.title=reason;button.onclick=()=>{if(!button.disabled)emit('goo:facet-stage',{value:item.value,mode})};controls.append(button)}
   row.append(value,controls);list.append(row);
  }
  const find=()=>{list.query=search.value;const projected=list.refresh?.()||{visible:v.items.length,total:v.items.length};matches.textContent=loadedMatchesText(projected)};
  search.oninput=()=>{this._find=search.value;find()};find();restoreFocus();

 }
}
if(global.customElements&&!global.customElements.get('goo-categorical-facet'))global.customElements.define('goo-categorical-facet',CategoricalFacet);
export {CategoricalFacet};
