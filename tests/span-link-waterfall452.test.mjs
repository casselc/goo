import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
let shown;
class Element {
 constructor(tag,text=''){this.tag=tag;this.textContent=text;this.children=[];this.dataset={};this.handlers={};this.hidden=false;this.classList={toggle(){}};}
 append(...nodes){for(let n of nodes){if(typeof n==='string')n=new Element('text',n);n.parentElement=this;this.children.push(n)}}
 set data(value){if(this.tag==='goo-trace-waterfall')value.rows.forEach((r,i)=>{const n=new Element('button');n.dataset={rowIndex:String(i),traceId:r.TraceId,spanId:r.SpanId};n.className='waterfall-row';this.append(n)})}
 addEventListener(name,fn){this.handlers[name]=fn}
 querySelectorAll(selector){return this.children.flatMap(n=>[...(selector==='.waterfall-row'&&n.className==='waterfall-row'?[n]:[]),...n.querySelectorAll(selector)])}
 closest(){return this}
 contains(n){return n===this||this.children.some(c=>c.contains(n))}
 focus(){this.focused=true}
 click(){this.clicks=(this.clicks||0)+1}
 get isConnected(){return this.connected!==false&&(!this.parentElement||this.parentElement.isConnected)}
 setAttribute(){}
}
const visible=n=>!!n?.isConnected&&!n.hidden&&(!n.parentElement||visible(n.parentElement));
mock.module('../ui/waterfall.mjs',{exports:{layout:()=>({})}});
mock.module('../ui/components.mjs',{exports:{actionVisible:visible,inspector:()=>({has:()=>false,show:value=>shown=value})}});
mock.module('../ui/record-presentation.mjs',{exports:{describe:()=>({duration:'1 ms',label:'span'}),service:()=>'',attributes(){},timestamp:()=>new Element('time')}});
mock.module('../ui/exception-detail.mjs',{exports:{mount(){}}});
mock.module('../ui/trace-page-breakdown.mjs',{exports:{analyze(){throw Error('No breakdown needed for link ownership test')}}});
mock.module('../ui/trace-page-breakdown-view.mjs',{exports:{}});
const {waterfall}=await import('../ui/visualizations.mjs');
const source={TraceId:'a'.repeat(32),SpanId:'b'.repeat(16),SpanName:'producer',linked_traces:['c'.repeat(32)],linked_spans:['d'.repeat(16)],linked_attributes:[{}]};
const target={TraceId:'c'.repeat(32),SpanId:'d'.repeat(16),SpanName:'consumer'};
const origin={signal:'spans',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0};
function setup(rows,options={}){globalThis.document={createElement:tag=>new Element(tag)};const host=new Element('section'),element=waterfall(host,rows,false,null,'details',{origin,...options});element.handlers.click({target:element.children[0]});return {host,element,actions:shown.actions,content:shown.content};}
test('outside target emits one owner intent without acquiring or selecting a row',()=>{let calls=[];const x=setup([source],{onLinkTarget:v=>calls.push(v)});assert.equal(x.actions[0].textContent,'Look up linked span');x.actions[0].onclick();assert.deepEqual(calls.map(({rowIndex,linkIndex})=>({rowIndex,linkIndex})),[{rowIndex:0,linkIndex:0}]);assert.equal(calls[0].opener,x.actions[0]);assert.equal(x.element.children[0].clicks,undefined)});
test('only a unique returned exact tuple is followed locally',()=>{let calls=0;const x=setup([source,target],{onLinkTarget:()=>calls++});assert.equal(x.actions[0].textContent,'Follow linked span');x.actions[0].onclick();assert.equal(x.element.children[1].clicks,1);assert.equal(calls,0);const ambiguous=setup([source,target,{...target}],{onLinkTarget:()=>calls++});assert.equal(ambiguous.actions[0].textContent,'Look up linked span');assert.match(ambiguous.content.children.at(-1).textContent,/Multiple returned/)});
test('cached action refuses after owner withdrawal, hidden control, or detached origin',()=>{let current=true,calls=0;const x=setup([source],{canAct:()=>current,onLinkTarget:()=>calls++});const button=x.actions[0];current=false;button.onclick();current=true;button.hidden=true;button.onclick();button.hidden=false;x.host.connected=false;button.onclick();assert.equal(calls,0)});
test('malformed references and absent owner have no enabled lookup',()=>{const invalid=setup([{...source,linked_spans:[]}],{onLinkTarget:()=>assert.fail('invalid link escaped')});assert.equal(invalid.actions.length,0);assert.match(invalid.content.children.at(-1).textContent,/unavailable/);const unowned=setup([source],{origin:undefined});assert.equal(unowned.actions.length,0);assert.match(unowned.content.children.at(-1).textContent,/outside returned spans/)});
test('link controls show their position and target; large reference lists are bounded explicitly',()=>{const linked={...source,linked_traces:Array(40).fill(target.TraceId),linked_spans:Array.from({length:40},(_,i)=>(i+1).toString(16).padStart(16,'0')),linked_attributes:Array.from({length:40},()=>({}))};const x=setup([linked],{onLinkTarget(){}});assert.equal(x.actions.length,32);assert.equal(x.actions[0].children[0].textContent,' 1 · 00000001');assert.equal(x.actions[1].children[0].textContent,' 2 · 00000002');assert.equal(x.content.children.at(-1).textContent,'8 additional reported links not displayed.')});
