const {test}=require('node:test');const assert=require('node:assert/strict');
const {bucketLabel,p95Label}=require('../ui/histograms.mjs');
test('native histogram labels preserve inclusive upper bounds and unbounded tails',()=>{
 assert.equal(bucketLabel([0,.01],0,'s'),'(-∞, 0] s');
 assert.equal(bucketLabel([0,.01],1,'s'),'(0, 0.01] s');
 assert.equal(bucketLabel([0,.01],2,'s'),'(0.01, +∞) s');
 assert.equal(p95Label({state:'no_data'},'s'),'P95: no observations');
 assert.match(p95Label({state:'bucket-bound',lower:0,upper:.01},'s'),/0, 0.01\] s · no interpolation/);
 assert.match(p95Label({state:'unbounded-upper',lower:30},'s'),/30, \+∞\)/);
});
