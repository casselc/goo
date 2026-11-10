import {test} from 'node:test';
import assert from 'node:assert/strict';
class Element{
 constructor(){this.attributes=new Map();this.listeners=new Map();this.scrollWidth=500;this.clientWidth=200;this.scrollLeft=0;this.table={identity:'same table'};this.isConnected=true}
 hasAttribute(key){return this.attributes.has(key)}getAttribute(key){return this.attributes.get(key)??null}setAttribute(key,value){this.attributes.set(key,String(value))}removeAttribute(key){this.attributes.delete(key)}
 querySelector(){return this.table}addEventListener(name,fn){this.listeners.set(name,fn)}removeEventListener(name,fn){if(this.listeners.get(name)===fn)this.listeners.delete(name)}
}
const resize=[],mutations=[],registry=new Map();
class Resize{constructor(fn){this.fn=fn;this.nodes=new Set();resize.push(this)}observe(node){this.nodes.add(node)}unobserve(node){this.nodes.delete(node)}disconnect(){this.nodes.clear();this.disconnected=true}}
class Mutation{constructor(fn){this.fn=fn;mutations.push(this)}observe(node,options){this.node=node;this.options=options}disconnect(){this.disconnected=true}}
globalThis.HTMLElement=Element;globalThis.ResizeObserver=Resize;globalThis.MutationObserver=Mutation;globalThis.customElements={get:key=>registry.get(key),define:(key,value)=>registry.set(key,value)};
await import('../ui/components.mjs');const Table=registry.get('goo-table');
test('402 overflow directions follow local scrolling without replacing rows or reading data',()=>{
 const old=globalThis.fetch;globalThis.fetch=()=>{throw Error('Passive table cannot fetch')};
 try{const table=new Table(),content=table.table;table.setAttribute('aria-label','Native export intervals');table.connectedCallback();assert.equal(table.getAttribute('data-overflow-inline'),'end');assert.equal(table.getAttribute('aria-label'),'Native export intervals');assert.equal(table.tabIndex,0);assert.equal(table.getAttribute('role'),'region');
 for(const [left,state]of [[100,'both'],[300,'start'],[10000,'start'],[-100,'both'],[0,'end']]){table.scrollLeft=left;table.listeners.get('scroll')();assert.equal(table.getAttribute('data-overflow-inline'),state)}
 table.clientWidth=500;resize.at(-1).fn();assert.equal(table.getAttribute('data-overflow-inline'),null);assert.equal(table.table,content);table.scrollWidth=500.5;assert.equal(table.refreshOverflow(),'');table.disconnectedCallback();
 }finally{globalThis.fetch=old}
});
test('402 late SSR contents, content resize, removal and reconnect retain observer ownership',()=>{
 const table=new Table();table.table=null;table.setAttribute('tabindex','-1');table.setAttribute('aria-labelledby','table-heading');table.connectedCallback();const first=resize.at(-1),mutation=mutations.at(-1);assert.equal(table.getAttribute('tabindex'),'-1');assert.equal(table.hasAttribute('aria-label'),false);assert.equal(mutation.options.attributes,undefined,'derived attributes must not trigger mutation loops');
 const content={id:1};table.table=content;mutation.fn();assert(first.nodes.has(content));table.table={id:2};mutation.fn();assert(!first.nodes.has(content));assert(first.nodes.has(table.table));table.disconnectedCallback();assert(first.disconnected);assert(mutation.disconnected);assert.equal(table.listeners.size,0);assert.equal(table.getAttribute('data-overflow-inline'),null);
 table.connectedCallback();const next=resize.at(-1);assert.notEqual(next,first);assert(next.nodes.has(table.table));assert.equal(table.getAttribute('data-overflow-inline'),'end');assert.equal(table.listeners.size,1);table.disconnectedCallback();
});
