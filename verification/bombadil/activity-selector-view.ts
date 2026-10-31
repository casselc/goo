import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({version:'goo.activitySelector283',phase:document.documentElement.dataset.selectorViewPhase||'initial',reads:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/tail').length,rows:document.querySelectorAll('.stream-entry .live-tail tbody tr').length,tail:document.querySelector('.stream-entry .live-tail')?.getAttribute('data-state'),saved:!!document.querySelector('.activity-view-options a[href^="#activity-view="]'),overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const complete=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const budget=always(()=>state.current.reads<=2);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.activitySelectorView',async(document,window,phase:string)=>{
 const button=(name:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 if(phase==='initial'||phase==='reopen')button('Start stream').click();
 else if(phase==='start'||phase==='restart')button('Pause stream').click();
 else if(phase==='pause'){const options=document.querySelector<HTMLDetailsElement>('.activity-view-options')!;options.open=true;const title=options.querySelector<HTMLInputElement>('[aria-label="Activity view title"]')!;title.value='Bombadil typed Activity';title.dispatchEvent(new Event('input',{bubbles:true}));button('Save Activity view').click()}
 else if(phase==='save')document.querySelector<HTMLAnchorElement>('.activity-view-options a[href^="#activity-view="]')!.click();
 document.documentElement.dataset.selectorViewPhase=({initial:'start',start:'pause',pause:'save',save:'reopen',reopen:'restart',restart:'stopping',stopping:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'?[step('initial')]:s.phase==='start'&&s.rows>0?[step('start')]:s.phase==='pause'&&s.tail==='paused'?[step('pause')]:s.phase==='save'&&s.saved?[step('save')]:s.phase==='reopen'&&s.rows===0?[step('reopen')]:s.phase==='restart'&&s.rows>0?[step('restart')]:s.phase==='stopping'&&s.tail==='paused'&&s.reads===2?[step('stopping')]:['Wait']});
