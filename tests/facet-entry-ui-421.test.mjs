import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Attributes from '../ui/attributes.mjs';
import * as Editor from '../ui/filter-editor.mjs';

class Element{
 constructor(tag){this.localName=tag;this.style={};this.classList={add:(...tokens)=>{this.className=[...(this.className||'').split(' ').filter(Boolean),...tokens].join(' ')}};this.children=[];this.attrs={};this.dataset={};this.value='';this.textContent='';this.open=false;this.hidden=false;this.disabled=false;this.inert=false;this.root=false;}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}
 append(...nodes){for(const node of nodes){if(typeof node==='object'){node.remove();node.parentElement=this;}this.children.push(node)}if(this.localName==='select'&&!this.value)this.value=this.children.find(c=>c.value)?.value||''}
 prepend(...nodes){for(const node of nodes){node.remove();node.parentElement=this;}this.children.unshift(...nodes)}
 before(node){const p=this.parentElement;node.remove();node.parentElement=p;p.children.splice(p.children.indexOf(this),0,node)}
 remove(){if(this.parentElement){const p=this.parentElement;this.parentElement=null;p.children=p.children.filter(c=>c!==this)}}
 replaceChildren(...nodes){for(const child of [...this.children])if(typeof child==='object')child.remove();this.children=[];this.append(...nodes)}
 setAttribute(key,value){this.attrs[key]=String(value)}getAttribute(key){return this.attrs[key]??null}removeAttribute(key){delete this.attrs[key]}
 addEventListener(){}
 contains(node){return node===this||this.children.some(c=>typeof c==='object'&&c.contains(node))}
 matches(selector){const option=selector.match(/^option\[value="([^"]+)"\]$/);if(option)return this.localName==='option'&&this.value===option[1];if(selector==='details:not([open])')return this.localName==='details'&&!this.open;if(selector===':disabled'){if(this.disabled)return true;for(let p=this.parentElement;p;p=p.parentElement)if(p.localName==='fieldset'&&p.disabled&&!p.querySelector(':scope > legend')?.contains(this))return true;return false}if(selector==='details[open]')return this.localName==='details'&&this.open;return this.localName===selector}
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
 const load=editor.focusField(field,{label:text,current:()=>current}),note=walk(host).find(e=>e.className==='filter-facet-scope evidence-note');assert.equal(note.children[0].textContent,'Latest completed Activity window · 2026-11-13 00:00:00–00:01:00 UTC · inspected record window differs');assert.equal(note.children[1].textContent,text);assert.match(note.id,/^filter-fields-\d+-facet-scope$/);assert.equal(load.getAttribute('aria-describedby'),note.id);assert.equal(loads,0);assert.equal(document.activeElement,load);
 current=false;load.onclick();assert.equal(loads,0);assert(walk(host).some(e=>e.textContent==='Scope changed. Choose Facet by… again.'));
 assert.equal(editor.focusField(other,{current:()=>false}),null);current=true;load.onclick();assert.equal(loads,1);
 const resource=editor.focusField(other,{label:'Current resource scope',current:()=>current});assert.equal(load.getAttribute('aria-describedby'),null);assert.equal(resource.getAttribute('aria-describedby'),note.id);assert.equal(note.isConnected,false);
 editor.setSuggestions({rows:0,fields:[]});assert.equal(resource.getAttribute('aria-describedby'),null);
}));
test('421 long observed values do not remove field-level facet discovery',()=>dom(()=>{
 const host=new Element('section');host.root=true;Attributes.mount(host,{Attributes:{region:'x'.repeat(1025)}},{raw:false,onFacetValues(){},canFacetValues:()=>true});
 assert(label(host,'Facet by attribute:region'));assert.equal(label(host,'Include attribute:region from record'),undefined);
}));

const appliedQuestion=where=>({signal:'spans',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'production'}},fromMs:1000,toMs:61000,limit:100,offset:0,where});
const region=(value,qualified=field)=>({field:structuredClone(qualified),op:'eq',value});
const constraintPanel=host=>walk(host).find(node=>node.className==='filter-field-constraints');

test('424 applied field constraints target the exact OR condition through the existing editor',()=>dom(()=>{
 const host=new Element('section');host.root=true;const where={all:[region('west'),region('resource',other),{any:[region('east'),region('south')]}]},question=appliedQuestion(where),before=structuredClone(question),changes=[];let loads=0;
 const editor=Editor.mount(host,{where,signal:'spans',onChange:(...args)=>changes.push(args)});editor.setSuggestions(observations);editor.setValueLoader(()=>loads++);
 const load=editor.focusField(field,{question}),panel=constraintPanel(host);assert(load);assert(panel);assert.equal(panel.querySelector('summary').getAttribute('aria-label'),'Applied constraints (3) · attribute:region');assert.equal(panel.open,false);panel.open=true;
 assert(label(host,'Remove condition 1 · attribute:region'));assert.equal(label(host,'Remove condition 2 · attribute:region'),undefined);assert.equal(label(host,'Remove condition 3 · attribute:region'),undefined);
 const edit=label(host,'Edit condition 2 · attribute:region');edit.onclick();assert.equal(document.activeElement,label(host,'Condition 5 operator'),'the second matching field is the first OR leaf, not the resource or first condition');
 assert.equal(document.activeElement.value,'eq');assert.equal(loads,0);assert.equal(changes.length,0);assert.deepEqual(editor.snapshot(),where);assert.deepEqual(question,before);
 document.activeElement.value='ne';document.activeElement.onchange();assert.equal(changes.length,1);assert.equal(editor.snapshot().all[2].any[0].op,'ne');assert.equal(editor.snapshot().all[0].op,'eq');assert.deepEqual(editor.snapshot().all[1],region('resource',other));assert.equal(constraintPanel(host),undefined,'editing retires the applied-constraint affordances');
}));

test('424 explicit conjunctive removal stages once, keeps other predicates and retires counted results through the owner',()=>dom(()=>{
 const host=new Element('section');host.root=true;const where={all:[region('west'),region('resource',other),{any:[region('east'),region('south')]}]},question=appliedQuestion(where),events=[];let editor,loads=0,stages=0;
 editor=Editor.mount(host,{where,signal:'spans',onChange:(...args)=>{events.push(args);editor.setCountedValues()},onConstraintStage:()=>stages++});editor.setSuggestions(observations);editor.setValueLoader(()=>loads++);editor.focusField(field,{question});const panel=constraintPanel(host);panel.open=true;
 editor.setCountedValues({state:'ready',question:{...question,groupBy:[field]},field,items:[{value:'west',count:'1'}],total:'1',readAt:'2026-11-13T00:00:00Z',basis:'held',label:'Original scope'});
 const counted=walk(host).find(node=>node.className==='filter-counted-values');assert.equal(counted.hidden,false);
 const remove=label(host,'Remove condition 1 · attribute:region'),old=remove.onclick;remove.onclick();assert.equal(events.length,1);assert.deepEqual(events[0],[{all:[region('resource',other),{any:[region('east'),region('south')]}]},null]);assert.equal(stages,1);assert.equal(loads,0);assert.equal(counted.hidden,true);assert.equal(constraintPanel(host),undefined);
 old.call(remove);assert.equal(events.length,1);assert.equal(stages,1);assert.equal(loads,0);assert.deepEqual(question,appliedQuestion(where));
}));

test('424 stale and withdrawn contextual controls cannot edit, remove, stage or reveal',()=>dom(()=>{
 const host=new Element('section');host.root=true;const where={all:[region('west'),region('resource',other)]},question=appliedQuestion(where);let current=true,mutable=true,changes=0,stages=0;
 const editor=Editor.mount(host,{where,signal:'spans',canMutate:()=>mutable,onChange:()=>changes++,onConstraintStage:()=>stages++});editor.setSuggestions(observations);editor.focusField(field,{question,current:()=>current});const panel=constraintPanel(host);panel.open=true;const edit=label(host,'Edit condition 1 · attribute:region'),remove=label(host,'Remove condition 1 · attribute:region'),before=editor.snapshot(),focus={};
 const deny=()=>{document.activeElement=focus;edit.onclick();remove.onclick();assert.deepEqual(editor.snapshot(),before);assert.equal(changes,0);assert.equal(stages,0);assert.equal(document.activeElement,focus)};
 panel.open=false;deny();assert.equal(panel.open,false);panel.open=true;
 for(const mode of ['hidden','inert']){host[mode]=true;deny();host[mode]=false;}
 for(const mode of ['disabled','hidden','inert']){edit[mode]=remove[mode]=true;deny();edit[mode]=remove[mode]=false;}
 edit.setAttribute('aria-disabled','true');remove.setAttribute('aria-disabled','true');deny();edit.removeAttribute('aria-disabled');remove.removeAttribute('aria-disabled');
 const observed=walk(host).find(node=>node.className==='filter-observations');observed.open=false;deny();assert.equal(observed.open,false);observed.open=true;
 current=false;deny();current=true;mutable=false;deny();mutable=true;document.hidden=true;deny();document.hidden=false;
 editor.set(structuredClone(where),'spans');deny();assert.equal(panel.isConnected,false);
}));

test('424 contextual constraints require the whole authored predicate and accept normalized present-empty forms',()=>dom(()=>{
 for(const authoredEmpty of [false,true])for(const suppliedEmpty of [false,true]){
  const host=new Element('section');host.root=true;const empty=explicit=>({field:structuredClone(field),op:'eq',...(explicit?{value:''}:{})}),where={all:[empty(authoredEmpty),region('resource',other)]},question=appliedQuestion({all:[empty(suppliedEmpty),region('resource',other)]}),before=structuredClone({where,question}),events=[];
  const editor=Editor.mount(host,{where,signal:'spans',onChange:(...args)=>events.push(args)});editor.setSuggestions(observations);
  assert(editor.focusField(field,{question:{...question,where:region('other')}}));assert.equal(constraintPanel(host),undefined);
  assert(editor.focusField(field,{question:{...question,where:{all:[...question.where.all].reverse()}}}));assert.equal(constraintPanel(host),undefined,'authored clause reorder must not retarget an ordinal');
  assert(editor.focusField(field,{question}));const panel=constraintPanel(host);assert(panel);panel.open=true;const code=walk(panel).find(node=>node.localName==='code');assert.match(code.textContent,/attribute:region (eq|equals) ""$/);assert(!code.textContent.includes('undefined'));
  label(host,'Edit condition 1 · attribute:region').onclick();assert.equal(document.activeElement,label(host,'Condition 2 operator'));assert.equal(events.length,0);
  label(host,'Remove condition 1 · attribute:region').onclick();assert.deepEqual(events,[[{all:[region('resource',other)]},null]]);assert.deepEqual(editor.snapshot(),{all:[region('resource',other)]});assert.deepEqual({where,question},before);
 }
}));

test('424 normalized integer scale-zero predicates remain editable and removable',()=>dom(()=>{
 const host=new Element('section');host.root=true;const numeric={numeric:{field:structuredClone(field),type:'int64',unit:'1',parse:'canonical-decimal',scale:0},op:'gte',value:'5'},where={all:[numeric,region('resource',other)]},question=appliedQuestion(structuredClone(where));delete question.where.all[0].numeric.scale;const events=[];
 const editor=Editor.mount(host,{where,signal:'spans',onChange:(...args)=>events.push(args)});editor.setSuggestions(observations);assert(editor.focusField(field,{question}));const panel=constraintPanel(host);assert(panel);panel.open=true;
 label(host,'Edit condition 1 · attribute:region').onclick();assert.equal(document.activeElement,label(host,'Condition 2 operator'));assert.equal(events.length,0);
 label(host,'Remove condition 1 · attribute:region').onclick();assert.deepEqual(events,[[{all:[region('resource',other)]},null]]);assert.equal(numeric.numeric.scale,0);assert.equal(question.where.all[0].numeric.scale,undefined);
}));
