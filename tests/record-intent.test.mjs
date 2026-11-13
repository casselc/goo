import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as P from '../ui/question-protocol.mjs';
import {normalizeIntent} from '../ui/scope-history.mjs';
import {Session,identity} from '../ui/question-session.mjs';
const q=()=>({signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0});
test('all ordinary constructors refuse executable properties before copying or serializing',()=>{
 let invoked=0;const evil=()=>{const value=q();Object.defineProperty(value,'where',{enumerable:true,get(){invoked++;return undefined}});return value;};
 for(const check of [x=>P.validateRecordQuestion(x),x=>P.recordStep(x),x=>P.questionLink(x),x=>normalizeIntent({kind:'question',question:x})])assert.throws(()=>check(evil()));
 const envelope={kind:'question'};Object.defineProperty(envelope,'question',{enumerable:true,get(){invoked++;return q()}});assert.throws(()=>normalizeIntent(envelope));
 const json={kind:'question',question:q(),toJSON(){invoked++;return {}}};assert.throws(()=>normalizeIntent(json));assert.equal(invoked,0);
});
test('invalid stage/start cannot revoke or replace an applied question',()=>{
 const events=[],session=new Session(e=>events.push(e)),question=q(),token=session.start(question);session.publish(token,{question,count:'1',basis:'held',state:'observed'});const before=session.snapshot(),count=events.length;
 for(const bad of [{...q(),limit:101},{...q(),where:{all:[],value:''}},{...q(),spanStatus:true}])for(const action of ['stage','start']){assert.throws(()=>session[action](bad));assert.deepEqual(session.snapshot(),before);assert.equal(events.length,count);}
});
test('history retains an exact legal continuation and refuses explicit invalid views',()=>{
 const question={...q(),limit:17,offset:34},before=structuredClone(question);assert.deepEqual(normalizeIntent({kind:'question',question}).question,before);assert.deepEqual(P.readQuestionLink(P.questionLink(question)).question,before);
 for(const view of [null,'',false])assert.throws(()=>normalizeIntent({kind:'question',question,view}));assert.deepEqual(question,before);
});

test('omitted offset adopts the explicit Go zero without changing authored configuration',()=>{
 const question=q();delete question.offset;const before=structuredClone(question),session=new Session(),token=session.start(question);
 assert.equal(session.publish(token,{question:{...question,offset:0},count:'1',state:'observed'}),true);assert.deepEqual(question,before);
});

test('qualified empty namespaces adopt Go omission while keeping missing distinct',()=>{
 const question={...q(),signal:'spans',reportedEdge:{parent:{service:'api',namespace:{mode:'value',value:''}},child:{service:'worker',namespace:{mode:'value',value:''}},discoveryNamespace:{mode:'value',value:''}}};
 const before=structuredClone(question),response=structuredClone(question);for(const ns of [response.reportedEdge.parent.namespace,response.reportedEdge.child.namespace,response.reportedEdge.discoveryNamespace])delete ns.value;
 const session=new Session(),token=session.start(question);assert.equal(session.publish(token,{question:response,count:'1'}),true);assert.deepEqual(question,before);
 const different=structuredClone(response);different.reportedEdge.child.namespace={mode:'missing'};assert.notEqual(identity(different),identity(response));
});
