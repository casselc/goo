import assert from 'node:assert/strict';
import {test,mock} from 'node:test';
class Element{
 constructor(tag='div'){this.tagName=tag;this.children=[];this.dataset={};this.attributes={};this.isConnected=true;this.title='';this.scrollTop=0;this.listeners=new Map();this.classList={add:(...names)=>{this.className=[this.className||'',...names].join(' ')}};}
 append(...nodes){for(const node of nodes){node.remove?.();node.parentElement=this;this.children.push(node)}}
 prepend(...nodes){for(const node of [...nodes].reverse()){node.remove?.();node.parentElement=this;this.children.unshift(node)}}
 replaceChildren(...nodes){for(const node of this.children)node.parentElement=null;this.children=[];this.append(...nodes)}
 remove(){if(this.parentElement)this.parentElement.children=this.parentElement.children.filter(node=>node!==this);this.parentElement=null}
 after(node){const parent=this.parentElement;if(parent){node.remove?.();node.parentElement=parent;parent.children.splice(parent.children.indexOf(this)+1,0,node)}}
 contains(node){return node===this||this.children.some(child=>child.contains?.(node))}
 setAttribute(key,value){this.attributes[key]=value}hasAttribute(key){return Object.hasOwn(this.attributes,key)}getAttribute(key){return this.attributes[key]}
 matches(selector){if(selector.startsWith('.'))return(this.className||'').split(' ').includes(selector.slice(1));if(selector.startsWith('[data-')){const key=selector.slice(6,-1).replace(/-([a-z])/g,(_,letter)=>letter.toUpperCase());return Object.hasOwn(this.dataset,key)}return this.tagName===selector}
 querySelectorAll(selector){return this.children.flatMap(child=>[...(child.matches?.(selector)?[child]:[]),...(child.querySelectorAll?.(selector)||[])])}querySelector(selector){return this.querySelectorAll(selector)[0]||null}
 closest(selector){return this.matches(selector)?this:this.parentElement?.closest(selector)||null}
 addEventListener(type,listener){this.listeners.set(type,listener)}dispatch(type){this.listeners.get(type)?.({type})}
}
const node=(tag,text,className)=>{const element=new Element(tag);element.textContent=text;if(className)element.className=className;return element};
let inspected,shared;
mock.module('../ui/attributes.mjs',{exports:{mountShared:(_host,_records,options)=>{shared=options}}});
mock.module('../ui/record-presentation.mjs',{exports:{table(){const element=node('table'),viewport=node('div'),header=node('thead'),body=node('tbody');element.append(header,body);viewport.append(element);return {element,viewport,header,body,row(){const row=node('tr'),cells=Array.from({length:4},()=>node('td'));row.append(...cells);return{row,cells}}}},description:()=>'',detail:(_row,_signal,options)=>{inspected=options;return node('section')}}});
mock.module('../ui/components.mjs',{exports:{node,inspector(panel){const body=node('div',undefined,'ui-inspector-body');panel.append(body);return {has:()=>false,show(){}}}}});
mock.module('../ui/record-columns.mjs',{exports:{mount:()=>({update(){}})}});
mock.module('../ui/kept-selection.mjs',{exports:{mount(){}}});
for(const name of ['discovery','retention','trace-search','visualizations','record-list','record-span-waterfall'])mock.module('../ui/'+name+'.mjs',{exports:{}});
const {mount}=await import('../ui/record-interactions.mjs');
test('record inspector and shared-field wrappers preserve host acceptance and reject retired owners',()=>{
 for(const accepted of [false,undefined,true]){
  let allowed=true,calls=0;inspected=shared=null;const target=node('section'),onFacet=()=>{calls++;return accepted},options={onFacet,canRefine:()=>allowed};
  const owner=mount(target,[{Body:'ordinary',Attributes:{tenant:'team'}}],'logs',{signal:'logs'},'basis',options,{attributeOptions:()=>options});
  target.querySelector('[data-inspect-record]').onclick();assert(inspected);assert.equal(inspected.onFacet({kind:'attribute',key:'tenant'},'team','include'),accepted);
  const input=target.querySelector('[data-record-selection]');input.checked=true;input.onchange();const disclosure=target.querySelectorAll('details').find(element=>element.children[0]?.textContent==='Fields shared by selected records');assert(disclosure);disclosure.open=true;disclosure.dispatch('toggle');assert(shared);assert.equal(shared.onFacet({kind:'attribute',key:'tenant'},'team','include'),accepted);assert.equal(calls,2);
  allowed=false;assert.equal(inspected.onFacet({},'team','include'),false);assert.equal(shared.onFacet({},'team','include'),false);assert.equal(calls,2);
  allowed=true;owner.dispose();assert.equal(inspected.onFacet({},'team','include'),false);assert.equal(shared.onFacet({},'team','include'),false);assert.equal(calls,2);
 }
});

test('shared-field actions require an actual host and use the canonical attribute-action owner',()=>{
 for(const hasHost of [false,true]){
  let calls=0;shared=null;const target=node('section'),options={resultToken:'opaque'},origin={signal:'logs',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}}},onFacet=()=>{calls++;return false};
  const owner=mount(target,[{Body:'ordinary'}],'logs',origin,'basis',options,{attributeOptions:()=>hasHost?{onFacet}:{}});const input=target.querySelector('[data-record-selection]');input.checked=true;input.onchange();const disclosure=target.querySelectorAll('details').find(element=>element.children[0]?.textContent==='Fields shared by selected records');disclosure.open=true;disclosure.dispatch('toggle');assert(shared);if(hasHost){assert.equal(shared.onFacet({},'value','include'),false);assert.equal(calls,1)}else{assert.equal(shared.onFacet,undefined);assert.equal(calls,0)}owner.dispose();
 }
});
