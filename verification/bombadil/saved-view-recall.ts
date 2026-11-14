// Exact definition recall only. No Run, saves, captures or telemetry actions.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state=extract(({document,window})=>{
 const root=document.querySelector<HTMLDetailsElement>('.scope-history');
 const id=document.documentElement.dataset.savedRecallActorId||'';
 const row=[...root?.querySelectorAll<HTMLElement>('.scope-history-row[data-kind="saved-view"]')||[]].find(e=>e.dataset.scopeId==='saved-view:'+id);
 const button=(label:string)=>[...row?.querySelectorAll<HTMLButtonElement>('button')||[]].find(e=>e.textContent===label);
 const stage=button('Stage')||button('Loading…');
 const filter=root?.querySelector<HTMLInputElement>('input[aria-label="Find recent or pinned scopes"]');
 const calls=((window as any).__gooSavedRecallCalls||[]) as {method:string,path:string}[];
 const metadata=calls.filter(c=>c.method==='GET'&&c.path==='/api/exploration-views/'+id).length;
 let localReferences=0,closedReferences=true,persistence=false;
 for(let i=0;i<window.localStorage.length;i++){
  const key=window.localStorage.key(i);if(!key?.startsWith('goo.scope-history.'))continue;
  persistence=true;
  try{const stored=JSON.parse(window.localStorage.getItem(key)||'');closedReferences&&=Object.keys(stored).sort().join(',')==='items,version'&&stored.version===1;for(const entry of stored.items||[]){const intent=entry.intent;if(intent?.kind!=='saved-view')continue;localReferences++;closedReferences&&=Object.keys(entry).sort().join(',')==='intent,pinned,usedAt'&&Object.keys(intent).sort().join(',')==='id,kind,title'&&/^[a-f0-9]{64}$/.test(intent.id)&&typeof intent.title==='string';}}catch{closedReferences=false}
 }
 const ready=(e:Element|null|undefined)=>!!e&&!!e.getClientRects().length&&!e.closest('[inert]')&&!e.matches(':disabled');
 return {version:'goo.saved-view-recall-actor/1',phase:Number(document.documentElement.dataset.savedRecallActorStep||0),ready:!!root,
  id,row:!!row,open:!!root?.open,stageReady:ready(stage),filterReady:ready(filter),pinned:row?.dataset.pinned==='true',
  pending:root?.getAttribute('aria-busy')==='true',pendingControlsDisabled:!root?.getAttribute('aria-busy')||root.getAttribute('aria-busy')!=='true'||!stage||stage.disabled,
  metadata,requests:calls.length,metadataBeforeHidden:Number(document.documentElement.dataset.savedRecallActorMetadata||0),requestsBeforeHidden:Number(document.documentElement.dataset.savedRecallActorRequests||0),
  forbidden:calls.filter(c=>c.path.startsWith('/api/')&&!(c.method==='GET'&&/^\/api\/exploration-views(?:\/[a-f0-9]{64})?$/.test(c.path))&&!/^\/api\/queries\/[^/]+\/cancel$/.test(c.path)).length,
  persistence,localReferences,closedReferences,
  filteredRows:root?.querySelectorAll('.scope-history-row[data-kind="saved-view"]').length||0,
  forgotten:document.documentElement.dataset.savedRecallActorForgotten==='true',
  overflow:document.documentElement.scrollWidth>window.innerWidth,
 };
});
export const noTelemetryOrPublication=always(()=>state.current.forbidden===0);
export const referenceOnlyStorage=always(()=>state.current.closedReferences);
export const pinnedReferencePersisted=always(()=>state.current.phase<3||state.current.phase>=10||state.current.persistence&&state.current.localReferences>=1);
export const pendingStageIsNativeDisabled=always(()=>state.current.pendingControlsDisabled);
export const fits=always(()=>!state.current.overflow);
export const exactStageWitness=always(()=>state.current.phase<6||state.current.metadata>=1);
export const retiredClosedStageCannotLoad=always(()=>state.current.phase!==8||state.current.metadata===state.current.metadataBeforeHidden&&state.current.requests===state.current.requestsBeforeHidden);
export const completed=eventually(()=>state.current.phase===10&&state.current.forgotten&&state.current.metadata>=1&&state.current.persistence).within(22,'seconds');

const begin=registerCustomAction('goo.savedRecallBegin',async(document,window)=>{
 const target=window as any,calls:{method:string,path:string}[]=[],original=window.fetch;
 target.__gooSavedRecallCalls=calls;
 window.fetch=function(input,options){const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href);calls.push({path:url.pathname,method:String(options?.method||(input instanceof Request?input.method:'GET')).toUpperCase()});return Reflect.apply(original,this,[input,options])};
 const id=window.location.hash.match(/^#exploration-view=([a-f0-9]{64})$/)?.[1];if(!id)throw Error('Seed an actual saved view URL');
 document.documentElement.dataset.savedRecallActorId=id;
 const root=document.querySelector<HTMLDetailsElement>('.scope-history');if(!root)throw Error('Recent / Pinned unavailable');
 root.querySelector<HTMLElement>(':scope > summary')?.click();document.documentElement.dataset.savedRecallActorStep='1';
});
const remember=registerCustomAction('goo.savedRecallRememberLocally',async(document)=>{
 const input=document.querySelector<HTMLInputElement>('.scope-history input[type="checkbox"]');if(!input)throw Error('Browser persistence choice unavailable');if(!input.checked)input.click();document.documentElement.dataset.savedRecallActorStep='2';
});
const pin=registerCustomAction('goo.savedRecallPinLocally',async(document)=>{
 const id=document.documentElement.dataset.savedRecallActorId,row=[...document.querySelectorAll<HTMLElement>('.scope-history-row[data-kind="saved-view"]')].find(e=>e.dataset.scopeId==='saved-view:'+id);if(!row)throw Error('Saved reference unavailable');
 const button=[...row.querySelectorAll<HTMLButtonElement>('button')].find(e=>e.textContent==='Pin');button?.click();document.documentElement.dataset.savedRecallActorStep='3';
});
const filter=registerCustomAction('goo.savedRecallFilterLocally',async(document)=>{
 const input=document.querySelector<HTMLInputElement>('.scope-history input[aria-label="Find recent or pinned scopes"]');if(!input)throw Error('Local history filter unavailable');input.value=document.documentElement.dataset.savedRecallActorId||'';input.dispatchEvent(new Event('input',{bubbles:true}));document.documentElement.dataset.savedRecallActorStep='4';
});
const stage=registerCustomAction('goo.savedRecallStageExactReference',async(document)=>{
 const id=document.documentElement.dataset.savedRecallActorId,row=[...document.querySelectorAll<HTMLElement>('.scope-history-row[data-kind="saved-view"]')].find(e=>e.dataset.scopeId==='saved-view:'+id),button=[...row?.querySelectorAll<HTMLButtonElement>('button')||[]].find(e=>e.textContent==='Stage');if(!button)throw Error('Exact Stage unavailable');button.click();document.documentElement.dataset.savedRecallActorStep='5';
});
const reopen=registerCustomAction('goo.savedRecallReopenPicker',async(document)=>{
 const root=document.querySelector<HTMLDetailsElement>('.scope-history');if(!root)throw Error('Recent picker unavailable');if(!root.open)root.querySelector<HTMLElement>(':scope > summary')?.click();document.documentElement.dataset.savedRecallActorStep='6';
});
const retire=registerCustomAction('goo.savedRecallRetiredStage',async(document,window)=>{
 const root=document.querySelector<HTMLDetailsElement>('.scope-history'),id=document.documentElement.dataset.savedRecallActorId,row=[...root?.querySelectorAll<HTMLElement>('.scope-history-row[data-kind="saved-view"]')||[]].find(e=>e.dataset.scopeId==='saved-view:'+id),button=[...row?.querySelectorAll<HTMLButtonElement>('button')||[]].find(e=>e.textContent==='Stage');if(!root||!button)throw Error('Retained Stage unavailable');
 const handler=button.onclick;if(typeof handler!=='function')throw Error('Retained Stage callback unavailable');root.querySelector<HTMLElement>(':scope > summary')?.click();
 document.documentElement.dataset.savedRecallActorMetadata=String(((window as any).__gooSavedRecallCalls||[]).filter((c:{method:string,path:string})=>c.method==='GET'&&c.path==='/api/exploration-views/'+id).length);
 document.documentElement.dataset.savedRecallActorRequests=String(((window as any).__gooSavedRecallCalls||[]).length);
 await handler.call(button,new MouseEvent('click'));document.documentElement.dataset.savedRecallActorStep='8';
});
const reopenForForget=registerCustomAction('goo.savedRecallReopenForForget',async(document)=>{
 const root=document.querySelector<HTMLDetailsElement>('.scope-history');if(!root)throw Error('Picker unavailable');if(!root.open)root.querySelector<HTMLElement>(':scope > summary')?.click();document.documentElement.dataset.savedRecallActorStep='9';
});
const forget=registerCustomAction('goo.savedRecallForgetLocally',async(document)=>{
 const id=document.documentElement.dataset.savedRecallActorId,row=[...document.querySelectorAll<HTMLElement>('.scope-history-row[data-kind="saved-view"]')].find(e=>e.dataset.scopeId==='saved-view:'+id),button=[...row?.querySelectorAll<HTMLButtonElement>('button')||[]].find(e=>e.textContent==='Forget');if(!button)throw Error('Forget unavailable');button.click();document.documentElement.dataset.savedRecallActorForgotten=String(![...document.querySelectorAll<HTMLElement>('.scope-history-row[data-kind="saved-view"]')].some(e=>e.dataset.scopeId==='saved-view:'+id));document.documentElement.dataset.savedRecallActorStep='10';
});
export const choices=actions(()=>{
 const s=state.current;
 if(s.phase===0)return s.ready?[begin()]:['Wait'];
 if(s.phase===1)return s.open&&s.row&&s.stageReady?[remember()]:['Wait'];
 if(s.phase===2)return s.open&&s.row?[pin()]:['Wait'];
 if(s.phase===3)return s.open&&s.pinned&&s.filterReady?[filter()]:['Wait'];
 if(s.phase===4)return s.stageReady&&s.filteredRows===1?[stage()]:['Wait'];
 if(s.phase===5)return !s.pending&&!s.open&&s.metadata>=1?[reopen()]:['Wait'];
 if(s.phase===6)return s.open&&s.stageReady?[retire()]:['Wait'];
 if(s.phase===8)return !s.open?[reopenForForget()]:['Wait'];
 if(s.phase===9)return s.open&&s.row&&s.stageReady?[forget()]:['Wait'];
 return ['Wait'];
});
