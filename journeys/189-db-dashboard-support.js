const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { base, json, observe, loaded, snapshot } = require('./187-http-dashboard-support');
const { fixture, traces, metrics, expected } = require('./189-db-dashboard-fixture');
async function seed(request) {
  const facts = fixture();
  const traceAck = await json(await request.post('http://127.0.0.1:4318/v1/traces', { data: traces(facts) }));
  const metricAck = await json(await request.post('http://127.0.0.1:4318/v1/metrics', { data: metrics(facts) }));
  let ready, nativeReady;
  await expect.poll(async () => {
    ready = await json(await request.post(base + '/api/explore', { data: { signal: 'spans', environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs, limit: 100, groupBy: [{ kind: 'column', key: 'service' }] } }));
    return ready.count;
  }, { timeout: 60000 }).toBe(String(facts.records.length));
  await expect.poll(async () => {
    nativeReady = await json(await request.post(base + '/api/metric-discovery', { data: { environment: facts.environment, service: 'db-api', namespace: { mode: 'value', value: 'commerce' }, fromMs: facts.fromMs, toMs: facts.toMs } }));
    return nativeReady.instruments?.length || 0;
  }, { timeout: 60000 }).toBe(3);
  return { ...facts, traceAck, metricAck, ready, nativeReady };
}
function canonical(value) {
  return Array.isArray(value) ? value.map(canonical) : value && typeof value === 'object' ? Object.fromEntries(Object.keys(value).sort().map(key => [key, canonical(value[key])])) : value;
}
function findResult(observation, question) {
  const key = JSON.stringify(canonical(question));
  return [...observation.responses].reverse().find(value => value.endpoint === '/api/record-aggregates' && JSON.stringify(canonical(value.body.question)) === key)?.body;
}
function assertPopulation(facts, result, profile, service = 'db-api', attributes = {}) {
  const records = expected(facts, profile, service).filter(record => record.timeMs >= result.question.fromMs && record.timeMs < result.question.toMs && Object.entries(attributes).every(([key, value]) => record.attributes[key] === value));
  expect(result.count).toBe(String(records.length));
  const rows = result.rows;
  if (result.version === 'goo/record-aggregates/v1') {
    for (const key of ['rows', 'fields', 'plan', 'resultToken']) expect(result[key]).toBeUndefined();
  } else expect(rows).toHaveLength(Math.min(100, records.length));
  const byID = new Map(records.map(record => [record.id, record]));
  expect(new Set((rows || []).map(row => row.SpanId)).size).toBe((rows || []).length);
  for (const row of rows || []) {
    const record = byID.get(row.SpanId);
    expect(record, 'unexpected source span ' + row.SpanId).toBeTruthy();
    expect(row.TraceId).toBe(record.traceId);
    expect(row.duration_ns).toBe(record.durationNs);
    expect(row.SpanAttributes).toEqual(record.attributes);
  }
  if (result.duration) {
    expect(result.duration.count).toBe(String(records.length));
    expect(result.duration.sumNs).toBe(String(records.reduce((sum, record) => sum + BigInt(record.durationNs), 0n)));
  }
  if (result.spanStatus) expect(result.spanStatus.error).toBe(String(records.filter(record => record.status === 2).length));
  return records;
}
async function write(facts, name, data) {
  const dir = path.resolve('artifacts/db-client-dashboard', facts.environment);
  await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2));
  console.log('189 artifact:', path.join(dir, name)); return dir;
}
module.exports = { base, json, observe, loaded, snapshot, seed, findResult, assertPopulation, write };
