// A UI-authored records+facet view; the host owns all reads and scope changes.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const panel=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(e=>!e.hidden&&e.getClientRects().length>0);
 const facet=panel?.querySelector<any>('goo-categorical-facet'),data=facet?.data;
 const count=(path:string)=>window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname===path).length;
 return {version:'goo.declarativeFacet310',phase:document.documentElement.dataset.declarativeFacetPhase||'initial',
  population:panel?.dataset.phase,facet:!!facet,values:data?.values?.state,total:data?.values?.total,
  staged:data?.hasStaged===true,pages:count('/api/explore/page'),reads:count('/api/explore'),
  valueReads:count('/api/explore/counted-values'),saves:count('/api/exploration-views'),
  overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const noHiddenReads=always(()=>state.current.reads===0&&state.current.saves===0&&state.current.pages<=2&&state.current.valueReads<=2);
export const localSelectionDoesNotRead=always(()=>!['missing','two','cleared','missingAgain','twoAgain','childStaged'].includes(state.current.phase)||state.current.pages===1&&state.current.valueReads===1);
export const firstRunDoesNotLoadValues=always(()=>state.current.phase!=='reading'||state.current.valueReads===0);
export const childRunDoesNotLoadValues=always(()=>state.current.phase!=='readingChild'||state.current.valueReads===1);
export const wholeScopeCounts=always(()=>state.current.values!=='ready'||state.current.total===(['readingChildValues','complete'].includes(state.current.phase)?'4':'107'));
export const clearOnlyPending=always(()=>state.current.phase!=='cleared'||!state.current.staged);
export const noViewportOverflow=always(()=>!state.current.overflow);
export const terminal=always(()=>state.current.phase!=='complete'||state.current.population==='applied'&&state.current.values==='ready'&&state.current.total==='4'&&state.current.pages===2&&state.current.valueReads===2);
const step=registerCustomAction('goo.declarativeFacet',async(document,window,phase:string)=>{
 const panel=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(e=>!e.hidden&&e.getClientRects().length>0)!;
 const button=(name:string)=>[...panel.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 const choose=(label:string)=>{const facet=panel.querySelector<any>('goo-categorical-facet')!,field=facet.data.declaration.field,name='Include '+field.kind+':'+field.key+' '+label;[...facet.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.getAttribute('aria-label')===name)!.click()};
 if(phase==='initial'||phase==='childStaged')panel.querySelector<HTMLButtonElement>('.population-run')!.click();
 else if(phase==='reading'||phase==='readingChild')button('Read values').click();
 else if(phase==='values'||phase==='cleared')choose('Missing key');
 else if(phase==='missing'||phase==='missingAgain')choose('Present empty string');
 else if(phase==='two')button('Clear staged facet').click();
 else if(phase==='twoAgain')button('Open selected facets').click();
 document.documentElement.dataset.declarativeFacetPhase=({initial:'reading',reading:'values',values:'missing',missing:'two',two:'cleared',cleared:'missingAgain',missingAgain:'twoAgain',twoAgain:'childStaged',childStaged:'readingChild',readingChild:'readingChildValues',readingChildValues:'complete'} as Record<string,string>)[phase];
});
export const explore=actions(()=>{
 const s=state.current;
 if(s.phase==='initial'&&s.population==='staged'&&s.facet)return[step('initial')];
 if(s.phase==='reading'&&s.population==='applied'&&s.values==='unloaded')return[step('reading')];
 if(s.phase==='values'&&s.values==='ready'&&s.total==='107')return[step('values')];
 if(['missing','two','cleared','missingAgain','twoAgain'].includes(s.phase))return[step(s.phase)];
 if(s.phase==='childStaged'&&s.population==='staged'&&s.facet)return[step('childStaged')];
 if(s.phase==='readingChild'&&s.population==='applied'&&s.values==='unloaded')return[step('readingChild')];
 if(s.phase==='readingChildValues'&&s.values==='ready'&&s.total==='4')return[step('readingChildValues')];
 return['Wait'];
});
