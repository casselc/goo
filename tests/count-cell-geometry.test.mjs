import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class Element {
 constructor(tag,text,className=''){this.tagName=tag;this.textContent=text;this.className=className;this.children=[];this.dataset={};this.style={};this.attributes={};this.classList={add:name=>{this.className+=' '+name}};}
 append(...children){this.children.push(...children)}
 setAttribute(key,value){this.attributes[key]=value}
}
const node=(...args)=>new Element(...args);
mock.module('../ui/components.mjs',{exports:{node}});
const {cell}=await import('../ui/comparison-visual.mjs');
const track=element=>element.children[1],label=element=>element.children[0];
const unavailable=element=>element.className.split(' ').includes('unavailable');

test('ordinary count cells keep their default geometry and exact common scale',()=>{
 for(const value of ['0','27','9007199254740993','18446744073709551615']){
  const first=node('main'),second=node('main'),options={max:'18446744073709551615',unit:'records'};
  const standard=cell(first,value,'selection',options),explicit=cell(second,value,'selection',{...options,geometry:true});
  assert.deepEqual(JSON.parse(JSON.stringify(standard)),JSON.parse(JSON.stringify(explicit)));assert.equal(standard.dataset.value,value);assert.equal(standard.dataset.geometry,'shown');assert.equal(track(standard).children.length,1);assert.equal(track(standard).hidden,false);assert.equal(unavailable(standard),false);
 }
 const exact=cell(node('main'),'1','baseline',{max:'16'});assert.equal(track(exact).children[0].style.width,'6.25%');
});
test('withheld geometry keeps known zero and large values known rather than unavailable',()=>{
 for(const value of ['0','27','9007199254740993','18446744073709551615']){
  const shown=cell(node('main'),value,'selection',{max:'18446744073709551615',unit:'records',geometry:false});
  assert.equal(shown.dataset.value,value);assert.equal(shown.dataset.scaleMax,'18446744073709551615');assert.equal(shown.dataset.geometry,'withheld');assert.equal(label(shown).textContent,value+' records');assert.equal(unavailable(shown),false);assert.equal(track(shown).hidden,true);assert.equal(track(shown).children.length,0);assert.equal(track(shown).attributes['aria-hidden'],'true');
 }
 const formatted=cell(node('main'),'27','selection',{max:'75',geometry:false,text:'27 rows'});assert.equal(formatted.dataset.value,'27');assert.equal(label(formatted).textContent,'27 rows');
});
test('withholding bars cannot manufacture a known value or reinterpret an invalid policy',()=>{
 for(const geometry of [true,false]){
  const unknown=cell(node('main'),null,'baseline',{max:'75',geometry});assert.equal(unknown.dataset.value,'unavailable');assert.equal(unknown.dataset.geometry,'unavailable');assert.equal(unavailable(unknown),true);assert.equal(label(unknown).textContent,'Unavailable');assert.equal(track(unknown).children.length,0);
 }
 const dash=cell(node('main'),null,'baseline',{max:'75',geometry:false,text:'—'});assert.equal(dash.dataset.value,'unavailable');assert.equal(label(dash).textContent,'—');
 for(const geometry of [null,0,1,'false',{},[]]){const host=node('main');assert.throws(()=>cell(host,'27','selection',{geometry}),/geometry must be a boolean/);assert.equal(host.children.length,0);}
});
