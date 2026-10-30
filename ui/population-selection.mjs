import {durationHistogramEdges} from './view-model.mjs';
/* Typed question transforms only. Go owns evaluation and aggregation. */
import * as Discovery from './discovery.mjs';
import * as Distributions from './distributions.mjs';
import * as Explanation from './explanation.mjs';
import * as QuestionSession from './question-session.mjs';
import * as UI from './components.mjs';
import * as Workspace from './workspace.mjs';
const global=globalThis;
const api=name=>({"GooWorkspace":Workspace,"GooDiscovery":Discovery})[name]||(null);
function question(base,{fromMs=base.fromMs,toMs=base.toMs,minMS='',maxMS='',values=[]}={}){
  api('GooWorkspace').validateRecordQuestion(base);
  if(!Number.isSafeInteger(fromMs)||!Number.isSafeInteger(toMs)||fromMs<base.fromMs||toMs>base.toMs||fromMs>=toMs)throw Error('Selection must stay inside the applied half-open window.');
  if((minMS!==''||maxMS!=='')&&base.signal!=='spans')throw Error('Duration selection requires span records.');
  const leaves=[];let low=null,high=null;if(minMS!==''){low=BigInt(api('GooDiscovery').millisecondsToNS(minMS));leaves.push({field:{kind:'column',key:'duration_ns'},op:'gte',value:String(low)})}if(maxMS!==''){high=BigInt(api('GooDiscovery').millisecondsToNS(maxMS));if(high===0n||low!==null&&high<=low)throw Error('Duration upper bound must exceed the lower bound.');leaves.push({field:{kind:'column',key:'duration_ns'},op:'lte',value:String(high-1n)})}
  for(const {field,value}of values)leaves.push({field:structuredClone(field),op:value===null?'missing':'eq',...(value===null?{}:{value})});
  const q={...structuredClone(base),fromMs,toMs,offset:0};if(leaves.length)q.where={all:[...(base.where?[structuredClone(base.where)]:[]),...leaves]};api('GooWorkspace').validateRecordQuestion(q);return q;
 }
// Only a generated time/duration region is complemented, never arbitrary parent predicates.
function partition(base,selection={}){
  if(base.signal!=='spans'||base.numeric)throw Error('Rest comparison requires raw span durations.');
  if(selection.values?.length)throw Error('Use a time/duration rectangle or clear the selected chart value before comparing with rest.');
  const narrowed=question(base,selection),bare=structuredClone(base);delete bare.where;
  const duration=question(bare,selection).where?.all||[];
  const leaf=(key,op,value)=>({field:{kind:'column',key},op,value:String(value)});
  const intent=[leaf('timestamp_ns','gte',BigInt(narrowed.fromMs)*1000000n),leaf('timestamp_ns','lte',BigInt(narrowed.toMs)*1000000n-1n),...duration];
  const outside=[];
  for(const p of intent){const v=BigInt(p.value);if(p.op==='gte'&&v>0n)outside.push(leaf(p.field.key,'lte',v-1n));else if(p.op==='lte'&&v<18446744073709551615n)outside.push(leaf(p.field.key,'gte',v+1n));}
  function make(predicate){const q={...structuredClone(base),offset:0,where:base.where?{all:[structuredClone(base.where),predicate]}:predicate};let count=0;function check(p,d=0){if(++count>32||d>4)throw Error('Region comparison exceeds the predicate budget. Simplify the parent filters.');for(const c of p.all||p.any||[])check(c,d+1)}check(q.where);api('GooWorkspace').validateRecordQuestion(q);return q}
  return {selected:make({all:intent}),rest:make({any:outside})};
 }
function rectangle(data,range){
  const order=api('GooDiscovery').durationOrder;
  if(data.question.signal!=='spans'||data.question.groupBy?.length!==1||data.question.groupBy[0].kind!=='column'||data.question.groupBy[0].key!=='duration_bucket'||data.state==='changed_between_reads')throw Error('Rectangle requires a consistent raw-span duration breakdown.');
  const points=data.timeSeries,keys=['fromColumn','toColumn','fromRow','toRow'];if(!points?.length||keys.some(k=>!Number.isInteger(range[k]))||range.fromColumn<0||range.toColumn>=points.length||range.fromColumn>range.toColumn||range.fromRow<0||range.toRow>=order.length||range.fromRow>range.toRow)throw Error('Invalid heatmap rectangle.');
  const ms=value=>value==='+inf'?'':String(BigInt(value)/1000000n);return {fromMs:points[range.fromColumn].fromMs,toMs:points[range.toColumn].toMs,minMS:ms(durationHistogramEdges[range.fromRow]),maxMS:ms(durationHistogramEdges[range.toRow+1])};
 }
function mount(host,data,{open,onBaseline,compact=false,canAct=()=>true}={}){
  const node=UI.node,details=node('details',undefined,'population-selection'),summary=node('summary','Investigate a chart selection'),form=node('form'),controls=node('div',undefined,'controls'),preview=node('p'),run=node('button','Run selected population'),compare=node('button','Compare selection to this baseline'),rest=node('button','Compare with rest');rest.type='button';preview.setAttribute('role','status');const exact=node('details'),predicates=node('code');exact.append(node('summary','Exact selected predicates'),predicates);form.append(controls,preview,exact,run,compare,rest);details.append(summary,form);host.append(details);let selected=null,pair=null,explanationOwner=null,paused=false;const allowed=()=>!paused&&details.isConnected&&canAct();const explanationHost=node('div');details.append(explanationHost);
  const fields={};for(const [key,title,value]of [['from','From (UTC)',new Date(data.question.fromMs).toISOString().slice(0,23)],['to','To (UTC, exclusive)',new Date(data.question.toMs).toISOString().slice(0,23)],...(data.question.signal==='spans'?[['min','Minimum span duration (ms)',''],['max','Maximum span duration (ms, exclusive)','']]:[])]){const label=node('label',title),input=node('input');input.type=['from','to'].includes(key)?'datetime-local':'text';if(input.type==='datetime-local')input.step='0.001';else input.inputMode='decimal';input.value=value;input.setAttribute('aria-label',title);label.append(input);controls.append(label);fields[key]=input}
  if(compact){const actions=node('div',undefined,'controls');actions.append(run,rest,compare);exact.querySelector('summary').textContent='Selection bounds and predicates';exact.insertBefore(controls,predicates);form.replaceChildren(preview,actions,exact)}
  function update(){explanationOwner?.dispose?.();explanationHost.replaceChildren();const bounds={fromMs:Date.parse(fields.from.value+'Z'),toMs:Date.parse(fields.to.value+'Z'),minMS:fields.min?.value||'',maxMS:fields.max?.value||'',values:details._values||[]};pair=null;rest.disabled=true;try{selected=question(data.question,{fromMs:Date.parse(fields.from.value+'Z'),toMs:Date.parse(fields.to.value+'Z'),minMS:fields.min?.value||'',maxMS:fields.max?.value||'',values:details._values||[]});preview.textContent='Preview · ['+new Date(selected.fromMs).toISOString()+', '+new Date(selected.toMs).toISOString()+')'+(fields.min?.value||fields.max?.value?' · span duration ['+(fields.min?.value||'0')+', '+(fields.max?.value||'+∞')+') ms':'');predicates.textContent=QuestionSession.selectorText(selected.where);run.disabled=!open||!allowed();compare.disabled=!open||!allowed()||!onBaseline||!!data.numeric||data.question.groupBy?.length!==1||data.state==='changed_between_reads';try{pair=partition(data.question,bounds);rest.disabled=compare.disabled;rest.title='Stage a disjoint region and remainder in the parent window. Run reads both independently.'}catch(e){rest.title=e.message}if(pair)explanationOwner=Explanation?.mount(explanationHost,{title:'Explain this selection',request:{version:'goo/explanation/v1',baseline:pair.rest,selection:pair.selected},available:data.fields||[],canAct:()=>allowed()&&details.getClientRects().length>0});else if(selected&&data.question.signal==='logs')explanationOwner=Explanation?.mount(explanationHost,{title:'Explain selection versus parent',note:'Selection and parent populations overlap; this is not a remainder comparison.',request:{version:'goo/explanation/v1',baseline:structuredClone(data.question),selection:structuredClone(selected)},available:data.fields||[],canAct:()=>allowed()&&details.getClientRects().length>0});}catch(e){selected=null;preview.textContent=e.message;run.disabled=compare.disabled=true}}
  rest.onclick=()=>{if(allowed()&&pair&&!rest.disabled)open(structuredClone(pair.selected),null,structuredClone(pair.rest))};form.oninput=update;form.onsubmit=e=>{e.preventDefault();if(allowed()&&selected&&!run.disabled)open(structuredClone(selected))};compare.type='button';compare.onclick=()=>{if(allowed()&&selected&&!compare.disabled)open(structuredClone(selected),Distributions.snapshot(data,true))};update();
  return {suspend(){paused=true;explanationOwner?.invalidate?.();run.disabled=compare.disabled=rest.disabled=true;preview.textContent='Run again to use this selection.'},invalidate(){paused=true;explanationOwner?.dispose?.();run.disabled=compare.disabled=rest.disabled=true},select(point,values=[],{reveal=true}={}){if(!allowed())return;fields.from.value=new Date(point.fromMs).toISOString().slice(0,23);fields.to.value=new Date(point.toMs).toISOString().slice(0,23);details._values=values;if(fields.min)fields.min.value=point.minMS||'';if(fields.max)fields.max.value=point.maxMS||'';if(reveal)details.open=true;update()},element:details};
 }

export {question,partition,rectangle,mount};
