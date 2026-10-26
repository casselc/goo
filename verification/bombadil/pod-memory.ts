import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>({phase:Number(document.documentElement.dataset.podMemoryPhase||0),source:Number(document.documentElement.dataset.podMemorySource||0),plans:Number(document.documentElement.dataset.podMemoryPlans||0),
 target:document.documentElement.dataset.podMemoryTarget||'',usageKey:document.documentElement.dataset.podMemoryUsageKey||'',ready:!!document.querySelector('tr[data-pod-variant]'),
 loaded:[...document.querySelectorAll('.pod-metrics-panel [role=status]')].some(e=>e.textContent?.includes('Independent pod memory observations')),
 nativeReady:!!document.querySelector('#measureRun')&&!document.querySelector<HTMLButtonElement>('#measureRun')!.disabled,
 nativeComplete:!!document.querySelector('#measureResult')?.children.length&&!document.querySelector<HTMLButtonElement>('#measureCapture')?.disabled,
 terminal:document.documentElement.dataset.podMemoryTerminal==='true'}));
export const bounded=always(()=>state.current.source<=5&&state.current.plans<=2);
export const complete=eventually(()=>state.current.terminal).within(25,'seconds');
const step=registerCustomAction('goo.podMemoryStep',async(document,window)=>{
 const root=document.documentElement,phase=Number(root.dataset.podMemoryPhase||0),panel=()=>document.querySelector<HTMLElement>('.pod-metrics-panel[data-pod-metrics="'+root.dataset.podMemoryTarget+'"]')!;
 if(phase===0){
  const environment=new URLSearchParams(window.location.hash.slice(1)).get('podMemoryEnvironment')!,original=window.fetch.bind(window);
  window.fetch=async(input,init)=>{const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href),method=init?.method||(input instanceof Request?input.method:'GET');
   if(method==='POST'&&url.pathname.startsWith('/api/')&&!url.pathname.endsWith('/cancel')){const key=url.pathname.endsWith('/plan')?'podMemoryPlans':'podMemorySource';root.dataset[key]=String(Number(root.dataset[key]||0)+1)}
   const response=await original(input,init);
   if(response.ok&&method==='POST'&&url.pathname==='/api/pod-discovery'){const data=await response.clone().json(),row=data.observations.find((r:any)=>r.resource['k8s.pod.uid']==='uid-a'&&r.resourceSchemaURL===''&&!Object.hasOwn(r.resource,'service.name')&&!Object.hasOwn(r.resource,'k8s.cluster.uid')&&!Object.hasOwn(r.resource,'k8s.namespace.name'));if(!row)throw Error('Known-empty-schema pod resource absent');root.dataset.podMemoryTarget=row.key}
   if(response.ok&&method==='POST'&&url.pathname==='/api/pod-dashboard/plan'){const data=await response.clone().json();if(data.widgets.length!==2||data.widgets.some((w:any)=>w.profile!=='collector-kubeletstats-v0.137.0'))throw Error('Receiver default pair differs');root.dataset.podMemoryUsageKey=data.widgets.find((w:any)=>w.question.definition.source.metric==='k8s.pod.memory.usage').key}
   if(response.ok&&method==='POST'&&url.pathname==='/api/measurements'){const data=await response.clone().json(),b=data.question.definition.source,expected=b.metric==='k8s.pod.memory.usage'?32768:16384;if(b.resource.kind!=='kubernetes-pod'||b.resource.id!=='uid-a'||b.version!=='0.137.0'||data.gauges.series.length!==1||data.gauges.series[0].samples.at(-1).value!==expected)throw Error('Actual exact pod memory differs')}
   return response;
  };
  document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();document.querySelector<HTMLButtonElement>('#entitiesKubernetes')!.click();const input=document.querySelector<HTMLInputElement>('[aria-label="Kubernetes environment"]')!;input.value=environment;input.dispatchEvent(new Event('input',{bubbles:true}));
  const minutes=document.querySelector<HTMLInputElement>('#minutes')!;minutes.value='2';minutes.dispatchEvent(new Event('input',{bubbles:true}));[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Find pods')!.click();
 }else if(phase===1){const input=document.querySelector<HTMLInputElement>('[aria-label="Find reported pod resource"]')!;input.value='uid-a';input.dispatchEvent(new Event('input',{bubbles:true}));if(root.dataset.podMemorySource!=='1')throw Error('Local search read source')}
 else if(phase===2){const button=document.querySelector<HTMLButtonElement>('button[data-pod-observe="'+root.dataset.podMemoryTarget+'"]')!;button.scrollIntoView({block:'center'});button.click()}
 else if(phase===3){if(root.dataset.podMemorySource!=='4')throw Error('Default budget differs');const summary=panel().querySelector<HTMLElement>('[data-pod-metric="'+root.dataset.podMemoryUsageKey+'"] details > summary')!;summary.scrollIntoView({block:'center'});summary.click()}
 else if(phase===4){if(root.dataset.podMemorySource!=='4')throw Error('Disclosure read source');const button=panel().querySelector<HTMLButtonElement>('[data-pod-metric="'+root.dataset.podMemoryUsageKey+'"] > button')!;button.scrollIntoView({block:'center'});button.focus();(window as any).__podMemoryActorOpener=button;button.click()}
 else if(phase===5){if(root.dataset.podMemorySource!=='4')throw Error('Native staging read source');document.querySelector<HTMLButtonElement>('#measureRun')!.click()}
 else if(phase===6)[...document.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Back to Kubernetes')!.click();
 else if(phase===7){if(root.dataset.podMemorySource!=='5'||panel().querySelectorAll('.pod-metric-widget').length!==2||document.querySelector<HTMLInputElement>('[aria-label="Find reported pod resource"]')!.value!=='uid-a'||document.activeElement!==(window as any).__podMemoryActorOpener)throw Error('Retained native panel/focus/read budget differs');root.dataset.podMemoryTerminal='true'}
 root.dataset.podMemoryPhase=String(phase+1);
});
export const workflow=actions(()=>{const s=state.current;if(s.phase>=8||s.phase===1&&(!s.ready||!s.target)||s.phase===3&&(!s.loaded||!s.usageKey)||s.phase===5&&!s.nativeReady||s.phase===6&&!s.nativeComplete)return ['Wait'];return [step()];});
