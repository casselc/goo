import {test} from 'node:test';
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
import {node,mount} from './support/native-histogram-reuse-371.mjs';
import * as Native from '../ui/native-histogram-declaration.mjs';
const {result}=createRequire(import.meta.url)('./fixtures/native-histogram.cjs');
const deferred=()=>{let resolve;const promise=new Promise(r=>resolve=r);return{promise,resolve}};
const id=n=>n.toString(16).padStart(64,'0');
function setup(){const question=result().question,source={version:1,id:id(1),title:'Saved histogram 雪',question:structuredClone(question),declaration:Native.create(question,'table')},host=node('details');host.open=true;document.body.append(host);let owner={question:structuredClone(question),generation:1,active:true};owner.question.definition.source.services=['destination'];owner.question.definition.title='Destination';
 const calls=[],pending=[],applied=[];const request=(path,body)=>{const d=deferred();calls.push({path,...(body?{body:structuredClone(body)}:{})});pending.push(d);return d.promise};
 const controller=mount(host,{current:()=>owner,get:path=>request(path),post:request,onApply:view=>{applied.push(view);owner.generation++;return true}});
 const section=host.children[0];section.open=true;const find=id=>{const walk=e=>e.id===id?e:e.children.map(walk).find(Boolean);return walk(host)};
 return{host,section,controller,calls,pending,source,applied,owner:()=>owner,setOwner:value=>owner=value,browse:find('measureNativeViewBrowse'),more:find('measureNativeViewMore'),rows:find('measureNativeViewRows'),find:find('measureNativeViewFind'),apply:find('measureNativeViewApply'),status:find('measureNativeViewStatus'),listStatus:find('measureNativeViewListStatus')};}
const summary=v=>({id:v.id,title:v.title,kind:v.declaration.kind.name,unit:v.question.definition.source.unit});
const page=(items,extra={})=>({items,partial:false,diagnostics:[],...extra});
test('discovery summaries require explicit authoritative Load, then preserve current destination on Apply',async()=>{
 const s=setup(),destination=structuredClone(s.owner().question),listing=s.browse.onclick();assert.deepEqual(s.calls,[{path:'/native-histogram-views'}]);s.pending.shift().resolve(page([summary(s.source)]));await listing;assert.equal(s.apply.disabled,true);assert.match(s.listStatus.textContent,/no recency or snapshot claim/);
 const row=s.rows.querySelector('button'),loading=row.onclick();assert.equal(s.calls[1].path,'/native-histogram-views/'+s.source.id);s.pending.shift().resolve(s.source);await loading;assert.equal(s.apply.disabled,false);
 const applying=s.apply.onclick();assert.deepEqual(s.calls[2],{path:'/native-histogram-views/reuse',body:{id:s.source.id,question:destination}});s.pending.shift().resolve(Native.reuse(s.source,destination));await applying;assert.equal(s.applied.length,1);assert.deepEqual(s.applied[0].question,destination);
});
test('metadata paging and local Find preserve loaded-page scope and report unavailable scanned artifacts',async()=>{
 const s=setup(),first=Array.from({length:20},(_,i)=>({...summary(s.source),id:id(i+1),title:'Display '+(i+1)})),loading=s.browse.onclick();s.pending.shift().resolve(page(first,{nextCursor:id(20)}));await loading;
 s.find.value='Display 17';s.find.oninput();assert.equal(s.rows.querySelectorAll('button').length,1);assert.equal(s.calls.length,1);
 const more=s.more.onclick();assert.equal(s.calls[1].path,'/native-histogram-views?after='+id(20));s.pending.shift().resolve({items:[{...summary(s.source),id:id(21),title:'Another'}],partial:true,diagnostics:[{id:id(22),reason:'unavailable-artifact'}]});await more;assert.match(s.listStatus.textContent,/21 verified/);assert.match(s.listStatus.textContent,/partial/);assert.equal(s.more.hidden,true);assert.equal(s.rows.querySelectorAll('button').length,1);
});
test('stale list and inspect closures cannot revive after close/reopen, and Browse never adopts a display',async()=>{
 const s=setup(),oldBrowse=s.browse.onclick,listing=s.browse.onclick();s.section.open=false;s.section.dispatchEvent({type:'toggle'});s.section.open=true;s.section.dispatchEvent({type:'toggle'});s.pending.shift().resolve(page([summary(s.source)]));await listing;assert.equal(s.rows.querySelectorAll('button').length,0);await oldBrowse();assert.equal(s.calls.length,1);
 const fresh=s.browse.onclick();s.pending.shift().resolve(page([summary(s.source)]));await fresh;const oldRow=s.rows.querySelector('button'),loading=oldRow.onclick();s.section.open=false;s.section.dispatchEvent({type:'toggle'});s.section.open=true;s.section.dispatchEvent({type:'toggle'});s.pending.shift().resolve(s.source);await loading;assert.equal(s.apply.disabled,true);await oldRow.onclick();assert.equal(s.calls.length,3);assert.equal(s.applied.length,0);
});
test('summary substitution and incompatible unit metadata never enable Apply',async()=>{
 for(const mismatch of ['title','unit']){const s=setup(),donor=structuredClone(s.source);if(mismatch==='unit'){donor.question.definition.source.unit='ms'}const listing=s.browse.onclick();s.pending.shift().resolve(page([summary(donor)]));await listing;const loading=s.rows.querySelector('button').onclick();const reply=structuredClone(donor);if(mismatch==='title')reply.title='Substituted';s.pending.shift().resolve(reply);await loading;assert.equal(s.apply.disabled,true);assert.match(s.status.textContent,/differs|unit/);assert.equal(s.applied.length,0)}
});
test('Find rerender retires row callbacks even if an old row is reattached in the same question',async()=>{
 const s=setup(),listing=s.browse.onclick();s.pending.shift().resolve(page([summary(s.source)]));await listing;const old=s.rows.querySelector('button'),captured=old.onclick;s.find.value='histogram';s.find.oninput();s.rows.append(old);const count=s.calls.length;await captured();assert.equal(s.calls.length,count);assert.equal(s.applied.length,0);assert.equal(s.apply.disabled,true);
});
test('a second row Inspect replaces an in-flight first selection and ignores the late first response',async()=>{
 const s=setup(),second={...structuredClone(s.source),id:id(2),title:'Second display'},listing=s.browse.onclick();s.pending.shift().resolve(page([summary(s.source),summary(second)]));await listing;
 const [a,b]=s.rows.querySelectorAll('button'),first=a.onclick(),firstResponse=s.pending.shift(),next=b.onclick();assert.equal(s.calls.length,3);assert.equal(s.calls[2].path,'/native-histogram-views/'+second.id);
 s.pending.shift().resolve(second);await next;assert.equal(s.apply.disabled,false);assert.match(s.status.textContent,/Second display/);assert.equal(a.getAttribute('aria-pressed'),'false');assert.equal(b.getAttribute('aria-pressed'),'true');assert.equal(s.rows.querySelectorAll('button')[1],b);
 firstResponse.resolve(s.source);await first;assert.match(s.status.textContent,/Second display/);
 const applying=s.apply.onclick();assert.equal(s.calls[3].body.id,second.id);s.pending.shift().resolve(Native.reuse(second,s.owner().question));await applying;assert.equal(s.applied.length,1);assert.equal(s.applied[0].title,second.title);assert(s.rows.querySelectorAll('button').every(button=>button.getAttribute('aria-pressed')==='false'));
});
