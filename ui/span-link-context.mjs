/* Application-owned, explicitly acquired context over one reported link tuple. */
import {node} from './components.mjs';
import {nativeControlReady,revealFocus} from './dom-visibility.mjs';
import * as Environment from './record-environment.mjs';
import {identity} from './question-session.mjs';
import * as Records from './record-presentation.mjs';
import {renderRefusal} from './source-refusal-view.mjs';

export function admitResult(data,intent){
 if(!data||identity(data.question)!==identity(intent.question)||!Array.isArray(data.rows)||data.rows.length>intent.question.limit)throw Error('Returned linked-span question differs from the requested target.');
 const q=intent.question,from=BigInt(q.fromMs)*1000000n,to=BigInt(q.toMs)*1000000n;
 for(const row of data.rows){
  if(row?.TraceId!==intent.target.traceId||row?.SpanId!==intent.target.spanId||!Environment.matches(Environment.selection(q),row.ResourceAttributes)||typeof row.timestamp_ns!=='string'||!/^\d+$/.test(row.timestamp_ns)||BigInt(row.timestamp_ns)<from||BigInt(row.timestamp_ns)>=to)throw Error('Returned span falls outside the exact target scope.');
 }
 return data;
}
export function mount(host,{fetchQuery,current=()=>true}={}){
 host.setAttribute('data-span-link-context','');host.setAttribute('aria-label','Linked span lookup');host.hidden=true;
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
  const parent=host.parentElement;held={intent:structuredClone(intent),opener,parentBasis,parentReadAt,siblings:[...parent.children].filter(element=>element!==host&& !element.classList.contains('inspector-heading')).map(element=>({element,hidden:element.hidden})),scroll:parent.scrollTop};
  for(const item of held.siblings)item.element.hidden=true;
  host.hidden=false;host.dataset.phase='staged';
  const backButton=node('button','Back to trace'),run=node('button','Run target lookup'),controls=node('div'),status=node('p'),results=node('div'),scope=node('p'),reference=node('details'),json=node('pre'),title=node('h3','Linked span');
  controls.className='controls';status.setAttribute('role','status');scope.className='evidence-note';
  intent=held.intent;const q=intent.question;scope.textContent=Environment.questionLabel(q)+' · '+new Date(q.fromMs).toISOString()+' to '+new Date(q.toMs).toISOString()+' (end excluded). Service, namespace and population filters are omitted.';
  const target=node('p',intent.target.traceId+' / '+intent.target.spanId);target.className='record-reference';target.style.overflowWrap='anywhere';
  status.textContent='Target lookup staged · Run to read spans. Reported links do not prove causality or complete delivery.';
  json.textContent=JSON.stringify({question:q,reportedLink:intent.origin,attributes:intent.attributes,parentBasis,parentReadAt},null,2);reference.append(node('summary','Question and reported link'),json);
  controls.append(backButton,run);host.replaceChildren(title,controls,target,scope,status,results,reference);
  const token=generation,owner=held;
  const valid=button=>token===generation&&held===owner&&alive()&&nativeControlReady(button);
  backButton.onclick=()=>{if(valid(backButton))back()};
  run.onclick=async()=>{
   if(!valid(run)||run.disabled)return;
   const moveFocus=globalThis.document?.activeElement===run;controller?.abort();const read=new AbortController();controller=read;if(moveFocus)revealFocus(backButton);run.disabled=true;host.dataset.phase='running';results.replaceChildren();status.textContent='Reading linked span…';
   try{
    const data=await fetchQuery(structuredClone(q),read.signal);read.signal.throwIfAborted();if(!valid(backButton))return;
    admitResult(data,intent);host.dataset.phase='ready';
    status.textContent=data.rows.length?(data.count+' matching span observations · '+data.rows.length+' on this page'+(data.hasMore?' · additional matches not shown':'')+'. Duplicate IDs do not identify a unique invocation. Coverage unknown.'):String(data.count)==='0'&&data.state!=='changed_between_reads'?'No matching spans in this environment and window · absence does not establish missing delivery.':'No span observations returned on this page. Coverage unknown.';
    if(data.state==='changed_between_reads')status.append(' Source reads changed; count and page may disagree.');
    if(data.rows.length){const table=Records.table({signal:'spans',context:false,mobileCompact:true,summaryCompact:false,caption:'Linked span observations · separate source read'});for(const row of data.rows){const item=table.row(row);table.body.append(item.row);const detail=node('details'),body=node('div');detail.append(node('summary','Span attributes'));Records.attributes(body,row);detail.append(body);item.cells[2].append(detail)}results.append(table.viewport)}
    json.textContent=JSON.stringify({question:q,reportedLink:intent.origin,attributes:intent.attributes,parentBasis,parentReadAt,lookup:{question:data.question,basis:data.basis,capturedAt:data.capturedAt,count:data.count,state:data.state,hasMore:data.hasMore}},null,2);
    const readTime=node('small','Read '+(data.capturedAt||'time not reported'));results.append(readTime);
   }catch(error){if(error.name==='AbortError')return;if(valid(backButton)){host.dataset.phase='unavailable';renderRefusal(status,error,{prefix:'Linked span unavailable · '})}}
   finally{if(controller===read)controller=null;if(valid(backButton))run.disabled=false;}
  };
  revealFocus(run);return true;
 }
 return {stage,back,invalidate,isOpen:()=>!!held};
}
