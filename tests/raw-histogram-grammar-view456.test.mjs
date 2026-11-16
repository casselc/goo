import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Element} from './support/raw-observation-dom.mjs';
import {readFileSync} from 'node:fs';
const bank=JSON.parse(JSON.parse(readFileSync(new URL('../testdata/http/raw-histogram447.json',import.meta.url))).cases.find(value=>value.name==='observed').response);
import {dataset} from '../ui/raw-histogram-dataset.mjs';
const registry=new Map();
class NativeElement extends Element{
 constructor(tag='custom-element'){super(tag);this.events=[];this.handlers=new Map()}
 matches(selector){if(selector==='[data-occurrence-key]')return this.dataset.occurrenceKey!==undefined;if(selector==='[data-point-key]')return this.dataset.pointKey!==undefined;return super.matches(selector)}
 addEventListener(name,handler){this.handlers.set(name,handler)}
 dispatchEvent(event){this.events.push(event);this.handlers.get(event.type)?.(event);return true}
}
globalThis.HTMLElement=NativeElement;globalThis.customElements={get:key=>registry.get(key),define:(key,value)=>registry.set(key,value)};
globalThis.document={activeElement:null,createElement:tag=>{const Type=registry.get(tag);const element=Type?new Type():new NativeElement(tag);element.localName=tag;return element}};
globalThis.CustomEvent=class{constructor(type,init){this.type=type;Object.assign(this,init)}};
await import('../ui/scope-view.mjs');
const Chart=registry.get('goo-chart');
test('common chart renders one raw histogram point with the same exact Table/bars component',()=>{
 const chart=new Chart();chart.root=true;chart.setAttribute('heading-level','5');chart.snapshot=dataset(bank,'1');chart.display={kind:'histogram',title:'One source occurrence'};
 let buckets=chart.querySelector('goo-raw-histogram-buckets');assert.ok(buckets,chart.textContent);assert.equal(buckets.getAttribute('mode'),'bars');assert.equal(buckets.getAttribute('heading-level'),'5');assert.equal(buckets.getAttribute('heading'),'One source occurrence');assert.equal(buckets.data.occurrenceKey,'1');assert.equal(chart.querySelector('.scope-view-unavailable'),null);
 buckets.connectedCallback();assert.match(buckets.textContent,/3/);assert.match(buckets.textContent,/2/);assert.equal(chart.events.length,0,'bucket count geometry does not select a record or invent membership');
 chart.display={kind:'table'};buckets=chart.querySelector('goo-raw-histogram-buckets');assert.equal(buckets.getAttribute('mode'),'table');assert.equal(chart.querySelector('.scope-view-unavailable'),null);buckets.connectedCallback();assert.equal(buckets.querySelectorAll('tbody').length,1);
 chart.display={kind:'histogram',requiresComplete:true};assert.equal(chart.querySelector('goo-raw-histogram-buckets'),null);assert.match(chart.textContent,/Complete-only display/);
});

test('457 common chart retains bucket page only for the same admitted snapshot',()=>{
 const source=structuredClone(bank),point=source.rows[0];point.rawRow.bounds=Array.from({length:204},(_,i)=>String(i));point.bounds=Array.from({length:204},(_,i)=>i);point.rawRow.bucket_counts=point.bucketCounts=Array(205).fill('1');point.rawRow.count=point.count='205';
 const chart=new Chart();chart.root=true;chart.snapshot=dataset(source,'1');chart.display={kind:'histogram'};
 const buckets=chart.querySelector('goo-raw-histogram-buckets');buckets.connectedCallback();buckets.querySelectorAll('button')[1].onclick();assert.equal(buckets.querySelectorAll('tbody')[0].children[0].dataset.bucketIndex,'100');
 chart.display={kind:'table',title:'Same point'};assert.equal(chart.querySelector('goo-raw-histogram-buckets'),buckets);assert.equal(buckets.getAttribute('mode'),'table');assert.equal(buckets.querySelectorAll('tbody')[0].children[0].dataset.bucketIndex,'100');
 chart.display={kind:'histogram'};assert.equal(chart.querySelector('goo-raw-histogram-buckets'),buckets);assert.equal(buckets.querySelectorAll('tbody')[0].children[0].dataset.bucketIndex,'100');
 chart.snapshot=structuredClone(chart.snapshot);const replacement=chart.querySelector('goo-raw-histogram-buckets');assert.notEqual(replacement,buckets);replacement.connectedCallback();assert.equal(replacement.querySelectorAll('tbody')[0].children[0].dataset.bucketIndex,'0');
 chart.display={kind:'histogram',requiresComplete:true};assert.equal(chart.querySelector('goo-raw-histogram-buckets'),null);chart.display={kind:'table'};assert.notEqual(chart.querySelector('goo-raw-histogram-buckets'),replacement,'removed/refused presentations do not retain detached page authority');
});
