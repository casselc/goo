// Additional-condition exploration only: no Apply, Run, saves or publications.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state=extract(({document,window})=>{
 const filters=document.querySelector<HTMLDetailsElement>('.entry-filter-options');
 const root=filters?.querySelector<HTMLElement>('.cel-expression-editor');
 const mode=root?.querySelector<HTMLSelectElement>('select[aria-label="Filter mode"]');
 const text=root?.querySelector<HTMLTextAreaElement>('textarea');
 const check=[...root?.querySelectorAll<HTMLButtonElement>('button')||[]].find(button=>button.textContent==='Check condition');
 const calls=((window as any).__gooCELActorCalls||[]) as {path:string,method:string}[];
 const paths=calls.map(entry=>entry.path);
 const phase=Number(document.documentElement.dataset.celActorStep||0);
 return {version:'goo.cel-editor-actor/1',phase,filters:!!filters,ready:!!root,open:!!filters?.open,
  mode:mode?.value||'',ownerPhase:root?.dataset.phase||'',text:text?.value||'',
  invalid:text?.getAttribute('aria-invalid')==='true',checkReady:!!check&&!check.disabled,
  diagnosticAssociated:!!text&&!!text.getAttribute('aria-describedby')?.split(' ').every(id=>!!document.getElementById(id)),
  savedDraft:document.documentElement.dataset.celActorDraft||'',
  planning:paths.filter(path=>path==='/api/cel/context'||path==='/api/cel/plan').length,
  contexts:paths.filter(path=>path==='/api/cel/context').length,plans:paths.filter(path=>path==='/api/cel/plan').length,
  forbiddenPosts:calls.filter(entry=>entry.method==='POST'&&entry.path.startsWith('/api/')&&!['/api/cel/context','/api/cel/plan'].includes(entry.path)).length,
  planningBeforeHidden:Number(document.documentElement.dataset.celActorPlanning||0),
  sourceReads:paths.filter(path=>['/api/explore','/api/explore/page','/api/tail','/api/record-aggregates'].includes(path)).length,
  overflow:document.documentElement.scrollWidth>window.innerWidth,
  enabledPrimary:[...document.querySelectorAll<HTMLButtonElement>('.stream-entry button')].some(button=>['Apply scope','Start stream'].includes(button.textContent||'')&&!button.disabled&&!!button.getClientRects().length),
 };
});
export const noSourceRead=always(()=>state.current.sourceReads===0&&state.current.forbiddenPosts===0);
export const accessibleDiagnostic=always(()=>!state.current.ready||state.current.diagnosticAssociated);
export const fits=always(()=>!state.current.overflow);
export const conditionsByDefault=always(()=>state.current.phase!==1||state.current.mode==='conditions');
export const invalidCannotApply=always(()=>!state.current.open||!['draft','checking','invalid'].includes(state.current.ownerPhase)||!state.current.enabledPrimary);
export const invalidDraftSurvives=always(()=>![5,6].includes(state.current.phase)||state.current.text===state.current.savedDraft);
export const hiddenCallbackCannotCheck=always(()=>state.current.phase!==10||(state.current.contexts>=1&&state.current.plans>=2&&state.current.planning===state.current.planningBeforeHidden));
export const completed=eventually(()=>state.current.phase===10).within(20,'seconds');

const open=registerCustomAction('goo.celOpenFilters',async(document,window)=>{
 // Observe invocation paths before contextual preparation starts. Preserve the
 // supplied request, receiver and response; unlike resource timing this also
 // records pending or aborted requests and needs no timing-buffer support.
 const target=window as any,calls:{path:string,method:string}[]=[],original=window.fetch;
 target.__gooCELActorCalls=calls;
 window.fetch=function(input,options){const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href);calls.push({path:url.pathname,method:String(options?.method||(input instanceof Request?input.method:'GET')).toUpperCase()});return Reflect.apply(original,this,[input,options])};
 const filters=document.querySelector<HTMLDetailsElement>('.entry-filter-options');if(!filters)throw Error('Activity Filters unavailable');filters.open=true;document.documentElement.dataset.celActorStep='1';
});
const expression=registerCustomAction('goo.celExpressionMode',async(document)=>{
 const mode=document.querySelector<HTMLSelectElement>('.entry-filter-options .cel-expression-editor select');if(!mode)throw Error('Filter mode unavailable');mode.value='expression';mode.dispatchEvent(new Event('change',{bubbles:true}));document.documentElement.dataset.celActorStep='2';
});
const invalid=registerCustomAction('goo.celInvalidDraft',async(document,_window,value:string)=>{
 const text=document.querySelector<HTMLTextAreaElement>('.entry-filter-options .cel-expression-editor textarea');if(!text)throw Error('Expression unavailable');text.value=value;text.dispatchEvent(new Event('input',{bubbles:true}));document.documentElement.dataset.celActorDraft=value;document.documentElement.dataset.celActorStep='3';
});
const checkInvalid=registerCustomAction('goo.celCheckInvalid',async(document)=>{
 [...document.querySelectorAll<HTMLButtonElement>('.entry-filter-options .cel-expression-editor button')].find(button=>button.textContent==='Check condition')?.click();document.documentElement.dataset.celActorStep='4';
});
const conditions=registerCustomAction('goo.celConditionsMode',async(document)=>{
 const mode=document.querySelector<HTMLSelectElement>('.entry-filter-options .cel-expression-editor select');if(!mode)throw Error('Filter mode unavailable');mode.value='conditions';mode.dispatchEvent(new Event('change',{bubbles:true}));document.documentElement.dataset.celActorStep='5';
});
const restore=registerCustomAction('goo.celRestoreDraft',async(document)=>{
 const mode=document.querySelector<HTMLSelectElement>('.entry-filter-options .cel-expression-editor select');if(!mode)throw Error('Filter mode unavailable');mode.value='expression';mode.dispatchEvent(new Event('change',{bubbles:true}));document.documentElement.dataset.celActorStep='6';
});
const valid=registerCustomAction('goo.celValidDraft',async(document,_window,value:string)=>{
 const text=document.querySelector<HTMLTextAreaElement>('.entry-filter-options .cel-expression-editor textarea');if(!text)throw Error('Expression unavailable');text.value=value;text.dispatchEvent(new Event('input',{bubbles:true}));document.documentElement.dataset.celActorStep='7';
});
const checkValid=registerCustomAction('goo.celCheckValid',async(document)=>{
 [...document.querySelectorAll<HTMLButtonElement>('.entry-filter-options .cel-expression-editor button')].find(button=>button.textContent==='Check condition')?.click();document.documentElement.dataset.celActorStep='8';
});
const retire=registerCustomAction('goo.celRetiredCheck',async(document,window)=>{
 const filters=document.querySelector<HTMLDetailsElement>('.entry-filter-options');const check=[...filters?.querySelectorAll<HTMLButtonElement>('.cel-expression-editor button')||[]].find(button=>button.textContent==='Check condition');
 if(!filters||!check)throw Error('Checked condition unavailable');filters.open=false;
 document.documentElement.dataset.celActorPlanning=String(((window as any).__gooCELActorCalls||[]).filter((entry:{path:string})=>['/api/cel/context','/api/cel/plan'].includes(entry.path)).length);
 // Invoke the retained handler directly after native disclosure closes. Its
 // visibility guard must refuse even if a driver bypasses normal click dispatch.
 await check.onclick?.call(check,new MouseEvent('click'));
 document.documentElement.dataset.celActorStep='10';
});
export const choices=actions(()=>{
 const s=state.current;
 if(s.phase===0)return s.filters?[open()]:['Wait'];
 if(s.phase===1)return s.ready?[expression()]:['Wait'];
 if(s.phase===2)return [invalid('log.body =='),invalid('log.severity_number >= 17')];
 if(s.phase===3)return s.checkReady?[checkInvalid()]:['Wait'];
 if(s.phase===4)return s.ownerPhase==='invalid'&&s.invalid?[conditions()]:['Wait'];
 if(s.phase===5)return [restore()];
 if(s.phase===6)return [valid('log.body.contains("ordinary")'),valid('log.severity_number >= 17u'),valid('log.body != "雪"')];
 if(s.phase===7)return s.checkReady?[checkValid()]:['Wait'];
 if(s.phase===8)return s.ownerPhase==='valid'?[retire()]:['Wait'];
 return ['Wait'];
});
