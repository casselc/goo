const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const shared = require('./187-http-dashboard-support');
const fixture = require('./200-runtime-fixture');
function observe(page) {
  const result = shared.observe(page);
  return { ...result, source: () => result.source, plans: () => result.plans };
}
async function seed(request) {
  const facts = fixture.fixture(), rawOTLP = { metrics: fixture.metrics(facts), traces: fixture.traces(facts) }, acknowledgements = [];
  for (const [signal, data] of Object.entries(rawOTLP)) acknowledgements.push({ signal,
    response: await shared.json(await request.post('http://127.0.0.1:4318/v1/' + signal, { data })) });
  let ready, discovery;
  await expect.poll(async () => {
    ready = await shared.json(await request.post(shared.base + '/api/explore', { data: { signal: 'spans', environment: facts.environment,
      fromMs: facts.fromMs, toMs: facts.toMs, limit: 100 } }));
    return ready.count;
  }, { timeout: 60000 }).toBe('120');
  await expect.poll(async () => {
    discovery = await shared.json(await request.post(shared.base + '/api/metric-discovery', { data: {
      environment: facts.environment, service: facts.service, namespace: { mode: 'value', value: facts.namespace }, fromMs: facts.fromMs, toMs: facts.toMs } }));
    return discovery.instruments?.some(i => i.metric === 'go.memory.used');
  }, { timeout: 60000 }).toBe(true);
  return { ...facts, rawOTLP, acknowledgements, ready, discovery };
}
async function write(facts, name, data) {
  const dir = path.resolve('artifacts/runtime-observations', facts.environment); await fs.mkdir(dir, { recursive: true });
  await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2)); console.log('200 artifact:', path.join(dir, name)); return dir;
}
async function captureBytes(facts, id, name) {
  expect(id).toMatch(/^[a-f0-9]{64}$/);
  const dir = path.resolve('artifacts/runtime-observations', facts.environment); await fs.mkdir(dir, { recursive: true });
  const env = { ...process.env };
  for (const key of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete env[key];
  await require('node:util').promisify(require('node:child_process').execFile)('docker', ['--host=unix:///var/run/docker.sock', 'cp', 'grammar-observability-packages-1:/app/artifacts/investigations/exploration/' + id + '.json', path.join(dir, name)], { env });
  const bytes = await fs.readFile(path.join(dir, name));
  return { file: name, bytes: bytes.length, sha256: require('node:crypto').createHash('sha256').update(bytes).digest('hex'), data: JSON.parse(bytes) };
}
async function discover(page, facts) {
  await page.goto(shared.base + '/#services'); await page.locator('#inventoryEnvironment').fill(facts.environment); await page.locator('#minutes').fill('2');
  const pending = page.waitForResponse(r => r.url().endsWith('/api/discovery') && r.request().method() === 'POST');
  await page.locator('#inventoryRun').click(); return shared.json(await pending);
}
async function open(page, facts, observation) {
  const catalog = await discover(page, facts), pending = page.waitForResponse(r => r.url().endsWith('/api/dashboards/plan'));
  await page.getByRole('button', { name: 'Open service ' + facts.namespace + ' / ' + facts.service, exact: true }).click();
  const response = await pending, plan = await shared.json(response);
  await expect(page.locator('.service-dashboard')).toHaveAttribute('data-state', 'loaded'); await observation.settle();
  return { catalog, request: response.request().postDataJSON(), plan };
}
function stable(value) { return JSON.stringify(sort(value)); }
function sort(value) { return Array.isArray(value) ? value.map(sort) : value && typeof value === 'object' ? Object.fromEntries(Object.keys(value).sort().map(k => [k, sort(value[k])])) : value; }
function assertRaw(facts, result) {
  const binding = result.question.definition.source, stream = binding.exactStream;
  expect(result.question.definition.stockDomain).toBe('nonnegative-integer');
  expect(stream).toBeTruthy(); expect(result.stocks.series).toHaveLength(1);
  expect(result.stocks.series[0].stream).toEqual(stream);
  const raw = facts.points.filter(p => p.metric === stream.name && p.unit === stream.unit && p.schemaURL === stream.schemaURL && p.resourceSchemaURL === stream.resourceSchemaURL && p.scope === stream.scope && p.version === stream.version && stable(p.resource) === stable(stream.resource) && stable(p.attributes) === stable(stream.attributes));
  expect(raw.length).toBeGreaterThan(0); const samples = result.stocks.series[0].samples;
  expect(samples).toHaveLength(raw.length);
  for (const sample of samples) {
    const point = raw.find(p => p.endNS === sample.atNS && p.startNS === sample.startNS);
    expect(point).toBeTruthy(); expect(sample.value).toBe(Number(point.value)); expect(sample.state).toBe('observed');
  }
  return raw;
}
module.exports = { ...shared, observe, seed, write, captureBytes, discover, open, assertRaw };
