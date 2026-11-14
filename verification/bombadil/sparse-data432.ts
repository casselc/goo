// Empty retained Activity observations are not a population census.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const owner=(window as any).__sparse432,params=new URL(window.location.href).searchParams,environment=params.get('sparse432Env')||'',bank=params.get('sparse432Bank')||'',shell=document.querySelector('.stream-entry'),tail=shell?.querySelector<HTMLElement>('.live-tail'),recovery=shell?.querySelector<HTMLElement>('.activity-recovery');
 const visible=(e:any)=>{if(!e?.isConnected||!e.getClientRects().length||e.closest('[hidden],[inert],[aria-hidden="true"]'))return false;for(let p=e.parentElement;p;p=p.parentElement)if(p.localName==='details'&&!p.open&&!p.querySelector(':scope>summary')?.contains(e))return false;return true;},button=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((e:any)=>e.textContent===label&&visible(e)&&!e.disabled),fixed=[...document.querySelectorAll<HTMLElement>('.activity-analysis')].find(visible),run=fixed?.querySelector<HTMLButtonElement>('#exploreRun'),phase=tail?.dataset.state,calls=owner?.calls||[],sources=calls.filter((call:any)=>call.source).length,emptyReady=!!button(shell,'Search wider window')&&Number(tail?.dataset.retained)===0&&recovery?.dataset.pollRows==='0'&&recovery?.dataset.limited==='false'&&Number.isFinite(Date.parse(recovery?.dataset.observedAt||'')),staged=!!fixed&&fixed.dataset.questionPhase==='draft'&&!!run&&visible(run)&&document.activeElement===run,applied=!!fixed&&fixed.dataset.questionPhase==='applied'&&owner?.runCount==='1'&&(fixed.querySelector('#exploreRows')?.textContent||'').includes(bank),returned=visible(shell)&&phase==='paused'&&emptyReady&&document.activeElement===button(shell,'Search wider window')&&recovery?.dataset.toMs===String(owner?.expected?.toMs);
 const initialSources=window.performance.getEntriesByType('resource').filter((entry:any)=>['/api/tail','/api/explore','/api/explore/page','/api/record-aggregates','/api/measurements'].includes(new URL(entry.name,window.location.href).pathname)).length;
 return {version:'goo.sparseRecovery432Actor/1',phase:owner?.phase||'initial',ready:document.documentElement.dataset.profile==='generic'&&!!environment&&!!bank&&!!button(shell,'Start stream'),emptyReady,staged,applied,returned,resumed:visible(shell)&&phase==='unknown'&&emptyReady&&calls.filter((call:any)=>call.path==='/api/tail').length===2,tailReads:calls.filter((call:any)=>call.path==='/api/tail').length,exploreReads:calls.filter((call:any)=>call.path==='/api/explore').length,localSafe:!owner?initialSources===0:owner.localAt===undefined||owner.localAt===sources,scopeSafe:owner?.scopeSafe!==false,forbidden:owner?.forbidden||[],overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const explicitReads=always(()=>state.current.localSafe&&state.current.forbidden.length===0&&state.current.tailReads<=2&&state.current.exploreReads<=1);
export const exactWindowAndScope=always(()=>state.current.scopeSafe);
export const containment=always(()=>!state.current.overflow);
const act=registerCustomAction('goo.sparseRecovery432',async(document,window,command:string)=>{
 const w=window as any,params=new URL(window.location.href).searchParams,environment=params.get('sparse432Env'),bank=params.get('sparse432Bank'),shell=document.querySelector('.stream-entry')!,visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert]'),button=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((e:any)=>e.textContent===label&&visible(e)&&!e.disabled),click=(e:any)=>{if(!e||!visible(e)||e.disabled)throw Error('Expected sparse-recovery control unavailable');e.scrollIntoView({block:'center',behavior:'instant'});e.focus({preventScroll:true});e.click()};
 if(command==='start'){
  if(!environment||!bank)throw Error('Sparse fixture identity required');
  const owner=w.__sparse432={phase:'reading',calls:[],forbidden:[],localAt:undefined,scopeSafe:true,expected:null,runCount:null},original=window.fetch.bind(window);
  window.addEventListener('goo:question-state',event=>{const detail=(event as CustomEvent).detail;if(detail.action==='stage')owner.staged=structuredClone(detail.requested)});
  window.fetch=async(input:any,init:any)=>{
   const url=new URL(typeof input==='string'?input:input.url,window.location.href),method=String(init?.method||input?.method||'GET').toUpperCase();
   if(url.pathname.startsWith('/api/')&&method==='POST'){
    const q=JSON.parse(init?.body||'{}'),source=['/api/tail','/api/explore'].includes(url.pathname);owner.calls.push({path:url.pathname,source});
    if(!source&&!/^\/api\/queries\/[^/]+\/cancel$/.test(url.pathname))owner.forbidden.push(url.pathname);
    if(source){owner.scopeSafe&&=q.environment===environment&&(q.signal===undefined||q.signal==='logs')&&JSON.stringify(q.where)===JSON.stringify({all:[{field:{kind:'resource',key:'sparse432.bank'},op:'eq',value:bank}]});if(url.pathname==='/api/explore')owner.scopeSafe&&=q.fromMs===owner.expected?.fromMs&&q.toMs===owner.expected?.toMs&&q.limit===100&&(q.offset===undefined||q.offset===0);}
   }
   const response=await original(input,init);
   if(url.pathname==='/api/explore')response.clone().json().then(data=>{owner.runCount=data.count}).catch(()=>{owner.scopeSafe=false});
   return response;
  };
  click(button(shell,'Start stream'));return;
 }
 const owner=w.__sparse432,sources=()=>owner.calls.filter((call:any)=>call.source).length;
 if(command==='widen'){
  const recovery=shell.querySelector<HTMLElement>('.activity-recovery')!;owner.expected={fromMs:Number(recovery.dataset.fromMs),toMs:Number(recovery.dataset.toMs)};owner.scopeSafe&&=owner.expected.toMs-owner.expected.fromMs===900000&&owner.expected.toMs===Number(recovery.dataset.pollToMs);owner.localAt=sources();click(button(shell,'Search wider window'));owner.phase='staging';
 }
 if(command==='run'){owner.localAt=undefined;const fixed=[...document.querySelectorAll<HTMLElement>('.activity-analysis')].find(visible);click(fixed?.querySelector('#exploreRun'));owner.phase='running'}
 if(command==='back'){owner.localAt=sources();click(button(document,'Return to paused activity'));owner.phase='returning'}
 if(command==='resume'){owner.localAt=undefined;click(button(shell,'Resume with recent logs'));owner.phase='resuming'}
 if(command==='finish')owner.phase='complete';
});
export const journey=actions(()=>{
 const s=state.current;
 if(s.phase==='initial')return s.ready?[act('start')]:['Wait'];
 if(s.phase==='reading')return s.tailReads===1&&s.emptyReady?[act('widen')]:['Wait'];
 if(s.phase==='staging')return s.staged&&s.localSafe?[act('run')]:['Wait'];
 if(s.phase==='running')return s.applied&&s.exploreReads===1?[act('back')]:['Wait'];
 if(s.phase==='returning')return s.returned&&s.localSafe?[act('resume')]:['Wait'];
 if(s.phase==='resuming')return s.resumed?[act('finish')]:['Wait'];
 return['Wait'];
});
