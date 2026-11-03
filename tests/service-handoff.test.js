const {test}=require('node:test'),assert=require('node:assert/strict');
const H=require('../ui/service-handoff.mjs'),E=require('../ui/record-environment.mjs');
const query=()=>({signal:'logs',environmentSelector:E.current('prod'),fromMs:1000,toMs:61000,limit:100,offset:0,where:{field:{kind:'column',key:'body'},op:'contains',value:'timeout'}});
const row=()=>({ServiceName:'edge-api',timestamp_ns:'1000000000',ResourceAttributes:{'deployment.environment.name':'prod','service.namespace':'shop'},Attributes:{'service.namespace':'wrong-log-namespace'}});
test('observed service context preserves original intent and explicitly encodes an equivalent destination',()=>{
 const q=query(),r=row(),before=structuredClone({q,r}),intent=H.create(q,r),scope=H.scope(intent);
 assert.deepEqual(intent.origin,q);assert.deepEqual(scope,{environment:'prod',namespace:{mode:'value',value:'shop'},fromMs:1000,toMs:61000});assert.equal(E.sameQuestionEnvironment(q,scope),true);assert.equal(scope.where,undefined);assert.equal(scope.signal,undefined);assert.equal(scope.service,undefined);assert.equal(intent.service,'edge-api');assert(Object.isFrozen(intent.origin.where.field));
 assert.deepEqual({q,r},before);q.where.value='changed';r.ResourceAttributes['service.namespace']='changed';scope.namespace.value='changed';assert.equal(intent.origin.where.value,'timeout');assert.equal(intent.namespace.value,'shop');assert.equal(H.scope(intent).namespace.value,'shop');
 const legacy={...query(),environment:'prod'};delete legacy.environmentSelector;assert.deepEqual(H.create(legacy,row()).origin,legacy);
});
test('namespace presence comes only from a known resource map and preserves empty versus missing',()=>{
 const r=row();delete r.ResourceAttributes['service.namespace'];assert.deepEqual(H.create(query(),r).namespace,{mode:'missing'});r.ResourceAttributes['service.namespace']='';assert.deepEqual(H.create(query(),r).namespace,{mode:'value',value:''});
 for(const value of [undefined,null,[],42]){const bad=row();if(value===undefined)delete bad.ResourceAttributes;else bad.ResourceAttributes=value;assert.throws(()=>H.create(query(),bad),/namespace reporting is unknown/)}
 for(const value of [null,12,'x\0y','x'.repeat(129)]){const bad=row();bad.ResourceAttributes['service.namespace']=value;assert.throws(()=>H.create(query(),bad))}
});
test('unsupported environment policies never weaken to a same-named legacy selector',()=>{
 for(const selector of [E.effective({mode:'named',value:'prod'}),E.effective({mode:'empty'}),E.effective({mode:'unreported'}),{...E.current('prod'),version:'future'},{...E.current('prod'),policy:'future'}])assert.notEqual(H.reason({...query(),environmentSelector:selector},row()),'');
 assert.throws(()=>H.create({...query(),environment:'prod'},row()));const q=query();delete q.environmentSelector;q.environment='';assert.throws(()=>H.create(q,row()));const r=row();r.ResourceAttributes['deployment.environment.name']='other';assert.throws(()=>H.create(query(),r),/environment/);
});
test('service identities and completed observation windows are bounded and checked, with no clock recomputation',()=>{
 for(const name of ['',null,'bad\0name','x'.repeat(129)])assert.throws(()=>H.create(query(),{...row(),ServiceName:name}));
 for(const timestamp of ['999999999','61000000000','01','x','1'.repeat(100)])assert.throws(()=>H.create(query(),{...row(),timestamp_ns:timestamp}));
 assert.equal(H.create(query(),{...row(),timestamp_ns:'60999999999'}).toMs,61000);const q={...query(),fromMs:4102444799000,toMs:4102444800000};assert.throws(()=>H.create(q,{...row(),timestamp_ns:'4102444799000000000'}),/bounds/);
 const intent=H.create(query(),row());for(const patch of [{fromMs:1001},{toMs:61001},{namespace:{mode:'missing',value:''}},{environmentSelector:E.effective({mode:'named',value:'prod'})},{extra:true},{version:'future'}])assert.throws(()=>H.validate({...intent,...patch}));
 assert.equal(H.reason(query(),row()),'');
});
