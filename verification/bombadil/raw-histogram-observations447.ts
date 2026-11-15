// Real raw-histogram acquisition followed by local held-data interactions.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state=extract(({document,window})=>{
 const o=(window as any).__raw447||{},calls=o.calls||[],fixture=JSON.parse(new URL(location.href).searchParams.get('raw447')||'null');
 const visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[aria-hidden="true"],[inert]');
 const panel=document.querySelector<HTMLElement>('[data-raw-observations]');
 const receipt=(label:string)=>[...(panel?.querySelectorAll('dt')||[])].find(e=>e.textContent===label)?.nextElementSibling?.textContent||'';
 const inspector=panel?.querySelector<HTMLElement>('[aria-label="Selected item details"]');
 const held={readAt:receipt('Source read UTC'),clock:receipt('Fixed window'),keys:[...panel?.querySelectorAll<HTMLElement>('tbody tr[data-occurrence-key]')||[]].map(e=>e.dataset.occurrenceKey),find:panel?.querySelector<HTMLInputElement>('[aria-label="Find held observations"]')?.value||'',selected:inspector?.dataset.key||'',detail:inspector?.textContent||''};
 const label=fixture?.label||'Read observations for request.duration [ms]';
 const source=calls.filter((c:any)=>c.path==='/api/histogram-observations').length,posts=calls.filter((c:any)=>c.method==='POST');
 const read=[...document.querySelectorAll<HTMLButtonElement>('#metricInstruments button')].find(b=>b.getAttribute('aria-label')===label&&visible(b)&&!b.disabled);
 return {version:'goo.rawHistogram447Actor/1',phase:Number(document.documentElement.dataset.raw447Phase||0),ready:!!document.querySelector('#metricDiscover'),hasRead:!!read,panelVisible:visible(panel),rows:panel?.querySelectorAll('tbody tr[data-occurrence-key]:not([hidden])').length||0,detail:visible(inspector),bucketRows:inspector?.querySelectorAll('[aria-label="Source histogram buckets"] tbody tr').length||0,source,posts:posts.length,forbidden:posts.filter((c:any)=>!['/api/metric-discovery','/api/histogram-observations','/api/cel/context','/api/cel/plan'].includes(c.path)).length,readFreeLocal:!o.beforeLocal||posts.length===o.beforeLocal,heldRestored:!o.held||JSON.stringify(o.held)===JSON.stringify(held),held,overflow:document.documentElement.scrollWidth>innerWidth};
});
export const explicitAcquisition=always(()=>state.current.source<=1);
export const noEvaluationOrWrites=always(()=>state.current.forbidden===0);
export const localAuthority=always(()=>state.current.phase<4||state.current.readFreeLocal);
export const heldReturn=always(()=>state.current.phase<7||state.current.heldRestored);
export const noOverflow=always(()=>!state.current.overflow);
export const completes=eventually(()=>state.current.phase===8&&state.current.source===1&&state.current.detail&&state.current.bucketRows>0).within(24,'seconds');

const step=registerCustomAction('goo.rawHistogram447Step',async(document,window,next:number)=>{
 const target:any=window,fixture=JSON.parse(new URL(location.href).searchParams.get('raw447')||'null');
 const visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[aria-hidden="true"],[inert]');
 const click=(e:any)=>{if(!e||!visible(e)||e.disabled)throw Error('Raw histogram control unavailable');e.click()};
 const panel=()=>document.querySelector<HTMLElement>('[data-raw-observations]')!;
 if(next===1){
  if(!fixture)throw Error('Missing histogram fixture');const o:any=target.__raw447={calls:[]},original=window.fetch;
  window.fetch=function(input,options){const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,location.href),method=String(options?.method||(input instanceof Request?input.method:'GET')).toUpperCase();let body;try{body=JSON.parse(String(options?.body))}catch{}o.calls.push({path:url.pathname,method,body});return Reflect.apply(original,this,[input,options])};
  click(document.querySelector('#exploreMeasurementsMode'));const disclosure=document.querySelector<HTMLDetailsElement>('#metricDiscovery');if(disclosure)disclosure.open=true;
  for(const [id,value]of [['metricEnvironment',fixture.environment],['metricService',fixture.service]]){const input=document.getElementById(id) as HTMLInputElement;if(!input)throw Error('Discovery input unavailable');input.value=value;input.dispatchEvent(new Event('input',{bubbles:true}))}
 }else if(next===2)click(document.querySelector('#metricDiscover'));
 else if(next===3)click([...document.querySelectorAll<HTMLButtonElement>('#metricInstruments button')].find(b=>b.getAttribute('aria-label')===(fixture.label||'Read observations for request.duration [ms]')&&visible(b)));
 else if(next===4){target.__raw447.beforeLocal=target.__raw447.calls.filter((c:any)=>c.method==='POST').length;const find=panel().querySelector<HTMLInputElement>('[aria-label="Find held observations"]')!;find.value=fixture.find||'standard';find.dispatchEvent(new Event('input',{bubbles:true}))}
 else if(next===5)click([...panel().querySelectorAll<HTMLButtonElement>('button[aria-label^="Inspect observation"]')].find(visible));
 else if(next===6){const root=panel(),receipt=(label:string)=>[...root.querySelectorAll('dt')].find(e=>e.textContent===label)?.nextElementSibling?.textContent||'',inspector=root.querySelector<HTMLElement>('[aria-label="Selected item details"]');target.__raw447.held={readAt:receipt('Source read UTC'),clock:receipt('Fixed window'),keys:[...root.querySelectorAll<HTMLElement>('tbody tr[data-occurrence-key]')].map(e=>e.dataset.occurrenceKey),find:root.querySelector<HTMLInputElement>('[aria-label="Find held observations"]')?.value||'',selected:inspector?.dataset.key||'',detail:inspector?.textContent||''};click(document.querySelector('#exploreRecordsMode'))}
 else if(next===7)click(document.querySelector('#exploreMeasurementsMode'));
 document.documentElement.dataset.raw447Phase=String(next);
});
export const choices=actions(()=>{const s=state.current;if(s.phase===0)return s.ready?[step(1)]:['Wait'];if(s.phase===1)return[step(2)];if(s.phase===2)return s.hasRead?[step(3)]:['Wait'];if(s.phase===3)return s.rows>0?[step(4)]:['Wait'];if(s.phase===4)return s.rows>0?[step(5)]:['Wait'];if(s.phase===5)return s.detail&&s.bucketRows>0?[step(6)]:['Wait'];if(s.phase===6)return !s.panelVisible?[step(7)]:['Wait'];if(s.phase===7)return s.panelVisible&&s.detail?[step(8)]:['Wait'];return['Wait']});
