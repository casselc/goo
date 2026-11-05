import {mock} from 'node:test';
class Element {
 constructor(tag='div',text=''){this.tagName=tag;this.textContent=text;this.children=[];this.dataset={};this.hidden=false;this.isConnected=true;this.value='';this.classList={add(){},toggle(){}};}
 append(...nodes){this.children.push(...nodes)} replaceChildren(...nodes){this.children=nodes} contains(node){return this===node||this.children.some(child=>child?.contains?.(node))} remove(){} setAttribute(k,v){this[k]=v} getAttribute(k){return this[k]} addEventListener(){} getClientRects(){return this.isConnected&&!this.hidden?[1]:[]} querySelector(){return null} querySelectorAll(){return []} closest(){return null} focus(){document.activeElement=this}
}
const node=(tag,text)=>new Element(tag,text);
mock.module('../../ui/components.mjs',{exports:{node,disclosure:node,stats(){},actions(){},inspector(){},durationMS(){},latencyModel(){},latency(){},parseAttribute(){}}});
const events=new EventTarget();
globalThis.location={hash:'',href:'http://example.test/'};
globalThis.document={hidden:false,activeElement:null};
globalThis.addEventListener=events.addEventListener.bind(events);globalThis.removeEventListener=events.removeEventListener.bind(events);
const {mount}=await import(process.env.GOO_OWNER_MODULE||'../../ui/population-workspace.mjs');
export const tick=async()=>{for(let i=0;i<16;i++)await Promise.resolve()};
export const question={signal:'logs',environment:'prod',fromMs:1,toMs:60000,limit:100,offset:0};
export function fixture({onAdopt}={}){
 location.hash='';const reads=[],adoptions=[],host=new Element(),panel=()=>host.children.find(c=>c['aria-label']==='Population workspace');let enabled=true;
 globalThis.fetch=(url,options={})=>new Promise((resolve,reject)=>reads.push({url:String(url),method:options.method||'GET',signal:options.signal,resolve:(id=String(url).split('/').at(-1))=>resolve({ok:true,json:async()=>({version:1,id,title:'Saved '+id,queries:[structuredClone(question)],layout:'charts-records',columns:['utc','service','summary']})}),reject:()=>reject(Error('offline'))}));
 // Stop at the real owner's admitted callback. Native coverage separately checks
 // rendering/controls; this fixture measures acquisition and adoption ownership.
 const owner=mount(host,{current:()=>null,active:()=>enabled,onRestore(q){adoptions.push(q);if(onAdopt){onAdopt(q);return;}throw Error('adoption observation barrier')}});
 const emit=(type='hashchange',newHash)=>{const e=new Event(type);if(newHash!==undefined)Object.defineProperty(e,'newURL',{value:'http://example.test/'+newHash});events.dispatchEvent(e)};
 const route=(id,types=['popstate','hashchange'])=>{location.hash=id?'#exploration-view='+id:'#explore';for(const type of types)emit(type)};
 return {owner,reads,adoptions,host,panel,emit,route,workspace:route=>events.dispatchEvent(new CustomEvent('goo:workspace',{detail:{route}})),setActive:v=>enabled=v};
}
