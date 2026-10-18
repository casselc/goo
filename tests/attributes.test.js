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
