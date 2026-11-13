import {test} from 'node:test';
import assert from 'node:assert/strict';
import {plan,resolve,remove} from '../ui/field-constraints.mjs';
import {identity} from '../ui/question-session.mjs';
const field={kind:'attribute',key:'cohort'},resource={kind:'resource',key:'cohort'},leaf=(value,f=field)=>({field:structuredClone(f),op:'eq',value});
const q=where=>({signal:'logs',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},fromMs:1000,toMs:61000,limit:75,offset:0,...(where?{where}:{})});
function matches(p,row){if(!p)return true;if(p.all)return p.all.every(c=>matches(c,row));if(p.any)return p.any.some(c=>matches(c,row));const f=p.numeric?.field||p.field,v=row[f.kind]?.[f.key];if(p.op==='missing')return v===undefined;if(p.op==='exists')return v!==undefined;if(p.numeric)return v!==undefined&&v!==''&&BigInt(v)>=BigInt(p.value);return v===p.value;}

test('424 exact namespace and authored clause paths include numeric fields and OR edit-only entries',()=>{
 const numeric={numeric:{field,type:'int64',unit:'1',parse:'canonical-decimal'},op:'gte',value:'5'},question=q({all:[leaf('resource',resource),leaf('a'),{any:[leaf('b'),leaf('c')]},numeric]});
 const intent=plan(question,field);
 assert.deepEqual(intent.clauses.map(c=>c.path),[['all',1],['all',2,'any',0],['all',2,'any',1],['all',3]]);
 assert.deepEqual(intent.clauses.map(c=>c.removable),[true,false,false,true]);
 assert.deepEqual(intent.clauses.at(-1).clause,numeric);
 assert.deepEqual(plan(question,resource).clauses.map(c=>c.path),[['all',0]]);
 assert.equal(plan(q(),field).clauses.length,0);
 assert.deepEqual(resolve(question,intent,intent.clauses[2]),['all',2,'any',1]);
 assert.throws(()=>remove(question,intent,intent.clauses[1]),/OR group/);
});

test('424 removal changes only the chosen conjunction branch and retains full scope metadata',()=>{
 const question=q({all:[leaf('api',{kind:'column',key:'service'}),{all:[leaf('a')]},{any:[leaf('b',resource),leaf('c',resource)]}]}),before=structuredClone(question),intent=plan(question,field),next=remove(question,intent,intent.clauses[0]);
 assert.deepEqual(next,{...before,where:{all:[before.where.all[0],before.where.all[2]]}});
 assert.deepEqual(question,before);assert.deepEqual(next.environmentSelector,before.environmentSelector);
 next.environmentSelector.scope.value='changed';assert.deepEqual(question,before);
 const lone=q(leaf('a')),single=plan(lone,field);assert.deepEqual(remove(lone,single,single.clauses[0]),q());
 const duplicate=q({all:[leaf('a'),leaf('a'),leaf('b',resource)]}),dupPlan=plan(duplicate,field);
 assert.deepEqual(remove(duplicate,dupPlan,dupPlan.clauses[1]).where,{all:[leaf('a'),leaf('b',resource)]});
});

test('424 stale question, path, content and forged eligibility refuse atomically',()=>{
 const question=q({all:[leaf('a'),leaf('b'),{any:[leaf('c')]}]}),intent=plan(question,field),before=structuredClone({question,intent});
 for(const change of [x=>x.toMs++,x=>x.environmentSelector.policy='current-only',x=>x.where.all.reverse(),x=>x.where.all[0].value='other']){const stale=structuredClone(question);change(stale);assert.throws(()=>resolve(stale,intent,intent.clauses[0]),/changed/);assert.throws(()=>remove(stale,intent,intent.clauses[0]),/changed/);}
 for(const change of [x=>x.path[1]=1,x=>x.clause.value='other',x=>x.clauseKey='other',x=>x.removable=false,x=>x.reason='other']){const choice=structuredClone(intent.clauses[0]);change(choice);assert.throws(()=>resolve(question,intent,choice),/changed/);assert.throws(()=>remove(question,intent,choice),/changed/);}
 const forged=structuredClone(intent.clauses[2]);forged.removable=true;forged.reason='';assert.throws(()=>remove(question,intent,forged),/changed/);
 assert.deepEqual({question,intent},before);
 // Canonical question identity intentionally equates empty authored values;
 // clause custody additionally binds exact authored content.
 const exists=q({field,op:'exists'}),existsPlan=plan(exists,field),changed=structuredClone(exists);changed.where.value='';assert.equal(identity(changed),identity(exists));assert.throws(()=>resolve(changed,existsPlan,existsPlan.clauses[0]),/changed/);
});

test('424 generated conjunctive removal is monotone and preserves unrelated predicates',()=>{
 for(let depth=0;depth<4;depth++)for(const selected of [null,'','a','b']){
  let where=selected===null?{field,op:'missing'}:leaf(selected);
  for(let level=0;level<depth;level++)where={all:[leaf('api',{kind:'column',key:'service'}),where,leaf('prod',resource)]};
  const question=q(where),before=structuredClone(question),intent=plan(question,field);assert.equal(intent.clauses.length,1);const next=remove(question,intent,intent.clauses[0]);
  for(const value of [undefined,'','a','b','c'])for(const service of ['api','worker'])for(const environment of ['prod','dev']){
   const row={attribute:value===undefined?{}:{cohort:value},column:{service},resource:{cohort:environment}};
   if(matches(question.where,row))assert(matches(next.where,row),'removal must not narrow the original population');
   assert.equal(matches(next.where,row),depth===0||service==='api'&&environment==='prod');
  }
  assert.deepEqual(question,before);
 }
});

test('424 invalid supplied objects refuse without evaluating getters or mutating input',()=>{
 let reads=0;const bad=q(leaf('a'));Object.defineProperty(bad.where,'value',{enumerable:true,get(){reads++;return 'a'}});assert.throws(()=>plan(bad,field));
 const question=q(leaf('a')),intent=plan(question,field),choice=structuredClone(intent.clauses[0]);Object.defineProperty(choice.clause,'value',{enumerable:true,get(){reads++;return 'a'}});assert.throws(()=>resolve(question,intent,choice));
 const path=structuredClone(intent.clauses[0]);path.path=[{get all(){reads++;return 'all'}}];assert.throws(()=>resolve(question,intent,path));
 assert.equal(reads,0);assert.deepEqual(question,q(leaf('a')));
});
