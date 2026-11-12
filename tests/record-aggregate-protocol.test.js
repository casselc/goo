const {test}=require('node:test'),assert=require('node:assert/strict');
const P=require('../ui/record-aggregate-protocol.mjs'),{aggregateFixture}=require('./fixtures/record-aggregate-fixture.cjs');
const clone=x=>structuredClone(x);
test('pure aggregate admission preserves service validation without changing either transport or requested intent',()=>{
 const {widget,result}=aggregateFixture(),requested=clone(widget.query),before=clone({requested,result});assert.equal(P.recordAggregates(requested,result),result);assert.deepEqual({requested,result},before);
 result.question.where={field:{kind:'column',key:'service'},op:'eq',value:'other'};assert.deepEqual(requested,before.requested);assert.throws(()=>P.recordAggregates(requested,result),/question differs/);
});
test('aggregate scope identity accepts Go omissions but refuses changed selectors, windows, measures and typed policies',()=>{
 const {widget,result}=aggregateFixture(),q={...clone(widget.query),environment:'',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},groupTimeSeries:false,reportedEdge:null,where:{field:{kind:'column',key:'service'},op:'eq',value:''}},wire={...result,question:clone(q)};delete wire.question.environment;delete wire.question.groupTimeSeries;delete wire.question.reportedEdge;delete wire.question.where.value;assert.equal(P.recordAggregates(q,wire),wire);
 for(const mutate of [r=>r.question.fromMs++,r=>r.question.toMs++,r=>r.question.where.value='api',r=>r.question.groupBy[0].key='name',r=>r.question.summary='',r=>r.question.environmentSelector.policy='current-only',r=>r.question.environmentSelector.scope.value='other']){const changed=clone(wire);mutate(changed);assert.throws(()=>P.recordAggregates(q,changed));}
});
test('closed aggregate DTO rejects unattested datasets, inspection payloads and nested unknown authority',()=>{
 const {widget,result}=aggregateFixture();for(const key of ['viewDataset','summaryDataset','timeDataset','histogramDataset','recordPage','rows','fields','hasMore','resultToken','plan','quality','completeness','unknown'])assert.throws(()=>P.recordAggregates(widget.query,{...result,[key]:null}),/Invalid record aggregates/);
 for(const mutate of [r=>r.groups[0].viewDataset={},r=>r.groups[0].g1='other',r=>r.duration.sourceBasis='forged',r=>r.spanStatus.quality={},r=>r.timeSeries[0].summaryDataset={},r=>r.reads.other={},r=>r.reads.count.complete=true]){const changed=clone(result);mutate(changed);assert.throws(()=>P.recordAggregates(widget.query,changed),/Invalid record aggregates/);}
 const getter={...result};Object.defineProperty(getter,'viewDataset',{get(){throw Error('getter invoked')},enumerable:true});assert.throws(()=>P.recordAggregates(widget.query,getter),/Invalid record aggregates/);assert.throws(()=>P.recordAggregates(widget.query,{...result,[Symbol('extra')]:1}),/Invalid record aggregates/);
});
test('aggregate receipts and requested measures remain mandatory; changed reads retain diagnostic arithmetic',()=>{
 const {widget,result}=aggregateFixture();for(const mutate of [r=>delete r.reads.groups,r=>r.reads.count.readAt='invalid',r=>r.reads.timeSeries.basis='bad',r=>delete r.duration,r=>delete r.groups[0].spanStatus,r=>r.duration.p95Ns='1',r=>r.timeSeries[0].toMs++,r=>r.groups[0].count='1',r=>r.count='02']){const changed=clone(result);mutate(changed);assert.throws(()=>P.recordAggregates(widget.query,changed));}
 const diagnostic={...clone(result),state:'changed_between_reads',groups:[],timeSeries:[{fromMs:1000,toMs:2000,state:'no_data',count:'0',duration:{state:'no_data',count:'0',sumNs:'0'},spanStatus:{count:'0',error:'0',ok:'0',unset:'0',unknown:'0'}}]};assert.equal(P.recordAggregates(widget.query,diagnostic),diagnostic);assert.equal(diagnostic.viewDataset,undefined);
 const zero={...clone(result),state:'no_data',count:'0',groups:[],duration:{state:'no_data',count:'0',sumNs:'0'},spanStatus:{count:'0',error:'0',ok:'0',unset:'0',unknown:'0'}};zero.timeSeries=[{fromMs:1000,toMs:2000,state:'no_data',count:'0',duration:clone(zero.duration),spanStatus:clone(zero.spanStatus)}];assert.equal(P.recordAggregates(widget.query,zero),zero);zero.groups=[{g0:'api',count:'1',duration:result.duration,spanStatus:result.spanStatus}];assert.throws(()=>P.recordAggregates(widget.query,zero));
});
test('projection capability preflight keeps numeric column cohorts on the existing full read path',()=>{
 const {widget}=aggregateFixture(),q={...widget.query,summary:'span-duration'};assert.equal(P.canRead(q),true);for(const [signal,key]of [['spans','duration_ns'],['spans','timestamp_ns'],['logs','severity_number']]){const query={...q,signal,summary:signal==='spans'?'span-duration':undefined,spanStatus:signal==='spans',groupBy:[{kind:'column',key}]};assert.equal(P.canRead(query),false);assert.equal(P.canRead({...query,groupBy:[{kind:'attribute',key}]}),true)}
 for(const patch of [{groupBy:[]},{groupBy:[{kind:'column',key:'not_a_column'}]},{offset:1},{numeric:{}},{numericHistogram:{}},{groupTimeSeries:true},{summary:'unsupported'},{signal:'logs',summary:'span-duration'}])assert.equal(P.canRead({...q,...patch}),false);assert.equal(P.canRead({...q,groupBy:[{kind:'resource',key:'service.namespace'}]}),true);
});

test('unrequested time receipts cannot enter a count/groups-only projection',()=>{const {widget,result}=aggregateFixture(),q={...widget.query};delete q.bucketMs;const data={...result,question:q};delete data.timeSeries;assert.throws(()=>P.recordAggregates(q,data),/Invalid record aggregates/);data.reads={count:result.reads.count,groups:result.reads.groups};assert.equal(P.recordAggregates(q,data),data);});

test('two string dimensions admit qualified tuples rather than collapsing equal first values',()=>{
 const {widget,result}=aggregateFixture(),q=clone(widget.query);delete q.bucketMs;q.summary='span-duration';q.groupBy=[{kind:'column',key:'service'},{kind:'resource',key:'service.namespace'}];const data=clone(result);data.question=clone(q);delete data.timeSeries;delete data.reads.timeSeries;data.groups[0].g0='api';data.groups[0].g1=null;
 assert.equal(P.canRead(q),true);assert.equal(P.recordAggregates(q,data),data);
 const extra=clone(data.groups[0]);extra.g1='';const diagnostic={...clone(data),state:'changed_between_reads',groups:[clone(data.groups[0]),extra]};assert.equal(P.recordAggregates(q,diagnostic),diagnostic);
 for(const edit of [d=>delete d.groups[0].g1,d=>d.groups[0].g1=1,d=>d.groups.push(clone(d.groups[0]))]){const bad=clone(data);edit(bad);assert.throws(()=>P.recordAggregates(q,bad));}
 assert.equal(P.canRead({...q,groupBy:[q.groupBy[0],{kind:'column',key:'duration_ns'}]}),false);assert.equal(P.canRead({...q,groupBy:[...q.groupBy,{kind:'attribute',key:'third'}]}),false);
});
