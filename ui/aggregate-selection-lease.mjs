/* Synchronous selection handoff only. Each acquisition owner supplies its own
   captured render/scope and native eligibility checks; holding may retire them. */
export function holdSelection(current,hold,handoff){
 if(!current())return false;
 hold();
 if(!current())return false;
 handoff?.();
 return true;
}
