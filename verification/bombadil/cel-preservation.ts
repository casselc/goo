// Saved-source exploration only. The scripted journey owns publication and
// dirty-recall choices; this actor never runs a query or writes a definition.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state=extract(({document,window})=>{
 const root=document.querySelector<HTMLElement>('#exploreFilterEditor .cel-expression-editor');
 const mode=root?.querySelector<HTMLSelectElement>('select[aria-label="Filter mode"]');
 const text=root?.querySelector<HTMLTextAreaElement>('textarea');
 const check=[...root?.querySelectorAll<HTMLButtonElement>('button')||[]].find(button=>button.textContent==='Check condition');
 const calls=((window as any).__gooCELPreservationCalls||[]) as {path:string,method:string,baseIdentity?:string}[];
 const planning=calls.filter(call=>['/api/cel/context','/api/cel/plan'].includes(call.path)).length;
 const phase=Number(document.documentElement.dataset.celPreservationStep||0);
 const expected=document.documentElement.dataset.celPreservationSource||'';
 const run=document.querySelector<HTMLButtonElement>('#exploreRun');
 return {version:'goo.cel-preservation-actor/1',phase,ready:!!root,
  originalAvailable:[...document.querySelectorAll<HTMLButtonElement>('button')].some(button=>(button.getAttribute('aria-label')==='Edit original question'||button.textContent==='Edit original question')&&!!button.getClientRects().length),
  mode:mode?.value||'',ownerPhase:root?.dataset.phase||'',text:text?.value||'',expectedSource:expected,
  sourceRestored:document.documentElement.dataset.celPreservationRestored==='true',
  checkReady:!!check&&!check.disabled&&!!check.getClientRects().length,
  textReady:!!text&&!text.disabled&&!!text.getClientRects().length&&!text.closest('[inert]')&&!!document.querySelector<HTMLDetailsElement>('#filterComposer')?.open,
  invalid:text?.getAttribute('aria-invalid')==='true',
  diagnosticAssociated:!!text&&!!text.getAttribute('aria-describedby')?.split(' ').every(id=>!!document.getElementById(id)),
  savedDraft:document.documentElement.dataset.celPreservationDraft||'',
  planning,plans:calls.filter(call=>call.path==='/api/cel/plan').length,
  originalBase:calls.filter(call=>call.path==='/api/cel/plan').every(call=>call.baseIdentity===document.documentElement.dataset.celPreservationBase),
  planningBeforeHidden:Number(document.documentElement.dataset.celPreservationPlanning||0),
  sourceReads:calls.filter(call=>['/api/explore','/api/explore/page','/api/tail','/api/record-aggregates','/api/population/explain','/api/explore/counted-values','/api/inventory','/api/discovery','/api/measurements','/api/metric-discovery'].includes(call.path)).length,
  forbiddenPosts:calls.filter(call=>call.method==='POST'&&call.path.startsWith('/api/')&&!['/api/cel/context','/api/cel/plan'].includes(call.path)&&!/^\/api\/queries\/[^/]+\/cancel$/.test(call.path)).length,
  enabledRun:!!run&&!run.disabled&&!!run.getClientRects().length,
  overflow:document.documentElement.scrollWidth>window.innerWidth,
 };
});
export const noReadOrPublication=always(()=>state.current.sourceReads===0&&state.current.forbiddenPosts===0);
export const accessibleDiagnostic=always(()=>!state.current.ready||state.current.diagnosticAssociated);
export const fits=always(()=>!state.current.overflow);
// This records the independently verified initial restoration; subsequent
// actions deliberately replace draft text. It is not continuous source equality.
export const initialStoredSourceWitness=always(()=>state.current.phase<2||state.current.sourceRestored);
export const retainedInvalidText=always(()=>![4,5].includes(state.current.phase)||state.current.text===state.current.savedDraft);
export const invalidCannotRun=always(()=>!([3,4,5,6].includes(state.current.phase)||state.current.phase===7&&state.current.ownerPhase!=='valid')||!state.current.enabledRun);
export const hiddenCannotPlan=always(()=>state.current.phase!==8||(state.current.plans>=1&&state.current.originalBase&&state.current.planning===state.current.planningBeforeHidden));
export const completed=eventually(()=>state.current.phase===8).within(20,'seconds');

const begin=registerCustomAction('goo.celPreservationOpenOriginal',async(document,window)=>{
 const target=window as any,calls:{path:string,method:string,baseIdentity?:string}[]=[],original=window.fetch;
 target.__gooCELPreservationCalls=calls;
 window.fetch=function(input,options){const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href);let baseIdentity;try{if(url.pathname==='/api/cel/plan'&&typeof options?.body==='string')baseIdentity=JSON.parse(options.body).binding?.baseIdentity}catch{}calls.push({path:url.pathname,method:String(options?.method||(input instanceof Request?input.method:'GET')).toUpperCase(),baseIdentity});return Reflect.apply(original,this,[input,options])};
 const id=window.location.hash.match(/^#exploration-view=([a-f0-9]{64})$/)?.[1];if(!id)throw Error('Saved view seed unavailable');
 const response=await window.fetch('/api/exploration-views/'+id);if(!response.ok)throw Error('Saved view unavailable');const saved=await response.json();
 const metadata=saved.authoring?.[saved.queries?.[0]?.signal];if(typeof metadata?.source!=='string'||!metadata.source)throw Error('Saved condition source unavailable');
 // Native textarea display normalizes line endings; metadata roundtrip identity
 // is independently checked by the scripted journey and domain properties.
 document.documentElement.dataset.celPreservationSource=metadata.source.replace(/\r\n?/g,'\n');document.documentElement.dataset.celPreservationBase=metadata.binding.baseIdentity;
 const open=[...document.querySelectorAll<HTMLButtonElement>('button')].find(button=>(button.getAttribute('aria-label')==='Edit original question'||button.textContent==='Edit original question')&&!!button.getClientRects().length);
 if(!open)throw Error('Open original control unavailable');open.click();document.documentElement.dataset.celPreservationStep='1';
});
const reveal=registerCustomAction('goo.celPreservationRevealSource',async(document)=>{
 const edit=[...document.querySelectorAll<HTMLButtonElement>('button')].find(button=>button.textContent==='Edit question'&&button.getAttribute('aria-expanded')==='false'&&!button.disabled&&!!button.getClientRects().length);edit?.click();
 const filters=document.querySelector<HTMLDetailsElement>('#filterComposer');if(!filters)throw Error('Fixed Filters unavailable');filters.open=true;
 const mode=document.querySelector<HTMLSelectElement>('#exploreFilterEditor .cel-expression-editor select');if(!mode)throw Error('Condition mode unavailable');mode.value='expression';mode.dispatchEvent(new Event('change',{bubbles:true}));
 const text=document.querySelector<HTMLTextAreaElement>('#exploreFilterEditor .cel-expression-editor textarea');if(!text||text.value!==document.documentElement.dataset.celPreservationSource)throw Error('Restored source differs from saved definition');
 document.documentElement.dataset.celPreservationRestored='true';document.documentElement.dataset.celPreservationStep='2';
});
const invalid=registerCustomAction('goo.celPreservationInvalidDraft',async(document,_window,value:string)=>{
 const text=document.querySelector<HTMLTextAreaElement>('#exploreFilterEditor .cel-expression-editor textarea');if(!text)throw Error('Source editor unavailable');text.value=value;text.dispatchEvent(new Event('input',{bubbles:true}));document.documentElement.dataset.celPreservationDraft=value;document.documentElement.dataset.celPreservationStep='3';
});
const conditions=registerCustomAction('goo.celPreservationConditions',async(document)=>{
 const mode=document.querySelector<HTMLSelectElement>('#exploreFilterEditor .cel-expression-editor select');if(!mode)throw Error('Condition mode unavailable');mode.value='conditions';mode.dispatchEvent(new Event('change',{bubbles:true}));document.documentElement.dataset.celPreservationStep='4';
});
const expression=registerCustomAction('goo.celPreservationExpression',async(document)=>{
 const mode=document.querySelector<HTMLSelectElement>('#exploreFilterEditor .cel-expression-editor select');if(!mode)throw Error('Condition mode unavailable');mode.value='expression';mode.dispatchEvent(new Event('change',{bubbles:true}));document.documentElement.dataset.celPreservationStep='5';
});
const restore=registerCustomAction('goo.celPreservationFreshDraft',async(document)=>{
 const text=document.querySelector<HTMLTextAreaElement>('#exploreFilterEditor .cel-expression-editor textarea');if(!text)throw Error('Source editor unavailable');text.value=document.documentElement.dataset.celPreservationSource||'';text.dispatchEvent(new Event('input',{bubbles:true}));document.documentElement.dataset.celPreservationStep='6';
});
const check=registerCustomAction('goo.celPreservationFreshCheck',async(document)=>{
 const button=[...document.querySelectorAll<HTMLButtonElement>('#exploreFilterEditor .cel-expression-editor button')].find(button=>button.textContent==='Check condition');if(!button)throw Error('Check unavailable');button.click();document.documentElement.dataset.celPreservationStep='7';
});
const retire=registerCustomAction('goo.celPreservationRetiredCheck',async(document,window)=>{
 const filters=document.querySelector<HTMLDetailsElement>('#filterComposer'),button=[...document.querySelectorAll<HTMLButtonElement>('#exploreFilterEditor .cel-expression-editor button')].find(button=>button.textContent==='Check condition');if(!filters||!button)throw Error('Checked source unavailable');filters.open=false;
 document.documentElement.dataset.celPreservationPlanning=String(((window as any).__gooCELPreservationCalls||[]).filter((call:{path:string})=>['/api/cel/context','/api/cel/plan'].includes(call.path)).length);
 await button.onclick?.call(button,new MouseEvent('click'));document.documentElement.dataset.celPreservationStep='8';
});
export const choices=actions(()=>{
 const s=state.current;
 if(s.phase===0)return s.originalAvailable?[begin()]:['Wait'];
 if(s.phase===1)return s.ready?[reveal()]:['Wait'];
 if(s.phase===2)return s.textReady&&s.checkReady?[invalid('log.body =='),invalid('log.severity_number >= 17')]:['Wait'];
 if(s.phase===3)return [conditions()];
 if(s.phase===4)return [expression()];
 if(s.phase===5)return s.textReady?[restore()]:['Wait'];
 if(s.phase===6)return s.checkReady?[check()]:['Wait'];
 if(s.phase===7)return s.ownerPhase==='valid'&&s.plans>=1?[retire()]:['Wait'];
 return ['Wait'];
});
