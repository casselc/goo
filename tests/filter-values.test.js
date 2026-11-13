require('./support/module-dependencies.cjs').install(__filename);
const {test}=require('node:test'),assert=require('node:assert/strict'),{query,result}=require('../ui/filter-values.mjs');
const scope={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,where:{field:{kind:'column',key:'service'},op:'eq',value:'api'},numeric:{field:{kind:'attribute',key:'latency'},type:'int64'},groupBy:[{kind:'column',key:'service'}],limit:100,offset:100,bucketMs:1000},field={kind:'attribute',key:'release'};
const response=q=>({version:'goo/counted-values/v1',reads:{count:{readAt:'2026-10-24T00:00:00Z',basis:'a'.repeat(64)},groups:{readAt:'2026-10-24T00:00:01Z',basis:'b'.repeat(64)}},question:q,state:'observed',count:'9007199254740995',capturedAt:'2026-10-24T00:00:00Z',basis:'source-read',groups:[{g0:null,count:'1'},{g0:'',count:'1'},{g0:'canary',count:'9007199254740993'}]});
test('lookup preserves population scope and strips presentation/measures while requesting one bounded row',()=>{const q=query(scope,field);assert.deepEqual(q.where,scope.where);assert.deepEqual(q.groupBy,[field]);assert.equal(q.limit,1);assert.equal(q.offset,0);assert.equal(q.numeric,undefined);assert.equal(q.bucketMs,undefined);assert.equal(scope.offset,100);assert.throws(()=>query(null,field));assert.throws(()=>query(scope,{kind:'column',key:'duration_ns'}))});
test('counted groups retain exact integers, empty and missing and require matching scope and independent sums',()=>{const q=query(scope,field),r=response(q),m=result(q,r);assert.equal(m.items[0].count,'9007199254740993');assert.ok(m.items.some(x=>x.value===null));assert.ok(m.items.some(x=>x.value===''));assert.equal(m.total,'9007199254740995');assert.throws(()=>result(q,{...r,count:'1'}));assert.throws(()=>result(q,{...r,state:'changed_between_reads'}));assert.throws(()=>result(q,{...r,question:{...q,environment:'other'}}));assert.throws(()=>result(q,{...r,groups:[{g0:'a',count:1}]}));assert.throws(()=>result(q,{...r,groups:[{g0:'',count:'1'},{g0:'',count:'9007199254740994'}]}));assert.throws(()=>result(q,{...r,groups:Array.from({length:1001},(_,i)=>({g0:String(i),count:'1'}))}));assert.deepEqual(result(q,{...r,state:'no_data',count:'0',groups:[]}).items,[])});


test('counted projection requires its version and refuses full inspection substitution',()=>{const q=query(scope,field),r=response(q);const {version,...inspection}=r;assert.throws(()=>result(q,inspection),/Unsupported/);assert.throws(()=>result(q,{...r,version:'goo/explore/v1'}),/Unsupported/);assert.throws(()=>result(q,{...r,count:'9007199254740995\n'}));});
test('counted query retains typed predicates and reported-edge membership without numeric measures',()=>{const numeric={numeric:{field:{kind:'resource',key:'budget'},type:'uint64',unit:'items',parse:'canonical-decimal'},op:'gte',value:'9007199254740993'},s={...scope,where:{all:[scope.where,numeric]},reportedEdge:{from:'test'}};const q=query(s,field);assert.deepEqual(q.where,s.where);assert.deepEqual(q.reportedEdge,s.reportedEdge);assert.notEqual(q.where,s.where);assert.equal(q.numeric,undefined);});

test('projection retains independent source receipts and refuses absent or invalid receipt metadata',()=>{const q=query(scope,field),r=response(q),m=result(q,r);assert.deepEqual(m.reads,r.reads);assert.notEqual(m.reads,r.reads);assert.throws(()=>result(q,{...r,reads:undefined}),/receipt/);assert.throws(()=>result(q,{...r,reads:{...r.reads,count:{readAt:'bad',basis:'x'}}}),/receipt/);});
test('lookup uses lean endpoint and ignores late completion after scope change or disposal',async()=>{
 const {mount}=require('../ui/filter-values.mjs');const old=global.GooQueries;let loader,resolve,request,paused=0,signal,now=scope;const states=[];
 const editor={setValueLoader:f=>{loader=f},setCountedValues:v=>states.push(v)};
 global.GooQueries={post:(path,q,s)=>{request={path,q};signal=s;return new Promise(r=>{resolve=r})}};
 const owner=mount(editor,{current:()=>now,pause:()=>paused++});
 try{const pending=loader(field);assert.equal(paused,1);assert.equal(request.path,'/api/explore/counted-values');assert.deepEqual(request.q,query(scope,field));now={...scope,environment:'other'};owner.sync();assert.equal(signal.aborted,true);resolve({ok:true,json:async()=>response(request.q)});await pending;assert.equal(states.at(-1),undefined);assert.equal(states.some(s=>s?.state==='ready'),false);
 now=scope;const next=loader(field);owner.dispose();assert.equal(signal.aborted,true);resolve({ok:true,json:async()=>response(request.q)});await next;assert.equal(loader,null);assert.equal(states.at(-1),undefined);
 }finally{global.GooQueries=old}
});

test('422 counted result owns its immutable question and a live source-generation lease',async()=>{
 const {mount}=require('../ui/filter-values.mjs'),old=global.GooQueries;let loader,now=structuredClone(scope);const ready=[];
 const editor={setValueLoader:f=>loader=f,setCountedValues:(data,options)=>{if(data?.state==='ready')ready.push({data,options})}};
 global.GooQueries={post:async(_path,q)=>({ok:true,json:async()=>response(q)})};
 const owner=mount(editor,{current:()=>now});try{
  await loader(field);const first=ready.at(-1);assert(first);assert.deepEqual(first.data.question,query(scope,field));assert.equal(first.options.current(),true);
  now={...scope,environment:'other'};assert.equal(first.options.current(),false,'scope drift refuses before owner.sync');now=structuredClone(scope);owner.invalidate();assert.equal(first.options.current(),false);
  await loader(field);const second=ready.at(-1);assert.equal(second.options.current(),true);assert.equal(first.options.current(),false,'identical scope cannot revive a retired generation');
  owner.dispose();assert.equal(second.options.current(),false);
 }finally{owner.dispose();global.GooQueries=old}
});

test('425 retained counted scope preserves ready DOM custody while suspended without acquisition authority',async()=>{
 const {mount}=require('../ui/filter-values.mjs'),old=global.GooQueries;let loader,active=structuredClone(scope),retained=structuredClone(scope),reads=0,pauses=0;const states=[];
 const editor={setValueLoader:f=>loader=f,setCountedValues:(data,options)=>states.push({data,options})};
 global.GooQueries={post:async(_path,q)=>{reads++;return{ok:true,json:async()=>response(q)}}};
 const owner=mount(editor,{current:()=>active,retainedCurrent:()=>retained,pause:()=>pauses++});try{
  await loader(field);const ready=states.at(-1);assert.equal(ready.data.state,'ready');assert.equal(ready.options.current(),true);const count=states.length;
  active=null;owner.sync();assert.equal(states.length,count,'suspension must not remove the retained result/opener');assert.equal(ready.options.current(),false,'held scope is not active authority');
  assert.equal(await loader(field),false);assert.equal(states.length,count);assert.equal(reads,1);assert.equal(pauses,1,'withdrawn callbacks cannot pause or acquire');
  active=structuredClone(scope);owner.sync();assert.equal(states.length,count);assert.equal(ready.options.current(),true,'return reuses the same completed generation');
  retained={...scope,toMs:62000};assert.equal(ready.options.current(),false);owner.sync();assert.equal(states.at(-1).data,undefined);retained=structuredClone(scope);assert.equal(ready.options.current(),false,'scope replacement retires generation even after identical values return');
 }finally{owner.dispose();global.GooQueries=old}
});

test('425 late counted completion during retained suspension never publishes hidden ready or error state',async()=>{
 const {mount}=require('../ui/filter-values.mjs'),old=global.GooQueries;for(const ok of [true,false]){let loader,active=structuredClone(scope),resolve,request,reads=0;const states=[];
 const editor={setValueLoader:f=>loader=f,setCountedValues:(data,options)=>states.push({data,options})};
 global.GooQueries={post:(_path,q)=>{reads++;request=q;return new Promise(r=>resolve=r)}};
 const owner=mount(editor,{current:()=>active,retainedCurrent:()=>scope});try{
  const pending=loader(field);assert.equal(states.at(-1).data.state,'loading');active=null;owner.sync();resolve({ok,status:503,json:async()=>ok?response(request):{error:'Unavailable'}});await pending;
  assert.equal(reads,1);assert.equal(states.at(-1).data,undefined);assert.equal(states.some(s=>['ready','error'].includes(s.data?.state)),false,'a retained selector cannot authorize hidden publication');
  active=structuredClone(scope);owner.sync();assert.equal(states.at(-1).data,undefined,'return cannot revive a dropped pending read');
 }finally{owner.dispose();global.GooQueries=old}}
});
