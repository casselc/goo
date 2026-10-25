const { randomBytes } = require('node:crypto');
const attrs = map => Object.entries(map).map(([key, stringValue]) => ({ key, value: { stringValue } }));
function fixture() {
  const environment = 'cumulative-stock-' + randomBytes(5).toString('hex');
  const from = BigInt(Date.now() - 30000) * 1000000n, to = from + 20000000000n;
  const facts = { environment, fromNS: String(from), toNS: String(to), points: [] };
  function add(instance, pool, state, offset, value, options = {}) {
    facts.points.push({ metric: options.metric || 'db.client.connection.count', unit: '{connection}', namespace: 'inventory', instance,
      attributes: { pool, state }, startNS: String(from - 1000000000n + BigInt(options.epoch || 0)),
      endNS: String(from + BigInt(offset) * 1000000000n), value: String(value), valueType: options.double ? 'double' : 'int',
      flags: options.flags || 0, temporality: options.delta ? 1 : 2, monotonic: false });
  }
  add('worker-a', 'main', 'used', 1, 10);
  add('worker-a', 'main', 'used', 9, 7);
  add('worker-a', 'main', 'used', 17, 3, { epoch: 15000000000 });
  add('worker-a', 'main', 'idle', 17, 12);
  add('worker-b', 'main', 'used', 17, 21);
  add('worker-a', 'archive', 'used', 2, 5);
  add('worker-a', 'missing', 'used', 17, 0, { flags: 1 });
  add('worker-a', 'conflicting', 'used', 17, 4);
  add('worker-a', 'conflicting', 'used', 17, 9);
  add('worker-a', 'generic', 'reported', 9, 2.5, { double: true });
  add('worker-a', 'generic', 'reported', 17, -1.25, { double: true });
  add('worker-a', 'delta-only', 'used', 17, 2, { delta: true, metric: 'work.unsupported.delta.stock' });
  return facts;
}
function payload(facts) {
  return { resourceMetrics: facts.points.map(p => ({ resource: { attributes: attrs({ 'service.name': 'stock-worker', 'service.namespace': p.namespace,
    'service.instance.id': p.instance, 'deployment.environment.name': facts.environment, 'telemetry.sdk.language': 'python' }) }, scopeMetrics: [{
      scope: { name: 'ordinary.stock', version: '1' }, metrics: [{ name: p.metric, unit: p.unit, sum: {
        aggregationTemporality: p.temporality, isMonotonic: false, dataPoints: [{ startTimeUnixNano: p.startNS, timeUnixNano: p.endNS,
          [p.valueType === 'double' ? 'asDouble' : 'asInt']: p.valueType === 'double' ? Number(p.value) : p.value,
          attributes: attrs(p.attributes), flags: p.flags }] } }] }] })) };
}
function question(facts) {
  return { definition: { apiVersion: 'goo/measurement-package/v1alpha1', name: 'reported-stock', title: 'Reported stock observations', kind: 'stock',
    environment: facts.environment, lookbackMinutes: 1, maxGapMs: 5000, grain: 'per-stream', source: { services: ['stock-worker'],
      namespace: { mode: 'value', value: 'inventory' }, scope: 'ordinary.stock', version: '1', schemaURL: '', metric: 'db.client.connection.count', unit: '{connection}' } },
    fromMs: Number(BigInt(facts.fromNS) / 1000000n), toMs: Number(BigInt(facts.toNS) / 1000000n), fromNS: facts.fromNS, toNS: facts.toNS };
}
module.exports = { fixture, payload, question };
