const { randomBytes } = require('node:crypto');
const attributes = values => Object.entries(values).map(([key, stringValue]) => ({ key, value: { stringValue } }));
function fixture() {
  const environment = 'db-dashboard-' + randomBytes(5).toString('hex'), now = Date.now(), records = [];
  function add(count, label, options = {}) {
    for (let i = 0; i < count; i++) records.push({
      id: randomBytes(8).toString('hex'), traceId: randomBytes(16).toString('hex'), environment,
      service: options.service || 'db-api', namespace: Object.hasOwn(options, 'namespace') ? options.namespace : 'commerce',
      name: 'DB ' + label, kind: options.kind ?? 3, timeMs: now - 20000 + i * 20,
      durationNs: String(options.durationNs || 30000000), status: options.status ?? (i % 5 === 0 ? 2 : 1),
      attributes: { 'test.case': label, ...(options.attributes || {}) },
    });
  }
  add(120, 'client-select', { attributes: { 'db.system.name': 'postgresql', 'db.namespace': 'catalog', 'db.operation.name': 'SELECT',
    'db.query.summary': 'SUPPLIED_SUMMARY_NOT_A_DEFAULT_189', 'db.query.text': 'SELECT /* RAW_QUERY_NOT_A_DEFAULT_189 */ item FROM inventory', 'db.response.status_code': '00000' } });
  for (const [index, record] of records.entries()) if (index % 10 === 0) {
    record.attributes['db.system.name'] = 'mysql'; record.attributes['db.namespace'] = 'archive';
  }
  add(12, 'client-fetch', { status: 1, attributes: { 'db.system.name': 'custom.store', 'db.namespace': 'cache', 'db.operation.name': 'FETCH', 'db.response.status_code': 'ERROR' } });
  add(3, 'missing-operation-namespace', { attributes: { 'db.system.name': 'postgresql' } });
  add(2, 'empty-operation-namespace', { attributes: { 'db.system.name': 'postgresql', 'db.operation.name': '', 'db.namespace': '' } });
  add(4, 'missing-system', { attributes: { 'db.operation.name': 'SELECT', 'db.namespace': 'catalog' } });
  add(3, 'empty-system', { attributes: { 'db.system.name': '', 'db.operation.name': 'SELECT', 'db.namespace': 'catalog' } });
  add(5, 'legacy-only', { attributes: { 'db.system': 'postgresql', 'db.operation.name': 'SELECT', 'db.namespace': 'catalog' } });
  add(20, 'server-db-field', { kind: 2, status: 2, durationNs: 900000000, attributes: { 'db.system.name': 'postgresql', 'db.operation.name': 'SELECT' } });
  add(6, 'internal-db-field', { kind: 1, attributes: { 'db.system.name': 'postgresql', 'db.operation.name': 'SELECT' } });
  add(8, 'http-client', { attributes: { 'http.request.method': 'GET', 'http.route': '/items' } });
  add(6, 'rpc-client', { attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(11, 'wrong-namespace', { namespace: 'admin', attributes: { 'db.system.name': 'postgresql', 'db.operation.name': 'SELECT', 'service.namespace': 'commerce' } });
  add(2, 'empty-resource-namespace', { namespace: '', attributes: { 'db.system.name': 'postgresql', 'db.operation.name': 'SELECT' } });
  add(2, 'missing-resource-namespace', { namespace: null, attributes: { 'db.system.name': 'postgresql', 'db.operation.name': 'SELECT' } });
  add(5, 'other-role', { service: 'db-other-role', kind: 2, attributes: { 'db.system.name': 'postgresql', 'db.operation.name': 'SELECT' } });
  add(9, 'generic-worker', { service: 'worker', kind: 5, attributes: { 'messaging.system': 'redis' } });
  add(5, 'hybrid-http', { service: 'hybrid', kind: 2, attributes: { 'http.request.method': 'GET', 'http.route': '/items' } });
  add(5, 'hybrid-rpc', { service: 'hybrid', kind: 2, attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(5, 'hybrid-db', { service: 'hybrid', kind: 3, attributes: { 'db.system.name': 'postgresql', 'db.namespace': 'catalog', 'db.operation.name': 'SELECT' } });
  return { version: 1, environment, fromMs: now - 120000, toMs: now + 1000, records };
}
function traces(facts) {
  const groups = new Map();
  for (const record of facts.records) {
    const key = JSON.stringify([record.service, record.namespace]);
    if (!groups.has(key)) groups.set(key, { resource: { attributes: attributes({ 'service.name': record.service,
      'service.instance.id': record.service + '-one', 'deployment.environment.name': record.environment,
      ...(record.namespace === null ? {} : { 'service.namespace': record.namespace }),
    }) }, scopeSpans: [{ scope: { name: 'ordinary.database', version: '1' }, spans: [] }] });
    const start = BigInt(record.timeMs) * 1000000n;
    groups.get(key).scopeSpans[0].spans.push({ traceId: record.traceId, spanId: record.id, name: record.name, kind: record.kind, status: { code: record.status },
      startTimeUnixNano: String(start), endTimeUnixNano: String(start + BigInt(record.durationNs)), attributes: attributes(record.attributes) });
  }
  return { resourceSpans: [...groups.values()] };
}
function metrics(facts) {
  const schema = 'https://opentelemetry.io/schemas/1.30.0', at = BigInt(facts.toMs - 6000) * 1000000n;
  const histogram = (name, unit, bounds) => ({ name, unit, histogram: { aggregationTemporality: 1, dataPoints: [{
    startTimeUnixNano: String(at - 2000000000n), timeUnixNano: String(at), count: '4', bucketCounts: ['1', '2', '1'], explicitBounds: bounds,
    attributes: attributes({ 'db.system.name': 'postgresql', 'db.namespace': 'catalog', 'db.operation.name': 'SELECT' }),
  }] } });
  return { resourceMetrics: [{ schemaUrl: schema, resource: { attributes: attributes({ 'service.name': 'db-api', 'service.namespace': 'commerce',
    'service.instance.id': 'db-api-one', 'deployment.environment.name': facts.environment }) },
  scopeMetrics: [{ schemaUrl: schema, scope: { name: 'ordinary.database', version: '1' }, metrics: [
    histogram('db.client.operation.duration', 's', [.01, .05]), histogram('db.client.operation.duration', 'ms', [10, 50]),
    histogram('db.client.duration', 's', [.01, .05]),
  ] }] }] };
}
function expected(facts, profile, service = 'db-api') {
  return facts.records.filter(record => record.service === service && record.namespace === 'commerce' &&
    (profile === 'db-client' ? record.kind === 3 && typeof record.attributes['db.system.name'] === 'string' && record.attributes['db.system.name'] !== '' :
      profile === 'rpc-server' ? record.kind === 2 && record.attributes['rpc.system'] && record.attributes['rpc.method'] :
        profile === 'http-server' ? record.kind === 2 && record.attributes['http.request.method'] : true));
}
module.exports = { fixture, traces, metrics, expected };
