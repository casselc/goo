const { randomBytes } = require('node:crypto');

const attributes = values => Object.entries(values).map(([key, stringValue]) => ({ key, value: { stringValue } }));
const schema = 'https://opentelemetry.io/schemas/1.30.0';
function fixture() {
  const environment = 'http-dashboard-' + randomBytes(5).toString('hex');
  const now = Date.now(), records = [];
  function add(count, label, options = {}) {
    for (let i = 0; i < count; i++) {
      records.push({
        id: randomBytes(8).toString('hex'), traceId: randomBytes(16).toString('hex'),
        environment, service: options.service || 'edge-api', namespace: options.namespace || 'commerce',
        name: 'HTTP ' + label, kind: options.kind ?? 2,
        timeMs: now - 20000 + i * 20, durationNs: String(options.durationNs || 20000000),
        status: options.status ?? (i % 5 === 0 ? 2 : 1),
        attributes: { 'test.case': label, ...(options.attributes || {}) },
      });
    }
  }
  add(120, 'incoming-current', { attributes: { 'http.request.method': 'GET', 'http.route': '/items', 'http.response.status_code': '200' } });
  add(12, 'incoming-custom', { attributes: { 'http.request.method': 'CUSTOM', 'http.route': '/custom', 'http.response.status_code': '503' }, status: 2 });
  add(3, 'incoming-missing-route', { attributes: { 'http.request.method': 'POST' } });
  add(2, 'incoming-empty-route', { attributes: { 'http.request.method': 'POST', 'http.route': '' } });
  add(40, 'outgoing-client', { kind: 3, durationNs: 900000000, status: 2, attributes: { 'http.request.method': 'GET', 'http.route': '/items' } });
  add(8, 'legacy-only', { attributes: { 'http.method': 'GET', 'http.route': '/legacy', 'http.status_code': '500' } });
  add(4, 'empty-method', { attributes: { 'http.request.method': '', 'http.route': '/items' } });
  add(3, 'missing-method', { attributes: { 'http.route': '/items' } });
  add(6, 'internal-http-field', { kind: 1, attributes: { 'http.request.method': 'GET', 'http.route': '/items' } });
  add(11, 'wrong-namespace', { namespace: 'admin', attributes: { 'http.request.method': 'GET', 'http.route': '/items', 'service.namespace': 'commerce' } });
  add(9, 'generic-worker', { service: 'worker', kind: 5, attributes: { 'messaging.system': 'redis' } });
  add(7, 'client-only', { service: 'outbound', kind: 3, attributes: { 'http.request.method': 'GET', 'http.route': '/remote' } });
  return { version: 1, environment, fromMs: now - 120000, toMs: now + 1000, records, schema };
}
function resource(record) {
  return { attributes: attributes({ 'service.name': record.service, 'service.namespace': record.namespace, 'service.instance.id': record.service + '-one', 'deployment.environment.name': record.environment }) };
}
function traces(facts) {
  const groups = new Map();
  for (const record of facts.records) {
    const key = record.service + '/' + record.namespace;
    if (!groups.has(key)) groups.set(key, { resource: resource(record), scopeSpans: [{ scope: { name: 'ordinary.http', version: '1' }, spans: [] }] });
    const start = BigInt(record.timeMs) * 1000000n;
    groups.get(key).scopeSpans[0].spans.push({ traceId: record.traceId, spanId: record.id, name: record.name, kind: record.kind, status: { code: record.status }, startTimeUnixNano: String(start), endTimeUnixNano: String(start + BigInt(record.durationNs)), attributes: attributes(record.attributes) });
  }
  return { resourceSpans: [...groups.values()] };
}
function metrics(facts) {
  const record = facts.records[0], at = BigInt(facts.toMs - 6000) * 1000000n;
  const histogram = (name, unit) => ({ name, unit, histogram: { aggregationTemporality: 1, dataPoints: [{ startTimeUnixNano: String(at - 2000000000n), timeUnixNano: String(at), count: '4', bucketCounts: ['1', '2', '1'], explicitBounds: [.1, .5], attributes: attributes({ 'http.request.method': 'GET', 'http.response.status_code': '200', 'http.route': '/items' }) }] } });
  return { resourceMetrics: [{ schemaUrl: schema, resource: resource(record), scopeMetrics: [{ schemaUrl: schema, scope: { name: 'otel.http', version: '1' }, metrics: [
    { name: 'aaa.runtime.workers', unit: '{worker}', gauge: { dataPoints: [{ timeUnixNano: String(at), asDouble: 3 }] } },
    histogram('http.client.request.duration', 's'), histogram('http.server.request.duration', 's'),
  ] }] }] };
}
function expected(facts, profile, service = 'edge-api', namespace = 'commerce') {
  return facts.records.filter(record => record.service === service && record.namespace === namespace &&
    (profile !== 'http-server' || record.kind === 2 && typeof record.attributes['http.request.method'] === 'string' && record.attributes['http.request.method'] !== ''));
}
module.exports = { fixture, traces, metrics, expected };
