// Start from a view authored through the actual HTTP service Customize journey.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const panel=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(e=>!e.hidden&&e.getClientRects().length>0),facet=panel?.querySelector<any>('goo-categorical-facet'),data=facet?.data;
 let configuredMeasure;try{const source=panel?.querySelector<HTMLTextAreaElement>('[aria-label="Composition source"]')?.value;configuredMeasure=source?JSON.parse(source).children.find((c:any)=>c.aggregate)?.aggregate.display.value:undefined}catch{}
 const count=(path:string)=>window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname===path).length;
 return {version:'goo.httpMethodFacet311',phase:document.documentElement.dataset.httpMethodFacetPhase||'initial',population:panel?.dataset.phase,facet:!!facet,values:data?.values?.state,total:data?.values?.total,
  literal:data?.declaration?.field?.kind==='attribute'&&data?.declaration?.field?.key==='http.request.method',
  template:panel?.querySelector<HTMLInputElement>('[aria-label="Reusable template"]')?.checked===true,
  profile:panel?.querySelector<HTMLSelectElement>('[aria-label="Template data profile"]')?.value,
  measure:panel?.querySelector<HTMLSelectElement>('[aria-label="Breakdown measure"]')?.value??configuredMeasure,
  reads:count('/api/explore'),pages:count('/api/explore/page'),valueReads:count('/api/explore/counted-values'),saves:count('/api/exploration-views'),staged:data?.hasStaged===true,
  overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const templateAndLiteralFieldRetained=always(()=>!state.current.population||state.current.literal&&state.current.template&&state.current.profile==='otel.http-server');
export const noHiddenReads=always(()=>state.current.reads<=2&&state.current.valueReads<=2&&state.current.pages===0&&state.current.saves===0);
export const localControlsDoNotRead=always(()=>!['measured','included','cleared','includedAgain','childStaged'].includes(state.current.phase)||state.current.reads===1&&state.current.valueReads===1);
export const measureRetained=always(()=>!['measured','included','cleared','includedAgain','childStaged','readingChild','readingChildValues','complete'].includes(state.current.phase)||state.current.measure==='duration.p95');
export const noAutomaticFacetRead=always(()=>state.current.phase!=='reading'||state.current.valueReads===0);
export const noAutomaticChildFacetRead=always(()=>state.current.phase!=='readingChild'||state.current.valueReads===1);
export const currentScopeCounts=always(()=>state.current.values!=='ready'||state.current.total===(['readingChildValues','complete'].includes(state.current.phase)?'3':'107'));
export const clearOnlyPending=always(()=>state.current.phase!=='cleared'||!state.current.staged);
export const noViewportOverflow=always(()=>!state.current.overflow);
export const terminal=always(()=>state.current.phase!=='complete'||state.current.population==='applied'&&state.current.values==='ready'&&state.current.total==='3'&&state.current.reads===2&&state.current.valueReads===2);
const step=registerCustomAction('goo.httpMethodFacet',async(document,window,phase:string)=>{
 const panel=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(e=>!e.hidden&&e.getClientRects().length>0)!;
 const button=(name:string)=>[...panel.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 const post=()=>panel.querySelector<HTMLButtonElement>('button[aria-label=\'Include attribute:http.request.method "POST"\']')!.click();
 if(phase==='initial'||phase==='childStaged')panel.querySelector<HTMLButtonElement>('.population-run')!.click();
 else if(phase==='reading'||phase==='readingChild')button('Read values').click();
 else if(phase==='values'){const select=panel.querySelector<HTMLSelectElement>('[aria-label="Breakdown measure"]')!;select.value='duration.p95';select.dispatchEvent(new Event('change',{bubbles:true}))}
 else if(phase==='measured'||phase==='cleared')post();
 else if(phase==='included')button('Clear staged facet').click();
 else if(phase==='includedAgain')button('Open selected facets').click();
 document.documentElement.dataset.httpMethodFacetPhase=({initial:'reading',reading:'values',values:'measured',measured:'included',included:'cleared',cleared:'includedAgain',includedAgain:'childStaged',childStaged:'readingChild',readingChild:'readingChildValues',readingChildValues:'complete'} as Record<string,string>)[phase];
});
export const explore=actions(()=>{
 const s=state.current;
 if(s.phase==='initial'&&s.population==='staged'&&s.facet)return[step('initial')];
 if(s.phase==='reading'&&s.population==='applied'&&s.values==='unloaded')return[step('reading')];
 if(s.phase==='values'&&s.values==='ready'&&s.total==='107')return[step('values')];
 if(['measured','included','cleared','includedAgain'].includes(s.phase))return[step(s.phase)];
 if(s.phase==='childStaged'&&s.population==='staged'&&s.facet)return[step('childStaged')];
 if(s.phase==='readingChild'&&s.population==='applied'&&s.values==='unloaded')return[step('readingChild')];
 if(s.phase==='readingChildValues'&&s.values==='ready'&&s.total==='3')return[step('readingChildValues')];
 return['Wait'];
});
