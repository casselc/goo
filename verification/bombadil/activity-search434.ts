// Local Find is not a source query; source-message search stages an explicit Run.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const owner=(window as any).__search434,params=new URL(window.location.href).searchParams,shell=document.querySelector('.stream-entry'),tail=shell?.querySelector<HTMLElement>('.live-tail'),source=shell?.querySelector<HTMLElement>('.activity-source-search');
 const visible=(e:any)=>{if(!e?.isConnected||!e.getClientRects().length||e.closest('[hidden],[inert],[aria-hidden="true"]'))return false;for(let p=e.parentElement;p;p=p.parentElement)if(p.localName==='details'&&!p.open&&!p.querySelector(':scope>summary')?.contains(e))return false;return true;};
 const button=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((e:any)=>e.textContent===label&&visible(e)&&!e.disabled),fixed=[...document.querySelectorAll<HTMLElement>('.activity-analysis')].find(visible),run=fixed?.querySelector<HTMLButtonElement>('#exploreRun'),calls=owner?.calls||[],sources=calls.filter((c:any)=>c.source).length,keys=[...tail?.querySelectorAll<HTMLElement>('[data-tail-key]')||[]].map(row=>row.dataset.tailKey),find=tail?.querySelector<HTMLInputElement>('input[type=search]');
 const initialSources=window.performance.getEntriesByType('resource').filter((entry:any)=>['/api/tail','/api/explore','/api/explore/page','/api/record-aggregates','/api/measurements'].includes(new URL(entry.name,window.location.href).pathname)).length;
 return {version:'goo.activitySearch434Actor/1',phase:owner?.phase||'initial',ready:document.documentElement.dataset.profile==='generic'&&!!params.get('search434Env')&&!!params.get('search434Bank')&&!!params.get('search434Term')&&!!button(shell,'Start stream'),loaded:keys.length===1&&tail?.dataset.state==='unknown'&&!!button(shell,'Pause stream'),found:tail?.dataset.state==='paused'&&find?.value===params.get('search434Term')&&!!button(shell,'Search log messages')&&Number(source?.dataset.toMs)-Number(source?.dataset.fromMs)===60000,staged:!!fixed&&fixed.dataset.questionPhase==='draft'&&!!run&&visible(run)&&document.activeElement===run,applied:!!fixed&&fixed.dataset.questionPhase==='applied'&&owner?.runCount==='1',returned:visible(shell)&&tail?.dataset.state==='paused'&&find?.value===params.get('search434Term')&&document.activeElement===button(shell,'Search log messages')&&JSON.stringify(keys)===JSON.stringify(owner?.keys)&&source?.dataset.toMs===String(owner?.expected?.toMs),tailReads:calls.filter((c:any)=>c.path==='/api/tail').length,exploreReads:calls.filter((c:any)=>c.path==='/api/explore').length,localSafe:!owner?initialSources===0:owner.localAt===undefined||owner.localAt===sources,scopeSafe:owner?.scopeSafe!==false,heldSafe:owner?.heldSafe!==false,forbidden:owner?.forbidden||[],overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const explicitReads=always(()=>state.current.localSafe&&state.current.forbidden.length===0&&state.current.tailReads<=1&&state.current.exploreReads<=1);
export const exactWindowAndScope=always(()=>state.current.scopeSafe&&state.current.heldSafe);
export const containment=always(()=>!state.current.overflow);
const act=registerCustomAction('goo.activitySearch434',async(document,window,command:string)=>{
 const w=window as any,params=new URL(window.location.href).searchParams,environment=params.get('search434Env'),bank=params.get('search434Bank'),term=params.get('search434Term'),shell=document.querySelector('.stream-entry')!,tail=shell.querySelector<HTMLElement>('.live-tail')!;
 const visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert]'),button=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((e:any)=>e.textContent===label&&visible(e)&&!e.disabled),click=(e:any)=>{if(!e||!visible(e)||e.disabled)throw Error('Expected source-message search control unavailable');e.scrollIntoView({block:'center',behavior:'instant'});e.focus({preventScroll:true});e.click();};
 if(command==='start'){
  if(!environment||!bank||!term)throw Error('Source-search fixture identity required');
  const owner=w.__search434={phase:'reading',calls:[],forbidden:[],localAt:undefined,scopeSafe:true,heldSafe:true,expected:null,keys:null,runCount:null},original=window.fetch.bind(window),parent={all:[{field:{kind:'resource',key:'search434.bank'},op:'eq',value:bank}]},where={all:[parent,{field:{kind:'column',key:'body'},op:'contains',value:term}]};
  const valid=(q:any)=>q.signal==='logs'&&q.environment===environment&&q.fromMs===owner.expected?.fromMs&&q.toMs===owner.expected?.toMs&&q.limit===100&&(q.offset===undefined||q.offset===0)&&JSON.stringify(q.where)===JSON.stringify(where);
  window.addEventListener('goo:question-state',event=>{const detail=(event as CustomEvent).detail;if(detail.action==='stage'){owner.staged=structuredClone(detail.requested);owner.scopeSafe&&=valid(detail.requested);}});
  window.fetch=async(input:any,init:any)=>{
   const url=new URL(typeof input==='string'?input:input.url,window.location.href),method=String(init?.method||input?.method||'GET').toUpperCase();
   if(url.pathname.startsWith('/api/')&&method!=='GET'){
    const q=JSON.parse(init?.body||'{}'),source=['/api/tail','/api/explore'].includes(url.pathname);owner.calls.push({path:url.pathname,source});
    if(!source&&!/^\/api\/queries\/[^/]+\/cancel$/.test(url.pathname))owner.forbidden.push(url.pathname);
    if(url.pathname==='/api/tail')owner.scopeSafe&&=q.environment===environment&&(q.signal===undefined||q.signal==='logs')&&JSON.stringify(q.where)===JSON.stringify(parent);
    if(url.pathname==='/api/explore')owner.scopeSafe&&=valid(q);
   }
   const response=await original(input,init);
   if(url.pathname==='/api/explore')response.clone().json().then(data=>{owner.runCount=data.count;owner.scopeSafe&&=data.rows?.length===1&&data.rows[0].Body===term;}).catch(()=>{owner.scopeSafe=false;});
   return response;
  };
  click(button(shell,'Start stream'));return;
 }
 const owner=w.__search434,sources=()=>owner.calls.filter((c:any)=>c.source).length;
 if(command==='find'){
  click(button(shell,'Pause stream'));owner.localAt=sources();
  const find=tail.querySelector<HTMLInputElement>('input[type=search]')!;find.focus({preventScroll:true});find.value=term!;find.dispatchEvent(new Event('input',{bubbles:true}));owner.phase='finding';
 }
 if(command==='stage'){
  const source=shell.querySelector<HTMLElement>('.activity-source-search')!;owner.expected={fromMs:Number(source.dataset.fromMs),toMs:Number(source.dataset.toMs)};owner.keys=[...tail.querySelectorAll<HTMLElement>('[data-tail-key]')].map(row=>row.dataset.tailKey);owner.scopeSafe&&=owner.expected.toMs-owner.expected.fromMs===60000&&owner.expected.toMs===Number(tail.dataset.lastTo);click(button(shell,'Search log messages'));owner.phase='staging';
 }
 if(command==='run'){owner.localAt=undefined;const fixed=[...document.querySelectorAll<HTMLElement>('.activity-analysis')].find(visible);click(fixed?.querySelector('#exploreRun'));owner.phase='running';}
 if(command==='back'){owner.localAt=sources();click(button(document,'Return to paused activity'));owner.phase='returning';}
 if(command==='finish'){owner.heldSafe&&=JSON.stringify([...tail.querySelectorAll<HTMLElement>('[data-tail-key]')].map(row=>row.dataset.tailKey))===JSON.stringify(owner.keys)&&tail.querySelector<HTMLInputElement>('input[type=search]')?.value===term;owner.phase='complete';}
});
export const journey=actions(()=>{
 const s=state.current;
 if(s.phase==='initial')return s.ready?[act('start')]:['Wait'];
 if(s.phase==='reading')return s.loaded&&s.tailReads===1?[act('find')]:['Wait'];
 if(s.phase==='finding')return s.found&&s.localSafe?[act('stage')]:['Wait'];
 if(s.phase==='staging')return s.staged&&s.localSafe?[act('run')]:['Wait'];
 if(s.phase==='running')return s.applied&&s.exploreReads===1?[act('back')]:['Wait'];
 if(s.phase==='returning')return s.returned&&s.localSafe?[act('finish')]:['Wait'];
 return ['Wait'];
});
