// Stored-page navigation only; never recaptures, refreshes telemetry or writes.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({
 stage:Number(document.documentElement.dataset.savedCatalogStage||0),
 rows:document.querySelectorAll('#captureCatalog tbody>tr').length,
 selected:!!document.querySelector('#generalCaptureWorkspace:not([hidden])'),
 failed:document.querySelector('#captureSavedStatus')?.textContent?.startsWith('Capture unavailable')||false,
 overflow:document.documentElement.scrollWidth>window.innerWidth,
 source:window.performance.getEntriesByType('resource').filter(e=>/^\/api\/(?:tail|explore|query|measurements|inventory|discovery|operations|counter-rates|histograms|http-package)$/.test(new URL(e.name).pathname)).length,
 typed:[...document.querySelectorAll('#captureCatalog tbody>tr')].every(row=>/^[a-f0-9]{64}$/.test((row as HTMLElement).dataset.captureId||'')&&row.querySelectorAll('td').length===3&&!!row.querySelector('[data-capture-col=kind]')?.textContent),
 clocks:[...document.querySelectorAll<HTMLTimeElement>('#captureCatalog time')].every(e=>e.textContent?.includes('UTC')&&Number.isFinite(Date.parse(e.dateTime))&&e.title==='Captured at '+e.dateTime),
}));
export const storedOnly=always(()=>state.current.source===0);
export const compactTypedRows=always(()=>!state.current.overflow&&state.current.typed&&state.current.clocks);
export const completes=eventually(()=>state.current.stage===7).within(20,'seconds');
const act=registerCustomAction('goo.savedCatalog',async(document,_window,step:number)=>{
 if(step===1)document.querySelector<HTMLButtonElement>('#captureCatalog tbody>tr button')!.click();
 if(step===2)document.querySelector<HTMLButtonElement>('[data-saved-open=evidence]')!.click();
 if(step===3)document.querySelector<HTMLElement>('#captureIDTools>summary')!.click();
 if(step===4)document.querySelector<HTMLInputElement>('#captureGeneralID')!.focus();
 if(step===5)document.querySelector<HTMLElement>('#captureIDTools>summary')!.click();
 if(step===6)document.querySelector<HTMLElement>('#captureLibraryPanel>summary')!.click();
 if(step===7)document.querySelector<HTMLElement>('#captureLibraryPanel>summary')!.click();
 document.documentElement.dataset.savedCatalogStage=String(step);
});
export const choices=actions(()=>{
 const s=state.current;if(!s.rows)return ['Wait'];if(s.stage===1&&!s.selected&&!s.failed)return ['Wait'];return s.stage<7?[act(s.stage+1)]:['Wait'];
});
