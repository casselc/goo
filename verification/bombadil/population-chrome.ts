// Real UI-saved fixed-query comparison. The actor checks compact scope and reads both window values,
// then stages one common filter and independently verifies both new populations.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const h=(window as any).__gooChrome339,visible=(e:HTMLElement)=>e.getClientRects().length>0&&!e.closest('[hidden],[inert]'),p=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(visible);
 return {version:'goo.populationChrome339',phase:h?.phase||0,ready:!!p,population:p?.dataset.phase,pairReady:!!p?.querySelector<any>('goo-group-count-comparison')?.comparison,facet:p?.querySelector<HTMLElement>('goo-categorical-facet')?.dataset.state,source:h?.source||0,lookups:h?.lookups||0,compares:h?.compares||0,writes:h?.writes||0,verified:h?.verified||0,immutable:!h?.fingerprint||!!h.originalGroup?.isConnected&&JSON.stringify(h.originalGroup.comparison)===h.fingerprint,scopeVerified:!!h?.scopeVerified,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const bounded=always(()=>state.current.source<=6&&state.current.lookups<=2&&state.current.compares<=2&&state.current.writes===0);
export const immutable=always(()=>state.current.immutable);
export const noOverflow=always(()=>!state.current.overflow);
export const complete=eventually(()=>state.current.phase===7&&state.current.scopeVerified&&state.current.verified===4&&state.current.source===6&&state.current.lookups===2&&state.current.compares===2).within(30,'seconds');
const step=registerCustomAction('goo.populationChrome',async(document,window,phase:number)=>{
 const visible=(e:HTMLElement)=>e.getClientRects().length>0&&!e.closest('[hidden],[inert]'),p=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(visible)!,group=()=>p.querySelector<any>('goo-group-count-comparison'),facet=()=>p.querySelector<any>('goo-categorical-facet'),button=(name:string,host:Element=p)=>[...host.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.getAttribute('aria-label')===name||b.textContent===name)!,change=(input:HTMLSelectElement,value:string)=>{if(input.value===value)return;input.value=value;input.dispatchEvent(new Event('change',{bubbles:true}))};
 let h=(window as any).__gooChrome339;
 if(phase===1){
  h=(window as any).__gooChrome339={phase:0,source:0,lookups:0,compares:0,writes:0,verified:0,reads:[],values:[]};
  const paths=new Set(['/api/explore','/api/explore/page','/api/record-aggregates','/api/tail','/api/inventory','/api/discovery','/api/measurements','/api/explore/counted-values']),fetch=window.fetch.bind(window);
  window.fetch=async(input,init)=>{const path=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;if(paths.has(path))h.source++;if(path==='/api/explore/counted-values')h.lookups++;if(path==='/api/compare')h.compares++;if((init?.method||'GET')!=='GET'&&['/api/captures','/api/kept-observations/records','/api/exploration-views','/api/documents'].includes(path))h.writes++;const response=await fetch(input,init);if(path==='/api/explore'||path==='/api/record-aggregates')h.reads.push({path,data:await response.clone().json()});if(path==='/api/explore/counted-values')h.values.push(await response.clone().json());return response};
  p.querySelector<HTMLButtonElement>('.population-run')!.click();
 }else if(phase===2){
  change(p.querySelector<HTMLSelectElement>('[aria-label="Comparison view"]')!,'groups');h.originalGroup=group();h.original=group().comparison;h.fingerprint=JSON.stringify(h.original);
  if(h.original.baseline.rows[0].values.count!=='13'||h.original.selection.rows[0].values.count!=='107')throw Error('Expected genuine full13/107 operands');h.verified++;
  const scope=p.querySelector<HTMLDetailsElement>('.population-scope-context')!;if(scope.open)throw Error('Fresh scope should begin compact');scope.querySelector<HTMLElement>(':scope>summary')!.click();for(const name of ['.population-window','.population-baseline-window'])if(!visible(scope.querySelector<HTMLElement>(name)!))throw Error('Exact window facts must be disclosed');for(const side of ['baseline','selection']){const q=h.original[side].scope.selector;if(!scope.textContent?.includes(new Date(q.fromMs).toISOString())||!scope.textContent?.includes(new Date(q.toMs).toISOString()))throw Error('Exact half-open bounds missing')}scope.querySelector<HTMLElement>(':scope>summary')!.click();if(scope.open||h.source!==2||h.compares!==1)throw Error('Scope disclosure mutated read budget');h.scopeVerified=true;
  if(!facet().data.baseline)throw Error('Facet lacks paired-window attribution');button('Read values',facet()).click();
 }else if(phase===3){
  const values=h.values.find(value=>value.count==='107'),baseline=h.values.find(value=>value.count==='13'),counts=Object.fromEntries(values.groups.map(g=>[String(g.g0),g.count]));if(h.values.length!==2||!baseline)throw Error('Both counted-value operands were not read');
  if(values.count!=='107'||counts.GET!=='75'||counts.POST!=='27'||counts['']!=='2'||counts.null!=='3'||values.groups.length!==4||!baseline.groups.some(group=>group.g0==='DELETE'&&group.count==='2'))throw Error('Expected independent category partitions');const row=facet().querySelector('[data-facet-value-key="1DELETE"]');if(!row||row.querySelector('[data-side=selection]')?.getAttribute('data-count')!=='0'||row.querySelector('[data-side=baseline]')?.getAttribute('data-count')!=='2')throw Error('Baseline-only category did not align an observed-partition zero');h.verified++;
  h.pendingRun=p.querySelector<HTMLButtonElement>('.population-run')!;h.pendingHandler=h.pendingRun.onclick;button('Include attribute:http.request.method "POST"',facet()).click();
 }else if(phase===4){
  const tray=p.querySelector('[aria-label="Facet selection"]')!,pending=tray.textContent;if(!p.querySelector<HTMLButtonElement>('.population-run')!.disabled)throw Error('Parent Run accepts unapplied Both-windows choices');if(h.pendingHandler)await h.pendingHandler.call(h.pendingRun);
  if(!tray.textContent?.includes('Both windows')||tray.textContent!==pending||h.source!==4||JSON.stringify(group().comparison)!==h.fingerprint)throw Error('Selection or retained Run handler read/replaced the held pair');h.verified++;
  button('Filter both windows',tray).click();
 }else if(phase===5){
  if(!p.querySelector('h3')?.textContent?.includes('Filter both windows')||h.source!==4)throw Error('Common continuation did not remain staged');p.querySelector<HTMLButtonElement>('.population-run')!.click();
 }else if(phase===6){
  change(p.querySelector<HTMLSelectElement>('[aria-label="Comparison view"]')!,'groups');const pair=group().comparison,canonical=(value:any):any=>Array.isArray(value)?value.map(canonical):value&&typeof value==='object'?Object.fromEntries(Object.keys(value).sort().map(key=>[key,canonical(value[key])])):value;
  if(pair.baseline.rows[0].values.count!=='4'||pair.selection.rows[0].values.count!=='27')throw Error('POST inclusion did not narrow both independent populations');
  for(const side of ['baseline','selection']){const original=h.original[side].scope.selector,expected={...structuredClone(original),offset:0,where:{all:[...(original.where?[structuredClone(original.where)]:[]),{field:{kind:'attribute',key:'http.request.method'},op:'eq',value:'POST'}]}};if(JSON.stringify(canonical(pair[side].scope.selector))!==JSON.stringify(canonical(expected)))throw Error('POST inclusion lost an operand definition')}
  if(facet().dataset.state!=='unloaded')throw Error('Child adopted the parent counted-values witness');h.verified++;
 }else if(phase===7){
  if(!h.scopeVerified||h.verified!==4||h.source!==6||h.lookups!==2||h.compares!==2||h.writes!==0)throw Error('Incomplete paired-facet campaign');
 }
 h.phase=phase;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready)return['Wait'];if(s.phase===0)return s.population==='staged'?[step(1)]:['Wait'];if(s.phase===1||s.phase===5)return s.population==='applied'&&s.pairReady?[step(s.phase+1)]:['Wait'];if(s.phase===2)return s.facet==='ready'?[step(3)]:['Wait'];if(s.phase<7)return[step(s.phase+1)];return['Wait']});
