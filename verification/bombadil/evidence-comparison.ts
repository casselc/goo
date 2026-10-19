import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{const c=document.querySelector('.evidence-comparison');const r=c?.querySelector('.evidence-comparison-result');return {ready:!!c,blocked:['incompatible','unavailable'].includes(r?.getAttribute('data-state')||''),hasDelta:!!r?.textContent?.includes('Comparison − baseline')}});
export const ready=eventually(()=>state.current.ready).within(10,'seconds');
export const incompatibleHasNoDelta=always(()=>!state.current.blocked||!state.current.hasDelta);
const compare=registerCustomAction('goo.compareStoredEvidence',async(document,_window,index:number)=>{const c=document.querySelector('.evidence-comparison');const a=c?.querySelector('.evidence-baseline') as HTMLSelectElement,b=c?.querySelector('.evidence-selection') as HTMLSelectElement;if(!a||!b)return;a.selectedIndex=0;a.dispatchEvent(new Event('change'));b.selectedIndex=index;b.dispatchEvent(new Event('change'));(c?.querySelector('button') as HTMLButtonElement)?.click()});
export const comparisons=actions(()=>state.current.ready?[compare(1),compare(2),compare(3)]:['Wait']);
