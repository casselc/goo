/* Service browsing composes existing typed record and native metric questions. */
(function(global){
 const clone=x=>structuredClone(x);
 function scopeQuestion(scope,service,signal='spans'){
  return {...clone(scope),signal,where:{field:{kind:'column',key:'service'},op:'eq',value:service},limit:100,offset:0};
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
   return {...f,available:!!field,field:field&&clone(field),question:field?{...scopeQuestion(scope,service),where:{all:[scopeQuestion(scope,service).where,{field:clone(field),op:'exists'}]},groupBy:[clone(field)],summary:'span-duration',spanStatus:true}:null};
  });
 }
 const node=(tag,text,cls)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n};
 function mount(host){
  let scope=null,inventory=null,selected=null,generation=0,controller=null,opener=null;
  const list=node('div',undefined,'service-browser-list'),detail=node('section',undefined,'service-browser-detail');detail.id='serviceDetail';detail.hidden=true;detail.setAttribute('aria-label','Service detail');host.insertBefore(list,host.querySelector('#inventoryRows'));host.insertBefore(detail,host.querySelector('#inventoryRows'));
  const inventoryParts=[host.querySelector('#inventoryRows'),host.querySelector('#inventoryExplain').closest('details'),host.querySelector('#serviceDiscoveryRun').closest('details')];
  const returnBar=node('div',undefined,'service-return controls');returnBar.id='serviceReturn';returnBar.hidden=true;
  document.querySelector('[data-workspace-panel="explore"]').prepend(returnBar);
  function invalidate(){inventoryParts.forEach(n=>n.hidden=false);generation++;controller?.abort();controller=null;selected=null;detail.hidden=true;detail.replaceChildren();returnBar.hidden=true;returnBar.replaceChildren()}
  function notify(action,extra={}){global.dispatchEvent(new CustomEvent('goo:service-browser-state',{detail:{action,generation,service:selected,scope:clone(scope),...extra}}))}
  function returnToService(){document.querySelector('[data-workspace="services"]').click();detail.scrollIntoView({block:'start'});detail.querySelector('h3')?.focus({preventScroll:true});notify('return')}
  function follow(question){returnBar.replaceChildren(node('span',selected+' · fixed service window'));const back=node('button','Back to service');back.onclick=returnToService;returnBar.append(back);returnBar.hidden=false;notify('question',{question:clone(question)});global.dispatchEvent(new CustomEvent('goo:open-question',{detail:question}))}
  function button(title,action){const b=node('button',title);b.onclick=action;return b}
  async function open(service,origin){
   invalidate();selected=service;opener=origin;inventoryParts.forEach(n=>n.hidden=true);detail.hidden=false;const g=generation;controller=new AbortController();const active=controller;
   const title=node('h3',service||'(empty service name)');title.tabIndex=-1;detail.append(button('Back to service list',()=>{invalidate();opener?.focus()}),title,node('p',scope.environment+' · ['+new Date(scope.fromMs).toISOString()+', '+new Date(scope.toMs).toISOString()+')','evidence-note'),node('p','Separate bounded reads of captured signals, span attributes and instrument metadata. Coverage and health remain unknown.'));
   const cancel=button('Cancel service reads',()=>{invalidate();notify('cancel')});detail.append(cancel);
   const sections=['Signals and operations','Suggested questions','Native instruments'].map(text=>{const section=node('section');section.append(node('h4',text),node('p','Reading…','service-loading'));detail.append(section);return section});
   detail.scrollIntoView({block:'start'});title.focus({preventScroll:true});notify('start');
   const requests=[['/api/discovery',scope],['/api/explore',scopeQuestion(scope,service)],['/api/metric-discovery',{...scope,service}]];
   const results=await Promise.allSettled(requests.map(async([url,q])=>{const r=await global.GooQueries.post(url,q,active.signal),data=await r.json();if(!r.ok)throw Error(data.error||'Source unavailable');return data}));
   if(g!==generation)return;cancel.remove();controller=null;sections.forEach(s=>s.querySelector('.service-loading').remove());
   sections[1].classList.add('service-question-grid');sections[2].classList.add('service-instrument-grid');
   const [signals,spans,metrics]=results;const error=(section,result)=>section.append(node('p','Unavailable · '+result.reason.message,'warning'));
   if(signals.status==='fulfilled'){
    const rows=signals.value.signals.filter(r=>r.service===service);
    for(const signal of ['spans','logs']){const row=rows.find(r=>r.signal===signal);sections[0].append(node('p',row?row.count+' captured '+signal:'No '+signal+' observed in this read; coverage unknown.'));if(row)sections[0].append(button('Explore service '+signal,()=>follow(scopeQuestion(scope,service,signal))))}
   }else error(sections[0],signals);
   const operations=inventory?.rows.filter(r=>r.service===service)||[];
   sections[0].append(node('p',inventory?operations.length+' operation names in the inventory read. Counts are span records.':'Operation inventory was not read for this window.'));
   for(const row of operations){const b=button('Explore operation '+(row.name||'(empty name)')+' · '+row.count,()=>follow(global.GooInventory.question(inventory,row)));sections[0].append(b)}
   if(spans.status==='fulfilled'){
    sections[1].append(button('Explore recorded span status',()=>follow({...scopeQuestion(scope,service),groupBy:[{kind:'column',key:'status'}],spanStatus:true})),node('p',spans.value.fieldsLimited?'Attribute discovery reached its limit; absent suggestions may have undiscovered fields.':'Suggestions use observed attribute keys; absent keys do not prove missing instrumentation.'));
    for(const suggestion of suggestions(scope,service,spans.value.fields)){
     const card=node('article',undefined,'definition-card');card.append(node('h4',suggestion.title),node('p',suggestion.note));
     const b=button(suggestion.available?'Explore '+suggestion.id:'No matching attributes observed',()=>follow(suggestion.question));b.disabled=!suggestion.available;card.append(b);sections[1].append(card)
    }
   }else error(sections[1],spans);
   if(metrics.status==='fulfilled'){
    const result=metrics.value;sections[2].append(node('p',result.warning));if(!result.instruments.length)sections[2].append(node('p','No native instruments observed in this window. Coverage unknown; resource utilization cannot be inferred from spans.'));
    for(const i of result.instruments){const card=node('article',undefined,'definition-card');card.append(node('h4',i.metric),node('p',i.kind+' · '+(i.unit||'unit omitted')+' · '+i.scope+' · version '+(i.version||'(empty)')+' · schema '+(i.schemaURL||'(empty)')+' · temporality '+i.temporality+' · monotonic '+i.monotonic),node('p','Observed dimensions: '+(i.dimensions||[]).join(', ')));
     const use=button(i.supported?'Preview measurement':'Unsupported by current kernels',()=>{returnBar.replaceChildren(node('span',service+' · fixed service window'),button('Back to service',returnToService));returnBar.hidden=false;notify('instrument',{question:{...scope,service}});global.dispatchEvent(new CustomEvent('goo:service-instrument',{detail:{definition:i.definition,...scope}}))});use.disabled=!i.supported;card.append(use);sections[2].append(card)}
   }else error(sections[2],metrics);
   const evidence=node('details');evidence.append(node('summary','Independent source reads and scope'));const pre=node('pre');pre.textContent=JSON.stringify(results.map((r,i)=>({source:requests[i][0],question:requests[i][1],status:r.status,...(r.status==='fulfilled'?{evidence:r.value}:{error:r.reason.message})})),null,2);evidence.append(pre);detail.append(evidence);notify('applied',{statuses:results.map(r=>r.status)})
  }
  function setInventory(value){invalidate();inventory=value;scope=value?clone(value.question):null;list.replaceChildren();if(!value)return;list.append(node('h3','Browse a service'),node('p','Open a service to read its signals, observed attributes and native instruments in this fixed window.'));
   for(const service of [...new Set(value.rows.map(r=>r.service))]){const b=button('Open service '+(service||'(empty name)'),()=>open(service,b));b.dataset.service=service;list.append(b)}
   search(host.querySelector('#inventorySearch').value);
  }
  function search(value){const text=value.toLowerCase();for(const b of list.querySelectorAll('button'))b.hidden=!inventory?.rows.some(r=>r.service===b.dataset.service&&(r.service+'\n'+r.name).toLowerCase().includes(text))}
  return {setInventory,search,open:(service,origin)=>{if(scope)return open(service,origin)},setScope:value=>{invalidate();list.replaceChildren();scope=clone(value);inventory=null},invalidate};
 }
 const api={scopeQuestion,suggestions,mount};if(typeof module!=='undefined')module.exports=api;else global.GooServiceBrowser=api;
})(typeof window==='undefined'?globalThis:window);
