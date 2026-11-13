import {mock} from 'node:test';
import {actionVisible} from '../../ui/dom-visibility.mjs';
class Element {
 constructor(tag='div',text=''){this.tagName=tag;this.localName=tag;this.textContent=text;this.style={};this.scrollTop=0;this.scrollLeft=0;this.children=[];this.dataset={};this.hidden=false;this.isConnected=true;this.value='';this.classList={add(){},toggle(){},contains(){return false},remove(){}};}
 append(...nodes){for(const node of nodes){if(node&&typeof node==='object'){node.remove?.();node.parentElement=this;}this.children.push(node)}} replaceChildren(...nodes){for(const child of this.children)if(child?.parentElement===this)child.parentElement=null;this.children=[];this.append(...nodes)} contains(node){return this===node||this.children.some(child=>child?.contains?.(node))} remove(){const parent=this.parentElement;if(parent){const index=parent.children.indexOf(this);if(index>=0)parent.children.splice(index,1)}this.parentElement=null} before(node){const parent=this.parentElement;if(!parent)return;node.remove?.();parent.children.splice(parent.children.indexOf(this),0,node);node.parentElement=parent} setAttribute(k,v){this[k]=v} getAttribute(k){return this[k]} removeAttribute(k){delete this[k]} addEventListener(){} getClientRects(){return this.isConnected&&!this.hidden?[1]:[]} querySelector(selector){return this.querySelectorAll(selector)[0]||null} querySelectorAll(selector){const walk=e=>e.children.filter(n=>n&&typeof n==='object').flatMap(n=>[n,...walk(n)]);return walk(this).filter(n=>selector.split(',').some(part=>{const s=part.trim();if(s.startsWith('.'))return (n.className||'').split(' ').includes(s.slice(1));if(s.startsWith('[aria-label='))return n['aria-label']===s.match(/=\"([^\"]+)/)?.[1];return n.tagName===s}))} closest(){return null} focus(){document.activeElement=this}
}
const node=(tag,text,cls)=>{const e=new Element(tag,text);if(cls)e.className=cls;return e};
mock.module('../../ui/components.mjs',{exports:{node,actionVisible,disclosure:node,stats(){},actions(){},inspector(){},durationMS(){},latencyModel(){},latency(){},parseAttribute(){}}});
const events=new EventTarget();
globalThis.location={hash:'',href:'http://example.test/'};
globalThis.document={hidden:false,activeElement:null,addEventListener(){},removeEventListener(){}};
globalThis.addEventListener=events.addEventListener.bind(events);globalThis.removeEventListener=events.removeEventListener.bind(events);
const {mount}=await import(process.env.GOO_OWNER_MODULE||'../../ui/population-workspace.mjs');
export const tick=async()=>{for(let i=0;i<16;i++)await Promise.resolve()};
export const question={signal:'logs',environment:'prod',fromMs:1,toMs:60000,limit:100,offset:0};
export function fixture({onAdopt,externalControls=false,externalLibrary=false}={}){
 location.hash='';const reads=[],adoptions=[],host=new Element(),controlsHost=externalControls?new Element('details'):undefined,libraryHost=externalLibrary?new Element('div'):undefined,panel=()=>host.children.find(c=>c['aria-label']==='Population workspace');let enabled=true;
 globalThis.fetch=(url,options={})=>new Promise((resolve,reject)=>reads.push({url:String(url),method:options.method||'GET',signal:options.signal,resolve:(id=String(url).split('/').at(-1))=>resolve({ok:true,json:async()=>({version:1,id,title:'Saved '+id,queries:[structuredClone(question)],layout:'charts-records',columns:['utc','service','summary']})}),reject:()=>reject(Error('offline'))}));
 // Stop at the real owner's admitted callback. Native coverage separately checks
 // rendering/controls; this fixture measures acquisition and adoption ownership.
 const owner=mount(host,{current:()=>null,controlsHost,libraryHost,active:()=>enabled,onRestore(q){adoptions.push(q);if(onAdopt){onAdopt(q);return;}throw Error('adoption observation barrier')}});
 const emit=(type='hashchange',newHash)=>{const e=new Event(type);if(newHash!==undefined)Object.defineProperty(e,'newURL',{value:'http://example.test/'+newHash});events.dispatchEvent(e)};
 const route=(id,types=['popstate','hashchange'])=>{location.hash=id?'#exploration-view='+id:'#explore';for(const type of types)emit(type)};
 return {owner,reads,adoptions,host,controlsHost,libraryHost,panel,emit,route,workspace:route=>events.dispatchEvent(new CustomEvent('goo:workspace',{detail:{route}})),setActive:v=>enabled=v};
}
