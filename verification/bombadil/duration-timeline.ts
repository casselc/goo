import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({version:'goo.duration287',phase:document.documentElement.dataset.durationPhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),reads:window.performance.getEntriesByType('resource').filter(e=>['/api/explore','/api/explore/page'].includes(new URL(e.name).pathname)).length,measure:document.querySelector<HTMLSelectElement>('[aria-label="Timeline measure"]')?.value,kind:document.querySelector<HTMLSelectElement>('[aria-label="Timeline view"]')?.value,bars:document.querySelectorAll('[data-composition-role="timeline"] .scope-time-interval').length,offset:document.querySelector<HTMLElement>('.population-record-page-facts')?.dataset.pageOffset,selected:!![...document.querySelectorAll<HTMLButtonElement>('[data-composition-role="timeline"] button')].find(b=>b.textContent==='Stage selected interval'&&!b.disabled),saved:!!document.querySelector('.population-save-status a'),overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const complete=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const budget=always(()=>state.current.reads<=4);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.durationTimeline',async(document,window,phase:string)=>{
 const button=(name:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 const choose=(label:string,value:string)=>{const s=document.querySelector<HTMLSelectElement>('[aria-label="'+label+'"]')!;s.value=value;s.dispatchEvent(new Event('change',{bubbles:true}))};
 if(phase==='initial'||phase==='stage'||phase==='reopen')button('Run population view').click();
 else if(phase==='read')choose('Timeline measure','count');
 else if(phase==='count')choose('Timeline measure','duration.mean');
 else if(phase==='mean')choose('Timeline measure','duration.p95');
 else if(phase==='p95')choose('Timeline view','table');
 else if(phase==='table')choose('Timeline view','time-bar');
 else if(phase==='bars')button('Next spans page').click();
 else if(phase==='page')document.querySelector<HTMLButtonElement>('[data-composition-role="timeline"] .scope-time-interval')!.click();
 else if(phase==='preview')button('Stage selected interval').click();
 else if(phase==='narrow'){const s=[...document.querySelectorAll<HTMLElement>('summary')].find(e=>e.textContent==='Preserve')!;(s.parentElement as HTMLDetailsElement).open=true;button('Save exploration view').click()}
 else if(phase==='save')window.location.hash=document.querySelector<HTMLAnchorElement>('.population-save-status a')!.hash;
 document.documentElement.dataset.durationPhase=({initial:'read',read:'count',count:'mean',mean:'p95',p95:'table',table:'bars',bars:'page',page:'preview',preview:'stage',stage:'narrow',narrow:'save',save:'reopen',reopen:'executing',executing:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='read'&&s.population==='applied'&&s.bars>0?[step('read')]:s.phase==='count'&&s.measure==='count'?[step('count')]:s.phase==='mean'&&s.measure==='duration.mean'?[step('mean')]:s.phase==='p95'&&s.measure==='duration.p95'?[step('p95')]:s.phase==='table'&&s.kind==='table'?[step('table')]:s.phase==='bars'&&s.bars>0?[step('bars')]:s.phase==='page'&&s.offset==='1'?[step('page')]:s.phase==='preview'&&s.selected?[step('preview')]:s.phase==='stage'&&s.population==='staged'?[step('stage')]:s.phase==='narrow'&&s.population==='applied'?[step('narrow')]:s.phase==='save'&&s.saved?[step('save')]:s.phase==='reopen'&&s.population==='staged'?[step('reopen')]:s.phase==='executing'&&s.population==='applied'&&s.reads===4?[step('executing')]:['Wait']});
