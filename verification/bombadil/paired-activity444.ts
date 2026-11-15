// Paired custody exploration: two ordinary source connections, then local work.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state=extract(({document,window})=>{
 const o=(window as any).__paired444||{},visible=(element:any)=>!!element?.getClientRects().length&&!element.closest('[hidden], [aria-hidden="true"], [inert]'),pair=document.querySelector('.paired-activity'),shell=document.querySelector('.stream-entry');
 const lane=(signal:string)=>{
  const region=pair?.querySelector('[data-paired-lane="'+signal+'"]'),tail=region?.querySelector<HTMLElement>('.live-tail'),rows=[...region?.querySelectorAll<HTMLElement>('[data-tail-key]')||[]],records=[...region?.querySelectorAll<any>('goo-record-summary')||[]].map(element=>element.record),owner=tail?.dataset.owner,notification=o.events?.[owner||''];
  return {owner:owner||'',phase:tail?.dataset.state||'',keys:rows.map(row=>row.dataset.tailKey),windows:rows.map(row=>[row.dataset.tailKey,row.dataset.fromMs,row.dataset.toMs]),lastTo:tail?.dataset.lastTo||'',find:region?.querySelector<HTMLInputElement>('[aria-label="Find loaded records"]')?.value||'',readWindow:notification?.readWindow||null,mixed:records.some(record=>!record||typeof record[signal==='logs'?'Body':'SpanName']!=='string'||typeof record[signal==='logs'?'SpanName':'Body']==='string')};
 };
 const logs=lane('logs'),spans=lane('spans'),calls=o.calls||[],tails=calls.filter((call:any)=>call.path==='/api/tail'),pure=new Set(['/api/cel/context','/api/cel/plan']),forbidden=calls.filter((call:any)=>call.path!=='/api/tail'&&!pure.has(call.path)&&!/^\/api\/queries\/[^/]+\/cancel$/.test(call.path)),phase=Number(document.documentElement.dataset.paired444Phase||0),held=o.held;
 const fixed=o.fixedQuestion,spanSelector=tails.find((call:any)=>call.body?.signal==='spans')?.body,origin=o.spanOrigin;
 const project=(value:any)=>JSON.stringify({keys:value.keys,windows:value.windows,lastTo:value.lastTo,find:value.find,readWindow:value.readWindow});
 return {version:'goo.paired-activity444/1',phase,ready:!!shell?.querySelector('[aria-label="Live environment"]')&&!!pair,pairedVisible:visible(pair),fixedVisible:visible(document.querySelector('#exploreRun'))&&!visible(shell),lanes:{logs,spans},tailRequests:tails.length,forbiddenPosts:forbidden.length,
  typedSelectors:tails.every((call:any)=>call.body&&call.body.signal!=='both'&&['logs','spans'].includes(call.body.signal||'logs')),
  bothObserved:logs.keys.length>0&&spans.keys.length>0,mixedFamilies:logs.mixed||spans.mixed,sourceFreeStaging:phase!==1||calls.length===0,
  logsOnlyPaused:phase!==3||logs.phase==='paused'&&spans.phase!=='paused',
  heldRestored:!held||project(logs)===project(held.logs)&&project(spans)===project(held.spans),
  focusRestored:!!o.fixedOpener&&document.activeElement===o.fixedOpener,
  fixedScope:phase!==6||!!fixed&&!!origin&&fixed.signal==='spans'&&fixed.environment===o.environment&&fixed.fromMs===origin.fromMs&&fixed.toMs===origin.toMs&&JSON.stringify(fixed.where||null)===JSON.stringify(spanSelector?.where||null)&&document.querySelector<HTMLSelectElement>('#exploreSignal')?.value==='spans',
  overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const validSourceSelectors=always(()=>state.current.typedSelectors&&!state.current.mixedFamilies);
export const sourceBudget=always(()=>state.current.tailRequests<=2&&state.current.forbiddenPosts===0&&state.current.sourceFreeStaging);
export const scopedPause=always(()=>state.current.logsOnlyPaused);
export const fixedStagingKeepsFamily=always(()=>state.current.fixedScope);
export const readFreeReturn=always(()=>state.current.phase!==7||state.current.pairedVisible&&state.current.bothObserved&&state.current.tailRequests===2&&state.current.heldRestored&&state.current.focusRestored&&state.current.lanes.logs.phase==='paused'&&state.current.lanes.spans.phase==='paused');
export const fits=always(()=>!state.current.overflow);
export const completes=eventually(()=>state.current.phase===7).within(20,'seconds');

const step=registerCustomAction('goo.pairedActivity444',async(document,window,next:number)=>{
 const target=window as any,pair=()=>document.querySelector<HTMLElement>('.paired-activity')!,lane=(signal:string)=>pair().querySelector<HTMLElement>('[data-paired-lane="'+signal+'"]')!,visible=(element:any)=>!!element?.getClientRects().length&&!element.closest('[hidden], [aria-hidden="true"], [inert]'),click=(button:any)=>{if(!button||button.disabled||!visible(button))throw Error('Paired action unavailable');button.scrollIntoView({block:'nearest'});button.click()},opener=(signal:string)=>lane(signal).querySelector<HTMLButtonElement>('[data-tail-inspect]')!;
 if(next===1){
  const environment=new URL(window.location.href).searchParams.get('paired444Environment');if(!environment)throw Error('paired444Environment required');
  const observation:any=target.__paired444={environment,calls:[],events:{}},original=window.fetch;
  window.fetch=function(input,options){const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href),method=String(options?.method||(input instanceof Request?input.method:'GET')).toUpperCase();if(method==='POST'&&url.pathname.startsWith('/api/')){let body;try{body=JSON.parse(String(options?.body||''))}catch{body=null}observation.calls.push({path:url.pathname,body})}return Reflect.apply(original,this,[input,options])};
  window.addEventListener('goo:tail-state',(event:Event)=>{const detail=(event as CustomEvent).detail;if(detail?.owner)observation.events[detail.owner]=structuredClone(detail)});
  window.addEventListener('goo:question-state',(event:Event)=>{const detail=(event as CustomEvent).detail;if(detail?.action==='stage')observation.fixedQuestion=structuredClone(detail.requested)});
  const env=document.querySelector<HTMLInputElement>('[aria-label="Live environment"]')!,signal=document.querySelector<HTMLSelectElement>('[aria-label="Live signal"]')!;env.value=environment;env.dispatchEvent(new Event('input',{bubbles:true}));signal.value='both';signal.dispatchEvent(new Event('change',{bubbles:true}));
  if(observation.calls.length)throw Error('Pair staging read source');
 }else if(next===2){click([...document.querySelectorAll<HTMLButtonElement>('.stream-entry-toolbar button')].find(button=>button.textContent==='Start streams'&&visible(button)));}
 else if(next===3){click(opener('logs'));}
 else if(next===4){
  const region=lane('logs'),find=region.querySelector<HTMLInputElement>('[aria-label="Find loaded records"]')!,record=region.querySelector<any>('goo-record-summary')?.record,value=String(record?.Body||'');if(!value)throw Error('Loaded log body required');
  find.value=value.slice(0,100);find.dispatchEvent(new Event('input',{bubbles:true}));click(pair().querySelector('[aria-label="Close selected item details"]'));
 }else if(next===5){const button=opener('spans'),row=button.closest<HTMLElement>('[data-tail-key]')!;target.__paired444.spanOrigin={fromMs:Number(row.dataset.fromMs),toMs:Number(row.dataset.toMs)};click(button);}
 else if(next===6){
  const project=(signal:string)=>{const region=lane(signal),tail=region.querySelector<HTMLElement>('.live-tail')!,rows=[...region.querySelectorAll<HTMLElement>('[data-tail-key]')],event=target.__paired444.events[tail.dataset.owner!];return {keys:rows.map(row=>row.dataset.tailKey),windows:rows.map(row=>[row.dataset.tailKey,row.dataset.fromMs,row.dataset.toMs]),lastTo:tail.dataset.lastTo||'',find:region.querySelector<HTMLInputElement>('[aria-label="Find loaded records"]')!.value,readWindow:event?.readWindow||null}};
  target.__paired444.held={logs:project('logs'),spans:project('spans')};const analyze=pair().querySelector<HTMLButtonElement>('[aria-label="Analyze observation window · spans"]')!;target.__paired444.fixedOpener=analyze;click(analyze);
 }else if(next===7){click([...document.querySelectorAll<HTMLButtonElement>('.entry-live-button')].find(button=>button.textContent==='Return to paused activity'&&visible(button)));}
 document.documentElement.dataset.paired444Phase=String(next);
});
export const explore=actions(()=>{
 const s=state.current;
 if(s.phase===0)return s.ready?[step(1)]:['Wait'];
 if(s.phase===1)return[step(2)];
 if(s.phase===2)return s.bothObserved?[step(3)]:['Wait'];
 if(s.phase===3)return[step(4)];
 if(s.phase===4)return[step(5)];
 if(s.phase===5)return s.lanes.logs.phase==='paused'&&s.lanes.spans.phase==='paused'?[step(6)]:['Wait'];
 if(s.phase===6)return s.fixedVisible?[step(7)]:['Wait'];
 return['Wait'];
});
