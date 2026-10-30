import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>({version:'goo.scoped261',phase:document.documentElement.dataset.scoped261||'initial',reads:Number(document.documentElement.dataset.scoped261Reads||0),bad:document.documentElement.dataset.scoped261Bad==='true',bars:document.querySelectorAll('.scoped-breakdown .scope-view-bar').length,table:!!document.querySelector('.scoped-breakdown goo-chart table'),selected:document.querySelector('.scoped-breakdown-selection')?.textContent||'',staged:document.querySelector('#exploreStatus')?.textContent?.includes('Group filter staged'),applied:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied'}));
export const onlyRunReads=always(()=>!state.current.bad&&state.current.reads<=2);
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
const act=registerCustomAction('goo.scoped261',async(document,window,phase:string)=>{
 document.documentElement.dataset.scoped261=phase;
 const click=(root:ParentNode,label:string)=>{const b=[...root.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===label&&!b.disabled&&b.getClientRects().length);if(!b)throw Error('Missing '+label);b.click()};
 if(phase==='started'){const native=window.fetch.bind(window);window.fetch=(input,init)=>{if(init?.method==='POST'&&new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname==='/api/explore'){document.documentElement.dataset.scoped261Reads=String(Number(document.documentElement.dataset.scoped261Reads||0)+1);if(!['started','rerun'].includes(document.documentElement.dataset.scoped261||''))document.documentElement.dataset.scoped261Bad='true'}return native(input,init)};click(document,'Run question')}
 else if(phase==='table'){const select=document.querySelector<HTMLSelectElement>('[aria-label="Breakdown view"]')!;select.value='table';select.dispatchEvent(new Event('change',{bubbles:true}))}
 else if(phase==='selected'){const button=document.querySelector<HTMLButtonElement>('.scoped-breakdown goo-chart tbody button')!;button.click()}
 else if(phase==='staged')click(document.querySelector('.scoped-breakdown')!,'Stage selected group');
 else if(phase==='rerun')click(document,'Run question');
});
export const follow=actions(()=>{const s=state.current;if(s.phase==='initial')return[act('started')];if(s.phase==='started'&&s.bars===2&&s.reads===1)return[act('table')];if(s.phase==='table'&&s.table)return[act('selected')];if(s.phase==='selected'&&s.selected)return[act('staged')];if(s.phase==='staged'&&s.staged&&s.reads===1)return[act('rerun')];if(s.phase==='rerun'&&s.applied&&s.table&&s.reads===2)return[act('complete')];return['Wait']});
