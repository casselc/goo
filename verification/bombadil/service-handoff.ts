// Real OTel rows supplied by the optional 316 native journey. URL parameters are
// fixture identity only; this driver never seeds records or saved definitions.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state=extract(({document,window})=>{
 const tracking=(window as any).__gooServiceHandoff316,phase=tracking?.phase||0,entry=document.querySelector<HTMLElement>('.stream-entry'),tail=entry?.querySelector<HTMLElement>('.live-tail'),rows=[...tail?.querySelectorAll<HTMLElement>('tbody>tr')||[]],detail=document.querySelector<HTMLElement>('.service-browser-detail[aria-label="Service handoff"]');
 const visible=(el:HTMLElement|null|undefined)=>!!el&&el.getClientRects().length>0&&!el.closest('[hidden], [inert]');
 const rowSignature=JSON.stringify(rows.map(row=>[row.dataset.tailKey,row.dataset.fromMs,row.dataset.toMs]));
 const viewport=tail?.querySelector<HTMLElement>('goo-table'),origin=rows.find(row=>row.dataset.tailKey===tracking?.key),summary=origin?.querySelector<HTMLElement>('[aria-label="Service actions"]');
 const staged=phase!==3||visible(detail)&&detail!.textContent!.includes(tracking.service)&&detail!.textContent!.includes(tracking.environment)&&detail!.textContent!.includes(tracking.namespace)&&detail!.textContent!.includes('current key only')&&detail!.textContent!.includes(new Date(tracking.fromMs).toISOString())&&detail!.textContent!.includes(new Date(tracking.toMs).toISOString())&&detail!.textContent!.includes('Activity filters are not carried')&&[...detail!.querySelectorAll<HTMLButtonElement>('button')].some(b=>b.textContent==='Read service overview'&&visible(b));
 const returned=phase!==4||visible(entry)&&tail?.dataset.state==='paused'&&rows.length===13&&rowSignature===tracking.signature&&viewport?.scrollTop===tracking.top&&viewport?.scrollLeft===tracking.left&&document.activeElement===summary;
 return {version:'goo.serviceHandoff316',phase,ready:visible(entry)&&!!entry?.querySelector('[aria-label="Live environment"]'),rows:rows.length,tailState:tail?.dataset.state,staged,returned,sourceAttempts:tracking?.sourceAttempts||0,tailAttempts:tracking?.tailAttempts||0,apiAttempts:tracking?.apiAttempts||0,passive:phase<3||tracking.apiAttempts===tracking.beforeStage,pausedExact:phase<2||tail?.dataset.state==='paused'&&rows.length===13&&rowSignature===tracking.signature};
});
export const oneExplicitTail=always(()=>state.current.sourceAttempts===state.current.tailAttempts&&state.current.tailAttempts<=1);
export const stagingAndBackArePassive=always(()=>state.current.passive);
export const exactServiceScope=always(()=>state.current.staged);
export const pausedParentRetained=always(()=>state.current.pausedExact&&state.current.returned);
export const completes=eventually(()=>state.current.phase===4&&state.current.returned&&state.current.rows===13&&state.current.tailAttempts===1).within(20,'seconds');

const step=registerCustomAction('goo.serviceHandoff',async(document,window,phase:number)=>{
 const entry=document.querySelector<HTMLElement>('.stream-entry')!,tail=entry.querySelector<HTMLElement>('.live-tail')!,find=(root:ParentNode,text:string)=>[...root.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===text)!;
 let tracking=(window as any).__gooServiceHandoff316;
 if(phase===1){
  const parameters=new URLSearchParams(window.location.search),environment=parameters.get('env'),namespace=parameters.get('ns'),service=parameters.get('service');if(!environment||namespace===null||!service)throw Error('316 fixture environment, exact namespace and service URL parameters are required');
  tracking=(window as any).__gooServiceHandoff316={phase:0,environment,namespace,service,apiAttempts:0,sourceAttempts:0,tailAttempts:0};
  const source=new Set(['/api/tail','/api/discovery','/api/inventory','/api/metric-discovery','/api/explore','/api/explore/page','/api/explore/counted-values','/api/record-aggregates','/api/measurements']);
  const fetch=window.fetch.bind(window);window.fetch=(input,init)=>{const path=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;if(path.startsWith('/api/'))tracking.apiAttempts++;if(source.has(path))tracking.sourceAttempts++;if(path==='/api/tail')tracking.tailAttempts++;return fetch(input,init)};
  const environmentInput=entry.querySelector<HTMLInputElement>('[aria-label="Live environment"]')!;environmentInput.value=environment;environmentInput.dispatchEvent(new Event('input',{bubbles:true}));const signal=entry.querySelector<HTMLSelectElement>('[aria-label="Live signal"]')!;signal.value='logs';signal.dispatchEvent(new Event('change',{bubbles:true}));const start=find(entry,'Start stream')||find(entry,'Apply scope');if(!start)throw Error('Explicit Start stream control is missing');start.click();
 }else if(phase===2){
  find(entry,'Pause stream').click();const rows=[...tail.querySelectorAll<HTMLElement>('tbody>tr')];if(rows.length!==13)throw Error('Expected thirteen real fixture observations');
  const row=rows.find(row=>row.querySelector<any>('goo-record-summary')?.record?.ServiceName===tracking.service)!;if(!row)throw Error('Reported fixture service row is missing');
  const record=row.querySelector<any>('goo-record-summary')!.record;if(record.ResourceAttributes?.['deployment.environment.name']!==tracking.environment||record.ResourceAttributes?.['service.namespace']!==tracking.namespace)throw Error('Real resource identity differs from fixture metadata');tracking.key=row.dataset.tailKey;tracking.fromMs=Number(row.dataset.fromMs);tracking.toMs=Number(row.dataset.toMs);tracking.signature=JSON.stringify(rows.map(row=>[row.dataset.tailKey,row.dataset.fromMs,row.dataset.toMs]));const viewport=tail.querySelector<HTMLElement>('goo-table')!;
  const summary=row.querySelector<HTMLElement>('[aria-label="Service actions"]')!;summary.focus();if(!(summary.parentElement as HTMLDetailsElement).open)summary.click();const open=find(row,'Open service');if(!open||open.disabled)throw Error('Real reported service cannot be opened');open.focus();tracking.top=viewport.scrollTop;tracking.left=viewport.scrollLeft;tracking.beforeStage=tracking.apiAttempts;
 }else if(phase===3){const row=[...tail.querySelectorAll<HTMLElement>('tbody>tr')].find(row=>row.dataset.tailKey===tracking.key)!;find(row,'Open service').click();}
 else if(phase===4){const detail=document.querySelector<HTMLElement>('.service-browser-detail[aria-label="Service handoff"]')!;find(detail,'Back to paused Activity').click();}
 tracking.phase=phase;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready&&s.phase<2)return ['Wait'];if(s.phase===0)return [step(1)];if(s.phase===1)return s.rows===13&&s.tailAttempts===1?[step(2)]:['Wait'];if(s.phase===2)return [step(3)];if(s.phase===3)return s.staged?[step(4)]:['Wait'];return ['Wait']});
