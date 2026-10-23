/* Pure application status policy. No requests, storage or component lifecycle. */
(function(global){
 function failure(error,{tail=false,journaled=false,localInspection=false,inRecovery=false}={}){
  const status=Number(error?.status)||0;
  if(error?.attempted===false)return {state:'local',message:'Save was not sent. '+error.message+' Keep this view open to retry unchanged.'};
  const recovery=journaled?(inRecovery?' Retry keeps the original request.':' Retry uses the unchanged request; Saved contains browser-local recovery.'):' Keep this view open; browser-local recovery is not confirmed.';
  if(tail&&status===410)return {state:'uncommitted',message:'This request was not committed; its observation proof expired or is unavailable.'+(localInspection?' Local inspection remains available.':'')+recovery};
  if(status===409)return {state:'conflict',message:'Request conflicts with an existing save. The original request has been kept for review.'+recovery};
  if(status===503)return {state:'unconfigured',message:'Retention is not configured. The original save request is kept.'+recovery};
  if(status===400)return {state:'refused',message:'Save request refused: '+error.message+recovery};
  return {state:'pending',message:'Save acknowledgment unavailable; commit outcome is unknown.'+recovery};
 }
 function availability(state){return {checking:'Checking retention configuration…',unconfigured:'Retention is not configured. Local inspection remains available.',unavailable:'Retention configuration could not be checked. Local inspection remains available.'}[state]||''}
 const api={failure,availability};if(typeof module!=='undefined')module.exports=api;else global.GooRetentionStatus=api;
})(typeof window==='undefined'?globalThis:window);
