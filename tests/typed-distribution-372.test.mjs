import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import {test} from 'node:test';
import assert from 'node:assert/strict';
import {node,mount} from './support/typed-distribution-372.mjs';
import {edges} from '../ui/numeric-bins.mjs';
import {createSourceAdapter} from '../ui/snapshots.mjs';
import {createActionContext} from '../ui/scope-actions.mjs';
import {histogramIntent} from '../ui/view-model.mjs';
import * as Protocol from '../ui/question-protocol.mjs';
const clone=structuredClone;
function fixture(){
 const numeric={field:{kind:'attribute',key:'payload.size'},type:'int64',unit:'items',parse:'canonical-decimal'},question={signal:'spans',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},fromMs:1000,toMs:5000,limit:19,offset:0,numeric,numericHistogram:{binning:'powers-of-two'},where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'worker'},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''}]}};
 const bounds=edges('int64'),raw={version:'goo.dataset/1',id:'typed372',revision:'r372',origin:'source-read',scope:{id:'q372',generation:'r372',signal:'spans',selector:question},shape:'histogram',histogram:{kind:'numeric',binning:'powers-of-two',binding:numeric,support:{total:'3',valid:'2',missing:'1',empty:'0',invalid:'0'},closure:'left-closed-right-open'},fields:[{key:'from',label:'From',type:'rational',unit:'items'},{key:'to',label:'To',type:'rational',unit:'items'},{key:'count',label:'Records',type:'number',unit:'records'}],rows:bounds.slice(0,-1).map((from,i)=>({key:'bin:'+i,values:{from,to:bounds[i+1],count:i===119?'2':'0'},state:i===119?'observed':'no_data'})),quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'r372',scopeId:'q372',aspects:['bins']}};
 return{question,count:'3',histogramDataset:createSourceAdapter({id:'typed372'}).adoptSnapshot(raw,{question})};
}
function setup(){const data=fixture(),host=node('section');document.body.append(host);const opened=[],pending=[],completedDisabled=[];let active=true;const fromBin=119,toBin=120,from=edges('int64')[fromBin],to=edges('int64')[toBin],intent=histogramIntent(data.histogramDataset,'bin:119');
 globalThis.fetch=(path,options)=>new Promise(resolve=>pending.push({path,options,resolve}));
 const context=createActionContext({epoch:()=>1,active:()=>active,canStage:()=>active}),binding=context.bind({role:'current',slot:'histogram',epoch:1,snapshot:data.histogramDataset,question:data.question,coherent:true});
 const owner=mount(host,data,{open:(...args)=>{completedDisabled.push(args[3]?.disabled);opened.push(args)},canAct:()=>active,resolveRange:intent=>context.histogram(binding,intent)});
 const section=host.children[0],chart=section.querySelector('goo-chart'),button=section.querySelectorAll('button')[0];
 const selected=clone(data.question);selected.where={all:[clone(data.question.where),{all:[{numeric:clone(data.question.numeric),op:'gte',value:from},{numeric:clone(data.question.numeric),op:'lt',value:to}]}]};
 const rest=clone(data.question);rest.where={all:[clone(data.question.where),{any:[...['missing','empty','invalid'].map(op=>({numeric:clone(data.question.numeric),op})),{numeric:clone(data.question.numeric),op:'lt',value:from},{numeric:clone(data.question.numeric),op:'gte',value:to}]}]};const project=q=>{const v=clone(q);delete v.numeric;delete v.numericHistogram;v.groupBy=[{kind:'column',key:'service'}];return v};const plan={binning:'powers-of-two',parent:data.question,selected,rest,fromBin,toBin,from,to,comparison:{selection:project(selected),baseline:project(rest)}};
 return{data,host,owner,section,chart,button,opened,pending,completedDisabled,intent,plan,active:v=>active=v};}
test('typed numeric distribution planning preserves exact bigint population question without reading telemetry',async()=>{
 const s=setup(),before=clone(s.data.question);s.chart.dispatchEvent({type:'goo-histogram-range',detail:s.intent});assert.equal(s.pending.length,0);const operation=s.button.onclick();assert.equal(s.pending.length,1);assert.equal(s.pending[0].path,'/api/numeric-histograms/selection');assert.deepEqual(JSON.parse(s.pending[0].options.body),{parent:before,fromBin:119,toBin:120});s.pending[0].resolve({ok:true,json:async()=>s.plan});await operation;
 assert.equal(s.opened.length,1);assert.deepEqual(s.completedDisabled,[true],'checked completion may hand off while its opener is pending-disabled');assert.deepEqual(s.opened[0][0],s.plan.selected);assert.deepEqual(s.data.question,before);assert.equal(s.plan.from,'18014398509481984');assert.equal(s.plan.to,'36028797018963968');assert.deepEqual(Protocol.readQuestionLink(Protocol.questionLink(s.opened[0][0])).question,s.plan.selected);
});
test('pending typed distribution continuation is retired by suspension, owner replacement and hidden surfaces',async()=>{
 for(const change of [s=>s.owner.suspend(),s=>s.owner.invalidate(),s=>s.active(false),s=>s.section.hidden=true]){const s=setup();s.chart.dispatchEvent({type:'goo-histogram-range',detail:s.intent});const operation=s.button.onclick();change(s);s.pending[0].resolve({ok:true,json:async()=>s.plan});await operation;assert.equal(s.opened.length,0)}
});
test('current numeric producer refuses changed thresholds and duplicate coverage even when projections agree',async()=>{
 for(const mutate of [
  p=>p.selected.where.all[1].all[0].value='18014398509481985',
  p=>p.selected.where.all[1].all[1].value='36028797018963967',
  p=>p.rest.where.all[1].any[3].value='18014398509481985',
  p=>p.rest.where.all[1].any[0]=clone(p.rest.where.all[1].any[2]),
 ]){const s=setup();s.chart.dispatchEvent({type:'goo-histogram-range',detail:s.intent});const operation=s.button.onclick(),changed=clone(s.plan);mutate(changed);changed.comparison.selection.where=clone(changed.selected.where);changed.comparison.baseline.where=clone(changed.rest.where);s.pending[0].resolve({ok:true,json:async()=>changed});await operation;assert.equal(s.opened.length,0)}
});
test('captured hidden, disabled and inert range controls do not plan, and hidden pending openers cannot stage',async()=>{
 for(const change of [s=>s.button.hidden=true,s=>s.button.disabled=true,s=>s.button.inert=true]){const s=setup();s.chart.dispatchEvent({type:'goo-histogram-range',detail:s.intent});const captured=s.button.onclick;change(s);await captured();assert.equal(s.pending.length,0);assert.equal(s.opened.length,0)}
 for(const change of [s=>s.button.hidden=true,s=>s.button.inert=true]){const s=setup();s.chart.dispatchEvent({type:'goo-histogram-range',detail:s.intent});const operation=s.button.onclick();change(s);s.pending[0].resolve({ok:true,json:async()=>s.plan});await operation;assert.equal(s.opened.length,0)}
});

test('production source retirement synchronously pauses the numeric controls and prevents late finally reactivation',async()=>{
 const source=readFileSync(new URL('../ui/distributions.mjs',import.meta.url),'utf8').split('\n').find(line=>line.trimStart().startsWith('function suspendActions(')).trim();
 for(const pending of [false,true]){
  const s=setup();s.chart.dispatchEvent({type:'goo-histogram-range',detail:s.intent});const operation=pending?s.button.onclick():null;
  const env={populationGeneration:1,distributionGeneration:1,numericOwner:s.owner,durationOwner:null,timelineView:null,distributionView:null};vm.createContext(env);vm.runInContext(source,env);env.suspendActions();
  assert.equal(env.populationGeneration,2);assert.equal(env.distributionGeneration,2);assert(s.section.querySelectorAll('button').every(button=>button.disabled));assert.match(s.section.textContent,/Selection paused/);
  if(pending){s.pending[0].resolve({ok:true,json:async()=>s.plan});await operation}
  assert.equal(s.opened.length,0);assert(s.section.querySelectorAll('button').every(button=>button.disabled));assert.match(s.section.textContent,/Selection paused/);
 }
});
