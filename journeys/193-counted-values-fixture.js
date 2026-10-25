const { randomBytes } = require('node:crypto');
const attributes = map => Object.entries(map).map(([key, stringValue]) => ({ key, value: { stringValue } }));
const field = { kind: 'attribute', key: 'release' };
function fixture() {
  const environment = 'counted-projection-' + randomBytes(5).toString('hex'), now = Date.now(), records = [];
  function add(population, release, units = '9007199254740993') {
    const id = randomBytes(8).toString('hex');
    records.push({ id, environment, service: 'ordinary-values-api', namespace: 'inventory', timeMs: now - 15000 + records.length * 10,
      body: 'counted ' + id, attributes: { population, ...(release === null ? {} : { release }), ...(units === null ? {} : { 'work.units': units }) } });
  }
  for (let i = 0; i < 140; i++) add('large', 'common');
  for (const value of ['canary', '', null]) add('large', value);
  for (const value of ['common', 'common', 'canary', '', null]) add('small', value);
  for (const units of ['9007199254740992', 'invalid', '', null]) add('large', 'excluded', units);
  return { environment, fromMs: now - 60000, toMs: now + 1000, records };
}
function question(facts, population = 'large') {
  return { signal: 'logs', environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs, limit: 100, offset: 0,
    where: { all: [
      { field: { kind: 'column', key: 'service' }, op: 'eq', value: 'ordinary-values-api' },
      { field: { kind: 'resource', key: 'service.namespace' }, op: 'eq', value: 'inventory' },
      { field: { kind: 'attribute', key: 'population' }, op: 'eq', value: population },
      { numeric: { field: { kind: 'attribute', key: 'work.units' }, type: 'int64', parse: 'canonical-decimal', unit: '{item}' }, op: 'gte', value: '9007199254740993' },
    ] } };
}
function logs(facts) {
  return { resourceLogs: [{ resource: { attributes: attributes({ 'service.name': 'ordinary-values-api', 'service.namespace': 'inventory', 'deployment.environment.name': facts.environment }) },
    scopeLogs: [{ scope: { name: 'ordinary.counted-values', version: '1' }, logRecords: facts.records.map(record => ({ timeUnixNano: String(BigInt(record.timeMs) * 1000000n),
      body: { stringValue: record.body }, severityNumber: 9, severityText: 'INFO', attributes: attributes(record.attributes) })) }] }] };
}
module.exports = { fixture, question, logs, field };
