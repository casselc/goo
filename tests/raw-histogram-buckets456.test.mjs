import {test} from 'node:test';
import assert from 'node:assert/strict';
import {facts} from '../ui/raw-histogram-values.mjs';
import {model} from '../ui/raw-histogram-buckets.mjs';
function row({bounds=['-10','0','10'],counts=['2','0','3','5'],flags=0,count}={}){
 const raw={bounds,bucket_counts:counts,count:count??counts.reduce((s,n)=>s+BigInt(n),0n).toString(),flags,temporality:2};
 return {occurrenceKey:'point-1',stream:{unit:'By'},endNS:'1700000000000000001',rawRow:raw,flags,flagsState:'valid',...facts(raw,{flags,flagsState:'valid'}).result};
}
test('one explicit point retains negative intervals, zero buckets, exact counts and source boundary text',()=>{
 const r=row({bounds:['-1e1','0','10']}),m=model(r);assert.equal(m.eligible,true);assert.equal(m.state,'observed');assert.equal(m.count,'10');assert.deepEqual(m.rows.map(r=>r.interval),['(−∞, -10]','(-10, 0]','(0, 10]','(10, +∞)']);assert.equal(m.rows[0].rawInterval,'(−∞, -1e1]');assert.deepEqual(m.rows.map(r=>r.count),['2','0','3','5']);assert.deepEqual(m.rows.map(r=>r.fraction),[.4,0,.6,1]);assert.equal(m.temporality,2);assert.equal(m.unit,'By');
 m.rows[0].count='changed';assert.equal(r.bucketCounts[0],'2');assert.equal(model(r).rows[0].count,'2');
});
test('uint64 counts stay exact while approximate geometry remains finite and tiny counts remain positive',()=>{
 const m=model(row({bounds:['0'],counts:['18446744073709551614','1']}));assert.equal(m.count,'18446744073709551615');assert.equal(m.maximum,'18446744073709551614');assert.equal(m.rows[0].count,'18446744073709551614');assert.equal(m.rows[1].count,'1');assert.equal(m.rows[0].fraction,1);assert.ok(m.rows[1].fraction>0&&m.rows[1].fraction<1);assert.ok(m.rows.every(r=>Number.isFinite(r.fraction)));
});
test('recorded zero count is empty, including a valid all-range bucket; unknown temporality does not invent evaluation',()=>{
 const r=row({bounds:[],counts:['0']}),m=model(r);assert.equal(m.state,'no_data');assert.equal(m.eligible,true);assert.equal(m.rows[0].interval,'(−∞, +∞)');assert.equal(m.rows[0].fraction,0);r.rawRow.temporality=99;r.temporality=99;assert.equal(model(r).eligible,true);
});
test('unknown recording retains valid raw table values but never plots; no-recorded withholds vector',()=>{
 for(const flags of [2,4,4294967294]){const m=model(row({flags}));assert.equal(m.state,'unknown');assert.equal(m.eligible,false);assert.equal(m.rows.length,4);assert.match(m.reason,/Recording validity unknown/)}
 const m=model(row({flags:1}));assert.equal(m.eligible,false);assert.equal(m.rows.length,0);assert.match(m.reason,/No recorded value/);
});
test('malformed, missing, count-inconsistent and forged vectors never produce a plot',()=>{
 for(const change of [r=>r.rawRow.count='11',r=>r.rawRow.bounds[1]='-20',r=>r.rawRow.bucket_counts=['10'],r=>r.rawRow.bounds=null,r=>r.bucketCounts[0]='100',r=>r.flags=2,r=>r.rawRow.flags=2]){const r=row();change(r);const m=model(r);assert.equal(m.eligible,false);assert.equal(m.rows.length,0)}
 for(const input of [null,{},[],{rawRow:{},stream:{unit:''},occurrenceKey:'x'}])assert.equal(model(input).eligible,false);
});
test('full 10001 bucket vector retained without plot truncation, all-domain normalization independent of pages',()=>{
 const m=model(row({bounds:Array.from({length:10000},(_,i)=>String(i)),counts:Array.from({length:10001},(_,i)=>String(i===10000?2:1))}));assert.equal(m.rows.length,10001);assert.equal(m.rows[0].fraction,.5);assert.equal(m.rows.at(-1).fraction,1);assert.equal(m.rows.at(-1).index,10000);
});
