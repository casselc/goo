import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
import {actionVisible} from '../ui/dom-visibility.mjs';
class Element{
 constructor(tag,text){this.tagName=tag;this.children=[];this.attributes={};this.dataset={};this._text=text??'';this.hidden=false;this.disabled=false;this.inert=false;this.connected=false}
 get isConnected(){return this.connected&&(!this.parentElement||this.parentElement.isConnected)}
 get textContent(){return String(this._text)+this.children.map(x=>x.textContent).join('')}
 set textContent(value){this._text=value??''}
 append(...nodes){for(const child of nodes){child.parentElement=this;child.connected=true;this.children.push(child)}}
 replaceChildren(...nodes){for(const child of this.children){child.parentElement=null;child.connected=false}this.children=[];this._text='';this.append(...nodes)}
 setAttribute(key,value){this.attributes[key]=String(value)}getAttribute(key){return this.attributes[key]??null}
 closest(selector){for(let e=this;e;e=e.parentElement){if(selector==='[inert]'&&e.inert)return e;if(selector==='[hidden], [aria-hidden="true"]'&&(e.hidden||e.getAttribute('aria-hidden')==='true'))return e}return null}
 getClientRects(){return this.hidden?[]:[{}]}
}
const node=(tag,text)=>new Element(tag,text);
globalThis.document={hidden:false};
mock.module('../ui/components.mjs',{exports:{node,actionVisible}});
mock.module('../ui/attributes.mjs',{exports:{streamInspection:()=>node('div','Exact stream metadata')}});
const {render}=await import('../ui/metric-discovery.mjs');
const find=(host,tag)=>host.children.flatMap(child=>[...(child.tagName===tag?[child]:[]),...find(child,tag)]);
const capability={version:'goo/instrument-capability-plan/v1alpha1',profile:'generic',title:'Generic binding',eligible:true,verification:'not-run',reasons:[],hints:[],coverage:{state:'unavailable',streams:[],dimensions:[]}};
function instrument(overrides={}){return {kind:'gauge',metric:'load',unit:'1',scope:'app',version:'',schemaURL:'',records:7,temporality:0,supported:true,definition:{kind:'gauge'},capabilityPlan:structuredClone(capability),...overrides}}
function setup(instruments,diagnostics=[]){const host=node('section');host.connected=true;const built=[];render(host,{instruments,diagnostics,sources:diagnostics.map(d=>({kind:d.kind,state:d.reason}))},{onBuild:i=>built.push(i)});return{host,built}}
test('observed exponential points stay readable but never executable, including direct callback',()=>{
 const source=instrument({kind:'exponential-histogram',metric:'response.duration',unit:'s',records:13,supported:false,definition:undefined,unsupportedReason:'Zero-threshold and optional-statistic presence are unavailable.'});
 const {host,built}=setup([source]);assert.match(host.textContent,/Exponential histogram · s · 13 exported points/);assert.match(host.textContent,/Preview unavailable/);assert.match(host.textContent,/Zero-threshold/);assert.doesNotMatch(host.textContent,/13 observations/);
 const table=find(host,'goo-table')[0];assert.equal(table.getAttribute('mobile-layout'),'stacked');const heads=find(table,'th').map(x=>x.textContent),row=find(table,'tbody')[0].children[0];assert.deepEqual(row.children.map(x=>x.dataset.label),heads);assert.match(row.children[2].textContent,/Preview unavailable · Zero-threshold/);
 const use=find(host,'button').find(x=>x.textContent==='Build measurement package');assert.equal(use.disabled,true);use.onclick();use.disabled=false;use.onclick();assert.equal(built.length,0);
 // Even a forged eligible/supported flag cannot turn this kind into an explicit histogram package.
 source.supported=true;source.definition={kind:'histogram'};use.onclick();assert.equal(built.length,0);
});
test('available ordinary instrument builds only through current visible row; local paging retires old callback',()=>{
 const instruments=Array.from({length:26},(_,i)=>instrument({metric:'load-'+i})),{host,built}=setup(instruments),use=find(host,'button').find(x=>x.textContent==='Build measurement package'),old=use.onclick;
 use.onclick();assert.equal(built[0],instruments[0]);use.hidden=true;old();use.hidden=false;use.inert=true;old();use.inert=false;use.disabled=true;old();use.disabled=false;assert.equal(built.length,1);
 find(host,'button').find(x=>x.textContent==='More instruments').onclick();assert.equal(built.length,1);old();assert.equal(built.length,1);find(host,'button').find(x=>x.textContent==='Previous instruments').onclick();host.append(use);old();assert.equal(built.length,1);
});
test('missing optional table is visible and an empty partial catalogue does not claim no instruments everywhere',()=>{
 const {host}=setup([],[{kind:'exponential-histogram',reason:'absent'}]);assert.equal(host.children[0].children[0].textContent,'Metadata and support · available sources only');assert.match(host.children[0].textContent,/Exponential histograms \(table absent\)/);assert.match(host.textContent,/Exponential histograms \(table absent\)/);assert.match(host.textContent,/Available sources only/);assert.match(host.textContent,/available sources for this window/);assert.doesNotMatch(host.textContent,/No instruments observed in this window/);
 const ordinary=setup([instrument()],[{kind:'exponential-histogram',reason:'absent'}]);assert.equal(find(ordinary.host,'button').find(x=>x.textContent==='Build measurement package').disabled,false);
});
test('catalog readiness refuses missing or ineligible supplied plans and keeps full caveats disclosed',()=>{
 for(const patch of [{capabilityPlan:undefined},{capabilityPlan:{...capability,eligible:false,reasons:['Current service evaluation policy requires a reported instance.']}},{supported:false},{definition:null},{kind:'exponential-histogram'}]){
  const i=instrument(patch),{host,built}=setup([i]),use=find(host,'button').find(x=>x.textContent==='Build measurement package');assert(use);assert.equal(use.disabled,true);use.disabled=false;use.onclick();assert.equal(built.length,0);assert.match(find(host,'tbody')[0].textContent,/Preview unavailable/);
 }
 const host=node('section');host.connected=true;render(host,{instruments:[instrument()],warning:'Full discovery caveat.',sources:[]},{onBuild(){}});const disclosure=host.children[0];assert.equal(disclosure.tagName,'details');assert.equal(disclosure.children[0].textContent,'Metadata and support');assert.match(disclosure.textContent,/Full discovery caveat/);assert.equal(disclosure.open,undefined);
});

test('raw gauge read remains independent of missing-instance evaluation policy and retires on local paging',()=>{
 const source=instrument({capabilityPlan:{...capability,eligible:false,reasons:['Reported instance required for evaluation.']}}),host=node('section');host.connected=true;const read=[];render(host,{instruments:[source,...Array.from({length:25},(_,i)=>instrument({metric:'other-'+i}))],sources:[]},{onBuild(){throw Error('not evaluable')},onReadObservations:(i,button)=>read.push({i,button})});
 const build=find(host,'button').find(x=>x.textContent==='Build measurement package'),raw=find(host,'button').find(x=>x.textContent==='Read observations');assert.equal(build.disabled,true);assert.equal(raw.disabled,false);assert.match(raw.getAttribute('aria-label'),/load/);raw.onclick();assert.equal(read[0].i,source);const old=raw.onclick;find(host,'button').find(x=>x.textContent==='More instruments').onclick();old();assert.equal(read.length,1);
});
test('441 discovered sums expose exact reporting variants visibly and independently of evaluator readiness',()=>{const profiles=[[1,true],[1,false],[2,true],[2,false],[0,false],[-7,true]],sources=profiles.map(([temporality,monotonic])=>instrument({kind:temporality===2&&!monotonic?'stock':'counter',metric:'jobs.total',temporality,monotonic,supported:false,definition:undefined,capabilityPlan:{...structuredClone(capability),eligible:false,reasons:['Instance identity is missing.']}})),host=node('section'),read=[];host.connected=true;render(host,{instruments:sources,diagnostics:[],sources:[]},{onBuild:()=>assert.fail('raw read cannot build'),onReadObservations:(i,opener)=>read.push({i,opener})});const rows=find(host,'tbody')[0].children;for(let j=0;j<profiles.length;j++){const [tempo,mono]=profiles[j],row=rows[j],label=tempo===1?'Delta':tempo===2?'Cumulative':tempo===0?'Unspecified (0)':'Unrecognized ('+tempo+')',button=find(row,'button').find(b=>b.textContent==='Read observations');assert.equal(row.dataset.temporality,String(tempo));assert.equal(row.dataset.monotonic,String(mono));assert.match(row.children[1].textContent,new RegExp(label.replace(/[()]/g,'\\$&')+' · '+(mono?'Monotonic':'Nonmonotonic')));assert.ok(button.getAttribute('aria-label').includes(label));assert.equal(button.disabled,false);button.onclick();assert.equal(read.at(-1).i,sources[j]);assert.doesNotMatch(row.textContent,/temporality not applicable/);}assert.equal(read.length,profiles.length);});
test('447 explicit histogram raw read remains available independently of evaluator readiness and keeps exact temporality',()=>{
 const i=instrument({kind:'histogram',metric:'work.payload',unit:'{item}',temporality:2,supported:false,definition:undefined,capabilityPlan:{...capability,eligible:false,reasons:['Unsupported evaluation unit.']}}),host=node('section'),reads=[];host.connected=true;render(host,{instruments:[i],sources:[]},{onBuild(){throw Error('must not build')},onReadObservations:value=>reads.push(value)});
 const buttons=find(host,'button'),read=buttons.find(b=>b.textContent==='Read observations'),build=buttons.find(b=>b.textContent==='Build measurement package');assert.equal(build.disabled,true);assert.ok(read);assert.equal(read.getAttribute('aria-label'),'Read observations for work.payload [{item}] · Cumulative');read.onclick();assert.deepEqual(reads,[i]);host.hidden=true;read.onclick();assert.equal(reads.length,1);host.hidden=false;
 const expo=node('section');expo.connected=true;render(expo,{instruments:[instrument({kind:'exponential-histogram',supported:false,definition:undefined})],sources:[]},{onBuild(){},onReadObservations(){throw Error('unsupported family read')}});assert.equal(find(expo,'button').some(b=>b.textContent==='Read observations'),false);
});
