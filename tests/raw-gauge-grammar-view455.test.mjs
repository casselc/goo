import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Element} from './support/raw-observation-dom.mjs';
import {bank} from './support/raw-gauge-points455-fixture.mjs';
import {dataset} from '../ui/raw-gauge-points.mjs';
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
test('generic chart grammar delegates points to the passive observation renderer rather than an unknown-value fallback',()=>{
 const chart=new Chart();chart.root=true;chart.snapshot=dataset(bank([{rawValue:'7'},{flags:2}]));chart.display={kind:'points',title:'Worker gauge'};
 const points=chart.querySelector('goo-raw-gauge-points');assert.ok(points,chart.textContent);assert.equal(points.getAttribute('mode'),'points');assert.equal(points.getAttribute('heading'),'Worker gauge');assert.equal(points.data.shape,'observations');assert.equal(chart.querySelector('.scope-view-unavailable'),null);
 points.dispatchEvent(new CustomEvent('goo-raw-gauge-point',{detail:{occurrenceKeys:['1']},bubbles:true}));assert.equal(chart.events.at(-1).type,'goo-select');assert.equal(chart.events.at(-1).detail.key,'1');assert.equal(chart.events.at(-1).detail.datasetId,chart.snapshot.id);
 const before=chart.events.length;points.dispatchEvent(new CustomEvent('goo-raw-gauge-point',{detail:{occurrenceKeys:['1','2']},bubbles:true}));assert.equal(chart.events.length,before,'a coordinate group must not select its first member');
 chart.selectedKey='1';assert.equal(points.selectedKey,'1');chart.display={kind:'table'};assert.equal(chart.querySelector('goo-raw-gauge-points').getAttribute('mode'),'table');assert.equal(chart.querySelector('.scope-view-unavailable'),null);points.dispatchEvent(new CustomEvent('goo-raw-gauge-point',{detail:{occurrenceKeys:['1']},bubbles:true}));assert.equal(chart.events.length,before,'retired renderer cannot publish selection');
});
test('table renderer shows UTC/raw value/recording state and remains bounded without a point plot',()=>{
 const chart=new Chart();chart.root=true;chart.snapshot=dataset(bank([{rawValue:'7'},{flags:2}]));chart.display={kind:'table'};const table=chart.querySelector('goo-raw-gauge-points');table.connectedCallback();assert.equal(table.querySelector('.raw-gauge-points-plot'),null);assert.equal(table.querySelectorAll('tbody').length,1);assert.match(table.textContent,/Observed UTC/);assert.match(table.textContent,/Recording unknown/);assert.doesNotMatch(table.textContent,/1700000000000000001/);assert.equal(table.querySelectorAll('[data-occurrence-key]').length,2);
});
