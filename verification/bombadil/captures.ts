// Read-only immutable capture exploration. Never save, refresh, export or emit workload.
import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const capture=extract(({document})=>{
 const el=(id:string)=>document.getElementById(id),read=(id:string)=>{try{return JSON.parse(el(id)?.textContent||'null')}catch{return null}};
 const receipt=read('captureGeneralReceipt'),evidence=read('savedexploreExplain'),rowCount=document.querySelectorAll('#savedexploreRows .record-row').length,loading=el('generalCaptureWorkspace')?.getAttribute('aria-busy')==='true';
 let scope=true,accounting=true,stable=true;
 if(receipt&&evidence){const expected={...receipt.receipt.question.explore,offset:evidence.question.offset,limit:100};scope=JSON.stringify(expected)===JSON.stringify(evidence.question)&&evidence.receipt.id===receipt.receipt.id&&evidence.basis===receipt.basis;const offset=evidence.question.offset,n=receipt.retainedRecords;accounting=n<=1000&&rowCount===Math.min(100,n-offset)&&offset>=0&&offset%100===0&&offset<1000&&(el('savedexploreCount')?.querySelector('span')?.textContent||'')===receipt.count;
  if(!loading)accounting=accounting&&((el('savedexplorePrevious') as HTMLButtonElement).disabled===(offset===0))&&((el('savedexploreNext') as HTMLButtonElement).disabled===(offset+rowCount>=n));
  const golden=JSON.parse(document.documentElement.dataset.gooCaptureGolden||'null');if(golden&&offset===0&&!loading)stable=JSON.stringify({receipt,question:evidence.question,rows:[...document.querySelectorAll('#savedexploreRows .record-row')].map(r=>[...r.querySelectorAll(':scope > td')].map(c=>c.textContent))})===JSON.stringify(golden);
 }
 const display=(el('captureEvidenceSummary')?.querySelector('goo-evidence-summary')as any)?.data;const summaryCoherent=!evidence||display?.mode==='frozen'&&display.facts.find((x:any)=>x.label==='Source records')?.value===receipt.count&&display.details.find((x:any)=>x.label==='Capture ID')?.value===evidence.receipt.id&&display.details.find((x:any)=>x.label==='Stored custody time')?.value===evidence.receipt.capturedAt&&display.details.find((x:any)=>x.label==='Displayed page')?.value===rowCount+' records · offset '+evidence.question.offset;
 const ids=[...document.querySelectorAll('[id]')].map(e=>e.id);
 return {summaryCoherent,version:'goo.captures.v1',stage:document.documentElement.dataset.gooCaptureStage||'initial',ready:!!receipt&&!!evidence,loading,scope,accounting,stable,uniqueIDs:new Set(ids).size===ids.length,readonly:!el('savedexploreBuilder')&&!el('savedexploreRun')&&!el('savedexploreSaveCapture')&&!document.querySelector('#captureExploreEvidence .query-workbench'),offset:evidence?.question?.offset||0,
 next:!(el('savedexploreNext') as HTMLButtonElement)?.disabled,previous:!(el('savedexplorePrevious') as HTMLButtonElement)?.disabled,viewer:!!document.querySelector('#savedexploreRows .record-attributes[open] .attribute-viewer'),receipt,evidence,rowCount};
});
export const ready=eventually(()=>capture.current.ready).within(10,'seconds');
export const complete=eventually(()=>capture.current.stage==='complete').within(25,'seconds');
export const frozenSummaryMatchesReceipt=always(()=>capture.current.loading||capture.current.summaryCoherent);
export const fixedQuestionAndBasis=always(()=>capture.current.scope);
export const retainedRecordsAndPaging=always(()=>capture.current.accounting);
export const replayedFirstPage=always(()=>capture.current.stable);
export const readOnlyControls=always(()=>!capture.current.ready||!!capture.current.readonly);
export const isolatedComponentIDs=always(()=>capture.current.uniqueIDs);
const act=registerCustomAction('goo.captures',async(document,_window,command:string,next?:string)=>{
 const el=(id:string)=>document.getElementById(id),click=(id:string)=>el(id)?.click();
 if(command==='remember'){const receipt=JSON.parse(el('captureGeneralReceipt')?.textContent||'null'),evidence=JSON.parse(el('savedexploreExplain')?.textContent||'null');document.documentElement.dataset.gooCaptureGolden=JSON.stringify({receipt,question:evidence.question,rows:[...document.querySelectorAll('#savedexploreRows .record-row')].map(r=>[...r.querySelectorAll(':scope > td')].map(c=>c.textContent))})}
 if(command==='scope'){const details=document.querySelector<HTMLDetailsElement>('#captureEvidenceSummary goo-evidence-summary > details');if(details)details.open=!details.open}if(command==='next')click('savedexploreNext');if(command==='previous')click('savedexplorePrevious');if(command==='distribution')click('savedresultDistribution');if(command==='records')click('savedresultRecords');
 if(command==='inspect'){const detail=document.querySelector<HTMLDetailsElement>('#savedexploreRows .record-attributes');if(detail)detail.open=true}
 if(command==='search'){const search=document.querySelector<HTMLInputElement>('#savedexploreRows .attribute-viewer input[type="search"]');if(search){search.value='duration';search.dispatchEvent(new Event('input',{bubbles:true}))}}
 if(next)document.documentElement.dataset.gooCaptureStage=next;
});
export const inspect=actions(()=>{const s=capture.current;if(!s.ready||s.loading)return['Wait'];if(s.stage==='initial')return[act('remember','remembered')];if(s.stage==='remembered')return[act('scope','scope')];if(s.stage==='scope')return[act('next','paging')];if(s.stage==='paging')return s.offset===100?[act('previous','returning')]:['Wait'];if(s.stage==='returning')return s.offset===0?[act('distribution','distribution')]:['Wait'];if(s.stage==='distribution')return[act('records','records')];if(s.stage==='records')return[act('inspect','inspected')];if(s.stage==='inspected')return s.viewer?[act('search','complete')]:['Wait'];const choices=[act('distribution'),act('records'),act('inspect')];if(s.next)choices.push(act('next'));if(s.previous)choices.push(act('previous'));if(s.viewer)choices.push(act('search'));return choices;});
