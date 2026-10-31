import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const pop=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(e=>!e.hidden&&e.getClientRects().length>0),element=document.querySelector<any>('goo-group-count-comparison'),g=element?.comparison,resources=window.performance.getEntriesByType('resource'),count=(path:string)=>resources.filter(e=>new URL(e.name).pathname===path).length;
 return {version:'goo.groupSide292',phase:document.documentElement.dataset.sidePhase||'initial',population:pop?.dataset.phase,childCount:pop?.querySelector('.population-count')?.textContent?.match(/^(\d+) matching/)?.[1],child:!!document.querySelector('.population-child'),full:count('/api/explore'),pages:count('/api/explore/page'),compares:count('/api/compare'),baselines:resources.filter(e=>/^\/api\/captures\/[a-f0-9]{64}$/.test(new URL(e.name).pathname)).length,pair:!!g,baseline:g?.baseline.rows[0].values.count,current:g?.selection.rows[0].values.count,find:element?.querySelector<HTMLInputElement>('[aria-label="Find compared groups"]')?.value,held:!document.documentElement.dataset.sideHeld||!g||JSON.stringify(g)===document.documentElement.dataset.sideHeld,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const readBudget=always(()=>state.current.full<=2&&state.current.pages===0&&state.current.compares<=1&&state.current.baselines<=1);
export const heldParent=always(()=>state.current.held);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.groupSide',async(document,_window,phase:string)=>{
 const visible=(e:Element)=>e.getClientRects().length>0,button=(text:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===text&&visible(b))!;
 const find=(value:string)=>{const e=document.querySelector<HTMLInputElement>('goo-group-count-comparison [aria-label="Find compared groups"]')!;e.value=value;e.dispatchEvent(new Event('input',{bubbles:true}))};
 const choose=(side:string)=>{const row=document.querySelector('goo-group-count-comparison tr[data-group-key="1shared"]')!,b=row.querySelector<HTMLButtonElement>('td[data-side='+side+'] button')!;b.click()};
 if(phase==='initial')([...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>/^(Run population view|Run selected population|Rerun fixed window)$/.test(b.textContent||'')&&visible(b))!).click();
 else if(phase==='reading'){document.documentElement.dataset.sideHeld=JSON.stringify(document.querySelector<any>('goo-group-count-comparison').comparison);find('shared')}
 else if(phase==='finding')choose('baseline');
 else if(phase==='staging')([...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>/^(Run population view|Run selected population|Rerun fixed window)$/.test(b.textContent||'')&&visible(b))!).click();
 else if(phase==='childRead')button('Back to parent population').click();
 else if(phase==='returned')choose('selection');
 else if(phase==='currentStaged')button('Back to parent population').click();
 document.documentElement.dataset.sidePhase=({initial:'reading',reading:'finding',finding:'staging',staging:'childRead',childRead:'returned',returned:'currentStaged',currentStaged:'backAgain',backAgain:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='reading'&&s.pair&&s.baseline==='5'&&s.current==='7'?[step('reading')]:s.phase==='finding'&&s.find==='shared'?[step('finding')]:s.phase==='staging'&&s.child&&s.population==='staged'&&s.full===1?[step('staging')]:s.phase==='childRead'&&s.child&&s.population==='applied'&&s.full===2&&s.childCount==='3'?[step('childRead')]:s.phase==='returned'&&!s.child&&s.find==='shared'&&s.held?[step('returned')]:s.phase==='currentStaged'&&s.child&&s.population==='staged'&&s.full===2?[step('currentStaged')]:s.phase==='backAgain'&&!s.child&&s.held?[step('backAgain')]:['Wait']});
