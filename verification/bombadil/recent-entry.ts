import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({phase:document.querySelector('.population-workspace')?.getAttribute('data-phase'),stage:Number(document.documentElement.dataset.recentStage||0),live:document.querySelectorAll('.live-tail:not([hidden])').length,overflow:document.documentElement.scrollWidth>window.innerWidth,logs:document.querySelector('[data-population-signal=logs]')?.hasAttribute('hidden')}));
export const safe=always(()=>!state.current.overflow&&state.current.live===0);
export const complete=eventually(()=>state.current.stage===2&&state.current.phase==='applied'&&!state.current.logs).within(25,'seconds');
const run=registerCustomAction('goo.recentActivity',async(document)=>{document.querySelector<HTMLButtonElement>('#exploreRecent')!.click();document.documentElement.dataset.recentStage='1'});
const logs=registerCustomAction('goo.recentLogs',async(document)=>{document.querySelector<HTMLButtonElement>('.population-signal-tabs button[data-signal=logs]')!.click();document.documentElement.dataset.recentStage='2'});
export const choices=actions(()=>state.current.stage===0?state.current.phase==='staged'?[run()]:['Wait']:state.current.stage===1?state.current.phase==='applied'?[logs()]:['Wait']:['Wait']);
