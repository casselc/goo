import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({phase:document.querySelector('.population-workspace')?.getAttribute('data-phase'),stage:Number(document.documentElement.dataset.liveStage||0),tail:document.querySelector('.population-live .live-tail')?.getAttribute('data-state'),rows:document.querySelectorAll('.population-live .live-tail tbody tr').length,overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const bounded=always(()=>state.current.rows<=100&&!state.current.overflow);
export const complete=eventually(()=>state.current.stage===2&&state.current.tail==='paused').within(25,'seconds');
const start=registerCustomAction('goo.populationTailStart',async(document)=>{[...document.querySelectorAll<HTMLButtonElement>('.population-live button')].find(b=>b.textContent==='Start population tail')!.click();document.documentElement.dataset.liveStage='1'});
const pause=registerCustomAction('goo.populationTailPause',async(document)=>{[...document.querySelectorAll<HTMLButtonElement>('.population-live button')].find(b=>b.textContent==='Pause stream')!.click();document.documentElement.dataset.liveStage='2'});
export const choices=actions(()=>state.current.stage===0?state.current.phase==='staged'?[start()]:['Wait']:state.current.stage===1?state.current.tail==='unknown'?[pause()]:['Wait']:['Wait']);
