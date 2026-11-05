import assert from 'node:assert/strict';
import {mock} from 'node:test';
import fs from 'node:fs';
const input=JSON.parse(fs.readFileSync(0,'utf8'));
class El {constructor(){this.hidden=false;this.dataset={};this.value='';this.children=[];this.previousElementSibling={};} append(...v){this.children.push(...v)} after(){} before(){} setAttribute(k,v){this[k]=v} getAttribute(k){return this[k]} addEventListener(){} closest(){return this} replaceChildren(...v){this.children=v} getClientRects(){return [1]} }
const elements=new Map(),el=id=>{if(!elements.has(id))elements.set(id,new El());return elements.get(id)};
globalThis.document={getElementById:el,querySelector:()=>null,hidden:false};
globalThis.location={hash:''};globalThis.history={replaceState(_a,_b,hash){location.hash=hash}};
const events=new EventTarget();globalThis.addEventListener=events.addEventListener.bind(events);globalThis.dispatchEvent=events.dispatchEvent.bind(events);
const emit=(type,detail)=>events.dispatchEvent(new CustomEvent(type,{detail,cancelable:true}));
const reads=[],shown=[],inspected=[];
const passive={mount:()=>({suspend(){},showSaved(){}})};
for(const name of ['distributions','inventory','saved-records'])mock.module('../../ui/'+name+'.mjs',{exports:passive});
for(const name of ['record-environment','documents','evidence-scope','explanation','measurements','saved-capture-protocol','workspace'])mock.module('../../ui/'+name+'.mjs',{exports:{}});
mock.module('../../ui/components.mjs',{exports:{node:()=>new El(),disclosure:()=>new El()}});
mock.module('../../ui/kept-saved.mjs',{exports:{library(){},mount:()=>({show(v){shown.push(v.id)},inspect(i){if(Number.isInteger(i)&&i>=0&&i<3){inspected.push(i);return true}return false}})}});
mock.module('../../ui/kept-client.mjs',{exports:{client:{open(id,{signal}){let resolve,reject;const promise=new Promise((a,b)=>{resolve=a;reject=b});reads.push({id,signal,resolve:()=>resolve({id,members:[{},{},{}]}),reject:()=>reject(Error('offline'))});return promise}}}});
const {mount}=await import(process.env.GOO_OWNER_MODULE||'../../ui/captures.mjs');mount({});
const tick=async()=>{for(let i=0;i<5;i++)await Promise.resolve()};
const route=id=>{location.hash='#kept='+id;emit('popstate');emit('hashchange')};
const explicit=(id,i)=>emit('goo:open-kept-member',{id,memberIndex:i});
for(let round=0;round<input.rounds;round++){
 const a=((round+input.seed)%15+1).toString(16).repeat(64),b=((round+input.seed+1)%15+1).toString(16).repeat(64),start=reads.length,paint=shown.length;
 route(a);for(let n=0;n<input.duplicates;n++)emit(n%2?'hashchange':'popstate');assert.equal(reads.length,start+1,'pending duplicate GET');
 explicit(a,2);assert.equal(reads.length,start+1);route(b);assert(reads[start].signal.aborted,'superseded acquisition not aborted');
 reads[start].resolve();await tick();assert.equal(shown.length,paint,'late success adopted');
 reads[start+1].resolve();await tick();assert.equal(shown.at(-1),b);assert.equal(inspected.length,round*2,'superseded hint leaked');
 for(let n=0;n<input.duplicates;n++)emit('hashchange');assert.equal(reads.length,start+2,'loaded duplicate GET');assert.equal(shown.length,paint+1,'loaded duplicate rebuild');
 explicit(b,1);explicit(b,99);assert.equal(inspected.at(-1),1);assert.equal(inspected.length,round*2+1,'invalid hint selected');assert.match(el('captureSavedStatus').textContent,/Matching record unavailable/,'refused inspect not explained');
 location.hash='#investigations';emit('hashchange');route(b);assert.equal(reads.length,start+3,'fresh visit reused old receipt');
 reads[start+2].reject();await tick();emit('hashchange');emit('popstate');assert.equal(reads.length,start+3,'failed visit passive duplicate retried');
 const click=new Event('click',{cancelable:true});Object.defineProperties(click,{button:{value:0},target:{value:{closest:()=>({getAttribute:()=>'#kept='+b})}}});events.dispatchEvent(click);assert(click.defaultPrevented);assert.equal(reads.length,start+4,'ordinary exact-link retry missing');explicit(b,2);reads[start+3].resolve();await tick();assert.equal(inspected.at(-1),2);
 location.hash='#document=other';emit('hashchange');route(a);const last=reads.at(-1);location.hash='#investigations';emit('hashchange');assert(last.signal.aborted,'same-workspace departure not aborted');const status=el('captureLibraryStatus').textContent;last.reject();await tick();assert.equal(el('captureLibraryStatus').textContent,status,'late failure published status');assert.equal(shown.length,paint+2,'late failure changed rendered receipt');
}
process.stdout.write(JSON.stringify({rounds:input.rounds,reads:reads.length}));
