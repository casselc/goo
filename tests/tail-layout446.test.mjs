import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as RecordColumns from '../ui/record-columns.mjs';
import * as RecordEnvironment from '../ui/record-environment.mjs';

const source=readFileSync(new URL('../ui/tail.mjs',import.meta.url),'utf8');
const prepare=source.split('\n').find(line=>line.startsWith('  function prepare('));
const exposed=source.split('\n').find(line=>line.startsWith('  return {setReadBlocked('));

test('actual Tail stage then Start retains saved field order independently for logs and spans',()=>{
 let current,calls=0;const control=()=>({setAttribute(){}}),context={RecordEnvironment,structuredClone,layouts:new Map(),resetFind(){},signalKind:'logs',columnControl:{
  // setLayout deliberately has the real passive controller contract: it validates
  // and projects, but does not emit a user's onChange event to populate caches.
  setLayout(value){current=RecordColumns.validate(value,['utc','service','summary'])},layout:()=>structuredClone(current)
 },section:control(),compact:true,table:{caption:{}},summaryHeader:{},resume:{},stop:{},viewport:control(),halt(){},clear(){},scope:{},clearFailure(){},follow:{},QuestionSession:{selectorText:()=>''},ready:()=>true,publish(){},reset(){},pauseStream(){},scopeSnapshot:()=>({}),connect(_reconnect,accepted){calls++;accepted()}};
 vm.runInNewContext(prepare+'\n'+exposed.replace('  return ','this.api='),context);
 const logs={signal:'logs',environment:'prod'},spans={signal:'spans',environment:'prod'},logLayout={columns:['summary','utc'],fieldColumns:[{kind:'attribute',key:'message.id'},{kind:'resource',key:'service.namespace'}]},spanLayout={columns:['utc','service','summary'],fieldColumns:[{kind:'column',key:'duration_ns'},{kind:'attribute',key:'http.route'}]},expectedLogs=structuredClone(logLayout),expectedSpans=structuredClone(spanLayout);
 context.api.stage(logs,logLayout);assert.deepEqual(context.api.layout(),expectedLogs);assert.equal(calls,0);
 logLayout.columns.reverse();logLayout.fieldColumns[0].key='external-mutation';
 assert.equal(context.api.start(logs),true);assert.deepEqual(context.api.layout(),expectedLogs);
 context.api.stage(spans,spanLayout);assert.deepEqual(context.api.layout(),expectedSpans);assert.equal(calls,1);
 spanLayout.columns.reverse();spanLayout.fieldColumns[0].key='external-mutation';
 assert.equal(context.api.start(spans),true);assert.deepEqual(context.api.layout(),expectedSpans);
 assert.equal(context.api.start(logs),true);assert.deepEqual(context.api.layout(),expectedLogs);
 assert.equal(context.api.start(spans),true);assert.deepEqual(context.api.layout(),expectedSpans);assert.equal(calls,4);
 assert.deepEqual(context.layouts.get('logs'),expectedLogs);assert.deepEqual(context.layouts.get('spans'),expectedSpans);
});
