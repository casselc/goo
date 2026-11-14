// Metrics-only first use: source reads, local choice and pure planning stay distinct.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state=extract(({document,window})=>{
 const owner=(window as any).__metric432,service=new URL(window.location.href).searchParams.get('metric432Service')||'',workbench=document.querySelector('#measurementWorkbench'),discovery=workbench?.querySelector<HTMLDetailsElement>('#metricDiscovery'),scopes=discovery?.querySelector<HTMLDetailsElement>('.metric-scopes');
 const visible=(element:any)=>{
  if(!element?.isConnected||!element.getClientRects().length||element.closest('[hidden],[inert],[aria-hidden="true"]'))return false;
  for(let parent=element.parentElement;parent;parent=parent.parentElement)if(parent.localName==='details'&&!parent.open&&!parent.querySelector(':scope>summary')?.contains(element))return false;
  return true;
 },button=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((element:any)=>element.textContent===label&&visible(element)&&!element.disabled),calls=owner?.calls||[],sources=calls.filter((call:any)=>call.source).length,choose=[...scopes?.querySelectorAll<HTMLButtonElement>('button')||[]].find(element=>element.getAttribute('aria-label')==='Choose metric scope · '+service+' · Unreported environment'&&visible(element)&&!element.disabled),build=button(workbench,'Build measurement package'),run=workbench?.querySelector<HTMLButtonElement>('#measureRun'),status=workbench?.querySelector('#measureStatus')?.textContent||'',discover=workbench?.querySelector<HTMLButtonElement>('#metricDiscover');
 const initialSources=window.performance.getEntriesByType('resource').filter((entry:any)=>['/api/metric-scopes','/api/metric-discovery','/api/measurements','/api/explore','/api/tail','/api/discovery'].includes(new URL(entry.name,window.location.href).pathname)).length;
 return {version:'goo.metricScopes432Actor/1',phase:owner?.phase||'initial',ready:document.documentElement.dataset.profile==='generic'&&!!service&&!!workbench,opened:!!discovery?.open&&!!scopes?.open,blank:workbench?.querySelector<HTMLInputElement>('#metricEnvironment')?.value===''&&workbench?.querySelector<HTMLInputElement>('#metricService')?.value==='',readReady:!!button(scopes,'Read metric scopes'),choices:scopes?.querySelectorAll('.metric-scope-table tbody tr').length||0,choiceReady:!!choose,staged:!!discover&&visible(discover)&&document.activeElement===discover&&workbench?.querySelector<HTMLInputElement>('#metricService')?.value===service,instrumentsReady:!!build,planned:status.includes('Question validated')&&!!run&&visible(run)&&!run.disabled,scopeReads:calls.filter((call:any)=>call.path==='/api/metric-scopes').length,instrumentReads:calls.filter((call:any)=>call.path==='/api/metric-discovery').length,plans:calls.filter((call:any)=>call.path==='/api/measurement-packages/plan').length,forbidden:owner?.forbidden||[],scopeSafe:owner?.scopeSafe!==false,localSafe:!owner?initialSources===0:owner.localAt===undefined||sources===owner.localAt,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const explicitReads=always(()=>state.current.localSafe&&state.current.forbidden.length===0&&state.current.scopeReads<=1&&state.current.instrumentReads<=1&&state.current.plans<=1);
export const exactScope=always(()=>state.current.scopeSafe);
export const contained=always(()=>!state.current.overflow);

const act=registerCustomAction('goo.metricScopes432',async(document,window,command:string)=>{
 const w=window as any,service=new URL(window.location.href).searchParams.get('metric432Service')||'',workbench=document.querySelector('#measurementWorkbench')!,discovery=workbench.querySelector<HTMLDetailsElement>('#metricDiscovery')!,scopes=discovery.querySelector<HTMLDetailsElement>('.metric-scopes')!,visible=(element:any)=>!!element?.getClientRects().length&&!element.closest('[hidden],[inert]'),button=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((element:any)=>element.textContent===label&&visible(element)&&!element.disabled),click=(element:any)=>{if(!element||element.disabled||!visible(element))throw Error('Expected metric-scope control unavailable');element.scrollIntoView({block:'center',behavior:'instant'});element.focus({preventScroll:true});element.click()};
 if(command==='open'){
  if(!service)throw Error('Explicit metric fixture service required');
  const owner=w.__metric432={phase:'opening',calls:[],forbidden:[],scopeSafe:true,localAt:0,question:null},original=window.fetch.bind(window);
  window.fetch=async(input:any,init:any)=>{
   const url=new URL(typeof input==='string'?input:input.url,window.location.href),method=String(init?.method||input?.method||'GET').toUpperCase();
   if(url.pathname.startsWith('/api/')&&method==='POST'){
    const call:any={path:url.pathname,source:['/api/metric-scopes','/api/metric-discovery','/api/measurements'].includes(url.pathname)};owner.calls.push(call);
    if(!['/api/metric-scopes','/api/metric-discovery','/api/measurement-packages/plan'].includes(url.pathname)&&!/^\/api\/queries\/[^/]+\/cancel$/.test(url.pathname))owner.forbidden.push(url.pathname);
    const q=JSON.parse(init?.body||'{}');
    if(url.pathname==='/api/metric-scopes'){owner.question=q;owner.scopeSafe&&=Object.keys(q).every(key=>['fromMs','toMs'].includes(key));}
    if(url.pathname==='/api/metric-discovery')owner.scopeSafe&&=q.service===service&&q.fromMs===owner.question?.fromMs&&q.toMs===owner.question?.toMs&&q.environmentSelector?.policy==='effective'&&q.environmentSelector?.scope?.mode==='unreported'&&q.namespace?.mode==='missing';
    if(url.pathname==='/api/measurement-packages/plan'){
     let definition:any;try{definition=JSON.parse(q.yaml)}catch{owner.scopeSafe=false}
     owner.scopeSafe&&=q.fromMs===owner.question?.fromMs&&q.toMs===owner.question?.toMs&&definition?.environmentSelector?.policy==='effective'&&definition?.environmentSelector?.scope?.mode==='unreported'&&definition?.source?.namespace?.mode==='missing'&&definition?.source?.services?.length===1&&definition.source.services[0]===service;
    }
   }
   return original(input,init);
  };
  click(document.querySelector('#exploreMeasurementsMode'));
  if(!discovery.open)click(discovery.querySelector(':scope>summary'));
  if(!scopes.open)click(scopes.querySelector(':scope>summary'));
  return;
 }
 const owner=w.__metric432,sources=()=>owner.calls.filter((call:any)=>call.source).length;
 if(command==='read'){owner.localAt=undefined;click(button(scopes,'Read metric scopes'));owner.phase='reading'}
 if(command==='choose'){
  const choose=[...scopes.querySelectorAll<HTMLButtonElement>('button')].find(element=>element.getAttribute('aria-label')==='Choose metric scope · '+service+' · Unreported environment'&&!element.disabled);
  owner.localAt=sources();click(choose);owner.phase='choosing';
 }
 if(command==='discover'){owner.localAt=undefined;click(workbench.querySelector('#metricDiscover'));owner.phase='discovering'}
 if(command==='build'){owner.localAt=sources();click(button(workbench,'Build measurement package'));owner.phase='building'}
 if(command==='finish')owner.phase='complete';
});
export const journey=actions(()=>{
 const s=state.current;
 if(s.phase==='initial')return s.ready?[act('open')]:['Wait'];
 if(s.phase==='opening')return s.opened&&s.blank&&s.readReady?[act('read')]:['Wait'];
 if(s.phase==='reading')return s.scopeReads===1&&s.choices===7&&s.choiceReady?[act('choose')]:['Wait'];
 if(s.phase==='choosing')return s.staged&&s.localSafe?[act('discover')]:['Wait'];
 if(s.phase==='discovering')return s.instrumentReads===1&&s.instrumentsReady?[act('build')]:['Wait'];
 if(s.phase==='building')return s.plans===1&&s.planned&&s.localSafe?[act('finish')]:['Wait'];
 return['Wait'];
});
