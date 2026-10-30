import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>({version:'goo.duration271',phase:document.documentElement.dataset.duration271||'initial',
 ready:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&!!document.querySelector('#exploreRun:not([disabled])')?.getClientRects().length,
 applied:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied',reads:Number(document.documentElement.dataset.duration271Reads||0),pairs:Number(document.documentElement.dataset.duration271Pairs||0),bad:document.documentElement.dataset.duration271Bad==='true',reason:document.documentElement.dataset.duration271Reason||'',
 bins:document.querySelectorAll('#distributionContent .latency-histogram button[data-key]').length,
 population:document.querySelector('.population-workspace')?.getAttribute('data-phase')||'',
 result:document.querySelector('.population-workspace [id$=cohortStatus]')?.textContent||''}));
export const explicitReadsOnly=always(()=>!state.current.bad&&state.current.reads<=3&&state.current.pairs<=1);
export const localSelectionDoesNotRead=always(()=>!['brush','manual','reset','rest'].includes(state.current.phase)||state.current.reads===1&&state.current.pairs===0);
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
const act=registerCustomAction('goo.duration271',async(document,window,phase:string)=>{
 const attrs=document.documentElement.dataset,store=window as any;
 const check=(ok:boolean,reason:string)=>{if(!ok){attrs.duration271Bad='true';attrs.duration271Reason=reason;throw Error(reason)}};
 const section=()=>document.querySelector('#distributionContent .latency-histogram')!;
 const chart=()=>section()?.querySelector('goo-chart') as any;
 const tray=()=>section()?.querySelector('.population-selection')!;
 const choose=()=>{const bin=section()?.querySelector<HTMLButtonElement>('button[data-key="bin:2"]');check(!!bin,'Exact duration bin2 required');bin!.focus();bin!.dispatchEvent(new KeyboardEvent('keydown',{key:' ',bubbles:true,cancelable:true}));bin!.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}));check(tray().querySelector<HTMLInputElement>('input[aria-label="Minimum span duration (ms)"]')?.value==='5'&&tray().querySelector<HTMLInputElement>('input[aria-label="Maximum span duration (ms, exclusive)"]')?.value==='10','Keyboard bin2 must select exact half-open5to10ms')};
 if(phase==='source'){
  attrs.duration271='source';store.__gooDuration271={};const native=window.fetch.bind(window);
  window.fetch=(input,init)=>{const method=init?.method||(typeof input==='object'&&'method' in input?input.method:'GET');if(method.toUpperCase()==='POST'){
   const path=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;
   if(path==='/api/explore'){attrs.duration271Reads=String(Number(attrs.duration271Reads||0)+1);if(!['source','execute'].includes(attrs.duration271||''))attrs.duration271Bad='true'}
   else if(path==='/api/compare'){attrs.duration271Pairs=String(Number(attrs.duration271Pairs||0)+1);if(attrs.duration271!=='execute')attrs.duration271Bad='true'}
   else if(!path.startsWith('/api/queries/'))attrs.duration271Bad='true';
  }return native(input,init)};const run=document.querySelector<HTMLButtonElement>('#exploreRun');check(!!run&&!run.disabled,'Explicit Run required');run!.click();return;
 }
 const saved=store.__gooDuration271;check(!!saved,'Actual duration source required');
 const unchanged=()=>check(chart()?.snapshot===saved.snapshot&&JSON.stringify(saved.snapshot)===saved.json,'Local selection must preserve immutable complete parent bins');
 if(phase==='brush'){
  document.querySelector<HTMLButtonElement>('#resultDistribution')!.click();const s=chart()?.snapshot;check(s?.histogram?.kind==='span-duration'&&s.histogram.total==='11'&&s.rows.length===8&&s.rows[7].values.to==='+inf','Real eight-bin duration population and open upper tail required');check(JSON.stringify(s.rows.map((r:any)=>r.values.count))===JSON.stringify(['2','1','2','1','1','1','1','2']),'Exact boundary values must belong to the expected bins');check(!s.histogram.support&&!s.histogram.binding,'Captured spans must not acquire numeric parse support');check(Object.isFrozen(s)&&Object.isFrozen(s.rows),'Source snapshot must be immutable');saved.snapshot=s;saved.json=JSON.stringify(s);choose();unchanged();attrs.duration271='brush';return;
 }
 if(phase==='manual'){
  const details=[...tray().querySelectorAll<HTMLDetailsElement>('details')].find(d=>d.querySelector('summary')?.textContent==='Selection bounds and predicates');check(!!details,'Exact manual bounds required');if(!details!.open)details!.querySelector<HTMLElement>('summary')!.click();const min=tray().querySelector<HTMLInputElement>('input[aria-label="Minimum span duration (ms)"]')!;min.value='7';min.dispatchEvent(new Event('input',{bubbles:true}));check(section().querySelectorAll('[data-range-selected=true]').length===0,'Manual bounds must clear obsolete brush geometry');check(tray().textContent?.includes('[7, 10)'),'Manual preview must show exact7to10ms');unchanged();attrs.duration271='manual';return;
 }
 if(phase==='reset'){choose();unchanged();attrs.duration271='reset';return}
 if(phase==='rest'){
  const button=[...tray().querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Compare with rest');check(!!button&&!button.disabled,'Disjoint remainder action required');button!.click();check(document.querySelector('.population-workspace')?.getAttribute('data-phase')==='staged','Compare with rest must stage before reading');check(Number(attrs.duration271Reads)===1&&Number(attrs.duration271Pairs||0)===0,'Rest staging must be source-free');attrs.duration271='rest';return;
 }
 if(phase==='execute'){
  const button=[...document.querySelectorAll<HTMLButtonElement>('.population-workspace button')].find(b=>b.textContent==='Run population view');check(!!button&&!button.disabled,'Explicit population Run required');attrs.duration271='execute';button!.click();return;
 }
 check(phase==='complete','Unknown duration phase');check(Number(attrs.duration271Reads)===3&&Number(attrs.duration271Pairs)===1,'Only explicitRun may read independent selected/rest populations');check(document.querySelector('.population-workspace [id$=cohortStatus]')?.textContent?.includes('2 selected records vs 9 baseline records'),'Boundary5ms inclusive10ms exclusive must yield2selected9rest');check(JSON.stringify(saved.snapshot)===saved.json&&Object.isFrozen(saved.snapshot),'Refinement must preserve retained parent11-span evidence');attrs.duration271='complete';
});
export const follow=actions(()=>{const s=state.current;
 if(s.phase==='initial'&&s.ready)return[act('source')];
 if(s.phase==='source'&&s.applied&&s.reads===1&&s.bins===8)return[act('brush')];
 for(const [a,b]of [['brush','manual'],['manual','reset'],['reset','rest']])if(s.phase===a)return[act(b)];
 if(s.phase==='rest'&&s.population==='staged')return[act('execute')];
 if(s.phase==='execute'&&s.population==='applied'&&s.reads===3&&s.pairs===1&&s.result.includes('2 selected records vs 9 baseline records'))return[act('complete')];
 return['Wait'];});
