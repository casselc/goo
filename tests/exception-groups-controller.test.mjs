import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class El{
 constructor(tag,text=''){this.tagName=tag.toUpperCase();this.localName=tag;this._text=text;this.children=[];this.dataset={};this.attributes=new Map();this.hidden=false;this.disabled=false;this.open=false;this.listeners={};this.value='';}
 get textContent(){return this._text+this.children.map(c=>c.textContent).join('');}set textContent(v){this._text=v;this.replaceChildren();}
 get isConnected(){return this.root||!!this.parent?.isConnected;}append(...nodes){for(const node of nodes){node.remove();node.parent=this;this.children.push(node);}}remove(){if(this.parent)this.parent.children=this.parent.children.filter(c=>c!==this);this.parent=null;}replaceChildren(...nodes){for(const node of this.children)node.parent=null;this.children=[];this.append(...nodes);}contains(value){return this===value||this.children.some(c=>c.contains(value));}setAttribute(k,v){this.attributes.set(k,String(v));}getAttribute(k){return this.attributes.get(k)||null;}addEventListener(k,f){this.listeners[k]=f;}
}
const walk=root=>root.children.flatMap(c=>[c,...walk(c)]),text=(root,value)=>walk(root).find(c=>c._text===value),label=(root,value)=>walk(root).find(c=>c.getAttribute('aria-label')===value);
const calls=[];
mock.module('../ui/components.mjs',{exports:{node:(tag,value,cls)=>{const e=new El(tag,value);e.className=cls||'';return e;}}});
mock.module('../ui/queries.mjs',{exports:{postJSON:(path,body,signal)=>new Promise(resolve=>calls.push({path,body,signal,resolve}))}});
mock.module('../ui/dom-visibility.mjs',{exports:{nativeControlReady:node=>{for(let p=node;p;p=p.parent)if(p.hidden||p.disabled||p.inert)return false;return node.isConnected;}}});
const {mount}=await import('../ui/exception-groups.mjs');
const q={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0},id='ef1:c:'+'a'.repeat(64),read={readAt:'2026-11-13T00:00:00Z',basis:'a'.repeat(64)};
function response(){return{ok:true,data:{version:'goo/exception-groups/v1',question:structuredClone(q),count:'150',state:'observed',capturedAt:read.readAt,basis:read.basis,reads:{count:{...read},groups:{...read}},warnings:[],groups:[{fingerprint:id,state:'classified',reason:'',count:'150',firstNS:'1000000000',lastNS:'60999999999',inputs:{type:'CheckoutError',stacktrace:'frame\nline'},affected:[{service:'api',namespace:null,version:''}],affectedCount:'1',affectedLimited:false}]}};}
function setup(){const host=new El('main');host.root=true;let active=true;const staged=[],owner=mount(host,{current:()=>active,onStage:(...args)=>staged.push(args)});owner.set(q);owner.root.open=true;return{host,owner,staged,retire:()=>{active=false;owner.sync();}};}

test('exception owner reads explicitly, stages exact group locally, and retires replaced controls',async()=>{
 const {host,owner,staged,retire}=setup(),before=calls.length,read=text(host,'Read exception groups');assert.equal(calls.length,before);const promise=read.onclick();assert.equal(calls.length,before+1);read.onclick();assert.equal(calls.length,before+1);calls.at(-1).resolve(response());await promise;assert.equal(owner.root.dataset.state,'observed');const old=text(host,'Stage group records');old.onclick();assert.equal(staged.length,1);assert.equal(staged[0][0].where.value,id);assert.equal(calls.length,before+1);
 label(host,'Find exception groups').value='Checkout';label(host,'Find exception groups').oninput();owner.root.append(old);old.onclick();assert.equal(staged.length,1);const current=walk(host).find(c=>c._text==='Stage group records'&&c!==old);owner.root.open=false;current.onclick();assert.equal(staged.length,1);owner.root.open=true;retire();current.onclick();assert.equal(staged.length,1);owner.dispose();
});

test('exception owner aborts changed and closed reads; failure withdraws old groups and receipts',async()=>{
 const {host,owner}=setup(),read=text(host,'Read exception groups');let pending=read.onclick(),first=calls.at(-1);owner.set({...q,toMs:62000});assert.equal(first.signal.aborted,true);first.resolve(response());await pending;assert.notEqual(owner.root.dataset.state,'observed');owner.set(q);pending=read.onclick();const second=calls.at(-1);owner.root.open=false;owner.root.listeners.toggle();assert.equal(second.signal.aborted,true);second.resolve(response());await pending;assert.equal(walk(host).filter(c=>c._text==='Stage group records').length,0);
 owner.root.open=true;pending=read.onclick();calls.at(-1).resolve(response());await pending;assert.equal(owner.root.dataset.state,'observed');pending=read.onclick();calls.at(-1).resolve({ok:false,data:{error:'Source unavailable'}});await pending;assert.equal(owner.root.dataset.state,'unavailable');assert.equal(walk(host).filter(c=>c._text==='Stage group records').length,0);owner.dispose();
});

test('exception search matches displayed unclassified reasons without reads and row descriptions are instance-unique',async()=>{
 const a=setup(),b=setup();let pending=text(a.host,'Read exception groups').onclick();const data=response();data.data.groups[0]={...data.data.groups[0],fingerprint:'ef1:u:missing-stack',state:'unclassified',reason:'missing-stack',inputs:null};calls.at(-1).resolve(data);await pending;pending=text(b.host,'Read exception groups').onclick();calls.at(-1).resolve(data);await pending;
 assert.notEqual(text(a.host,'Stage group records').getAttribute('aria-describedby'),text(b.host,'Stage group records').getAttribute('aria-describedby'));
 const before=calls.length,find=label(a.host,'Find exception groups');for(const value of ['Stack not reported','Unclassified']){find.value=value;find.oninput();assert(text(a.host,'Stage group records'));}assert.equal(calls.length,before);a.owner.dispose();b.owner.dispose();
});

test('same exception type has a distinct literal stack cue and contextual stage description',async()=>{
 const {host,owner}=setup(),pending=text(host,'Read exception groups').onclick(),data=response();data.data.groups[0].count='75';data.data.groups.push({...structuredClone(data.data.groups[0]),fingerprint:'ef1:c:'+'b'.repeat(64),inputs:{type:'CheckoutError',stacktrace:'checkout\n  cancel:99'}});calls.at(-1).resolve(data);await pending;
 const cues=walk(host).filter(e=>e.className==='exception-group-kind');assert.equal(cues.length,2);assert.notEqual(cues[0].textContent,cues[1].textContent);assert(cues.some(c=>c.textContent.includes('cancel:99')));const summaries=walk(host).filter(e=>e._text==='Signature details');assert.equal(new Set(summaries.map(e=>e.getAttribute('aria-label'))).size,2);for(const summary of summaries)assert.match(summary.getAttribute('aria-label'),/^Signature details · CheckoutError · [ab]{8}$/);for(const stage of walk(host).filter(e=>e._text==='Stage group records'))assert.match(stage.getAttribute('aria-describedby'),/context/);owner.dispose();
});
