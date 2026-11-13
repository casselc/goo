import {test,mock} from 'node:test';
import assert from 'node:assert/strict';

// Exercise the real dashboard and native-readiness helpers. Expansion validation
// belongs to view-template-handoff tests; this seam controls completion timing.
mock.module('../ui/view-template-handoff.mjs',{exports:{
 eligible:()=>true,customizeLabel:()=> 'widget view',templateDestination:w=>({query:structuredClone(w.query)}),
 async expand(widget,{current,post}){if(!current())return null;const result=await post({widget});return current()?result.view:null}
}});
const {Element,node,fixture,observed,tick,walk}=await import('./support/dashboard-owner-354.mjs');
const parentMatches=(element,selector)=>selector==='details:not([open])'?element.localName==='details'&&!element.open:selector==='fieldset:disabled'?element.localName==='fieldset'&&element.disabled:selector==='[inert]'?element.inert:element.hidden||element.getAttribute('aria-hidden')==='true';
Element.prototype.closest=function(selector){for(let element=this;element;element=element.parentElement)if(parentMatches(element,selector))return element;return null};
Element.prototype.querySelector=function(selector){return this.children.find(element=>element?.localName===(selector===':scope > legend'?'legend':'summary'))||null};
Element.prototype.matches=function(selector){if(selector!==':disabled')return false;if(this.disabled)return true;for(let fieldset=this.parentElement?.closest('fieldset:disabled');fieldset;fieldset=fieldset.parentElement?.closest('fieldset:disabled'))if(!fieldset.querySelector(':scope > legend')?.contains(this))return true;return false};

async function setup({http=false,canLeave=()=>true}={}){
 const callbacks=[],events=[],f=fixture(1,{canLeave,onCustomize:(...args)=>callbacks.push(['customize',...args]),onUseTemplate:(...args)=>callbacks.push(['reuse',...args])});
 if(http){const widget=f.plan.widgets[0];widget.key='http-server-responses';widget.family='http-server';widget.query.signal='spans';widget.query.groupBy=[{kind:'attribute',key:'http.response.status_code'}];}
 globalThis.dispatchEvent=event=>events.push(event);await tick();f.requests[0].resolve(observed(f.requests[0]));const owner=await f.loaded,box=f.box(f.plan.widgets[0].key);
 const find=text=>walk(box).find(element=>element.textContent===text),controls=['Open widget question','Add widget question','Capture widget evidence','Use saved template','Customize view'].map(find);
 assert(controls.every(Boolean));return {...f,owner,box,controls,callbacks,events,find};
}
function disclosure(f){const menu=node('details'),summary=node('summary','Widget actions');menu.open=true;menu.append(summary);f.host.append(menu);f.box.remove();menu.append(f.box);return menu}
const click=control=>control.onclick({preventDefault(){},stopImmediatePropagation(){}});

test('420 ordinary widget actions refuse closed, disabled, inert, hidden and retired controls',async()=>{
 const f=await setup(),menu=disclosure(f),fieldset=node('fieldset');menu.append(fieldset);f.box.remove();fieldset.append(f.box);
 const states=[
  [()=>menu.open=false,()=>menu.open=true],
  [()=>fieldset.disabled=true,()=>fieldset.disabled=false],
  [()=>f.box.inert=true,()=>f.box.inert=false],
  [()=>f.host.hidden=true,()=>f.host.hidden=false],
  [()=>document.hidden=true,()=>document.hidden=false]
 ];
 for(const [withdraw,restore]of states){withdraw();for(const control of f.controls)await click(control);restore();}
 for(const control of f.controls){control.disabled=true;await click(control);control.disabled=false;control.setAttribute('aria-disabled','true');await click(control);control.setAttribute('aria-disabled','false');}
 assert.equal(f.requests.length,1);assert.equal(f.events.length,0);assert.equal(f.callbacks.length,0);
 assert.equal(click(f.controls[0]),true,'visible native link may navigate');click(f.controls[1]);click(f.controls[2]);click(f.controls[3]);assert.equal(f.events.length,2);assert.equal(f.callbacks.length,1);assert.equal(f.callbacks[0][2],f.controls[3]);
 f.retire();for(const control of f.controls)await click(control);assert.equal(f.requests.length,1);assert.equal(f.events.length,2);assert.equal(f.callbacks.length,1);f.owner.dispose();
});

test('420 a valid asynchronous customization completes through its own disabled opener only once',async()=>{
 const f=await setup(),button=f.find('Customize view'),pending=click(button);assert.equal(button.disabled,true);assert.equal(f.requests.length,2);await click(button);assert.equal(f.requests.length,2);
 f.requests[1].resolve({ok:true,data:{view:{title:'Prepared view'}}});await pending;assert.equal(f.callbacks.length,1);assert.equal(f.callbacks[0][0],'customize');assert.equal(f.callbacks[0][2],button);assert.equal(button.disabled,false);f.owner.dispose();
});

test('420 closed or retired asynchronous customization cannot hand off a late configuration',async()=>{
 for(const withdrawal of ['closed','inert','fieldset','aria-disabled','detached','retired']){
  const f=await setup(),menu=disclosure(f),button=f.find('Customize view'),pending=click(button);assert.equal(f.requests.length,2);
  if(withdrawal==='closed')menu.open=false;
  if(withdrawal==='inert')f.box.inert=true;
  if(withdrawal==='fieldset'){const fieldset=node('fieldset');menu.append(fieldset);f.box.remove();fieldset.append(f.box);fieldset.disabled=true;}
  if(withdrawal==='aria-disabled')button.setAttribute('aria-disabled','true');
  if(withdrawal==='detached')button.remove();
  if(withdrawal==='retired')f.retire();
  f.requests[1].resolve({ok:true,data:{view:{title:'Late view'}}});await pending;assert.equal(f.callbacks.length,0,withdrawal+' withdraws handoff');assert.equal(f.requests.length,2);f.owner.dispose();
 }
});


test('420 HTTP customization completion rechecks parent departure while allowing its disabled opener',async()=>{
 for(const withdraw of [false,true]){
  let allowed=true;const f=await setup({http:true,canLeave:()=>allowed}),context=walk(f.host).find(element=>element.className==='dashboard-http-server-context');assert(context);context.open=true;
  const button=f.find('Customize view'),pending=click(button);assert.equal(button.disabled,true);assert.equal(f.requests.length,2);await click(button);assert.equal(f.requests.length,2);
  if(withdraw)allowed=false;
  f.requests[1].resolve({ok:true,data:{view:{title:'Prepared HTTP context'}}});await pending;
  assert.equal(f.callbacks.length,withdraw?0:1,'departure permission is checked independently from the disabled opener');
  assert.equal(button.disabled,false);f.owner.dispose();
 }
});
