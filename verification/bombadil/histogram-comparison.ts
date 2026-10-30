import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>({version:'goo.histogram270',phase:document.documentElement.dataset.histogram270||'initial',
 ready:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&!!document.querySelector('#exploreRun:not([disabled])')?.getClientRects().length,
 applied:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied',
 reads:Number(document.documentElement.dataset.histogram270Reads||0),plans:Number(document.documentElement.dataset.histogram270Plans||0),bad:document.documentElement.dataset.histogram270Bad==='true',reason:document.documentElement.dataset.histogram270Reason||'',
 pairValid:(document.querySelector('#cohortMeasures goo-aligned-histogram') as any)?.comparison?.selection?.histogram?.support?.valid||'',
 child:document.querySelector('.population-child .population-workspace')?.getAttribute('data-phase')||''}));
export const explicitReadsOnly=always(()=>!state.current.bad&&state.current.reads<=3&&state.current.plans<=1);
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
const act=registerCustomAction('goo.histogram270',async(document,window,phase:string)=>{
 const attrs=document.documentElement.dataset,store=window as any;
 const check=(ok:boolean,reason:string)=>{if(!ok){attrs.histogram270Bad='true';attrs.histogram270Reason=reason;throw Error(reason)}};
 const reveal=(el:Element)=>{const ancestors=[];for(let p=el.parentElement;p;p=p.parentElement)if(p.localName==='details')ancestors.unshift(p);for(const p of ancestors)if(!(p as HTMLDetailsElement).open)p.querySelector<HTMLElement>(':scope > summary')?.click()};
 const click=(id:string)=>{const button=document.querySelector<HTMLButtonElement>('#'+id);check(!!button&&!button.disabled,'Enabled '+id+' required');reveal(button!);button!.click()};
 const host=()=>document.querySelector('#cohortMeasures .numeric-distribution-comparison')!;
 const plot=()=>host()?.querySelector('goo-aligned-histogram') as any;
 if(phase==='source'){
  attrs.histogram270='source';store.__gooHistogram270={};const native=window.fetch.bind(window);
  window.fetch=(input,init)=>{const method=init?.method||(typeof input==='object'&&'method' in input?input.method:'GET');if(method.toUpperCase()==='POST'){
   const path=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;
   if(path==='/api/explore'){attrs.histogram270Reads=String(Number(attrs.histogram270Reads||0)+1);if(!['source','selection','stage'].includes(attrs.histogram270||''))attrs.histogram270Bad='true'}
   else if(path==='/api/numeric-histograms/comparison-selection'){attrs.histogram270Plans=String(Number(attrs.histogram270Plans||0)+1);if(attrs.histogram270!=='stage')attrs.histogram270Bad='true'}
   else if(path!=='/api/compare'&&!path.startsWith('/api/queries/'))attrs.histogram270Bad='true';
  }return native(input,init)};click('exploreRun');return;
 }
 const saved=store.__gooHistogram270;check(!!saved,'Real source baseline required');
 if(phase==='pin'){attrs.histogram270='pin';click('distributionPin');return}
 if(phase==='selection'){
  const value=document.querySelector<HTMLInputElement>('#exploreFilterEditor input[aria-label="Condition 1 value"]');check(!!value&&value.value==='baseline','Actual baseline population filter required');reveal(value!);value!.value='selection';value!.dispatchEvent(new Event('input',{bubbles:true}));attrs.histogram270='selection';click('exploreRun');return;
 }
 if(phase==='compare'){attrs.histogram270='compare';click('distributionCompare');return}
 const modulePath='/aligned-histogram.mjs',aligned=await import(modulePath),model=()=>aligned.model(plot().comparison,plot().display);
 const choose=(scale:string)=>{const select=host()?.querySelector<HTMLSelectElement>('select[aria-label="Distribution scale"]');check(!!select,'Explicit denominator control required');select!.value=scale;select!.dispatchEvent(new Event('change',{bubbles:true}))};
 const fact=(shares:string[])=>{const m=model(),tail=m?.all.at(-1);check(!!m&&m.sides[0].histogram.support.total==='4'&&m.sides[0].histogram.support.valid==='2'&&m.sides[1].histogram.support.total==='4'&&m.sides[1].histogram.support.valid==='3','Independent total and valid coverage required');check(tail.counts[0]==='1'&&tail.counts[1]==='2'&&tail.to==='9223372036854775808'&&JSON.stringify(tail.shares)===JSON.stringify(shares),'Exact aligned tail count/denominator/sentinel must remain unchanged');check(plot().comparison===saved.pair&&JSON.stringify(saved.pair)===saved.json,'Local scale must not replace immutable operands')};
 if(phase==='population'){
  click('resultCompare');const p=plot()?.comparison;check(!!p&&p.baseline.rows.length===128&&p.selection.rows.length===128&&Object.isFrozen(p),'Actual independently captured aligned pair required');saved.pair=p;saved.json=JSON.stringify(p);choose('population-share');fact(['25','50']);check(Number(attrs.histogram270Reads)===2&&Number(attrs.histogram270Plans||0)===0,'Comparison scale must not read telemetry');attrs.histogram270='population';return;
 }
 if(phase==='valid'){choose('valid-share');fact(['50','200/3']);attrs.histogram270='valid';return}
 if(phase==='table'){
  const details=[...plot().querySelectorAll<HTMLDetailsElement>('details')].find(d=>d.querySelector('summary')?.textContent==='All exact aligned bins');check(!!details,'Full exact aligned bins required');details!.querySelector<HTMLElement>('summary')!.click();check(plot().querySelectorAll('tr[data-bin]').length===128,'Visual trim must retain all128 exactbins');check(plot().querySelector('tr[data-bin="bin:127"] th')?.getAttribute('title')?.includes('9223372036854775808'),'Exact sentinel must remain inspectable');details!.querySelector<HTMLElement>('summary')!.click();fact(['50','200/3']);attrs.histogram270='table';return;
 }
 if(phase==='brush'){
  const path='/view-model.mjs',views=await import(path);plot().dispatchEvent(new CustomEvent('goo-aligned-range',{bubbles:true,detail:{side:'baseline',intent:views.histogramIntent(saved.pair.baseline,'bin:127')}}));const button=[...host().querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Open baseline range');check(!!button&&!button.disabled,'Exact nonempty baseline range required');saved.open=button;check(Number(attrs.histogram270Reads)===2&&Number(attrs.histogram270Plans||0)===0,'Range brushing must neither plan nor read');attrs.histogram270='brush';return;
 }
 if(phase==='stage'){attrs.histogram270='stage';saved.open.click();return}
 check(phase==='complete','Unknown aligned comparison phase');check(Number(attrs.histogram270Reads)===3&&Number(attrs.histogram270Plans)===1,'Only explicit Open may plan and read the selected side');check(Object.isFrozen(saved.pair)&&JSON.stringify(saved.pair)===saved.json,'Opening a side must preserve the captured immutable comparison');check(document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied','Explicitly opened baseline population must apply');attrs.histogram270='complete';
});
export const follow=actions(()=>{const s=state.current;
 if(s.phase==='initial'&&s.ready)return[act('source')];
 if(s.phase==='source'&&s.applied&&s.reads===1)return[act('pin')];
 if(s.phase==='pin'&&s.pairValid==='2')return[act('selection')];
 if(s.phase==='selection'&&s.applied&&s.reads===2)return[act('compare')];
 if(s.phase==='compare'&&s.pairValid==='3')return[act('population')];
 for(const [a,b]of [['population','valid'],['valid','table'],['table','brush'],['brush','stage']])if(s.phase===a)return[act(b)];
 if(s.phase==='stage'&&s.applied&&s.reads===3&&s.plans===1)return[act('complete')];
 return['Wait'];});
