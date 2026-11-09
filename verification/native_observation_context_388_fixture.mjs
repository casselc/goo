import fs from 'node:fs';
import assert from 'node:assert/strict';
import {plan,radiusNS} from '../ui/native-observation-context.mjs';
const cases=JSON.parse(fs.readFileSync(0,'utf8'));let plans=0;
globalThis.fetch=()=>{throw Error('Observation projection must not read telemetry')};
assert.equal(radiusNS,'60000000000');
const norm=x=>{x=structuredClone(x);if(x.question.environmentSelector&&x.question.environment==='')delete x.question.environment;return x};
for(const c of cases){const before=JSON.stringify(c);
 for(const expected of c.Plans){const actual=plan(c.Result,c.Stream,expected.question.signal,c.AtNS);assert.deepEqual(norm(actual),norm(expected));plans++}
 for(const mutate of [r=>r.question.definition.kind='counter',r=>r.question.definition.grain='consumer-group',r=>r.question.definition.environmentSelector={version:'goo/environment-selector/v1',policy:'current-only',scope:{mode:'empty'}},r=>r.gauges.series.push(structuredClone(r.gauges.series[0])),r=>r.gauges.series[0].samples.push(structuredClone(r.gauges.series[0].samples[0])),r=>r.gauges.series[0].samples[0].state='invented',r=>r.gauges.series[0].samples[0].atNS='0'+r.gauges.series[0].samples[0].atNS]){const bad=structuredClone(c.Result);mutate(bad);assert.throws(()=>plan(bad,c.Stream,'logs',c.AtNS))}
 for(const signal of ['metrics','','Spans'])assert.throws(()=>plan(c.Result,c.Stream,signal,c.AtNS));
 for(const at of ['0'+c.AtNS,String(BigInt(c.AtNS)+1n),'+1','not-reported'])assert.throws(()=>plan(c.Result,c.Stream,'logs',at));
 const foreign=structuredClone(c.Stream);foreign.resource['service.instance.id']='foreign';assert.throws(()=>plan(c.Result,foreign,'logs',c.AtNS));
 assert.equal(JSON.stringify(c),before);
}
process.stdout.write(JSON.stringify({cases:cases.length,plans}));
