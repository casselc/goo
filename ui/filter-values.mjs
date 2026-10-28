/* App-owned explicit breakdown reads. The editor remains a supplied-data renderer. */
import * as FilterEditor from './filter-editor.mjs';
import * as Queries from './queries.mjs';
import * as QuestionSession from './question-session.mjs';
const global=globalThis;
const protocol=()=>QuestionSession||(null);
const editorAPI=()=>FilterEditor||(null);
function query(scope,field){if(!scope)throw Error('Apply or Run this scope before loading counted values.');editorAPI().validate({field,op:'exists'},scope.signal);const q={signal:scope.signal,environment:scope.environment,fromMs:scope.fromMs,toMs:scope.toMs,limit:1,offset:0,groupBy:[structuredClone(field)]};if(scope.where)q.where=editorAPI().normalize(scope.where);if(scope.reportedEdge)q.reportedEdge=structuredClone(scope.reportedEdge);if(!q.environment||!Number.isSafeInteger(q.fromMs)||!Number.isSafeInteger(q.toMs)||q.fromMs<0||q.toMs<=q.fromMs||q.toMs-q.fromMs>7200000)throw Error('Require an applied fixed window of at most 120 minutes.');return q}
function result(request,data){
  if(data?.version!=='goo/counted-values/v1')throw Error('Unsupported counted value response.');
  for(const kind of ['count','groups']){const receipt=data.reads?.[kind];if(!receipt||typeof receipt.basis!=='string'||receipt.basis.length!==64||!/^[0-9a-f]{64}$/.test(receipt.basis)||!Number.isFinite(Date.parse(receipt.readAt)))throw Error('Invalid counted value read receipt.');}
  if(protocol().identity(request)!==protocol().identity(data?.question))throw Error('Returned value scope differs from the request.');
  if(!['observed','no_data'].includes(data.state))throw Error('Counted values unavailable: source reads changed or are incomplete.');
  const decimal=v=>typeof v==='string'&&/^(0|[1-9][0-9]*)$/.test(v)&&!/[\r\n]/.test(v)&&BigInt(v)<=18446744073709551615n;
  if(!decimal(data.count)||!Array.isArray(data.groups)||data.groups.length>1000||typeof data.basis!=='string'||!data.basis||!Number.isFinite(Date.parse(data.capturedAt)))throw Error('Invalid counted value response.');
  const seen=new Set();let sum=0n;const items=data.groups.map(g=>{if(!Object.hasOwn(g,'g0')||!(g.g0===null||typeof g.g0==='string')||!decimal(g.count)||BigInt(g.count)===0n||seen.has(g.g0))throw Error('Invalid counted group.');seen.add(g.g0);sum+=BigInt(g.count);return {value:g.g0,count:g.count}});
  if(sum!==BigInt(data.count)||data.state==='no_data'&&data.count!=='0')throw Error('Counted groups disagree with the independent count read.');
  items.sort((a,b)=>BigInt(a.count)===BigInt(b.count)?String(a.value).localeCompare(String(b.value)):BigInt(a.count)>BigInt(b.count)?-1:1);
  return {state:'ready',field:structuredClone(request.groupBy[0]),items,total:data.count,readAt:data.capturedAt,basis:data.basis,reads:structuredClone(data.reads),label:request.signal+' · '+request.environment+' · '+new Date(request.fromMs).toISOString()+' → '+new Date(request.toMs).toISOString()};
 }
function mount(editor,{current,pause=()=>{}}){let generation=0,controller=null,bound=null,disposed=false;
  const identity=()=>{try{const q=current();return q?protocol().identity(query(q,bound?.field||{kind:'column',key:'service'})):null}catch{return null}};
  function invalidate(){generation++;controller?.abort();controller=null;bound=null;editor.setCountedValues()}
  async function load(field){pause();invalidate();if(disposed)return;let request;try{request=query(current(),field)}catch(e){editor.setCountedValues({state:'error',field,error:e.message});return}bound={field:structuredClone(field),key:protocol().identity(request)};const own=++generation;controller=new AbortController();editor.setCountedValues({state:'loading',field});try{const response=await Queries.postJSON('/api/explore/counted-values',request,controller.signal),data=response.data;if(own!==generation||disposed)return;if(identity()!==bound.key){invalidate();return}if(!response.ok)throw Error(data.error||'Counted values unavailable');editor.setCountedValues(result(request,data))}catch(e){if(own===generation&&!disposed&&e.name!=='AbortError'){if(bound&&identity()!==bound.key){invalidate();return}editor.setCountedValues({state:'error',field,error:e.message})}}}
  editor.setValueLoader(load);
  return {invalidate,sync(){if(bound&&identity()!==bound.key)invalidate()},dispose(){disposed=true;invalidate();editor.setValueLoader(null)}};
 }

export {query,result,mount};
