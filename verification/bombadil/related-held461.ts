// Source setup is explicit; after Hold the independent driver permits local work only.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{const root=document.querySelector('.paired-activity'),view=root?.querySelector<any>('goo-held-chronology'),phase=Number(document.documentElement.dataset.related461Phase||0),held=view?.data,o=(window as any).__related461||{};return{version:'goo.relatedHeld461Actor/1',phase,ready:!!document.querySelector('[aria-label="Live signal"]'),rows:root?.querySelectorAll('[data-tail-key]').length||0,related:view?.dataset.related||'none',calls:o.calls||[],unchanged:!o.bank||o.bank===JSON.stringify(held),hash:window.location.hash,initialHash:o.hash||window.location.hash,overflow:document.documentElement.scrollWidth>window.innerWidth}});
export const noLocalAcquisition=always(()=>state.current.calls.length===0);
export const heldBankUnchanged=always(()=>state.current.unchanged);
export const noNavigation=always(()=>state.current.hash===state.current.initialHash);
export const fits=always(()=>!state.current.overflow);
export const completes=eventually(()=>state.current.phase===8&&state.current.related==='none').within(30,'seconds');
const step=registerCustomAction('goo.related461',async(document,window,next:number)=>{
 const root=document.querySelector('.paired-activity'),view=root?.querySelector<any>('goo-held-chronology'),visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert]'),click=(e:any)=>{if(!e||e.disabled||!visible(e))throw Error('Current related-record control unavailable');e.scrollIntoView({block:'nearest'});e.click()},button=(host:any,name:string)=>[...host.querySelectorAll('button')].find((e:any)=>e.textContent===name);
 if(next===1){const select=document.querySelector<HTMLSelectElement>('[aria-label="Live signal"]')!;select.value='both';select.dispatchEvent(new Event('change',{bubbles:true}))}
 else if(next===2)click([...document.querySelectorAll('.stream-entry-toolbar button')].find(e=>/^(Start streams|Apply scope)$/.test(e.textContent||'')));
 else if(next===3){click(button(root,'Hold chronology'));const target=window as any,o=target.__related461={calls:[],bank:JSON.stringify(view.data),hash:window.location.hash},original=window.fetch;window.fetch=function(input,options){const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href),method=String(options?.method||(input instanceof Request?input.method:'GET')).toUpperCase();if(method==='POST'&&url.pathname.startsWith('/api/'))o.calls.push(url.pathname);return Reflect.apply(original,this,[input,options])}}
 else if(next===4){click(button(view,'More records'));const row=[...view.querySelectorAll('[data-chronology-key]')].find((e:any)=>e.dataset.signal==='spans'&&e.textContent.includes('child operation'));click(row?.querySelector('[data-chronology-inspect]'))}
 else if(next===5)click(button(root.querySelector('.ui-inspector'),'Related held records'));
 else if(next===6)click(view.querySelector('input[type=checkbox]'));
 else if(next===7)click(button(view,'Return to chronology'));
 else if(next===8)click(root.querySelector('[aria-label="Close selected item details"]'));
 document.documentElement.dataset.related461Phase=String(next);
});
export const explore=actions(()=>{const s=state.current;if(s.phase===0)return s.ready?[step(1)]:['Wait'];if(s.phase===1)return[step(2)];if(s.phase===2)return s.rows===120?[step(3)]:['Wait'];if(s.phase<8)return[step(s.phase+1)];return['Wait']});
