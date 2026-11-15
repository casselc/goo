import {identity} from './question-session.mjs';

/* Private application state for an applied predicate and an explicit draft.
 * Validation belongs to the owner; this class neither reads nor starts anything. */
export class PredicateDraft {
 #applied; #draft; #error=null; #pending=false; #revision=0; #validate;
 constructor({where,validate=()=>{}}={}){
  if(typeof validate!=='function')throw TypeError('Predicate validation must be a function.');
  this.#validate=validate;
  const checked=this.#checked(where);this.#applied=structuredClone(checked);this.#draft=checked;
 }
 #checked(where){const candidate=structuredClone(where);this.#validate(candidate);return candidate}
 applied(){return structuredClone(this.#applied)}
 draft(){return structuredClone(this.#draft)}
 token(){return this.#revision}
 dirty(){return this.#error!=null||this.#pending||identity({where:this.#draft})!==identity({where:this.#applied})}
 snapshot(){return {applied:this.applied(),draft:this.draft(),error:structuredClone(this.#error),pending:this.#pending,revision:this.#revision,dirty:this.dirty()}}
 stage(where,{error=null,pending=false}={}){
  const next=structuredClone(where);let failure=structuredClone(error);
  if(failure==null&&!pending)try{this.#checked(next)}catch(error){failure=error instanceof Error?error.message:String(error);if(!failure)failure='Invalid predicate draft.'}
  this.#draft=next;this.#error=failure;this.#pending=!!pending;return ++this.#revision;
 }
 candidate(){
  if(this.#error!=null)throw Error(this.#error instanceof Error?this.#error.message:String(this.#error)||'Invalid predicate draft.');
  if(this.#pending)throw Error('Check or discard the pending predicate draft.');
  return this.#checked(this.#draft);
 }
 accept(token){
  if(token!==this.#revision)return false;
  let candidate;try{candidate=this.candidate()}catch{return false}
  // An owner validator may retire this draft while checking it.
  if(token!==this.#revision)return false;
  this.#applied=structuredClone(candidate);this.#draft=candidate;this.#error=null;this.#pending=false;this.#revision++;return true;
 }
 discard(){this.#draft=this.applied();this.#error=null;this.#pending=false;return ++this.#revision}
 reset(where){const checked=this.#checked(where);this.#applied=structuredClone(checked);this.#draft=checked;this.#error=null;this.#pending=false;return ++this.#revision}
 retire(){return ++this.#revision}
}
