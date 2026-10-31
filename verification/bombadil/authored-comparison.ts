import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const c=document.querySelector<any>('[data-composition-role=comparison] goo-measure-comparison')?.comparison,resources=window.performance.getEntriesByType('resource'),count=(path:string)=>resources.filter(e=>new URL(e.name).pathname===path).length;
 return {version:'goo.authored290',phase:document.documentElement.dataset.authoredPhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),full:count('/api/explore'),pages:count('/api/explore/page'),compares:count('/api/compare'),baselines:resources.filter(e=>/^\/api\/captures\/[a-f0-9]{64}$/.test(new URL(e.name).pathname)).length,saves:count('/api/exploration-views'),pair:!!c,baseline:c?.baseline.rows[0].values.count,current:c?.selection.rows[0].values.count,change:c?.changes.count.value,offset:document.querySelector<HTMLElement>('.population-record-page-facts')?.dataset.pageOffset,held:!document.documentElement.dataset.authoredHeld||!c||JSON.stringify(c)===document.documentElement.dataset.authoredHeld,saved:document.querySelector<HTMLAnchorElement>('.population-save-status a')?.getAttribute('href'),overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const sourceBudget=always(()=>state.current.full<=1&&state.current.pages<=1&&state.current.compares<=1&&state.current.baselines<=1&&state.current.saves<=1);
export const heldPair=always(()=>state.current.held&&(!['paged','saving'].includes(state.current.phase)||state.current.pair));
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.authoredComparison',async(document,window,phase:string)=>{
 const button=(text:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===text)!;
 if(phase==='initial')button('Run population view').click();
 else if(phase==='reading') {document.documentElement.dataset.authoredHeld=JSON.stringify(document.querySelector<any>('[data-composition-role=comparison] goo-measure-comparison').comparison);button('Next spans page').click()}
 else if(phase==='paged') {const menu=document.querySelector<HTMLDetailsElement>('.population-save')!;menu.open=true;const input=document.querySelector<HTMLInputElement>('[aria-label="Exploration view title"]')!;input.value='Authored campaign '+Date.now();input.dispatchEvent(new Event('input',{bubbles:true}));button('Save exploration view').click()}
 else if(phase==='saving'){window.location.hash=document.querySelector<HTMLAnchorElement>('.population-save-status a')!.hash;}
 document.documentElement.dataset.authoredPhase=({initial:'reading',reading:'paged',paged:'saving',saving:'reopened',reopened:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='reading'&&s.pair&&s.baseline==='2'&&s.current==='4'&&s.change==='2'?[step('reading')]:s.phase==='paged'&&s.offset==='1'&&s.pages===1&&s.pair&&s.held?[step('paged')]:s.phase==='saving'&&s.saves===1&&!!s.saved?[step('saving')]:s.phase==='reopened'&&s.population==='staged'&&!s.pair?[step('reopened')]:['Wait']});
