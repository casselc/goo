import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as Composition from '../ui/view-composition.mjs';
import * as CELProvenance from '../ui/cel-provenance.mjs';

// Execute the real owned recall/load functions. Configuration GET is observed;
// this fixture supplies no telemetry read, publication or execution adapter.
const source=readFileSync(new URL('../ui/population-workspace.mjs',import.meta.url),'utf8');
function declaration(prefix){
 const start=source.indexOf(prefix);assert(start>=0);const brace=source.indexOf('){',start)+1;
 let depth=0,quote=null,escaped=false;
 for(let i=brace;i<source.length;i++){
  const c=source[i];if(quote){if(escaped)escaped=false;else if(c==='\\')escaped=true;else if(c===quote)quote=null;continue}
  if(c==='"'||c==="'"||c==='`'){quote=c;continue}
  if(c==='{')depth++;if(c==='}'&&!--depth)return source.slice(start,i+1);
 }
 throw Error('Missing production declaration');
}
class Element{
 constructor(tag,text=''){this.tag=tag;this.textContent=text;this.children=[];this.dataset={};this.removed=false;this.hidden=false}
 append(...nodes){this.children.push(...nodes)}remove(){this.removed=true}querySelector(){return null}
}
function fixture({pending=false}={}){
 const id='a'.repeat(64),host=new Element('host'),panel=new Element('section'),calls={configuration:0,restored:0,rendered:0,aborts:0,invalidated:0};
 const q={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:100,where:{field:{kind:'column',key:'body'},op:'eq',value:'saved'}};
 const view={id,version:1,title:'Saved original',queries:[q],layout:'records',columns:['utc','summary']};
 panel.dataset.phase=pending?'loading':'ready';
 const old={id,generation:1,pending,retired:false,controller:{abort(){calls.aborts++}}};
 const s={recallConflict:null,savedVisit:old,savedID:id,generation:1,busy:false,controller:null,panel,
  admitSavedView:value=>Composition.view(value),rememberSaved(){},node:(...args)=>new Element(...args),libraryHost:host,controlsHost:host,acknowledgedRecall:()=>null,canRecall:()=>false,
  nativeControlReady:n=>!n.removed,active:()=>true,location:{hash:'#exploration-view='+id},document:{activeElement:null},
  AbortController,structuredClone,Composition,CELProvenance,CapturedCount:{queryVariant:()=>false},sourceErrors:new Set(),sourceDrafts:new Map(),filterDrafts:new Map(),
  discardedAuthoring:new Set(),draftRevision:0,clearViews(){},closeChild(){},retireCountComparison(){},resetLive(){},visibility(){},
  invalidate(){calls.invalidated++},render(){calls.rendered++},currentRun:g=>g===s.generation,
  async api(path){assert.equal(path,'/api/exploration-views/'+id);calls.configuration++;return view},
  onRestore(restored){assert.deepEqual(restored,q);calls.restored++},
 };
 s.history={replaceState(a,b,hash){s.location.hash=hash}};
 vm.runInNewContext(['function clearRecallConflict(','function retireSavedLoad(','function retireSavedVisit(','function savedVisitCurrent(','async function load('].map(declaration).join('\n')+'\nthis.load=load;',s);
 return {s,id,host:panel,old,calls};
}

test('Keep then explicit discard can reopen the same saved view and retire the previous visit',async()=>{
 for(const pending of [false,true]){
  const f=fixture({pending});await f.s.load(f.id);assert.equal(f.host.children.length,1);
  const retiredDiscard=f.host.children[0].children.find(n=>n.textContent==='Open saved view and discard draft');
  f.host.children[0].children.find(n=>n.textContent==='Keep current draft').onclick();
  assert.equal(f.calls.configuration,0);assert.equal(f.calls.restored,0);
  await f.s.load(f.id);const current=f.host.children.at(-1);
  retiredDiscard.onclick();assert.equal(f.calls.configuration,0,'retired discard callback cannot act for the newer prompt');assert.equal(f.s.recallConflict.element,current);
  current.children.find(n=>n.textContent==='Open saved view and discard draft').onclick();
  await new Promise(resolve=>setImmediate(resolve));
  assert.equal(f.calls.configuration,1,'explicit same-ID discard must read the saved configuration again');
  assert.equal(f.calls.restored,1,'explicit discard must restore the original question');
  assert.equal(f.calls.rendered,1);assert.notEqual(f.s.savedVisit,f.old);assert.equal(f.s.recallConflict,null);
  assert.equal(f.calls.aborts,pending?1:0,'pending previous configuration load is retired');
  assert.equal(f.old.retired,pending);
 }
});

test('clean duplicate history notification for the current saved view remains a no-op',async()=>{
 const f=fixture();f.s.canRecall=()=>true;await f.s.load(f.id);await f.s.load(f.id);
 assert.equal(f.calls.configuration,0);assert.equal(f.calls.restored,0);assert.equal(f.s.savedVisit,f.old);assert.equal(f.host.children.length,0);
});


test('clean saved CEL navigation checks the current acknowledged view, independently of the requested ID',async()=>{
 const f=fixture(),next='b'.repeat(64);f.s.location.hash='#exploration-view='+next;let acknowledged;f.s.acknowledgedRecall=key=>{acknowledged=key;return {queries:[]}};f.s.canRecall=ack=>!!ack;f.s.api=async path=>{assert.equal(path,'/api/exploration-views/'+next);f.calls.configuration++;return {id:next,version:1,title:'Other',queries:[{signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:100}],layout:'records',columns:['utc','summary']}};
 await f.s.load(next);assert.equal(acknowledged,f.id);assert.equal(f.calls.configuration,1);assert.equal(f.s.recallConflict,null);
});
