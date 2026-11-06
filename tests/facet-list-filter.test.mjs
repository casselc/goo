import test from 'node:test';
import assert from 'node:assert/strict';
class Element{
 constructor(){this.attributes=new Map();this.rows=[];this.isConnected=false;}
 getAttribute(key){return this.attributes.get(key)??null}hasAttribute(key){return this.attributes.has(key)}
 setAttribute(key,value){const previous=this.getAttribute(key);this.attributes.set(key,String(value));if(this.constructor.observedAttributes?.includes(key))this.attributeChangedCallback(key,previous,String(value));}
 querySelectorAll(){return this.rows;}
}
const registry=new Map();globalThis.HTMLElement=Element;globalThis.customElements={get:key=>registry.get(key),define:(key,value)=>registry.set(key,value)};
await import('../ui/components.mjs');const FacetList=registry.get('goo-facet-list');
const row=text=>({text,hidden:false,getAttribute:key=>key==='data-facet-text'?text:null});
function instrument(list){const refresh=list.refresh;let scans=0;list.refresh=function(){scans++;return refresh.call(this)};return ()=>scans;}
test('atomic filtering scans once, reflects its query and retains exact current row/action identities connected or detached',()=>{
 const previousFetch=globalThis.fetch;globalThis.fetch=()=>{throw Error('A passive facet cannot fetch');};
 try{for(const connected of [false,true]){
  const list=new FacetList();list.isConnected=connected;const counts=instrument(list),rows=[row('Prefix Ω🙂 terminal-one'),row('PREFIX ω🙂 terminal-two'),row('Missing key'),row('Present empty string')];list.rows=rows;
  const before=counts();assert.deepEqual(list.filter('ω🙂'),{visible:2,total:4});assert.equal(counts()-before,1);assert.equal(list.getAttribute('query'),'ω🙂');assert.deepEqual(rows.map(item=>item.hidden),[false,false,true,true]);assert.equal(list.rows,rows);
  list.rows.push(row('Added Ω🙂 terminal-three'));const mutation=counts();assert.deepEqual(list.filter('ω🙂'),{visible:3,total:5});assert.equal(counts()-mutation,1);
  list.rows[0].getAttribute=()=> 'replacement';list.rows.splice(1,1);assert.deepEqual(list.filter('ω🙂'),{visible:1,total:4});assert.equal(list.rows[0],rows[0]);assert.deepEqual(list.filter(''),{visible:4,total:4});
 }}finally{globalThis.fetch=previousFetch;}
});
test('ordinary attribute/property writes and SSR/pre-upgrade properties keep their existing single-pass behavior',()=>{
 const list=new FacetList();list.rows=[row('Alpha'),row('Beta')];list.setAttribute('query','alpha');list.isConnected=true;const scans=instrument(list);list.connectedCallback();assert.equal(scans(),1);assert.deepEqual(list.rows.map(item=>item.hidden),[false,true]);
 list.query='beta';assert.equal(scans(),2);assert.equal(list.getAttribute('query'),'beta');assert.deepEqual(list.rows.map(item=>item.hidden),[true,false]);
 const upgraded=new FacetList();upgraded.rows=[row('Ω'),row('other')];Object.defineProperty(upgraded,'query',{value:'ω',writable:true,configurable:true});upgraded.isConnected=true;const upgradeScans=instrument(upgraded);upgraded.connectedCallback();assert.equal(upgradeScans(),1);assert.equal(Object.hasOwn(upgraded,'query'),false);assert.equal(upgraded.query,'ω');assert.deepEqual(upgraded.rows.map(item=>item.hidden),[false,true]);
});
