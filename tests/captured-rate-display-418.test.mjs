import {createSourceAdapter} from '../ui/snapshots.mjs';
import {test} from 'node:test';
import assert from 'node:assert/strict';
import {validateComparison,comparisonPresentation} from '../ui/view-comparison.mjs';
import {validateSnapshot} from '../ui/view-schema.mjs';
import * as Summary from '../ui/summary-declaration.mjs';
import {formatCell,exactText,defaultSummaryFields} from '../ui/view-values.mjs';
const rate={key:'record.rate',label:'Captured log records/s',type:'rational',unit:'records/s',nullable:true,aggregation:{family:'rate',kind:'records-per-second',method:'exact-rational'}};
const count={key:'count',label:'Captured records',type:'number',unit:'records'};
const q=(fromMs,toMs)=>({signal:'logs',environment:'prod',fromMs,toMs,limit:100,offset:0,groupBy:[{kind:'column',key:'service'}]});
function side(id,question,total,value){return {version:'goo.dataset/1',id,revision:id+'-revision',origin:'browser-held',capturedAt:'2026-11-13T00:00:00Z',scope:{id:id+'-scope',generation:id+'-generation',signal:question.signal,selector:question},shape:'summary',fields:[structuredClone(count),structuredClone(rate)],rows:[{key:'population',values:{count:total,'record.rate':value},cells:{'record.rate':{state:'derived',rate:{count:total,elapsedMs:String(question.toMs-question.fromMs)}}}}],quality:{extent:'unknown',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'}};}
function comparison(){return {version:'goo.comparison/1',state:'observed',baseline:side('a',q(1000,61000),'120','2'),selection:side('b',q(100000,220000),'240','2'),changes:{count:{state:'derived',unit:'records',value:'120'},'record.rate':{state:'derived',unit:'records/s',aggregation:structuredClone(rate.aggregation),value:'0'}}};}
function unavailable(c){for(const change of Object.values(c.changes))Object.assign(change,{state:'unavailable',value:null,reason:'Independent source reads disagree'});}

test('418 equal captured rates remain equal across different counts and independently exposed windows',()=>{
 const c=comparison(),before=structuredClone(c);assert.equal(validateComparison(c).ok,true);assert.equal(c.changes.count.value,'120');assert.equal(c.changes['record.rate'].value,'0');assert.deepEqual(c,before);
 assert.equal(formatCell(rate,c.baseline.rows[0].values['record.rate'],c.baseline.rows[0].cells['record.rate']),'2 records/s');assert.match(exactText(rate,'2',c.selection.rows[0].cells['record.rate']),/240 captured records over 120000 ms/);
 assert.deepEqual(defaultSummaryFields(c.baseline.fields),[count],'rate remains an explicit presentation choice');
});
test('418 canonical rate differences retain large rational values rather than float or relative-count changes',()=>{
 const c=comparison();c.baseline.scope.selector.toMs=4000;c.baseline.rows[0].values.count='1';c.baseline.rows[0].values['record.rate']='1/3';c.baseline.rows[0].cells['record.rate'].rate={count:'1',elapsedMs:'3000'};c.selection.scope.selector.toMs=103000;c.selection.rows[0].values.count='2';c.selection.rows[0].values['record.rate']='2/3';c.selection.rows[0].cells['record.rate'].rate={count:'2',elapsedMs:'3000'};c.changes.count.value='1';c.changes['record.rate'].value='1/3';assert.equal(validateComparison(c).ok,true);c.changes['record.rate'].value='0.3333333333333333';assert.equal(validateComparison(c).ok,false);
});
test('418 known zero rate is derived while diagnostic differences remain unavailable',()=>{
 const c=comparison();c.state='no_data';c.baseline.rows[0].values.count='0';c.baseline.rows[0].values['record.rate']='0';c.baseline.rows[0].cells['record.rate'].rate.count='0';c.changes.count.value='240';c.changes['record.rate'].value='2';assert.equal(validateComparison(c).ok,true);c.state='diagnostic';unavailable(c);assert.equal(validateComparison(c).ok,true);c.changes['record.rate'].state='derived';c.changes['record.rate'].value='2';assert.equal(validateComparison(c).ok,false);
});
test('418 rate admission refuses fabricated exposure, counts, grouping, aggregation and signed changes',()=>{
 for(const mutate of [c=>c.baseline.rows[0].cells['record.rate'].rate.elapsedMs='120000',c=>c.baseline.rows[0].cells['record.rate'].rate.count='121',c=>delete c.baseline.rows[0].cells['record.rate'].rate,c=>c.baseline.scope.selector.groupBy.push({kind:'attribute',key:'region'}),c=>c.baseline.fields[1].aggregation.method='approximate',c=>c.changes['record.rate'].unit='requests/s',c=>c.changes['record.rate'].value='120']){const c=comparison();mutate(c);assert.equal(validateComparison(c).ok,false);}
});
test('418 only ordinary one-group summary rates extend the ungrouped profile; timeline grouping stays refused',()=>{
 const question=q(1000,61000);assert.ok(Summary.availableFields(question).includes('record.rate'));
 for(const patch of [{groupBy:[...question.groupBy,{kind:'attribute',key:'region'}]},{numeric:{field:{kind:'attribute',key:'size'},type:'uint64',unit:'items',parse:'canonical-decimal'}},{groupBy:[{kind:'column',key:'body_pattern_v1'}]}])assert.equal(Summary.rateEligible({...question,...patch}),false);
 assert.equal(Summary.rateEligible({...question,signal:'spans',summary:'span-duration'}),true);assert.equal(Summary.rateEligible({...question,signal:'spans',spanStatus:true}),false);
 const d=comparison().baseline;d.shape='time-series';d.interval={fromMs:'1000',toMs:'61000',bucketMs:'60000'};d.scope.selector.groupBy=[];d.scope.selector.bucketMs=60000;d.rows[0].key='1000';d.rows[0].state='observed';d.rows[0].values.fromMs='1000';d.rows[0].values.toMs='61000';d.fields.push({key:'fromMs',label:'From',type:'timestamp',unit:'ms'},{key:'toMs',label:'To',type:'timestamp',unit:'ms'});assert.equal(validateSnapshot(d).ok,true);d.scope.selector.groupBy=question.groupBy;assert.equal(validateSnapshot(d).ok,false);
});

test('418 supplied rate source clocks are valid independent RFC3339 values, or explicitly absent',()=>{
 const c=comparison();c.baseline.capturedAt='2026-11-13T01:00:00.123456789+01:00';c.selection.capturedAt='2026-11-13T00:02:00.123456789123Z';assert.equal(validateComparison(c).ok,true);
 for(const invalid of ['not a source clock','2026-11-13','2026-11-13T00:00:00','2026-99-13T00:00:00Z','2026-02-30T00:00:00Z','2026-11-13T24:00:00Z']){const bad=structuredClone(c);bad.baseline.capturedAt=invalid;assert.equal(validateComparison(bad).ok,false);}
 delete c.baseline.capturedAt;delete c.selection.capturedAt;assert.equal(validateComparison(c).ok,true,'missing read clocks stay unknown rather than invented');
});

test('418 source adapters admit the same grouped summary profile without admitting grouped interval rates',()=>{
 const source=createSourceAdapter({id:'rate418-source'}),raw=comparison().selection;raw.origin='source-read';const before=structuredClone(raw);
 assert.doesNotThrow(()=>source.adoptSnapshot(raw,{question:raw.scope.selector}));assert.deepEqual(raw,before);
 const invalid=structuredClone(raw);invalid.scope.selector.groupBy.push({kind:'attribute',key:'other'});assert.throws(()=>source.adoptSnapshot(invalid,{question:invalid.scope.selector}));
});

test('418 one zero operand keeps rate comparison chrome observed and scopes the empty-record notice',()=>{
 const c=comparison();c.state='no_data';c.baseline.rows[0].values.count='0';c.baseline.rows[0].values['record.rate']='0';c.baseline.rows[0].cells['record.rate'].rate.count='0';c.changes.count.value='240';c.changes['record.rate'].value='2';const before=structuredClone(c);
 assert.equal(validateComparison(c).ok,true);assert.deepEqual(comparisonPresentation(c,['Frozen baseline','Selected window']),{state:'observed',notice:'Frozen baseline has no records'});assert.deepEqual(c,before);assert.equal(c.selection.rows[0].values['record.rate'],'2');
 c.selection.rows[0].values.count='0';c.selection.rows[0].values['record.rate']='0';c.selection.rows[0].cells['record.rate'].rate.count='0';c.changes.count.value='0';c.changes['record.rate'].value='0';assert.equal(validateComparison(c).ok,true);assert.deepEqual(comparisonPresentation(c),{state:'no_data',notice:'Neither population has captured records'});
 c.state='diagnostic';unavailable(c);assert.equal(validateComparison(c).ok,true);assert.deepEqual(comparisonPresentation(c),{state:'unknown',notice:'Diagnostic comparison'});
});
