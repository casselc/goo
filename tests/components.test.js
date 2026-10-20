const {test}=require('node:test'),assert=require('node:assert/strict');
const {latencyModel,durationMS}=require('../ui/components');
test('latency display preserves exact strings, zero and absent quantiles without inventing p90',()=>{
 const summary={state:'observed',count:'130',sumNs:'3250000000',p50Ns:'10000000',p95Ns:'40000000',p99Ns:'40000000',minNs:'10000000',maxNs:'40000000'},before=JSON.stringify(summary),model=latencyModel(summary);
 assert.equal(model.mean,'25.000 ms');assert.deepEqual(model.quantiles.map(q=>[q.label,q.ns,q.position]),[['p50','10000000',25],['p95','40000000',100],['p99','40000000',100]]);assert.equal(JSON.stringify(summary),before);
 assert.deepEqual(latencyModel({...summary,p95Ns:undefined}).quantiles.map(q=>q.label),['p50','p99']);assert.equal(latencyModel({...summary,count:'0'}).state,'unavailable');assert.equal(latencyModel({...summary,count:'invalid'}).state,'unavailable');assert.equal(latencyModel({state:'no_data',count:'0'}).state,'no_data');assert.deepEqual(latencyModel({state:'observed',count:'2',sumNs:'0',p50Ns:'0'}).quantiles[0],{label:'p50',ns:'0',value:'0.000 ms',position:0});assert.equal(durationMS('18446744073709551615'),'18446744073709.552 ms');
});
