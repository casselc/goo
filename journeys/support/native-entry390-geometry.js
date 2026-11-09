// Browser-side measured paint/clip facts. No scroll or application mutation.
function geometry390({rootSelector,state='data'}){
 const root=document.querySelector(rootSelector);
 const rect=e=>{if(!e)return null;const r=e.getBoundingClientRect();return{absoluteDocumentTop:r.top+scrollY,x:r.x,y:r.y,top:r.top,left:r.left,right:r.right,bottom:r.bottom,width:r.width,height:r.height}};
 const visible=e=>!!e&&e.getClientRects().length>0&&!e.closest('[hidden],[inert],[aria-hidden="true"]');
 const clipped=value=>/^(auto|scroll|hidden|clip)$/.test(value);
 function clipFacts(e){
  let clip={left:0,top:0,right:innerWidth,bottom:innerHeight};const ancestors=[];
  for(let p=e;p;p=p.parentElement){const css=getComputedStyle(p),r=p.getBoundingClientRect(),x=clipped(css.overflowX),y=clipped(css.overflowY);if(!x&&!y)continue;
   ancestors.push({tag:p.tagName,id:p.id,className:String(p.className),scrollTop:p.scrollTop,scrollLeft:p.scrollLeft,scrollHeight:p.scrollHeight,clientHeight:p.clientHeight,overflowX:css.overflowX,overflowY:css.overflowY,rect:rect(p)});
   if(x)clip={...clip,left:Math.max(clip.left,r.left),right:Math.min(clip.right,r.right)};
   if(y)clip={...clip,top:Math.max(clip.top,r.top),bottom:Math.min(clip.bottom,r.bottom)};
  }
  return{clip,clipAncestors:ancestors};
 }
 function item(e){
  const r=rect(e),facts=clipFacts(e),c=facts.clip;
  return{tag:e?.tagName||null,visible:visible(e),rect:r,...facts,intersection:r?{width:Math.max(0,Math.min(r.right,c.right)-Math.max(r.left,c.left)),height:Math.max(0,Math.min(r.bottom,c.bottom)-Math.max(r.top,c.top)),centerInside:(r.left+r.right)/2>=c.left&&(r.left+r.right)/2<=c.right&&(r.top+r.bottom)/2>=c.top&&(r.top+r.bottom)/2<=c.bottom}:null};
 }
 const candidate=state==='refused'?document.getElementById('measureStatus'):state==='empty'?root?.querySelector('.native-result-summary')||root:root?.querySelector('goo-value-timeline,goo-histogram,goo-chart,goo-table,table');
 const firstItem=state==='data'?(candidate?.querySelector('[data-key],tbody>tr,.histogram-bar,[data-bin]')||candidate):candidate;
 const timeline=root?.querySelector('goo-value-timeline'),numeric=timeline?.querySelector('.value-timeline-plot'),rail=timeline?.querySelector('.value-timeline-unavailable-plot');
 const marks=[...timeline?.querySelectorAll('.value-timeline-mark')||[]].map(e=>({...item(e),key:e.dataset.key,state:e.dataset.state,value:e.hasAttribute('data-value')?e.dataset.value:null,fromNS:e.dataset.fromNs,toNS:e.dataset.toNs}));
 const summary=root?.querySelector('.native-result-summary')||root?.querySelector(':scope>h3'),main=document.querySelector('main'),workspace=root?.closest('[data-workspace-panel]');
 const scroll=e=>e?{tag:e.tagName,id:e.id,scrollTop:e.scrollTop,scrollHeight:e.scrollHeight,clientHeight:e.clientHeight,scrollWidth:e.scrollWidth,clientWidth:e.clientWidth,rect:rect(e)}:null;
 const primary=item(candidate);
 return{viewport:{width:innerWidth,height:innerHeight},document:scroll(document.scrollingElement),main:scroll(main),workspace:scroll(workspace),root:{selector:rootSelector,visible:visible(root),rect:rect(root)},clip:primary.clip,ancestors:primary.clipAncestors,primary,firstItem:item(firstItem),numericPlot:rect(numeric),unavailablePlot:rect(rail),marks,timeline:timeline?.data||null,summary:{...item(summary),text:summary?.textContent||null,fromNS:summary?.dataset.fromNs||null,toNS:summary?.dataset.toNs||null},status:{text:document.getElementById('measureStatus')?.textContent||null},scrollX,scrollY};
}
module.exports={geometry390};
