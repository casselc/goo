const { randomBytes } = require('node:crypto');
const attributes = map => Object.entries(map).map(([key, stringValue]) => ({ key, value: { stringValue } }));
function fixture() {
  const environment = 'generic-python-' + randomBytes(5).toString('hex'), now = Date.now();
  const records = Array.from({ length: 150 }, (_, i) => ({
    id: randomBytes(8).toString('hex'), traceId: randomBytes(16).toString('hex'),
    timeMs: now - 15000 + i * 30, durationNs: String(i % 7 === 0 ? 90000000 : 20000000),
    environment, service: i < 140 ? 'python-orders' : 'python-worker', namespace: 'ordinary-python',
    name: i < 140 ? 'GET /orders' : 'process batch', kind: i < 140 ? 2 : 1, status: i % 10 === 0 ? 2 : 1,
    attributes: i < 140 ? { 'http.request.method': 'GET', 'http.route': '/orders', 'fixture.source': 'python-otlp' } : { 'fixture.source': 'python-otlp' },
  }));
  return { environment, fromMs: now - 30000, toMs: now + 1000, records };
}
function resource(record) {
  return { attributes: attributes({ 'service.name': record.service, 'service.namespace': record.namespace,
    'deployment.environment.name': record.environment, 'telemetry.sdk.language': 'python', 'telemetry.sdk.name': 'opentelemetry', 'telemetry.sdk.version': '1.38.0' }) };
}
function traces(facts) {
  return { resourceSpans: facts.records.map(record => {
    const start = BigInt(record.timeMs) * 1000000n;
    return { resource: resource(record), scopeSpans: [{ scope: { name: 'opentelemetry.instrumentation.fastapi', version: '0.59b0' }, spans: [{
      traceId: record.traceId, spanId: record.id, name: record.name, kind: record.kind, status: { code: record.status },
      startTimeUnixNano: String(start), endTimeUnixNano: String(start + BigInt(record.durationNs)), attributes: attributes(record.attributes),
    }] }] };
  }) };
}
function logs(facts) {
  return { resourceLogs: facts.records.slice(0, 140).map(record => ({ resource: resource(record), scopeLogs: [{
    scope: { name: 'python.logging', version: '1' }, logRecords: [{ timeUnixNano: String(BigInt(record.timeMs) * 1000000n),
      observedTimeUnixNano: String(BigInt(record.timeMs) * 1000000n), severityNumber: record.status === 2 ? 17 : 9,
      severityText: record.status === 2 ? 'ERROR' : 'INFO', body: { stringValue: 'Python request ' + record.id },
      traceId: record.traceId, spanId: record.id, attributes: attributes({ 'fixture.source': 'python-otlp' }),
    }] }] })) };
}
module.exports = { fixture, traces, logs };
