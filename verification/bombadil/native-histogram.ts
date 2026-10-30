import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>({version:'goo.native272',phase:document.documentElement.dataset.native272||'initial',ready:!!document.querySelector('#captureMeasurementEvidence goo-value-timeline'),posts:Number(document.documentElement.dataset.native272Posts||0),bad:document.documentElement.dataset.native272Bad==='true',reason:document.documentElement.dataset.native272Reason||''}));
export const immutableInspection=always(()=>!state.current.bad&&state.current.posts===0);
export const completes=eventually(()=>state.current.phase==='complete').within(25,'seconds');
const step=registerCustomAction('goo.native272',async(document,window,phase:string)=>{
 const attrs=document.documentElement.dataset,store=window as any;
 const check=(ok:boolean,reason:string)=>{if(!ok){attrs.native272Bad='true';attrs.native272Reason=reason;throw Error(reason)}};
 const selection=()=>document.querySelector('#captureMeasurementEvidence .native-timeline-selection')!;
 const chart=()=>selection()?.querySelector('goo-chart') as any;
 const choose=(key:string)=>{const button=document.querySelector<HTMLButtonElement>('#captureMeasurementEvidence goo-value-timeline [data-key="'+key+'"]');check(!!button,'Retained interval required');button!.click()};
 const reveal=async()=>{const summary=[...selection().querySelectorAll<HTMLElement>('summary')].find(x=>x.textContent==='Selected native bucket counts');check(!!summary,'Native bucket disclosure required');if(!(summary!.parentElement as HTMLDetailsElement).open)summary!.click();for(let i=0;i<100&&!chart();i++)await new Promise(resolve=>window.setTimeout(resolve,10));check(!!chart(),'Lazy native chart must finish mounting')};
 if(phase==='known'){
  store.__native272={};const fetch=window.fetch.bind(window);window.fetch=(input,init)=>{const method=init?.method||(typeof input==='object'&&'method'in input?input.method:'GET');if(method.toUpperCase()==='POST')attrs.native272Posts=String(Number(attrs.native272Posts||0)+1);return fetch(input,init)};
  choose('0');await reveal();const s=chart()?.snapshot;check(s?.histogram.kind==='otel-explicit'&&s.histogram.total==='4'&&s.histogram.grain==='observations'&&s.histogram.closure==='left-open-right-closed','Native observation semantics required');check(s.rows[0].values.from==='-inf'&&s.rows.at(-1).values.to==='+inf','Infinite tails required');check(s.rows[0].values.to==='3602879701896397/36028797018963968','Finite source binary64 bound must stay exact');check(Object.isFrozen(s),'Immutable histogram required');store.__native272={snapshot:s,json:JSON.stringify(s)};
 }else if(phase==='table'){
  const select=selection().querySelector<HTMLSelectElement>('select[aria-label="Native histogram display"]');check(!!select,'Local histogram/table selector required');select!.value='table';select!.dispatchEvent(new Event('change',{bubbles:true}));check(chart()?.querySelectorAll('tbody tr').length===3,'Same three observation buckets required');check(chart()?.snapshot===store.__native272.snapshot,'Local kind switch must preserve exact handle');
 }else if(phase==='zero'){
  choose('2');await reveal();const s=chart()?.snapshot;check(s?.histogram.total==='0'&&s.rows.length===3&&s.rows.every((r:any)=>r.values.count==='0'&&r.state==='no_data'),'Known empty observations must retain all exact zero bins');
 }else if(phase==='unknown'){
  choose('1');check(!chart(),'No-recorded-value interval must not become a histogram');check(selection().textContent?.includes('no-recorded-value'),'Unknown reason must remain inspectable');
 }else{
  check(phase==='complete','Unknown campaign action');check(JSON.stringify(store.__native272.snapshot)===store.__native272.json,'Selecting other intervals must preserve retained first snapshot');check(Number(attrs.native272Posts||0)===0,'Saved inspection must not POST');
 }
 attrs.native272=phase;
});
export const follow=actions(()=>{const s=state.current;if(s.phase==='initial'&&s.ready)return[step('known')];for(const [a,b]of[['known','table'],['table','zero'],['zero','unknown'],['unknown','complete']])if(s.phase===a)return[step(b)];return['Wait']});
