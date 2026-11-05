// Real UI-saved fixed-query comparison. The actor reads Selected-window values,
// then stages one common filter and independently verifies both new populations.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const h=(window as any).__gooPaired337,visible=(e:HTMLElement)=>e.getClientRects().length>0&&!e.closest('[hidden],[inert]'),p=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(visible);
 return {version:'goo.pairedFacet337',phase:h?.phase||0,ready:!!p,population:p?.dataset.phase,pairReady:!!p?.querySelector<any>('goo-group-count-comparison')?.comparison,facet:p?.querySelector<HTMLElement>('goo-categorical-facet')?.dataset.state,source:h?.source||0,lookups:h?.lookups||0,compares:h?.compares||0,writes:h?.writes||0,verified:h?.verified||0,immutable:!h?.fingerprint||!!h.originalGroup?.isConnected&&JSON.stringify(h.originalGroup.comparison)===h.fingerprint,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const bounded=always(()=>state.current.source<=5&&state.current.lookups<=1&&state.current.compares<=2&&state.current.writes===0);
export const immutable=always(()=>state.current.immutable);
export const noOverflow=always(()=>!state.current.overflow);
export const complete=eventually(()=>state.current.phase===7&&state.current.verified===4&&state.current.source===5&&state.current.lookups===1&&state.current.compares===2).within(30,'seconds');
const step=registerCustomAction('goo.pairedFacet',async(document,window,phase:number)=>{
 const visible=(e:HTMLElement)=>e.getClientRects().length>0&&!e.closest('[hidden],[inert]'),p=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(visible)!,group=()=>p.querySelector<any>('goo-group-count-comparison'),facet=()=>p.querySelector<any>('goo-categorical-facet'),button=(name:string,host:Element=p)=>[...host.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.getAttribute('aria-label')===name||b.textContent===name)!,change=(input:HTMLSelectElement,value:string)=>{if(input.value===value)return;input.value=value;input.dispatchEvent(new Event('change',{bubbles:true}))};
 let h=(window as any).__gooPaired337;
 if(phase===1){
  h=(window as any).__gooPaired337={phase:0,source:0,lookups:0,compares:0,writes:0,verified:0,reads:[],values:[]};
  const paths=new Set(['/api/explore','/api/explore/page','/api/record-aggregates','/api/tail','/api/inventory','/api/discovery','/api/measurements','/api/explore/counted-values']),fetch=window.fetch.bind(window);
  window.fetch=async(input,init)=>{const path=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;if(paths.has(path))h.source++;if(path==='/api/explore/counted-values')h.lookups++;if(path==='/api/compare')h.compares++;if((init?.method||'GET')!=='GET'&&['/api/captures','/api/kept-observations/records','/api/exploration-views','/api/documents'].includes(path))h.writes++;const response=await fetch(input,init);if(path==='/api/explore'||path==='/api/record-aggregates')h.reads.push({path,data:await response.clone().json()});if(path==='/api/explore/counted-values')h.values.push(await response.clone().json());return response};
  p.querySelector<HTMLButtonElement>('.population-run')!.click();
 }else if(phase===2){
  change(p.querySelector<HTMLSelectElement>('[aria-label="Comparison view"]')!,'groups');h.originalGroup=group();h.original=group().comparison;h.fingerprint=JSON.stringify(h.original);
  if(h.original.baseline.rows[0].values.count!=='13'||h.original.selection.rows[0].values.count!=='107')throw Error('Expected genuine full13/107 operands');h.verified++;
  if(!facet().querySelector('.categorical-facet-scope')?.textContent?.includes('Selected window values'))throw Error('Facet lacks selected-window attribution');button('Read values',facet()).click();
 }else if(phase===3){
  const values=h.values.at(-1),counts=Object.fromEntries(values.groups.map(g=>[String(g.g0),g.count]));
  if(values.count!=='107'||counts.GET!=='75'||counts.POST!=='27'||counts['']!=='2'||counts.null!=='3'||values.groups.length!==4)throw Error('Expected full selected-window category population');h.verified++;
  h.pendingRun=p.querySelector<HTMLButtonElement>('.population-run')!;h.pendingHandler=h.pendingRun.onclick;button('Exclude attribute:http.request.method "GET"',facet()).click();
 }else if(phase===4){
  const tray=p.querySelector('[aria-label="Facet selection"]')!,pending=tray.textContent;if(!p.querySelector<HTMLButtonElement>('.population-run')!.disabled)throw Error('Parent Run accepts unapplied Both-windows choices');if(h.pendingHandler)await h.pendingHandler.call(h.pendingRun);
  if(!tray.textContent?.includes('Both windows')||tray.textContent!==pending||h.source!==3||JSON.stringify(group().comparison)!==h.fingerprint)throw Error('Selection or retained Run handler read/replaced the held pair');h.verified++;
  button('Filter both windows',tray).click();
 }else if(phase===5){
  if(!p.querySelector('h3')?.textContent?.includes('Filter both windows')||h.source!==3)throw Error('Common continuation did not remain staged');p.querySelector<HTMLButtonElement>('.population-run')!.click();
 }else if(phase===6){
  change(p.querySelector<HTMLSelectElement>('[aria-label="Comparison view"]')!,'groups');const pair=group().comparison,canonical=(value:any):any=>Array.isArray(value)?value.map(canonical):value&&typeof value==='object'?Object.fromEntries(Object.keys(value).sort().map(key=>[key,canonical(value[key])])):value;
  if(pair.baseline.rows[0].values.count!=='6'||pair.selection.rows[0].values.count!=='32')throw Error('Common exclusion did not narrow both independent populations');
  for(const side of ['baseline','selection']){const original=h.original[side].scope.selector,expected={...structuredClone(original),offset:0,where:{all:[...(original.where?[structuredClone(original.where)]:[]),{any:[{field:{kind:'attribute',key:'http.request.method'},op:'missing'},{field:{kind:'attribute',key:'http.request.method'},op:'ne',value:'GET'}]}]}};if(JSON.stringify(canonical(pair[side].scope.selector))!==JSON.stringify(canonical(expected)))throw Error('Common exclusion lost an operand definition')}
  if(facet().dataset.state!=='unloaded')throw Error('Child adopted the parent counted-values witness');h.verified++;
 }else if(phase===7){
  if(h.verified!==4||h.source!==5||h.lookups!==1||h.compares!==2||h.writes!==0)throw Error('Incomplete paired-facet campaign');
 }
 h.phase=phase;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready)return['Wait'];if(s.phase===0)return s.population==='staged'?[step(1)]:['Wait'];if(s.phase===1||s.phase===5)return s.population==='applied'&&s.pairReady?[step(s.phase+1)]:['Wait'];if(s.phase===2)return s.facet==='ready'?[step(3)]:['Wait'];if(s.phase<7)return[step(s.phase+1)];return['Wait']});
