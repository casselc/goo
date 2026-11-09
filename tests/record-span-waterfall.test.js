const {test}=require('node:test'),assert=require('node:assert/strict');
const W=require('../ui/waterfall.mjs'),R=require('../ui/record-span-waterfall.mjs'),P=require('../ui/record-page.mjs');
const trace='1'.repeat(32),other='2'.repeat(32),id=n=>n.toString(16).padStart(16,'0');
const span=(n,start,duration,parent='',t=trace)=>({TraceId:t,SpanId:id(n),ParentSpanId:parent,timestamp_ns:String(start),duration_ns:String(duration),ServiceName:'api',SpanName:'span '+n,Attributes:{empty:'',nil:null,literal:'Missing'},ResourceAttributes:{empty:'resource'}});
const source=rows=>({question:{signal:'spans',environment:'prod',fromMs:1000,toMs:2000,limit:100,offset:0},basis:'a'.repeat(64),capturedAt:'2026-11-01T00:00:00Z',state:'observed',rows});
const options={columns:['utc','service','summary','context'],fieldColumns:[{kind:'attribute',key:'empty'},{kind:'resource',key:'empty'},{kind:'attribute',key:'nil'},{kind:'attribute',key:'absent'}]};
test('positive sub-microsecond values remain distinct from a reported zero in timing labels',()=>{assert.equal(W.durationLabel(0n),'0.000 ms');assert.equal(W.durationLabel(1n),'<0.001 ms');assert.equal(W.durationLabel(999n),'<0.001 ms');assert.equal(W.durationLabel(1000n),'0.001 ms');assert.equal(W.durationLabel(null),'Duration unavailable');});

test('trace-relative timing removes intertrace idle gaps but retains one comparable duration scale',()=>{
 const rows=[span(2,1000001n,1n,id(1)),span(1,1000000n,5000000n),span(3,100000000000n,500000000n,'',other)],before=structuredClone(rows),m=W.layout(rows,{origin:'trace-relative'});
 assert.equal(m.scaleExtent,500000000n);assert.deepEqual(m.spans.map(s=>s.index),[1,0,2]);assert.equal(m.spans[0].width,1);assert.equal(m.spans[2].width,100);assert.equal(m.spans[0].left,0);assert.equal(m.spans[2].left,0);assert.equal(m.spans[1].depth,1);
 const moved=W.layout(rows.map(r=>({...r,timestamp_ns:String(BigInt(r.timestamp_ns)+1700000000000000000n)})),{origin:'trace-relative'});assert.deepEqual(m.spans.map(s=>[s.index,s.left,s.width]),moved.spans.map(s=>[s.index,s.left,s.width]));assert.deepEqual(rows,before);
 const global=W.layout(rows);assert.equal(global.originMode,'global');assert(global.spans.find(s=>s.index===2).left>90);
});
test('parent multimap never invents edges for duplicate, missing or invalid identities; cycles stay inspectable',()=>{
 const rows=[span(1,0,4),span(1,1,3),span(2,2,1,id(1)),span(3,3,1,id(8)),span(4,4,1,id(5)),span(5,5,1,id(4)),span(6,6,1,id(1),'invalid'),span(7,7,1,id(6),'invalid')],m=W.layout(rows,{origin:'trace-relative'}),byIndex=new Map(m.spans.map(s=>[s.index,s]));
 assert.equal(byIndex.get(2).parentIndex,null);assert.equal(byIndex.get(2).ambiguousParent,true);assert.equal(byIndex.get(3).missingParent,true);assert.equal(byIndex.get(4).cycle,true);assert.equal(byIndex.get(5).cycle,true);assert.equal(byIndex.get(6).parentIndex,null);assert.notEqual(byIndex.get(6).groupKey,byIndex.get(7).groupKey);assert.equal(m.spans.length,rows.length);assert.equal(new Set(m.spans.map(s=>s.index)).size,rows.length);
 assert.match(W.parentLabel(byIndex.get(2)),/Ambiguous/);assert.match(W.parentLabel(byIndex.get(0)),/Duplicate/);
 for(const s of m.spans){assert(Number.isFinite(s.left));assert(Number.isFinite(s.width));assert(s.left>=0&&s.left+s.width<=100)}
});
test('unknown timing remains an observation with unavailable geometry, while zero duration stays known zero',()=>{
 const rows=[{...span(1,5,1),duration_ns:null},{...span(2,6,1),timestamp_ns:'-1'},span(3,7,0,'',other)],m=W.layout(rows,{origin:'trace-relative'}),byIndex=new Map(m.spans.map(s=>[s.index,s]));
 assert.equal(byIndex.get(0).timingValid,false);assert.equal(byIndex.get(0).left,null);assert.equal(byIndex.get(0).width,null);assert.equal(byIndex.get(2).timingValid,true);assert.equal(byIndex.get(2).width,0);assert.equal(m.groups[0].origin,null);assert.equal(m.groups[0].extent,null);assert.equal(m.scaleExtent,0n);
});
test('waterfall page model retains all qualified fields and custody but never adds completeness or a population total',()=>{
 const page=P.fromExploration(source([span(1,1000,5000000)])),m=R.waterfallModel(page,options);
 assert.equal(m.page,page);assert.equal(m.rows[0].index,0);assert.deepEqual(m.rows[0].fields.map(f=>[f.state,f.text]),[['empty','Present empty string'],['present','resource'],['null','Null'],['missing','Missing']]);assert(!Object.hasOwn(m,'count'));assert(!Object.hasOwn(m,'completeness'));assert.equal(m.geometry.originMode,'trace-relative');
 assert.throws(()=>R.waterfallModel(P.fromExploration({...source([]),question:{...source([]).question,signal:'logs'}})),/spans query/);
 const projected=R.timingRecord({TraceId:'x'.repeat(100000),SpanId:{huge:'x'.repeat(100000)},timestamp_ns:'9'.repeat(100000),ServiceName:'x'.repeat(100000),Attributes:{large:'x'.repeat(100000)}});assert(JSON.stringify(projected).length<1000);assert(!Object.hasOwn(projected,'Attributes'));assert.equal(W.layout([projected]).spans[0].timingValid,false);
});
test('passive mounting hosts keep original positions despite visual order, remain inert and retire on replacement',()=>{
 const previous=global.document;
 class Element{
  constructor(tag){this.tag=tag;this.children=[];this.dataset={};this.attributes=new Map();this.style={};this.isConnected=true}
  get childNodes(){return this.children}
  append(...nodes){this.children.push(...nodes)}
  replaceChildren(...nodes){this.children=nodes}
  setAttribute(k,v){this.attributes.set(k,v)}
 }
 global.document={createElement:tag=>new Element(tag)};
 try{
  const list=new R.RecordSpanWaterfall();list.isConnected=false;list.dataset={};list.getAttribute=()=>null;list.replaceChildren=(...nodes)=>{list.nodes=nodes};list.append=(...nodes)=>list.nodes.push(...nodes);
  const page=P.fromExploration(source([span(2,1001,1,id(1)),span(1,1000,4),{...span(3,1002,1),duration_ns:null}]));list.page=page;list.display={columns:['utc','summary']};
  const hosts=list.recordItems();assert.equal(hosts.length,3);assert(Object.isFrozen(hosts));assert.deepEqual(hosts.map(h=>h.index),[0,1,2]);assert(Object.isFrozen(hosts[0]));
  const section=list.nodes.find(n=>n.className==='record-span-waterfall-group');assert.equal(section.children[0].children[0].textContent,'Trace '+trace.slice(0,8)+'…'+trace.slice(-8));assert.equal(section.children[0].children[0].title,trace);assert.equal(section.children[1],hosts[1].element);assert.equal(section.children[2],hosts[0].element);
  for(const h of hosts){assert.equal(h.actionsHost.hidden,true);assert.deepEqual(h.actionsHost.children,[]);assert.equal(h.detailHost.hidden,true);assert.equal(h.timingButton.tag,'button');assert.equal(h.actionsHost.tag,'div');assert(!h.timingButton.children.includes(h.actionsHost));assert(!h.timingButton.attributes.get('aria-label').includes('api'),'service hidden in bar label')}
  assert.equal(hosts[2].timingButton.dataset.timingState,'unavailable');list.isConnected=true;list.connectedCallback();assert.equal(list.recordItems(),hosts,'connection preserves explicitly mounted hosts');hosts[0].timingButton.onclick();assert.equal(hosts[0].details.open,true);
  list.isConnected=false;list.display=options;const fresh=list.recordItems();assert.notEqual(fresh,hosts);assert.notEqual(fresh[0].element,hosts[0].element);list.page={invalid:true};list.render();assert.deepEqual(list.recordItems(),[]);assert.equal(list.dataset.state,'unavailable');assert.equal(list.dataset.signal,undefined);
 }finally{global.document=previous}
});
test('serialized attributes and detached property replacement use the same admitted page and display contract',()=>{
 const list=new R.RecordSpanWaterfall(),attrs=new Map();list.isConnected=false;list.getAttribute=k=>attrs.get(k)||null;const page=P.fromExploration(source([span(1,0,0)]));
 attrs.set('page',JSON.stringify(page));attrs.set('display',JSON.stringify(options));assert.deepEqual(list.page,page);assert.deepEqual(list.display,options);attrs.set('page','{broken');list.attributeChangedCallback('page');assert.equal(list.page,null);list.page=page;assert.equal(list.page,page);list.display={columns:['summary']};assert.equal(list.display,null);
});

test('summary-provided is a passive SSR/detached display option; standalone facts remain available',()=>{const previous=global.document;class Element{constructor(tag){this.tag=tag;this.children=[];this.dataset={};this.attributes=new Map();this.style={};this.isConnected=true}get childNodes(){return this.children}append(...nodes){this.children.push(...nodes)}replaceChildren(...nodes){this.children=nodes}setAttribute(k,v){this.attributes.set(k,v)}}global.document={createElement:tag=>new Element(tag)};try{const attrs=new Map(),list=new R.RecordSpanWaterfall();list.isConnected=false;list.dataset={};list.getAttribute=k=>attrs.has(k)?attrs.get(k):null;list.setAttribute=(key,value)=>{attrs.set(key,value);list.attributeChangedCallback(key)};list.removeAttribute=key=>{attrs.delete(key);list.attributeChangedCallback(key)};list.replaceChildren=(...nodes)=>{list.nodes=nodes};list.append=(...nodes)=>list.nodes.push(...nodes);list.page=P.fromExploration(source([span(1,1000,4),span(2,1001,1,id(1))]));list.display=options;const standalone=list.recordItems();assert(list.nodes.some(node=>node.className==='record-span-waterfall-facts'));const page=list.page;list.summaryProvided=true;assert.equal(attrs.get('summary-provided'),'');const embedded=list.recordItems();assert.equal(list.page,page);assert.notEqual(embedded,standalone);assert(!list.nodes.some(node=>node.className==='record-span-waterfall-facts'));assert(list.nodes.some(node=>node.className==='record-span-waterfall-scale'&&node.textContent.includes('Relative trace time')&&node.textContent.includes('coverage unknown')));assert.equal(list.nodes.at(-1).children[0].textContent,'Timing scope');assert.equal(embedded.length,standalone.length);list.summaryProvided=false;const restored=list.recordItems();assert.equal(list.page,page);assert.notEqual(restored,embedded);assert(list.nodes.some(node=>node.className==='record-span-waterfall-facts'));attrs.set('summary-provided','');list.attributeChangedCallback('summary-provided');list.isConnected=true;list.connectedCallback();assert.equal(list.summaryProvided,true);assert(!list.nodes.some(node=>node.className==='record-span-waterfall-facts'))}finally{global.document=previous}});
