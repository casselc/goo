/* Supplied authored tags; persistence and annotations belong to application owners. */
import {node} from './components.mjs';
const global=globalThis;
 function tags(value){if(!Array.isArray(value)||value.length>16||new Set(value).size!==value.length||value.some(t=>typeof t!=='string'||!/^[a-z0-9][a-z0-9._:/-]{0,63}$/.test(t)))throw Error('Invalid tags');return [...value]}
 if(global.customElements&&global.HTMLElement){class GooTagList extends global.HTMLElement{
  static get observedAttributes(){return ['tags','interactive']}
  set tags(value){try{this._tags=tags(value);this._invalid=false}catch{this._tags=[];this._invalid=true}if(this.isConnected)this.render()}
  get tags(){if(this._tags)return [...this._tags];try{return tags(JSON.parse(this.getAttribute('tags')||'[]'))}catch{return []}}
  connectedCallback(){if(Object.hasOwn(this,'tags')){const v=this.tags;delete this.tags;this.tags=v}this.render()}
  attributeChangedCallback(name){if(name==='tags'){this._tags=null;this._invalid=false}if(this.isConnected)this.render()}
  render(){this.replaceChildren();let items;try{items=this._tags||tags(JSON.parse(this.getAttribute('tags')||'[]'));if(this._invalid)throw Error()}catch{this.dataset.state='invalid';this.append(node('span','Tags unavailable'));return}this.dataset.state='ready';for(const tag of items){const el=node(this.hasAttribute('interactive')?'button':'span',tag);if(el.tagName==='BUTTON'){el.type='button';el.onclick=()=>this.dispatchEvent(new CustomEvent('goo-tag-select',{bubbles:true,detail:{tag}}))}this.append(el)}}
 };if(!customElements.get('goo-tag-list'))customElements.define('goo-tag-list',GooTagList)}
export {tags};
