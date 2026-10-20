/* Service browsing composes existing typed record and native metric questions. */
(function(global){
 const identity=typeof module!=='undefined'?require('./service-identity'):global.GooServiceIdentity;
 const clone=x=>structuredClone(x);
 function scopeQuestion(scope,service,signal='spans'){
  return {environment:scope.environment,fromMs:scope.fromMs,toMs:scope.toMs,signal,where:identity.serviceWhere(service,scope.namespace),limit:100,offset:0};
 }
 // Exact OTel keys, including legacy aliases, are chosen from source-discovered fields.
 // They are suggestions, not a claim that every span follows a convention.
 const families=[
  {id:'http',title:'HTTP route latency and recorded errors',keys:['http.route'],note:'Captured spans by route; includes mixed span kinds and retries.'},
  {id:'database',title:'Database latency and recorded errors',keys:['db.system.name','db.system'],note:'Captured database spans by reported system; not query throughput.'},
  {id:'messaging',title:'Messaging latency and recorded errors',keys:['messaging.destination.name','messaging.system'],note:'Captured messaging spans; not queue wait, delivery or end-to-end latency.'},
 ];
 function suggestions(scope,service,fields=[]){
  return families.map(f=>{const field=f.keys.map(key=>fields.find(x=>x.kind==='attribute'&&x.key===key)).find(Boolean);
   return {...f,available:!!field,field:field&&clone(field),question:field?{...scopeQuestion(scope,service),where:identity.and(scopeQuestion(scope,service).where,{field:clone(field),op:'exists'}),groupBy:[clone(field)],summary:'span-duration',spanStatus:true}:null};
  });
 }
 const node=(tag,text,cls)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n};
 function state(kind,text,cls){const e=node('goo-result-state',text,cls);e.setAttribute('state',kind);return e}
 function mount(host,{catalog}={}){
  let scope=null,inventory=null,selected=null,generation=0,controller=null,opener=null,selectedScope=null;
  const detail=node('section',undefined,'service-browser-detail');detail.id='serviceDetail';detail.hidden=true;detail.setAttribute('aria-label','Service detail');host.append(detail);
  const inventoryParts=[catalog];
  const returnBar=node('div',undefined,'service-return controls');returnBar.id='serviceReturn';returnBar.hidden=true;
  document.querySelector('[data-workspace-panel="explore"]').prepend(returnBar);
  function invalidate(){inventoryParts.forEach(n=>n.hidden=false);generation++;controller?.abort();controller=null;selected=null;detail.hidden=true;detail.replaceChildren();returnBar.hidden=true;returnBar.replaceChildren()}
  function notify(action,extra={}){global.dispatchEvent(new CustomEvent('goo:service-browser-state',{detail:{action,generation,service:selected,scope:clone(selectedScope||scope),...extra}}))}
  function returnToService(){document.querySelector('[data-workspace="services"]').click();detail.scrollIntoView({block:'start'});detail.querySelector('h3')?.focus({preventScroll:true});notify('return')}
  function follow(question){returnBar.replaceChildren(node('span',selected+' · fixed service window'));const back=node('button','Back to service');back.onclick=returnToService;returnBar.append(back);returnBar.hidden=false;notify('question',{question:clone(question)});global.dispatchEvent(new CustomEvent('goo:open-question',{detail:question}))}
  function button(title,action){const b=node('button',title);b.onclick=action;return b}
  async function open(service,origin,namespace){
   const activeScope={...clone(scope),...(namespace?{namespace:clone(namespace)}:{})};
   invalidate();selected=service;selectedScope=activeScope;opener=origin;inventoryParts.forEach(n=>n.hidden=true);detail.hidden=false;const g=generation;controller=new AbortController();const active=controller;
   const title=node('h3',service||'(empty service name)');title.tabIndex=-1;detail.append(button(origin?.closest('.distributed-view')?'Back to distributed view':'Back to service list',()=>{invalidate();if(opener?.isConnected)opener.focus();else (catalog.querySelector('.inventory-table tbody button')||document.getElementById('inventorySearch')).focus()}),title,node('p',activeScope.environment+' · '+identity.label(activeScope.namespace)+' · ['+new Date(activeScope.fromMs).toISOString()+', '+new Date(activeScope.toMs).toISOString()+')','evidence-note'),node('p','Separate bounded reads of captured signals, span attributes, operation inventory and native instruments. Coverage and health remain unknown.'));
   const cancel=button('Cancel service reads',()=>{invalidate();notify('cancel')});detail.append(cancel);
   const dashboard=node('section',undefined,'service-dashboard');dashboard.append(state('loading','Planning service overview…','service-loading'));detail.append(dashboard);
   const sections=['Signals and operations','Suggested questions','Native instruments'].map(text=>{const section=node('section');section.append(node('h4',text),state('loading','Reading…','service-loading'));detail.append(section);return section});
   detail.scrollIntoView({block:'start'});title.focus({preventScroll:true});notify('start');
   const requests=[['/api/discovery',activeScope],['/api/explore',scopeQuestion(activeScope,service)],['/api/metric-discovery',{...activeScope,service}],['/api/inventory',activeScope]];
   const results=await Promise.allSettled(requests.map(async([url,q])=>{const r=await global.GooQueries.post(url,q,active.signal),data=await r.json();if(!r.ok)throw Error(data.error||'Source unavailable');return data}));
   if(g!==generation)return;sections.forEach(s=>s.querySelector('.service-loading').remove());
   sections[1].classList.add('service-question-grid');sections[2].classList.add('service-instrument-grid');
   const [signals,spans,metrics,operationsRead]=results;inventory=operationsRead.status==='fulfilled'?operationsRead.value:null;const error=(section,result)=>section.append(state('unavailable','Unavailable · '+result.reason.message,'warning'));
   if(signals.status==='fulfilled'){
    const rows=signals.value.signals.filter(r=>r.service===service);
    for(const signal of ['spans','logs','metrics']){const matches=rows.filter(r=>r.signal===signal),row=matches.length?{count:matches.reduce((n,r)=>n+BigInt(r.count),0n).toString()}:null;sections[0].append(node('p',row?row.count+(signal==='metrics'?' captured metric data-point records':' captured '+signal):'No '+signal+' observed in this read; coverage unknown.'));if(signal!=='metrics')sections[0].append(button('Explore service '+signal,()=>follow(scopeQuestion(activeScope,service,signal))))}
   }else error(sections[0],signals);
   const operations=inventory?.rows.filter(r=>r.service===service)||[];
   const ops=node('div');ops.id='inventoryRows';sections[0].append(node('p',inventory?operations.length+' observed operation names for this service · '+operations.reduce((n,r)=>n+BigInt(r.count),0n)+' captured spans.':'Operation inventory unavailable · '+operationsRead.reason.message),ops);
   if(inventory){const search=node('input');search.id='operationSearch';search.type='search';search.placeholder='Find an observed operation';search.setAttribute('aria-label','Find an observed operation');sections[0].insertBefore(search,ops);const paint=()=>{ops.replaceChildren();const viewport=node('goo-table');viewport.setAttribute('density','compact');viewport.setAttribute('aria-label','Observed operations for '+service);const table=node('table',undefined,'inventory-table'),head=node('tr');['Observed operation','Captured spans','Recorded errors','Duration buckets','Investigate'].forEach(t=>{const th=node('th',t);th.scope='col';head.append(th)});table.append(node('caption','Observed operations for '+service));const thead=node('thead');thead.append(head);table.append(thead);const body=node('tbody');for(const row of operations.filter(r=>r.name.toLowerCase().includes(search.value.toLowerCase()))){const tr=node('tr');tr.dataset.count=row.count;tr.dataset.errors=row.errors;tr.dataset.buckets=JSON.stringify(row.buckets);tr.append(node('td',row.name||'(empty name)'),node('td',row.count),node('td',row.errors));const buckets=node('td'),counts=node('details');counts.append(node('summary','Duration bucket counts'));row.buckets.forEach((n,i)=>counts.append(node('p',inventory.bucketLabels[i]+': '+n)));buckets.append(counts);tr.append(buckets);const action=node('td'),open=button('Explore spans',()=>follow(global.GooInventory.question(inventory,row)));open.setAttribute('aria-label','Explore '+service+' / '+row.name);action.append(open);tr.append(action);[...tr.children].forEach((c,i)=>c.dataset.label=['Observed operation','Captured spans','Recorded errors','Duration buckets','Investigate'][i]);body.append(tr)}table.append(body);viewport.append(table);ops.append(viewport);if(!body.children.length)ops.append(state('empty',operations.length?'No local operation matches. Captured totals above are unchanged.':'No operations observed in this read; coverage unknown.'))};search.oninput=paint;paint();
    const capture=button('Capture scoped span inventory',()=>global.dispatchEvent(new CustomEvent('goo:save-capture',{detail:{question:{kind:'inventory',inventory:clone(inventory.question)},title:scope.environment+' '+identity.label(activeScope.namespace)+' span inventory'}})));capture.id='inventorySaveCapture';sections[0].append(capture,node('p','Inventory capture rereads service/operation pairs in this environment, selected namespace and fixed window, including other service names. It does not archive this catalog.','evidence-note'));const explain=node('details');explain.append(node('summary','Scoped operation inventory evidence'));const pre=node('pre');pre.id='inventoryExplain';pre.textContent=JSON.stringify(inventory,null,2);explain.append(pre);sections[0].append(explain);
   }
   if(spans.status==='fulfilled'){
    sections[1].append(button('Explore recorded span status',()=>follow({...scopeQuestion(activeScope,service),groupBy:[{kind:'column',key:'status'}],spanStatus:true})),node('p',spans.value.fieldsLimited?'Attribute discovery reached its limit; absent suggestions may have undiscovered fields.':'Suggestions use observed attribute keys; absent keys do not prove missing instrumentation.'));
    for(const suggestion of suggestions(activeScope,service,spans.value.fields)){
     const card=node('article',undefined,'definition-card');card.append(node('h4',suggestion.title),node('p',suggestion.note));
     const b=button(suggestion.available?'Explore '+suggestion.id:'No matching attributes observed',()=>follow(suggestion.question));b.disabled=!suggestion.available;card.append(b);sections[1].append(card)
    }
   }else error(sections[1],spans);
   if(metrics.status==='fulfilled'){
    const result=metrics.value;const discover=button('Discover metrics',()=>global.dispatchEvent(new CustomEvent('goo:discover-metrics',{detail:{...activeScope,service}})));discover.setAttribute('aria-label','Discover metrics for '+service);sections[2].append(discover);sections[2].append(node('p',result.warning));if(!result.instruments.length)sections[2].append(state('empty','No native instruments observed in this window. Coverage unknown; resource utilization cannot be inferred from spans.'));
    for(const i of result.instruments){const card=node('article',undefined,'definition-card');card.append(node('h4',i.metric),node('p',i.kind+' · '+(i.unit||'unit omitted')+' · '+i.scope+' · version '+(i.version||'(empty)')+' · schema '+(i.schemaURL||'(empty)')+' · temporality '+i.temporality+' · monotonic '+i.monotonic),node('p','Observed dimensions: '+(i.dimensions||[]).join(', ')));
     const use=button(i.supported?'Preview measurement':'Unsupported by current kernels',()=>{returnBar.replaceChildren(node('span',service+' · fixed service window'),button('Back to service',returnToService));returnBar.hidden=false;notify('instrument',{question:{...activeScope,service}});global.dispatchEvent(new CustomEvent('goo:service-instrument',{detail:{definition:i.definition,...activeScope}}))});use.disabled=!i.supported;card.append(use);sections[2].append(card)}
   }else error(sections[2],metrics);
   if(signals.status==='rejected'){dashboard.replaceChildren(node('h4','Service overview'),state('unavailable','Overview unavailable · service discovery failed; signal support cannot be inferred.'));dashboard.dataset.state='unavailable'}else await global.GooServiceDashboard.load(dashboard,{scope:clone(activeScope),service,signals:signals.status==='fulfilled'?[...new Set(signals.value.signals.filter(r=>r.service===service).map(r=>r.signal))]:[],fields:spans.status==='fulfilled'?spans.value.fields:[],instruments:metrics.status==='fulfilled'?metrics.value.instruments:[]},{signal:active.signal,current:()=>g===generation&&detail.isConnected});if(g!==generation)return;
   const evidence=node('details');evidence.className='service-source-evidence';evidence.append(node('summary','Independent source reads and scope'));const pre=node('pre');pre.textContent=JSON.stringify(results.map((r,i)=>({source:requests[i][0],question:requests[i][1],status:r.status,...(r.status==='fulfilled'?{evidence:r.value}:{error:r.reason.message})})),null,2);evidence.append(pre);detail.append(evidence);cancel.remove();controller=null;notify('applied',{statuses:results.map(r=>r.status)})
  }
  return {open:(service,origin,namespace)=>{if(scope)return open(service,origin,namespace)},setScope:value=>{invalidate();scope=value?clone(value):null;inventory=null},invalidate};
 }
 const api={scopeQuestion,suggestions,mount};if(typeof module!=='undefined')module.exports=api;else global.GooServiceBrowser=api;
})(typeof window==='undefined'?globalThis:window);
