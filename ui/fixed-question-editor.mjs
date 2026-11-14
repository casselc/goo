import {actionVisible} from './dom-visibility.mjs';
/* Local read/Edit presentation only. The existing question session owns execution. */
export function mount({container,button,editor,state,error=()=>null,current=()=>true,onVisibility=()=>{}}){
 let expanded=false,disposed=false,lease=null,lastApplied=null,lastTerminal=null;
 const dirty=()=>{const s=state();return !!error()||s.phase==='draft'&&!s.requested;};
 function canEdit(){return !disposed&&current()&&!editor.hidden&&!editor.inert&&!editor.closest?.('[inert]')&&actionVisible(editor);}
 function update(){if(disposed)return;const s=state(),required=dirty()||['failed','cancelled'].includes(s.phase),initial=s.phase==='initial';
  if(s.phase==='applied'&&lastApplied!==s.generation){expanded=false;lastApplied=s.generation;}
  if(['failed','cancelled'].includes(s.phase)&&lastTerminal!==s.generation){expanded=true;lastTerminal=s.generation;}
  const open=initial||required||expanded;container.classList.toggle('fixed-question',!initial);container.classList.toggle('fixed-editing',open);editor.hidden=!open;editor.inert=!open||!current();onVisibility(open&&current());button.hidden=initial;button.disabled=required;button.textContent=required?'Editing question':open?'Hide question controls':'Edit question';button.setAttribute('aria-expanded',String(open));button.setAttribute('aria-controls',editor.id);button.title=required?'Run a valid draft explicitly before hiding its controls.':'';
  const owner=lease={};button.onclick=()=>{if(disposed||lease!==owner||!current()||button.disabled||button.hidden||!actionVisible(button)||button.closest?.('[inert]'))return;const focusOwned=editor.contains?.(document.activeElement);expanded=!open;update();if(expanded){editor.tabIndex=-1;editor.focus?.({preventScroll:true});editor.scrollIntoView?.({block:'start'});}else if(focusOwned)button.focus?.({preventScroll:true});};
 }
 function reset(){expanded=false;lease=null;lastApplied=lastTerminal=null;}
 function open(){if(disposed||!current())return;expanded=true;update();}
 function block(event){if(canEdit())return;event.preventDefault();event.stopImmediatePropagation();}
 for(const type of ['click','submit','input','change','keydown'])editor.addEventListener(type,block,true);
 return {update,canEdit,reset,open,dispose(){disposed=true;lease=null;for(const type of ['click','submit','input','change','keydown'])editor.removeEventListener(type,block,true);}};
}
