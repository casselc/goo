import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as Distributions from '../ui/distributions.mjs';
import * as D from '../ui/view-declaration.mjs';
import * as QuestionSession from '../ui/question-session.mjs';
import {stagedView} from '../ui/view-template-handoff.mjs';
import {createSourceAdapter} from '../ui/snapshots.mjs';
import {measureValue} from '../ui/view-model.mjs';
const vectors=JSON.parse(readFileSync(new URL('./fixtures/quantitative-view-397-wire.json',import.meta.url)));
const measures=['numeric.mean','numeric.sum','numeric.min','numeric.max','duration.mean','duration.min','duration.max','duration.p50','duration.p95','duration.p99'];
function data(signal='spans',extra={}){
 const question=structuredClone(vectors[0].question);question.signal=signal;if(signal==='logs')delete question.summary;question.where={all:[{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''},{any:[{field:{kind:'attribute',key:'tenant'},op:'eq',value:"a'雪"},{field:{kind:'attribute',key:'tenant'},op:'missing'}]}]};question.environmentSelector={version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'unreported'}};delete question.environment;Object.assign(question,extra);
 const fields=[{key:'category',type:'string',label:'Class',nullable:true},{key:'count',type:'number',unit:'records',label:'Captured records'}],rows=[{key:'zero',values:{category:'',count:'2'},cells:{}},{key:'unavailable',values:{category:null,count:'3'},cells:{}}];
 for(const key of measures.filter(key=>signal==='spans'||key.startsWith('numeric.'))){const [family,kind]=key.split('.'),quantile={p50:'0.5',p95:'0.95',p99:'0.99'}[kind];fields.push({key,type:'rational',label:key,unit:family==='duration'?'ns':'bytes',nullable:true,aggregation:{family,kind:quantile?'quantile':kind,method:quantile?'quantilesExactHigh':kind==='mean'?'exact-rational':'exact',...(quantile?{quantile}:{})}});for(const [index,row] of rows.entries()){row.values[key]=index===0?'0':null;row.cells[key]={state:index===0?'derived':'no_valid_values',support:index===0?{total:'2',valid:'2',missing:'0',empty:'0',invalid:'0'}:{total:'3',valid:'0',missing:'1',empty:'1',invalid:'1'}};}}
 const raw={version:'goo.dataset/1',id:'quant397',revision:'basis397',origin:'source-read',scope:{id:'q397',generation:'basis397',signal,selector:question},shape:'ranked',dimensions:['category'],fields,rows,quality:{extent:'complete',basis:'captured-query',value:'recorded',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'basis397',scopeId:'q397',aspects:['groups']},partition:{kind:'disjoint-record-groups',measure:'count',total:'5'}};
 return {question,basis:'basis397',state:'observed',count:'5',viewDataset:createSourceAdapter({id:'quant397'}).adoptSnapshot(raw,{question})};
}
test('397 shared declaration wire vectors admit finite quantitative Bar/Table contracts',()=>{for(const v of vectors){if(v.accepted)assert.doesNotThrow(()=>D.validate(v.declaration,v.question),v.name);else assert.throws(()=>D.validate(v.declaration,v.question),undefined,v.name)}});
test('397 exact quantitative handoff preserves bindings, predicates, paging, selected measure and unavailable support',()=>{
 for(const signal of ['logs','spans'])for(const value of measures.filter(key=>signal==='spans'||key.startsWith('numeric.')))for(const kind of ['bar','table']){
  const origin=data(signal),before=structuredClone(origin),presentation={kind,value,scale:'max',order:'label',query:'雪',requiresComplete:true},view=Distributions.distributionCustomization(origin,presentation),declaration=view.displays[signal];
  assert.deepEqual(view.queries,[origin.question]);assert.equal(declaration.display.value,value);assert.equal(declaration.kind.name,kind);assert.equal(declaration.display.query,'雪');assert.equal(declaration.display.requiresComplete,true);assert.deepEqual(stagedView(view,origin.question),view);assert.equal(Object.hasOwn(view,'basis'),false);assert.equal(Object.hasOwn(view,'viewDataset'),false);
  const destination=structuredClone(origin.question);destination.fromMs+=10000;destination.toMs+=10000;destination.limit=3;destination.where={field:{kind:'column',key:'service'},op:'eq',value:'other'};assert.deepEqual(D.reuse(declaration,origin.question,destination),declaration);
  if(value.startsWith('numeric.'))for(const mutate of [q=>q.numeric.unit='requests',q=>q.numeric.field.kind='resource',q=>q.numeric.field.key='other',q=>q.numeric.parse='lenient',q=>q.numeric.type='uint64']){const bad=structuredClone(destination);mutate(bad);assert.throws(()=>D.reuse(declaration,origin.question,bad));}
  const zero=measureValue(origin.viewDataset,origin.viewDataset.rows[0],value),absent=measureValue(origin.viewDataset,origin.viewDataset.rows[1],value);assert.equal(zero.exact,'0');assert.equal(absent.exact,null);assert.deepEqual(absent.support,{total:'3',valid:'0',missing:'1',empty:'1',invalid:'1'});assert.deepEqual(origin,before);
 }
});
const source=readFileSync(new URL('../ui/distributions.mjs',import.meta.url),'utf8'),rootSource=readFileSync(new URL('../ui/exploration.mjs',import.meta.url),'utf8');
function node(tag,text){return {localName:tag,textContent:text,disabled:false,hidden:false,inert:false,attributes:{},dataset:{},isConnected:true,append(...children){(this.children??=[]).push(...children)},matches(){return this.fieldset===true},getAttribute(key){return this.attributes[key]??null},closest(){return this.inert?this:null}}}
test('397 actual producer/root wrappers use current quantitative presentation and refuse retired controls',()=>{
 for(const signal of ['logs','spans']){
  const origin=data(signal),events=[],applied={basis:origin.basis,question:structuredClone(origin.question)},presentation={kind:'table',value:signal==='logs'?'numeric.sum':'duration.p95',scale:'max',query:'雪',order:'label',requiresComplete:true};
  const env={readOnly:false,customizeOpener:null,current:origin,lastResult:origin,distributionView:{presentation:()=>presentation},distributionCustomization:Distributions.distributionCustomization,Distributions,QuestionSession,stagedView,node,session:{snapshot:()=>({phase:'applied',applied})},plotReady:()=>env.ready,ready:true,container:{isConnected:true,classList:{contains:()=>false}},UI:{actionVisible:element=>!element.hidden&&element.isConnected},structuredClone,CustomEvent:class{constructor(type,options){this.type=type;this.detail=options.detail}},global:{dispatchEvent:event=>events.push(event)}};
  vm.createContext(env);const guard=rootSource.split('\n').find(line=>line.trimStart().startsWith('function canCustomizeDistribution(')),options=rootSource.split('\n').find(line=>line.includes('onCustomizeDistribution:(view,origin,opener)=>')),start=options.indexOf('onCustomizeDistribution:')+'onCustomizeDistribution:'.length,callback=options.slice(start,options.indexOf(',stagePopulation:',start));env.onCustomizeDistribution=vm.runInContext(guard+'\n('+callback+')',env);
  vm.runInContext(['refreshCustomization','mountCustomization'].map(name=>source.split('\n').find(line=>line.trimStart().startsWith('function '+name+'('))).join('\n'),env);
  const host=node('div');env.mountCustomization(host,origin);const button=host.children[0];assert.equal(button.onclick(),true);assert.deepEqual(events[0].detail.question,origin.question);assert.equal(events[0].detail.view.displays[signal].display.value,presentation.value);
  env.ready=false;assert.equal(button.onclick(),false);env.ready=true;for(let i=0;i<3;i++)assert.equal(button.onclick(),true);
  for(const patch of [{hidden:true},{inert:true},{disabled:true},{fieldset:true}]){Object.assign(button,patch);assert.equal(button.onclick(),false);for(const key of Object.keys(patch))button[key]=false;}
  presentation.query='雪'.repeat(86);env.refreshCustomization(origin);assert.equal(button.disabled,true);assert.equal(button.onclick(),false);presentation.query='fixed';env.refreshCustomization(origin);assert.equal(button.disabled,false);assert.equal(button.onclick(),true);
  const replacement=node('div');env.mountCustomization(replacement,origin);assert.equal(button.onclick(),false);const current=replacement.children[0];env.current=data(signal);assert.equal(current.onclick(),false);env.current=origin;applied.basis='superseded';assert.equal(current.onclick(),false);applied.basis=origin.basis;env.readOnly=true;assert.equal(current.onclick(),false);assert.equal(events.length,5);
 }
});

test('397 quantitative continuation refuses unsupported profiles and missing selected field without fallback',()=>{
 for(const extra of [{offset:2},{spanStatus:true},{reportedEdge:{}},{numericHistogram:{binning:'powers-of-two'}},{groupTimeSeries:true}])assert.throws(()=>Distributions.distributionCustomization(data('spans',extra),{kind:'table',value:'numeric.mean',scale:'max'}));
 assert.throws(()=>Distributions.distributionCustomization(data('logs',{summary:'span-duration'}),{kind:'bar',value:'duration.mean',scale:'max'}));
 const origin=data('logs'),raw=structuredClone(origin.viewDataset);raw.fields=raw.fields.filter(field=>field.key!=='numeric.mean');for(const row of raw.rows){delete row.values['numeric.mean'];delete row.cells['numeric.mean']}
 origin.viewDataset=createSourceAdapter({id:'missing397'}).adoptSnapshot(raw,{question:origin.question});assert.throws(()=>Distributions.distributionCustomization(origin,{kind:'bar',value:'numeric.mean',scale:'max'}));
 for(const value of ['numeric.mean','duration.mean'])assert.throws(()=>Distributions.distributionCustomization(data(),{kind:'bar',value,scale:'population'}));
});
