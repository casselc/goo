import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Element} from './support/raw-observation-dom.mjs';
import {displayChoice} from '../ui/display-choice.mjs';
const options=[{value:'histogram',label:'Bars'},{value:'table',label:'Table'}];
function environment(){const body=new Element('body');body.root=true;globalThis.document={body,hidden:false,activeElement:null,createElement:tag=>new Element(tag)};return body}
test('native display choices retain explicit labels, checked state, grouping and focus',()=>{
 const body=environment(),calls=[],picker=displayChoice({legend:'Bucket display',options,value:'histogram',onChange:(...args)=>calls.push(args)});body.append(picker);assert.equal(picker.localName,'fieldset');assert.equal(picker.querySelector('legend').textContent,'Bucket display');const [bars,table]=picker.querySelectorAll('input');assert.equal(bars.type,'radio');assert.equal(bars.checked,true);assert.equal(table.checked,false);assert.equal(bars.name,table.name);assert.equal(table.getAttribute('aria-label'),'Table');
 table.checked=true;table.focus();table.onchange();assert.deepEqual(calls, [['table',table]]);assert.equal(document.activeElement,table);
 const another=displayChoice({legend:'Observation display',options,value:'table'});assert.notEqual(another.querySelector('input').name,bars.name);
});
test('withdrawn native choices do not publish changes or perform source IO',()=>{
 const body=environment();let active=true,calls=0,reads=0;globalThis.fetch=()=>{reads++;throw Error('No source access')};const picker=displayChoice({legend:'Display',options,value:'histogram',active:()=>active,onChange:()=>calls++}),input=picker.querySelectorAll('input')[1];body.append(picker);
 input.onchange();assert.equal(calls,0);input.checked=true;active=false;input.onchange();active=true;picker.hidden=true;input.onchange();picker.hidden=false;picker.setAttribute('inert','');input.onchange();picker.removeAttribute('inert');input.disabled=true;input.onchange();input.disabled=false;picker.remove();input.onchange();assert.equal(calls,0);body.append(picker);input.onchange();assert.equal(calls,1);assert.equal(reads,0);delete globalThis.fetch;
});
test('ambiguous display configuration refuses rather than inventing a default',()=>{
 environment();for(const override of [{legend:''},{options:[options[0]]},{options:[options[0],options[0]]},{value:'missing'},{options:[...options,{value:'bad',label:undefined}]}])assert.throws(()=>displayChoice({legend:'Display',options,value:'histogram',...override}),TypeError);
});

test('display choices detach option values and labels from later caller mutation',()=>{
 const body=environment(),choices=structuredClone(options),calls=[],picker=displayChoice({legend:'Display',options:choices,value:'histogram',onChange:value=>calls.push(value)});body.append(picker);const table=picker.querySelectorAll('input')[1];choices[1].value='other';choices[1].label='Other';table.checked=true;table.onchange();assert.equal(table.value,'table');assert.equal(table.getAttribute('aria-label'),'Table');assert.deepEqual(calls,['table']);
});
