import assert from 'node:assert/strict';
import {test,mock} from 'node:test';
class El {constructor(){this.hidden=false;this.dataset={};this.value='';this.children=[];this.previousElementSibling={};} append(...v){this.children.push(...v)} after(){} before(){} setAttribute(k,v){this[k]=v} getAttribute(k){return this[k]} addEventListener(){} closest(){return this} replaceChildren(...v){this.children=v} getClientRects(){return [1]} get firstChild(){return this.children[0]} showModal(){this.open=true} close(){this.open=false} focus(){} }
const elements=new Map(),el=id=>{if(!elements.has(id))elements.set(id,new El());return elements.get(id)};
globalThis.document={getElementById:el,querySelector:()=>null,hidden:false};
globalThis.location={hash:''};globalThis.history={replaceState(_a,_b,hash){location.hash=hash}};
const events=new EventTarget();globalThis.addEventListener=events.addEventListener.bind(events);globalThis.dispatchEvent=events.dispatchEvent.bind(events);
const emit=(type,detail)=>events.dispatchEvent(new CustomEvent(type,{detail,cancelable:true}));
const reads=[],shown=[],inspected=[],painted=[];
const passive={mount:()=>({suspend(){},showSaved(page){painted.push(page.receipt.id)}})};
for(const name of ['distributions','inventory','saved-records'])mock.module('../ui/'+name+'.mjs',{exports:passive});
for(const name of ['record-environment','evidence-scope','explanation','measurements'])mock.module('../ui/'+name+'.mjs',{exports:{}});
mock.module('../ui/components.mjs',{exports:{node:()=>new El(),disclosure:()=>new El()}});
mock.module('../ui/kept-saved.mjs',{exports:{library(){},mount:()=>({show(v){shown.push(v.id)},inspect(i){if(Number.isInteger(i)&&i>=0&&i<3){inspected.push(i);return true}return false}})}});
mock.module('../ui/kept-client.mjs',{exports:{client:{open(id,{signal}){let resolve,reject;const promise=new Promise((a,b)=>{resolve=a;reject=b});reads.push({id,signal,resolve:()=>resolve({id,members:[{},{},{}]}),reject:()=>reject(Error('offline'))});return promise}}}});
mock.module('../ui/saved-capture-protocol.mjs',{exports:{validate(){},validateRefresh(){}}});
const attached=[];mock.module('../ui/documents.mjs',{exports:{context:()=>null,attach:async(target,reference)=>{attached.push({target,reference});return {id:'revision'}}}});
mock.module('../ui/workspace.mjs',{exports:{routeFromHash:hash=>hash==='#explore'?'explore':'investigations'}});
globalThis.Option=class extends El {constructor(text,value){super();this.textContent=text;this.value=value}};
const {mount}=await import(process.env.GOO_OWNER_MODULE||'../ui/captures.mjs');
const pending=[];globalThis.fetch=(url,options={})=>new Promise((resolve,reject)=>pending.push({url,method:options.method,signal:options.signal,resolve:page=>resolve({ok:true,json:async()=>page}),reject}));
const tick=async()=>{for(let i=0;i<12;i++)await Promise.resolve()};
const a='a'.repeat(64),b='b'.repeat(64);
const page=id=>({receipt:{id,title:id,capturedAt:'2026-01-01T00:00:00Z',question:{inventory:{fromMs:0,toMs:1,environment:'prod'}}},inventory:{count:1,state:'ready',basis:'stored'}});
const route=id=>{location.hash='#capture='+id;emit('popstate');emit('hashchange')};
test('capture acquisition belongs to route visit, joins notifications, retires before abort and retries explicitly',async()=>{
 const owner=mount({});route(a);assert.equal(pending.length,1);const first=pending[0];
 route(b);assert(first.signal.aborted);first.resolve(page(a));await tick();assert.notEqual(location.hash,'#capture='+a);assert.notEqual(el('captureGeneralID').value,a);
 pending[1].resolve(page(b));await tick();assert.equal(el('captureGeneralID').value,b);const status=el('captureSavedStatus').textContent,paintCount=painted.length;
 emit('hashchange');emit('popstate');assert.equal(pending.length,2);assert.equal(el('captureSavedStatus').textContent,status);assert.equal(painted.length,paintCount,'passive route must preserve evidence DOM');
 location.hash='#investigations';emit('hashchange');route(b);assert.equal(pending.length,3);
 pending[2].reject(Error('offline'));await tick();assert.match(el('captureLibraryStatus').textContent,/offline/);emit('popstate');emit('hashchange');assert.equal(pending.length,3);
 owner.open(b);assert.equal(pending.length,4);pending[3].resolve(page(b));await tick();
 owner.open(b,'next');assert.equal(pending.length,5);emit('hashchange');assert.equal(pending.length,5);assert.match(pending[4].url,/cursor=next/);
 location.hash='#investigations';emit('hashchange');const failureStatus=el('captureLibraryStatus').textContent;assert(pending[4].signal.aborted);pending[4].reject(Error('late failure'));await tick();assert.equal(el('captureLibraryStatus').textContent,failureStatus);assert.equal(location.hash,'#investigations');
 owner.open(a);assert.equal(pending.length,6);pending[5].resolve(page(a));await tick();assert.equal(location.hash,'#capture='+a);
 owner.open(b);const last=pending[6];emit('goo:workspace',{route:'explore'});assert(last.signal.aborted);last.resolve(page(b));await tick();assert.equal(el('captureGeneralID').value,a);
 const refresh=el('captureGeneralRefresh').onclick();assert.equal(pending.length,8);emit('popstate');emit('hashchange');assert.equal(pending.length,8,'passive notifications must not supersede refresh');pending[7].reject(Error('refresh offline'));await refresh;assert.match(el('captureSavedStatus').textContent,/Refresh unavailable/);
 const departedRefresh=el('captureGeneralRefresh').onclick();location.hash='#investigations';emit('hashchange');pending[8].resolve(page(b));await departedRefresh;assert.equal(location.hash,'#investigations','retired refresh must not restore URL');assert.equal(el('captureGeneralID').value,a);
 for(const aba of [false,true,'queued']){location.hash='#explore';emit('hashchange');emit('goo:save-capture',{question:{kind:'inventory'},title:'held save'});pending.at(-1).resolve({items:[]});await tick();el('captureDraftInvestigation').value='target';const before=painted.length,save=el('captureDraftConfirm').onclick(),post=pending.at(-1);assert.equal(post.method,'POST');assert.equal(post.signal,undefined);if(aba==='queued'){const departure=new Event('hashchange');Object.defineProperty(departure,'newURL',{value:'https://example.test/#investigations'});events.dispatchEvent(departure)}else {location.hash='#investigations';emit('hashchange');if(aba){location.hash='#explore';emit('popstate');emit('hashchange')}}const status=el('captureSavedStatus').textContent;post.resolve(page(b));await save;assert.equal(el('captureDialog').open,false,'original completed Save dialog must close after departure');assert.equal(painted.length,before,'departed Save must not adopt');assert.equal(location.hash,aba?'#explore':'#investigations');assert.equal(el('captureSavedStatus').textContent,status);assert.equal(attached.length,aba==='queued'?3:aba?2:1,'requested attachment must still complete');}
 location.hash='#explore';emit('hashchange');emit('goo:save-capture',{question:{kind:'inventory'}});pending.at(-1).resolve({items:[]});await tick();el('captureDraftInvestigation').value='target';const validSave=el('captureDraftConfirm').onclick(),validPost=pending.at(-1);emit('popstate');emit('hashchange');validPost.resolve(page(b));await validSave;assert.equal(location.hash,'#capture='+b,'unchanged Save adopts after passive events');assert.equal(attached.length,4);assert.match(el('captureSavedStatus').textContent,/Evidence attached/);
 const startDraft=async title=>{emit('goo:save-capture',{question:{kind:'inventory'},title});pending.at(-1).resolve({items:[]});await tick()};
 await startDraft('older');el('captureDraftInvestigation').value='target';const olderSave=el('captureDraftConfirm').onclick(),olderPost=pending.at(-1);await startDraft('newer');const newerStatus=el('captureDraftStatus').textContent;olderPost.resolve(page(a));await olderSave;assert.equal(el('captureDialog').open,true,'older completion must not close newer dialog');assert.equal(el('captureDraftTitle').value,'newer');assert.equal(el('captureDraftStatus').textContent,newerStatus);assert.equal(attached.length,5,'superseded draft still completes requested attachment');
 const failedSave=el('captureDraftConfirm').onclick(),failedPost=pending.at(-1);location.hash='#explore';emit('hashchange');failedPost.reject(Error('save offline'));await failedSave;assert.equal(el('captureDialog').open,true);assert.equal(el('captureDraftConfirm').disabled,false);assert.equal(el('captureDraftInvestigation').disabled,false);assert.match(el('captureDraftStatus').textContent,/save offline/);assert.equal(location.hash,'#explore');
 const oldestRetry=el('captureDraftConfirm').onclick(),retryPost=pending.at(-1);await startDraft('latest');const latestStatus=el('captureDraftStatus').textContent;retryPost.reject(Error('stale failure'));await oldestRetry;assert.equal(el('captureDialog').open,true);assert.equal(el('captureDraftTitle').value,'latest');assert.equal(el('captureDraftStatus').textContent,latestStatus);assert.equal(el('captureDraftConfirm').disabled,false);

});
