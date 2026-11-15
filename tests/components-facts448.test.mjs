import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Element} from './support/raw-observation-dom.mjs';

test('448 shared typed facts retain native lifecycle, exact clocks and compact presentation without IO',async t=>{
 const keys=['document','HTMLElement','customElements','fetch'],prior=Object.fromEntries(keys.map(k=>[k,globalThis[k]])),registry=new Map();let reads=0;
 class HTMLElement extends Element{
  constructor(){super('custom-element')}
  setAttribute(key,value){const old=this.getAttribute(key);super.setAttribute(key,value);if(old!==String(value)&&this.constructor.observedAttributes?.includes(key))this.attributeChangedCallback?.(key,old,String(value))}
 }
 globalThis.HTMLElement=HTMLElement;
 globalThis.customElements={get:name=>registry.get(name),define:(name,value)=>registry.set(name,value)};
 globalThis.document={activeElement:null,createElement(tag){const C=registry.get(tag),e=C?new C():new Element(tag);e.localName=tag;return e}};
 globalThis.fetch=()=>{reads++;throw Error('Display facts may not acquire data')};
 t.after(()=>{for(const key of keys)if(prior[key]===undefined)delete globalThis[key];else globalThis[key]=prior[key]});
 const {stats}=await import('../ui/components.mjs?facts448');
 const connect=e=>{e.root=true;e.connectedCallback();return e};
 const values=e=>e.querySelectorAll('dd');
 await t.test('detached properties, native UTC time and updates preserve the complete source instant',()=>{
  const input=[{label:'Read UTC',value:'2026-11-15T05:31:18.793761062Z',format:'utc'}],e=stats(input,{density:'compact'});
  assert.equal(e.getAttribute('density'),'compact');assert.equal(e.children.length,0);input[0].value='changed';connect(e);
  const time=e.querySelector('time');assert.equal(time.textContent,'2026-11-15 05:31:18.793 UTC');assert.equal(time.dateTime,'2026-11-15T05:31:18.793761062Z');assert.equal(time.title,time.dateTime);assert.equal(time.tabIndex,0);
  e.root=false;e.items=[{label:'Event UTC',value:'1794720678793761062',format:'unix-ns'}];assert.equal(e.querySelector('time'),time);connect(e);
  const event=e.querySelector('time');assert.equal(event.dateTime,'2026-11-15T05:31:18.793761062Z');assert.match(event.title,/1794720678793761062 ns since Unix epoch/);assert.equal(event.tabIndex,0);
  const detached=e.items;detached[0].value='0';assert.equal(e.items[0].value,'1794720678793761062');
 });
 await t.test('SSR JSON items and typed attributes render without an application adapter',()=>{
  const e=document.createElement('goo-stat-list');e.setAttribute('density','compact');e.setAttribute('items',JSON.stringify([{label:'Event',value:'1000000001',format:'unix-ns'}]));connect(e);
  assert.equal(e.querySelector('dl').className.trim(),'ui-stats');assert.equal(e.querySelector('time').dateTime,'1970-01-01T00:00:01.000000001Z');assert.match(e.querySelector('time').title,/1000000001 ns/);
  e.setAttribute('items',JSON.stringify([{label:'Count',value:0}]));assert.equal(values(e)[0].textContent,'0');assert.equal(e.querySelector('time'),null);
 });
 await t.test('pre-definition own properties upgrade through the custom element lifecycle',()=>{
  const e=new Element('goo-stat-list');e.items=[{label:'Read UTC',value:'2026-11-15T00:00:00Z',format:'utc'}];Object.setPrototypeOf(e,registry.get('goo-stat-list').prototype);connect(e);
  assert.equal(Object.hasOwn(e,'items'),false);assert.equal(e.querySelector('time').dateTime,'2026-11-15T00:00:00Z');
 });
 await t.test('half-open ranges render two exact boundaries and retain dates across midnight',()=>{
  const e=connect(stats([{label:'Window',value:'[1000000001, 2000000009)',format:'unix-ns-range'}],{density:'compact'}));let clocks=e.querySelectorAll('time');
  assert.deepEqual(clocks.map(x=>x.dateTime),['1970-01-01T00:00:01.000000001Z','1970-01-01T00:00:02.000000009Z']);assert.equal(values(e)[0].textContent,'[1970-01-01 00:00:01.000, 00:00:02.000) UTC');assert.match(values(e)[0].firstChild.title,/\[1000000001, 2000000009\) ns/);
  e.items=[{label:'Window',value:'[86399999999999, 86400000000001)',format:'unix-ns-range'}];clocks=e.querySelectorAll('time');assert.match(clocks[0].textContent,/1970-01-01/);assert.match(clocks[1].textContent,/1970-01-02/);assert.equal(clocks[1].dateTime,'1970-01-02T00:00:00.000000001Z');
 });
 await t.test('bad typed facts are explicitly unavailable while plain counts and text stay unchanged',()=>{
  const bad=[{format:'utc',value:'2026-02-30T00:00:00Z'},{format:'utc',value:'bad'},{format:'unix-ns',value:123},{format:'unix-ns',value:'1e9'},{format:'unix-ns-range',value:'[2, 1)'},{format:'unix-ns-range',value:'[1, 2]'},{format:'html',value:'<img src=x>'},{format:'utc',value:null}];
  const e=connect(stats([...bad.map((item,i)=>({label:'Bad '+i,...item})),{label:'Exact count',value:'18446744073709551615'},{label:'Zero',value:0},{label:'Text',value:'<script>text only</script>'}]));
  assert.deepEqual(values(e).slice(0,bad.length).map(x=>x.textContent),bad.map(()=> 'Unavailable'));assert.equal(e.querySelector('time'),null);assert.deepEqual(values(e).slice(bad.length).map(x=>x.textContent),['18446744073709551615','0','<script>text only</script>']);assert.equal(e.querySelector('script'),null);assert.equal(e.querySelector('img'),null);
  assert.throws(()=>stats([{label:'Node',value:document.createElement('strong')}]),/clone|cloned/i);
 });
 assert.equal(reads,0);
});
