// Fresh generic entry uses observed scope discovery; no manifest, guessed label or writes.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state=extract(({document,window})=>{
 const o=(window as any).__entry430,visible=(e:any)=>{
  if(!e?.isConnected||!e.getClientRects().length||e.closest('[hidden],[inert],[aria-hidden="true"]'))return false;
  for(let p=e.parentElement;p;p=p.parentElement)if(p.localName==='details'&&!p.open&&!p.querySelector(':scope>summary')?.contains(e))return false;
  return true;
 },shell=document.querySelector('.stream-entry'),box=shell?.querySelector('.environment-adoption'),tail=shell?.querySelector('.live-tail'),findButton=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((e:any)=>e.textContent===label&&visible(e)&&!e.disabled),expected=new URL(window.location.href).searchParams.get('entry430')||'',previewRow=[...box?.querySelectorAll('tbody tr')||[]].find(e=>e.querySelector('th[scope="row"],td')?.textContent==='Named environment: '+JSON.stringify(expected)),previewButton=previewRow?.querySelector('button'),calls=o?.calls||[],source=calls.filter((c:any)=>c.source).length,start=findButton(shell,'Start stream'),pause=findButton(shell,'Pause stream'),resume=findButton(shell,'Resume with recent logs'),phase=tail?.getAttribute('data-state'),stage=findButton(box,'Stage in Activity'),rows=tail?.querySelectorAll('tbody tr').length||0;
 const initialSources=window.performance.getEntriesByType('resource').filter((e:any)=>['/api/tail','/api/environment-scopes','/api/environment-scopes/records','/api/explore','/api/compare'].includes(new URL(e.name,window.location.href).pathname)).length;
 return{version:'goo.unknownDataEntry430Actor/1',phase:o?.phase||'initial',ready:document.documentElement.dataset.profile==='generic'&&!!box&&!!expected,opened:!!box?.hasAttribute('open'),readReady:!!findButton(box,'Read environment scopes'),previewReady:!!previewButton&&visible(previewButton)&&!previewButton.disabled,stageReady:!!stage,staged:phase==='paused'&&!!start&&rows===0&&document.activeElement===start,streaming:rows>0&&['unknown','partial'].includes(phase||''),paused:phase==='paused'&&!!resume,rows,pauseFocused:!!pause&&document.activeElement===pause,resumeFocused:!!resume&&document.activeElement===resume,discoveryReads:calls.filter((c:any)=>c.path==='/api/environment-scopes').length,previewReads:calls.filter((c:any)=>c.path==='/api/environment-scopes/records').length,tailReads:calls.filter((c:any)=>c.path==='/api/tail').length,forbidden:o?.forbidden||[],focusSafe:o?.focusSafe!==false,scopeSafe:o?.scopeSafe!==false,localSafe:!o?initialSources===0:o.localAt===undefined||source===o.localAt,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const sourceAndScope=always(()=>state.current.localSafe&&state.current.scopeSafe&&state.current.forbidden.length===0&&state.current.discoveryReads<=1&&state.current.previewReads<=1&&state.current.tailReads<=2);
export const keyboardFocus=always(()=>state.current.focusSafe);
export const contained=always(()=>!state.current.overflow);

const act=registerCustomAction('goo.unknownDataEntry430',async(document,window,command:string)=>{
 const w=window as any,shell=document.querySelector('.stream-entry')!,box=shell.querySelector<HTMLDetailsElement>('.environment-adoption')!,tail=shell.querySelector('.live-tail'),expected=new URL(window.location.href).searchParams.get('entry430'),visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert]'),button=(root:any,text:string)=>[...root?.querySelectorAll('button')||[]].find((e:any)=>e.textContent===text&&visible(e)&&!e.disabled),click=(e:any)=>{if(!e||e.disabled||!visible(e))throw Error('Expected unknown-data entry control unavailable');e.scrollIntoView({block:'center',behavior:'instant'});e.focus({preventScroll:true});e.click()};
 if(command==='open'){
  if(!expected)throw Error('entry430 fixture environment required');
  const o=w.__entry430={phase:'opening',calls:[],forbidden:[],focusSafe:true,scopeSafe:true,localAt:0},original=window.fetch.bind(window);
  window.fetch=async(input:any,init:any)=>{
   const url=new URL(typeof input==='string'?input:input.url,window.location.href),method=String(init?.method||input?.method||'GET').toUpperCase();
   if(url.pathname.startsWith('/api/')&&method==='POST'){
    const call:any={path:url.pathname,source:['/api/environment-scopes','/api/environment-scopes/records','/api/tail'].includes(url.pathname)};o.calls.push(call);
    if(!call.source&&!/^\/api\/queries\/[^/]+\/cancel$/.test(url.pathname))o.forbidden.push(url.pathname);
    if(call.source){const q=JSON.parse(init?.body||'{}');if(url.pathname==='/api/environment-scopes')o.scopeSafe&&=q.signal==='logs';else if(url.pathname.endsWith('/records'))o.scopeSafe&&=q.signal==='logs'&&q.scope?.mode==='named'&&q.scope?.value===expected;else o.scopeSafe&&=(q.signal===undefined||q.signal==='logs')&&q.environmentSelector?.policy==='effective'&&q.environmentSelector?.scope?.mode==='named'&&q.environmentSelector?.scope?.value===expected;}
   }
   return original(input,init);
  };
  if(!box.open)click(box.querySelector(':scope>summary'));return;
 }
 const o=w.__entry430,sources=()=>o.calls.filter((c:any)=>c.source).length;
 if(command==='read'){o.localAt=undefined;click(button(box,'Read environment scopes'));o.phase='reading'}
 if(command==='preview'){
  const row=[...box.querySelectorAll('tbody tr')].find(e=>e.querySelector('th[scope="row"],td')?.textContent==='Named environment: '+JSON.stringify(expected));
  click(row?.querySelector('button'));o.phase='previewing';
 }
 if(command==='stage'){o.localAt=sources();click(button(box,'Stage in Activity'));o.phase='staging'}
 if(command==='start'){o.localAt=undefined;click(button(shell,'Start stream'));o.phase='streaming'}
 if(command==='pause'){o.localAt=sources();click(button(shell,'Pause stream'));o.focusSafe&&=document.activeElement===button(shell,'Resume with recent logs');o.phase='paused'}
 if(command==='resume'){o.localAt=undefined;click(button(shell,'Resume with recent logs'));o.focusSafe&&=document.activeElement===button(shell,'Pause stream');o.phase='resuming'}
 if(command==='pauseAgain'){o.localAt=sources();click(button(shell,'Pause stream'));o.focusSafe&&=document.activeElement===button(shell,'Resume with recent logs');o.phase='pausedAgain'}
 if(command==='finish')o.phase='complete';
});
export const journey=actions(()=>{
 const s=state.current;
 if(s.phase==='initial')return s.ready?[act('open')]:['Wait'];
 if(s.phase==='opening')return s.opened&&s.readReady?[act('read')]:['Wait'];
 if(s.phase==='reading')return s.discoveryReads===1&&s.previewReady?[act('preview')]:['Wait'];
 if(s.phase==='previewing')return s.previewReads===1&&s.stageReady?[act('stage')]:['Wait'];
 if(s.phase==='staging')return s.staged?[act('start')]:['Wait'];
 if(s.phase==='streaming')return s.streaming?[act('pause')]:['Wait'];
 if(s.phase==='paused')return s.paused&&s.resumeFocused?[act('resume')]:['Wait'];
 if(s.phase==='resuming')return s.streaming&&s.pauseFocused?[act('pauseAgain')]:['Wait'];
 if(s.phase==='pausedAgain')return s.paused&&s.resumeFocused&&s.tailReads===2?[act('finish')]:['Wait'];
 return['Wait'];
});
