const {test}=require('node:test'),assert=require('node:assert/strict');
const G=require('../ui/group-count-comparison.mjs'),S=require('../ui/source-snapshots.mjs'),C=require('../ui/captured-count-comparison.mjs');
const clone=x=>structuredClone(x),query=()=>({signal:'spans',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0,groupBy:[{kind:'attribute',key:'cohort'}]});
function summary(q,count,id){return {version:'goo.dataset/1',id:id+':summary',revision:id,origin:'browser-held',sourceBasis:'source-'+id,capturedAt:'2026-10-31T00:00:00Z',shape:'summary',scope:{id:'scope-'+id,generation:id,signal:q.signal,selector:q},quality:{extent:'unknown',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},fields:[{key:'count',label:'Captured records',type:'number',unit:'records'}],rows:[{key:'population',values:{count}}]};}
function fixture(){
 const a=new Map([[null,'2'],['','1'],['Missing value','1'],['api','2'],['gone','1']]),b=new Map([[null,'2'],['','1'],['Missing value','1'],['api','3'],['fresh','2']]);
 const cell=(m,k)=>({presence:m.has(k)?'present':'absent',state:'derived',value:m.get(k)||'0'}),baseline=summary(query(),'7','baseline'),selection=summary({...query(),fromMs:61000,toMs:121000},'9','selection');
 return {version:'goo.group-count-comparison/1',baseline,selection,field:{kind:'attribute',key:'cohort'},state:'observed',baselineAccounting:'coherent',selectionAccounting:'coherent',rows:[...new Set([...a.keys(),...b.keys()])].map(value=>({key:G.categoryKey(value),value,baseline:cell(a,value),selection:cell(b,value),change:{state:'derived',value:String(BigInt(b.get(value)||'0')-BigInt(a.get(value)||'0')),unit:'records'}}))};
}
function diagnostic(){const x=fixture();x.state='diagnostic';x.reason='Source reads changed.';x.baselineAccounting='diagnostic';for(const row of x.rows){row.baseline={presence:row.baseline.presence==='present'?'present':'unknown',state:'unavailable',value:null,reason:x.reason};row.change={state:'unavailable',value:null,unit:'records',reason:x.reason}}return x;}
function transport(group=fixture()){
 const rows=side=>group.rows.filter(row=>row[side].presence==='present').map(row=>({g0:row.value,count:row[side].value})),request={baseline:{question:clone(group.baseline.scope.selector),basis:group.baseline.sourceBasis,capturedAt:group.baseline.capturedAt,count:group.baseline.rows[0].values.count,state:'observed',groups:rows('baseline')},selection:{question:clone(group.selection.scope.selector),basis:group.selection.sourceBasis,capturedAt:group.selection.capturedAt,count:group.selection.rows[0].values.count,state:'observed',groups:rows('selection')}};
 return {request,data:{viewComparison:{version:'goo.comparison/1',baseline:clone(group.baseline),selection:clone(group.selection),state:group.state,changes:{count:{state:'derived',value:'2',unit:'records'}}},groupComparison:group}};
}
test('group counts preserve typed categories, exact union and one shared domain during local exploration',()=>{
 const raw=fixture(),before=clone(raw);assert.equal(G.validateGroupComparison(raw).ok,true);const held=G.asGroupComparison(raw);raw.rows[0].baseline.value='100';assert.deepEqual(held,before);assert(Object.isFrozen(held.rows[0].baseline));
 assert.notEqual(G.categoryLabel(null),G.categoryLabel(''));assert.notEqual(G.categoryLabel(null),G.categoryLabel('Missing value'));assert.equal(G.categoryLabel('Missing value'),'"Missing value"');
 const all=G.projection(held),filtered=G.projection(held,{query:'gone',order:'label'});assert.equal(all.domainMax,'3');assert.equal(filtered.domainMax,'3');assert.equal(filtered.rows.length,1);assert.equal(filtered.total,6);assert.equal(all.rows[0].value,'fresh');assert.equal(all.rows[0].change.value,'2');
 assert.equal(G.projection(held,{query:'no such value'}).matched,0);assert.equal(G.projection(held,{query:'no such value'}).total,6);
});
test('coherent zero and diagnostic absence remain different without client-created differences',()=>{
 const zero=fixture();zero.baseline.rows[0].values.count='0';zero.state='no_data';zero.rows=zero.rows.filter(row=>row.selection.presence==='present');for(const row of zero.rows){row.baseline={presence:'absent',state:'derived',value:'0'};row.change.value=row.selection.value}assert.equal(G.validateGroupComparison(zero).ok,true);assert.equal(G.projection(zero).domainMax,'3');
 const d=diagnostic();assert.equal(G.validateGroupComparison(d).ok,true);assert.equal(G.projection(d).domainMax,null);assert(G.projection(d).rows.every(row=>row.change.value===null));
 const extreme=fixture();extreme.rows=[{key:'1huge',value:'huge',baseline:{presence:'present',state:'derived',value:'18446744073709551615'},selection:{presence:'absent',state:'derived',value:'0'},change:{state:'derived',value:'-18446744073709551615',unit:'records'}}];extreme.baseline.rows[0].values.count='18446744073709551615';extreme.selection.rows[0].values.count='0';extreme.state='no_data';assert.equal(G.validateGroupComparison(extreme).ok,true);
});
test('group profile rejects false zero, changed identities, accounting drift and diagnostic arithmetic',()=>{
 for(const mutate of [x=>x.rows[0].key='1forged',x=>x.rows.push({...clone(x.rows[0]),key:'1new'}),x=>x.rows.pop(),x=>x.rows[0].baseline.value='0',x=>x.rows[0].change.value='1',x=>x.rows[0].change.value='-0',x=>x.rows[0].change.aggregation={},x=>x.rows[0].value='x'.repeat(4097),x=>x.rows[0].value='\ud800',x=>x.baseline.rows[0].values.count='8',x=>x.field.key='different',x=>{x.field={kind:'column',key:'service'};for(const side of ['baseline','selection'])x[side].scope.selector.groupBy=[x.field]},x=>x.rows[0].selection.value='18446744073709551616',x=>x.extra='authority']){const x=fixture();mutate(x);assert.equal(G.validateGroupComparison(x).ok,false)}
 for(const mutate of [x=>x.rows[0].baseline.value='0',x=>x.rows.at(-1).baseline={presence:'absent',state:'derived',value:'0'},x=>x.rows[0].change={state:'derived',value:'0',unit:'records'}]){const x=diagnostic();mutate(x);assert.equal(G.validateGroupComparison(x).ok,false)}
 const getter=fixture();Object.defineProperty(getter.baseline.rows[0].values,'count',{enumerable:true,get(){throw Error('getter executed')}});assert.equal(G.validateGroupComparison(getter).ok,false);assert.throws(()=>G.asGroupComparison(getter),/data properties/);
});
test('source admission binds groups to both summary identities and requested original partitions',()=>{
 const {request,data}=transport(),admitted=S.importComparisonResult(data,request);assert.equal(admitted.groupComparison.rows.length,6);assert.equal(admitted.groupComparison.baseline.id,admitted.viewComparison.baseline.id);assert.notEqual(admitted.groupComparison.baseline.id,data.groupComparison.baseline.id);data.groupComparison.rows[0].baseline.value='200';assert.equal(admitted.groupComparison.rows[0].baseline.value,'2');
 for(const mutate of [x=>x.groupComparison.baseline.capturedAt='2026-10-31T01:00:00Z',x=>{for(const v of [x.groupComparison,x.viewComparison])v.baseline.capturedAt='2026-10-31T01:00:00Z'},x=>{for(const v of [x.groupComparison,x.viewComparison])v.baseline.sourceBasis='other'},x=>x.groupComparison=null,x=>delete x.viewComparison]){const next=transport();mutate(next.data);assert.throws(()=>S.importComparisonResult(next.data,next.request))}
 const shifted=transport();shifted.request.baseline.groups[0].g0='another';assert.throws(()=>S.importComparisonResult(shifted.data,shifted.request));
});
test('comparison declaration keeps only bounded local view and order settings',()=>{
 const q=query(),value=C.create(q,'a'.repeat(64));value.display.view='groups';value.display.order='label';assert.equal(C.validate(value,q).display.view,'groups');for(const changed of [{view:'heatmap'},{order:'percent'},{view:null},{order:null}])assert.throws(()=>C.validate({...value,display:{...value.display,...changed}},q));
});


test('invalid or removed generic component input cannot resurrect an older supplied pair',()=>{
 const {GroupCountComparison}=require('../ui/group-count-comparison-view.mjs'),element=new GroupCountComparison(),attributes=new Map(),oldDocument=global.document;
 element.isConnected=false;element.getAttribute=key=>attributes.has(key)?attributes.get(key):null;element.replaceChildren=(...nodes)=>{element.nodes=nodes};Object.defineProperty(element,'firstChild',{get:()=>element.nodes?.[0]});
 global.document={createElement:tag=>({tag,textContent:'',setAttribute(){}})};
 try{
  element.comparison=fixture();assert(element.comparison);element.isConnected=true;attributes.set('comparison','{broken');element.attributeChangedCallback('comparison');assert.equal(element.comparison,null);const refused=element.firstChild.textContent;element.display={order:'label'};assert.equal(element.firstChild.textContent,refused);
  element.isConnected=false;element.display={order:'invalid'};element.comparison=fixture();element.isConnected=true;element.render();assert.match(element.firstChild.textContent,/checked group display/);
  element.isConnected=false;attributes.delete('display');element.attributeChangedCallback('display');assert.equal(element.display.order,'change');attributes.delete('comparison');element.attributeChangedCallback('comparison');assert.equal(element.comparison,null);
  const recovering=new GroupCountComparison();let rendered=0;recovering.render=()=>{rendered++};recovering.isConnected=true;recovering.display={order:'label'};recovering.display={order:'wrong'};const count=rendered;recovering.display={order:'label'};assert.equal(rendered,count+1);
 }finally{global.document=oldDocument}
});


test('full 2000-group union retains one domain through local row paging and enforces side bounds',()=>{
 const x=fixture();x.baseline.rows[0].values.count='1000';x.selection.rows[0].values.count='1000';x.rows=[];
 for(const side of ['baseline','selection'])for(let i=0;i<1000;i++){const value=(side==='baseline'?'b':'s')+String(i).padStart(4,'0'),present={presence:'present',state:'derived',value:'1'},absent={presence:'absent',state:'derived',value:'0'};x.rows.push({key:G.categoryKey(value),value,baseline:side==='baseline'?present:absent,selection:side==='selection'?present:absent,change:{state:'derived',value:side==='baseline'?'-1':'1',unit:'records'}})}
 assert.equal(G.validateGroupComparison(x).ok,true);const held=G.asGroupComparison(x),last=G.projection(held,{page:999});assert.equal(last.total,2000);assert.equal(last.page,39);assert.equal(last.rows.length,50);assert.equal(last.domainMax,'1');assert.equal(G.projection(held,{query:'b0001'}).domainMax,'1');
 const excessive=clone(x);excessive.rows[1000].baseline={presence:'present',state:'derived',value:'1'};excessive.rows[1000].change.value='0';excessive.baseline.rows[0].values.count='1001';assert.equal(G.validateGroupComparison(excessive).ok,false);
});


test('diagnostic status remains explicit when no groups or no local search matches exist',()=>{
 const {groupStatus}=require('../ui/group-count-comparison-view.mjs');
 assert.match(groupStatus({total:0,matched:0,diagnostic:true,domainMax:null}),/Source reads disagree/);assert.doesNotMatch(groupStatus({total:0,matched:0,diagnostic:true}),/No groups observed/);
 assert.match(groupStatus({total:3,matched:0,diagnostic:true}),/No matching groups.*Source reads disagree/);assert.equal(groupStatus({total:0,matched:0,diagnostic:false}),'No groups observed in either supplied population.');
});


test('long Unicode group labels use a bounded preview with the exact quoted value retained',()=>{
 const value='😀'.repeat(300),label=G.categoryPreview(value);assert(label.long);assert.equal(Array.from(label.preview).length,81);assert(label.preview.isWellFormed());assert.equal(label.label,JSON.stringify(value));assert.equal(G.categoryPreview('api').preview,'"api"');assert.equal(G.categoryPreview(null).long,false);
});


test('source admission follows Go nullable group defaults for an omitted request g0',()=>{
 const {request,data}=transport();delete request.baseline.groups.find(row=>row.g0===null).g0;delete request.selection.groups.find(row=>row.g0===null).g0;const admitted=S.importComparisonResult(data,request);assert.equal(admitted.groupComparison.rows.find(row=>row.key==='0').value,null);assert.equal(admitted.groupComparison.rows.find(row=>row.key==='0').baseline.value,'2');
});
