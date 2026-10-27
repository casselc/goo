const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const shared = require('./187-http-dashboard-support'), pod = require('./201-kubernetes-support'), fixture = require('./204b-container-readiness-fixture');
function observe(page) {
  const result = shared.observe(page);
  page.on('request', request => {
    const endpoint = new URL(request.url()).pathname; if (!endpoint.startsWith('/api/')) return;
    const entry = result.requests.at(-1); if (!entry || entry.endpoint !== endpoint) return;
    if (request.method() === 'POST' && endpoint === '/api/pod-dashboard/plan') entry.effect = 'plan';
    entry.source = result.requests.filter(r => r.effect === 'source').length; entry.plans = result.requests.filter(r => r.effect === 'plan').length;
  });
  return { ...result, source: () => result.requests.filter(r => r.effect === 'source').length, plans: () => result.requests.filter(r => r.effect === 'plan').length };
}
async function seed(request) {
  const facts = fixture.fixture(), rawOTLP = { metrics: fixture.metrics(facts), traces: fixture.traces(facts), logs: fixture.logs(facts) }, acknowledgements = [];
  for (const [signal, data] of Object.entries(rawOTLP)) acknowledgements.push({ signal, response: await shared.json(await request.post('http://127.0.0.1:4318/v1/' + signal, { data })) });
  let ready; const population = facts.points.length + facts.records.length;
  await expect.poll(async () => {
    ready = await shared.json(await request.post(shared.base + '/api/pod-discovery', { data: { environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs } }));
    return ready.observations?.reduce((sum, row) => sum + row.families.reduce((s, family) => s + Number(family.count), 0), 0);
  }, { timeout: 60000 }).toBe(population);
  pod.assertInventory(facts, ready); return { ...facts, rawOTLP, acknowledgements, ready };
}
async function write(facts, name, data) { const dir = path.resolve('artifacts/container-readiness', facts.environment); await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2)); console.log('204b artifact:', path.join(dir, name)); return dir; }
async function openMetrics(page, row) {
  const discoveryWait = page.waitForResponse(r => r.url().endsWith('/api/pod-metric-discovery') && r.request().method() === 'POST');
  const planWait = page.waitForResponse(r => r.url().endsWith('/api/pod-dashboard/plan') && r.request().method() === 'POST');
  await page.locator('button[data-pod-observe="' + row.key + '"]').click();
  const discoveryResponse = await discoveryWait, discovery = await shared.json(discoveryResponse), planResponse = await planWait, plan = await shared.json(planResponse);
  const panel = page.locator('.pod-metrics-panel[data-pod-metrics="' + row.key + '"]');
  await expect(panel.locator('[role=status]')).toContainText('Independent reported pod observations');
  return { discoveryRequest: discoveryResponse.request().postDataJSON(), discovery, planRequest: planResponse.request().postDataJSON(), plan };
}
function assertGaugeRaw(facts, result) {
  const binding = result.question.definition.source, stream = binding.exactStream;
  expect(result.question.definition.kind).toBe('gauge'); expect(binding.resource).toEqual({ kind: 'kubernetes-pod', id: stream.resource['k8s.pod.uid'] });
  expect(binding.services || []).toEqual([]); expect(result.gauges.series).toHaveLength(1); expect(result.gauges.series[0].stream).toEqual(stream);
  const raw = facts.points.filter(p => p.kind === 'gauge' && p.metric === stream.name && p.unit === stream.unit && p.schemaURL === (stream.schemaURL || '') && p.resourceSchemaURL === (stream.resourceSchemaURL || '') && p.scope === stream.scope && p.version === (stream.version || '') && pod.canonical(p.resource) === pod.canonical(stream.resource) && pod.canonical(p.attributes) === pod.canonical(stream.attributes || {}));
  expect(raw.length).toBeGreaterThan(0);
  for (const sample of result.gauges.series[0].samples) {
    const at = raw.filter(p => p.endNS === sample.atNS); expect(at.length).toBeGreaterThan(0);
    if (['observed', 'stale'].includes(sample.state)) { expect(at.some(p => Number(p.value) === sample.value)).toBe(true); }
    else expect(sample.value).toBeUndefined();
  }
  return raw;
}
function assertRaw(facts,result) {
 if(result.question.definition.kind!=='state') return assertGaugeRaw(facts,result);
 const definition=result.question.definition, stream=definition.source.exactStream;
 expect(definition.stateProfile).toBe('k8scluster-container-ready/0.137.0');
 expect(stream.unit).toBe(''); expect(definition.source.unit ?? '').toBe('');
 expect(definition.source.metric).toBe(stream.name);expect(definition.source.scope).toBe(stream.scope);
 expect(definition.source.version).toBe(stream.version);expect(definition.source.schemaURL).toBe(stream.schemaURL);
 expect(definition.source.resource).toEqual({kind:'kubernetes-container',id:stream.resource['container.id']});
 expect(definition.source.services||[]).toEqual([]); expect(result.states.series).toHaveLength(1);
 expect(result.states.series[0].stream).toEqual(stream);
 const raw=facts.points.filter(p=>p.kind==='gauge'&&p.metric===stream.name&&p.unit===stream.unit&&p.scope===stream.scope&&p.version===(stream.version||'')&&p.schemaURL===(stream.schemaURL||'')&&p.resourceSchemaURL===(stream.resourceSchemaURL||'')&&pod.canonical(p.resource)===pod.canonical(stream.resource)&&pod.canonical(p.attributes)===pod.canonical(stream.attributes||{}));
 expect(raw.length).toBeGreaterThan(0);
 const labels={0:'Not ready',1:'Ready'};
 for(const sample of result.states.series[0].samples) {
  const at=raw.filter(p=>p.endNS===sample.atNS);expect(at.length).toBeGreaterThan(0);
  if(['observed','stale'].includes(sample.state)) {
   expect(at.some(p=>Number(p.value)===sample.code)).toBe(true);
   const supported=Number.isInteger(sample.code)&&sample.code>=0&&sample.code<=1;
   expect(sample.codeState).toBe(supported?'supported':'unsupported');
   if(supported)expect(sample.label).toBe(labels[sample.code]);else expect(sample.label||'').toBe('');
  } else {expect(sample.code).toBeUndefined();expect(sample.label||'').toBe('');}
 }
 return raw;
}
async function captureBytes(facts, id, name) {
  expect(id).toMatch(/^[a-f0-9]{64}$/); const dir = path.resolve('artifacts/container-readiness', facts.environment); await fs.mkdir(dir, { recursive: true });
  const env = { ...process.env }; for (const key of ['DOCKER_HOST','DOCKER_CONTEXT','DOCKER_TLS','DOCKER_TLS_VERIFY','DOCKER_CERT_PATH']) delete env[key];
  await require('node:util').promisify(require('node:child_process').execFile)('docker', ['--host=unix:///var/run/docker.sock','cp','grammar-observability-packages-1:/app/artifacts/investigations/exploration/' + id + '.json',path.join(dir,name)], { env });
  const bytes = await fs.readFile(path.join(dir,name)); return { file:name, bytes:bytes.length, sha256:require('node:crypto').createHash('sha256').update(bytes).digest('hex'), data:JSON.parse(bytes) };
}
module.exports = { ...shared, observe, seed, write, discover: pod.discover, canonical: pod.canonical, assertInventory: pod.assertInventory, openMetrics, assertRaw, captureBytes };
