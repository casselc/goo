const {test}=require('node:test'),assert=require('node:assert/strict');const {summary}=require('../ui/captures');const {routeFromHash}=require('../ui/workspace');
test('capture summaries retain exact counts and retention separately from saved question identity',()=>{const page={receipt:{id:'x'},explore:{count:'18446744073709551615',basis:'source',state:'observed'},retainedRecords:1000,recordsLimited:true};assert.deepEqual(summary(page),{count:'18446744073709551615',basis:'source',state:'observed',retainedRecords:1000,recordsLimited:true});assert.equal(routeFromHash('#capture='+'a'.repeat(64)),'investigations');assert.equal(routeFromHash('#capture=../'),'explore');});

test('duration display preserves uint64 and rational precision',()=>{
 const {durationMS}=require('../ui/distributions');
 assert.equal(durationMS('21115000000','205'),'103.000 ms');
 assert.equal(durationMS('18446744073709551615'),'18446744073709.552 ms');
 assert.equal(durationMS('0','0'),'—');
 assert.equal(durationMS('499'),'0.000 ms');
 assert.equal(durationMS('500'),'0.001 ms');
});
