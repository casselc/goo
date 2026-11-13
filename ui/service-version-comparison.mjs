/* Guided resource-version cohorts; query intent and application-owned reads only. */
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
import {identity} from './question-session.mjs';
import * as Environment from './record-environment.mjs';
import * as ServiceIdentity from './service-identity.mjs';
import * as Aggregate from './record-aggregate-protocol.mjs';
import * as Comparison from './captured-count-comparison.mjs';
import * as Declaration from './view-declaration.mjs';
import * as Composition from './view-composition.mjs';
import * as UI from './components.mjs';
import * as Queries from './queries.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
const clone=structuredClone,versionField={kind:'resource',key:'service.version'};

function parent(question){
 validateRecordQuestion(question);if(question.where)validatePredicate(question.where,question.signal);
 if(question.signal!=='spans'||(question.offset??0)!==0||question.numeric||question.numericHistogram||question.reportedEdge||question.groupTimeSeries||question.groupBy?.length!==1||question.groupBy[0].kind==='column'&&['duration_bucket','body_pattern_v1'].includes(question.groupBy[0].key))throw Error('Choose a page-zero ordinary span breakdown.');
 const q=clone(question);q.limit=100;q.offset=0;for(const key of ['summary','spanStatus','bucketMs'])delete q[key];return q;
}
function projected(question){const q=parent(question);q.groupBy=[clone(versionField)];q.limit=1;return q;}
export function discoveryQuestion(widget,data,{service,scope}={}){
 if(!widget?.query||widget.measurement||!data?.question||!['observed','no_data'].includes(data.state)||identity(widget.query)!==identity(data.question))throw Error('Read a coherent span widget before discovering versions.');
 const q=projected(widget.query),ns=scope?.namespace;
 if(!ns||!['missing','value'].includes(ns.mode))throw Error('Choose one qualified service namespace.');
 const leaves=p=>!p||p.any?[]:p.all?p.all.flatMap(leaves):[p],required=[{field:{kind:'column',key:'service'},op:'eq',value:service},ServiceIdentity.predicate(ns)];
 if(!scope||q.fromMs!==scope.fromMs||q.toMs!==scope.toMs||!Environment.sameQuestionEnvironment(q,scope)||required.some(expected=>!leaves(q.where).some(p=>p.field?.kind===expected.field.kind&&p.field.key===expected.field.key&&p.op===expected.op&&(p.value??'')===(expected.value??''))))throw Error('Widget differs from the selected qualified service scope.');
 return q;
}
export function admit(question,data){
 Aggregate.recordAggregates(question,data);
 if(question.signal!=='spans'||question.groupBy?.length!==1||question.groupBy[0].kind!=='resource'||question.groupBy[0].key!=='service.version'||data.state==='changed_between_reads')throw Error('Version discovery requires coherent complete groups.');
 return data;
}
export function comparison(parentQuestion,discovery,baselineValue,selectedValue,measure='count'){
 const q=parent(parentQuestion),discoveryQuery=projected(parentQuestion);admit(discoveryQuery,discovery);
 if(!['count','duration.mean','status.error','status.error_fraction'].includes(measure)||baselineValue===selectedValue||![baselineValue,selectedValue].every(value=>discovery.groups.some(g=>g.g0===value)))throw Error('Choose two distinct discovered versions and a supported measure.');
 function cohort(value){const out=clone(q),leaf={field:clone(versionField),op:value===null?'missing':'eq',...(value===null?{}:{value})};out.where={all:[...(q.where?[clone(q.where)]:[]),leaf]};if(measure==='duration.mean')out.summary='span-duration';if(measure.startsWith('status.'))out.spanStatus=true;validateRecordQuestion(out);validatePredicate(out.where,out.signal);return out;}
 const selected=cohort(selectedValue),baseline=cohort(baselineValue),declaration=Comparison.createQuery(selected,baseline,measure);
 const short=value=>{let text=valueLabel(value);if(new TextEncoder().encode(text).length<=40)return text;while(new TextEncoder().encode(text+'…').length>40)text=Array.from(text).slice(0,-1).join('');return text+'…';};
 const title='Comparison '+short(selectedValue)+' · baseline '+short(baselineValue);declaration.display.title=title;if(measure.startsWith('status.'))declaration.display.view='summary';
 const composition=Composition.setComparison(Composition.create(selected,Declaration.create(selected)),declaration,selected);
 return Composition.view({version:1,title,queries:[selected],composition});
}
export function choiceReason(value){try{validatePredicate({field:versionField,op:value===null?'missing':'eq',...(value===null?{}:{value})},'spans');return '';}catch(error){return error.message;}}
export const valueLabel=value=>value===null?'Not reported':value===''?'Present empty':JSON.stringify(value);

export function mount(host,{widget,data,context,signal,current=()=>true,ready=()=>true,onPendingChange=()=>{},onStage}={}){
 const query=discoveryQuestion(widget,data,context),n=UI.node,root=n('details',undefined,'service-version-comparison'),summary=n('summary','Compare versions'),controls=n('div',undefined,'controls'),discover=n('button','Discover versions'),stage=n('button','Stage version comparison'),status=n('p'),pair=n('div',undefined,'controls');
 const readDetails=UI.disclosure('Read details'),readBody=n('dl',undefined,'evidence-summary-details');readDetails.append(readBody,n('p','Independent count and group reads; this is not the earlier widget receipt. Reported versions and captured differences do not establish deployment causality.','evidence-note'));readDetails.hidden=true;
 root.append(summary,controls,pair,status,readDetails);controls.append(discover);status.setAttribute('role','status');stage.disabled=true;pair.hidden=true;root.dataset.state='initial';host.append(root);
 let generation=0,controller=null,snapshot=null,busy=false,retired=false;
 const valid=()=>!retired&&!signal?.aborted&&current()&&root.isConnected&&host.contains(root),action=control=>valid()&&root.open&&nativeControlReady(control);
 const baseline=n('select'),selection=n('select'),measure=n('select');baseline.setAttribute('aria-label','Baseline version');selection.setAttribute('aria-label','Comparison version');measure.setAttribute('aria-label','Version comparison measure');
 for(const [value,label]of [['count','Captured count'],['duration.mean','Mean span duration'],['status.error','Recorded Error count'],['status.error_fraction','Recorded Error fraction']]){const option=n('option',label);option.value=value;measure.append(option);}measure.value='count';for(const [control,text]of [[baseline,'Baseline version'],[selection,'Comparison version'],[measure,'Measure']]){const label=n('label',text);label.append(control);pair.append(label);}pair.append(stage);
 function update(){stage.disabled=busy||!snapshot||!baseline.value||!selection.value||baseline.value===selection.value||!ready();discover.disabled=busy||!ready();}
 function cancel(){generation++;controller?.abort();controller=null;if(busy){busy=false;onPendingChange(false);}update();}
 function retire(){retired=true;cancel();}
 signal?.addEventListener('abort',retire,{once:true});
 root.addEventListener('toggle',()=>{if(!root.open&&busy){cancel();root.dataset.state='paused';status.textContent='Discovery stopped. Discover explicitly to read again.';}else update();});
 discover.onclick=async()=>{
  if(!action(discover)||busy||!ready())return;
  snapshot=null;pair.hidden=true;readDetails.hidden=true;readDetails.open=false;readBody.replaceChildren();stage.disabled=true;baseline.replaceChildren();selection.replaceChildren();busy=true;controller=new AbortController();const own=++generation;root.dataset.state='loading';status.textContent='Reading reported resource versions…';onPendingChange(true);update();
  try{const response=await Queries.postJSON('/api/record-aggregates',clone(query),signal?AbortSignal.any([signal,controller.signal]):controller.signal);if(own!==generation||!valid()||!root.open)return;if(!response.ok)throw Error(response.data?.error||'Version discovery unavailable.');snapshot=clone(admit(query,response.data));
   for(const select of [baseline,selection]){const placeholder=n('option','Choose version');placeholder.value='';select.append(placeholder);for(const [index,g]of snapshot.groups.entries()){const option=n('option',valueLabel(g.g0)+' · '+g.count+' spans');option.value=String(index+1);const reason=choiceReason(g.g0);if(reason){option.disabled=true;option.textContent+=' · cannot select';option.title=reason;}select.append(option);}}
   pair.hidden=false;root.dataset.state=snapshot.state;
   const missing=snapshot.groups.find(g=>g.g0===null)?.count||'0',empty=snapshot.groups.find(g=>g.g0==='')?.count||'0';
   status.textContent=snapshot.count+' spans · '+snapshot.groups.length+' version groups · '+missing+' unreported · '+empty+' empty · coverage unknown';
   const facts=[['Discovery completed',snapshot.capturedAt],['Count read',snapshot.reads.count.readAt],['Count basis',snapshot.reads.count.basis],['Groups read',snapshot.reads.groups.readAt],['Groups basis',snapshot.reads.groups.basis],['Discovery basis',snapshot.basis],['Fixed window · UTC','['+new Date(query.fromMs).toISOString()+', '+new Date(query.toMs).toISOString()+')'],['Environment',Environment.questionLabel(query)],['Version field','resource service.version']];
   for(const [label,value]of facts){const row=n('div');row.append(n('dt',label),n('dd',value));readBody.append(row);}
   const exact=n('div'),definition=n('pre');definition.append(n('code',JSON.stringify(query,null,2)));const value=n('dd');value.append(definition);exact.append(n('dt','Exact discovery question'),value);readBody.append(exact);readDetails.hidden=false;

  }catch(error){if(own===generation&&valid()&&error.name!=='AbortError'){snapshot=null;root.dataset.state='unavailable';status.textContent=error.message;}}
  finally{if(own===generation){busy=false;controller=null;onPendingChange(false);update();}}
 };
 for(const select of [baseline,selection,measure])select.onchange=update;
 stage.onclick=()=>{if(!action(stage)||stage.disabled||busy||!snapshot||!ready())return;try{const view=comparison(widget.query,snapshot,snapshot.groups[Number(baseline.value)-1]?.g0,snapshot.groups[Number(selection.value)-1]?.g0,measure.value);onStage?.(view,stage);}catch(error){status.textContent=error.message;}};
 update();return {root,sync:update,dispose(){retire();signal?.removeEventListener('abort',retire);root.remove();}};
}
