import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({phase:document.querySelector('.population-workspace')?.getAttribute('data-phase'),stage:Number(document.documentElement.dataset.populationStage||0),layout:(document.querySelector('[aria-label="Population layout"]')as HTMLSelectElement)?.value,records:document.querySelectorAll('[data-population-signal]').length,overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const compact=always(()=>!state.current.overflow&&state.current.records<=2);
export const complete=eventually(()=>state.current.stage===2&&state.current.phase==='applied'&&state.current.layout==='charts-records').within(25,'seconds');
const run=registerCustomAction('goo.populationRun',async(document)=>{[...document.querySelectorAll<HTMLButtonElement>('.population-workspace button')].find(b=>b.textContent==='Run population view')!.click();document.documentElement.dataset.populationStage='1'});
const layout=registerCustomAction('goo.populationLayout',async(document)=>{const input=document.querySelector('[aria-label="Population layout"]')as HTMLSelectElement;input.value='charts-records';input.dispatchEvent(new Event('change',{bubbles:true}));document.documentElement.dataset.populationStage='2'});
export const choices=actions(()=>state.current.stage===0?state.current.phase==='staged'?[run()]:['Wait']:state.current.stage===1?state.current.phase==='applied'?[layout()]:['Wait']:['Wait']);
