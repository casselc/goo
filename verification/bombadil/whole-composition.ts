import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{const chart=document.querySelector<any>('.population-breakdown goo-chart'),root=document.querySelector('.population-composition'),facts=document.querySelector<HTMLElement>('[data-composition-role="records"] .population-record-page-facts'),apply=[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Apply composition');return {version:'goo.whole280',phase:document.documentElement.dataset.wholePhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),reads:window.performance.getEntriesByType('resource').filter(e=>['/api/explore','/api/explore/page'].includes(new URL(e.name).pathname)).length,composed:!!root,first:root?.firstElementChild?.getAttribute('data-composition-role'),records:!!facts,offset:facts?.dataset.pageOffset,whole:!!apply,blocked:apply?.disabled,selected:document.querySelectorAll('input[data-record-selection]:checked').length,inspect:!![...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.startsWith('Inspect display · Shared composition280-')),held:!chart||!document.documentElement.dataset.wholeBasis||chart.snapshot.revision===document.documentElement.dataset.wholeBasis,scopeOK:!chart||chart.snapshot.scope.selector.environment===new URL(window.location.href).searchParams.get('currentEnvironment'),saved:!!document.querySelector('.population-save-status a'),overflow:document.documentElement.scrollWidth>window.innerWidth}});
export const complete=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const readBudget=always(()=>state.current.reads<=(['page','save','reopen','complete'].includes(state.current.phase)?2:state.current.phase==='initial'?0:1));
export const heldInput=always(()=>state.current.held&&state.current.scopeOK);
export const selectedNotDiscarded=always(()=>state.current.phase!=='clearLive'||state.current.selected===1&&state.current.blocked===true);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.reuseWholeComposition',async(document,window,phase:string)=>{
 const button=(name:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 const open=(name:string)=>{const s=[...document.querySelectorAll<HTMLElement>('summary')].find(e=>e.textContent===name);if(s)(s.parentElement as HTMLDetailsElement).open=true};
 if(phase==='initial')button('Run population view').click();
 else if(phase==='read'){const chart=document.querySelector<any>('.population-breakdown goo-chart');document.documentElement.dataset.wholeBasis=chart.snapshot.revision;open('View settings');open('Saved exploration views');button('Load saved views').click()}
 else if(phase==='catalog')[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent?.startsWith('Inspect display · Shared composition280-'))!.click();
 else if(phase==='inspect'){const box=document.querySelector<HTMLInputElement>('input[data-record-selection]')!;box.checked=true;box.dispatchEvent(new Event('change'))}
 else if(phase==='selected')button('Clear live configuration').click();
 else if(phase==='clearLive')button('Clear selected records').click();
 else if(phase==='cleared')button('Apply composition').click();
 else if(phase==='apply'){open('View settings');open('Composition');const input=document.querySelector<HTMLInputElement>('input[aria-label="Show records"]')!;input.checked=true;input.dispatchEvent(new Event('change'))}
 else if(phase==='visible')button('Next spans page').click();
 else if(phase==='page'){open('Preserve');button('Save exploration view').click()}
 else if(phase==='save')window.location.hash=document.querySelector<HTMLAnchorElement>('.population-save-status a')!.hash;
 else if(phase==='reopen'){if(window.performance.getEntriesByType('resource').filter(e=>['/api/explore','/api/explore/page'].includes(new URL(e.name).pathname)).length!==2)throw Error('Reopening configuration read telemetry')}
 document.documentElement.dataset.wholePhase=({initial:'read',read:'catalog',catalog:'inspect',inspect:'selected',selected:'clearLive',clearLive:'cleared',cleared:'apply',apply:'visible',visible:'page',page:'save',save:'reopen',reopen:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='read'&&s.population==='applied'?[step('read')]:s.phase==='catalog'&&s.inspect?[step('catalog')]:s.phase==='inspect'&&s.whole&&s.blocked?[step('inspect')]:s.phase==='selected'&&s.selected===1?[step('selected')]:s.phase==='clearLive'&&s.selected===1?[step('clearLive')]:s.phase==='cleared'&&s.whole&&!s.blocked?[step('cleared')]:s.phase==='apply'&&s.composed&&!s.records?[step('apply')]:s.phase==='visible'&&s.records?[step('visible')]:s.phase==='page'&&s.offset==='1'?[step('page')]:s.phase==='save'&&s.saved?[step('save')]:s.phase==='reopen'&&s.population==='staged'?[step('reopen')]:['Wait']});
