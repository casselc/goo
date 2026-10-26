const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path'), crypto = require('node:crypto');
const { fixture, traces, metrics } = require('./197-pool-signals-fixture');
const shared = require('./187-http-dashboard-support');
function observe(page) { const o = shared.observe(page); page.on('request', request => { if (new URL(request.url()).pathname === '/api/pool-signals/plan') { const entry = o.requests.at(-1); if (entry?.endpoint === '/api/pool-signals/plan') entry.effect = 'pure'; } }); return { ...o, source: () => o.requests.filter(r => r.effect === 'source').length, plans: () => o.requests.filter(r => r.effect === 'pure').length }; }
async function seed(request) {
  const facts = fixture(), rawOTLP = { traces: traces(facts), metrics: metrics(facts) }, acknowledgements = [];
  for (const [signal, body] of Object.entries(rawOTLP)) acknowledgements.push({ signal, body: await shared.json(await request.post('http://127.0.0.1:4318/v1/' + signal, { data: body })) });
  let ready, discovery;
  await expect.poll(async () => { ready = await shared.json(await request.post(shared.base + '/api/explore', { data: { signal: 'spans', environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs, limit: 100 } })); return ready.count; }, { timeout: 60000 }).toBe('120');
  await expect.poll(async () => { discovery = await shared.json(await request.post(shared.base + '/api/metric-discovery', { data: { environment: facts.environment, service: facts.service, namespace: { mode: 'value', value: facts.namespace }, fromMs: facts.fromMs, toMs: facts.toMs } })); return discovery.instruments.length; }, { timeout: 60000 }).toBe(7);
  return { ...facts, rawOTLP, acknowledgements, ready, discovery };
}
async function write(facts, name, data) { const dir = path.resolve('artifacts/pool-signals', facts.environment); await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2)); console.log('197 artifact:', path.join(dir, name)); return dir; }
async function captureBytes(facts, id, name) {
  expect(id).toMatch(/^[a-f0-9]{64}$/); const dir = path.resolve('artifacts/pool-signals', facts.environment); await fs.mkdir(dir, { recursive: true });
  const env = { ...process.env }; for (const key of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete env[key];
  await require('node:util').promisify(require('node:child_process').execFile)('docker', ['--host=unix:///var/run/docker.sock', 'cp', 'grammar-observability-packages-1:/app/artifacts/investigations/exploration/' + id + '.json', path.join(dir, name)], { env });
  const bytes = await fs.readFile(path.join(dir, name)); return { filename: name, bytes: bytes.length, sha256: crypto.createHash('sha256').update(bytes).digest('hex'), data: JSON.parse(bytes) };
}
async function open(page, facts, observation) {
  await page.goto(shared.base + '/#services'); await page.locator('#inventoryEnvironment').fill(facts.environment); await page.locator('#minutes').fill('2');
  let pending = page.waitForResponse(r => r.url().endsWith('/api/discovery')); await page.locator('#inventoryRun').click(); const catalog = await shared.json(await pending);
  pending = page.waitForResponse(r => r.url().endsWith('/api/dashboards/plan')); await page.getByRole('button', { name: 'Open service inventory / pool-api', exact: true }).click();
  const response = await pending, plan = await shared.json(response); await shared.loaded(page, 'http-server'); await observation.settle();
  return { catalog, request: response.request().postDataJSON(), plan, countResult: observation.responses.find(r => r.endpoint === '/api/measurements' && r.body?.question.definition.source.metric === 'db.client.connection.count').body };
}
module.exports = { ...shared, observe, seed, write, captureBytes, open };
