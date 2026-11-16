/* Supplied reading state only; the application owns panel visibility and actions. */
import * as UI from './components.mjs';
import {adopt,topics,pages} from './guidance-content.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
export {initial,adopt,withContext,transition,contexts,topics,pages} from './guidance-content.mjs';
const n=UI.node;
if(globalThis.customElements&&globalThis.HTMLElement){
 class Guidance extends globalThis.HTMLElement{
  static get observedAttributes(){return ['data']}
  set data(value){try{this._data=adopt(value);this._error=''}catch{this._data=null;this._error='Guidance unavailable: invalid reading state.'}if(this.isConnected)this.render()}
  get data(){if(this._data!==undefined)return this._data;try{const text=this.getAttribute('data');if(!text||text.length>512)throw Error();this._data=adopt(JSON.parse(text));this._error='';return this._data}catch{this._error='Guidance unavailable: invalid reading state.';return null}}
  connectedCallback(){if(Object.hasOwn(this,'data')){const value=this.data;delete this.data;this.data=value}this.render()}
  attributeChangedCallback(){this._data=undefined;if(this.isConnected)this.render()}
  render(){
   const focused=this.contains(document.activeElement)?document.activeElement?.dataset?.guidanceAction:null,revision=this._revision=(this._revision||0)+1,state=this.data;
   this.replaceChildren();this.dataset.state=state?'ready':'unavailable';
   const header=n('div',undefined,'guidance-header'),heading=n('h2',state?(state.guidePage===null?'Help · '+topics[state.context].title:'First investigation'):'Help');heading.tabIndex=-1;heading.dataset.guidanceHeading='';header.append(heading);
   const button=(label,action,disabled=false)=>{const control=n('button',label);control.type='button';control.disabled=disabled;control.dataset.guidanceAction=action;control.onclick=()=>{if(revision!==this._revision||!this.isConnected||!this.contains(control)||!nativeControlReady(control))return;control.dispatchEvent(new CustomEvent('goo-guidance-action',{bubbles:true,composed:true,detail:{action}}))};return control};
   header.append(button('Close help','close'));this.append(header);
   const body=n('div',undefined,'guidance-body');this.append(body);
   if(!state){const error=n('p',this._error);error.setAttribute('role','status');body.append(error);return}
   if(state.guidePage===null){for(const [title,text]of topics[state.context].topics)body.append(UI.disclosure(title,[n('p',text)],{className:'ui-context-help'}));body.append(button('First investigation guide','start-guide'))}
   else{
    const page=pages[state.guidePage],position=n('p','Guide page '+(state.guidePage+1)+' of '+pages.length,'guidance-position');position.setAttribute('role','status');position.setAttribute('aria-atomic','true');body.append(position,n('h3',page.title),n('p',page.body));
    const expectation=n('p',undefined,'guidance-look-for');expectation.append(n('strong','Look for '),page.lookFor);body.append(expectation);
    const controls=n('div',undefined,'guidance-controls');controls.append(button('Previous','previous',state.guidePage===0),button('Next','next',state.guidePage===pages.length-1),button('Replay','replay'),button('Back to help','about'));body.append(controls);
   }
   body.append(button('Return to work','work'));
   if(focused){const equivalent=Array.from(this.querySelectorAll('button')).find(control=>control.dataset.guidanceAction===focused&&!control.disabled);if(equivalent&&nativeControlReady(equivalent))equivalent.focus({preventScroll:true});else if(nativeControlReady(heading))heading.focus({preventScroll:true})}
  }
 }
 if(!globalThis.customElements.get('goo-guidance'))globalThis.customElements.define('goo-guidance',Guidance);
}
