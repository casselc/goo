const { randomBytes } = require('node:crypto');
const previous = require('./187-http-dashboard-fixture');
const attributes = values => Object.entries(values).map(([key, stringValue]) => ({ key, value: { stringValue } }));
function fixture() {
  const facts = previous.fixture();
  facts.environment = 'service-projection-' + randomBytes(5).toString('hex');
  let incoming = 0, custom = 0;
  for (const record of facts.records) {
    record.environment = facts.environment;
    if (record.attributes['test.case'] === 'incoming-current') record.durationNs = String(BigInt(++incoming) * 1000000n);
    if (record.attributes['test.case'] === 'incoming-custom') record.durationNs = String(BigInt(++custom) * 5000000n);
  }
  facts.logs = Array.from({ length: 160 }, (_, index) => ({ id: randomBytes(8).toString('hex'), environment: facts.environment, service: 'edge-api',
    namespace: index < 155 ? 'commerce' : 'admin', timeMs: facts.toMs - 15000 + index * 20,
    severity: index % 7 === 0 ? 'ERROR' : 'INFO', body: 'Observed request log ' + index, attributes: { 'fixture.record': String(index) } }));
  return facts;
}
function logs(facts) {
  return { resourceLogs: facts.logs.map(record => ({ resource: { attributes: attributes({ 'service.name': record.service, 'service.namespace': record.namespace,
    'service.instance.id': 'edge-api-one', 'deployment.environment.name': facts.environment }) }, scopeLogs: [{ scope: { name: 'ordinary.logs', version: '1' },
      logRecords: [{ timeUnixNano: String(BigInt(record.timeMs) * 1000000n), severityNumber: record.severity === 'ERROR' ? 17 : 9,
        severityText: record.severity, body: { stringValue: record.body }, attributes: attributes(record.attributes) }] }] })) };
}
function value(record, field) {
  if (field.kind === 'attribute') return Object.hasOwn(record.attributes, field.key) ? record.attributes[field.key] : null;
  if (field.kind === 'resource') return ({ 'service.namespace': record.namespace, 'service.instance.id': record.service + '-one', 'deployment.environment.name': record.environment })[field.key] ?? null;
  return ({ service: record.service, name: record.name, severity: record.severity, kind: ({ 1: 'Internal', 2: 'Server', 3: 'Client', 4: 'Producer', 5: 'Consumer' })[record.kind] })[field.key] ?? null;
}
function matches(record, predicate) {
  if (!predicate) return true;
  if (predicate.all) return predicate.all.every(child => matches(record, child));
  if (predicate.any) return predicate.any.some(child => matches(record, child));
  if (predicate.not) return !matches(record, predicate.not);
  const actual = value(record, predicate.field), expected = predicate.value ?? '';
  if (predicate.op === 'eq') return actual !== null && actual === expected;
  if (predicate.op === 'ne') return actual !== null && actual !== expected;
  if (predicate.op === 'missing') return actual === null;
  if (predicate.op === 'exists') return actual !== null;
  throw Error('Unexpected fixture predicate ' + predicate.op);
}
function members(facts, query) {
  return (query.signal === 'spans' ? facts.records : facts.logs).filter(record => record.environment === query.environment && record.timeMs >= query.fromMs && record.timeMs < query.toMs && matches(record, query.where));
}
module.exports = { fixture, logs, traces: previous.traces, metrics: previous.metrics, members, value };
