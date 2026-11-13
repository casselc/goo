import {test} from 'node:test';
import assert from 'node:assert/strict';
class Element {
 constructor(tag,text,cls){this.tag=tag;this.textContent=text??'';this.className=cls||'';this.children=[];this.parent=null;this.dataset={};this.attributes=new Map();this.events=new Map();this.hidden=false;this.open=false;this.value='';}
 get isConnected(){return this.connected===true||!!this.parent?.isConnected}get firstChild(){return this.children[0]}get lastChild(){return this.children.at(-1)}
 append(...nodes){for(const node of nodes){node.remove();node.parent=this;this.children.push(node)}}replaceChildren(...nodes){for(const child of this.children)child.parent=null;this.children=[];this.append(...nodes)}remove(){if(this.parent){this.parent.children=this.parent.children.filter(node=>node!==this);this.parent=null}}
 setAttribute(k,v){this.attributes.set(k,String(v))}removeAttribute(k){this.attributes.delete(k)}getClientRects(){return this.isConnected&&!this.closest('[hidden],[inert]')?[{}]:[]}closest(){for(let e=this;e;e=e.parent)if(e.hidden||e.attributes.has('inert'))return e;return null}
 addEventListener(type,callback){const set=this.events.get(type)||new Set();set.add(callback);this.events.set(type,set)}removeEventListener(type,callback){this.events.get(type)?.delete(callback)}emit(type,detail){for(const callback of this.events.get(type)||[])callback({detail})}
}
function dom(run){const old=globalThis.document;globalThis.document={createElement:tag=>new Element(tag)};try{return run()}finally{globalThis.document=old}}
const V=await import('../ui/comparison-visual.mjs'),E=await import('../ui/evidence-scope.mjs');
const q=(fromMs)=>({signal:'logs',environment:'prod',fromMs,toMs:fromMs+60000,limit:2,offset:0,groupBy:[{kind:'column',key:'service'}]});
function fixture(a='5',b='9'){
 const side=(role,count,from)=>({version:'goo.dataset/1',id:role,revision:role,origin:'browser-held',sourceBasis:'basis-'+role,capturedAt:'2026-11-04T01:02:03.123456789Z',shape:'summary',scope:{id:role,generation:role,signal:'logs',selector:q(from)},quality:{extent:'unknown',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},fields:[{key:'count',label:'Captured records',type:'number',unit:'records'}],rows:[{key:'population',values:{count}}]});
 const comparison={version:'goo.comparison/1',state:a==='0'||b==='0'?'no_data':'observed',baseline:side('baseline',a,60000),selection:side('selected',b,120000),changes:{count:{state:'derived',value:String(BigInt(b)-BigInt(a)),unit:'records'}}};
 const sides=['Previous window','Selected window'].map((label,index)=>{const d=comparison[index?'selection':'baseline'];return {label,data:E.comparisonContextSide(d.scope.selector,{mode:index?'fixed':'frozen',state:'observed',count:d.rows[0].values.count,readAt:d.capturedAt,basis:d.sourceBasis,...(!index?{capturedAt:'2026-11-04T02:00:00Z'}:{})})}});return {comparison,sides};
}
test('shared facts retain exact uint counts, distinct fixed windows, source clocks and custody without mutating inputs',()=>{const value=fixture('18446744073709551614','18446744073709551615'),before=structuredClone(value),facts=V.countPairFacts(value);assert.equal(facts.valid,true);assert.deepEqual(facts.sides.map(side=>[side.count,side.fromMs,side.toMs,side.readAt]),[['18446744073709551614',60000,120000,'2026-11-04T01:02:03.123456789Z'],['18446744073709551615',120000,180000,'2026-11-04T01:02:03.123456789Z']]);assert.ok(facts.sides[0].data.details.some(row=>row.label==='Stored custody time'));facts.sides[0].data.details[0].value='changed';assert.deepEqual(value,before);assert.equal(V.countPairFacts(fixture('0','9')).state,'no_data');});
test('malformed count facts refuse unavailable context without promoting supplied claims',()=>{for(const value of [null,{}, {...fixture(),sides:[]}, {...fixture(),sides:[{label:'',data:fixture().sides[0].data},fixture().sides[1]]}, {...fixture(),sides:[{label:'x'.repeat(121),data:fixture().sides[0].data},fixture().sides[1]]}])assert.equal(V.countPairFacts(value).valid,false);for(const mutate of [v=>v.sides[0].data.state='healthy',v=>v.comparison.baseline.capturedAt='not-a-clock',v=>v.comparison.baseline.scope.selector.fromMs=-1,v=>{v.comparison.baseline.scope.selector.fromMs=Number.MAX_SAFE_INTEGER-60000;v.comparison.baseline.scope.selector.toMs=Number.MAX_SAFE_INTEGER},v=>v.comparison.baseline.rows[0].values.count='-1']){const value=fixture();mutate(value);assert.equal(V.countPairFacts(value).valid,false);}});
test('both origins share Summary/Groups with host-owned callbacks, cached nodes and native disclosure state; zero IO',()=>dom(()=>{const previousFetch=globalThis.fetch;globalThis.fetch=()=>{throw Error('Passive count pair must not fetch')};try{for(const labels of [['Rest of parent','Selected'],['Previous window','Selected window'],['Frozen baseline','Current held'],['Held parent (overlapping)','Selected']]){const input=fixture();input.sides.forEach((side,index)=>side.label=labels[index]);const host=new Element('main');host.connected=true;const nodes={},changes=[];let current=true;const render=(view='summary')=>V.countPair(host,{...input,groups:{held:'supplied-group-model'},nodes,view,current:()=>current,onDisplay:change=>changes.push(change)});const first=render();assert.deepEqual(first.summary.display.labels,labels);assert.deepEqual(first.groups.display.labels,labels);assert.equal(first.summary.hidden,false);assert.equal(first.groups.hidden,true);nodes.chrome.scope.open=true;const select=first.viewSelect;select.value='groups';select.onchange();assert.equal(first.groups.hidden,false);assert.deepEqual(changes,[{view:'groups'}]);first.groups.emit('goo:groupdisplaychange',{order:'label'});assert.deepEqual(changes.at(-1),{order:'label'});first.groups.emit('goo:groupdisplaychange',null);assert.equal(changes.length,2);current=false;select.value='summary';select.onchange();assert.equal(select.value,'groups');assert.equal(changes.length,2);current=true;const second=render('groups');assert.equal(second.root,first.root);assert.equal(second.summary,first.summary);assert.equal(second.groups,first.groups);assert.equal(second.viewSelect,select);assert.equal(nodes.chrome.scope.open,true);host.remove();host.connected=false;select.value='summary';select.onchange();assert.equal(changes.length,2);}}finally{globalThis.fetch=previousFetch}}));
test('scope disclosure renders UTC windows and read clocks through the existing fact formatter',()=>dom(()=>{const input=fixture(),host=new Element('main');host.connected=true;const shown=V.countPair(host,input),rows=shown.nodes.chrome.scopeBody.children[0].children[1].children,window=rows.find(row=>row.children[0].textContent==='Window · UTC').children[1].firstChild,read=rows.find(row=>row.children[0].textContent==='Source read').children[1].firstChild;assert.match(window.textContent,/UTC$/);assert.match(window.textContent,/1970-01-01 00:01:00/);assert.equal(window.title,'[60000000000, 120000000000) ns since Unix epoch');assert.equal(read.tag,'time');assert.equal(read.dateTime,'2026-11-04T01:02:03.123456789Z');assert.match(read.textContent,/01:02:03\.123 UTC/);}));
test('missing source read clock remains explicitly unavailable without discarding exact counts',()=>dom(()=>{const input=fixture();delete input.comparison.baseline.capturedAt;const facts=V.countPairFacts(input);assert.equal(facts.valid,true);assert.equal(facts.sides[0].count,'5');assert.equal(facts.sides[0].readAt,null);const host=new Element('main');host.connected=true;const shown=V.countPair(host,input);assert.equal(shown.nodes.chrome.operands.firstChild.lastChild.textContent,'Read time unavailable');}));
test('shared count wrapper stays neutral and selected role labels are not duplicated',()=>dom(()=>{const host=new Element('main');host.connected=true;for(const label of ['Selected','Selected window']){const input=fixture();input.sides[1].label=label;const rendered=V.countPair(host,input);assert.equal(rendered.root.tag,'div');assert.equal(rendered.root.attributes.get('role'),'group');assert.equal(rendered.nodes.chrome.operands.lastChild.firstChild.textContent,label);assert.equal(rendered.nodes.chrome.operands.firstChild.firstChild.textContent,'Previous window');}}));

test('passive role labels copy bounded literal names and reject accessors without executing them',async()=>{
 const {comparisonRoleLabels}=await import('../ui/comparison-role-labels.mjs');assert.deepEqual(comparisonRoleLabels(),['Operand A','Operand B']);
 const labels=['Previous window','Selected window'],copy=comparisonRoleLabels(labels);labels[0]='changed';assert.equal(copy[0],'Previous window');assert(Object.isFrozen(copy));
 for(const value of [null,[],['one'],['one','two','three'],['','two'],['one','bad\nlabel'],['a'.repeat(121),'two']])assert.throws(()=>comparisonRoleLabels(value));
 let reads=0;const accessor=['one','two'];Object.defineProperty(accessor,0,{get(){reads++;return 'bad'}});assert.throws(()=>comparisonRoleLabels(accessor));assert.equal(reads,0);
});
test('mean summary shares independent operand facts and exact supplied support without offering category means',()=>dom(()=>{
 const input=fixture('9007199254740993','3');
 for(const [index,side]of [input.comparison.baseline,input.comparison.selection].entries()){
  side.scope.signal='spans';side.scope.selector.signal='spans';side.scope.selector.summary='span-duration';
  side.fields.push({key:'duration.mean',label:'Mean duration',type:'rational',unit:'ns',aggregation:{family:'duration',kind:'mean',method:'exact-rational'}});
  side.rows[0].values['duration.mean']=index?'0':'9007199254740993/2';side.rows[0].cells={'duration.mean':{state:'derived',support:{total:side.rows[0].values.count,valid:side.rows[0].values.count,missing:'0',empty:'0',invalid:'0'}}};
 }
 input.comparison.changes['duration.mean']={state:'derived',value:'-9007199254740993/2',unit:'ns'};
 const host=new Element('main');host.connected=true;const calls=[];const shown=V.countPair(host,{...input,value:'duration.mean',groups:{supplied:true},onDisplay:value=>calls.push(value)});
 assert.equal(shown.root.attributes.get('aria-label'),'Mean duration comparison');assert.deepEqual(shown.summary.display.fields,['duration.mean']);assert.equal(shown.groups,null);assert.equal(shown.nodes.chrome.controls.hidden,true);assert.equal(shown.nodes.chrome.operands.firstChild.children[1].textContent,'9007199254740993 duration contributors');assert.equal(shown.nodes.chrome.operands.lastChild.children[1].textContent,'3 duration contributors');
 shown.viewSelect.value='groups';shown.viewSelect.onchange();assert.deepEqual(calls,[]);assert.equal(shown.viewSelect.value,'summary');
}));

test('result-first pair has one collapsed source context and preserves exact supplied identity without acquiring data',()=>dom(()=>{
 const input=fixture('132','168'),before=structuredClone(input),host=new Element('main');host.connected=true;
 const shown=V.countPair(host,input),chrome=shown.nodes.chrome;
 assert.equal(chrome.scope.open,false);
 assert.equal(chrome.operands.parent,chrome.scope);
 assert.equal(shown.root.children.includes(chrome.operands),false);
 assert.equal(shown.summary.attributes.has('data-first'),true);
 assert.equal(shown.summary.attributes.has('external-context'),true);
 const rows=chrome.scopeBody.firstChild.children[1].children;
 for(const [label,value]of [['Dataset','baseline'],['Revision','baseline'],['Origin','browser-held'],['Extent','unknown'],['Sampling','unknown'],['Source basis','basis-baseline']]){
  const row=rows.find(row=>row.children[0].textContent===label);assert(row,label);assert.equal(row.children[1].textContent,value);
 }
 assert.equal(chrome.notice.textContent,'Independent populations · upstream coverage unknown.');
 assert.deepEqual(input,before);
}));

test('one empty captured operand scopes its notice rather than labeling the whole comparison No data',()=>dom(()=>{
 const input=fixture('0','9'),before=structuredClone(input),host=new Element('main');host.connected=true;const shown=V.countPair(host,input);
 assert.equal(shown.nodes.chrome.notice.textContent,'Previous window has no records · Independent populations · upstream coverage unknown.');assert.equal(shown.nodes.chrome.notice.dataset.state,'observed');assert.deepEqual(input,before);assert.equal(input.comparison.state,'no_data');
}));
