import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{
 const el=(id:string)=>document.getElementById(id),read=(id:string)=>{try{return JSON.parse(el(id)?.textContent||'null')}catch{return null}};
 const d=read('serviceDiscoveryExplain'),body=el('serviceDetail')?.querySelector('details pre')?.textContent;let browse:any=null;try{browse=JSON.parse(body||'null')}catch{}
 const q=read('exploreExplain')?.question;
 return {ready:!!el('inventorySignals'),stage:document.documentElement.dataset.metricServicesStage||'initial',discovered:!!d,loaded:!!browse,discovery:d,
 coherent:!browse||browse.every((x:any)=>x.question.environment===d?.question.environment&&x.question.fromMs===d?.question.fromMs&&x.question.toMs===d?.question.toMs),
 recordSignal:!q||['spans','logs'].includes(q.signal),overflow:document.documentElement.scrollWidth>innerWidth,
 preview:el('measureStatus')?.textContent?.includes('Question validated'),returning:el('serviceReturn')?.hidden===false,unsupported:[...document.querySelectorAll<HTMLButtonElement>('#serviceDetail button')].filter(b=>b.textContent==='Unsupported by current kernels').every(b=>b.disabled)};
});
export const complete=eventually(()=>state.current.stage==='complete').within(25,'seconds');
export const fixedMetadataScope=always(()=>state.current.coherent);
export const noMetricRecordQuery=always(()=>state.current.recordSignal);
export const supportedActions=always(()=>state.current.unsupported);
export const noOverflow=always(()=>!state.current.overflow);
const act=registerCustomAction('goo.metricServices',async(document,window,command:string)=>{
 const el=(id:string)=>document.getElementById(id);const button=(host:string,text:string)=>[...document.querySelectorAll<HTMLButtonElement>(host+' button')].find(b=>b.textContent===text);
 if(command==='start'){const env=new URLSearchParams(window.location.hash.slice(1)).get('metricServices')!;document.querySelector<HTMLButtonElement>('[data-workspace="services"]')!.click();const input=el('inventoryEnvironment') as HTMLInputElement;input.value=env;input.dispatchEvent(new Event('input',{bubbles:true}));el('inventorySignals')!.click()}
 if(command==='open')button('#serviceDiscoveryRows','Open service metrics-worker')!.click();
 if(command==='preview'){const card=[...document.querySelectorAll<HTMLElement>('#serviceDetail article')].find(c=>c.querySelector('h4')?.textContent==='worker.depth');card?.querySelector<HTMLButtonElement>('button')!.click()}
 if(command==='return')button('#serviceReturn','Back to service')!.click();
 document.documentElement.dataset.metricServicesStage=command==='return'?'complete':command;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready)return['Wait'];if(s.stage==='initial')return[act('start')];if(s.stage==='start')return s.discovered?[act('open')]:['Wait'];if(s.stage==='open')return s.loaded?[act('preview')]:['Wait'];if(s.stage==='preview')return s.preview&&s.returning?[act('return')]:['Wait'];return['Wait']});
