import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const phase=document.documentElement.dataset.histogram407Phase||'initial';
 const pop=document.querySelector<HTMLElement>('.population-workspace:not([hidden])');
 const chart=pop?.querySelector<any>('[data-composition-role=histogram] goo-chart');
 const expected=JSON.parse(new URL(window.location.href).searchParams.get('expected')||'null');
 const buttons=[...(pop?.querySelectorAll<HTMLButtonElement>('button')||[])];
 const apply=buttons.find(e=>e.textContent==='Apply template');
 const calls=(window as any).__histogram407Calls||[];
 return{version:'goo.histogram407Actor/1',phase,population:pop?.dataset.phase,inspect:buttons.some(e=>e.textContent==='Inspect template · '+expected?.title),applyReady:!!apply&&!apply.disabled,chartCount:pop?.querySelectorAll('[data-composition-role=histogram] goo-chart').length||0,overflow:document.documentElement.scrollWidth>window.innerWidth,calls:calls.map((c:any)=>phase==='read'?c:{path:c.path,method:c.method,status:c.status}),histogram:chart?{display:chart.display,...(phase==='read'?{snapshot:chart.snapshot}:{})}:null};
});
export const finishes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const budget=always(()=>state.current.calls.filter((c:any)=>c.path==='/api/explore').length<=1&&state.current.calls.filter((c:any)=>c.path==='/api/view-templates/expand').length<=1&&state.current.calls.every((c:any)=>c.method!=='POST'||['/api/explore','/api/view-templates/expand'].includes(c.path)));
export const staging=always(()=>!['library','preview','applied'].includes(state.current.phase)||state.current.chartCount===0);
export const viewport=always(()=>!state.current.overflow);
const step=registerCustomAction('goo.histogram407',async(document,window,phase:string)=>{
 const pop=()=>document.querySelector<HTMLElement>('.population-workspace:not([hidden])')!;
 const button=(text:string)=>[...pop().querySelectorAll<HTMLButtonElement>('button')].find(e=>e.textContent===text)!;
 const visible=(e:any)=>e?.getClientRects().length>0&&!e.closest('[hidden],[inert]');
 const open=(e:HTMLElement)=>{const parents:HTMLDetailsElement[]=[];for(let p=e.parentElement;p;p=p.parentElement)if(p.localName==='details')parents.unshift(p as HTMLDetailsElement);for(const p of parents)if(!p.open)p.querySelector<HTMLElement>(':scope>summary')!.click()};
 const expected=JSON.parse(new URL(window.location.href).searchParams.get('expected')||'null');
 const stable=(x:any):any=>Array.isArray(x)?x.map(stable):x&&typeof x==='object'?Object.fromEntries(Object.keys(x).sort().map(k=>[k,stable(x[k])])):x;
 const same=(a:any,b:any)=>JSON.stringify(stable(a))===JSON.stringify(stable(b));
 if(phase==='initial'){
  const original=window.fetch.bind(window);(window as any).__histogram407Calls=[];
  window.fetch=async function(input:any,init:any){const path=new URL(typeof input==='string'?input:input.url,window.location.href).pathname,record:any={path,method:init?.method||'GET'};if(path.startsWith('/api/')){if(init?.body)try{record.request=JSON.parse(init.body)}catch{}(window as any).__histogram407Calls.push(record)}const response=await original(input,init);record.status=response.status;try{record.response=await response.clone().json()}catch{}return response};
  const load=button('Load saved views');open(load);if(!visible(load)||load.disabled)throw Error('Saved library unavailable');load.click();
 }else if(phase==='library'){
  const inspect=button('Inspect template · '+expected.title);if(!visible(inspect)||inspect.disabled)throw Error('Template inspect unavailable');inspect.click();
 }else if(phase==='preview'){
  const apply=button('Apply template');if(!visible(apply)||apply.disabled)throw Error('Template Apply unavailable');apply.click();
 }else if(phase==='applied'){
  if(pop().querySelector('[data-composition-role=histogram] goo-chart'))throw Error('Apply manufactured source bins');
  const run=pop().querySelector<HTMLButtonElement>('button.population-run[aria-label="Run population view"]');if(!run)throw Error('Explicit Run unavailable');open(run);if(!visible(run)||run.disabled)throw Error('Run unavailable');run.click();
 }else if(phase==='read'){
  const chart=pop().querySelector<any>('[data-composition-role=histogram] goo-chart')!,calls=(window as any).__histogram407Calls,reads=calls.filter((c:any)=>c.path==='/api/explore'),expands=calls.filter((c:any)=>c.path==='/api/view-templates/expand');
  if(reads.length!==1||reads[0].status!==200||!same(reads[0].request,expected.question))throw Error('Template replaced destination query');
  if(expands.length!==1||expands[0].status!==200||!same(expands[0].response.view.queries,[expected.question]))throw Error('Expansion changed destination scope');
  if(chart.display.kind!=='table'||chart.display.requiresComplete!==true||!same(chart.snapshot,reads[0].response.histogramDataset))throw Error('Authored table changed source bins or completeness policy');
  const bins=chart.snapshot.rows.map((r:any)=>({from:r.values.from,to:r.values.to,count:r.values.count}));
  if(!same(bins,expected.bins)||reads[0].response.count!==expected.count||!same(chart.snapshot.scope.selector,expected.question)||chart.snapshot.revision!==reads[0].response.basis)throw Error('Source bin membership, scope or basis changed');
 }
 document.documentElement.dataset.histogram407Phase=({initial:'library',library:'preview',preview:'applied',applied:'read',read:'complete'} as any)[phase];
});
export const explore=actions(()=>{const s=state.current;return s.phase==='initial'&&s.population==='staged'?[step('initial')]:s.phase==='library'&&s.inspect?[step('library')]:s.phase==='preview'&&s.applyReady?[step('preview')]:s.phase==='applied'&&s.population==='staged'?[step('applied')]:s.phase==='read'&&s.population==='applied'&&s.histogram?[step('read')]:['Wait'];});
