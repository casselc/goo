import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {mount} from '../ui/raw-observations.mjs';
import {Element} from './support/raw-observation-dom.mjs';
const fixture=JSON.parse(readFileSync(new URL('../testdata/http/raw-gauge438.json',import.meta.url))).cases.find(c=>c.name==='observed');
const request=()=>JSON.parse(fixture.body),response=()=>JSON.parse(fixture.response);
class EventsElement extends Element{constructor(tag){super(tag);this.events=new Map()}addEventListener(type,fn){this.events.set(type,fn)}}
function setup(){globalThis.Node=EventsElement;const body=new EventsElement('body');body.root=true;globalThis.document={body,hidden:false,activeElement:null,createElement:tag=>new EventsElement(tag)};const host=new EventsElement('main'),opener=new EventsElement('button');body.append(host,opener);const calls=[],pending=[],staged=[];let active=true;const owner=mount(host,{active:()=>active,onRecordContext:(...args)=>{staged.push(args);return true},post:async(path,input,signal)=>{calls.push({path,input,signal});const next=pending.shift();return typeof next==='function'?next(input):next||{ok:true,data:response()}}});return {owner,host,opener,calls,pending,staged,depart:()=>active=false,return:()=>active=true};}
const button=(h,text)=>h.querySelectorAll('button').find(b=>b.textContent===text||b.getAttribute('aria-label')===text);
function radio(h,value){return h.querySelectorAll('input').find(input=>input.type==='radio'&&input.value===value)}
function mode(h,value){const input=radio(h,value);input.checked=true;input.focus();input.onchange();return input;}
function emit(plot,keys,opener){plot.events.get('goo-raw-gauge-point')({detail:{occurrenceKeys:keys},composedPath:()=>[opener,plot]});}
function point(plot){const opener=new EventsElement('button');plot.append(opener);return opener;}
test('gauge points use the whole held bank, lazy and source-free; inspector uses actual point opener',async()=>{
 const h=setup();await h.owner.read(request(),h.opener);const plot=h.host.querySelector('goo-raw-gauge-points');assert.equal(plot.data,undefined);assert.equal(radio(h.host,'table').checked,true);assert.equal(h.host.querySelector('input').type,'search');
 const find=h.host.querySelector('input');find.value='does not match';find.oninput();mode(h.host,'points');assert.equal(plot.data.rows.length,response().rows.length);assert.equal(plot.hidden,false);assert.equal(find.parentElement.parentElement.hidden,true);assert.equal(h.calls.length,1);
 const opener=point(plot),key=plot.data.rows[0].key;emit(plot,[key],opener);assert.match(h.host.textContent,/Gauge observation 1/);assert.equal(plot.selectedKey,key);assert.equal(h.calls.length,1);assert.ok(h.host.querySelectorAll('goo-stat-list').some(list=>list.items?.some(item=>item.label==='Observation time')));assert.equal(button(h.host,'Nearby logs').onclick(),true);assert.equal(h.staged[0][0].origin.occurrenceKey,key);
 assert.match(h.host.textContent,/All 1 loaded · Table Find not applied/);const tableRadio=mode(h.host,'table');assert.equal(find.value,'does not match');assert.equal(document.activeElement,tableRadio);assert.match(h.host.textContent,/Selected observation is outside Table Find/);assert.equal(h.calls.length,1);
});
test('ambiguous, foreign and hidden plot events never choose an arbitrary held occurrence',async()=>{
 const h=setup();await h.owner.read(request(),h.opener);mode(h.host,'points');const plot=h.host.querySelector('goo-raw-gauge-points'),opener=point(plot),key=plot.data.rows[0].key;
 for(const keys of [[key,key],['foreign'],[],[null]])emit(plot,keys,opener);assert.equal(plot.selectedKey,null);assert.doesNotMatch(h.host.textContent,/Gauge observation 1Raw/);
 opener.hidden=true;emit(plot,[key],opener);assert.equal(plot.selectedKey,null);opener.hidden=false;mode(h.host,'table');emit(plot,[key],opener);assert.equal(plot.selectedKey,null);assert.equal(h.calls.length,1);
});
test('point-selected occurrence is revealed on its table page locally, preserving exact occurrence identity',async()=>{
 const h=setup(),data=response(),first=data.rows[0];data.rows=Array.from({length:205},(_,index)=>({...structuredClone(first),occurrenceKey:String(index+1)}));h.pending.push({ok:true,data});await h.owner.read(request(),h.opener);mode(h.host,'points');const plot=h.host.querySelector('goo-raw-gauge-points'),opener=point(plot);emit(plot,['205'],opener);const tableRadio=mode(h.host,'table');assert.equal(h.host.querySelectorAll('tr').length,6);assert.equal(document.activeElement,tableRadio);assert.equal(button(h.host,'Inspect observation 205').closest('tr').dataset.selected,'true');assert.equal(h.calls.length,1);
});
test('held points survive failed reread; prior Nearby lease retires and replacement bank revokes old plot',async()=>{
 const h=setup();await h.owner.read(request(),h.opener);mode(h.host,'points');const plot=h.host.querySelector('goo-raw-gauge-points'),opener=point(plot),key=plot.data.rows[0].key;emit(plot,[key],opener);const nearby=button(h.host,'Nearby logs'),old=nearby.onclick;assert.equal(old(),true);
 h.pending.push({ok:false,status:503,data:{error:'unavailable'}});await h.owner.read(request(),h.opener);assert.equal(h.host.querySelector('goo-raw-gauge-points'),plot);assert.equal(plot.hidden,false);assert.equal(old(),false);emit(plot,[key],opener);assert.equal(button(h.host,'Nearby logs').onclick(),true);assert.equal(h.calls.length,2);
 await h.owner.read(request(),h.opener);const oldCount=h.host.querySelectorAll('goo-stat-list').length;emit(plot,[key],opener);assert.equal(h.host.querySelectorAll('goo-stat-list').length,oldCount);assert.notEqual(h.host.querySelector('goo-raw-gauge-points'),plot);
});
test('departed and retired plot callbacks refuse; radio groups are independent per owner',async()=>{
 const h=setup();await h.owner.read(request(),h.opener);mode(h.host,'points');const plot=h.host.querySelector('goo-raw-gauge-points'),opener=point(plot),key=plot.data.rows[0].key;h.depart();emit(plot,[key],opener);assert.equal(plot.selectedKey,null);h.return();emit(plot,[key],opener);assert.equal(plot.selectedKey,key);h.owner.retire();emit(plot,[key],opener);assert.equal(h.host.querySelector('goo-raw-gauge-points'),null);
 const other=setup();await other.owner.read(request(),other.opener);const another=setup();await another.owner.read(request(),another.opener);assert.notEqual(radio(other.host,'points').name,radio(another.host,'points').name);
});
test('pending reread keeps held points inspectable while Nearby is revoked; cancel never adopts late bank',async()=>{
 const h=setup();await h.owner.read(request(),h.opener);mode(h.host,'points');const plot=h.host.querySelector('goo-raw-gauge-points'),opener=point(plot),key=plot.data.rows[0].key;emit(plot,[key],opener);let complete;h.pending.push(()=>new Promise(resolve=>complete=resolve));const running=h.owner.read(request(),h.opener);emit(plot,[key],opener);assert.match(h.host.textContent,/Gauge observation 1/);assert.equal(button(h.host,'Nearby logs').disabled,true);button(h.host,'Cancel read').onclick();complete({ok:true,data:{...response(),readAt:'2026-11-15T00:00:00Z'}});await running;assert.equal(h.host.querySelector('goo-raw-gauge-points'),plot);assert.equal(plot.hidden,false);assert.doesNotMatch(h.host.textContent,/2026-11-15T00:00:00Z/);emit(plot,[key],opener);assert.equal(button(h.host,'Nearby logs').disabled,false);assert.equal(h.calls.length,2);
});
test('explicit inspection reveals the shared heading and Close restores the actual visible opener',async()=>{
 const h=setup();await h.owner.read(request(),h.opener);const tableOpener=button(h.host,'Inspect observation 1');tableOpener.onclick();assert.equal(document.activeElement.localName,'h4');assert.match(document.activeElement.textContent,/Gauge observation 1/);button(h.host,'Close details').onclick();assert.equal(document.activeElement,tableOpener);
 mode(h.host,'points');const plot=h.host.querySelector('goo-raw-gauge-points'),pointOpener=point(plot),key=plot.data.rows[0].key;emit(plot,[key],pointOpener);assert.equal(document.activeElement.localName,'h4');button(h.host,'Close details').onclick();assert.equal(document.activeElement,pointOpener);assert.equal(h.calls.length,1);
});
test('Close after display switch reveals an equivalent visible Table action or the current mode radio',async()=>{
 const h=setup();await h.owner.read(request(),h.opener);mode(h.host,'points');const plot=h.host.querySelector('goo-raw-gauge-points'),pointOpener=point(plot),key=plot.data.rows[0].key;emit(plot,[key],pointOpener);mode(h.host,'table');button(h.host,'Close details').onclick();assert.equal(document.activeElement,button(h.host,'Inspect observation 1'));
 button(h.host,'Inspect observation 1').onclick();mode(h.host,'points');button(h.host,'Close details').onclick();assert.equal(document.activeElement,radio(h.host,'points'));
 emit(plot,[key],pointOpener);mode(h.host,'table');const find=h.host.querySelector('input');find.value='outside selected observation';find.oninput();button(h.host,'Close details').onclick();assert.equal(document.activeElement,radio(h.host,'table'));assert.equal(find.value,'outside selected observation');assert.equal(h.calls.length,1);
});

test('457 inspection marks only the visible selected table row without replacing its opener',async()=>{
 const h=setup(),data=response(),first=data.rows[0];data.rows=[first,{...structuredClone(first),occurrenceKey:'2'}];h.pending.push({ok:true,data});await h.owner.read(request(),h.opener);
 const rows=h.host.querySelectorAll('tr').filter(row=>row.dataset.occurrenceKey),firstOpener=button(h.host,'Inspect observation 1');firstOpener.onclick();assert.equal(rows[0].getAttribute('aria-current'),'true');assert.equal(rows[1].getAttribute('aria-current'),null);assert.equal(rows[0].querySelector('button'),firstOpener);
 button(h.host,'Inspect observation 2').onclick();assert.equal(rows[0].getAttribute('aria-current'),null);assert.equal(rows[0].dataset.selected,undefined);assert.equal(rows[1].getAttribute('aria-current'),'true');assert.equal(rows[1].dataset.selected,'true');assert.equal(h.calls.length,1);
});
test('457 cached inspection reuses its content and renews context actions without rebuilding the point',async()=>{
 const h=setup();await h.owner.read(request(),h.opener);const inspect=button(h.host,'Inspect observation 1');inspect.onclick();const body=h.host.querySelector('.ui-inspector-body'),content=body.firstChild,nearby=button(h.host,'Nearby logs');button(h.host,'Close details').onclick();inspect.onclick();assert.equal(body.firstChild,content);assert.equal(button(h.host,'Nearby logs'),nearby);assert.equal(nearby.onclick(),true);assert.equal(h.calls.length,1);
 h.pending.push({ok:false,status:503,data:{error:'unavailable'}});await h.owner.read(request(),h.opener);assert.equal(nearby.disabled,true);inspect.onclick();assert.equal(body.firstChild,content);assert.equal(nearby.disabled,false);assert.equal(nearby.onclick(),true);assert.equal(h.calls.length,2);
});
