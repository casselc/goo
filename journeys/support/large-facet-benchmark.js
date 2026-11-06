/* Supplied-component diagnostic, not a substitute for the live OTel journey.
 * The caller owns browser scheduling and observes network traffic independently.
 */
async function runLargeFacetBenchmark(page,{size=1000,labelBytes=384,iterations=4}={}){
 if(![1000,2000].includes(size)||![384,2048].includes(labelBytes)||!Number.isInteger(iterations)||iterations<1||iterations>10)throw Error('Choose a bounded large-facet diagnostic.');
 return page.evaluate(async({size,labelBytes,iterations})=>{
  const {create}=await import('/facet-declaration.mjs');await import('/categorical-facet.mjs');
  const field={kind:'attribute',key:'benchmark.category'},q={signal:'logs',environment:'benchmark',fromMs:1000,toMs:2000,limit:1,offset:0,groupBy:[{kind:'column',key:'service'}]},paired=size===2000;
  const label=(side,index)=>{const suffix=' terminal-'+side+'-'+String(index).padStart(4,'0');return 'A common category prefix '.padEnd(labelBytes-suffix.length,'x')+suffix;};
  const items=side=>Array.from({length:1000},(_,index)=>({value:side==='selected'&&index===0?null:side==='selected'&&index===1?'':label(side,index),count:side==='selected'?'1':'2'}));
  const state=side=>({state:'ready',field,items:items(side),total:side==='selected'?'1000':'2000',readAt:side==='selected'?'2026-11-06T01:02:03.123456789Z':'2026-11-06T01:02:04.987654321Z',basis:(side==='selected'?'a':'b').repeat(64),reads:{count:{readAt:'2026-11-06T01:02:03.123456789Z',basis:'c'.repeat(64)},groups:{readAt:'2026-11-06T01:02:04.987654321Z',basis:'d'.repeat(64)}}});
  const model={declaration:create(q,field,'Large facet benchmark'),values:state('selected'),...(paired?{baseline:state('previous'),baselineLabel:'Previous window'}:{})},receiptBefore=JSON.stringify([model.values.reads,model.baseline?.reads]);
  const host=document.createElement('section');host.dataset.largeFacetBenchmark='';host.setAttribute('aria-label','Supplied large-facet benchmark');host.style.cssText='position:fixed;inset:8px;overflow:auto;z-index:10000;background:#101621;padding:8px';const element=document.createElement('goo-categorical-facet'),prototype=customElements.get('goo-facet-list').prototype,originalRefresh=prototype.refresh;let refreshCalls=0;
  prototype.refresh=function(...args){refreshCalls++;return originalRefresh.apply(this,args);};
  const paint=()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))),observations=[];
  try{
   document.body.append(host);element.interaction={current:()=>true};const started=performance.now();element.data=model;host.append(element);const mountSyncMs=performance.now()-started;await paint();const mountPaintMs=performance.now()-started,mountRefreshCalls=refreshCalls,domElements=element.querySelectorAll('*').length,rowCount=element.querySelectorAll('[data-facet-item]').length;
   const search=element.querySelector('[aria-label="Search loaded facet values"]');if(!search)throw Error('Supplied facet did not mount its search.');
   const queries=['terminal-selected-0997',...(paired?['terminal-previous-0997']:[]),'not-an-observed-category',''];
   for(let iteration=0;iteration<iterations;iteration++)for(const query of queries){
    search.focus({preventScroll:true});search.value=query;search.setSelectionRange(query.length,query.length);const before=refreshCalls,start=performance.now();search.dispatchEvent(new InputEvent('input',{bubbles:true,inputType:'insertText',data:query}));const syncMs=performance.now()-start;await paint();
    observations.push({iteration,query,syncMs,paintMs:performance.now()-start,refreshPasses:refreshCalls-before,matches:element.querySelector('[aria-label="Loaded facet matches"]').textContent,visible:element.querySelectorAll('[data-facet-item]:not([hidden])').length,focused:document.activeElement===search,caret:search.selectionStart});
   }
   const target=label('selected',997),selected=element.querySelector('[data-facet-value-key="'+CSS.escape('1'+target)+'"]'),baseline=paired?element.querySelector('[data-facet-value-key="'+CSS.escape('1'+label('previous',997))+'"]'):null;
   const supplied=element.data;
   return {kind:'supplied-component',size,labelBytes,paired,iterations,styleFingerprint:{display:getComputedStyle(element).display,fontSize:getComputedStyle(element).fontSize,rowDisplay:getComputedStyle(element.querySelector('[data-facet-item]')).display},viewport:{width:innerWidth,height:innerHeight},componentWidth:element.getBoundingClientRect().width,mountSyncMs,mountPaintMs,mountRefreshCalls,domElements,rowCount,observations,exactSuffixIdentity:!!selected&&selected.dataset.facetText===JSON.stringify(target),missingDistinctFromEmpty:!!element.querySelector('[data-facet-value-key="0"]')&&!!element.querySelector('[data-facet-value-key="1"]'),selectedCounts:paired&&selected?[...selected.querySelectorAll('[data-count]')].map(cell=>cell.dataset.count):null,previousOnlyCounts:baseline?[...baseline.querySelectorAll('[data-count]')].map(cell=>cell.dataset.count):null,receiptStable:receiptBefore===JSON.stringify([supplied.values.reads,supplied.baseline?.reads]),syntheticInput:true};
  }finally{prototype.refresh=originalRefresh;host.remove();}
 },{size,labelBytes,iterations});
}
module.exports={runLargeFacetBenchmark};
