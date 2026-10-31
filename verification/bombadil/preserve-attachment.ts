import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{const card=document.querySelector<HTMLElement>('#documentCards>[data-kind="view"]');return {version:'goo.attachment293',phase:document.documentElement.dataset.attachmentPhase||'initial',population:document.querySelector('.population-workspace:not([hidden])')?.getAttribute('data-phase'),reference:document.documentElement.dataset.attachmentReference,cards:document.querySelectorAll('#documentCards>[data-kind="view"]').length,exact:!!card&&card.dataset.reference===document.documentElement.dataset.attachmentReference&&card.querySelector<HTMLTextAreaElement>('textarea')?.value==='Independent responder preserved this exact definition.',stored:!!document.querySelector('#documentRevision')?.textContent?.includes('stored revision'),reads:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore').length,overflow:document.documentElement.scrollWidth>window.innerWidth}});
export const complete=eventually(()=>state.current.phase==='complete'&&state.current.exact&&state.current.stored).within(25,'seconds');
export const noSourceReads=always(()=>state.current.reads===0);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.preserveAttachment',async(document,window,phase:string)=>{
 const button=(name:string)=>[...document.querySelectorAll<HTMLButtonElement>('button')].find(e=>e.textContent===name)!;
 const summary=(name:string)=>[...document.querySelectorAll<HTMLElement>('summary')].find(e=>e.textContent===name)!;
 if(phase==='initial'){document.documentElement.dataset.attachmentReference=window.location.hash.slice('#exploration-view='.length);(summary('Preserve').parentElement as HTMLDetailsElement).open=true;(summary('Add to investigation').parentElement as HTMLDetailsElement).open=true;const input=document.querySelector<HTMLTextAreaElement>('.population-save textarea[aria-label="Saved view finding"]')!;input.value='Independent responder preserved this exact definition.';input.dispatchEvent(new Event('input',{bubbles:true}));button('Add saved view to investigation').click()}
 else if(phase==='attached')document.querySelector<HTMLButtonElement>('#documentSave')!.click();
 document.documentElement.dataset.attachmentPhase=({initial:'attached',attached:'saving',saving:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='attached'&&s.cards===1&&s.exact?[step('attached')]:s.phase==='saving'&&s.stored&&s.exact?[step('saving')]:['Wait']});
