// Saved-only measure exploration: no source reads or receipt mutations.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{
 const el=(id:string)=>document.getElementById(id);let pair:any=null;try{pair=JSON.parse(el('paircohortExplain')?.textContent||'null')}catch{}
 const stage=document.documentElement.dataset.gooChangesStage||'initial',metric=(el('paircomparisonMeasure') as HTMLSelectElement)?.value,signature=pair?JSON.stringify(pair):'',golden=document.documentElement.dataset.gooChangesGolden;
 const m=pair?.result?.measures,coherent=!pair||pair.request.baseline.count==='80'&&pair.request.selection.count==='124'&&pair.request.selection.spanStatus.error==='20'&&pair.request.selection.spanStatus.unset==='44'&&m.errorDifferencePP==='13.63'&&m.p95DifferenceNs==='90000000';const summary=el('paircohortMeasures')?.textContent||'';
 const display=!pair||metric==='share'||metric==='mean'||metric==='p99'||metric==='p95'&&summary.includes('+90.000 ms')||metric==='error-rate'&&summary.includes('+13.63 pp');
 return {version:'goo.changes.v1',stage,ready:!!pair,coherent,display,stable:!golden||!pair||golden===signature,aligned:[...document.querySelectorAll('#paircohortValues .comparison-bar-cell')].every(e=>{if(e.getAttribute('data-scale-max')!==document.querySelector('#paircohortValues .comparison-bar-cell')?.getAttribute('data-scale-max'))return false;const fill=e.querySelector<HTMLElement>('.comparison-bar-fill');return !fill||parseFloat(fill.style.width)>=0&&parseFloat(fill.style.width)<=100}),metric,signature};
});
export const ready=eventually(()=>state.current.ready).within(10,'seconds');
export const complete=eventually(()=>state.current.stage==='complete').within(25,'seconds');
export const fullEvidence=always(()=>state.current.coherent);
export const correctDisplay=always(()=>state.current.display);
export const stableReceipt=always(()=>state.current.stable);
export const sharedVisualScales=always(()=>state.current.aligned);
const act=registerCustomAction('goo.changes',async(document,_window,command:string,next?:string)=>{
 const el=(id:string)=>document.getElementById(id);
 if(command==='remember')document.documentElement.dataset.gooChangesGolden=el('paircohortExplain')?.textContent?JSON.stringify(JSON.parse(el('paircohortExplain')!.textContent!)):'';
 if(['share','mean','p95','p99','error-rate'].includes(command)){const select=el('paircomparisonMeasure') as HTMLSelectElement;select.value=command;select.dispatchEvent(new Event('change',{bubbles:true}))}
 if(next)document.documentElement.dataset.gooChangesStage=next;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready)return['Wait'];if(s.stage==='initial')return[act('remember','remembered')];if(s.stage==='remembered')return[act('error-rate','errors')];if(s.stage==='errors')return[act('p95','latency')];if(s.stage==='latency')return[act('mean','mean')];if(s.stage==='mean')return[act('share','complete')];return['share','mean','p95','p99','error-rate'].map(x=>act(x))});
