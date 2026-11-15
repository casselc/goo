import {test} from 'node:test';
import assert from 'node:assert/strict';
import {mount} from '../ui/native-record-navigation.mjs';
import {Element} from './support/raw-observation-dom.mjs';

function fixture(){
 const body=new Element('body');body.root=true;globalThis.Node=Element;
 globalThis.document={body,hidden:false,activeElement:null,createElement:tag=>new Element(tag)};
 const listeners=new Map();globalThis.addEventListener=(name,fn)=>{const fs=listeners.get(name)||[];fs.push(fn);listeners.set(name,fs)};
 const emit=(name,detail)=>{for(const fn of listeners.get(name)||[])fn({detail})};
 globalThis.location={hash:'#explore'};const historyCalls=[];
 globalThis.history={pushState(_state,_title,hash){historyCalls.push(['push',hash]);location.hash=hash},replaceState(_state,_title,hash){historyCalls.push(['replace',hash]);location.hash=hash}};
 globalThis.scrollY=175;const scrolls=[];globalThis.scrollTo=options=>scrolls.push(options);
 const host=new Element('section'),nativeHost=new Element('section'),port=new Element('goo-table'),detailPort=new Element('div'),opener=new Element('button');
 detailPort.className='ui-inspector-body';port.scrollTop=80;port.scrollLeft=11;detailPort.scrollTop=43;detailPort.scrollLeft=4;
 body.append(host,nativeHost);nativeHost.append(port,detailPort);detailPort.append(opener);
 const state={active:true,alive:true,stage:[],shows:0,suspends:0,resources:0};
 const api=mount(host,{nativeHost,active:()=>state.active,stageQuestion:q=>{state.stage.push(q);state.onStage?.();state.active=false;nativeHost.hidden=true;emit('goo:explore-mode',{measurements:false})},showNative:()=>{state.shows++;state.active=true;nativeHost.hidden=false;emit('goo:explore-mode',{measurements:true})},suspendRecords:()=>state.suspends++,holdResourceOrigin:()=>{state.resources++;throw Error('Raw context must not acquire an evaluated resource origin')}});
 const plan={relation:'raw-observation-context',scope:'service',fromNS:'1000000000',toNS:'3000000000',atNS:'2000000000',radiusNS:'60000000000',question:{signal:'logs',environment:'prod',fromMs:1000,toMs:3000,limit:100,offset:0,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'worker'},{field:{kind:'resource',key:'service.namespace'},op:'missing'},{field:{kind:'column',key:'timestamp_ns'},op:'gte',value:'1000000000'},{field:{kind:'column',key:'timestamp_ns'},op:'lte',value:'2999999999'}]}},origin:{question:{environment:'prod',service:'worker',fromMs:1000,toMs:3000},instrument:{kind:'gauge',metric:'queue.depth'},readAt:'2026-11-15T12:00:00.123456789Z',occurrenceKey:'raw-1',stream:{resource:{'service.name':'worker'}}}};
 return {api,host,nativeHost,port,detailPort,opener,state,plan,emit,historyCalls,scrolls};
}

test('449 raw context stages an ordinary question and returns held scroll and focus without evaluated origin',()=>{
 const f=fixture();assert.equal(f.api.stage(f.plan,f.opener,{current:()=>f.state.alive}),true);
 assert.equal(f.state.stage.length,1);assert.deepEqual(f.state.stage[0],f.plan.question);assert.equal(f.state.resources,0);assert.equal(f.host.hidden,false);
 assert.match(f.host.textContent,/Back to observations/);assert.match(f.host.textContent,/Service-wide · instance not reported/);assert.match(f.host.textContent,/Shared time does not establish cause/);
 f.port.scrollTop=0;f.port.scrollLeft=0;f.detailPort.scrollTop=0;f.detailPort.scrollLeft=0;
 const back=f.host.querySelector('button');back.focus();back.onclick();assert.equal(f.state.shows,1);assert.equal(f.state.suspends,1);assert.equal(f.state.stage.length,1);
 assert.equal(f.port.scrollTop,80);assert.equal(f.port.scrollLeft,11);assert.equal(f.detailPort.scrollTop,43);assert.equal(f.detailPort.scrollLeft,4);assert.equal(document.activeElement,f.opener);assert.equal(f.host.hidden,true);assert.equal(location.hash,'#explore');assert.equal(f.scrolls.at(-1).top,175);
});

test('449 retired or non-synchronous capabilities cannot stage or restore raw context',()=>{
 for(const current of [undefined,()=>false,()=>Promise.resolve(true),()=>{throw Error('retired')}]){const f=fixture();assert.equal(f.api.stage(f.plan,f.opener,{current}),false);assert.equal(f.state.stage.length,0);assert.equal(f.historyCalls.length,0)}
 const f=fixture();assert.equal(f.api.stage(f.plan,f.opener,{current:()=>f.state.alive}),true);f.state.alive=false;assert.equal(f.api.restore('#explore'),false);assert.equal(f.state.shows,0);assert.equal(f.state.suspends,0);assert.equal(f.host.hidden,true);
});

test('449 hidden and disabled context actions refuse and a synchronous withdrawal creates no return citation',()=>{
 for(const withdraw of [f=>f.opener.disabled=true,f=>f.nativeHost.inert=true,f=>f.nativeHost.hidden=true,f=>document.hidden=true]){const f=fixture();withdraw(f);assert.equal(f.api.stage(f.plan,f.opener,{current:()=>true}),false);assert.equal(f.state.stage.length,0)}
 const f=fixture();f.state.onStage=()=>f.state.alive=false;assert.equal(f.api.stage(f.plan,f.opener,{current:()=>f.state.alive}),false);assert.equal(f.historyCalls.length,0);assert.equal(f.host.hidden,true);
});

test('449 reported empty instance remains explicit and background return cannot move focus',()=>{
 const f=fixture();f.plan.scope='reported-instance';f.plan.origin.stream.resource['service.instance.id']='';assert.equal(f.api.stage(f.plan,f.opener,{current:()=>true}),true);assert.match(f.host.textContent,/Reported empty instance label/);assert.doesNotMatch(f.host.textContent,/Service-wide · instance not reported/);
 document.hidden=true;assert.equal(f.api.restore('#explore'),false);assert.equal(f.state.shows,0);document.hidden=false;assert.equal(f.api.restore('#explore'),true);assert.equal(document.activeElement,f.opener);
});
