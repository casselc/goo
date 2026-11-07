import fs from 'node:fs';
import assert from 'node:assert/strict';
import {fixture,Saved,Context} from '../tests/support/saved-native-context-364.mjs';
const input=JSON.parse(fs.readFileSync(0,'utf8')),f=fixture(),clone=structuredClone;
let held=null,current=null,anchor=null,suspended=false,active=true,disposed=false,epoch=0,checks=0;const old=[],sources=[];
const bytes=JSON.stringify(input.pages);
// Go's record DTO echoes an inert blank scalar beside a typed policy; UI
// derives the scalar's absence from the checked native definition.
for(let i=0;i<input.pages.length;i++){const d=input.details[i],actual=Context.plan(input.pages[i].measurement.question,d.stream,d.question.signal,d.fromNS,d.toNS),wire=clone(d.question);if(wire.environmentSelector&&wire.environment==='')delete wire.environment;assert.deepEqual(wire,actual.question);assert.equal(d.grain,actual.grain);d.question=actual.question;}
function frozen(v){if(v&&typeof v==='object'){assert(Object.isFrozen(v));for(const x of Object.values(v))frozen(x)}}
function state(){return [f.opened.length,f.host.selectedKey,f.host.children.length]}
function save(){if(anchor)old.push({epoch,fn:anchor.onclick})}
function show(index){if(disposed)return;save();held=input.pages[index];f.owner.show(clone(held));current=f.render();current.detail=clone(input.details[index]);anchor=current.anchor();suspended=false;epoch++;assert.deepEqual(current.result,held.measurement);frozen(current.result);checks++}
function click(){if(!anchor)return;const n=f.opened.length,allowed=!disposed&&!suspended&&active&&!anchor.hidden&&anchor.isConnected;anchor.onclick();assert.equal(f.opened.length,n+(allowed?1:0));if(allowed){const [detail,source]=f.opened.at(-1);assert.deepEqual(detail,input.details[input.pages.indexOf(held)]);frozen(detail);frozen(source.receipt);sources.push({source,epoch});assert(source.current())}checks++}
function suspend(){if(disposed)return;save();f.owner.suspend();suspended=true;epoch++;for(const x of sources)if(x.epoch===epoch-1)assert(x.source.current(),'suspension destroyed retained source authority');checks++}
function resume(){if(disposed)return;save();const before=current,children=[...f.host.children],selection=f.host.selectedKey;f.owner.resume();if(suspended&&active){suspended=false;epoch++}assert.equal(f.render(),before);assert.deepEqual(f.host.children,children);assert.equal(f.host.selectedKey,selection);checks++}
function stale(){const before=state();for(const x of old)if(x.epoch<epoch)x.fn();assert.deepEqual(state(),before);checks++}
function hidden(){if(!anchor)return;const before=state();anchor.hidden=true;anchor.onclick();anchor.hidden=false;assert.deepEqual(state(),before);f.active(false);active=false;anchor.onclick();assert.deepEqual(state(),before);f.active(true);active=true;checks++}
// Every generated schedule includes positive admission and explicit retired/ABA controls.
for(let i=0;i<input.pages.length;i++){const p=input.pages[i],d=input.details[i];assert.deepEqual(Saved.admit(p,d),d);for(const mutate of [x=>x.question.toMs++,x=>x.stream.attributes.poison='foreign',x=>x.fromNS=String(BigInt(x.fromNS)+1n),x=>x.readAt+='x',x=>x.basis={forged:true}]){const bad=clone(d);mutate(bad);assert.throws(()=>Saved.admit(p,bad))}const copy=clone(p);copy.receipt.question.measurement.toMs++;assert.throws(()=>Saved.admit(copy,d));const gauge=clone(p);gauge.measurement.question.definition.kind="gauge";gauge.receipt.question.measurement.definition.kind="gauge";assert.throws(()=>Saved.admit(gauge,d));const bucket=clone(p);delete bucket.measurement[p.measurement.question.definition.kind+"s"];bucket.measurement.histogramTime={buckets:[{fromNS:d.fromNS,toNS:d.toNS}]};assert.throws(()=>Saved.admit(bucket,d));}
show(0);click();suspend();click();stale();resume();click();hidden();show(1);for(const x of sources)assert(!x.source.current());show(0);stale();click();
for(const action of input.actions){if(action===0)click();if(action===1)suspend();if(action===2)resume();if(action===3)show(0);if(action===4)show(1);if(action===5)stale();if(action===6)hidden();}
save();f.owner.dispose();disposed=true;epoch++;stale();for(const x of sources)assert(!x.source.current());assert.equal(JSON.stringify(input.pages),bytes);process.stdout.write(JSON.stringify({complete:true,checks,opens:f.opened.length}));
