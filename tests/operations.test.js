const {test}=require('node:test');
const assert=require('node:assert/strict');
const {select,distribution,waterfallRows}=require('../ui/operations.mjs');
const epoch=1792209000000000000n;
const span=(id,offset,duration,outcome='error')=>({traceID:'a',spanID:id,parentSpanID:'',startNS:String(epoch+offset),durationNS:String(duration),operation:'checkout.reserve',role:'server',outcome,attributes:{}});
test('heatmap partitions exact durations and half-open timestamps at a modern epoch',()=>{
 const rows=[span('a',0n,999999n),span('b',1n,1000000n),span('c',59999999999n,10000000n),span('d',60000000000n,10000000000n),span('e',60000000001n,18446744073709551615n)];
 const cells=distribution(rows,Number(epoch/1000000n),Number(epoch/1000000n)+60000);
 assert.equal(cells.reduce((n,c)=>n+c.count,0),3);
 assert.deepEqual(cells.map(c=>c.band),[0,1,2]);
 for(const cell of cells){const actual=select(rows,cell);assert.equal(actual.length,cell.count);}
 const all=distribution(rows,Number(epoch/1000000n),Number(epoch/1000000n)+120000);
 assert.equal(all.find(c=>c.band===7).count,2);
 assert.equal(all.find(c=>c.band===7).maxNS,undefined);
});
test('operation, role, outcome and trace filtering compose without mutating evidence',()=>{
 const rows=[span('a',0n,1n),span('b',1n,1n,'ok'),{...span('c',2n,1n),role:'client'},{...span('d',3n,1n),operation:'other'},{...span('e',4n,1n),traceID:'b'}];
 const before=JSON.stringify(rows);
 assert.deepEqual(select(rows,{operation:'checkout.reserve',role:'server',outcome:'error',traceID:'a'}).map(s=>s.spanID),['a']);
 assert.equal(select(rows,{operation:'absent'}).length,0);
 assert.equal(JSON.stringify(rows),before);
 const out=waterfallRows(rows,'bound-service');assert.equal(out[0].timestamp_ns,String(epoch));assert.equal(out[0].duration_ns,'1');assert.equal(Object.hasOwn(out[0],'StatusCode'),false);assert.equal(Object.hasOwn(out[1],'StatusCode'),false);
});
test('subject navigation separates complete scoped identities and keeps repeated events',()=>{
 const {subjectKey,subjectEvents,subjectSpans}=require('../ui/operations.mjs');
 const subject={environment:'local',service:'checkout',namespace:'reservation',tenant:'a',id:'same'};
 const event=(s,time)=>({subject:s,timeNS:String(time),name:'accepted'});
 const rows=[{...span('1',0n,1n),events:[event(subject,epoch+2n),event(subject,epoch+2n)]},{...span('2',1n,1n),traceID:'b',events:[event(subject,epoch+1n)]}];
 for(const key of ['environment','service','namespace','tenant','id'])rows.push({...span(key,0n,1n),events:[event({...subject,[key]:'different'},epoch)]});
 rows.push({...span('unscoped',0n,1n),events:[{name:'accepted',timeNS:String(epoch),diagnostic:'subject-binding-required'}]});
 assert.equal(subjectEvents(rows,subject).length,3);
 assert.deepEqual(subjectSpans(rows,subject).map(s=>s.spanID),['1','2']);
 assert.equal(subjectEvents(rows,null).length,0);
 assert.notEqual(subjectKey({...subject,tenant:'a/b',id:'c'}),subjectKey({...subject,tenant:'a',id:'b/c'}));
 assert.equal(subjectEvents(rows,subject)[0].span.spanID,'2');
});

test('declared outcomes never infer reported OTel status, including old, empty and future evidence',()=>{const {waterfallFacts}=require('../ui/operations.mjs');for(const outcome of ['ok','error','panic'])for(const reportedStatus of [undefined,'','Ok','Error','Unset','VendorStatus']){const s={...span('id',0n,1n,outcome),...(reportedStatus===undefined?{}:{reportedStatus})},before=structuredClone(s),row=waterfallRows([s],'svc')[0],facts=waterfallFacts([s])[0];assert.equal(row.operationOutcome,outcome);assert.equal(row.StatusCode,reportedStatus);assert.equal(Object.hasOwn(row,'StatusCode'),reportedStatus!==undefined);assert.equal(facts[0].value,outcome);assert.equal(facts[1].value,reportedStatus===undefined?'Not retained':reportedStatus===''?'Unknown (empty)':reportedStatus==='VendorStatus'?'Unknown (VendorStatus)':reportedStatus);assert.deepEqual(s,before)}});
