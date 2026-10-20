// App-owned live stream exploration. No saves or workload actions.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{
 const tail=document.querySelector<HTMLElement>('.live-tail'),rows=[...document.querySelectorAll<HTMLElement>('.live-tail tbody>tr')],keys=rows.map(r=>r.dataset.tailKey),stage=document.documentElement.dataset.tailStage||'initial',golden=document.documentElement.dataset.tailPausedKeys;
 return {version:'goo.tail.v1',stage,ready:!!document.querySelector('#exploreTail'),running:document.querySelector('#exploration')?.getAttribute('data-question-phase')==='running',applied:!document.querySelector<HTMLButtonElement>('#exploreTail')?.disabled,visible:!!tail&&!tail.hidden,phase:tail?.dataset.state,rows:rows.length,gap:tail?.dataset.gap==='true',bounds:rows.length<=100&&new Set(keys).size===keys.length&&Number(tail?.dataset.retained||0)===rows.length&&Number(tail?.dataset.observed||0)===Number(tail?.dataset.retained||0)+Number(tail?.dataset.evicted||0),paused:stage!=='paused'||!golden||golden===JSON.stringify(keys),overflow:document.documentElement.scrollWidth>innerWidth};
});
export const ready=eventually(()=>state.current.ready).within(20,'seconds');
export const complete=eventually(()=>state.current.stage==='complete').within(25,'seconds');
export const boundedObservations=always(()=>state.current.bounds&&!state.current.overflow&&state.current.paused);
const act=registerCustomAction('goo.tailControl',async(document,_window,command:string,next?:string)=>{
 const button=(name:string)=>[...document.querySelectorAll<HTMLButtonElement>('.live-tail button')].find(b=>b.textContent===name);
 if(command==='run')(document.querySelector('#exploreRun')as HTMLButtonElement)?.click();
 if(command==='start')(document.querySelector('#exploreTail')as HTMLButtonElement)?.click();
 if(command==='pause'){button('Pause stream')?.click();document.documentElement.dataset.tailPausedKeys=JSON.stringify([...document.querySelectorAll<HTMLElement>('.live-tail tbody>tr')].map(r=>r.dataset.tailKey))}
 if(command==='resume')button('Resume with recent logs')?.click();
 if(command==='focus')(document.querySelector('.live-tail .record-attributes>summary')as HTMLElement)?.focus();
 if(command==='close')button('Close live logs')?.click();
 if(next)document.documentElement.dataset.tailStage=next;
});
export const explore=actions(()=>{const s=state.current;if(!s.ready||s.running)return['Wait'];if(s.stage==='initial')return s.applied?[act('start','connecting')]:[act('run')];if(s.stage==='connecting')return s.rows>0&&s.phase!=='loading'?[act('pause','paused')]:['Wait'];if(s.stage==='paused')return[act('resume','resuming')];if(s.stage==='resuming')return s.gap&&s.rows>0&&s.phase!=='loading'?[act('focus','inspecting')]:['Wait'];if(s.stage==='inspecting')return s.phase==='paused'?[act('close','complete')]:['Wait'];return['Wait']});
