/* App-owned explicit detail opening/return. No source or subscription behavior. */
import {actionVisible,revealFocus} from './dom-visibility.mjs';
let nextID=0;
function mount(host,{fallback,active=()=>true}={}){
 const heading=host.querySelector('.ui-inspector-heading>:is(h2,h3,h4,h5,h6)'),close=host.querySelector('[aria-label="Close selected item details"]');
 if(!host.id)host.id='activity-details-'+(++nextID);
 heading.id=host.id+'-heading';heading.tabIndex=-1;host.setAttribute('aria-labelledby',heading.id);
 let opener=null;
 function reset(){opener?.setAttribute('aria-expanded','false');opener=null}
 function recover(){if(active())revealFocus(fallback?.())}
 function dismiss(){const origin=opener;reset();if(active())revealFocus(origin?.disabled?null:origin,fallback?.());return false}
 function open(origin){if(!active()||host.hidden||!actionVisible(origin))return false;reset();opener=origin;origin.setAttribute('aria-controls',host.id);origin.setAttribute('aria-expanded','true');revealFocus(heading,fallback?.());return true}
 const escape=event=>{if(event.key==='Escape'&&!event.defaultPrevented&&!event.target.closest('details,input,textarea,select,[contenteditable="true"],[role="combobox"],[role="listbox"]')&&!host.hidden&&active()){event.preventDefault();event.stopPropagation();close.click()}};
 host.addEventListener('keydown',escape);
 return {open,close:dismiss,reset,recover,dispose(){reset();host.removeEventListener('keydown',escape)}};
}
export {mount};
