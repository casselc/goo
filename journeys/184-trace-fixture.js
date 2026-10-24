// Ordinary OTLP trace fixtures. No supplied UI data or mocked source responses.
const { randomBytes } = require('node:crypto');

function attributes(values) {
  return Object.entries(values).map(([key, stringValue]) => ({ key, value: { stringValue } }));
}

function fixture() {
  const environment = 'trace-breakdown-' + randomBytes(5).toString('hex');
  const fromMs = Date.now() - 30000;
  const originNs = BigInt(fromMs + 5000) * 1000000n;
  const traces = [];
  function trace(name, definitions) {
    const traceId = randomBytes(16).toString('hex');
    const records = definitions.map(([id, parent, start, duration, service, operation], index) => ({
      traceId, spanId: id.toString(16).padStart(16, '0'),
      parentSpanId: parent ? parent.toString(16).padStart(16, '0') : '',
      startNs: String(originNs + BigInt(start) * 1000000n),
      durationNs: String(BigInt(duration) * 1000000n),
      service: service || 'checkout-api', operation: operation || name,
      attributes: { 'work.case': name, 'work.ordinal': String(index) },
      resource: { 'service.namespace': 'commerce' },
    }));
    traces.push({ case: name, traceId, records });
  }
  trace('overlap', [
    [1, 0, 0, 100, 'checkout-api', 'GET /checkout'],
    [2, 1, 10, 50, 'checkout-api', 'prepare'],
    [3, 1, 40, 40, 'catalog-api', 'lookup'],
    [4, 2, 20, 10, 'storage-worker', 'read'],
    [5, 1, 90, 0, 'checkout-api', 'marker'],
  ]);
  trace('missing-parent', [[11, 999, 0, 10]]);
  trace('outside-parent', [[21, 0, 0, 30], [22, 21, 20, 20, 'catalog-api']]);
  trace('grandchild-outside-child', [
    [31, 0, 0, 100], [32, 31, 10, 20], [33, 32, 25, 20, 'catalog-api'],
  ]);
  trace('duplicate-identity', [[41, 0, 0, 100], [42, 41, 10, 20], [42, 41, 40, 20]]);
  trace('parent-cycle', [[51, 52, 0, 20], [52, 51, 0, 20]]);
  trace('bounded-page', [
    [61, 0, 0, 1000],
    ...Array.from({ length: 125 }, (_, i) => [100 + i, 61, 5 + i * 5, 2,
      i % 2 ? 'catalog-api' : 'checkout-api', 'fanout']),
  ]);
  return { version: 1, environment, fromMs, toMs: fromMs + 20000, traces };
}

function payload(facts) {
  const resources = new Map();
  for (const item of facts.traces) for (const record of item.records) {
    const key = record.service;
    if (!resources.has(key)) resources.set(key, []);
    resources.get(key).push({
      traceId: record.traceId, spanId: record.spanId,
      ...(record.parentSpanId ? { parentSpanId: record.parentSpanId } : {}),
      name: record.operation, kind: 1, status: { code: 1 },
      startTimeUnixNano: record.startNs,
      endTimeUnixNano: String(BigInt(record.startNs) + BigInt(record.durationNs)),
      attributes: attributes(record.attributes),
    });
  }
  return { resourceSpans: [...resources].map(([service, spans]) => ({
    resource: { attributes: attributes({ 'service.name': service,
      'deployment.environment.name': facts.environment, 'service.namespace': 'commerce' }) },
    scopeSpans: [{ scope: { name: 'goo-journey-184' }, spans }],
  })) };
}

module.exports = { fixture, payload };
