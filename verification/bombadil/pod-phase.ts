import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>({phase:Number(document.documentElement.dataset.podPhasePhase||0),source:Number(document.documentElement.dataset.podPhaseSource||0),plans:Number(document.documentElement.dataset.podPhasePlans||0),
 target:document.documentElement.dataset.podPhaseTarget||'',usageKey:document.documentElement.dataset.podPhaseUsageKey||'',ready:!!document.querySelector('tr[data-pod-variant]'),
 loaded:[...document.querySelectorAll('.pod-metrics-panel [role=status]')].some(e=>e.textContent?.includes('Independent reported pod observations')),
 nativeReady:!!document.querySelector('#measureRun')&&!document.querySelector<HTMLButtonElement>('#measureRun')!.disabled,
 nativeComplete:!!document.querySelector('#measureResult')?.children.length&&!document.querySelector<HTMLButtonElement>('#measureCapture')?.disabled,
 terminal:document.documentElement.dataset.podPhaseTerminal==='true'}));
export const bounded=always(()=>state.current.source<=4&&state.current.plans<=2);
export const complete=eventually(()=>state.current.terminal).within(25,'seconds');
const step=registerCustomAction('goo.podPhaseStep',async(document,window)=>{
 const root=document.documentElement,phase=Number(root.dataset.podPhasePhase||0),panel=()=>document.querySelector<HTMLElement>('.pod-metrics-panel[data-pod-metrics="'+root.dataset.podPhaseTarget+'"]')!;
 if(phase===0){
  const environment=new URLSearchParams(window.location.hash.slice(1)).get('podPhaseEnvironment')!,original=window.fetch.bind(window);
  window.fetch=async(input,init)=>{const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href),method=init?.method||(input instanceof Request?input.method:'GET');
   if(method==='POST'&&url.pathname.startsWith('/api/')&&!url.pathname.endsWith('/cancel')){const key=url.pathname.endsWith('/plan')?'podPhasePlans':'podPhaseSource';root.dataset[key]=String(Number(root.dataset[key]||0)+1)}
   const response=await original(input,init);
   if(response.ok&&method==='POST'&&url.pathname==='/api/pod-discovery'){const data=await response.clone().json(),row=data.observations.find((r:any)=>r.resource['k8s.pod.uid']==='reported-unknown'&&r.resourceSchemaURL===''&&!Object.hasOwn(r.resource,'service.name')&&!Object.hasOwn(r.resource,'k8s.cluster.uid')&&!Object.hasOwn(r.resource,'k8s.namespace.name'));if(!row)throw Error('Known-empty-schema pod resource absent');root.dataset.podPhaseTarget=row.key}
   if(response.ok&&method==='POST'&&url.pathname==='/api/pod-dashboard/plan'){const data=await response.clone().json();if(data.widgets.length!==1||data.widgets.some((w:any)=>w.profile!=='collector-k8scluster-v0.137.0'))throw Error('Phase default differs');root.dataset.podPhaseUsageKey=data.widgets.find((w:any)=>w.question.definition.source.metric==='k8s.pod.phase').key}
   if(response.ok&&method==='POST'&&url.pathname==='/api/measurements'){const data=await response.clone().json(),b=data.question.definition.source,sample=data.states?.series?.[0]?.samples.at(-1);if(data.question.definition.kind!=='state'||data.question.definition.stateProfile!=='k8scluster-pod-phase/0.137.0'||b.resource.kind!=='kubernetes-pod'||b.resource.id!=='reported-unknown'||b.version!=='0.137.0'||sample?.code!==5||sample?.label!=='Unknown'||sample?.state!=='observed')throw Error('Actual reported Unknown phase differs')}
   return response;
  };
  document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();document.querySelector<HTMLButtonElement>('#entitiesKubernetes')!.click();const input=document.querySelector<HTMLInputElement>('[aria-label="Kubernetes environment"]')!;input.value=environment;input.dispatchEvent(new Event('input',{bubbles:true}));
  const minutes=document.querySelector<HTMLInputElement>('#minutes')!;minutes.value='2';minutes.dispatchEvent(new Event('input',{bubbles:true}));[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Find pods')!.click();
 }else if(phase===1){const input=document.querySelector<HTMLInputElement>('[aria-label="Find reported pod resource"]')!;input.value='reported-unknown';input.dispatchEvent(new Event('input',{bubbles:true}));if(root.dataset.podPhaseSource!=='1')throw Error('Local search read source')}
 else if(phase===2){const button=document.querySelector<HTMLButtonElement>('button[data-pod-observe="'+root.dataset.podPhaseTarget+'"]')!;button.scrollIntoView({block:'center'});button.click()}
 else if(phase===3){if(root.dataset.podPhaseSource!=='3')throw Error('Default budget differs');const summary=panel().querySelector<HTMLElement>('[data-pod-metric="'+root.dataset.podPhaseUsageKey+'"] details > summary')!;summary.scrollIntoView({block:'center'});summary.click()}
 else if(phase===4){if(root.dataset.podPhaseSource!=='3')throw Error('Disclosure read source');const button=panel().querySelector<HTMLButtonElement>('[data-pod-metric="'+root.dataset.podPhaseUsageKey+'"] > button')!;button.scrollIntoView({block:'center'});button.focus();(window as any).__podPhaseActorOpener=button;button.click()}
 else if(phase===5){if(root.dataset.podPhaseSource!=='3')throw Error('Native staging read source');document.querySelector<HTMLButtonElement>('#measureRun')!.click()}
 else if(phase===6)[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Back to Kubernetes')!.click();
 else if(phase===7){if(root.dataset.podPhaseSource!=='4'||panel().querySelectorAll('.pod-metric-widget').length!==1||document.querySelector<HTMLInputElement>('[aria-label="Find reported pod resource"]')!.value!=='reported-unknown'||document.activeElement!==(window as any).__podPhaseActorOpener)throw Error('Retained native panel/focus/read budget differs');root.dataset.podPhaseTerminal='true'}
 root.dataset.podPhasePhase=String(phase+1);
});
export const workflow=actions(()=>{const s=state.current;if(s.phase>=8||s.phase===1&&(!s.ready||!s.target)||s.phase===3&&(!s.loaded||!s.usageKey)||s.phase===5&&!s.nativeReady||s.phase===6&&!s.nativeComplete)return ['Wait'];return [step()];});
