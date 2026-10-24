import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
// Run on a staged uint64 numeric question URL. Actor edits supplied controls only;
// no action executes a source read or publication.
const state=extract(({document,window})=>({
 ready:!!document.querySelector('#exploreFilterEditor [aria-label="Condition 1 value"]'),
 stage:Number(document.documentElement.dataset.typedFilterStage||0),
 effects:Number(document.documentElement.dataset.typedFilterEffects||0),
 runDisabled:document.querySelector<HTMLButtonElement>('#exploreRun')?.disabled,
 value:document.querySelector<HTMLInputElement>('#exploreFilterEditor [aria-label="Condition 1 value"]')?.value,
 type:document.querySelector<HTMLSelectElement>('#exploreFilterEditor [aria-label="Condition 1 type"]')?.value,
 unit:document.querySelector<HTMLInputElement>('#exploreFilterEditor [aria-label="Condition 1 unit"]')?.value,
 overflow:document.documentElement.scrollWidth>window.innerWidth,
}));
export const noImplicitEffects=always(()=>state.current.effects===0);
export const compact=always(()=>!state.current.overflow);
export const invalidBlocked=always(()=>![2,3].includes(state.current.stage)||state.current.runDisabled===true);
export const bindingNotInferred=always(()=>state.current.stage<1||state.current.type==='uint64'&&state.current.unit==='items');
export const completed=eventually(()=>state.current.stage===4&&state.current.value==='9007199254740993'&&!state.current.runDisabled).within(25,'seconds');
const initialize=registerCustomAction('goo.typedFilterInitialize',async(document)=>{
 const panel=document.querySelector<HTMLDetailsElement>('#filterComposer');if(panel)panel.open=true;
 const win=document.defaultView!,fetch=win.fetch.bind(win);
 win.fetch=((input,init)=>{if(init?.method==='POST')document.documentElement.dataset.typedFilterEffects=String(Number(document.documentElement.dataset.typedFilterEffects||0)+1);return fetch(input,init)}) as typeof win.fetch;
 document.documentElement.dataset.typedFilterStage='1';
});
const edit=registerCustomAction('goo.typedFilterEdit',async(document,_window,value:string,stage:string)=>{
 const control=document.querySelector<HTMLInputElement>('#exploreFilterEditor [aria-label="Condition 1 value"]')!;if(!control||control.offsetParent===null)throw Error('Numeric editor must be visible before editing.');
 control.value=value;control.dispatchEvent(new Event('input',{bubbles:true}));document.documentElement.dataset.typedFilterStage=stage;
});
export const choices=actions(()=>!state.current.ready?['Wait']:state.current.stage===0?[initialize()]:state.current.stage===1?[edit('01','2')]:state.current.stage===2?[edit('18446744073709551616','3')]:state.current.stage===3?[edit('9007199254740993','4')]:['Wait']);
