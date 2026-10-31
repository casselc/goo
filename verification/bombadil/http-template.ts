import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const params=new URL(window.location.href).searchParams,environment=params.get('templateEnvironment'),service=params.get('templateService'),namespace=params.get('templateNamespace');
 const chart=document.querySelector<any>('.population-breakdown goo-chart'),q=chart?.snapshot?.scope?.selector;
 const leaves=(p:any):any[]=>!p?[]:Array.isArray(p.all)?p.all.flatMap(leaves):p.any?[]:[p];
 const predicates=leaves(q?.where),has=(kind:string,key:string,op:string,value:string)=>predicates.some(p=>p.field?.kind===kind&&p.field.key===key&&p.op===op&&(p.value||'')===value);
 const sourcePaths=['/api/discovery','/api/inventory','/api/metric-discovery','/api/explore','/api/record-aggregates','/api/measurements'];
 return {inventoryReady:!!document.querySelector<HTMLButtonElement>('#inventoryRun')?.onclick,version:'goo.httpTemplate278',phase:document.documentElement.dataset.templatePhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),reads:window.performance.getEntriesByType('resource').filter(e=>sourcePaths.includes(new URL(e.name).pathname)).length,base:document.documentElement.dataset.templateBase===undefined?null:Number(document.documentElement.dataset.templateBase),expansions:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/view-templates/expand').length,serviceReady:document.querySelector('.service-dashboard')?.getAttribute('data-state')==='loaded',serviceAvailable:[...document.querySelectorAll('button')].some(b=>(b.getAttribute('aria-label')||b.textContent)==='Open service '+namespace+' / '+service),customize:!!document.querySelector('button[aria-label="Customize HTTP route view"]'),value:chart?.display?.value,scale:chart?.display?.scale,total:chart?.snapshot?.partition?.total,scopeOK:!q?null:q.environment===environment&&q.signal==='spans'&&q.groupBy?.length===1&&q.groupBy[0].kind==='attribute'&&q.groupBy[0].key==='http.route'&&has('column','service','eq',service!)&&has('resource','service.namespace','eq',namespace!)&&has('column','kind','eq','Server')&&has('attribute','http.request.method','ne',''),saved:!!document.querySelector('.population-save-status a'),overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const explicitReads=always(()=>state.current.base===null||state.current.reads<=state.current.base+(['customize'].includes(state.current.phase)?0:1));
export const customizeIsConfigurationOnly=always(()=>state.current.phase!=='customize'||state.current.reads===state.current.base);
export const checkedScope=always(()=>state.current.scopeOK===null||state.current.scopeOK===true&&state.current.total==='5');
export const durationScale=always(()=>!state.current.value||state.current.value==='count'||state.current.scale==='max');
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.httpTemplate',async(document,window,phase:string)=>{
 const button=(name:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 const choose=(value:string)=>{const s=document.querySelector<HTMLSelectElement>('select[aria-label="Breakdown measure"]')!;s.value=value;s.dispatchEvent(new Event('change'))};
 const params=new URL(window.location.href).searchParams,sourcePaths=['/api/discovery','/api/inventory','/api/metric-discovery','/api/explore','/api/record-aggregates','/api/measurements'],reads=()=>window.performance.getEntriesByType('resource').filter(e=>sourcePaths.includes(new URL(e.name).pathname)).length;
 if(phase==='initial'){
  const env=document.querySelector<HTMLInputElement>('#inventoryEnvironment')!,minutes=document.querySelector<HTMLInputElement>('#minutes')!;env.value=params.get('templateEnvironment')!;env.dispatchEvent(new Event('input'));minutes.value='2';minutes.dispatchEvent(new Event('input'));document.querySelector<HTMLButtonElement>('#inventoryRun')!.click();
 }else if(phase==='inventory'){
  const name='Open service '+params.get('templateNamespace')+' / '+params.get('templateService');[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>(b.getAttribute('aria-label')||b.textContent)===name)!.click();
 }else if(phase==='service'){
  document.documentElement.dataset.templateBase=String(reads());document.querySelector<HTMLButtonElement>('button[aria-label="Customize HTTP route view"]')!.click();
 }else if(phase==='customize'){
  if(reads()!==Number(document.documentElement.dataset.templateBase))throw Error('Customize read telemetry');button('Run population view').click();
 }else if(phase==='run'){
  if(reads()!==Number(document.documentElement.dataset.templateBase)+1)throw Error('Run did not acquire exactly one population');choose('count');
 }else if(phase==='count')choose('duration.p95');
 else if(phase==='duration'){
  document.querySelectorAll<HTMLDetailsElement>('.population-save').forEach(d=>d.open=true);const name=document.querySelector<HTMLInputElement>('input[aria-label="Exploration view title"]')!;name.value='Bombadil HTTP route display';name.dispatchEvent(new Event('input'));button('Save exploration view').click();
 }else if(phase==='save')window.location.hash=document.querySelector<HTMLAnchorElement>('.population-save-status a')!.hash;
 else if(phase==='reopen'){if(reads()!==Number(document.documentElement.dataset.templateBase)+1)throw Error('Configuration reopen read telemetry')}
 document.documentElement.dataset.templatePhase=({initial:'inventory',inventory:'service',service:'customize',customize:'run',run:'count',count:'duration',duration:'save',save:'reopen',reopen:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.inventoryReady?[step('initial')]:s.phase==='inventory'&&s.serviceAvailable?[step('inventory')]:s.phase==='service'&&s.serviceReady&&s.customize?[step('service')]:s.phase==='customize'&&s.population==='staged'?[step('customize')]:s.phase==='run'&&s.population==='applied'&&s.scopeOK===true?[step('run')]:['count','duration'].includes(s.phase)?[step(s.phase)]:s.phase==='save'&&s.saved?[step('save')]:s.phase==='reopen'&&s.population==='staged'?[step('reopen')]:['Wait']});
