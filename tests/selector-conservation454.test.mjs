import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {populationQuery} from '../ui/record-population.mjs';
import {query as countedValues} from '../ui/filter-values.mjs';
import {expectedSideQuestion} from '../ui/explanation.mjs';
import {identity} from '../ui/question-session.mjs';
const bank=JSON.parse(readFileSync(new URL('./fixtures/selector-conservation454.json',import.meta.url),'utf8'));
const clone=structuredClone;
const field={kind:'column',key:'service'};
test('shared Go/JS casebank conserves complete selectors while stripping only projections',()=>{
 assert.equal(bank.version,'goo/selector-conservation/v1');
 for(const entry of bank.cases){const before=clone(entry.query),actual=populationQuery(entry.query,{limit:entry.limit,offset:entry.offset});assert.equal(identity(actual),identity(entry.population),entry.name);assert.deepEqual(entry.query,before,entry.name+' source mutation');for(const key of ['where','reportedEdge','linksTo','environmentSelector'])if(entry.query[key]!==undefined){assert.deepEqual(actual[key],entry.query[key],entry.name+' '+key);assert.notEqual(actual[key],entry.query[key],entry.name+' detached '+key)}for(const key of ['numeric','numericHistogram','groupBy','summary','spanStatus','logSeverity','bucketMs','groupTimeSeries'])assert.equal(Object.hasOwn(actual,key),false,entry.name+' stripped '+key);}
});
test('invalid original scopes are rejected before projections can repair them',()=>{
 for(const entry of bank.invalid)assert.throws(()=>populationQuery(entry.query,{limit:entry.limit,offset:entry.offset}),entry.name);
 const base=bank.cases[0].query;
 for(const input of [null,{...base,linksTo:null},{...base,numeric:null},{...base,where:null},{...base,reportedEdge:null},{...base,groupBy:null},{...base,extraSelector:true}])assert.throws(()=>populationQuery(input));
 for(const options of [{limit:0},{offset:-1},{limit:null},{offset:null},{signal:'logs'},{where:undefined},null])assert.throws(()=>populationQuery(base,options));
});
test('counted values derive population with exact selectors and numeric predicates intact',()=>{
 for(const entry of bank.cases){const actual=countedValues(entry.query,field),expected={...populationQuery(entry.query,{limit:1}),groupBy:[field]};assert.equal(identity(actual),identity(expected),entry.name);if(entry.query.where?.all?.some(p=>p.numeric))assert.deepEqual(actual.where,entry.query.where);}
 const source=bank.cases.find(entry=>entry.query.linksTo)?.query;const q=countedValues(source,field);q.linksTo.spanId='f'.repeat(16);assert.notEqual(q.linksTo.spanId,source.linksTo.spanId);
});
test('explanation refinements intentionally keep original projections as well as selectors',()=>{
 for(const entry of bank.cases){const baseline=entry.query,request={version:'goo/explanation/v1',baseline,selection:baseline,fields:[field]};const actual=expectedSideQuestion(request,field,'api','selection','include');for(const key of ['linksTo','reportedEdge','numeric','numericHistogram','groupBy','summary','spanStatus','logSeverity','bucketMs','groupTimeSeries'])assert.deepEqual(actual[key],baseline[key],entry.name+' '+key);assert.equal(actual.offset,0);}
});
test('population conservation and counted planning perform no source IO',()=>{
 const previous=globalThis.fetch;let calls=0;globalThis.fetch=()=>{calls++;throw Error('unexpected source read')};try{for(const entry of bank.cases){populationQuery(entry.query);countedValues(entry.query,field)}assert.equal(calls,0)}finally{globalThis.fetch=previous}
});
