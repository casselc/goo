/* Pure categorical intent. Owners retain draft, acquisition and revision authority. */
import {text} from './intent-shape.mjs';
import {validate,validateField} from './predicate-validation.mjs';

function checked(value,signal,fieldOnly=false){
 const check=fieldOnly?validateField:validate;
 if(signal!==undefined)return check(value,signal);
 let error;for(const candidate of ['logs','spans'])try{return check(value,candidate)}catch(e){error=e}
 throw error;
}
export function actions(field){
 try{checked(field,undefined,true)}catch{return []}
 const items=[{mode:'include',label:'Include'},{mode:'exclude',label:'Exclude'}];
 if(['resource','attribute'].includes(field.kind))items.push({mode:'reported',label:'Is reported'},{mode:'missing',label:'Is missing'});
 return items;
}
export function intent(field,value,mode){
 checked(field,undefined,true);
 if(['reported','missing'].includes(mode)){
  if(!['resource','attribute'].includes(field.kind)||value!==null)throw Error('Presence actions require an attribute field and no value.');
  return {field:structuredClone(field),value:null,mode:mode==='reported'?'exclude':'include'};
 }
 if(!['include','exclude'].includes(mode)||!(value===null||text(value,1024)))throw Error('Choose a reported string value or missing key.');
 return {field:structuredClone(field),value,mode};
}
export function predicate(field,value,mode,signal){
 const selected=intent(field,value,mode),leaf=(op,v)=>({field:structuredClone(selected.field),op,...(v===undefined||v===''?{}:{value:v})});
 const p=selected.mode==='include'?(selected.value===null?leaf('missing'):leaf('eq',selected.value)):(selected.value===null?leaf('exists'):{any:[leaf('missing'),leaf('ne',selected.value)]});
 return checked(p,signal);
}
export function reason(field,value,mode,signal){try{predicate(field,value,mode,signal);return ''}catch(error){return error.message}}
