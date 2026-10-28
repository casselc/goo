import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const panel=document.querySelector<HTMLElement>('.stream-entry>.ui-inspector'),heading=panel?.querySelector<HTMLElement>('.ui-inspector-heading>:is(h2,h3,h4)'),focused=document.activeElement as HTMLElement;
 const rect=focused?.getBoundingClientRect(),inView=!!rect&&rect.top>=0&&rect.bottom<=window.innerHeight;
 return {stage:Number(document.documentElement.dataset.inspectStage||0),reads:Number(document.documentElement.dataset.inspectReads||0),ready:[...document.querySelectorAll<HTMLButtonElement>('button')].some(b=>b.textContent==='Start stream'&&b.getClientRects().length>0),rows:document.querySelectorAll('[data-tail-inspect]').length,open:!!panel&&!panel.hidden,heading:focused===heading,returned:!!focused?.matches('[data-tail-inspect][aria-expanded=false]'),inView,expanded:document.querySelectorAll('[data-tail-inspect][aria-expanded=true]').length,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const sourceFree=always(()=>state.current.reads<=1&&!state.current.overflow);
export const visibleDetails=always(()=>![2,4].includes(state.current.stage)||state.current.open&&state.current.heading&&state.current.inView&&state.current.expanded===1);
export const visibleReturn=always(()=>![3,5].includes(state.current.stage)||!state.current.open&&state.current.returned&&state.current.inView&&state.current.expanded===0);
export const scopeStaging=always(()=>state.current.stage<6||!state.current.open&&state.current.reads===1);
export const complete=eventually(()=>state.current.stage===7).within(18,'seconds');
const act=registerCustomAction('goo.inspectAndReturn',async(document,window,next:number)=>{
 if(next===1){
  const original=window.fetch.bind(window);window.fetch=async(input,init)=>{if(String(input)==='/api/tail')document.documentElement.dataset.inspectReads=String(Number(document.documentElement.dataset.inspectReads||0)+1);return original(input,init)};
  [...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Start stream'&&b.getClientRects().length>0)!.click();
 }else if(next===2||next===4){const buttons=[...document.querySelectorAll<HTMLButtonElement>('[data-tail-inspect]')],button=buttons[next===2?0:1];button.scrollIntoView({block:'center'});button.focus();button.click()}
 else if(next===3){document.activeElement!.dispatchEvent(new KeyboardEvent('keydown',{key:'Escape',bubbles:true,cancelable:true}))}
 else if(next===5)document.querySelector<HTMLButtonElement>('.stream-entry>.ui-inspector [aria-label="Close selected item details"]')!.click();
 else if(next===6){const menu=document.querySelector<HTMLDetailsElement>('.live-tail tbody tr .ui-action-menu')!;menu.querySelector<HTMLElement>('summary')!.click()}
 else document.querySelector<HTMLButtonElement>('.live-tail tbody tr [aria-label="Include column:service from row"]')!.click();
 document.documentElement.dataset.inspectStage=String(next);
});
export const choices=actions(()=>state.current.stage===0&&state.current.ready?[act(1)]:state.current.stage===1&&state.current.rows>1?[act(2)]:state.current.stage>=2&&state.current.stage<7?[act(state.current.stage+1)]:['Wait']);
