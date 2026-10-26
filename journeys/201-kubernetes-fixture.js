// Generated ordinary OTLP resource observations, not a live Kubernetes API or agent.
const { randomBytes } = require('node:crypto');
const schemaURL = 'https://opentelemetry.io/schemas/1.37.0';
const attrs = values => Object.entries(values).map(([key, value]) => ({ key, value: { stringValue: String(value) } }));
function fixture() {
  const toMs = Date.now(), environment = 'pod-observed-' + randomBytes(5).toString('hex');
  const primary = { 'deployment.environment.name': environment, 'k8s.pod.uid': 'uid-a', 'k8s.pod.name': 'shared-pod', 'k8s.cluster.name': 'shared-cluster', 'telemetry.sdk.language': 'python' };
  const facts = { environment, fromMs: toMs - 120000, toMs, schemaURL, primary, points: [], records: [] };
  function point(kind, resource, options = {}) {
    facts.points.push({ kind, family: kind, resource: { ...resource }, resourceSchemaURL: options.schemaURL ?? schemaURL,
      schemaURL, scope: 'ordinary.pod', version: '1', metric: 'fixture.pod.' + kind, unit: '{item}', attributes: {},
      startNS: String(BigInt(toMs - 10000) * 1000000n), endNS: String(BigInt(toMs - (options.age || 3000)) * 1000000n),
      value: '3', temporality: 2, monotonic: false, count: '2', bounds: [1], bucketCounts: ['1', '1'], sum: 2, case: options.case || 'primary' });
  }
  for (const kind of ['gauge', 'sum', 'histogram']) point(kind, primary);
  point('gauge', primary, { schemaURL: '', case: 'literal-empty-schema' });
  point('gauge', primary, { schemaURL: 'https://opentelemetry.io/schemas/1.36.0', case: 'different-schema' });
  point('gauge', { ...primary, 'k8s.pod.uid': 'uid-b', 'k8s.cluster.uid': 'cluster-b', 'k8s.namespace.name': 'team-b' }, { case: 'same-name-other-uid-cluster' });
  point('gauge', { ...primary, 'k8s.namespace.name': '' }, { case: 'empty-namespace' });
  point('gauge', { ...primary, 'k8s.namespace.name': 'team-a' }, { case: 'reported-namespace' });
  point('gauge', { ...primary, 'k8s.cluster.uid': '' }, { case: 'empty-cluster-uid' });
  point('gauge', { ...primary, 'service.name': '', 'service.instance.id': '' }, { case: 'empty-service-fields' });
  point('gauge', { ...primary, 'cloud.account.id': 'other' }, { case: 'raw-map-variant' });
  const missingUID = { ...primary }; delete missingUID['k8s.pod.uid'];
  point('gauge', missingUID, { case: 'missing-pod-uid' });
  point('gauge', { ...primary, 'k8s.pod.uid': '' }, { case: 'empty-pod-uid' });
  point('gauge', { ...primary, 'deployment.environment.name': environment + '-excluded' }, { case: 'wrong-environment' });
  point('gauge', primary, { age: 150000, case: 'outside-window' });
  function record(family, resource, options = {}) {
    const spanId = randomBytes(8).toString('hex'), id = family === 'spans' ? spanId : randomBytes(8).toString('hex');
    facts.records.push({ family, signal: family, id, spanId, traceId: randomBytes(16).toString('hex'), resource: { ...resource },
      resourceSchemaURL: options.schemaURL ?? schemaURL, timeNS: String(BigInt(toMs - 3000 - facts.records.length) * 1000000n),
      name: 'pod-context-work', body: 'Pod resource observed ' + id, attributes: { 'work.id': id }, case: options.case || 'primary' });
  }
  record('logs', primary); record('logs', primary); record('spans', primary);
  record('logs', primary, { schemaURL: '', case: 'literal-empty-schema' });
  record('spans', primary, { schemaURL: '', case: 'unknown-trace-schema' });
  for (const family of ['logs', 'spans']) record(family, { ...primary, 'service.name': 'pod-context-app', 'service.instance.id': 'app-worker' }, { case: 'extra-service-attrs' });
  for (const family of ['logs', 'spans']) {
    record(family, { ...primary, 'k8s.cluster.uid': '', 'service.name': 'excluded-empty-cluster' }, { case: 'empty-cluster-control' });
    record(family, { ...primary, 'k8s.namespace.name': '', 'service.name': 'excluded-empty-namespace' }, { case: 'empty-namespace-control' });
  }
  return facts;
}
function metrics(facts) { return { resourceMetrics: facts.points.map(p => {
  const point = { startTimeUnixNano: p.startNS, timeUnixNano: p.endNS, attributes: attrs(p.attributes) };
  const metric = { name: p.metric, unit: p.unit, ...(p.kind === 'gauge' ? { gauge: { dataPoints: [{ ...point, asInt: p.value }] } } : p.kind === 'sum' ? { sum: { aggregationTemporality: p.temporality, isMonotonic: p.monotonic, dataPoints: [{ ...point, asInt: p.value }] } } : { histogram: { aggregationTemporality: p.temporality, dataPoints: [{ ...point, count: p.count, sum: p.sum, explicitBounds: p.bounds, bucketCounts: p.bucketCounts }] } }) };
  return { resource: { attributes: attrs(p.resource) }, schemaUrl: p.resourceSchemaURL, scopeMetrics: [{ scope: { name: p.scope, version: p.version }, schemaUrl: p.schemaURL, metrics: [metric] }] };
}) }; }
function traces(facts) { return { resourceSpans: facts.records.filter(r => r.family === 'spans').map(r => ({ resource: { attributes: attrs(r.resource) }, schemaUrl: r.resourceSchemaURL,
  scopeSpans: [{ scope: { name: 'ordinary.pod', version: '1' }, spans: [{ traceId: r.traceId, spanId: r.spanId, name: r.name, kind: 1,
    startTimeUnixNano: r.timeNS, endTimeUnixNano: String(BigInt(r.timeNS) + 1000000n), attributes: attrs(r.attributes), status: { code: 1 } }] }] })) }; }
function logs(facts) { return { resourceLogs: facts.records.filter(r => r.family === 'logs').map(r => ({ resource: { attributes: attrs(r.resource) }, schemaUrl: r.resourceSchemaURL,
  scopeLogs: [{ scope: { name: 'ordinary.pod', version: '1' }, logRecords: [{ timeUnixNano: r.timeNS, observedTimeUnixNano: r.timeNS,
    traceId: r.traceId, spanId: r.spanId, severityNumber: 9, severityText: 'INFO', body: { stringValue: r.body }, attributes: attrs(r.attributes) }] }] })) }; }
module.exports = { fixture, metrics, traces, logs, schemaURL };
