/* Synchronous local-definition handlers only. Never use this to revive a source request. */
export function retainedDefinitionControls(read){
 let token={},generation=read().generation;
 const current=()=>{const state=read();return token!==null&&state.eligible===true&&state.generation===generation;};
 return Object.freeze({
  current,
  bind(action){if(typeof action!=='function')throw TypeError('Supply a local definition action.');const owner=token;return function(...args){if(owner===null||owner!==token||!current())return;return action.apply(this,args);};},
  retire(){token=null;},
  renew(){const state=read();if(state.eligible!==true)return false;generation=state.generation;token={};return true;}
 });
}

// Preserve raw local edits across an existing definition editor's remount.
// Disabled state and event handlers belong to the new owner and are never copied.
export function definitionControlState(root){
 return [...root.querySelectorAll('input,textarea,select,details')].map((element,index)=>({index,tag:element.tagName,label:element.getAttribute('aria-label'),type:element.type,value:element.value,checked:element.checked,open:element.open,start:element.selectionStart,end:element.selectionEnd,direction:element.selectionDirection,focused:element===globalThis.document?.activeElement,...(element.tagName==='SELECT'?{options:[...element.options].map(option=>option.cloneNode(true))}:{})}));
}
export function restoreDefinitionControlState(root,state){
 const elements=[...root.querySelectorAll('input,textarea,select,details')];
 for(const saved of state){const element=elements[saved.index];if(!element||element.tagName!==saved.tag||element.getAttribute('aria-label')!==saved.label||element.type!==saved.type)continue;
  if(saved.options)element.replaceChildren(...saved.options.map(option=>option.cloneNode(true)));
  if(saved.value!==undefined)element.value=saved.value;if(saved.checked!==undefined)element.checked=saved.checked;if(saved.open!==undefined)element.open=saved.open;
  if(Number.isInteger(saved.start)&&typeof element.setSelectionRange==='function'){try{element.setSelectionRange(saved.start,saved.end,saved.direction)}catch{}}
  if(saved.focused)element.focus({preventScroll:true});
 }
}
