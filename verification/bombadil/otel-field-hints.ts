// One explicit source read; semantic browsing and exact legacy selection stay local.
import{always,eventually}from'@antithesishq/bombadil';
import{extract,actions,registerCustomAction}from'@antithesishq/bombadil/browser';
export{noUncaughtExceptions,noUnhandledPromiseRejections}from'@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{const root=document.querySelector('#exploreFilterEditor'),rows=[...(root?.querySelectorAll('.filter-observations details[data-field]')||[])],step=Number(document.documentElement.dataset.hintStep||0);return{step,ready:!!document.getElementById('exploreRun'),phase:document.getElementById('exploration')?.dataset.questionPhase,overflow:document.documentElement.scrollWidth>window.innerWidth,methods:rows.length,legacy:!!root?.querySelector('[data-field="attribute:http.method"] .field-hint-label')?.textContent?.includes('(legacy)'),field:root?.querySelector<HTMLInputElement>('[aria-label="Condition 1 field"]')?.value,reads:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore').length}});
export const noHiddenReads=always(()=>state.current.reads<=1);
export const exactLegacyField=always(()=>state.current.step<3||state.current.field==='http.method'&&state.current.phase==='draft');
export const distinctFields=always(()=>state.current.step!==2||state.current.methods===2&&state.current.legacy);
export const fits=always(()=>!state.current.overflow);
export const progress=eventually(()=>state.current.step===3).within(20,'seconds');
const act=registerCustomAction('goo.discoverOTelField',async(document,_window,step:number)=>{
 if(step===1)document.getElementById('exploreRun')?.click();
 if(step===2){(document.getElementById('filterComposer') as HTMLDetailsElement).open=true;const panel=document.querySelector<HTMLDetailsElement>('#exploreFilterEditor .filter-observations');if(panel)panel.open=true;const search=panel?.querySelector<HTMLInputElement>('input');if(search){search.value='HTTP request method';search.dispatchEvent(new Event('input',{bubbles:true}))}panel?.querySelectorAll<HTMLDetailsElement>('details[data-field]').forEach(d=>d.open=true)}
 if(step===3)document.querySelector<HTMLButtonElement>('#exploreFilterEditor button[aria-label=\'Add attribute:http.method "GET"\']')?.click();
 document.documentElement.dataset.hintStep=String(step)
});
export const choices=actions(()=>{const s=state.current;if(!s.ready)return['Wait'];if(s.step===0)return[act(1)];if(s.step===1)return s.phase==='applied'?[act(2)]:['Wait'];return s.step===2?[act(3)]:['Wait']});
