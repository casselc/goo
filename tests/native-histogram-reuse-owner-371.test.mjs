import {test} from 'node:test';
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
import {node,mount} from './support/native-histogram-reuse-371.mjs';
import * as Native from '../ui/native-view-declaration.mjs';
const {result}=createRequire(import.meta.url)('./fixtures/native-histogram.cjs');
const deferred=()=>{let resolve;const promise=new Promise(r=>resolve=r);return{promise,resolve}};
function setup(family='histogram'){
 const question=result().question;if(family==='gauge'){question.definition.kind='gauge';question.definition.source.unit='By'}const source={version:1,id:'a'.repeat(64),title:'Interval display',question:structuredClone(question),declaration:Native.create(question,'table')};
 const host=node('section');document.body.append(host);
 let owner={question:structuredClone(question),generation:1,active:true};
 const calls=[],applied=[],pending=[];
 const controller=mount(host,{current:()=>owner,get:path=>{const d=deferred();calls.push({path});pending.push(d);return d.promise},post:(path,body)=>{const d=deferred();calls.push({path,body:structuredClone(body)});pending.push(d);return d.promise},onApply:view=>{applied.push(view);owner={...owner,generation:owner.generation+1};return true}});
 const find=id=>{const walk=e=>e.id===id?e:e.children.map(walk).find(Boolean);return walk(host)};
 const details=host.children[0];details.open=true;
 const input=find('measureNativeViewID'),load=find('measureNativeViewLoad'),apply=find('measureNativeViewApply'),status=find('measureNativeViewStatus');input.value=source.id;
 return{host,source,controller,calls,applied,pending,details,input,load,apply,status,owner:()=>owner,setOwner:v=>owner=v};
}
test('mounted reuse performs metadata reads and explicitly applies the exact current question',async()=>{
 const s=setup(),before=structuredClone(s.owner().question),loading=s.load.onclick();
 assert.deepEqual(s.calls,[{path:'/native-views/'+s.source.id}]);
 s.pending.shift().resolve(s.source);await loading;assert.equal(s.apply.disabled,false);assert.match(s.status.textContent,/current validated question/);
 const applying=s.apply.onclick();assert.deepEqual(s.calls[1],{path:'/native-views/reuse',body:{id:s.source.id,question:before}});
 s.pending.shift().resolve(Native.reuse(s.source,before));await applying;
 assert.equal(s.applied.length,1);assert.deepEqual(s.applied[0].question,before);assert.equal(Object.hasOwn(s.applied[0],'id'),false);assert.equal(s.apply.disabled,true);assert.match(s.status.textContent,/Run measurement to read the current question/);
});
test('late metadata load and apply cannot adopt into replaced owners or reuse stale action closures',async()=>{
 for(const stage of ['load','apply']){
  const s=setup(),oldLoad=s.load.onclick,loading=s.load.onclick();s.pending.shift().resolve(s.source);await loading;
  let operation;if(stage==='load')operation=s.load.onclick();else operation=s.apply.onclick();
  const previous=s.owner();s.setOwner({...previous,generation:previous.generation+1});s.controller.invalidate();
  s.pending.shift().resolve(stage==='load'?s.source:Native.reuse(s.source,previous.question));await operation;
  assert.equal(s.applied.length,0);assert.equal(s.apply.disabled,true);
  const count=s.calls.length;await oldLoad();assert.equal(s.calls.length,count,'old event closure stays retired after refresh');
 }
});
test('loaded ID and destination response are independently checked before adoption',async()=>{
 const s=setup(),loading=s.load.onclick();s.pending.shift().resolve({...s.source,id:'b'.repeat(64)});await loading;
 assert.equal(s.apply.disabled,true);assert.match(s.status.textContent,/ID changed/);
 const again=s.load.onclick();s.pending.shift().resolve(s.source);await again;
 const applying=s.apply.onclick(),changed=Native.reuse(s.source,s.owner().question);changed.question.toMs++;
 s.pending.shift().resolve(changed);await applying;assert.equal(s.applied.length,0);assert.match(s.status.textContent,/not applied/);
});
test('successful save acknowledgement survives independent metadata interaction but not owner cancellation',async()=>{
 for(const cancel of [false,true]){
  const s=setup(),opener=node('button','Save display');s.host.append(opener);s.details.open=false;
  const expected={version:1,title:s.source.title,question:s.owner().question,declaration:s.source.declaration};
  const saving=s.controller.save(expected,opener),saveAck=s.pending.shift();
  s.input.value='b'.repeat(64);s.input.oninput();
  if(cancel){s.setOwner({...s.owner(),generation:2});s.controller.invalidate()}
  saveAck.resolve(s.source);const actual=await saving;
  assert.equal(actual?.id??null,cancel?null:s.source.id);assert.deepEqual(s.calls[0],{path:'/native-views',body:expected});
 }
});

test('gauge owner stages only exact destination intent and fences replaced or hidden metadata callbacks',async()=>{
 const s=setup('gauge'),destination=structuredClone(s.owner().question);destination.definition.source.services=['different'];destination.definition.maxGapMs=500;destination.fromMs++;s.setOwner({...s.owner(),question:destination});const loading=s.load.onclick();s.pending.shift().resolve(s.source);await loading;assert.equal(s.apply.disabled,false);const applying=s.apply.onclick();assert.deepEqual(s.calls[1].body.question,destination);s.pending.shift().resolve(Native.reuse(s.source,destination));await applying;assert.equal(s.applied.length,1);assert.deepEqual(s.applied[0].question,destination);assert.equal(s.applied[0].declaration.kind.name,'table');assert.equal(s.calls.length,2);
 const h=setup('gauge'),pending=h.load.onclick();h.details.open=false;h.details.listeners.get('toggle')?.();h.pending.shift().resolve(h.source);await pending;assert.equal(h.apply.disabled,true);assert.equal(h.applied.length,0);
});

test('gauge metadata with omitted unit is refused before it can be applied',async()=>{
 const s=setup('gauge');delete s.source.question.definition.source.unit;const loading=s.load.onclick();s.pending.shift().resolve(s.source);await loading;assert.equal(s.apply.disabled,true);assert.match(s.status.textContent,/Display unavailable/);assert.doesNotMatch(s.status.textContent,/undefined/);assert.equal(s.applied.length,0);
});
