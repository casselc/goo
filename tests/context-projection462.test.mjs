import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Element} from './support/raw-observation-dom.mjs';
import {lanes,item} from './support/held-chronology458-fixture.mjs';
const registry=new Map();
globalThis.HTMLElement=class extends Element{constructor(tag='element'){super(tag)}};
globalThis.customElements={get:key=>registry.get(key),define:(key,value)=>registry.set(key,value)};
globalThis.document={hidden:false,activeElement:null,createElement:tag=>{const Type=registry.get(tag),node=Type?new Type():new HTMLElement(tag);node.localName=tag;return node}};
const {prepare}=await import('../ui/held-chronology.mjs');
await import('../ui/held-chronology-view.mjs');
function data(){const value=lanes();value.spans.rows.push(item('spans',1));return prepare(value)}
function mount(data){const View=registry.get('goo-held-chronology'),view=new View();view.data=data;view.root=true;view.connectedCallback();return view}

test('internal parent projections are shared safely while public context inputs and readback stay detached',()=>{
 const input=structuredClone(data());for(const item of input.lanes.spans.rows){item.record.TraceId='a'.repeat(32);item.record.SpanId='b'.repeat(16)}const view=mount(input);view.selectedKey='spans:'+input.lanes.spans.rows[0].key;view.showRelated('spans',input.lanes.spans.rows[0].key);const contexts=view.querySelectorAll('goo-record-context');assert.equal(contexts[0]._records,contexts[1]._records);assert.equal(Object.isFrozen(contexts[0]._records),true);const read=contexts[0].records;read[0].SpanId='forged';assert.equal(contexts[1].records[0].SpanId,'b'.repeat(16));const raw=[{TraceId:'a'.repeat(32),SpanId:'b'.repeat(16)}];contexts[0].records=raw;raw[0].SpanId='changed';assert.equal(contexts[0].records[0].SpanId,'b'.repeat(16));
});
