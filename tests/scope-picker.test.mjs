import {test,mock} from 'node:test';import assert from 'node:assert/strict';
class El{
 constructor(tag,text=''){this.tag=tag;this.textContent=text;this.children=[];this.dataset={};this.attributes=new Map();this.hidden=false;this.disabled=false;this.inert=false;this.value='';this.listeners={};this.open=false;}
 get parentElement(){return this.parent||null;}get isConnected(){return this.root||!!this.parent?.isConnected;}
 append(...ns){for(const n of ns){n.remove();n.parent=this;this.children.push(n);}}prepend(n){n.parent=this;this.children.unshift(n);}remove(){if(this.parent)this.parent.children=this.parent.children.filter(x=>x!==this);this.parent=null;}replaceChildren(...ns){for(const n of this.children)n.parent=null;this.children=[];this.append(...ns);}
 setAttribute(k,v){this.attributes.set(k,String(v));}getAttribute(k){return this.attributes.get(k)??null;}addEventListener(k,f){this.listeners[k]=f;}getClientRects(){return this.isConnected?[{}]:[];}contains(e){return e===this||walk(this).includes(e);}focus(){document.activeElement=this;}
 matches(s){return s===':disabled'&&this.disabled;}
 closest(s){for(let e=this;e;e=e.parent){if(s==='[inert]'&&e.inert||s==='details:not([open])'&&e.tag==='details'&&!e.open||s==='fieldset:disabled'&&e.tag==='fieldset'&&e.disabled||s.includes('[hidden]')&&(e.hidden||e.getAttribute('aria-hidden')==='true'))return e;}return null;}
 querySelector(s){if(s===':scope > summary'||s===':scope > legend')return this.children.find(n=>n.tag===s.split(' > ')[1]);return walk(this).find(n=>n.tag===s);}
}
const walk=e=>e.children.flatMap(x=>[x,...walk(x)]);mock.module('../ui/components.mjs',{exports:{node:(tag,text)=>new El(tag,text)}});const {mount,scopeSummary}=await import('../ui/scope-picker.mjs');const {Journal}=await import('../ui/scope-history.mjs');const q={kind:'question',question:{signal:'logs',environment:'prod',fromMs:0,toMs:60000,limit:100,offset:0}};
test('picker stages only through supplied owner and retires reattached controls',()=>{globalThis.document={createTextNode:t=>new El('text',t)};const host=new El('main');host.root=true;const journal=new Journal({storage:()=>({getItem:()=>null,removeItem(){}}),now:()=>1000}),staged=[],held=[];journal.remember(q);const owner=mount(host,{journal,current:()=>q,onStage:v=>{staged.push(v);return true},visits:()=>[{id:1,title:'Old view',current:false}],onVisit:id=>held.push(id)});owner.root.open=true;owner.refresh();let stage=walk(owner.root).find(x=>x.tag==='button'&&x.textContent==='Stage');const retired=stage;owner.refresh();host.append(retired);retired.onclick();assert.equal(staged.length,0);stage=walk(owner.root).find(x=>x.tag==='button'&&x.textContent==='Stage');stage.onclick();assert.equal(staged.length,1);assert.deepEqual(staged[0].question,q.question);assert.equal(owner.root.open,false);owner.root.open=true;owner.refresh();walk(owner.root).find(x=>x.dataset.visitId==='1').onclick();assert.deepEqual(held,[1]);owner.root.open=true;owner.refresh();walk(owner.root).find(x=>x.textContent==='Pin current scope').onclick();assert.equal(journal.list()[0].pinned,true);walk(owner.root).find(x=>x.textContent==='Clear recent').onclick();assert.equal(journal.list().length,1);walk(owner.root).find(x=>x.textContent==='Reset scope history').onclick();assert.deepEqual(journal.list(),[]);});


test('scope presentation separates compact UTC from exact bounds and preserves distinguishing filters',()=>{
 const question={...q.question,fromMs:1794520750373,toMs:1794520814373,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'edge-'+'x'.repeat(128)},{field:{kind:'column',key:'severity'},op:'eq',value:'INFO'}]}};
 const result=scopeSummary({kind:'question',question});
 assert(result.primary.includes('INFO'),'the last distinguishing filter must remain visible');
 assert(result.meta.includes('Fixed question · Logs · "prod"'));
 assert(result.window.endsWith(' UTC'));assert(!result.window.includes('.373'));
 assert(result.exactWindow.includes(new Date(question.fromMs).toISOString()));assert(result.exactWindow.includes(new Date(question.toMs).toISOString()));assert(result.exactWindow.includes('end excluded'));
 const rolling=scopeSummary({kind:'activity',selector:{signal:'spans',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'unreported'}}}});
 assert.equal(rolling.primary,'All spans');assert.equal(rolling.window,'Rolling 60s');assert(rolling.meta.includes('Unreported environment'));assert.equal(rolling.policy,'Current / legacy');
 const nextDay=scopeSummary({kind:'question',question:{...q.question,fromMs:86399000,toMs:86401000}});
 assert(nextDay.window.includes('1970-01-01'));assert(nextDay.window.includes('1970-01-02'));
});

const deferred=()=>{let resolve,reject;const promise=new Promise((yes,no)=>{resolve=yes;reject=no});return {promise,resolve,reject}};
const flush=()=>new Promise(resolve=>setImmediate(resolve));
function savedFixture(onStage){globalThis.document={hidden:false,createTextNode:t=>new El('text',t)};const host=new El('main');host.root=true;const journal=new Journal({storage:()=>({getItem:()=>null,removeItem(){}}),now:()=>1000}),intent={kind:'saved-view',id:'a'.repeat(64),title:'HTTP errors'};journal.remember(intent);const owner=mount(host,{journal,current:()=>intent,onStage});owner.root.open=true;owner.refresh();const find=text=>walk(owner.root).find(x=>x.tag==='button'&&x.textContent===text);return {host,journal,owner,intent,find,status:()=>walk(owner.root).find(e=>e.tag==='p'),search:()=>walk(owner.root).find(e=>e.tag==='input'&&e.type==='search')};}

test('saved view summaries identify exact revisions without invented query facts',()=>{
 const s=scopeSummary({kind:'saved-view',id:'a'.repeat(64),title:'Checkout comparison'});
 assert.equal(s.primary,'Checkout comparison');assert.equal(s.meta,'Saved view');assert.equal(s.policy,'Exact revision');assert(s.exact.includes('a'.repeat(64)));assert(!s.exact.includes('Rolling'));assert(!s.exact.includes('environment'));
});

test('saved metadata Stage awaits adoption, reports refusal, and never fetches from the picker',async()=>{
 const load=deferred();let supplied;const f=savedFixture((intent,options)=>{supplied={intent,options};return load.promise});
 const stage=f.find('Stage');stage.onclick();assert.equal(stage.disabled,true);assert.equal(stage.textContent,'Loading…');assert.equal(f.owner.root.open,true);assert.equal(f.owner.root.getAttribute('aria-busy'),'true');assert.equal(supplied.options.current(),true);assert.deepEqual(supplied.intent,f.intent);
 load.resolve(true);await flush();assert.equal(f.owner.root.open,false);assert.equal(f.owner.root.getAttribute('aria-busy'),'false');
 const refusal=deferred(),g=savedFixture(()=>refusal.promise);g.find('Stage').onclick();refusal.reject(Error('Saved definition unavailable.'));await flush();assert.equal(g.owner.root.open,true);assert.equal(g.status().textContent,'Saved definition unavailable.');assert.equal(g.find('Stage').disabled,false);
 const keep=deferred(),h=savedFixture(()=>keep.promise);h.find('Stage').onclick();keep.resolve(false);await flush();assert.equal(h.owner.root.open,true);assert.equal(h.status().textContent,'Current investigation kept.');
});

test('local reference filtering, pinning and removal issue no metadata requests and retire old Stage',async()=>{
 let calls=0,signal;const pending=deferred(),f=savedFixture((_i,options)=>{calls++;signal=options.signal;return pending.promise});
 const old=f.find('Stage');old.onclick();assert.equal(calls,1);f.search().value='missing';f.search().oninput();assert.equal(signal.aborted,true);assert.equal(f.find('Stage'),undefined);pending.resolve(true);await flush();assert.equal(f.owner.root.open,true);
 f.search().value='HTTP';f.search().oninput();f.find('Pin').onclick();assert.equal(calls,1);assert.equal(f.journal.list()[0].pinned,true);old.onclick();assert.equal(calls,1);f.find('Forget').onclick();assert.equal(calls,1);assert.equal(f.journal.list().length,0);
});

test('closed, inert, disabled or disposed Stage callbacks cannot acquire saved metadata',async()=>{
 for(const retire of ['closed','inert','disabled','dispose']){
  let calls=0;const f=savedFixture(()=>{calls++;return true}),stage=f.find('Stage');
  if(retire==='closed')f.owner.root.open=false;else if(retire==='inert')f.host.inert=true;else if(retire==='disabled')stage.disabled=true;else f.owner.dispose();stage.onclick();assert.equal(calls,0,retire);
 }
 const pending=deferred();let supplied;const f=savedFixture((_i,options)=>{supplied=options;return pending.promise});f.find('Stage').onclick();f.owner.root.open=false;f.owner.root.listeners.toggle();assert.equal(supplied.signal.aborted,true);pending.resolve(true);await flush();assert.equal(f.owner.root.open,false);
});

test('retired saved successes and errors cannot close or overwrite a newer same-reference check',async()=>{
 const first=deferred(),second=deferred();let count=0;const f=savedFixture(()=>++count===1?first.promise:second.promise);
 f.find('Stage').onclick();f.owner.refresh();f.find('Stage').onclick();assert.equal(count,2);
 first.reject(Error('Old definition error'));await flush();assert.equal(f.status().textContent,'Loading saved view…');assert.equal(f.owner.root.open,true);assert.equal(f.find('Loading…').disabled,true);
 second.resolve(false);await flush();assert.equal(f.status().textContent,'Current investigation kept.');assert.equal(f.find('Stage').disabled,false);
 const stale=deferred(),g=savedFixture(()=>stale.promise);g.find('Stage').onclick();g.owner.root.open=false;g.owner.root.listeners.toggle();g.owner.root.open=true;g.owner.refresh();stale.resolve(true);await flush();assert.equal(g.owner.root.open,true);assert.equal(g.find('Stage').disabled,false);
});

test('native completion withdrawal restores a retryable control without adopting metadata',async()=>{
 for(const withdrawal of ['inert','hidden','fieldset']){
  const load=deferred();let context;const f=savedFixture((_intent,options)=>{context=options;return load.promise});f.find('Stage').onclick();
  if(withdrawal==='inert')f.host.inert=true;else if(withdrawal==='hidden')f.host.hidden=true;else{const outer=new El('fieldset');outer.root=true;outer.disabled=true;outer.append(f.host);}
  assert.equal(context.current(),false,withdrawal);load.resolve(true);await flush();assert.equal(f.owner.root.open,true);assert.equal(f.owner.root.getAttribute('aria-busy'),'false');assert.equal(f.find('Stage').disabled,false);
 }
});

test('another enabled Stage explicitly supersedes a pending reference without reviving its reply',async()=>{
 const a=deferred(),b=deferred(),attempts=[],f=savedFixture((intent,options)=>{attempts.push({intent,options});return intent.id==='a'.repeat(64)?a.promise:b.promise});
 f.journal.remember({kind:'saved-view',id:'b'.repeat(64),title:'Another saved view'});f.owner.refresh();
 const row=id=>walk(f.owner.root).find(e=>e.dataset.scopeId==='saved-view:'+id),button=id=>row(id).children[1].children[0];
 const first=button('a'.repeat(64)),second=button('b'.repeat(64));first.onclick();assert.equal(first.disabled,true);assert.equal(second.disabled,false);second.onclick();assert.equal(attempts.length,2);assert.equal(attempts[0].options.signal.aborted,true);assert.equal(attempts[0].options.current(),false);assert.equal(attempts[1].options.current(),true);assert.equal(first.disabled,false);assert.equal(first.textContent,'Stage');
 a.resolve(true);await flush();assert.equal(f.owner.root.open,true);assert.equal(second.disabled,true);b.resolve(true);await flush();assert.equal(f.owner.root.open,false);
});
