import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Fields from '../ui/field-presentation.mjs';
import * as Session from '../ui/question-session.mjs';
const field={kind:'column',key:'duration_ns'};
test('fixed duration thresholds use exact decimal units across selector and parts',()=>{
 for(const [value,want] of [['0','0 ns'],['1','1 ns'],['999','999 ns'],['1000','1 µs'],['1001','1.001 µs'],['999999','999.999 µs'],['1000000','1 ms'],['100000001','100.000001 ms'],['100000000','100 ms'],['1000000000','1 s'],['1000000001','1.000000001 s'],['9007199254740993','9007199.254740993 s'],['18446744073709551615','18446744073.709551615 s']]){
  for(const op of ['eq','ne','gte','lte']){
   const p={field,op,value},before=structuredClone(p),identity=Session.identity({where:p});
   assert.equal(Fields.valueText(field,value),want);assert.equal(Fields.parts(p).value,want);assert.equal(Fields.selectorText(p),'Span duration '+Fields.operator(op)+' '+want);assert.equal(Session.selectorText(p),Fields.selectorText(p));assert.deepEqual(p,before);assert.equal(Session.identity({where:p}),identity);assert.equal(Fields.referenceText(p),JSON.stringify(before));
  }
 }
 assert.equal(Fields.label(field),'Span duration (ns)','authoring input keeps its raw unit');
});
test('duration display does not infer units for attributes, bindings or malformed values',()=>{
 for(const value of ['','01','-1','1.2','1e6','18446744073709551616','not recorded',100000000,null]){
  assert.equal(Fields.valueText(field,value),(value??'')===''?'Empty string':JSON.stringify(value));
 }
 assert.equal(Fields.valueText({kind:'attribute',key:'duration_ns'},'100000000'),'"100000000"');
 const numeric={numeric:{field:{kind:'attribute',key:'duration_ns'},type:'uint64',unit:'ns',parse:'decimal'},op:'gte',value:'100000000'};
 assert.equal(Fields.parts(numeric).value,'100000000');assert.match(Fields.selectorText(numeric),/100000000$/);
 for(const op of ['missing','exists'])assert.equal(Fields.selectorText({field,op}),'Span duration '+Fields.operator(op));
 const group={all:[{field,op:'gte',value:'100000000'},{field:{kind:'column',key:'severity_number'},op:'eq',value:'17'}]};
 assert.equal(Fields.selectorText(group),'(Span duration ≥ 100 ms AND OTel severity number = "17")');assert.equal(Fields.referenceText(group),JSON.stringify(group));
});
test('exact duration strings round trip to nanoseconds without precision loss',()=>{
 for(const ns of [0n,1n,999n,1000n,999999n,1000000n,999999999n,1000000000n,9007199254740993n,18446744073709551615n]){
  const shown=Fields.valueText(field,String(ns)),[decimal,unit]=shown.split(' '),digits={ns:0,'µs':3,ms:6,s:9}[unit],parts=decimal.split('.');
  const recovered=BigInt(parts[0])*10n**BigInt(digits)+BigInt((parts[1]||'').padEnd(digits,'0')||'0');assert.equal(recovered,ns);
 }
});
