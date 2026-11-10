import {readFileSync} from 'node:fs';
import {plan} from '../../ui/native-observation-context.mjs';
export const clone=structuredClone;
export function measurement({unavailable=false}={}){
 const frame=JSON.parse(readFileSync(new URL('../fixtures/native-observation-context-388.json',import.meta.url))).find(c=>!c.refused),r=clone(frame.result);
 if(unavailable){const sample=r.gauges.series[0].samples[1];sample.state='unknown';delete sample.value;sample.reason='No recorded value';r.rows[1].flags=1;r.rows[1].value='NaN';}
 return r;
}
export function page(options){const r=measurement(options);return {receipt:{version:1,id:'a'.repeat(64),title:'Retained gauge observations',capturedAt:'2026-11-10T07:00:00Z',question:{kind:'measurement',measurement:clone(r.question)}},measurement:r,retainedRecords:r.rows.length,recordsLimited:false};}
export function detail(r,signal='logs',atNS=r.gauges.series[0].samples[0].atNS){const stream=r.gauges.series[0].stream;return {...plan(r,stream,signal,atNS),measurement:clone(r.question),stream:clone(stream),basis:clone(r.basis),readAt:r.readAt};}
export class Element{
 constructor(tag='div'){this.localName=tag;this.children=[];this.parentElement=null;this.isConnected=true;this.hidden=false;this.inert=false;this.open=false;this.listeners={};this.dataset={};this.attributes={};this.selectedKey='held-observation';this.onclick=null;this.onauxclick=null;this.disabled=false;}
 append(...items){this.children.push(...items);for(const item of items)if(item&&typeof item==='object')item.parentElement=this;}
 replaceChildren(...items){for(const child of this.children)if(child&&typeof child==='object'){child.isConnected=false;child.parentElement=null;}this.children=[];this.append(...items);}
 contains(item){return this.children.some(child=>child===item||child&&typeof child==='object'&&child.contains?.(item));}
 closest(selector){for(let p=this;p;p=p.parentElement)if(selector==='[inert]'&&p.inert)return p;return null;}
 setAttribute(key,value){this.attributes[key]=String(value);}
 getAttribute(key){return this.attributes[key]??null;}
 addEventListener(key,fn){this.listeners[key]=fn;}
 getClientRects(){return this.isConnected&&!this.hidden?[{}]:[];}
}
export const visible=node=>{for(let p=node;p;p=p.parentElement)if(!p.isConnected||p.hidden)return false;return !!node;};
export const event=(extra={})=>({button:0,prevented:0,stopped:0,preventDefault(){this.prevented++},stopImmediatePropagation(){this.stopped++},stopPropagation(){},...extra});
