// Bounded discovery exploration: queries, draft/validate/preview/use; never save/export/workload.
import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';

const discovery = extract(({document}) => {
  const el=(id:string)=>document.getElementById(id);
  let malformed=false;
  const read=(id:string)=>{const s=el(id)?.textContent||'';if(!s)return null;try{return JSON.parse(s)}catch{malformed=true;return null}};
  const canonical=(value:any):string=>JSON.stringify(value,(key,v)=>v&&typeof v==='object'&&!Array.isArray(v)?Object.fromEntries(Object.entries(v).sort(([a],[b])=>a.localeCompare(b))):v);
  const definition=(q:any)=>canonical({signal:q?.signal,environment:q?.environment,where:q?.where||null,groupBy:q?.groupBy||[]});
  const packageActive=el('packageAuthoring')?.hidden===false&&el('packageEditor')?.hidden===false;
  const question=read('exploreExplain')?.question,context=read('surroundingExplain'),traceEvidence=read('exploreContextExplain'),plan=packageActive?read('packagePlan'):null,previewText=packageActive?(el('packagePreviewResult')?.querySelector('pre')?.textContent||''):'';
  let preview:any=null;try{if(previewText)preview=JSON.parse(previewText)}catch{malformed=true}
  let intent:any=null,draft:any=null;try{intent=JSON.parse(document.documentElement.dataset.gooDiscoveryIntent||'null');draft=JSON.parse(document.documentElement.dataset.gooDiscoveryDraft||'null')}catch{malformed=true}
  const rawCount=el('exploreCount')?.querySelector('span')?.textContent||'',groups=[...document.querySelectorAll<HTMLElement>('#distributionContent .distribution-row')];
  let population=true,buckets=true;
  if(groups.length){if(!/^\d+$/.test(rawCount))population=false;else{let sum=0n;for(const g of groups){const text=g.querySelector(':scope > span')?.textContent||'',match=text.match(/% · (\d+)$/);if(!match)population=false;else sum+=BigInt(match[1]);}population=population&&sum===BigInt(rawCount)}}
  if(question?.groupBy?.[0]?.key==='duration_bucket'){
    const labels=['[0,1) ms','[1,5) ms','[5,10) ms','[10,50) ms','[50,100) ms','[100,500) ms','[500,1000) ms','[1000,+inf) ms'];
    let previous=-1;for(const row of groups){try{const value=JSON.parse(row.querySelector('.distribution-label')?.textContent||'null'),index=labels.indexOf(value);if(index<=previous||index<0)buckets=false;previous=index}catch{buckets=false}}
  }
  let contextBounded=true,anchorProvenance=true;
  if(context){try{const {origin,anchor,anchorQuestion,anchorBasis,context:c}=context,q=c.question,ms=Number(BigInt(anchor.timestamp_ns)/1000000n),radius=Number((el('surroundingRadius') as HTMLSelectElement).value),p=q.where;
    contextBounded=q.signal==='logs'&&q.environment===origin.environment&&q.fromMs===Math.max(origin.fromMs,ms-radius)&&q.toMs===Math.min(origin.toMs,ms+radius+1)&&q.limit===100&&q.offset>=0&&q.offset<=1000&&!(q.groupBy?.length)&&p?.field?.kind==='column'&&p.field.key==='service'&&p.op==='eq'&&(p.value||'')===anchor.ServiceName&&!p.all&&!p.any;
    anchorProvenance=typeof anchorBasis==='string'&&anchorBasis.length>0&&anchorQuestion?.environment===origin.environment&&anchorQuestion.fromMs<=ms&&anchorQuestion.toMs>ms&&typeof context.originBasis==='string'&&context.originBasis.length>0;
    const rows=document.querySelectorAll('#surroundingRows .record-row').length;if(rows>100)contextBounded=false;
  }catch{contextBounded=false;anchorProvenance=false}}
  let quickAnd=true;
  if(question&&intent){quickAnd=question.signal===intent.signal;for(const p of intent.predicates)if(!(question.where?.all||[]).some((a:any)=>canonical(a)===canonical(p)))quickAnd=false;}
  const packagePreserved=!plan||!draft||definition(plan.question)===definition(draft);
  const previewSeparate=!preview||!!plan&&canonical(preview.question)===canonical(plan.question)&&typeof preview.basis==='string'&&preview.basis.length>0;
  const reused=document.documentElement.dataset.gooDiscoveryReuse==='true';
  const reusePreserved=!reused||!question||!draft||definition(question)===definition(draft);
  const visible=(id:string)=>el(id)?.hidden===false;
  const packageInvalidated=!visible('packageAuthoring')||!!plan||['packagePreview','packageSave','packageUse'].every(id=>(el(id) as HTMLButtonElement)?.disabled);
  return {version:'goo.discovery.v1',stage:document.documentElement.dataset.gooDiscoveryStage||'initial',ready:!!el('exploreRun'),busy:(el('exploreRun') as HTMLButtonElement)?.disabled,
    question,hasRecords:!!document.querySelector('#exploreRows .record-row'),population,buckets,contextBounded,anchorProvenance,quickAnd,packagePreserved,previewSeparate,reusePreserved,packageInvalidated,malformed,
    trace:visible('exploreContext'),nearby:visible('surroundingContext'),traceReady:visible('exploreContext')&&!el('exploreContextSummary')?.textContent?.startsWith('Reading')&&!!document.querySelector('#exploreTrace .waterfall-row'),contextReady:visible('surroundingContext')&&!!context,
    selected:document.querySelectorAll('#exploreTrace .waterfall-row[aria-pressed="true"]').length,traceMore:!(el('exploreTraceMore') as HTMLButtonElement)?.disabled,nearbyMore:!(el('surroundingNext') as HTMLButtonElement)?.disabled,
    draftReady:visible('packageDraftSetup')&&visible('packageEditor'),planReady:!!plan,previewReady:!!preview,route:document.querySelector('[data-workspace][aria-current="page"]')?.getAttribute('data-workspace'),
    timeContext:context,traceQuestion:traceEvidence?.trace?.question||null,packageQuestion:plan?.question||null,recordCount:rawCount,rowCount:document.querySelectorAll('#exploreRows .record-row').length,contextQuestion:context?.context?.question||null,contextCount:context?.context?.count||null};
});
export const ready = eventually(()=>discovery.current.ready).within(10,'seconds');
export const prefixCompletes = eventually(()=>discovery.current.stage==='complete').within(50,'seconds');
export const coherentPopulation = always(()=>discovery.current.population);
export const orderedDurationPartition = always(()=>discovery.current.buckets);
export const boundedTimeContext = always(()=>discovery.current.contextBounded&&discovery.current.anchorProvenance);
export const explicitAndSearch = always(()=>discovery.current.quickAnd);
export const reusableDefinition = always(()=>discovery.current.packagePreserved&&discovery.current.reusePreserved);
export const independentPreview = always(()=>discovery.current.previewSeparate);
export const invalidDraftDisablesActions = always(()=>discovery.current.packageInvalidated);
export const oneInspector = always(()=>!(discovery.current.trace&&discovery.current.nearby)&&discovery.current.selected<=1);
export const parseableEvidence = always(()=>!discovery.current.malformed);

const act=registerCustomAction('goo.discovery',async(document,window,command:string,next?:string)=>{
  const el=(id:string)=>document.getElementById(id),click=(id:string)=>el(id)?.click();
  const set=(id:string,value:string)=>{const n=el(id) as HTMLInputElement|HTMLSelectElement;if(n){n.value=value;n.dispatchEvent(new Event('input',{bubbles:true}));n.dispatchEvent(new Event('change',{bubbles:true}))}};
  const stage=()=>{if(next)document.documentElement.dataset.gooDiscoveryStage=next};
  const clearIntent=()=>{delete document.documentElement.dataset.gooDiscoveryIntent;delete document.documentElement.dataset.gooDiscoveryReuse};
  const read=(id:string)=>{try{return JSON.parse(el(id)?.textContent||'null')}catch{return null}};
  if(command==='start'){const env=new URLSearchParams(window.location.hash.slice(1)).get('environment');if(env)set('exploreEnvironment',env);set('minutes','5');set('exploreSignal','spans');click('exploreRun');}
  if(command==='query'){clearIntent();click('exploreRun')}
  if(command==='reset'){clearIntent();click('exploreClear');set('discoveryService','');set('discoveryMessage','');set('discoverySeverity','');set('discoveryDuration','');(el('discoveryErrors') as HTMLInputElement).checked=false;set('exploreGroup','');click('exploreRun')}
  if(command==='logs'||command==='spans'){clearIntent();set('exploreSignal',command);click('exploreRun')}
  if(command==='quick'){
    const signal=(el('exploreSignal') as HTMLSelectElement).value,raw=document.querySelector('#exploreRows .raw-record pre')?.textContent||'{}';let row:any={};try{row=JSON.parse(raw)}catch{}
    clearIntent();click('exploreClear');set('exploreJoin','any');set('discoveryService',signal==='logs'?row.ServiceName||'':'');set('discoveryMessage',signal==='logs'?(row.Body||'').slice(0,24):'');set('discoverySeverity',signal==='logs'?row.SeverityText||'':'');set('discoveryDuration',signal==='spans'?'0':'');(el('discoveryErrors') as HTMLInputElement).checked=false;
    const predicates:any[]=[],add=(key:string,op:string,value:string)=>predicates.push({field:{kind:'column',key},op,value});
    if(signal==='spans')add('duration_ns','gte','0');else{if(row.ServiceName)add('service','eq',row.ServiceName);if(row.Body)add('body','contains',row.Body.slice(0,24));if(row.SeverityText)add('severity','eq',row.SeverityText)}
    document.documentElement.dataset.gooDiscoveryIntent=JSON.stringify({signal,predicates});click('exploreRun');
  }
  if(command==='traces'||command==='buckets'){clearIntent();const name=command==='traces'?'Find matching traces':'Span latency distribution';[...document.querySelectorAll<HTMLButtonElement>('#discoverySuggestions button')].find(b=>b.textContent===name)?.click()}
  if(command==='trace'){document.querySelector<HTMLButtonElement>('#distributionContent button[aria-label^="Open matching trace"]:not(:disabled)')?.click()}
  if(command==='tracePage')click('exploreTraceMore');if(command==='select')document.querySelector<HTMLButtonElement>('#exploreTrace .waterfall-row')?.click();if(command==='closeTrace')click('exploreContextClose');
  if(command==='nearby')document.querySelector<HTMLButtonElement>('#exploreRows button[aria-label^="Surrounding logs for"]')?.click();
  if(command==='nearbyPage')click('surroundingNext');if(command==='radius')set('surroundingRadius',(el('surroundingRadius') as HTMLSelectElement).value==='30000'?'60000':'30000');
  if(command==='reanchor'){const anchor=read('surroundingExplain')?.anchor;const rows=[...document.querySelectorAll<HTMLElement>('#surroundingRows .record-row')];const row=rows.find(r=>r.textContent?.includes('No valid reported trace ID')&&!r.classList.contains('log-anchor'))||rows.find(r=>!r.classList.contains('log-anchor'));row?.querySelector<HTMLButtonElement>('button[aria-label^="Surrounding logs for"]')?.click()}
  if(command==='closeNearby')click('surroundingClose');
  if(command==='draft'){clearIntent();document.documentElement.dataset.gooDiscoveryDraft=JSON.stringify(read('exploreExplain')?.question);click('exploreCreatePackage')}
  if(command==='build'){set('packageName','bombadil-discovery-preview');set('packageTitle','Discovery exploratory preview');click('packageBuild')}
  if(command==='preview')click('packagePreview');
  if(command==='use'){clearIntent();document.documentElement.dataset.gooDiscoveryReuse='true';click('packageUse')}
  if(command==='edit'){const input=el('packageYAML') as HTMLTextAreaElement;input.value+='\n# exploratory draft edit\n';input.dispatchEvent(new Event('input',{bubbles:true}))}
  if(command==='validate')click('packageValidate');
  if(command==='explore'){clearIntent();document.querySelector<HTMLButtonElement>('[data-workspace="explore"]')?.click()}
  stage();
});
export const exploreDiscovery = actions(()=>{
  const s=discovery.current;if(!s.ready||s.busy)return ['Wait'];
  // Prefix guarantees the new workflows are reached before free exploration.
  switch(s.stage){
    case 'initial':return [act('start','loaded')];
    case 'loaded':return s.question?[act('quick','searched')]:['Wait'];
    case 'searched':return s.question?[act('traces','trace_groups')]:['Wait'];
    case 'trace_groups':return s.question?.groupBy?.[0]?.key==='trace_id'?[act('trace','trace_open')]:['Wait'];
    case 'trace_open':return s.traceReady?[s.traceMore?act('tracePage','trace_paged'):act('closeTrace','trace_closed')]:['Wait'];
    case 'trace_paged':return s.traceReady?[act('closeTrace','trace_closed')]:['Wait'];
    case 'trace_closed':return [act('buckets','buckets')];
    case 'buckets':return s.question?.groupBy?.[0]?.key==='duration_bucket'?[act('draft','draft')]:['Wait'];
    case 'draft':return s.draftReady?[act('build','built')]:['Wait'];
    case 'built':return s.planReady?[act('preview','previewed')]:['Wait'];
    case 'previewed':return s.previewReady?[act('use','reused')]:['Wait'];
    case 'reused':return s.route==='explore'&&s.question?[act('logs','logs_loaded')]:['Wait'];
    case 'logs_loaded':return s.question?.signal==='logs'?[act('quick','logs_searched')]:['Wait'];
    case 'logs_searched':return s.question?.signal==='logs'?[act('nearby','around')]:['Wait'];
    case 'around':return s.contextReady?[s.nearbyMore?act('nearbyPage','around_paged'):act('radius','radius')]:['Wait'];
    case 'around_paged':return s.contextReady?[act('radius','radius')]:['Wait'];
    case 'radius':return s.contextReady?[act('reanchor','reanchored')]:['Wait'];
    case 'reanchored':return s.contextReady?[act('closeNearby','complete')]:['Wait'];
  }
  const choices=[act('query'),act('reset'),act('logs'),act('spans')];
  if(s.route==='definitions'){choices.splice(0,choices.length);if(s.draftReady)return[act('build'),act('explore')];if(s.planReady)choices.push(act('preview'),act('edit'),act('use'));else choices.push(act('validate'));choices.push(act('explore'));return choices}
  if(s.question){choices.push(act('quick'));if(s.question.signal==='spans')choices.push(act('buckets'),act('traces'));choices.push(act('draft'))}
  if(s.question?.groupBy?.[0]?.key==='trace_id')choices.push(act('trace'));
  if(s.question?.signal==='logs'&&s.hasRecords)choices.push(act('nearby'));
  if(s.traceReady)choices.push(act('select'),act('closeTrace'),...(s.traceMore?[act('tracePage')]:[]));
  if(s.contextReady)choices.push(act('radius'),act('reanchor'),act('closeNearby'),...(s.nearbyMore?[act('nearbyPage')]:[]));
  return choices;
});
