import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as QuestionProtocol from '../ui/question-protocol.mjs';
// Execute the production closures in their real lexical names. Full browser
// journeys independently exercise the mounted application and physical reads.
const distributionSource=readFileSync(new URL('../ui/distributions.mjs',import.meta.url),'utf8');
const explorationSource=readFileSync(new URL('../ui/exploration.mjs',import.meta.url),'utf8');
const line=name=>distributionSource.split('\n').find(line=>line.trimStart().startsWith('function '+name+'(')).trim();
function fixture(){return{question:{signal:'spans',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'empty'}},fromMs:1000,toMs:5000,limit:19,offset:0,numeric:{field:{kind:'attribute',key:'size'},type:'int64',unit:'items',parse:'canonical-decimal'},numericHistogram:{binning:'powers-of-two'},where:{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''}}}}
test('actual distribution producer retires same-data ABA and requires a current visible owner',()=>{
 const data=fixture(),calls=[],env={current:data,populationGeneration:1,distributionGeneration:1,container:{},UI:{actionVisible:element=>env.visible&&!element.hidden},visible:true,numericCanAct:()=>env.active,active:true,openPopulation:(...args)=>{calls.push(args);return true},numericOwner:{suspend(){calls.push('numeric suspended')}},durationOwner:null,timelineView:null,distributionView:null};
 vm.createContext(env);vm.runInContext(line('populationOpen')+'\n'+line('suspendActions'),env);
 const old=env.populationOpen(data),q=data.question,opener={closest(){return this.inert?this:null}};assert.equal(old(q,null,null,opener),true);assert.equal(calls[0][3],data);assert.equal(calls[0][4],opener);
 env.suspendActions();assert.equal(env.populationGeneration,2);assert.equal(calls.pop(),'numeric suspended');env.current=fixture();env.current=data;assert.equal(old(q),false,'same data object after Back cannot revive old producer');
 const fresh=env.populationOpen(data);env.visible=false;assert.equal(fresh(q),false);env.visible=true;env.active=false;assert.equal(fresh(q),false);env.active=true;assert.equal(fresh(q),true);assert.equal(calls.length,2);opener.hidden=true;assert.equal(fresh(q,null,null,opener),false);opener.hidden=false;opener.inert=true;assert.equal(fresh(q,null,null,opener),false);assert.equal(calls.length,2);
});
test('actual exploration continuation stages cloned whole typed questions only from the current source owner',()=>{
 const sourceLine=explorationSource.split('\n').find(line=>line.includes('openPopulation:(q,baseline,rest,origin,opener)=>'));
 const callback=sourceLine.slice(sourceLine.indexOf('openPopulation:')+'openPopulation:'.length,sourceLine.indexOf(',openTrace:',sourceLine.indexOf('openPopulation:')));
 const data=fixture(),events=[],env={lastResult:data,scopeActionsLive:true,plotReady:()=>env.ready,ready:true,UI:{actionVisible:()=>env.visible},visible:true,container:{},QuestionProtocol,structuredClone,CustomEvent:class{constructor(type,options){this.type=type;this.detail=options.detail}},global:{dispatchEvent:event=>events.push(event)}};
 const open=vm.runInNewContext('('+callback+')',env),opener={};assert.equal(open(data.question,null,null,data,opener),true);assert.equal(events[0].detail.stage,true);assert.deepEqual(events[0].detail.question,data.question);assert.equal(events[0].detail.opener,opener);
 data.question.where.value='caller mutation';assert.equal(events[0].detail.question.where.value,'');
 for(const change of [()=>env.scopeActionsLive=false,()=>{env.scopeActionsLive=true;env.ready=false},()=>{env.ready=true;env.visible=false},()=>{env.visible=true;env.lastResult=fixture()}]){change();assert.equal(open(data.question,null,null,data),false)}
 assert.equal(events.length,1);
});

test('shared typed time and duration transforms preserve the parent policy and namespace predicates',async()=>{
 const {question,partition}=await import('../ui/population-selection.mjs');
 for(const signal of ['logs','spans']){
  const base=fixture().question;delete base.numeric;delete base.numericHistogram;base.signal=signal;base.offset=17;
  const before=structuredClone(base),selection={fromMs:1500,toMs:4000,...(signal==='spans'?{minMS:'0.000001',maxMS:'1'}:{})},selected=question(base,selection);
  assert.deepEqual(base,before);assert.deepEqual(selected.environmentSelector,base.environmentSelector);assert.deepEqual(signal==='spans'?selected.where.all[0]:selected.where,base.where);assert.equal(selected.offset,0);assert.equal(selected.fromMs,1500);assert.equal(selected.toMs,4000);QuestionProtocol.validateRecordQuestion(selected);
  if(signal==='spans'){assert.deepEqual(selected.where.all.slice(1).map(p=>[p.op,p.value]),[['gte','1'],['lte','999999']]);const pair=partition(base,selection);for(const q of [pair.selected,pair.rest]){assert.deepEqual(q.environmentSelector,base.environmentSelector);assert.deepEqual(q.where.all[0],base.where);assert.equal(q.fromMs,base.fromMs);assert.equal(q.toMs,base.toMs);QuestionProtocol.validateRecordQuestion(q)}}
 }
});
