// The campaign observes a real missing owned source, then proves terminal
// quiescence and rejection of retained Retry callbacks after scope/route changes.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const o=(window as any).__tail433,tail=document.querySelector<HTMLElement>('.stream-entry .live-tail'),start=[...document.querySelectorAll<HTMLButtonElement>('.stream-entry button')].find(e=>e.textContent==='Start stream'&&!e.disabled&&e.getClientRects().length),retry=tail?.querySelector<HTMLButtonElement>('.tail-retry');
 return{version:'goo.tailRefusal433Actor/1',phase:o?.phase||'initial',ready:document.documentElement.dataset.profile==='generic'&&!!start,terminal:tail?.dataset.state==='unavailable'&&tail?.dataset.failureReason==='source_missing'&&!!tail.querySelector('.source-refusal[data-source-refusal-kind="unavailable"]')&&!!retry&&!retry.hidden&&!retry.disabled,retained:Number(tail?.dataset.retained||0),quiet:!!o?.since&&window.performance.now()-o.since>2300,reads:o?.reads||0,safe:!o||o.reads<=1&&o.forbidden.length===0,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const terminalDoesNotPoll=always(()=>state.current.safe);
export const contained=always(()=>!state.current.overflow);
const act=registerCustomAction('goo.tailRefusal433',async(document,window,command:string)=>{
 const w=window as any;
 if(command==='start'){
  const owner=w.__tail433={phase:'starting',reads:0,forbidden:[],since:0,retry:null},original=window.fetch.bind(window);
  window.fetch=async(input:any,init:any)=>{const p=new URL(typeof input==='string'?input:input.url,window.location.href).pathname,method=String(init?.method||input?.method||'GET').toUpperCase();if(p.startsWith('/api/')&&method==='POST'){if(p==='/api/tail')owner.reads++;else if(!/^\/api\/queries\/[^/]+\/cancel$/.test(p))owner.forbidden.push(p)}return original(input,init)};
  const start=[...document.querySelectorAll<HTMLButtonElement>('.stream-entry button')].find(e=>e.textContent==='Start stream'&&!e.disabled&&e.getClientRects().length)!;start.focus();start.click();return;
 }
 const o=w.__tail433;
 if(command==='observe'){const retry=document.querySelector<HTMLButtonElement>('.tail-retry')!;o.retry={element:retry,handler:retry.onclick};o.since=window.performance.now();o.phase='quiet'}
 if(command==='dirty'){const env=document.querySelector<HTMLInputElement>('.stream-entry input[aria-label="Live environment"]')!;env.value='retired-tail433';env.dispatchEvent(new Event('input',{bubbles:true}));o.retry.handler.call(o.retry.element);o.phase='dirty'}
 if(command==='hide'){window.location.hash='services';o.retry.handler.call(o.retry.element);o.phase='hidden'}
 if(command==='finish')o.phase='complete';
});
export const journey=actions(()=>{const s=state.current;if(s.phase==='initial')return s.ready?[act('start')]:['Wait'];if(s.phase==='starting')return s.terminal&&s.reads===1&&s.retained===0?[act('observe')]:['Wait'];if(s.phase==='quiet')return s.quiet&&s.terminal&&s.safe?[act('dirty')]:['Wait'];if(s.phase==='dirty')return s.safe?[act('hide')]:['Wait'];if(s.phase==='hidden')return s.safe?[act('finish')]:['Wait'];return['Wait']});
