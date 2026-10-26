const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path'), crypto = require('node:crypto');
const { fixture, traces, metrics } = require('./196-pool-fixture');
const { base, json, observe, loaded } = require('./187-http-dashboard-support');
async function seed(request) {
  const facts = fixture(), rawOTLP = { traces: traces(facts), metrics: metrics(facts) }, acknowledgements = [];
  for (const [signal, body] of Object.entries(rawOTLP)) acknowledgements.push({ signal, body: await json(await request.post('http://127.0.0.1:4318/v1/' + signal, { data: body })) });
  let ready, discovery;
  await expect.poll(async () => { ready = await json(await request.post(base + '/api/explore', { data: { signal: 'spans', environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs, limit: 100 } })); return ready.count; }, { timeout: 60000 }).toBe('120');
  await expect.poll(async () => { discovery = await json(await request.post(base + '/api/metric-discovery', { data: { environment: facts.environment, service: facts.service, namespace: { mode: 'value', value: facts.namespace }, fromMs: facts.fromMs, toMs: facts.toMs } })); return discovery.instruments.length; }, { timeout: 60000 }).toBe(4);
  return { ...facts, rawOTLP, acknowledgements, ready, discovery };
}
async function write(facts, name, data) { const dir = path.resolve('artifacts/pool-connections', facts.environment); await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2)); console.log('196 artifact:', path.join(dir, name)); return dir; }
async function captureBytes(facts, id, name) {
  expect(id).toMatch(/^[a-f0-9]{64}$/); const dir = path.resolve('artifacts/pool-connections', facts.environment); await fs.mkdir(dir, { recursive: true });
  const env = { ...process.env }; for (const key of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete env[key];
  await require('node:util').promisify(require('node:child_process').execFile)('docker', ['--host=unix:///var/run/docker.sock', 'cp', 'grammar-observability-packages-1:/app/artifacts/investigations/exploration/' + id + '.json', path.join(dir, name)], { env });
  const bytes = await fs.readFile(path.join(dir, name)); return { filename: name, bytes: bytes.length, sha256: crypto.createHash('sha256').update(bytes).digest('hex'), data: JSON.parse(bytes) };
}
function metadataControlRequest(scope, service, observed, kind) {
  const supplied = structuredClone(observed), streams = observed.capabilityPlan.coverage.streams;
  const incomplete = streams.filter(s => !s.stream.attributes['db.client.connection.pool.name'] || !s.stream.attributes['db.client.connection.state']);
  expect(incomplete).toHaveLength(2);
  const positive = streams.find(s => s.stream.attributes['db.client.connection.pool.name'] && s.stream.attributes['db.client.connection.state']);
  const selected = kind === 'missing-keys' ? incomplete : [...Array.from({ length: 25 }, () => incomplete[0]), positive];
  // Keep exact stream identities; omit redundant per-key display/coverage metadata.
  supplied.capabilityPlan.coverage = { state: 'observed', totalStreams: selected.length, omittedStreams: 0, dimensions: [], streams: selected.map(s => ({ stream: structuredClone(s.stream) })) };
  supplied.capabilityPlan.hints = []; supplied.capabilityPlan.reasons = [];
  const request = { scope, service, signals: ['metrics'], fields: [], instruments: [supplied], profile: 'all-spans' };
  expect(Buffer.byteLength(JSON.stringify(request))).toBeLessThanOrEqual(16384);
  return request;
}
module.exports = { base, json, observe, loaded, seed, write, captureBytes, metadataControlRequest };
