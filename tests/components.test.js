const {test}=require('node:test'),assert=require('node:assert/strict');
const {latencyModel,durationMS}=require('../ui/components.mjs');
test('latency display preserves exact strings, zero and absent quantiles without inventing p90',()=>{
 const summary={state:'observed',count:'130',sumNs:'3250000000',p50Ns:'10000000',p95Ns:'40000000',p99Ns:'40000000',minNs:'10000000',maxNs:'40000000'},before=JSON.stringify(summary),model=latencyModel(summary);
 assert.equal(model.mean,'25.000 ms');assert.deepEqual(model.quantiles.map(q=>[q.label,q.ns,q.position]),[['p50','10000000',25],['p95','40000000',100],['p99','40000000',100]]);assert.equal(JSON.stringify(summary),before);
 assert.deepEqual(latencyModel({...summary,p95Ns:undefined}).quantiles.map(q=>q.label),['p50','p99']);assert.equal(latencyModel({...summary,count:'0'}).state,'unavailable');assert.equal(latencyModel({...summary,count:'invalid'}).state,'unavailable');assert.equal(latencyModel({state:'no_data',count:'0'}).state,'no_data');assert.deepEqual(latencyModel({state:'observed',count:'2',sumNs:'0',p50Ns:'0'}).quantiles[0],{label:'p50',ns:'0',value:'0.000 ms',position:0});assert.equal(durationMS('18446744073709551615'),'18446744073709.552 ms');
});
test('caller-supplied latency domain aligns independent summaries without rounding geometry',()=>{
 const base={state:'observed',count:'2',sumNs:'4',p50Ns:'2',p95Ns:'2',p99Ns:'2',minNs:'2',maxNs:'2'},next={...base,sumNs:'16',p50Ns:'8',p95Ns:'8',p99Ns:'8'};
 assert.equal(latencyModel(base,'8').quantiles[0].position,25);assert.equal(latencyModel(next,'8').quantiles[0].position,100);assert.equal(latencyModel(base,'8').ceilingNs,'8');assert.equal(latencyModel(next,'4').state,'unavailable');assert.equal(latencyModel(base,'invalid').state,'unavailable');assert.equal(latencyModel({...base,sumNs:'0',p50Ns:'0',p95Ns:'0',p99Ns:'0'},'0').quantiles[0].position,0);
});
