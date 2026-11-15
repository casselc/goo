'use strict';
const path=require('node:path');
// Passive geometry and authority measurements; the existing journey owns IO.
function predicateFrame(dir,name){const label=process.env.GOO_PREDICATE451_LABEL;if(label==='baseline')return path.join(dir,'baseline-'+name);return 'docs/journeys/screenshots/'+(label==='final'?name.replace(/^450-/,'451-'):name)}
async function measurePredicateDraft(page,{dir,report,width,posts,sql}){
 const label=process.env.GOO_PREDICATE451_LABEL;if(!label)return;
 const before={posts:posts(),sql:sql()},measurement=await page.evaluate(()=>{
  const pair=document.querySelector('.paired-activity'),visible=e=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert],[aria-hidden="true"]'),r=e=>{if(!e)return null;const v=e.getBoundingClientRect();return {top:v.top,left:v.left,width:v.width,height:v.height}};
  return {domNodes:document.querySelectorAll('*').length,pair:r(pair),visibleControls:[...pair.querySelectorAll('button,input,select,textarea,summary')].filter(visible).length,lanes:Object.fromEntries(['logs','spans'].map(signal=>{const lane=pair.querySelector('[data-paired-lane="'+signal+'"]'),condition=lane.querySelector('.paired-lane-condition');return[signal,{firstRow:r(lane.querySelector('tbody>tr[data-tail-key]')),rows:lane.querySelectorAll('tbody>tr[data-tail-key]').length,condition:{text:condition?.textContent||'',title:condition?.title||''},openEditors:lane.querySelectorAll('[data-lane-filter][open]').length}]}))};
 });
 const after={posts:posts(),sql:sql()};if(before.posts!==after.posts||before.sql!==after.sql)throw Error('Passive layout measurement acquired source authority');
 report.predicate451??={version:'goo.predicateDraftMeasure451/1',label,screens:[]};report.predicate451.screens.push({width,...measurement,sourceBefore:before,sourceAfter:after});
 return measurement;
}
module.exports={predicateFrame,measurePredicateDraft};
