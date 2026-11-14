import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
import {AuthoringSlot,recordCELContract} from '../ui/cel-authoring.mjs';

class Element {
 constructor(tag,text=''){this.tag=tag;this.textContent=text;this.children=[];this.attributes={};this.dataset={};this.hidden=false;this.inert=false;this.disabled=false;this.value='';this.selectionStart=0;this.selectionEnd=0}
 append(...nodes){for(const node of nodes){node.parentElement=this;this.children.push(node)}}
 setAttribute(name,value){this.attributes[name]=String(value)}getAttribute(name){return this.attributes[name]??null}
 focus(){globalThis.document.activeElement=this}
}
const walk=root=>[root,...root.children.flatMap(walk)];
const ready=element=>{for(let e=element;e;e=e.parentElement)if(e.hidden||e.inert||e.disabled)return false;return true};
const completion=element=>{if(element.hidden||element.inert)return false;for(let e=element.parentElement;e;e=e.parentElement)if(e.hidden||e.inert||e.disabled)return false;return true};
mock.module('../ui/components.mjs',{exports:{node:(...args)=>new Element(...args)}});
mock.module('../ui/dom-visibility.mjs',{exports:{nativeControlReady:ready,completionControlReady:completion}});
const {mount}=await import('../ui/cel-expression-editor.mjs');
const clone=x=>structuredClone(x),flush=()=>new Promise(resolve=>setImmediate(resolve));
const deferred=()=>{let resolve,reject;const promise=new Promise((a,b)=>{resolve=a;reject=b});return {promise,resolve,reject}};
const base=()=>({signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:20,where:{any:[{field:{kind:'attribute',key:'tenant'},op:'missing'},{field:{kind:'column',key:'service'},op:'eq',value:'api'}]}});
function receipt(request,value='ok'){
 const predicate={field:{kind:'column',key:'body'},op:'eq',value},question=clone(request.base);question.where={all:[question.where,predicate]};
 return {version:'goo.cel-plan/1',celLibrary:recordCELContract.library,expression:request.expression,binding:clone(request.binding),predicate,question,plan:{rows:'parameterized'}};
}
function setup({prepare,check}={}){
 globalThis.document={activeElement:null};const host=new Element('section'),conditions=new Element('fieldset'),changes=[],calls=[],owners=[];let available=true,discarded=0,prepares=0;
 const widget=mount(host,{conditionsHost:conditions,scopeSummary:'prod · logs',current:()=>available,owner:signal=>{prepares++;if(prepare)return prepare(signal);const owner=new AuthoringSlot({base:base(),baseIdentity:'a'.repeat(64),parentRevision:'1',slot:'condition',epoch:'mount'});owners.push(owner);return owner},check:(request,signal)=>{calls.push({request,signal});return check?check(request,signal):receipt(request)},onChange:(...args)=>changes.push(args),onDiscard:()=>discarded++});
 const find=(tag,label)=>walk(host).find(e=>e.tag===tag&&(e.textContent===label||e.getAttribute('aria-label')===label));
 const mode=find('select','Filter mode'),text=find('textarea','Additional condition (CEL)'),verify=find('button','Check condition'),clear=find('button','Clear condition'),discard=find('button','Discard expression'),feedback=walk(host).find(e=>e.getAttribute('role')==='status');
 return {widget,host,conditions,changes,calls,owners,mode,text,verify,clear,discard,feedback,prepares:()=>prepares,discarded:()=>discarded,setAvailable:value=>available=value,async expression(){mode.value='expression';mode.onchange();await flush()},type(value){text.value=value;text.oninput()},async run(){await verify.onclick();await flush()}};
}

test('CEL mode is lazy, native and retains one textarea with caret across mode switches',async()=>{
 const h=setup();assert.equal(h.prepares(),0);assert.equal(h.mode.value,'conditions');assert.equal(h.conditions.hidden,false);
 await h.expression();assert.equal(h.prepares(),1);assert.equal(h.conditions.hidden,true);
 const textarea=h.text;h.type('log.body ==');h.text.selectionStart=4;h.text.selectionEnd=7;
 h.mode.value='conditions';h.mode.onchange();assert.equal(h.text,textarea);assert.equal(h.widget.snapshot().text,'log.body ==');assert.equal(h.conditions.disabled,true);assert.equal(h.discard.hidden,false);
 h.mode.value='expression';h.mode.onchange();await flush();assert.equal(h.text,textarea);assert.equal(h.text.selectionStart,4);assert.equal(h.text.selectionEnd,7);assert.equal(h.calls.length,0);
 assert.equal(h.text.getAttribute('aria-describedby').includes(h.feedback.id),true);
});

test('explicit check stages complete parent once and clear removes only the owned condition',async()=>{
 const h=setup();await h.expression();h.type('log.body == "18446744073709551615"');assert.equal(h.changes.at(-1)[0],undefined);assert.equal(h.feedback.getAttribute('aria-live'),'off');
 await h.run();assert.equal(h.calls.length,1);assert.equal(h.calls[0].request.expression,'log.body == "18446744073709551615"');
 const staged=h.changes.filter(([q])=>q);assert.equal(staged.length,1);assert.deepEqual(staged[0][0].where.all[0],base().where);assert.equal(h.feedback.textContent,'Condition staged.');assert.equal(h.feedback.getAttribute('aria-live'),'polite');
 await h.clear.onclick();assert.deepEqual(h.changes.at(-1)[0],base());assert.equal(h.text.value,'');assert.equal(h.widget.canEditConditions(),true);
});

test('compiler diagnostics associate to persistent textarea without taking focus',async()=>{
 const failure={code:'syntax',phase:'parse',message:'Expected expression.',span:{start:11,end:11,line:1,column:12}},h=setup({check:()=>({ok:false,data:{diagnostic:failure}})});
 await h.expression();h.type('log.body ==');h.text.focus();await h.run();
 assert.equal(document.activeElement,h.text);assert.equal(h.text.getAttribute('aria-invalid'),'true');assert.equal(h.feedback.textContent,failure.message);assert.equal(h.feedback.getAttribute('aria-live'),'polite');assert.equal(h.widget.snapshot().diagnostic.span.start,11);
 h.type('log.body == "ok"');assert.equal(h.feedback.getAttribute('aria-live'),'off');assert.equal(h.text.getAttribute('aria-invalid'),'false');
});

test('malformed Unicode, NUL and byte budgets retain the draft without a planning request',async()=>{
 for(const source of ['\ud800','log.body == "\0"','é'.repeat(2049)]){
  const h=setup();await h.expression();h.type(source);await h.run();assert.equal(h.calls.length,0);assert.equal(h.widget.snapshot().text,source);assert.equal(h.text.getAttribute('aria-invalid'),'true');assert.equal(h.changes.some(([q])=>q),false);
 }
});

test('late ignored-abort success and error cannot replace a newer check',async()=>{
 for(const rejectOld of [false,true]){
  const pending=[],h=setup({check:()=>{const wait=deferred();pending.push(wait);return wait.promise}});await h.expression();h.type('log.body == "old"');const old=h.verify.onclick();await flush();
  h.type('log.body == "new"');const latest=h.verify.onclick();await flush();assert.equal(pending.length,2);assert.equal(h.calls[0].signal.aborted,true);
  if(rejectOld)pending[0].reject(Error('Obsolete failure'));else pending[0].resolve(receipt(h.calls[0].request,'old'));
  await old;assert.equal(h.widget.snapshot().diagnostic,null);assert.equal(h.host.children[0].dataset.phase,'checking');
  pending[1].resolve(receipt(h.calls[1].request,'new'));await latest;assert.equal(h.changes.filter(([q])=>q).length,1);assert.equal(h.changes.at(-1)[0].where.all[1].value,'new');
 }
});

test('native hidden or inert ancestry withdraws a pending completion even before sync',async()=>{
 for(const attribute of ['hidden','inert','disabled']){
  const wait=deferred(),h=setup({check:()=>wait.promise});await h.expression();h.type('log.body == "ok"');const pending=h.verify.onclick();await flush();h.host[attribute]=true;
  wait.resolve(receipt(h.calls[0].request));await pending;assert.equal(h.changes.some(([q])=>q),false);
  h.host[attribute]=false;h.widget.sync();assert.equal(h.verify.disabled,false);
 }
});

test('hide/dispose and Conditions mode reject deferred results and direct callbacks',async()=>{
 for(const retire of ['hide','dispose','conditions']){
  const wait=deferred(),h=setup({check:()=>wait.promise});await h.expression();h.type('log.body == "ok"');const pending=h.verify.onclick();await flush();
  if(retire==='hide')h.widget.setVisible(false);else if(retire==='dispose')h.widget.dispose();else{h.mode.value='conditions';h.mode.onchange()}
  wait.resolve(receipt(h.calls[0].request));await pending;assert.equal(h.changes.some(([q])=>q),false);
  const count=h.calls.length;await h.verify.onclick();assert.equal(h.calls.length,count);
 }
});

test('retired preparation errors are ignored after hide and restore',async()=>{
 const wait=deferred(),h=setup({prepare:()=>wait.promise});h.mode.value='expression';h.mode.onchange();await flush();h.widget.setVisible(false);h.widget.setVisible(true);wait.reject(Error('Retired context error'));await flush();
 assert.equal(h.widget.snapshot().diagnostic,null);assert.equal(h.feedback.textContent,'');
});

test('retained clear callback cannot erase text edited during its await',async()=>{
 const h=setup();await h.expression();h.type('log.body == "old"');await h.run();const clearing=h.clear.onclick();h.type('log.body == "new"');await clearing;assert.equal(h.text.value,'log.body == "new"');assert.equal(h.changes.at(-1)[0],undefined);
});

test('external parent retirement retains text and requires explicit discard before re-adoption',async()=>{
 const h=setup();await h.expression();h.type('log.body == "old"');await h.run();h.widget.invalidate();
 assert.equal(h.widget.snapshot().text,'log.body == "old"');assert.equal(h.widget.snapshot().needsAdoption,true);assert.equal(h.verify.disabled,true);assert.equal(h.clear.disabled,true);assert.equal(h.discard.hidden,false);
 h.type('log.body == "changed"');assert.equal(h.verify.disabled,true);const count=h.calls.length;await h.verify.onclick();assert.equal(h.calls.length,count);
 h.discard.onclick();assert.equal(h.discarded(),1);assert.equal(h.widget.snapshot().text,'');assert.equal(h.widget.snapshot().needsAdoption,false);assert.equal(h.widget.canEditConditions(),true);
});

test('long compiler diagnostics have one concise announcement and one optional full detail',async()=>{
 const raw='CEL typecheck: ERROR: <input>:1:23: mismatched input expecting expression\n | log.body == "timeout" &&\n | ......................^\nExpected tokens: IDENTIFIER, STRING, NUMBER, TRUE, FALSE, LPAREN\nAdditional compiler context stays available in details.';
 const diagnostic={code:'syntax',phase:'parse',message:raw,span:{start:22,end:22,line:1,column:23}};
 const h=setup({check:()=>({ok:false,data:{diagnostic}})});await h.expression();h.type('log.body == "timeout" &&');h.text.focus();await h.run();
 assert.equal(h.feedback.textContent,'Check expression syntax · line 1, column 23.');
 assert.equal(h.changes.at(-1)[1],'Check expression before applying.');assert.equal(h.changes.at(-1)[2].diagnostic.message,raw);
 assert.equal(h.widget.snapshot().diagnostic.message,raw);assert.equal(document.activeElement,h.text);
 const details=walk(h.host).find(e=>e.tag==='details'),pre=walk(h.host).find(e=>e.tag==='pre');
 assert.equal(details.hidden,false);assert.equal(details.open,false);assert.equal(pre.textContent,raw);assert.equal(pre.parentElement,details);assert.notEqual(details.parentElement,h.feedback);
 assert.equal(walk(h.host).filter(e=>e.textContent===raw).length,1);assert.equal(h.feedback.getAttribute('aria-live'),'polite');assert.equal(h.text.getAttribute('aria-describedby'),h.text.id+'-hint '+h.feedback.id);
 h.type('log.body == "ok"');assert.equal(details.hidden,true);assert.equal(pre.textContent,'');
});
