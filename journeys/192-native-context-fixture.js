const { randomBytes } = require('node:crypto');
const attributes = map => Object.entries(map).map(([key, stringValue]) => ({ key, value: { stringValue } }));
function fixture() {
  const environment = 'native-context-' + randomBytes(5).toString('hex');
  const fromNS = BigInt(Date.now() - 15000) * 1000000n + 234567n, toNS = fromNS + 2000321000n;
  const records = [], traceId = randomBytes(16).toString('hex');
  function add(label, at, options = {}) {
    records.push({ id: randomBytes(8).toString('hex'), traceId: options.traceId || randomBytes(16).toString('hex'),
      environment, service: options.service || 'native-worker', namespace: Object.hasOwn(options, 'namespace') ? options.namespace : 'inventory',
      instance: options.instance || 'worker-a', timestampNS: String(at), timeMs: Number(at / 1000000n),
      durationNs: '10000', kind: 1, status: 1, name: label, attributes: { 'fixture.case': label, 'record.only': 'ordinary-record' } });
  }
  for (let i = 0; i < 112; i++) add('inside-' + i, fromNS + 1000000n + BigInt(i) * 1000000n, i === 0 ? { traceId } : {});
  for (const [label, at] of [['before-lower', fromNS - 1n], ['at-lower', fromNS], ['after-lower', fromNS + 1n], ['before-upper', toNS - 1n], ['at-upper', toNS], ['after-upper', toNS + 1n]]) add(label, at, ['before-lower', 'at-upper'].includes(label) ? { traceId } : {});
  add('other-instance', fromNS + 5000000n, { instance: 'worker-b', traceId });
  add('empty-namespace', fromNS + 6000000n, { namespace: '' });
  add('missing-namespace', fromNS + 7000000n, { namespace: null });
  add('other-namespace', fromNS + 8000000n, { namespace: 'admin' });
  add('same-trace-outside-interval', toNS + 2000000n, { traceId });
  return { environment, fromNS: String(fromNS), toNS: String(toNS), emptyFromNS: String(toNS + 1000000000n), emptyToNS: String(toNS + 2000000000n), records, traceId };
}
function resource(facts, namespace = 'inventory', instance = 'worker-a') {
  const map = { 'service.name': 'native-worker', 'service.instance.id': instance, 'deployment.environment.name': facts.environment };
  if (namespace !== null) map['service.namespace'] = namespace;
  return map;
}
function traces(facts) {
  return { resourceSpans: facts.records.map(record => ({ resource: { attributes: attributes(resource(facts, record.namespace, record.instance)) }, scopeSpans: [{
    scope: { name: 'ordinary.records', version: '1' }, spans: [{ traceId: record.traceId, spanId: record.id, name: record.name, kind: record.kind,
      status: { code: record.status }, startTimeUnixNano: record.timestampNS, endTimeUnixNano: String(BigInt(record.timestampNS) + BigInt(record.durationNs)), attributes: attributes(record.attributes) }],
  }] })) };
}
function logs(facts) {
  return { resourceLogs: facts.records.map(record => ({ resource: { attributes: attributes(resource(facts, record.namespace, record.instance)) }, scopeLogs: [{
    scope: { name: 'ordinary.logs', version: '2' }, logRecords: [{ timeUnixNano: record.timestampNS, observedTimeUnixNano: record.timestampNS,
      traceId: record.traceId, spanId: record.id, severityNumber: 9, severityText: 'INFO', body: { stringValue: record.name }, attributes: attributes(record.attributes) }],
  }] })) };
}
function metrics(facts) {
  const dimensions = attributes({ 'metric.only': 'not-a-record-filter', 'service.namespace': 'metric-label', 'service.instance.id': 'metric-label-instance' });
  const points = [
    { startTimeUnixNano: facts.fromNS, timeUnixNano: facts.toNS, asInt: '7', attributes: dimensions },
    { startTimeUnixNano: facts.emptyFromNS, timeUnixNano: facts.emptyToNS, asInt: '0', attributes: dimensions },
  ];
  return { resourceMetrics: ['inventory', '', null].map(namespace => ({ resource: { attributes: attributes(resource(facts, namespace)) }, scopeMetrics: [{
    scope: { name: 'ordinary.native', version: '9' }, schemaUrl: 'https://opentelemetry.io/schemas/1.30.0', metrics: [
      { name: 'work.completed', unit: '{work}', sum: { aggregationTemporality: 1, isMonotonic: true, dataPoints: points } },
      { name: 'work.latency', unit: 's', histogram: { aggregationTemporality: 1, dataPoints: [{ startTimeUnixNano: facts.fromNS, timeUnixNano: facts.toNS,
        count: '3', bucketCounts: ['1', '2'], explicitBounds: [0.02], attributes: dimensions }] } },
      { name: 'work.unavailable', unit: '{work}', sum: { aggregationTemporality: 1, isMonotonic: true, dataPoints: [{ startTimeUnixNano: facts.fromNS, timeUnixNano: facts.toNS, asInt: '7', flags: 1, attributes: dimensions }] } },
    ],
  }] })) };
}
function members(facts, namespace = 'inventory', interval = { fromNS: facts.fromNS, toNS: facts.toNS }) {
  return facts.records.filter(record => record.namespace === namespace && record.instance === 'worker-a' && BigInt(record.timestampNS) >= BigInt(interval.fromNS) && BigInt(record.timestampNS) < BigInt(interval.toNS));
}
module.exports = { fixture, traces, logs, metrics, members };
