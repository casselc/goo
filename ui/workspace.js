/* Product navigation owns layout, never telemetry execution or evidence mutation. */
(function(global){
 const routes={explore:{title:'Explore',description:'Follow a question through captured telemetry.'},services:{title:'Services',description:'Discover emitting services and open focused investigations.'},operations:{title:'Operations',description:'Investigate application behavior through shared definitions.'},investigations:{title:'Saved investigations',description:'Reopen evidence, preserve context, and compare revisions.'},definitions:{title:'Definitions',description:'Inspect the contracts behind your investigations.'},examples:{title:'Examples',description:'Investigate independent sample applications using shared telemetry packages.'}};
 // Versioned links contain questions, never evidence or source credentials.
 function recordStep(q,view='Records'){return {version:1,kind:'record',scope:{environment:q.environment,fromMs:q.fromMs,toMs:q.toMs},view,record:structuredClone(q)}}
 function measurementStep(q){return {version:1,kind:'measurement',scope:{environment:q.definition.environment,fromMs:q.fromMs,toMs:q.toMs},view:'Overview',measurement:structuredClone(q.definition)}}
 function validateStep(s){
  const b=s?.scope;if(s?.version!==1||!b||typeof b.environment!=='string'||!b.environment||!Number.isSafeInteger(b.fromMs)||!Number.isSafeInteger(b.toMs)||b.fromMs<0||b.toMs<=b.fromMs||b.toMs-b.fromMs>7200000||b.toMs>=4102444800000)throw Error('Invalid question step scope.');
  if(s.kind==='record'){if(s.measurement||!s.record||s.record.environment!==b.environment||s.record.fromMs!==b.fromMs||s.record.toMs!==b.toMs)throw Error('Record step scope differs.');validateLink({version:1,question:s.record,view:s.view})}
  else if(s.kind==='measurement'){if(s.record||s.view!=='Overview'||!s.measurement||s.measurement.environment!==b.environment||typeof s.measurement.source!=='object')throw Error('Invalid measurement step.');}
  else throw Error('Unsupported question step.');return s;
 }
 function stepLink(s){validateStep(s);const hash='#step='+encodeURIComponent(JSON.stringify(s));if(hash.length>48000)throw Error('Question step is too large to share.');return hash}
 function readStepLink(hash){if(!hash.startsWith('#step='))return null;if(hash.length>48000)throw Error('Question step is too large.');return validateStep(JSON.parse(decodeURIComponent(hash.slice(6))))}
 function sameWindow(a,b){return a.environment===b.environment&&a.fromMs===b.fromMs&&a.toMs===b.toMs}
 function questionLink(question,view='Records'){
  const value={version:1,question:structuredClone(question),view};validateLink(value);
  const hash=stepLink(recordStep(question,view));if(hash.length>16000)throw Error('Question is too large to share in a link.');return hash;
 }
 function validateLink(value){
  const q=value?.question;if(value?.version!==1||!['Records','Distribution'].includes(value.view)||!q||!['spans','logs'].includes(q.signal)||typeof q.environment!=='string'||!q.environment.length||q.environment.length>256||!Number.isSafeInteger(q.fromMs)||!Number.isSafeInteger(q.toMs)||q.fromMs<0||q.toMs<=q.fromMs||q.toMs-q.fromMs>7200000||q.toMs>8640000000000000)throw Error('Invalid shared question or fixed window.');
  const allowed=['signal','environment','fromMs','toMs','limit','offset','where','groupBy','numeric','summary','spanStatus','bucketMs','groupTimeSeries'];if(Object.keys(q).some(k=>!allowed.includes(k)))throw Error('Unsupported shared question field.');
  function field(f){if(!f||!['column','attribute','resource'].includes(f.kind)||typeof f.key!=='string'||!f.key.length||f.key.length>512)throw Error('Invalid shared field.');}
  let count=0;function predicate(p,depth=0){if(++count>64||depth>8||!p||typeof p!=='object')throw Error('Invalid shared selectors.');if(p.all||p.any){const a=p.all||p.any;if(!Array.isArray(a)||!a.length||p.all&&p.any)throw Error('Invalid shared selectors.');a.forEach(x=>predicate(x,depth+1));}else{field(p.field);if(!['eq','ne','contains','exists','missing','gte','lte'].includes(p.op)||p.value!==undefined&&typeof p.value!=='string')throw Error('Invalid shared predicate.');}}
  if(q.where)predicate(q.where);if(q.groupBy){if(!Array.isArray(q.groupBy)||q.groupBy.length>2)throw Error('Invalid shared grouping.');q.groupBy.forEach(field);}
  if(q.limit!==undefined&&(!Number.isInteger(q.limit)||q.limit<1||q.limit>200)||q.offset!==undefined&&(!Number.isInteger(q.offset)||q.offset<0||q.offset>1000))throw Error('Invalid shared page.');
  if(q.numeric){field(q.numeric.field);if(!['attribute','resource'].includes(q.numeric.field.kind)||!['int64','uint64'].includes(q.numeric.type)||typeof q.numeric.unit!=='string'||q.numeric.unit.length>128||q.numeric.parse!=='canonical-decimal')throw Error('Invalid shared numeric binding.');}
  if(q.summary!==undefined&&!['','span-duration'].includes(q.summary)||q.spanStatus!==undefined&&typeof q.spanStatus!=='boolean'||q.groupTimeSeries!==undefined&&typeof q.groupTimeSeries!=='boolean'||q.bucketMs!==undefined&&![0,1000,10000,60000,300000].includes(q.bucketMs))throw Error('Invalid shared presentation.');
  return value;
 }
 function readQuestionLink(hash){if(hash.startsWith('#step=')){const s=readStepLink(hash);return s.kind==='record'?{version:1,question:s.record,view:s.view}:null}if(!hash.startsWith('#question='))return null;if(hash.length>16000)throw Error('Question link is too large.');return validateLink(JSON.parse(decodeURIComponent(hash.slice(10))));}
 function routeFromHash(hash){if(/^#(?:investigation|capture|document)=[a-f0-9]{64}$/.test(hash))return 'investigations';const key=hash.replace(/^#/,'');return Object.hasOwn(routes,key)?key:'explore';}
 function mount({onExamples}={}){
  let current='';const buttons=[...document.querySelectorAll('[data-workspace]')],panels=[...document.querySelectorAll('[data-workspace-panel]')];
  function navigate(route,{writeHash=true,focus=false}={}){if(!Object.hasOwn(routes,route))route='explore';current=route;for(const panel of panels)panel.hidden=panel.dataset.workspacePanel!==route;for(const button of buttons){if(button.dataset.workspace===route)button.setAttribute('aria-current','page');else button.removeAttribute('aria-current')}
   document.getElementById('workspaceTitle').textContent=routes[route].title;document.getElementById('workspaceDescription').textContent=routes[route].description;document.title=routes[route].title+' · goo';
   const operations=document.getElementById('operations'),host=document.getElementById(route==='investigations'?'savedEvidenceHost':'liveEvidenceHost');if(route==='operations'||route==='investigations')host.append(operations);
   if(writeHash&&routeFromHash(location.hash)!==route)history.pushState(null,'','#'+route);if(focus)document.getElementById('workspaceTitle').focus({preventScroll:true});if(route==='examples')onExamples?.();global.dispatchEvent(new CustomEvent('goo:workspace',{detail:{route}}));
  }
  for(const button of buttons)button.onclick=()=>navigate(button.dataset.workspace,{focus:true});global.addEventListener('popstate',()=>navigate(routeFromHash(location.hash),{writeHash:false}));global.addEventListener('hashchange',()=>navigate(routeFromHash(location.hash),{writeHash:false}));navigate(routeFromHash(location.hash),{writeHash:false});
  return {navigate,current:()=>current};
 }
 const api={recordStep,measurementStep,validateStep,stepLink,readStepLink,sameWindow,routes,routeFromHash,questionLink,readQuestionLink,validateLink,mount};if(typeof module!=='undefined')module.exports=api;else global.GooWorkspace=api;
})(typeof window==='undefined'?globalThis:window);
