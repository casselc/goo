import assert from 'node:assert/strict';
import fs from 'node:fs';
import * as Constraints from '../../ui/field-constraints.mjs';
const samples=JSON.parse(fs.readFileSync(0,'utf8')),clone=structuredClone;
let reads=0;globalThis.fetch=()=>{reads++;throw Error('constraint planning acquired data')};
const key=f=>f.kind+':'+f.key;
function matches(p,row){if(!p)return true;if(p.all)return p.all.every(c=>matches(c,row));if(p.any)return p.any.some(c=>matches(c,row));const v=row.get(key(p.numeric?.field||p.field));if(p.op==='missing')return v===undefined;if(p.op==='exists')return v!==undefined;if(p.op==='eq')return v!==undefined&&v===(p.value??'');throw Error('unsupported independent oracle operator');}
for(const sample of samples){
 const field={kind:sample.Kind,key:'cohort'},other={kind:sample.Kind==='attribute'?'resource':'attribute',key:'cohort'},tenant={kind:'attribute',key:'tenant'},value=sample.Value;
 const chosen={field:clone(field),op:'eq',...(value===''?{}:{value})},retained={field:tenant,op:'eq',value:sample.Tenant},mixed={any:[{field:clone(field),op:'missing'},{field:clone(other),op:'eq',value:'blue'}]};
 const where={all:[clone(chosen),clone(retained),clone(mixed)]};if(sample.Duplicate)where.all.splice(1,0,clone(chosen));
 const question={signal:sample.Signal,environmentSelector:{version:'goo/environment-selector/v1',policy:sample.Policy,scope:{mode:'named',value:sample.Environment}},fromMs:sample.From,toMs:sample.From+60000,limit:100,offset:0,where};
 const before=clone(question),intent=Constraints.plan(question,field),choice=intent.clauses[0];assert.equal(intent.clauses.length,sample.Duplicate?3:2);assert.deepEqual(Constraints.resolve(question,intent,choice),['all',0]);assert.equal(choice.removable,true);
 const after=Constraints.remove(question,intent,choice),expected=clone(question);expected.where.all.shift();assert.deepEqual(after,expected,'only the exact authored conjunct may disappear');assert.deepEqual(question,before,'planning and removal must not mutate original question');
 // Independent finite-domain semantics distinguishes absent, present-empty and
 // literal alternatives, including namespace lookalikes and foreign tenants.
 for(const v of [undefined,'','Missing key',value,'other'])for(const peer of [undefined,'blue','green'])for(const t of [sample.Tenant,'foreign']){
  const row=new Map([[key(tenant),t]]);if(v!==undefined)row.set(key(field),v);if(peer!==undefined)row.set(key(other),peer);
  const original=matches(question.where,row),broader=matches(after.where,row);assert(!original||broader,'AND removal cannot lose a matching record');assert.equal(broader,matches(expected.where,row));
 }
 const orChoice=intent.clauses.at(-1);assert.equal(orChoice.removable,false);assert.throws(()=>Constraints.remove(question,intent,orChoice),/OR/);assert.deepEqual(Constraints.resolve(question,intent,orChoice),sample.Duplicate?['all',3,'any',0]:['all',2,'any',0]);
 // A changed complete question, not only a changed predicate, retires intent.
 for(const mutate of [q=>q.toMs++,q=>q.offset=1,q=>q.environmentSelector.policy=q.environmentSelector.policy==='effective'?'current-only':'effective',q=>q.where.all.reverse()]){const changed=clone(question);mutate(changed);assert.throws(()=>Constraints.resolve(changed,intent,choice));assert.throws(()=>Constraints.remove(changed,intent,choice));}
 for(const mutate of [c=>c.path[1]=99,c=>c.clause.value='forged',c=>c.clauseKey+='x',c=>c.removable=false]){const forged=clone(choice);mutate(forged);assert.throws(()=>Constraints.resolve(question,intent,forged));}
 // After one state transition the old choice cannot address its shifted sibling.
 assert.throws(()=>Constraints.remove(after,intent,choice));
 const next=Constraints.plan(after,field);if(sample.Duplicate){const selected=next.clauses.find(c=>c.removable);const twice=Constraints.remove(after,next,selected);assert.deepEqual(twice.where,{all:[retained,mixed]});assert.throws(()=>Constraints.resolve(twice,next,selected));}
 const numeric=clone(question);numeric.where.all[0]={numeric:{field:clone(field),type:'int64',unit:'items',parse:'canonical-decimal'},op:'gte',value:'0'};const numericPlan=Constraints.plan(numeric,field);assert.deepEqual(numericPlan.clauses[0].clause.numeric,numeric.where.all[0].numeric);const numericAfter=Constraints.remove(numeric,numericPlan,numericPlan.clauses[0]);const numericExpected=clone(numeric);numericExpected.where.all.shift();assert.deepEqual(numericAfter,numericExpected);
 assert.equal(reads,0);
}
process.stdout.write(JSON.stringify(samples.length));
