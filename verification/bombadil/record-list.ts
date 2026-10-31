import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const pop=document.querySelector<HTMLElement>('.population-workspace:not([hidden])'),records=pop?.querySelector('[data-composition-role=records]'),list=records?.querySelector('goo-record-list'),chart=pop?.querySelector<any>('[data-composition-role=aggregate] goo-chart'),snapshot=chart?.snapshot,now=snapshot&&JSON.stringify({id:snapshot.id,basis:snapshot.revision,clock:snapshot.capturedAt,total:snapshot.partition?.total}),held=document.documentElement.dataset.recordListHeld;
 const observation=(window as any).__gooRecordList297Observation||{reads:0,pages:0,writes:0};
 return {version:'goo.recordList297',phase:document.documentElement.dataset.recordListPhase||'initial',population:pop?.getAttribute('data-phase'),kind:pop?.querySelector<HTMLSelectElement>('[aria-label="Records view"]')?.value,rows:list?.querySelectorAll('.record-list-item').length||0,details:list?.querySelectorAll('.record-detail-row:not([hidden]) .record-detail').length||0,offset:Number(records?.querySelector<HTMLElement>('.population-record-page-facts')?.dataset.pageOffset||0),count:records?.querySelector('.population-count')?.textContent,held:!held||held===now,...observation,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete'&&state.current.kind==='record-list'&&state.current.offset===100&&state.current.rows===3&&state.current.details===1&&state.current.held).within(25,'seconds');
export const budget=always(()=>state.current.reads<=1&&state.current.pages<=1&&state.current.writes===0);
export const heldAggregate=always(()=>state.current.held);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.recordList',async(document,window,phase:string)=>{
 const pop=document.querySelector<HTMLElement>('.population-workspace:not([hidden])')!,button=(name:string)=>[...pop.querySelectorAll<HTMLButtonElement>('button')].find(e=>e.textContent===name)!,setKind=(value:string)=>{const select=pop.querySelector<HTMLSelectElement>('[aria-label="Records view"]')!;select.value=value;select.dispatchEvent(new Event('change',{bubbles:true}))};
 if(phase==='initial'){const observation={reads:0,pages:0,writes:0},original=window.fetch.bind(window);(window as any).__gooRecordList297Observation=observation;window.fetch=((input:any,init:any)=>{const url=new URL(typeof input==='string'?input:input.url||String(input),window.location.href),method=(init?.method||input?.method||'GET').toUpperCase();if(method==='POST'){if(url.pathname==='/api/explore')observation.reads++;if(url.pathname==='/api/explore/page')observation.pages++;if(['/api/exploration-views','/api/captures','/api/documents','/api/kept-observations/records'].includes(url.pathname))observation.writes++;}return original(input,init)}) as any;button('Run population view').click()}
 else if(phase==='reading'){const snapshot=pop.querySelector<any>('[data-composition-role=aggregate] goo-chart')!.snapshot;document.documentElement.dataset.recordListHeld=JSON.stringify({id:snapshot.id,basis:snapshot.revision,clock:snapshot.capturedAt,total:snapshot.partition?.total});pop.querySelector<HTMLButtonElement>('goo-record-list [aria-label="Inspect record"]')!.click()}
 else if(phase==='detailed')setKind('record-table');
 else if(phase==='table')setKind('record-list');
 else if(phase==='listed')button('Next spans page').click();
 else if(phase==='paged')pop.querySelector<HTMLButtonElement>('goo-record-list [aria-label="Inspect record"]')!.click();
 document.documentElement.dataset.recordListPhase=({initial:'reading',reading:'detailed',detailed:'table',table:'listed',listed:'paging',paged:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='reading'&&s.kind==='record-list'&&s.rows===100?[step('reading')]:s.phase==='detailed'&&s.details===1?[step('detailed')]:s.phase==='table'&&s.kind==='record-table'?[step('table')]:s.phase==='listed'&&s.kind==='record-list'&&s.rows===100&&s.details===0?[step('listed')]:s.phase==='paging'&&s.kind==='record-list'&&s.offset===100&&s.rows===3?[step('paged')]:['Wait']});
