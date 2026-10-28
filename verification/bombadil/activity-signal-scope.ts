// Explore a saved nested selector across signals. Transport observations belong
// to this driver, never the passive UI components or product runtime.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({
 ready:!![...document.querySelectorAll('button')].find(button=>button.textContent==='Start stream'&&button.getClientRects().length>0),
 stage:Number(document.documentElement.dataset.signalScopeStage||0),
 reads:Number(document.documentElement.dataset.signalScopeReads||0),
 changed:document.documentElement.dataset.signalScopeChanged==='true',
 phase:document.querySelector<HTMLElement>('.live-tail')?.dataset.state,
 signal:document.querySelector<HTMLSelectElement>('[aria-label="Live signal"]')?.value,
 overflow:document.documentElement.scrollWidth>window.innerWidth
}));
export const exactScope=always(()=>!state.current.changed&&!state.current.overflow);
export const explicitReads=always(()=>state.current.reads<=state.current.stage);
export const complete=eventually(()=>state.current.stage===4&&state.current.reads===4&&state.current.signal==='spans'&&['unknown','partial'].includes(state.current.phase||'')).within(18,'seconds');
const step=registerCustomAction('goo.keepSignalScope',async(document,window,next:number)=>{
 if(next===1){
  const original=window.fetch.bind(window);let expected='';
  window.fetch=async(input,init)=>{
   if(String(input)==='/api/tail'){
    const q=JSON.parse(String(init?.body)),where=JSON.stringify(q.where);
    if(!expected){expected=where;if(!where.includes('scope-api')||!where.includes('"any"')||!where.includes('"missing"')||!where.includes('"resource"'))document.documentElement.dataset.signalScopeChanged='true'}
    else if(where!==expected)document.documentElement.dataset.signalScopeChanged='true';
    document.documentElement.dataset.signalScopeReads=String(Number(document.documentElement.dataset.signalScopeReads||0)+1);
   }
   return original(input,init);
  };
  [...document.querySelectorAll<HTMLButtonElement>('button')].find(button=>button.textContent==='Start stream'&&button.getClientRects().length>0)!.click();
 }else{
  const signal=document.querySelector<HTMLSelectElement>('[aria-label="Live signal"]')!;
  signal.value=next%2===0?'spans':'logs';signal.dispatchEvent(new Event('change',{bubbles:true}));
 }
 document.documentElement.dataset.signalScopeStage=String(next);
});
export const choices=actions(()=>state.current.stage===0&&state.current.ready?[step(1)]:state.current.stage>0&&state.current.stage<4&&['unknown','partial'].includes(state.current.phase||'')?[step(state.current.stage+1)]:['Wait']);
