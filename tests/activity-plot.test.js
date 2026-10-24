const {test}=require('node:test');
const assert=require('node:assert/strict');
const {plotWindow}=require('../ui/stream-entry');

test('Activity row plots retain the original applied window across later same-selector polls',()=>{
 const selector={signal:'spans',environment:'production',where:{field:{kind:'resource',key:'team'},op:'eq',value:'search'}};
 const origin={selector:structuredClone(selector),readWindow:{fromMs:1000,toMs:61000},epoch:4};
 const context={selector,readWindow:{fromMs:11000,toMs:71000},draft:{environment:'other'}};
 const before=JSON.stringify({origin,context});
 const q=plotWindow(context,origin,4);
 assert.deepEqual(q,{...selector,fromMs:1000,toMs:61000,limit:100,offset:0});
 assert.equal(JSON.stringify({origin,context}),before);
 q.where.value='mutated';
 assert.equal(origin.selector.where.value,'search');
 assert.equal(context.selector.where.value,'search');
});

test('Activity discovery plots use only the latest completed applied window',()=>{
 const context={selector:{environment:'production',where:{field:{kind:'attribute',key:'team'},op:'eq',value:'search'}},readWindow:{fromMs:11000,toMs:71000}};
 const q=plotWindow(context,null,8);
 assert.equal(q.signal,'logs');
 assert.equal(q.fromMs,11000);
 assert.equal(q.toMs,71000);
 assert.deepEqual(q.where,context.selector.where);
 assert.throws(()=>plotWindow({...context,readWindow:null},null,8));
});

test('retained row callbacks refuse changed epochs and applied selectors',()=>{
 const selector={signal:'logs',environment:'production',where:{field:{kind:'attribute',key:'team'},op:'eq',value:'search'}};
 const origin={selector,readWindow:{fromMs:1000,toMs:61000},epoch:4};
 const context={selector:structuredClone(selector),readWindow:origin.readWindow};
 assert.throws(()=>plotWindow(context,origin,5));
 for(const changed of [{...selector,environment:'staging'},{...selector,signal:'spans'},{...selector,where:{...selector.where,value:'payments'}}])assert.throws(()=>plotWindow({...context,selector:changed},origin,4));
 assert.throws(()=>plotWindow(context,{...origin,readWindow:{fromMs:1000,toMs:61001}},4));
});
