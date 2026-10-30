// Isolated public summary/comparison conformance. No product runtime or source IO.
const {chromium}=require('playwright');
const fs=require('node:fs/promises'),path=require('node:path'),assert=require('node:assert/strict');
const mean={key:'numeric.mean',label:'Numeric mean',type:'rational',nullable:true,unit:'USD',aggregation:{family:'numeric',kind:'mean',method:'exact-rational'}};
const latency={key:'duration.p95',label:'p95 duration',type:'rational',nullable:true,unit:'ns',aggregation:{family:'duration',kind:'quantile',method:'quantilesExactHigh',quantile:'0.95'}};
function snapshot(id,value,duration,capturedAt){return {
 version:'goo.dataset/1',id,revision:'r-'+id,origin:'source-read',capturedAt,
 shape:'summary',scope:{id:'scope-'+id,generation:1,signal:'spans',selector:{signal:'spans',environment:'test',summary:'span-duration',fromMs:id==='baseline'?0:10000,toMs:id==='baseline'?10000:20000,numeric:{field:{kind:'attribute',key:'price'},type:'int64',parse:'canonical-decimal',unit:'USD'}}},
 fields:[{key:'count',label:'Captured records',type:'number',nullable:true,unit:'records'},mean,latency],
 rows:[{key:'population',values:{count:'3','numeric.mean':value,'duration.p95':duration},cells:{'numeric.mean':{state:'derived',support:{total:'3',valid:'3',missing:'0',empty:'0',invalid:'0'}},'duration.p95':{state:'derived',support:{total:'3',valid:'3',missing:'0',empty:'0',invalid:'0'}}}}],
 quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'r-'+id,scopeId:'scope-'+id,aspects:['summary']}
}}
const baseline=snapshot('baseline','-1/3','4000000','2026-10-30T03:00:00.123456789Z'),selection=snapshot('selection','5/3','1000000','2026-10-30T04:00:00.987654321Z');
const comparison={version:'goo.comparison/1',state:'observed',baseline,selection,changes:{count:{state:'derived',value:'0',unit:'records'},'numeric.mean':{state:'derived',value:'2',unit:'USD',aggregation:mean.aggregation},'duration.p95':{state:'derived',value:'-3000000',unit:'ns',aggregation:latency.aggregation}}};
(async()=>{let browser;try{
 browser=await chromium.launch({executablePath:process.env.CHROMIUM_PATH||'/usr/bin/chromium',args:['--no-sandbox']});
 const page=await browser.newPage({viewport:{width:390,height:844},hasTouch:true}),requests=[],errors=[];
 page.on('request',request=>requests.push(new URL(request.url()).pathname));page.on('pageerror',error=>errors.push(error.message));
 await page.route('http://summary.invalid/**',async route=>{const name=new URL(route.request().url()).pathname;
  if(name==='/'){await route.fulfill({contentType:'text/html',body:`<!doctype html><html lang="en"><meta name="viewport" content="width=device-width"><title>Summary conformance</title><link rel="stylesheet" href="/components.css"><link rel="stylesheet" href="/chart-frame.css"><link rel="stylesheet" href="/measure-summary.css"><goo-measure-summary id="pre"></goo-measure-summary><script>window.supplied=${JSON.stringify(baseline)};window.comparison=${JSON.stringify(comparison)};document.getElementById('pre').snapshot=window.supplied;</script><script type="module" src="/measure-summary.mjs"></script><script type="module" src="/scope-view.mjs"></script><div id="root"></div>`});return}
  const local=path.resolve(__dirname,'../ui'+name);assert(local.startsWith(path.resolve(__dirname,'../ui')+path.sep));await route.fulfill({contentType:name.endsWith('.css')?'text/css':'text/javascript',body:await fs.readFile(local)});
 });
 await page.goto('http://summary.invalid/');await page.waitForFunction(()=>customElements.get('goo-measure-comparison')&&customElements.get('goo-chart'));
 await page.evaluate(async()=>{const {createSourceAdapter,asSnapshot,asComparison,hasGuarantee}=await import('/snapshots.mjs');const source=createSourceAdapter({id:'summary-conformance'});window.adopt=value=>source.adoptSnapshot(value,{question:value.scope.selector});window.asSnapshot=asSnapshot;window.asComparison=asComparison;window.hasGuarantee=hasGuarantee});
 const result=await page.evaluate(()=>{
  const root=document.querySelector('#root'),summary=document.createElement('goo-measure-summary');summary.id='summary';const source=window.adopt(window.supplied);summary.snapshot=source;root.append(summary);const shared=summary.snapshot===source&&summary.snapshot===summary.snapshot&&Object.isFrozen(source.rows[0].values);
  const pre=!!document.querySelector('#pre .measure-stats'),duration=summary.textContent.includes('4 ms'),rational=summary.textContent.includes('≈ −0.333 USD');
  const isolated=structuredClone(window.supplied);summary.snapshot=isolated;isolated.rows[0].values.count='999';const assigned=summary.snapshot.rows[0].values.count==='3';const read=summary.snapshot;try{read.rows[0].values.count='123'}catch{}const getter=summary.snapshot===read&&summary.snapshot.rows[0].values.count==='3'&&Object.isFrozen(read.rows[0].values);
  const ssr=document.createElement('goo-measure-summary');ssr.setAttribute('snapshot',JSON.stringify(window.supplied));root.append(ssr);const attrs=!!ssr.querySelector('.measure-stats')&&ssr.snapshot.id.startsWith('goo.local/')&&ssr.snapshot===ssr.snapshot&&ssr.snapshot.origin==='browser-held'&&!window.hasGuarantee(ssr.snapshot,'summary');
  const cached=ssr.snapshot;ssr.setAttribute('display',JSON.stringify({title:'Imported summary'}));const cachedAttribute=ssr.snapshot===cached;
  const compare=document.createElement('goo-measure-comparison');compare.id='compare';const selection=window.adopt(window.comparison.selection),envelope=window.asComparison({...window.comparison,baseline:source,selection});compare.comparison=envelope;root.append(compare);const comparisonShared=compare.comparison===envelope&&compare.comparison.baseline===source&&compare.comparison.selection===selection;
  const importedComparison=document.createElement('goo-measure-comparison');importedComparison.setAttribute('comparison',JSON.stringify(envelope));root.append(importedComparison);const heldComparison=importedComparison.comparison===importedComparison.comparison&&importedComparison.comparison.baseline.origin==='browser-held'&&!window.hasGuarantee(importedComparison.comparison.baseline,'summary')&&!window.hasGuarantee(importedComparison.comparison.selection,'summary');
  const tracks=[...compare.querySelectorAll('tr[data-field="numeric.mean"] .measure-comparison-track')];
  const common=tracks.length===2&&tracks[0].dataset.min===tracks[1].dataset.min&&tracks[0].dataset.max===tracks[1].dataset.max&&tracks[0].style.getPropertyValue('--measure-zero')===tracks[1].style.getPropertyValue('--measure-zero');
  const signed=tracks[0]?.dataset.min==='-1/3'&&tracks[0]?.dataset.max==='5/3';
  const delta=compare.querySelector('tr[data-field="duration.p95"] td[data-side=change]').textContent.includes('−3 ms');
  const clocks=compare.textContent.includes('03:00:00.123 UTC')&&compare.textContent.includes('04:00:00.987 UTC')&&compare.textContent.includes('1970-01-01 00:00:00.000 UTC')&&compare.textContent.includes('1970-01-01 00:00:20.000 UTC');
  const generic=document.createElement('goo-chart');generic.snapshot=window.supplied;generic.display={kind:'table'};root.append(generic);const passiveTable=!!generic.querySelector('table')&&!generic.querySelector('table button')&&![...generic.querySelectorAll('th')].some(th=>th.textContent==='Selection');
  return {pre,duration,rational,assigned,getter,attrs,shared,comparisonShared,cachedAttribute,heldComparison,common,signed,delta,clocks,passiveTable,overflow:document.documentElement.scrollWidth>innerWidth};
 });
 for(const [key,value] of Object.entries(result))assert.equal(value,key==='overflow'?false:true,key);
 const exact=page.locator('#summary [data-field="numeric.mean"] .measure-exact');await exact.locator('summary').focus();await page.keyboard.press('Enter');assert(await exact.locator('p').isVisible());assert((await exact.textContent()).includes('-1/3 USD'));
 const scope=page.locator('#compare .measure-provenance').first();await scope.locator('summary').tap();assert((await scope.textContent()).includes('2026-10-30T03:00:00.123456789Z'));assert(await scope.locator('pre').isVisible());
 await page.locator('#compare').evaluate(element=>element.style.width='300px');
 const baselineExact=page.locator('#compare tr[data-field="numeric.mean"] td[data-side=baseline] .measure-exact');await baselineExact.locator('summary').focus();await page.keyboard.press('Enter');assert(await baselineExact.locator('p').isVisible());
 const fit=await page.locator('#compare tr[data-field="numeric.mean"]').evaluate(row=>{
  const cells=[...row.querySelectorAll('td')].map(cell=>cell.getBoundingClientRect()),tracks=[...row.querySelectorAll('.measure-comparison-track')].map(track=>track.getBoundingClientRect());
  return {allSides:cells.length===3&&cells.every(rect=>rect.width>0&&rect.left>=0&&rect.right<=innerWidth),paired:Math.abs(cells[0].top-cells[1].top)<1,sameScaleWidth:Math.abs(tracks[0].width-tracks[1].width)<1,noScroll:row.closest('goo-table').scrollWidth<=row.closest('goo-table').clientWidth};
 });
 for(const [key,value] of Object.entries(fit))assert(value,key);
 const accessible=await page.locator('#compare table').ariaSnapshot();assert(accessible.includes('columnheader "Baseline"')&&accessible.includes('columnheader "Selection"')&&accessible.includes('columnheader "Change"'));
 const withdrawal=await page.evaluate(()=>{
  const summary=document.querySelector('#summary'),partial=structuredClone(window.supplied);partial.origin='browser-held';partial.sourceBasis='original-claim';partial.quality.extent='unknown';delete partial.completeness;delete partial.capturedAt;partial.rows[0].values['numeric.mean']=null;partial.rows[0].cells['numeric.mean']={state:'no_valid_values',support:{total:'3',valid:'0',missing:'1',empty:'1',invalid:'1'}};summary.snapshot=partial;
  const unknown=summary.textContent.includes('browser-held')&&summary.textContent.includes('unknown captured summary')&&summary.textContent.includes('Capture time not supplied')&&summary.textContent.includes('No valid values')&&!summary.textContent.includes('03:00:00');
  const compare=document.querySelector('#compare'),diagnostic=structuredClone(window.comparison);diagnostic.state='diagnostic';for(const change of Object.values(diagnostic.changes)){change.state='unavailable';change.value=null;change.reason='Independent diagnostic only'}compare.comparison=diagnostic;
  const diagnosticOnly=!compare.querySelector('.measure-comparison-track')&&compare.textContent.includes('Unavailable')&&compare.textContent.includes('Independent diagnostic only');
  diagnostic.changes.count={state:'derived',value:'7',unit:'records'};compare.comparison=diagnostic;const invalid=!compare.querySelector('table')&&compare.textContent.toLowerCase().includes('unavailable');
  summary.snapshot={};const clear=!summary.querySelector('.measure-stats');
  summary.snapshot=window.supplied;let getterReads=0;const fields=[];Object.defineProperty(fields,'0',{enumerable:true,get(){getterReads++;return 'count'}});summary.display={fields};const optionsClosed=getterReads===0&&!summary.querySelector('.measure-stats');
  return {unknown,diagnosticOnly,invalid,clear,optionsClosed};
 });
 for(const [key,value] of Object.entries(withdrawal))assert(value,key);
 assert.equal(errors.length,0,errors.join('; '));assert(!requests.some(name=>name.startsWith('/api/')||name==='/app.mjs'));
 console.log('Summary components: rational/duration/exact support, immutable SSR/preupgrade properties, shared signed comparison scale, independent clocks/provenance, keyboard/touch,390px, unavailable/stale clearing and no application IO pass');
}finally{await browser?.close()}})().catch(error=>{console.error(error);process.exitCode=1});
