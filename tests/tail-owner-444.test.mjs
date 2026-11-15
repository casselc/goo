import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as TailPresentation from '../ui/tail-presentation.mjs';
const source=readFileSync(new URL('../ui/tail.mjs',import.meta.url),'utf8');
const between=(a,b)=>source.slice(source.indexOf(a),source.indexOf(b,source.indexOf(a)));
const control=()=>({dataset:{},state:'unknown',textContent:'',hidden:false,disabled:false});

test('actual state and Find notifications are owner-tagged detached scope observations',()=>{
 const events=[],states=[],finds=[],selector={signal:'spans',environment:'prod',where:{field:{kind:'attribute',key:'region'},op:'eq',value:'west'}},completedRead={fromMs:1000,toMs:61000,revision:2},s={TailPresentation,bufferFacts:control(),owner:'tail-fixture',signalKind:'spans',selector,completedRead,phase:'paused',staged:false,find:{value:'needle'},findDraft:'needle',findRevision:1,lastTo:61000,connection:'a'.repeat(64),buffer:{rows:[{key:'one'}],evicted:0,observed:1},gap:false,generation:3,structuredClone,status:control(),section:control(),counts:control(),pause:control(),resume:control(),controller:null,saving:false,compact:false,refreshBatch(){},global:{dispatchEvent:e=>events.push(e)},CustomEvent:class{constructor(type,{detail}){this.type=type;this.detail=detail}},
  onState:value=>{states.push(structuredClone(value));value.selector.environment='callback edit';value.completedRead.revision=9;value.find.text='changed';value.keys.push('bad')},onFind:value=>{finds.push(structuredClone(value));value.selector.where.value='callback edit'}};
 vm.runInNewContext(source.match(/function scopeSnapshot\(\)\{[^\n]+\}/)[0]+between('  function notification(','  function halt(')+';this.publish=publish;this.notifyFind=notifyFind;',s);
 s.publish('unknown','Listening');s.notifyFind();assert.equal(states.length,1);assert.equal(finds.length,1);assert.equal(events.length,2);
 for(const {detail}of events){assert.equal(detail.owner,'tail-fixture');assert.equal(detail.signal,'spans');assert.equal(detail.connection,'a'.repeat(64));assert.equal(detail.selector.environment,'prod');assert.equal(detail.selector.where.value,'west');assert.equal(detail.completedRead.revision,2);assert.equal(detail.find.text,'needle');assert.deepEqual([...detail.keys],['one']);assert.equal(detail.readWindow.fromMs,1000);assert.equal(detail.generation,3)}
 assert.equal(selector.environment,'prod');assert.equal(completedRead.revision,2);events[0].detail.selector.environment='global edit';assert.equal(selector.environment,'prod');assert.equal(finds[0].selector.environment,'prod');
});

test('actual exposed start refuses an inactive lane before clearing or preparing held state',()=>{
 let allowed=false,prepared=0,connected=0;const held={phase:'paused',selector:{environment:'prod'}},context={usable:()=>allowed,resume:{dataset:{}},scopeSnapshot:()=>structuredClone(held),reset(){},pauseStream(){},prepare(){prepared++},connect(_reconnect,onAccepted){connected++;onAccepted?.()}};context.ready=()=>context.usable()&&!context.resume.dataset.readBlocked;
 const line=source.split('\n').find(line=>line.startsWith('  return {setReadBlocked('));assert(line);vm.runInNewContext(line.replace('  return ','this.api='),context);
 assert.equal(context.api.start({signal:'logs',environment:'prod'}),false);assert.equal(prepared,0);assert.equal(connected,0);assert.deepEqual(context.api.scope(),held);
 allowed=true;assert.equal(context.api.start({signal:'logs',environment:'prod'}),true);assert.equal(prepared,1);assert.equal(connected,1);
});


test('inactive or unavailable Find callbacks restore the last owned text and preserve held projection',()=>{
 let allowed=true;const calls={apply:0,notify:0,refresh:0,focus:0},find={value:'held',ready:true,setCustomValidity(){},removeAttribute(){},setAttribute(){},focus(){calls.focus++}},clearFind={ready:true},context={usable:()=>allowed,nativeControlReady:control=>control.ready,find,clearFind,findDraft:'held',findText:'held',findRevision:4,RecordPresentation:{findQuery:value=>{if(value==='invalid')throw Error('Invalid Find')}},findStatus:{},applyFind(){calls.apply++},refreshBatch(){calls.refresh++},notifyFind(){calls.notify++}};
 vm.runInNewContext(between('  function resetFind(','  // Observers receive detached metadata.'),context);
 allowed=false;find.value='unowned';find.oninput();assert.equal(find.value,'held');assert.equal(context.findRevision,4);assert.equal(context.findText,'held');clearFind.onclick();assert.deepEqual(calls,{apply:0,notify:0,refresh:0,focus:0});
 allowed=true;find.ready=false;find.value='hidden';find.oninput();assert.equal(find.value,'held');assert.equal(context.findRevision,4);find.ready=true;
 find.value='invalid';find.oninput();assert.equal(context.findDraft,'invalid');assert.equal(context.findText,'held');assert.equal(context.findRevision,5);assert.equal(calls.refresh,1);assert.equal(calls.notify,1);
 allowed=false;find.value='lost invalid draft';find.oninput();assert.equal(find.value,'invalid','withdrawal preserves even an owned invalid draft');assert.equal(context.findRevision,5);
 allowed=true;clearFind.ready=false;clearFind.onclick();assert.equal(find.value,'invalid');assert.equal(context.findRevision,5);
 clearFind.ready=true;clearFind.onclick();assert.equal(find.value,'');assert.equal(context.findDraft,'');assert.equal(context.findText,'');assert.equal(context.findRevision,6);assert.equal(calls.apply,1);assert.equal(calls.notify,2);assert.equal(calls.focus,1);
});

test('captured Pause and Close actions cannot mutate inactive or unavailable retained lanes',()=>{
 let allowed=false;const calls={pause:0,reset:0,connect:0,align:0},control=()=>({ready:true,hidden:false,disabled:false,focus(){}}),pause=control(),resume=control(),stop=control(),context={usable:()=>allowed,nativeControlReady:e=>e.ready,pause,resume,stop,document:{activeElement:null},compact:true,selector:{environment:'prod'},terminalFailure:null,staged:false,viewport:{},header:{},body:{children:[]},pauseStream(){calls.pause++},reset(){calls.reset++},connect(){calls.connect++},alignPausedRecords(){calls.align++}};
 vm.runInNewContext(between('  pause.onclick=','let pointerFocus='),context);pause.onclick();stop.onclick();resume.onclick();assert.deepEqual(calls,{pause:0,reset:0,connect:0,align:0});
 allowed=true;pause.ready=stop.ready=resume.ready=false;pause.onclick();stop.onclick();resume.onclick();assert.deepEqual(calls,{pause:0,reset:0,connect:0,align:0});
 pause.ready=stop.ready=true;pause.onclick();stop.onclick();assert.equal(calls.pause,1);assert.equal(calls.reset,1);
});

test('captured Keep Retain Clear and retention-check callbacks cannot write or alter inactive selections',async()=>{
 const calls={write:0,mutate:0,metadata:0},context={usable:()=>false,nativeControlReady:()=>true,keep:{},save:{},dismiss:{},checkRetention:{},tags:{value:'held'},noteInput:{value:'note'},find:{focus(){calls.mutate++}},selected:{items:new Map([['held',{}]])},nodes:new Map(),keepBusy:false,kept:null,saving:false,keepIntent:null,pendingIntent:null,selectionGeneration:1,selector:{environment:'prod'},document:{activeElement:null},refreshSelection(){calls.mutate++},clearSelection(){calls.mutate++},checkConfiguration(){calls.metadata++},Keeps:{key:()=>'',send:async()=>{calls.write++}},RetentionClient:{key:()=>'',send:async()=>{calls.write++}},UI:{actionVisible:()=>true}};
 vm.runInNewContext(between('  keep.onclick=','\n  async function checkConfiguration(')+between('  dismiss.onclick=','\n  const expiryTimer=')+between('  checkRetention.onclick=','\n  function refreshSelection('),context);
 await context.keep.onclick();await context.save.onclick();context.dismiss.onclick();context.checkRetention.onclick();assert.deepEqual(calls,{write:0,mutate:0,metadata:0});
 context.usable=()=>true;context.nativeControlReady=()=>false;await context.keep.onclick();await context.save.onclick();context.dismiss.onclick();context.checkRetention.onclick();assert.deepEqual(calls,{write:0,mutate:0,metadata:0});
});

test('row inspection callback requires active ownership native readiness and the exact retained row',()=>{
 let allowed=false,ready=true;const calls={pause:0,inspect:0},row={},item={key:'held',record:{}},nodes=new Map([['held',row]]),context={usable:()=>allowed,nativeControlReady:()=>ready,row,item,nodes,inspect:{},r:item.record,pauseStream(){calls.pause++},onInspect(){calls.inspect++}};
 const guard=source.match(/const current=button=>([^;]+);const scopeActions=/)[1],handler=source.match(/inspect\.onclick=(\(\)=>\{[^\n]+?\});const recordActions/)[1];
 vm.runInNewContext('this.current=button=>'+guard+';this.callback='+handler,context);
 context.callback();assert.deepEqual(calls,{pause:0,inspect:0});allowed=true;ready=false;context.callback();assert.deepEqual(calls,{pause:0,inspect:0});ready=true;nodes.set('held',{});context.callback();assert.deepEqual(calls,{pause:0,inspect:0});nodes.set('held',row);context.callback();assert.deepEqual(calls,{pause:1,inspect:1});context.pauseStream=()=>{calls.pause++;allowed=false};context.callback();assert.deepEqual(calls,{pause:2,inspect:1},'owner withdrawal while pausing prevents inspection handoff');
});

test('inactive annotation edits restore owned tags and note without changing preservation eligibility',()=>{
 const calls={refresh:0},context={usable:()=>false,nativeControlReady:()=>true,tags:{value:'injected'},noteInput:{value:'injected'},tagsDraft:'held-tag',noteDraft:'held-note',refreshSelection(){calls.refresh++}};
 vm.runInNewContext(between('  function annotationInput(','\n  function refreshSelection('),context);context.tags.oninput();assert.equal(context.tags.value,'held-tag');assert.equal(context.noteInput.value,'held-note');assert.equal(calls.refresh,0);
 context.usable=()=>true;context.nativeControlReady=()=>false;context.tags.value='unavailable';context.tags.oninput();assert.equal(context.tags.value,'held-tag');assert.equal(calls.refresh,0);
 context.nativeControlReady=()=>true;context.tags.value='new-tag';context.tags.oninput();assert.equal(context.tagsDraft,'new-tag');assert.equal(context.noteDraft,'held-note');assert.equal(calls.refresh,1);
});
