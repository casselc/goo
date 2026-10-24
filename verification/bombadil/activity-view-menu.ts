import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({ready:[...document.querySelectorAll('button')].some(b=>b.textContent==='Start stream'),step:Number(document.documentElement.dataset.viewMenuStep||0),loaded:!!document.querySelector('.activity-view-options .definition-row button'),rows:document.querySelectorAll('.stream-entry .live-tail tbody>tr').length,reads:window.performance.getEntriesByType('resource').filter(e=>/\/api\/(tail|explore|compare|captures)(\/|$)/.test(new URL(e.name).pathname)).length,overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const configurationOnly=always(()=>state.current.rows===0&&state.current.reads===0);
export const fits=always(()=>!state.current.overflow);
export const progress=eventually(()=>state.current.step===3).within(20,'seconds');
const act=registerCustomAction('goo.activityViewMenu',async(document,_window,next:number)=>{const options=document.querySelector('.activity-view-options')as HTMLDetailsElement;options.open=next!==3;if(next===2)(options.querySelector('.definition-row button')as HTMLButtonElement).click();document.documentElement.dataset.viewMenuStep=String(next)});
export const choices=actions(()=>!state.current.ready?['Wait']:state.current.step===0?[act(1)]:state.current.step===1&&state.current.loaded?[act(2)]:state.current.step===2?[act(3)]:['Wait']);
