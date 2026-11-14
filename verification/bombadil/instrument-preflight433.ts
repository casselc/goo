import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const o=(window as any).__preflight433,rows=[...document.querySelectorAll<HTMLTableRowElement>('#metricInstruments tr[data-instrument]')],blocked=rows.filter(r=>r.dataset.instrument!=='valid.gauge'),valid=rows.find(r=>r.dataset.instrument==='valid.gauge'),build=(r:any)=>[...r?.querySelectorAll('button')||[]].find((e:any)=>e.textContent==='Build measurement package');
 return {version:'goo.instrumentPreflight433Actor/1',phase:o?.phase||'initial',ready:document.documentElement.dataset.profile==='generic'&&!!document.querySelector('#metricDiscovery'),rows:rows.length,blocked:blocked.length===3&&blocked.every(r=>build(r)?.disabled&&/instance/i.test(r.textContent||'')),valid:!!build(valid)&&!build(valid).disabled,planned:document.querySelector('#measureStatus')?.textContent?.includes('Question validated')||false,safe:!o||o.forbidden.length===0&&o.discoveries<=1&&o.plans<=1&&o.runs===0,discoveries:o?.discoveries||0,plans:o?.plans||0,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const finishes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const explicitReadAndPlan=always(()=>state.current.safe);
export const contained=always(()=>!state.current.overflow);
const act=registerCustomAction('goo.instrumentPreflight433',async(document,window,command:string)=>{
 const w=window as any,fixture=JSON.parse(new URL(window.location.href).searchParams.get('preflight433')||'{}'),click=(e:any)=>{if(!e||e.disabled)throw Error('Expected admitted native control');e.scrollIntoView({block:'center'});e.focus();e.click()};
 if(command==='open'){
  if(!fixture.environment||!fixture.service)throw Error('Explicit fixture scope required');
  const owner=w.__preflight433={phase:'opening',discoveries:0,plans:0,runs:0,forbidden:[]},original=window.fetch.bind(window);
  window.fetch=async(input:any,init:any)=>{const p=new URL(typeof input==='string'?input:input.url,window.location.href).pathname,method=String(init?.method||input?.method||'GET').toUpperCase();if(p.startsWith('/api/')&&method==='POST'){if(p==='/api/metric-discovery')owner.discoveries++;else if(p==='/api/measurement-packages/plan')owner.plans++;else if(p==='/api/measurements')owner.runs++;else if(!/^\/api\/queries\/[^/]+\/cancel$/.test(p))owner.forbidden.push(p)}return original(input,init)};
  click(document.querySelector('#exploreMeasurementsMode'));const d=document.querySelector<HTMLDetailsElement>('#metricDiscovery')!;if(!d.open)click(d.querySelector(':scope>summary'));
  for(const [id,value] of [['metricEnvironment',fixture.environment],['metricService',fixture.service]]){const e=document.getElementById(id) as HTMLInputElement;e.value=value;e.dispatchEvent(new Event('input',{bubbles:true}));e.dispatchEvent(new Event('change',{bubbles:true}))}return;
 }
 const owner=w.__preflight433;
 if(command==='discover'){click(document.querySelector('#metricDiscover'));owner.phase='discovering'}
 if(command==='refuse'){document.querySelectorAll<HTMLButtonElement>('#metricInstruments tr[data-instrument]:not([data-instrument="valid.gauge"]) button').forEach(e=>{if(e.textContent==='Build measurement package')e.click()});owner.phase='refused'}
 if(command==='build'){click([...document.querySelectorAll('#metricInstruments tr[data-instrument="valid.gauge"] button')].find(e=>e.textContent==='Build measurement package'));owner.phase='building'}
 if(command==='finish')owner.phase='complete';
});
export const journey=actions(()=>{const s=state.current;if(s.phase==='initial')return s.ready?[act('open')]:['Wait'];if(s.phase==='opening')return[act('discover')];if(s.phase==='discovering')return s.rows===4&&s.blocked&&s.valid&&s.discoveries===1?[act('refuse')]:['Wait'];if(s.phase==='refused')return s.plans===0&&s.safe?[act('build')]:['Wait'];if(s.phase==='building')return s.planned&&s.plans===1&&s.safe?[act('finish')]:['Wait'];return['Wait']});
