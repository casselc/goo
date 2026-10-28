/* Product navigation owns layout, never telemetry execution or evidence mutation. */
import * as QuestionProtocol from './question-protocol.mjs';
const global=globalThis;
const routes={explore:{title:'Explore',description:'Follow a question through captured telemetry.',measurementsDescription:'Inspect measurements in their selected time window.'},services:{title:'Services & resources',description:'Discover services or host resources and inspect their reported telemetry.'},operations:{title:'Operations',description:'Investigate application behavior through shared definitions.'},investigations:{title:'Saved investigations',description:'Reopen evidence, preserve context, and compare revisions.'},definitions:{title:'Definitions',description:'Inspect the contracts behind your investigations.'},examples:{title:'Examples',description:'Investigate independent sample applications using shared telemetry packages.'}};
const protocol=QuestionProtocol;
function routeFromHash(hash){if(hash==='#activity-views')return 'investigations';if(/^#(?:investigation|capture|document|annotation|evidence)=[a-f0-9]{64}$/.test(hash))return 'investigations';const key=hash.replace(/^#/,'');return Object.hasOwn(routes,key)?key:'explore';}
function mount({onExamples}={}){
  let current='',definitionFamily='records',measurements=false;const buttons=[...document.querySelectorAll('[data-workspace]')],panels=[...document.querySelectorAll('[data-workspace-panel]')];
  const sidebar=document.querySelector('.app-sidebar'),menu=sidebar.querySelector('.entry-menu');
  function closeMenu({focus=false}={}){menu.setAttribute('aria-expanded','false');sidebar.classList.remove('navigation-expanded');if(focus)menu.focus()}
  menu.onclick=()=>{const expanded=menu.getAttribute('aria-expanded')!=='true';menu.setAttribute('aria-expanded',String(expanded));sidebar.classList.toggle('navigation-expanded',expanded)};
  sidebar.addEventListener('keydown',event=>{if(event.key==='Escape'&&menu.getAttribute('aria-expanded')==='true'){event.preventDefault();closeMenu({focus:true})}});
  function syncContext(){document.getElementById('workspaceDescription').textContent=current==='explore'&&measurements?routes.explore.measurementsDescription:routes[current].description;const lookback=document.getElementById('minutes'),scope=lookback.closest('.scope-control'),label=scope.querySelector('label[for="minutes"]');scope.hidden=current==='investigations'||(current==='definitions'&&definitionFamily!=='contracts')||(current==='explore'&&measurements);scope.closest('.app-topbar').hidden=scope.hidden;label.textContent=({services:'New discovery lookback',operations:'New operation read lookback',definitions:'HTTP preview lookback',examples:'New example read lookback'})[current]||'New query lookback';lookback.setAttribute('aria-label',label.textContent+' minutes');lookback.title='Default for a new read. Saved evidence and fixed question bounds stay unchanged until an explicit scope edit.';}
  global.addEventListener('goo:explore-mode',event=>{measurements=event.detail.measurements;syncContext()});
  function setDefinitionFamily(family){definitionFamily=family;syncContext()}
  function navigate(route,{writeHash=true,focus=false}={}){if(!Object.hasOwn(routes,route))route='explore';current=route;closeMenu();syncContext();for(const panel of panels)panel.hidden=panel.dataset.workspacePanel!==route;for(const button of buttons){if(button.dataset.workspace===route)button.setAttribute('aria-current','page');else button.removeAttribute('aria-current')}
   document.getElementById('workspaceTitle').textContent=routes[route].title;document.title=routes[route].title+' · goo';
   if(writeHash&&routeFromHash(location.hash)!==route)history.pushState(null,'','#'+route);if(focus)document.getElementById('workspaceTitle').focus({preventScroll:true});if(route==='examples')onExamples?.();global.dispatchEvent(new CustomEvent('goo:workspace',{detail:{route}}));
  }
  for(const button of buttons)button.onclick=()=>navigate(button.dataset.workspace,{focus:true});global.addEventListener('popstate',()=>navigate(routeFromHash(location.hash),{writeHash:false}));global.addEventListener('hashchange',()=>navigate(routeFromHash(location.hash),{writeHash:false}));navigate(routeFromHash(location.hash),{writeHash:false});
  return {navigate,current:()=>current,setDefinitionFamily};
 }
const api={...protocol,routes,routeFromHash,mount};
const questionFromEvidence=api.questionFromEvidence;
const captureQuestion=api.captureQuestion;
const evidenceDifference=api.evidenceDifference;
const recordStep=api.recordStep;
const measurementStep=api.measurementStep;
const validateStep=api.validateStep;
const stepLink=api.stepLink;
const readStepLink=api.readStepLink;
const sameWindow=api.sameWindow;
const questionLink=api.questionLink;
const readQuestionLink=api.readQuestionLink;
const validateLink=api.validateLink;
const validateNumericPredicate=api.validateNumericPredicate;
export {questionFromEvidence,captureQuestion,evidenceDifference,recordStep,measurementStep,validateStep,stepLink,readStepLink,sameWindow,questionLink,readQuestionLink,validateLink,validateNumericPredicate,routes,routeFromHash,mount};
