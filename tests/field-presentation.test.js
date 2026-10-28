const {test}=require('node:test'),assert=require('node:assert/strict'),fields=require('../ui/field-presentation.mjs'),session=require('../ui/question-session'),fixtures=require('../testdata/protocol/record-fields.json');
test('Go record wire fixtures retain qualified fields, empty defaults and exact values in display',()=>{assert.equal(fixtures.version,1);for(const c of fixtures.cases){const before=JSON.stringify(c.predicate);assert.deepEqual(fields.parts(c.predicate),c.display,c.name);assert.deepEqual(fields.parts(c.canonical),c.display,c.name+' canonical');assert.equal(session.identity({where:c.predicate}),session.identity({where:c.canonical}),c.name+' identity');assert.equal(JSON.stringify(c.predicate),before)}});
test('nested live selector text keeps literal strings and presence distinct',()=>{const where={any:[{field:{kind:'resource',key:'region'},op:'missing'},{field:{kind:'attribute',key:'region'},op:'eq'}]};assert.equal(fields.selectorText(where),'(resource region is missing OR attribute region = "")');assert.equal(fields.selectorText(where),session.selectorText(where));assert.notEqual(fields.selectorText(where.any[1]),fields.selectorText({...where.any[1],value:'Empty string'}));assert.equal(fields.qualified({kind:'attribute',key:'地域:route'}),'attribute:地域:route')});

test('literal sentinel text stays distinct from an empty record string',()=>{const p={field:{kind:'attribute',key:'region'},op:'eq'};assert.notEqual(fields.parts(p).value,fields.parts({...p,value:'Empty string'}).value);assert.equal(fields.parts({...p,value:'Empty string'}).value,'\"Empty string\"')});

test('event timestamp predicates display exact UTC without rounding or changing the wire value',()=>{
 const p={field:{kind:'column',key:'timestamp_ns'},op:'lte',value:'1700000000000999999'},before=structuredClone(p);
 assert.equal(fields.parts(p).value,'2023-11-14T22:13:20.000999999Z');
 assert.equal(fields.selectorText(p),'Event time (UTC) ≤ 2023-11-14T22:13:20.000999999Z');assert.deepEqual(p,before);
 assert.equal(fields.parts({...p,value:'0'}).value,'1970-01-01T00:00:00.000000000Z');
 assert.equal(fields.parts({...p,value:'18446744073709551615'}).value,'2554-07-21T23:34:33.709551615Z');
 for(const value of ['-1','01','not a clock','18446744073709551616'])assert.equal(fields.parts({...p,value}).value,JSON.stringify(value));
 assert.equal(fields.parts({...p,field:{kind:'attribute',key:'timestamp_ns'}}).value,JSON.stringify(p.value));
});
