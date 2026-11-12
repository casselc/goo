// Local scope reuse over a real staged fixed question; no source or save authority.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const o=(window as any).__scopeHistoryActor,root=document.querySelector<HTMLDetailsElement>('.scope-history'),visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert]');
 return {version:'goo.scopeHistoryActor/1',phase:o?.phase||'initial',ready:visible(document.querySelector('#exploreRun'))&&!!root,pinned:root?.querySelectorAll('[data-pinned="true"]').length||0,recent:root?.querySelectorAll('[data-pinned="false"]').length||0,posts:o?.posts||[],rows:document.querySelectorAll('#exploreRows .record-row').length,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const localOnly=always(()=>state.current.posts.length===0&&state.current.rows===0);
export const contained=always(()=>!state.current.overflow);
export const completes=eventually(()=>state.current.phase==='complete'&&state.current.pinned===0&&state.current.recent===0).within(30,'seconds');
const act=registerCustomAction('goo.scopeHistory',async(document,window,command:string)=>{
 const w=window as any,root=document.querySelector<HTMLDetailsElement>('.scope-history');if(!root)throw Error('Scope picker unavailable');
 const button=(text:string,selector='')=>[...root.querySelectorAll<HTMLButtonElement>((selector?selector+' ':'')+'button')].find(e=>e.textContent===text),click=(node:any)=>{if(!node||node.disabled)throw Error('Expected scope action unavailable');node.scrollIntoView({block:'nearest',behavior:'instant'});node.focus();node.click()};
 if(!root.open)root.querySelector('summary')?.click();
 if(command==='pin'){const o=w.__scopeHistoryActor={phase:'pinning',posts:[]},original=window.fetch.bind(window);window.fetch=async(input:any,init:any)=>{const method=String(init?.method||input?.method||'GET').toUpperCase();if(method==='POST')o.posts.push(typeof input==='string'?input:input.url);return original(input,init)};click(button('Pin current scope'));o.phase='pinned';return}
 const o=w.__scopeHistoryActor;
 if(command==='stage'){click(button('Stage','[data-pinned="true"]'));o.phase='staged'}
 if(command==='unpin'){click(button('Unpin','[data-pinned="true"]'));o.phase='unpinned'}
 if(command==='forget'){click(button('Forget','[data-pinned="false"]'));o.phase='complete'}
});
export const journey=actions(()=>{const s=state.current;if(s.phase==='initial')return s.ready?[act('pin')]:['Wait'];if(s.phase==='pinned')return s.pinned===1?[act('stage')]:['Wait'];if(s.phase==='staged')return[act('unpin')];if(s.phase==='unpinned')return s.recent===1?[act('forget')]:['Wait'];return['Wait']});
