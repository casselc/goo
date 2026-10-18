/* Product navigation owns layout, never telemetry execution or evidence mutation. */
(function(global){
 const routes={explore:{title:'Explore',description:'Follow a question through captured telemetry.'},services:{title:'Services',description:'Discover emitting services and open focused investigations.'},operations:{title:'Operations',description:'Investigate application behavior through shared definitions.'},investigations:{title:'Saved investigations',description:'Reopen evidence, preserve context, and compare revisions.'},definitions:{title:'Definitions',description:'Inspect the contracts behind your investigations.'},examples:{title:'Examples',description:'Explore reference workflows in the reservation sample.'}};
 function routeFromHash(hash){if(/^#investigation=[a-f0-9]{64}$/.test(hash))return 'investigations';const key=hash.replace(/^#/,'');return Object.hasOwn(routes,key)?key:'explore';}
 function mount({onExamples}={}){
  let current='';const buttons=[...document.querySelectorAll('[data-workspace]')],panels=[...document.querySelectorAll('[data-workspace-panel]')];
  function navigate(route,{writeHash=true,focus=false}={}){if(!Object.hasOwn(routes,route))route='explore';current=route;for(const panel of panels)panel.hidden=panel.dataset.workspacePanel!==route;for(const button of buttons){if(button.dataset.workspace===route)button.setAttribute('aria-current','page');else button.removeAttribute('aria-current')}
   document.getElementById('workspaceTitle').textContent=routes[route].title;document.getElementById('workspaceDescription').textContent=routes[route].description;document.title=routes[route].title+' · goo';
   const operations=document.getElementById('operations'),host=document.getElementById(route==='investigations'?'savedEvidenceHost':'liveEvidenceHost');if(route==='operations'||route==='investigations')host.append(operations);
   if(writeHash)history.pushState(null,'','#'+route);if(focus)document.getElementById('workspaceTitle').focus({preventScroll:true});if(route==='examples')onExamples?.();global.dispatchEvent(new CustomEvent('goo:workspace',{detail:{route}}));
  }
  for(const button of buttons)button.onclick=()=>navigate(button.dataset.workspace,{focus:true});global.addEventListener('popstate',()=>navigate(routeFromHash(location.hash),{writeHash:false}));global.addEventListener('hashchange',()=>navigate(routeFromHash(location.hash),{writeHash:false}));navigate(routeFromHash(location.hash),{writeHash:false});
  return {navigate,current:()=>current};
 }
 const api={routes,routeFromHash,mount};if(typeof module!=='undefined')module.exports=api;else global.GooWorkspace=api;
})(typeof window==='undefined'?globalThis:window);
