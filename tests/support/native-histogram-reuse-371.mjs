import {mock} from 'node:test';
import {actionVisible} from '../../ui/dom-visibility.mjs';
export class Element{
 constructor(tag='section'){this.tagName=tag;this.children=[];this.attributes={};this.dataset={};this._connected=false;this._text='';this.style={setProperty(key,value){this[key]=value}};this.listeners=new Map();this.className='';this.value='';this.classList={add:(name)=>{if(!this.className.split(' ').includes(name))this.className+=' '+name},toggle:(name,on)=>{this.className=this.className.split(' ').filter(value=>value&&value!==name).join(' ');if(on)this.classList.add(name)}}}
 get isConnected(){return this._connected}set isConnected(value){this._connected=value;for(const child of this.children)child.isConnected=value}
 set textContent(value){this._text=String(value??'')}get textContent(){return this._text+this.children.map(child=>child.textContent).join('')}
 append(...children){for(const child of children){if(child.parentElement)child.parentElement.children=child.parentElement.children.filter(item=>item!==child);child.parentElement=this;child.isConnected=this.isConnected;this.children.push(child)}}
 prepend(...children){this.append(...children);this.children=[...children,...this.children.filter(child=>!children.includes(child))]}
 replaceChildren(...children){for(const child of this.children){child.parentElement=null;child.isConnected=false}this.children=[];this._text='';this.append(...children)}
 contains(element){return element===this||this.children.some(child=>child.contains(element))}
 setAttribute(key,value){this.attributes[key]=String(value)}getAttribute(key){return this.attributes[key]??null}
 matches(selector){if(selector.startsWith('.'))return this.className.split(' ').includes(selector.slice(1));if(selector==='[aria-label]')return this.getAttribute('aria-label')!==null;if(selector.startsWith('[aria-label=\"'))return this.getAttribute('aria-label')===selector.match(/^\[aria-label="(.*)"\]$/)?.[1];return this.tagName===selector}
 querySelectorAll(selector){return this.children.flatMap(child=>[...(child.matches(selector)?[child]:[]),...child.querySelectorAll(selector)])}querySelector(selector){return this.querySelectorAll(selector)[0]||null}
 addEventListener(type,handler){this.listeners.set(type,handler)}dispatchEvent(event){this.listeners.get(event.type)?.(event);return true}focus(){globalThis.document.activeElement=this}
 closest(selector){for(let e=this;e;e=e.parentElement){if(selector==='[inert]'&&e.inert)return e;if(selector==='[hidden], [aria-hidden="true"]'&&(e.hidden||e.getAttribute('aria-hidden')==='true'))return e;if(selector==='details:not([open])'&&e.tagName==='details'&&!e.open)return e}return null}
 getClientRects(){return this.hidden?[]:[{}]}
 setSelectionRange(start,end,direction){this.selectionStart=start;this.selectionEnd=end;this.selectionDirection=direction}
}
const node=(tag,text,className)=>{const element=new Element(tag);element.textContent=text;if(className)element.className=className;return element};
globalThis.HTMLElement=Element;const body=node('body');body.isConnected=true;globalThis.document={body,activeElement:body};
mock.module('../../ui/components.mjs',{exports:{node,actionVisible}});

export {node};
export const {mount}=await import('../../ui/native-histogram-reuse.mjs');
