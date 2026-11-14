/* Explicit raw acquisition belongs to this application adapter; held display is local. */
import * as UI from './components.mjs';
import * as Attributes from './attributes.mjs';
import * as Environment from './record-environment.mjs';
import * as Evidence from './evidence-summary.mjs';
import * as Refusal from './source-refusal.mjs';
import {renderRefusal} from './source-refusal-view.mjs';
import {compactScalar} from './numeric-values.mjs';
import {nativeControlReady,completionControlReady} from './dom-visibility.mjs';
import {object,text} from './intent-shape.mjs';
const n=UI.node,version='goo/raw-gauge-observations/v1',has=(o,k)=>Object.hasOwn(o,k);
const fail=()=>{throw Error('Raw gauge response does not match this read.');};
const stable=v=>JSON.stringify(v,(_,x)=>x&&typeof x==='object'&&!Array.isArray(x)?Object.fromEntries(Object.entries(x).sort(([a],[b])=>a.localeCompare(b))):x);
export function checkedRequest(input){
 object(input,['question','instrument']);const q=input.question,i=input.instrument;object(q,['environment','environmentSelector','namespace','service','fromMs','toMs']);Environment.requireNarrow(q);
 if(!text(q.service,128,{empty:false})||!Number.isSafeInteger(q.fromMs)||!Number.isSafeInteger(q.toMs)||q.fromMs<0||q.toMs<=q.fromMs||q.toMs-q.fromMs>7200000||q.toMs>=4102444800000)fail();
 if(q.namespace!=null){object(q.namespace,['mode','value']);if(!['missing','value'].includes(q.namespace.mode)||!text(q.namespace.value??'',128)||q.namespace.mode==='missing'&&(q.namespace.value??'')!=='')fail();}
 object(i,['kind','scope','version','schemaURL','metric','unit']);if(i.kind!=='gauge'||!text(i.metric,256,{empty:false})||!text(i.scope,256)||!text(i.version,128)||!text(i.schemaURL,1024)||!text(i.unit,64))fail();
 return structuredClone({question:{...Environment.questionFields(q),...(q.namespace?{namespace:q.namespace}:{}),service:q.service,fromMs:q.fromMs,toMs:q.toMs},instrument:i});
}
function questionKey(q){return stable({...q,environment:q.environment||undefined,namespace:q.namespace?{mode:q.namespace.mode,...(q.namespace.value?{value:q.namespace.value}:{})}:undefined});}
function map(value){
 if(!value||typeof value!=='object'||![Object.prototype,null].includes(Object.getPrototypeOf(value)))fail();
 for(const key of Reflect.ownKeys(value)){const d=Object.getOwnPropertyDescriptor(value,key);if(typeof key!=='string'||!d.enumerable||!has(d,'value')||typeof d.value!=='string'||!key.isWellFormed()||!d.value.isWellFormed())fail();}
}
function timestamp(value){
 if(value===null||value===undefined)return {value:null,state:'absent'};
 if(typeof value!=='string')return {value:null,state:'invalid'};
 if(!/^(0|[1-9]\d*)$/.test(value)||BigInt(value)>9223372036854775807n)return {value,state:'invalid'};
 return {value,state:'valid'};
}
function scalarValue(raw){
 if(raw===null||raw===undefined)return {value:null,state:'absent'};
 if(typeof raw!=='number'&&typeof raw!=='string')return {value:null,state:'invalid'};
 const text=String(raw);
 if(/^nan$/i.test(text)||/^[+-]?inf(?:inity)?$/i.test(text))return {value:null,state:'nonfinite'};
 const digit='[0-9](?:_?[0-9])*',hex='[0-9a-f](?:_?[0-9a-f])*';
 const decimal=new RegExp('^[+-]?(?:'+digit+'(?:\\.(?:'+digit+')?)?|\\.'+digit+')(?:[eE][+-]?'+digit+')?$');
 let value;
 if(decimal.test(text))value=Number(text.replaceAll('_',''));
 else {
  const match=new RegExp('^([+-]?)0[xX]('+hex+'(?:\\.(?:'+hex+')?)?|\\.'+hex+')[pP]([+-]?'+digit+')$','i').exec(text);
  if(!match)return {value:null,state:'invalid'};
  const [whole,fraction='']=match[2].replaceAll('_','').split('.');let mantissa=Number.parseInt(whole||'0',16);for(let i=0;i<fraction.length;i++)mantissa+=Number.parseInt(fraction[i],16)*16**(-i-1);value=(match[1]==='-'?-1:1)*mantissa*2**Number(match[3].replaceAll('_',''));
 }
 return Number.isFinite(value)?{value,state:'finite'}:{value:null,state:'nonfinite'};
}
export function checkedResult(value,request){
 const expected=checkedRequest(request);object(value,['version','question','instrument','readAt','rows','limit']);
 const encoded=JSON.stringify(value).replace(/[<>&\u2028\u2029]/g,c=>'\\u'+c.charCodeAt(0).toString(16).padStart(4,'0'));
 if(new TextEncoder().encode(encoded).length>8*1024*1024||value.version!==version||value.limit!==10000||!Array.isArray(value.rows)||value.rows.length>value.limit||typeof value.readAt!=='string'||!/^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(?:\.\d{1,9})?Z$/.test(value.readAt)||!Number.isFinite(Date.parse(value.readAt))||new Date(value.readAt).toISOString().slice(0,19)!==value.readAt.slice(0,19))fail();
 const echo=checkedRequest({question:value.question,instrument:value.instrument});if(questionKey(echo.question)!==questionKey(expected.question)||stable(echo.instrument)!==stable(expected.instrument))fail();const keys=new Set();
 for(const row of value.rows){
  object(row,['occurrenceKey','stream','startNS','endNS','startState','rawValue','value','valueState','flags','flagsState','rawRow','diagnostics']);if(!text(row.occurrenceKey,128,{empty:false})||keys.has(row.occurrenceKey))fail();keys.add(row.occurrenceKey);
  const s=row.stream,i=expected.instrument,q=expected.question;object(s,['resourceSchemaURL','resource','scope','version','schemaURL','name','unit','attributes']);map(s.resource);map(s.attributes);
  if(typeof s.resourceSchemaURL!=='string'||s.name!==i.metric||s.unit!==i.unit||s.scope!==i.scope||s.version!==i.version||s.schemaURL!==i.schemaURL||s.resource['service.name']!==q.service||!Environment.matches(Environment.requireNarrow(q),s.resource))fail();
  if(q.namespace&&(q.namespace.mode==='missing'?has(s.resource,'service.namespace'):!has(s.resource,'service.namespace')||s.resource['service.namespace']!==(q.namespace.value||'')))fail();
  if(!row.rawRow||typeof row.rawRow!=='object'||Array.isArray(row.rawRow)||!Array.isArray(row.diagnostics))fail();const raw=row.rawRow;
  for(const [wire,key]of [['resource','resource'],['attributes','attributes'],['resourceSchemaURL','resource_schema_url'],['scope','scope'],['version','version'],['schemaURL','schema_url'],['name','name'],['unit','unit']])if(!has(raw,key)||stable(s[wire])!==stable(raw[key]))fail();
  const end=timestamp(raw.end_ns);if(end.state!=='valid'||row.endNS!==end.value||BigInt(row.endNS)<BigInt(q.fromMs)*1000000n||BigInt(row.endNS)>=BigInt(q.toMs)*1000000n)fail();
  const diagnostics=[],add=(field,code)=>diagnostics.push({field,code}),start=timestamp(raw.start_ns);if(start.state!=='valid')add('startNS',start.state);else if(BigInt(start.value)>BigInt(end.value)){start.state='invalid';add('startNS','after-observation');}
  if(row.startState!==start.state||row.startNS!==start.value)fail();
  if(stable(row.rawValue??null)!==stable(raw.value??null))fail();const scalar=scalarValue(raw.value);if(scalar.state!=='finite')add('value',scalar.state);
  // JSON.parse may round a fractional flags token to an integer. Retain the
  // source adapter's conservative invalid state; rounded bits cannot prove flags.
  let flags=null,flagsState='absent';if(raw.flags!==null&&raw.flags!==undefined){if(!Number.isInteger(raw.flags)||raw.flags<0||raw.flags>4294967295||row.flagsState==='invalid'&&typeof raw.flags==='number')flagsState='invalid';else {flags=raw.flags;flagsState='valid';if((BigInt(flags)&~1n)!==0n)add('flags','unknown-bits');if((flags&1)!==0){scalar.value=null;scalar.state='absent';add('value','no-recorded-value');}}}
  if(flagsState!=='valid')add('flags',flagsState);
  if(row.valueState!==scalar.state||row.value!==scalar.value||row.flagsState!==flagsState||row.flags!==flags||stable(row.diagnostics)!==stable(diagnostics))fail();
 }
 return structuredClone(value);
}
function instance(row){const r=row.stream.resource;return has(r,'service.instance.id')?(r['service.instance.id']||'Empty instance label'):'Instance not reported';}
function scalar(row){const amount=row.valueState==='finite'?(compactScalar(row.rawValue)||compactScalar(row.value)):null;return amount?amount.text+(row.stream.unit?' '+row.stream.unit:''):'Unavailable · '+row.valueState;}
function renderRows(host,data,{active=()=>true}={}){
 const controls=n('div',undefined,'controls'),label=n('label','Find held observations '),find=n('input'),count=n('p'),previous=n('button','Previous observations'),next=n('button','More observations'),viewport=n('goo-table'),table=n('table'),head=n('thead'),header=n('tr'),body=n('tbody'),panel=n('section');find.type='search';find.setAttribute('aria-label','Find held observations');label.append(find);count.setAttribute('role','status');previous.type=next.type='button';controls.append(label,previous,next,count);viewport.setAttribute('density','compact');viewport.setAttribute('mobile-layout','stacked');viewport.setAttribute('aria-label','Raw gauge observations');for(const name of ['Observed UTC','Value','Reported instance','Details']){const th=n('th',name);th.scope='col';header.append(th);}head.append(header);table.append(n('caption','Exported gauge observations · repeated rows retained'),head,body);viewport.append(table);host.replaceChildren(controls,viewport,panel);const inspector=UI.inspector(panel),rows=data.rows.map((row,index)=>({row,index,text:stable(row).toLocaleLowerCase()}));let offset=0,revision=0;
 function draw(){const epoch=++revision,query=find.value.toLocaleLowerCase(),matched=rows.filter(row=>row.text.includes(query));offset=Math.min(offset,Math.max(0,Math.floor((matched.length-1)/100)*100));body.replaceChildren();
  for(const {row,index}of matched.slice(offset,offset+100)){const tr=n('tr'),time=n('td'),value=n('td'),producer=n('td',instance(row)),action=n('td'),inspect=n('button','Inspect');tr.dataset.occurrenceKey=row.occurrenceKey;time.append(Evidence.timestamp(row.endNS));value.textContent=scalar(row)+(row.stream.unit?'':' · unit not reported');inspect.type='button';inspect.setAttribute('aria-label','Inspect observation '+(index+1));inspect.onclick=()=>{if(epoch!==revision||!active()||!nativeControlReady(inspect))return;const content=n('div');content.append(UI.stats([{label:'Observation time',value:Evidence.unixNs(row.endNS)?.exact||'Unavailable'},{label:'Reported start',value:row.startNS==='0'?'Not reported · raw 0':row.startState==='valid'?Evidence.unixNs(row.startNS)?.exact||row.startNS:'Unavailable · '+row.startState},{label:'Raw value',value:has(row,'rawValue')?JSON.stringify(row.rawValue):'Not reported'},{label:'Unit',value:row.stream.unit||'Not reported'},{label:'Value state',value:row.valueState},{label:'Flags',value:row.flagsState==='valid'?String(row.flags):'Unavailable · '+row.flagsState},{label:'Source read',value:data.readAt}]));content.append(n('p','Occurrence '+(index+1)+' in this read; no continuous producer identity is inferred.','evidence-note'));Attributes.mountStream(content,row.stream);if(row.diagnostics.length){const list=n('ul');for(const d of row.diagnostics)list.append(n('li',d.field+': '+d.code));content.append(UI.disclosure('Diagnostics',[list]));}content.append(UI.disclosure('Raw source row',[n('pre',JSON.stringify(row.rawRow,null,2))]));inspector.show({key:row.occurrenceKey,title:'Gauge observation '+(index+1),description:row.stream.name,content,opener:inspect});};action.append(inspect);tr.append(time,value,producer,action);[...tr.children].forEach((cell,i)=>cell.dataset.label=header.children[i].textContent);body.append(tr);}
  previous.hidden=next.hidden=matched.length<=100;previous.disabled=offset===0;next.disabled=offset+100>=matched.length;count.textContent=(matched.length?offset+1:0)+'–'+Math.min(offset+100,matched.length)+' of '+matched.length+' matching · '+rows.length+' held observations'+(!matched.length&&rows.length?' · No matches':'');
 }
 find.oninput=()=>{if(!active()||!nativeControlReady(find))return;offset=0;draw();};previous.onclick=()=>{if(!active()||!nativeControlReady(previous))return;offset-=100;draw();};next.onclick=()=>{if(!active()||!nativeControlReady(next))return;offset+=100;draw();};draw();if(!rows.length)host.append(n('p','No exported observations returned for this instrument and fixed window.'));
}

export function mount(host,{active=()=>true,post}={}){
 let generation=0,controller=null,lease=null,held=null,readButton=null,lastOpener=null;const section=n('section'),heading=n('h3','Raw gauge observations'),controls=n('div',undefined,'controls'),cancel=n('button','Cancel read'),close=n('button','Close observations'),status=n('p'),receipt=n('div'),display=n('div');section.dataset.rawGaugeObservations='';section.hidden=true;status.setAttribute('role','status');cancel.type=close.type='button';cancel.hidden=true;controls.append(cancel,close);section.append(heading,controls,status,receipt,display);host.append(section);
 function stopRead(){generation++;controller?.abort();controller=null;if(readButton)readButton.disabled=false;readButton=null;cancel.hidden=true;}
 function retire(){stopRead();lease=null;held=null;section.hidden=true;display.replaceChildren();receipt.replaceChildren();status.replaceChildren();}
 // Navigation ends acquisition authority, not the fixed observations already held.
 function suspend(){if(!held){retire();return;}const pending=controller!==null;stopRead();if(pending)status.textContent='Read canceled by navigation. Previous held observations retained.';}
 cancel.onclick=()=>{if(!nativeControlReady(cancel))return;stopRead();status.textContent=held?'Read canceled. Previous held observations retained.':'Read canceled. No observations adopted.';if(active()&&nativeControlReady(lastOpener))lastOpener.focus({preventScroll:true});};close.onclick=()=>{if(!nativeControlReady(close))return;const origin=lastOpener;retire();if(active()&&nativeControlReady(origin))origin.focus({preventScroll:true});};
 async function read(request,opener){if(!active()||!nativeControlReady(opener))return;let checked;try{checked=checkedRequest(request);}catch(error){return;}const identity=stable(checked);if(lease!==identity){retire();lease=identity;}controller?.abort();const current=++generation,abort=new AbortController(),followFocus=document.activeElement===opener;controller=abort;readButton=opener;lastOpener=opener;opener.disabled=true;section.hidden=false;cancel.hidden=false;heading.textContent='Raw observations · '+checked.instrument.metric;status.textContent='Reading bounded gauge observations…';if(followFocus){heading.tabIndex=-1;heading.scrollIntoView({block:'start',behavior:'instant'});heading.focus({preventScroll:true});}
  try{const reply=await post('/api/metric-observations',checked,abort.signal);if(current!==generation)return;if(!active()||!completionControlReady(opener)||!completionControlReady(section)){retire();return;}if(!reply.ok)throw Refusal.responseError(reply,'Gauge observations unavailable.');let data;try{data=checkedResult(reply.data,checked);}catch{throw new Refusal.SourceReadError({kind:'malformed',detail:'The raw observation response did not match its fixed read.'});}held=data;receipt.replaceChildren(UI.stats([{label:'Source read UTC',value:data.readAt},{label:'Fixed window',value:new Date(checked.question.fromMs).toISOString()+' → '+new Date(checked.question.toMs).toISOString()}]));renderRows(display,data,{active:()=>active()&&held===data&&lease===identity});status.textContent=data.rows.length+' exported observations · no evaluation or completeness claim.';
  }catch(error){if(current!==generation)return;if(!active()||!completionControlReady(opener)||!completionControlReady(section)){retire();return;}const failure=error instanceof Refusal.SourceReadError?error:new Refusal.SourceReadError({kind:error?.name==='AbortError'?'cancelled':error instanceof SyntaxError?'malformed':'unavailable',detail:'The raw observation read could not be completed.'});renderRefusal(status,failure,{prefix:held?'Previous held observations retained. ':''});
  }finally{if(current===generation){controller=null;opener.disabled=false;readButton=null;cancel.hidden=true;}}
 }
 return {read,retire,suspend};
}
