// Actual Services -> reported response group -> scoped records -> held Services.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{
 const w=document.defaultView as any,o=w.__http410,calls=o?.calls||[],visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert]'),d=document.querySelector('.service-dashboard'),context=d?.querySelector('details.dashboard-http-server-context'),card=d?.querySelector('[data-widget="http-server-responses"]'),last=calls.findLast((c:any)=>c.path==='/api/explore'),source=calls.filter((c:any)=>c.source).length,preserved=!o?.held||o.held.every((x:any)=>x.node.isConnected&&x.node.querySelector('.dashboard-widget-evidence pre')?.textContent===x.evidence);
 return{version:'goo.httpServer410Actor/1',phase:o?.phase||'initial',ready:visible(document.getElementById('captureGeneralLive')),catalogReady:[...document.querySelectorAll('#serviceCatalog button')].some(e=>e.getAttribute('aria-label')==='Open service '+o?.namespace+' / '+o?.service),serviceReady:visible(d)&&['loaded','partial'].includes(d?.getAttribute('data-state')||'')&&!d?.querySelector('[aria-busy=true]')&&!!context,contextOpen:!!context?.hasAttribute('open')&&visible(card),runReady:visible(document.getElementById('exploreRun')),recordsReady:!!last?.response&&visible(document.getElementById('exploreRun')),returned:visible(d)&&preserved,preserved,sourceSafe:!o||o.stageBefore===undefined||o.phase!=='staged'||source===o.stageBefore,returnSafe:!o||o.returnBefore===undefined||o.phase!=='complete'||source===o.returnBefore,sources:source,forbidden:o?.forbidden||[],calls:calls.map((c:any)=>({path:c.path,status:c.status,method:c.method})),...(o?.phase==='running'&&last?.response?{observed:last}:{})};
});
export const completes=eventually(()=>state.current.phase==='complete'&&state.current.returned).within(40,'seconds');
export const scopeSafe=always(()=>state.current.sourceSafe&&state.current.returnSafe&&state.current.preserved&&state.current.forbidden.length===0);
const act=registerCustomAction('goo.httpServer410',async(document,window,command:string)=>{
 const w=window as any,visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert]'),click=(e:any)=>{if(!e||e.disabled||!visible(e))throw Error('Missing current response-context action');e.scrollIntoView({block:'center',behavior:'instant'});e.focus();e.click()};
 if(command==='init'){
  const expected=JSON.parse(new URL(window.location.href).searchParams.get('expected')||'null'),o=w.__http410={phase:'capture',...expected,calls:[],forbidden:[]},original=window.fetch.bind(window);
  window.fetch=async(input:any,init:any)=>{const u=new URL(typeof input==='string'?input:input.url,window.location.href),method=String(init?.method||input?.method||'GET').toUpperCase(),source=method==='POST'&&['/api/discovery','/api/inventory','/api/metric-discovery','/api/explore','/api/record-aggregates','/api/measurements','/api/tail'].includes(u.pathname),c:any={path:u.pathname,method,source};if(init?.body)try{c.request=JSON.parse(init.body)}catch{}if(u.pathname.startsWith('/api/'))o.calls.push(c);if(method==='POST'&&['/api/captures','/api/documents','/api/exploration-views'].includes(u.pathname))o.forbidden.push(u.pathname);const response=await original(input,init);c.status=response.status;try{c.response=await response.clone().json()}catch{}return response};return;
 }
 const o=w.__http410,source=()=>o.calls.filter((c:any)=>c.source).length;
 if(command==='browse'){click(document.getElementById('captureGeneralLive'));o.phase='browsing'}
 if(command==='service'){click([...document.querySelectorAll('#serviceCatalog button')].find(e=>e.getAttribute('aria-label')==='Open service '+o.namespace+' / '+o.service));o.phase='service'}
 if(command==='context'){o.openBefore=source();click(document.querySelector('details.dashboard-http-server-context>summary'));o.phase='context'}
 if(command==='stage'){
  if(source()!==o.openBefore)throw Error('Opening response context read telemetry');const d=document.querySelector('.service-dashboard')!;o.held=[...d.querySelectorAll('.dashboard-widget')].map(e=>({node:e,evidence:e.querySelector('.dashboard-widget-evidence pre')?.textContent}));o.stageBefore=source();click(document.querySelector('[data-widget="http-server-responses"] tr[data-diagnosis-value="null"] a'));o.phase='staged';
 }
 if(command==='run'){click(document.getElementById('exploreRun'));o.phase='running'}
 if(command==='back'){
  const last=o.calls.findLast((c:any)=>c.path==='/api/explore'),stable=(x:any):any=>Array.isArray(x)?x.map(stable):x&&typeof x==='object'?Object.fromEntries(Object.keys(x).sort().map(k=>[k,stable(x[k])])):x;
  if(JSON.stringify(stable(last.request))!==JSON.stringify(stable(o.question))||last.response.count!=='3')throw Error('Response group changed exact missing-key population');if(source()!==o.stageBefore+1)throw Error('Response drill acquired extra source data');o.returnBefore=source();const services=document.querySelector('[data-workspace=services]');if(!visible(services))click(document.querySelector('.entry-menu'));click(services);o.phase='returning';
 }
 if(command==='complete')o.phase='complete';
});
export const journey=actions(()=>{const s=state.current;if(s.phase==='initial')return s.ready?[act('init')]:['Wait'];if(s.phase==='capture')return[act('browse')];if(s.phase==='browsing')return s.catalogReady?[act('service')]:['Wait'];if(s.phase==='service')return s.serviceReady?[act('context')]:['Wait'];if(s.phase==='context')return s.contextOpen?[act('stage')]:['Wait'];if(s.phase==='staged')return s.runReady?[act('run')]:['Wait'];if(s.phase==='running')return s.recordsReady&&s.observed?[act('back')]:['Wait'];if(s.phase==='returning')return s.returned?[act('complete')]:['Wait'];return['Wait']});
