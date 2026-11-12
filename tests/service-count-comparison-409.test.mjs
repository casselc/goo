import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as Comparison from '../ui/service-count-comparison.mjs';
import * as Composition from '../ui/view-composition.mjs';
import * as CapturedCount from '../ui/captured-count-comparison.mjs';
import {fixture,observed,tick,walk} from './support/dashboard-owner-354.mjs';

const clone=structuredClone;
const scope={environment:'prod',fromMs:120000,toMs:180000};
function question(signal='spans'){
 const q={signal,environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},fromMs:120000,toMs:180000,limit:3,offset:0,bucketMs:1000,groupBy:[{kind:'attribute',key:'operation'}],where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''},{numeric:{field:{kind:'attribute',key:'work.items'},type:'int64',unit:'items',parse:'canonical-decimal'},op:'gte',value:'2'}]}};
 if(signal==='spans'){q.summary='span-duration';q.spanStatus=true;}
 return q;
}
function button(f,key='logs0'){return walk(f.box(key)).find(element=>element.className==='dashboard-compare-records');}

test('409 count projection preserves full population filters and exact previous-window recipe',()=>{
 for(const signal of ['logs','spans']){
  const q=question(signal),before=clone(q),data={question:clone(q),state:'observed'},view=Comparison.create({query:q},data,{service:'api',scope:{...scope,environment:'',environmentSelector:clone(q.environmentSelector),namespace:{mode:'value',value:''}}});
  const expected=clone(q);delete expected.summary;delete expected.spanStatus;
  assert.deepEqual(view.queries,[expected]);assert.deepEqual(q,before);assert.deepEqual(data,{question:before,state:'observed'});
  Composition.view(view);const declaration=Composition.comparison(view.composition);
  assert(CapturedCount.queryVariant(declaration));assert.equal(declaration.display.value,'count');assert.deepEqual(declaration.baselineQuery,{...expected,fromMs:60000,toMs:120000});
  assert.deepEqual(Composition.aggregate(view.composition).display.value,'count');
  for(const mutate of [x=>x.offset=1,x=>x.limit=101,x=>x.numeric=clone(x.where.all[2].numeric),x=>{x.numeric=clone(x.where.all[2].numeric);x.numericHistogram={binning:'powers-of-two'}},x=>x.reportedEdge={},x=>x.groupTimeSeries=true,x=>x.groupBy=[],x=>x.groupBy.push({kind:'column',key:'service'}),x=>{x.fromMs=50;x.toMs=150}]){const bad=clone(q);mutate(bad);assert.throws(()=>Comparison.project(bad));}
  for(const mutate of [x=>x.question.limit++,x=>x.state='changed_between_reads']){const bad=clone(data);mutate(bad);assert.throws(()=>Comparison.create({query:q},bad));}
  assert.throws(()=>Comparison.create({measurement:{},query:q},data));
  for(const mutate of [x=>x.service='foreign',x=>x.scope.fromMs++,x=>x.scope.environment='other',x=>x.scope.namespace={mode:'missing'}]){const context={service:'api',scope:{...scope,environment:'',environmentSelector:clone(q.environmentSelector),namespace:{mode:'value',value:''}}};mutate(context);assert.throws(()=>Comparison.create({query:q},data,context));}
 }
});

test('409 actual dashboard action waits for all reads and stages without source IO',async()=>{
 const staged=[],f=fixture(2,{onCompare:(...args)=>staged.push(args)},scope);await tick();assert.equal(f.requests.length,2);
 f.requests[0].resolve(observed(f.requests[0]));await tick();const compare=button(f);assert(compare);assert.equal(compare.disabled,true);compare.onclick();assert.equal(staged.length,0);
 f.requests[1].resolve(observed(f.requests[1]));const owner=await f.loaded;assert.equal(compare.disabled,false);const before={reads:f.requests.length,plans:f.metadata.length};compare.onclick();assert.equal(staged.length,1);assert.equal(staged[0][1],compare);assert.deepEqual(staged[0][0].queries,[f.plan.widgets[0].query]);assert.deepEqual({reads:f.requests.length,plans:f.metadata.length},before);
 for(const mode of ['disabled','inert','hidden','aria-disabled']){if(mode==='aria-disabled')compare.setAttribute(mode,'true');else compare[mode]=true;compare.onclick();assert.equal(staged.length,1);if(mode==='aria-disabled')compare.setAttribute(mode,'false');else compare[mode]=false;}
 const current=owner.results[0].result;current.question.limit++;compare.onclick();assert.equal(staged.length,1);assert.equal(compare.disabled,true);owner.dispose();compare.disabled=false;compare.onclick();assert.equal(staged.length,1);
});

test('409 foreign and stale widget leases never acquire comparison authority',async()=>{
 const staged=[],foreign=fixture(1,{onCompare:view=>staged.push(view)},scope);foreign.plan.widgets[0].query.where.value='foreign';await tick();foreign.requests[0].resolve(observed(foreign.requests[0]));const foreignOwner=await foreign.loaded;assert.equal(button(foreign),undefined);foreignOwner.dispose();
 const f=fixture(1,{onCompare:view=>staged.push(view)},scope);await tick();f.requests[0].resolve(observed(f.requests[0]));const owner=await f.loaded,old=button(f),invoke=old.onclick;assert(old);owner.dispose();invoke();assert.equal(staged.length,0);assert.equal(f.requests.length,1);
});


test('409 parent handoff refuses both open and overview pending before preserving the service',()=>{
 const source=readFileSync(new URL('../ui/service-browser.mjs',import.meta.url),'utf8'),start=source.indexOf('const stageWidget=')+'const stageWidget='.length,end=source.indexOf(';\n     try{await ServiceDashboard.load',start);
 assert(end>start);
 const events=[],env={g:1,generation:1,version:2,revision:2,request:{signal:{aborted:false}},pending:{open:true,dashboard:false},Visuals:{actionVisible:()=>true},clone:structuredClone,returnPath:{show:value=>events.push(['return',value])},identity:{name:()=> 'api'},service:'api',overviewScope:{namespace:undefined},returnToService(){},notify:(...args)=>events.push(['notify',...args]),global:{dispatchEvent:event=>events.push(['dispatch',event.detail])},CustomEvent:class{constructor(type,{detail}){this.type=type;this.detail=detail}}};
 vm.createContext(env);vm.runInContext('var stageWidget='+source.slice(start,end),env);
 const view=Comparison.create({query:question('logs')},{question:question('logs'),state:'observed'}),origin={};
 env.stageWidget(view.queries[0],origin,{view},'compare');assert.equal(events.length,0);
 env.pending.open=false;env.pending.dashboard=true;env.stageWidget(view.queries[0],origin,{view},'compare');assert.equal(events.length,0);
 env.pending.dashboard=false;env.stageWidget(view.queries[0],origin,{view},'compare');assert.equal(events.length,3);assert.equal(events[0][1].backLabel,'Back to service');assert.equal(events[1][1],'compare');assert.equal(events[2][1].stage,true);assert.deepEqual(events[2][1].view,view);assert.deepEqual(events[2][1].question,view.queries[0]);
});
