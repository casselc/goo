import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({phase:document.querySelector('.population-workspace')?.getAttribute('data-phase'),stage:Number(document.documentElement.dataset.fieldStage||0),reads:window.performance.getEntriesByType('resource').filter(e=>e.name.endsWith('/api/explore')).length,basis:Number(document.documentElement.dataset.fieldReads||0),overflow:document.documentElement.scrollWidth>window.innerWidth,staged:!!document.querySelector('.population-selection .selection-clause'),fields:document.querySelectorAll('[data-population-signal=logs] th[data-record-field]').length}));
export const local=always(()=>!state.current.overflow&&(state.current.stage<2||state.current.reads===state.current.basis));
export const bounded=always(()=>state.current.fields<=8);
export const complete=eventually(()=>state.current.stage===6&&state.current.staged).within(25,'seconds');
const step=registerCustomAction('goo.recordFields',async(document,window,stage:number)=>{
 const panel=document.querySelector('.population-workspace')!;
 if(stage===0)[...panel.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Run population view')!.click();
 if(stage===1){panel.querySelector<HTMLButtonElement>('[data-signal=logs]')?.click();document.documentElement.dataset.fieldReads=String(window.performance.getEntriesByType('resource').filter(e=>e.name.endsWith('/api/explore')).length)}
 if(stage>=2&&stage<5){const picker=panel.querySelector<HTMLElement>('[data-population-signal=logs] goo-record-columns')!;picker.querySelector<HTMLDetailsElement>('details')!.open=true;const label=stage===2?'Show Service':stage===3?'Hide Service column':'Move resource:region later';const button=[...picker.querySelectorAll<HTMLButtonElement>('button')].find(b=>(b.getAttribute('aria-label')||b.textContent)===label);button!.click()}
 if(stage===5){const cell=[...panel.querySelectorAll<HTMLTableCellElement>('[data-population-signal=logs] td[data-record-field]')].find(e=>e.dataset.recordField==='["attribute","region"]'&&e.querySelector('span')?.textContent==='west')!;cell.querySelector<HTMLDetailsElement>('details')!.open=true;cell.querySelector<HTMLButtonElement>('[aria-label="Exclude attribute:region column value"]')!.click()}
 document.documentElement.dataset.fieldStage=String(stage+1);
});
export const choices=actions(()=>state.current.stage===0?state.current.phase==='staged'?[step(0)]:['Wait']:state.current.stage===1?state.current.phase==='applied'?[step(1)]:['Wait']:state.current.stage<6?[step(state.current.stage)]:['Wait']);
