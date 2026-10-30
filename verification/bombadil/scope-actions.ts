import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

// Exercise real mounted callbacks and real source results. No action context,
// source adapter, execution epoch or source witness is fabricated by this driver.
const state=extract(({document})=>({
 version:'goo.scope266',phase:document.documentElement.dataset.scope266||'initial',
 ready:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&!!document.querySelector('#exploreRun:not([disabled])')?.getClientRects().length,
 reads:Number(document.documentElement.dataset.scope266Reads||0),bad:document.documentElement.dataset.scope266Bad==='true',reason:document.documentElement.dataset.scope266Reason||'',
 applied:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied',draft:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft',
 visible:!!document.querySelector('#exploration')?.getClientRects().length,
 bars:document.querySelectorAll('[aria-label="Scoped breakdown"] goo-chart .scope-view-bar').length,
 intervals:document.querySelectorAll('[aria-label="Scoped timeline"] .scope-time-interval').length,
 sameRevision:document.documentElement.dataset.scope266SameRevision==='true',distinctClock:document.documentElement.dataset.scope266DistinctClock==='true',
 withdrawn:!!document.querySelector('[aria-label="Scoped breakdown"] .scoped-breakdown-selection')?.textContent?.includes('Run again'),
 staged:document.querySelector('#exploreStatus')?.textContent?.includes('Group filter staged')||false
}));
export const explicitRunsOnly=always(()=>!state.current.bad&&state.current.reads<=3);
export const identicalContentsDoNotAuthorizeOldActions=always(()=>!['stale','suspend','suspended','resume','stage','complete'].includes(state.current.phase)||state.current.sameRevision&&state.current.distinctClock);
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');

const act=registerCustomAction('goo.scope266',async(document,window,phase:string)=>{
 const attrs=document.documentElement.dataset,store=window as any;
 const check=(condition:boolean,reason:string)=>{if(!condition){attrs.scope266Bad='true';attrs.scope266Reason=reason;throw Error(reason)}};
 const run=()=>{const button=document.querySelector<HTMLButtonElement>('#exploreRun');check(!!button&&!button.disabled&&!!button.getClientRects().length,'The visible fixed question must be run explicitly');button!.click()};
 const navigate=(workspace:string)=>{const control=document.querySelector<HTMLElement>('[data-workspace="'+workspace+'"]');check(!!control,'Workspace control is required: '+workspace);control!.click()};
 const applied=()=>document.querySelector('#exploration')?.getAttribute('data-question-phase')==='applied';
 const showRecords=()=>{const tab=document.querySelector<HTMLButtonElement>('#resultRecords');check(!!tab,'The Records result tab is required');if(tab!.getAttribute('aria-selected')!=='true')tab!.click()};
 const replay=(saved:any)=>{
  check(!!saved?.facet&&!!saved.groupStage&&!!saved.windowStage,'Retained facet, group and interval callbacks are required');
  saved.facet.dispatchEvent(new CustomEvent('goo-scope',{detail:structuredClone(saved.intent)}));
  // Invoke retired listeners independently of their disabled-button affordance.
  saved.groupStage.dispatchEvent(new MouseEvent('click',{bubbles:true}));saved.windowStage.dispatchEvent(new MouseEvent('click',{bubbles:true}));
 };
 const retain=async()=>{
  showRecords();
  const group=document.querySelector('[aria-label="Scoped breakdown"]')!,timeline=document.querySelector('[aria-label="Scoped timeline"]')!;
  check(!!group&&!!timeline,'Applied group and interval views are required');
  const bar=group.querySelector<HTMLButtonElement>('.scope-view-bar'),interval=timeline.querySelector<HTMLButtonElement>('.scope-time-interval[data-state="observed"]');
  check(!!bar&&!!interval,'Known source group and interval selections are required');bar!.click();interval!.click();
  const snapshot=(group.querySelector('goo-chart') as any).snapshot,modulePath='/view-model.mjs',{scopeIntent}=await import(modulePath);
  const groupStage=[...group.querySelectorAll<HTMLButtonElement>('button')].find(button=>button.textContent==='Stage selected group'),windowStage=[...timeline.querySelectorAll<HTMLButtonElement>('button')].find(button=>button.textContent==='Stage selected interval');
  check(!!groupStage&&!groupStage.disabled&&!!windowStage&&!windowStage.disabled,'Current source selections must enable their own staging actions');
  return {facet:group.querySelector('goo-scope-facet'),intent:scopeIntent(snapshot,'category',snapshot.rows[0].values.category),groupStage,windowStage,snapshot};
 };
 if(phase==='source'){
  attrs.scope266='source';store.__gooScope266={};const native=window.fetch.bind(window);
  window.fetch=(input,init)=>{
   const method=init?.method||(typeof input==='object'&&'method' in input?input.method:'GET');
   if(method.toUpperCase()==='POST'){
    const pathname=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;
    if(pathname==='/api/explore'){
     attrs.scope266Reads=String(Number(attrs.scope266Reads||0)+1);
     if(!['source','rerun','resume'].includes(attrs.scope266||''))attrs.scope266Bad='true';
    }else if(!pathname.startsWith('/api/queries/'))attrs.scope266Bad='true';
   }
   return native(input,init);
  };
  run();return;
 }
 const saved=store.__gooScope266;check(!!saved,'An actual fixed source execution is required');
 if(phase==='rerun'){
  saved.old=await retain();attrs.scope266='rerun';run();return;
 }
 if(phase==='stale'){
  showRecords();
  const current=(document.querySelector('[aria-label="Scoped breakdown"] goo-chart') as any)?.snapshot;
  check(!!current&&current.revision===saved.old.snapshot.revision,'The explicit rerun must return the same immutable source revision');
  check(current.capturedAt!==saved.old.snapshot.capturedAt,'The identical-content rerun must have a distinct actual read clock');
  attrs.scope266SameRevision='true';attrs.scope266DistinctClock='true';replay(saved.old);
  check(applied()&&Number(attrs.scope266Reads)===2,'Retired callbacks must neither stage nor query after an identical-content rerun');
  saved.suspended=await retain();attrs.scope266='stale';return;
 }
 if(phase==='suspend'){attrs.scope266='suspend';navigate('services');return}
 if(phase==='suspended'){attrs.scope266='suspended';navigate('explore');return}
 if(phase==='resume'){
  showRecords();
  replay(saved.suspended);
  check(applied()&&Number(attrs.scope266Reads)===2,'Returning to held contents must not revive retired callbacks or read telemetry');
  const groups=document.querySelector('[aria-label="Scoped breakdown"]')!,facets=[...groups.querySelectorAll<HTMLButtonElement>('goo-scope-facet button')];
  check(!!groups.querySelector('.scope-view-bar')&&facets.length>0&&facets.every(button=>button.disabled),'Held charts may remain readable but their scope facets must stay withdrawn');
  check(saved.suspended.groupStage.disabled&&saved.suspended.windowStage.disabled,'Returning must require an explicit Run before staging');
  attrs.scope266='resume';run();return;
 }
 if(phase==='stage'){
  saved.current=await retain();saved.current.groupStage.click();
  check(document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&Number(attrs.scope266Reads)===3,'A current selection must stage its captured question without an additional source read');
  check(!!document.querySelector('#exploreStatus')?.textContent?.includes('Group filter staged'),'Current source staging must be independently visible');
  attrs.scope266='stage';return;
 }
 check(phase==='complete','Unknown scoped action phase');replay(saved.current);
 check(document.querySelector('#exploration')?.getAttribute('data-question-phase')==='draft'&&Number(attrs.scope266Reads)===3,'Staging must retire the successful execution before another retained callback can act');
 check(attrs.scope266SameRevision==='true'&&attrs.scope266DistinctClock==='true','Completion requires the observed identical-content execution boundary');
 attrs.scope266='complete';
});

export const follow=actions(()=>{
 const s=state.current;
 if(s.phase==='initial'&&s.ready)return[act('source')];
 if(s.phase==='source'&&s.applied&&s.reads===1&&s.bars===2&&s.intervals===5)return[act('rerun')];
 if(s.phase==='rerun'&&s.applied&&s.reads===2&&s.bars===2&&s.intervals===5)return[act('stale')];
 if(s.phase==='stale'&&s.applied&&s.reads===2)return[act('suspend')];
 if(s.phase==='suspend'&&!s.visible&&s.reads===2)return[act('suspended')];
 if(s.phase==='suspended'&&s.visible&&s.withdrawn&&s.reads===2)return[act('resume')];
 if(s.phase==='resume'&&s.applied&&s.reads===3&&s.bars===2&&s.intervals===5)return[act('stage')];
 if(s.phase==='stage'&&s.draft&&s.staged&&s.reads===3)return[act('complete')];
 return['Wait'];
});
