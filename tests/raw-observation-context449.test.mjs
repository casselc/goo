import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {plan,radiusNS} from '../ui/raw-observation-context.mjs';
import * as Environment from '../ui/record-environment.mjs';
import {validateRecordQuestion} from '../ui/question-protocol.mjs';

const clone=structuredClone;
function fixture(family='gauge',name='observed'){
 const file={gauge:'raw-gauge438',sum:'raw-sum441',histogram:'raw-histogram447'}[family];
 const item=JSON.parse(readFileSync(new URL('../testdata/http/'+file+'.json',import.meta.url))).cases.find(c=>c.name===name&&c.status===200);
 assert.ok(item,'independent Go fixture '+family+'/'+name);return JSON.parse(item.response);
}
function resource(data,key,value){for(const row of data.rows)for(const r of [row.stream.resource,row.rawRow.resource]){if(value===undefined)delete r[key];else r[key]=value;}}
function clock(data,{fromMs,toMs,atNS}){Object.assign(data.question,{fromMs,toMs});for(const row of data.rows){row.endNS=row.rawRow.end_ns=atNS;row.startNS=row.rawRow.start_ns='0';row.startState='valid';row.diagnostics=row.diagnostics.filter(d=>d.field!=='startNS');}}
const predicate=(p,kind,key)=>p.question.where.all.find(x=>x.field.kind===kind&&x.field.key===key);

test('449 all raw families and diagnostic magnitudes stage nearby records without evaluator results',()=>{
 for(const [family,names]of [['gauge',['observed','no-recorded']],['sum',['observed','stock','unknown-temporality','no-recorded']],['histogram',['observed','no-recorded','unknown-flags','count-mismatch','count-only','all-values','nonfinite-stat','missing-stat-columns']]])for(const name of names)for(const signal of ['logs','spans']){
  const data=fixture(family,name),before=clone(data),row=data.rows[0],p=plan(data,row.occurrenceKey,signal);
  assert.equal(p.scope,'service');assert.equal(p.relation,'raw-observation-context');assert.equal(p.grain,'raw-observation-context');assert.equal(p.question.signal,signal);assert.equal(p.radiusNS,radiusNS);assert.equal(p.origin.readAt,data.readAt);assert.deepEqual(p.origin.instrument,data.instrument);assert.deepEqual(p.origin.question,data.question);assert.deepEqual(p.origin.stream,row.stream);assert.equal(p.origin.occurrenceKey,row.occurrenceKey);assert.equal(p.origin.endNS,row.endNS);assert.equal(Object.hasOwn(p,'measurement'),false);assert.deepEqual(data,before);validateRecordQuestion(p.question);
 }
});

test('449 contextual scopes preserve namespace and instance absence, empty values and exact reported strings',()=>{
 for(const family of ['gauge','sum','histogram'])for(const namespace of [undefined,'','team'])for(const instance of [undefined,'','worker-a'])for(const signal of ['logs','spans']){
  const data=fixture(family);resource(data,'service.namespace',namespace);resource(data,'service.instance.id',instance);
  data.rows[0].stream.attributes={'not.a.record.filter':'private'};data.rows[0].rawRow.attributes=clone(data.rows[0].stream.attributes);
  const p=plan(data,data.rows[0].occurrenceKey,signal);
  assert.deepEqual(predicate(p,'column','service'),{field:{kind:'column',key:'service'},op:'eq',value:data.question.service});
  assert.deepEqual(predicate(p,'resource','service.namespace'),{field:{kind:'resource',key:'service.namespace'},op:namespace===undefined?'missing':'eq',...(namespace===undefined?{}:{value:namespace})});
  assert.deepEqual(predicate(p,'resource','service.instance.id'),instance===undefined?undefined:{field:{kind:'resource',key:'service.instance.id'},op:'eq',value:instance});
  assert.equal(p.scope,instance===undefined?'service':'reported-instance');
  assert.deepEqual(p.question.where.all.map(x=>x.field.key),['service','service.namespace',...(instance===undefined?[]:['service.instance.id']),'timestamp_ns','timestamp_ns']);
 }
});

test('449 environment policies remain exact, including empty and unreported effective scopes',()=>{
 for(const selection of [Environment.current('prod'),Environment.effective({mode:'named',value:'prod'}),Environment.effective({mode:'empty'}),Environment.effective({mode:'unreported'})]){
  const data=fixture();delete data.question.environment;data.question.environmentSelector=selection;
  const value=selection.scope.mode==='named'?'prod':selection.scope.mode==='empty'?'':undefined;resource(data,'deployment.environment.name',value);resource(data,'deployment.environment',undefined);
  const before=clone(data),p=plan(data,data.rows[0].occurrenceKey,'logs');assert.deepEqual(p.question.environmentSelector,selection);assert.equal(Object.hasOwn(p.question,'environment'),false);assert.deepEqual(data,before);
 }
});

test('449 nanosecond windows clip exactly and millisecond envelopes cannot broaden exact timestamp predicates',()=>{
 for(const [fromMs,toMs,atNS]of [[0,1,'0'],[1000,1001,'1000000001'],[1000,301001,'1000000000'],[1000,301001,'150000000123'],[1000,301001,'301000999999'],[1000000000000,1000000200000,'1000000090000000001']]){
  const data=fixture();clock(data,{fromMs,toMs,atNS});const p=plan(data,data.rows[0].occurrenceKey,'spans'),at=BigInt(atNS),lo=BigInt(fromMs)*1000000n,hi=BigInt(toMs)*1000000n,r=BigInt(radiusNS),from=at-r>lo?at-r:lo,to=at+r<hi?at+r:hi;
  assert.equal(p.fromNS,String(from));assert.equal(p.toNS,String(to));assert.equal(p.question.fromMs,Number(from/1000000n));assert.equal(p.question.toMs,Number((to+999999n)/1000000n));assert.deepEqual(p.question.where.all.slice(-2),[{field:{kind:'column',key:'timestamp_ns'},op:'gte',value:String(from)},{field:{kind:'column',key:'timestamp_ns'},op:'lte',value:String(to-1n)}]);
 }
});

test('449 occurrence membership is unique even when exports are identical, and all outputs detach from the bank',()=>{
 const data=fixture(),second=clone(data.rows[0]);second.occurrenceKey='second';data.rows.push(second);const before=clone(data),a=plan(data,'1','logs'),b=plan(data,'second','logs');assert.deepEqual(a.question,b.question);assert.notEqual(a.origin.occurrenceKey,b.origin.occurrenceKey);
 a.origin.stream.resource['service.name']='changed';a.origin.question.environment='other';a.origin.instrument.unit='changed';a.question.where.all[0].value='elsewhere';assert.deepEqual(data,before);assert.equal(b.origin.stream.resource['service.name'],'api');assert.equal(b.question.where.all[0].value,'api');
});

test('449 malformed banks, foreign occurrences, unsupported signals and unrepresentable resource values refuse atomically',()=>{
 for(const mutate of [d=>d.rows=[],d=>d.rows.push(clone(d.rows[0])),d=>d.rows[0].endNS='1500000001',d=>d.rows[0].rawRow.end_ns=d.rows[0].endNS='2000000000',d=>d.rows[0].stream.resource['service.name']='foreign',d=>d.question.environment='other',d=>d.instrument.kind='exponential-histogram',d=>d.readAt='not-a-clock',d=>resource(d,'service.instance.id','x'.repeat(1025)),d=>resource(d,'service.namespace','x'.repeat(1025)),d=>resource(d,'service.instance.id','bad\0id')]){
  const data=fixture();mutate(data);const before=clone(data);assert.throws(()=>plan(data,'1','logs'));assert.deepEqual(data,before);
 }
 for(const signal of ['both','metrics','',null,0])assert.throws(()=>plan(fixture(),'1',signal));
 for(const key of ['foreign','',null,1])assert.throws(()=>plan(fixture(),key,'logs'));
 const data=fixture();data.question.environmentSelector=Environment.effective({mode:'all'});delete data.question.environment;assert.throws(()=>plan(data,'1','logs'));
});
