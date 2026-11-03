// Synthetic supplied SSR/SPA display contracts only. Actual OTel/app behavior is
// covered by the 315 native journey; this catalog has no query/source owner.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state=extract(({document,window})=>{
 const facets=[document.querySelector<any>('#component-categorical-facet'),document.querySelector<any>('#component-categorical-facet-spa goo-categorical-facet')];
 const coarse=window.matchMedia('(pointer:coarse)').matches;
 const projections=facets.map(facet=>{
  if(!facet||facet.dataset.state!=='ready')return null;
  const data=facet.data,items=[...facet.querySelectorAll<HTMLElement>('[data-facet-item]')],visible=items.filter(item=>!item.hidden),matches=facet.querySelector('.categorical-facet-matches'),scope=facet.querySelector('.categorical-facet-scope')?.textContent||'',count=facet.querySelector('.categorical-facet-count')?.textContent||'';
  return {valid:data.declaration.field.kind==='attribute'&&data.declaration.field.key==='http.route'&&data.declaration.display.basis==='applied'&&data.values.total==='130'&&data.values.items.length===4&&scope.includes('attribute:http.route')&&scope.includes('Applied fixed scope')&&count==='130 matching records · upstream coverage unknown',loaded:items.length,visible:visible.length,matches:matches?.textContent,live:matches?.getAttribute('aria-live')==='polite'&&matches?.getAttribute('role')==='status',targets:[...facet.querySelectorAll<HTMLElement>('button,input')].filter(control=>control.getClientRects().length>0).every(control=>control.getBoundingClientRect().height+0.01>=(coarse?44:32)),detailsTarget:(facet.querySelector('summary')?.getBoundingClientRect().height||0)+0.01>=(coarse?44:32)};
 });
 const tracking=(window as any).__gooCompactFacet315;
 return {version:'goo.compact-facet315',phase:Number(document.documentElement.dataset.compactFacet315||0),ready:projections.every(Boolean),projections,apiResources:window.performance.getEntriesByType('resource').filter(entry=>new URL(entry.name).pathname.startsWith('/api/')).length,fetches:tracking?.fetches||0,intents:tracking?.intents||0,staleChanged:tracking?.staleChanged||false,coarse};
});
export const remainsPassive=always(()=>state.current.apiResources===0&&state.current.fetches===0&&!state.current.staleChanged);
export const suppliedMeaningStaysExact=always(()=>state.current.projections.every(p=>!p||p.valid&&p.loaded===4&&p.live&&p.matches===`${p.visible} of 4 loaded values`));
export const usableTargets=always(()=>state.current.projections.every(p=>!p||p.targets&&p.detailsTarget));
export const completes=eventually(()=>state.current.phase===10&&state.current.ready&&state.current.intents===2).within(20,'seconds');

const step=registerCustomAction('goo.compactFacet',async(document,window,phase:number)=>{
 const ssr=document.querySelector<any>('#component-categorical-facet')!,spa=document.querySelector<any>('#component-categorical-facet-spa goo-categorical-facet')!;
 const current=(window as any).__gooCompactFacet315;
 const find=(facet:HTMLElement,value:string)=>{const input=facet.querySelector<HTMLInputElement>('input[type=search]')!;input.focus();input.value=value;input.dispatchEvent(new Event('input',{bubbles:true}));if(document.activeElement!==input)throw Error('Local facet Find lost focus');};
 const details=(facet:HTMLElement)=>{const summary=facet.querySelector<HTMLElement>('summary')!;summary.focus();summary.click();if(document.activeElement!==summary)throw Error('Native facet disclosure lost keyboard focus');};
 if(phase===1){
  const tracking=(window as any).__gooCompactFacet315={fetches:0,intents:0,staleChanged:false};
  const fetch=window.fetch.bind(window);window.fetch=(input,init)=>{if(new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname.startsWith('/api/'))tracking.fetches++;return fetch(input,init)};
  for(const facet of [ssr,spa])facet.addEventListener('goo:facet-load',()=>tracking.intents++);
  find(ssr,'/orders');
 }else if(phase===2)details(ssr);
 else if(phase===3)find(ssr,'');
 else if(phase===4){const button=ssr.querySelector<HTMLButtonElement>('button[aria-label="Read values"]')!;current.staleRead=button;button.focus();button.click();if(ssr.querySelector('button[aria-label="Read values"]')!==document.activeElement)throw Error('Supplied re-render lost Read values focus');}
 else if(phase===5){const before=current.intents;current.staleRead.click();current.staleChanged ||= before!==current.intents;}
 else if(phase===6)find(spa,'/orders');
 else if(phase===7)details(spa);
 else if(phase===8){const button=spa.querySelector<HTMLButtonElement>('button[aria-label="Read values"]')!;button.focus();button.click();}
 else if(phase===9){
  const button=spa.querySelector<HTMLButtonElement>('button[aria-label="Read values"]')!,parent=spa.parentNode!,next=spa.nextSibling,before=current.intents;
  spa.remove();button.click();parent.insertBefore(spa,next);button.click();current.staleChanged ||= before!==current.intents;
 }else if(phase===10){find(spa,'');const summary=spa.querySelector<HTMLElement>('summary')!;if(spa.querySelector<HTMLDetailsElement>('details')?.open)summary.click();}
 document.documentElement.dataset.compactFacet315=String(phase);
});
export const explore=actions(()=>state.current.ready&&state.current.phase<10?[step(state.current.phase+1)]:['Wait']);
