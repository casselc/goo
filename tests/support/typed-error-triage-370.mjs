import {mock} from 'node:test';
import * as Components from '../../ui/components.mjs';
import * as Visualizations from '../../ui/visualizations.mjs';
import {actionVisible} from '../../ui/dom-visibility.mjs';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as ErrorTriage from '../../ui/error-triage.mjs';
import * as QuestionProtocol from '../../ui/question-protocol.mjs';

// Minimal connected DOM; production groups owns anchors, leases and rebinding.
// Source/result admission and real browser focus are checked by the native journey.
export class Element {
 constructor(tag='section',text=''){this.localName=tag;this.children=[];this.parentElement=null;this._text=text??'';this.attributes=new Map();this.dataset={};this.hidden=false;this.inert=false;this.open=false;this.root=false;}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}
 get textContent(){return this._text+this.children.map(child=>child.textContent).join('')}
 set textContent(value){this._text=value??'';this.replaceChildren()}
 append(...nodes){for(const node of nodes){node.remove();node.parentElement=this;this.children.push(node)}}
 replaceChildren(...nodes){for(const child of this.children)child.parentElement=null;this.children=[];this.append(...nodes)}
 remove(){if(this.parentElement){this.parentElement.children=this.parentElement.children.filter(child=>child!==this);this.parentElement=null}}
 contains(node){return this===node||this.children.some(child=>child.contains(node))}
 setAttribute(name,value){this.attributes.set(name,String(value))}
 getAttribute(name){return this.attributes.get(name)??null}
 matches(selector){if(selector===':disabled')return !!this.disabled;if(selector==='details:not([open])')return this.localName==='details'&&!this.open;if(selector==='fieldset:disabled')return this.localName==='fieldset'&&!!this.disabled;return this.localName===selector}
 closest(selector){for(let node=this;node;node=node.parentElement){if(selector==='fieldset:disabled'&&node.localName==='fieldset'&&node.disabled)return node;if(selector==='[inert]'&&node.inert)return node;if(selector==='[hidden], [aria-hidden="true"]'&&(node.hidden||node.getAttribute('aria-hidden')==='true'))return node;if(selector==='details:not([open])'&&node.localName==='details'&&!node.open)return node}return null}
 querySelector(selector){if(selector===':scope > summary')return this.children.find(child=>child.localName==='summary')||null;if(selector===':scope > legend')return this.children.find(child=>child.localName==='legend')||null;return this.querySelectorAll(selector)[0]||null}
 querySelectorAll(selector){return this.children.flatMap(child=>[...(child.localName===selector?[child]:[]),...child.querySelectorAll(selector)])}
 getClientRects(){return this.isConnected?[{}]:[]}
 focus(){document.activeElement=this}
}
mock.module('../../ui/components.mjs',{exports:{...Components,node:(tag,text)=>new Element(tag,text)}});
mock.module('../../ui/visualizations.mjs',{exports:{...Visualizations,actionVisible}});
mock.module('../../ui/scoped-breakdown.mjs',{exports:{mount(){return {suspendActions(){},resumeActions(){}}}}});
const {groups}=await import('../../ui/record-results.mjs');
export function event(options={}){return{button:0,defaultPrevented:false,stopped:false,...options,preventDefault(){this.defaultPrevented=true},stopImmediatePropagation(){this.stopped=true}}}
export function mounted(data,actions){globalThis.document={hidden:false,activeElement:null};const host=new Element();host.root=true;const owner=groups(host,data,{actions});return{host,owner,anchors:()=>host.querySelectorAll('a')}}
// Exercise the actual closure body without mounting an unrelated editor/trace
// stack. The native journey proves the complete source owner and route schedule.
export function admission(data){
 const source=readFileSync(new URL('../../ui/exploration.mjs',import.meta.url),'utf8').split('\n').find(line=>line.trimStart().startsWith('function groupActions(data)'));
 if(!source)throw Error('Production group action owner is required');
 const state={phase:'applied',generation:7},env={readOnly:false,triageGroupLive:true,lastResult:data,ErrorTriage,QuestionProtocol,activeFlag:true,refinable:true,population:false,container:{isConnected:true,classList:{contains:()=>env.population}},session:{snapshot:()=>state},active:()=>env.activeFlag,canRefine:()=>env.refinable};
 const groupActions=vm.runInNewContext('('+source.trim()+')',env);
 return{env,state,actions:()=>groupActions(data)};
}
