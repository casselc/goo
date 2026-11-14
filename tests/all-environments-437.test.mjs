import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Environment from '../ui/record-environment.mjs';
import {validateRecordQuestion} from '../ui/record-intent.mjs';
import * as Protocol from '../ui/question-protocol.mjs';
import * as History from '../ui/scope-history.mjs';
import {scopeSummary} from '../ui/scope-picker.mjs';
import {RecordScopeDraft} from '../ui/record-scope-draft.mjs';
import {traceQuestion} from '../ui/record-results.mjs';
import {surroundingQuestion} from '../ui/discovery.mjs';
import {inspectionQuestion,transportIdentity} from '../ui/tail.mjs';
const all=()=>Environment.effective({mode:'all'});
const query=()=>({signal:'logs',environmentSelector:all(),fromMs:1000,toMs:61000,limit:100,offset:0,where:{any:[{field:{kind:'column',key:'body'},op:'contains',value:'canary'},{field:{kind:'attribute',key:'region'},op:'missing'}]}});

test('437 explicit All is selector-only, never a name or observed environment',()=>{
 assert.deepEqual(all(),{version:Environment.version,policy:'effective',scope:{mode:'all'}});
 assert.deepEqual(Environment.normalize({...all(),scope:{mode:'all',value:''}}),all());
 for(const scope of [{mode:'all',value:'prod'},{mode:'all',value:null}])assert.throws(()=>Environment.effective(scope));
 assert.throws(()=>Environment.normalize({...all(),policy:'current-only'}));assert.throws(()=>Environment.normalizeScope({mode:'all'}));
 assert.equal(Environment.label(all()),'All environments');assert.notEqual(Environment.label(Environment.current('all')),Environment.label(all()));
 assert.throws(()=>Environment.selection({environment:''}));assert.equal(Environment.sameQuestionEnvironment({environment:'all'},query()),false);
 for(const resource of [undefined,null,{},[],{'deployment.environment.name':'prod'},{'deployment.environment.name':''},{'deployment.environment':'all'},{bad:4}])assert.equal(Environment.matches(all(),resource),true);
});

test('437 observed facts preserve key presence/precedence while invalid facts cannot authorize narrowing',()=>{
 for(const [resource,scope,conflict]of [[{},{mode:'unreported'},false],[{'deployment.environment':''},{mode:'empty'},false],[{'deployment.environment':'legacy'},{mode:'named',value:'legacy'},false],[{'deployment.environment.name':'','deployment.environment':'prod'},{mode:'empty'},true],[{'deployment.environment.name':'all','deployment.environment':'other'},{mode:'named',value:'all'},true]]){
  const before=structuredClone(resource),fact=Environment.observe(resource),q=query(),narrow=Environment.observedQuestion(q,resource);
  assert.deepEqual(fact.scope,scope);assert.equal(fact.conflict,conflict);assert.deepEqual(narrow.environmentSelector,Environment.effective(scope));assert.deepEqual(narrow.where,q.where);assert.equal(narrow.fromMs,q.fromMs);assert.deepEqual(resource,before);assert.equal(Environment.isAll(narrow),false);
 }
 for(const resource of [undefined,null,[],new Map(),new Date(),{bad:1}])assert.throws(()=>Environment.observe(resource));
 const long={'deployment.environment.name':'x'.repeat(129)},fact=Environment.observe(long);assert.match(Environment.observedLabel(fact),/xxxx/);assert.throws(()=>Environment.observedQuestion(query(),long));
});

test('437 logs preserve complete All intent through fixed links, draft, local history and inspection',()=>{
 const q=query(),before=structuredClone(q);validateRecordQuestion(q);const step=Protocol.recordStep(q);assert.deepEqual(Protocol.readStepLink(Protocol.stepLink(step)),step);
 const draft=new RecordScopeDraft('');draft.restore(q);assert.deepEqual(draft.snapshot(),Environment.questionFields(q));assert.equal(draft.matches(q),true);draft.setName('all');assert.equal(draft.matches(q),false);
 const store=new Map(),storage={getItem:k=>store.get(k)??null,setItem:(k,v)=>store.set(k,v),removeItem:k=>store.delete(k)},journal=new History.Journal({storage:()=>storage,now:()=>1000,path:'/all'});journal.persistence(true);journal.remember({kind:'question',question:q,view:'Records'});journal.pin({kind:'question',question:q,view:'Records'});
 const restored=new History.Journal({storage:()=>storage,now:()=>1000,path:'/all'});assert.deepEqual(restored.list()[0].intent.question,q);assert.match(scopeSummary(restored.list()[0].intent).meta,/All environments/);
 const inspected=inspectionQuestion(q,'logs',{readWindow:{fromMs:1000,toMs:61000}});assert.deepEqual(inspected.environmentSelector,all());assert.equal(JSON.parse(transportIdentity(q)).environmentSelector.scope.mode,'all');assert.deepEqual(q,before);
});

test('437 unsupported execution profiles refuse All rather than creating a broader context',()=>{
 const q=query();assert.throws(()=>validateRecordQuestion({...q,signal:'spans'}));assert.throws(()=>transportIdentity({...q,signal:'spans'}));assert.throws(()=>inspectionQuestion(q,'spans',{readWindow:{fromMs:1000,toMs:61000}}));
 assert.throws(()=>traceQuestion(q,'1'.repeat(32)));assert.throws(()=>surroundingQuestion(q,{timestamp_ns:'2000000000',ServiceName:'app',ResourceAttributes:{}}));
 assert.throws(()=>Protocol.validateTraceSearchRequest({version:'goo/trace-search/v1',...Environment.questionFields(q),fromMs:1000,toMs:61000,conditions:[{field:{kind:'column',key:'service'},op:'eq',value:'app'}],limit:10}));
 assert.throws(()=>Protocol.measurementStep({definition:{environmentSelector:all()}}));
});

test('437 actual Apply callback refuses hidden, disabled and departed controls after local All staging',async t=>{
 const {readFile}=await import('node:fs/promises'),{nativeControlReady}=await import('../ui/dom-visibility.mjs');
 const source=await readFile(new URL('../ui/stream-entry.mjs',import.meta.url),'utf8');let reads=0,active=true;
 const saved=globalThis.document;globalThis.document={hidden:false};t.after(()=>{globalThis.document=saved});
 const apply={isConnected:true,disabled:false,hidden:false,parentElement:null,getClientRects:()=>[1],getAttribute:()=>null,matches(s){return s===':disabled'&&this.disabled},closest(s){return s.includes('[hidden]')&&this.hidden?this:null}};
 const binding=source.match(/apply\.onclick=\(\)=>\{[^\n]+?\};signal\.onchange=/)[0].replace(/;signal\.onchange=$/,';');
 const install=()=>new Function('apply','enabled','records','route','nativeControlReady','start',binding)(apply,active,true,'explore',nativeControlReady,()=>reads++);
 install();apply.onclick();assert.equal(reads,1);
 for(const state of ['hidden','disabled','detached','departed']){apply.hidden=state==='hidden';apply.disabled=state==='disabled';apply.isConnected=state!=='detached';active=state!=='departed';install();apply.onclick();assert.equal(reads,1,state)}
});

test('437 supplied environment pills keep unavailable, unreported and unselectable names distinct',async t=>{
 const {environment}=await import('../ui/record-presentation.mjs'),saved=globalThis.document;
 globalThis.document={createElement:tag=>({localName:tag,textContent:'',className:'',title:''})};t.after(()=>{globalThis.document=saved});
 assert.equal(environment({ResourceAttributes:{}}).textContent,'Unreported environment');
 for(const row of [{},{ResourceAttributes:null},{ResourceAttributes:[]},{ResourceAttributes:new Map()},{ResourceAttributes:new Date()},{ResourceAttributes:{bad:1}}])assert.equal(environment(row).textContent,'Environment unavailable');
 const conflict=environment({ResourceAttributes:{'deployment.environment.name':'','deployment.environment':'prod'}});assert.equal(conflict.textContent,'Empty environment label');assert.match(conflict.title,/conflict, current key wins/);
 const long=environment({ResourceAttributes:{'deployment.environment.name':'x'.repeat(129)}});assert.match(long.textContent,/xxxxx/);assert.notEqual(long.textContent,'Environment unavailable');
});
