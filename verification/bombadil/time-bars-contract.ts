// Browser-side public display contract, shared by source-free and live app actors.
import {extract} from '@antithesishq/bombadil/browser';
export const timeBarContract=extract(({document})=>{
 const plots=[...document.querySelectorAll<any>('goo-time-bars')];let bars=0,zeroBars=0,unknownBars=0;
 const valid=plots.every(plot=>{
  const nodes=[...(plot.querySelectorAll(':scope > .time-bars-figure > .time-bars > .time-bar') as NodeListOf<HTMLElement>)];bars+=nodes.length;
  if(plot.dataset.state==='unavailable')return nodes.length===0;
  const data=plot.data;if(!data||!Array.isArray(data.points)||nodes.length!==data.points.length||!/^\d+$/.test(data.ceiling))return false;
  const ceiling=BigInt(data.ceiling);
  return nodes.every((bar,i)=>{
   const p=data.points[i],state=p.state||'observed',known=state==='observed'||state==='no_data';
   if(bar.dataset.key!==p.key||bar.dataset.state!==state||bar.getAttribute('aria-label')!==p.label)return false;
   if(bar.tagName==='BUTTON'&&bar.getAttribute('aria-pressed')!==String(plot.selectedKey===p.key))return false;
   if(bar.tagName==='A'&&(bar.getAttribute('href')!==p.href||bar.hasAttribute('aria-current')!==(plot.selectedKey===p.key)))return false;
   const pseudo=getComputedStyle(bar,'::after');
   if(!known){unknownBars++;return !bar.hasAttribute('data-count')&&bar.dataset.positive==='false'&&pseudo.backgroundImage.includes('gradient')}
   if(typeof p.value!=='string'||!/^\d+$/.test(p.value)||bar.dataset.count!==p.value)return false;
   const count=BigInt(p.value);if(count>ceiling||state==='no_data'&&count!==0n)return false;
   const height=parseFloat(bar.style.getPropertyValue('--bar-height')),expected=ceiling===0n?0:Number(count*10000n/ceiling)/100;
   if(!Number.isFinite(height)||height<0||height>100||height!==expected||bar.dataset.positive!==String(count>0n))return false;
   if(count===0n){zeroBars++;return parseFloat(pseudo.height)===0&&parseFloat(pseudo.minHeight)===0}
   return parseFloat(pseudo.minHeight)===2;
  });
 });return {valid,plots:plots.length,bars,zeroBars,unknownBars};
});
