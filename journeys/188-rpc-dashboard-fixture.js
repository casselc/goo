const { randomBytes } = require('node:crypto');
const attributes = values => Object.entries(values).map(([key, stringValue]) => ({ key, value: { stringValue } }));
function fixture() {
  const environment = 'rpc-dashboard-' + randomBytes(5).toString('hex');
  const now = Date.now(), records = [];
  function add(count, label, options = {}) {
    for (let i = 0; i < count; i++) records.push({
      id: randomBytes(8).toString('hex'), traceId: randomBytes(16).toString('hex'),
      environment, service: options.service || 'rpc-api',
      namespace: Object.hasOwn(options, 'namespace') ? options.namespace : 'commerce',
      name: options.name || 'RPC ' + label, kind: options.kind ?? 2,
      timeMs: now - 20000 + i * 20, durationNs: String(options.durationNs || 25000000),
      status: options.status ?? (i % 5 === 0 ? 2 : 1),
      attributes: { 'test.case': label, ...(options.attributes || {}) },
    });
  }
  add(120, 'incoming-method', { attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem', 'rpc.grpc.status_code': '0' } });
  // Method names are not global RPC operation identities: this group spans two systems/services.
  for (const [index, record] of records.entries()) {
    record.attributes['rpc.service'] = index % 10 === 0 ? 'catalog.AlternateInventory' : 'catalog.Inventory';
    if (index % 10 === 0) record.attributes['rpc.system'] = 'custom.protocol';
  }
  // A recorded nonzero RPC code with SpanStatus OK is not silently promoted to a failure.
  add(12, 'incoming-custom', { attributes: { 'rpc.system': 'custom.protocol', 'rpc.method': 'CustomCall', 'rpc.grpc.status_code': '13' }, status: 1 });
  add(3, 'missing-method', { attributes: { 'rpc.system': 'grpc' } });
  add(2, 'empty-method', { attributes: { 'rpc.system': 'grpc', 'rpc.method': '' } });
  add(4, 'missing-system', { attributes: { 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(3, 'empty-system', { attributes: { 'rpc.system': '', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(40, 'outgoing-client', { kind: 3, durationNs: 900000000, status: 2, attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(6, 'internal-rpc-field', { kind: 1, attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(8, 'http-client', { kind: 3, attributes: { 'http.request.method': 'GET', 'http.route': '/items' } });
  add(5, 'unsupported-current-key', { attributes: { 'rpc.system.name': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(11, 'wrong-namespace', { namespace: 'admin', attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem', 'service.namespace': 'commerce' } });
  add(2, 'empty-namespace', { namespace: '', attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(2, 'missing-namespace', { namespace: null, attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(7, 'client-only', { service: 'rpc-outbound', kind: 3, attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(9, 'generic-worker', { service: 'worker', kind: 5, attributes: { 'messaging.system': 'redis' } });
  return { version: 1, environment, fromMs: now - 120000, toMs: now + 1000, records };
}
function traces(facts) {
  const groups = new Map();
  for (const record of facts.records) {
    const key = JSON.stringify([record.service, record.namespace]);
    if (!groups.has(key)) groups.set(key, { resource: { attributes: attributes({
      'service.name': record.service, 'service.instance.id': record.service + '-one',
      'deployment.environment.name': record.environment,
      ...(record.namespace === null ? {} : { 'service.namespace': record.namespace }),
    }) }, scopeSpans: [{ scope: { name: 'ordinary.rpc', version: '1' }, spans: [] }] });
    const start = BigInt(record.timeMs) * 1000000n;
    groups.get(key).scopeSpans[0].spans.push({ traceId: record.traceId, spanId: record.id,
      name: record.name, kind: record.kind, status: { code: record.status },
      startTimeUnixNano: String(start), endTimeUnixNano: String(start + BigInt(record.durationNs)),
      attributes: attributes(record.attributes) });
  }
  return { resourceSpans: [...groups.values()] };
}
function metrics(facts) {
  const schema = 'https://opentelemetry.io/schemas/1.30.0', at = BigInt(facts.toMs - 6000) * 1000000n;
  const histogram = (unit, bounds) => ({ name: 'rpc.server.duration', unit, histogram: { aggregationTemporality: 1, dataPoints: [{
    startTimeUnixNano: String(at - 2000000000n), timeUnixNano: String(at), count: '4', bucketCounts: ['1', '2', '1'], explicitBounds: bounds,
    attributes: attributes({ 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' }),
  }] } });
  return { resourceMetrics: [{ schemaUrl: schema, resource: { attributes: attributes({
    'service.name': 'rpc-api', 'service.namespace': 'commerce', 'service.instance.id': 'rpc-api-one', 'deployment.environment.name': facts.environment,
  }) }, scopeMetrics: [{ schemaUrl: schema, scope: { name: 'ordinary.rpc', version: '1' }, metrics: [
    { name: 'aaa.runtime.workers', unit: '{worker}', gauge: { dataPoints: [{ timeUnixNano: String(at), asDouble: 3 }] } },
    histogram('s', [.01, .05]), histogram('ms', [10, 50]),
  ] }] }] };
}
function expected(facts, profile, service = 'rpc-api') {
  return facts.records.filter(record => record.service === service && record.namespace === 'commerce' &&
    (profile === 'rpc-server' ? record.kind === 2 && typeof record.attributes['rpc.system'] === 'string' && record.attributes['rpc.system'] !== '' && typeof record.attributes['rpc.method'] === 'string' && record.attributes['rpc.method'] !== '' :
      profile === 'http-server' ? record.kind === 2 && typeof record.attributes['http.request.method'] === 'string' && record.attributes['http.request.method'] !== '' : true));
}
module.exports = { fixture, traces, metrics, expected };
