const { randomBytes } = require('node:crypto');
const attributes = values => Object.entries(values).map(([key, stringValue]) => ({ key, value: { stringValue } }));
function fixture() {
  const environment = 'messaging-dashboard-' + randomBytes(5).toString('hex'), now = Date.now(), records = [];
  function add(count, label, options = {}) {
    for (let i = 0; i < count; i++) records.push({ id: randomBytes(8).toString('hex'), traceId: randomBytes(16).toString('hex'), environment,
      service: options.service || 'messaging-app', namespace: Object.hasOwn(options, 'namespace') ? options.namespace : 'commerce',
      name: 'Messaging ' + label, kind: options.kind ?? 5, timeMs: now - 20000 + i * 20, durationNs: String(options.durationNs || 30000000),
      status: options.status ?? (i % 5 === 0 ? 2 : 1), attributes: { 'test.case': label, ...(options.attributes || {}) } });
  }
  add(120, 'consumer-orders', { attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': 'orders', 'messaging.operation.name': 'process', 'messaging.consumer.group.name': 'workers-a' } });
  for (const [i, record] of records.entries()) if (i % 10 === 0) { record.attributes['messaging.system'] = 'rabbitmq'; record.attributes['messaging.consumer.group.name'] = 'workers-b'; }
  add(12, 'consumer-returns', { status: 1, attributes: { 'messaging.system': 'custom.bus', 'messaging.destination.name': 'returns', 'messaging.operation.name': 'receive' } });
  add(3, 'missing-destination', { attributes: { 'messaging.system': 'kafka', 'messaging.destination': 'legacy-orders' } });
  add(2, 'empty-destination', { attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': '', 'messaging.operation.name': '' } });
  add(4, 'missing-system', { attributes: { 'messaging.destination.name': 'orders' } });
  add(3, 'empty-system', { attributes: { 'messaging.system': '', 'messaging.destination.name': 'orders' } });
  add(120, 'producer-orders', { kind: 4, durationNs: 50000000, attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': 'orders', 'messaging.operation.name': 'publish' } });
  add(20, 'server-system-field', { kind: 2, durationNs: 900000000, attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': 'orders' } });
  add(6, 'internal-system-field', { kind: 1, attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': 'orders' } });
  add(8, 'http-client', { kind: 3, attributes: { 'http.request.method': 'GET', 'http.route': '/items' } });
  add(6, 'rpc-client', { kind: 3, attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(5, 'db-server', { kind: 2, attributes: { 'db.system.name': 'postgresql', 'db.operation.name': 'SELECT' } });
  add(11, 'wrong-namespace', { namespace: 'admin', attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': 'orders', 'service.namespace': 'commerce' } });
  add(2, 'empty-resource-namespace', { namespace: '', attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': 'orders' } });
  add(2, 'missing-resource-namespace', { namespace: null, attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': 'orders' } });
  add(5, 'other-role', { service: 'messaging-other-role', kind: 2, attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': 'orders' } });
  add(9, 'generic-worker', { service: 'worker', kind: 1, attributes: { 'code.function.name': 'work' } });
  add(3, 'hybrid-http', { service: 'hybrid', kind: 2, attributes: { 'http.request.method': 'GET', 'http.route': '/items' } });
  add(3, 'hybrid-rpc', { service: 'hybrid', kind: 2, attributes: { 'rpc.system': 'grpc', 'rpc.method': 'catalog.Inventory/GetItem' } });
  add(3, 'hybrid-db', { service: 'hybrid', kind: 3, attributes: { 'db.system.name': 'postgresql', 'db.namespace': 'catalog', 'db.operation.name': 'SELECT' } });
  add(3, 'hybrid-consumer', { service: 'hybrid', kind: 5, attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': 'orders' } });
  add(3, 'hybrid-producer', { service: 'hybrid', kind: 4, attributes: { 'messaging.system': 'kafka', 'messaging.destination.name': 'orders' } });
  const linkedConsumer = records.find(record => record.attributes['test.case'] === 'consumer-orders' && record.attributes['messaging.system'] === 'kafka');
  const linkedProducer = records.find(record => record.attributes['test.case'] === 'producer-orders');
  linkedConsumer.links = [{ traceId: linkedProducer.traceId, spanId: linkedProducer.id, attributes: { 'test.link': 'reported-only' } }];
  return { version: 1, environment, fromMs: now - 120000, toMs: now + 1000, records, linkedConsumer: linkedConsumer.id, linkedProducer: linkedProducer.id };
}
function traces(facts) {
  const groups = new Map();
  for (const record of facts.records) {
    const key = JSON.stringify([record.service, record.namespace]);
    if (!groups.has(key)) groups.set(key, { resource: { attributes: attributes({ 'service.name': record.service, 'service.instance.id': record.service + '-one',
      'deployment.environment.name': record.environment, ...(record.namespace === null ? {} : { 'service.namespace': record.namespace }) }) }, scopeSpans: [{ scope: { name: 'ordinary.messaging', version: '1' }, spans: [] }] });
    const start = BigInt(record.timeMs) * 1000000n;
    groups.get(key).scopeSpans[0].spans.push({ traceId: record.traceId, spanId: record.id, name: record.name, kind: record.kind, status: { code: record.status },
      startTimeUnixNano: String(start), endTimeUnixNano: String(start + BigInt(record.durationNs)), attributes: attributes(record.attributes),
      ...(record.links ? { links: record.links.map(link => ({ traceId: link.traceId, spanId: link.spanId, attributes: attributes(link.attributes) })) } : {}) });
  }
  return { resourceSpans: [...groups.values()] };
}
function metrics(facts) {
  const schema = 'https://opentelemetry.io/schemas/1.30.0', at = BigInt(facts.toMs - 6000) * 1000000n;
  const histogram = (name, unit, bounds) => ({ name, unit, histogram: { aggregationTemporality: 1, dataPoints: [{ startTimeUnixNano: String(at - 2000000000n), timeUnixNano: String(at),
    count: '4', bucketCounts: ['1', '2', '1'], explicitBounds: bounds, attributes: attributes({ 'messaging.system': 'kafka', 'messaging.destination.name': 'orders', 'messaging.operation.name': 'process' }) }] } });
  return { resourceMetrics: [{ schemaUrl: schema, resource: { attributes: attributes({ 'service.name': 'messaging-app', 'service.namespace': 'commerce', 'service.instance.id': 'messaging-app-one', 'deployment.environment.name': facts.environment }) },
    scopeMetrics: [{ schemaUrl: schema, scope: { name: 'ordinary.messaging', version: '1' }, metrics: [histogram('messaging.process.duration', 's', [.01, .05]), histogram('messaging.process.duration', 'ms', [10, 50]), histogram('messaging.publish.duration', 's', [.01, .05])] }] }] };
}
function expected(facts, profile, service = 'messaging-app') {
  return facts.records.filter(record => record.service === service && record.namespace === 'commerce' &&
    (profile === 'messaging-consumer' ? record.kind === 5 && record.attributes['messaging.system'] :
      profile === 'messaging-producer' ? record.kind === 4 && record.attributes['messaging.system'] :
        profile === 'http-server' ? record.kind === 2 && record.attributes['http.request.method'] :
          profile === 'rpc-server' ? record.kind === 2 && record.attributes['rpc.system'] && record.attributes['rpc.method'] :
            profile === 'db-client' ? record.kind === 3 && record.attributes['db.system.name'] : true));
}
module.exports = { fixture, traces, metrics, expected };
