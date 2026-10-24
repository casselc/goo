import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
// Run against the real general numeric-comparison capture emitted by journey175.
// The sequence ranks stored evidence locally, stages its exact question, then
// authorizes one fresh read. Pinning/ranking/editing after that remains local.
const state=extract(({document})=>({
 phase:document.documentElement.dataset.ncPhase||'initial',reads:Number(document.documentElement.dataset.ncReads||0),
 frozen:document.getElementById('paircohortExplain')?.textContent||'',golden:document.documentElement.dataset.ncGolden||'',
 saved:!!document.querySelector('#paircohortMeasures .numeric-comparison[data-numeric-state=comparable]'),
 live:!!document.querySelector('#cohortMeasures .numeric-comparison[data-numeric-state=comparable]'),
 runReady:!!document.getElementById('exploreRun')&&!(document.getElementById('exploreRun') as HTMLButtonElement).disabled,
 count:document.getElementById('exploreCount')?.textContent||'',
 signed:[...document.querySelectorAll<HTMLElement>('.numeric-comparison .signed-comparison-cell')].every(el=>el.dataset.value==='unavailable'||!!el.dataset.scaleMin&&!!el.dataset.scaleMax),
 domains:[...document.querySelectorAll('.numeric-comparison figure.signed-comparison tr')].every(row=>{const cells=[...row.querySelectorAll<HTMLElement>('.signed-comparison-cell')];return cells.length===0||cells.length===2&&cells[0].dataset.scaleMin===cells[1].dataset.scaleMin&&cells[0].dataset.scaleMax===cells[1].dataset.scaleMax}),
 invalidDisabled:(document.getElementById('exploreRun') as HTMLButtonElement|null)?.disabled,
}));
export const complete=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const atMostOneRead=always(()=>state.current.reads<=1);
export const stageAndStoredActionsAreSourceFree=always(()=>!['initial','mean','sum','min','max','coverage','staged'].includes(state.current.phase)||state.current.reads===0);
export const storedEvidenceFixed=always(()=>!state.current.golden||!state.current.frozen||state.current.frozen===state.current.golden);
export const signedRowsHaveSharedDomains=always(()=>state.current.signed&&state.current.domains);
export const completionRequiresOneExplicitRead=always(()=>state.current.phase!=='complete'||state.current.reads===1);
export const invalidNumericAuthoringBlocksRun=always(()=>state.current.phase!=='invalid'||state.current.invalidDisabled===true);
const step=registerCustomAction('goo.numericComparison',async(document,window,phase:string)=>{
 const root=document.documentElement,globals=window as any;
 if(phase==='mean'){
  root.dataset.ncReads='0';root.dataset.ncGolden=document.getElementById('paircohortExplain')!.textContent!;
  const pair=JSON.parse(root.dataset.ncGolden);globals.ncQuestion=pair.request.selection.question;
  const original=window.fetch.bind(window);window.fetch=async(input:any,init?:RequestInit)=>{const url=typeof input==='string'?input:input instanceof URL?input.href:input.url;if(init?.method==='POST'&&/\/api\/(explore|captures|measurements)$/.test(url))root.dataset.ncReads=String(Number(root.dataset.ncReads||0)+1);return original(input,init)};
 }
 if(['mean','sum','min','max'].includes(phase)){const select=document.getElementById('paircomparisonMeasure') as HTMLSelectElement;select.value='numeric-'+phase;select.dispatchEvent(new Event('change'))}
 if(phase==='coverage'){const summary=document.querySelector<HTMLElement>('#paircohortMeasures .numeric-comparison details > summary')!;summary.click()}
 if(phase==='staged')window.location.hash=globals.GooQuestionProtocol.questionLink(globals.ncQuestion);
 if(phase==='reading')(document.getElementById('exploreRun') as HTMLButtonElement).click();
 if(phase==='pinning'){const pin=document.getElementById('baselinePin') as HTMLButtonElement;const details=pin.closest('details');if(details)details.open=true;pin.click()}
 if(phase==='invalid'||phase==='complete'){
  const outer=document.getElementById('measureComposer') as HTMLDetailsElement,inner=outer.querySelector('details') as HTMLDetailsElement;outer.open=true;inner.open=true;const unit=document.getElementById('numericUnit') as HTMLInputElement;unit.value=phase==='invalid'?'':globals.ncQuestion.numeric.unit;unit.dispatchEvent(new Event('input'));
 }
 root.dataset.ncPhase=phase;
});
export const explore=actions(()=>{const s=state.current;if(s.phase==='initial')return s.saved?[step('mean')]:['Wait'];if(s.phase==='mean')return[step('sum')];if(s.phase==='sum')return[step('min')];if(s.phase==='min')return[step('max')];if(s.phase==='max')return[step('coverage')];if(s.phase==='coverage')return[step('staged')];if(s.phase==='staged')return s.runReady?[step('reading')]:['Wait'];if(s.phase==='reading')return s.count.includes('128')?[step('pinning')]:['Wait'];if(s.phase==='pinning')return s.live?[step('invalid')]:['Wait'];if(s.phase==='invalid')return[step('complete')];return['Wait']});
