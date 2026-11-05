import {test} from 'node:test';
import assert from 'node:assert/strict';
import {countedState,projection,valueKey} from '../ui/categorical-pair.mjs';
import {model,loadedMatchesText,CategoricalFacet} from '../ui/categorical-facet.mjs';
import {create} from '../ui/facet-declaration.mjs';
const field={kind:'attribute',key:'http.request.method'},receipt={readAt:'2026-11-05T00:00:01Z',basis:'a'.repeat(64)};
const ready=items=>({state:'ready',field,items,total:String(items.reduce((sum,item)=>sum+BigInt(item.count),0n)),readAt:receipt.readAt,basis:'partition',reads:{count:receipt,groups:{...receipt,basis:'b'.repeat(64)}},label:'logs · prod · [2026-11-05T00:00:00Z, 2026-11-05T00:01:00Z)'});
const values=(...pairs)=>ready(pairs.map(([value,count])=>({value,count})));
const declaration=create({signal:'logs',environment:'prod',fromMs:0,toMs:60000,offset:0,limit:100,groupBy:[]},field);
test('union exposes baseline-only categories while null, empty and literal sentinel strings remain distinct',()=>{
 const selected=values(['GET','75'],['POST','27'],['','2'],[null,'3']),baseline=values(['GET','5'],['POST','4'],['DELETE','2'],['','1'],[null,'1'],['Missing key','1']);
 const original=structuredClone({selected,baseline}),p=projection(selected,baseline,field),rows=new Map(p.rows.map(row=>[row.value,row]));
 assert.equal(p.complete,true);assert.equal(p.max,'75');assert.equal(p.rows.length,6);assert.deepEqual(rows.get('DELETE').selected,{state:'absent',count:'0'});assert.deepEqual(rows.get('DELETE').baseline,{state:'observed',count:'2'});assert.equal(rows.get(null).selected.count,'3');assert.equal(rows.get('').selected.count,'2');assert.equal(rows.get('Missing key').selected.count,'0');assert.notEqual(valueKey(null),valueKey(''));assert.notEqual(valueKey(null),valueKey('null'));assert.deepEqual({selected,baseline},original);
});
test('missing categories in unavailable operands remain unknown while the successful witness is retained',()=>{
 const baseline=values(['DELETE','2'],[null,'1']);for(const state of ['unloaded','loading','error']){
  const p=projection({state},baseline,field);assert.equal(p.complete,false);assert.equal(p.max,'2');assert(p.rows.every(row=>row.selected.count===null&&row.selected.state==='unknown'));assert.equal(p.rows.find(row=>row.value==='DELETE').baseline.count,'2');
  const swapped=projection(baseline,{state},field);assert(swapped.rows.every(row=>row.baseline.count===null&&row.baseline.state==='unknown'));assert.equal(swapped.rows.find(row=>row.value==='DELETE').selected.count,'2');
 }
});
test('a validated empty grouping admits relative absence, unlike a failed read',()=>{
 const p=projection(ready([]),values(['DELETE','2']),field);assert.deepEqual(p.rows[0].selected,{state:'absent',count:'0'});assert.equal(p.complete,true);assert.deepEqual(projection(ready([]),ready([]),field),{rows:[],max:'0',complete:true});assert.deepEqual(projection({state:'error'},{state:'loading'},field),{rows:[],max:'0',complete:false});
});
test('uint64 counts stay exact and ordering uses the whole reported union',()=>{
 const p=projection(values(['a','18446744073709551614']),values(['b','18446744073709551615']),field);assert.equal(p.max,'18446744073709551615');assert.deepEqual(p.rows.map(row=>row.value),['b','a']);assert.equal(p.rows[1].selected.count,'18446744073709551614');
 const ties=projection(values(['雪','1'],['','1']),values(['z','1'],[null,'1']),field);assert.deepEqual(ties.rows.map(row=>row.value),[null,'','z','雪']);
});
test('individually bounded partitions admit a 2000-value union without truncating it',()=>{
 const side=prefix=>ready(Array.from({length:1000},(_,i)=>({value:prefix+i,count:'1'}))),p=projection(side('a'),side('b'),field);assert.equal(p.rows.length,2000);assert.equal(p.max,'1');assert.equal(loadedMatchesText({visible:1,total:2000},2000),'1 of 2000 loaded values');assert.throws(()=>loadedMatchesText({visible:1,total:2000}));assert.throws(()=>countedState(ready(Array.from({length:1001},(_,i)=>({value:String(i),count:'1'}))),field));
});
test('both operands must independently validate their actual grouping and receipts',()=>{
 const selected=values(['GET','1']);for(const mutate of [v=>v.total='2',v=>v.field={...field,key:'other'},v=>v.items.push({...v.items[0]}),v=>v.items[0].count='01',v=>v.reads.groups.basis='short',v=>v.items[0].value='\ud800']){
  const bad=structuredClone(selected);mutate(bad);assert.throws(()=>projection(selected,bad,field));assert.throws(()=>projection(bad,selected,field));
 }
});
test('paired model accepts independent states, owns both copies and withdraws invalid replacements',()=>{
 const data={declaration,values:values(['GET','2']),baseline:values(['DELETE','1']),baselineLabel:'Previous window'},element=new CategoricalFacet();element.data=data;data.baseline.items[0].count='99';assert.equal(element.data.baseline.items[0].count,'1');const copy=element.data;copy.values.items[0].count='77';assert.equal(element.data.values.items[0].count,'2');
 assert.doesNotThrow(()=>model({declaration,values:{state:'error'},baseline:values(['DELETE','1'])}));assert.throws(()=>model({declaration,values:{state:'unloaded'},baselineLabel:'Previous window'}));assert.throws(()=>model({...data,baselineLabel:'\0'}));element.data=data;assert.equal(element.data,null);
});
