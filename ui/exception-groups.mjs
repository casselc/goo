/* Query-only exception populations. The application owns reads and navigation. */
import * as Protocol from './question-protocol.mjs';
import * as Predicates from './predicate-validation.mjs';
import * as Session from './question-session.mjs';
import * as Identity from './exception-fingerprint-identity.mjs';
import * as Environment from './record-environment.mjs';
import * as EvidenceSummary from './evidence-summary.mjs';
import * as UI from './components.mjs';
import * as Queries from './queries.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
let nextInstance=0;
const key=value=>value?Session.identity(value):'';
const clone=structuredClone,bytes=value=>new TextEncoder().encode(value).length;
const uint=value=>typeof value==='string'&&/^(0|[1-9]\d*)$/.test(value)&&BigInt(value)<=18446744073709551615n;
const literal=value=>typeof value==='string'&&value.isWellFormed();
const hash=value=>typeof value==='string'&&/^[a-f0-9]{64}$/.test(value);
const utc=value=>typeof value==='string'&&/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?Z$/.test(value)&&Number.isFinite(Date.parse(value));
const closed=(value,keys)=>{if(!value||typeof value!=='object'||Array.isArray(value)||![Object.prototype,null].includes(Object.getPrototypeOf(value))||Reflect.ownKeys(value).some(key=>typeof key!=='string'||!keys.includes(key)||!Object.hasOwn(Object.getOwnPropertyDescriptor(value,key),'value')))throw Error('Invalid exception-group response fields.');};

export function question(parent){
 Protocol.validateRecordQuestion(parent);
 if(parent.reportedEdge||(parent.offset??0)!==0||parent.numeric||parent.numericHistogram)throw Error('Choose a page-zero ordinary log or span population.');
 const q=clone(parent);for(const key of ['groupBy','bucketMs','groupTimeSeries','summary','spanStatus'])delete q[key];q.limit=100;q.offset=0;
 if(q.where)Predicates.validate(q.where,q.signal);return q;
}
export function stageQuestion(parent,fingerprint){
 const q=question(parent);Identity.decode(fingerprint);
 const leaf={field:{...Identity.field},op:'eq',value:fingerprint};q.where=q.where?{all:[clone(q.where),leaf]}:leaf;
 Protocol.validateRecordQuestion(q);Predicates.validate(q.where,q.signal);return q;
}
export function admit(request,result){
 const q=question(request);
 closed(result,['version','question','count','state','capturedAt','basis','reads','warnings','groups']);
 if(result.version!=='goo/exception-groups/v1'||Session.identity(result.question)!==Session.identity(q)||!uint(result.count)||!['observed','no_data','changed_between_reads'].includes(result.state)||!utc(result.capturedAt)||!hash(result.basis)||!Array.isArray(result.warnings)||result.warnings.some(v=>!literal(v))||!Array.isArray(result.groups)||result.groups.length>1000)throw Error('Exception groups do not match the fixed population.');
 closed(result.reads,['count','groups']);for(const name of ['count','groups']){const read=result.reads[name];closed(read,['readAt','basis']);if(!utc(read.readAt)||!hash(read.basis))throw Error('Exception read receipt is unavailable.');}
 const seen=new Set(),start=BigInt(q.fromMs)*1000000n,end=BigInt(q.toMs)*1000000n;let total=0n;
 for(const group of result.groups){
  closed(group,['fingerprint','state','reason','count','firstNS','lastNS','inputs','affected','affectedCount','affectedLimited']);
  const identity=Identity.decode(group.fingerprint);
  if(q.signal==='logs'&&['none','multiple'].includes(identity.reason))throw Error('Span-event classification cannot describe a log signature.');
  if(seen.has(identity.id)||group.state!==identity.state||group.reason!==identity.reason||!uint(group.count)||group.count==='0'||!uint(group.firstNS)||!uint(group.lastNS)||BigInt(group.firstNS)<start||BigInt(group.lastNS)>=end||BigInt(group.firstNS)>BigInt(group.lastNS))throw Error('Invalid exception group identity, count or record window.');
  seen.add(identity.id);total+=BigInt(group.count);
  if(identity.classified){closed(group.inputs,['type','stacktrace']);if(!literal(group.inputs.type)||!group.inputs.type||bytes(group.inputs.type)>256||!literal(group.inputs.stacktrace)||!group.inputs.stacktrace||bytes(group.inputs.stacktrace)>8192)throw Error('Exact reported signature inputs are unavailable.');}
  else if(group.inputs!==null)throw Error('Unclassified exception groups cannot supply a complete signature.');
  if(!Array.isArray(group.affected)||group.affected.length>32||!uint(group.affectedCount)||BigInt(group.affectedCount)<1n||BigInt(group.affectedCount)>BigInt(group.count)||typeof group.affectedLimited!=='boolean'||group.affected.length!==Number(BigInt(group.affectedCount)>32n?32n:BigInt(group.affectedCount))||group.affectedLimited!==(BigInt(group.affectedCount)>32n))throw Error('Invalid bounded affected-context coverage.');
  const contexts=new Set();for(const value of group.affected){closed(value,['service','namespace','version']);if(!literal(value.service)||value.namespace!==null&&!literal(value.namespace)||value.version!==null&&!literal(value.version))throw Error('Invalid qualified exception context.');const key=JSON.stringify([value.service,value.namespace,value.version]);if(contexts.has(key))throw Error('Duplicate exception context.');contexts.add(key);}
 }
 if(result.state!=='changed_between_reads'&&(total!==BigInt(result.count)||result.state!==(result.count==='0'?'no_data':'observed')))throw Error('Exception group counts do not conserve the population.');
 if(bytes(JSON.stringify(result))>8*1024*1024)throw Error('Exception groups exceed the response bound.');
 return clone(result);
}
const contextLabel=value=>(value.namespace===null?'Namespace not reported':value.namespace===''?'Namespace empty':JSON.stringify(value.namespace))+' / '+JSON.stringify(value.service)+' · version '+(value.version===null?'not reported':JSON.stringify(value.version));
function clock(node,value){const time=EvidenceSummary.unixNs(value);if(!time){node.textContent='Unavailable';return;}const e=UI.node('time',time.text);e.dateTime=time.exact;e.title=time.exact+' · '+time.title;node.append(e);}

export function mount(host,{current=()=>true,onStage=()=>{}}={}){
 const instance=++nextInstance,n=UI.node,root=n('details',undefined,'exception-groups'),summary=n('summary','Exception groups'),read=n('button','Read exception groups'),status=n('p'),body=n('div'),tools=n('div',undefined,'controls'),find=n('input'),order=n('select'),facts=n('details'),factBody=n('dl');
 root.append(summary,n('p','Exact reported type + stack · fingerprint v1','evidence-note'),read,status,tools,body,facts);facts.append(n('summary','Group read facts'),factBody);facts.hidden=tools.hidden=true;find.type='search';find.placeholder='Find a signature or context';find.setAttribute('aria-label','Find exception groups');order.setAttribute('aria-label','Order exception groups');for(const [value,label]of [['count','Most captured records'],['first','First record in window'],['last','Last record in window'],['label','Signature']]){const option=n('option',label);option.value=value;order.append(option);}tools.append(find,order);status.setAttribute('role','status');host.append(root);
 let source=null,snapshot=null,controller=null,revision=0,displayRevision=0,busy=false,disposed=false;
 const alive=()=>!disposed&&root.isConnected&&current(),ready=control=>alive()&&root.open&&root.contains(control)&&nativeControlReady(control);
 function withdraw(){revision++;displayRevision++;controller?.abort();controller=null;busy=false;snapshot=null;body.replaceChildren();factBody.replaceChildren();facts.hidden=tools.hidden=true;read.disabled=!source||!alive();root.dataset.state='initial';}
 function set(value){let next=null;try{if(value)next=question(value);}catch(error){withdraw();source=null;read.disabled=true;status.textContent=error.message;return;}if(key(next)!==key(source)){withdraw();source=next;status.textContent='';find.value='';}read.disabled=busy||!source||!alive();}
 function render(){
  body.replaceChildren();if(!snapshot||!['observed','no_data'].includes(snapshot.state))return;
  const own=revision,display=++displayRevision,term=find.value.toLocaleLowerCase(),rows=[...snapshot.groups].filter(g=>[g.inputs?.type,g.inputs?.stacktrace,g.reason,Identity.reasons[g.reason],g.state,g.fingerprint,...g.affected.map(contextLabel)].join(' ').toLocaleLowerCase().includes(term));
  rows.sort((a,b)=>{if(order.value==='label')return (a.inputs?.type||Identity.reasons[a.reason]).localeCompare(b.inputs?.type||Identity.reasons[b.reason])||a.fingerprint.localeCompare(b.fingerprint);const key=order.value==='count'?'count':order.value==='first'?'firstNS':'lastNS',delta=BigInt(a[key])-BigInt(b[key]);return delta===0n?a.fingerprint.localeCompare(b.fingerprint):delta>0n?(order.value==='first'?1:-1):(order.value==='first'?-1:1);});
  if(!rows.length){body.append(n('p',snapshot.count==='0'?'No matching records.':'No loaded groups match this search.'));return;}
  const viewport=n('goo-table'),table=n('table'),thead=n('thead'),head=n('tr'),tbody=n('tbody');viewport.setAttribute('density','compact');table.setAttribute('aria-label','Captured exception groups');for(const text of ['Signature','Records / share','Window first / last','Reported contexts','Investigate']){const th=n('th',text);th.scope='col';head.append(th);}thead.append(head);table.append(thead,tbody);viewport.append(table);body.append(viewport);
  for(const [index,g]of rows.entries()){
   const row=n('tr'),title=n('th'),label=n('strong',g.inputs?JSON.stringify(g.inputs.type):Identity.reasons[g.reason]),count=n('td'),times=n('td'),contexts=n('td'),actions=n('td'),details=n('details'),open=n('button','Stage group records');row.dataset.fingerprint=g.fingerprint;row.dataset.state=g.state;title.scope='row';label.id='exception-group-'+instance+'-'+own+'-'+index;title.append(label);const cue=n('span',g.inputs?(()=>{const preview=g.inputs.stacktrace.split(/\r?\n/u).filter(line=>line.length).slice(0,2).join(' · '),points=Array.from(preview);return points.slice(0,120).join('')+(points.length>120?'…':'')+' · '+g.fingerprint.slice(-8);})():'Unclassified','exception-group-kind');cue.id=label.id+'-context';if(g.inputs)cue.title=g.inputs.stacktrace;title.append(cue);open.setAttribute('aria-describedby',label.id+' '+cue.id);
   const percentage=snapshot.count==='0'?'':(Number((BigInt(g.count)*10000n)/BigInt(snapshot.count))/100).toFixed(2)+'%';count.append(n('strong',g.count),n('small',percentage));count.title=g.count+'/'+snapshot.count+' matching '+source.signal+' records (exact share)';const first=n('span'),last=n('span');clock(first,g.firstNS);clock(last,g.lastNS);times.append(first,last);times.title='First/last matching record timestamps in this searched window';
   contexts.append(n('span',g.affectedCount+' qualified '+(g.affectedCount==='1'?'context':'contexts')+(g.affectedLimited?' · first 32 shown':'')));details.append(n('summary','Signature details'));
   if(g.inputs){const type=n('code',JSON.stringify(g.inputs.type)),stack=n('pre');stack.textContent=g.inputs.stacktrace;stack.tabIndex=0;stack.setAttribute('aria-label','Reported exception stack trace');details.append(n('p','Reported type'),type,n('p','Reported stack'),stack);}else details.append(n('p',Identity.reasons[g.reason]));details.append(n('code',g.fingerprint));const list=n('ul');list.tabIndex=0;list.setAttribute('aria-label','Reported qualified contexts');for(const value of g.affected)list.append(n('li',contextLabel(value)));details.append(list);title.append(details);
   let staged;try{staged=stageQuestion(source,g.fingerprint);}catch(error){open.disabled=true;open.title=error.message;}
   open.onclick=()=>{if(own!==revision||display!==displayRevision||snapshot?.groups.find(x=>x.fingerprint===g.fingerprint)!==g||!ready(open)||open.disabled||!staged)return;onStage(clone(staged),open);};actions.append(open);[title,count,times,contexts,actions].forEach((cell,i)=>{cell.dataset.label=['Signature','Records / share','Window first / last','Reported contexts','Investigate'][i];row.append(cell);});tbody.append(row);
  }
 }
 read.onclick=async()=>{
  if(!ready(read)||busy||!source)return;
  withdraw();const own=revision,requested=clone(source);busy=true;controller=new AbortController();read.disabled=true;root.dataset.state='loading';status.textContent='Reading exception groups…';
  try{const response=await Queries.postJSON('/api/exception-groups',requested,controller.signal);if(own!==revision||!alive()||!root.open)return;if(!response.ok)throw Error(response.data?.error||'Exception groups unavailable.');snapshot=admit(requested,response.data);root.dataset.state=snapshot.state;
   const factsRows=[['Population',source.signal+' records'],['Fixed scope',Environment.questionLabel(source)+' · ['+new Date(source.fromMs).toISOString()+', '+new Date(source.toMs).toISOString()+')'],['Acquired',snapshot.capturedAt],['Count read',snapshot.reads.count.readAt],['Count basis',snapshot.reads.count.basis],['Group read',snapshot.reads.groups.readAt],['Group basis',snapshot.reads.groups.basis],['Grain','One matching record, not one exception event'],['Window first / last','Matching record timestamps only; not event time or lifetime first seen']];for(const [key,value]of factsRows){const item=n('div');item.append(n('dt',key),n('dd',value));factBody.append(item);}for(const warning of snapshot.warnings){const item=n('div');item.append(n('dt','Source note'),n('dd',warning));factBody.append(item);}facts.hidden=false;
   if(snapshot.state==='changed_between_reads'){status.textContent='Source reads differ. Read groups again; ranking and group actions are unavailable.';snapshot=null;return;}
   status.textContent=snapshot.count+' matching '+source.signal+' records · '+snapshot.groups.length+' groups · coverage unknown';tools.hidden=false;render();
  }catch(error){if(own===revision&&alive()&&error.name!=='AbortError'){snapshot=null;root.dataset.state='unavailable';status.textContent=error.message;}}
  finally{if(own===revision){busy=false;controller=null;read.disabled=!source||!alive();}}
 };
 find.oninput=order.onchange=render;root.addEventListener('toggle',()=>{if(!root.open&&busy){withdraw();status.textContent='Read stopped. Read groups explicitly to try again.';}});
 return {root,set,sync(){read.disabled=busy||!source||!alive();if(!alive()&&busy)withdraw();},invalidate(){withdraw();source=null;read.disabled=true;status.textContent='';},dispose(){disposed=true;withdraw();root.remove();}};
}
