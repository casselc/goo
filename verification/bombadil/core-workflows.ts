// Read-only live exploration: reopen a question, reuse a definition, discover
// signals/parent links, then cancel. Never saves definitions/evidence or emits work.
import { always,eventually } from '@antithesishq/bombadil';
import { extract,actions,registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions,noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{
 const el=(id:string)=>document.getElementById(id),read=(id:string)=>{try{return JSON.parse(el(id)?.textContent||'null')}catch{return null}};
 const receipt=read('captureGeneralReceipt'),e=read('exploreExplain'),d=read('serviceDiscoveryExplain'),fixed=receipt?.receipt?.question?.explore,stage=document.documentElement.dataset.gooWorkflowStage||'initial';
 const fixedValid=!['fixed','library'].includes(stage)||!!e&&!!fixed&&e.question.environment===fixed.environment&&e.question.fromMs===fixed.fromMs&&e.question.toMs===fixed.toMs&&JSON.stringify(e.question.where)===JSON.stringify(fixed.where)&&e.duration?.count==='205'&&e.duration.p95Ns==='195000000';
 const reusableValid=stage!=='reused'||!!e&&e.question.environment===fixed?.environment&&e.question.summary==='span-duration'&&e.question.bucketMs===60000&&e.question.toMs-e.question.fromMs===300000&&e.question.fromMs!==fixed.fromMs&&e.duration?.count===el('exploreCount')?.querySelector('span')?.textContent;
 let discoveryValid=true;if(stage==='signals'){discoveryValid=!!d&&d.question.environment===fixed?.environment&&d.signals.some((x:any)=>x.service==='log-only'&&x.signal==='logs')&&d.dependencies.filter((x:any)=>x.kind==='edge').length===1&&d.dependencies.some((x:any)=>x.kind==='edge'&&x.parent==='api'&&x.child==='worker'&&x.count==='1')&&d.dependencies.some((x:any)=>x.kind==='ambiguous'&&x.count==='1')&&d.dependencies.some((x:any)=>x.kind==='unresolved'&&x.count==='1')&&d.signals.every((x:any)=>x.firstMs>=d.question.fromMs&&x.lastMs<d.question.toMs&&x.firstMs<=x.lastMs)}
 const cancelledValid=stage!=='cancelled'||!d&&!el('serviceDiscoveryRows')?.textContent&&(el('inventoryStatus')?.textContent||'').includes('cancelled');
 const selected=[...document.querySelectorAll('[data-workspace][aria-current="page"]')].map(x=>x.getAttribute('data-workspace')),visible=[...document.querySelectorAll<HTMLElement>('[data-workspace-panel]')].filter(x=>!x.hidden).map(x=>x.dataset.workspacePanel);
 return {version:'goo.core-workflows.v1',stage,ready:!!fixed,environment:fixed?.environment,opened:!!e,reused:!!e&&e.question.fromMs!==fixed?.fromMs&&!((el('exploreRun') as HTMLButtonElement)?.disabled),hasQuestion:[...document.querySelectorAll('#questionLibraryRows button')].some(x=>x.textContent==='Run '+fixed?.environment),discovered:!!d,fixedValid,reusableValid,discoveryValid,cancelledValid,navigation:selected.length===1&&visible.length===1&&selected[0]===visible[0]};
});
export const ready=eventually(()=>state.current.ready).within(10,'seconds');
export const complete=eventually(()=>state.current.stage==='complete').within(25,'seconds');
export const frozenQuestionHandoff=always(()=>state.current.fixedValid);
export const reusableMeasureScope=always(()=>state.current.reusableValid);
export const honestServiceLinks=always(()=>state.current.discoveryValid);
export const cancellationClearsEvidence=always(()=>!!state.current.cancelledValid);
export const coherentWorkspace=always(()=>!state.current.ready||state.current.navigation);
const act=registerCustomAction('goo.core-workflows',async(document,_window,command:string,next?:string)=>{
 const el=(id:string)=>document.getElementById(id),click=(id:string)=>el(id)?.click(),nav=(route:string)=>document.querySelector<HTMLElement>('[data-workspace="'+route+'"]')?.click(),set=(id:string,value:string,event='input')=>{const input=el(id) as HTMLInputElement;input.value=value;input.dispatchEvent(new Event(event,{bubbles:true}))};
 const fixed=JSON.parse(el('captureGeneralReceipt')?.textContent||'null')?.receipt?.question?.explore;
 if(command==='live')click('captureGeneralLive');
 if(command==='library'){const details=el('questionLibraryRows')?.closest('details');if(details)details.open=true;click('questionLibraryRefresh');set('questionLibrarySearch',fixed.environment)}
 if(command==='reuse'){const button=[...document.querySelectorAll<HTMLButtonElement>('#questionLibraryRows button')].find(x=>x.textContent==='Run '+fixed.environment);button?.click()}
 if(command==='discover'){nav('services');set('minutes','10','change');set('inventoryEnvironment',fixed.environment);click('inventoryRun')}
 if(command==='cancel'){click('inventoryRun');click('inventoryCancel')}
 if(command==='explore')nav('explore');if(command==='services')nav('services');
 if(next)document.documentElement.dataset.gooWorkflowStage=next;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready)return['Wait'];if(s.stage==='initial')return[act('live','opening')];if(s.stage==='opening')return s.opened?[act('noop','fixed')]:['Wait'];if(s.stage==='fixed')return[act('library','library')];if(s.stage==='library')return s.hasQuestion?[act('reuse','reusing')]:['Wait'];if(s.stage==='reusing')return s.reused?[act('noop','reused')]:['Wait'];if(s.stage==='reused')return[act('discover','discovering')];if(s.stage==='discovering')return s.discovered?[act('noop','signals')]:['Wait'];if(s.stage==='signals')return[act('cancel','cancelled')];if(s.stage==='cancelled')return[act('noop','complete')];return[act('explore'),act('services')]});
