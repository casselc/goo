const {test}=require('node:test'),assert=require('node:assert/strict'),{mount}=require('../ui/attributes.mjs');
class Element{
 constructor(tag){this.tagName=tag;this.children=[];this.dataset={};this.value='';this.textContent='';this.attributes={};this.disabled=false;}
 get isConnected(){return this.connected!==false&&(!this.parentElement||this.parentElement.isConnected)}
 append(...children){for(const child of children){if(child?.parentElement)child.parentElement.children=child.parentElement.children.filter(n=>n!==child);if(child instanceof Element)child.parentElement=this;this.children.push(child)}if(this.tagName==='select'&&!this.value)this.value=children.find(child=>child?.value)?.value||''}
 prepend(...children){this.append(...children);this.children=[...children,...this.children.filter(child=>!children.includes(child))]}
 replaceChildren(...children){for(const child of this.children)if(child instanceof Element)child.parentElement=null;this.children=[];this.append(...children)}
 setAttribute(key,value){this.attributes[key]=value}getAttribute(key){return this.attributes[key]}
 addEventListener(){}
 closest(selector){for(let n=this;n;n=n.parentElement){if(selector==='.ui-action-menu'&&n.className?.split(' ').includes('ui-action-menu')||selector==='details:not([open])'&&n.tagName==='details'&&!n.open||selector==='[hidden], [aria-hidden="true"]'&&(n.hidden||n.getAttribute('aria-hidden')==='true')||selector==='[inert]'&&n.inert||selector==='fieldset:disabled'&&n.tagName==='fieldset'&&n.disabled)return n}return null}
 querySelector(selector){return this.children.find(n=>n.tagName===(selector===':scope > legend'?'legend':'summary'))??null}
 contains(element){for(let n=element;n;n=n.parentElement)if(n===this)return true;return false}
 getClientRects(){return this.rect===false?[]:[{}]}
 matches(selector){return selector===':disabled'&&this.disabled}
}
const descendants=node=>[node,...node.children.filter(child=>child instanceof Element).flatMap(descendants)];
test('attribute inspection reports host refusal truthfully while ordinary void/true callbacks stage normally',()=>{
 const previous=global.document;global.document={createElement:tag=>new Element(tag)};
 try{for(const accepted of [false,undefined,true]){let observed=null;const host=new Element('section');mount(host,{Attributes:{'http.route':'/checkout'}},{raw:false,onFacet:(field,value,mode)=>{observed={field,value,mode};return accepted}});const all=descendants(host),button=all.find(node=>node.getAttribute('aria-label')==='Include attribute:http.route from record'),status=all.find(node=>node.className==='attribute-feedback');assert(button);assert.equal(button.disabled,false);button.closest('details:not([open])').open=true;button.onclick();assert.deepEqual(observed,{field:{kind:'attribute',key:'http.route'},value:'/checkout',mode:'include'});assert.equal(status.textContent,accepted===false?'Not staged.':'Staged include for attribute:http.route. Apply or Open in the selection tray; displayed measurements are unchanged.');}}
 finally{global.document=previous}
});


test('attribute rows keep Copy primary and put qualified operations in a closed native disclosure by default',()=>{
 const previous=global.document;global.document={createElement:tag=>new Element(tag)};
 try{
  let chosen,plotOrigin,traceOrigin,invoked=0;const record={Attributes:{'work.region':'eu-west',large:'9007199254740993'}},before=structuredClone(record),host=new Element('section');
  mount(host,record,{raw:false,onFacet:(field,value,mode)=>{invoked++;chosen={field,value,mode}},onFilter(){invoked++},onBreakdown(){invoked++},onPlot(field,opener){invoked++;plotOrigin={field,opener}},onTraceSearch(seed,opener){invoked++;traceOrigin={seed,opener}}});
  const all=descendants(host),row=all.find(e=>e.getAttribute('aria-label')==='Span attributes: work.region'),actions=descendants(row).find(e=>e.className==='attribute-actions'),group=actions.children[0],copy=group.children[0],menu=group.children[1];
  assert.equal(copy.getAttribute('aria-label'),'Copy value for work.region');assert.equal(copy.textContent,'Copy');assert.equal(menu.tagName,'details');assert.equal(menu.open,false);assert.equal(menu.children[0].textContent,'⋯');assert.equal(menu.children[0].getAttribute('aria-label'),'Field actions');
  const secondary=descendants(menu).filter(e=>e.tagName==='button');assert.equal(secondary.length,6);assert(secondary.every(button=>button!==copy));
  for(const button of secondary)button.onclick();assert.equal(invoked,0,'closed menu callbacks cannot stage or query');
  const include=secondary.find(e=>e.getAttribute('aria-label')==='Include attribute:work.region from record');menu.open=true;include.onclick();assert.deepEqual(chosen,{field:{kind:'attribute',key:'work.region'},value:'eu-west',mode:'include'});assert.equal(invoked,1);
  for(const blocked of ['disabled','hidden','inert']){include[blocked]=true;include.onclick();include[blocked]=false;assert.equal(invoked,1,'unavailable action cannot stage: '+blocked)}menu.open=false;include.onclick();assert.equal(invoked,1,'captured callback refuses after menu closes');
  menu.open=true;secondary.find(e=>e.getAttribute('aria-label')==='Plot values of attribute:work.region').onclick();assert.equal(plotOrigin.opener,menu.children[0]);assert.deepEqual(plotOrigin.field,{kind:'attribute',key:'work.region'});secondary.find(e=>e.getAttribute('aria-label')==='Find traces containing attribute:work.region').onclick();assert.equal(traceOrigin.opener,menu.children[0]);assert.deepEqual(traceOrigin.seed,{field:{kind:'attribute',key:'work.region'},op:'eq',value:'eu-west'});menu.open=false;assert.equal(plotOrigin.opener,menu.children[0],'return origin survives native menu close');
  assert.equal(descendants(host).find(e=>e.tagName==='goo-field-value'&&e.value==='9007199254740993').value,'9007199254740993');assert.deepEqual(record,before);
 }finally{global.document=previous}
});
