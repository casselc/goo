import {test} from 'node:test';
import assert from 'node:assert/strict';
import {summary} from '../ui/tail-presentation.mjs';
const base={phase:'paused',message:'Paused while reading earlier observations. Resume starts a new recent window.',gap:false,retained:8,evicted:0,observed:8,compact:true};
test('compact buffer presentation reduces repetition while retaining all facts and uncertainty',()=>{
 const paused=summary(base);assert.equal(paused.text,'Paused');assert.equal(paused.counts,'8 loaded');assert.match(paused.detail,/Resume/);assert.match(paused.facts,/0 evicted.*8 observations.*Not a population count/);
 const loss=summary({...base,gap:true,evicted:24,observed:32});assert.match(loss.text,/continuity gap/);assert.match(loss.detail,/no replay guarantee/);assert.match(loss.counts,/24 evicted/);assert.match(loss.facts,/32 observations/);
 for(const phase of ['unknown','partial','stale','unavailable']){const message=phase==='unknown'?'No recent observations loaded. Coverage unknown.':phase==='stale'?'Stream capacity reached; reconnecting in 2 seconds.':'Source read unavailable.';assert.equal(summary({...base,phase,message}).text,message)}
 const original=summary({...base,compact:false});assert.equal(original.text,base.message);assert.match(original.counts,/retained observations.*0 evicted.*8 observations received/);
});
