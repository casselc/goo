const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const shared = require('./187-http-dashboard-support'), fixture = require('./205-workload-context-fixture');
function observe(page) { const result=shared.observe(page);page.on('request',request=>{const endpoint=new URL(request.url()).pathname;if(!endpoint.startsWith('/api/'))return;const entry=result.requests.at(-1);if(!entry||entry.endpoint!==endpoint)return;if(request.method()==='POST'&&endpoint==='/api/pod-dashboard/plan')entry.effect='plan';entry.source=result.requests.filter(r=>r.effect==='source').length;entry.plans=result.requests.filter(r=>r.effect==='plan').length});return{...result,source:()=>result.requests.filter(r=>r.effect==='source').length,plans:()=>result.requests.filter(r=>r.effect==='plan').length}; }
function canonical(value) { return JSON.stringify(Array.isArray(value) ? value.map(v => JSON.parse(canonical(v))) : value && typeof value === 'object' ? Object.fromEntries(Object.keys(value).sort().map(k => [k, JSON.parse(canonical(value[k]))])) : value); }
function expected(facts) {
  const groups = new Map();
  const observations = [...facts.points.map(p => ({ family: p.family, resource: p.resource, schema: p.resourceSchemaURL, at: p.endNS })), ...facts.records.map(r => ({ family: r.family, resource: r.resource, schema: r.family === 'spans' ? null : r.resourceSchemaURL, at: r.timeNS }))];
  for (const row of observations) {
    if (row.resource['deployment.environment.name'] !== facts.environment || BigInt(row.at) < BigInt(facts.fromMs) * 1000000n || BigInt(row.at) >= BigInt(facts.toMs) * 1000000n) continue;
    const key = canonical([row.resource, row.schema]);
    if (!groups.has(key)) groups.set(key, { resource: row.resource, resourceSchemaURL: row.schema, families: new Map() });
    const families = groups.get(key).families;
    if (!families.has(row.family)) families.set(row.family, { family: row.family, count: '0', firstNS: row.at, lastNS: row.at });
    const family = families.get(row.family); family.count = String(BigInt(family.count) + 1n);
    if (BigInt(row.at) < BigInt(family.firstNS)) family.firstNS = row.at;
    if (BigInt(row.at) > BigInt(family.lastNS)) family.lastNS = row.at;
  }
  return [...groups.values()].map(row => ({ ...row, families: [...row.families.values()].sort((a, b) => a.family.localeCompare(b.family)) }));
}
function assertInventory(facts, data) {
  const wanted = expected(facts); expect(data.observationsOmitted).toBe(0); expect(data.observations).toHaveLength(wanted.length);
  expect(data.sourceGroups).toBe(wanted.reduce((sum, row) => sum + row.families.length, 0));
  const found = data.observations.map(row => ({ resource: row.resource, resourceSchemaURL: row.resourceSchemaURL, families: [...row.families].sort((a, b) => a.family.localeCompare(b.family)) }));
  expect(found.map(canonical).sort()).toEqual(wanted.map(canonical).sort());
  expect(new Set(data.observations.map(row => row.key)).size).toBe(wanted.length);
}
async function seed(request) {
  const facts = fixture.fixture(), rawOTLP = { metrics: fixture.metrics(facts), traces: fixture.traces(facts), logs: fixture.logs(facts) }, acknowledgements = [];
  for (const [signal, data] of Object.entries(rawOTLP)) acknowledgements.push({ signal, response: await shared.json(await request.post('http://127.0.0.1:4318/v1/' + signal, { data })) });
  let ready; const population = expected(facts).reduce((sum, row) => sum + row.families.reduce((s, family) => s + Number(family.count), 0), 0);
  await expect.poll(async () => {
    ready = await shared.json(await request.post(shared.base + '/api/pod-discovery', { data: { environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs } }));
    return ready.observations?.reduce((sum, row) => sum + row.families.reduce((s, family) => s + Number(family.count), 0), 0);
  }, { timeout: 60000 }).toBe(population);
  assertInventory(facts, ready); return { ...facts, rawOTLP, acknowledgements, ready };
}
async function write(facts, name, data) { const dir = path.resolve('artifacts/workload-context', facts.environment); await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2)); console.log('205 artifact:', path.join(dir, name)); return dir; }
async function discover(page, facts) {
  await page.goto(shared.base + '/#services'); await page.locator('#entitiesKubernetes').click();
  await page.getByLabel('Kubernetes environment', { exact: true }).fill(facts.environment); await page.locator('#minutes').fill('2');
  const pending = page.waitForResponse(r => r.url().endsWith('/api/pod-discovery') && r.request().method() === 'POST');
  await page.getByRole('button', { name: 'Find pods', exact: true }).click(); const response = await pending;
  return { request: response.request().postDataJSON(), response: await shared.json(response) };
}
module.exports = { ...shared, observe, seed, write, discover, assertInventory, canonical };
function workloadKey(resource,kind){const state=k=>!Object.hasOwn(resource,k)?{state:'missing'}:resource[k]===''?{state:'empty',value:''}:{state:'present',value:resource[k]};return JSON.stringify([kind,...['k8s.cluster.uid','k8s.cluster.name','k8s.namespace.name'].map(state),state('k8s.'+kind+'.uid'),state('k8s.'+kind+'.name')]);}
function expectedMembers(facts,resource,kind,signal,question){const keys=['k8s.cluster.uid','k8s.cluster.name','k8s.namespace.name','k8s.'+kind+'.uid','k8s.'+kind+'.name'];return facts.records.filter(r=>r.family===signal&&r.resource['deployment.environment.name']===facts.environment&&BigInt(r.timeNS)>=BigInt(question.fromMs)*1000000n&&BigInt(r.timeNS)<BigInt(question.toMs)*1000000n&&keys.every(k=>Object.hasOwn(resource,k)?r.resource[k]===resource[k]:!Object.hasOwn(r.resource,k))).map(r=>r.id).sort();}
function assertWorkloadQuestion(facts,resource,kind,signal,parent,q){expect(q.environment).toBe(facts.environment);expect(q.signal).toBe(signal);expect(q.fromMs).toBe(parent.fromMs);expect(q.toMs).toBe(parent.toMs);const keys=['k8s.cluster.uid','k8s.cluster.name','k8s.namespace.name','k8s.'+kind+'.uid','k8s.'+kind+'.name'],predicates=q.where.all;expect(predicates).toHaveLength(6);expect(predicates).toContainEqual({field:{kind:'resource',key:'deployment.environment.name'},op:'eq',value:facts.environment});for(const k of keys)expect(predicates).toContainEqual(Object.hasOwn(resource,k)?{field:{kind:'resource',key:k},op:'eq',value:resource[k]}:{field:{kind:'resource',key:k},op:'missing'});return expectedMembers(facts,resource,kind,signal,q);}
async function captureBytes(facts,id,name){expect(id).toMatch(/^[a-f0-9]{64}$/);const dir=path.resolve('artifacts/workload-context',facts.environment);await fs.mkdir(dir,{recursive:true});const env={...process.env};for(const k of ['DOCKER_HOST','DOCKER_CONTEXT','DOCKER_TLS','DOCKER_TLS_VERIFY','DOCKER_CERT_PATH'])delete env[k];await require('node:util').promisify(require('node:child_process').execFile)('docker',['--host=unix:///var/run/docker.sock','cp','grammar-observability-packages-1:/app/artifacts/investigations/exploration/'+id+'.json',path.join(dir,name)],{env});const bytes=await fs.readFile(path.join(dir,name));return{file:name,bytes:bytes.length,sha256:require('node:crypto').createHash('sha256').update(bytes).digest('hex'),data:JSON.parse(bytes)};}
Object.assign(module.exports,{workloadKey,expectedMembers,assertWorkloadQuestion,captureBytes});
