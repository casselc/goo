const {test}=require('node:test'),assert=require('node:assert/strict'),{question}=require('../ui/population-selection');
const base={signal:'spans',environment:'prod',fromMs:1000,toMs:6000,limit:100,offset:100,where:{any:[{field:{kind:'attribute',key:'route'},op:'eq',value:''},{field:{kind:'attribute',key:'route'},op:'missing'}]},groupBy:[{kind:'attribute',key:'route'}],summary:'span-duration'};
test('population transform narrows an OR population and preserves exact half-open duration/time bounds',()=>{const before=structuredClone(base),q=question(base,{fromMs:1500,toMs:5000,minMS:'0.000001',maxMS:'1',values:[{field:{kind:'attribute',key:'route'},value:null}]});assert.deepEqual(base,before);assert.equal(q.offset,0);assert.deepEqual(q.where.all[0],base.where);assert.deepEqual(q.where.all.slice(1).map(p=>[p.op,p.value]),[['gte','1'],['lte','999999'],['missing',undefined]]);assert.deepEqual(q.groupBy,base.groupBy);for(const selection of [{fromMs:999},{toMs:6001},{fromMs:5000,toMs:4000},{minMS:'1',maxMS:'1'},{maxMS:'0'},{minMS:'0.0000001'}])assert.throws(()=>question(base,selection));assert.throws(()=>question({...base,signal:'logs',summary:undefined},{minMS:'1'}))});
test('population transform preserves reported edges and distinguishes missing from empty',()=>{const values=[{field:{kind:'resource',key:'service.instance.id'},value:''}],q=question(base,{values});assert.equal(q.where.all[1].op,'eq');assert.equal(q.where.all[1].value,'');assert.deepEqual(q.where.all[0],base.where);assert.throws(()=>question(base,{values:[{field:{kind:'bad',key:'x'},value:'x'}]}))});
test('rectangle maps complete raw span buckets to exact half-open ranges',()=>{
 const {rectangle}=require('../ui/population-selection');const data={question:{...base,groupBy:[{kind:'column',key:'duration_bucket'}]},timeSeries:[{fromMs:1000,toMs:2000},{fromMs:2000,toMs:2500}]};
 for(let a=0;a<8;a++)for(let b=a;b<8;b++){
  const before=structuredClone(data),r=rectangle(data,{fromColumn:0,toColumn:1,fromRow:a,toRow:b}),q=question(data.question,r);assert.deepEqual(data,before);assert.equal(q.fromMs,1000);assert.equal(q.toMs,2500);assert.deepEqual(q.where.all[0],base.where);assert.equal(q.offset,0);assert.equal(q.where.all[1].op,'gte');if(b===7)assert.equal(q.where.all.length,2);else assert.equal(q.where.all[2].op,'lte');
 }
 for(const r of [{fromColumn:1,toColumn:0,fromRow:0,toRow:1},{fromColumn:0,toColumn:2,fromRow:0,toRow:1},{fromColumn:0,toColumn:1,fromRow:-1,toRow:1}])assert.throws(()=>rectangle(data,r));
 assert.throws(()=>rectangle({...data,state:'changed_between_reads'},{fromColumn:0,toColumn:1,fromRow:0,toRow:1}));
});
test('region and rest retain the complete parent window and exact uint64 boundaries',()=>{
 const {partition}=require('../ui/population-selection'),before=structuredClone(base),p=partition(base,{fromMs:1500,toMs:5000,minMS:'0.000001',maxMS:'1'});
 for(const q of [p.selected,p.rest]){assert.equal(q.fromMs,base.fromMs);assert.equal(q.toMs,base.toMs);assert.equal(q.offset,0);assert.deepEqual(q.where.all[0],base.where);assert.deepEqual(q.groupBy,base.groupBy)}
 assert.deepEqual(p.selected.where.all[1].all.map(x=>[x.field.key,x.op,x.value]),[['timestamp_ns','gte','1500000000'],['timestamp_ns','lte','4999999999'],['duration_ns','gte','1'],['duration_ns','lte','999999']]);
 assert.deepEqual(p.rest.where.all[1].any.map(x=>[x.op,x.value]),[['lte','1499999999'],['gte','5000000000'],['lte','0'],['gte','1000000']]);assert.deepEqual(base,before);
 assert.throws(()=>partition({...base,signal:'logs'}));assert.throws(()=>partition({...base,numeric:{}}));assert.throws(()=>partition(base,{values:[{field:{kind:'attribute',key:'x'},value:null}]}));
 const zero=partition({...base,fromMs:0,where:undefined},{minMS:'0'});assert(!zero.rest.where.any.some(p=>p.value==='-1'));
 const edge={...base,reportedEdge:{parent:{service:'a',namespace:{mode:'missing'}},child:{service:'b',namespace:{mode:'missing'}}}};assert.deepEqual(partition(edge).rest.reportedEdge,edge.reportedEdge);
});
