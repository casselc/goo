import assert from 'node:assert/strict';
import {mock} from 'node:test';
import fs from 'node:fs';
const input=JSON.parse(fs.readFileSync(0,'utf8'));
class El {constructor(){this.hidden=false;this.dataset={};this.value='';this.children=[];this.previousElementSibling={};this.isConnected=true;} append(...v){this.children.push(...v)} after(){} before(){} setAttribute(k,v){this[k]=v} getAttribute(k){return this[k]} addEventListener(){} closest(){return this} replaceChildren(...v){this.children=v} getClientRects(){return [1]} focus(){focuses++} scrollIntoView(){} showModal(){this.open=true} close(){this.open=false} get firstChild(){return this.children[0]} }
let focuses=0;globalThis.Option=class extends El {constructor(text,value){super();this.textContent=text;this.value=value}};
const elements=new Map(),el=id=>{if(!elements.has(id))elements.set(id,new El());return elements.get(id)};
globalThis.document={getElementById:el,querySelector:()=>null,hidden:false};
globalThis.location={hash:''};globalThis.history={replaceState(_a,_b,hash){location.hash=hash}};
const events=new EventTarget();globalThis.addEventListener=events.addEventListener.bind(events);globalThis.dispatchEvent=events.dispatchEvent.bind(events);
const emit=(type,detail)=>events.dispatchEvent(new CustomEvent(type,{detail,cancelable:true}));
const reads=[],shown=[],writes=[],attachments=[];let page;
const passive={mount:()=>({suspend(){},showSaved(){}})};
for(const name of ['distributions','inventory'])mock.module('../../ui/'+name+'.mjs',{exports:passive});
mock.module('../../ui/saved-records.mjs',{exports:{mount:(_host,options)=>{page=options.onPage;return {suspend(){},showSaved(v){shown.push({id:v.receipt.id,cursor:v.cursor})}}}}});
mock.module('../../ui/documents.mjs',{exports:{context:()=>null,attach(target,reference){let resolve;const promise=new Promise(r=>{resolve=r});attachments.push({target,reference,resolve:()=>resolve({id:'d'.repeat(64)})});return promise}}});
mock.module('../../ui/workspace.mjs',{exports:{routeFromHash:hash=>/^#(?:capture|kept|document)=/.test(hash)||hash==='#investigations'?'investigations':hash==='#services'?'services':'explore'}});
for(const name of ['explanation','measurements'])mock.module('../../ui/'+name+'.mjs',{exports:{}});
mock.module('../../ui/record-environment.mjs',{exports:{questionLabel:()=> 'prod'}});
mock.module('../../ui/evidence-scope.mjs',{exports:{record:x=>x,element:()=>new El()}});
mock.module('../../ui/saved-capture-protocol.mjs',{exports:{validate(){},validateRefresh(){}}});
mock.module('../../ui/components.mjs',{exports:{node:()=>new El(),disclosure:()=>new El(),actionVisible:()=>true}});
mock.module('../../ui/kept-saved.mjs',{exports:{library(){},mount:()=>({show(){},inspect(){return true}})}});
mock.module('../../ui/kept-client.mjs',{exports:{client:{open:async id=>({id,members:[]})}}});
function capture(id,cursor=''){const question={signal:'logs',environment:'prod',fromMs:0,toMs:60000,limit:100,offset:cursor?100:0};return {cursor,receipt:{id,title:'Capture '+id,capturedAt:'2026-11-05T00:00:00Z',question:{kind:'explore',explore:question}},explore:{question,count:2,state:'ok',rows:[],basis:{}},retainedRecords:2};}
globalThis.fetch=(url,options={})=>{if(String(url)==='/api/captures'&&options.method==='POST'){let resolve,reject;const promise=new Promise((a,b)=>{resolve=a;reject=b});const body=JSON.parse(options.body);writes.push({body,signal:options.signal,resolve:()=>resolve({ok:true,json:async()=>capture('f'.repeat(64))}),reject:()=>reject(Error('offline write'))});return promise}const match=String(url).match(/^\/api\/captures\/([a-f0-9]{64})(?:\?cursor=(.*))?$/);if(!match)return Promise.resolve({ok:true,json:async()=>({items:[],matched:0,total:0})});let resolve,reject;const promise=new Promise((a,b)=>{resolve=a;reject=b});const id=match[1],cursor=decodeURIComponent(match[2]||'');reads.push({id,cursor,signal:options.signal,resolve:()=>resolve({ok:true,json:async()=>capture(id,cursor)}),reject:()=>reject(Error('offline'))});return promise;};
const {mount}=await import(process.env.GOO_OWNER_MODULE||'../../ui/captures.mjs');mount({});
const tick=async()=>{for(let i=0;i<12;i++)await Promise.resolve()};
const route=id=>{location.hash='#capture='+id;emit('goo:workspace',{route:'investigations'});emit('popstate');emit('hashchange')};
const leave=()=>{location.hash='#investigations';emit('goo:workspace',{route:'investigations'});emit('hashchange')};
function retry(id){el('captureGeneralID').value=id;el('captureGeneralOpen').onclick();}
for(let round=0;round<input.rounds;round++){
 const a=((round+input.seed)%15+1).toString(16).repeat(64),b=((round+input.seed+1)%15+1).toString(16).repeat(64),start=reads.length,paint=shown.length;
 route(a);for(let n=0;n<input.duplicates;n++)emit(n%2?'hashchange':'popstate');assert.equal(reads.length,start+1,'pending duplicate GET');
 route(b);assert(reads[start].signal?.aborted,'superseded acquisition not aborted');reads[start].resolve();await tick();assert.equal(shown.length,paint,'late distinct-ID success adopted');
 reads[start+1].resolve();await tick();assert.deepEqual(shown.at(-1),{id:b,cursor:''});assert.equal(shown.length,paint+1);
 for(let n=0;n<input.duplicates;n++)emit(n%2?'hashchange':'popstate');assert.equal(reads.length,start+2,'loaded duplicate GET');assert.equal(shown.length,paint+1,'loaded duplicate rebuild');
 page('next-page');assert.equal(reads.length,start+3,'page intent did not acquire');emit('hashchange');emit('popstate');assert.equal(reads.length,start+3,'passive route replaced pending cursor');reads[start+2].resolve();await tick();assert.deepEqual(shown.at(-1),{id:b,cursor:'next-page'});
 leave();route(b);assert.equal(reads.length,start+4,'fresh same-ID visit reused old receipt');reads[start+3].reject();await tick();emit('hashchange');emit('popstate');assert.equal(reads.length,start+4,'failed passive visit retried');
 retry(b);assert.equal(reads.length,start+5,'explicit retry missing');reads[start+4].resolve();await tick();assert.equal(shown.length,paint+3);
 route(a);const pending=reads.at(-1);leave();assert(pending.signal?.aborted,'same-workspace departure not aborted');await tick();const status=el('captureLibraryStatus').textContent;pending.resolve();await tick();assert.equal(shown.length,paint+3,'late same-workspace success adopted');assert.equal(location.hash,'#investigations','late success restored departed route');assert.equal(el('captureLibraryStatus').textContent,status,'late success changed status');
 route(a);const failed=reads.at(-1);leave();assert(failed.signal?.aborted,'failure departure not aborted');await tick();const failureStatus=el('captureLibraryStatus').textContent;failed.reject();await tick();assert.equal(shown.length,paint+3);assert.equal(el('captureLibraryStatus').textContent,failureStatus,'late failure published status');assert.equal(location.hash,'#investigations');
 route(b);reads.at(-1).resolve();await tick();const ready=shown.length;const refresh=el('captureGeneralRefresh').onclick();const write=writes.at(-1);for(let n=0;n<input.duplicates;n++)emit(n%2?'hashchange':'popstate');assert.equal(reads.length,start+8,'passive event acquired during refresh');leave();write.resolve();await refresh;await tick();assert.equal(shown.length,ready,'late refresh success adopted');assert.equal(location.hash,'#investigations','late refresh restored departed route');assert.equal(write.body.parent,b);assert.equal(write.signal,undefined,'authorized POST made abortable');
 route(b);reads.at(-1).resolve();await tick();const after=shown.length;const failedRefresh=el('captureGeneralRefresh').onclick();const failedWrite=writes.at(-1);leave();await tick();const savedStatus=el('captureSavedStatus').textContent;failedWrite.reject();await failedRefresh;await tick();assert.equal(shown.length,after,'late refresh failure rebuilt receipt');assert.equal(el('captureSavedStatus').textContent,savedStatus,'late refresh failure published status');assert.equal(location.hash,'#investigations');
 for(const transition of ['depart','aba','queued-aba','stay']){
  location.hash='#explore';emit('goo:workspace',{route:'explore'});emit('hashchange');
  emit('goo:save-capture',{question:{kind:'explore',explore:capture(b).explore.question},title:'Saved from Activity'});el('captureDraftInvestigation').value='d'.repeat(64);const confirming=el('captureDraftConfirm').onclick(),save=writes.at(-1),beforeSave=shown.length;
  assert.equal(save.signal,undefined,'Save publication made abortable');assert.equal(save.body.question.kind,'explore');
  if(transition==='queued-aba'){const event=new Event('hashchange');Object.defineProperty(event,'newURL',{value:'http://goo.invalid/#services'});events.dispatchEvent(event)}else if(transition!=='stay'){location.hash='#services';emit('goo:workspace',{route:'services'});emit('hashchange');if(transition==='aba'){location.hash='#explore';emit('goo:workspace',{route:'explore'});emit('hashchange')}}else {emit('popstate');emit('hashchange')}
  await tick();const statusBeforeSave=el('captureSavedStatus').textContent,focusBeforeSave=focuses;save.resolve();await tick();assert.equal(el('captureDialog').open,false,'original completed Save dialog remained open');
  const attached=attachments.at(-1);assert.equal(attached.target,'d'.repeat(64),'authorized attachment lost');assert.deepEqual(attached.reference,{kind:'capture',reference:'f'.repeat(64),caption:'Capture '+'f'.repeat(64)});
  if(transition!=='stay'){assert.equal(shown.length,beforeSave,'departed Save adopted receipt');assert.equal(location.hash,transition.endsWith('aba')?'#explore':'#services','departed Save restored route');assert.equal(el('captureSavedStatus').textContent,statusBeforeSave,'departed Save published status');assert.equal(focuses,focusBeforeSave,'departed Save moved focus')}
  else {assert.equal(shown.length,beforeSave+1,'current Save failed to adopt');assert.equal(location.hash,'#capture='+'f'.repeat(64))}
  attached.resolve();await confirming;await tick();
  if(transition!=='stay'){assert.equal(el('captureSavedStatus').textContent,statusBeforeSave,'departed attachment published status');assert.equal(focuses,focusBeforeSave,'departed attachment moved focus')}
 }
 assert.equal(reads.length,start+9,'unexpected read budget');leave();await tick();
}
process.stdout.write(JSON.stringify({rounds:input.rounds,reads:reads.length,adoptions:shown.length,writes:writes.length,attachments:attachments.length}));
