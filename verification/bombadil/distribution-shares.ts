import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

// One real grouped population is viewed through local transformations before an
// explicit scope change. The driver never constructs a source or action witness.
const state=extract(({document})=>({
 version:'goo.shares267',phase:document.documentElement.dataset.shares267||'initial',
 ready:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&!!document.querySelector('#exploreRun:not([disabled])')?.getClientRects().length,
 reads:Number(document.documentElement.dataset.shares267Reads||0),bad:document.documentElement.dataset.shares267Bad==='true',reason:document.documentElement.dataset.shares267Reason||'',
 applied:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied',draft:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft',
 bars:document.querySelectorAll('[aria-label="Scoped breakdown"] .scope-view-bar').length,
 width:document.querySelector<HTMLElement>('[aria-label="Scoped breakdown"] .scope-view-fill')?.style.width||'',
 denominator:document.documentElement.dataset.shares267Denominator||'',fraction:document.documentElement.dataset.shares267Fraction||'',
 sameRevision:document.documentElement.dataset.shares267SameRevision==='true',distinctClock:document.documentElement.dataset.shares267DistinctClock==='true'
}));
export const explicitScopeReadsOnly=always(()=>!state.current.bad&&state.current.reads<=3);
export const localFindKeepsParentPopulation=always(()=>!['find','sort','max','population','measure','rerun','stale','stage'].includes(state.current.phase)||state.current.denominator==='4'&&state.current.fraction==='1/4');
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');

const act=registerCustomAction('goo.shares267',async(document,window,phase:string)=>{
 const attrs=document.documentElement.dataset,store=window as any;
 const check=(condition:boolean,reason:string)=>{if(!condition){attrs.shares267Bad='true';attrs.shares267Reason=reason;throw Error(reason)}};
 const showDistribution=()=>{const tab=document.querySelector<HTMLButtonElement>('#resultDistribution');check(!!tab,'The ordinary Distribution tab is required');if(tab!.getAttribute('aria-selected')!=='true')tab!.click()};
 const run=()=>{const button=document.querySelector<HTMLButtonElement>('#exploreRun');check(!!button&&!button.disabled&&!!button.getClientRects().length,'The fixed question must run explicitly');button!.click()};
 const region=()=>document.querySelector('[aria-label="Scoped breakdown"]')!;
 const chart=()=>region()?.querySelector('goo-chart') as any;
 const find=(value:string)=>{const input=document.querySelector<HTMLInputElement>('#facetSearch');check(!!input,'The shared Find input is required');input!.value=value;input!.dispatchEvent(new Event('input',{bubbles:true}))};
 const choose=(label:string,value:string)=>{const control=region()?.querySelector<HTMLSelectElement>('select[aria-label="'+label+'"]');check(!!control,'Missing presentation control '+label);control!.value=value;control!.dispatchEvent(new Event('change',{bubbles:true}))};
 const applied=()=>document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied';
 if(phase==='source'){
  attrs.shares267='source';store.__gooShares267={};const native=window.fetch.bind(window);
  window.fetch=(input,init)=>{
   const method=init?.method||(typeof input==='object'&&'method' in input?input.method:'GET');
   if(method.toUpperCase()==='POST'){
    const pathname=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;
    if(pathname==='/api/explore'){
     attrs.shares267Reads=String(Number(attrs.shares267Reads||0)+1);
     if(!['source','rerun','execute'].includes(attrs.shares267||''))attrs.shares267Bad='true';
    }else if(!pathname.startsWith('/api/queries/'))attrs.shares267Bad='true';
   }
   return native(input,init);
  };
  run();return;
 }
 const saved=store.__gooShares267,modulePath='/view-model.mjs',views=await import(modulePath);
 check(!!saved,'A real grouped source execution is required');
 const observeShare=(expectedTotal:string,expectedFraction:string)=>{
  const current=chart(),snapshot=current?.snapshot,row=snapshot?.rows.find((row:any)=>row.values.category===saved.worker);
  check(!!row,'The supplied worker group must remain inspectable');const share=views.populationShare(snapshot,row.key);
  check(share.ok&&share.total===expectedTotal&&share.fraction===expectedFraction,'Shares must use the original exhaustive group-read denominator');
  attrs.shares267Denominator=share.total;attrs.shares267Fraction=share.fraction;
  return share;
 };
 const unchanged=()=>check(chart()?.snapshot===saved.current&&JSON.stringify(chart().snapshot)===saved.currentJSON,'Local presentation must preserve the original immutable snapshot');
 const quarter=()=>{const share=observeShare('4','1/4'),current=chart();check(share.percent==='25%'&&current.querySelectorAll('.scope-view-bar').length===1&&current.querySelector('.scope-view-fill')?.style.width==='25%','Finding one of four captured records must retain25%, not renormalize to100%')};
 if(phase==='find'){
  showDistribution();const snapshot=chart()?.snapshot;
  check(snapshot?.partition?.total==='4'&&snapshot.rows.length===2,'The real source must contain the expected three-to-one record partition');
  const worker=snapshot.rows.find((row:any)=>row.values.count==='1');check(!!worker&&typeof worker.values.category==='string','A unique supplied worker group is required');
  saved.original=snapshot;saved.current=snapshot;saved.currentJSON=JSON.stringify(snapshot);saved.worker=worker.values.category;saved.workerKey=worker.key;
  check(!document.querySelector('#exploreGroups goo-chart'),'Records must not repeat the ordinary Distribution breakdown');
  find(saved.worker);quarter();unchanged();check(Number(attrs.shares267Reads)===1,'Find must not read telemetry');attrs.shares267='find';return;
 }
 if(phase==='sort'){
  const more=[...region().querySelectorAll<HTMLDetailsElement>('details')].find(details=>details.querySelector('summary')?.textContent==='More options');check(!!more,'More options must contain local ordering');if(!more!.open)more!.querySelector('summary')!.click();
  find('');choose('Group order','value-asc');check(chart().querySelector('.scope-view-bar')?.dataset.key===saved.workerKey,'Numeric ordering must use the supplied captured counts');find(saved.worker);quarter();unchanged();check(Number(attrs.shares267Reads)===1,'Local ordering must not read telemetry');attrs.shares267='sort';return;
 }
 if(phase==='max'){
  choose('Count scale','max');const width=parseFloat(chart()?.querySelector('.scope-view-fill')?.style.width||'');
  check(Math.abs(width-100/3)<0.01,'Max scale must keep the largest full-result group even when Find hides it');observeShare('4','1/4');unchanged();attrs.shares267='max';return;
 }
 if(phase==='population'){
  choose('Count scale','population');quarter();unchanged();attrs.shares267='population';return;
 }
 if(phase==='measure'){
  find('');choose('Breakdown measure','duration.p95');
  const field=chart()?.snapshot.fields.find((field:any)=>field.key==='duration.p95');check(field?.aggregation?.method==='quantilesExactHigh','Changing measures must retain the exact captured empirical quantile method');
  check(chart().querySelectorAll('.scope-view-bar').length===2,'The supplied duration measure must render without another source read');
  choose('Breakdown measure','count');find(saved.worker);choose('Breakdown view','table');check(!!chart().querySelector('table')&&chart().textContent.includes('25%'),'Exact tables must retain full-population shares');choose('Breakdown view','bar');quarter();unchanged();check(Number(attrs.shares267Reads)===1,'Kind and measure switching must remain local');attrs.shares267='measure';return;
 }
 if(phase==='rerun'){
  chart().querySelector<HTMLButtonElement>('.scope-view-bar')!.click();const current=chart(),stage=[...region().querySelectorAll<HTMLButtonElement>('button')].find(button=>button.textContent==='Stage selected group');
  check(!!stage&&!stage.disabled,'A current supplied selection must enable staging');saved.old={chart:current,stage,intent:views.selectionIntent(current.snapshot,saved.workerKey),snapshot:current.snapshot};attrs.shares267='rerun';run();return;
 }
 if(phase==='stale'){
  showDistribution();const snapshot=chart()?.snapshot;check(snapshot?.revision===saved.old.snapshot.revision&&snapshot.capturedAt!==saved.old.snapshot.capturedAt,'A second execution must have identical contents but a distinct actual read clock');
  attrs.shares267SameRevision='true';attrs.shares267DistinctClock='true';saved.current=snapshot;saved.currentJSON=JSON.stringify(snapshot);
  saved.old.chart.dispatchEvent(new CustomEvent('goo-select',{detail:structuredClone(saved.old.intent)}));saved.old.stage.dispatchEvent(new MouseEvent('click',{bubbles:true}));
  check(applied()&&Number(attrs.shares267Reads)===2,'Retired ordinary-view callbacks must not stage or rerun');find(saved.worker);quarter();unchanged();attrs.shares267='stale';return;
 }
 if(phase==='stage'){
  quarter();chart().querySelector<HTMLButtonElement>('.scope-view-bar')!.click();saved.currentStage=[...region().querySelectorAll<HTMLButtonElement>('button')].find(button=>button.textContent==='Stage selected group');
  check(!!saved.currentStage&&!saved.currentStage.disabled,'The current worker selection must be stageable');saved.currentStage.click();check(document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&Number(attrs.shares267Reads)===2,'A group selection must stage without a source read');attrs.shares267='stage';return;
 }
 if(phase==='execute'){
  saved.currentStage.dispatchEvent(new MouseEvent('click',{bubbles:true}));check(document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&Number(attrs.shares267Reads)===2,'Staging must retire its previous binding');attrs.shares267='execute';run();return;
 }
 check(phase==='complete','Unknown ordinary breakdown phase');showDistribution();observeShare('1','1/1');
 const snapshot=chart()?.snapshot;check(chart().querySelectorAll('.scope-view-bar').length===1&&chart().textContent.includes('100%'),'Only an explicitly rerun narrowed scope may show the worker as100%');
 check(snapshot.scope.selector.where?.all?.some((p:any)=>p.field?.kind==='column'&&p.field.key==='service'&&p.op==='eq'&&p.value===saved.worker),'The executed source scope must contain the staged worker predicate');
 check(snapshot.scope.selector.numeric?.unit==='USD'&&snapshot.scope.selector.summary==='span-duration','Scope refinement must preserve the declared numeric and duration profiles');
 check(Number(attrs.shares267Reads)===3&&attrs.shares267SameRevision==='true'&&attrs.shares267DistinctClock==='true','Completion requires all three explicit source reads and the observed stale-execution fence');attrs.shares267='complete';
});

export const follow=actions(()=>{
 const s=state.current;
 if(s.phase==='initial'&&s.ready)return[act('source')];
 if(s.phase==='source'&&s.applied&&s.reads===1&&s.bars===2)return[act('find')];
 if(s.phase==='find'&&s.width==='25%'&&s.reads===1)return[act('sort')];
 if(s.phase==='sort'&&s.width==='25%'&&s.reads===1)return[act('max')];
 if(s.phase==='max'&&s.reads===1)return[act('population')];
 if(s.phase==='population'&&s.width==='25%'&&s.reads===1)return[act('measure')];
 if(s.phase==='measure'&&s.width==='25%'&&s.reads===1)return[act('rerun')];
 if(s.phase==='rerun'&&s.applied&&s.reads===2&&s.bars>0)return[act('stale')];
 if(s.phase==='stale'&&s.width==='25%'&&s.reads===2)return[act('stage')];
 if(s.phase==='stage'&&s.draft&&s.reads===2)return[act('execute')];
 if(s.phase==='execute'&&s.applied&&s.reads===3&&s.bars===1)return[act('complete')];
 return['Wait'];
});
