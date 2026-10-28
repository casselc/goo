/* Local action eligibility only; no subscriptions or source operations. */
  function actionVisible(element){return !!element?.isConnected&&!document.hidden&&!element.closest('[hidden], [aria-hidden="true"], details:not([open])')&&element.getClientRects().length>0}
export {actionVisible};
