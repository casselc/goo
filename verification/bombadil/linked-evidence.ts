import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>({count:document.querySelectorAll('.step-evidence-entry').length,latest:[...document.querySelectorAll('.step-evidence-entry strong')].filter(e=>e.textContent==='Latest linked capture').length,ready:!!document.querySelector('#documentCards [data-kind=question]')}));
export const ready=eventually(()=>state.current.count===4).within(10,'seconds');
export const historyRemainsAvailable=always(()=>!state.current.ready||state.current.count===4&&state.current.latest===1);
const reopen=registerCustomAction('goo.inspectLinkedEvidence',async(document,_window,index:number)=>{(document.querySelector('#activeInvestigationOpen') as HTMLButtonElement)?.click();(document.querySelectorAll('.step-evidence-entry button')[index] as HTMLButtonElement)?.click()});
export const evidence=actions(()=>state.current.count===4?[reopen(0),reopen(1),reopen(2),reopen(3)]:['Wait']);
