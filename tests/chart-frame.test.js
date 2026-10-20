const {test}=require('node:test'),assert=require('node:assert/strict');
const {model}=require('../ui/chart-frame');
test('chart frame preserves caller window and units without deriving health or a window',()=>{
 const input={title:'HTTP requests',unit:'source records',window:'[12:00, 12:05) UTC',state:'stale',note:'Last completed read 12:05:02 UTC',headingLevel:2},before=structuredClone(input);
 assert.deepEqual(model(input),{...input,headingLevel:'2'});assert.deepEqual(input,before);
 assert.equal(model({state:'healthy'}).state,'unknown');assert.equal(model({}).window,'');assert.equal(model({headingLevel:1}).headingLevel,'3');
});
test('chart frame bounds chrome and retains explicit absent-data states',()=>{
 for(const state of ['observed','no_data','unknown','loading','partial','stale','unavailable','incompatible'])assert.equal(model({state}).state,state);
 assert.equal(model({title:'x'.repeat(5000)}).title.length,4096);assert.equal(model(null).title,'Chart');assert.equal(model({unit:42}).unit,'');
});
