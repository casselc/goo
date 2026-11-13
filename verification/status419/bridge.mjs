import fs from 'node:fs';
import assert from 'node:assert/strict';
import {validateComparison} from '../../ui/view-comparison.mjs';
import {fractionFacts} from '../../ui/view-schema.mjs';
globalThis.fetch=()=>{throw Error('Supplied status comparison performed source IO')};
const cases=JSON.parse(fs.readFileSync(0,'utf8')),before=structuredClone(cases);
for(const c of cases){
 const admitted=validateComparison(c.comparison);assert(admitted.ok,c.name+': '+admitted.reasons.join('; '));
 for(const [i,side]of [c.comparison.baseline,c.comparison.selection].entries()){
  const row=side.rows[0];for(const [key,value]of Object.entries(c.expected[i]))assert.equal(row.values[key],value,c.name+' '+key);
  const facts=fractionFacts(side,row,'status.error_fraction');assert(facts,c.name+' missing denominator support');
  assert.equal(facts.numerator.value,row.values['status.error']);assert.equal(facts.denominator.value,row.values.count);
 }
 for(const [key,want]of Object.entries(c.changes))assert.equal(c.comparison.changes[key].value,want,c.name+' change '+key);
 const mutants=[
  x=>x.baseline.rows[0].values['status.unset']='999999999999999999',
  x=>x.baseline.rows[0].values['status.error']='1/2',
  x=>x.baseline.rows[0].values['status.error_fraction']='1/0',
  x=>x.baseline.fields.find(f=>f.key==='status.error_fraction').aggregation.denominator='status.ok',
  x=>delete x.selection.fields.find(f=>f.key==='status.error_fraction').aggregation,
  x=>x.changes['status.error_fraction'].aggregation.kind='fraction',
  x=>x.changes['status.error_fraction'].value='99',
  x=>x.changes['status.error'].value='999999999999999999',
  x=>x.selection.quality.extent='complete'
 ];
 for(const mutate of mutants){const bad=structuredClone(c.comparison);mutate(bad);assert.equal(validateComparison(bad).ok,false,c.name+' accepted changed status support');}
}
assert.deepEqual(cases,before,'admission changed supplied facts');process.stdout.write(JSON.stringify(cases.map(c=>c.name)));
