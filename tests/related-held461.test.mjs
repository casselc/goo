import {test} from 'node:test';
import assert from 'node:assert/strict';
import {prepare,related} from '../ui/held-chronology.mjs';
import {lane,item} from './support/held-chronology458-fixture.mjs';
const trace='a'.repeat(32),other='c'.repeat(32),span='b'.repeat(16);
function bank(){return prepare(Object.fromEntries(['logs','spans'].map(signal=>[signal,lane(signal,[item(signal,0,undefined,{TraceId:trace,SpanId:span}),item(signal,1,undefined,{TraceId:other,SpanId:span}),item(signal,2,undefined,{TraceId:trace,SpanId:span})])])))}
test('related scope admits an exact held occurrence and preserves duplicates, lane facts and independent provenance',()=>{
 const data=bank(),before=JSON.stringify(data),result=related(data,{signal:'logs',key:data.lanes.logs.rows[0].key});assert.equal(result.rows.length,4);assert.deepEqual(result.matched,{logs:2,spans:2});assert.equal(result.data,data);assert.equal(JSON.stringify(data),before);assert.equal(new Set(result.rows.map(row=>row.key)).size,4);assert.equal(related(data,{signal:'logs',key:data.lanes.logs.rows[0].key,spanOnly:true}).rows.length,4);
 for(const value of ['0'.repeat(32),'A'.repeat(32),'bad',null]){const copy=structuredClone(data);copy.lanes.logs.rows[0].record.TraceId=value;assert.throws(()=>related(copy,{signal:'logs',key:copy.lanes.logs.rows[0].key}))}assert.throws(()=>related(data,{signal:'logs',key:'forged'}));
});
test('generated held-bank subsets never cross trace identity or collapse occurrences',()=>{
 for(let seed=1;seed<=80;seed++){
  const data=structuredClone(bank());for(const [si,signal]of ['logs','spans'].entries())for(const [i,row]of data.lanes[signal].rows.entries()){row.record.TraceId=(seed+i+si)%3?trace:other;row.record.SpanId=(seed+i)%2?span:'d'.repeat(16)}
  const anchor=data.lanes.logs.rows[0];for(const spanOnly of [false,true]){const result=related(data,{signal:'logs',key:anchor.key,spanOnly}),oracle=['logs','spans'].flatMap(signal=>data.lanes[signal].rows.filter(row=>row.record.TraceId===anchor.record.TraceId&&(!spanOnly||row.record.SpanId===anchor.record.SpanId)).map(row=>signal+':'+row.key));assert.deepEqual(result.rows.map(row=>row.key).sort(),oracle.sort())}
 }
});
