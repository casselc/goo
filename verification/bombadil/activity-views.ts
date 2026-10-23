// Saved configuration inspection must never start a stream.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({stage:Number(document.documentElement.dataset.activityViewStage||0),ready:[...document.querySelectorAll('button')].some(b=>b.textContent==='Start stream'),rows:document.querySelectorAll('.stream-entry .live-tail tbody>tr').length,phase:document.querySelector<HTMLElement>('.stream-entry .live-tail')?.dataset.state,overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const staged=always(()=>!state.current.ready||state.current.rows===0&&state.current.phase==='paused'&&!state.current.overflow);
export const complete=eventually(()=>state.current.stage===2).within(20,'seconds');
const act=registerCustomAction('goo.activityViewInspect',async(document,_window,next:number)=>{const options=document.querySelector<HTMLDetailsElement>('.stream-entry .activity-view-options');if(options)options.open=next===1;document.documentElement.dataset.activityViewStage=String(next)});
export const choices=actions(()=>!state.current.ready?['Wait']:state.current.stage===0?[act(1)]:state.current.stage===1?[act(2)]:['Wait']);
