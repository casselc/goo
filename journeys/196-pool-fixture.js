const { randomBytes } = require('node:crypto');
const schemaURL = 'https://opentelemetry.io/schemas/1.37.0';
const keys = { pool: 'db.client.connection.pool.name', state: 'db.client.connection.state' };
const attrs = map => Object.entries(map).map(([key, stringValue]) => ({ key, value: { stringValue } }));
function fixture() {
  const environment = 'named-pools-' + randomBytes(5).toString('hex'), toMs = Date.now() - 10000, fromMs = toMs - 110000;
  const facts = { environment, service: 'pool-api', namespace: 'inventory', fromMs, toMs, schemaURL, points: [], records: [] };
  function add(instance, pool, state, age, value, options = {}) {
    facts.points.push({ metric: options.metric || 'db.client.connection.count', unit: options.unit || '{connection}', namespace: facts.namespace, instance,
      attributes: { ...(pool === null ? {} : { [keys.pool]: pool }), ...(state === null ? {} : { [keys.state]: state }) },
      startNS: String(BigInt(toMs - (options.epochAge || 120000)) * 1000000n), endNS: String(BigInt(toMs - age) * 1000000n),
      value: String(value), valueType: options.double ? 'double' : 'int', flags: options.flags || 0, temporality: options.delta ? 1 : 2,
      monotonic: false, schemaURL: options.schemaURL || schemaURL });
  }
  add('worker-a', 'orders', 'used', 30000, 10);
  add('worker-a', 'orders', 'used', 15000, 7);
  add('worker-a', 'orders', 'used', 3000, 3, { epochAge: 10000 });
  add('worker-a', 'orders', 'idle', 4000, 12);
  add('worker-a', 'orders', 'reserved', 2000, 2);
  add('worker-b', 'orders', 'used', 3000, 21);
  add('worker-b', 'orders', 'idle', 3000, 0);
  add('worker-a', 'archive', 'used', 90000, 5);
  add('worker-a', 'missing-report', 'used', 3000, 0, { flags: 1 });
  add('worker-a', 'conflicting', 'used', 3000, 4);
  add('worker-a', 'conflicting', 'used', 3000, 9);
  // Missing-key observations remain generic/ineligible, never fabricated zeros.
  add('worker-a', null, 'used', 3000, 6);
  add('worker-a', 'unclassified-state', null, 3000, 1);
  // Separate binding: invalid named-pool counts must not poison the valid family.
  add('worker-a', 'probe', 'used', 3000, -1.25, { metric: 'work.generic.stock', unit: '1', double: true });
  add('worker-a', 'probe', 'used', 15000, 2.5, { metric: 'work.generic.stock', unit: '1', double: true });
  add('worker-a', 'old-schema', 'used', 3000, 7, { schemaURL: 'https://opentelemetry.io/schemas/1.36.0' });
  for (let i = 0; i < 120; i++) facts.records.push({ id: randomBytes(8).toString('hex'), traceId: randomBytes(16).toString('hex'),
    environment, service: facts.service, namespace: facts.namespace, timeMs: toMs - 15000 + i * 10, durationNs: '20000000', kind: 2, status: 1,
    name: 'GET /items', attributes: { 'http.request.method': 'GET', 'http.route': '/items' } });
  return facts;
}
function resource(facts, instance = 'worker-a') { return { 'service.name': facts.service, 'service.namespace': facts.namespace,
  'service.instance.id': instance, 'deployment.environment.name': facts.environment, 'telemetry.sdk.language': 'python' }; }
function traces(facts) {
  return { resourceSpans: [{ resource: { attributes: attrs(resource(facts)) }, scopeSpans: [{ scope: { name: 'ordinary.http', version: '1' }, spans: facts.records.map(r => ({
    traceId: r.traceId, spanId: r.id, name: r.name, kind: r.kind, status: { code: r.status }, startTimeUnixNano: String(BigInt(r.timeMs) * 1000000n),
    endTimeUnixNano: String(BigInt(r.timeMs) * 1000000n + BigInt(r.durationNs)), attributes: attrs(r.attributes) })) }] }] };
}
function metrics(facts) {
  const resourceMetrics = facts.points.map(p => ({ resource: { attributes: attrs(resource(facts, p.instance)) }, scopeMetrics: [{ schemaUrl: p.schemaURL,
    scope: { name: 'ordinary.pool', version: '1' }, metrics: [{ name: p.metric, unit: p.unit, sum: { aggregationTemporality: p.temporality, isMonotonic: false,
      dataPoints: [{ startTimeUnixNano: p.startNS, timeUnixNano: p.endNS, attributes: attrs(p.attributes), flags: p.flags,
        [p.valueType === 'double' ? 'asDouble' : 'asInt']: p.valueType === 'double' ? Number(p.value) : p.value }] } }] }] }));
  resourceMetrics.push({ resource: { attributes: attrs(resource(facts)) }, scopeMetrics: [{ scope: { name: 'ordinary.http', version: '1' }, metrics: [{
    name: 'http.server.request.duration', unit: 's', histogram: { aggregationTemporality: 1, dataPoints: [{
      startTimeUnixNano: String(BigInt(facts.toMs - 5000) * 1000000n), timeUnixNano: String(BigInt(facts.toMs - 1000) * 1000000n),
      count: '120', sum: 2.4, explicitBounds: [0.02], bucketCounts: ['120', '0'], attributes: attrs({ 'http.request.method': 'GET' }) }] } }] }] });
  return { resourceMetrics };
}
module.exports = { fixture, traces, metrics, schemaURL, keys };
