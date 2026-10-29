import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({
 stage:Number(document.documentElement.dataset.captureSearchStage||0),rows:document.querySelectorAll('#captureCatalog tbody tr').length,
 source:window.performance.getEntriesByType('resource').filter(e=>/^\/api\/(tail|explore|measurements|inventory|discovery|query)$/.test(new URL(e.name).pathname)).length,
 overflow:document.documentElement.scrollWidth>window.innerWidth,
 selected:!!document.querySelector('#generalCaptureWorkspace:not([hidden])'),
 status:document.querySelector('#captureLibraryStatus')?.textContent||'',
}));
export const storedOnly=always(()=>state.current.source===0&&!state.current.overflow);
export const finished=eventually(()=>state.current.stage===5&&state.current.selected).within(18,'seconds');
const act=registerCustomAction('goo.captureSearch',async(document,_window,step:number)=>{
 const form=document.querySelector<HTMLFormElement>('.capture-search')!;
 if(step===1){const input=form.querySelector<HTMLInputElement>('input[type=search]')!;input.value='Shared title';input.dispatchEvent(new Event('input',{bubbles:true}))}
 if(step===2||step===4)form.requestSubmit();
 if(step===3){const select=[...form.querySelectorAll<HTMLSelectElement>('select')].find(s=>[...s.options].some(o=>o.value==='oldest'))!;select.value='oldest';select.dispatchEvent(new Event('change',{bubbles:true}))}
 if(step===5)document.querySelector<HTMLButtonElement>('#captureCatalog tbody tr button')!.click();
 document.documentElement.dataset.captureSearchStage=String(step);
});
export const choices=actions(()=>{const s=state.current;if(s.stage===0&&!s.rows)return ['Wait'];if((s.stage===2||s.stage===4)&&(!s.status.includes('2 matching')||s.rows!==2))return ['Wait'];return s.stage<5?[act(s.stage+1)]:['Wait']});
