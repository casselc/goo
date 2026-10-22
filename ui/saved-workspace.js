/* Saved task composition only. Existing owners retain drafts, requests and custody. */
(function(global){
 function mount(){
  const el=id=>document.getElementById(id),panel=document.querySelector('[data-workspace-panel="investigations"]');
  function openDocuments(){el('documentPanel').open=true;if(!el('documentEditor').hidden)el('activeInvestigationOpen').click();el('captureLibraryPanel').open=false;el('documentPanel').scrollIntoView({block:'start'});if(el('documentLibrary').hidden)el('documentLibraryLoad').click();el('documentSearch').focus({preventScroll:true})}
  function openEvidence(){el('documentPanel').open=false;el('captureLibraryPanel').open=true;el('captureLibraryPanel').scrollIntoView({block:'start'});el('captureCatalogRefresh').focus({preventScroll:true})}
  panel.querySelector('[data-saved-open="investigations"]').onclick=openDocuments;
  panel.querySelector('[data-saved-open="evidence"]').onclick=openEvidence;
  el('savedRecoveryNotice').onclick=()=>{el('savedTools').open=true;const recovery=el('savedRecovery').querySelector('details');recovery.open=true;recovery.querySelector('summary').tabIndex=-1;recovery.querySelector('summary').focus();recovery.scrollIntoView({block:'start'})};
  function selectEvidence(){el('documentPanel').open=false;el('captureLibraryPanel').open=false}
  global.addEventListener('goo:show-general-capture',selectEvidence);global.addEventListener('goo:show-retained-evidence',selectEvidence);global.addEventListener('goo:show-operation-capture',()=>{selectEvidence();el('savedTools').open=false});
  global.addEventListener('goo:show-document',()=>{el('captureLibraryPanel').open=false;el('savedTools').open=false});
  function revealLink(){if(location.hash.startsWith('#annotation='))el('savedTools').open=true;}
  global.addEventListener('hashchange',revealLink);global.addEventListener('popstate',revealLink);revealLink();
 }
 global.GooSavedWorkspace={mount};
})(window);
