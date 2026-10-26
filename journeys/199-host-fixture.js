// Generated ordinary OTLP telemetry, not a live operating-system scraper.
const { randomBytes } = require('node:crypto');
const schemaURL = 'https://opentelemetry.io/schemas/1.37.0';
const attrs = map => Object.entries(map).map(([key, value]) => ({ key, value: { stringValue: String(value) } }));
function fixture() {
  const toMs = Date.now(), environment = 'host-observed-' + randomBytes(5).toString('hex');
  const facts = { environment, fromMs: toMs - 120000, toMs, schemaURL, points: [], records: [] };
  const base = { 'deployment.environment.name': environment, 'host.id': 'host-a', 'host.name': 'shared-name', 'cloud.account.id': 'account-a', 'telemetry.sdk.language': 'python' };
  function point(metric, kind, unit, value, attributes = {}, options = {}) {
    facts.points.push({ metric, kind, unit, value: String(value), valueType: options.integer ? 'int' : 'double',
      resource: { ...base, ...(options.resource || {}) }, resourceSchemaURL: schemaURL,
      schemaURL: options.schemaURL || schemaURL, scope: 'ordinary.host', version: '1', attributes,
      startNS: kind === 'stock' ? String(BigInt(toMs - (options.epochAge || 90000)) * 1000000n) : '0',
      endNS: String(BigInt(toMs - (options.age || 3000)) * 1000000n), flags: options.flags || 0,
      temporality: kind === 'stock' ? 2 : 0, monotonic: false, case: options.case || 'primary' });
  }
  const cpu = { 'cpu.mode': 'user', 'cpu.logical_number': '0' };
  point('system.cpu.utilization', 'gauge', '1', '0.25', cpu);
  point('system.cpu.utilization', 'gauge', '1', '0.4', { ...cpu, 'cpu.logical_number': '1' });
  point('system.cpu.utilization', 'gauge', '1', '0.6', { ...cpu, 'cpu.mode': 'system' });
  point('system.memory.usage', 'stock', 'By', '4096', { 'system.memory.state': 'used' }, { integer: true, age: 20000 });
  point('system.memory.usage', 'stock', 'By', '2048', { 'system.memory.state': 'used' }, { integer: true, epochAge: 10000 });
  point('system.memory.usage', 'stock', 'By', '8192', { 'system.memory.state': 'free' }, { integer: true });
  point('system.cpu.utilization', 'gauge', '1', '0.8', cpu, { resource: { 'cloud.account.id': 'account-b' }, case: 'same-id-different-account' });
  point('system.cpu.utilization', 'gauge', '1', '0.35', cpu, { resource: { 'service.name': '', 'service.instance.id': '' }, case: 'present-empty-service' });
  point('system.cpu.utilization', 'gauge', '1', '0.5', cpu, { resource: { 'host.id': 'cpu-only' }, case: 'partial-cpu' });
  point('system.cpu.utilization', 'gauge', '1', '0.1', cpu, { resource: { 'host.id': 'stale' }, age: 90000, case: 'stale' });
  point('system.cpu.utilization', 'gauge', '1', '0', cpu, { resource: { 'host.id': 'missing' }, flags: 1, case: 'missing' });
  for (const value of ['0.2', '0.7']) point('system.cpu.utilization', 'gauge', '1', value, cpu, { resource: { 'host.id': 'conflict' }, case: 'conflict' });
  point('system.cpu.utilization', 'gauge', '1', '1.5', cpu, { resource: { 'host.id': 'out-of-range' }, case: 'out-of-range' });
  point('system.cpu.time', 'counter', 's', '0.125', cpu, { case: 'unsupported-counter' });
  point('system.cpu.utilization', 'gauge', '1', '0.9', cpu, { schemaURL: 'https://opentelemetry.io/schemas/1.36.0', case: 'unsupported-schema' });
  const traceId = randomBytes(16).toString('hex'), spanId = randomBytes(8).toString('hex');
  for (const signal of ['logs', 'spans']) facts.records.push({ signal, id: randomBytes(8).toString('hex'), traceId, spanId,
    timeNS: String(BigInt(toMs - 3000) * 1000000n), resource: { ...base, 'service.name': 'host-context-app', 'service.instance.id': 'app-worker' },
    attributes: { 'work.fixture': environment }, body: 'Host context observed', name: 'host-context-work' });
  return facts;
}
function metrics(facts) {
  return { resourceMetrics: facts.points.map(p => {
    const dataPoint = { startTimeUnixNano: p.startNS, timeUnixNano: p.endNS, attributes: attrs(p.attributes), flags: p.flags,
      [p.valueType === 'int' ? 'asInt' : 'asDouble']: p.valueType === 'int' ? p.value : Number(p.value) };
    const payload = p.kind === 'gauge' ? { gauge: { dataPoints: [dataPoint] } } : { sum: { aggregationTemporality: 2, isMonotonic: p.kind === 'counter', dataPoints: [dataPoint] } };
    return { resource: { attributes: attrs(p.resource) }, schemaUrl: p.resourceSchemaURL, scopeMetrics: [{ schemaUrl: p.schemaURL, scope: { name: p.scope, version: p.version }, metrics: [{ name: p.metric, unit: p.unit, ...payload }] }] };
  }) };
}
function traces(facts) { return { resourceSpans: facts.records.filter(r => r.signal === 'spans').map(r => ({ resource: { attributes: attrs(r.resource) }, schemaUrl: schemaURL, scopeSpans: [{ scope: { name: 'ordinary.host.context', version: '1' }, spans: [{ traceId: r.traceId, spanId: r.spanId, name: r.name, kind: 2, startTimeUnixNano: r.timeNS, endTimeUnixNano: String(BigInt(r.timeNS) + 1000000n), attributes: attrs(r.attributes), status: { code: 1 } }] }] })) }; }
function logs(facts) { return { resourceLogs: facts.records.filter(r => r.signal === 'logs').map(r => ({ resource: { attributes: attrs(r.resource) }, schemaUrl: schemaURL, scopeLogs: [{ scope: { name: 'ordinary.host.context', version: '1' }, logRecords: [{ timeUnixNano: r.timeNS, observedTimeUnixNano: r.timeNS, traceId: r.traceId, spanId: r.spanId, severityNumber: 9, severityText: 'INFO', body: { stringValue: r.body }, attributes: attrs(r.attributes) }] }] })) }; }
module.exports = { fixture, metrics, traces, logs, schemaURL };
