/* Local reading state only. DOM, source authority and inspection belong to callers. */
import {findQuery} from './record-presentation.mjs';
const states=new WeakSet();
function state(value){if(value.related)Object.freeze(value.related);if(value.returnPoint)Object.freeze(value.returnPoint);const next=Object.freeze(value);states.add(next);return next}
function checked(value){if(!states.has(value))throw Error('Use an initialized chronology reading state.');return value}
export function initial(){return state({query:'',offset:0,related:null,returnPoint:null})}
export function find(value,query){checked(value);if(value.related)throw Error('Return to chronology before changing Find.');findQuery(query);return state({...value,query,offset:0})}
export function fit(value,length){checked(value);if(!Number.isInteger(length)||length<0||length>200)throw Error('Use the bounded held row count.');const offset=Math.min(value.offset,Math.max(0,Math.floor((length-1)/100)*100));return offset===value.offset?value:state({...value,offset})}
export function page(value,delta,length){checked(value);if(![-1,1].includes(delta))throw Error('Choose the previous or next page.');return fit(state({...value,offset:Math.max(0,value.offset+delta*100)}),length)}
export function enter(value,{signal,key},{selectedKey=null,scroll=0}={}){checked(value);if(!['logs','spans'].includes(signal)||typeof key!=='string'||!/^[a-f0-9]{64}:\d+$/.test(key)||selectedKey!==null&&(typeof selectedKey!=='string'||!/^(logs|spans):[a-f0-9]{64}:\d+$/.test(selectedKey))||!Number.isFinite(scroll)||scroll<0)throw Error('Use a held occurrence and local return point.');return state({...value,offset:0,related:{signal,key,spanOnly:false},returnPoint:value.returnPoint||{query:value.query,offset:value.offset,selectedKey,scroll}})}
export function span(value,spanOnly){checked(value);if(!value.related||typeof spanOnly!=='boolean')throw Error('Choose a related span scope.');return state({...value,offset:0,related:{...value.related,spanOnly}})}
export function leave(value){checked(value);if(!value.related||!value.returnPoint)throw Error('No related reading scope to return from.');const snapshot=value.returnPoint;return {state:state({query:snapshot.query,offset:snapshot.offset,related:null,returnPoint:null}),snapshot}}
