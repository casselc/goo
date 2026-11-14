import {test,mock} from 'node:test';import assert from 'node:assert/strict';
class El{constructor(tag,text){this.tag=tag;this.textContent=text;this.children=[];this.dataset={questionPhase:'initial'};this.parent=null;this.hidden=false}get isConnected(){return this.root||!!this.parent?.isConnected}append(...nodes){for(const n of nodes){n.parent=this;this.children.push(n)}}replaceChildren(...nodes){this.children.forEach(n=>n.parent=null);this.children=[];this.append(...nodes)}setAttribute(){}querySelectorAll(){return []}contains(){return false}remove(){if(this.parent)this.parent.children=this.parent.children.filter(n=>n!==this);this.parent=null}}
const apis=[];let mountHook=null;mock.module('../ui/components.mjs',{exports:{node:(tag,text)=>new El(tag,text)}});mock.module('../ui/exploration.mjs',{exports:{mount:view=>{const api={suspensions:0,openQuestion(question,_fields,options={}){this.question=structuredClone(question);this.options=structuredClone(options);view.dataset.questionPhase='staged'},suspend(){this.suspensions++},dispose(){},resume(){},focusRun(){}};apis.push(api);mountHook?.(view,api);return api}}});
const {mount}=await import('../ui/investigation-trail.mjs');
const {questionLink}=await import('../ui/question-protocol.mjs');
test('consumed local question links stay local while ordinary question links enter the trail',()=>{
 const keys=['document','addEventListener','dispatchEvent','scrollY','location','history','CustomEvent'],previous=new Map(keys.map(k=>[k,globalThis[k]])),listeners=new Map(),pushes=[];
 Object.assign(globalThis,{document:{addEventListener(type,fn){listeners.set(type,fn)},activeElement:null},addEventListener(){},dispatchEvent(){},scrollY:0,location:{hash:'#explore',href:'http://goo.test/',origin:'http://goo.test',pathname:'/'},history:{pushState(...args){pushes.push(args)}},CustomEvent:class{}});
 try{
  const host=new El('main');host.root=true;mount(host,{lazy:true});const before=apis.length,q={signal:'logs',environment:'prod',fromMs:0,toMs:60000,limit:100,offset:0},hash=questionLink(q),link={href:'http://goo.test/'+hash},click=listeners.get('click');
  click({defaultPrevented:true,button:0,target:{closest(){return link}}});assert.equal(apis.length,before);assert.equal(pushes.length,0);
  let prevented=false;click({defaultPrevented:false,button:0,target:{closest(){return link}},preventDefault(){prevented=true}});assert(prevented);assert.equal(apis.length,before+1);assert.deepEqual(apis.at(-1).question,q);assert.equal(pushes.length,1);assert.equal(pushes[0][2],hash);
 }finally{for(const [key,value]of previous){if(value===undefined)delete globalThis[key];else globalThis[key]=value}}
});
test('per-visit lease rejects identical-question ABA and releases temporary external return chrome',()=>{const previous=new Map(['document','addEventListener','dispatchEvent','scrollY','location','CustomEvent'].map(k=>[k,globalThis[k]]));Object.assign(globalThis,{document:{addEventListener(){},activeElement:null},addEventListener(){},dispatchEvent(){},scrollY:0,location:{hash:'#explore'},CustomEvent:class{}});try{const host=new El('main');host.root=true;const trail=mount(host,{lazy:true}),q={signal:'logs',environment:'prod'};trail.openQuestion(q);assert.equal(host.children[0].hidden,true,'one question has no history navigation');trail.openQuestion(q,[],{retainedReturn:true});const visit=trail.visit();assert(visit.current());assert.equal(host.children[0].hidden,true);host.children[0].children[0].onclick();assert(visit.current());assert(visit.suspend());visit.releaseReturn();assert.equal(host.children[0].hidden,false);trail.resume();assert.equal(visit.current(),false);assert.equal(host.children[0].hidden,false);trail.openQuestion(q);assert.equal(visit.current(),false);assert.equal(visit.suspend(),false);assert.equal(apis.at(-1).suspensions,0)}finally{for(const [k,v]of previous){if(v===undefined)delete globalThis[k];else globalThis[k]=v}}});


test('template entry stages grouped and ungrouped whole checked roots and rejects specialized destinations before mounting',()=>{
 const keys=['document','addEventListener','dispatchEvent','scrollY','location','CustomEvent'],previous=new Map(keys.map(k=>[k,globalThis[k]])),listeners=new Map();
 Object.assign(globalThis,{document:{addEventListener(){},activeElement:null},addEventListener(type,fn){listeners.set(type,fn)},dispatchEvent(){},scrollY:0,location:{hash:'#explore'},CustomEvent:class{}});
 try{const host=new El('main');host.root=true;mount(host,{lazy:true});const before=apis.length,q={signal:'spans',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'empty'}},fromMs:1000,toMs:61000,groupBy:[{kind:'column',key:'name'}],summary:'span-duration',spanStatus:true,bucketMs:1000,limit:100,offset:0,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:'team'}]}},original=structuredClone(q),enter=listeners.get('goo:open-population');
 enter({detail:{question:q,templates:true}});assert.equal(apis.length,before+1);assert.deepEqual(apis.at(-1).question,original);assert.equal(apis.at(-1).options.stage,true);assert.equal(apis.at(-1).options.templates,true);q.where.all[0].value='mutated';assert.deepEqual(apis.at(-1).question,original);
 const ungrouped={...structuredClone(original),groupBy:[],limit:7};enter({detail:{question:ungrouped,templates:true}});assert.equal(apis.length,before+2);assert.deepEqual(apis.at(-1).question,ungrouped);assert.equal(apis.at(-1).options.stage,true);assert.equal(apis.at(-1).options.templates,true);
 for(const bad of [{...original,offset:1},{...original,groupTimeSeries:true},{...original,groupBy:[{kind:'column',key:'name'},{kind:'column',key:'service'}]},{...ungrouped,numeric:{field:{kind:'attribute',key:'value'},type:'int64',parse:'canonical-decimal',unit:'{item}'}}])enter({detail:{question:bad,templates:true}});assert.equal(apis.length,before+2,'invalid entry must not replace the active visit');
 }finally{for(const [key,value]of previous){if(value===undefined)delete globalThis[key];else globalThis[key]=value}}
});


test('construction events retain typed scope until the mounted API acknowledges an exact saved reference',()=>{
 const keys=['document','addEventListener','dispatchEvent','scrollY','location','CustomEvent'],previous=new Map(keys.map(key=>[key,globalThis[key]])),listeners=new Map();
 const emit=(type,detail)=>{for(const listener of listeners.get(type)||[])listener({type,detail})};
 Object.assign(globalThis,{document:{addEventListener(){},activeElement:null},addEventListener(type,listener){if(!listeners.has(type))listeners.set(type,[]);listeners.get(type).push(listener)},dispatchEvent(event){emit(event.type,event.detail)},scrollY:0,location:{hash:'#explore'},CustomEvent:class{constructor(type,{detail}={}){this.type=type;this.detail=detail}}});
 try{
  const host=new El('main');host.root=true;const trail=mount(host,{lazy:true}),question={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0},ref={kind:'saved-view',id:'a'.repeat(64),title:'Whole comparison'};let readyAPI;
  mountHook=(view,api)=>{
   readyAPI=api;api.currentSavedView=()=>api.saved||null;view.dataset.questionPhase='draft';
   assert.doesNotThrow(()=>emit('goo:question-state',{phase:'draft',action:'stage',requested:question}));
   assert.deepEqual(trail.currentConfiguration(),{kind:'question',question,view:'Records'});assert.equal(trail.recallState(),null);
   // An event alone does not acknowledge a saved definition before API readiness.
   api.saved=ref;assert.doesNotThrow(()=>emit('goo:scope-used',ref));
   assert.equal(trail.visits()[0].configuration.kind,'question');
  };
  assert.doesNotThrow(()=>trail.openQuestion(question));mountHook=null;
  assert.deepEqual(trail.currentConfiguration(),ref);
  emit('goo:scope-used',ref);assert.deepEqual(trail.visits()[0].configuration,ref);
  emit('goo:question-state',{phase:'applied',action:'publish',applied:{question}});
  assert.deepEqual(trail.visits()[0].configuration,ref,'an acknowledged composition never collapses to its selected question');
  const returned=trail.currentConfiguration();returned.title='Caller mutation';assert.equal(trail.currentConfiguration().title,ref.title);
  readyAPI.saved=null;emit('goo:question-state',{phase:'draft',action:'stage',requested:question});assert.equal(trail.currentConfiguration().kind,'question','edited unacknowledged scope is not falsely pinned as its former saved revision');
 }finally{mountHook=null;for(const [key,value]of previous){if(value===undefined)delete globalThis[key];else globalThis[key]=value}}
});
