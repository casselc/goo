import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>({phase:Number(document.documentElement.dataset.containerReadyPhase||0),source:Number(document.documentElement.dataset.containerReadySource||0),plans:Number(document.documentElement.dataset.containerReadyPlans||0),
 target:document.documentElement.dataset.containerReadyTarget||'',usageKey:document.documentElement.dataset.containerReadyUsageKey||'',ready:!!document.querySelector('tr[data-pod-variant]'),
 loaded:[...document.querySelectorAll('.pod-metrics-panel [role=status]')].some(e=>e.textContent?.includes('Independent reported pod observations')),
 nativeReady:!!document.querySelector('#measureRun')&&!document.querySelector<HTMLButtonElement>('#measureRun')!.disabled,
 nativeComplete:!!document.querySelector('#measureResult')?.children.length&&!document.querySelector<HTMLButtonElement>('#measureCapture')?.disabled,
 terminal:document.documentElement.dataset.containerReadyTerminal==='true'}));
export const bounded=always(()=>state.current.source<=4&&state.current.plans<=2);
export const complete=eventually(()=>state.current.terminal).within(25,'seconds');
const step=registerCustomAction('goo.containerReadyStep',async(document,window)=>{
 const root=document.documentElement,phase=Number(root.dataset.containerReadyPhase||0),panel=()=>document.querySelector<HTMLElement>('.pod-metrics-panel[data-pod-metrics="'+root.dataset.containerReadyTarget+'"]')!;
 if(phase===0){
  const environment=new URLSearchParams(window.location.hash.slice(1)).get('containerReadyEnvironment')!,original=window.fetch.bind(window);
  window.fetch=async(input,init)=>{const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href),method=init?.method||(input instanceof Request?input.method:'GET');
   if(method==='POST'&&url.pathname.startsWith('/api/')&&!url.pathname.endsWith('/cancel')){const key=url.pathname.endsWith('/plan')?'containerReadyPlans':'containerReadySource';root.dataset[key]=String(Number(root.dataset[key]||0)+1)}
   const response=await original(input,init);
   if(response.ok&&method==='POST'&&url.pathname==='/api/pod-discovery'){const data=await response.clone().json(),row=data.observations.find((r:any)=>r.resource['container.id']==='reported-container-not-ready'&&r.resourceSchemaURL===''&&!Object.hasOwn(r.resource,'service.name')&&!Object.hasOwn(r.resource,'k8s.cluster.uid')&&!Object.hasOwn(r.resource,'k8s.namespace.name'));if(!row)throw Error('Known-empty-schema pod resource absent');root.dataset.containerReadyTarget=row.key}
   if(response.ok&&method==='POST'&&url.pathname==='/api/pod-dashboard/plan'){const data=await response.clone().json();if(data.widgets.length!==1||data.widgets.some((w:any)=>w.profile!=='collector-k8scluster-v0.137.0'))throw Error('Readiness default differs');root.dataset.containerReadyUsageKey=data.widgets.find((w:any)=>w.question.definition.source.metric==='k8s.container.ready').key}
   if(response.ok&&method==='POST'&&url.pathname==='/api/measurements'){const data=await response.clone().json(),b=data.question.definition.source,sample=data.states?.series?.[0]?.samples.at(-1);if(data.question.definition.kind!=='state'||data.question.definition.stateProfile!=='k8scluster-container-ready/0.137.0'||b.resource.kind!=='kubernetes-container'||b.resource.id!=='reported-container-not-ready'||b.version!=='0.137.0'||sample?.code!==0||sample?.label!=='Not ready'||sample?.state!=='observed')throw Error('Actual reported Not ready differs')}
   return response;
  };
  document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();document.querySelector<HTMLButtonElement>('#entitiesKubernetes')!.click();const input=document.querySelector<HTMLInputElement>('[aria-label="Kubernetes environment"]')!;input.value=environment;input.dispatchEvent(new Event('input',{bubbles:true}));
  const minutes=document.querySelector<HTMLInputElement>('#minutes')!;minutes.value='2';minutes.dispatchEvent(new Event('input',{bubbles:true}));[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Find pods')!.click();
 }else if(phase===1){const input=document.querySelector<HTMLInputElement>('[aria-label="Find reported pod resource"]')!;input.value='reported-container-not-ready';input.dispatchEvent(new Event('input',{bubbles:true}));if(root.dataset.containerReadySource!=='1')throw Error('Local search read source')}
 else if(phase===2){const button=document.querySelector<HTMLButtonElement>('button[data-pod-observe="'+root.dataset.containerReadyTarget+'"]')!;button.scrollIntoView({block:'center'});button.click()}
 else if(phase===3){if(root.dataset.containerReadySource!=='3')throw Error('Default budget differs');const summary=panel().querySelector<HTMLElement>('[data-pod-metric="'+root.dataset.containerReadyUsageKey+'"] details > summary')!;summary.scrollIntoView({block:'center'});summary.click()}
 else if(phase===4){if(root.dataset.containerReadySource!=='3')throw Error('Disclosure read source');const button=panel().querySelector<HTMLButtonElement>('[data-pod-metric="'+root.dataset.containerReadyUsageKey+'"] > button')!;button.scrollIntoView({block:'center'});button.focus();(window as any).__containerReadyActorOpener=button;button.click()}
 else if(phase===5){if(root.dataset.containerReadySource!=='3')throw Error('Native staging read source');document.querySelector<HTMLButtonElement>('#measureRun')!.click()}
 else if(phase===6)[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Back to Kubernetes')!.click();
 else if(phase===7){if(root.dataset.containerReadySource!=='4'||panel().querySelectorAll('.pod-metric-widget').length!==1||document.querySelector<HTMLInputElement>('[aria-label="Find reported pod resource"]')!.value!=='reported-container-not-ready'||document.activeElement!==(window as any).__containerReadyActorOpener)throw Error('Retained native panel/focus/read budget differs');root.dataset.containerReadyTerminal='true'}
 root.dataset.containerReadyPhase=String(phase+1);
});
export const workflow=actions(()=>{const s=state.current;if(s.phase>=8||s.phase===1&&(!s.ready||!s.target)||s.phase===3&&(!s.loaded||!s.usageKey)||s.phase===5&&!s.nativeReady||s.phase===6&&!s.nativeComplete)return ['Wait'];return [step()];});
