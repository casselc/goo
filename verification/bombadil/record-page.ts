import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{const charts=[...document.querySelectorAll<any>('[data-composition-role="aggregate"] goo-chart,[data-composition-role="timeline"] goo-chart')];return {version:'goo.page289',phase:document.documentElement.dataset.pagePhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),full:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore').length,pages:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore/page').length,offset:document.querySelector<HTMLElement>('.population-record-page-facts')?.dataset.pageOffset,held:!document.documentElement.dataset.pageHeld||JSON.stringify(charts.map(e=>({id:e.snapshot.id,basis:e.snapshot.revision,clock:e.snapshot.capturedAt})))===document.documentElement.dataset.pageHeld,overflow:document.documentElement.scrollWidth>window.innerWidth}});
export const complete=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const sourceBudget=always(()=>state.current.full<=1&&state.current.pages<=2);
export const heldCharts=always(()=>state.current.held);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.recordPage',async(document,_window,phase:string)=>{const button=(name:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 if(phase==='initial')button('Run population view').click();
 else if(phase==='read'){document.documentElement.dataset.pageHeld=JSON.stringify([...document.querySelectorAll<any>('[data-composition-role="aggregate"] goo-chart,[data-composition-role="timeline"] goo-chart')].map(e=>({id:e.snapshot.id,basis:e.snapshot.revision,clock:e.snapshot.capturedAt})));button('Next spans page').click()}
 else if(phase==='next')button('Previous spans page').click();
 document.documentElement.dataset.pagePhase=({initial:'read',read:'next',next:'previous',previous:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='read'&&s.population==='applied'?[step('read')]:s.phase==='next'&&s.offset==='1'&&s.pages===1?[step('next')]:s.phase==='previous'&&s.offset==='0'&&s.pages===2?[step('previous')]:['Wait']});
