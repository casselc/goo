const {test}=require('node:test'),assert=require('node:assert/strict');
const {traceView}=require('../ui/record-results.mjs'),Composition=require('../ui/view-composition.mjs');
const trace='0123456789abcdef0123456789abcdef';
test('trace view stages exact trace waterfall intent without source data or inherited inspection filters',()=>{
 const origin={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,where:{field:{kind:'column',key:'service'},op:'eq',value:'one'},groupBy:[{kind:'column',key:'service'},{kind:'column',key:'name'}],limit:20,offset:100,summary:'span-duration',spanStatus:true,bucketMs:1000};
 const before=structuredClone(origin),view=traceView(origin,trace);
 assert.deepEqual(view.queries,[{signal:'spans',environment:'prod',fromMs:1000,toMs:61000,where:{all:[{field:{kind:'column',key:'trace_id'},op:'eq',value:trace}]},limit:100,offset:0}]);
 assert.equal(Composition.recordOnly(view.composition),true);assert.equal(Composition.records(view.composition).kind.name,'record-span-waterfall');assert.deepEqual(Object.keys(view).sort(),['composition','queries','title','version']);assert.deepEqual(origin,before);
 const another=traceView(origin,trace);view.composition.children[0].records.display.columns.pop();assert.notDeepEqual(view,another);
});
test('trace view preserves typed environment policy and rejects invalid trace or bounds',()=>{
 const origin={signal:'spans',environment:'prod',fromMs:1000,toMs:61000};
 for(const id of ['0'.repeat(32),'ABCDEF0123456789abcdef0123456789','trace'])assert.throws(()=>traceView(origin,id));
 const typed={signal:'spans',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},fromMs:1000,toMs:61000};assert.deepEqual(traceView(typed,trace).queries[0].environmentSelector,typed.environmentSelector);assert.equal(traceView(typed,trace).queries[0].environment,undefined);
 assert.throws(()=>traceView({...origin,toMs:1000},trace));assert.throws(()=>traceView({...origin,environment:''},trace));
});
test('trace views preserve every supported typed environment representation without aliasing',()=>{
 for(const fixture of require('../testdata/protocol/environment-selectors.json').cases)for(const legacy of [{},{environment:''}]){
  const origin={signal:'logs',...legacy,environmentSelector:structuredClone(fixture.selector),fromMs:1000,toMs:61000},before=structuredClone(origin),q=traceView(origin,trace).queries[0];
  assert.deepEqual(q.environmentSelector,origin.environmentSelector,fixture.name);assert.equal(Object.hasOwn(q,'environment'),Object.hasOwn(origin,'environment'));assert.notEqual(q.environmentSelector,origin.environmentSelector);q.environmentSelector.scope.mode='broken';assert.deepEqual(origin,before);
 }
});
