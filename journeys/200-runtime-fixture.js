// Generated ordinary OTLP fixtures, not telemetry from a running Go application.
const { randomBytes } = require('node:crypto');
const schemaURL = 'https://opentelemetry.io/schemas/1.37.0';
const attrs = values => Object.entries(values).map(([key, value]) => ({ key, value: { stringValue: String(value) } }));
function fixture() {
  const toMs = Date.now(), environment = 'runtime-observed-' + randomBytes(5).toString('hex');
  const facts = { environment, service: 'runtime-worker', namespace: 'jobs', fromMs: toMs - 120000, toMs, schemaURL, points: [], records: [] };
  function point(metric, unit, value, instance = 'worker-a', attributes = {}, options = {}) {
    facts.points.push({ metric, unit, value: String(value), valueType: 'int', kind: options.kind || 'stock',
      resource: { 'deployment.environment.name': environment, 'service.name': facts.service, 'service.namespace': facts.namespace,
        'service.instance.id': instance, 'telemetry.sdk.language': 'go', ...(options.resource || {}) },
      attributes, schemaURL: options.schemaURL || schemaURL, resourceSchemaURL: schemaURL, scope: 'ordinary.runtime', version: '1',
      startNS: String(BigInt(toMs - (options.epochAge || 90000)) * 1000000n),
      endNS: String(BigInt(toMs - (options.age || 3000)) * 1000000n), flags: options.flags || 0,
      temporality: options.temporality || 2, monotonic: options.monotonic || false, case: options.case || 'primary' });
  }
  point('go.memory.used', 'By', '65536', 'worker-a', { 'go.memory.type': 'stack' }, { age: 20000 });
  point('go.memory.used', 'By', '32768', 'worker-a', { 'go.memory.type': 'stack' }, { epochAge: 10000 });
  point('go.memory.used', 'By', '16384', 'worker-a', { 'go.memory.type': 'other' });
  point('go.memory.used', 'By', '131072', 'worker-b', { 'go.memory.type': 'stack' });
  point('go.memory.gc.goal', 'By', '262144');
  point('go.memory.gc.goal', 'By', '524288', 'worker-b');
  point('go.goroutine.count', '{goroutine}', '7');
  point('go.goroutine.count', '{goroutine}', '11', 'worker-b');
  point('go.memory.used', 'By', '8192', 'worker-a', { 'go.memory.type': 'custom-reported' });
  point('go.memory.used', 'By', '4096', 'worker-a', {});
  point('go.memory.used', 'By', '999', 'wrong-schema', { 'go.memory.type': 'stack' }, { schemaURL: 'https://opentelemetry.io/schemas/1.36.0', case: 'wrong-schema' });
  point('go.memory.gc.goal', 'bytes', '888', 'wrong-unit', {}, { case: 'wrong-unit' });
  point('go.memory.used', 'By', '777', 'delta', { 'go.memory.type': 'stack' }, { temporality: 1, case: 'delta' });
  point('go.memory.used', 'By', '666', 'monotonic', { 'go.memory.type': 'stack' }, { monotonic: true, case: 'monotonic' });
  point('go.memory.gc.goal', 'By', '555', 'gauge', {}, { kind: 'gauge', case: 'wrong-kind' });
  point('go.memory.used', 'By', '222', 'different-resource', { 'go.memory.type': 'stack' }, { resource: { 'cloud.account.id': 'other' }, case: 'different-resource' });
  for (let i = 0; i < 120; i++) { const spanId = randomBytes(8).toString('hex'); facts.records.push({ id: spanId, traceId: randomBytes(16).toString('hex'),
    spanId, environment, service: facts.service, namespace: facts.namespace, name: 'runtime-work', kind: 1,
    instance: i % 2 ? 'worker-b' : 'worker-a', timeMs: toMs - 4000 - i, durationNs: '1000000', status: 1, attributes: { 'work.fixture': environment } }); }
  return facts;
}
function metrics(facts) {
  return { resourceMetrics: facts.points.map(p => {
    const dataPoint = { attributes: attrs(p.attributes), startTimeUnixNano: p.startNS, timeUnixNano: p.endNS, asInt: p.value, flags: p.flags };
    return { resource: { attributes: attrs(p.resource) }, schemaUrl: p.resourceSchemaURL, scopeMetrics: [{ schemaUrl: p.schemaURL,
      scope: { name: p.scope, version: p.version }, metrics: [{ name: p.metric, unit: p.unit,
        ...(p.kind === 'gauge' ? { gauge: { dataPoints: [dataPoint] } } : { sum: { aggregationTemporality: p.temporality, isMonotonic: p.monotonic, dataPoints: [dataPoint] } }) }] }] };
  }) };
}
function traces(facts) {
  return { resourceSpans: facts.records.map(r => ({ resource: { attributes: attrs({ 'deployment.environment.name': facts.environment,
    'service.name': facts.service, 'service.namespace': facts.namespace, 'service.instance.id': r.instance, 'telemetry.sdk.language': 'go' }) },
    schemaUrl: schemaURL, scopeSpans: [{ scope: { name: 'ordinary.runtime', version: '1' }, spans: [{ traceId: r.traceId, spanId: r.spanId,
      name: r.name, kind: r.kind, startTimeUnixNano: String(BigInt(r.timeMs) * 1000000n),
      endTimeUnixNano: String(BigInt(r.timeMs) * 1000000n + BigInt(r.durationNs)), attributes: attrs(r.attributes), status: { code: r.status } }] }] })) };
}
module.exports = { fixture, metrics, traces, schemaURL };
