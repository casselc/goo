// A small native custom-element lifecycle fixture; production Chart owns rendering.
export class Element{
 constructor(tag='section'){this.localName=tag;this.children=[];this.parentElement=null;this.attributes=new Map();this.dataset={};this.listeners=new Map();this._text='';this.root=false;this.className='';this.style={};this.classList={add:name=>{this.className+=' '+name},remove:name=>{this.className=this.className.split(' ').filter(value=>value!==name).join(' ')},toggle:(name,on)=>{if(on)this.classList.add(name);else this.classList.remove(name)}}}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}get childNodes(){return this.children}get options(){return this.children.filter(child=>child.localName==='option')}
 get textContent(){return this._text+this.children.map(child=>child.textContent).join('')}set textContent(value){this._text=String(value??'');this.replaceChildren()}
 append(...nodes){for(let item of nodes){if(typeof item==='string'){const text=new Element('#text');text.textContent=item;item=text}item.remove();item.parentElement=this;this.children.push(item)}}
 replaceChildren(...nodes){this.children.forEach(child=>child.parentElement=null);this.children=[];this.append(...nodes)}remove(){if(this.parentElement)this.parentElement.children=this.parentElement.children.filter(child=>child!==this);this.parentElement=null}
 setAttribute(name,value){const previous=this.getAttribute(name);this.attributes.set(name,String(value));if(this.constructor.observedAttributes?.includes(name))this.attributeChangedCallback?.(name,previous,String(value))}getAttribute(name){return this.attributes.get(name)??null}hasAttribute(name){return this.attributes.has(name)}removeAttribute(name){const previous=this.getAttribute(name);this.attributes.delete(name);if(this.constructor.observedAttributes?.includes(name))this.attributeChangedCallback?.(name,previous,null)}
 querySelectorAll(selector){return this.children.flatMap(child=>{const match=selector==='*'||(selector.startsWith('.')?child.className.split(' ').includes(selector.slice(1)):child.localName===selector);return[...(match?[child]:[]),...child.querySelectorAll(selector)]})}
 querySelector(selector){return this.querySelectorAll(selector)[0]??null}contains(element){return this===element||this.children.some(child=>child.contains(element))}
 addEventListener(name,handler){const callbacks=this.listeners.get(name)||[];callbacks.push(handler);this.listeners.set(name,callbacks)}dispatchEvent(event){for(const handler of this.listeners.get(event.type)||[])handler(event);return true}getClientRects(){return this.isConnected&&!this.hidden?[{}]:[]}focus(){document.activeElement=this}
}
const registry=new Map();globalThis.HTMLElement=Element;globalThis.customElements={get:name=>registry.get(name),define:(name,constructor)=>registry.set(name,constructor)};globalThis.CustomEvent=class{constructor(type,options){this.type=type;Object.assign(this,options)}};
export const node=tag=>{const constructor=tag==='goo-chart'?registry.get(tag):null,element=constructor?new constructor():new Element(tag);element.localName=tag;return element};
const body=node('body');body.root=true;globalThis.document={body,activeElement:null,createElement:node,createTextNode:text=>{const element=node('#text');element.textContent=text;return element}};
await import('../../ui/scope-view.mjs');
export const Chart=registry.get('goo-chart');
export const Breakdown=await import('../../ui/scoped-breakdown.mjs');
export const Results=await import('../../ui/record-results.mjs');
