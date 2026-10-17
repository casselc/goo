const {test}=require('node:test');const assert=require('node:assert/strict');const {label}=require('../ui/ratios.js');
test('ratio labels preserve exact counts and separate unknown from zero traffic',()=>{
 assert.equal(label({state:'ok',percent:20,errors:'2',calls:'10'}),'20.00% (2/10)');
 assert.equal(label({state:'no_data'}),'No data · zero traffic');
 assert.equal(label({state:'unknown',reason:'missing-outcome'}),'Unknown · missing-outcome');
});
