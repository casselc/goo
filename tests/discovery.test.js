const {test}=require('node:test');const assert=require('node:assert/strict');
const {millisecondsToNS,surroundingQuestion,sameLog,severity,searchPredicates}=require('../ui/discovery.mjs');
test('duration search preserves exact ns and rejects coercion',()=>{assert.equal(millisecondsToNS('0.000001'),'1');assert.equal(millisecondsToNS('18446744073709.551615'),'18446744073709551615');for(const s of ['1e2','-1','1.0000001','18446744073709.551616','', ' 1 '])assert.throws(()=>millisecondsToNS(s));});
test('surrounding logs retain environment and fixed bounds and exact service namespace but omit original search filters',()=>{const origin={signal:'logs',environment:'prod',fromMs:1000,toMs:100000,where:{any:[{field:{kind:'column',key:'body'},op:'contains',value:'error'}]},groupBy:[{kind:'column',key:'severity'}]};const row={timestamp_ns:'1500123456',ServiceName:'worker',ResourceAttributes:{}};const q=surroundingQuestion(origin,row);assert.deepEqual(q,{signal:'logs',environment:'prod',fromMs:1000,toMs:31501,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'worker'},{field:{kind:'resource',key:'service.namespace'},op:'missing'}]},limit:100,offset:0});assert.equal(surroundingQuestion(origin,row,300000,100).toMs,100000);assert.equal(origin.where.any[0].value,'error');assert.throws(()=>surroundingQuestion(origin,{...row,timestamp_ns:'100000000000'}));assert.throws(()=>surroundingQuestion(origin,row,300001));});
test('anchor matches retain nanoseconds and literal values; severity is display only',()=>{const a={timestamp_ns:'1500123456',ServiceName:'a',Body:'\n<img>',ResourceAttributes:{b:'2',a:'1'}};assert.ok(sameLog(a,{...a,ResourceAttributes:{a:'1',b:'2'}}));assert.ok(!sameLog(a,{...a,timestamp_ns:'1500123457'}));assert.equal(severity('ERROR2'),'error');assert.equal(severity('<script>'),'unset');const predicates=searchPredicates('logs',{message:'a.*b',level:'ERROR',minimum:'invalid'});assert.equal(predicates[0].value,'a.*b');assert.equal(predicates[0].op,'contains');assert.equal(searchPredicates('spans',{minimum:'0.1',errors:true})[0].value,'100000');});
test('surrounding planning distinguishes reported namespaces and detaches typed environment',()=>{
 const origin={signal:'logs',environmentSelector:{version:'goo/environment-selector/v1',scope:{mode:'named',value:'prod'},policy:'effective'},fromMs:1000,toMs:100000,where:{field:{kind:'column',key:'body'},op:'contains',value:'original'}};
 for(const namespace of [undefined,'','payments']){
  const row={timestamp_ns:'50000999999',ServiceName:'worker',ResourceAttributes:Object.create(null)};
  if(namespace!==undefined)row.ResourceAttributes['service.namespace']=namespace;
  const before=structuredClone(origin),q=surroundingQuestion(origin,row,1000,100);
  assert.deepEqual([q.fromMs,q.toMs,q.limit,q.offset],[49000,51001,100,100]);
  assert.deepEqual(q.where.all[1],{field:{kind:'resource',key:'service.namespace'},op:namespace===undefined?'missing':'eq',...(namespace===undefined?{}:{value:namespace})});
  q.environmentSelector.scope.value='changed';q.where.all[0].value='changed';assert.deepEqual(origin,before);assert.equal(row.ServiceName,'worker');
 }
});
test('surrounding planning refuses unknown or invalid qualified identity',()=>{
 const origin={environment:'prod',fromMs:1000,toMs:61000},valid={timestamp_ns:'2000000000',ServiceName:'worker',ResourceAttributes:{}};
 for(const row of [{...valid,ServiceName:''},{...valid,ServiceName:null},{...valid,ServiceName:'a\0b'},{...valid,ServiceName:'é'.repeat(513)}, {timestamp_ns:valid.timestamp_ns,ServiceName:'worker'},... [null,[],new Map(),Object.create({}), 'unknown'].map(ResourceAttributes=>({...valid,ResourceAttributes})),...[null,undefined,1,{},'a\0b','é'.repeat(513)].map(value=>({...valid,ResourceAttributes:{'service.namespace':value}}))])assert.throws(()=>surroundingQuestion(origin,row));
});

test('surrounding planning enforces the bounded paging offset',()=>{const q={environment:'prod',fromMs:1000,toMs:61000},row={timestamp_ns:'2000000000',ServiceName:'api',ResourceAttributes:{}};for(const offset of [-1,0.5,'100',NaN,1001])assert.throws(()=>surroundingQuestion(q,row,30000,offset));assert.equal(surroundingQuestion(q,row,30000,1000).offset,1000)});
