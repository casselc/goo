/* A native local display choice. The caller owns rendering and source authority. */
import {nativeControlReady} from './dom-visibility.mjs';
let instance=0;
export function displayChoice({legend,options,value,active=()=>true,onChange=()=>{}}){
 if(typeof legend!=='string'||!legend||!Array.isArray(options)||options.length<2||options.some(option=>typeof option?.value!=='string'||!option.value||typeof option.label!=='string'||!option.label)||new Set(options.map(option=>option.value)).size!==options.length||!options.some(option=>option.value===value))throw TypeError('Supply named display choices and an initial value.');
 const fieldset=document.createElement('fieldset'),heading=document.createElement('legend'),name='goo-display-choice-'+(++instance);fieldset.className='ui-display-choice';heading.textContent=legend;fieldset.append(heading);
 for(const {value:choice,label:text} of options){
  const label=document.createElement('label'),input=document.createElement('input');input.type='radio';input.name=name;input.value=choice;input.checked=choice===value;input.setAttribute('aria-label',text);label.append(input,text);fieldset.append(label);
  input.onchange=()=>{if(!active()||!nativeControlReady(input)||!input.checked)return;onChange(choice,input)};
 }
 return fieldset;
}
