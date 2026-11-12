import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as T from '../ui/service-triage-model.mjs';
import * as P from '../ui/record-aggregate-protocol.mjs';
const scope={environment:'prod',fromMs:61000,toMs:121000},clone=x=>structuredClone(x);
function group(service,ns,count,error='0',duration='10') {return {g0:service,g1:ns,count,spanStatus:{count,error,ok:(BigInt(count)-BigInt(error)).toString(),unset:'0',unknown:'0'},duration:{state:'observed',count,sumNs:(BigInt(count)*BigInt(duration)).toString(),minNs:duration,p50Ns:duration,p95Ns:duration,p99Ns:duration,maxNs:duration}};}
function response(q,groups){
 const count=groups.reduce((n,g)=>n+BigInt(g.count),0n).toString(),sumNs=groups.reduce((n,g)=>n+BigInt(g.duration.sumNs),0n).toString(),receipt={basis:'a'.repeat(64),readAt:'2026-11-12T00:00:00Z'},values=groups.map(g=>BigInt(g.duration.minNs)),min=values.reduce((a,b)=>a<b?a:b,values[0]),max=values.reduce((a,b)=>a>b?a:b,values[0]);
 const duration=count==='0'?{state:'no_data',count,sumNs}:{state:'observed',count,sumNs,minNs:String(min),p50Ns:String(max),p95Ns:String(max),p99Ns:String(max),maxNs:String(max)};
 return {version:'goo/record-aggregates/v1',question:clone(q),capturedAt:receipt.readAt,basis:receipt.basis,state:count==='0'?'no_data':'observed',count,groups:clone(groups),duration,spanStatus:Object.fromEntries(['count','error','ok','unset','unknown'].map(k=>[k,groups.reduce((n,g)=>n+BigInt(g.spanStatus[k]),0n).toString()])),warnings:[],reads:{count:clone(receipt),groups:clone(receipt)}};
}
function snapshot(current,baseline,s=scope){const p=T.plan(s);return T.admit(p,response(p.current,current),response(p.baseline,baseline));}
test('fleet plan freezes role, qualified grouping, exact environment/namespace and adjacent windows',()=>{
 const s={environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'unreported'}},namespace:{mode:'value',value:''},fromMs:61000,toMs:121000},p=T.plan(s);
 assert.deepEqual(p.current.groupBy,[{kind:'column',key:'service'},{kind:'resource',key:'service.namespace'}]);assert.equal(P.canRead(p.current),true);assert.equal(p.baseline.fromMs,1000);assert.equal(p.baseline.toMs,61000);assert.deepEqual(p.current.environmentSelector,s.environmentSelector);assert.deepEqual(p.current.where.all,[{field:{kind:'column',key:'kind'},op:'eq',value:'Server'},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''}]);s.namespace.value='mutated';assert.equal(p.scope.namespace.value,'');
 for(const bad of [{...scope,fromMs:1000},{...scope,toMs:61000},{...scope,service:'api'},{...scope,namespace:{mode:'all'}},{...scope,fromMs:4102444800000,toMs:4102444860000},{...scope,namespace:{mode:'value',value:'雪'.repeat(43)}},{...scope,toMs:scope.fromMs+7200001},{...scope,fromMs:61000.5}])assert.throws(()=>T.plan(bad));
});
test('qualified service union preserves missing/empty names, baseline-only services and non-server catalog entries',()=>{
 const value=snapshot([group('api',null,'3','1'),group('api','','2')],[group('api',null,'1'),group('retired','team','4','2')]),catalog=[{service:'logs',namespace:{mode:'missing'}},{service:'api',namespace:{mode:'missing'}}],rows=T.rows(value,catalog);assert.equal(rows.length,4);const missing=rows.find(r=>r.service==='api'&&r.namespace.mode==='missing'),empty=rows.find(r=>r.service==='api'&&r.namespace.mode==='value');assert.equal(missing.countDelta,'2');assert.equal(missing.errorDelta.n,'1');assert.equal(missing.errorDelta.d,'3');assert.equal(empty.countDelta,'2');assert.equal(empty.baseline,null);assert.equal(empty.errorDelta,null);assert.equal(rows.find(r=>r.service==='retired').countDelta,'-4');const logs=rows.find(r=>r.service==='logs');assert.equal(logs.current,null);assert.equal(logs.baseline,null);assert.equal(logs.countDelta,null);assert.equal(logs.durationDelta,null);
 const q=T.question(value.plan,missing);assert.deepEqual(q.where.all.slice(-2),[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'resource',key:'service.namespace'},op:'missing'}]);assert.equal(q.fromMs,scope.fromMs);assert.equal(value.plan.current.where.all.length,1);
});
test('exact count/error/mean/delta ranking exceeds JavaScript integer precision and leaves absent latency last',()=>{
 const huge='9007199254740993',value=snapshot([group('big',null,huge,'1','11'),group('small',null,'9007199254740992','1','12')],[group('big',null,'9007199254740992'),group('small',null,'9007199254740992')]),rows=T.rows(value,[{service:'unobserved',namespace:{mode:'missing'}}]);assert.equal(T.sortRows(rows,'count')[0].service,'big');assert.equal(T.sortRows(rows,'error-rate')[0].service,'small');assert.equal(T.sortRows(rows,'mean')[0].service,'small');assert.equal(T.sortRows(rows,'p95')[0].service,'small');assert.equal(T.sortRows(rows,'count-change')[0].service,'big');assert.equal(rows.find(r=>r.service==='big').countDelta,'1');assert.deepEqual(rows.find(r=>r.service==='big').durationDelta,{n:(BigInt(huge)*9007199254740992n).toString(),d:(BigInt(huge)*9007199254740992n).toString()});assert.equal(T.sortRows(rows,'mean').at(-1).service,'unobserved');
});
test('admission refuses altered scope, changed/partial aggregates and duplicate qualified tuples',()=>{
 const p=T.plan(scope),current=response(p.current,[group('api',null,'2','1')]),baseline=response(p.baseline,[]),accepted=T.admit(p,current,baseline);current.groups[0].count='8';assert.equal(accepted.current.groups[0].count,'2');
 for(const mutate of [d=>d.question.fromMs++,d=>d.question.environment='other',d=>d.state='changed_between_reads',d=>d.count='3',d=>d.groups.push(clone(d.groups[0])),d=>delete d.groups[0].g1,d=>d.groups[0].g0=null,d=>d.groups[0].g1=3,d=>d.groups[0].duration.sumNs='21',d=>d.groups[0].spanStatus.error='0']){const data=clone(accepted.current);mutate(data);assert.throws(()=>T.admit(p,data,baseline));}
 const altered=clone(p);altered.baseline.toMs++;assert.throws(()=>T.admit(altered,accepted.current,baseline));
 const selected=T.plan({...scope,namespace:{mode:'value',value:'team'}});assert.throws(()=>T.admit(selected,response(selected.current,[group('api','other','1')]),response(selected.baseline,[])));
});
test('two empty observations keep catalog rows unavailable and zero counts separate from unknown rates',()=>{
 const value=snapshot([],[]),rows=T.rows(value,[{service:'metrics',namespace:{mode:'missing'}}]);assert.equal(rows.length,1);assert.equal(rows[0].countDelta,null);assert.equal(rows[0].errorDelta,null);assert.equal(rows[0].p95Delta,null);assert.equal(value.current.state,'no_data');
});

test('unspecified declared namespaces remain unavailable rather than aliases of missing observed namespace',()=>{
 const value=snapshot([group('api',null,'2')],[]),items=T.rows(value,[{service:'api',namespace:null}]);assert.equal(items.length,2);const wildcard=items.find(r=>!r.qualified),observed=items.find(r=>r.qualified);assert.equal(wildcard.namespace,null);assert.equal(wildcard.current,null);assert.notEqual(wildcard.key,observed.key);assert.throws(()=>T.question(value.plan,wildcard));assert.equal(T.sortRows(items,'name').length,2);
});

test('accepted catalog declarations do not inherit the independent source-group limit',()=>{
 const value=snapshot([],[]),catalog=Array.from({length:1100},(_,i)=>({service:'declared-'+i,namespace:null})),items=T.rows(value,catalog);assert.equal(items.length,1100);assert.equal(items.every(row=>row.qualified===false&&row.current===null&&row.baseline===null&&row.countDelta===null),true);assert.throws(()=>T.rows(value,{}),/catalog/);
});
