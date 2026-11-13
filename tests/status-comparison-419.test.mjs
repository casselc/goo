import {test} from 'node:test';
import assert from 'node:assert/strict';
import {validateComparison} from '../ui/view-comparison.mjs';
import {formatComparisonChange,formatCell,exactText} from '../ui/view-values.mjs';
import {fractionFacts} from '../ui/view-schema.mjs';
const classes=['error','ok','unset','unknown'];
const fields=[{key:'count',label:'Captured spans',type:'number',unit:'records'},...classes.map(key=>({key:'status.'+key,label:key,type:'number',unit:'records',nullable:true,aggregation:{family:'span-status',kind:'count',method:'exact'}})),{key:'status.error_fraction',label:'Recorded Error fraction',type:'rational',unit:'1',nullable:true,aggregation:{family:'ratio',kind:'fraction',method:'exact-rational',numerator:'status.error',denominator:'count'}}];
function side(id,count,values,fraction){return {version:'goo.dataset/1',id,revision:id,origin:'browser-held',scope:{id,generation:id,signal:'spans',selector:{signal:'spans',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0,spanStatus:true,groupBy:[{kind:'column',key:'name'}]}},shape:'summary',fields:structuredClone(fields),rows:[{key:'population',values:{count,'status.error_fraction':fraction,...Object.fromEntries(classes.map((key,i)=>['status.'+key,values[i]]))},cells:{'status.error_fraction':{state:count==='0'?'no_data':'derived'},...Object.fromEntries(classes.map(key=>['status.'+key,{state:'derived'}]))}}],quality:{extent:'unknown',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'}};}
function fixture(){return {version:'goo.comparison/1',state:'observed',baseline:side('a','100',['10','50','30','10'],'1/10'),selection:side('b','200',['10','70','100','20'],'1/20'),changes:Object.fromEntries(fields.map((field,i)=>[field.key,{state:'derived',value:['100','0','20','70','10','-1/20'][i],unit:field.unit, ...(field.aggregation?{aggregation:{...field.aggregation,...(field.aggregation.family==='ratio'?{kind:'difference'}:{})}}:{})}]))};}

test('419 recorded Error comparisons retain exact counts, partition and signed fraction differences',()=>{
 const c=fixture(),before=structuredClone(c);assert.deepEqual(validateComparison(c),{ok:true,reasons:[]});assert.deepEqual(c,before);
 const f=fields.at(-1),facts=fractionFacts(c.selection,c.selection.rows[0],f.key);assert.equal(facts.numerator.value,'10');assert.equal(facts.denominator.value,'200');assert.equal(facts.partition.find(f=>f.field==='status.unset').value,'100');
 assert.equal(formatCell(f,c.selection.rows[0].values[f.key],c.selection.rows[0].cells[f.key]),'5%');assert.equal(formatComparisonChange(f,c.changes[f.key]),'−5 pp');assert.match(exactText(f,'1/20',c.selection.rows[0].cells[f.key],facts),/10.*200/);
});
test('419 percentage-point formatting preserves signs, exact zero and very small rational differences',()=>{
 const field=fields.at(-1),delta=value=>({state:'derived',value});
 assert.equal(formatComparisonChange(field,delta('0')),'0 pp');assert.equal(formatComparisonChange(field,delta('1/2')),'50 pp');assert.equal(formatComparisonChange(field,delta('-1/2')),'−50 pp');assert.equal(formatComparisonChange(field,delta('1/3')),'≈ 33.33 pp');assert.equal(formatComparisonChange(field,delta('-1/100000000')),'> −0.01 pp');assert.equal(formatComparisonChange(field,delta('-99999999/100000000')),'< −99.99 pp');assert.equal(formatComparisonChange(field,delta('99999999/100000000')),'>99.99 pp');assert.equal(formatComparisonChange(field,{state:'unavailable',value:null}),'Unavailable');
});
test('419 zero population keeps known status counts and makes the fraction change unavailable',()=>{
 const c=fixture();c.state='no_data';c.baseline=side('a','0',['0','0','0','0'],null);for(const field of fields.slice(0,-1)){c.changes[field.key].value=c.selection.rows[0].values[field.key]};c.changes['status.error_fraction']={state:'unavailable',value:null,unit:'1',aggregation:{...fields.at(-1).aggregation,kind:'difference'},reason:'Baseline has no captured spans'};
 assert.equal(validateComparison(c).ok,true);const bad=structuredClone(c);bad.changes['status.error_fraction'].state='derived';bad.changes['status.error_fraction'].value='1/20';assert.equal(validateComparison(bad).ok,false);
});
test('419 rejects forged partitions, ratios, signed-change metadata and arithmetic',()=>{
 for(const mutate of [c=>c.baseline.rows[0].values['status.unset']='31',c=>c.selection.rows[0].values['status.error_fraction']='1/10',c=>c.changes['status.error_fraction'].value='-1/10',c=>c.changes['status.error_fraction'].aggregation.kind='fraction',c=>delete c.changes['status.error_fraction'].aggregation,c=>c.changes['status.error_fraction'].aggregation.denominator='status.ok',c=>c.changes['status.error'].unit='1']){const c=fixture();mutate(c);assert.equal(validateComparison(c).ok,false);}
});
test('419 diagnostic changes remain withheld despite individually known Error counts',()=>{
 const c=fixture();c.state='diagnostic';for(const delta of Object.values(c.changes))Object.assign(delta,{state:'unavailable',value:null,reason:'Independent reads disagree'});assert.equal(validateComparison(c).ok,true);c.changes['status.error'].state='derived';c.changes['status.error'].value='0';assert.equal(validateComparison(c).ok,false);
});
