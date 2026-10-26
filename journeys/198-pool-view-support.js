const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path'), crypto = require('node:crypto');
const { fixture, traces, metrics } = require('./198-pool-view-fixture');
const shared = require('./187-http-dashboard-support');
const { open } = require('./197-pool-signals-support');
const pure = new Set(['/api/pool-signals/plan', '/api/pool-signals/view']);
function observe(page) { const o = shared.observe(page); page.on('request', r => { if (pure.has(new URL(r.url()).pathname)) { const e = o.requests.at(-1); if (pure.has(e?.endpoint)) e.effect = 'pure'; } }); return { ...o, source: () => o.requests.filter(r => r.effect === 'source').length, poolPlans: () => o.requests.filter(r => r.endpoint === '/api/pool-signals/plan').length, views: () => o.requests.filter(r => r.endpoint === '/api/pool-signals/view').length }; }
async function seed(request) {
 const facts = fixture(), rawOTLP = { traces: traces(facts), metrics: metrics(facts) }, acknowledgements = [];
 for (const [signal, body] of Object.entries(rawOTLP)) acknowledgements.push({ signal, response: await shared.json(await request.post('http://127.0.0.1:4318/v1/' + signal, { data: body })) });
 let ready, discovery;
 await expect.poll(async () => { ready = await shared.json(await request.post(shared.base + '/api/explore', { data: { signal: 'spans', environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs, limit: 100 } })); return ready.count; }, { timeout: 60000 }).toBe('120');
 await expect.poll(async () => { discovery = await shared.json(await request.post(shared.base + '/api/metric-discovery', { data: { environment: facts.environment, service: facts.service, namespace: { mode: 'value', value: facts.namespace }, fromMs: facts.fromMs, toMs: facts.toMs } })); return discovery.instruments.length; }, { timeout: 60000 }).toBe(4);
 return { ...facts, rawOTLP, acknowledgements, ready, discovery };
}
async function write(facts, name, data) { const dir = path.resolve('artifacts/pool-view', facts.environment); await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2)); console.log('198 artifact:', path.join(dir, name)); return dir; }
async function captureBytes(facts, id, name) {
 expect(id).toMatch(/^[a-f0-9]{64}$/); const dir = path.resolve('artifacts/pool-view', facts.environment); await fs.mkdir(dir, { recursive: true }); const env = { ...process.env };
 for (const k of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete env[k];
 await require('node:util').promisify(require('node:child_process').execFile)('docker', ['--host=unix:///var/run/docker.sock', 'cp', 'grammar-observability-packages-1:/app/artifacts/investigations/exploration/' + id + '.json', path.join(dir, name)], { env });
 const bytes = await fs.readFile(path.join(dir, name)); return { file: name, bytes: bytes.length, sha256: crypto.createHash('sha256').update(bytes).digest('hex'), data: JSON.parse(bytes) };
}
module.exports = { ...shared, open, observe, seed, write, captureBytes };
