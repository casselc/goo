const {test}=require('node:test'),assert=require('node:assert/strict');
const {displayJSON,model}=require('../ui/retention-presentation');
test('display exact JSON numeric tokens without altering quoted data',()=>{
 const raw='{"n":18446744073709551615,"negative":-9007199254740993,"zero":-0,"fraction":1.000,"ok":12,"text":"18446744073709551615 \\"n\\": 1e99","nested":[1e999,0.125]}';
 const parsed=displayJSON(raw);assert.equal(parsed.exact,5);assert.equal(parsed.value.n,'Exact JSON number: 18446744073709551615');assert.equal(parsed.value.ok,12);assert.equal(parsed.value.nested[1],0.125);assert.equal(parsed.value.text,JSON.parse(raw).text);
});
test('typed tail presentation preserves distinct windows and never fabricates capture totals',()=>{
 const tail={id:'t',selection:{connection:'connection',selector:{environment:'prod',signal:'spans'},members:[0,1].map(i=>({record:{timestamp_ns:'150000000000'},fromMs:100000+i,toMs:160000+i,observedAt:'2026-10-23T00:00:00Z',ref:{key:'occurrence:'+i},basis:'basis'+i,late:!!i,limited:!i}))}};
 const value=model({origin:'tail-selection',tail,manifestJSON:JSON.stringify(tail),retainedAt:'2026-10-23T01:00:00Z'});assert.equal(value.clock,'Retained at');assert.equal(value.original,undefined);assert.equal(value.members.length,2);assert.notDeepEqual(value.members[0].provenance,value.members[1].provenance);assert.equal(value.details.some(d=>d.label.includes('Source count')),false);
 assert.throws(()=>model({origin:'saved-capture',tail,manifestJSON:JSON.stringify(tail)}));
 assert.throws(()=>model({origin:'mystery',manifestJSON:'{}'}));
});
