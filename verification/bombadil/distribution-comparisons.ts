import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
// Dedicated real exploration view emitted by journey176. Counts/share/brush are
// local; exactly one explicit two-side read and one side continuation are allowed.
const state=extract(({document})=>{
 const root=document.documentElement,plot=document.querySelector<any>('.population-breakdown goo-aligned-histogram'),comparison=plot?.comparison,display=plot?.display,sides=plot?[...plot.querySelectorAll<any>('goo-time-bars')].map(e=>e.data):[],last=root.dataset.dcLastQuery?JSON.parse(root.dataset.dcLastQuery):null,parent=root.dataset.dcParent?JSON.parse(root.dataset.dcParent):null;
 return {phase:root.dataset.dcPhase||'initial',reads:Number(root.dataset.dcReads||0),plans:Number(root.dataset.dcPlans||0),staged:document.querySelector('.population-workspace')?.getAttribute('data-phase')==='staged',ready:plot?.dataset.state==='observed',mode:plot?.dataset.scale||'',sameEvidence:!root.dataset.dcGolden||!!comparison&&Object.isFrozen(comparison)&&Object.isFrozen(comparison.baseline)&&Object.isFrozen(comparison.selection)&&JSON.stringify(comparison)===root.dataset.dcGolden,
 aligned:sides.length!==2||JSON.stringify(sides[0].points.map((p:any)=>p.key))===JSON.stringify(sides[1].points.map((p:any)=>p.key))&&sides[0].ceiling===sides[1].ceiling,
 shareScale:display?.scale!=='population-share'||sides.every(s=>s.ceiling==='100'),
 childStaged:document.querySelector('.population-child .population-workspace')?.getAttribute('data-phase')==='staged',count:document.querySelector('.population-child .population-count')?.textContent||document.getElementById('exploreCount')?.textContent||'',intent:document.querySelector('.numeric-distribution-comparison [role=status]')?.textContent||'',
 baselineDisabled:[...document.querySelectorAll<HTMLButtonElement>('.numeric-distribution-comparison button')].find(b=>b.textContent==='Open baseline range')?.disabled,
 originalSide:!last||!parent||last.environment===parent.environment&&last.fromMs===parent.fromMs&&last.toMs===parent.toMs&&JSON.stringify(last.where).includes(JSON.stringify(parent.where)),
 };
});
export const completed=eventually(()=>state.current.phase==='complete'&&state.current.intent.includes('[8, 32)')).within(35,'seconds');
export const boundedExplicitReads=always(()=>state.current.reads<=3&&state.current.plans<=1);
export const initialConfigurationSourceFree=always(()=>!['initial','initialized'].includes(state.current.phase)||state.current.reads===0&&state.current.plans===0);
export const togglesAndBrushSourceFree=always(()=>!['count','share','keyboard'].includes(state.current.phase)||state.current.reads===2&&state.current.plans===0);
export const sidePlanningSourceFree=always(()=>state.current.phase!=='planned'||state.current.reads===2&&state.current.plans===1);
export const sharedBinsAndScales=always(()=>state.current.aligned&&state.current.shareScale);
export const suppliedEvidenceFixed=always(()=>state.current.sameEvidence);
export const requestedSideRetainsItsScope=always(()=>state.current.originalSide);
export const completionRequiresActualContinuation=always(()=>state.current.phase!=='complete'||state.current.reads===3&&state.current.plans===1&&state.current.baselineDisabled===true);
const step=registerCustomAction('goo.distributionComparison',async(document,window,phase:string)=>{
 const root=document.documentElement,globals=window as any,host=()=>document.querySelector<HTMLElement>('.population-breakdown .numeric-distribution-comparison')!;
 if(phase==='initialized'){root.dataset.dcReads='0';root.dataset.dcPlans='0';const original=window.fetch.bind(window);window.fetch=async(input:any,init?:RequestInit)=>{const url=typeof input==='string'?input:input instanceof URL?input.href:input.url;if(init?.method==='POST'&&/\/api\/(explore|captures|measurements|inventory|discovery|metric-discovery)$/.test(url)){root.dataset.dcReads=String(Number(root.dataset.dcReads||0)+1);if(Number(root.dataset.dcReads)===3)root.dataset.dcLastQuery=String(init.body)}if(init?.method==='POST'&&url.endsWith('/api/numeric-histograms/comparison-selection')){root.dataset.dcPlans=String(Number(root.dataset.dcPlans||0)+1);const body=JSON.parse(String(init.body));root.dataset.dcParent=JSON.stringify(body.comparison[body.side].question)}return original(input,init)}}
 if(phase==='reading')(document.querySelector('.population-run') as HTMLButtonElement).click();
 if(phase==='count'||phase==='share'){const select=host().querySelector('select[aria-label="Distribution scale"]') as HTMLSelectElement;if(phase==='count'){const plot=host().querySelector('goo-aligned-histogram') as any;globals.dcComparison=plot.comparison;root.dataset.dcGolden=JSON.stringify(plot.comparison)}select.value=phase==='share'?'population-share':'count';select.dispatchEvent(new Event('change'));if((host().querySelector('goo-aligned-histogram') as any).comparison!==globals.dcComparison)throw Error('Local display replaced the immutable comparison operands.')}
 function brush(side:string,key:string){const first=host().querySelector<HTMLButtonElement>('.aligned-histogram-side[data-side='+side+'] button[data-key="'+key+'"]')!;first.focus();first.dispatchEvent(new KeyboardEvent('keydown',{key:' ',bubbles:true,cancelable:true}));first.dispatchEvent(new KeyboardEvent('keydown',{key:'ArrowRight',shiftKey:true,bubbles:true,cancelable:true}));document.activeElement!.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}))}
 if(phase==='keyboard')brush('baseline','bin:65');
 if(phase==='opening')[...host().querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Open baseline range')!.click();
 if(phase==='running')(document.querySelector('.population-child .population-run') as HTMLButtonElement).click();
 if(phase==='back')[...document.querySelectorAll<HTMLButtonElement>('.population-child button')].find(b=>b.textContent==='Back to parent population')!.click();
 if(phase==='complete')brush('selection','bin:68');root.dataset.dcPhase=phase;
});
export const explore=actions(()=>{const s=state.current;if(s.phase==='initial')return s.staged?[step('initialized')]:['Wait'];if(s.phase==='initialized')return[step('reading')];if(s.phase==='reading')return s.ready?[step('count')]:['Wait'];if(s.phase==='count')return[step('share')];if(s.phase==='share')return[step('keyboard')];if(s.phase==='keyboard')return s.intent.includes('[1, 4)')?[step('opening')]:['Wait'];if(s.phase==='opening')return s.childStaged?[step('planned')]:['Wait'];if(s.phase==='planned')return[step('running')];if(s.phase==='running')return s.reads===3&&s.count.includes('56')?[step('back')]:['Wait'];if(s.phase==='back')return s.ready?[step('complete')]:['Wait'];return['Wait']});
