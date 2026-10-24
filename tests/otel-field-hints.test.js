const {test}=require('node:test'),assert=require('node:assert/strict'),fields=require('../ui/field-presentation'),{suggestionData,matches}=require('../ui/attributes'),{normalize,validate}=require('../ui/filter-editor');
const f=(kind,key)=>({kind,key});
test('OTel hints require exact qualified keys; aliases remain independent and detached',()=>{
 for(const [kind,key,label] of [['attribute','http.route','HTTP route'],['attribute','rpc.system.name','RPC system'],['attribute','db.operation.name','Database operation'],['attribute','messaging.destination.name','Message destination'],['resource','service.name','Service']]){
  const field=f(kind,key),before=structuredClone(field),h=fields.hint(field);assert.equal(h.label,label);assert.equal(h.legacy,false);h.label='mutated';assert.equal(fields.hint(field).label,label);assert.deepEqual(field,before);
 }
 for(const field of [f('resource','http.route'),f('attribute','service.name'),f('attribute','HTTP.ROUTE'),f('column','http.route'),f('attribute','custom.http.route'),f('attribute','__proto__'),null])assert.equal(fields.hint(field),null);
 for(const key of ['http.method','http.status_code','db.name','rpc.service','messaging.destination'])assert.equal(fields.hint(f('attribute',key)).legacy,true);
 const predicate={field:f('attribute','http.method'),op:'eq',value:'GET'};assert.deepEqual(normalize(predicate),predicate);assert.equal(fields.qualified(predicate.field),'attribute:http.method');
});
test('enrichment never invents observed fields, maps old keys, or changes typed values',()=>{
 const record={Attributes:{'http.method':'GET','http.status_code':'503','custom.http.route':'/raw'},ResourceAttributes:{'service.name':'worker'}},before=structuredClone(record),data=suggestionData([record]);
 assert.equal(data.fields.some(x=>x.field.key==='http.request.method'),false);assert.equal(data.fields.some(x=>x.field.key==='http.route'),false);
 for(const item of data.fields){fields.hint(item.field);fields.searchText(item.field);assert.ok(matches({key:item.field.key,field:item.field,value:item.values[0]},item.field.key));}
 assert.deepEqual(record,before);assert.ok(matches({key:'http.route',field:f('attribute','http.route'),value:'/orders/{id}'},'route template'));
 assert.equal(matches({key:'custom.http.route',field:f('attribute','custom.http.route'),value:'x'},'route template'),false);
 assert.throws(()=>validate({field:f('attribute','http.response.status_code'),op:'gte',value:'500'},'spans')); // Hints do not introduce numeric coercion.
 assert.match(fields.hint(f('attribute','http.response.status_code')).description,/separate from reported span status/);
 assert.match(fields.hint(f('column','status')).description,/Unset or absent/);
});

test('semantic search keeps original exporter column names searchable',()=>{for(const [key,logical] of [['TraceId','trace_id'],['StatusCode','status'],['ServiceName','service']])assert.equal(matches({key,field:f('column',logical),value:'sample'},key),true)});
