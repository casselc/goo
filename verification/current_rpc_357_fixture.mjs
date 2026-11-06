import fs from 'node:fs';import assert from 'node:assert/strict';import * as D from '../ui/service-dashboard.mjs';
const {scope,service,question,rows}=JSON.parse(fs.readFileSync(0,'utf8'));
const clone=structuredClone,raw=rows.map(r=>({SpanId:r.SpanId,ServiceName:r.ServiceName,SpanKind:r.SpanKind,ResourceAttributes:clone(r.ResourceAttributes),Attributes:clone(r.SpanAttributes),timestamp_ns:String(BigInt(r.timestamp_ms)*1000000n)}));
const data={state:'observed',question:clone(question),rows:raw},before=JSON.stringify({scope,question,rows,data});
const current=D.rpcServerCurrentObserved(scope,service,data,question),legacy=D.rpcServerObserved(scope,service,data,question);
for(const fn of [D.rpcServerCurrentObserved,D.rpcServerObserved]){
 const changed=clone(data);changed.question.fromMs++;assert.equal(fn(scope,service,changed,question),false);
 const absent=clone(data);absent.state='no_data';assert.equal(fn(scope,service,absent,question),false);
 const poison=clone(data);poison.rows=raw.map(r=>({...r,Attributes:[]}));assert.equal(fn(scope,service,poison,question),false);
 const outside=clone(data);outside.rows=raw.map(r=>({...r,timestamp_ns:String(BigInt(scope.toMs)*1000000n)}));assert.equal(fn(scope,service,outside,question),false);
 const beyond={state:'observed',question:clone(question),rows:[...Array.from({length:100},()=>({...clone(raw[0]),ServiceName:'foreign'})),...raw]};assert.equal(fn(scope,service,beyond,question),false);
}
const input={scope:clone(scope),service,signals:['spans','logs'],profile:'all-spans',instruments:[]};const planned=D.planningInput(input);assert.equal(planned.profile,'all-spans');planned.scope.fromMs++;assert.notDeepEqual(planned.scope,input.scope);assert.equal(JSON.stringify({scope,question,rows,data}),before);
process.stdout.write(JSON.stringify({current,legacy,complete:true}));
