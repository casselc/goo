// Positive partial-source first-use path. Presence witnesses are never totals.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state=extract(({document,window})=>{
 const owner=(window as any).__source431,params=new URL(window.location.href).searchParams,kind=params.get('source431Kind'),environment=params.get('source431Env')||'',shell=document.querySelector('.stream-entry'),box=shell?.querySelector<HTMLDetailsElement>('.environment-adoption'),tail=shell?.querySelector('.live-tail');
 const visible=(element:any)=>{
  if(!element?.isConnected||!element.getClientRects().length||element.closest('[hidden],[inert],[aria-hidden="true"]'))return false;
  for(let parent=element.parentElement;parent;parent=parent.parentElement)if(parent.localName==='details'&&!parent.open&&!parent.querySelector(':scope>summary')?.contains(element))return false;
  return true;
 },button=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((element:any)=>element.textContent===label&&visible(element)&&!element.disabled),calls=owner?.calls||[],sources=calls.filter((call:any)=>call.source).length,signal=shell?.querySelector<HTMLSelectElement>('[aria-label="Live signal"]'),rows=tail?.querySelectorAll('tbody tr').length||0,phase=tail?.getAttribute('data-state');
 const caps=[...box?.querySelectorAll('.source-capability-table tbody tr')||[]],spans=caps.find(row=>row.querySelector('th')?.textContent==='Spans'),capReady=caps.length===6&&spans?.querySelector('[data-state="compatible"]')?.textContent==='Records observed'&&caps.filter(row=>row!==spans).every(row=>row.querySelector('[data-state="absent"]')?.textContent==='Table absent')&&!button(box,'Choose logs')&&!!button(box,'Choose spans'),previewRow=[...box?.querySelectorAll('.environment-label-table tbody tr')||[]].find(row=>row.querySelector('th[scope="row"]')?.textContent==='Named environment: '+JSON.stringify(environment)),preview=previewRow?.querySelector('button'),start=button(shell,'Start stream'),resume=button(shell,'Resume with recent spans');
 const initialSources=window.performance.getEntriesByType('resource').filter((entry:any)=>['/api/source-capabilities','/api/environment-scopes','/api/environment-scopes/records','/api/tail','/api/explore','/api/metric-discovery'].includes(new URL(entry.name,window.location.href).pathname)).length;
 return {version:'goo.sourceCapabilities431Actor/1',phase:owner?.phase||'initial',ready:document.documentElement.dataset.profile==='generic'&&kind==='spans'&&!!environment&&!!box,opened:!!box?.open,capReadReady:!!button(box,'Read available signals'),capReady,chosen:signal?.value==='spans'&&document.activeElement===button(box,'Read environment scopes'),previewReady:!!preview&&visible(preview)&&!preview.disabled,stageReady:!!button(box,'Stage in Activity'),staged:phase==='paused'&&rows===0&&!!start&&document.activeElement===start,streaming:rows>0&&['unknown','partial'].includes(phase||''),paused:phase==='paused'&&!!resume&&document.activeElement===resume,rows,capabilityReads:calls.filter((call:any)=>call.path==='/api/source-capabilities').length,environmentReads:calls.filter((call:any)=>call.path==='/api/environment-scopes').length,previewReads:calls.filter((call:any)=>call.path==='/api/environment-scopes/records').length,tailReads:calls.filter((call:any)=>call.path==='/api/tail').length,forbidden:owner?.forbidden||[],scopeSafe:owner?.scopeSafe!==false,focusSafe:owner?.focusSafe!==false,capabilityObserved:owner?.capabilityObserved===true,localSafe:!owner?initialSources===0:owner.localAt===undefined||sources===owner.localAt,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const readBudgets=always(()=>state.current.capabilityReads<=1&&state.current.environmentReads<=1&&state.current.previewReads<=1&&state.current.tailReads<=1&&state.current.forbidden.length===0&&state.current.localSafe);
export const exactScopeAndFocus=always(()=>state.current.scopeSafe&&state.current.focusSafe);
export const narrowContainment=always(()=>!state.current.overflow);

const act=registerCustomAction('goo.sourceCapabilities431',async(document,window,command:string)=>{
 const w=window as any,shell=document.querySelector('.stream-entry')!,box=shell.querySelector<HTMLDetailsElement>('.environment-adoption')!,tail=shell.querySelector('.live-tail'),params=new URL(window.location.href).searchParams,kind=params.get('source431Kind'),environment=params.get('source431Env'),visible=(element:any)=>!!element?.getClientRects().length&&!element.closest('[hidden],[inert]'),button=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((element:any)=>element.textContent===label&&visible(element)&&!element.disabled),click=(element:any)=>{if(!element||element.disabled||!visible(element))throw Error('Expected source capability control unavailable');element.scrollIntoView({block:'center',behavior:'instant'});element.focus({preventScroll:true});element.click()};
 if(command==='open'){
  if(kind!=='spans'||!environment)throw Error('Explicit spans-only fixture identity required');
  const owner=w.__source431={phase:'opening',calls:[],forbidden:[],scopeSafe:true,focusSafe:true,capabilityObserved:false,localAt:0},original=window.fetch.bind(window);
  window.fetch=async(input:any,init:any)=>{
   const url=new URL(typeof input==='string'?input:input.url,window.location.href),method=String(init?.method||input?.method||'GET').toUpperCase();
   if(url.pathname.startsWith('/api/')&&method==='POST'){
    const call:any={path:url.pathname,source:['/api/source-capabilities','/api/environment-scopes','/api/environment-scopes/records','/api/tail'].includes(url.pathname)};owner.calls.push(call);
    if(!call.source&&!/^\/api\/queries\/[^/]+\/cancel$/.test(url.pathname))owner.forbidden.push(url.pathname);
    if(call.source){const q=JSON.parse(init?.body||'{}');if(url.pathname==='/api/environment-scopes')owner.scopeSafe&&=q.signal==='spans';else if(url.pathname==='/api/environment-scopes/records')owner.scopeSafe&&=q.signal==='spans'&&q.scope?.mode==='named'&&q.scope?.value===environment;else if(url.pathname==='/api/tail')owner.scopeSafe&&=q.signal==='spans'&&q.environmentSelector?.policy==='effective'&&q.environmentSelector?.scope?.mode==='named'&&q.environmentSelector?.scope?.value===environment;}
   }
   return original(input,init);
  };
  if(!box.open)click(box.querySelector(':scope>summary'));return;
 }
 const owner=w.__source431,sources=()=>owner.calls.filter((call:any)=>call.source).length;
 if(command==='capabilities'){owner.localAt=undefined;click(button(box,'Read available signals'));owner.phase='checking'}
 if(command==='choose'){owner.capabilityObserved=true;owner.localAt=sources();click(button(box,'Choose spans'));owner.focusSafe&&=document.activeElement===button(box,'Read environment scopes');owner.phase='choosing'}
 if(command==='environments'){owner.localAt=undefined;click(button(box,'Read environment scopes'));owner.phase='discovering'}
 if(command==='preview'){const row=[...box.querySelectorAll('.environment-label-table tbody tr')].find(row=>row.querySelector('th[scope="row"]')?.textContent==='Named environment: '+JSON.stringify(environment));click(row?.querySelector('button'));owner.phase='previewing'}
 if(command==='stage'){owner.localAt=sources();click(button(box,'Stage in Activity'));owner.phase='staging'}
 if(command==='start'){owner.localAt=undefined;click(button(shell,'Start stream'));owner.phase='streaming'}
 if(command==='pause'){owner.localAt=sources();click(button(shell,'Pause stream'));owner.focusSafe&&=document.activeElement===button(shell,'Resume with recent spans');owner.phase='pausing'}
 if(command==='finish')owner.phase='complete';
});
export const journey=actions(()=>{
 const s=state.current;
 if(s.phase==='initial')return s.ready?[act('open')]:['Wait'];
 if(s.phase==='opening')return s.opened&&s.capReadReady?[act('capabilities')]:['Wait'];
 if(s.phase==='checking')return s.capabilityReads===1&&s.capReady?[act('choose')]:['Wait'];
 if(s.phase==='choosing')return s.chosen&&s.localSafe?[act('environments')]:['Wait'];
 if(s.phase==='discovering')return s.environmentReads===1&&s.previewReady?[act('preview')]:['Wait'];
 if(s.phase==='previewing')return s.previewReads===1&&s.stageReady?[act('stage')]:['Wait'];
 if(s.phase==='staging')return s.staged?[act('start')]:['Wait'];
 if(s.phase==='streaming')return s.streaming&&s.tailReads===1?[act('pause')]:['Wait'];
 if(s.phase==='pausing')return s.paused&&s.capabilityObserved?[act('finish')]:['Wait'];
 return['Wait'];
});
