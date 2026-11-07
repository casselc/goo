import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Triage from '../ui/error-triage.mjs';
import * as Protocol from '../ui/question-protocol.mjs';
import {mounted,event,admission} from './support/typed-error-triage-370.mjs';

const field=(kind,key)=>({kind,key}),clause=(field,op,value)=>({field,op,...(value===undefined?{}:{value})});
const selectors=[{policy:'current-only',scope:{mode:'named',value:'prod'}},{policy:'effective',scope:{mode:'named',value:'prod'}},{policy:'effective',scope:{mode:'empty'}},{policy:'effective',scope:{mode:'unreported'}}].map(s=>({version:'goo/environment-selector/v1',...s}));
const parent=(selector=selectors[0],signal='spans',namespace='team-π')=>({signal,environmentSelector:structuredClone(selector),fromMs:1000,toMs:61000,limit:17,offset:9,where:{all:[clause(field('resource','service.namespace'),namespace===null?'missing':'eq',namespace??undefined),{any:[clause(field('attribute','tenant'),'eq','inside'),clause(field('attribute','tenant'),'eq','inside-2')]}]}});
const shown=()=>({question:Triage.triageQuery(parent()),groups:[{g0:'api',g1:'Failure',count:'2'},{g0:'api',g1:null,count:'1'},{g0:'api',g1:'',count:'1'}],state:'observed'});
const actions=(data,current=()=>true)=>[{label:'Inspect group',href:g=>Protocol.questionLink(Triage.inspectGroup(data.question,g)),current:g=>current()&&data.groups.includes(g)},{label:'Compare within service',href:g=>Protocol.questionLink(Triage.compareInService(data.question,g.g0)),current:g=>current()&&data.groups.includes(g)}];

test('triage continuations retain all typed policies, namespaces and nested parent predicates',()=>{
 for(const selector of selectors)for(const signal of ['logs','spans'])for(const namespace of [null,'','team-π'])for(const service of [null,'','api"雪'])for(const type of [null,'','Failure"雪']){
  const source=parent(selector,signal,namespace),before=structuredClone(source),q=Triage.triageQuery(source),group={g0:service,g1:type},inspect=Triage.inspectGroup(q,group),compare=Triage.compareInService(q,service);
  for(const child of [inspect,compare]){assert.deepEqual(child.environmentSelector,selector);assert.equal(child.fromMs,source.fromMs);assert.equal(child.toMs,source.toMs);assert.equal(child.bucketMs,q.bucketMs);assert.equal(child.offset,0);assert.equal(child.limit,100);assert.deepEqual(child.where.all.slice(0,q.where.all.length),q.where.all);Protocol.validateRecordQuestion(child)}
  assert.deepEqual(compare.where.all.at(-1),clause(field('column','service'),service===null?'missing':'eq',service??undefined));assert.deepEqual(inspect.where.all.at(-2),compare.where.all.at(-1));assert.deepEqual(inspect.where.all.at(-1),clause(q.groupBy[1],type===null?'missing':'eq',type??undefined));assert.deepEqual(inspect.groupBy,[q.groupBy[1]]);assert.deepEqual(source,before);
 }
});

test('mounted current anchors preserve ordinary and modifier hrefs without acquiring data',()=>{
 const oldFetch=globalThis.fetch;globalThis.fetch=()=>{throw Error('Group staging must not read source')};
 try{const data=shown(),before=structuredClone(data),m=mounted(data,actions(data));assert.equal(m.anchors().length,6);for(const [index,anchor]of m.anchors().entries()){assert.equal(anchor.localName,'a');assert.match(anchor.href,/^#step=/);for(const options of [{},{ctrlKey:true},{metaKey:true},{shiftKey:true},{button:1}]){const e=event(options);assert.equal(anchor.onclick(e),true);assert.equal(e.defaultPrevented,false);assert.equal(e.stopped,false)}assert.equal(anchor.textContent,index%2?'Compare within service':'Inspect group')}assert.deepEqual(data,before)}finally{globalThis.fetch=oldFetch}
});

test('actual fixed-owner admission requires a coherent current applied result and an original group',()=>{
 const data=shown(),a=admission(data),controls=a.actions(),group=data.groups[0];assert.equal(controls.length,2);assert.equal(controls[0].current(group),true);assert.match(controls[0].href(group),/^#step=/);assert.equal(controls[0].current({...group}),false);assert.throws(()=>controls[0].href({...group}));
 for(const [object,key,value]of [[a.env,'triageGroupLive',false],[a.env,'activeFlag',false],[a.env,'refinable',false],[a.env,'population',true],[a.env.container,'isConnected',false],[a.state,'phase','running'],[a.state,'generation',8],[a.env,'lastResult',structuredClone(data)]]){const old=object[key];object[key]=value;assert.equal(controls[0].current(group),false,key);assert.throws(()=>controls[0].href(group),undefined,key);object[key]=old}
 a.env.readOnly=true;assert.equal(a.actions().length,0);a.env.readOnly=false;for(const state of ['changed_between_reads','unavailable','loading']){data.state=state;assert.equal(a.actions().length,0)}data.state='observed';const old=data.question.groupBy;data.question.groupBy=[old[1],old[0]];assert.equal(a.actions().length,0);data.question.groupBy=old;data.groups.splice(0,1);assert.equal(controls[0].current(group),false);
});

test('suspend and resume renew only current anchors while preserving held DOM and focus',()=>{
 const data=shown();let active=true;const m=mounted(data,actions(data,()=>active)),anchor=m.anchors()[0],old=anchor.onclick,children=m.host.children;anchor.focus();m.owner.suspendActions();active=false;const retired=event({ctrlKey:true});assert.equal(old(retired),false);assert(retired.defaultPrevented&&retired.stopped);assert.equal(m.host.children,children);active=true;assert.equal(m.owner.resumeActions(),true);assert.equal(m.anchors()[0],anchor);assert.equal(document.activeElement,anchor);assert.equal(old(event()),false);assert.equal(anchor.onclick(event()),true);assert.equal(m.host.children,children);
 const previous=anchor.onclick;m.owner.suspendActions();m.owner.resumeActions();assert.equal(previous(event()),false);assert.equal(anchor.onclick(event()),true);
});

test('hidden, inert, detached, unowned and superseded anchors refuse before delegated navigation',()=>{
 const data=shown();let active=true;const m=mounted(data,actions(data,()=>active)),anchor=m.anchors()[0],run=()=>{const e=event();assert.equal(anchor.onclick(e),false);assert(e.defaultPrevented&&e.stopped)};
 m.host.hidden=true;run();m.host.hidden=false;m.host.inert=true;run();m.host.inert=false;active=false;run();active=true;anchor.remove();run();m.host.append(anchor);assert.equal(anchor.onclick(event()),true);
 const held=m.anchors()[1],saved=held.onclick;m.owner.suspendActions();m.host.replaceChildren();m.host.append(held);const stale=event();assert.equal(saved(stale),false);assert(stale.defaultPrevented&&stale.stopped);
 const absent=shown(),n=mounted(absent,actions(absent));absent.groups.splice(0,1);const unavailable=event();assert.equal(n.anchors()[0].onclick(unavailable),false);assert(unavailable.defaultPrevented);
});

test('disabled or unrepresentable group actions stay non-links and cannot manufacture groups',()=>{
 const data=shown(),m=mounted(data,[{...actions(data)[0],disabled:true},{label:'Unavailable',href(){throw Error('Unrepresentable question')}}]);assert.equal(m.anchors().length,0);assert.match(m.host.textContent,/Inspect group/);assert.match(m.host.textContent,/Unavailable/);const empty=mounted({...data,groups:[]},actions(data));assert.equal(empty.anchors().length,0);
});

test('mixed local owner schedules never revive a saved callback or issue source reads',()=>{
 const oldFetch=globalThis.fetch;globalThis.fetch=()=>{throw Error('Local group ownership attempted a read')};
 try{for(let seed=1;seed<=16;seed++){const data=shown();let active=true,paused=false,n=seed;const m=mounted(data,actions(data,()=>active)),anchor=m.anchors()[0],retired=[];
  for(let step=0;step<30;step++){n=(n*1664525+1013904223)>>>0;switch(n%5){case 0:retired.push(anchor.onclick);m.owner.suspendActions();paused=true;break;case 1:active=true;m.owner.resumeActions();paused=false;break;case 2:active=!active;break;case 3:m.host.hidden=!m.host.hidden;break;case 4:{const e=event();assert.equal(anchor.onclick(e),active&&!paused&&!m.host.hidden);assert.equal(e.defaultPrevented,!(active&&!paused&&!m.host.hidden));break}}for(const fn of retired){const e=event();assert.equal(fn(e),false);assert(e.defaultPrevented)}}
 }}finally{globalThis.fetch=oldFetch}
});
