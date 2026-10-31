import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const element=document.querySelector<any>('[data-composition-role=comparison] goo-group-count-comparison'),g=element?.comparison,resources=window.performance.getEntriesByType('resource'),count=(path:string)=>resources.filter(e=>new URL(e.name).pathname===path).length;
 return {version:'goo.groups291',phase:document.documentElement.dataset.groupsPhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),full:count('/api/explore'),pages:count('/api/explore/page'),compares:count('/api/compare'),baselines:resources.filter(e=>/^\/api\/captures\/[a-f0-9]{64}$/.test(new URL(e.name).pathname)).length,saves:count('/api/exploration-views'),pair:!!g,baseline:g?.baseline.rows[0].values.count,current:g?.selection.rows[0].values.count,rows:g?.rows.length,domain:element?.dataset.domainMax,visibleRows:element?.querySelectorAll('tbody tr').length,find:element?.querySelector<HTMLInputElement>('[aria-label="Find compared groups"]')?.value,offset:document.querySelector<HTMLElement>('.population-record-page-facts')?.dataset.pageOffset,held:!document.documentElement.dataset.groupsHeld||!g||JSON.stringify(g)===document.documentElement.dataset.groupsHeld,saved:document.querySelector<HTMLAnchorElement>('.population-save-status a')?.getAttribute('href'),overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const sourceBudget=always(()=>state.current.full<=1&&state.current.pages<=1&&state.current.compares<=1&&state.current.baselines<=1&&state.current.saves<=1);
export const heldGroups=always(()=>state.current.held);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.groupCount',async(document,window,phase:string)=>{
 const button=(text:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===text)!;
 const pick=(label:string,value:string)=>{const e=document.querySelector<HTMLSelectElement>('[aria-label="'+label+'"]')!;e.value=value;e.dispatchEvent(new Event('change',{bubbles:true}))};
 const find=(value:string)=>{const e=document.querySelector<HTMLInputElement>('goo-group-count-comparison [aria-label="Find compared groups"]')!;e.value=value;e.dispatchEvent(new Event('input',{bubbles:true}))};
 if(phase==='initial')button('Run population view').click();
 else if(phase==='reading'){document.documentElement.dataset.groupsHeld=JSON.stringify(document.querySelector<any>('goo-group-count-comparison').comparison);find('shared');pick('Group comparison order','label')}
 else if(phase==='finding'){pick('Comparison view','summary');pick('Comparison view','groups')}
 else if(phase==='toggling')button('Next spans page').click();
 else if(phase==='paged'){find('');const menu=document.querySelector<HTMLDetailsElement>('.population-save')!;menu.open=true;const input=document.querySelector<HTMLInputElement>('[aria-label="Exploration view title"]')!;input.value='Group comparison campaign '+Date.now();input.dispatchEvent(new Event('input',{bubbles:true}));button('Save exploration view').click()}
 else if(phase==='saving')window.location.hash=document.querySelector<HTMLAnchorElement>('.population-save-status a')!.hash;
 document.documentElement.dataset.groupsPhase=({initial:'reading',reading:'finding',finding:'toggling',toggling:'paged',paged:'saving',saving:'reopened',reopened:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='reading'&&s.pair&&s.baseline==='61'&&s.current==='64'&&s.rows===61&&s.domain==='4'?[step('reading')]:s.phase==='finding'&&s.visibleRows===1&&s.find==='shared'&&s.domain==='4'?[step('finding')]:s.phase==='toggling'&&s.pair&&s.held?[step('toggling')]:s.phase==='paged'&&s.pages===1&&s.offset==='1'&&s.held&&s.find==='shared'?[step('paged')]:s.phase==='saving'&&s.saves===1&&!!s.saved?[step('saving')]:s.phase==='reopened'&&s.population==='staged'&&!s.pair?[step('reopened')]:['Wait']});
