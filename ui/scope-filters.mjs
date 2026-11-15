import * as FieldRefinement from './field-refinement.mjs';
import {validate} from './predicate-validation.mjs';
import {PredicateDraft} from './predicate-draft.mjs';

// Chip provenance is presentation state; the shared draft owns predicate adoption.
export class ScopeFilters {
 #base=null; #chips=[]; #appliedBase=null; #appliedChips=[];
 #state=new PredicateDraft({validate:where=>{let error;for(const signal of ['logs','spans'])try{validate(where,signal);return}catch(e){error=e}throw error}});
 get base(){return structuredClone(this.#base)}
 set base(value){this.#base=structuredClone(value);this.#stage()}
 get draft(){return structuredClone(this.#chips)}
 get applied(){return structuredClone(this.#appliedChips)}
 #project(){return this.#base?(this.#chips.length?{all:[structuredClone(this.#base),...structuredClone(this.#chips)]}:structuredClone(this.#base)):this.#chips.length?{all:structuredClone(this.#chips)}:undefined}
 #stage(){this.#state.stage(this.#project())}
 stage(field,value,mode,signal){
  const p=FieldRefinement.predicate(field,value,mode,signal),key=x=>{const leaf=x.any?x.any[1]:x;return JSON.stringify([leaf.field.kind,leaf.field.key,['exists','missing'].includes(leaf.op)?null:leaf.value??''])},i=this.#chips.findIndex(x=>key(x)===key(p)),next=structuredClone(this.#chips);
  if(i>=0)next[i]=p;else{if(next.length>=16)throw Error('Use at most 16 live filters.');next.push(p)}
  const where={all:[...(this.#base?[this.#base]:[]),...next]};let accepted=false,error;for(const candidate of signal?[signal]:['logs','spans'])try{validate(where,candidate);accepted=true;break}catch(e){error=e}if(!accepted)throw error;
  this.#chips=next;this.#stage();
 }
 replace(where,options){this.#base=structuredClone(where||null);this.#chips=[];this.#state.stage(this.#project(),options)}
 report(error,pending=false){this.#state.stage(this.#state.draft(),{error,pending})}
 remove(index){this.#chips.splice(index,1);this.#stage()}
 apply(){if(!this.#state.accept(this.#state.token())){this.#state.candidate();throw Error('Predicate draft changed. Apply explicitly again.')}this.#appliedChips=structuredClone(this.#chips);this.#appliedBase=structuredClone(this.#base)}
 discard(){this.#state.discard();this.#base=structuredClone(this.#appliedBase);this.#chips=structuredClone(this.#appliedChips)}
 reset(){this.restore(undefined)}
 restore(where){this.#state.reset(where||undefined);this.#base=structuredClone(where||null);this.#appliedBase=structuredClone(this.#base);this.#chips=[];this.#appliedChips=[]}
 get dirty(){return this.#state.dirty()}
 get blocked(){try{this.#state.candidate();return false}catch{return true}}
 where(applied=false){return applied?this.#state.applied():this.#state.draft()}
}
