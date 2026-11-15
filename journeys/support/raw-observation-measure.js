'use strict';
const fs=require('node:fs/promises'),path=require('node:path');

// Read the semantic clock, not its human display. The fallback supports the
// previously embedded binary during before/after measurements.
async function sourceReadAt(panel){return panel.locator('dt').filter({hasText:/^Source read UTC$/}).evaluate(e=>{const value=e.nextElementSibling;return value.querySelector('time[datetime]')?.dateTime||value.textContent})}
function framePath(target,dir){return process.env.GOO_RAW448_LABEL==='baseline'?path.join(dir,'baseline-'+path.basename(target)):target}
function sourceWork(report){return report.steps.map(step=>{const before=step.before,after=step.after;return {name:step.name||step.action,...(typeof step.sourceBefore==='number'?{reads:step.sourceAfter-step.sourceBefore}:before&&after?{reads:after.source-before.source}:{}),...(typeof step.sqlBefore==='number'?{sql:step.sqlAfter-step.sqlBefore}:before&&after?{sql:after.sql-before.sql}:{}),posts:step.posts||step.calls?.filter(call=>call.method==='POST').map(call=>call.path)||[]}})}
function finishMeasurements(report){if(report.raw448)report.raw448.sourceWork=sourceWork(report)}
async function measureRaw(page,{family,dir,report,counters,heldRows,state='held inspection'}){
 const label=process.env.GOO_RAW448_LABEL;if(!label)return;
 const original=page.viewportSize(),before=counters(),result={version:'goo.raw-observation-measure448/1',label,family,state,heldRows,density:[]};
 for(const width of [1440,390,320]){
  await page.setViewportSize({width,height:1000});await page.locator('[data-raw-observations]').evaluate(e=>e.scrollIntoView({block:'start'}));
  await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
  const value=await page.evaluate(()=>{
   const panel=document.querySelector('[data-raw-observations]'),rect=e=>{if(!e)return null;const r=e.getBoundingClientRect();return {top:r.top,left:r.left,width:r.width,height:r.height,bottom:r.bottom,right:r.right}},visible=e=>!!e?.getClientRects().length&&!e.closest('[hidden],[aria-hidden="true"],[inert]')&&(()=>{for(let p=e.parentElement;p;p=p.parentElement)if(p.localName==='details'&&!p.open&&!p.querySelector(':scope>summary')?.contains(e))return false;return true})();
   const rows=[...panel.querySelectorAll('tr[data-occurrence-key]')].filter(visible),first=rows[0],inspector=panel.querySelector('[aria-label="Selected item details"]'),body=inspector?.querySelector('.ui-inspector-body'),receipt=panel.querySelector('[data-raw-read-context]')||panel.children[3],bucket=panel.querySelector('table[aria-label="Source histogram buckets"]')||panel.querySelector('[aria-label="Source histogram buckets"] table');
   const p=rect(panel),r=rect(first),i=visible(inspector)?rect(inspector):null,b=rect(bucket?.querySelector('tbody>tr'));
   const inViewport=e=>{if(!visible(e))return false;const r=e.getBoundingClientRect();let top=Math.max(0,r.top),bottom=Math.min(innerHeight,r.bottom),left=Math.max(0,r.left),right=Math.min(innerWidth,r.right);for(let parent=e.parentElement;parent;parent=parent.parentElement){const style=getComputedStyle(parent),box=parent.getBoundingClientRect();if(/auto|scroll|hidden|clip/.test(style.overflowY)){top=Math.max(top,box.top);bottom=Math.min(bottom,box.bottom)}if(/auto|scroll|hidden|clip/.test(style.overflowX)){left=Math.max(left,box.left);right=Math.min(right,box.right)}}return bottom>top&&right>left};
   const controls=[...panel.querySelectorAll('button,input,select,summary,a,[tabindex]')].filter(visible),control=e=>({tag:e.localName,label:(e.getAttribute('aria-label')||e.textContent||'').slice(0,120),disabled:!!e.disabled,rect:rect(e)});
   return {width:innerWidth,height:innerHeight,documentOverflow:document.documentElement.scrollWidth>innerWidth,panel:p,receipt:rect(receipt),readContextCharacters:receipt?.textContent.length||0,firstRecord:r,firstRecordBelowPanel:r?r.top-p.top:null,inspector:i,inspectorBody:rect(body),inspectorFacts:rect(body?.querySelector('goo-stat-list')),firstBucket:b,firstBucketBelowInspector:b&&i?b.top-i.top:null,displayedRows:rows.length,visibleRows:rows.filter(inViewport).length,keys:rows.map(e=>e.dataset.occurrenceKey),controls:controls.map(control),viewportControls:controls.filter(inViewport).map(control),clocks:[...panel.querySelectorAll('time')].filter(visible).map(e=>({text:e.textContent,exact:e.dateTime,title:e.title}))};
  });
  if(value.documentOverflow)throw Error('Raw '+family+' measurement overflows at '+width);
  result.density.push(value);await page.screenshot({path:path.join(dir,'448-'+label+'-'+family+'-'+width+'.png')});
  if(process.env.GOO_RAW448_SCREENSHOTS==='1'&&label!=='baseline')await page.screenshot({path:'docs/journeys/screenshots/448-'+family+'-'+width+'.png'});
 }
 await page.setViewportSize(original);const after=counters();if(JSON.stringify(after)!==JSON.stringify(before))throw Error('Raw '+family+' density measurement acquired source work');result.measurementWork={before,after};report.raw448=result;finishMeasurements(report);await fs.writeFile(path.join(dir,'raw448-measure.json'),JSON.stringify(result,null,2));
}
module.exports={sourceReadAt,framePath,measureRaw,finishMeasurements};
