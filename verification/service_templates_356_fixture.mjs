import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as T from '../ui/view-template-handoff.mjs';
import * as C from '../ui/view-composition.mjs';
const input=JSON.parse(fs.readFileSync(0,'utf8')),{source,target,baseline,expanded}=input;
const clone=structuredClone;
function measures(q,count){const d={};if(q.summary)d.duration=count==='0'?{state:'no_data',count,sumNs:'0'}:{state:'observed',count,sumNs:'20',minNs:'10',p50Ns:'10',p95Ns:'10',p99Ns:'10',maxNs:'10'};if(q.spanStatus)d.spanStatus={count,error:'0',ok:count,unset:'0',unknown:'0'};return d}
function envelope(q,empty){const count=empty?'0':'2',clock='2026-11-06T00:00:00Z',receipt={readAt:clock,basis:'a'.repeat(64)},data={version:'goo/record-aggregates/v1',question:clone(q),capturedAt:clock,basis:'b'.repeat(64),count,state:empty?'no_data':'observed',warnings:[],reads:{count:clone(receipt),groups:clone(receipt)},groups:empty?[]:[{g0:'destination',count,...measures(q,count)}],...measures(q,count)};if(q.bucketMs){data.reads.timeSeries=clone(receipt);data.timeSeries=Array.from({length:Math.ceil((q.toMs-q.fromMs)/q.bucketMs)},(_,i)=>{const n=i===0?count:'0';return{fromMs:q.fromMs+i*q.bucketMs,toMs:Math.min(q.fromMs+(i+1)*q.bucketMs,q.toMs),state:n==='0'?'no_data':'observed',count:n,...measures(q,n)}})}return data}
const widget={key:'fixture',title:'Reusable destination',state:'supported',query:clone(target)},data=envelope(target,false),before=clone({source,target,widget,data});
for(const empty of [false,true])assert.deepEqual(T.templateDestination(widget,envelope(target,empty)),target);
assert.deepEqual(T.templateQuestion(target),target);
const detached=T.templateDestination(widget,data);detached.where.all[0].value='mutated destination';assert.deepEqual({source,target,widget,data},before);
for(const poison of [w=>w.state='unsupported',w=>w.state='not_observed',w=>w.query.offset=1,w=>w.query.groupBy=[],w=>w.query.groupTimeSeries=true,w=>w.query.numeric={field:{kind:'attribute',key:'bytes'},type:'uint64',parse:'canonical-decimal',unit:'By'},w=>w.query.numericHistogram={field:{kind:'attribute',key:'bytes'}},w=>w.query.reportedEdge={parent:{service:'a'},child:{service:'b'}},w=>{w.measurement={};delete w.query}]){const bad=clone(widget);poison(bad);assert.throws(()=>T.templateDestination(bad,data))}
for(const poison of [d=>d.state='changed_between_reads',d=>d.question.fromMs++,d=>d.groups[0].count='3',d=>d.rows=[],d=>delete d.reads.groups]){const bad=clone(data);poison(bad);assert.throws(()=>T.templateDestination(widget,bad))}
const destination=T.templateDestination(widget,data),request=T.authoredRequest(source,destination,baseline);
if(!expanded){process.stdout.write(JSON.stringify(request));}
else{
 const checked=T.authoredResponse(expanded,source,destination,baseline);assert.deepEqual(checked.queries,[target]);T.stagedView(checked,target);C.view(checked);
 const changed=clone(target);changed.where.all[0].value='foreign service';assert.throws(()=>T.authoredResponse(expanded,source,changed,baseline));
 for(const poison of [s=>s.title+=' changed',s=>C.aggregate(s.composition).display.title='changed selected definition']){const selected=clone(source);poison(selected);assert.throws(()=>T.authoredResponse(expanded,selected,target,baseline))}
 const changedWindow=clone(target);changedWindow.toMs++;assert.throws(()=>T.authoredResponse(expanded,source,changedWindow,baseline));
 const swapped=clone(expanded);swapped.view.queries[0]=clone(source.queries[0]);assert.throws(()=>T.authoredResponse(swapped,source,target,baseline));
 if(C.comparison(source.composition)){assert.throws(()=>T.authoredRequest(source,target));assert.equal(C.comparison(checked.composition).baselineCapture,baseline);assert.notEqual(baseline,C.comparison(source.composition).baselineCapture)}
 if(C.timeline(source.composition)){const noInterval=clone(target);delete noInterval.bucketMs;assert.throws(()=>T.authoredRequest(source,noInterval,baseline));const wrongInterval=clone(target);wrongInterval.bucketMs=target.bucketMs===1000?10000:1000;assert.throws(()=>T.authoredRequest(source,wrongInterval,baseline))}
 let posts=0;assert.equal(await T.expandAuthored(source,target,{baselineCapture:baseline,current:()=>false,post:async()=>{posts++;return expanded}}),null);assert.equal(posts,0);
 let current=true,release;const pending=T.expandAuthored(source,target,{baselineCapture:baseline,current:()=>current,post:()=>{posts++;return new Promise(r=>release=r)}});current=false;release(expanded);assert.equal(await pending,null);assert.equal(posts,1);
 const mutable=clone(source),original=clone(mutable),response=await T.expandAuthored(mutable,target,{baselineCapture:baseline,post:async()=>{mutable.title+=' mutated caller';return expanded}});assert.deepEqual(response,checked);assert.notDeepEqual(mutable,original);
 response.queries[0].where.all[0].value='output edit';C.aggregate(response.composition).display.title='output edit';assert.deepEqual({source,target,widget,data},before);process.stdout.write(JSON.stringify({complete:true,refusals:20}));
}
