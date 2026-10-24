import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({step:Number(document.documentElement.dataset.valuesStep||0),phase:document.getElementById('exploration')?.dataset.questionPhase,ready:!!document.getElementById('exploreRun'),loaded:document.querySelector('#exploreFilterEditor .filter-counted-values')?.textContent?.includes('103 matching records'),hidden:(document.querySelector('#exploreFilterEditor .filter-counted-values') as HTMLElement)?.hidden,overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const staged=always(()=>state.current.step<4||state.current.phase==='draft'&&state.current.hidden);
export const fits=always(()=>!state.current.overflow);
export const progress=eventually(()=>state.current.step===4).within(25,'seconds');
const act=registerCustomAction('goo.countValues',async(document,_window,step:number)=>{
 if(step===1)document.getElementById('exploreRun')?.click();
 if(step===2){(document.getElementById('filterComposer') as HTMLDetailsElement).open=true;const panel=document.querySelector<HTMLDetailsElement>('#exploreFilterEditor .filter-observations');if(panel)panel.open=true;const search=panel?.querySelector<HTMLInputElement>('input');if(search){search.value='attribute:release';search.dispatchEvent(new Event('input',{bubbles:true}))}panel?.querySelectorAll<HTMLDetailsElement>(':scope>div>details').forEach(d=>d.open=true);document.querySelector<HTMLButtonElement>('button[aria-label="Load scoped values for attribute:release"]')?.click()}
 if(step===3){const search=document.querySelector<HTMLInputElement>('#exploreFilterEditor input[aria-label="Search counted values"]');if(search){search.value='canary';search.dispatchEvent(new Event('input',{bubbles:true}))}}
 if(step===4)[...document.querySelectorAll<HTMLButtonElement>('#exploreFilterEditor .filter-counted-values button')].find(b=>b.getAttribute('aria-label')==='Stage counted attribute:release "canary"')?.click();
 document.documentElement.dataset.valuesStep=String(step)
});
export const choices=actions(()=>{const s=state.current;if(!s.ready)return['Wait'];if(s.step===0)return[act(1)];if(s.step===1)return s.phase==='applied'?[act(2)]:['Wait'];if(s.step===2)return s.loaded?[act(3)]:['Wait'];if(s.step===3)return[act(4)];return['Wait']});
