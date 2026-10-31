const {test}=require('node:test'),assert=require('node:assert/strict');
const {fromExploration}=require('../ui/record-page.mjs'),{displayOptions,listModel,fieldValue,RecordList,contextRecord,contextRecords}=require('../ui/record-list.mjs');
const response=(signal='spans')=>({question:{signal,environment:'prod',fromMs:1000,toMs:2000,limit:2,offset:2},basis:'a'.repeat(64),capturedAt:'2026-10-31T12:00:00Z',state:'observed',rows:[{timestamp_ns:'1793448000123456789',ServiceName:'api',SpanName:'work',Body:'message',duration_ns:'9000000',Attributes:{empty:'',nil:null,literal:'Missing',nested:{enabled:false}},ResourceAttributes:{empty:'resource-value'}}]});
const layout=()=>({columns:['utc','service','summary','context'],fieldColumns:[{kind:'attribute',key:'empty'},{kind:'resource',key:'empty'},{kind:'attribute',key:'nil'},{kind:'attribute',key:'absent'},{kind:'attribute',key:'literal'},{kind:'attribute',key:'nested'}]});
test('record list keeps rich page custody separate from counts and projects exact qualified value states',()=>{
 const raw=response(),page=fromExploration(raw),display=layout(),before=structuredClone({raw,display}),model=listModel(page,display);
 assert.equal(model.page,page);assert.equal(model.rows[0].index,0);assert.equal(model.rows[0].description.duration,'9.000 ms');assert.equal(model.rows[0].description.time.exact,'2026-10-31T12:00:00.123456789Z');
 assert.deepEqual(model.rows[0].fields.map(x=>[x.state,x.text]),[['empty','Present empty string'],['present','resource-value'],['null','Null'],['missing','Missing'],['present','\"Missing\"'],['present','{"enabled":false}']]);
 assert.equal(model.page.page.returned,1);assert.equal(model.page.page.hasMore,null);assert(!Object.hasOwn(model,'count'));assert.deepEqual({raw,display},before);
 assert.equal(fieldValue(page.rows[0],{kind:'resource',key:'nil'}).state,'missing');
});
test('list display options reject invalid or accessor-bearing configuration before reading it',()=>{
 for(const value of [{columns:['summary']},{columns:['utc','summary'],fieldColumns:null},{columns:['utc','summary'],fieldColumns:[{kind:'attribute',key:'x',extra:true}]},{columns:['utc','summary'],data:[]},{columns:['utc','summary'],fieldColumns:[{kind:'attribute',get key(){throw Error('getter invoked')}}]}])assert.throws(()=>displayOptions(value));
 const list=['utc','summary'];Object.defineProperty(list,'0',{enumerable:true,get(){throw Error('getter invoked')}});assert.throws(()=>displayOptions({columns:list}),/plain/);
 const config=layout(),admitted=displayOptions(config);config.columns.reverse();config.fieldColumns[0].key='changed';assert.equal(admitted.columns[0],'utc');assert.equal(admitted.fieldColumns[0].key,'empty');assert(Object.isFrozen(admitted.fieldColumns[0]));
 assert.throws(()=>listModel(fromExploration(response('logs')),{columns:['utc','summary'],fieldColumns:[{kind:'column',key:'name'}]}));
});
test('standalone properties support detached updates, SSR attributes and invalid input withdrawal',()=>{
 const element=new RecordList(),attrs=new Map();element.isConnected=false;element.getAttribute=key=>attrs.get(key)||null;element.replaceChildren=()=>{};
 const page=fromExploration(response());element.display=layout();element.page=page;assert.equal(element.page,page);assert.equal(element.display.columns[0],'utc');
 element.page={...page,count:'10'};assert.equal(element.page,null);element.display=layout();assert.equal(element.page,null,'display update cannot resurrect prior page');
 attrs.set('page',JSON.stringify(page));element.attributeChangedCallback('page');assert.deepEqual(element.page,page);attrs.set('display',JSON.stringify(layout()));element.attributeChangedCallback('display');assert.deepEqual(element.display,layout());
 attrs.set('page','{broken');element.attributeChangedCallback('page');assert.equal(element.page,null);attrs.delete('page');element.attributeChangedCallback('page');assert.equal(element.page,null);
 element.page=page;element.display={columns:['summary']};assert.equal(element.display,null);element.display={columns:['utc','summary']};assert.deepEqual(element.display,{columns:['utc','summary'],fieldColumns:[]});
});


test('invalid replacement withdraws old observed metadata as well as the rendered records',()=>{
 const oldDocument=global.document,element=new RecordList(),attrs=new Map();element.isConnected=false;element.getAttribute=key=>attrs.get(key)||null;element.dataset={signal:'spans',state:'observed'};element.replaceChildren=(...nodes)=>{element.nodes=nodes};element.append=(...nodes)=>element.nodes.push(...nodes);
 global.document={createElement:tag=>({tag,textContent:'',setAttribute(){}})};
 try{element.page=fromExploration(response());element.display={columns:['utc','summary']};element.page={bad:true};element.render();assert.equal(element.dataset.state,'unavailable');assert.equal(element.dataset.signal,undefined);assert.equal(element.nodes.length,1);assert.match(element.nodes[0].textContent,/record-page/);}finally{global.document=oldDocument}
});


test('compact ancestry inputs contain bounded identifiers only, not repeated rich page payloads',()=>{
 const R=require('../ui/record-presentation.mjs'),trace='a'.repeat(32),root={TraceId:trace,SpanId:'1'.repeat(16),ParentSpanId:'',Attributes:{large:'x'.repeat(100000)},Body:'body'},child={...root,SpanId:'2'.repeat(16),ParentSpanId:root.SpanId},rows=[root,child],before=structuredClone(rows),projected=contextRecords(rows);
 assert.deepEqual(projected,[{TraceId:trace,SpanId:root.SpanId,ParentSpanId:''},{TraceId:trace,SpanId:child.SpanId,ParentSpanId:root.SpanId}]);assert(JSON.stringify(projected).length<300);assert.deepEqual(rows,before);
 assert.deepEqual(R.ancestry(child,rows),R.ancestry(contextRecord(child),projected));
 for(const parent of [null,'','0'.repeat(16),'broken','a'.repeat(100000),{large:'x'.repeat(100000)},['0'.repeat(16)],{toString:'0'.repeat(16)}]){const record={...child,ParentSpanId:parent};assert.deepEqual(R.ancestry(record,[root,record]),R.ancestry(contextRecord(record),contextRecords([root,record])));assert(JSON.stringify(contextRecord(record)).length<200);}
 const bad={TraceId:'a'.repeat(100000),SpanId:{large:'x'.repeat(100000)},ParentSpanId:['x'.repeat(100000)]};const compact=contextRecord(bad);assert(JSON.stringify(compact).length<100);assert.deepEqual(R.ancestry(bad,[bad]),R.ancestry(compact,[compact]));
 const duplicates=[root,{...root},child];assert.deepEqual(R.ancestry(child,duplicates),R.ancestry(contextRecord(child),contextRecords(duplicates)));
});

// Exercise the actual passive render without a browser/custom-element registry.
// The application owns any listeners later placed into the returned inert slots.
test('record interaction hosts are inert, page-local and withdrawn on replacement',()=>{
 const oldDocument=global.document;
 class Element{
  constructor(tag){this.tag=tag;this.children=[];this.dataset={};this.attributes=new Map();this.isConnected=true}
  get childNodes(){return this.children}
  append(...nodes){this.children.push(...nodes)}
  replaceChildren(...nodes){this.children=nodes}
  setAttribute(key,value){this.attributes.set(key,value)}
 }
 global.document={createElement:tag=>new Element(tag)};
 try{
  const list=new RecordList();list.isConnected=false;list.dataset={};list.getAttribute=()=>null;list.replaceChildren=(...nodes)=>{list.nodes=nodes};list.append=(...nodes)=>list.nodes.push(...nodes);
  const page=fromExploration(response());list.page=page;list.display={columns:['utc','summary']};list.render();
  const hosts=list.recordItems();assert.equal(hosts.length,1);assert.equal(hosts[0].index,0);assert.equal(hosts[0].element.dataset.recordIndex,'0');assert(Object.isFrozen(hosts));assert(Object.isFrozen(hosts[0]));
  for(const host of [hosts[0].actionsHost,hosts[0].detailHost]){assert.equal(host.hidden,true);assert.deepEqual(host.children,[]);assert.equal(host.onclick,undefined)}
  assert.equal(hosts[0].details.children[0].textContent,'Details');assert.equal(hosts[0].details.open,undefined);
  hosts[0].actionsHost.append(new Element('button'));list.isConnected=true;list.connectedCallback();assert.equal(list.recordItems(),hosts,'connection retains explicit host mounts from detached app construction');assert.equal(list.recordItems()[0].actionsHost.children.length,1);list.isConnected=false;
  list.page=fromExploration({...response(),basis:'b'.repeat(64)});assert.notEqual(list.recordItems()[0].element,hosts[0].element,'explicit host access renders detached replacement and retires old handles');
  const replacement=list.recordItems();list.display={columns:['utc','summary','context']};assert.notEqual(list.recordItems()[0].element,replacement[0].element);
  list.page={bad:true};list.render();assert.deepEqual(list.recordItems(),[]);assert.equal(list.dataset.state,'unavailable');
 }finally{global.document=oldDocument}
});
