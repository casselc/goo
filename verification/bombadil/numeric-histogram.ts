import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

// Real int64 OTLP data, including exact extrema and an invalid max+1 value.
const state=extract(({document})=>({version:'goo.histogram269',phase:document.documentElement.dataset.histogram269||'initial',
 ready:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&!!document.querySelector('#exploreRun:not([disabled])')?.getClientRects().length,
 applied:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied',
 reads:Number(document.documentElement.dataset.histogram269Reads||0),plans:Number(document.documentElement.dataset.histogram269Plans||0),bad:document.documentElement.dataset.histogram269Bad==='true',reason:document.documentElement.dataset.histogram269Reason||'',
 bins:document.querySelector('.numeric-distribution goo-chart')?.querySelectorAll('goo-time-bars button').length||0,
 sameRevision:document.documentElement.dataset.histogram269SameRevision==='true',distinctClock:document.documentElement.dataset.histogram269DistinctClock==='true'}));
export const explicitReadsOnly=always(()=>!state.current.bad&&state.current.reads<=3&&state.current.plans<=1);
export const brushDoesNotRead=always(()=>!['range','rerun','stale','current'].includes(state.current.phase)||state.current.plans===0);
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
const act=registerCustomAction('goo.histogram269',async(document,window,phase:string)=>{
 const attrs=document.documentElement.dataset,store=window as any;
 const check=(ok:boolean,reason:string)=>{if(!ok){attrs.histogram269Bad='true';attrs.histogram269Reason=reason;throw Error(reason)}};
 const section=()=>document.querySelector('.numeric-distribution')!;
 const chart=()=>section()?.querySelector('goo-chart') as any;
 const run=()=>{const button=document.querySelector<HTMLButtonElement>('#exploreRun');check(!!button&&!button.disabled,'Explicit Run required');button!.click()};
 const open=()=>[...section().querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent==='Run selected population')!;
 if(phase==='source'){
  attrs.histogram269='source';store.__gooHistogram269={};const native=window.fetch.bind(window);
  window.fetch=(input,init)=>{const method=init?.method||(typeof input==='object'&&'method' in input?input.method:'GET');if(method.toUpperCase()==='POST'){
   const path=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;
   if(path==='/api/explore'){attrs.histogram269Reads=String(Number(attrs.histogram269Reads||0)+1);if(!['source','rerun','execute'].includes(attrs.histogram269||''))attrs.histogram269Bad='true'}
   else if(path==='/api/numeric-histograms/selection'){attrs.histogram269Plans=String(Number(attrs.histogram269Plans||0)+1);if(attrs.histogram269!=='execute')attrs.histogram269Bad='true'}
   else if(!path.startsWith('/api/queries/'))attrs.histogram269Bad='true';
  }return native(input,init)};run();return;
 }
 const saved=store.__gooHistogram269,modulePath='/view-model.mjs',views=await import(modulePath);check(!!saved,'Actual source execution required');
 const snapshot=()=>chart()?.snapshot;
 const geometry=()=>{const s=snapshot();check(s?.shape==='histogram'&&s.rows.length===128&&s.rows[127].key==='bin:127'&&s.rows[127].values.from==='4611686018427387904'&&s.rows[127].values.to==='9223372036854775808','Exact full signed histogram geometry and max+1 sentinel required');check(Object.isFrozen(s)&&Object.isFrozen(s.rows),'Public snapshot must be immutable');return s};
 const range=()=>{const s=geometry();chart().dispatchEvent(new CustomEvent('goo-histogram-range',{detail:views.histogramIntent(s,'bin:127')}));check(!!open()&&!open().disabled,'Current exact tail selection must enable explicit population run')};
 if(phase==='range'){
  const s=geometry(),support=s.histogram.support;check(support.total==='18'&&support.valid==='13'&&support.missing==='1'&&support.empty==='1'&&support.invalid==='3','Coverage must separate valid bins from all captured records');check(s.rows[127].values.count==='1'&&s.rows.reduce((n:any,r:any)=>n+BigInt(r.values.count),0n)===13n,'Bins must sum valid coverage, preserving exactMAX membership');saved.original=s;saved.json=JSON.stringify(s);range();check(chart().snapshot===s&&JSON.stringify(s)===saved.json&&Number(attrs.histogram269Reads)===1&&Number(attrs.histogram269Plans||0)===0,'Range selection must preserve source handle and perform no planning/read');check(section().textContent?.includes('1 of 18 records selected'),'Selection context must retain complete captured population');saved.old={chart:chart(),snapshot:s,intent:views.histogramIntent(s,'bin:127'),button:open()};attrs.histogram269='range';return;
 }
 if(phase==='rerun'){attrs.histogram269='rerun';run();return}
 if(phase==='stale'){
  const s=geometry();check(s.revision===saved.old.snapshot.revision&&s.capturedAt!==saved.old.snapshot.capturedAt,'New execution must preserve content revision with distinct read clock');attrs.histogram269SameRevision='true';attrs.histogram269DistinctClock='true';saved.old.chart.dispatchEvent(new CustomEvent('goo-histogram-range',{detail:structuredClone(saved.old.intent)}));saved.old.button.dispatchEvent(new MouseEvent('click',{bubbles:true}));check(Number(attrs.histogram269Reads)===2&&Number(attrs.histogram269Plans||0)===0,'Old range and Run callbacks must not plan or read');attrs.histogram269='stale';return;
 }
 if(phase==='current'){range();check(Number(attrs.histogram269Reads)===2&&Number(attrs.histogram269Plans||0)===0,'Current brush remains source-free');attrs.histogram269='current';return}
 if(phase==='execute'){attrs.histogram269='execute';open().click();return}
 check(phase==='complete','Unknown histogram phase');const s=geometry();check(s.histogram.support.total==='1'&&s.histogram.support.valid==='1'&&s.rows[127].values.count==='1','Explicit tail run must measure exactMAX as one valid record');check(s.rows.slice(0,127).every((r:any)=>r.values.count==='0'),'Full zero bins remain supplied after narrowed run');check(s.scope.selector.numeric?.type==='int64'&&s.scope.selector.numericHistogram?.binning==='powers-of-two','Run retains declared histogram binding');const predicates=JSON.stringify(s.scope.selector.where);check(!predicates.includes('9223372036854775808'),'Geometry sentinel must never become an invalid numeric threshold');check(Number(attrs.histogram269Reads)===3&&Number(attrs.histogram269Plans)===1&&attrs.histogram269SameRevision==='true'&&attrs.histogram269DistinctClock==='true','Completion requires three explicit source reads and one checked Go plan');attrs.histogram269='complete';
});
export const follow=actions(()=>{const s=state.current;
 if(s.phase==='initial'&&s.ready)return[act('source')];
 if(s.phase==='source'&&s.applied&&s.reads===1&&s.bins>0)return[act('range')];
 if(s.phase==='range')return[act('rerun')];
 if(s.phase==='rerun'&&s.applied&&s.reads===2&&s.bins>0)return[act('stale')];
 if(s.phase==='stale')return[act('current')];
 if(s.phase==='current')return[act('execute')];
 if(s.phase==='execute'&&s.applied&&s.reads===3&&s.plans===1&&s.bins>0)return[act('complete')];
 return['Wait'];});
