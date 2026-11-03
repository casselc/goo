// Starts at a genuine UI-generated Share question link. One explicit gauge read;
// all later actions inspect/filter retained streams without acquiring telemetry.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const held=(window as any).__gooTypedNative319,run=document.querySelector<HTMLButtonElement>('#measureRun'),result=document.querySelector<HTMLElement>('#measureResult'),streams=result?.querySelectorAll('.native-stream-list>details').length||0;
 const resultStable=!held?.fingerprint||JSON.stringify(held.result)===held.fingerprint;
 const scope=held?.result?.question,expected=held?.expected,policy=!scope||JSON.stringify(scope.definition.environmentSelector)===JSON.stringify(expected.measurement.environmentSelector)&&scope.fromMs===expected.scope.fromMs&&scope.toMs===expected.scope.toMs;
 return {version:'goo.typedNative319',phase:held?.phase||0,ready:!!run&&!run.disabled,streams,sourceAttempts:held?.sourceAttempts||0,measurementAttempts:held?.measurementAttempts||0,captureWrites:held?.captureWrites||0,policy,resultStable,valuesVerified:!!held?.valuesVerified,shared:!!held?.shared,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const oneExplicitGaugeRead=always(()=>state.current.sourceAttempts===state.current.measurementAttempts&&state.current.measurementAttempts<=1&&state.current.captureWrites===0);
export const immutableTypedResult=always(()=>state.current.policy&&state.current.resultStable);
export const noHorizontalOverflow=always(()=>!state.current.overflow);
export const completes=eventually(()=>state.current.phase===8&&state.current.valuesVerified&&state.current.shared&&state.current.measurementAttempts===1).within(25,'seconds');
const step=registerCustomAction('goo.typedNative',async(document,window,phase:number)=>{
 let held=(window as any).__gooTypedNative319;const result=document.querySelector<HTMLElement>('#measureResult')!,streams=()=>[...result.querySelectorAll<HTMLDetailsElement>('.native-stream-list>details')],change=(input:HTMLInputElement|HTMLSelectElement,value:string,type='input')=>{input.value=value;input.dispatchEvent(new Event(type,{bubbles:true}))};
 if(phase===1){
  const expected=JSON.parse(decodeURIComponent(window.location.hash.slice(6)));if(expected.kind!=='measurement'||expected.measurement.kind!=='gauge'||!expected.measurement.environmentSelector||expected.measurement.source.exactStream)throw Error('Expected shared typed multi-stream gauge question');
  held=(window as any).__gooTypedNative319={phase:0,expected,sourceAttempts:0,measurementAttempts:0,captureWrites:0};
  const source=new Set(['/api/explore','/api/explore/page','/api/explore/counted-values','/api/record-aggregates','/api/tail','/api/discovery','/api/inventory','/api/metric-discovery','/api/measurements']),fetch=window.fetch.bind(window);
  window.fetch=async(input,init)=>{const path=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;if(source.has(path))held.sourceAttempts++;if(path==='/api/captures'&&(init?.method||'GET')!=='GET')held.captureWrites++;if(path==='/api/measurements')held.measurementAttempts++;const response=await fetch(input,init);if(path==='/api/measurements'){held.result=await response.clone().json();held.fingerprint=JSON.stringify(held.result)}return response};
  document.querySelector<HTMLButtonElement>('#measureRun')!.click();
 }else if(phase===2){
  const series=held.result?.gauges?.series;if(series?.length!==2||JSON.stringify(series.map((s:any)=>String(s.samples.at(-1).value)).sort())!==JSON.stringify(['123','456']))throw Error('Expected two separate actual gauge values 123 and 456');
  if(JSON.stringify(series[0].stream.resource)===JSON.stringify(series[1].stream.resource))throw Error('Expected distinct raw resource identities');
  for(const value of ['123','456'])if(![...result.querySelectorAll('td[data-label="Recorded value"]')].some(cell=>cell.textContent===value))throw Error('Exact retained sample value missing from UI');held.valuesVerified=true;for(const details of streams())details.open=true;
 }else if(phase===3){
  const input=result.querySelector<HTMLInputElement>('[aria-label="Find a stream"]')!;change(input,'deployment.environment.name');if(streams().length!==1)throw Error('Raw current-key Find must isolate one retained stream');change(input,'');if(streams().length!==2)throw Error('Clear Find must restore both independent streams');
 }else if(phase===4){
  const select=result.querySelector<HTMLSelectElement>('[aria-label="Stream state"]')!;change(select,'review','change');result.querySelector<HTMLButtonElement>('.native-stream-tools button')!.click();if(streams().length!==2||select.value!=='all')throw Error('Clear local stream filters must restore both streams');
 }else if(phase===5){
  const details=streams()[1],summary=details.querySelector<HTMLElement>(':scope>summary')!;summary.focus();summary.click();summary.click();if(document.activeElement!==summary)throw Error('Stream disclosure lost focus');
 }else if(phase===6){
  document.querySelector<HTMLButtonElement>('#measureShare')!.click();const link=document.querySelector<HTMLInputElement>('#measureQuestionLink')!;const shared=JSON.parse(decodeURIComponent(new URL(link.value).hash.slice(6)));if(JSON.stringify(shared)!==JSON.stringify(held.expected))throw Error('Share changed the typed definition or fixed scope');held.shared=true;
 }else if(phase===7){
  const limits=[...result.querySelectorAll<HTMLDetailsElement>('details')].find(d=>d.querySelector(':scope>summary')?.textContent==='Evidence scope and limits');if(!limits)throw Error('Evidence limits disclosure missing');limits.open=true;limits.open=false;
 }else if(phase===8){
  if(held.sourceAttempts!==1||held.measurementAttempts!==1||held.captureWrites)throw Error('Local inspection acquired or saved additional evidence');if(streams().length!==2)throw Error('Local inspection lost an independent stream');
 }
 held.phase=phase;
});
export const explore=actions(()=>{const s=state.current;if(s.phase===0)return s.ready?[step(1)]:['Wait'];if(s.phase===1)return s.streams===2&&s.ready?[step(2)]:['Wait'];if(s.phase<8)return[step(s.phase+1)];return ['Wait']});
