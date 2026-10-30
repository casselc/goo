'use strict';
const {test}=require('node:test'),assert=require('node:assert/strict');
const {createActionContext}=require('../ui/scope-actions.mjs');
const {createSourceAdapter,importSnapshot}=require('../ui/snapshots.mjs');
const {histogramIntent}=require('../ui/view-model.mjs');

function fixture(type='decimal64'){
 const numeric={field:{kind:'attribute',key:'work.cost'},type,parse:'canonical-decimal',unit:'USD',...(type==='decimal64'?{scale:2}:{})};
 const question={signal:'spans',environment:'test',fromMs:1000,toMs:5000,limit:100,offset:0,numeric,numericHistogram:{binning:'powers-of-two'},where:{field:{kind:'resource',key:'team'},op:'eq',value:'core'}};
 const edges=[];if(type!=='uint64')for(let i=63;i>=0;i--)edges.push(-(2n**BigInt(i)));edges.push(0n);for(let i=0;i<=(type==='uint64'?64:63);i++)edges.push(2n**BigInt(i));
 const text=value=>{let n=value,d=type==='decimal64'?100n:1n,a=n<0n?-n:n,b=d;while(b){const r=a%b;a=b;b=r}n/=a;d/=a;return String(n)+(d===1n?'':'/'+d)};
 return {version:'goo.dataset/1',id:'histogram',revision:'same-content',origin:'source-read',scope:{id:'scope',generation:'same-content',signal:'spans',selector:question},shape:'histogram',histogram:{kind:'numeric',binning:'powers-of-two',binding:numeric,support:{total:'5',valid:'3',missing:'1',empty:'0',invalid:'1'},closure:'left-closed-right-open'},fields:[{key:'from',label:'From',type:'rational',unit:'USD'},{key:'to',label:'To',type:'rational',unit:'USD'},{key:'count',label:'Valid captured records',type:'number',unit:'records'}],rows:edges.slice(0,-1).map((n,i)=>({key:'bin:'+i,values:{from:text(n),to:text(edges[i+1]),count:i===edges.length-2?'3':'0'},state:i===edges.length-2?'observed':'no_data'})),quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'same-content',scopeId:'scope',aspects:['bins']}};
}
function host(raw=fixture()){
 let epoch=1,active=true,canStage=true;const source=createSourceAdapter({id:'histogram.actions'}),snapshot=source.adoptSnapshot(raw,{question:raw.scope.selector});
 const context=createActionContext({epoch:()=>epoch,active:()=>active,canStage:()=>canStage});
 const bind=(value=snapshot,extra={})=>context.bind({role:'current',slot:'histogram',epoch,snapshot:value,question:value.scope.selector,coherent:true,...extra});
 return {snapshot,context,bind,setEpoch:v=>epoch=v,setActive:v=>active=v,setCanStage:v=>canStage=v};
}
const denied=(r,code)=>{assert.equal(r.ok,false);assert.equal(r.code,code)};

test('histogram planner geometry preserves exact coefficient sentinels and captured question',()=>{
 for(const type of ['decimal64','uint64']){
  const h=host(fixture(type)),binding=h.bind(),last=h.snapshot.rows.at(-1),intent=histogramIntent(h.snapshot,last.key,last.key),r=h.context.histogram(binding,intent);
  assert.equal(r.ok,true);assert.equal(r.to,type==='uint64'?'18446744073709551616':'9223372036854775808');assert.equal(r.fromBin,h.snapshot.rows.length-1);assert.equal(r.toBin,h.snapshot.rows.length);assert.deepEqual(r.question,h.snapshot.scope.selector);assert(Object.isFrozen(r.question));
  assert.equal(r.question.where.value,'core');assert.equal(r.question.numeric.type,type);if(type==='decimal64'){assert.notEqual(r.to,last.values.to);assert.equal(r.question.numeric.scale,2)}
  denied(h.context.resolve(binding,intent),'intent-type');assert.throws(()=>h.context.project(binding,h.snapshot),e=>e.code==='projection-shape');
  const {fromKey,toKey,...identity}=intent;denied(h.context.inspect(binding,{...identity,type:'select',key:last.key}),'intent-type');
 }
});
test('range endpoints must be ordered exact supplied keys without changing source',()=>{
 const h=host(),binding=h.bind(),before=JSON.stringify(h.snapshot),intent=histogramIntent(h.snapshot,'bin:3','bin:5'),r=h.context.histogram(binding,intent);
 assert.equal(r.ok,true);assert.equal(r.fromBin,3);assert.equal(r.toBin,6);assert.equal(JSON.stringify(h.snapshot),before);
 denied(h.context.histogram(binding,{...intent,fromKey:'bin:999'}),'invalid-membership');denied(h.context.histogram(binding,{...intent,from:'forged'}),'intent-type');denied(h.context.histogram(binding,{...intent,revision:'other'}),'identity-mismatch');
 denied(h.context.histogram(binding,{...intent,fromKey:'bin:5',toKey:'bin:3'}),'invalid-membership');
});
test('captured histogram binding refuses reruns, foreign owners, held roles and changed reads',()=>{
 const h=host(),old=h.bind(),intent=histogramIntent(h.snapshot,'bin:0','bin:1');h.bind();denied(h.context.histogram(old,intent),'stale-binding');
 const current=h.bind();h.setEpoch(2);denied(h.context.histogram(current,intent),'stale-epoch');h.setEpoch(1);
 const foreign=host();denied(foreign.context.histogram(current,intent),'foreign-binding');
 denied(h.context.histogram(h.bind(h.snapshot,{role:'baseline'}),intent),'read-only-role');
 denied(h.context.histogram(h.bind(importSnapshot(h.snapshot)),intent),'source-required');
 denied(h.context.histogram(h.bind(h.snapshot,{coherent:false}),intent),'incoherent');
});
test('planning rechecks the same binding after asynchronous work and never accepts a latest binding',async()=>{
 const h=host(),binding=h.bind(),intent=histogramIntent(h.snapshot,'bin:0','bin:1');assert.equal(h.context.histogram(binding,intent).ok,true);
 await Promise.resolve();h.context.revoke();h.bind();denied(h.context.histogram(binding,intent),'stale-binding');
 const current=h.bind();h.setActive(false);denied(h.context.histogram(current,intent),'inactive');h.setActive(true);h.setCanStage(false);denied(h.context.histogram(current,intent),'not-applied');
});
