import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({version:'goo.timeline281',phase:document.documentElement.dataset.timelinePhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),reads:window.performance.getEntriesByType('resource').filter(e=>['/api/explore','/api/explore/page'].includes(new URL(e.name).pathname)).length,bars:document.querySelectorAll('[data-composition-role="timeline"] .scope-time-interval').length,offset:document.querySelector<HTMLElement>('.population-record-page-facts')?.dataset.pageOffset,selected:!![...document.querySelectorAll<HTMLButtonElement>('[data-composition-role="timeline"] button')].find(b=>b.textContent==='Stage selected interval'&&!b.disabled),saved:!!document.querySelector('.population-save-status a'),overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const complete=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const readBudget=always(()=>state.current.reads<=(['narrow','reset','save','reopen','complete'].includes(state.current.phase)?3:['page','select','stage'].includes(state.current.phase)?2:1));
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.countTimeline',async(document,window,phase:string)=>{
 const button=(name:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 const open=(name:string)=>{const s=[...document.querySelectorAll<HTMLElement>('summary')].find(e=>e.textContent===name);if(s)(s.parentElement as HTMLDetailsElement).open=true};
 if(phase==='initial'||phase==='stage')button('Run population view').click();
 else if(phase==='read')button('Next spans page').click();
 else if(phase==='page')document.querySelector<HTMLButtonElement>('[data-composition-role="timeline"] .scope-time-interval')!.click();
 else if(phase==='select')button('Stage selected interval').click();
 else if(phase==='narrow')[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>/^Reset (timeline |starting )?window$/.test(b.textContent||''))!.click();
 else if(phase==='reset'){open('Preserve');button('Save exploration view').click()}
 else if(phase==='save')window.location.hash=document.querySelector<HTMLAnchorElement>('.population-save-status a')!.hash;
 document.documentElement.dataset.timelinePhase=({initial:'read',read:'page',page:'select',select:'stage',stage:'narrow',narrow:'reset',reset:'save',save:'reopen',reopen:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='read'&&s.population==='applied'&&s.bars>0?[step('read')]:s.phase==='page'&&s.offset==='1'?[step('page')]:s.phase==='select'&&s.selected?[step('select')]:s.phase==='stage'&&s.population==='staged'?[step('stage')]:s.phase==='narrow'&&s.population==='applied'?[step('narrow')]:s.phase==='reset'&&s.population==='staged'?[step('reset')]:s.phase==='save'&&s.saved?[step('save')]:s.phase==='reopen'&&s.population==='staged'?[step('reopen')]:['Wait']});
