import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{const h=(window as any).__typedTrace340;return{version:'goo.typedTraceSearch340',phase:h?.phase||0,source:h?.source||0,verified:h?.verified||0,policyPreserved:h?.policyPreserved||false,canRun:!!document.querySelector<HTMLButtonElement>('#exploreRun')?.getClientRects().length&&document.querySelector<HTMLButtonElement>('#exploreRun')?.disabled===false,parentReady:!!document.querySelector('#exploreRows .record-row'),searchReady:!!document.querySelector('.trace-search:not([hidden]) tr[data-trace-id]'),contextReady:document.querySelector('#exploreLogStatus')?.textContent?.includes('captured logs')||false,overflow:document.documentElement.scrollWidth>window.innerWidth}});
export const bounded=always(()=>state.current.source<=4);
export const noOverflow=always(()=>!state.current.overflow);
export const complete=eventually(()=>state.current.phase===8&&state.current.source===4&&state.current.verified===4&&state.current.policyPreserved).within(35,'seconds');
const step=registerCustomAction('goo.typedTraceSearch',async(document,window,phase:number)=>{
 const button=(name:string,host:Element|Document=document)=>[...host.querySelectorAll<HTMLButtonElement>('button')].find(x=>x.textContent===name)!,panel=()=>document.querySelector<HTMLElement>('.trace-search:not([hidden])')!;
 let h=(window as any).__typedTrace340;
 if(phase===1){
  h=(window as any).__typedTrace340={phase:0,source:0,verified:0,policyPreserved:false,parent:null,search:null,trace:[]};const original=window.fetch.bind(window);
  const same=(a:any,b:any)=>JSON.stringify(a)===JSON.stringify(b);
  window.fetch=async(input,init)=>{const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href),method=init?.method||'GET';if(method==='POST'&&['/api/explore','/api/trace-search'].includes(url.pathname))h.source++;const response=await original(input,init);if(method==='POST'&&response.ok&&['/api/explore','/api/trace-search'].includes(url.pathname)){const data=await response.clone().json();if(url.pathname==='/api/trace-search'){if(!same(data.request.environmentSelector,h.parent.question.environmentSelector)||data.request.environment!==undefined||data.request.conditions.length!==2||data.matchedTraces!=='15'||data.items.length!==15||data.hasMore)throw Error('Typed trace search lost requested scope or full bounded population');const cross=data.items.find((x:any)=>x.witnesses[0].service!==x.witnesses[1].service);if(!cross)throw Error('No independently reported cross-service witnesses');for(const w of cross.witnesses)for(const key of ['deployment.environment.name','deployment.environment'])if(!w.fieldValues.some((v:any)=>v.field.kind==='resource'&&v.field.key===key))throw Error('Environment witness presence omitted');h.search=data;h.traceId=cross.traceId;h.verified++;}else if(!h.parent)h.parent=data;else{const q=data.question;if(!same(q.environmentSelector,h.parent.question.environmentSelector)||q.fromMs!==h.parent.question.fromMs||q.toMs!==h.parent.question.toMs||q.where?.all?.length!==1||q.where.all[0].field.key!=='trace_id'||q.where.all[0].value!==h.traceId)throw Error('Checked exact trace continuation changed policy/window/ID');h.trace.push(data)}}return response};
  (document.querySelector('#exploreRun') as HTMLButtonElement).click();
 }else if(phase===2){
  if(!h.parent?.question.environmentSelector||h.parent.question.environmentSelector.policy!=='effective')throw Error('Expected genuine typed effective parent');h.verified++;
  document.querySelector<HTMLButtonElement>('#exploreRows .record-row button[aria-label="Inspect record"]')!.click();button('Find traces with this operation').click();
 }else if(phase===3){
  const before=h.source;button('Add span condition',panel()).click();const condition=panel().querySelectorAll('.trace-search-condition')[1],input=condition.querySelector<HTMLInputElement>('input[aria-label$=" field"]')!,operator=condition.querySelector<HTMLSelectElement>('select[aria-label$=" operator"]')!,value=condition.querySelector<HTMLInputElement>('input[aria-label$=" value"]')!;
  input.value='status_class';input.dispatchEvent(new Event('input',{bubbles:true}));operator.value='eq';operator.dispatchEvent(new Event('change',{bubbles:true}));value.value='error';value.dispatchEvent(new Event('input',{bubbles:true}));if(h.source!==before)throw Error('Condition staging read telemetry');button('Run trace search',panel()).click();
 }else if(phase===4){
  const before=h.source,row=panel().querySelector<HTMLElement>('tr[data-trace-id="'+h.traceId+'"]')!;row.querySelector<HTMLElement>('details>summary')!.click();if(h.source!==before||h.source!==2)throw Error('Witness disclosure read telemetry');h.verified++;
 }else if(phase===5){button('Open trace (fresh read)',panel().querySelector('tr[data-trace-id="'+h.traceId+'"]')!).click();
 }else if(phase===6){
  if(h.source!==4||h.trace.length!==2||!h.trace.some((x:any)=>x.question.signal==='spans'&&x.count==='3')||!h.trace.some((x:any)=>x.question.signal==='logs'&&x.count==='1'))throw Error('Expected exact typed spans and logs');h.policyPreserved=true;h.verified++;
  (document.querySelector('#exploreContextClose') as HTMLButtonElement).click();
 }else if(phase===7){if(h.source!==4||!panel()||!panel().querySelector('tr[data-trace-id="'+h.traceId+'"]'))throw Error('Return lost held trace-search result or reread');button('Close trace search',panel()).click();
 }else if(phase===8){if(h.source!==4||h.verified!==4||!h.policyPreserved||document.querySelector('.trace-search:not([hidden])'))throw Error('Incomplete typed trace search campaign')}
 h.phase=phase;
});
export const explore=actions(()=>{const s=state.current;if(s.phase===0)return s.canRun?[step(1)]:['Wait'];if(s.phase===1)return s.parentReady?[step(2)]:['Wait'];if(s.phase===2)return[step(3)];if(s.phase===3)return s.searchReady?[step(4)]:['Wait'];if(s.phase===4)return[step(5)];if(s.phase===5)return s.contextReady?[step(6)]:['Wait'];if(s.phase<8)return[step(s.phase+1)];return['Wait']});
