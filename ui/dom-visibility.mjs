/* Local action eligibility only; no subscriptions or source operations. */
function actionVisible(element){
 if(!element?.isConnected||document.hidden||element.closest('[hidden], [aria-hidden="true"]')||!element.getClientRects().length)return false;
 // A closed disclosure hides its content, but its first summary stays actionable.
 for(let closed=element.closest('details:not([open])');closed;closed=closed.parentElement?.closest('details:not([open])'))if(!closed.querySelector(':scope > summary')?.contains(element))return false;
 return true;
}
// A checked asynchronous completion may hand off while its own opener is
// deliberately disabled. The owner must still recheck its request/source lease.
function completionControlReady(element){
 if(!actionVisible(element)||element.closest('[inert]')||element.getAttribute('aria-disabled')==='true')return false;
 // Native disabled fieldsets still withdraw the action. Their first legend is
 // exempt, including when fieldsets nest; the opener's own disabled flag is not.
 for(let fieldset=element.parentElement?.closest('fieldset:disabled');fieldset;fieldset=fieldset.parentElement?.closest('fieldset:disabled'))if(!fieldset.querySelector(':scope > legend')?.contains(element))return false;
 return true;
}
function nativeControlReady(element){return completionControlReady(element)&&!element.disabled&&!element.matches(':disabled')}
// Explicit retained returns may reveal native disclosure ancestry. Invocation
// never reveals controls, and hidden/inert/disabled owners remain withdrawn.
function focusCandidate(element,{disclosures=false}={}){
 if(!element?.isConnected||document.hidden||element.closest('[hidden], [aria-hidden="true"], [inert]')||element.disabled||element.matches(':disabled')||element.getAttribute('aria-disabled')==='true')return false;
 for(let parent=element;parent;parent=parent.parentElement){const style=globalThis.getComputedStyle?.(parent);if(style?.display==='none'||['hidden','collapse'].includes(style?.visibility))return false;}
 if(disclosures){const ancestors=[];for(let parent=element.parentElement;parent;parent=parent.parentElement)if(parent.matches('details:not([open])')&&!parent.querySelector(':scope > summary')?.contains(element))ancestors.push(parent);for(const details of ancestors.reverse())details.open=true;}
 return nativeControlReady(element);
}
function revealFocus(origin,fallback,options){
 const target=focusCandidate(origin,options)?origin:focusCandidate(fallback)?fallback:null;
 target?.scrollIntoView({block:'center',inline:'nearest',behavior:'instant'});
 target?.focus({preventScroll:true});return target;
}
export {actionVisible,revealFocus,nativeControlReady,completionControlReady};
