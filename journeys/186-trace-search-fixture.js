// Ordinary OTLP spans. Raw invalid/zero-identity controls belong to backend
// source-parity checks and are deliberately not passed off as ingested telemetry.
const { randomBytes } = require('node:crypto');
const attributes = values => Object.entries(values).map(([key, stringValue]) => ({ key, value: { stringValue } }));
const conditionA = { all: [
  { field: { kind: 'column', key: 'service' }, op: 'eq', value: 'db-service' },
  { field: { kind: 'column', key: 'duration_ns' }, op: 'gte', value: '50000000' },
] };
const conditionB = { field: { kind: 'column', key: 'status_class' }, op: 'eq', value: 'error' };

function fixture() {
  const environment = 'trace-search-' + randomBytes(5).toString('hex');
  const now = Date.now(), fromMs = now - 60000, toMs = now, traces = [];
  const atMs = now - 20000;
  function add(role, definitions, options = {}) {
    const traceId = options.traceId || randomBytes(16).toString('hex');
    const sourceEnvironment = options.environment || environment;
    const originMs = options.atMs ?? atMs;
    const records = definitions.map(([id, parent, start, duration, service, operation, status = 1, kind = 1], index) => ({
      traceId, spanId: id.toString(16).padStart(16, '0'),
      parentSpanId: parent ? parent.toString(16).padStart(16, '0') : '',
      startNs: String(BigInt(originMs) * 1000000n + BigInt(start) * 1000000n),
      durationNs: String(BigInt(duration) * 1000000n), service, operation, status, kind,
      environment: sourceEnvironment,
      attributes: { team: 'core', 'work.role': role, 'work.ordinal': String(index),
        ...(service === 'db-service' ? { 'db.system.name': 'postgresql', 'db.operation.name': 'SELECT' } : {}) },
      resource: { 'service.namespace': 'commerce' },
    }));
    const trace = { role, traceId, environment: sourceEnvironment, records };
    traces.push(trace); return trace;
  }
  const cross = () => [
    [1, 0, 0, 200, 'checkout-api', 'GET /checkout', 1, 2],
    [2, 1, 10, 80, 'db-service', 'SELECT inventory', 1, 3],
    [3, 1, 100, 20, 'queue-worker', 'publish receipt', 2, 4],
  ];
  add('cross-service', cross());
  add('same-span', [
    [1, 0, 0, 300, 'checkout-api', 'GET /checkout', 1, 2],
    [2, 1, 10, 120, 'db-service', 'SELECT inventory', 2, 3],
  ]);
  add('repeated-witness', [
    [1, 0, 0, 2000, 'checkout-api', 'GET /checkout', 1, 2],
    ...Array.from({ length: 120 }, (_, index) => [100 + index, 1, 10 + index * 5, 80, 'db-service', 'SELECT inventory', 1, 3]),
    ...Array.from({ length: 120 }, (_, index) => [300 + index, 1, 700 + index * 5, 20, 'queue-worker', 'publish receipt', 2, 4]),
  ]);
  for (let index = 0; index < 12; index++) add('additional-match-' + index, cross());
  add('condition-a-only', cross().slice(0, 2));
  add('condition-b-only', [cross()[0], cross()[2]]);
  add('fast-database-error', [cross()[0], [2, 1, 10, 10, 'db-service', 'SELECT inventory', 2, 3]]);
  add('wrong-environment', cross(), { environment: environment + '-other' });
  add('outside-window', cross(), { atMs: fromMs - 10000 });
  return { version: 1, environment, fromMs, toMs, conditionA, conditionB, traces };
}
function query(facts, where = { all: [facts.conditionA,
  { field: { kind: 'resource', key: 'service.namespace' }, op: 'eq', value: 'commerce' },
  { field: { kind: 'attribute', key: 'team' }, op: 'eq', value: 'core' },
] }) {
  return { signal: 'spans', environment: facts.environment, fromMs: facts.fromMs,
    toMs: facts.toMs, limit: 100, offset: 0, where };
}
function payload(facts) {
  const resources = new Map();
  for (const trace of facts.traces) for (const record of trace.records) {
    const key = record.environment + ':' + record.service;
    if (!resources.has(key)) resources.set(key, []);
    resources.get(key).push(record);
  }
  return { resourceSpans: [...resources.values()].map(records => ({
    resource: { attributes: attributes({ ...records[0].resource,
      'service.name': records[0].service, 'deployment.environment.name': records[0].environment }) },
    scopeSpans: [{ scope: { name: 'goo-journey-186' }, spans: records.map(record => ({
      traceId: record.traceId, spanId: record.spanId,
      ...(record.parentSpanId ? { parentSpanId: record.parentSpanId } : {}),
      name: record.operation, kind: record.kind, status: { code: record.status },
      startTimeUnixNano: record.startNs,
      endTimeUnixNano: String(BigInt(record.startNs) + BigInt(record.durationNs)),
      attributes: attributes(record.attributes),
    })) }],
  })) };
}
module.exports = { fixture, query, payload, conditionA, conditionB };
