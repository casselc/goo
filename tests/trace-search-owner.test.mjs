import assert from 'node:assert/strict';
import {test,mock} from 'node:test';
import * as Environment from '../ui/record-environment.mjs';
class Element {
 constructor(tag='div',text=''){this.localName=tag;this.textContent=text||'';this.children=[];this.attributes={};this.dataset={};this.hidden=false;this.isConnected=true;this.disabled=false;this.listeners=new Map();this.scrolls=[];this.focuses=[];}
 set textContent(value){this._text=value??'';if(this.children){for(const child of this.children)child.parentElement=null;this.children=[];}}get textContent(){return this._text+(this.children||[]).map(child=>child.textContent).join('');}
 append(...children){for(const child of children){child.parentElement=this;this.children.push(child)}}
 replaceChildren(...children){for(const child of this.children)child.parentElement=null;this.children=[];this._text='';this.append(...children)}
 setAttribute(key,value){this.attributes[key]=String(value)} getAttribute(key){return this.attributes[key]}
 contains(node){return node===this||this.children.some(child=>child.contains(node))}
 addEventListener(type,callback){this.listeners.set(type,callback)} removeEventListener(type){this.listeners.delete(type)}
 remove(){this.isConnected=false;if(this.parentElement)this.parentElement.children=this.parentElement.children.filter(child=>child!==this)}
 querySelectorAll(selector){return this.children.flatMap(child=>[...(selector==='details[data-condition-editor]'&&child.localName==='details'&&Object.hasOwn(child.dataset,'conditionEditor')?[child]:[]),...child.querySelectorAll(selector)])}
 scrollIntoView(options){this.scrolls.push(options)} focus(options){this.focuses.push(options);document.activeElement=this}
}
const node=(...args)=>new Element(...args),events=new EventTarget(),documentEvents=new EventTarget();
globalThis.document={hidden:false,body:node('body'),activeElement:null,addEventListener:documentEvents.addEventListener.bind(documentEvents),removeEventListener:documentEvents.removeEventListener.bind(documentEvents)};
globalThis.addEventListener=events.addEventListener.bind(events);globalThis.removeEventListener=events.removeEventListener.bind(events);
let requests=[],editors=[];
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
const eq=(key,value)=>({field:{kind:'column',key},op:'eq',value}),all=element=>[element,...element.children.flatMap(all)],button=(host,text)=>all(host).find(node=>node.localName==='button'&&node.textContent===text);
function fixture(selector){
 requests=[];editors=[];let active=true,clean=true;const host=node('main'),opener=node('button','Inspect span'),opened=[],question={signal:'spans',environment:'',environmentSelector:selector,fromMs:1000,toMs:61000,where:eq('service','api'),limit:100,offset:0};
 host.append(opener);const owner=Search.mount(host,{current:()=>({question,active,clean}),onOpenTrace:(...args)=>opened.push(args)});assert(owner.open(eq('name','work'),opener));
 return {host,owner,opener,opened,question,setActive:value=>active=value,setClean:value=>clean=value};
}
function response(request,resources){
 const fieldValues=Search.conditionFields(request.conditions[0]).map(field=>({field,value:field.kind==='resource'?(Object.hasOwn(resources,field.key)?resources[field.key]:null):field.key==='name'?'work':'api'}));
 return {ok:true,data:{request,readAt:'2026-11-05T00:00:00Z',basis:'a'.repeat(64),plan:{sql:'SELECT',parameters:{}},warnings:[],state:'observed',matchedTraces:'1',hasMore:false,items:[{traceId:'1'.repeat(32),witnesses:[{conditionIndex:0,matchingRecords:'1',spanId:'2'.repeat(16),parentSpanId:'',service:'api',operation:'work',kind:'Server',timestampNs:'1000000000',durationNs:'1',fieldValues}]}]}};
}
test('mounted typed search opens a fresh exact trace with the same policy/window and no parent predicate',async()=>{
 for(const [selector,resources]of [[Environment.current('prod'),{[Environment.currentKey]:'prod'}],[Environment.effective({mode:'named',value:'prod'}),{[Environment.legacyKey]:'prod'}],[Environment.effective({mode:'empty'}),{[Environment.currentKey]:''}],[Environment.effective({mode:'unreported'}),{}]]){
  const f=fixture(selector),before=structuredClone(f.question),pending=button(f.host,'Run trace search').onclick();assert.equal(requests.length,1);const read=requests[0];assert.equal(read.url,'/api/trace-search');assert.deepEqual(read.request.environmentSelector,selector);assert.equal(Object.hasOwn(read.request,'environment'),false);read.resolve(response(read.request,resources));await pending;
  const open=button(f.host,'Open trace (fresh read)');assert(open);open.onclick();assert.equal(requests.length,1);assert.equal(f.opened.length,1);assert.deepEqual(f.opened[0][0],{signal:'spans',environmentSelector:selector,fromMs:1000,toMs:61000,limit:100,offset:0});assert.equal(f.opened[0][1],'1'.repeat(32));assert.equal(f.opened[0][0].where,undefined);assert.deepEqual(f.question,before);
  f.owner.invalidate();open.onclick();assert.equal(f.opened.length,1);f.owner.dispose();
 }
});
test('mounted search retirement aborts pending reads and refuses late success or a retained run handler',async()=>{
 for(const method of ['invalidate','dispose']){
  const f=fixture(Environment.effective({mode:'empty'})),run=button(f.host,'Run trace search'),pending=run.onclick(),read=requests[0];f.owner[method]();assert(read.signal.aborted);read.resolve(response(read.request,{[Environment.currentKey]:''}));await pending;assert.equal(button(f.host,'Open trace (fresh read)'),undefined);await run.onclick();assert.equal(requests.length,1);assert.equal(f.opened.length,0);if(method!=='dispose')f.owner.dispose();
 }
});
test('condition edits abort the old request and cannot admit a late result into the new draft',async()=>{
 const f=fixture(Environment.effective({mode:'unreported'})),pending=button(f.host,'Run trace search').onclick(),read=requests[0];editors[0].onChange(eq('name','changed'),null);assert(read.signal.aborted);read.resolve(response(read.request,{}));await pending;assert.equal(button(f.host,'Open trace (fresh read)'),undefined);const next=button(f.host,'Run trace search').onclick();assert.equal(requests.length,2);assert.deepEqual(requests[1].request.conditions,[eq('name','changed')]);f.owner.dispose();requests[1].resolve(response(requests[1].request,{}));await next;
});
test('refused replacement clears the old scope label instead of displaying stale policy context',()=>{
 const f=fixture(Environment.effective({mode:'named',value:'prod'})),panel=f.host.children.find(node=>node.attributes['aria-label']==='Find traces containing span conditions'),scope=panel.children[2];assert.match(scope.textContent,/current\/legacy precedence/);f.question.reportedEdge={};assert.equal(f.owner.open(eq('name','other'),f.opener),false);assert.equal(scope.textContent,'');assert.equal(requests.length,0);assert.equal(button(f.host,'Run trace search').disabled,true);f.owner.dispose();
});

test('explicit Close restores the current opener viewport and focus without another read',()=>{
 const f=fixture(Environment.effective({mode:'named',value:'prod'})),close=button(f.host,'Close trace search');close.onclick();assert.deepEqual(f.opener.scrolls,[{block:'center',inline:'nearest',behavior:'instant'}]);assert.deepEqual(f.opener.focuses,[{preventScroll:true}]);assert.equal(document.activeElement,f.opener);assert.equal(requests.length,0);assert.equal(close.parentElement.parentElement.hidden,true);f.owner.dispose();
});
test('retirement and departed Close handlers never scroll or focus a previous opener',()=>{
 for(const method of ['invalidate','dispose','inactive','dirty']){
  const f=fixture(Environment.effective({mode:'empty'})),close=button(f.host,'Close trace search'),other=node('button','New owner');document.activeElement=other;
  if(method==='inactive')f.setActive(false);else if(method==='dirty')f.setClean(false);else f.owner[method]();
  assert.deepEqual(f.opener.scrolls,[]);assert.deepEqual(f.opener.focuses,[]);close.onclick();assert.deepEqual(f.opener.scrolls,[]);assert.deepEqual(f.opener.focuses,[]);assert.equal(document.activeElement,other);assert.equal(requests.length,0);if(method!=='dispose')f.owner.dispose();
 }
});

test('typed trace-search refusal retries the same fixed question explicitly and cancellation retires a late refusal',async()=>{
 const f=fixture(Environment.effective({mode:'named',value:'prod'})),run=button(f.host,'Run trace search'),first=run.onclick(),request=structuredClone(requests[0].request);document.activeElement=f.opener;
 requests[0].resolve({ok:false,status:502,data:{error:'Source read timed out.',refusal:{kind:'timeout',detail:'The source read exceeded its time allowance.'}}});await first;const refused=all(f.host).find(node=>node.dataset.sourceRefusalKind==='timeout');assert(refused);assert.equal(refused.children[1].localName,'details');assert.equal(!!refused.children[1].open,false);assert.equal(document.activeElement,f.opener);assert.equal(button(f.host,'Open trace (fresh read)'),undefined);assert.equal(requests.length,1);
 const retry=run.onclick();assert.deepEqual(requests[1].request,request);requests[1].resolve(response(requests[1].request,{[Environment.legacyKey]:'prod'}));await retry;assert(button(f.host,'Open trace (fresh read)'));const next=run.onclick();assert.equal(requests.length,3);button(f.host,'Cancel trace search').onclick();assert.equal(requests[2].signal.aborted,true);requests[2].resolve({ok:false,status:502,data:{error:'Source unavailable.',refusal:{kind:'unavailable',detail:'The source read could not be completed.'}}});await next;assert.equal(all(f.host).filter(node=>node.dataset.sourceRefusalKind).length,0);assert.equal(button(f.host,'Open trace (fresh read)'),undefined);assert.equal(requests.length,3);f.owner.dispose();
});
