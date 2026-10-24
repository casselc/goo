// Real ordinary-OTLP observations. Synthetic component cases live separately.
const { randomBytes } = require('node:crypto');
const attributes = values => Object.entries(values).map(([key, stringValue]) => ({ key, value: { stringValue } }));

function fixture() {
  const environment = 'trace-comparison-' + randomBytes(5).toString('hex');
  const now = Date.now(), traces = [];
  function add(role, definitions, atMs) {
    const traceId = randomBytes(16).toString('hex');
    const records = definitions.map(([id, parent, start, duration, service, operation, kind = 1], index) => ({
      traceId, spanId: id.toString(16).padStart(16, '0'),
      parentSpanId: parent ? parent.toString(16).padStart(16, '0') : '',
      startNs: String(BigInt(atMs) * 1000000n + BigInt(start) * 1000000n),
      durationNs: String(BigInt(duration) * 1000000n), service, operation, kind,
      attributes: { 'work.role': role, 'work.ordinal': String(index), 'work.same': 'stable',
        'work.changed': role.includes('reference') ? 'old' : 'new',
        'work.namespace': role.includes('reference') ? 'attribute-old' : 'attribute-new',
        ...(role.includes('reference') ? { 'work.empty': '', 'work.removed': 'retained' } : { 'work.added': 'new' }) },
      resource: { 'service.namespace': 'commerce',
        'service.version': role.includes('reference') ? 'v1' : 'v2',
        'work.namespace': role.includes('reference') ? 'resource-old' : 'resource-new' },
    }));
    const trace = { role, traceId, fromMs: atMs - 1000, toMs: atMs + 5000, records };
    traces.push(trace); return trace;
  }
  add('reference', [
    [1, 0, 0, 100, 'checkout-api', 'GET /checkout', 2],
    [2, 1, 10, 20, 'catalog-api', 'lookup', 3],
    [3, 1, 40, 10, 'checkout-api', 'prepare'],
    [4, 1, 55, 10, 'checkout-api', 'prepare'],
    [5, 1, 70, 10, 'cache-worker', 'cache lookup'],
    [6, 1, 90, 0, 'checkout-api', 'marker'],
    [7, 1, 95, 0, 'queue-worker', 'dispatch', 3],
  ], now - 45000);
  add('current', [
    [11, 0, 0, 130, 'checkout-api', 'GET /checkout', 2],
    [12, 11, 10, 10, 'catalog-api', 'lookup', 3],
    [13, 11, 35, 15, 'checkout-api', 'prepare'],
    [14, 11, 45, 15, 'checkout-api', 'prepare'],
    [15, 11, 70, 20, 'queue-worker', 'enqueue', 4],
    [16, 11, 95, 0, 'checkout-api', 'marker'],
    [17, 11, 100, 0, 'queue-worker', 'dispatch', 4],
  ], now - 15000);
  add('problem-reference', [
    [21, 0, 0, 100, 'checkout-api', 'GET /checkout', 2],
    [22, 21, 10, 20, 'catalog-api', 'lookup', 3],
    [23, 999, 140, 10, 'queue-worker', 'orphan'],
  ], now - 45000);
  add('problem-current', [
    [31, 0, 0, 100, 'checkout-api', 'GET /checkout', 2],
    [32, 31, 90, 30, 'catalog-api', 'lookup', 3],
    [33, 999, 140, 10, 'queue-worker', 'orphan'],
  ], now - 15000);
  add('partial-current', [
    [41, 0, 0, 1000, 'checkout-api', 'GET /checkout', 2],
    ...Array.from({ length: 125 }, (_, i) => [100 + i, 41, 5 + i * 5, 2, 'catalog-api', 'fanout']),
  ], now - 15000);
  return { version: 1, environment, traces };
}

function question(facts, trace) {
  return { signal: 'spans', environment: facts.environment, fromMs: trace.fromMs,
    toMs: trace.toMs, limit: 100, offset: 0,
    where: { all: [{ field: { kind: 'column', key: 'trace_id' }, op: 'eq', value: trace.traceId }] } };
}

function payload(facts) {
  return { resourceSpans: facts.traces.flatMap(trace => {
    const services = [...new Set(trace.records.map(record => record.service))];
    return services.map(service => {
      const records = trace.records.filter(record => record.service === service);
      return { resource: { attributes: attributes({ ...records[0].resource, 'service.name': service,
        'deployment.environment.name': facts.environment }) }, scopeSpans: [{
        scope: { name: 'goo-journey-185' }, spans: records.map(record => ({
          traceId: record.traceId, spanId: record.spanId,
          ...(record.parentSpanId ? { parentSpanId: record.parentSpanId } : {}),
          name: record.operation, kind: record.kind, status: { code: 1 },
          startTimeUnixNano: record.startNs,
          endTimeUnixNano: String(BigInt(record.startNs) + BigInt(record.durationNs)),
          attributes: attributes(record.attributes),
        })),
      }] };
    });
  }) };
}
module.exports = { fixture, question, payload };
