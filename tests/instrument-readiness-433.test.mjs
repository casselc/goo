import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as InstrumentCapabilities from '../ui/instrument-capabilities.mjs';
import {nativeControlReady,actionVisible} from '../ui/dom-visibility.mjs';
const capability={version:InstrumentCapabilities.version,eligible:true,verification:'not-run',reasons:[],hints:[]};
const instrument=overrides=>({kind:'gauge',supported:true,definition:{kind:'gauge'},metric:'cpu',capabilityPlan:structuredClone(capability),...overrides});
const blocked=[{supported:false},{definition:null},{kind:'exponential-histogram'},{capabilityPlan:undefined},{capabilityPlan:{...capability,eligible:false,reasons:['Instance identity unavailable.']}},{capabilityPlan:{...capability,version:'future'}},{capabilityPlan:{...capability,verification:'verified'}},{unsupportedReason:'Unsupported instance.'}];
test('shared readiness accepts only executable supplied plans without mutating observed metadata',()=>{
 const good=instrument(),before=structuredClone(good);assert.equal(InstrumentCapabilities.ready(good),true);assert.deepEqual(good,before);for(const patch of blocked)assert.equal(InstrumentCapabilities.ready(instrument(patch)),false);for(const kind of ['gauge','stock','counter','histogram'])assert.equal(InstrumentCapabilities.ready(instrument({kind})),true);
});
class Element{
 constructor(tag,text=''){this.tagName=tag;this.textContent=text;this.children=[];this.dataset={};this.isConnected=true;this.disabled=false;this.hidden=false;this.inert=false}
 append(...nodes){for(const child of nodes){child.parentElement=this;this.children.push(child)}}
 closest(selector){for(let p=this;p;p=p.parentElement){if(selector==='[inert]'&&p.inert)return p;if(selector==='[hidden], [aria-hidden="true"]'&&p.hidden)return p}return null}
 getClientRects(){return this.hidden?[]:[{}]}getAttribute(){return null}matches(selector){return selector===':disabled'&&this.disabled}
}
const node=(tag,text)=>new Element(tag,text);
globalThis.document={hidden:false};
test('actual Measurements Build callback independently refuses missing, unsupported, ineligible, stale and hidden metadata',()=>{
 const source=readFileSync(new URL('../ui/measurements.mjs',import.meta.url),'utf8'),start=source.indexOf('onBuild:i=>')+'onBuild:'.length,end=source.indexOf('}})}catch',start),callback=source.slice(start,end+1);assert(end>start);
 for(const patch of [{},...blocked]){
  const i=instrument(patch),opened=[],nodes={measureEditor:{}},state={i,g:1,discoveryGeneration:1,ownerVisible:()=>true,result:{instruments:[i]},InstrumentCapabilities,Visuals:{actionVisible},host:node('section'),el:id=>nodes[id],open:p=>opened.push(p),discoveryBounds:{fromMs:1,toMs:2},JSON};
  const invoke=()=>vm.runInNewContext('('+callback+')(i)',state);invoke();assert.equal(opened.length,InstrumentCapabilities.ready(i)?1:0);
  if(!InstrumentCapabilities.ready(i))continue;
  for(const edit of [()=>state.discoveryGeneration++,()=>state.ownerVisible=()=>false,()=>state.host.hidden=true,()=>state.host.inert=true,()=>state.result.instruments=[],()=>i.capabilityPlan.eligible=false]){state.discoveryGeneration=1;state.ownerVisible=()=>true;state.host.hidden=false;state.host.inert=false;state.result.instruments=[i];i.capabilityPlan.eligible=true;const count=opened.length;edit();invoke();assert.equal(opened.length,count)}
 }
});
test('actual Services Preview rendering and callbacks use the same supplied policy and native control ownership',()=>{
 const source=readFileSync(new URL('../ui/service-browser.mjs',import.meta.url),'utf8'),begin=source.indexOf('    for(const i of result.instruments){'),end=source.indexOf('\n   }else error(sections[2]',begin),loop=source.slice(begin,end),button=source.slice(source.indexOf('  function button('),source.indexOf('\n',source.indexOf('  function button(')));
 assert(begin>=0&&end>begin);
 for(const patch of [{},...blocked]){
  const i=instrument(patch),events=[],sections=[node('section'),node('section'),node('section')],state={result:{instruments:[i]},sections,node,InstrumentCapabilities,nativeControlReady,Visuals:{actionVisible},generation:1,g:1,identity:{name:()=> 'worker'},service:'worker',activeScope:{fromMs:1,toMs:2},returnPath:{show:x=>events.push(x)},notify:()=>{},returnToService:()=>{},global:{dispatchEvent:event=>events.push(event)},CustomEvent:class{constructor(type,options){this.type=type;this.detail=options.detail}}};
  globalThis.document={hidden:false};state.document=document;vm.runInNewContext(button+'\n'+loop,state);const card=sections[2].children[0],use=card.children.find(c=>c.tagName==='button'),eligible=InstrumentCapabilities.ready(i);assert.equal(use.disabled,!eligible);assert.equal(use.textContent,'Preview measurement');use.onclick();assert.equal(events.length,eligible?2:0);
  use.disabled=false;if(!eligible){use.onclick();assert.equal(events.length,0);continue}
  for(const edit of [()=>use.hidden=true,()=>{use.hidden=false;use.inert=true},()=>{use.inert=false;state.generation++}]){use.hidden=false;use.inert=false;state.generation=1;const count=events.length;edit();use.onclick();assert.equal(events.length,count)}
 }
});

test('Services global metadata keeps a visible partial-source cue and disclosed source exclusions',async()=>{
 const {sourceNote}=await import('../ui/source-capabilities.mjs'),source=readFileSync(new URL('../ui/service-browser.mjs',import.meta.url),'utf8'),begin=source.indexOf("const metadata=node('details');"),end=source.indexOf('if(!result.instruments.length)',begin),statement=source.slice(begin,end);assert(begin>=0&&end>begin);
 for(const partial of [false,true]){
  const result={sources:partial?[{kind:'exponential-histogram',state:'absent'}]:[],records:7,instruments:[instrument()],warning:'Coverage unknown.'},sections=[node('section'),node('section'),node('section')];vm.runInNewContext(statement,{result,sections,node,sourceNote});
  const disclosure=sections[2].children[1];assert.equal(disclosure.tagName,'details');assert.equal(disclosure.children[0].textContent,'Metadata and support'+(partial?' · available sources only':''));assert(disclosure.children.some(c=>c.textContent==='Coverage unknown.'));if(partial)assert(disclosure.children.some(c=>c.textContent===sourceNote(result.sources)));assert.equal(sections[2].children[0].textContent,'7 exported points · 1 instrument identities.');
 }
});
