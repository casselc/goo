// Read-only inventory discovery and local exploration; no saves or workload mutations.
import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const inventory=extract(({document})=>{
 const el=(id:string)=>document.getElementById(id),read=(id:string)=>{try{return JSON.parse(el(id)?.textContent||'null')}catch{return null}};
 const capture=read('inventoryExplain'),opened=read('exploreExplain');let accounting=true,rendered=true;
 if(capture){try{accounting=capture.rows.reduce((s:bigint,r:any)=>s+BigInt(r.count),0n)===BigInt(capture.count)&&capture.rows.every((r:any)=>r.buckets.length===8&&r.buckets.reduce((s:bigint,b:string)=>s+BigInt(b),0n)===BigInt(r.count)&&BigInt(r.errors)<=BigInt(r.count));for(const row of document.querySelectorAll<HTMLElement>('#inventoryRows tbody tr')){const cells=row.querySelectorAll('td');rendered=rendered&&cells[1]?.textContent===row.dataset.count&&cells[2]?.textContent===row.dataset.errors&&JSON.parse(row.dataset.buckets||'[]').reduce((s:bigint,b:string)=>s+BigInt(b),0n)===BigInt(row.dataset.count||'-1')}}catch{accounting=false}}
 const intent=read('inventoryIntent'),scope=!intent||!opened||opened.question.environment===intent.environment&&opened.question.fromMs===intent.fromMs&&opened.question.toMs===intent.toMs&&JSON.stringify(opened.question.where)===JSON.stringify(intent.where);
 return {version:'goo.inventory.v1',stage:document.documentElement.dataset.gooInventoryStage||'initial',ready:!!el('inventoryRun'),busy:(el('inventoryRun') as HTMLButtonElement)?.disabled,exploreBusy:(el('exploreRun') as HTMLButtonElement)?.disabled,capture,opened,accounting,rendered,scope,hasRows:!!document.querySelector('#inventoryRows tbody button'),staleCleared:!(el('inventoryStatus')?.textContent||'').startsWith('Scope changed')||!capture&&!document.querySelector('#inventoryRows tbody tr')};
});
export const ready=eventually(()=>inventory.current.ready).within(10,'seconds');
export const prefixComplete=eventually(()=>inventory.current.stage==='complete').within(25,'seconds');
export const exactPopulation=always(()=>inventory.current.accounting&&inventory.current.rendered);
export const fixedDrilldown=always(()=>inventory.current.scope);
export const editedScopeClearsEvidence=always(()=>inventory.current.staleCleared);
const act=registerCustomAction('goo.inventory',async(document,window,command:string)=>{
 const el=(id:string)=>document.getElementById(id),click=(id:string)=>el(id)?.click(),set=(id:string,v:string)=>{const input=el(id) as HTMLInputElement;if(input){input.value=v;input.dispatchEvent(new Event('input',{bubbles:true}))}};
 const stage=(s:string)=>document.documentElement.dataset.gooInventoryStage=s;
 if(command==='start'){const env=new URLSearchParams(window.location.hash.slice(1)).get('environment')||document.documentElement.dataset.gooInventoryEnvironment;document.querySelector<HTMLElement>('[data-workspace="services"]')?.click();if(env)set('inventoryEnvironment',env);click('inventoryRun');stage('catalog')}
 if(['searchMissing','searchClear','filter','clear'].includes(command)){const control=document.querySelector('#operationSearch');const parent=control?.closest('details');if(parent)(parent as HTMLDetailsElement).open=true}
 if(command==='searchMissing')set('operationSearch','__missing__');
 if(command==='searchClear')set('operationSearch','');
 if(command==='filter'){set('operationSearch','__missing__');stage('filtered')}
 if(command==='clear'){set('operationSearch','');stage('cleared')}
 if(command==='open'){const capture=JSON.parse(el('inventoryExplain')?.textContent||'null'),row=capture.rows[0],q=(await import(new URL('/inventory.mjs',window.location.href).href)).question(capture,row);let marker=el('inventoryIntent');if(!marker){marker=document.createElement('script');marker.id='inventoryIntent';marker.type='application/json';document.body.append(marker)}marker.textContent=JSON.stringify(q);document.querySelector<HTMLElement>('#inventoryRows tbody button')?.click();stage('opening')}
 if(command==='browse'){const b=[...document.querySelectorAll<HTMLElement>('#serviceCatalog button')].find(b=>b.textContent?.startsWith('Open service '));if(b){b.click();stage('capturing')}}if(command==='complete'){stage('complete')}
 if(command==='reset'){el('inventoryIntent')?.remove();document.querySelector<HTMLElement>('[data-workspace="services"]')?.click();click('inventoryRun')}
 if(command==='edit'){set('inventoryEnvironment',(el('inventoryEnvironment') as HTMLInputElement).value+'-no-data')}
});
export const explore=actions(()=>{const s=inventory.current;if(!s.ready)return ['Wait'];if(s.stage==='initial')return[act('start')];if(s.busy||s.exploreBusy)return['Wait'];if(s.stage==='catalog')return [act('browse')];if(s.stage==='capturing')return s.hasRows?[act('filter')]:['Wait'];if(s.stage==='filtered')return[act('clear')];if(s.stage==='cleared')return[act('open')];if(s.stage==='opening')return s.opened?[act('complete')]:['Wait'];return[act('searchMissing'),act('searchClear'),act('reset'),act('edit')];});
