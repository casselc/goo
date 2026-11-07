import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
import {node,Element} from './support/typed-distribution-372.mjs';
Element.prototype.remove=function(){if(this.parentElement)this.parentElement.children=this.parentElement.children.filter(e=>e!==this);this.parentElement=null;this.isConnected=false};
const apis=[];
mock.module('../ui/exploration.mjs',{exports:{mount:view=>{const api={suspensions:0,runs:0,focuses:0,openQuestion(q,_fields,options){this.question=structuredClone(q);this.options=structuredClone(options);view.dataset.questionPhase='staged'},suspend(){this.suspensions++},dispose(){},resume(){},focusRun(){this.focuses++}};apis.push(api);return api}}});
const {mount}=await import('../ui/investigation-trail.mjs');
test('typed distribution entry validates before visit replacement and stages complete questions; Back is source-free',()=>{
 const listeners=new Map();globalThis.addEventListener=(name,fn)=>listeners.set(name,fn);globalThis.dispatchEvent=()=>{};globalThis.scrollY=0;globalThis.scrollTo=()=>{};globalThis.location={hash:'#explore'};globalThis.history={replaceState(){}};globalThis.CustomEvent=class{};document.addEventListener=()=>{};
 const host=node('main');document.body.append(host);mount(host,{lazy:true});const enter=listeners.get('goo:open-population');
 const q={signal:'spans',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'empty'}},fromMs:1000,toMs:5000,limit:19,offset:0,numeric:{field:{kind:'attribute',key:'payload.size'},type:'int64',parse:'canonical-decimal',unit:'items'},numericHistogram:{binning:'powers-of-two'},where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'worker'},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''},{numeric:{field:{kind:'attribute',key:'payload.size'},type:'int64',parse:'canonical-decimal',unit:'items'},op:'gte',value:'18014398509481984'}]}};
 const before=structuredClone(q);enter({detail:{question:q,stage:true}});assert.equal(apis.length,1);assert.deepEqual(apis[0].question,before);assert.equal(apis[0].options.stage,true);
 for(const bad of [{...q,environmentSelector:{...q.environmentSelector,policy:'unknown'}},{...q,toMs:999},{...q,unexpected:true}])enter({detail:{question:bad,stage:true}});
 assert.equal(apis.length,1);assert.equal(apis[0].suspensions,0,'invalid admission cannot suspend the active parent');
 const child=structuredClone(q);child.where.all.push({numeric:child.numeric,op:'lt',value:'36028797018963968'});enter({detail:{question:child,stage:true}});assert.equal(apis.length,2);assert.equal(apis[0].suspensions,1);assert.deepEqual(apis[1].question,child);assert.equal(apis[1].options.stage,true);
 q.where.all[0].value='mutated caller';assert.deepEqual(apis[0].question,before);
 host.children[0].children[0].onclick();assert.equal(apis[0].focuses,1);assert.equal(apis[1].suspensions,1);assert.equal(apis.reduce((n,a)=>n+a.runs,0),0);
});
