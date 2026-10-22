/* Product navigation owns layout, never telemetry execution or evidence mutation. */
(function(global){
 const routes={explore:{title:'Explore',description:'Follow a question through captured telemetry.'},services:{title:'Services',description:'Discover emitting services and open focused investigations.'},operations:{title:'Operations',description:'Investigate application behavior through shared definitions.'},investigations:{title:'Saved investigations',description:'Reopen evidence, preserve context, and compare revisions.'},definitions:{title:'Definitions',description:'Inspect the contracts behind your investigations.'},examples:{title:'Examples',description:'Investigate independent sample applications using shared telemetry packages.'}};
 // Versioned links contain questions, never evidence or source credentials.
 function exactBounds(q){return q.fromNS!==undefined||q.toNS!==undefined?{fromNS:q.fromNS,toNS:q.toNS}:{}}
 function questionFromEvidence(receipt){
  if(!/^[a-f0-9]{64}$/.test(receipt?.id))throw Error('An exact capture ID is required.');
  const q=receipt.question;let step;if(q?.kind==='measurement'&&q.measurement)step=measurementStep(q.measurement);else if(q?.kind==='explore'&&q.explore)step=recordStep(q.explore);else throw Error('This evidence cannot become a reusable question.');
  validateStep(step);return {kind:'question',step,evidence:[receipt.id]};
 }
 function captureQuestion(step){validateStep(step);return step.kind==='record'?{kind:'explore',explore:structuredClone(step.record)}:{kind:'measurement',measurement:{definition:structuredClone(step.measurement),fromMs:step.scope.fromMs,toMs:step.scope.toMs,...exactBounds(step.scope)}}}
 function evidenceDifference(step,capture){
  const identify=global.GooQuestionSession?.identity||(typeof module!=='undefined'?require('./question-session').identity:null);const expected=captureQuestion(step),actual=capture,changes=[];if(expected.kind!==actual.kind)return ['Signal family differs'];
  const a=expected.explore||expected.measurement,b=actual.explore||actual.measurement;if(!b)return ['Question unavailable'];
  if((a.environment||a.definition?.environment)!==(b.environment||b.definition?.environment))changes.push('Environment differs');
  if(a.fromMs!==b.fromMs||a.toMs!==b.toMs||a.fromNS!==b.fromNS||a.toNS!==b.toNS)changes.push('Fixed window differs');
  function definition(q){if(q.definition){const d=structuredClone(q.definition);delete d.environment;return identify?identify(d):stable(d)}const d=structuredClone(q);for(const k of ['environment','fromMs','toMs','offset','limit'])delete d[k];return identify?identify(d):stable(d)}
  function stable(x){if(Array.isArray(x))return '['+x.map(stable).join(',')+']';if(x&&typeof x==='object')return '{'+Object.keys(x).sort().map(k=>JSON.stringify(k)+':'+stable(x[k])).join(',')+'}';return JSON.stringify(x)}
  if(definition(a)!==definition(b))changes.push('Question definition differs');return changes;
 }
 function recordStep(q,view='Records'){return {version:1,kind:'record',scope:{environment:q.environment,fromMs:q.fromMs,toMs:q.toMs},view,record:structuredClone(q)}}
 function measurementStep(q){return {version:1,kind:'measurement',scope:{environment:q.definition.environment,fromMs:q.fromMs,toMs:q.toMs,...exactBounds(q)},view:'Overview',measurement:structuredClone(q.definition)}}
 function validateStep(s){
  const b=s?.scope;if(s?.version!==1||!b||typeof b.environment!=='string'||!b.environment||!Number.isSafeInteger(b.fromMs)||!Number.isSafeInteger(b.toMs)||b.fromMs<0||b.toMs<=b.fromMs||b.toMs-b.fromMs>7200000||b.toMs>=4102444800000)throw Error('Invalid question step scope.');
  if(b.fromNS!==undefined||b.toNS!==undefined){if(s.kind!=='measurement'||s.measurement?.kind==='lifecycle'||typeof b.fromNS!=='string'||typeof b.toNS!=='string'||!/^(0|[1-9][0-9]*)$/.test(b.fromNS)||!/^(0|[1-9][0-9]*)$/.test(b.toNS))throw Error('Invalid exact native scope.');const f=BigInt(b.fromNS),t=BigInt(b.toNS);if(t<=f||t-f>7200000000000n||t>=4102444800000000000n||BigInt(b.fromMs)!==f/1000000n||BigInt(b.toMs)!==(t+999999n)/1000000n)throw Error('Exact scope differs from enclosing bounds.');}
  if(s.kind==='record'){if(s.measurement||!s.record||s.record.environment!==b.environment||s.record.fromMs!==b.fromMs||s.record.toMs!==b.toMs)throw Error('Record step scope differs.');validateLink({version:1,question:s.record,view:s.view})}
  else if(s.kind==='measurement'){if(s.record||s.view!=='Overview'||!s.measurement||s.measurement.environment!==b.environment||typeof s.measurement.source!=='object')throw Error('Invalid measurement step.');}
  else throw Error('Unsupported question step.');return s;
 }
 function stepLink(s){validateStep(s);const hash='#step='+encodeURIComponent(JSON.stringify(s));if(hash.length>48000)throw Error('Question step is too large to share.');return hash}
 function readStepLink(hash){if(!hash.startsWith('#step='))return null;if(hash.length>48000)throw Error('Question step is too large.');return validateStep(JSON.parse(decodeURIComponent(hash.slice(6))))}
 function sameWindow(a,b){return a.environment===b.environment&&a.fromMs===b.fromMs&&a.toMs===b.toMs&&a.fromNS===b.fromNS&&a.toNS===b.toNS}
 function questionLink(question,view='Records'){
  const value={version:1,question:structuredClone(question),view};validateLink(value);
  const hash=stepLink(recordStep(question,view));if(hash.length>16000)throw Error('Question is too large to share in a link.');return hash;
 }
 function validateLink(value){
  const q=value?.question;if(value?.version!==1||!['Records','Distribution'].includes(value.view)||!q||!['spans','logs'].includes(q.signal)||typeof q.environment!=='string'||!q.environment.length||q.environment.length>256||!Number.isSafeInteger(q.fromMs)||!Number.isSafeInteger(q.toMs)||q.fromMs<0||q.toMs<=q.fromMs||q.toMs-q.fromMs>7200000||q.toMs>8640000000000000)throw Error('Invalid shared question or fixed window.');
  const allowed=['signal','environment','fromMs','toMs','limit','offset','where','groupBy','numeric','summary','spanStatus','bucketMs','groupTimeSeries','reportedEdge'];if(Object.keys(q).some(k=>!allowed.includes(k)))throw Error('Unsupported shared question field.');
  function field(f){if(!f||!['column','attribute','resource'].includes(f.kind)||typeof f.key!=='string'||!f.key.length||f.key.length>512)throw Error('Invalid shared field.');}
  let count=0;function predicate(p,depth=0){if(++count>64||depth>8||!p||typeof p!=='object')throw Error('Invalid shared selectors.');if(p.all||p.any){const a=p.all||p.any;if(!Array.isArray(a)||!a.length||p.all&&p.any)throw Error('Invalid shared selectors.');a.forEach(x=>predicate(x,depth+1));}else{field(p.field);if(!['eq','ne','contains','exists','missing','gte','lte'].includes(p.op)||p.value!==undefined&&typeof p.value!=='string')throw Error('Invalid shared predicate.');}}
  if(q.reportedEdge){const e=q.reportedEdge,ns=n=>n&&['missing','value'].includes(n.mode)&&(!n.value||typeof n.value==='string'&&n.value.length<=128)&&!(n.mode==='missing'&&n.value);if(q.signal!=='spans'||!e.parent||!e.child||[e.parent,e.child].some(s=>typeof s.service!=='string'||s.service.length>128||!ns(s.namespace))||e.discoveryNamespace&&!ns(e.discoveryNamespace))throw Error('Invalid reported edge identity.');}
  if(q.where)predicate(q.where);if(q.groupBy){if(!Array.isArray(q.groupBy)||q.groupBy.length>2)throw Error('Invalid shared grouping.');q.groupBy.forEach(field);}
  if(q.limit!==undefined&&(!Number.isInteger(q.limit)||q.limit<1||q.limit>200)||q.offset!==undefined&&(!Number.isInteger(q.offset)||q.offset<0||q.offset>1000))throw Error('Invalid shared page.');
  if(q.numeric){field(q.numeric.field);if(!['attribute','resource'].includes(q.numeric.field.kind)||!['int64','uint64'].includes(q.numeric.type)||typeof q.numeric.unit!=='string'||q.numeric.unit.length>128||q.numeric.parse!=='canonical-decimal')throw Error('Invalid shared numeric binding.');}
  if(q.summary!==undefined&&!['','span-duration'].includes(q.summary)||q.spanStatus!==undefined&&typeof q.spanStatus!=='boolean'||q.groupTimeSeries!==undefined&&typeof q.groupTimeSeries!=='boolean'||q.bucketMs!==undefined&&![0,1000,10000,60000,300000].includes(q.bucketMs))throw Error('Invalid shared presentation.');
  return value;
 }
 function readQuestionLink(hash){if(hash.startsWith('#step=')){const s=readStepLink(hash);return s.kind==='record'?{version:1,question:s.record,view:s.view}:null}if(!hash.startsWith('#question='))return null;if(hash.length>16000)throw Error('Question link is too large.');return validateLink(JSON.parse(decodeURIComponent(hash.slice(10))));}
 function routeFromHash(hash){if(/^#(?:investigation|capture|document|annotation|evidence)=[a-f0-9]{64}$/.test(hash))return 'investigations';const key=hash.replace(/^#/,'');return Object.hasOwn(routes,key)?key:'explore';}
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
 const api={questionFromEvidence,captureQuestion,evidenceDifference,recordStep,measurementStep,validateStep,stepLink,readStepLink,sameWindow,routes,routeFromHash,questionLink,readQuestionLink,validateLink,mount};if(typeof module!=='undefined')module.exports=api;else global.GooWorkspace=api;
})(typeof window==='undefined'?globalThis:window);
