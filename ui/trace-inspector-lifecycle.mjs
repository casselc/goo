/* Trace owner keyboard policy. Native dialog supplies modal focus containment. */
const nativeEscapeOwners='details,input,textarea,select,[contenteditable]:not([contenteditable="false"]),[role="combobox"],[role="listbox"]';
function closeOnEscape(event,host){const owner=event.target?.closest?.(nativeEscapeOwners);return event.key==='Escape'&&!event.defaultPrevented&&!event.isComposing&&!!host?.contains(event.target)&&(!owner||!host.contains(owner))}
export {closeOnEscape};
function containTab(event,host){
 if(event.key!=='Tab'||event.defaultPrevented||event.altKey||event.ctrlKey||event.metaKey||!host?.contains(event.target))return false;
 const candidates=[...host.querySelectorAll('button,a[href],input,select,textarea,summary,[tabindex]')].filter(e=>!e.disabled&&e.tabIndex>=0&&!e.closest('[hidden],[inert],[aria-hidden="true"]')&&e.getClientRects().length),active=host.ownerDocument.activeElement;
 if(!candidates.length)return false;const first=candidates[0],last=candidates.at(-1),target=event.shiftKey&&active===first?last:!event.shiftKey&&active===last?first:null;if(!target)return false;event.preventDefault();event.stopPropagation();target.focus();return true;
}
export {containTab};
