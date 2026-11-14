import {test} from 'node:test';
import assert from 'node:assert/strict';
import {recordLabel,matchesLoaded} from '../ui/record-presentation.mjs';

test('record severity follows every OTel UInt8 band independently of producer text',()=>{
 const labels=['Trace','Debug','Info','Warn','Error','Fatal'];
 for(let value=0;value<256;value++)for(const numeric of [value,String(value)]){
  const record={SeverityNumber:numeric,SeverityText:'FATAL'},before=structuredClone(record),label=recordLabel(record,'logs');
  const expected=value===0?'Unspecified':value>24?'Invalid':labels[Math.floor((value-1)/4)];
  assert.equal(label.band,expected.toLowerCase());assert.equal(label.numeric,true);
  assert.equal(label.tone,value>=17&&value<=24?'error':value>=13&&value<=16?'warn':'neutral');
  assert.equal(label.text,expected==='Fatal'?'Fatal':expected+' · producer: FATAL');
  assert(label.title.includes('SeverityNumber '+value));assert(label.title.includes('producer severity text: FATAL'));
  assert.deepEqual(record,before);
 }
});

test('missing and malformed numeric severity never inherit producer alarm colors',()=>{
 for(const value of [undefined,null,-1,256,1.5,'09','+9',' 9',NaN,{},false,9n]){
  const label=recordLabel({SeverityNumber:value,SeverityText:'ERROR'},'logs');assert.equal(label.tone,'neutral');assert.equal(label.numeric,false);assert(label.text.endsWith(' · producer: ERROR'));
 }
 assert.equal(recordLabel({SeverityNumber:9,SeverityText:'FATAL'},'logs').text,'Info · producer: FATAL');
 assert.equal(recordLabel({SeverityNumber:17,SeverityText:'INFO'},'logs').text,'Error · producer: INFO');
 assert.equal(recordLabel({SeverityNumber:9,SeverityText:'INFO'},'logs').text,'Info');
 assert.equal(recordLabel({SeverityNumber:0},'logs').text,'Unspecified');
 assert.equal(recordLabel({},'logs').text,'Unreported');
 assert.equal(recordLabel({SeverityNumber:9,StatusCode:'Error'},'spans').tone,'error');
 assert.equal(recordLabel({StatusCode:'not-an-error'},'spans').tone,'neutral');
});

test('Find matches visible canonical severity and literal producer text without inventing missing data',()=>{
 const info={SeverityNumber:9,SeverityText:'FATAL',Body:'ordinary'};
 assert.equal(matchesLoaded(info,'logs',[],'info'),true);assert.equal(matchesLoaded(info,'logs',[],'fatal'),true);
 assert.equal(matchesLoaded(info,'logs',[],'error'),false);
 assert.equal(matchesLoaded({SeverityText:'ERROR'},'logs',[],'unreported'),false);
 assert.equal(matchesLoaded({SeverityNumber:0},'logs',[],'unspecified'),true);
});

test('shared custom-element label renders canonical tone with attributed producer text',async t=>{
 const previous={HTMLElement:globalThis.HTMLElement,customElements:globalThis.customElements,document:globalThis.document};
 class Element{
  constructor(tag='element'){this.localName=tag;this.attributes=new Map();this.dataset={};this.children=[];this.isConnected=false;this.textContent=''}
  setAttribute(key,value){this.attributes.set(key,String(value))}getAttribute(key){return this.attributes.get(key)??null}hasAttribute(key){return this.attributes.has(key)}
  append(...children){this.children.push(...children)}replaceChildren(...children){this.children=[...children]}querySelector(){return null}
 }
 const classes=new Map();globalThis.HTMLElement=Element;globalThis.customElements={get:name=>classes.get(name),define:(name,element)=>classes.set(name,element)};globalThis.document={createElement:tag=>new Element(tag)};
 t.after(()=>{for(const [key,value]of Object.entries(previous))if(value===undefined)delete globalThis[key];else globalThis[key]=value});
 await import('../ui/record-presentation.mjs?severity-tone-test');
 const Summary=classes.get('goo-record-summary'),element=new Summary();element.setAttribute('signal','logs');element.record={SeverityNumber:9,SeverityText:'FATAL',Body:'ordinary'};element.isConnected=true;element.connectedCallback();
 const label=element.children[0];assert.equal(label.textContent,'Info · producer: FATAL');assert.equal(label.dataset.tone,'neutral');assert.equal(label.dataset.severityBand,'info');assert.equal(label.getAttribute('aria-label'),null);assert.match(label.title,/OTel severity Info .*producer severity text: FATAL/);
});
