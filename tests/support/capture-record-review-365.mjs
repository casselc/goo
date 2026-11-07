import {mock} from 'node:test';
export class Element{
 constructor(tag='div',text=''){this.localName=tag;this.text=String(text||'');this.children=[];this.parentElement=null;this.root=false;this.hidden=false;this.disabled=false;this.value='';this.open=false;this.listeners={};this.attributes={}}
 get textContent(){return this.text+this.children.map(c=>c.textContent).join('')}set textContent(value){this.text=String(value);this.replaceChildren()}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}get firstChild(){return this.children[0]}
 append(...children){for(const child of children){child.remove();child.parentElement=this;this.children.push(child)}}
 replaceChildren(...children){for(const c of this.children)c.parentElement=null;this.children=[];this.append(...children)}remove(){if(this.parentElement){const list=this.parentElement.children;list.splice(list.indexOf(this),1);this.parentElement=null}}
 setAttribute(key,value){this.attributes[key]=String(value)}addEventListener(name,handler){this.listeners[name]=handler}
}
export const walk=root=>[root,...root.children.flatMap(walk)];
export const visible=el=>{if(!el?.isConnected)return false;for(let p=el;p;p=p.parentElement)if(p.hidden||p.localName==='details'&&!p.open&&el.localName!=='summary')return false;return true};
mock.module('../../ui/components.mjs',{exports:{node:(...args)=>new Element(...args),actionVisible:visible}});
globalThis.document={createElement:tag=>new Element(tag)};
export const Review=await import('../../ui/capture-record-review.mjs');
export const ids={capture:'a'.repeat(64),task:'b'.repeat(64),first:'c'.repeat(64),reviewed:'d'.repeat(64),branch:'e'.repeat(64),merged:'f'.repeat(64),manifest:'1'.repeat(64)};
export const task=()=>({id:ids.task,version:1,name:'investigate',title:'Investigation review',labels:[{key:'investigate',title:'Investigate'},{key:'routine',title:'Routine'}],origin:'local-user-unverified'});
export const resolved=()=>({reference:{captureId:ids.capture,index:100},recordDigest:'9'.repeat(64),record:{ServiceName:'worker',timestamp_ns:'1794045600000000123'},receipt:{id:ids.capture,capturedAt:'2026-11-07T10:01:00Z',question:{kind:'explore',explore:{signal:'logs'}}},sourceReadAt:'2026-11-07T10:00:00Z'});
export const decision=(id=ids.first,parents=[],review='unreviewed')=>({id,version:1,taskID:ids.task,target:{captureId:ids.capture,index:100},parents,evidence:{reference:{captureId:ids.capture,index:100},recordDigest:'9'.repeat(64),receipt:{id:ids.capture}},judgment:{kind:'label',value:'investigate'},review,state:'active',note:'kept note',origin:'local-user-unverified'});
export function fixture(){const requests=[];globalThis.fetch=(url,options)=>new Promise(resolve=>requests.push({url,options,body:options.body?JSON.parse(options.body):undefined,reply(value,ok=true){resolve({ok,json:async()=>value})}}));let active=true;const host=new Element();host.root=true;const owner=Review.mount(host,{target:{captureId:ids.capture,index:100},current:()=>active}),all=()=>walk(host),button=text=>all().find(e=>e.localName==='button'&&e.textContent===text),field=label=>all().find(e=>e.localName==='label'&&e.text===label)?.firstChild,disclosure=text=>all().find(e=>e.localName==='details'&&e.firstChild?.textContent===text);return{owner,host,requests,all,button,field,disclosure,active:value=>active=value,open(text){const d=disclosure(text);d.open=true;return d.listeners.toggle?.()},respond(value,ok=true){requests.at(-1).reply(value,ok)}}}
export const tick=()=>new Promise(resolve=>setImmediate(resolve));
export function choose(f,value='label:investigate'){const select=f.all().find(e=>e.attributes['aria-label']==='Review judgment');select.value=value;select.onchange()}
export async function ready(f){f.open('Review example');f.respond(resolved());await tick();f.field('Exact task ID').value=ids.task;f.button('Load task').onclick();f.respond(task());await tick();choose(f)}

export const manifest=()=>({id:ids.manifest,version:1,taskID:ids.task,mode:'pinned-review-branches',historical:true,selected:[ids.reviewed],knownHeads:[ids.reviewed],revisions:[ids.first,ids.reviewed],items:[{decisionID:ids.reviewed,label:'investigate',evidence:decision().evidence}]});
