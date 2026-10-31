import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const pop=document.querySelector<HTMLElement>('.population-workspace:not([hidden])'),records=pop?.querySelector('[data-composition-role=records]'),list=records?.querySelector('goo-record-list'),kind=pop?.querySelector<HTMLSelectElement>('[aria-label="Records view"]'),observation=(window as any).__gooInteractions298||{reads:0,pages:0,keeps:0,writes:0};
 return {version:'goo.recordInteractions298',phase:document.documentElement.dataset.interactionsPhase||'initial',population:pop?.getAttribute('data-phase'),kind:kind?.value,blocked:kind?.disabled,selected:records?.querySelectorAll('input[data-record-selection]:checked').length||0,rows:list?.querySelectorAll('.record-list-item').length||0,inspect:list?.querySelectorAll('.record-detail-row:not([hidden]) .record-detail').length||0,kept:!!records?.querySelector('a[href^="#kept="]')||[...records?.querySelectorAll('a')||[]].some(e=>e.textContent==='Open saved'),offset:Number(records?.querySelector<HTMLElement>('.population-record-page-facts')?.dataset.pageOffset||0),overflow:document.documentElement.scrollWidth>window.innerWidth,...observation};
});
export const completes=eventually(()=>state.current.phase==='complete'&&state.current.offset===100&&state.current.rows===2&&state.current.inspect===1).within(27,'seconds');
export const selectionsBlockReplacement=always(()=>state.current.selected===0||state.current.blocked===true);
export const boundedPageSelection=always(()=>state.current.selected<=20);
export const budget=always(()=>state.current.reads<=1&&state.current.pages<=1&&state.current.keeps<=1&&state.current.writes===0);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.recordInteractions',async(document,window,phase:string)=>{
 const pop=document.querySelector<HTMLElement>('.population-workspace:not([hidden])')!,button=(name:string)=>[...pop.querySelectorAll<HTMLButtonElement>('button')].find(e=>e.textContent===name)!,setKind=(value:string)=>{const select=pop.querySelector<HTMLSelectElement>('[aria-label="Records view"]')!;select.value=value;select.dispatchEvent(new Event('change',{bubbles:true}))};
 if(phase==='initial'){const observation={reads:0,pages:0,keeps:0,writes:0},original=window.fetch.bind(window);(window as any).__gooInteractions298=observation;window.fetch=((input:any,init:any)=>{const url=new URL(typeof input==='string'?input:input.url||String(input),window.location.href),method=(init?.method||input?.method||'GET').toUpperCase();if(method==='POST'){if(url.pathname==='/api/explore')observation.reads++;if(url.pathname==='/api/explore/page')observation.pages++;if(url.pathname==='/api/kept-observations/records')observation.keeps++;if(['/api/exploration-views','/api/captures','/api/documents'].includes(url.pathname))observation.writes++;}return original(input,init)}) as any;button('Run population view').click()}
 else if(phase==='reading')pop.querySelector<HTMLInputElement>('input[data-record-selection]')!.click();
 else if(phase==='selected'){setKind('record-table');if(pop.querySelector<HTMLSelectElement>('[aria-label="Records view"]')!.value!=='record-list')throw Error('Selection was discarded by kind replacement');button('Keep selected').click()}
 else if(phase==='keeping')button('Clear selected records').click();
 else if(phase==='cleared')setKind('record-table');
 else if(phase==='table')setKind('record-list');
 else if(phase==='listed')button('Next spans page').click();
 else if(phase==='paged')pop.querySelector<HTMLButtonElement>('goo-record-list [aria-label="Inspect record"]')!.click();
 document.documentElement.dataset.interactionsPhase=({initial:'reading',reading:'selected',selected:'keeping',keeping:'cleared',cleared:'table',table:'listed',listed:'paging',paged:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='reading'&&s.kind==='record-list'&&s.rows===100?[step('reading')]:s.phase==='selected'&&s.selected===1&&s.blocked?[step('selected')]:s.phase==='keeping'&&s.kept?[step('keeping')]:s.phase==='cleared'&&s.selected===0?[step('cleared')]:s.phase==='table'&&s.kind==='record-table'?[step('table')]:s.phase==='listed'&&s.kind==='record-list'?[step('listed')]:s.phase==='paging'&&s.offset===100&&s.rows===2?[step('paged')]:['Wait']});
