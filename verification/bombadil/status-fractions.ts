import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

// Actual OTLP fixture: one recorded error in 100 frontend spans. The whole
// population has 103 spans; neither denominator is inferred from visible rows.
const state=extract(({document})=>({
 version:'goo.fractions268',phase:document.documentElement.dataset.fractions268||'initial',
 ready:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&!!document.querySelector('#exploreRun:not([disabled])')?.getClientRects().length,
 reads:Number(document.documentElement.dataset.fractions268Reads||0),bad:document.documentElement.dataset.fractions268Bad==='true',reason:document.documentElement.dataset.fractions268Reason||'',
 applied:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied',draft:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft',
 bars:document.querySelectorAll('[aria-label="Scoped breakdown"] .scope-view-bar').length,
 fraction:document.documentElement.dataset.fractions268Fraction||'',sameRevision:document.documentElement.dataset.fractions268SameRevision==='true',distinctClock:document.documentElement.dataset.fractions268DistinctClock==='true'
}));
export const explicitReadsOnly=always(()=>!state.current.bad&&state.current.reads<=3);
export const rowDenominatorSurvivesExploration=always(()=>['initial','source'].includes(state.current.phase)||state.current.fraction==='1/100');
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');

const act=registerCustomAction('goo.fractions268',async(document,window,phase:string)=>{
 const attrs=document.documentElement.dataset,store=window as any,key='status.error_fraction';
 const check=(condition:boolean,reason:string)=>{if(!condition){attrs.fractions268Bad='true';attrs.fractions268Reason=reason;throw Error(reason)}};
 const region=()=>document.querySelector('[aria-label="Scoped breakdown"]')!;
 const chart=()=>region()?.querySelector('goo-chart') as any;
 const show=()=>{const tab=document.querySelector<HTMLButtonElement>('#resultDistribution');check(!!tab,'Distribution tab required');if(tab!.getAttribute('aria-selected')!=='true')tab!.click()};
 const run=()=>{const button=document.querySelector<HTMLButtonElement>('#exploreRun');check(!!button&&!button.disabled,'Explicit Run required');button!.click()};
 const choose=(label:string,value:string)=>{const control=region()?.querySelector<HTMLSelectElement>('select[aria-label="'+label+'"]');check(!!control,'Missing '+label);control!.value=value;control!.dispatchEvent(new Event('change',{bubbles:true}))};
 const find=(value:string)=>{const input=document.querySelector<HTMLInputElement>('#facetSearch');check(!!input,'Shared Find required');input!.value=value;input!.dispatchEvent(new Event('input',{bubbles:true}))};
 if(phase==='source'){
  attrs.fractions268='source';store.__gooFractions268={};const native=window.fetch.bind(window);
  window.fetch=(input,init)=>{const method=init?.method||(typeof input==='object'&&'method' in input?input.method:'GET');if(method.toUpperCase()==='POST'){
   const path=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;
   if(path==='/api/explore'){attrs.fractions268Reads=String(Number(attrs.fractions268Reads||0)+1);if(!['source','rerun','execute'].includes(attrs.fractions268||''))attrs.fractions268Bad='true'}
   else if(!path.startsWith('/api/queries/'))attrs.fractions268Bad='true';
  }return native(input,init)};run();return;
 }
 const saved=store.__gooFractions268,modulePath='/view-model.mjs',views=await import(modulePath);check(!!saved,'Actual source run required');
 const front=()=>chart()?.snapshot.rows.find((row:any)=>row.values.category===saved.name);
 const exact=()=>{const current=chart(),row=front();check(!!row&&row.values.count==='100'&&row.values['status.error']==='1'&&row.values['status.ok']==='99'&&row.values[key]==='1/100','Recorded fraction must retain its own exact class counts and row denominator');const field=current.snapshot.fields.find((field:any)=>field.key===key);check(field?.aggregation?.numerator==='status.error'&&field.aggregation.denominator==='count'&&field.unit==='1','Typed fraction operands must remain declared');attrs.fractions268Fraction=row.values[key]};
 const onePercent=()=>{exact();const bar=[...chart().querySelectorAll<HTMLElement>('.scope-view-bar')].find((bar:any)=>bar.dataset.key===front().key) as HTMLElement|undefined;check(!!bar&&bar.querySelector<HTMLElement>('.scope-view-fill')?.style.width==='1%'&&bar.textContent?.includes('1%'),'One error in100 must use1% geometry, even without a100% neighbor')};
 const unchanged=()=>check(chart().snapshot===saved.current&&JSON.stringify(chart().snapshot)===saved.json,'Local display must retain immutable snapshot identity and content');
 if(phase==='find'){
  show();const snapshot=chart()?.snapshot;check(snapshot?.rows.length===3&&snapshot.partition?.total==='103','Actual three-service103-span source required');const row=snapshot.rows.find((row:any)=>row.values.count==='100');check(!!row,'Frontend100-span group required');saved.name=row.values.category;saved.key=row.key;saved.current=snapshot;saved.json=JSON.stringify(snapshot);choose('Breakdown measure',key);find(saved.name);onePercent();unchanged();check(!document.querySelector('#exploreGroups goo-chart'),'Records must not duplicate the status breakdown');check(Number(attrs.fractions268Reads)===1,'Find must not read');attrs.fractions268='find';return;
 }
 if(phase==='sort'){
  const more=[...region().querySelectorAll<HTMLDetailsElement>('details')].find(d=>d.querySelector('summary')?.textContent==='More options');check(!!more,'Local ordering required');if(!more!.open)more!.querySelector('summary')!.click();find('');choose('Group order','value-desc');const first=chart().snapshot.rows.find((row:any)=>row.key===chart().querySelector('.scope-view-bar')?.dataset.key);check(first?.values.count==='1'&&first.values[key]==='1','Fraction ordering must rank worker100% before frontend1%');find(saved.name);onePercent();unchanged();attrs.fractions268='sort';return;
 }
 if(phase==='table'){
  choose('Breakdown view','table');check(!!chart().querySelector('table')&&chart().textContent.includes('1%'),'Table must format typed fraction as percent');exact();choose('Breakdown view','bar');onePercent();unchanged();attrs.fractions268='table';return;
 }
 if(phase==='classes'){
  find('');choose('Breakdown measure','status.unset');check(chart().textContent.includes('2 records'),'Unset status remains a separate exact count');choose('Breakdown measure',key);find(saved.name);onePercent();unchanged();check(Number(attrs.fractions268Reads)===1,'Measure/kind/order/Find must remain local');attrs.fractions268='classes';return;
 }
 if(phase==='rerun'){
  chart().querySelector<HTMLButtonElement>('.scope-view-bar')!.click();const stage=[...region().querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Stage selected group');check(!!stage&&!stage.disabled,'Current selection should stage');saved.old={chart:chart(),stage,snapshot:chart().snapshot,intent:views.selectionIntent(chart().snapshot,saved.key)};attrs.fractions268='rerun';run();return;
 }
 if(phase==='stale'){
  show();const snapshot=chart()?.snapshot;check(snapshot?.revision===saved.old.snapshot.revision&&snapshot.capturedAt!==saved.old.snapshot.capturedAt,'Rerun must observe same contents with a distinct read clock');attrs.fractions268SameRevision='true';attrs.fractions268DistinctClock='true';saved.current=snapshot;saved.json=JSON.stringify(snapshot);saved.old.chart.dispatchEvent(new CustomEvent('goo-select',{detail:structuredClone(saved.old.intent)}));saved.old.stage.dispatchEvent(new MouseEvent('click',{bubbles:true}));check(document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied'&&Number(attrs.fractions268Reads)===2,'Retired callback must not stage or read');choose('Breakdown measure',key);find(saved.name);onePercent();attrs.fractions268='stale';return;
 }
 if(phase==='stage'){
  onePercent();chart().querySelector<HTMLButtonElement>('.scope-view-bar')!.click();saved.stage=[...region().querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Stage selected group');check(!!saved.stage&&!saved.stage.disabled,'Current frontend selection required');saved.stage.click();check(document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&Number(attrs.fractions268Reads)===2,'Current group stages without I/O');attrs.fractions268='stage';return;
 }
 if(phase==='execute'){saved.stage.dispatchEvent(new MouseEvent('click',{bubbles:true}));check(Number(attrs.fractions268Reads)===2,'Staging invalidates former binding');attrs.fractions268='execute';run();return}
 check(phase==='complete','Unknown status fraction phase');show();choose('Breakdown measure',key);find(saved.name);onePercent();const snapshot=chart().snapshot;check(snapshot.rows.length===1&&snapshot.partition.total==='100','Explicit narrowed run must retain100-span frontend denominator');check(snapshot.scope.selector.spanStatus===true&&snapshot.scope.selector.summary==='span-duration'&&snapshot.scope.selector.numeric?.unit==='USD','Staging must preserve all requested profiles');check(snapshot.scope.selector.where?.all?.some((p:any)=>p.field?.kind==='column'&&p.field.key==='service'&&p.value===saved.name),'Explicit run must use staged service predicate');check(Number(attrs.fractions268Reads)===3&&attrs.fractions268SameRevision==='true'&&attrs.fractions268DistinctClock==='true','Three explicit reads and stale-execution fence required');attrs.fractions268='complete';
});
export const follow=actions(()=>{
 const s=state.current;
 if(s.phase==='initial'&&s.ready)return[act('source')];
 if(s.phase==='source'&&s.applied&&s.reads===1&&s.bars===3)return[act('find')];
 for(const [from,to]of [['find','sort'],['sort','table'],['table','classes'],['classes','rerun']])if(s.phase===from&&s.reads===1)return[act(to)];
 if(s.phase==='rerun'&&s.applied&&s.reads===2&&s.bars>0)return[act('stale')];
 if(s.phase==='stale'&&s.reads===2)return[act('stage')];
 if(s.phase==='stage'&&s.draft&&s.reads===2)return[act('execute')];
 if(s.phase==='execute'&&s.applied&&s.reads===3&&s.bars===1)return[act('complete')];
 return['Wait'];
});
