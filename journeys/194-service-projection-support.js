const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { fixture, traces, metrics, logs, members, value } = require('./194-service-projection-fixture');
const { base, json, observe, loaded, snapshot } = require('./187-http-dashboard-support');
async function seed(request) {
  const facts = fixture();
  const acknowledgements = [];
  for (const [signal, payload] of [['traces', traces(facts)], ['logs', logs(facts)], ['metrics', metrics(facts)]]) acknowledgements.push({ signal, response: await json(await request.post('http://127.0.0.1:4318/v1/' + signal, { data: payload })) });
  const readiness = [];
  for (const [signal, count] of [['spans', facts.records.length], ['logs', facts.logs.length]]) {
    let result;
    await expect.poll(async () => { result = await json(await request.post(base + '/api/explore', { data: { signal, environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs, limit: 100 } })); return result.count; }, { timeout: 60000 }).toBe(String(count));
    readiness.push(result);
  }
  let nativeReady;
  await expect.poll(async () => { nativeReady = await json(await request.post(base + '/api/metric-discovery', { data: { environment: facts.environment, service: 'edge-api', namespace: { mode: 'value', value: 'commerce' }, fromMs: facts.fromMs, toMs: facts.toMs } })); return nativeReady.instruments?.length || 0; }, { timeout: 60000 }).toBe(3);
  return { ...facts, acknowledgements, readiness, nativeReady };
}
function canonical(value) { return Array.isArray(value) ? value.map(canonical) : value && typeof value === 'object' ? Object.fromEntries(Object.keys(value).sort().map(key => [key, canonical(value[key])])) : value; }
function find(observation, query) { return [...observation.responses].reverse().find(response => response.endpoint === '/api/record-aggregates' && JSON.stringify(canonical(response.body.question)) === JSON.stringify(canonical(query)))?.body; }
function duration(records) {
  const sorted = records.map(record => BigInt(record.durationNs)).sort((a, b) => a < b ? -1 : a > b ? 1 : 0), n = sorted.length;
  if (!n) return { state: 'no_data', count: '0', sumNs: '0' };
  const quantile = p => String(sorted[Math.min(Math.floor(n * p / 100), n - 1)]);
  return { state: 'observed', count: String(n), sumNs: String(sorted.reduce((sum, value) => sum + value, 0n)), minNs: String(sorted[0]), maxNs: String(sorted.at(-1)), p50Ns: quantile(50), p95Ns: quantile(95), p99Ns: quantile(99) };
}
function status(records) { return { count: String(records.length), error: String(records.filter(record => record.status === 2).length), ok: String(records.filter(record => record.status === 1).length), unset: String(records.filter(record => record.status === 0).length), unknown: String(records.filter(record => ![0, 1, 2].includes(record.status)).length) }; }
function assertAggregate(facts, result) {
  expect(result.version).toBe('goo/record-aggregates/v1');
  for (const key of ['rows', 'fields', 'plan', 'token', 'resultToken', 'hasMore', 'traceExamples']) expect(result[key]).toBeUndefined();
  const q = result.question, records = members(facts, q);
  expect(result.count).toBe(String(records.length)); expect(result.state).toBe(records.length ? 'observed' : 'no_data');
  if (q.summary) expect(result.duration).toEqual(duration(records)); else expect(result.duration).toBeUndefined();
  if (q.spanStatus) expect(result.spanStatus).toEqual(status(records)); else expect(result.spanStatus).toBeUndefined();
  const grouped = new Map();
  for (const record of records) { const v = value(record, q.groupBy[0]), key = JSON.stringify(v); if (!grouped.has(key)) grouped.set(key, { value: v, records: [] }); grouped.get(key).records.push(record); }
  expect(result.groups).toHaveLength(grouped.size);
  for (const group of result.groups) { const expected = grouped.get(JSON.stringify(group.g0)); expect(expected).toBeTruthy(); expect(group.count).toBe(String(expected.records.length)); if (q.summary) expect(group.duration).toEqual(duration(expected.records)); if (q.spanStatus) expect(group.spanStatus).toEqual(status(expected.records)); }
  if (q.bucketMs) {
    expect(result.timeSeries).toHaveLength(Math.ceil((q.toMs - q.fromMs) / q.bucketMs));
    for (const [index, point] of result.timeSeries.entries()) {
      const from = q.fromMs + index * q.bucketMs, to = Math.min(q.toMs, from + q.bucketMs), included = records.filter(record => record.timeMs >= from && record.timeMs < to);
      expect(point.fromMs).toBe(from); expect(point.toMs).toBe(to); expect(point.count).toBe(String(included.length));
      expect(point.state).toBe(included.length ? 'observed' : 'no_data');
      if (q.summary) expect(point.duration).toEqual(duration(included)); if (q.spanStatus) expect(point.spanStatus).toEqual(status(included));
    }
  } else expect(result.timeSeries).toBeUndefined();
  expect(Object.keys(result.reads).sort()).toEqual(['count', 'groups', ...(q.bucketMs ? ['timeSeries'] : [])].sort());
  for (const read of Object.values(result.reads)) { expect(read.basis).toBeTruthy(); expect(Number.isFinite(Date.parse(read.readAt))).toBe(true); }
  return records;
}
async function write(facts, name, data) { const dir = path.resolve('artifacts/service-dashboard-projection', facts.environment); await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2)); console.log('194 artifact:', path.join(dir, name)); return dir; }
async function captureBytes(facts, id, name) {
  expect(id).toMatch(/^[a-f0-9]{64}$/);
  const dir = path.resolve('artifacts/service-dashboard-projection', facts.environment); await fs.mkdir(dir, { recursive: true });
  const environment = { ...process.env }; for (const key of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete environment[key];
  await require('node:util').promisify(require('node:child_process').execFile)('docker', ['--host=unix:///var/run/docker.sock', 'cp', 'grammar-observability-packages-1:/app/artifacts/investigations/exploration/' + id + '.json', path.join(dir, name)], { env: environment });
  const bytes = await fs.readFile(path.join(dir, name)); return { filename: name, bytes: bytes.length, sha256: require('node:crypto').createHash('sha256').update(bytes).digest('hex'), data: JSON.parse(bytes) };
}
module.exports = { captureBytes, base, json, observe, loaded, snapshot, seed, find, assertAggregate, write };
