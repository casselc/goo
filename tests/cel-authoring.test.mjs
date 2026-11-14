import {test} from 'node:test';
import assert from 'node:assert/strict';
import {AuthoringSlot,recordCELContract} from '../ui/cel-authoring.mjs';

const clone=value=>structuredClone(value);
const base=()=>({signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:20,offset:4,where:{any:[{field:{kind:'attribute',key:'team'},op:'eq',value:'edge'},{field:{kind:'column',key:'service'},op:'eq',value:'api'}]}});
const options=(overrides={})=>({base:base(),baseIdentity:'a'.repeat(64),parentRevision:'parent-1',slot:'condition',epoch:'mount-1',...overrides});
function receipt(check,value='ok'){
 const predicate={field:{kind:'column',key:'body'},op:'eq',value};
 const question=clone(check.request.base);
 question.where=question.where?{all:[question.where,predicate]}:predicate;
 return {loweringDigest:'b'.repeat(64),version:'goo.cel-plan/1',celLibrary:recordCELContract.library,expression:check.request.expression,predicate,question,binding:clone(check.request.binding),plan:{rows:'parameterized'}};
}
const diagnosis={code:'syntax',phase:'parse',message:'Expected expression.',span:{start:2,end:2,line:1,column:3}};
function check(owner,text='log.body == "ok"'){
 owner.edit(text);const pending=owner.beginCheck();assert.ok(pending);return pending;
}

test('authoring checks exact wire binding and retains unsupported immutable parent',()=>{
 const input=options(),owner=new AuthoringSlot(input),pending=check(owner);
 input.base.where.any[0].value='caller changed';
 assert.deepEqual(pending.request.binding,{...recordCELContract,signal:'logs',baseIdentity:'a'.repeat(64),slot:'condition',epoch:'mount-1',revision:'2'});
 assert.equal(owner.accept(pending.token,receipt(pending)),true);
 const question=owner.stage(pending.token);
 assert.deepEqual(question.where.all[0],base().where);
 assert.equal(question.offset,4);
 question.where.all[0].any[0].value='consumer changed';
 assert.equal(owner.stage(pending.token).where.all[0].any[0].value,'edge');
 assert.equal(owner.stage(),null);
 const state=owner.snapshot();state.lastValid.question.where.all[0].any=[];
 assert.equal(owner.stage(pending.token).where.all[0].any.length,2);
});

test('repeated edits replace one condition without accumulating depth',()=>{
 const owner=new AuthoringSlot(options());
 for(let i=0;i<20;i++){
  const pending=check(owner,`log.body == "${i}"`);
  assert.equal(owner.accept(pending.token,receipt(pending,String(i))),true);
  assert.deepEqual(owner.stage(pending.token).where,{all:[base().where,{field:{kind:'column',key:'body'},op:'eq',value:String(i)}]});
 }
 const restore=owner.clear();assert.deepEqual(owner.stage(restore),base());
 assert.equal(owner.snapshot().text,'');assert.ok(owner.snapshot().lastValid);
});

test('text ABA and repeated checks cannot revive deferred success or failure',()=>{
 const owner=new AuthoringSlot(options()),old=check(owner,'A');
 check(owner,'B');const current=check(owner,'A');
 assert.equal(owner.accept(old.token,receipt(old)),false);
 assert.equal(owner.fail(old.token,diagnosis),false);
 assert.equal(owner.accept({},receipt(current)),false);
 assert.equal(owner.accept(current.token,receipt(current)),true);
 const next=owner.beginCheck();
 assert.equal(owner.stage(current.token),null);
 assert.equal(owner.accept(current.token,receipt(current)),false);
 assert.equal(owner.fail(current.token,diagnosis),false);
 assert.equal(owner.fail(next.token,diagnosis),true);
 assert.equal(owner.stage(next.token),null);
 assert.equal(owner.snapshot().lastValid.expression,'A');
});

test('invalid text survives mode changes and visual adoption needs explicit discard',()=>{
 const owner=new AuthoringSlot(options()),valid=check(owner);
 owner.accept(valid.token,receipt(valid));
 const invalid=check(owner,'log.body ==');owner.fail(invalid.token,diagnosis);
 owner.setMode('conditions');owner.setMode('expression');
 assert.equal(owner.snapshot().text,'log.body ==');
 assert.deepEqual(owner.snapshot().diagnostic,diagnosis);
 assert.equal(owner.stage(valid.token),null);
 assert.equal(owner.adoptVisual('log.body == "new"'),false);
 assert.equal(owner.snapshot().text,'log.body ==');
 assert.equal(owner.adoptVisual('log.body == "new"',{discardDraft:true}),true);
 assert.equal(owner.snapshot().mode,'conditions');assert.equal(owner.snapshot().text,'log.body == "new"');
 assert.equal(owner.snapshot().phase,'draft');assert.equal(owner.stage(invalid.token),null);
 assert.ok(owner.snapshot().lastValid);
});

test('hide, restore, changed parent and disposal retire direct staging callbacks',()=>{
 const owner=new AuthoringSlot(options());
 const first=check(owner);owner.accept(first.token,receipt(first));
 const oldCallback=()=>owner.stage(first.token);
 owner.setVisible(false);assert.equal(oldCallback(),null);assert.equal(owner.beginCheck(),null);
 owner.setVisible(true);assert.equal(oldCallback(),null);
 const second=check(owner);owner.accept(second.token,receipt(second));
 assert.equal(oldCallback(),null);
 const changed=options({base:{...base(),signal:'spans',where:{field:{kind:'column',key:'service'},op:'eq',value:'other'}},baseIdentity:'b'.repeat(64),parentRevision:'parent-2'});
 owner.replaceParent(changed);assert.equal(owner.stage(second.token),null);
 owner.replaceParent(options());assert.equal(owner.accept(second.token,receipt(second)),false);
 const third=check(owner);owner.accept(third.token,receipt(third));owner.dispose();
 assert.equal(owner.stage(third.token),null);assert.equal(owner.beginCheck(),null);
 assert.equal(owner.accept(third.token,receipt(third)),false);assert.equal(owner.fail(third.token,diagnosis),false);
 assert.equal(owner.edit('new'),false);assert.equal(owner.clear(),false);assert.equal(owner.setVisible(true),false);
 assert.equal(owner.snapshot().text,third.request.expression);assert.ok(owner.snapshot().lastValid);
});

test('parent ABA and parent revision change reject originating errors',()=>{
 const owner=new AuthoringSlot(options()),a=check(owner);
 owner.replaceParent(options({baseIdentity:'b'.repeat(64),parentRevision:'parent-B'}));
 owner.replaceParent(options());
 assert.equal(owner.fail(a.token,diagnosis),false);
 const current=owner.beginCheck();owner.replaceParent(options({parentRevision:'parent-2'}));
 assert.equal(owner.accept(current.token,receipt(current)),false);
});

test('receipt must bind exact source, full compiler contract and whole parent meaning',()=>{
 for(const mutate of [
  r=>r.binding.library='other',r=>r.binding.revision='rounded',r=>r.binding.epoch='other',
  r=>r.binding.signal='spans',r=>r.binding.baseIdentity='b'.repeat(64),r=>r.binding.target='other',
  r=>r.binding.slot='other',r=>r.binding.extra=true,r=>r.version='other',r=>r.expression='other',
  r=>r.celLibrary='other',r=>r.question.environment='other',r=>r.question.where=r.predicate,
  r=>r.question.where.all.reverse(),r=>r.question.offset=0,r=>delete r.predicate,
 ]){
  const owner=new AuthoringSlot(options()),pending=check(owner),out=receipt(pending);mutate(out);
  assert.equal(owner.accept(pending.token,out),false);assert.equal(owner.stage(pending.token),null);
  assert.equal(owner.accept(pending.token,receipt(pending)),true);
 }
});

test('public request mutation cannot rewrite the private originating request',()=>{
 const owner=new AuthoringSlot(options()),pending=check(owner),valid=receipt(pending);
 pending.request.binding.revision='9007199254740993';pending.request.base.environment='tampered';pending.request.expression='other';
 assert.equal(owner.accept(pending.token,valid),true);
 assert.equal(owner.stage(pending.token).environment,'prod');
});

test('owner tokens admit well formed bounded strings and mode inputs are closed',()=>{
 for(const overrides of [{epoch:'bad\u0000token'},{slot:'a b'},{parentRevision:'\ud800'},{epoch:'é'.repeat(65)},{baseIdentity:'bad'}])assert.throws(()=>new AuthoringSlot(options(overrides)));
 const owner=new AuthoringSlot(options());assert.throws(()=>owner.setMode('raw'));assert.throws(()=>owner.edit(null));
 const pending=check(owner);assert.throws(()=>owner.fail(pending.token,{}));
 assert.equal(owner.snapshot().phase,'checking');
});
