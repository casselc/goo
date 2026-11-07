import * as Measurements from './measurements.mjs';
import {admitHeld} from './native-held-record-context.mjs';
import {actionVisible} from './visualizations.mjs';

const clone=value=>structuredClone(value);
function freeze(value){if(value&&typeof value==='object'){for(const item of Object.values(value))freeze(item);Object.freeze(value)}return value}
const canonical=value=>JSON.stringify(value,(_key,item)=>item&&typeof item==='object'&&!Array.isArray(item)?Object.fromEntries(Object.keys(item).sort().map(key=>[key,item[key]])):item);
const equal=(a,b)=>canonical(a)===canonical(b);
function checked(page){
 const receipt=page?.receipt,measurement=page?.measurement;
 if(!/^[a-f0-9]{64}$/.test(receipt?.id)||receipt.question?.kind!=='measurement'||!measurement?.question||!equal(receipt.question.measurement,measurement.question))throw Error('Stored measurement differs from its immutable receipt.');
 return freeze(clone({receipt,measurement}));
}
function admit(page,detail){return admitHeld(checked(page),detail)}
function mount(host,{active=()=>true,onOpenContext=()=>false}={}){
 let held=null,revision=0,actionLease=0,suspended=false,disposed=false;const anchors=new Map();
 const connected=()=>host.isConnected&&!disposed;
 const sourceCurrent=(value,version)=>connected()&&held===value&&revision===version;
 function bind(handler,anchor){const lease=actionLease,version=revision,value=held;anchors.set(anchor,handler);return function(event){if(!sourceCurrent(value,version)||suspended||lease!==actionLease||!active()||!anchor.isConnected||!actionVisible(anchor))return false;return handler.call(this,event)}}
 function show(page){
  if(disposed)return false;const next=checked(page);revision++;actionLease++;held=next;suspended=false;anchors.clear();const version=revision;
  Measurements.render(host,next.measurement,{saved:true,summaryProvided:true,canAct:()=>sourceCurrent(next,version)&&!suspended&&active(),bindRecordContextAction:bind,onRecordContext:(detail,opener)=>{
   if(!sourceCurrent(next,version)||suspended||!active()||!opener?.isConnected||!actionVisible(opener))return false;
   let admitted;try{admitted=admitHeld(next,detail)}catch{return false}
   const source=Object.freeze({current:()=>sourceCurrent(next,version),receipt:next.receipt,measurement:next.measurement,opener});
   return onOpenContext(admitted,source)===true;
  }});return true;
 }
 function suspend(){if(disposed)return;suspended=true;actionLease++}
 function resume(){if(disposed||!suspended||!connected()||!active())return false;suspended=false;actionLease++;for(const [anchor,handler]of anchors){if(anchor.isConnected)anchor.onclick=bind(handler,anchor);else anchors.delete(anchor)}return true}
 function dispose(){if(disposed)return;disposed=true;revision++;actionLease++;held=null;anchors.clear();host.replaceChildren()}
 return{show,suspend,resume,dispose};
}
export {admit,mount};
