import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Preservation from '../ui/preservation.mjs';
import * as CapturedCount from '../ui/captured-count-comparison.mjs';
import {comparisonScopes} from '../ui/capture-question-preview.mjs';
const field={kind:'column',key:'service'},base={signal:'logs',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},fromMs:1000,toMs:61000,limit:10,offset:0,groupBy:[field]},selection={...base,where:{field:{kind:'column',key:'body'},op:'contains',value:'east'}},baseline={...base,where:{field:{kind:'column',key:'body'},op:'contains',value:'west'}};
const declaration=()=>CapturedCount.createQuery(selection,baseline);
test('428 pair configuration carries ordered full queries without presentation, source or held evidence',()=>{
 const selected=structuredClone(selection);selected.offset=20;const question=Preservation.comparisonQuestion([selected],declaration());
 assert.deepEqual(question,{kind:'comparison',comparison:{baseline,selection}});assert.notEqual(question.comparison.baseline,baseline);question.comparison.baseline.where.value='changed';assert.equal(baseline.where.value,'west');
});
test('428 paired preview exposes exact fixed questions, policies and predicates in semantic supplied facts',()=>{
 const scopes=comparisonScopes(Preservation.comparisonQuestion([selection],declaration()));assert.deepEqual(scopes.map(s=>s.label),['Baseline','Selected']);
 for(const side of scopes){assert.equal(side.data.mode,'fixed');assert.equal(side.data.state,'unknown');assert.equal(side.data.facts[1].value,'[1000000000, 61000000000)');assert.equal(side.data.facts[1].format,'unix-ns-range');assert.equal(side.data.details[0].value,'effective');assert.equal(side.data.details[1].value,'column:service');}
 assert.match(scopes[0].selector,/west/);assert.match(scopes[1].selector,/east/);
});
test('428 pair admission rejects foreign profiles and altered environment/grouping instead of selecting another metric',()=>{
 for(const mutate of [q=>q.spanStatus=true,q=>q.numeric={field:{kind:'attribute',key:'n'},unit:'count'},q=>q.logSeverity=true,q=>q.summary='span-duration',q=>q.environmentSelector.scope.value='other',q=>q.groupBy=[{kind:'column',key:'span_name'}]]){
  const q=structuredClone(selection);mutate(q);assert.throws(()=>Preservation.comparisonQuestion([q],declaration()));
 }
 const d=declaration();d.display.value='duration.mean';assert.throws(()=>Preservation.comparisonQuestion([selection],d));assert.throws(()=>Preservation.comparisonQuestion([selection,baseline],declaration()));
});

test('428 long valid selectors retain complete disclosed text without overflowing shared summary fact budgets',async()=>{
 const {model}=await import('../ui/evidence-summary.mjs');const clauses=Array.from({length:12},(_,i)=>({field:{kind:'attribute',key:'long'+i},op:'eq',value:'x'.repeat(900)+i}));
 const q={...selection,where:{all:clauses}},pair=Preservation.comparisonQuestion([q],CapturedCount.createQuery(q,baseline)),scope=comparisonScopes(pair)[1];assert(scope.selector.length>8192);assert(scope.selector.endsWith('11")'));assert.equal(model(scope.data).valid,true);assert.equal(scope.data.facts.some(f=>f.value.includes('x'.repeat(900))),false);
});
