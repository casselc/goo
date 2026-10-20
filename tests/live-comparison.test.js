const {test}=require('node:test'),assert=require('node:assert/strict'),{displayState}=require('../ui/live-comparison');
test('freeze preserves unavailable/incompatible numeric evidence instead of announcing observed',()=>{
 for(const phase of ['current','frozen']){assert.equal(displayState(phase,{state:'comparable'}),'observed');assert.equal(displayState(phase,{state:'unavailable'}),'unavailable');assert.equal(displayState(phase,{state:'incompatible'}),'incompatible');assert.equal(displayState(phase,{state:'unsupported'}),'unavailable')}
});
test('pausing an unstarted panel does not invent a stale result',()=>{assert.equal(displayState('paused',null),'unknown');assert.equal(displayState('paused',{state:'comparable'}),'stale');assert.equal(displayState('loading',null),'loading');assert.equal(displayState('unavailable',null),'unavailable');assert.equal(displayState('stale',{state:'comparable'}),'stale')});
