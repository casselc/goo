import {test} from 'node:test';
import assert from 'node:assert/strict';
import {plan,reason} from '../ui/counted-selection-comparison.mjs';
import * as Composition from '../ui/view-composition.mjs';
import * as Comparison from '../ui/captured-count-comparison.mjs';
import {identity} from '../ui/question-session.mjs';
const field={kind:'attribute',key:'cohort'},other={kind:'resource',key:'cohort'},tenant={kind:'attribute',key:'tenant'};
const leaf=(f,op,value)=>({field:structuredClone(f),op,...(value===undefined?{}:{value})});
const q=(where,signal='logs')=>({signal,environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},fromMs:1000,toMs:61000,limit:1,offset:0,groupBy:[structuredClone(field)],...(where?{where}:{})});
const sides=view=>({selected:view.queries[0],rest:Composition.comparison(view.composition).baselineQuery});
function matches(p,row){if(!p)return true;if(p.all)return p.all.every(child=>matches(child,row));if(p.any)return p.any.some(child=>matches(child,row));const k=p.field.kind+':'+p.field.key,has=Object.hasOwn(row,k),v=row[k];if(p.op==='missing')return !has;if(p.op==='exists')return has;if(p.op==='eq')return has&&v===(p.value??'');if(p.op==='ne')return has&&v!==(p.value??'');throw Error('Unknown independent oracle operator');}
function subsets(values){return Array.from({length:2**values.length-1},(_,bits)=>values.filter((_,i)=>(bits+1)&(1<<i)));}

test('425 existing fixed comparison grammar retains original context and independent operands',()=>{
 const where={all:[leaf(tenant,'eq','api'),{any:[leaf(other,'eq','blue'),leaf(field,'eq','common')]}]},question=q(where),before=structuredClone(question),values=[null,''];
 const view=plan(question,field,values),{selected,rest}=sides(view),declaration=Composition.comparison(view.composition);
 assert.equal(view.title,'Selected vs rest of scope');assert(Comparison.queryVariant(declaration));assert.equal(declaration.display.value,'count');assert.equal(declaration.display.view,'summary');
 assert.deepEqual(Composition.view(view),view);assert.equal(Composition.aggregate(view.composition).display.value,'count');
 const expectedContext={...before,limit:100,offset:0};delete expectedContext.where;for(const side of [selected,rest]){const context=structuredClone(side);delete context.where;assert.deepEqual(context,expectedContext);}
 assert.deepEqual(selected.where.all.slice(0,-1),where.all);assert.deepEqual(rest.where.all.slice(0,-1),where.all);assert.deepEqual(question,before);assert.deepEqual(values,[null,'']);
 selected.environmentSelector.scope.value='changed';rest.where.all[0].value='changed';assert.deepEqual(question,before);assert.equal(view.queries.length,1);
});

test('425 generated categorical selection and rest partition the same parent without losing namespaces or missing/empty/literal values',()=>{
 const alternatives=[null,'','Missing key','雪'],domain=[undefined,'','Missing key','雪','common','unseen'];
 for(const signal of ['logs','spans'])for(const f of [field,other])for(const values of subsets(alternatives))for(const where of [undefined,leaf(tenant,'eq','api'),{all:[leaf(tenant,'eq','api'),{any:[leaf(field,'eq','common'),leaf(other,'eq','blue')]}]}]){
  const parent=q(where,signal);parent.groupBy=[structuredClone(f)];const {selected,rest}=sides(plan(parent,f,values));
  for(const a of domain)for(const b of [...domain,'blue'])for(const t of ['api','foreign']){
   const row={'attribute:tenant':t};if(a!==undefined)row['attribute:cohort']=a;if(b!==undefined)row['resource:cohort']=b;
   const parentMatch=matches(parent.where,row),value=row[f.kind+':'+f.key],selectedValue=values.includes(value===undefined?null:value),inSelected=matches(selected.where,row),inRest=matches(rest.where,row);
   assert.equal(inSelected,parentMatch&&selectedValue);assert.equal(inRest,parentMatch&&!selectedValue);assert.equal(inSelected&&inRest,false);assert.equal(inSelected||inRest,parentMatch);
  }
 }
});

test('425 both selected and rest use the original fixed question rather than loaded counts or retained rows',()=>{
 const question=q(leaf(tenant,'eq','api')),old=plan(question,field,['Missing key']),oldSides=sides(old);
 for(const change of [q=>q.fromMs++,q=>q.toMs++,q=>q.environmentSelector.scope.value='stage',q=>q.environmentSelector.policy='current-only',q=>q.where.value='other']){
  const next=structuredClone(question);change(next);const fresh=sides(plan(next,field,['Missing key']));assert.notEqual(identity(fresh.selected),identity(oldSides.selected));assert.notEqual(identity(fresh.rest),identity(oldSides.rest));
 }
 assert.deepEqual(old,plan(question,field,['Missing key']));assert(!JSON.stringify(old).includes('capturedAt'));assert(!JSON.stringify(old).includes('baselineCapture'));
});

test('425 unsupported profiles, foreign grouping and inadmissible values refuse before constructing a comparison',()=>{
 for(const mutate of [q=>q.summary='span-duration',q=>q.spanStatus=true,q=>q.bucketMs=10000,q=>{q.groupTimeSeries=true;q.bucketMs=10000},q=>q.numeric={field,type:'int64',unit:'1',parse:'canonical-decimal'},q=>q.offset=1,q=>q.groupBy=[other],q=>q.groupBy.push(other)]){const parent=q(undefined,'spans');mutate(parent);assert.throws(()=>plan(parent,field,['a']));assert(reason(parent,field,['a']));}
 for(const values of [[],[null,null],['',''],[1],['\0'],['x'.repeat(1025)],Array(2)])assert.throws(()=>plan(q(),field,values));
 let read=0;const hostile=['a'];Object.defineProperty(hostile,'0',{get(){read++;return'a'},enumerable:true});assert.throws(()=>plan(q(),field,hostile));assert.equal(read,0);
 const timestamp={kind:'column',key:'timestamp_ns'},parent=q();parent.groupBy=[timestamp];assert.throws(()=>plan(parent,timestamp,['1']),/categorical string/);
});

test('425 full predicate budgets apply to both independently generated sides and refusal does not mutate input',()=>{
 const crowded=q({all:Array.from({length:28},(_,i)=>leaf({kind:'attribute',key:'keep'+i},'eq','x'))}),before=structuredClone(crowded),values=['a','b'];
 assert.throws(()=>plan(crowded,field,values),/32|nested/);assert.deepEqual(crowded,before);assert.deepEqual(values,['a','b']);
 const nested=q({any:[{any:[{any:[{any:[leaf(tenant,'eq','api')]}]}]}]});assert.throws(()=>plan(nested,field,['a']),/nested/);
 assert.equal(reason(q(),field,[null,'']),'');
});
