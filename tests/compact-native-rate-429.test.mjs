import {test} from 'node:test';
import assert from 'node:assert/strict';
import {compactScalar} from '../ui/numeric-values.mjs';
import * as Native from '../ui/native-presentation.mjs';
import {model,timeBounds} from '../ui/value-timeline.mjs';

test('429 scalar presentation distinguishes exact zero, tiny nonzero and approximate large quantities',()=>{
 for(const [raw,text,rounded]of [['0','0',false],['-0','0',false],['6','6',false],['123/100','1.23',false],['1/1000000','1e-6',false],['-1/1000000','-1e-6',false],['2/3','≈0.667',true],['123456789','≈123M',true],['999500','≈1M',true],['9007199254740993','≈9.01e+15',true]])assert.deepEqual(compactScalar(raw),{text,exact:raw,rounded});
 for(const raw of ['1/0','NaN','Infinity','1e9999',' 1','1 ','1/2/3','',null,Infinity])assert.equal(compactScalar(raw),null);
 // Exercise orders where floating point or fixed decimal formatting would erase a rate.
 for(let exponent=-250;exponent<=250;exponent++)for(const sign of ['','-']){
  const raw=sign+'12345e'+exponent,value=compactScalar(raw);
  assert.ok(value&&value.text!=='0'&&value.text!=='≈0',raw);assert.equal(value.exact,raw);assert.equal(value.text.replace(/^≈/,'').startsWith('-'),sign==='-');
 }
});

test('429 counter projections keep original units, exact rational disclosure and latest unknown evidence',()=>{
 const stream={name:'network.io',unit:'By',resource:{'service.instance.id':'one'},attributes:{}},series={stream,state:'unknown',intervals:[{fromNS:'1',toNS:'2',state:'ok',delta:'1',perSecond:0.000001,exactPerSecond:'1/1000000'}]},before=structuredClone(series);
 const summary=Native.latest(series,'counter');assert.equal(summary.value,'1e-6 By/s');assert.equal(summary.valueTitle,'Exact rate 1/1000000 By/s');assert.deepEqual(series,before);const recorded=structuredClone(series);delete recorded.intervals[0].exactPerSecond;assert.equal(Native.latest(recorded,'counter').valueTitle,'Recorded rate 0.000001 By/s');recorded.intervals[0].exactPerSecond=null;assert.equal(Native.latest(recorded,'counter').valueTitle,'Recorded rate 0.000001 By/s');
 assert.equal(Native.windowValue(stream,{state:'observed',reasons:[],delta:'1',rate:'1/1000000'}),'1 increments · 1e-6 By/s');
 series.intervals.push({fromNS:'2',toNS:'3',state:'unknown',reason:'counter-decrease'});assert.equal(Native.latest(series,'counter').value,'Unavailable');assert.equal(Native.latest(series,'counter').valueTitle,undefined);
 assert.equal(Native.rateText('0'),'0');assert.equal(Native.rateText('1/0'),'Unavailable');
});

test('429 compact timeline labels do not round supplied geometry, domains or exact UTC bounds',()=>{
 const data={fromNS:'1790000000000000000',toNS:'1790000000000000100',points:[{key:'tiny',fromNS:'1790000000000000001',toNS:'1790000000000000040',value:'0.0000001',label:'tiny'},{key:'large',fromNS:'1790000000000000040',toNS:'1790000000000000099',value:'123456789.123',label:'large'}]},before=structuredClone(data),scaled=model(data);
 assert.equal(scaled.min,'0');assert.equal(scaled.max,'123456789.123');assert.equal(scaled.points[0].value,'0.0000001');assert.equal(scaled.points[1].value,'123456789.123');assert.equal(compactScalar(scaled.max).text,'≈123M');assert.deepEqual(data,before);
 const close=model({fromNS:'0',toNS:'100',points:[{key:'a',fromNS:'10',toNS:'20',value:'1001',label:'a'},{key:'b',fromNS:'30',toNS:'40',value:'1002',label:'b'}]});assert.equal(close.min,'0');assert.equal(close.max,'1002');assert.notEqual(compactScalar(close.min).text,compactScalar(close.max).text);
 const bounds=timeBounds(data);assert.match(bounds[0].text,/UTC$/);assert.equal(bounds[1].exact,'2026-09-21T14:13:20.000000100Z');assert.match(bounds[1].title,/1790000000000000100 ns/);
});


test('429 rendered shared timeline exposes compact scale with exact focusable detail and unchanged marks',async t=>{
 const saved={HTMLElement:globalThis.HTMLElement,customElements:globalThis.customElements,document:globalThis.document};
 class Element{constructor(tag='goo-value-timeline'){this.localName=tag;this.children=[];this.attributes={};this.dataset={};this.style={};this.isConnected=false}append(...nodes){this.children.push(...nodes)}replaceChildren(...nodes){this.children=[...nodes]}setAttribute(key,value){this.attributes[key]=String(value)}getAttribute(key){return this.attributes[key]??null}hasAttribute(key){return Object.hasOwn(this.attributes,key)}get childNodes(){return this.children}querySelectorAll(){return []}}
 const registry=new Map();globalThis.HTMLElement=Element;globalThis.customElements={get:key=>registry.get(key),define:(key,value)=>registry.set(key,value)};globalThis.document={createElement:tag=>new Element(tag)};
 t.after(()=>{for(const [key,value]of Object.entries(saved)){if(value===undefined)delete globalThis[key];else globalThis[key]=value}});
 await import('../ui/value-timeline.mjs?compact429');const Timeline=registry.get('goo-value-timeline'),plot=new Timeline();
 const supplied={fromNS:'0',toNS:'100',points:[{key:'tiny',fromNS:'10',toNS:'20',value:'1e-8',label:'Exact tiny rate'},{key:'large',fromNS:'30',toNS:'40',value:'123456789.123',label:'Exact large rate'}]},before=structuredClone(supplied);plot.data=supplied;plot.render();
 const [scale,marks]=plot.children;assert.equal(scale.textContent,'0 to ≈123M · dashed baseline is zero');assert.equal(scale.title,'0 to 123456789.123 · exact displayed domain');assert.equal(scale.tabIndex,0);assert.deepEqual(marks.children.slice(1).map(mark=>mark.dataset.value),['1e-8','123456789.123']);assert.deepEqual(plot.data,before);
});
