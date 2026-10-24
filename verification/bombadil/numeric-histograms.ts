import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
// Supply a staged real fixture question with numericHistogram enabled. Only the
// single explicitly authorized Run click may execute; all other actions are local.
const state=extract(({document,window})=>{
 const root=document.documentElement,chart=document.querySelector('#numericSummary .numeric-distribution'),bars=chart?.querySelector('goo-time-bars') as HTMLElement&{data?:{points?:Array<{from:string,to:string}>}};
 const canonical=(s:string)=>/^(0|[1-9][0-9]*|-[1-9][0-9]*)$/.test(s);
 return {ready:!!document.querySelector('#numericHistogram'),stage:Number(root.dataset.numericHistogramStage||0),reads:Number(root.dataset.numericHistogramReads||0),allowed:Number(root.dataset.numericHistogramAllowed||0),otherEffects:Number(root.dataset.numericHistogramOther||0),runDisabled:document.querySelector<HTMLButtonElement>('#exploreRun')?.disabled,chartReady:!!chart?.querySelector('goo-time-bars button'),selected:!!chart?.querySelector('[data-range-selected=true]'),exact:(bars?.data?.points||[]).every(p=>canonical(p.from)&&canonical(p.to)&&BigInt(p.from)<BigInt(p.to)),noValid:chart?.textContent?.includes('No valid numeric values to plot.')||false,selectable:chart?.querySelectorAll('goo-time-bars button').length||0,overflow:root.scrollWidth>window.innerWidth,exploreVisible:(document.querySelector('#exploration') as HTMLElement)?.offsetParent!==null};
});
export const noImplicitSourceRead=always(()=>state.current.reads<=state.current.allowed&&state.current.otherEffects===0);
export const exactRanges=always(()=>state.current.exact);
export const noSelectableInvalidPopulation=always(()=>!state.current.noValid||state.current.selectable===0);
export const invalidBindingBlocksRun=always(()=>state.current.stage!==5||state.current.runDisabled===true);
export const noPageOverflow=always(()=>!state.current.overflow);
export const retainedOwnerDoesNotReopen=always(()=>state.current.stage!==9||state.current.exploreVisible===false);
export const completed=eventually(()=>state.current.stage===9&&state.current.reads===1).within(40,'seconds');
const initialize=registerCustomAction('goo.numericHistogramInitialize',async(document)=>{
 for(const selector of ['#measureComposer','.numeric-binding-editor']){const panel=document.querySelector<HTMLDetailsElement>(selector);if(panel)panel.open=true}
 const unit=document.querySelector<HTMLInputElement>('#numericUnit')!;document.documentElement.dataset.numericHistogramUnit=unit.value;
 const win=document.defaultView!,original=win.fetch.bind(win);
 win.fetch=((input,init)=>{const method=init?.method||(input instanceof win.Request?input.method:'GET'),url=typeof input==='string'?input:input instanceof win.Request?input.url:String(input);if(method==='POST'){const root=document.documentElement;if(new URL(url,win.location.href).pathname==='/api/explore')root.dataset.numericHistogramReads=String(Number(root.dataset.numericHistogramReads||0)+1);else root.dataset.numericHistogramOther=String(Number(root.dataset.numericHistogramOther||0)+1)}return original(input,init)}) as typeof win.fetch;
 document.documentElement.dataset.numericHistogramStage='1';
});
const run=registerCustomAction('goo.numericHistogramRun',async(document,_window,next:string)=>{
 const button=document.querySelector<HTMLButtonElement>('#exploreRun')!;if(!button||button.disabled)throw Error('Only a valid explicit Run may execute the real fixture.');document.documentElement.dataset.numericHistogramAllowed=String(Number(document.documentElement.dataset.numericHistogramAllowed||0)+1);button.click();document.documentElement.dataset.numericHistogramStage=next;
});
const brush=registerCustomAction('goo.numericHistogramBrush',async(document,_window,next:string)=>{
 const chart=document.querySelector('#numericSummary .numeric-distribution')!,buttons=chart.querySelectorAll<HTMLButtonElement>('goo-time-bars button'),first=buttons[0];if(!first)throw Error('Real supplied numeric bins are required.');first.focus();first.dispatchEvent(new KeyboardEvent('keydown',{key:' ',bubbles:true}));first.dispatchEvent(new KeyboardEvent('keydown',{key:'ArrowRight',shiftKey:true,bubbles:true}));const focused=document.activeElement!;focused.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true}));document.documentElement.dataset.numericHistogramStage=next;
});
const details=registerCustomAction('goo.numericHistogramDetails',async(document)=>{const panel=document.querySelector<HTMLDetailsElement>('#numericSummary .numeric-distribution details');if(panel)panel.open=true;document.documentElement.dataset.numericHistogramStage='4'});
const toggle=registerCustomAction('goo.numericHistogramToggle',async(document)=>{const checkbox=document.querySelector<HTMLInputElement>('#numericHistogram')!;checkbox.checked=false;checkbox.dispatchEvent(new Event('change',{bubbles:true}));checkbox.checked=true;checkbox.dispatchEvent(new Event('change',{bubbles:true}));document.documentElement.dataset.numericHistogramStage='10'});
const unit=registerCustomAction('goo.numericHistogramUnit',async(document,_window,invalid:boolean)=>{const control=document.querySelector<HTMLInputElement>('#numericUnit')!;control.value=invalid?'':document.documentElement.dataset.numericHistogramUnit||'items';control.dispatchEvent(new Event('input',{bubbles:true}));document.documentElement.dataset.numericHistogramStage=invalid?'5':'6'});
const leave=registerCustomAction('goo.numericHistogramLeave',async(document)=>{document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();document.documentElement.dataset.numericHistogramStage='9'});
export const choices=actions(()=>!state.current.ready?['Wait']:state.current.stage===0?[initialize()]:state.current.stage===1?[run('2')]:state.current.stage===2?(state.current.chartReady?[brush('3')]:['Wait']):state.current.stage===3?[details()]:state.current.stage===4?[toggle()]:state.current.stage===10?[unit(true)]:state.current.stage===5?[unit(false)]:state.current.stage===6?[leave()]:['Wait']);
