import {test} from 'node:test';
import assert from 'node:assert/strict';
import {ScopeFilters} from '../ui/scope-filters.mjs';
import {PredicateDraft} from '../ui/predicate-draft.mjs';
import {validate} from '../ui/predicate-validation.mjs';
const field={kind:'resource',key:'service.instance.id'},empty={field,op:'eq',value:''},missing={field,op:'missing'};
test('451 shared scope Discard restores applied base and chips, including exact empty/missing distinction',()=>{
 const scope=new ScopeFilters();scope.restore(empty);scope.stage({kind:'resource',key:'service.namespace'},null,'missing','logs');scope.apply();const applied=scope.where(true),chips=scope.applied;
 scope.replace(missing);scope.stage({kind:'attribute',key:'region'},'west','include','logs');assert.equal(scope.dirty,true);scope.discard();
 assert.deepEqual(scope.where(),applied);assert.deepEqual(scope.where(true),applied);assert.deepEqual(scope.applied,chips);assert.deepEqual(scope.draft,chips);assert.equal(scope.dirty,false);assert.deepEqual(scope.base,empty);
 const exposed=scope.where();exposed.all[0].field.key='mutated';assert.deepEqual(scope.where(),applied);
});
test('451 shared scope error/pending drafts refuse Apply and Discard never clears applied restrictions',()=>{
 for(const pending of [false,true]){const scope=new ScopeFilters();scope.restore(missing);scope.replace(empty);scope.report(pending?null:'Invalid predicate',pending);assert.equal(scope.blocked,true);assert.throws(()=>scope.apply());assert.deepEqual(scope.where(true),missing);scope.discard();assert.deepEqual(scope.where(),missing);assert.equal(scope.blocked,false);assert.equal(scope.dirty,false);}
});
test('451 semantic empty value normalization is clean but missing remains a distinct draft',()=>{
 const state=new PredicateDraft({where:empty,validate:p=>validate(p,'logs')});state.stage({op:'eq',field:{key:field.key,kind:field.kind}});assert.equal(state.dirty(),false);
 state.stage(missing);assert.equal(state.dirty(),true);const token=state.token();state.stage(empty);state.stage(missing);assert.equal(state.accept(token),false);assert.deepEqual(state.applied(),empty);assert.equal(state.accept(state.token()),true);assert.deepEqual(state.applied(),missing);
});
