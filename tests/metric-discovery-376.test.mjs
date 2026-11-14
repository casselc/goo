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
 const {host,built}=setup([source]);assert.match(host.textContent,/Exponential histogram · s · 13 exported points/);assert.match(host.textContent,/Observed · unsupported for evaluation/);assert.match(host.textContent,/Zero-threshold/);assert.doesNotMatch(host.textContent,/13 observations/);
 const table=find(host,'goo-table')[0];assert.equal(table.getAttribute('mobile-layout'),'stacked');const heads=find(table,'th').map(x=>x.textContent),row=find(table,'tbody')[0].children[0];assert.deepEqual(row.children.map(x=>x.dataset.label),heads);assert.equal(row.children[2].textContent,'Observed · unsupported for evaluationSource and representation requirements unmet.');
 const use=find(host,'button').find(x=>x.textContent==='Unsupported instrument');assert.equal(use.disabled,true);use.onclick();use.disabled=false;use.onclick();assert.equal(built.length,0);
 // Even a forged eligible/supported flag cannot turn this kind into an explicit histogram package.
 source.supported=true;source.definition={kind:'histogram'};use.onclick();assert.equal(built.length,0);
});
test('available ordinary instrument builds only through current visible row; local paging retires old callback',()=>{
 const instruments=Array.from({length:26},(_,i)=>instrument({metric:'load-'+i})),{host,built}=setup(instruments),use=find(host,'button').find(x=>x.textContent==='Build measurement package'),old=use.onclick;
 use.onclick();assert.equal(built[0],instruments[0]);use.hidden=true;old();use.hidden=false;use.inert=true;old();use.inert=false;use.disabled=true;old();use.disabled=false;assert.equal(built.length,1);
 find(host,'button').find(x=>x.textContent==='More instruments').onclick();assert.equal(built.length,1);old();assert.equal(built.length,1);find(host,'button').find(x=>x.textContent==='Previous instruments').onclick();host.append(use);old();assert.equal(built.length,1);
});
test('missing optional table is visible and an empty partial catalogue does not claim no instruments everywhere',()=>{
 const {host}=setup([],[{kind:'exponential-histogram',reason:'absent'}]);assert.match(host.textContent,/Exponential histograms \(table absent\)/);assert.match(host.textContent,/Available sources only/);assert.match(host.textContent,/available sources for this window/);assert.doesNotMatch(host.textContent,/No instruments observed in this window/);
 const ordinary=setup([instrument()],[{kind:'exponential-histogram',reason:'absent'}]);assert.equal(find(ordinary.host,'button').find(x=>x.textContent==='Build measurement package').disabled,false);
});
