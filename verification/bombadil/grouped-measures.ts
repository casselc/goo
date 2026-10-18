// Saved-only group ranking, local search and stable record pages. No source reads.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{
 const el=(id:string)=>document.getElementById(id);let e:any=null;try{e=JSON.parse(el('savedexploreExplain')?.textContent||'null')}catch{}
 const stage=document.documentElement.dataset.gooGroupsStage||'initial',loading=el('generalCaptureWorkspace')?.getAttribute('aria-busy')==='true';const rows=[...document.querySelectorAll<HTMLElement>('#saveddistributionContent .distribution-row')];
 let grouped=true;if(e){grouped=e.duration?.p95Ns==='100000000'&&e.duration.sumNs==='2500000000'&&e.groups?.length===4&&e.groups.reduce((s:bigint,g:any)=>s+BigInt(g.count),0n)===124n&&e.groups.reduce((s:bigint,g:any)=>s+BigInt(g.duration.sumNs),0n)===2500000000n}
 const metric=(el('saveddistributionMeasure') as HTMLSelectElement)?.value;const ranking=!rows.length||metric==='count'||rows[0]?.dataset.value==='Missing key';
 const golden=document.documentElement.dataset.gooGroupsGolden,signature=e?JSON.stringify({basis:e.basis,duration:e.duration,groups:e.groups}):'';
 return {version:'goo.groups.v1',stage,ready:!!e?.groups?.length,loading,grouped,ranking,stable:!golden||!e||golden===signature,offset:e?.question?.offset||0,rows:rows.length,visible:rows.filter(x=>!x.hidden).length,metric,signature};
});
export const ready=eventually(()=>state.current.ready).within(10,'seconds');
export const complete=eventually(()=>state.current.stage==='complete').within(25,'seconds');
export const fullGroupAccounting=always(()=>state.current.grouped);
export const exactRanking=always(()=>state.current.ranking);
export const immutableMeasurePages=always(()=>state.current.stable);
const act=registerCustomAction('goo.groups',async(document,_window,command:string,next?:string)=>{
 const el=(id:string)=>document.getElementById(id),click=(id:string)=>el(id)?.click();
 if(command==='remember'){const e=JSON.parse(el('savedexploreExplain')?.textContent||'null');document.documentElement.dataset.gooGroupsGolden=JSON.stringify({basis:e.basis,duration:e.duration,groups:e.groups})}
 if(command==='distribution')click('savedresultDistribution');
 if(['count','mean','p95','p99','max'].includes(command)){const select=el('saveddistributionMeasure') as HTMLSelectElement;select.value=command;select.dispatchEvent(new Event('change',{bubbles:true}))}
 if(command==='search'||command==='clear'){const search=el('savedfacetSearch') as HTMLInputElement;search.value=command==='search'?'/slow':'';search.dispatchEvent(new Event('input',{bubbles:true}))}
 if(command==='next'){click('savedresultRecords');click('savedexploreNext')}
 if(command==='previous')click('savedexplorePrevious');
 if(next)document.documentElement.dataset.gooGroupsStage=next;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready||s.loading)return['Wait'];if(s.stage==='initial')return[act('remember','remembered')];if(s.stage==='remembered')return[act('distribution','distribution')];if(s.stage==='distribution')return[act('p95','ranked')];if(s.stage==='ranked')return[act('search','filtered')];if(s.stage==='filtered')return s.visible===1?[act('clear','cleared')]:['Wait'];if(s.stage==='cleared')return[act('next','paging')];if(s.stage==='paging')return s.offset===100?[act('previous','returning')]:['Wait'];if(s.stage==='returning')return s.offset===0?[act('distribution','complete')]:['Wait'];return['count','mean','p95','p99','max'].map(x=>act(x))});
