// An external driver explores held records; it never invents observation evidence.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const o=(window as any).__chrono458||{},pair=document.querySelector('.paired-activity'),element=pair?.querySelector<any>('goo-held-chronology'),visible=(value:any)=>!!value?.getClientRects().length&&!value.closest('[hidden],[aria-hidden="true"],[inert]'),calls=o.calls||[],tail=calls.filter((call:any)=>call.path==='/api/tail'),phase=Number(document.documentElement.dataset.chrono458Phase||0),held=visible(element)?element.data:null;
 const lanes=Object.fromEntries(['logs','spans'].map(signal=>{const lane=pair?.querySelector('[data-paired-lane="'+signal+'"]'),root=lane?.querySelector<HTMLElement>('.live-tail');return [signal,{phase:root?.dataset.state||'',keys:[...lane?.querySelectorAll<HTMLElement>('[data-tail-key]')||[]].map(row=>row.dataset.tailKey),find:lane?.querySelector<HTMLInputElement>('[aria-label="Find loaded records"]')?.value||''}]}));
 const forbidden=calls.filter((call:any)=>call.path!=='/api/tail'&&!['/api/cel/context','/api/cel/plan'].includes(call.path)&&!/^\/api\/queries\/[^/]+\/cancel$/.test(call.path)),dataClean=!held||Object.values(held.lanes).every((lane:any)=>lane.rows.every((item:any)=>!Object.hasOwn(item,'membership'))),fixed=o.question,origin=o.origin;
 return {version:'goo.heldChronology458Actor/1',phase,ready:!!document.querySelector('[aria-label="Live environment"]'),pairVisible:visible(pair),chronologyVisible:visible(element),fixedVisible:visible(document.querySelector('#exploreRun'))&&!visible(pair),lanes,tailRequests:tail.length,forbidden:forbidden.length,bothObserved:lanes.logs.keys.length>0&&lanes.spans.keys.length>0,paused:lanes.logs.phase==='paused'&&lanes.spans.phase==='paused',heldClean:dataClean,heldRows:held?Object.values(held.lanes).reduce((sum:number,lane:any)=>sum+lane.rows.length,0):0,detailVisible:visible(pair?.querySelector('.ui-inspector')),fixedScope:phase<10||!!fixed&&!!origin&&fixed.signal===origin.signal&&fixed.environment===o.environment&&fixed.fromMs===origin.readWindow.fromMs&&fixed.toMs===origin.readWindow.toMs,focusRestored:!o.fixedOpener||document.activeElement===o.fixedOpener,keysPreserved:!o.original||JSON.stringify(lanes.logs.keys)===JSON.stringify(o.original.logs)&&JSON.stringify(lanes.spans.keys)===JSON.stringify(o.original.spans),overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const sourceBoundary=always(()=>state.current.tailRequests<=2&&state.current.forbidden===0&&state.current.heldClean);
export const heldModeIsPaused=always(()=>!state.current.chronologyVisible||state.current.paused);
export const originalScope=always(()=>state.current.fixedScope);
export const fits=always(()=>!state.current.overflow);
export const complete=eventually(()=>state.current.phase===12).within(30,'seconds');
export const localReturn=always(()=>state.current.phase!==12||state.current.pairVisible&&!state.current.chronologyVisible&&state.current.paused&&state.current.keysPreserved&&state.current.tailRequests===2);
const step=registerCustomAction('goo.heldChronology458',async(document,window,next:number)=>{
 const target=window as any,pair=()=>document.querySelector<HTMLElement>('.paired-activity')!,chronology=()=>pair().querySelector<any>('goo-held-chronology'),visible=(element:any)=>!!element?.getClientRects().length&&!element.closest('[hidden],[aria-hidden="true"],[inert]'),click=(button:any)=>{if(!button||button.disabled||!visible(button))throw Error('Chronology action unavailable');button.scrollIntoView({block:'nearest'});button.click()},button=(name:string)=>[...pair().querySelectorAll<HTMLButtonElement>('button')].find(value=>value.textContent===name&&visible(value)),find=()=>chronology().querySelector<HTMLInputElement>('[aria-label="Find held records"]')!;
 if(next===1){
  const raw=new URL(window.location.href).searchParams.get('chrono458');if(!raw)throw Error('chrono458 parameters required');const config=JSON.parse(raw),o=target.__chrono458={...config,calls:[]},original=window.fetch;
  window.fetch=function(input,options){const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href),method=String(options?.method||(input instanceof Request?input.method:'GET')).toUpperCase();if(method==='POST'&&url.pathname.startsWith('/api/')){let body;try{body=JSON.parse(String(options?.body||''))}catch{}o.calls.push({path:url.pathname,body})}return Reflect.apply(original,this,[input,options])};
  window.addEventListener('goo:question-state',(event:Event)=>{const detail=(event as CustomEvent).detail;if(detail?.action==='stage')o.question=structuredClone(detail.requested)});
  const env=document.querySelector<HTMLInputElement>('[aria-label="Live environment"]')!,signal=document.querySelector<HTMLSelectElement>('[aria-label="Live signal"]')!;env.value=config.environment;env.dispatchEvent(new Event('input',{bubbles:true}));signal.value='both';signal.dispatchEvent(new Event('change',{bubbles:true}));
 }else if(next===2){click([...document.querySelectorAll<HTMLButtonElement>('.stream-entry-toolbar button')].find(value=>/^(Start streams|Apply scope)$/.test(value.textContent||'')&&visible(value)));}
 else if(next===3){const o=target.__chrono458;o.original=Object.fromEntries(['logs','spans'].map(signal=>[signal,[...pair().querySelectorAll<HTMLElement>('[data-paired-lane="'+signal+'"] [data-tail-key]')].map(row=>row.dataset.tailKey)]));click(button('Hold chronology'));}
 else if(next===4){find().value='actor';find().dispatchEvent(new Event('input',{bubbles:true}));}
 else if(next===5){click(chronology().querySelector('[data-chronology-inspect]'));}
 else if(next===6){click(pair().querySelector('[aria-label="Close selected item details"]'));}
 else if(next===7){click(button('Back to lanes'));}
 else if(next===8){click(button('Hold chronology'));}
 else if(next===9){find().value='actor spans';find().dispatchEvent(new Event('input',{bubbles:true}));const row=chronology().querySelector<HTMLElement>('[data-chronology-key][data-signal="spans"]')!;const held=chronology().data,item=held.lanes.spans.rows.find((item:any)=>item.key===row.dataset.sourceKey);target.__chrono458.origin={signal:'spans',readWindow:item.readWindow};click(row.querySelector('[data-chronology-inspect]'));}
 else if(next===10){const analyze=pair().querySelector<HTMLButtonElement>('[aria-label="Analyze observation window · spans"]')!;target.__chrono458.fixedOpener=analyze;click(analyze);}
 else if(next===11){click([...document.querySelectorAll<HTMLButtonElement>('button')].find(value=>value.textContent==='Return to paused activity'&&visible(value)));}
 else if(next===12){click(button('Back to lanes'));}
 document.documentElement.dataset.chrono458Phase=String(next);
});
export const explore=actions(()=>{
 const s=state.current;if(s.phase===0)return s.ready?[step(1)]:['Wait'];if(s.phase===1)return[step(2)];if(s.phase===2)return s.bothObserved?[step(3)]:['Wait'];if(s.phase===3)return s.chronologyVisible&&s.heldRows>0?[step(4)]:['Wait'];if(s.phase===4)return[step(5)];if(s.phase===5)return s.detailVisible?[step(6)]:['Wait'];if(s.phase===6)return[step(7)];if(s.phase===7)return[step(8)];if(s.phase===8)return s.chronologyVisible?[step(9)]:['Wait'];if(s.phase===9)return s.detailVisible?[step(10)]:['Wait'];if(s.phase===10)return s.fixedVisible?[step(11)]:['Wait'];if(s.phase===11)return s.chronologyVisible&&s.focusRestored?[step(12)]:['Wait'];return['Wait'];
});
