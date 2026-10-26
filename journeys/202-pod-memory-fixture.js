// Generated ordinary OTLP, not a running kubeletstats receiver or Kubernetes agent.
const previous = require('./201-kubernetes-fixture');
const { randomBytes } = require('node:crypto');
const receiverScope = 'github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver';
const receiverVersion = '0.137.0', semconvSchema = 'https://opentelemetry.io/schemas/1.37.0';
const attrs = values => Object.entries(values).map(([key, value]) => ({ key, value: { stringValue: String(value) } }));
function fixture() {
  const facts = previous.fixture();
  facts.environment = 'pod-memory-' + randomBytes(5).toString('hex'); facts.primary['deployment.environment.name'] = facts.environment;
  for (const record of facts.records) record.resource['deployment.environment.name'] = facts.environment;
  facts.points = []; facts.receiverScope = receiverScope; facts.receiverVersion = receiverVersion;
  function point(metric, value, options = {}) {
    facts.points.push({ kind: options.kind || 'gauge', family: options.kind || 'gauge', metric, unit: options.unit ?? 'By',
      value: String(value), valueType: options.double ? 'double' : 'int', resource: { ...facts.primary, ...(options.resource || {}) },
      resourceSchemaURL: options.resourceSchemaURL ?? '', schemaURL: options.schemaURL ?? '', scope: options.scope || receiverScope,
      version: options.version || receiverVersion, attributes: options.attributes || {}, flags: options.flags || 0,
      startNS: '0', endNS: String(BigInt(facts.toMs - (options.age || 3000)) * 1000000n),
      temporality: 2, monotonic: false, case: options.case || 'primary' });
  }
  point('k8s.pod.memory.usage', 65536, { age: 20000 }); point('k8s.pod.memory.usage', 32768);
  point('k8s.pod.memory.working_set', 16384);
  point('k8s.pod.memory.usage', 999999, { resourceSchemaURL: semconvSchema, case: 'same-map-different-resource-schema' });
  point('k8s.pod.memory.working_set', 888888, { resourceSchemaURL: semconvSchema, case: 'same-map-different-resource-schema' });
  const otherCluster = { 'k8s.cluster.uid': 'cluster-2', 'k8s.namespace.name': 'team-b' };
  point('k8s.pod.memory.usage', 2222, { resource: otherCluster, case: 'same-uid-other-cluster' });
  point('k8s.pod.memory.working_set', 1111, { resource: otherCluster, case: 'same-uid-other-cluster' });
  point('k8s.pod.memory.usage', 8192, { resource: { 'k8s.pod.uid': 'uid-b' }, case: 'partial-usage-only' });
  point('k8s.pod.memory.usage', 4096, { resource: { 'k8s.pod.uid': 'stale' }, age: 90000, case: 'stale' });
  for (const value of [4096, 8192]) point('k8s.pod.memory.usage', value, { resource: { 'k8s.pod.uid': 'conflict' }, case: 'conflict' });
  point('k8s.pod.memory.usage', 0, { resource: { 'k8s.pod.uid': 'missing-data' }, flags: 1, case: 'missing-data' });
  point('k8s.pod.memory.usage', -1, { resource: { 'k8s.pod.uid': 'negative' }, case: 'negative-bytes' });
  point('k8s.pod.memory.usage', 1.5, { resource: { 'k8s.pod.uid': 'fractional' }, double: true, case: 'fractional-bytes' });
  point('k8s.pod.memory.usage', '9007199254740992', { resource: { 'k8s.pod.uid': 'unsafe' }, case: 'unsafe-bytes' });
  const semconv = { resource: { 'k8s.pod.uid': 'semconv' }, resourceSchemaURL: semconvSchema, schemaURL: semconvSchema, scope: 'ordinary.pod.memory', version: 'fixture-1', case: 'semconv-usage' };
  point('k8s.pod.memory.usage', 12345, semconv); point('k8s.pod.memory.working_set', 54321, { ...semconv, case: 'unsupported-semconv-working-set' });
  point('k8s.pod.memory.usage', 777, { version: 'v0.137.0', case: 'unproven-v-prefix' });
  point('k8s.pod.memory.usage', 666, { scope: 'kubeletstatsreceiver', case: 'short-scope' });
  point('k8s.pod.memory.usage', 555, { unit: 'MiBy', case: 'wrong-unit' });
  point('k8s.pod.memory.usage', 444, { schemaURL: semconvSchema, case: 'receiver-different-scope-schema' });
  point('k8s.pod.phase', 2, { unit: '1', case: 'unsupported-phase' });
  point('k8s.container.ready', 1, { unit: '', case: 'unsupported-ready' });
  point('k8s.pod.memory.usage', 333, { kind: 'sum', case: 'wrong-kind' });
  point('k8s.pod.memory.usage', 456, { resource: { 'k8s.namespace.name': '' }, case: 'empty-namespace' });
  point('k8s.pod.memory.usage', 457, { resource: { 'k8s.cluster.uid': '' }, case: 'empty-cluster' });
  point('k8s.pod.memory.usage', 458, { resource: { 'service.name': '', 'service.instance.id': '' }, case: 'empty-service-fields' });
  point('k8s.pod.memory.usage', 459, { resource: { 'k8s.pod.uid': '' }, case: 'empty-pod-uid' });
  point('k8s.pod.memory.usage', 460, { case: 'missing-pod-uid' }); delete facts.points.at(-1).resource['k8s.pod.uid'];
  return facts;
}
function metrics(facts) { return { resourceMetrics: facts.points.map(p => {
  const value = p.valueType === 'double' ? { asDouble: Number(p.value) } : { asInt: p.value };
  const dataPoint = { attributes: attrs(p.attributes), startTimeUnixNano: p.startNS, timeUnixNano: p.endNS, flags: p.flags, ...value };
  const body = p.kind === 'gauge' ? { gauge: { dataPoints: [dataPoint] } } : { sum: { aggregationTemporality: p.temporality, isMonotonic: p.monotonic, dataPoints: [dataPoint] } };
  return { resource: { attributes: attrs(p.resource) }, schemaUrl: p.resourceSchemaURL, scopeMetrics: [{ scope: { name: p.scope, version: p.version }, schemaUrl: p.schemaURL, metrics: [{ name: p.metric, unit: p.unit, ...body }] }] };
}) }; }
module.exports = { fixture, metrics, traces: previous.traces, logs: previous.logs, receiverScope, receiverVersion, semconvSchema };
