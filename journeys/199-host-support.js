const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const shared = require('./187-http-dashboard-support');
const fixture = require('./199-host-fixture');
function observe(page) {
  const result = shared.observe(page);
  page.on('request', request => {
    const endpoint = new URL(request.url()).pathname;
    if (endpoint === '/api/host-dashboard/plan') {
      const observed = result.requests.at(-1);
      if (observed?.endpoint === endpoint) observed.effect = 'pure';
    }
  });
  return { ...result, source: () => result.requests.filter(r => r.effect === 'source').length,
    hostPlans: () => result.requests.filter(r => r.endpoint === '/api/host-dashboard/plan').length };
}
async function seed(request) {
  const facts = fixture.fixture(), rawOTLP = { metrics: fixture.metrics(facts), traces: fixture.traces(facts), logs: fixture.logs(facts) }, acknowledgements = [];
  for (const [signal, data] of Object.entries(rawOTLP)) acknowledgements.push({ signal,
    response: await shared.json(await request.post('http://127.0.0.1:4318/v1/' + signal, { data })) });
  let ready;
  await expect.poll(async () => {
    ready = await shared.json(await request.post(shared.base + '/api/host-discovery', { data: {
      environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs } }));
    return ready.hosts?.reduce((sum, host) => sum + host.records, 0);
  }, { timeout: 60000 }).toBe(facts.points.length);
  return { ...facts, rawOTLP, acknowledgements, ready };
}
async function write(facts, name, data) {
  const dir = path.resolve('artifacts/hosts', facts.environment); await fs.mkdir(dir, { recursive: true });
  await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2)); console.log('199 artifact:', path.join(dir, name)); return dir;
}
async function captureBytes(facts, id, name) {
  expect(id).toMatch(/^[a-f0-9]{64}$/);
  const dir = path.resolve('artifacts/hosts', facts.environment); await fs.mkdir(dir, { recursive: true });
  const env = { ...process.env };
  for (const key of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete env[key];
  await require('node:util').promisify(require('node:child_process').execFile)('docker', ['--host=unix:///var/run/docker.sock', 'cp', 'grammar-observability-packages-1:/app/artifacts/investigations/exploration/' + id + '.json', path.join(dir, name)], { env });
  const bytes = await fs.readFile(path.join(dir, name));
  return { file: name, bytes: bytes.length, sha256: require('node:crypto').createHash('sha256').update(bytes).digest('hex'), data: JSON.parse(bytes) };
}
module.exports = { ...shared, observe, seed, write, captureBytes };
