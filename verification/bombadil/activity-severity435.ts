// Numeric bands are supplied aggregate projections, independent of producer text.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const owner=(window as any).__severity435,params=new URL(window.location.href).searchParams,shell=document.querySelector('.stream-entry'),tail=shell?.querySelector<HTMLElement>('.live-tail'),volume=shell?.querySelector<HTMLElement>('.activity-log-context');
 const visible=(e:any)=>{if(!e?.isConnected||!e.getClientRects().length||e.closest('[hidden],[inert],[aria-hidden="true"]'))return false;for(let p=e.parentElement;p;p=p.parentElement)if(p.localName==='details'&&!p.open&&!p.querySelector(':scope>summary')?.contains(e))return false;return true;},button=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((e:any)=>e.textContent===label&&visible(e)&&!e.disabled),calls=owner?.calls||[],sources=calls.length,chooser=volume?.querySelector<HTMLSelectElement>('select[aria-label="Numeric severity"]');
 return {version:'goo.activitySeverity435Actor/1',phase:owner?.phase||'initial',ready:document.documentElement.dataset.profile==='generic'&&!!params.get('severity435Env')&&!!params.get('severity435Bank')&&!!button(shell,'Start stream'),loaded:Number(tail?.dataset.retained)===8&&['unknown','partial'].includes(tail?.dataset.state||'')&&!!button(shell,'Pause stream'),paused:tail?.dataset.state==='paused'&&!!button(shell,'Show volume'),volume:visible(volume)&&volume?.dataset.phase==='ready'&&!!owner?.aggregate&&!!chooser&&!chooser.disabled,band:visible(volume)&&volume?.dataset.severityBand==='error'&&chooser?.value==='error'&&!!button(volume,'Filter Error'),staged:!!button(shell,'Apply scope')&&tail?.dataset.state==='paused',applied:owner?.tailApplied===true&&Number(tail?.dataset.retained)===1,tailReads:calls.filter((c:any)=>c.path==='/api/tail').length,aggregateReads:calls.filter((c:any)=>c.path==='/api/record-aggregates').length,localSafe:!owner||owner.localAt===undefined||owner.localAt===sources,scopeSafe:owner?.scopeSafe!==false,forbidden:owner?.forbidden||[],overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const explicitReadBoundaries=always(()=>state.current.localSafe&&state.current.tailReads<=2&&state.current.aggregateReads<=1&&state.current.forbidden.length===0);
export const exactNumericScope=always(()=>state.current.scopeSafe);
export const containment=always(()=>!state.current.overflow);
const act=registerCustomAction('goo.activitySeverity435',async(document,window,command:string)=>{
 const w=window as any,params=new URL(window.location.href).searchParams,environment=params.get('severity435Env'),bank=params.get('severity435Bank'),shell=document.querySelector('.stream-entry')!,tail=shell.querySelector<HTMLElement>('.live-tail')!,volume=shell.querySelector<HTMLElement>('.activity-log-context')!;
 const visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert]'),button=(root:any,label:string)=>[...root?.querySelectorAll('button')||[]].find((e:any)=>e.textContent===label&&visible(e)&&!e.disabled),click=(e:any)=>{if(!e||!visible(e)||e.disabled)throw Error('Expected numeric severity control unavailable');e.scrollIntoView({block:'center',behavior:'instant'});e.focus({preventScroll:true});e.click();};
 if(command==='start'){
  if(!environment||!bank)throw Error('Severity fixture identity required');
  const owner=w.__severity435={phase:'reading',calls:[],forbidden:[],localAt:undefined,scopeSafe:true,tailApplied:false,aggregate:null},original=window.fetch.bind(window),parent={any:[{field:{kind:'resource',key:'severity435.bank'},op:'eq',value:bank},{field:{kind:'column',key:'service'},op:'eq',value:'never-'+bank}]},range={all:[{field:{kind:'column',key:'severity_number'},op:'gte',value:'17'},{field:{kind:'column',key:'severity_number'},op:'lte',value:'20'}]},where={all:[parent,range]};
  window.fetch=async(input:any,init:any)=>{
   const url=new URL(typeof input==='string'?input:input.url,window.location.href),method=String(init?.method||input?.method||'GET').toUpperCase();
   if(url.pathname.startsWith('/api/')&&method!=='GET'){
    const q=JSON.parse(init?.body||'{}'),source=['/api/tail','/api/record-aggregates'].includes(url.pathname);
    if(source)owner.calls.push({path:url.pathname,query:q});else if(!/^\/api\/queries\/[^/]+\/cancel$/.test(url.pathname))owner.forbidden.push(url.pathname);
    if(source)owner.scopeSafe&&=q.environment===environment&&(q.signal===undefined||q.signal==='logs');
    if(url.pathname==='/api/tail'){
     const number=owner.calls.filter((c:any)=>c.path==='/api/tail').length;
     owner.scopeSafe&&=JSON.stringify(q.where)===JSON.stringify(number===1?parent:where);
     if(number===2)owner.tailApplied=true;
    }
    if(url.pathname==='/api/record-aggregates')owner.scopeSafe&&=q.logSeverity===true&&JSON.stringify(q.where)===JSON.stringify(parent)&&q.toMs-q.fromMs===60000&&q.toMs===Number(tail.dataset.lastTo);
   }
   const response=await original(input,init);
   if(url.pathname==='/api/record-aggregates')response.clone().json().then(data=>{owner.aggregate=data;owner.scopeSafe&&=data.count==='8'&&data.logSeverity?.error==='1'&&data.groups?.find((g:any)=>g.g0==='FATAL')?.logSeverity?.info==='1';}).catch(()=>{owner.scopeSafe=false;});
   return response;
  };
  click(button(shell,'Start stream'));return;
 }
 const owner=w.__severity435;
 if(command==='pause'){click(button(shell,'Pause stream'));owner.phase='paused';}
 if(command==='volume'){click(button(shell,'Show volume'));owner.phase='volume';}
 if(command==='band'){
  owner.localAt=owner.calls.length;const chooser=volume.querySelector<HTMLSelectElement>('select[aria-label="Numeric severity"]')!;chooser.focus({preventScroll:true});chooser.value='error';chooser.dispatchEvent(new Event('change',{bubbles:true}));
  const chart=(volume.querySelector('goo-time-bars') as any)?.data,aggregate=owner.aggregate,ceiling=aggregate.timeSeries.reduce((n:bigint,p:any)=>BigInt(p.count)>n?BigInt(p.count):n,1n).toString();owner.scopeSafe&&=chart.ceiling===ceiling&&JSON.stringify(chart.points.map((p:any)=>p.value))===JSON.stringify(aggregate.timeSeries.map((p:any)=>p.logSeverity.error));owner.phase='band';
 }
 if(command==='filter'){click(button(volume,'Filter Error'));owner.phase='staging';}
 if(command==='apply'){owner.localAt=undefined;click(button(shell,'Apply scope'));owner.phase='applying';}
 if(command==='finish'){click(button(shell,'Pause stream'));owner.localAt=owner.calls.length;owner.phase='complete';}
});
export const journey=actions(()=>{
 const s=state.current;
 if(s.phase==='initial')return s.ready?[act('start')]:['Wait'];
 if(s.phase==='reading')return s.loaded&&s.tailReads===1?[act('pause')]:['Wait'];
 if(s.phase==='paused')return s.paused?[act('volume')]:['Wait'];
 if(s.phase==='volume')return s.volume&&s.aggregateReads===1?[act('band')]:['Wait'];
 if(s.phase==='band')return s.band&&s.localSafe?[act('filter')]:['Wait'];
 if(s.phase==='staging')return s.staged&&s.localSafe?[act('apply')]:['Wait'];
 if(s.phase==='applying')return s.applied&&s.tailReads===2?[act('finish')]:['Wait'];
 return ['Wait'];
});
