// One explicit fixed read, then source-free observed-value exploration and staging.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({step:Number(document.documentElement.dataset.suggestionStep||0),phase:document.getElementById('exploration')?.dataset.questionPhase,ready:!!document.getElementById('exploreRun'),rows:document.querySelectorAll('#exploreRows .record-row').length,overflow:document.documentElement.scrollWidth>window.innerWidth,choices:document.querySelectorAll('#exploreFilterEditor button[aria-label^="Add attribute:release "]').length,note:document.querySelector('#exploreFilterEditor .filter-observations>p')?.textContent||''}));
export const noHiddenExecution=always(()=>state.current.step<3||state.current.phase==='draft'&&state.current.rows===0&&state.current.note.includes('Run or Start'));
export const distinctChoices=always(()=>state.current.step!==2||state.current.choices===4);
export const fits=always(()=>!state.current.overflow);
export const progress=eventually(()=>state.current.step===4).within(25,'seconds');
const act=registerCustomAction('goo.observeFields',async(document,_window,step:number)=>{
 if(step===1)document.getElementById('exploreRun')?.click();
 if(step===2){(document.getElementById('filterComposer') as HTMLDetailsElement).open=true;const panel=document.querySelector<HTMLDetailsElement>('#exploreFilterEditor .filter-observations');if(panel)panel.open=true;const search=panel?.querySelector<HTMLInputElement>('input');if(search){search.value='attribute:release';search.dispatchEvent(new Event('input',{bubbles:true}))}panel?.querySelectorAll<HTMLDetailsElement>(':scope>div>details').forEach(d=>d.open=true)}
 if(step===3)document.querySelector<HTMLButtonElement>('#exploreFilterEditor button[aria-label="Add attribute:release Present empty string"]')?.click();
 document.documentElement.dataset.suggestionStep=String(step)
});
export const choices=actions(()=>{const s=state.current;if(!s.ready)return['Wait'];if(s.step===0)return[act(1)];if(s.step===1)return s.phase==='applied'?[act(2)]:['Wait'];if(s.step<4)return[act(s.step+1)];return['Wait']});
