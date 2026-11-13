import fs from 'node:fs';
import assert from 'node:assert/strict';
import {validateComparison} from '../../ui/view-comparison.mjs';
import {measureValue} from '../../ui/view-model.mjs';
import * as Captured from '../../ui/captured-count-comparison.mjs';
globalThis.fetch=()=>{throw Error('A supplied comparison performed IO')};
const cases=JSON.parse(fs.readFileSync(0,'utf8')),before=structuredClone(cases);
for(const c of cases){
 const valid=validateComparison(c.comparison);assert(valid.ok,c.name+': '+valid.reasons.join('; '));
 for(const [i,side]of [c.comparison.baseline,c.comparison.selection].entries()){
  const value=measureValue(side,side.rows[0],'record.rate');assert.equal(value.exact,[c.baselineRate,c.selectionRate][i],c.name+' exact operand');
 }
 const a=c.comparison.baseline.scope.selector,b=c.comparison.selection.scope.selector;
 const declaration={version:'goo.view/1',scope:{kind:'record-query',signal:b.signal},kind:{name:'captured-count-comparison',version:1},baselineCapture:'a'.repeat(64),display:{value:'record.rate',view:'summary'}};
 Captured.validate(declaration,b);assert.throws(()=>Captured.validate({...declaration,display:{...declaration.display,view:'groups'}},b));
 const fresh={...declaration,kind:{name:'query-count-comparison',version:1},baselineQuery:structuredClone(b)};delete fresh.baselineCapture;assert.throws(()=>Captured.validate(fresh,b));
 const rejects=[
  x=>x.baseline.rows[0].cells['record.rate'].rate.elapsedMs='1',
  x=>x.selection.rows[0].cells['record.rate'].rate.count=x.selection.rows[0].values.count==='0'?'1':'0',
  x=>x.selection.rows[0].values['record.rate']='1/0',
  x=>x.baseline.capturedAt='not a source clock',
  x=>x.baseline.scope.selector.groupBy.push({kind:'attribute',key:'region'}),
  x=>x.baseline.fields.find(f=>f.key==='record.rate').aggregation.method='average-of-rates',
  x=>delete x.baseline.fields.find(f=>f.key==='record.rate').aggregation,
  x=>x.baseline.fields.find(f=>f.key==='record.rate').unit='records',
  x=>x.changes['record.rate'].value='999999999999999999999'
 ];
 for(const mutate of rejects){const bad=structuredClone(c.comparison);mutate(bad);assert.equal(validateComparison(bad).ok,false,c.name+' admitted corrupt rate metadata');}
 const blocked=structuredClone(c.comparison);blocked.state='incompatible';blocked.reason='independent unsupported profile';for(const change of Object.values(blocked.changes)){change.state='unavailable';change.value=null;change.reason='incompatible-fields';}
 assert.equal(validateComparison(blocked).ok,true,c.name+' conservative unavailable comparison');blocked.changes['record.rate'].state='derived';blocked.changes['record.rate'].value='0';assert.equal(validateComparison(blocked).ok,false,c.name+' incompatible change invented');
}
assert.deepEqual(cases,before,'Validation changed supplied observations');
process.stdout.write(JSON.stringify(cases.map(c=>c.name)));
