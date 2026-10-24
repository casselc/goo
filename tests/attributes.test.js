const {test}=require('node:test');
const assert=require('node:assert/strict');
const {groups,matches,type,text}=require('../ui/attributes.js');
test('inspection preserves scope, empty values, scalar types and exact integer strings',()=>{
 const record={SpanName:'work',duration_ns:'18446744073709551615',ResourceAttributes:{same:'resource'},Attributes:{same:'',nullable:null,flag:false,count:'9007199254740993'},SpanAttributes:{same:'duplicate-alias'}};
 const result=groups(record);
 assert.equal(result[1].entries[0].field.kind,'resource');
 assert.equal(result[2].entries.find(e=>e.key==='same').value,'');
 assert.equal(result[2].entries.filter(e=>e.key==='same').length,1);
 assert.equal(result[0].entries.find(e=>e.key==='duration_ns').value,'18446744073709551615');
 assert.deepEqual(result[2].entries.find(e=>e.key==='nullable').field,{kind:'attribute',key:'nullable'});
 assert.equal(type(null),'null');assert.equal(type(false),'boolean');assert.equal(type('false'),'string');assert.equal(type([]),'array');
 assert.equal(text('9007199254740993'),'9007199254740993');assert.equal(text(' a\n b '),' a\n b ');
 assert.ok(!result[2].entries.some(e=>e.key==='absent'));
 assert.equal(record.Attributes.same,'');
});
test('inspection searches literal hostile keys/values without evaluating them',()=>{
 const entry={key:'__proto__.<script>',value:'a\\b\n<img src=x onerror=alert(1)> 雪'};
 assert.ok(matches(entry,'<img'));assert.ok(matches(entry,'__PROTO__'));assert.ok(matches(entry,'雪'));
 assert.equal(groups(JSON.parse('{"__proto__":"literal"}'))[0].entries[0].field,undefined);
 assert.ok(!matches(entry,'a.*b'));assert.equal(groups(JSON.parse('{"Attributes":{"__proto__":"literal"}}'))[2].entries[0].value,'literal');
});
test('field display distinguishes absence from present null, empty, false and exact strings',()=>{
 const {fieldModel}=require('../ui/attributes.js');
 assert.deepEqual(fieldModel(undefined),{state:'missing',kind:null,label:'Missing value'});
 assert.deepEqual(fieldModel(null),{state:'present',kind:'null',label:'Null'});
 assert.deepEqual(fieldModel(''),{state:'present',kind:'string',label:'Empty string'});
 assert.deepEqual(fieldModel(false),{state:'present',kind:'boolean',label:'false'});
 assert.deepEqual(fieldModel('900719925474099300'),{state:'present',kind:'string',label:'900719925474099300'});
 for(const state of ['missing','unknown','unavailable'])assert.equal(fieldModel('retained but withheld',state).kind,null);
 assert.equal(fieldModel('', 'unrecognized').state,'unknown');
 assert.equal(fieldModel({present:false}).kind,'object');assert.equal(fieldModel([]).kind,'array');
});

test('field properties withhold non-JSON values without coercion or executable hooks',()=>{
 const {isJSONValue,fieldModel}=require('../ui/attributes.js');const cycle={};cycle.self=cycle;let reads=0;
 const accessor={get secret(){reads++;return 'side effect'}};
 class HookedArray extends Array{toJSON(){reads++;return 'changed'}}
 const inheritedHook=Object.setPrototypeOf([], {toJSON(){reads++;return 'changed'}});
 for(const value of [cycle,1n,NaN,Infinity,{nested:1n},{nested:undefined},[undefined],Array(1),new Date(),new Map(),new HookedArray(),inheritedHook,accessor,{toJSON(){reads++;return 'changed'}}]){
  assert.equal(isJSONValue(value),false);assert.equal(fieldModel(value).state,'unavailable');
 }
 assert.equal(reads,0);const shared={flag:false};assert.equal(isJSONValue({a:shared,b:shared}),true);
 assert.equal(isJSONValue(JSON.parse('{"__proto__":"literal","value":null}')),true);
 assert.equal(isJSONValue({nested:['900719925474099300',null,false,'']}),true);
});

test('shared record fields retain namespaces and exact strings without inferring absence',()=>{
 const {commonGroups}=require('../ui/attributes'),records=[{ServiceName:'api',ResourceAttributes:{region:'west'},Attributes:{region:'',large:'9007199254740993',null:null,flag:false,one:'x'}},{ServiceName:'worker',ResourceAttributes:{region:'west'},Attributes:{region:'',large:'9007199254740993',null:null,flag:false}}],before=structuredClone(records),model=commonGroups(records);
 assert.equal(model[0].entries.some(e=>e.key==='ServiceName'),false);assert.equal(model[1].entries[0].value,'west');assert.deepEqual(model[2].entries.map(e=>e.key),['region','large']);assert.equal(model[2].entries[0].value,'');model[1].entries[0].field.kind='attribute';assert.deepEqual(records,before);assert.throws(()=>commonGroups([]));assert.throws(()=>commonGroups(Array(21).fill(records[0])));assert.equal(commonGroups([{Attributes:{a:'x'}},{Attributes:{a:'y'}}])[2].entries.length,0);
});

test('record population actions withhold values the typed query cannot represent',()=>{
 const {selectable}=require('../ui/attributes'),field={kind:'attribute',key:'region'};for(const value of [null,false,1,'x'.repeat(1025),'雪'.repeat(400),'a\0b','\ud800'])assert.equal(selectable({field,value}),false);assert.equal(selectable({field,value:'x'.repeat(1024)}),true);assert.equal(selectable({field,value:'😀'}),true);assert.equal(selectable({field:{kind:'attribute',key:'雪'.repeat(50)},value:'ok'}),false);
});

test('native identity snapshots preserve qualified fields, omitted schemas and exact values',()=>{const {streamRecord}=require('../ui/attributes'),s={name:'http.server.request.duration',unit:'s',resource:{region:'',large:'9007199254740993'},attributes:{region:'west'},scope:'otel.http',version:'1',schemaURL:'schema'};const r=streamRecord(s);assert.equal(r.ResourceAttributes.region,'');assert.equal(r.Attributes.region,'west');assert.equal(r.ResourceAttributes.large,'9007199254740993');assert.equal(r.resourceSchema,undefined);assert.equal(r.schema,'schema');assert.equal(Object.hasOwn(r,'timestamp_ns'),false);r.ResourceAttributes.region='mutated';assert.equal(s.resource.region,'')});

test('optional derived record projections retain qualified field identity without changing null or number strings',()=>{const entries=groups({exception_event:'yes',exception_type:null,exception_type_state:'multiple',severity_number:'17'})[0].entries;for(const item of entries)assert.deepEqual(item.field,{kind:'column',key:item.key});assert.equal(entries.find(e=>e.key==='exception_type').value,null);assert.equal(entries.find(e=>e.key==='severity_number').value,'17')});
