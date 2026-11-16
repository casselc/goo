import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Element} from './support/raw-observation-dom.mjs';
import {initial,transition,withContext} from '../ui/guidance-content.mjs';
const registry=new Map(),events=[];
class NativeElement extends Element{
 constructor(tag='element'){super(tag)}
 setAttribute(key,value){const old=this.getAttribute(key);super.setAttribute(key,value);if(this.constructor.observedAttributes?.includes(key)&&old!==String(value))this.attributeChangedCallback?.(key,old,String(value))}
 dispatchEvent(event){events.push({target:this,event});return true}
}
globalThis.HTMLElement=NativeElement;globalThis.customElements={get:key=>registry.get(key),define:(key,value)=>registry.set(key,value)};
globalThis.CustomEvent=class{constructor(type,init){this.type=type;Object.assign(this,init)}};
globalThis.document={hidden:false,activeElement:null,createElement:tag=>new NativeElement(tag)};
await import('../ui/guidance-view.mjs');
const View=registry.get('goo-guidance');
function mount(state=initial()){const view=new View();view.data=state;view.root=true;view.connectedCallback();return view}
const action=(view,name)=>view.querySelectorAll('button').find(button=>button.dataset.guidanceAction===name);
test('help topics are native and closed, all interactions emit only explicit reading intents from actual button',()=>{
 const view=mount();assert.equal(view.querySelector('h2').tabIndex,-1);assert.equal(view.querySelectorAll('details').length,3);assert.ok(view.querySelectorAll('details').every(item=>!item.open));
 const start=action(view,'start-guide');start.onclick();assert.equal(events.at(-1).target,start);assert.equal(events.at(-1).event.type,'goo-guidance-action');assert.deepEqual(events.at(-1).event.detail,{action:'start-guide'});assert.equal(view.data.guidePage,null);
 action(view,'work').onclick();assert.deepEqual(events.at(-1).event.detail,{action:'work'});assert.equal(events.at(-1).event.composed,true);
});
test('manual reading pages have bounded controls, no completion claim and context changes keep position',()=>{
 const view=mount(transition(initial(),'start-guide'));assert.match(view.textContent,/Guide page 1 of 5/);assert.equal(action(view,'previous').disabled,true);assert.equal(action(view,'next').disabled,false);
 let state=view.data;for(let i=0;i<4;i++)state=transition(state,'next');view.data=withContext(state,'services');assert.match(view.textContent,/Guide page 5 of 5/);assert.equal(action(view,'next').disabled,true);assert.match(view.textContent,/Look for/);assert.doesNotMatch(view.textContent,/task completed|you completed|Done/);
});
test('hidden, detached, disabled and stale handlers cannot emit intents',()=>{
 const view=mount(),start=action(view,'start-guide'),before=events.length;view.data=initial('saved');start.onclick();assert.equal(events.length,before);
 const current=action(view,'start-guide');view.hidden=true;current.onclick();view.hidden=false;view.root=false;current.onclick();view.root=true;current.disabled=true;current.onclick();assert.equal(events.length,before);
});
test('SSR and pre-upgrade values work, malformed state safely refuses and source work never occurs',()=>{
 let reads=0;globalThis.fetch=()=>{reads++;throw Error('No IO')};const view=new View();view.setAttribute('data',JSON.stringify(initial('services')));view.root=true;view.connectedCallback();assert.match(view.textContent,/Help · Services/);
 const early=new View();Object.defineProperty(early,'data',{configurable:true,writable:true,value:initial('saved')});early.root=true;early.connectedCallback();assert.match(early.textContent,/Help · Saved/);
 view.data={...initial(),extra:true};assert.equal(view.dataset.state,'unavailable');assert.match(view.textContent,/invalid reading state/);assert.equal(action(view,'start-guide'),undefined);view.data=initial();assert.equal(view.dataset.state,'ready');assert.equal(reads,0);delete globalThis.fetch;
});
test('updates preserve reading control focus only, never acquire focus from app controls',()=>{
 const view=mount();const external=new NativeElement('button');document.activeElement=external;view.data=initial('saved');assert.equal(document.activeElement,external);
 action(view,'start-guide').focus();view.data=initial('services');assert.equal(document.activeElement,action(view,'start-guide'));
});

test('removed or newly disabled reading buttons restore focus to the guide heading',()=>{
 const view=mount();action(view,'start-guide').focus();view.data=transition(view.data,'start-guide');assert.equal(document.activeElement,view.querySelector('h2'));
 let state=view.data;for(let i=0;i<3;i++)state=transition(state,'next');view.data=state;action(view,'next').focus();view.data=transition(state,'next');assert.equal(document.activeElement,view.querySelector('h2'));
 action(view,'about').focus();view.data=transition(view.data,'about');assert.equal(document.activeElement,view.querySelector('h2'));
});
