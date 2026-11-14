/* App-owned identity discovery. Choosing a scope only stages existing native discovery. */
import * as UI from './components.mjs';
import * as Queries from './queries.mjs';
import * as Shape from './intent-shape.mjs';
import * as Environment from './record-environment.mjs';
import * as ServiceIdentity from './service-identity.mjs';
import * as SourceRefusal from './source-refusal.mjs';
import {families} from './source-capabilities.mjs';
import {nativeControlReady,completionControlReady,revealFocus} from './dom-visibility.mjs';

const profiles=families.filter(family=>!['logs','spans'].includes(family.kind));
const states=new Set(['compatible','absent','incompatible','unreadable']);
const invalid=()=>{throw Error('Metric scope response rejected.');};
function object(value,keys){try{return Shape.object(value,keys)}catch{invalid()}}
function array(value){try{return Shape.array(value)}catch{invalid()}}
function text(value,max=4096){return typeof value==='string'&&value.isWellFormed()&&new TextEncoder().encode(value).length<=max}
function utc(value){return text(value,64)&&/^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(?:\.\d{1,9})?Z$/.test(value)&&Number.isFinite(Date.parse(value))&&new Date(value).toISOString().slice(0,19)===value.slice(0,19)}
function namespace(value,selectable){
  object(value,['mode','value']);if(!['missing','value'].includes(value.mode)||value.value!==undefined&&!text(value.value)||value.mode==='missing'&&(value.value??'')!=='')invalid();
  if(selectable&&!Shape.text(value.value??'',128))invalid();
  return {mode:value.mode,...(value.mode==='value'&&value.value?{value:value.value}:{})};
}
function environment(value){object(value,['mode','value']);if(!['named','empty','unreported'].includes(value.mode)||value.value!==undefined&&!text(value.value)||value.mode!=='named'&&(value.value??'')!=='')invalid();return {mode:value.mode,...(value.mode==='named'?{value:value.value??''}:{})}}
function selector(value,scope){let selected;try{selected=Environment.normalize(value)}catch{invalid()}if(selected.policy!=='effective'||JSON.stringify(selected.scope)!==JSON.stringify(scope))invalid();return selected}
function windowQuestion(value){object(value,['fromMs','toMs']);if(!Number.isSafeInteger(value.fromMs)||!Number.isSafeInteger(value.toMs)||value.fromMs<0||value.toMs<=value.fromMs||value.toMs-value.fromMs>7200000||value.toMs>4102444800000)invalid();return value;}
function validateResult(value,question){
  windowQuestion(question);windowQuestion(value?.question);
  object(value,['version','question','capturedAt','sources','diagnostics','choices','unit','warning']);object(value.question,['fromMs','toMs']);
  if(value.version!=='goo/metric-scopes/v1'||value.unit!=='exported-points'||!utc(value.capturedAt)||value.question.fromMs!==question.fromMs||value.question.toMs!==question.toMs||!text(value.warning,1024))invalid();
  array(value.sources);if(value.sources.length!==4)invalid();const sourceKinds=new Set();
  for(const source of value.sources){object(source,['kind','table','state','detail']);const profile=profiles.find(p=>p.kind===source.kind);if(!profile||sourceKinds.has(source.kind)||source.table!==profile.table||!states.has(source.state)||!text(source.detail,256))invalid();sourceKinds.add(source.kind)}
  if(!value.sources.some(source=>source.state==='compatible'))invalid();
  array(value.diagnostics);const excluded=value.sources.filter(source=>source.state!=='compatible'),diagnosed=new Set();if(value.diagnostics.length!==excluded.length)invalid();for(const diagnostic of value.diagnostics){object(diagnostic,['kind','reason']);if(diagnosed.has(diagnostic.kind)||!excluded.some(source=>source.kind===diagnostic.kind&&source.state===diagnostic.reason))invalid();diagnosed.add(diagnostic.kind);}
  array(value.choices);if(value.choices.length>128)invalid();const identities=new Set();
  for(const choice of value.choices){
    object(choice,['service','namespace','environment','environmentSelector','points','firstMs','lastMs','selectable','reason']);
    if(!text(choice.service)||typeof choice.selectable!=='boolean'||typeof choice.points!=='string'||(!/^[1-9]\d{0,38}$/.test(choice.points)||BigInt(choice.points)>(1n<<128n)-1n)||!Number.isSafeInteger(choice.firstMs)||!Number.isSafeInteger(choice.lastMs)||choice.firstMs<question.fromMs||choice.lastMs>=question.toMs||choice.lastMs<choice.firstMs||choice.reason!==undefined&&!text(choice.reason,256))invalid();
    const ns=namespace(choice.namespace,choice.selectable),scope=environment(choice.environment);
    if(choice.environmentSelector!==undefined)selector(choice.environmentSelector,scope);
    if(choice.selectable){if(!Shape.text(choice.service,128,{empty:false})||!choice.environmentSelector||choice.reason)invalid();}
    else if(!choice.reason)invalid();
    const key=JSON.stringify([choice.service,ns,scope]);if(identities.has(key))invalid();identities.add(key);
  }
  if(new TextEncoder().encode(JSON.stringify(value)).length>1<<20)invalid();
  return structuredClone(value);
}
function selection(choice,question){
  windowQuestion(question);
  if(choice.selectable!==true)throw Error('Choose a supported reported service and scope.');
  const ns=namespace(choice.namespace,true),scope=environment(choice.environment),selected=selector(choice.environmentSelector,scope);
  if(!Shape.text(choice.service,128,{empty:false}))invalid();
  return {service:choice.service,namespace:ns,environment:'',environmentSelector:selected,fromMs:question.fromMs,toMs:question.toMs};
}
function literal(value){const raw=JSON.stringify(value),chars=Array.from(raw);return chars.length>100?chars.slice(0,100).join('')+'…':raw}
function environmentLabel(scope){return scope.mode==='unreported'?'Unreported environment':scope.mode==='empty'?'Empty environment label':'Environment '+literal(scope.value)+' · current/legacy precedence'}
function namespaceLabel(ns){return ns.mode==='missing'?'Namespace missing':ns.value?'Namespace '+literal(ns.value):'Namespace present empty'}
function choiceName(choice){return 'Choose metric scope · '+ServiceIdentity.name(choice.service,choice.namespace)+' · '+Environment.label(choice.environmentSelector)}
function mount(host,{active=()=>true,onRead,onSelect}={}){
  const n=UI.node,box=n('details',undefined,'metric-scopes'),summary=n('summary','Find metric scopes'),controls=n('div',undefined,'controls'),minutes=n('input'),label=n('label','Window (minutes)'),read=n('button','Read metric scopes'),status=n('p',undefined,'metric-scopes-status'),windowNote=n('p',undefined,'metric-scope-window'),output=n('div');
  minutes.type='number';minutes.min='1';minutes.max='120';minutes.value='2';minutes.setAttribute('aria-label','Metric scope window (minutes)');read.type='button';status.setAttribute('role','status');label.append(minutes);controls.append(label,read);windowNote.hidden=true;box.append(summary,controls,windowNote,status,output);host.append(box);
  let generation=0,controller=null,disposed=false,adopted=null,fixedWindow=null;
  const current=control=>!disposed&&box.open&&active()&&!document.hidden&&nativeControlReady(control);
  function clear(message=''){generation++;controller?.abort();controller=null;adopted=null;read.disabled=false;box.removeAttribute('aria-busy');output.replaceChildren();windowNote.hidden=!fixedWindow;status.textContent=message;}
  function render(result,g){
    const viewport=n('goo-table'),table=n('table'),caption=n('caption','All metric services and environments · '+new Date(result.question.fromMs).toISOString()+' → '+new Date(result.question.toMs).toISOString()),head=n('thead'),heading=n('tr'),body=n('tbody');viewport.className='metric-scope-table';viewport.setAttribute('density','compact');viewport.setAttribute('aria-label','Observed metric scopes');
    for(const text of ['Service','Scope','Exported points','Use']){const th=n('th',text);th.scope='col';heading.append(th)}head.append(heading);
    for(const choice of [...result.choices].sort((a,b)=>Number(b.selectable)-Number(a.selectable))){
      const row=n('tr'),service=n('th',choice.service?literal(choice.service):'Empty or unreported service'),scope=n('td'),points=n('td',choice.points),action=n('td'),choose=n('button','Choose metric scope');service.scope='row';service.title=JSON.stringify(choice.service);scope.append(n('div',namespaceLabel(choice.namespace)),n('div',environmentLabel(choice.environment)));choose.type='button';choose.disabled=!choice.selectable;choose.setAttribute('aria-label',choice.selectable?choiceName(choice):'Metric scope unavailable');choose.title=choice.reason||'Stage instrument discovery in this fixed scope. No source read.';if(choice.reason)action.append(n('small',choice.reason));
      choose.onclick=()=>{if(!current(choose)||g!==generation||adopted!==result)return;try{if(onSelect?.(selection(choice,result.question),choose)===false)return;clear();box.open=false}catch(error){status.textContent=error.message}};action.append(choose);row.append(service,scope,points,action);body.append(row);
    }
    table.append(caption,head,body);viewport.append(table);windowNote.hidden=true;output.replaceChildren(viewport);const excluded=result.sources.filter(source=>source.state!=='compatible');if(excluded.length){const details=n('details',undefined,'metric-source-exclusions'),list=n('ul');for(const source of excluded)list.append(n('li',source.table+' · '+source.detail));details.append(n('summary',excluded.length+' sources excluded'),list);output.append(details);}if(!result.choices.length)output.append(n('p','No metric service scopes observed in this window.'));
  }
  read.onclick=async()=>{
    if(!current(read))return;const count=Number(minutes.value);if(!fixedWindow&&(!Number.isInteger(count)||count<1||count>120)){status.textContent='Choose 1–120 minutes.';return}onRead?.();if(!current(read))return;
    clear();const g=generation,toMs=Date.now(),question=fixedWindow?structuredClone(fixedWindow):{fromMs:toMs-count*60000,toMs},own=new AbortController();controller=own;read.disabled=true;box.setAttribute('aria-busy','true');status.textContent='Reading metric service scopes…';
    try{const reply=await Queries.postJSON('/api/metric-scopes',question,own.signal);if(g!==generation||disposed||!active()||document.hidden||!box.open||!completionControlReady(read))return;if(!reply.ok)throw SourceRefusal.responseError(reply,'Metric scope discovery unavailable.');adopted=validateResult(reply.data,question);render(adopted,g);status.textContent=adopted.choices.length+' observed service scopes';}
    catch(error){if(g===generation&&!disposed&&active()&&!document.hidden&&box.open&&completionControlReady(read)&&error.name!=='AbortError')status.textContent=error instanceof SourceRefusal.SourceReadError?error.message:'Metric scope discovery unavailable.';}
    finally{if(controller===own){controller=null;read.disabled=false;box.removeAttribute('aria-busy')}}
  };
  minutes.oninput=()=>{fixedWindow=null;windowNote.hidden=true;clear();};box.addEventListener('toggle',()=>{if(!box.open)clear()});
  const visibility=()=>{if(document.hidden)clear('Metric scope discovery paused. Read again explicitly.');};document.addEventListener('visibilitychange',visibility);
  return {clear,open(window){clear();fixedWindow=window?structuredClone(windowQuestion(window)):null;windowNote.hidden=!fixedWindow;if(fixedWindow){minutes.value=String((fixedWindow.toMs-fixedWindow.fromMs)/60000);windowNote.textContent='Fixed scope window: '+new Date(fixedWindow.fromMs).toISOString()+' → '+new Date(fixedWindow.toMs).toISOString();}box.open=true;revealFocus(read,summary,{disclosures:true});},dispose(){clear();disposed=true;document.removeEventListener('visibilitychange',visibility);box.remove()}};
}
export {validateResult,selection,choiceName,mount};
