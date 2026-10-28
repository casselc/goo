const {test}=require('node:test'),assert=require('node:assert/strict');const {summary}=require('../ui/captures.mjs');const {routeFromHash}=require('../ui/workspace.mjs');
test('capture summaries retain exact counts and retention separately from saved question identity',()=>{const page={receipt:{id:'x'},explore:{count:'18446744073709551615',basis:'source',state:'observed'},retainedRecords:1000,recordsLimited:true};assert.deepEqual(summary(page),{count:'18446744073709551615',basis:'source',state:'observed',retainedRecords:1000,recordsLimited:true});assert.equal(routeFromHash('#capture='+'a'.repeat(64)),'investigations');assert.equal(routeFromHash('#capture=../'),'explore');});

test('duration display preserves uint64 and rational precision',()=>{
 const {durationMS}=require('../ui/distributions.mjs');
 assert.equal(durationMS('21115000000','205'),'103.000 ms');
 assert.equal(durationMS('18446744073709551615'),'18446744073709.552 ms');
 assert.equal(durationMS('0','0'),'—');
 assert.equal(durationMS('499'),'0.000 ms');
 assert.equal(durationMS('500'),'0.001 ms');
});

test('group duration ranking compares exact rational means and strips measures from count comparisons',()=>{
 const {compareMetric,groupMetric,metricWidth,snapshot}=require('../ui/distributions.mjs');
 const a={g0:'a',count:'2',duration:{state:'observed',count:'2',sumNs:'18446744073709551616',p95Ns:'18446744073709551615'}};
 const b={g0:'b',count:'3',duration:{state:'observed',count:'3',sumNs:'27670116110564327425',p95Ns:'18446744073709551614'}};
 assert.equal(compareMetric(a,b,'mean'),1);assert.equal(compareMetric(a,b,'p95'),-1);
 assert.equal(metricWidth(groupMetric(a,'mean'),groupMetric(b,'mean')),'99.99');
 assert.equal(metricWidth({n:0n,d:1n},{n:0n,d:1n}),'0.00');
 assert.deepEqual(snapshot({question:{},basis:'x',count:'5',state:'observed',groups:[a,b]}).groups,[{g0:'a',count:'2'},{g0:'b',count:'3'}]);
});

test('measure comparisons retain full summaries and rank exact change rather than rounded labels',()=>{
 const {snapshot,compareMeasureDifference,groupMetric}=require('../ui/distributions.mjs');
 const d=(sum,count)=>({state:'observed',count,sumNs:sum,p95Ns:'10'});
 const a={value:'a',measures:{baselineDuration:d('18446744073709551616','2'),selectionDuration:d('18446744073709551618','2')}};
 const b={value:'b',measures:{baselineDuration:d('18446744073709551616','2'),selectionDuration:d('18446744073709551619','2')}};
 assert.equal(compareMeasureDifference(a,b,'mean'),1);
 assert.equal(compareMeasureDifference(a,{value:'missing'},'mean'),-1);
 const status={count:'5',error:'1',ok:'0',unset:'3',unknown:'1'},group={g0:null,count:'5',spanStatus:status};
 assert.deepEqual(groupMetric(group,'error-rate'),{n:100n,d:5n});
 assert.deepEqual(snapshot({question:{spanStatus:true},basis:'x',state:'observed',count:'5',spanStatus:status,groups:[group]},true).groups,[{...group,duration:undefined}]);
 assert.equal(snapshot({question:{},groups:[group]},true).spanStatus,undefined);
});
