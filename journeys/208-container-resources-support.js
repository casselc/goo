const shared = require('./207-container-cpu-support');
const fixture = require('./208-container-resources-fixture');
const fs = require('node:fs/promises'), path = require('node:path');
const { expect } = require('@playwright/test');
function observe(page) {
  const observed = shared.observe(page);
  page.on('request', request => {
    const endpoint = new URL(request.url()).pathname;
    const entry = observed.requests.at(-1);
    if (entry && entry.endpoint === endpoint && request.method() === 'POST' &&
        ['/api/container-memory/view', '/api/container-cpu/view'].includes(endpoint)) entry.effect = 'composition';
  });
  return { ...observed,
    compositions: () => observed.requests.filter(r => r.effect === 'composition').length };
}
async function seed(request) {
  const facts = fixture.fixture();
  const rawOTLP = { metrics: fixture.metrics(facts), traces: fixture.traces(facts), logs: fixture.logs(facts) };
  const acknowledgements = [];
  for (const [signal, data] of Object.entries(rawOTLP)) acknowledgements.push({ signal,
    response: await shared.json(await request.post('http://127.0.0.1:4318/v1/' + signal, { data })) });
  let ready;
  await expect.poll(async () => {
    ready = await shared.json(await request.post(shared.base + '/api/pod-discovery', {
      data: { environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs } }));
    return ready.observations?.reduce((sum, row) => sum + row.families.reduce((n, f) => n + Number(f.count), 0), 0);
  }, { timeout: 60000 }).toBe(facts.points.length);
  shared.assertInventory(facts, ready);
  return { ...facts, rawOTLP, acknowledgements, ready };
}
async function write(facts, name, data) {
  const dir = path.resolve('artifacts/container-resources', facts.environment);
  await fs.mkdir(dir, { recursive: true });
  await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2));
  console.log('208 artifact:', path.join(dir, name));
  return dir;
}
module.exports = { ...shared, observe, seed, write };
async function openResources(page, row) {
  const d = page.waitForResponse(r => r.url().endsWith('/api/pod-metric-discovery') && r.request().method() === 'POST');
  const p = page.waitForResponse(r => r.url().endsWith('/api/pod-dashboard/plan') && r.request().method() === 'POST');
  await page.locator('[data-container-resources-open="' + row.key + '"]').click();
  const dr = await d, pr = await p;
  return { discoveryRequest: dr.request().postDataJSON(), discovery: await shared.json(dr), planRequest: pr.request().postDataJSON(), plan: await shared.json(pr) };
}
async function discoverVariant(page, panel, key) {
  const details = panel.locator('.container-resources-sources');
  if (!await details.evaluate(e => e.open)) await details.locator(':scope > summary').click();
  await panel.getByLabel('Container resource source variant', { exact: true }).selectOption(key);
  const d = page.waitForResponse(r => r.url().endsWith('/api/pod-metric-discovery') && r.request().method() === 'POST');
  const p = page.waitForResponse(r => r.url().endsWith('/api/pod-dashboard/plan') && r.request().method() === 'POST');
  await panel.getByRole('button', { name: 'Discover source', exact: true }).click();
  const dr = await d, pr = await p;
  return { discoveryRequest: dr.request().postDataJSON(), discovery: await shared.json(dr), planRequest: pr.request().postDataJSON(), plan: await shared.json(pr) };
}
const familyPanel = (panel, family) => panel.locator('[data-container-' + family + ']');
const loadIntents = [];
async function loadFamily(page, panel, family) {
  const target = familyPanel(panel, family);
  await shared.openConfiguration(target);
  const pending = page.waitForResponse(r => r.url().endsWith('/api/container-' + family + '/view') && r.request().method() === 'POST');
  const selection = await target.locator('.container-memory-choices select').evaluateAll(ns=>ns.map(n=>({label:n.getAttribute('aria-label'),key:n.value})));
  loadIntents.push({phase:'invoke',family,selection,selected:selection.filter(r=>r.key).length});
  await target.locator('[data-container-' + family + '-load]').click();
  const response = await pending;
  const result = { request: response.request().postDataJSON(), response: await shared.json(response) };
  loadIntents.push({phase:'complete',family,selection});
  return result;
}
async function captureBytes(facts, id, name) {
  expect(id).toMatch(/^[a-f0-9]{64}$/);
  const dir = path.resolve('artifacts/container-resources', facts.environment);
  await fs.mkdir(dir, { recursive: true });
  const env = { ...process.env };
  for (const key of ['DOCKER_HOST','DOCKER_CONTEXT','DOCKER_TLS','DOCKER_TLS_VERIFY','DOCKER_CERT_PATH']) delete env[key];
  await require('node:util').promisify(require('node:child_process').execFile)('docker', ['--host=unix:///var/run/docker.sock','cp','grammar-observability-packages-1:/app/artifacts/investigations/exploration/' + id + '.json',path.join(dir,name)], { env });
  const bytes = await fs.readFile(path.join(dir,name));
  return { file:name, bytes:bytes.length, sha256:require('node:crypto').createHash('sha256').update(bytes).digest('hex'), data:JSON.parse(bytes) };
}
async function snapshot(panel) {
  return panel.evaluate(e => ({ text:e.textContent, families: [...e.querySelectorAll('[data-container-memory],[data-container-cpu]')].map(n => ({ family:n.hasAttribute('data-container-cpu')?'cpu':'memory', text:n.textContent, html:n.innerHTML, rows:[...n.querySelectorAll('tr[data-memory-role],tr[data-cpu-role]')].map(r=>({role:r.dataset.memoryRole||r.dataset.cpuRole,state:r.dataset.state,text:r.textContent})), configurationOpen:n.querySelector('details').open,pending:n.querySelector('[role=status]')?.textContent?.startsWith('Reading '),completed:!!n.querySelector('.container-memory-view,.container-cpu-view') })) }));
}
function primaryRow(facts, initial) {
  return initial.response.observations.find(r => r.resourceSchemaURL === '' && shared.canonical(r.resource) === shared.canonical(facts.points.find(p=>p.case==='primary-usage').resource));
}
Object.assign(module.exports, { openResources, discoverVariant, familyPanel, loadFamily, captureBytes, snapshot, primaryRow, loadIntents });
