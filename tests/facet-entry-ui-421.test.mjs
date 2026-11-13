import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Attributes from '../ui/attributes.mjs';
import * as Editor from '../ui/filter-editor.mjs';

class Element{
 constructor(tag){this.localName=tag;this.children=[];this.attrs={};this.dataset={};this.value='';this.textContent='';this.open=false;this.hidden=false;this.disabled=false;this.inert=false;this.root=false;}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}
 append(...nodes){for(const node of nodes){if(typeof node==='object'){node.remove();node.parentElement=this;}this.children.push(node)}if(this.localName==='select'&&!this.value)this.value=this.children.find(c=>c.value)?.value||''}
 prepend(...nodes){for(const node of nodes){node.remove();node.parentElement=this;}this.children.unshift(...nodes)}
 before(node){const p=this.parentElement;node.remove();node.parentElement=p;p.children.splice(p.children.indexOf(this),0,node)}
 remove(){if(this.parentElement){const p=this.parentElement;this.parentElement=null;p.children=p.children.filter(c=>c!==this)}}
 replaceChildren(...nodes){for(const child of [...this.children])if(typeof child==='object')child.remove();this.children=[];this.append(...nodes)}
 setAttribute(key,value){this.attrs[key]=String(value)}getAttribute(key){return this.attrs[key]??null}removeAttribute(key){delete this.attrs[key]}
 addEventListener(){}
 contains(node){return node===this||this.children.some(c=>typeof c==='object'&&c.contains(node))}
 matches(selector){if(selector==='details:not([open])')return this.localName==='details'&&!this.open;if(selector===':disabled'){if(this.disabled)return true;for(let p=this.parentElement;p;p=p.parentElement)if(p.localName==='fieldset'&&p.disabled&&!p.querySelector(':scope > legend')?.contains(this))return true;return false}if(selector==='details[open]')return this.localName==='details'&&this.open;return this.localName===selector}
 closest(selector){for(let node=this;node;node=node.parentElement){if(selector==='details:not([open])'&&node.matches(selector)||selector==='fieldset:disabled'&&node.localName==='fieldset'&&node.disabled||selector==='.ui-action-menu'&&node.className?.split(' ').includes('ui-action-menu')||selector.includes('[hidden]')&&node.hidden||selector.includes('[aria-hidden')&&node.getAttribute('aria-hidden')==='true'||selector.includes('[inert]')&&node.inert||selector==='goo-grouping-control'&&node.localName===selector)return node}return null}
 querySelectorAll(selector){if(selector===':scope > summary'||selector===':scope > legend')return this.children.filter(c=>c.localName===selector.slice(9));if(selector==='details[open]>summary')return walk(this).filter(c=>c.localName==='summary'&&c.parentElement?.localName==='details'&&c.parentElement.open);return walk(this).slice(1).filter(c=>selector.split(',').some(s=>c.matches(s.trim())))}
 querySelector(selector){return this.querySelectorAll(selector)[0]||null}
 getClientRects(){return this.isConnected&&!this.closest('[hidden], [aria-hidden="true"]')?[{}]:[]}
 focus(){document.activeElement=this}scrollIntoView(){}
}
const walk=node=>[node,...node.children.filter(c=>typeof c==='object').flatMap(walk)];
const label=(host,text)=>walk(host).find(node=>node.getAttribute('aria-label')===text);
function dom(fn){const before=globalThis.document;globalThis.document={hidden:false,createElement:tag=>new Element(tag)};try{return fn()}finally{globalThis.document=before}}
const field={kind:'attribute',key:'region'},other={kind:'resource',key:'region'};
const observations={label:'Supplied records',fields:[{field,values:['west','']},{field:other,values:['east']},{field:{kind:'column',key:'duration_ns'},values:['4']}],rows:2,records:2};

test('421 field entry reveals the exact existing Load despite search, with no read or predicate mutation',()=>dom(()=>{
 const host=new Element('section');host.root=true;let loads=0,changes=0;const editor=Editor.mount(host,{signal:'spans',onChange:()=>changes++});editor.setSuggestions(observations);editor.setValueLoader(()=>loads++);
 const search=label(host,'Search observed fields');search.value='does-not-match';search.oninput();assert.equal(label(host,'Load scoped values for attribute:region'),undefined);
 const before=editor.snapshot(),load=editor.focusField(field);assert(load);assert.equal(load,label(host,'Load scoped values for attribute:region'));assert.equal(document.activeElement,load);assert.equal(search.value,'');assert.equal(load.closest('details:not([open])'),null);assert.equal(loads,0);assert.equal(changes,0);assert.deepEqual(editor.snapshot(),before);
 assert.notEqual(editor.focusField(other),load);assert.equal(document.activeElement,label(host,'Load scoped values for resource:region'));
 load.onclick();assert.equal(loads,1);
}));

test('421 field focus and existing Load refuse unsupported, absent, withdrawn and stale controls',()=>dom(()=>{
 const host=new Element('section');host.root=true;let mutable=true,loads=0;const editor=Editor.mount(host,{signal:'spans',canMutate:()=>mutable});editor.setSuggestions(observations);editor.setValueLoader(()=>loads++);
 assert.equal(editor.focusField({kind:'attribute',key:'absent'}),null);assert.equal(editor.focusField({kind:'column',key:'duration_ns'}),null);assert.equal(editor.focusField({kind:'bad',key:'region'}),null);
 const load=editor.focusField(field),menu=load.parentElement;menu.open=false;load.onclick();assert.equal(loads,0);menu.open=true;
 for(const mode of ['hidden','inert','disabled']){load[mode]=true;load.onclick();load[mode]=false;assert.equal(loads,0)}
 mutable=false;assert.equal(editor.focusField(field),null);load.onclick();assert.equal(loads,0);mutable=true;
 host.inert=true;assert.equal(editor.focusField(field),null);host.inert=false;host.hidden=true;assert.equal(editor.focusField(field),null);host.hidden=false;
 editor.setSuggestions({fields:[],records:0});load.onclick();assert.equal(loads,0);assert.equal(editor.focusField(field),null);
}));

test('421 attribute action is local, field-qualified and guarded by native and owner readiness',()=>dom(()=>{
 const host=new Element('section');host.root=true;let allowed=true;const calls=[],record={Attributes:{region:''},ResourceAttributes:{region:'west'}};
 Attributes.mount(host,record,{raw:false,onFacetValues:(chosen,opener)=>{calls.push({chosen,opener});return true},canFacetValues:chosen=>allowed&&chosen.kind==='attribute'});
 const facet=label(host,'Facet by attribute:region'),resource=label(host,'Facet by resource:region');assert(facet);assert.equal(resource.disabled,true);facet.onclick();assert.equal(calls.length,0,'closed menu must refuse');
 const menu=facet.closest('details:not([open])');menu.open=true;facet.onclick();assert.deepEqual(calls[0].chosen,field);assert.equal(calls[0].opener,facet);assert.notEqual(calls[0].chosen,field);
 for(const key of ['hidden','inert','disabled']){facet[key]=true;facet.onclick();facet[key]=false;assert.equal(calls.length,1)}
 allowed=false;facet.onclick();assert.equal(calls.length,1);menu.open=false;facet.onclick();assert.equal(calls.length,1);
 assert.equal(record.Attributes.region,'');
}));

test('421 contextual scope preview is text-only and its changed-window lease blocks the existing load',()=>dom(()=>{
 const host=new Element('section');host.root=true;let current=true,loads=0;const editor=Editor.mount(host,{signal:'spans'});editor.setSuggestions(observations);editor.setValueLoader(()=>loads++);
 const text='Latest completed Activity window · 2026-11-13T00:00:00Z → 2026-11-13T00:01:00Z · inspected record window differs';
 const load=editor.focusField(field,{label:text,current:()=>current}),note=walk(host).find(e=>e.className==='filter-facet-scope evidence-note');assert.equal(note.textContent,text);assert.match(note.id,/^filter-fields-\d+-facet-scope$/);assert.equal(load.getAttribute('aria-describedby'),note.id);assert.equal(loads,0);assert.equal(document.activeElement,load);
 current=false;load.onclick();assert.equal(loads,0);assert(walk(host).some(e=>e.textContent==='Scope changed. Choose Facet by… again.'));
 assert.equal(editor.focusField(other,{current:()=>false}),null);current=true;load.onclick();assert.equal(loads,1);
 const resource=editor.focusField(other,{label:'Current resource scope',current:()=>current});assert.equal(load.getAttribute('aria-describedby'),null);assert.equal(resource.getAttribute('aria-describedby'),note.id);assert.equal(note.isConnected,false);
 editor.setSuggestions({rows:0,fields:[]});assert.equal(resource.getAttribute('aria-describedby'),null);
}));
test('421 long observed values do not remove field-level facet discovery',()=>dom(()=>{
 const host=new Element('section');host.root=true;Attributes.mount(host,{Attributes:{region:'x'.repeat(1025)}},{raw:false,onFacetValues(){},canFacetValues:()=>true});
 assert(label(host,'Facet by attribute:region'));assert.equal(label(host,'Include attribute:region from record'),undefined);
}));
