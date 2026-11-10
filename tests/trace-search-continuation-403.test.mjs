import assert from 'node:assert/strict';
import {test,mock} from 'node:test';
import * as Environment from '../ui/record-environment.mjs';
class Element {
 constructor(tag='div',text=''){this.localName=tag;this.textContent=text||'';this.children=[];this.attributes={};this.dataset={};this.hidden=false;this.isConnected=true;this.disabled=false;this.listeners=new Map();this.scrolls=[];this.focuses=[];}
 set textContent(value){this._text=value??'';if(this.children){for(const child of this.children)child.parentElement=null;this.children=[];}}get textContent(){return this._text+(this.children||[]).map(child=>child.textContent).join('');}
 append(...children){for(const child of children){child.parentElement=this;this.children.push(child)}}
 replaceChildren(...children){for(const child of this.children)child.parentElement=null;this.children=[];this._text='';this.append(...children)}
 setAttribute(key,value){this.attributes[key]=String(value)} getAttribute(key){return this.attributes[key]}
 closest(selector){for(let node=this;node;node=node.parentElement)if(selector==='[inert]'&&node.inert)return node;return null}
 contains(node){return node===this||this.children.some(child=>child.contains(node))}
 addEventListener(type,callback){this.listeners.set(type,callback)} removeEventListener(type){this.listeners.delete(type)}
 remove(){this.isConnected=false;if(this.parentElement)this.parentElement.children=this.parentElement.children.filter(child=>child!==this)}
 querySelectorAll(selector){return this.children.flatMap(child=>[...(selector==='details[data-condition-editor]'&&child.localName==='details'&&Object.hasOwn(child.dataset,'conditionEditor')?[child]:[]),...child.querySelectorAll(selector)])}
 scrollIntoView(options){this.scrolls.push(options)} focus(options){this.focuses.push(options);document.activeElement=this}
}
const node=(...args)=>new Element(...args),events=new EventTarget(),documentEvents=new EventTarget();
globalThis.document={hidden:false,body:node('body'),activeElement:null,addEventListener:documentEvents.addEventListener.bind(documentEvents),removeEventListener:documentEvents.removeEventListener.bind(documentEvents)};
globalThis.dispatchEvent=events.dispatchEvent.bind(events);globalThis.addEventListener=events.addEventListener.bind(events);globalThis.removeEventListener=events.removeEventListener.bind(events);
let requests=[],editors=[],saved=[];
mock.module('../ui/documents.mjs',{exports:{context:()=>null,stageTraceSearchQuestion:(step,context)=>saved.push({step:structuredClone(step),context})}});
mock.module('../ui/components.mjs',{exports:{node}});
mock.module('../ui/attributes.mjs',{exports:{isJSONValue:value=>{try{return JSON.stringify(value)!==undefined}catch{return false}}}});
mock.module('../ui/evidence-summary.mjs',{exports:{unixNs:()=>null}});
mock.module('../ui/field-presentation.mjs',{exports:{selectorText:value=>JSON.stringify(value),operator:value=>value,label:field=>field.key}});
mock.module('../ui/filter-editor.mjs',{exports:{validate(){},normalize:structuredClone,mount(_host,options){editors.push(options);return{set(){}}}}});
mock.module('../ui/numeric-distribution-comparison.mjs',{exports:{PendingScope:class{start(){return this}stop(){}}}});
mock.module('../ui/queries.mjs',{exports:{postJSON:(url,request,signal)=>new Promise(resolve=>requests.push({url,request,signal,resolve}))}});
mock.module('../ui/trace-page-breakdown.mjs',{exports:{compareLabel:(a,b)=>a<b?-1:a>b?1:0}});
mock.module('../ui/trace-page-breakdown-view.mjs',{exports:{durationText:value=>value}});
mock.module('../ui/visualizations.mjs',{exports:{actionVisible:element=>{for(let current=element;current;current=current.parentElement)if(current.hidden||!current.isConnected)return false;return true}}});
const Search=await import('../ui/trace-search.mjs');
const Position=await import('../ui/trace-search-position.mjs');
const {createHash}=await import('node:crypto');
const eq=(key,value)=>({field:{kind:'column',key},op:'eq',value}),all=element=>[element,...element.children.flatMap(all)],button=(host,text)=>all(host).find(node=>node.localName==='button'&&node.textContent===text);
function fixture(selector){
 requests=[];editors=[];let active=true,clean=true;const host=node('main'),opener=node('button','Inspect span'),opened=[],question={signal:'spans',environment:'',environmentSelector:selector,fromMs:1000,toMs:61000,where:eq('service','api'),limit:100,offset:0};
 host.append(opener);const owner=Search.mount(host,{current:()=>({question,active,clean}),onOpenTrace:(...args)=>opened.push(args)});assert(owner.open(eq('name','work'),opener));
 return {host,owner,opener,opened,question,setActive:value=>active=value,setClean:value=>clean=value};
}
const resources={[Environment.currentKey]:'prod'};
const id=n=>n.toString(16).padStart(32,'0');
function response(input,{ids=[1,2],total='5',remaining=total}={}){
 const request=input.request||input,position=input.position,fieldValues=Search.conditionFields(request.conditions[0]).map(field=>({field,value:field.kind==='resource'?(Object.hasOwn(resources,field.key)?resources[field.key]:null):field.key==='name'?'work':'api'}));
 const items=ids.map(value=>({traceId:id(value),witnesses:[{conditionIndex:0,matchingRecords:'1',spanId:'2'.repeat(16),parentSpanId:'',service:'api',operation:'work',kind:'Server',timestampNs:'1000000000',durationNs:'1',fieldValues}]})),hasMore=BigInt(remaining)>BigInt(items.length);
 return {ok:true,data:{request,...(position?{position}:{}),remainingTraces:remaining,readAt:'2026-11-10T00:00:00Z',basis:'a'.repeat(64),plan:{sql:'SELECT',parameters:{}},warnings:[],state:remaining==='0'?'no_data':'observed',matchedTraces:total,hasMore,items,...(hasMore?{nextPosition:Position.makePosition(request,items.at(-1).traceId)}:{})}};
}
function ready(){const f=fixture(Environment.current('prod'));assert(f.owner.openRequest(Search.buildRequest(f.question,[eq('service','api')],2)));return f}
async function readFirst(f){const pending=button(f.host,'Run trace search').onclick(),read=requests.at(-1);read.resolve(response(read.request));await pending;return read}

test('403 bounded synchronous SHA256 agrees with standard UTF8 vectors',()=>{
 for(const value of ['', 'abc', '雪😀<>&\u2028\u2029', 'a'.repeat(65536)])assert.equal(Position.sha256(value),createHash('sha256').update(value).digest('hex'));
 const f=ready(),request=Search.buildRequest(f.question,[eq('service','api')],2),before=structuredClone(request),p=Position.makePosition(request,id(2));assert.equal(Position.validatePosition(p,request),p);assert.deepEqual(request,before);assert.throws(()=>Position.validatePosition({...p,extra:true},request));assert.throws(()=>Position.validatePosition(p,{...request,limit:3}));f.owner.dispose();
});
test('403 result admission binds whole question, position, lexical IDs and full versus remaining counts',()=>{
 const f=ready(),request=Search.buildRequest(f.question,[eq('service','api')],2),position=Position.makePosition(request,id(2)),good=response({request,position},{ids:[3,4],remaining:'3'}).data;assert.equal(Search.validateResult(good,request,position),good);
 const cases=[v=>v.remainingTraces='6',v=>v.state='no_data',v=>v.position.lastTraceId=id(1),v=>v.nextPosition.lastTraceId=id(3),v=>v.items[0].traceId=id(2),v=>v.request.limit=3,v=>delete v.remainingTraces];for(const mutate of cases){const bad=structuredClone(good);mutate(bad);assert.throws(()=>Search.validateResult(bad,request,position));}
 assert.throws(()=>Search.validateResult(good,request));const empty=response({request,position},{ids:[],total:'5',remaining:'0'}).data;assert.equal(Search.validateResult(empty,request,position),empty);f.owner.dispose();
});
test('403 live pages replace evidence, capture current position, and save only authored intent',async()=>{
 const f=ready(),first=await readFirst(f),oldNext=button(f.host,'Next page (fresh read)'),oldOpen=button(f.host,'Open trace (fresh read)'),nextPosition=response(first.request).data.nextPosition;
 const pending=oldNext.onclick(),next=requests.at(-1);assert.equal(next.url,'/api/trace-search/page');assert.deepEqual(next.request,{request:first.request,position:nextPosition});oldNext.onclick();oldOpen.onclick();assert.equal(requests.length,2);assert.equal(f.opened.length,0);
 next.resolve(response(next.request,{ids:[3,4],remaining:'3'}));await pending;assert.equal(all(f.host).filter(e=>e.dataset.traceId).length,2);assert.deepEqual(all(f.host).filter(e=>e.dataset.traceId).map(e=>e.dataset.traceId),[id(3),id(4)]);
 const captured=[],listener=e=>captured.push(structuredClone(e.detail));events.addEventListener('goo:save-capture',listener);try{button(f.host,'Capture search evidence').onclick();assert.deepEqual(captured[0].question,{kind:'trace-search',traceSearch:first.request,traceSearchPosition:nextPosition});saved=[];button(f.host,'Save search question').onclick();assert.deepEqual(saved[0].step.traceSearch,first.request);assert.equal(Object.hasOwn(saved[0].step,'traceSearchPosition'),false);assert.equal(requests.length,2)}finally{events.removeEventListener('goo:save-capture',listener)}
 const firstPage=button(f.host,'First page (fresh read)'),reset=firstPage.onclick(),read=requests.at(-1);assert.equal(read.url,'/api/trace-search');assert.deepEqual(read.request,first.request);read.resolve(response(read.request));await reset;await oldNext.onclick();await firstPage.onclick();assert.equal(requests.length,3);f.owner.dispose();
});
test('403 hidden, inert, edited and departed page callbacks cannot execute or admit late data',async()=>{
 const f=ready();await readFirst(f);const next=button(f.host,'Next page (fresh read)');next.hidden=true;await next.onclick();next.hidden=false;next.inert=true;await next.onclick();next.inert=false;assert.equal(requests.length,1);
 all(f.host).find(e=>Object.hasOwn(e.dataset,'traceSearchConditions')).open=true;const pending=next.onclick(),read=requests.at(-1),limit=all(f.host).find(e=>e.attributes['aria-label']==='Trace list limit');limit.value='3';limit.oninput();assert(read.signal.aborted);read.resolve(response(read.request,{ids:[3,4],remaining:'3'}));await pending;assert.equal(button(f.host,'Next page (fresh read)'),undefined);await next.onclick();assert.equal(requests.length,2);const fresh=button(f.host,'Run trace search').onclick(),current=requests.at(-1);assert.equal(current.url,'/api/trace-search');assert.equal(current.request.limit,3);f.owner.invalidate();current.resolve(response(current.request,{ids:[1,2,3]}));await fresh;assert.equal(button(f.host,'Open trace (fresh read)'),undefined);f.owner.dispose();
});
test('403 failed continuation offers an exact explicit retry with no earlier page basis',async()=>{
 const f=ready();await readFirst(f);const pending=button(f.host,'Next page (fresh read)').onclick(),read=requests.at(-1);read.resolve({ok:false,status:502,data:{error:'Unavailable'}});await pending;assert.equal(button(f.host,'Open trace (fresh read)'),undefined);const retry=button(f.host,'Retry page (fresh read)'),again=retry.onclick(),current=requests.at(-1);assert.deepEqual(current.request,read.request);current.resolve(response(current.request,{ids:[],total:'5',remaining:'0'}));await again;assert.equal(button(f.host,'Next page (fresh read)'),undefined);assert(button(f.host,'First page (fresh read)'));assert(all(f.host).some(e=>e.textContent.includes('No further matching reported trace IDs')));await retry.onclick();assert.equal(requests.length,3);f.owner.dispose();
});
test('403 accepted pages fold conditions while source-free disclosure preserves evidence and edits retire it',async()=>{
 const f=ready(),conditions=all(f.host).find(e=>Object.hasOwn(e.dataset,'traceSearchConditions'));
 assert.equal(conditions.open,true);assert.equal(conditions.children[0].textContent,'Search conditions · 1 · limit 2');
 await readFirst(f);assert.equal(conditions.open,false);const row=all(f.host).find(e=>e.dataset.traceId),next=button(f.host,'Next page (fresh read)');
 conditions.open=true;assert.equal(requests.length,1);assert.equal(all(f.host).find(e=>e.dataset.traceId),row);assert(next);
 const limit=all(f.host).find(e=>e.attributes['aria-label']==='Trace list limit');limit.value='3';limit.oninput();assert.equal(conditions.open,true);assert.equal(conditions.children[0].textContent,'Search conditions · 1 · limit 3');assert.equal(all(f.host).find(e=>e.dataset.traceId),undefined);await next.onclick();assert.equal(requests.length,1);
 const pending=button(f.host,'Run trace search').onclick(),read=requests.at(-1);read.resolve(response(read.request,{ids:[1,2,3]}));await pending;assert.equal(conditions.open,false);assert.equal(all(f.host).filter(e=>e.dataset.traceId).length,3);f.owner.dispose();
});
