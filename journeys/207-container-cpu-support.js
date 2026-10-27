const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const shared = require('./187-http-dashboard-support'), pod = require('./201-kubernetes-support'), fixture = require('./207-container-cpu-fixture');
function observe(page) {
  const result = shared.observe(page);
  page.on('request', request => {
    const endpoint = new URL(request.url()).pathname; if (!endpoint.startsWith('/api/')) return;
    const entry = result.requests.at(-1); if (!entry || entry.endpoint !== endpoint) return;
    if (request.method() === 'POST' && endpoint === '/api/pod-dashboard/plan') entry.effect = 'plan';
    if (request.method() === 'POST' && endpoint === '/api/container-cpu/view') entry.effect = 'composition';
    entry.source = result.requests.filter(r => r.effect === 'source').length; entry.plans = result.requests.filter(r => r.effect === 'plan').length;
  });
  return { ...result, source: () => result.requests.filter(r => r.effect === 'source').length, plans: () => result.requests.filter(r => r.effect === 'plan').length, compositions:()=>result.requests.filter(r=>r.effect==='composition').length };
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
async function write(facts, name, data) { const dir = path.resolve('artifacts/container-cpu', facts.environment); await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2)); console.log('207 artifact:', path.join(dir, name)); return dir; }
async function openPodDefaults(page, row) {
  const discoveryWait = page.waitForResponse(r => r.url().endsWith('/api/pod-metric-discovery') && r.request().method() === 'POST');
  const planWait = page.waitForResponse(r => r.url().endsWith('/api/pod-dashboard/plan') && r.request().method() === 'POST');
  await page.locator('button[data-pod-observe="' + row.key + '"]').click();
  const discoveryResponse = await discoveryWait, discovery = await shared.json(discoveryResponse), planResponse = await planWait, plan = await shared.json(planResponse);
  const panel = page.locator('.pod-metrics-panel[data-pod-metrics="' + row.key + '"]');
  await expect(panel.locator('[role=status]')).toContainText('Independent observations in the fixed window');
  return { discoveryRequest: discoveryResponse.request().postDataJSON(), discovery, planRequest: planResponse.request().postDataJSON(), plan };
}
function assertRaw(facts, result) {
  const binding = result.question.definition.source, stream = binding.exactStream;
  expect(result.question.definition.kind).toBe('gauge');expect(binding.unit).toBe('{cpu}');expect(stream.unit).toBe('{cpu}'); expect(binding.resource).toEqual({ kind: 'kubernetes-container', id: stream.resource['container.id'] });
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
async function captureBytes(facts, id, name) {
  expect(id).toMatch(/^[a-f0-9]{64}$/); const dir = path.resolve('artifacts/container-cpu', facts.environment); await fs.mkdir(dir, { recursive: true });
  const env = { ...process.env }; for (const key of ['DOCKER_HOST','DOCKER_CONTEXT','DOCKER_TLS','DOCKER_TLS_VERIFY','DOCKER_CERT_PATH']) delete env[key];
  await require('node:util').promisify(require('node:child_process').execFile)('docker', ['--host=unix:///var/run/docker.sock','cp','grammar-observability-packages-1:/app/artifacts/investigations/exploration/' + id + '.json',path.join(dir,name)], { env });
  const bytes = await fs.readFile(path.join(dir,name)); return { file:name, bytes:bytes.length, sha256:require('node:crypto').createHash('sha256').update(bytes).digest('hex'), data:JSON.parse(bytes) };
}
module.exports = { ...shared, observe, seed, write, discover: pod.discover, canonical: pod.canonical, assertInventory: pod.assertInventory, openPodDefaults, assertRaw, captureBytes };
async function openCPU(page,row){const d=page.waitForResponse(r=>r.url().endsWith('/api/pod-metric-discovery')&&r.request().method()==='POST'),p=page.waitForResponse(r=>r.url().endsWith('/api/pod-dashboard/plan')&&r.request().method()==='POST');await page.locator('[data-container-cpu-open="'+row.key+'"]').click();const dr=await d,pr=await p;return{discoveryRequest:dr.request().postDataJSON(),discovery:await shared.json(dr),planRequest:pr.request().postDataJSON(),plan:await shared.json(pr)};}
async function discoverVariant(page,panel,key){await panel.getByLabel('Container CPU source variant',{exact:true}).selectOption(key);const d=page.waitForResponse(r=>r.url().endsWith('/api/pod-metric-discovery')&&r.request().method()==='POST'),p=page.waitForResponse(r=>r.url().endsWith('/api/pod-dashboard/plan')&&r.request().method()==='POST');await panel.getByRole('button',{name:'Discover source',exact:true}).click();const dr=await d,pr=await p;return{discoveryRequest:dr.request().postDataJSON(),discovery:await shared.json(dr),planRequest:pr.request().postDataJSON(),plan:await shared.json(pr)};}
async function loadCPU(page,panel){const pending=page.waitForResponse(r=>r.url().endsWith('/api/container-cpu/view')&&r.request().method()==='POST');await panel.locator('[data-container-cpu-load]').click();const response=await pending;return{request:response.request().postDataJSON(),response:await shared.json(response)};}
async function cpuSnapshot(panel){return panel.evaluate(e=>({text:e.textContent,signals:[...e.querySelectorAll('tr[data-cpu-role]')].map(r=>({role:r.dataset.cpuRole,state:r.dataset.state,text:r.textContent,title:r.querySelector('td')?.title,bars:[...r.querySelectorAll('.comparison-bar-cell')].map(c=>({value:c.dataset.value,scaleMax:c.dataset.scaleMax,fill:c.querySelector('.comparison-bar-fill')?.style.width,trackWidth:c.querySelector('.comparison-bar-track')?.getBoundingClientRect().width}))})),percentages:[...e.querySelectorAll('.container-memory-percentages span')].map(n=>({text:n.textContent,title:n.title}))}));}
Object.assign(module.exports,{openCPU,discoverVariant,loadCPU,cpuSnapshot});
async function openConfiguration(panel){const config=panel.locator('.container-memory-configuration');if(!(await config.getAttribute('open'))&&!(await config.evaluate(e=>e.open)))await config.locator(':scope > summary').click();await expect(config).toHaveAttribute('open','');return config;}
Object.assign(module.exports,{openConfiguration});
