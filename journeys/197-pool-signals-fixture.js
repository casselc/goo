const previous = require('./196-pool-fixture');
const { randomBytes } = require('node:crypto');
const attrs = map => Object.entries(map).map(([key, stringValue]) => ({ key, value: { stringValue } }));
function fixture() {
  const facts = previous.fixture(); facts.environment = 'pool-signals-' + randomBytes(5).toString('hex'); facts.records.forEach(r => { r.environment = facts.environment; });
  facts.points = facts.points.filter(p => p.metric === 'db.client.connection.count' && p.schemaURL === previous.schemaURL);
  facts.histograms = [];
  function capacity(instance, pool, age, value, options = {}) {
    facts.points.push({ metric: 'db.client.connection.max', unit: options.unit || '{connection}', namespace: options.namespace || facts.namespace, instance,
      attributes: { 'db.client.connection.pool.name': pool, ...(options.attributes || {}) }, startNS: String(BigInt(facts.toMs - (options.epochAge || 120000)) * 1000000n),
      endNS: String(BigInt(facts.toMs - age) * 1000000n), value: String(value), valueType: 'int', flags: 0, temporality: 2, monotonic: false,
      schemaURL: options.schemaURL || previous.schemaURL, scope: options.scope || 'ordinary.pool', version: '1', resourceExtra: options.resourceExtra || {}, control: options.control || '' });
  }
  capacity('worker-a', 'orders', 15000, 20); capacity('worker-a', 'orders', 3000, 0, { epochAge: 10000 });
  capacity('worker-b', 'orders', 3000, 30); capacity('worker-a', 'missing-report', 3000, 9);
  capacity('worker-a', 'orders', 3000, 999, { resourceExtra: { 'cloud.region': 'other' }, control: 'different-resource' });
  capacity('worker-a', 'orders', 3000, 888, { attributes: { 'connection.extra': 'other' }, control: 'extra-dimension' });
  capacity('worker-a', 'orders', 3000, 666, { schemaURL: 'https://opentelemetry.io/schemas/1.36.0', control: 'different-schema' });
  capacity('worker-a', 'orders', 3000, 444, { namespace: 'admin', control: 'different-namespace' });
  function wait(instance, pool, temporality, options = {}) {
    facts.histograms.push({ metric: 'db.client.connection.wait_time', unit: options.unit || 's', namespace: facts.namespace, instance,
      attributes: { 'db.client.connection.pool.name': pool, ...(options.attributes || {}) }, scope: 'ordinary.pool', version: '1', schemaURL: previous.schemaURL,
      startNS: String(BigInt(facts.toMs - 5000) * 1000000n), endNS: String(BigInt(facts.toMs - 1000) * 1000000n), temporality,
      bounds: [0.01, 0.05, 0.1], bucketCounts: ['1', '2', '1', '0'], count: '4', sum: 0.13, resourceExtra: {}, control: options.control || '' });
  }
  wait('worker-a', 'orders', 1); wait('worker-b', 'orders', 2); wait('worker-a', 'archive', 1);
  wait('worker-a', 'orders', 1, { unit: 'ms', control: 'different-unit' });
  wait('worker-a', 'orders', 1, { attributes: { 'connection.extra': 'other' }, control: 'extra-dimension' });
  return facts;
}
function resource(facts, point) { return { 'service.name': facts.service, 'service.namespace': point.namespace, 'service.instance.id': point.instance,
  'deployment.environment.name': facts.environment, 'telemetry.sdk.language': 'python', ...(point.resourceExtra || {}) }; }
function metrics(facts) {
  const payload = previous.metrics({ ...facts, points: [] });
  for (const p of facts.points) payload.resourceMetrics.push({ resource: { attributes: attrs(resource(facts, p)) }, scopeMetrics: [{ schemaUrl: p.schemaURL,
    scope: { name: p.scope || 'ordinary.pool', version: p.version || '1' }, metrics: [{ name: p.metric, unit: p.unit, sum: {
      aggregationTemporality: p.temporality, isMonotonic: false, dataPoints: [{ startTimeUnixNano: p.startNS, timeUnixNano: p.endNS,
        asInt: p.value, attributes: attrs(p.attributes), flags: p.flags }] } }] }] });
  for (const p of facts.histograms) payload.resourceMetrics.push({ resource: { attributes: attrs(resource(facts, p)) }, scopeMetrics: [{ schemaUrl: p.schemaURL,
    scope: { name: p.scope, version: p.version }, metrics: [{ name: p.metric, unit: p.unit, histogram: { aggregationTemporality: p.temporality,
      dataPoints: [{ startTimeUnixNano: p.startNS, timeUnixNano: p.endNS, attributes: attrs(p.attributes), count: p.count, sum: p.sum,
        explicitBounds: p.bounds, bucketCounts: p.bucketCounts }] } }] }] });
  return payload;
}
module.exports = { fixture, metrics, traces: previous.traces };
