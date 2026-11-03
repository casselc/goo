// A real saved service view authored by the optional 317 native journey. One
// explicit Run acquires data; chart controls and held-parent navigation are local.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const held=(window as any).__gooCompactChart317,phase=held?.phase||0,visible=(el:HTMLElement)=>el.getClientRects().length>0&&!el.closest('[hidden], [inert]'),panels=[...document.querySelectorAll<HTMLElement>('.population-workspace')].filter(visible),panel=panels[0],chart=panel?.querySelector<any>('[data-composition-role="aggregate"] goo-chart'),snapshot=chart?.snapshot;
 const parent=held?.parent,heldChart=parent?.querySelector<any>('[data-composition-role="aggregate"] goo-chart'),quality=chart?.querySelector<HTMLElement>('.scope-view-quality'),facts=chart?.querySelector<HTMLDetailsElement>('.scope-view-read-facts'),settings=panel?.querySelector<HTMLDetailsElement>('.scoped-breakdown-settings');
 const preserved=phase<2||!!heldChart&&JSON.stringify(heldChart.snapshot)===held.snapshot;
 const q=snapshot?.quality,expectedQuality=q?[`${q.extent} groups`,`${q.value} values`,`${q.sampling} upstream sampling`,...(q.freshness!=='fixed'?[`${q.freshness} snapshot`]:[])].join(' · '):'';
 const certainty=panel?.dataset.phase!=='applied'||!!quality&&visible(quality)&&quality.textContent===expectedQuality;
 const readFactsExact=panel?.dataset.phase!=='applied'||!!facts&&[snapshot.origin,`${q.freshness} snapshot`,snapshot.capturedAt].filter(Boolean).every(value=>facts.textContent?.includes(value));
 const chartFirst=panel?.dataset.phase!=='applied'||!!settings&&!!chart&&(chart.compareDocumentPosition(settings)&4)!==0;
 const stagedChild=phase<8||phase>=10||panels.length===1&&panel?.dataset.phase==='staged'&&document.querySelectorAll('.population-child').length===1;
 const returned=phase<10||panel===parent&&panel?.dataset.phase==='applied'&&!!snapshot&&JSON.stringify(snapshot)===held.snapshot&&document.querySelectorAll('.population-child').length===0;
 return {version:'goo.compactChart317',phase,population:panel?.dataset.phase,ready:!!panel,groupRows:snapshot?.rows?.length||0,sourceAttempts:held?.sourceAttempts||0,exploreAttempts:held?.exploreAttempts||0,preserved,certainty,readFactsExact,chartFirst,stagedChild,returned,settings:!!settings,readFacts:!!facts,selected:!!panel&&[...panel.querySelectorAll<HTMLButtonElement>('button')].some(b=>b.textContent==='Stage selected group'&&!b.disabled),overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const oneExplicitSourceRead=always(()=>state.current.sourceAttempts===state.current.exploreAttempts&&state.current.exploreAttempts<=1);
export const heldSnapshotUnchanged=always(()=>state.current.preserved);
export const certaintyAlwaysVisible=always(()=>state.current.certainty&&state.current.readFactsExact);
export const chartBeforeSettings=always(()=>state.current.chartFirst);
export const selectedGroupLifetime=always(()=>state.current.stagedChild&&state.current.returned);
export const noOverflow=always(()=>!state.current.overflow);
export const completes=eventually(()=>state.current.phase===10&&state.current.returned&&state.current.certainty&&state.current.exploreAttempts===1).within(25,'seconds');
const step=registerCustomAction('goo.compactChart',async(document,window,phase:number)=>{
 const visible=(el:HTMLElement)=>el.getClientRects().length>0&&!el.closest('[hidden], [inert]'),panel=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(visible)!,chart=panel.querySelector<any>('[data-composition-role="aggregate"] goo-chart'),find=(root:ParentNode,text:string)=>[...root.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===text)!;
 let held=(window as any).__gooCompactChart317;
 if(phase===1){
  held=(window as any).__gooCompactChart317={phase:0,sourceAttempts:0,exploreAttempts:0};const source=new Set(['/api/explore','/api/explore/page','/api/explore/counted-values','/api/record-aggregates','/api/tail','/api/discovery','/api/inventory','/api/metric-discovery','/api/measurements']),fetch=window.fetch.bind(window);window.fetch=(input,init)=>{const path=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;if(source.has(path))held.sourceAttempts++;if(path==='/api/explore')held.exploreAttempts++;return fetch(input,init)};panel.querySelector<HTMLButtonElement>('.population-run')!.click();
 }else if(phase===2){
  if(!chart?.snapshot||chart.snapshot.rows.length<2)throw Error('Expected multiple actual captured service groups');held.parent=panel;held.snapshot=JSON.stringify(chart.snapshot);const summary=chart.querySelector<HTMLElement>('.scope-view-read-facts>summary')!;summary.focus();summary.click();if(document.activeElement!==summary)throw Error('Read facts lost native disclosure focus');
 }else if(phase===3){chart.querySelector<HTMLElement>('.scope-view-read-facts>summary')!.click();const summary=panel.querySelector<HTMLElement>('.scoped-breakdown-settings>summary')!;summary.focus();summary.click();if(document.activeElement!==summary)throw Error('Chart settings lost native disclosure focus');}
 else if(phase===4){const select=panel.querySelector<HTMLSelectElement>('[aria-label="Breakdown view"]')!;select.value='table';select.dispatchEvent(new Event('change',{bubbles:true}));const order=panel.querySelector<HTMLSelectElement>('[aria-label="Group order"]')!;order.value='label';order.dispatchEvent(new Event('change',{bubbles:true}));if(chart.display.kind!=='table'||chart.display.order!=='label')throw Error('Chart settings did not apply local presentation');}
 else if(phase===5){const input=panel.querySelector<HTMLInputElement>('[aria-label="spans display Find"]')!,value=chart.snapshot.rows[0].values[chart.snapshot.dimensions[0]];if(typeof value!=='string'||!value)throw Error('Expected a real named route group');input.value=value;input.dispatchEvent(new Event('input',{bubbles:true}));if(chart.display.query!==value||chart.querySelectorAll('tbody>tr').length!==1)throw Error('Local Find did not isolate the supplied route group');input.value='';input.dispatchEvent(new Event('input',{bubbles:true}));if(chart.display.query!==''||chart.querySelectorAll('tbody>tr').length!==chart.snapshot.rows.length)throw Error('Clear Find did not restore all held groups');}
 else if(phase===6){const select=chart.querySelector<HTMLButtonElement>('tbody button[aria-label^="Select "]')!;select.focus();select.click();held.stage=find(panel,'Stage selected group');if(!held.stage||held.stage.disabled)throw Error('Actual group selection did not become actionable');}
 else if(phase===7)held.stage.click();
 else if(phase===8){const open=find(panel,'Open selected facets');if(!open||open.disabled)throw Error('Staged group continuation is unavailable');open.click();}
 else if(phase===9){if(!held.stage.disabled)throw Error('Retired group action remains enabled');held.stage.click();}
 else if(phase===10){find(document,'Back to parent population').click();held.stage.click();}
 held.phase=phase;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready)return ['Wait'];if(s.phase===0)return s.population==='staged'?[step(1)]:['Wait'];if(s.phase===1)return s.population==='applied'&&s.groupRows>=2&&s.settings&&s.readFacts?[step(2)]:['Wait'];if(s.phase>=2&&s.phase<10)return[step(s.phase+1)];return ['Wait']});
