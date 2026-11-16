import {test} from 'node:test';
import assert from 'node:assert/strict';
import {prepare,adopt,model} from '../ui/held-chronology.mjs';
import {lane,item,lanes} from './support/held-chronology458-fixture.mjs';

test('held chronology preserves every occurrence and sorts exact nanoseconds with stable ties',()=>{
 const input={logs:lane('logs',[item('logs',0),item('logs',1,'1700000000000000002'),item('logs',2,'1700000000000000001'),item('logs',3,'bad'),item('logs',4,null)]),spans:lane('spans',[item('spans',0),item('spans',1,'9999999999999999999999999999999999')])};
 const m=model(prepare(input));assert.deepEqual(m.rows.map(row=>[row.signal,row.index]),[['logs',1],['logs',0],['logs',2],['spans',0],['logs',3],['logs',4],['spans',1]]);assert.equal(m.placed.length,4);assert.equal(m.unplaced.length,3);assert.deepEqual(m.counts,{logs:5,spans:2});assert.equal(new Set(m.rows.map(row=>row.key)).size,7);assert.equal(m.rows[1].timeNS,'1700000000000000001');assert.equal(m.rows[1].item.key,m.rows[3].item.key,'same source key in independent lanes remains two occurrences');assert.notEqual(m.rows[1].key,m.rows[3].key);
 assert.deepEqual(model(m.data).rows.map(row=>row.key),m.rows.map(row=>row.key));
});
test('zero, leading-zero and negative public-profile times retain the common exact formatter semantics',()=>{
 const source={logs:lane('logs',[item('logs',0,'0'),item('logs',1,'0001'),item('logs',2,'-1')]),spans:lane('spans',[])};const m=model(prepare(source));assert.equal(m.placed.length,3);assert.deepEqual(m.placed.map(row=>row.timeNS),['0001','0','-1']);assert.equal(m.unplaced.length,0);
});
test('prepared held data is detached, immutable and strips preservation capabilities before copying',()=>{
 const source=lanes();let tokenReads=0;Object.defineProperty(source.logs.rows[0],'membership',{enumerable:true,get(){tokenReads++;throw Error('Must not copy capabilities')}});const data=prepare(source);assert.equal(tokenReads,0);assert.equal(Object.isFrozen(data.lanes.logs.rows[0].record),true);assert.equal(JSON.stringify(data).includes('membership'),false);assert.equal(JSON.stringify(data).includes('SECRET-TOKEN'),false);assert.equal(data.lanes.logs.rows[0].connection,undefined);
 source.logs.rows[0].record.Body='changed';source.logs.selector.environment='changed';assert.equal(data.lanes.logs.rows[0].record.Body,'message 0');assert.equal(data.lanes.logs.selector.environment,'prod');assert.throws(()=>data.lanes.logs.rows.push({}));assert.throws(()=>data.lanes.logs.rows[0].record.Body='changed');
 assert.equal(adopt(data),data);const copied=adopt(JSON.parse(JSON.stringify(data)));assert.deepEqual(copied,data);const forged=JSON.parse(JSON.stringify(data));forged.lanes.logs.rows[0].membership={token:'SECRET-TOKEN'};assert.throws(()=>adopt(forged));forged.lanes.logs.rows[0]=structuredClone(data.lanes.logs.rows[0]);forged.quality={extent:'complete'};assert.throws(()=>adopt(forged));
});
test('lane and occurrence source windows remain independent; Find uses visible record values only',()=>{
 const source=lanes();source.logs.rows[0].readWindow={fromMs:1699999940000,toMs:1700000000000};source.logs.rows[0].record.ResourceAttributes={secret:'NEVER-FIND'};source.logs.selector.where={field:{kind:'resource',key:'region'},op:'eq',value:'NEVER-FIND'};source.logs.find={text:'nomatch',revision:4};source.spans.phase='unavailable';source.spans.completedRead=null;source.spans.readWindow=null;source.spans.rows=[];source.spans.gap=true;source.spans.evicted=12;
 const data=prepare(source),m=model(data);assert.equal(m.rows.length,1,'lane Find does not silently narrow held chronology');assert.equal(m.rows[0].item.readWindow.fromMs,1699999940000);assert.equal(data.lanes.logs.readWindow.fromMs,1700000000000);assert.equal(data.lanes.spans.completedRead,null);assert.equal(data.lanes.spans.phase,'unavailable');assert.equal(data.lanes.spans.evicted,12);assert.deepEqual(model(data,{query:'MESSAGE 0'}).matched,{logs:1,spans:0});assert.equal(model(data,{query:'NEVER-FIND'}).rows.length,0);assert.equal(model(data,{query:'SECRET-TOKEN'}).rows.length,0);assert.throws(()=>model(data,{query:'x'.repeat(257)}));
});
test('malformed source metadata, wrong identities and overbudget rows refuse the entire held view',()=>{
 for(const mutate of [s=>s.logs.selector.signal='spans',s=>s.logs.selector.environment='',s=>s.logs.completedRead.observedAt='bad',s=>s.logs.completedRead.rowCount=-1,s=>s.logs.completedRead.fromMs++,s=>s.logs.rows[0].readWindow.toMs++,s=>s.logs.rows[0].observedAt='bad',s=>s.logs.rows[0].limited='false',s=>s.logs.rows[0].late=null,s=>s.logs.rows[0].record.Body=1,s=>s.logs.rows.push(structuredClone(s.logs.rows[0])),s=>s.logs.rows[0].key='physical-id',s=>s.logs.phase='complete',s=>s.logs.evicted=-1]){const source=lanes();mutate(source);assert.throws(()=>prepare(source));}
 const source=lanes();source.logs.rows=Array.from({length:101},(_,index)=>item('logs',index));assert.throws(()=>prepare(source));source.logs.rows=source.logs.rows.slice(0,100);assert.equal(model(prepare(source)).counts.logs,100);
});
test('plain-data admission refuses getters, class instances, holes, nonfinite values and cycles without IO',()=>{
 let getterReads=0,fetches=0;const original=globalThis.fetch;globalThis.fetch=()=>{fetches++;throw Error('No source IO')};
 try{
  const source=lanes();Object.defineProperty(source.logs.rows[0],'record',{enumerable:true,get(){getterReads++;throw Error('bad')}});assert.throws(()=>prepare(source));assert.equal(getterReads,0);
  for(const value of [new Date(),()=>{},NaN,Infinity]){const bad=lanes();bad.logs.rows[0].record.extra=value;assert.throws(()=>prepare(bad));}
  const cycle=lanes();cycle.logs.rows[0].record.extra=cycle;assert.throws(()=>prepare(cycle));const holes=lanes();holes.logs.rows=new Array(1);assert.throws(()=>prepare(holes));
  const deep=lanes();let node=deep.logs.rows[0].record;for(let i=0;i<35;i++){node.next={};node=node.next}assert.throws(()=>prepare(deep));
  assert.equal(model(prepare(lanes())).rows.length,2);assert.equal(fetches,0);
 }finally{globalThis.fetch=original}
});

test('the actual log-tail transport omission means logs without rewriting the original selector',()=>{
 const source=lanes();delete source.logs.selector.signal;const data=prepare(source);assert.equal(Object.hasOwn(data.lanes.logs.selector,'signal'),false);assert.deepEqual(model(data).counts,{logs:1,spans:1});for(const value of ['',null,undefined,0,false]){const bad=lanes();bad.logs.selector.signal=value;assert.throws(()=>prepare(bad));}const wrong=lanes();delete wrong.spans.selector.signal;assert.throws(()=>prepare(wrong));
});

test('one-sided held views preserve a strictly unconfigured sibling without inventing its scope',()=>{
 const unconfigured=()=>({selector:null,phase:'stopped',staged:false,find:{text:'',revision:2},readWindow:null,completedRead:null,connection:null,evicted:0,gap:false,rows:[]});
 for(const phase of ['stopped','paused']){const source=lanes();source.spans={...unconfigured(),phase};const data=prepare(source);assert.equal(data.lanes.spans.selector,null);assert.deepEqual(model(data).counts,{logs:1,spans:0});assert.equal(adopt(JSON.parse(JSON.stringify(data))).lanes.spans.selector,null);}
 for(const change of [lane=>lane.phase='unavailable',lane=>lane.phase='loading',lane=>lane.rows=[item('spans',0)],lane=>lane.readWindow={fromMs:1700000000000,toMs:1700000060000},lane=>lane.completedRead=heldRead(),lane=>lane.connection='b'.repeat(64),lane=>lane.evicted=1,lane=>lane.gap=true,lane=>lane.find.text='old Find']){const source=lanes();source.spans=unconfigured();change(source.spans);assert.throws(()=>prepare(source));}
 function heldRead(){return lane('spans',[]).completedRead}
});
