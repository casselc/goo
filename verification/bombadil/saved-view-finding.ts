import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({version:'goo.finding282',phase:document.documentElement.dataset.findingPhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),reads:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore').length,cards:document.querySelectorAll('#documentCards>[data-kind="view"]').length,reference:document.querySelector<HTMLElement>('#documentCards>[data-kind="view"]')?.dataset.reference,saveReady:!document.querySelector<HTMLButtonElement>('#documentSave')?.disabled,stored:document.querySelector('#documentRevision')?.textContent?.includes('stored revision'),overflow:document.documentElement.scrollWidth>window.innerWidth}));
export const complete=eventually(()=>state.current.phase==='complete').within(25,'seconds');
export const budget=always(()=>state.current.reads<=2);
export const definitionIdentity=always(()=>!state.current.cards||/^[a-f0-9]{64}$/.test(state.current.reference||''));
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.savedViewFinding',async(document,window,phase:string)=>{
 const button=(name:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 if(phase==='initial'||phase==='definition')button('Run population view').click();
 else if(phase==='read'){const s=[...document.querySelectorAll<HTMLElement>('summary')].find(e=>e.textContent==='Preserve')!;(s.parentElement as HTMLDetailsElement).open=true;const text=document.querySelector<HTMLTextAreaElement>('.population-save textarea[aria-label="Saved view finding"]')!;text.value='Captured count pattern warrants follow-up; definition only.';text.dispatchEvent(new Event('input',{bubbles:true}));button('Add saved view to investigation').click()}
 else if(phase==='attach')document.querySelector<HTMLButtonElement>('#documentSave')!.click();
 else if(phase==='savedoc')document.querySelector<HTMLAnchorElement>('#documentCards>[data-kind="view"] a')!.click();
 document.documentElement.dataset.findingPhase=({initial:'read',read:'attach',attach:'savedoc',savedoc:'definition',definition:'executing',executing:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='read'&&s.population==='applied'?[step('read')]:s.phase==='attach'&&s.cards===1&&s.saveReady?[step('attach')]:s.phase==='savedoc'&&s.stored?[step('savedoc')]:s.phase==='definition'&&s.population==='staged'?[step('definition')]:s.phase==='executing'&&s.population==='applied'&&s.reads===2?[step('executing')]:['Wait']});
