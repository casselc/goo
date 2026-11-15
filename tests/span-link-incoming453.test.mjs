import {test} from 'node:test';
import assert from 'node:assert/strict';
import {planIncoming} from '../ui/span-link-intent.mjs';
import {matches,validateTarget} from '../ui/span-link-membership.mjs';
import {validateRecordQuestion,recordStep,stepLink,readStepLink,captureQuestion} from '../ui/question-protocol.mjs';
import {Session,identity} from '../ui/question-session.mjs';
import * as Environment from '../ui/record-environment.mjs';
const target={traceId:'a'.repeat(32),spanId:'b'.repeat(16)},other={traceId:'c'.repeat(32),spanId:'d'.repeat(16)};
const origin=()=>({signal:'spans',environment:'prod',fromMs:1000,toMs:61000,where:{field:{kind:'column',key:'service'},op:'eq',value:'checkout'},groupBy:[{kind:'column',key:'service'}],summary:'span-duration',limit:17,offset:34});
const row=()=>({TraceId:target.traceId,SpanId:target.spanId});
const linked=(traces=[target.traceId],spans=[target.spanId],attributes=[{}])=>({linked_traces:traces,linked_spans:spans,linked_attributes:attributes});
test('incoming planner keeps exact environment/window and drops prior selection/projection without IO',()=>{
 for(const scope of [{environment:'prod'},{environmentSelector:Environment.current('prod')},{environmentSelector:Environment.effective({mode:'named',value:'prod'})},{environmentSelector:Environment.effective({mode:'empty'})},{environmentSelector:Environment.effective({mode:'unreported'})}]){
  const q=origin();delete q.environment;Object.assign(q,scope);const input={origin:q,rows:[row()],rowIndex:0},before=structuredClone(input),out=planIncoming(input);
  assert.equal(out.relation,'reported-span-link-incoming');assert.deepEqual(out.question,{signal:'spans',...scope,fromMs:1000,toMs:61000,linksTo:target,limit:100,offset:0});assert.deepEqual(out.origin,{question:q,rowIndex:0,source:target});assert.deepEqual(input,before);
  out.question.linksTo.spanId='changed';out.target.traceId='changed';out.origin.source.spanId='changed';out.origin.question.where.value='changed';assert.deepEqual(input,before);assert.equal(out.readAt,undefined);assert.equal(out.custody,undefined);
 }
});
test('membership pairs the trace and span at the same link position and counts repeated links once',()=>{
 assert.equal(matches(linked(),target),true);assert.equal(matches(linked([target.traceId,other.traceId],[other.spanId,target.spanId],[{},{}]),target),false);
 assert.equal(matches(linked([target.traceId,target.traceId],[target.spanId,target.spanId],[{},{}]),target),true);
 assert.equal(matches(linked([other.traceId,target.traceId],[other.spanId,target.spanId],[null,{}]),target),true,'unrelated malformed entry does not hide valid pair');
 assert.equal(matches(linked([target.traceId,target.traceId],[target.spanId,target.spanId],[null,{}]),target),true,'any valid matching position is sufficient');
 for(const bad of [{},{...linked(),linked_spans:[]},{...linked(),linked_attributes:[]},linked([target.traceId],[target.spanId],[null]),linked([target.traceId],[target.spanId],[{bad:1}])])assert.equal(matches(bad,target),false);
 const sparse=linked();delete sparse.linked_spans[0];assert.equal(matches(sparse,target),false);
});
test('typed membership survives shared questions, steps, captures and session identities',()=>{
 const q=planIncoming({origin:origin(),rows:[row()],rowIndex:0}).question,before=structuredClone(q);validateRecordQuestion(q);const step=recordStep(q);assert.deepEqual(readStepLink(stepLink(step)),step);assert.deepEqual(captureQuestion(step).explore,q);
 const session=new Session(),token=session.start(q);assert.equal(session.publish(token,{question:structuredClone(q),basis:'held',count:'1',state:'observed'}),true);assert.notEqual(identity(q),identity({...q,linksTo:other}));assert.notEqual(identity(q),identity({...q,linksTo:undefined}));assert.deepEqual(q,before);
});
test('invalid tuples, mixed signals, forged fields, null selectors and bad origins refuse',()=>{
 const q=planIncoming({origin:origin(),rows:[row()],rowIndex:0}).question;
 for(const bad of [null,{},[],{...target,traceId:'0'.repeat(32)},{...target,spanId:'0'.repeat(16)},{...target,traceId:'A'.repeat(32)},{...target,spanId:'b'.repeat(15)},{...target,extra:true}]){assert.throws(()=>validateTarget(bad));assert.throws(()=>validateRecordQuestion({...q,linksTo:bad}))}
 assert.throws(()=>validateRecordQuestion({...q,signal:'logs'}));for(const input of [{origin:{...origin(),signal:'logs'},rows:[row()],rowIndex:0},{origin:origin(),rows:[{...row(),SpanId:''}],rowIndex:0},{origin:origin(),rows:[row()],rowIndex:1},{origin:origin(),rows:[row()],rowIndex:-1}])assert.throws(()=>planIncoming(input));
});
test('counted-value projections retain membership and draft recovery retains the exact selector',async()=>{
 const {query}=await import('../ui/filter-values.mjs'),{Journal}=await import('../ui/investigation-draft.mjs'),q=planIncoming({origin:origin(),rows:[row()],rowIndex:0}).question;
 const projected=query({...q,where:origin().where},{kind:'attribute',key:'release'});assert.deepEqual(projected.linksTo,target);projected.linksTo.traceId='changed';assert.deepEqual(q.linksTo,target);assert.throws(()=>query({...q,linksTo:null},{kind:'attribute',key:'release'}));assert.throws(()=>query({...q,signal:'logs'},{kind:'attribute',key:'release'}));
 const draft={parent:null,document:{title:'Links',question:'',finding:'',openQuestions:'',cards:[{kind:'question',step:recordStep(q)}]},composer:{role:'finding',text:'',editing:null,captureID:'',caption:'',destination:'',comparisonID:'',comparisonFinding:''},openedQuestion:''},values=new Map(),storage={getItem:key=>values.get(key)??null,setItem:(key,value)=>values.set(key,value),removeItem:key=>values.delete(key)},options={storage,path:'/',now:()=>1700000000000,id:()=> 'e'.repeat(64)},journal=new Journal(options);journal.inspect();journal.checkpoint(draft);assert.deepEqual(new Journal(options).inspect().entry.draft,draft);
});
