// Read-only exploration of saved full-population measures and fixed time buckets.
import { always,eventually } from '@antithesishq/bombadil';
import { extract,actions,registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions,noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{
 const el=(id:string)=>document.getElementById(id);let e:any=null;try{e=JSON.parse(el('savedexploreExplain')?.textContent||'null')}catch{}
 const d=e?.duration,points=e?.timeSeries||[],loading=el('generalCaptureWorkspace')?.getAttribute('aria-busy')==='true';
 let intervals=true;for(let i=0;i<points.length;i++){const p=points[i];intervals=intervals&&p.fromMs===e.question.fromMs+i*e.question.bucketMs&&p.toMs===Math.min(p.fromMs+e.question.bucketMs,e.question.toMs)&&((p.count==='0')===(p.state==='no_data'))}
 const total=points.reduce((n:bigint,p:any)=>n+BigInt(p.count),0n);
 const golden=document.documentElement.dataset.gooBasicsGolden;
 const signature=e?JSON.stringify({d,points,from:e.question.fromMs,to:e.question.toMs,basis:e.basis}):'';
 return {version:'goo.core-basics.v1',stage:document.documentElement.dataset.gooBasicsStage||'initial',ready:!!d,loading,offset:e?.question?.offset||0,
 measures:!d||d.count==='205'&&d.sumNs==='21115000000'&&d.p95Ns==='195000000'&&el('saveddurationSummary')?.textContent?.includes('195.000 ms'),
 intervals:!d||intervals&&total===205n&&points.length===document.querySelectorAll('#savedtimeChart .time-bar').length,
 stable:!golden||!e||golden===signature,signature,next:!(el('savedexploreNext') as HTMLButtonElement)?.disabled,previous:!(el('savedexplorePrevious') as HTMLButtonElement)?.disabled};
});
export const ready=eventually(()=>state.current.ready).within(10,'seconds');
export const complete=eventually(()=>state.current.stage==='complete').within(20,'seconds');
export const fullPopulationMeasures=always(()=>!!state.current.measures);
export const fixedHalfOpenIntervals=always(()=>state.current.intervals);
export const pagesPreserveMeasures=always(()=>state.current.stable);
const act=registerCustomAction('goo.core-basics',async(document,_window,command:string,next?:string)=>{
 const el=(id:string)=>document.getElementById(id);if(command==='remember'){const e=JSON.parse(el('savedexploreExplain')?.textContent||'null');document.documentElement.dataset.gooBasicsGolden=JSON.stringify({d:e.duration,points:e.timeSeries,from:e.question.fromMs,to:e.question.toMs,basis:e.basis})}
 if(command==='chart')document.querySelector<HTMLButtonElement>('#savedtimeChart .time-bar')?.click();if(command==='next')el('savedexploreNext')?.click();if(command==='previous')el('savedexplorePrevious')?.click();if(command==='distribution')el('savedresultDistribution')?.click();if(command==='records')el('savedresultRecords')?.click();if(next)document.documentElement.dataset.gooBasicsStage=next;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready||s.loading)return['Wait'];if(s.stage==='initial')return[act('remember','remembered')];if(s.stage==='remembered')return[act('chart','chart')];if(s.stage==='chart')return[act('next','paging')];if(s.stage==='paging')return s.offset===100?[act('previous','returning')]:['Wait'];if(s.stage==='returning')return s.offset===0?[act('records','complete')]:['Wait'];const a=[act('chart'),act('records'),act('distribution')];if(s.next)a.push(act('next'));if(s.previous)a.push(act('previous'));return a});
