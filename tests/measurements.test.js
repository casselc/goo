const {test}=require('node:test');const assert=require('node:assert/strict');const {subjectQuestion}=require('../ui/measurements.js');
test('lifecycle drilldown preserves package identity, fixed horizon and subject tenant',()=>{const r={question:{fromMs:100000,toMs:200000,definition:{environment:'prod',source:{services:['a','b']},lifecycle:{deadlineMs:10000,allowanceMs:3000,subject:'subject.id',tenant:'tenant.id',accepted:'job.accepted',terminal:'job.finished'}}}};const q=subjectQuestion(r,{subject_id:'s1',tenant_id:'t2'});assert.equal(q.fromMs,87000);assert.equal(q.toMs,200000);assert.equal(q.environment,'prod');assert.equal(q.where.all[1].value,'s1');assert.equal(q.where.all[2].value,'t2');assert.deepEqual(q.where.all[3].any.map(p=>p.value),['job.accepted','job.finished']);assert.deepEqual(q.where.all[0].any.map(p=>p.value),['a','b']);assert.equal(q.signal,'logs')});
test('drilldown rejects a window wider than the query contract',()=>{assert.throws(()=>subjectQuestion({question:{fromMs:100000,toMs:7300000,definition:{lifecycle:{deadlineMs:1,allowanceMs:0}}}},{}),/Narrow/) });

test('native interval context delegates exact bounds and preserves unknown/reset context',()=>{
 const {recordContextPlans}=require('../ui/measurements');
 const fixtures=require('./fixtures/native-record-context.json');
 const c=fixtures.find(c=>c.expected&&c.parent.definition.kind==='counter');
 const before=JSON.stringify(c),point={fromNS:c.fromNS,toNS:c.toNS,state:'unknown',reason:'reset'};
 const result=recordContextPlans({question:c.parent},{stream:c.stream},point);
 assert.equal(result.reason,undefined);assert.equal(result.plans.length,2);
 assert.deepEqual(result.plans.map(p=>p.question.signal||'logs'),['logs','spans']);
 for(const plan of result.plans){assert.equal(plan.fromNS,c.fromNS);assert.equal(plan.toNS,c.toNS);assert.equal(plan.grain,'interval-context')}
 assert.equal(JSON.stringify(c),before);
});
test('gauge context does not invent a nearby window or offer a one-nanosecond shortcut',()=>{
 const {recordContextPlans}=require('../ui/measurements');
 const result=recordContextPlans({question:{definition:{kind:'gauge'}}},{},{fromNS:'1',toNS:'1'});
 assert.deepEqual(result.plans,[]);assert.match(result.reason,/no interval/);
});
test('invalid native interval has no partially successful context links',()=>{
 const {recordContextPlans}=require('../ui/measurements');
 const result=recordContextPlans({question:{definition:{kind:'counter'}}},{stream:{}},{fromNS:'oops',toNS:'1'});
 assert.deepEqual(result.plans,[]);assert.match(result.reason,/unavailable/);
});

test('native context links preserve modified and non-primary navigation',()=>{const {ordinaryContextClick}=require('../ui/measurements');assert.equal(ordinaryContextClick({button:0}),true);for(const event of [{button:1},{button:2},{button:0,ctrlKey:true},{button:0,metaKey:true},{button:0,shiftKey:true},{button:0,altKey:true}])assert.equal(ordinaryContextClick(event),false)});
