import {validateRecordQuestion} from './record-intent.mjs';
import {identity} from './question-session.mjs';

/* One additional-condition owner. No DOM, transport, execution or evidence state. */
import {recordCELContract} from './cel-provenance.mjs';
export {recordCELContract};
const clone=value=>structuredClone(value);
const bindingKeys=['operation','signal','baseIdentity','slot','epoch','revision','library','compiler','target','profile'];
const sameBinding=(a,b)=>a&&b&&bindingKeys.every(key=>a[key]===b[key])&&Object.keys(a).length===bindingKeys.length;
function token(value,name){
 if(typeof value!=='string'||!value.isWellFormed()||!value||new TextEncoder().encode(value).length>128||/[\p{White_Space}\p{Cc}]/u.test(value))throw Error(`Invalid ${name} token.`);
 return value;
}
function parent(options){
 validateRecordQuestion(options.base);
 if(typeof options.baseIdentity!=='string'||!/^[a-f0-9]{64}$/.test(options.baseIdentity))throw Error('Require the supplied base content identity.');
 return {base:clone(options.base),baseIdentity:options.baseIdentity,parentRevision:token(options.parentRevision,'parent revision'),slot:token(options.slot,'slot'),epoch:token(options.epoch,'epoch')};
}

export class AuthoringSlot {
 #parent; #revision=0n; #pending=null; #stageLease=null; #active=true; #disposed=false;
 #text=''; #mode='conditions'; #phase='clear'; #diagnostic=null; #lastValid=null; #current=null;
 constructor(options){this.#parent=parent(options)}
 #retire(phase='draft'){
  this.#revision++;this.#pending=null;this.#stageLease=null;this.#current=null;this.#diagnostic=null;this.#phase=phase;
 }
 #binding(){return {...recordCELContract,signal:this.#parent.base.signal,baseIdentity:this.#parent.baseIdentity,slot:this.#parent.slot,epoch:this.#parent.epoch,revision:String(this.#revision)}}
 #owns(lease){return !this.#disposed&&this.#active&&lease===this.#pending&&lease!==null}
 snapshot(){return clone({base:this.#parent.base,baseIdentity:this.#parent.baseIdentity,parentRevision:this.#parent.parentRevision,slot:this.#parent.slot,epoch:this.#parent.epoch,revision:String(this.#revision),text:this.#text,mode:this.#mode,phase:this.#phase,diagnostic:this.#diagnostic,lastValid:this.#lastValid,active:this.#active&&!this.#disposed})}
 edit(text){
  if(this.#disposed)return false;
  if(typeof text!=='string')throw Error('Expression text must be a string.');
  this.#text=text;this.#retire();return true;
 }
 setMode(mode){
  if(this.#disposed)return false;
  if(!['conditions','expression'].includes(mode))throw Error('Unknown authoring mode.');
  this.#mode=mode;return true;
 }
 // Visual controls must deliberately discard an existing authored draft. The
 // supplied expression still requires an independent checked planning result.
 adoptVisual(text,{discardDraft=false}={}){
  if(this.#disposed||this.#text!==''&&!discardDraft)return false;
  this.edit(text);this.#mode='conditions';return true;
 }
 beginCheck(){
  if(this.#disposed||!this.#active)return null;
  this.#retire('checking');
  const request={base:clone(this.#parent.base),expression:this.#text,binding:this.#binding()};
  const lease=Object.freeze({});
  this.#pending=lease;
  // The private originating request cannot be changed by a caller retaining its
  // public copy. Repeated checks get distinct leases even for unchanged text.
  this.#current={request:clone(request),parentRevision:this.#parent.parentRevision};
  return {request,token:lease};
 }
 accept(lease,receipt){
  if(!this.#owns(lease))return false;
  const request=this.#current.request;
  if(!receipt||receipt.version!=='goo.cel-plan/1'||receipt.celLibrary!==request.binding.library||receipt.expression!==request.expression||typeof receipt.loweringDigest!=='string'||!/^[a-f0-9]{64}$/.test(receipt.loweringDigest)||!sameBinding(receipt.binding,request.binding)||!receipt.predicate)return false;
  try{
   validateRecordQuestion(receipt.question);
   const expected=clone(request.base);
   expected.where=expected.where?{all:[expected.where,clone(receipt.predicate)]}:clone(receipt.predicate);
   validateRecordQuestion(expected);
   if(identity(receipt.question)!==identity(expected))return false;
  }catch{return false}
  this.#lastValid=clone(receipt);this.#current=clone(receipt);this.#stageLease=lease;this.#pending=null;this.#diagnostic=null;this.#phase='valid';return true;
 }
 fail(lease,diagnostic){
  if(!this.#owns(lease))return false;
  if(!diagnostic||typeof diagnostic.code!=='string'||typeof diagnostic.phase!=='string'||typeof diagnostic.message!=='string')throw Error('Require a structured planning diagnostic.');
  this.#pending=null;this.#current=null;this.#diagnostic=clone(diagnostic);this.#phase='invalid';return true;
 }
 stage(lease){
  if(this.#disposed||!this.#active||!lease||lease!==this.#stageLease)return null;
  if(this.#phase==='clear')return clone(this.#parent.base);
  return this.#phase==='valid'?clone(this.#current.question):null;
 }
 clear(){
  if(this.#disposed)return false;
  this.#text='';this.#retire('clear');this.#stageLease=Object.freeze({});return this.#stageLease;
 }
 replaceParent(options){
  if(this.#disposed)return false;
  const next=parent(options);this.#parent=next;this.#retire();return true;
 }
 setVisible(active){
  if(this.#disposed)return false;
  active=!!active;
  if(active!==this.#active){this.#active=active;this.#retire()}
  return true;
 }
 dispose(){if(!this.#disposed){this.#disposed=true;this.#active=false;this.#retire('disposed')}}
}
