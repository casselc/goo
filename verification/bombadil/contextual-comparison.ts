import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const c=document.querySelector<any>('goo-measure-comparison')?.comparison,resources=window.performance.getEntriesByType('resource'),count=(path:string)=>resources.filter(e=>new URL(e.name).pathname===path).length,picker=document.querySelector<HTMLSelectElement>('[aria-label="Baseline capture"]'),expected=document.documentElement.dataset.contextualBaseline;
 return {version:'goo.contextual295',phase:document.documentElement.dataset.contextualPhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),editors:document.querySelectorAll('.population-workspace:not([hidden]) .population-count-editor').length,full:count('/api/explore'),compares:count('/api/compare'),writes:count('/api/exploration-views')+count('/api/captures')+count('/api/documents'),pair:!!c,baseline:c?.baseline.rows[0].values.count,current:c?.selection.rows[0].values.count,choice:!!expected&&!!picker?.querySelector('option[value="'+expected+'"]'),prepare:[...document.querySelectorAll<HTMLButtonElement>('button')].some(b=>b.textContent==='Prepare baseline'&&!b.disabled&&b.getClientRects().length>0),overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete'&&state.current.baseline==='2'&&state.current.current==='3').within(25,'seconds');
export const oneEditor=always(()=>state.current.editors<=1);
export const sourceBudget=always(()=>state.current.full<=1&&state.current.compares<=2&&state.current.writes===0);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.contextualComparison',async(document,window,phase:string)=>{
 const button=(text:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===text&&b.getClientRects().length>0)!;
 const open=()=>{const box=document.querySelector<HTMLDetailsElement>('.population-workspace:not([hidden]) .population-count-editor')!;for(let e:HTMLElement|null=box;e;e=e.parentElement)if(e instanceof HTMLDetailsElement)e.open=true;return box};
 if(phase==='initial'){open();document.documentElement.dataset.contextualBaseline=document.querySelector<HTMLInputElement>('[aria-label="Baseline capture ID"]')!.value;button('Run population view').click()}
 else if(phase==='reading'){open();button('Remove count comparison').click()}
 else if(phase==='removed'){open();button('Load baseline captures').click()}
 else if(phase==='loading'){const pick=document.querySelector<HTMLSelectElement>('[aria-label="Baseline capture"]')!;pick.value=document.documentElement.dataset.contextualBaseline!;pick.dispatchEvent(new Event('change',{bubbles:true}))}
 else if(phase==='chosen')button('Prepare baseline').click();
 document.documentElement.dataset.contextualPhase=({initial:'reading',reading:'removed',removed:'loading',loading:'chosen',chosen:'preparing',preparing:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'&&s.editors===1?[step('initial')]:s.phase==='reading'&&s.pair&&s.baseline==='2'&&s.current==='3'?[step('reading')]:s.phase==='removed'&&!s.pair&&s.editors===1?[step('removed')]:s.phase==='loading'&&s.choice?[step('loading')]:s.phase==='chosen'&&s.prepare?[step('chosen')]:s.phase==='preparing'&&s.pair&&s.compares===2&&s.baseline==='2'&&s.current==='3'?[step('preparing')]:['Wait']});
