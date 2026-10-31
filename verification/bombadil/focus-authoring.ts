import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{const pop=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(e=>!e.hidden&&e.getClientRects().length);return {version:'goo.focusAuthoring286',phase:document.documentElement.dataset.focusAuthoringPhase||'initial',population:pop?.getAttribute('data-phase'),reads:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore').length,plans:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/population/plan').length,composed:!!pop?.querySelector('[data-composition-role=records]'),saved:!!pop?.querySelector('.population-save-status a'),overflow:document.documentElement.scrollWidth>window.innerWidth}});
export const complete=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const budget=always(()=>state.current.reads<=4&&state.current.plans<=1);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.focusAuthoring',async(document,window,phase:string)=>{
 const pop=()=>[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(e=>!e.hidden&&e.getClientRects().length)!,button=(name:string)=>[...pop().querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!,open=(name:string)=>{const s=[...pop().querySelectorAll<HTMLElement>('summary')].find(s=>s.textContent===name);if(s)(s.parentElement as HTMLDetailsElement).open=true};
 if(phase==='initial')document.querySelector<HTMLButtonElement>('#exploreRecent')!.click();
 else if(phase==='recent')button('Customize logs').click();
 else if(phase==='focus'||phase==='reopen')button('Run population view').click();
 else if(phase==='execute'){open('Save this exploration view');const title=pop().querySelector<HTMLInputElement>('[aria-label="Exploration view title"]')!;title.value='Bombadil focused logs';title.dispatchEvent(new Event('input',{bubbles:true}));button('Save exploration view').click()}
 else if(phase==='save')window.location.hash=pop().querySelector<HTMLAnchorElement>('.population-save-status a')!.hash;
 document.documentElement.dataset.focusAuthoringPhase=({initial:'recent',recent:'focus',focus:'execute',execute:'save',save:'reopen',reopen:'executing',executing:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'?[step('initial')]:s.phase==='recent'&&s.population==='applied'&&s.reads===2?[step('recent')]:s.phase==='focus'&&s.population==='staged'?[step('focus')]:s.phase==='execute'&&s.population==='applied'&&s.composed&&s.reads===3?[step('execute')]:s.phase==='save'&&s.saved?[step('save')]:s.phase==='reopen'&&s.population==='staged'?[step('reopen')]:s.phase==='executing'&&s.population==='applied'&&s.reads===4?[step('executing')]:['Wait']});
