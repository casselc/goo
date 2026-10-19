import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{const card=document.querySelector('#documentCards [data-kind=comparison]'),result=card?.querySelector('.evidence-comparison-result');return{exists:!!card,state:result?.getAttribute('data-state'),text:result?.textContent||''}});
export const ready=eventually(()=>state.current.text.includes('Rate change (comparison − baseline): 2 {request}/s')).within(10,'seconds');
export const frozenDecision=always(()=>!state.current.state||state.current.state==='comparable');
const reopen=registerCustomAction('goo.reopenNativeWindow',async(document)=>{(document.querySelector('#documentCards [data-kind=comparison] .evidence-comparison > button')as HTMLButtonElement)?.click()});
export const inspect=actions(()=>state.current.state==='comparable'?[reopen()]:['Wait']);
