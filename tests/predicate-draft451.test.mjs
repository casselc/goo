import {test} from 'node:test';
import assert from 'node:assert/strict';
import {PredicateDraft} from '../ui/predicate-draft.mjs';
import {validate} from '../ui/predicate-validation.mjs';
const field={kind:'attribute',key:'region'},leaf=(op,value)=>({field:{...field},op,...(value===undefined?{}:{value})}),make=where=>new PredicateDraft({where,validate:p=>validate(p,'spans')});

test('applied/draft lifecycle preserves authored OR, missing/empty and exact uint64',()=>{
 const original={any:[leaf('missing'),leaf('eq','')]},next={all:[original,{field:{kind:'column',key:'duration_ns'},op:'gte',value:'18446744073709551615'}]},state=make(original);
 state.stage(next);assert.equal(state.dirty(),true);assert.deepEqual(state.applied(),original);assert.deepEqual(state.candidate(),next);
 const token=state.token();assert.equal(state.accept(token),true);assert.equal(state.accept(token),false);assert.equal(state.dirty(),false);assert.deepEqual(state.applied(),next);
 state.stage(leaf('eq','another'));state.discard();assert.deepEqual(state.draft(),next);assert.equal(state.dirty(),false);
 state.reset(undefined);assert.equal(state.applied(),undefined);assert.equal(state.draft(),undefined);assert.equal(state.dirty(),false);
});
test('canonical omission and property order are clean without rewriting literals or group order',()=>{
 const state=make(leaf('eq',''));state.stage({op:'eq',field:{key:'region',kind:'attribute'}});assert.equal(state.dirty(),false);
 state.stage(leaf('missing'));assert.equal(state.dirty(),true);state.reset({any:[leaf('missing'),leaf('eq','')]});state.stage({any:[leaf('eq',''),leaf('missing')]});assert.equal(state.dirty(),true);
});
test('invalid, error and pending drafts never erase applied condition',()=>{
 const original=leaf('exists'),state=make(original);
 for(const [draft,options] of [[{any:[]},{}],[leaf('eq','changed'),{error:'Compiler refusal'}],[undefined,{pending:true}]]){
  state.stage(draft,options);assert.equal(state.dirty(),true);assert.throws(()=>state.candidate());assert.equal(state.accept(state.token()),false);assert.deepEqual(state.applied(),original);state.discard();assert.equal(state.dirty(),false);
 }
 state.stage(leaf('eq','kept'),{error:new Error('Bad input')});assert.throws(()=>state.candidate(),/Bad input/);
 assert.throws(()=>state.reset({all:[]}));assert.deepEqual(state.applied(),original);assert.equal(state.snapshot().error.message,'Bad input');
});
test('every edit, discard, reset and retirement invalidates stale and ABA tokens',()=>{
 const a=leaf('eq','a'),b=leaf('eq','b'),state=make(a);
 state.stage(b);const token=state.token();state.stage(a);state.stage(b);assert.equal(state.accept(token),false);assert.deepEqual(state.applied(),a);
 for(const mutate of [()=>state.stage(b),()=>state.discard(),()=>state.reset(a),()=>state.retire()]){const stale=state.token();mutate();assert.equal(state.accept(stale),false)}
 state.stage(b,{pending:true});const before=state.snapshot();state.retire();const after=state.snapshot();assert.deepEqual({...after,revision:before.revision},before);
});
test('all inputs, candidates and snapshots are detached; validation has no mutation channel',()=>{
 const original=leaf('eq','a'),state=make(original);original.value='caller';assert.equal(state.applied().value,'a');
 const draft=leaf('eq','b');state.stage(draft);draft.field.key='caller';
 for(const value of [state.applied(),state.draft(),state.candidate(),state.snapshot().draft,state.snapshot().applied]){value.field.key='mutated';value.value='mutated'}
 assert.equal(state.applied().field.key,'region');assert.equal(state.draft().value,'b');
 const local=new PredicateDraft({where:leaf('exists'),validate:p=>{validate(p,'spans');if(p)p.field.key='validator-copy'}});local.stage(leaf('eq','c'));assert.equal(local.draft().field.key,'region');const candidate=local.candidate();candidate.value='caller-copy';assert.equal(local.draft().value,'c');
});
test('validator retirement during acceptance prevents a stale commit',()=>{
 let state,retire=false;state=new PredicateDraft({where:leaf('exists'),validate:p=>{validate(p,'spans');if(retire)state.retire()}});state.stage(leaf('eq','b'));retire=true;assert.equal(state.accept(state.token()),false);assert.equal(state.applied().op,'exists');
});

test('invalid drafts cannot appear clean through canonical null omission',()=>{
 for(const [applied,draft] of [[undefined,null],[leaf('eq',''),{...leaf('eq',''),numeric:null}]]){
  const state=make(applied);state.stage(draft);assert.equal(state.dirty(),true);assert.equal(typeof state.snapshot().error,'string');assert(state.snapshot().error.length>0);assert.deepEqual(state.draft(),draft);assert.deepEqual(state.applied(),applied);assert.throws(()=>state.candidate());assert.equal(state.accept(state.token()),false);state.discard();assert.equal(state.dirty(),false);
 }
});
