import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{const chart=document.querySelector<any>('[data-composition-role="aggregate"] goo-chart'),root=document.querySelector('.population-composition'),facts=document.querySelector<HTMLElement>('[data-composition-role="records"] .population-record-page-facts');return {version:'goo.composition279',phase:document.documentElement.dataset.compositionPhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),reads:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore').length,composed:!!root,first:root?.firstElementChild?.getAttribute('data-composition-role'),records:!!document.querySelector('[data-composition-role="records"]'),offset:facts?.dataset.pageOffset,basis:chart?.snapshot?.revision,total:chart?.snapshot?.partition?.total,held:!chart||!document.documentElement.dataset.compositionBasis||chart.snapshot.revision===document.documentElement.dataset.compositionBasis,saved:!!document.querySelector('.population-save-status a'),overflow:document.documentElement.scrollWidth>window.innerWidth}});
export const completes=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const readBudget=always(()=>state.current.reads<=(['page','save','reopen','complete'].includes(state.current.phase)?2:1));
export const groupBasisHeld=always(()=>state.current.held);
export const partition=always(()=>!state.current.total||state.current.total==='3');
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.composeView',async(document,window,phase:string)=>{
 const button=(name:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 const settings=()=>{document.querySelectorAll<HTMLDetailsElement>('.population-save,.population-composition-settings').forEach(d=>d.open=true)};
 const show=(checked:boolean)=>{settings();const input=document.querySelector<HTMLInputElement>('input[aria-label="Show records"]')!;input.checked=checked;input.dispatchEvent(new Event('change'))};
 if(phase==='initial')button('Run population view').click();
 else if(phase==='read'){settings();button('Compose aggregate and records').click();const chart=document.querySelector<any>('[data-composition-role="aggregate"] goo-chart')!;document.documentElement.dataset.compositionBasis=chart.snapshot.revision}
 else if(phase==='compose'){settings();button('Reverse component order').click()}
 else if(phase==='order')show(false);
 else if(phase==='hidden')show(true);
 else if(phase==='visible')button('Next spans page').click();
 else if(phase==='page'){settings();button('Save exploration view').click()}
 else if(phase==='save')window.location.hash=document.querySelector<HTMLAnchorElement>('.population-save-status a')!.hash;
 else if(phase==='reopen'){if(window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore').length!==2)throw Error('Configuration reopen acquired telemetry')}
 document.documentElement.dataset.compositionPhase=({initial:'read',read:'compose',compose:'order',order:'hidden',hidden:'visible',visible:'page',page:'save',save:'reopen',reopen:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='read'&&s.population==='applied'?[step('read')]:s.phase==='compose'&&s.composed?[step('compose')]:s.phase==='order'&&s.first==='records'?[step('order')]:s.phase==='hidden'&&!s.records?[step('hidden')]:s.phase==='visible'&&s.records?[step('visible')]:s.phase==='page'&&s.offset==='1'?[step('page')]:s.phase==='save'&&s.saved?[step('save')]:s.phase==='reopen'&&s.population==='staged'?[step('reopen')]:['Wait']});
