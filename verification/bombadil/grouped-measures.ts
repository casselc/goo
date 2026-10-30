// Saved-only group ranking, local search and stable record pages. No source reads.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
import {timeBarContract} from './time-bars-contract.ts';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{
 const el=(id:string)=>document.getElementById(id);let e:any=null;try{e=JSON.parse(el('savedexploreExplain')?.textContent||'null')}catch{}
 const stage=document.documentElement.dataset.gooGroupsStage||'initial',loading=el('generalCaptureWorkspace')?.getAttribute('aria-busy')==='true';const rows=[...document.querySelectorAll<HTMLElement>('#saveddistributionContent .scope-view-bar')];
 let grouped=true;if(e){grouped=e.duration?.p95Ns==='100000000'&&e.duration.sumNs==='2500000000'&&e.groups?.length===4&&e.groups.reduce((s:bigint,g:any)=>s+BigInt(g.count),0n)===124n&&e.groups.reduce((s:bigint,g:any)=>s+BigInt(g.duration.sumNs),0n)===2500000000n}
 const metric=document.querySelector<HTMLSelectElement>('#saveddistributionContent [aria-label="Breakdown measure"]')?.value;const chart=document.querySelector<any>('#saveddistributionContent goo-chart'),byKey=new Map((chart?.snapshot?.rows||[]).map((row:any)=>[row.key,row]));let previous:bigint[]|null=null,ranking=true;for(const node of rows){const row:any=byKey.get(node.dataset.key),raw=row?.values?.[metric||'count'];if(typeof raw!=='string'){ranking=false;break}const match=/^(-?\d+)(?:\/(\d+))?$/.exec(raw);if(!match){ranking=false;break}const next=[BigInt(match[1]),BigInt(match[2]||'1')];if(previous&&previous[0]*next[1]<next[0]*previous[1])ranking=false;previous=next;}
 const golden=document.documentElement.dataset.gooGroupsGolden,signature=e?JSON.stringify({basis:e.basis,duration:e.duration,groups:e.groups}):'';
 const facets=[...document.querySelectorAll<any>('goo-facet-list')].every(list=>[...list.querySelectorAll<HTMLElement>(':scope > [data-facet-item]')].every(row=>row.hidden===!(row.dataset.facetText||'').toLocaleLowerCase().includes(list.query.toLocaleLowerCase())));return {timeBars:document.querySelectorAll('#savedtimeChart goo-time-bars').length,facets,version:'goo.groups.v1',stage,ready:!!e?.groups?.length,loading,grouped,ranking,stable:!golden||!e||golden===signature,offset:e?.question?.offset||0,rows:rows.length,visible:rows.filter(x=>!x.hidden).length,metric,signature};
});
export const ready=eventually(()=>state.current.ready).within(10,'seconds');
export const complete=eventually(()=>state.current.stage==='complete').within(25,'seconds');
export const localFacetContract=always(()=>state.current.facets);
export const suppliedTimeBarGeometry=always(()=>timeBarContract.current.valid);
export const fullGroupAccounting=always(()=>state.current.grouped);
export const exactRanking=always(()=>state.current.ranking);
export const immutableMeasurePages=always(()=>state.current.stable);
const act=registerCustomAction('goo.groups',async(document,_window,command:string,next?:string)=>{
 const el=(id:string)=>document.getElementById(id),click=(id:string)=>el(id)?.click();
 if(command==='remember'){const e=JSON.parse(el('savedexploreExplain')?.textContent||'null');document.documentElement.dataset.gooGroupsGolden=JSON.stringify({basis:e.basis,duration:e.duration,groups:e.groups})}
 if(command==='chart'){const plot=document.querySelector<any>('#savedtimeChart goo-time-bars'),bar=plot?.querySelector('button.time-bar') as HTMLButtonElement|null;if(bar){bar.focus();plot.selectedKey=bar.dataset.key;if(document.activeElement!==bar||plot.querySelector('.time-bar[data-key="'+bar.dataset.key+'"]')!==bar)throw Error('Controlled chart selection replaced focus');bar.click()}}
 if(command==='distribution')click('savedresultDistribution');
 if(['count','mean','p95','p99','max'].includes(command)){const select=document.querySelector<HTMLSelectElement>('#saveddistributionContent [aria-label="Breakdown measure"]')!;select.value=command==='count'?'count':'duration.'+command;select.dispatchEvent(new Event('change',{bubbles:true}))}
 if(command==='search'||command==='clear'){const search=el('savedfacetSearch') as HTMLInputElement;search.value=command==='search'?'/slow':'';search.dispatchEvent(new Event('input',{bubbles:true}))}
 if(command==='next'){click('savedresultRecords');click('savedexploreNext')}
 if(command==='previous')click('savedexplorePrevious');
 if(next)document.documentElement.dataset.gooGroupsStage=next;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready||s.loading)return['Wait'];if(s.stage==='initial')return[act('remember','remembered')];if(s.stage==='remembered')return[act('distribution','distribution')];if(s.stage==='distribution')return[act('p95','ranked')];if(s.stage==='ranked')return[act('search','filtered')];if(s.stage==='filtered')return s.visible===1?[act('clear','cleared')]:['Wait'];if(s.stage==='cleared')return[act('next','paging')];if(s.stage==='paging')return s.offset===100?[act('previous','returning')]:['Wait'];if(s.stage==='returning')return s.offset===0?[act('distribution','complete')]:['Wait'];return[...['count','mean','p95','p99','max'].map(x=>act(x)),...(s.timeBars?[act('chart')]:[])]});
