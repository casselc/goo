import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Protocol from '../ui/question-protocol.mjs';
import {admitSelectionPlan} from '../ui/numeric-distribution.mjs';
import {identity} from '../ui/question-session.mjs';
const plan=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(plan);
globalThis.fetch=()=>{throw Error('Typed continuation admission must not read telemetry')};
admitSelectionPlan(plan,plan.parent);
let refusals=0;
for(const mutate of [p=>p.selected.environmentSelector.policy=p.parent.environmentSelector.policy==='effective'?'current-only':'effective',p=>p.selected.fromMs++,p=>p.selected.where.all[0]={field:{kind:'column',key:'service'},op:'eq',value:'intruder'},p=>p.selected.numeric.unit='other',p=>p.rest.where.all[1].any[0].numeric.field.key='other',p=>p.comparison.selection.numeric=p.parent.numeric,p=>p.numericComparison.baseline.numericHistogram=p.parent.numericHistogram]){const bad=structuredClone(plan);mutate(bad);assert.throws(()=>admitSelectionPlan(bad,plan.parent));refusals++}
// Keep all advertised projections coherent while changing the semantic child.
// Admission must reject the altered cohort itself, not merely stale pair copies.
const sync=p=>{for(const name of ['comparison','numericComparison','distributionComparison'])if(p[name])for(const [side,child]of [['selection','selected'],['baseline','rest']])p[name][side].where=structuredClone(p[child].where)};
const derived=(p,child)=>p[child].where.all[1];
for(const mutate of [
 p=>{const g=derived(p,'selected'),parts=g.all??[g],leaf=parts.find(v=>v.op==='gte');if(leaf)leaf.value=String(BigInt(leaf.value)+1n);else p.selected.where.all[1]={all:[...parts,{numeric:structuredClone(p.parent.numeric),op:'gte',value:'0'}]}},
 p=>{const g=derived(p,'selected'),parts=g.all??[g],leaf=parts.find(v=>v.op==='lt');if(leaf)leaf.value=String(BigInt(leaf.value)-1n);else p.selected.where.all[1]={all:[...parts,{numeric:structuredClone(p.parent.numeric),op:'lt',value:'0'}]}},
 p=>{const parts=derived(p,'rest').any,leaf=parts.find(v=>v.op==='lt'||v.op==='gte');if(leaf)leaf.value=String(BigInt(leaf.value)+(leaf.op==='lt'?1n:-1n));else parts.push({numeric:structuredClone(p.parent.numeric),op:'lt',value:'0'})},
 p=>{const parts=derived(p,'rest').any;parts[0]=structuredClone(parts.find(v=>v.op==='invalid'))},
]){const bad=structuredClone(plan);mutate(bad);sync(bad);assert.throws(()=>admitSelectionPlan(bad,plan.parent));refusals++}
assert.equal(refusals,11);

const children=[plan.selected,plan.rest,plan.comparison.selection,plan.comparison.baseline,plan.numericComparison.selection,plan.numericComparison.baseline];
for(const question of children){
 Protocol.validateLink({version:1,question,view:'Records'});
 const step=Protocol.readStepLink(Protocol.questionLink(question));
 assert.equal(identity(step.record),identity(question));
 assert.deepEqual(step.record,question);
 assert.deepEqual(step.scope.environmentSelector,question.environmentSelector);
}
assert.equal(JSON.stringify(plan),before);
process.stdout.write(JSON.stringify({selected:plan.selected,rest:plan.rest,checks:children.length}));
