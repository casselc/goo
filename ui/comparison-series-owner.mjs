import './time-range.mjs';
/* Application adapters: retained aggregates only, explicit stored reads and staged intent. */
import * as ComparisonSeries from './comparison-series.mjs';
import * as Workspace from './workspace.mjs';
const global=globalThis;
const n=(tag,text)=>{const e=document.createElement(tag);if(text!==undefined)e.textContent=text;return e};
function side(data,label){const q=data?.question;if(!q||!data.timeSeries?.length)return null;return {label,fromMs:q.fromMs,toMs:q.toMs,points:data.timeSeries.map(p=>({...p,state:data.state==='changed_between_reads'?'unknown':p.state}))}}
function create(host,initial,{onSelect,canSelect=()=>true,fixedDomain=false}={}){
  let results=structuredClone(initial),chosen=null,ceiling=null,overflow=false;
  const section=n('section'),heading=n('h5','Record activity comparison'),label=n('label','Time alignment '),select=n('select'),chart=n('goo-comparison-series');section.className='comparison-series-owner';select.setAttribute('aria-label','Comparison time alignment');
  for(const [value,text]of[['elapsed','Elapsed from each start'],['utc','Absolute UTC']]){const o=n('option',text);o.value=value;select.append(o)}label.append(select);select.onchange=()=>chart.setAttribute('alignment',select.value);
  const range=n('goo-time-range'),selection=n('div'),selectionLabel=n('p'),action=onSelect?n('button','Stage selected interval'):n('a','Stage selected interval');selection.hidden=true;
  function choose(index,fromMs,toMs){const r=results[index];if(!r||fromMs<r.question.fromMs||toMs>r.question.toMs||fromMs>=toMs)return;chosen={...structuredClone(r.question),fromMs,toMs,offset:0};selectionLabel.textContent=(index?'Comparison':'Baseline')+' · selected UTC range'+(fixedDomain?' · pinned; Run reads fresh data':'');selectionLabel.title='Selected source read: '+(r.readAt||r.capturedAt||'unavailable');range.data={fromMs:r.question.fromMs,toMs:r.question.toMs,selection:{fromMs,toMs}};selection.hidden=false;if(!onSelect)action.href=Workspace.questionLink(chosen,'Records')}
  chart.addEventListener('goo-series-select',e=>{const {side,index}=e.detail,p=results[side]?.timeSeries?.[index];if(p)choose(side,p.fromMs,p.toMs)});
  chart.addEventListener('goo-series-range',e=>choose(e.detail.side,e.detail.fromMs,e.detail.toMs));
  range.addEventListener('goo-range-select',e=>{chosen={...chosen,...e.detail};if(!onSelect)action.href=Workspace.questionLink(chosen,'Records')});
  if(onSelect){action.disabled=!canSelect();action.onclick=()=>chosen&&canSelect()&&onSelect(structuredClone(chosen),action)}
  selection.append(selectionLabel,range,action);section.append(heading,label,chart,selection);host.append(section);
  function update(next){results=structuredClone(next);const sides=results.map((r,i)=>side(r,i?'Comparison':'Baseline'));if(sides.some(s=>!s)){chart.data=null;overflow=false;return false}const m=ComparisonSeries.model({sides});if(fixedDomain&&ceiling===null&&m.state==='observed')ceiling=m.ceiling;const data={sides,...(fixedDomain&&ceiling!==null?{ceiling}:{})};overflow=ComparisonSeries.model(data).state==='overflow';chart.data=data;section.dataset.ceiling=ceiling??m.ceiling??'';return true}
  update(results);return {element:section,update,get overflow(){return overflow},rescale(){ceiling=null;update(results)},dispose(){section.remove()}};
 }
function mount(host,results,options={}){if(results.some(r=>!side(r,'')))return null;return create(host,results,options).element}
function stored(host,ids,{current=()=>host.isConnected}={}){
  const section=n('section'),button=n('button','Show captured activity comparison'),status=n('p'),content=n('div');let generation=0,controller;section.append(button,status,content);host.append(section);status.setAttribute('role','status');button.onclick=async()=>{const g=++generation;controller?.abort();controller=new AbortController();button.disabled=true;status.textContent='Loading stored aggregates; no telemetry reread.';try{const pages=await Promise.all(ids.map(async id=>{const r=await fetch('/api/captures/'+id,{signal:controller.signal}),p=await r.json();if(!r.ok||p.receipt?.id!==id)throw Error('Stored series unavailable');return p}));if(g!==generation||!current()||!section.isConnected)return;content.replaceChildren();const view=mount(content,pages.map(p=>p.explore));status.textContent=view?'Retained aggregate intervals; independent reads, completeness unknown.':'These captures do not retain record time series. No source read was made.';button.hidden=!!view}catch(e){if(g===generation&&current()&&section.isConnected&&e.name!=='AbortError')status.textContent=e.message}finally{if(g===generation)button.disabled=false}};
  return {dispose(){generation++;controller?.abort();section.remove()}};
 }
export {mount,create,stored};
