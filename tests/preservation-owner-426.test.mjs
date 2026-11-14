import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as Preservation from '../ui/preservation.mjs';
import * as Composition from '../ui/view-composition.mjs';
import * as CapturedCount from '../ui/captured-count-comparison.mjs';
import * as QuestionProtocol from '../ui/question-protocol.mjs';
import * as ViewArtifact from '../ui/view-artifact.mjs';
import * as RecordEnvironment from '../ui/record-environment.mjs';
import * as QuestionSession from '../ui/question-session.mjs';
import {plan} from '../ui/counted-selection-comparison.mjs';
import {nativeControlReady,completionControlReady} from '../ui/dom-visibility.mjs';

// Deliberately report geometry for closed disclosure contents: geometry alone
// cannot establish that a native action is available.
class Element {
 constructor(tag,text=''){this.localName=tag;this._text=text;this.children=[];this.attrs={};this.dataset={};this.open=false;this.hidden=false;this.inert=false;this.disabled=false;this.value='';this.root=false;this.className='';this.classList={add:name=>this.className+=' '+name};}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}
 get textContent(){return this._text+this.children.map(x=>x?.textContent??x).join('')}set textContent(v){this._text=String(v);this.replaceChildren()}
 append(...nodes){for(const n of nodes){if(typeof n==='object'){n.remove();n.parentElement=this;}this.children.push(n)}}replaceChildren(...nodes){for(const n of [...this.children])n?.remove?.();this.children=[];this.append(...nodes)}remove(){if(this.parentElement){this.parentElement.children=this.parentElement.children.filter(n=>n!==this);this.parentElement=null}}
 contains(n){return n===this||this.children.some(c=>c?.contains?.(n))}setAttribute(k,v){this.attrs[k]=String(v)}getAttribute(k){return this.attrs[k]??null}
 matches(s){if(s==='details:not([open])')return this.localName==='details'&&!this.open;if(s===':disabled'){for(let n=this;n;n=n.parentElement)if(n.disabled&&(n===this||n.localName==='fieldset')&&!n.querySelector(':scope > legend')?.contains(this))return true;return false;}if(s==='fieldset:disabled')return this.localName==='fieldset'&&this.disabled;if(s==='[hidden]')return this.hidden;if(s==='[inert]')return this.inert;if(s==='[aria-hidden="true"]')return this.getAttribute('aria-hidden')==='true';return s===this.localName}
 closest(s){for(let n=this;n;n=n.parentElement)if(s.split(',').some(x=>n.matches(x.trim())))return n;return null}
 querySelector(s){if(s===':scope > summary'||s===':scope > legend')return this.children.find(n=>n.localName===s.slice(9))||null;return walk(this).find(n=>n!==this&&n.matches(s))||null}getClientRects(){return this.isConnected&&!this.hidden?[{}]:[]}select(){document.activeElement=this}focus(){document.activeElement=this}addEventListener(){}
}
const walk=n=>[n,...n.children.filter(c=>typeof c==='object').flatMap(walk)],node=(tag,text)=>new Element(tag,text);
const source=readFileSync(new URL('../ui/population-workspace.mjs',import.meta.url),'utf8');
const between=(a,b)=>source.slice(source.indexOf(a),source.indexOf(b,source.indexOf(a)));
const mountSource=between('  function mountPreservedQuestion(','  function mountTemplateAuthor(');
const linkSource=between('  function preservationCurrent(','  let facetRead=');
const renderLine=source.split('\n').find(line=>line.trimStart().startsWith('const owner=preservationOwner='));
const saveSource=renderLine.slice(0,renderLine.indexOf(';if(CapturedCount.queryVariant'))+';panel.append(saveBox);globalThis.controls={saveBox,save,name,saved,meaningHost};';
const field={kind:'attribute',key:'cohort'},parent={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:1,offset:0,groupBy:[field]};
const view=plan(parent,field,[null,'']);
function fixture({deferred=false,savedID=null}={}){
 const previous=globalThis.document,document={hidden:false,activeElement:null,createElement:tag=>node(tag)};globalThis.document=document;
 const panel=node('section');panel.root=true;const env={node,structuredClone,RecordEnvironment,QuestionSession,Preservation,Composition,CapturedCount,QuestionProtocol,nativeControlReady,completionControlReady,document,panel,queries:structuredClone(view.queries),composition:structuredClone(view.composition),generation:7,preservationOwner:null,disposed:false,draftRevision:3,savedRevision:savedID?3:-1,savedID,title:view.title,authoring:{},discardedAuthoring:new Set(),sourceErrors:new Set(),filterDrafts:new Map(),selection:[],results:[{widget:{key:'logs'},result:{}}],views:new Map(),busy:false,childOwner:null,attachmentPreserveOpen:true,enabled:true,writes:[],staged:[],events:[],meanings:0,CustomEvent,location:{origin:'https://goo.test',pathname:'/',search:''},global:{dispatchEvent:e=>env.events.push(e)},Documents:{context:()=>({}),stageRecordQuestion:q=>env.staged.push(q)},Meaning:{mount:()=>env.meanings++},hasPendingRecordSelection:()=>false,refreshViewAttachment(){},active:()=>env.enabled,currentRun:g=>env.enabled&&g===env.generation&&!env.childOwner,settings:()=>structuredClone({...view,title:env.title}),markDraft:()=>{env.draftRevision++;env.savedID=null}};
 env.api=async(_path,body)=>{env.writes.push(body);if(deferred)return new Promise(resolve=>env.resolve=resolve);return{id:'a'.repeat(64)}};
 vm.createContext(env);vm.runInContext(linkSource+mountSource+saveSource,env);const controls=env.controls;controls.saveBox.open=true;
 return {...controls,env,panel,button:text=>walk(panel).find(e=>e.localName==='button'&&e.textContent===text),link:()=>walk(panel).find(e=>e.localName==='a'&&e.textContent.startsWith('Open saved view ')),close:()=>globalThis.document=previous};
}
const click=()=>new Event('click',{cancelable:true});

test('426 withdrawn native preservation controls cannot save, share, stage, capture or edit title',async()=>{
 const f=fixture();try{const controls=[f.save,f.button('Share selected question'),f.button('Add selected question to investigation'),f.button('Capture selected query')],initial=f.env.title;
  const deny=async()=>{for(const c of controls)await c.onclick();f.name.value='unavailable title';f.name.oninput();assert.equal(f.env.title,initial);assert.equal(f.env.writes.length,0);assert.equal(f.env.staged.length,0);assert.equal(f.env.events.length,0);assert(!walk(f.panel).find(e=>e.getAttribute('aria-label')==='Shared question link').value)};
  f.saveBox.open=false;assert.equal(f.save.getClientRects().length,1);await deny();f.saveBox.open=true;
  for(const key of ['hidden','inert']){f.saveBox[key]=true;await deny();f.saveBox[key]=false;}
  f.saveBox.setAttribute('aria-hidden','true');await deny();f.saveBox.setAttribute('aria-hidden','false');
  for(const c of [...controls,f.name]){c.disabled=true;c.setAttribute('aria-disabled','true')}await deny();for(const c of [...controls,f.name]){c.disabled=false;c.setAttribute('aria-disabled','false')}
  f.env.enabled=false;await deny();f.env.enabled=true;f.env.preservationOwner={};await deny();
 }finally{f.close()}
});

test('426 valid view preservation keeps both queries while selected question/evidence actions remain explicit',async()=>{
 const f=fixture();try{
  f.button('Share selected question').onclick();const shared=walk(f.panel).find(e=>e.getAttribute('aria-label')==='Shared question link');assert.equal(shared.hidden,false);assert.deepEqual(QuestionProtocol.readQuestionLink(new URL(shared.value).hash).question,view.queries[0]);
  f.button('Add selected question to investigation').onclick();assert.equal(f.env.staged.length,1);
  f.button('Capture selected query').onclick();assert.equal(f.env.events.length,1);assert.deepEqual(f.env.events[0].detail.question.explore,view.queries[0]);
  await f.save.onclick();assert.equal(f.env.writes.length,1);assert.deepEqual(f.env.writes[0].queries,view.queries);assert.deepEqual(Composition.comparison(f.env.writes[0].composition).baselineQuery,Composition.comparison(view.composition).baselineQuery);assert.equal(f.env.savedID,'a'.repeat(64));
  let link=f.link();let e=click();link.onclick(e);assert.equal(e.defaultPrevented,false);
  for(const mode of ['closed','inert','aria']){if(mode==='closed')f.saveBox.open=false;if(mode==='inert')f.saveBox.inert=true;if(mode==='aria')link.setAttribute('aria-disabled','true');e=click();link.onclick(e);assert.equal(e.defaultPrevented,true);f.saveBox.open=true;f.saveBox.inert=false;link.setAttribute('aria-disabled','false')}
  f.env.generation++;e=click();link.onclick(e);assert.equal(e.defaultPrevented,false,'held parent return keeps the same render owner across child query generations');await f.save.onclick();assert.equal(f.env.writes.length,2,'the same held Preserve controls remain usable after child return');link=f.link();assert(nativeControlReady(link));f.env.preservationOwner={};e=click();link.onclick(e);assert.equal(e.defaultPrevented,true,'an identical saved ID does not revive a departed render');
 }finally{f.close()}
});

test('426 pending save may complete while its own button is disabled, but cannot publish a withdrawn receipt',async()=>{
 for(const mode of ['ready','closed','inert','revision','generation','owner','fieldset']){const f=fixture({deferred:true});try{
  const pending=f.save.onclick();assert.equal(f.env.writes.length,1);assert.equal(f.save.disabled,true);
  if(mode==='closed')f.saveBox.open=false;if(mode==='inert')f.saveBox.inert=true;if(mode==='revision')f.env.draftRevision++;if(mode==='generation')f.env.generation++;if(mode==='owner')f.env.preservationOwner={};if(mode==='fieldset'){const fieldset=node('fieldset');fieldset.disabled=true;f.panel.append(fieldset);fieldset.append(f.saveBox)}
  f.env.resolve({id:'a'.repeat(64)});await pending;
  assert.equal(f.env.savedID,mode==='ready'?'a'.repeat(64):null,mode);assert.equal(!!f.link(),mode==='ready',mode);assert.equal(f.env.meanings,mode==='ready'?1:0,mode);
 }finally{f.close()}}
});


function portableFixture(kind){
 const f=fixture({deferred:true}),env=f.env;env.ViewArtifact=ViewArtifact;env.URL={createObjectURL:()=> 'blob:checked',revokeObjectURL(){}};env.Blob=Blob;env.countPending=false;env.compositionReady=()=>true;env.refreshArtifactExport=env.refreshArtifactImport=()=>{};env.listViews=async()=>{};env.library=node('details');env.library.open=true;f.panel.append(env.library);
 if(kind==='export'){vm.runInContext(between('  function mountArtifactExport(','  function mountArtifactImport(')+';mountArtifactExport(panel);',env);f.box=f.panel.children.at(-1)}
 else{vm.runInContext(between('  function mountArtifactImport(','  function mountViewAttachment(')+';mountArtifactImport();',env);f.box=env.library.children.at(-1)}
 f.box.open=true;f.find=text=>walk(f.box).find(n=>['button','a'].includes(n.localName)&&n.textContent===text);f.response={artifact:ViewArtifact.artifact(env.settings(),{namespace:'team',name:'view',revision:'1'}),yaml:'version: goo.view-artifact/1'};return f;
}

test('426 withdrawn export completion drops output but releases its own pending state for retry',async()=>{
 for(const mode of ['inert','generation']){const f=portableFixture('export');try{const run=f.find('Export YAML'),form=walk(f.box).find(e=>e.localName==='form'),output=walk(f.box).find(e=>e.localName==='textarea'),event={preventDefault(){}};
  const pending=form.onsubmit(event);assert.equal(run.disabled,true);if(mode==='inert')f.box.inert=true;else f.env.generation++;f.env.resolve(f.response);await pending;assert.equal(output.hidden,true);assert.equal(run.disabled,false);assert.equal(f.find('Download YAML').hidden,true);
  f.box.inert=false;const retry=form.onsubmit(event);f.env.resolve(f.response);await retry;assert.equal(output.hidden,false);assert.equal(run.disabled,false);assert.equal(f.find('Download YAML').hidden,false);
 }finally{f.close()}}
});

test('426 withdrawn preview and import completions release pending state without adopting admission or receipt',async()=>{
 for(const mode of ['inert','generation']){const f=portableFixture('import');try{const preview=f.find('Preview YAML'),save=f.find('Import view'),source=walk(f.box).find(e=>e.localName==='textarea');source.value='version: goo.view-artifact/1';source.oninput();
  let pending=preview.onclick();assert.equal(preview.disabled,true);if(mode==='inert')f.box.inert=true;else f.env.generation++;f.env.resolve(f.response);await pending;assert.equal(preview.disabled,false);assert.equal(save.disabled,true);
  f.box.inert=false;pending=preview.onclick();f.env.resolve(f.response);await pending;assert.equal(preview.disabled,false);assert.equal(save.disabled,false,f.box.textContent);
  pending=save.onclick();assert.equal(save.disabled,true);if(mode==='inert')f.box.inert=true;else f.env.generation++;f.env.resolve({id:'a'.repeat(64)});await pending;assert.equal(preview.disabled,false);assert.equal(save.disabled,true);assert(!f.box.textContent.includes('Imported team'));
  f.box.inert=false;pending=preview.onclick();f.env.resolve(f.response);await pending;assert.equal(save.disabled,false,'fresh preview restores explicit import after withdrawal');
 }finally{f.close()}}
});
