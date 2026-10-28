/* Local action eligibility only; no subscriptions or source operations. */
  function actionVisible(element){return !!element?.isConnected&&!document.hidden&&!element.closest('[hidden], [aria-hidden="true"], details:not([open])')&&element.getClientRects().length>0}
function revealFocus(origin,fallback){
 const target=actionVisible(origin)?origin:fallback;
 target?.scrollIntoView({block:'center',inline:'nearest',behavior:'instant'});
 target?.focus({preventScroll:true});return target;
}
export {actionVisible,revealFocus};
