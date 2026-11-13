import {test} from 'node:test';
import assert from 'node:assert/strict';
import {create,current} from '../ui/activity-facet-entry.mjs';
import {fieldReason,query} from '../ui/filter-values.mjs';
const field={kind:'attribute',key:'cohort'},context={active:true,clean:true,phase:'paused',selector:{signal:'logs',environment:'',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},where:{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''}},readWindow:{fromMs:60000,toMs:120000}};
test('421 facet entry preserves exact latest scope while distinguishing the inspected window',()=>{
 const original=structuredClone(context),intent=create(context,field,{fromMs:0,toMs:60000});
 assert.deepEqual(intent.question,query({...context.selector,...context.readWindow},field));assert.match(intent.label,/Latest completed Activity window/);assert.match(intent.label,/inspected record window differs/);assert.equal(current(intent,context),true);
 for(const patch of [{active:false},{clean:false},{phase:'loading'},{phase:'stopped'},{readWindow:{fromMs:60001,toMs:120001}},{selector:{...context.selector,where:{field,op:'missing'}}}])assert.equal(current(intent,{...context,...patch}),false);
 assert.deepEqual(context,original);assert.notDeepEqual(create(context,{kind:'resource',key:'cohort'},intent.observationWindow).question,intent.question);
});
test('421 counted fields require string source types and reject unsupported numeric columns before projection',()=>{
 for(const key of ['duration_ns','timestamp_ns','severity_number']){assert(fieldReason({kind:'column',key},'logs'));assert.throws(()=>query({...context.selector,...context.readWindow},{kind:'column',key}));}
 assert.equal(fieldReason(field,'logs'),'');assert.equal(fieldReason({kind:'resource',key:'cohort'},'logs'),'');assert(fieldReason({kind:'column',key:'not_a_field'},'logs'));
 assert.throws(()=>create({...context,clean:false},field,{fromMs:0,toMs:60000}));assert.throws(()=>create(context,field,{fromMs:10,toMs:0}));
});
