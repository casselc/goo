/* Application-owned, explicitly acquired context over reported span references. */
import {node} from './components.mjs';
import {nativeControlReady,revealFocus} from './dom-visibility.mjs';
import * as Environment from './record-environment.mjs';
import {identity} from './question-session.mjs';
import * as Records from './record-presentation.mjs';
import {renderRefusal} from './source-refusal-view.mjs';
import {matches as matchesLink} from './span-link-membership.mjs';

const incoming=intent=>intent.relation==='reported-span-link-incoming';
export function admitResult(data,intent){
 if(!data||identity(data.question)!==identity(intent.question)||!Array.isArray(data.rows)||data.rows.length>intent.question.limit)throw Error('Returned span-reference question differs from the request.');
 const q=intent.question,from=BigInt(q.fromMs)*1000000n,to=BigInt(q.toMs)*1000000n;
 for(const row of data.rows){
  const member=incoming(intent)?matchesLink(row,intent.target):row?.TraceId===intent.target.traceId&&row?.SpanId===intent.target.spanId;
  if(!member||!Environment.matches(Environment.selection(q),row.ResourceAttributes)||typeof row.timestamp_ns!=='string'||!/^\d+$/.test(row.timestamp_ns)||BigInt(row.timestamp_ns)<from||BigInt(row.timestamp_ns)>=to)throw Error('Returned span falls outside the requested reference scope.');
 }
 return data;
}
export function mount(host,{fetchQuery,current=()=>true}={}){
 host.setAttribute('data-span-link-context','');host.setAttribute('aria-label','Span reference lookup');host.hidden=true;
 let generation=0,controller=null,held=null;
 const alive=()=>!!held&&current()&&host.isConnected&&!host.hidden;
 function restore(){if(!held)return;for(const item of held.siblings)if(item.element.isConnected)item.element.hidden=item.hidden;host.parentElement.scrollTop=held.scroll;}
 function invalidate(){generation++;controller?.abort();controller=null;restore();held=null;host.hidden=true;host.replaceChildren();}
 function back(){
  if(!held)return false;const opener=held.opener,parent=host.parentElement;invalidate();
  const box=opener?.getBoundingClientRect?.(),frame=parent?.getBoundingClientRect?.();
  const visible=box&&frame&&box.top>=Math.max(0,frame.top)&&box.bottom<=Math.min(globalThis.innerHeight,frame.bottom)&&box.left>=Math.max(0,frame.left)&&box.right<=Math.min(globalThis.innerWidth,frame.right);
  if(visible&&nativeControlReady(opener))opener.focus({preventScroll:true});else revealFocus(opener,undefined,{disclosures:true});
  return true;
 }
 function stage({intent,opener,parentBasis=null,parentReadAt=null}){
  if(!current()||!nativeControlReady(opener))return false;
  invalidate();
  const parent=host.parentElement;held={intent:structuredClone(intent),opener,parentBasis,parentReadAt,siblings:[...parent.children].filter(element=>element!==host&&!element.classList.contains('inspector-heading')).map(element=>({element,hidden:element.hidden})),scroll:parent.scrollTop};
  for(const item of held.siblings)item.element.hidden=true;
  intent=held.intent;const inverse=incoming(intent),q=intent.question;let offset=q.offset??0,hasMore=false;
  host.hidden=false;host.dataset.phase='staged';host.dataset.relation=inverse?'incoming':'target';
  const backButton=node('button','Back to trace'),run=node('button',inverse?'Run reference search':'Run target lookup'),controls=node('div'),status=node('p'),results=node('div'),scope=node('p'),reference=node('details'),json=node('pre'),title=node('h3',inverse?'Referencing spans':'Linked span');
  const paging=node('nav'),previous=node('button','Previous'),next=node('button','Next');
  paging.setAttribute('aria-label','Reference pages');paging.className='controls';paging.hidden=true;
  previous.setAttribute('aria-label','Previous reference page');next.setAttribute('aria-label','Next reference page');paging.append(previous,next);
  controls.className='controls';status.setAttribute('role','status');scope.className='evidence-note';
  scope.textContent=Environment.questionLabel(q)+' · '+new Date(q.fromMs).toISOString()+' to '+new Date(q.toMs).toISOString()+' (end excluded). Service, namespace and population filters are omitted.';
  const target=node('p',intent.target.traceId+' / '+intent.target.spanId);target.className='record-reference';target.style.overflowWrap='anywhere';
  status.textContent=inverse?'Reference search staged · Run to find spans reporting this exact link. References do not establish causality or delivery.':'Target lookup staged · Run to read spans. Reported links do not prove causality or complete delivery.';
  const provenance=lookup=>JSON.stringify({question:q,[inverse?'selectedSpan':'reportedLink']:intent.origin,attributes:intent.attributes,parentBasis,parentReadAt,...(lookup?{lookup}:{})},null,2);
  json.textContent=provenance();reference.append(node('summary',inverse?'Question and reference target':'Question and reported link'),json);
  controls.append(backButton,run);if(inverse)controls.append(paging);
  host.replaceChildren(title,controls,target,scope,status,results,reference);
  const token=generation,owner=held;
  const valid=button=>token===generation&&held===owner&&alive()&&nativeControlReady(button);
  backButton.onclick=()=>{if(valid(backButton))back()};
  async function readPage(page,opener){
   if(!valid(opener)||run.disabled||!Number.isInteger(page)||page<0||page>1000)return;
   const moveFocus=globalThis.document?.activeElement===opener;
   controller?.abort();const read=new AbortController();controller=read;
   if(moveFocus)revealFocus(backButton);
   run.disabled=true;previous.disabled=true;next.disabled=true;host.dataset.phase='running';results.replaceChildren();status.textContent=inverse?'Reading referencing spans…':'Reading linked span…';
   offset=page;const requested={...structuredClone(q),offset:page};
   json.textContent=provenance({question:requested,phase:'running'});
   try{
    const data=await fetchQuery(requested,read.signal);read.signal.throwIfAborted();if(!valid(backButton))return;
    admitResult(data,{...intent,question:requested});offset=page;hasMore=!!data.hasMore;host.dataset.phase='ready';
    const empty=String(data.count)==='0'&&data.state!=='changed_between_reads';
    status.textContent=data.rows.length?(data.count+' matching span observations · '+(inverse?(offset+1)+'–'+(offset+data.rows.length)+' shown':data.rows.length+' on this page'+(data.hasMore?' · additional matches not shown':''))+(inverse?'. Fresh read; pages may change.':'. Duplicate IDs do not identify a unique invocation. Coverage unknown.')):empty?(inverse?'No spans report this link in this environment and window. Coverage unknown.':'No matching spans in this environment and window · absence does not establish missing delivery.'):'No span observations returned on this page. Coverage unknown.';
    if(data.state==='changed_between_reads')status.append(' Source reads changed; count and page may disagree.');
    if(inverse&&hasMore&&offset+q.limit>1000)status.append(' Paging limit reached. Narrow the window for further results.');
    if(data.rows.length){
     const table=Records.table({signal:'spans',context:false,mobileCompact:true,summaryCompact:false,caption:inverse?'Spans reporting this link · separate source read':'Linked span observations · separate source read'});
     for(const row of data.rows){const item=table.row(row);table.body.append(item.row);const detail=node('details'),body=node('div');detail.append(node('summary','Span attributes'));Records.attributes(body,row);detail.append(body);item.cells[2].append(detail)}results.append(table.viewport);
    }
    json.textContent=provenance({question:data.question,basis:data.basis,capturedAt:data.capturedAt,count:data.count,state:data.state,hasMore:data.hasMore});
    results.append(node('small','Read '+(data.capturedAt||'time not reported')));
    paging.hidden=!inverse;
   }catch(error){if(error.name==='AbortError')return;if(valid(backButton)){host.dataset.phase='unavailable';paging.hidden=true;json.textContent=provenance({question:requested,phase:'unavailable'});renderRefusal(status,error,{prefix:inverse?'Reference search unavailable · ':'Linked span unavailable · '})}}
   finally{if(controller===read)controller=null;if(valid(backButton)){run.disabled=false;previous.disabled=offset===0;next.disabled=!hasMore||offset+q.limit>1000;}}
  }
  run.onclick=()=>readPage(offset,run);
  previous.onclick=()=>readPage(Math.max(0,offset-q.limit),previous);
  next.onclick=()=>readPage(offset+q.limit,next);
  revealFocus(run);return true;
 }
 return {stage,back,invalidate,isOpen:()=>!!held};
}
