const previous = require('./197-pool-signals-fixture');
const { randomBytes } = require('node:crypto');
const poolKey = 'db.client.connection.pool.name', stateKey = 'db.client.connection.state';
function fixture() {
 const facts = previous.fixture(); facts.environment = 'pool-view-' + randomBytes(5).toString('hex'); facts.records.forEach(r => { r.environment = facts.environment; });
 // Keep exact compatible bindings plus same-binding identity controls; prior197
 // already covers incompatible unit/schema descriptors. This limits metadata size.
 facts.points = facts.points.filter(p => p.attributes[poolKey] !== 'missing-report' && !(p.instance === 'worker-b' && p.attributes[stateKey] === 'idle') && !['extra-dimension', 'different-namespace'].includes(p.control) && p.schemaURL === facts.schemaURL && p.unit === '{connection}' && p.attributes[poolKey] && (p.metric !== 'db.client.connection.count' || p.attributes[stateKey]));
 facts.histograms = facts.histograms.filter(p => p.temporality === 1 && p.unit === 's' && !p.control);
 const maxA = facts.points.find(p => p.metric === 'db.client.connection.max' && p.instance === 'worker-a' && p.attributes[poolKey] === 'orders' && !p.control && p.value === '0'); maxA.value = '10';
 facts.points.find(p => p.metric === 'db.client.connection.max' && p.instance === 'worker-b').value = '0';
 function add(metric, pool, value, age = 3000, epochAge = 120000) {
  facts.points.push({ metric, unit: '{connection}', namespace: facts.namespace, instance: 'worker-a', attributes: { [poolKey]: pool, ...(metric.endsWith('.count') ? { [stateKey]: 'used' } : {}) }, startNS: String(BigInt(facts.toMs - epochAge) * 1000000n), endNS: String(BigInt(facts.toMs - age) * 1000000n), value: String(value), valueType: 'int', flags: 0, temporality: 2, monotonic: false, schemaURL: facts.schemaURL, scope: 'ordinary.pool', version: '1', resourceExtra: {}, control: '' });
 }
 add('db.client.connection.count', 'time-mismatch', 3); add('db.client.connection.max', 'time-mismatch', 10, 2000);
 add('db.client.connection.count', 'epoch-mismatch', 3); add('db.client.connection.max', 'epoch-mismatch', 10, 3000, 10000);
 add('db.client.connection.count', 'used-exceeds', 11); add('db.client.connection.max', 'used-exceeds', 10);
 add('db.client.connection.max', 'conflicting', 10);
 add('db.client.connection.max', 'archive', 5, 90000);
 add('db.client.connection.count', 'wait-only', 2);
 const waitOnly = structuredClone(facts.histograms.find(p => p.attributes[poolKey] === 'archive')); waitOnly.attributes[poolKey] = 'wait-only'; facts.histograms.push(waitOnly);
 return facts;
}
module.exports = { fixture, metrics: previous.metrics, traces: previous.traces };
