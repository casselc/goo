import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

// The only source read is the ordinary staged question's Run. All subsequent
// actions consume the actual admitted chart handle; this driver creates no
// source adapter and cannot turn a serialized claim into a source witness.
const state=extract(({document})=>({
 version:'goo.snapshot265',phase:document.documentElement.dataset.snapshot265||'initial',
 ready:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&!!document.querySelector('#exploreRun:not([disabled])')?.getClientRects().length,
 reads:Number(document.documentElement.dataset.snapshot265Reads||0),bad:document.documentElement.dataset.snapshot265Bad==='true',
 reason:document.documentElement.dataset.snapshot265Reason||'',entry:document.documentElement.dataset.snapshot265Entry||'',
 complete:document.documentElement.dataset.snapshot265Complete==='true',frozen:document.documentElement.dataset.snapshot265Frozen==='true',
 identity:document.documentElement.dataset.snapshot265Identity==='true',rows:Number(document.documentElement.dataset.snapshot265Rows||0),
 bars:document.querySelectorAll('[aria-label="Scoped breakdown"] goo-chart .scope-view-bar').length,
 refused:!!document.querySelector('[aria-label="Scoped breakdown"] goo-chart .scope-view-unavailable')
}));
export const oneSourceRead=always(()=>!state.current.bad&&state.current.reads<=1);
export const importedAndProjectedAreNotComplete=always(()=>!['copy','subset','invalid','complete'].includes(state.current.phase)||!state.current.complete);
export const sourceIdentityIsImmutable=always(()=>!['source','copy','subset','restore'].includes(state.current.phase)||state.current.frozen&&state.current.identity);
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');

const act=registerCustomAction('goo.snapshot265',async(document,window,phase:string)=>{
 const attrs=document.documentElement.dataset;
 const check=(condition:boolean,reason:string)=>{if(!condition){attrs.snapshot265Bad='true';attrs.snapshot265Reason=reason;throw Error(reason)}};
 if(phase==='started'){
  attrs.snapshot265=phase;
  const native=window.fetch.bind(window);
  window.fetch=(input,init)=>{
   const method=init?.method||(typeof input==='object'&&'method' in input?input.method:'GET');
   if(method.toUpperCase()==='POST'){
    const pathname=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;
    if(pathname==='/api/explore'){
     attrs.snapshot265Reads=String(Number(attrs.snapshot265Reads||0)+1);
     if(attrs.snapshot265!=='started')attrs.snapshot265Bad='true';
    }else if(!pathname.startsWith('/api/queries/'))attrs.snapshot265Bad='true';
   }
   return native(input,init);
  };
  const run=[...document.querySelectorAll<HTMLButtonElement>('button')].find(button=>button.textContent==='Run question'&&!button.disabled&&button.getClientRects().length);
  check(!!run,'A staged question with an explicit Run is required');run!.click();return;
 }
 const modulePath='/snapshots.mjs',snapshots=await import(modulePath),store=window as any;
 const immutable=(root:any)=>{
  const pending=[root],seen=new Set();
  while(pending.length){const value=pending.pop();if(value===null||typeof value!=='object'||seen.has(value))continue;if(!Object.isFrozen(value))return false;seen.add(value);pending.push(...Object.values(value))}
  return true;
 };
 if(phase==='source'){
  const chart=document.querySelector('[aria-label="Scoped breakdown"] goo-chart') as any,source=chart?.snapshot;
  check(!!chart&&snapshots.isSnapshot(source)&&source.shape==='ranked'&&source.rows.length>=2,'Run must supply an actual admitted multi-row grouped snapshot');
  check(snapshots.getSnapshotInfo(source)?.entry==='source'&&snapshots.hasGuarantee(source,'groups'),'Actual source snapshot must carry its existing groups witness');
  chart.display={kind:'bar',category:'category',value:'count',requiresComplete:true};
  const before=source.rows[0].values.count;
  check(!Reflect.set(source.rows[0].values,'count','999')&&source.rows[0].values.count===before,'Admitted values must reject mutation');
  check(!Reflect.set(source.rows,'length',0),'Admitted row arrays must reject mutation');
  store.__gooSnapshot265={chart,source,originalRows:source.rows.length};
 }else{
  const saved=store.__gooSnapshot265;check(!!saved,'The actual source handle must be retained');
  const {chart,source}=saved;
  if(phase==='copy'){
   const copied=structuredClone(source);copied.rows=copied.rows.slice(0,1);chart.snapshot=copied;
   check(chart.snapshot!==copied&&snapshots.getSnapshotInfo(chart.snapshot)?.entry==='held','Plain property copies must enter conservatively as held data');
   check(chart.snapshot.origin==='browser-held'&&chart.snapshot.completeness===undefined,'Plain copies must lose source claims');
  }else if(phase==='subset'){
   check(snapshots.selectRows(source,source.rows.map((row:any)=>row.key).reverse())===source,'All-row selection must preserve original handle identity and parent order');
   const subset=snapshots.selectRows(source,[source.rows[source.rows.length-1].key]);chart.snapshot=subset;
   check(chart.snapshot===subset&&snapshots.getSnapshotInfo(subset)?.entry==='projection','Public row projection must retain its admitted identity');
   check(subset.id!==source.id&&subset.revision!==source.revision&&subset.quality.extent==='partial'&&subset.completeness===undefined,'A proper subset must have derivative identity and no complete-parent claim');
   check(subset.rows[0]===source.rows[source.rows.length-1],'Projection must share already immutable exact row values');
  }else if(phase==='restore'){
   chart.snapshot=source;check(chart.snapshot===source,'Reusing the original handle must restore its exact identity');
  }else if(phase==='invalid'){
   chart.snapshot={};check(chart.snapshot===null&&!chart.querySelector('.scope-view-bars')&&!!chart.querySelector('.scope-view-unavailable'),'Invalid property data must clear the previous successful chart and show refusal');
  }else check(phase==='complete','Unknown boundary action');
  check(source.rows.length===saved.originalRows&&immutable(source),'Copies, projections and invalid input must not mutate the source handle');
 }
 const saved=store.__gooSnapshot265,chart=saved.chart,current=chart.snapshot,info=snapshots.getSnapshotInfo(current);
 const complete=snapshots.hasGuarantee(current,'groups');
 attrs.snapshot265Entry=info?.entry||'invalid';attrs.snapshot265Complete=String(complete);attrs.snapshot265Frozen=String(!!current&&immutable(current)&&immutable(info));
 attrs.snapshot265Identity=String(current===null?phase==='invalid'||phase==='complete':chart.snapshot===current&&chart.read('snapshot')===current);
 attrs.snapshot265Rows=String(current?.rows.length||0);
 check(Number(attrs.snapshot265Reads)===1,'Local admission, projection and restoration must not read telemetry');
 if(['source','restore'].includes(phase))check(current===saved.source&&complete&&chart.querySelectorAll('.scope-view-bar').length===saved.originalRows,'The original source witness must render without being reconstructed');
 if(['copy','subset'].includes(phase))check(!complete&&!chart.querySelector('.scope-view-bars')&&!!chart.querySelector('.scope-view-unavailable'),'Held or projected snapshots must refuse the complete-only chart');
 if(phase==='complete')check(current===null&&info===null&&!complete&&!!chart.querySelector('.scope-view-unavailable'),'Completion requires an observed invalid-input refusal, not an old success');
 attrs.snapshot265=phase;
});

export const follow=actions(()=>{
 const s=state.current;
 if(s.phase==='initial'&&s.ready)return[act('started')];
 if(s.phase==='started'&&s.reads===1&&s.bars>=2)return[act('source')];
 if(s.phase==='source'&&s.entry==='source'&&s.complete&&s.frozen&&s.identity)return[act('copy')];
 if(s.phase==='copy'&&s.entry==='held'&&!s.complete&&s.refused&&s.rows===1)return[act('subset')];
 if(s.phase==='subset'&&s.entry==='projection'&&!s.complete&&s.refused&&s.rows===1)return[act('restore')];
 if(s.phase==='restore'&&s.entry==='source'&&s.complete&&s.frozen&&s.identity&&s.bars>=2)return[act('invalid')];
 if(s.phase==='invalid'&&s.entry==='invalid'&&!s.complete&&s.refused&&s.rows===0)return[act('complete')];
 return['Wait'];
});
