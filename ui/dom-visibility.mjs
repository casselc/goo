/* Local action eligibility only; no subscriptions or source operations. */
function actionVisible(element){
 if(!element?.isConnected||document.hidden||element.closest('[hidden], [aria-hidden="true"]')||!element.getClientRects().length)return false;
 // A closed disclosure hides its content, but its first summary stays actionable.
 for(let closed=element.closest('details:not([open])');closed;closed=closed.parentElement?.closest('details:not([open])'))if(!closed.querySelector(':scope > summary')?.contains(element))return false;
 return true;
}
function revealFocus(origin,fallback){
 const target=actionVisible(origin)?origin:fallback;
 target?.scrollIntoView({block:'center',inline:'nearest',behavior:'instant'});
 target?.focus({preventScroll:true});return target;
}
export {actionVisible,revealFocus};
