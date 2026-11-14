import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Severity from '../ui/log-severity.mjs';
import {validateRecordQuestion,questionLink,readQuestionLink} from '../ui/question-protocol.mjs';
import {identity} from '../ui/question-session.mjs';
import {recordAggregates,canRead} from '../ui/record-aggregate-protocol.mjs';
import {adoptExploration} from '../ui/source-snapshots.mjs';
import {eligible as displayEligible} from '../ui/view-declaration.mjs';
import {eligibilityReason as cohortReason} from '../ui/count-cohort-context.mjs';
import {packageQuestion} from '../ui/telemetry-package-intent.mjs';
const copy=value=>structuredClone(value);
const summary=values=>({count:'0',...Object.fromEntries(Severity.keys.map(key=>[key,'0'])),...values});
function fixture(){
 const q={signal:'logs',environment:'prod',fromMs:1000,toMs:3000,limit:1,offset:0,groupBy:[{kind:'column',key:'severity'}],bucketMs:1000,logSeverity:true},receipt={readAt:'2026-11-14T00:00:00Z',basis:'a'.repeat(64)};
 const data={version:'goo/record-aggregates/v1',question:copy(q),capturedAt:receipt.readAt,basis:'b'.repeat(64),count:'3',state:'observed',groups:[{g0:'ERROR',count:'2',logSeverity:summary({count:'2',trace:'1',unspecified:'1'})},{g0:'TRACE',count:'1',logSeverity:summary({count:'1',error:'1'})}],logSeverity:summary({count:'3',trace:'1',error:'1',unspecified:'1'}),timeSeries:[{fromMs:1000,toMs:2000,count:'2',state:'observed',logSeverity:summary({count:'2',trace:'1',error:'1'})},{fromMs:2000,toMs:3000,count:'1',state:'observed',logSeverity:summary({count:'1',unspecified:'1'})}],warnings:[],reads:{count:copy(receipt),groups:copy(receipt),timeSeries:copy(receipt)}};
 return {q,data};
}
test('OTel severity classes partition every physical UInt8 value with exact band predicates',()=>{
 const expected={trace:4,debug:4,info:4,warn:4,error:4,fatal:4,unspecified:1,invalid:231},actual=Object.fromEntries(Severity.keys.map(key=>[key,0]));
 const matches=(p,n)=>p.all?p.all.every(child=>matches(child,n)):p.op==='eq'?n===Number(p.value):p.op==='gte'?n>=Number(p.value):n<=Number(p.value);
 for(let n=0;n<=255;n++){const key=Severity.classify(n);actual[key]++;assert.equal(Severity.classify(String(n)),key);assert.deepEqual(Severity.keys.filter(band=>matches(Severity.bandPredicate(band),n)),[key]);}
 assert.deepEqual(actual,expected);assert.equal(Severity.classify(0),'unspecified');assert.equal(Severity.classify(24),'fatal');assert.equal(Severity.classify(25),'invalid');
 for(const value of [-1,256,1.5,NaN,Infinity,null,undefined,'01','1.0','1e1','','-1',' 1'])assert.throws(()=>Severity.classify(value));assert.throws(()=>Severity.bandPredicate('all'));
});
test('numeric severity query intent is logs-only, ordinary, shareable and identity-aware',()=>{
 const {q}=fixture();assert.equal(validateRecordQuestion(q),q);assert.equal(canRead(q),true);assert.deepEqual(readQuestionLink(questionLink(q)).question,q);
 const plain={...q};delete plain.logSeverity;assert.equal(identity({...plain,logSeverity:false}),identity(plain));assert.notEqual(identity(q),identity(plain));
 assert.equal(displayEligible(q),false);assert.match(cohortReason(q),/numeric log severity/);assert.equal(displayEligible(plain),true);
 const packaged=packageQuestion({source:{adapter:'clickhouse.otel/v1',signal:'logs',environment:'prod'},logSeverity:true,groupBy:q.groupBy,bucketMs:q.bucketMs},q);assert.equal(packaged.logSeverity,true);assert.deepEqual(packaged.groupBy,q.groupBy);packaged.groupBy[0].key='service';assert.equal(q.groupBy[0].key,'severity');
 for(const patch of [{logSeverity:null},{logSeverity:'true'},{signal:'spans'},{numeric:{field:{kind:'attribute',key:'size'},type:'uint64',parse:'canonical-decimal',unit:'1'}},{numericHistogram:{binning:'powers-of-two'}},{groupTimeSeries:true}]){assert.throws(()=>validateRecordQuestion({...q,...patch}));assert.equal(canRead({...q,...patch}),false);}
});
test('aggregate numeric bands conserve exact counts across independent text groups and intervals',()=>{
 const {q,data}=fixture(),before=copy(data);assert.equal(recordAggregates(q,data),data);assert.deepEqual(data,before);assert.equal(data.groups[0].g0,'ERROR');assert.equal(data.groups[0].logSeverity.error,'0');
 const huge=summary({count:'18446744073709551615',fatal:'18446744073709551615'});assert.equal(Severity.validateSummary(huge),huge);const empty=summary();assert.equal(Severity.validateSummary(empty),empty);
 const zero=copy(data);zero.state='no_data';zero.count='0';zero.logSeverity=summary();zero.groups=[];for(const point of zero.timeSeries){point.count='0';point.state='no_data';point.logSeverity=summary();}assert.equal(recordAggregates(q,zero),zero);
});
test('summary admission rejects missing, malformed, overflowing and hidden DTO fields before access',()=>{
 const valid=summary({count:'1',info:'1'});
 for(const mutate of [s=>delete s.info,s=>s.extra='0',s=>s.info='01',s=>s.count='18446744073709551616',s=>s.info='-1',s=>s.count='2',s=>s.info=1,s=>s.info=null,s=>s[Symbol('extra')]='0']){const value=copy(valid);mutate(value);assert.throws(()=>Severity.validateSummary(value));}
 for(const key of ['info','count']){let accessed=false;const getter=copy(valid);Object.defineProperty(getter,key,{get(){accessed=true;return '1'},enumerable:true});assert.throws(()=>Severity.validateSummary(getter));assert.equal(accessed,false);}
 assert.throws(()=>Severity.validateSummary({...valid},'2'));assert.throws(()=>Severity.validateSummary(null));
});
test('requested severity is mandatory at every aggregate projection and equal totals cannot hide class drift',()=>{
 const {q,data}=fixture();
 for(const mutate of [r=>delete r.logSeverity,r=>delete r.groups[0].logSeverity,r=>delete r.timeSeries[0].logSeverity,r=>r.groups[0].logSeverity.count='1',r=>r.logSeverity.info='1',r=>{r.groups[1].logSeverity.error='0';r.groups[1].logSeverity.fatal='1'},r=>{r.timeSeries[0].logSeverity.trace='0';r.timeSeries[0].logSeverity.debug='1'}]){const changed=copy(data);mutate(changed);assert.throws(()=>recordAggregates(q,changed));}
 // A server-declared changed read retains internally valid diagnostics only.
 const diagnostic=copy(data);diagnostic.state='changed_between_reads';diagnostic.timeSeries[0].logSeverity.trace='0';diagnostic.timeSeries[0].logSeverity.debug='1';assert.equal(recordAggregates(q,diagnostic),diagnostic);
 delete diagnostic.timeSeries[0].logSeverity;assert.throws(()=>recordAggregates(q,diagnostic));
 const plain=copy(q);delete plain.logSeverity;const unsolicited=copy(data);unsolicited.question=plain;assert.throws(()=>recordAggregates(plain,unsolicited));
});
test('ordinary exploration adoption validates numeric severity without granting new dataset authority',()=>{
 const {q,data}=fixture();assert.equal(adoptExploration(data,q).logSeverity,data.logSeverity);
 for(const mutate of [r=>delete r.logSeverity,r=>delete r.groups[0].logSeverity,r=>delete r.timeSeries[1].logSeverity,r=>{r.timeSeries[1].logSeverity.unspecified='0';r.timeSeries[1].logSeverity.invalid='1'}]){const changed=copy(data);mutate(changed);assert.throws(()=>adoptExploration(changed,q));}
 const diagnostic=copy(data);diagnostic.state='changed_between_reads';diagnostic.groups=[];assert.equal(adoptExploration(diagnostic,q).state,'changed_between_reads');
 const plain=copy(q);delete plain.logSeverity;const unsolicited=copy(data);unsolicited.question=plain;assert.throws(()=>adoptExploration(unsolicited,plain));
});
