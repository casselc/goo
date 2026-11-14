import test from 'node:test';
import assert from 'node:assert/strict';
import {actions,intent,predicate,reason} from '../ui/field-refinement.mjs';
import {selectionQueries} from '../ui/categorical-selection.mjs';
import {ScopeFilters} from '../ui/stream-entry.mjs';

// Independent ordinary-map oracle: inequality is reported-only, exclusion isn't.
function matches(p,map){
 if(p.all)return p.all.every(c=>matches(c,map));
 if(p.any)return p.any.some(c=>matches(c,map));
 const present=Object.hasOwn(map,p.field.key),v=map[p.field.key];
 if(p.op==='exists')return present;
 if(p.op==='missing')return !present;
 return present&&(p.op==='eq'?v===(p.value??''):v!==(p.value??''));
}

test('include/exclude partition missing, empty and literal values for both attribute maps and signals',()=>{
 for(const signal of ['logs','spans'])for(const kind of ['resource','attribute'])for(const value of [null,'','west','0','missing','雪','quote"\\']){
  const field={kind,key:'example.zone'},before=structuredClone(field),yes=predicate(field,value,'include',signal),no=predicate(field,value,'exclude',signal);
  for(const map of [{},{'example.zone':''},{'example.zone':'west'},{'example.zone':'0'},{'example.zone':'missing'},{'example.zone':'雪'},{'example.zone':'quote"\\'}]){
   assert.notEqual(matches(yes,map),matches(no,map));
   assert.equal(matches(yes,map),value===null?!Object.hasOwn(map,field.key):Object.hasOwn(map,field.key)&&map[field.key]===value);
  }
  assert.deepEqual(field,before);
  assert.deepEqual(predicate(field,null,'reported',signal),{field,op:'exists'});
  assert.deepEqual(predicate(field,null,'missing',signal),{field,op:'missing'});
 }
});

test('presence canonicalization is bounded and does not infer missing from empty',()=>{
 const field={kind:'attribute',key:'http.route'};
 assert.deepEqual(intent(field,null,'reported'),{field,value:null,mode:'exclude'});
 assert.deepEqual(intent(field,null,'missing'),{field,value:null,mode:'include'});
 assert.deepEqual(predicate(field,'','include'),{field,op:'eq'});
 assert.deepEqual(actions(field).map(a=>a.mode),['include','exclude','reported','missing']);
 assert.deepEqual(actions({kind:'column',key:'service'}).map(a=>a.mode),['include','exclude']);
 assert.deepEqual(actions({kind:'invalid',key:'x'}),[]);
 for(const [f,v,m]of [[field,'','reported'],[field,undefined,'include'],[field,1,'include'],[field,'x','other'],[{kind:'column',key:'service'},null,'missing'],[{kind:'attribute',key:'x'.repeat(129)},'x','include'],[field,'x'.repeat(1025),'exclude'],[field,'\0','include']])assert.notEqual(reason(f,v,m),'');
 assert.notEqual(reason({kind:'column',key:'name'},'work','include','logs'),'');
 assert.equal(reason({kind:'column',key:'name'},'work','include','spans'),'');
});

test('Activity and categorical refinement agree without mutating parent or applied state',()=>{
 const field={kind:'attribute',key:'example.zone'},base={signal:'logs',environment:'prod',fromMs:1000,toMs:2000,where:{field:{kind:'resource',key:'service.namespace'},op:'eq',value:'payments'}},snapshot=structuredClone(base);
 for(const value of [null,'','west'])for(const mode of ['include','exclude','reported','missing']){
  if(['reported','missing'].includes(mode)&&value!==null)continue;
  const activity=new ScopeFilters();activity.restore(base.where);const old=activity.where(true);activity.stage(field,value,mode,'logs');
  const fixed=selectionQueries([base],[{signal:'logs',field,value,mode}])[0];
  assert.deepEqual(activity.where(),fixed.where);
  assert.deepEqual(activity.where(true),old);
  assert.equal(activity.dirty,true);
  assert.deepEqual(base,snapshot);
 }
});

test('refused Activity changes leave staged and applied predicates intact',()=>{
 const field={kind:'attribute',key:'example.zone'},filters=new ScopeFilters();filters.stage(field,'west','include');filters.apply();
 for(const args of [[field,{},'include'],[field,'x'.repeat(1025),'exclude'],[{kind:'column',key:'service'},null,'reported']]){
  const snapshot=structuredClone({draft:filters.draft,applied:filters.applied,where:filters.where()});assert.throws(()=>filters.stage(...args));assert.deepEqual({draft:filters.draft,applied:filters.applied,where:filters.where()},snapshot);
 }
 const full=new ScopeFilters();for(let i=0;i<10;i++)full.stage({kind:'attribute',key:'field'+i},'v','exclude');const snapshot=structuredClone(full.draft);assert.throws(()=>full.stage({kind:'attribute',key:'overflow'},'v','exclude'),/32/);assert.deepEqual(full.draft,snapshot);
});
