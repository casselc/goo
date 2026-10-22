import{always,eventually}from'@antithesishq/bombadil';
import{extract,actions,registerCustomAction}from'@antithesishq/bombadil/browser';
export{noUncaughtExceptions,noUnhandledPromiseRejections}from'@antithesishq/bombadil/browser/defaults/properties';
const s=extract(({document,window})=>({phase:document.querySelector('.stream-entry .live-tail')?.getAttribute('data-state'),retained:Number(document.querySelector('.stream-entry .live-tail')?.getAttribute('data-retained')||0),overflow:document.documentElement.scrollWidth>window.innerWidth,stage:Number(document.documentElement.dataset.entryStage||0)}));
export const bounded=always(()=>s.current.retained<=100&&!s.current.overflow);
export const pauses=eventually(()=>s.current.stage===1&&s.current.phase==='paused').within(20,'seconds');
const pause=registerCustomAction('goo.pauseEntry',async(document)=>{document.querySelector<HTMLButtonElement>('.stream-entry .live-tail button')!.click();document.documentElement.dataset.entryStage='1'});
export const choices=actions(()=>s.current.stage===0&&['unknown','partial'].includes(s.current.phase||'')?[pause()]:['Wait']);
