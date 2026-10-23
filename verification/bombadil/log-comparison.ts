import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({canStart:[...document.querySelectorAll<HTMLButtonElement>('.evidence-comparison button')].some(b=>b.textContent==='Compare captures'&&!b.disabled),ready:!!document.querySelector('.evidence-comparison-result[data-state=comparable]'),headers:document.querySelectorAll('.comparison-breakdown table th').length,sourceReads:window.performance.getEntriesByType('resource').filter(e=>/\/api\/explore$/.test(new URL(e.name).pathname)).length,stage:Number(document.documentElement.dataset.logCompareStage||0)}));
export const noRecordRead=always(()=>state.current.sourceReads===0);
export const logColumns=always(()=>!state.current.ready||state.current.headers===5);
export const complete=eventually(()=>state.current.stage===2&&state.current.ready).within(10,'seconds');
const step=registerCustomAction('goo.logComparison',async(document,_window,stage:number)=>{if(stage===0)[...document.querySelectorAll<HTMLButtonElement>('.evidence-comparison button')].find(b=>b.textContent==='Compare captures')!.click();else{const box=document.querySelector<HTMLElement>('.comparison-breakdown>.scroll')!;box.scrollLeft=box.scrollWidth}document.documentElement.dataset.logCompareStage=String(stage+1)});
export const choices=actions(()=>state.current.stage===0&&state.current.canStart?[step(0)]:state.current.stage===1&&state.current.ready?[step(1)]:['Wait']);
