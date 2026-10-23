// Staged selector exploration; Start and Apply are deliberately outside this campaign.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({step:Number(document.documentElement.dataset.filterStep||0),rows:document.querySelectorAll('.stream-entry .live-tail tbody>tr').length,phase:document.querySelector<HTMLElement>('.stream-entry .live-tail')?.dataset.state,overflow:document.documentElement.scrollWidth>window.innerWidth,invalid:document.querySelector('.filter-editor [role="status"]')?.textContent?.includes('unavailable'),blocked:[...document.querySelectorAll<HTMLButtonElement>('.stream-entry button')].find(b=>b.textContent==='Apply scope')?.disabled}));
export const noExecution=always(()=>state.current.phase==='paused'&&state.current.rows===0&&!state.current.overflow);
export const invalidBlocked=always(()=>!state.current.invalid||state.current.blocked===true);
export const progress=eventually(()=>state.current.step===9).within(25,'seconds');
const edit=registerCustomAction('goo.editFilter',async(document,_window,step:number)=>{
 const panel=document.querySelector<HTMLDetailsElement>('.entry-filter-options');if(panel)panel.open=step!==9;
 function select(label:string,value:string){const s=document.querySelector<HTMLSelectElement>('.filter-editor [aria-label="'+label+'"]');if(s){s.value=value;s.dispatchEvent(new Event('change',{bubbles:true}))}}
 function click(name:string){[...document.querySelectorAll<HTMLButtonElement>('.filter-editor button')].find(b=>b.textContent===name)?.click()}
 if(step<=2)select('Group 1 match',step===1?'all':'any');
 if(step===3||step===4)select('Condition 2 operator',step===3?'missing':'eq');
 if(step===5)click('Add nested group 1');
 if(step===6||step===7){const f=document.querySelector<HTMLInputElement>('.filter-editor [aria-label="Condition 5 field"]');if(f){f.value=step===6?'bad-column':'service';f.dispatchEvent(new Event('input',{bubbles:true}))}}
 if(step===8)click('Remove condition 4');
 document.documentElement.dataset.filterStep=String(step)
});
export const choices=actions(()=>state.current.step<9?[edit(state.current.step+1)]:['Wait']);
