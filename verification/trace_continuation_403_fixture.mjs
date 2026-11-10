import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Position from '../ui/trace-search-position.mjs';
import * as Search from '../ui/trace-search.mjs';
import * as Saved from '../ui/saved-capture-protocol.mjs';
import * as Questions from '../ui/question-protocol.mjs';
const vectors=JSON.parse(fs.readFileSync(0,'utf8'));let controls=0;
globalThis.fetch=()=>{throw Error('Page admission must not read telemetry')};
for(const v of vectors){
 const before=JSON.stringify(v);assert.equal(Position.requestHash(v.request),v.hash);assert.equal(Search.validateResult(v.result,v.request,v.result.position),v.result);Saved.validateVisit(v.page,v.page.receipt.id,'');
 const step=Questions.traceSearchStep(v.request);assert.equal(Object.hasOwn(step.traceSearch,'position'),false);assert.equal(Object.hasOwn(Questions.captureQuestion(step),'traceSearchPosition'),false);
 for(const mutate of [q=>q.limit=q.limit===1?2:1,q=>q.fromMs++,q=>q.toMs++,q=>q.conditions.reverse(),q=>q.conditions[0]={field:{kind:'column',key:'service'},op:'eq',value:'foreign'},q=>{delete q.environmentSelector;q.environment='foreign'}]){
  const q=structuredClone(v.request);mutate(q);if(JSON.stringify(q)===JSON.stringify(v.request))continue;const p=Position.makePosition(v.request,'f'.repeat(32));assert.throws(()=>Position.validatePosition(p,q));controls++;
 }
 for(const mutate of [p=>p.requestHash='a'.repeat(64),p=>p.lastTraceId='0'.repeat(32),p=>p.lastTraceId='ABCDEF'.repeat(5)+'AB',p=>p.version='wrong',p=>p.extra=true]){const p=Position.makePosition(v.request,'f'.repeat(32));mutate(p);assert.throws(()=>Position.validatePosition(p,v.request));controls++;}
 for(const mutate of [r=>r.remainingTraces='-1',r=>r.remainingTraces='18446744073709551616',r=>r.hasMore=!r.hasMore,r=>r.request.toMs++,r=>r.position={version:Position.version,requestHash:'a'.repeat(64),lastTraceId:'f'.repeat(32)},r=>r.nextPosition={version:Position.version,requestHash:v.hash,lastTraceId:'0'.repeat(32)}]){const r=structuredClone(v.result);mutate(r);assert.throws(()=>Search.validateResult(r,v.request,v.result.position));controls++;}
 for(const mutate of [p=>p.receipt.question.traceSearch.toMs++,p=>p.receipt.id='bad',p=>p.receipt.capturedAt='not-clock',p=>p.explore={},p=>p.traceSearch.remainingTraces='-1']){const p=structuredClone(v.page);mutate(p);assert.throws(()=>Saved.validate(p));controls++;}
 if(v.result.position){const p=structuredClone(v.page);delete p.receipt.question.traceSearchPosition;assert.throws(()=>Saved.validate(p));controls++;}
 assert.equal(JSON.stringify(v),before);
}
process.stdout.write(JSON.stringify({Cases:vectors.length,Controls:controls}));
