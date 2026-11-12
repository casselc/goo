import assert from 'node:assert/strict';
import {test,mock} from 'node:test';
let focused=null;
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
 focus(){focused=this}scrollIntoView(){this.scrolled=true}
 addEventListener(type,listener){this.listeners.set(type,listener)}dispatch(type){this.listeners.get(type)?.({type})}
}
const node=(tag,text,className)=>{const element=new Element(tag);element.textContent=text;if(className)element.className=className;return element};
let inspected,shared,inspectedActions;
mock.module('../ui/attributes.mjs',{exports:{mountShared:(_host,_records,options)=>{shared=options}}});
mock.module('../ui/record-presentation.mjs',{exports:{table(){const element=node('table'),viewport=node('div'),header=node('thead'),body=node('tbody');element.append(header,body);viewport.append(element);return {element,viewport,header,body,row(){const row=node('tr'),cells=Array.from({length:4},()=>node('td'));row.append(...cells);return{row,cells}}}},description:()=>'',detail:(_row,_signal,options)=>{inspected=options;return node('section')}}});
mock.module('../ui/components.mjs',{exports:{node,actions(primary,secondary,{label}){const group=node('div'),details=node('details'),summary=node('summary',label);details.append(summary,...secondary);group.append(...primary,details);return group},actionVisible:element=>{if(!element.isConnected||element.visible===false)return false;for(let parent=element.parentElement;parent;parent=parent.parentElement)if(parent.tagName==='details'&&!parent.open)return false;return true},inspector(panel){const header=node('div',undefined,'ui-inspector-heading'),heading=node('h4'),body=node('div',undefined,'ui-inspector-body'),cache=new Set();header.append(heading);panel.append(header,body);panel.hidden=true;return {has:key=>cache.has(key),show({key,actions=[]}){cache.add(key);panel.hidden=false;inspectedActions=actions;panel.append(...actions)}}}}});
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


test('explicit record opening focuses details for fresh and cached inspection only while current and visible',()=>{
 let active=true;const target=node('section'),owner=mount(target,[{Body:'ordinary'}],'logs',{signal:'logs'},'basis',{}, {active:()=>active}),opener=target.querySelector('[data-inspect-record]'),heading=target.querySelector('.ui-inspector-heading').querySelector('h4');
 for(const cached of [false,true]){focused=null;opener.onclick();assert.equal(focused,heading,cached?'cached selection':'fresh selection');assert.equal(heading.tabIndex,-1);assert.equal(heading.scrolled,true)}
 focused=null;opener.visible=false;opener.onclick();assert.equal(focused,null,'a nonvisible opener cannot move keyboard focus');
 opener.visible=true;active=false;opener.onclick();assert.equal(focused,null,'a paused host cannot move keyboard focus');
 active=true;owner.dispose();opener.onclick();assert.equal(focused,null,'a retired owner cannot move keyboard focus');
});


test('secondary record reads require an open row menu or a visible inspector action and retain the owner',()=>{
 const target=node('section'),reads=[],row={SpanName:'/operations/shared-prefix/settlement',TraceId:'a'.repeat(32),SpanId:'b'.repeat(16)};
 const owner=mount(target,[row],'spans',{signal:'spans'},'fixed-basis',{}, {openTrace:(...args)=>reads.push(args)});
 const expand=target.querySelector('[data-expand-trace]'),menu=expand.parentElement,inspect=target.querySelector('[data-inspect-record]');
 assert.equal(menu.tagName,'details');assert.equal(menu.querySelector('summary').getAttribute('aria-label'),'More record actions');
 expand.onclick();assert.equal(reads.length,0);
 menu.open=true;expand.disabled=true;expand.onclick();assert.equal(reads.length,0);expand.disabled=false;expand.onclick();assert.equal(reads.length,1);assert.equal(reads[0][4],menu.querySelector('summary'),'closed row menu returns focus to its persistent summary');assert.equal(reads[0][1],row.TraceId);
 menu.open=false;inspect.onclick();assert.equal(reads.length,1);assert.equal(inspect.getAttribute('aria-pressed'),'true');
 const footer=inspectedActions[0];footer.onclick();assert.equal(reads.length,2);assert.equal(reads[1][4],footer);assert.equal(reads[1][3],'fixed-basis');
 owner.dispose();menu.open=true;expand.onclick();footer.onclick();assert.equal(reads.length,2);
});
