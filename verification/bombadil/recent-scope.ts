import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({version:'goo.recentScope285',phase:document.documentElement.dataset.recentScopePhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),reads:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore').length,plans:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/population/plan').length,saved:!!document.querySelector('.population-workspace:not([hidden]) .population-save-status a'),overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const complete=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const budget=always(()=>state.current.reads<=6&&state.current.plans<=1);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.recentScope',async(document,window,phase:string)=>{
 const pop=()=>document.querySelector<HTMLElement>('.population-workspace:not([hidden])')!,button=(name:string)=>[...(pop()?.querySelectorAll<HTMLButtonElement>('button')||[])].find(b=>b.textContent===name)!,open=(name:string)=>{const s=[...pop().querySelectorAll<HTMLElement>('summary')].find(s=>s.textContent===name);if(s)(s.parentElement as HTMLDetailsElement).open=true};
 if(phase==='initial'){const minutes=document.querySelector<HTMLInputElement>('#minutes')!;minutes.value='5';minutes.dispatchEvent(new Event('input',{bubbles:true}));document.querySelector<HTMLButtonElement>('#exploreRecent')!.click()}
 else if(phase==='recent'){open('Select a time range');const input=pop().querySelector<HTMLInputElement>('goo-time-range input[aria-label="Range start (UTC)"]')!;input.value=String(Math.round((Number(input.min)+Number(input.max))/2));input.dispatchEvent(new Event('input',{bubbles:true}))}
 else if(phase==='select')button('Open selected population').click();
 else if(phase==='child'||phase==='reopen'){[...pop().querySelectorAll<HTMLButtonElement>('button')].find(b=>/^Run (population view|selected population)$/.test(b.textContent||''))!.click()}
 else if(phase==='execute'){open('View settings');open('Preserve');const title=pop().querySelector<HTMLInputElement>('[aria-label="Exploration view title"]')!;title.value='Bombadil recent typed scope';title.dispatchEvent(new Event('input',{bubbles:true}));button('Save exploration view').click()}
 else if(phase==='save')window.location.hash=pop().querySelector<HTMLAnchorElement>('.population-save-status a')!.hash;
 document.documentElement.dataset.recentScopePhase=({initial:'recent',recent:'select',select:'child',child:'execute',execute:'save',save:'reopen',reopen:'executing',executing:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'?[step('initial')]:s.phase==='recent'&&s.population==='applied'&&s.reads===2?[step('recent')]:s.phase==='select'?[step('select')]:s.phase==='child'&&s.population==='staged'?[step('child')]:s.phase==='execute'&&s.population==='applied'&&s.reads===4?[step('execute')]:s.phase==='save'&&s.saved?[step('save')]:s.phase==='reopen'&&s.population==='staged'?[step('reopen')]:s.phase==='executing'&&s.population==='applied'&&s.reads===6?[step('executing')]:['Wait']});
