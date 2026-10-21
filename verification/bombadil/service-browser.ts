import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{
 const el=(id:string)=>document.getElementById(id);let evidence:any=null;try{evidence=JSON.parse(el('serviceDetail')?.querySelector('.service-source-evidence pre')?.textContent||'null')}catch{}
 const inventory=(()=>{try{return JSON.parse(el('serviceDiscoveryExplain')?.textContent||'null')}catch{return null}})();
 const coherent=!evidence||evidence.every((r:any)=>r.question.environment===inventory?.question.environment&&r.question.fromMs===inventory?.question.fromMs&&r.question.toMs===inventory?.question.toMs);
 return {stage:document.documentElement.dataset.serviceBrowserStage||'initial',ready:!!el('inventoryRun'),inventory:!!inventory,loading:!!el('serviceDetail')?.querySelector('.service-loading'),detail:el('serviceDetail')?.hidden===false,evidence:!!evidence,coherent,returning:el('serviceReturn')?.hidden===false,overflow:document.documentElement.scrollWidth>innerWidth,unsupported:[...document.querySelectorAll<HTMLButtonElement>('#serviceDetail button')].filter(b=>b.textContent==='Unsupported by current kernels').every(b=>b.disabled)};
});
export const complete=eventually(()=>state.current.stage==='complete').within(25,'seconds');
export const fixedServiceScope=always(()=>state.current.coherent);
export const unsupportedCannotExecute=always(()=>state.current.unsupported);
export const noOverflow=always(()=>!state.current.overflow);
const act=registerCustomAction('goo.serviceBrowser',async(document,window,command:string)=>{
 const el=(id:string)=>document.getElementById(id),button=(text:string)=>[...document.querySelectorAll<HTMLButtonElement>('#serviceDetail button, #serviceReturn button, #serviceCatalog button')].find(b=>b.textContent===text);
 if(command==='start'){const environment=new URLSearchParams(window.location.hash.slice(1)).get('serviceBrowser')!;document.querySelector<HTMLButtonElement>('[data-workspace="services"]')!.click();const input=el('inventoryEnvironment') as HTMLInputElement;input.value=environment;input.dispatchEvent(new Event('input',{bubbles:true}));el('inventoryRun')!.click()}
 if(command==='browse')button('Open service catalog')!.click();if(command==='follow'){(document.querySelector('.service-question-grid')as HTMLDetailsElement).open=true;button('Explore http')!.click()};if(command==='return')button('Back to service')!.click();
 document.documentElement.dataset.serviceBrowserStage=command==='return'?'complete':command;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready||s.loading)return ['Wait'];if(s.stage==='initial')return[act('start')];if(s.stage==='start')return s.inventory?[act('browse')]:['Wait'];if(s.stage==='browse')return s.evidence?[act('follow')]:['Wait'];if(s.stage==='follow')return s.returning?[act('return')]:['Wait'];return ['Wait']});
