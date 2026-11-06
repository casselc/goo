import fs from 'node:fs';
import assert from 'node:assert/strict';
import {fixture, observed, walk, tick, Dashboard} from '../tests/support/dashboard-owner-354.mjs';
import {reasons} from '../ui/source-refusal.mjs';

const {operations, scopeMode} = JSON.parse(fs.readFileSync(0, 'utf8'));
const owners=[], seen=new Set(), permits=new Set(), histories=[], snapshots=[];
let live=1, checks=0, maxActive=0, clock=0;
const stable=value=>JSON.stringify(value), clone=structuredClone;
const getBox=(owner,key)=>owner.boxes.get(key);
const retry=(owner,key)=>{const box=getBox(owner,key);return box?walk(box).find(el=>el.localName==='button'&&el.textContent==='Retry'):undefined};
const tree=element=>({tag:element.localName,text:element.textContent,data:element.dataset,attributes:element.attributes,children:element.children.map(child=>child&&typeof child==='object'?tree(child):child)});
const serialize=element=>stable(tree(element));
function freezeWitness(plan){assert(Object.isFrozen(plan));assert(Object.isFrozen(plan.widgets));for(const widget of plan.widgets){assert(Object.isFrozen(widget));if(widget.query)assert(Object.isFrozen(widget.query));if(widget.measurement)assert(Object.isFrozen(widget.measurement));}}
const f=fixture(6,{current:()=>live===1,onPlan:freezeWitness});
function configure(plan,id,mode,reason='window'){
 const scope={environment:'prod',fromMs:1000+id*100000,toMs:61000+id*100000};
 if(id===1){scope.fromMs=1000;scope.toMs=61000}
 if(mode){delete scope.environment;scope.environmentSelector={version:'goo/environment-selector/v1',policy:mode===1?'current-only':'effective',scope:mode<3?{mode:'named',value:'prod'}:{mode:mode===3?'empty':'unreported'}}}
 plan.scope=scope;plan.profile='all-spans';plan.declarations=[];plan.warnings=[];
 for(const [index,widget] of plan.widgets.entries()){
  widget.state=index===4?'unsupported':index===5?'not_observed':'supported';
  widget.reason=index>=4?'No supported matching capability.':undefined;
  if(index===2){delete widget.query;widget.family='go-runtime';widget.measurement={definition:{kind:'gauge',environment:'prod',source:{metric:'fixture.owner'+id+'.logs2',unit:'By'}},fromMs:scope.fromMs,toMs:scope.toMs};if(mode>=2)widget.state='unsupported';}
  else {widget.query={signal:'logs',...(scope.environmentSelector?{environmentSelector:clone(scope.environmentSelector)}:{environment:scope.environment}),fromMs:scope.fromMs,toMs:scope.toMs,limit:1,offset:0,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'attribute',key:'fixture.widget'},op:'eq',value:id+':'+widget.key}]},groupBy:[{kind:'column',key:'service'}]};}
 }
 return plan;
}
configure(f.plan,1,0);
function addOwner(id,plan,controller,promise){const owner={id,plan:clone(plan),controller,promise,retired:false,boxes:new Map(),slots:new Map(),healthy:new Map(),retiredBoxes:[]};for(const widget of plan.widgets)owner.slots.set(widget.key,{widget:clone(widget),state:widget.state==='supported'?'waiting':'unsupported',attempts:0});promise.then(handle=>{owner.handle=handle});owners.push(owner);return owner}
addOwner(1,f.plan,f.controller,f.loaded);
const current=()=>owners.at(-1);
function locateRequest(request){const signature=stable(request.question);for(const owner of owners)for(const [key,slot] of owner.slots){if(stable(slot.widget.query||slot.widget.measurement)===signature)return{owner,key,slot}}throw Error('Acquired a different fixed question');}
function sync(){
 for(const request of f.requests){if(seen.has(request))continue;seen.add(request);const {owner,key,slot}=locateRequest(request);request.owner=owner;request.key=key;assert(!owner.retired,'retired owner acquired');assert.equal(slot.widget.state,'supported','unsupported capability acquired');if(slot.attempts)assert(permits.delete(owner.id+':'+key),'unrequested or duplicate retry');slot.attempts++;slot.state='pending';}
 for(const owner of owners){
  const active=f.requests.filter(r=>r.owner===owner&&!r.settled&&!r.signal.aborted).length;maxActive=Math.max(maxActive,active);assert(active<=2,'initial and retries exceeded two slots');
  if(owner.retired){for(const saved of owner.retiredBoxes){assert.equal(serialize(saved.box),saved.snapshot,'late result mutated a retired card');assert.deepEqual(saved.box.children,saved.children)}continue}
  for(const [key,slot] of owner.slots){const box=walk(f.host).find(el=>el.localName==='article'&&el.dataset.widget===key);if(box)owner.boxes.set(key,box);if(slot.state==='healthy'&&box){assert.equal(box.dataset.state,'observed');let witness=owner.healthy.get(key);if(!witness){witness={box,children:box.children.slice(),snapshot:serialize(box),attempts:slot.attempts};owner.healthy.set(key,witness)}assert.equal(box,witness.box);assert.deepEqual(box.children,witness.children,'healthy sibling rebuilt');assert.equal(serialize(box),witness.snapshot,'healthy clock/evidence changed');assert.equal(slot.attempts,witness.attempts,'healthy sibling reread');}
   if(slot.state==='local-failure'&&box)assert.equal(retry(owner,key),undefined,'local protocol/client failure offered Retry');
   const button=retry(owner,key);if(button&&!histories.some(item=>item.button===button))histories.push({owner,key,button});
  }
  if(owner.handle){const keys=owner.handle.results.map(entry=>entry.widget.key);assert.equal(new Set(keys).size,keys.length,'duplicate result slots');for(const entry of owner.handle.results){assert(Object.isFrozen(entry),'caller can replace slot authority');const witness=owner.healthy.get(entry.widget.key);if(witness){if(!witness.entry){witness.entry=entry;witness.result=entry.result;witness.receipt=stable(entry.result)}assert.equal(entry,witness.entry,'healthy result slot replaced');assert.equal(entry.result,witness.result,'healthy observed result replaced');assert.equal(stable(entry.result),witness.receipt,'healthy read clock or evidence changed');}}}
 }
 checks++;
}
async function flush(){await tick();sync()}
async function settle(owner,key,outcome='success'){
 const request=f.requests.find(r=>r.owner===owner&&r.key===key&&!r.settled);if(!request)return;
 if(outcome==='aggregate'&&request.url==='/api/measurements')outcome='mismatch';
 request.settled=true;const slot=owner.slots.get(key),currentOwner=!owner.retired&&live===owner.id&&!request.signal.aborted;
 if(currentOwner)slot.state=outcome==='success'?'healthy':['client','mismatch','aggregate'].includes(outcome)?'local-failure':'source-failure';
 const readAt=new Date(Date.UTC(2026,10,6,0,0,++clock)).toISOString();
 if(outcome==='network')request.reject(new TypeError('https://private.invalid/secret-query'));
 else if(outcome==='client')request.resolve({ok:false,status:400,data:{error:'Invalid fixed scope'}});
 else if(outcome==='success'||outcome==='mismatch'||outcome==='aggregate'){
  const response=request.url==='/api/measurements'?{ok:true,status:200,data:{question:clone(request.question),state:'observed',readAt,basis:'b'.repeat(64)}}:observed(request,readAt);
  if(outcome==='mismatch'){if(response.data.question.definition)response.data.question.toMs++;else response.data.question.toMs++;}
  if(outcome==='aggregate'&&response.data.groups)response.data.groups[0].count='99';
  request.resolve(response);
 }else request.resolve({ok:false,status:502,data:{error:reasons[outcome],refusal:{kind:outcome,detail:'A bounded source read failed.'}}});
 await flush();
 if(currentOwner&&outcome==='network')assert(!serialize(getBox(owner,key)).includes('private.invalid'),'network diagnostic leaked');
 if(currentOwner&&slot.state==='source-failure')assert(retry(owner,key),'source refusal lacks explicit Retry');
}
async function clickRetry(owner,key,double=false){const button=retry(owner,key);if(!button)return;const slot=owner.slots.get(key),allowed=!owner.retired&&live===owner.id&&slot.state==='source-failure'&&!permits.has(owner.id+':'+key);if(allowed){permits.add(owner.id+':'+key);slot.state='queued'}button.onclick();if(double)button.onclick();await flush();}
function retire(owner){if(owner.retired)return;owner.retired=true;owner.controller.abort();owner.retiredBoxes=[...owner.boxes.values()].map(box=>({box,children:box.children.slice(),snapshot:serialize(box)}));for(const key of owner.slots.keys())permits.delete(owner.id+':'+key);}
async function replace(reason){const previous=current();retire(previous);const id=owners.length+1;live=id;const plan=configure(clone(previous.plan),id,scopeMode,reason),controller=new AbortController();const input={service:'api',scope:clone(plan.scope),profile:reason==='profile'?'auto':'all-spans'},promise=Dashboard.load(f.host,input,{signal:controller.signal,current:()=>live===id,onPlan:freezeWitness});addOwner(id,plan,controller,promise);f.metadata.at(-1).resolve(plan);await flush();}
async function blocked(owner,key,mode){const button=retry(owner,key);if(!button)return;const before=f.requests.length;if(mode==='hidden')f.host.hidden=true;else if(mode==='inert')f.host.inert=true;else f.host.root=false;button.onclick();assert.equal(f.requests.length,before,'hidden/inert/detached Retry acquired');f.host.hidden=f.host.inert=false;f.host.root=true;await flush();}
async function oldCallback(){const item=histories.find(item=>item.owner.retired||item.button!==retry(item.owner,item.key));if(!item)return;const before=f.requests.length;item.button.onclick();assert.equal(f.requests.length,before,'old handler acquired');await flush()}

await flush();
// Every generated schedule first crosses the contested queue and a real failure
// replacement. Random rule subsets then explore owner replacement and late work.
const first=current();await settle(first,'logs1','timeout');await clickRetry(first,'logs1',true);await settle(first,'logs0');await settle(first,'logs2','unavailable');await settle(first,'logs3');await settle(first,'logs1','malformed');await oldCallback();await blocked(first,'logs1','hidden');await blocked(first,'logs1','inert');await blocked(first,'logs1','detached');await clickRetry(first,'logs1',true);await settle(first,'logs1');await oldCallback();await clickRetry(first,'logs2',true);await settle(first,'logs2');await first.promise;sync();
assert.equal(first.slots.get('logs0').attempts,1);assert.equal(first.slots.get('logs3').attempts,1);assert.equal(first.slots.get('logs1').attempts,3);assert.equal(first.slots.get('logs2').attempts,2);
await replace('scope');await settle(current(),'logs0','client');await settle(current(),'logs1','mismatch');await settle(current(),'logs2','network');await settle(current(),'logs3','budget');await clickRetry(current(),'logs3',true);await settle(current(),'logs3');await clickRetry(current(),'logs2',true);await settle(current(),'logs2');
for(const action of operations||[]){const owner=current(),key='logs'+action.key;
 if(action.type==='settle')await settle(owner,key,action.outcome);
 else if(action.type==='retry')await clickRetry(owner,key,action.double);
 else if(action.type==='blocked')await blocked(owner,key,action.mode);
 else if(action.type==='old')await oldCallback();
 else if(action.type==='replace')await replace(action.reason);
 else if(action.type==='retire'){retire(owner);await flush()}
 else if(action.type==='late'){const pending=f.requests.find(r=>!r.settled&&r.signal.aborted);if(pending)await settle(pending.owner,pending.key);}
 else if(action.type==='authority'){const before=f.requests.length;const handle=owner.handle;if(handle?.plan.widgets.length){assert.throws(()=>handle.plan.widgets[0].state='supported');assert.throws(()=>handle.plan.widgets[0].query={});}if(handle?.results.length)assert.throws(()=>handle.results[0].widget={});assert.equal(f.requests.length,before);await flush()}
}
retire(current());for(const request of f.requests.filter(r=>!r.settled))await settle(request.owner,request.key);await Promise.all(owners.map(owner=>owner.promise));await flush();assert.equal(f.requests.filter(r=>!r.settled&&!r.signal.aborted).length,0);assert.equal(f.metadata.length,owners.length,'retry replanned overview');
process.stdout.write(JSON.stringify({checks,maxActive,owners:owners.length,requests:f.requests.length,critical:{record:first.slots.get('logs1').attempts,native:first.slots.get('logs2').attempts,healthy:first.slots.get('logs0').attempts},complete:true}));
