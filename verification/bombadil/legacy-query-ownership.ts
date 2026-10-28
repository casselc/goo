// Bounded source-read/navigation campaign. Never saves, exports or changes workload.
import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>({
 ready:!!document.querySelector('[data-workspace][aria-current=page]'),
 route:document.querySelector('[data-workspace][aria-current=page]')?.getAttribute('data-workspace'),
 probe:(window as any).__gooLegacyOwnerProbe||null,
 operationReady:!(document.querySelector('#runOperations') as HTMLButtonElement)?.disabled,
 httpVisible:!document.querySelector('#definitionContractsPanel')?.hasAttribute('hidden'),
 httpReady:!(document.querySelector('#runHttpPackage') as HTMLButtonElement)?.disabled,
 operationReading:document.querySelector('#operationStatus')?.textContent?.startsWith('Reading operation evidence')||false,
 httpReading:document.querySelector('#httpPackageStatus')?.textContent?.startsWith('Evaluating Go package')||false,
}));
export const noHiddenLegacyReads=always(()=>!state.current.probe||state.current.probe.hidden===0);
export const sourceGETsHaveCancellationCapabilities=always(()=>!state.current.probe||state.current.probe.missingIDs===0);
export const hiddenOwnersStopReading=always(()=>(state.current.route==='operations'||!state.current.operationReading)&&((state.current.route==='definitions'&&state.current.httpVisible)||!state.current.httpReading));
export const ownerProbeInstalled=eventually(()=>!!state.current.probe).within(10,'seconds');
export const sourceReadExercised=eventually(()=>(state.current.probe?.starts||0)>0).within(15,'seconds');
const act=registerCustomAction('goo.legacyOwner',async(document,window,command:string)=>{
 const w=window as any;
 if(command==='install'){
  if(w.__gooLegacyOwnerProbe)return;
  const probe=w.__gooLegacyOwnerProbe={hidden:0,missingIDs:0,starts:0};const original=window.fetch.bind(window);
  window.fetch=(input,init)=>{
   const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:(input as Request).url,window.location.href);
   const method=init?.method||((input as Request).method)||'GET';
   if(method==='GET'&&['/api/operations','/api/counter-rates','/api/histograms','/api/http-package'].includes(url.pathname)){
    probe.starts++;const owner=url.pathname==='/api/http-package'?'definitions':'operations';
    if(document.querySelector('[data-workspace][aria-current=page]')?.getAttribute('data-workspace')!==owner||(url.pathname==='/api/http-package'&&document.querySelector('#definitionContractsPanel')?.hasAttribute('hidden')))probe.hidden++;
    if(!/^[a-f0-9-]{36}$/.test(new Headers(init?.headers).get('X-Goo-Query-ID')||''))probe.missingIDs++;
   }
   return original(input,init);
  };return;
 }
 if(command==='inspect'){(document.querySelector('#runOperations') as HTMLButtonElement).click();return}
 if(command==='contracts'){(document.querySelector('#definitionsContracts') as HTMLButtonElement).click();return}
 if(command==='records'){(document.querySelector('#definitionsRecords') as HTMLButtonElement).click();return}
 if(command==='http'){(document.querySelector('#runHttpPackage') as HTMLButtonElement).click();return}
 if(command==='scope'){const field=document.querySelector('#minutes') as HTMLInputElement;field.value=field.value==='4'?'5':'4';field.dispatchEvent(new Event('input',{bubbles:true}));return}
 (document.querySelector('[data-workspace="'+command+'"]') as HTMLButtonElement).click();
});
export const ownerActions=actions(()=>{
 const s=state.current;if(!s.ready)return ['Wait'];if(!s.probe)return [act('install')];
 if(s.probe.starts===0)return s.route==='operations'&&s.operationReady?[act('inspect')]:[act('operations')];
 const commands=['operations','definitions','services','explore'].map(x=>act(x));
 if(s.probe.starts<6){if(s.route==='operations'&&s.operationReady)commands.push(act('inspect'));if(s.route==='definitions'){commands.push(act('contracts'),act('records'));if(s.httpVisible&&s.httpReady)commands.push(act('http'))}}
 if(s.operationReading||s.httpReading)commands.push(act('scope'));
 return commands;
});
