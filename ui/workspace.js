/* Product navigation owns layout, never telemetry execution or evidence mutation. */
(function(global){
 const routes={explore:{title:'Explore',description:'Follow a question through captured telemetry.'},services:{title:'Services & resources',description:'Discover services or host resources and inspect their reported telemetry.'},operations:{title:'Operations',description:'Investigate application behavior through shared definitions.'},investigations:{title:'Saved investigations',description:'Reopen evidence, preserve context, and compare revisions.'},definitions:{title:'Definitions',description:'Inspect the contracts behind your investigations.'},examples:{title:'Examples',description:'Investigate independent sample applications using shared telemetry packages.'}};
 const protocol=typeof module!=='undefined'?require('./question-protocol'):global.GooQuestionProtocol;
 function routeFromHash(hash){if(hash==='#activity-views')return 'investigations';if(/^#(?:investigation|capture|document|annotation|evidence)=[a-f0-9]{64}$/.test(hash))return 'investigations';const key=hash.replace(/^#/,'');return Object.hasOwn(routes,key)?key:'explore';}
 function mount({onExamples}={}){
  let current='';const buttons=[...document.querySelectorAll('[data-workspace]')],panels=[...document.querySelectorAll('[data-workspace-panel]')];
  function navigate(route,{writeHash=true,focus=false}={}){if(!Object.hasOwn(routes,route))route='explore';current=route;for(const panel of panels)panel.hidden=panel.dataset.workspacePanel!==route;for(const button of buttons){if(button.dataset.workspace===route)button.setAttribute('aria-current','page');else button.removeAttribute('aria-current')}
   const lookback=document.getElementById('minutes'),scope=lookback.closest('.scope-control'),label=scope.querySelector('label[for="minutes"]');scope.hidden=route==='investigations';label.textContent=({services:'New discovery lookback',operations:'New operation read lookback',definitions:'HTTP preview lookback',examples:'New example read lookback'})[route]||'New query lookback';lookback.setAttribute('aria-label',label.textContent+' minutes');lookback.title='Default for a new read. Saved evidence and fixed question bounds stay unchanged until an explicit scope edit.';
   document.getElementById('workspaceTitle').textContent=routes[route].title;document.getElementById('workspaceDescription').textContent=routes[route].description;document.title=routes[route].title+' · goo';
   if(writeHash&&routeFromHash(location.hash)!==route)history.pushState(null,'','#'+route);if(focus)document.getElementById('workspaceTitle').focus({preventScroll:true});if(route==='examples')onExamples?.();global.dispatchEvent(new CustomEvent('goo:workspace',{detail:{route}}));
  }
  for(const button of buttons)button.onclick=()=>navigate(button.dataset.workspace,{focus:true});global.addEventListener('popstate',()=>navigate(routeFromHash(location.hash),{writeHash:false}));global.addEventListener('hashchange',()=>navigate(routeFromHash(location.hash),{writeHash:false}));navigate(routeFromHash(location.hash),{writeHash:false});
  return {navigate,current:()=>current};
 }
 const api={...protocol,routes,routeFromHash,mount};if(typeof module!=='undefined')module.exports=api;else global.GooWorkspace=api;
})(typeof window==='undefined'?globalThis:window);
