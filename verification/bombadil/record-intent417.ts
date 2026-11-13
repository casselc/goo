// Served configuration, one explicit read, and rejection without losing evidence.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const o=(window as any).__recordIntent417,rows=[...document.querySelectorAll('#exploreRows .record-row')].map(n=>n.textContent),button=document.getElementById('exploreRun') as HTMLButtonElement|null,status=document.getElementById('exploreStatus')?.textContent||'',ownerPhase=document.getElementById('exploration')?.getAttribute('data-question-phase');
 return {version:'goo.recordIntent417Actor/1',phase:o?.phase||'initial',ready:!!button?.getClientRects().length&&!button.disabled&&status.includes('Fixed question staged.')&&rows.length===0,rows:rows.length,ownerPhase,status,posts:o?.calls.filter((c:any)=>c.method==='POST').length||0,forbidden:o?.forbidden||[],held:!o?.held||JSON.stringify(rows)===JSON.stringify(o.held),observed:o?.observed||null,hashSame:!!o&&window.location.hash===o.qhash,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const publicationIsExplicit=always(()=>state.current.forbidden.length===0);
export const oneExplicitRead=always(()=>state.current.posts<=1);
export const initialStagingIsLocal=always(()=>!['initial','ready'].includes(state.current.phase)||state.current.posts===0);
export const rejectionPreservesAppliedEvidence=always(()=>state.current.phase!=='refused'||state.current.posts===1&&state.current.rows===100&&state.current.held&&state.current.ownerPhase==='applied');
export const contained=always(()=>!state.current.overflow);
export const finishes=eventually(()=>state.current.phase==='complete'&&state.current.rows===0&&state.current.posts===1&&state.current.hashSame).within(30,'seconds');
const act=registerCustomAction('goo.recordIntent417',async(document,window,command:string)=>{
 const w=window as any,rows=()=>[...document.querySelectorAll('#exploreRows .record-row')].map(n=>n.textContent),status=()=>document.getElementById('exploreStatus')?.textContent||'',ownerPhase=()=>document.getElementById('exploration')?.getAttribute('data-question-phase');
 if(command==='init'){
  const expected=JSON.parse(new URL(window.location.href).searchParams.get('expected')||'null');if(!expected?.question||typeof expected.qhash!=='string'||typeof expected.invalidraw!=='string'||expected.count!=='140')throw Error('Expected one fixed ordinary-record bank');
  const Protocol=await import(new URL('/question-protocol.mjs',window.location.href).href),Session=await import(new URL('/question-session.mjs',window.location.href).href);Protocol.validateRecordQuestion(expected.question);if(Session.identity(Protocol.readQuestionLink(expected.qhash).question)!==Session.identity(expected.question))throw Error('Expected actor link changed the question');let rejected=false;try{Protocol.readQuestionLink(expected.invalidraw)}catch{rejected=true}if(!rejected)throw Error('Expected a refused raw question link');
  const o=w.__recordIntent417={...expected,phase:'ready',calls:[],forbidden:[],identity:Session.identity},original=window.fetch.bind(window);
  window.fetch=async(input:any,init:any)=>{
   const url=new URL(typeof input==='string'?input:input.url,window.location.href),method=String(init?.method||input?.method||'GET').toUpperCase(),call:any={path:url.pathname,method};if(init?.body)try{call.request=JSON.parse(init.body)}catch{}
   if(url.pathname.startsWith('/api/'))o.calls.push(call);if(!['GET','HEAD'].includes(method)&&!(method==='POST'&&url.pathname==='/api/explore'))o.forbidden.push(method+' '+url.pathname);
   const response=await original(input,init);call.status=response.status;
   if(method==='POST'&&url.pathname==='/api/explore'){
    const data=await response.clone().json();call.response=data;
    if(response.ok)o.observed={request:call.request,question:data.question,count:data.count,state:data.state,rowIds:(data.rows||[]).map((r:any)=>r.Attributes?.['work.id']),bodies:(data.rows||[]).map((r:any)=>r.Body)};
   }
   return response;
  };
  return;
 }
 const o=w.__recordIntent417,posts=()=>o.calls.filter((c:any)=>c.method==='POST').length;
 if(command==='run'){
  if(posts()!==0||rows().length!==0||!status().includes('Fixed question staged.'))throw Error('Initial actor state borrowed records or acquired telemetry');
  const button=document.getElementById('exploreRun') as HTMLButtonElement;if(!button?.getClientRects().length||button.disabled||button.closest('[hidden],[inert]'))throw Error('Explicit Run unavailable');button.scrollIntoView({block:'nearest',behavior:'instant'});button.focus();button.click();o.phase='reading';
 }
 if(command==='refuse'){
  const observed=o.observed,shown=rows(),domIds=shown.map(text=>text?.match(/ordinary work (\d+)/u)?.[1]);if(posts()!==1||observed?.count!==o.count||observed.state!=='observed'||observed.rowIds.length!==100||new Set(observed.rowIds).size!==100||observed.rowIds.some((id:any)=>typeof id!=='string'||!/^\d+$/.test(id)||Number(id)<0||Number(id)>=140)||o.identity(observed.request)!==o.identity(o.question)||o.identity(observed.question)!==o.identity(o.question)||shown.length!==100||JSON.stringify([...domIds].sort())!==JSON.stringify([...observed.rowIds].sort())||!observed.bodies.every((body:any)=>typeof body==='string'&&shown.some(text=>text?.includes(body)))||ownerPhase()!=='applied')throw Error('Explicit actor read did not yield the expected exact population');
  o.held=shown;window.location.hash=o.invalidraw;o.phase='refusing';
 }
 if(command==='observe-refusal'){
  if(posts()!==1||rows().length!==100||JSON.stringify(rows())!==JSON.stringify(o.held)||ownerPhase()!=='applied'||!status().includes('Cannot restore question'))throw Error('Invalid link changed the held evidence or source budget');o.phase='refused';
 }
 if(command==='restage'){
  if(posts()!==1||JSON.stringify(rows())!==JSON.stringify(o.held))throw Error('Refusal did not remain held');window.location.hash=o.qhash;o.phase='recovering';
 }
 if(command==='finish'){
  if(posts()!==1||rows().length!==0||!status().includes('Fixed question staged.')||window.location.hash!==o.qhash)throw Error('Valid restaging did not remain source-free');o.phase='complete';
 }
});
export const journey=actions(()=>{
 const s=state.current;if(s.phase==='initial')return s.ready?[act('init')]:['Wait'];if(s.phase==='ready')return[act('run')];if(s.phase==='reading')return s.observed&&s.rows===100&&s.ownerPhase==='applied'?[act('refuse')]:['Wait'];if(s.phase==='refusing')return s.status.includes('Cannot restore question')&&s.rows===100&&s.held&&s.ownerPhase==='applied'?[act('observe-refusal')]:['Wait'];if(s.phase==='refused')return[act('restage')];if(s.phase==='recovering')return s.ready&&s.hashSame?[act('finish')]:['Wait'];return['Wait'];
});
