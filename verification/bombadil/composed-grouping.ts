import{always,eventually}from'@antithesishq/bombadil';
import{extract,actions,registerCustomAction}from'@antithesishq/bombadil/browser';
export{noUncaughtExceptions,noUnhandledPromiseRejections}from'@antithesishq/bombadil/browser/defaults/properties';
const s=extract(({document,window})=>{const root=document.querySelector('#exploreFilterEditor');return{step:Number(document.documentElement.dataset.composeStep||0),ready:!!document.getElementById('exploreRun'),phase:document.getElementById('exploration')?.dataset.questionPhase,fields:[...(root?.querySelectorAll('.grouping-fields > button')||[])].map(b=>b.getAttribute('aria-label')),reads:window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname==='/api/explore').length,overflow:document.documentElement.scrollWidth>window.innerWidth}});
export const bounded=always(()=>s.current.fields.length<=2&&!s.current.overflow&&s.current.reads<=1);
export const staged=always(()=>s.current.step<3||s.current.phase==='draft');
export const exact=always(()=>![3,5,8].includes(s.current.step)||s.current.fields.join('|')==='Remove grouping attribute:http.route|Remove grouping attribute:http.method');
export const clearThenCompose=always(()=>s.current.step!==6&&s.current.step!==7||s.current.fields.join('|')===(s.current.step===6?'':'Remove grouping attribute:http.route'));
export const progress=eventually(()=>s.current.step===8).within(20,'seconds');
const act=registerCustomAction('goo.composeGrouping',async(document,_window,step:number)=>{
 if(step===1)document.getElementById('exploreRun')?.click();
 if(step===2)(document.getElementById('filterComposer') as HTMLDetailsElement).open=true;
 function choose(field:string){const pick=document.querySelector<HTMLSelectElement>('#exploreFilterEditor goo-grouping-control select');if(pick){pick.value=field;pick.dispatchEvent(new Event('change',{bubbles:true}))}}
 if(step===3||step===5)choose(JSON.stringify({kind:'attribute',key:'http.method'}));
 if(step===4)document.querySelector<HTMLButtonElement>('#exploreFilterEditor [aria-label="Remove grouping attribute:http.method"]')?.click();
 if(step===6)choose('clear');
 if(step===7)choose(JSON.stringify({kind:'attribute',key:'http.route'}));
 if(step===8){const detail=document.querySelector<HTMLDetailsElement>('#exploreFilterEditor .grouping-custom');if(detail){detail.open=true;const input=detail.querySelector<HTMLInputElement>('input');if(input)input.value='attribute:http.method';detail.querySelector<HTMLFormElement>('form')?.requestSubmit()}}
 document.documentElement.dataset.composeStep=String(step)
});
export const choices=actions(()=>{const v=s.current;if(!v.ready)return['Wait'];if(v.step===0)return[act(1)];if(v.step===1)return v.phase==='applied'?[act(2)]:['Wait'];return v.step<8?[act(v.step+1)]:['Wait']});
