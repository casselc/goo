import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({version:'goo.recordStep284',phase:document.documentElement.dataset.stepHandoffPhase||'initial',reads:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore').length,rows:document.querySelectorAll('#exploreRows .record-row').length,cards:document.querySelectorAll('#documentCards>[data-kind="question"]').length,stored:document.querySelector('#documentRevision')?.textContent?.includes('stored revision'),scope:document.querySelector('#exploreQuestionState')?.textContent,overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const complete=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const budget=always(()=>state.current.reads<=2);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.recordStepHandoff',async(document,window,phase:string)=>{
 const click=(id:string)=>document.querySelector<HTMLButtonElement>('#'+id)!.click();
 if(phase==='initial'||phase==='question')click('exploreRun');
 else if(phase==='read'){document.querySelector<HTMLDetailsElement>('.question-preserve')!.open=true;click('exploreAddStep')}
 else if(phase==='pin')click('documentSave');
 else if(phase==='savedoc')[...document.querySelectorAll<HTMLButtonElement>('#documentCards button')].find(b=>b.textContent==='Reopen question')!.click();
 document.documentElement.dataset.stepHandoffPhase=({initial:'read',read:'pin',pin:'savedoc',savedoc:'question',question:'executing',executing:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'?[step('initial')]:s.phase==='read'&&s.rows===1?[step('read')]:s.phase==='pin'&&s.cards===1?[step('pin')]:s.phase==='savedoc'&&s.stored?[step('savedoc')]:s.phase==='question'&&s.rows===0?[step('question')]:s.phase==='executing'&&s.rows===1&&s.reads===2?[step('executing')]:['Wait']});
